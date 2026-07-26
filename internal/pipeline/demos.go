package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// Demos stage outputs, under the lesson's generated dir.
const (
	DemosDirName         = "demos"
	DemoManifestFileName = "manifest.json" // inside demos/
)

// tapeRenderTimeout bounds one VHS recording.
const tapeRenderTimeout = 5 * time.Minute

// DemoSpec is one [DEMO: description] marker from lesson.md.
type DemoSpec struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// DemoEntry is one rendered demo in the manifest consumed by scenegraph.
type DemoEntry struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	// Path is relative to the lesson's generated dir, e.g. "demos/demo-1.mp4".
	Path       string `json:"path"`
	DurationMs int    `json:"durationMs"`
}

// DemoManifest is the persisted demos/manifest.json.
type DemoManifest struct {
	Demos []DemoEntry `json:"demos"`
}

var demoMarkerRe = regexp.MustCompile(`\[DEMO:\s*([^\]]+)\]`)

// extractDemoMarkers finds [DEMO: ...] markers in outline order and assigns
// stable ids demo-1, demo-2, ...
func extractDemoMarkers(body string) []DemoSpec {
	matches := demoMarkerRe.FindAllStringSubmatch(body, -1)
	specs := make([]DemoSpec, 0, len(matches))
	for i, m := range matches {
		specs = append(specs, DemoSpec{
			ID:          fmt.Sprintf("demo-%d", i+1),
			Description: strings.TrimSpace(m[1]),
		})
	}
	return specs
}

// TapeRunner validates and renders VHS tapes. Implementations: docker
// sandbox, host vhs, fakes in tests.
type TapeRunner interface {
	Name() string
	// Validate checks tape syntax (vhs validate). tapePath is relative to
	// workDir.
	Validate(ctx context.Context, workDir, tapePath string) error
	// RenderTape records the tape; its Output directive lands relative to
	// workDir.
	RenderTape(ctx context.Context, workDir, tapePath string) error
}

// DockerTapeRunner runs vhs inside the coursesmith sandbox, so the demo's
// python3 commands execute in isolation.
type DockerTapeRunner struct{}

func (DockerTapeRunner) Name() string { return "docker-sandbox" }

func (DockerTapeRunner) vhs(ctx context.Context, workDir string, args ...string) error {
	abs, err := filepath.Abs(workDir)
	if err != nil {
		return fmt.Errorf("resolving %s: %w", workDir, err)
	}
	full := append([]string{
		"run", "--rm",
		"--network", "none",
		"-v", abs + ":/work",
		"-w", "/work",
		SandboxImage,
		"vhs",
	}, args...)
	cmd := exec.CommandContext(ctx, "docker", full...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("vhs %s: %w\n%s", strings.Join(args, " "), err, tailLines(stderr.String(), 10))
	}
	return nil
}

func (r DockerTapeRunner) Validate(ctx context.Context, workDir, tapePath string) error {
	return r.vhs(ctx, workDir, "validate", tapePath)
}

func (r DockerTapeRunner) RenderTape(ctx context.Context, workDir, tapePath string) error {
	return r.vhs(ctx, workDir, tapePath)
}

// HostTapeRunner runs a locally installed vhs — demos then execute
// UNSANDBOXED on the host.
type HostTapeRunner struct{}

func (HostTapeRunner) Name() string { return "host-vhs (UNSANDBOXED)" }

func (HostTapeRunner) vhs(ctx context.Context, workDir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "vhs", args...)
	cmd.Dir = workDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("vhs %s: %w\n%s", strings.Join(args, " "), err, tailLines(stderr.String(), 10))
	}
	return nil
}

func (r HostTapeRunner) Validate(ctx context.Context, workDir, tapePath string) error {
	return r.vhs(ctx, workDir, "validate", tapePath)
}

func (r HostTapeRunner) RenderTape(ctx context.Context, workDir, tapePath string) error {
	return r.vhs(ctx, workDir, tapePath)
}

// ResolveTapeRunner picks docker (with the sandbox image) over host vhs.
// The note is non-empty when the choice degrades isolation; a nil runner
// means demos cannot be recorded at all.
func ResolveTapeRunner(ctx context.Context) (TapeRunner, string) {
	if _, err := exec.LookPath("docker"); err == nil {
		probe, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if exec.CommandContext(probe, "docker", "image", "inspect", SandboxImage).Run() == nil {
			return DockerTapeRunner{}, ""
		}
	}
	if _, err := exec.LookPath("vhs"); err == nil {
		return HostTapeRunner{}, "recording demos with host vhs, UNSANDBOXED — build the sandbox: " + sandboxBuildHelp
	}
	return nil, ""
}

// tapeHeader is the engine-owned settings block prepended to every tape.
// The LLM never writes these (lintTapeBody rejects them). The recording is
// themed to the derived dark video tokens so the demo sits natively inside
// the dark TerminalScene window instead of flashing a white rectangle, and
// sized 1440x640 (~15 lines) so short REPL sessions don't leave a mostly
// empty window on screen.
func tapeHeader(id string, theme SceneTheme) string {
	bg := theme.BgBottom
	if bg == "" {
		bg = "#11151c"
	}
	fg := theme.Text
	if fg == "" {
		fg = "#f9fafb"
	}
	accent := theme.Accent
	if accent == "" {
		accent = fg
	}
	vhsTheme := fmt.Sprintf(`{ "background": %q, "foreground": %q, "cursor": %q, "yellow": %q, "brightYellow": %q }`,
		bg, fg, accent, accent, accent)
	return strings.Join([]string{
		fmt.Sprintf("Output %s/%s.mp4", DemosDirName, id),
		"Set FontSize 34",
		"Set Width 1440",
		"Set Height 640",
		"Set Padding 32",
		"Set TypingSpeed 80ms",
		"Set Shell bash",
		"Set Theme " + vhsTheme,
		"", "",
	}, "\n")
}

// lintTapeBody enforces the real-execution contract on an LLM-written tape
// body and returns the cleaned body.
func lintTapeBody(content string) (string, error) {
	body := stripFences(content)
	sawPython := false
	for i, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		first, _, _ := strings.Cut(trimmed, " ")
		switch first {
		case "Output", "Set", "Source", "Require":
			return "", fmt.Errorf("line %d: %s directives are engine-owned — write only the tape body", i+1, first)
		}
		if strings.HasPrefix(trimmed, "Type") {
			typed := strings.ToLower(trimmed)
			if strings.Contains(typed, "echo ") || strings.Contains(typed, "echo\"") || strings.Contains(typed, "echo'") {
				return "", fmt.Errorf("line %d: echo is forbidden — all output must come from really executing python3", i+1)
			}
			if strings.Contains(typed, "python3") {
				sawPython = true
			}
		}
	}
	if !sawPython {
		return "", fmt.Errorf("the tape never runs python3 — demos must execute real Python")
	}
	if strings.TrimSpace(body) == "" {
		return "", fmt.Errorf("tape body is empty")
	}
	return strings.TrimSpace(body) + "\n", nil
}

// demoPromptData feeds prompts/demo_tape.tmpl.
type demoPromptData struct {
	Audience    string
	LessonTitle string
	Description string
	CodeContext string
	Critique    string
}

// generateTapeBody asks the content model for a tape body and lints it.
func generateTapeBody(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec DemoSpec, critique string) (string, error) {
	data := demoPromptData{
		Audience:    cfg.Style.Audience,
		LessonTitle: l.FrontMatter.Title,
		Description: spec.Description,
		CodeContext: verifiedOutputsSummary(l),
		Critique:    critique,
	}
	system, user, err := e.renderPrompt(demoTapeTemplateName, data)
	if err != nil {
		return "", err
	}
	body, err := e.completeWithRepair(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.4, 2048, false,
		func(content string) (string, error) { return lintTapeBody(content) })
	if err != nil {
		return "", fmt.Errorf("generating tape for %s: %w", spec.ID, err)
	}
	return body, nil
}

// runDemosStage turns every [DEMO: ...] marker into a really-executed
// terminal recording: LLM-written VHS tape → vhs validate (one retry with
// the error) → vhs render in the sandbox → demos/manifest.json.
func runDemosStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	specs := extractDemoMarkers(l.Body)
	if len(specs) == 0 {
		fmt.Fprintf(e.out(), "  → demos     no [DEMO] markers — nothing to record\n")
		return nil
	}
	if e.TapeRunner == nil {
		return fmt.Errorf(
			"no way to record terminal demos — install docker and run `%s` (or install vhs on the host)",
			sandboxBuildHelp,
		)
	}
	if strings.Contains(e.TapeRunner.Name(), "UNSANDBOXED") {
		fmt.Fprintf(e.out(), "  ⚠ demos     %s\n", "recording without isolation — build the sandbox: "+sandboxBuildHelp)
	}

	demosDir := filepath.Join(l.GeneratedDir(), DemosDirName)
	if err := os.MkdirAll(demosDir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", demosDir, err)
	}

	manifest := DemoManifest{}
	for _, spec := range specs {
		if err := ctx.Err(); err != nil {
			return err
		}
		fmt.Fprintf(e.out(), "  → demos     %s: writing tape (%s)...\n", spec.ID, cfg.Pipeline.LLMContent)
		body, err := generateTapeBody(ctx, e, l, cfg, spec, "")
		if err != nil {
			return err
		}
		tapeRel := filepath.Join(DemosDirName, spec.ID+".tape")
		tapeAbs := filepath.Join(l.GeneratedDir(), tapeRel)
		if err := writeFileAtomic(tapeAbs, []byte(tapeHeader(spec.ID, videoThemeForConfig(cfg, l.FrontMatter.Title))+body)); err != nil {
			return err
		}

		// vhs validate, with one regeneration round on failure.
		if err := e.TapeRunner.Validate(ctx, l.GeneratedDir(), tapeRel); err != nil {
			fmt.Fprintf(e.out(), "    tape invalid — regenerating with the error\n")
			body, regenErr := generateTapeBody(ctx, e, l, cfg, spec, err.Error())
			if regenErr != nil {
				return regenErr
			}
			if err := writeFileAtomic(tapeAbs, []byte(tapeHeader(spec.ID, videoThemeForConfig(cfg, l.FrontMatter.Title))+body)); err != nil {
				return err
			}
			if err := e.TapeRunner.Validate(ctx, l.GeneratedDir(), tapeRel); err != nil {
				return fmt.Errorf("%s: tape still invalid after retry: %w", spec.ID, err)
			}
		}

		fmt.Fprintf(e.out(), "    recording %s (%s)...\n", spec.ID, e.TapeRunner.Name())
		renderCtx, cancel := context.WithTimeout(ctx, tapeRenderTimeout)
		err = e.TapeRunner.RenderTape(renderCtx, l.GeneratedDir(), tapeRel)
		cancel()
		if err != nil {
			return fmt.Errorf("recording %s: %w", spec.ID, err)
		}

		mp4Rel := filepath.Join(DemosDirName, spec.ID+".mp4")
		durMs, err := mediaDurationMs(filepath.Join(l.GeneratedDir(), mp4Rel))
		if err != nil {
			return fmt.Errorf("%s: %w", spec.ID, err)
		}
		manifest.Demos = append(manifest.Demos, DemoEntry{
			ID:          spec.ID,
			Description: spec.Description,
			Path:        mp4Rel,
			DurationMs:  durMs,
		})
	}

	if err := writeJSON(filepath.Join(demosDir, DemoManifestFileName), manifest); err != nil {
		return err
	}
	fmt.Fprintf(e.out(), "    %d demo(s) recorded\n", len(manifest.Demos))
	return nil
}

// loadDemoManifest reads demos/manifest.json; missing file returns an empty
// manifest (a lesson may have no demos).
func loadDemoManifest(l *project.Lesson) (*DemoManifest, error) {
	path := filepath.Join(l.GeneratedDir(), DemosDirName, DemoManifestFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &DemoManifest{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var m DemoManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing %s (delete it and re-run the demos stage): %w", path, err)
	}
	return &m, nil
}
