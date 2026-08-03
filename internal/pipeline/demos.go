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
	// CaptureManifestFileName is what the capture stage wrote, kept apart from
	// manifest.json so the two stages have one writer each.
	CaptureManifestFileName = "captures.json" // inside demos/
)

// tapeRenderTimeout bounds one VHS recording.
const tapeRenderTimeout = 5 * time.Minute

// captureRenderTimeout bounds a tool capture, which waits on a real agent or a
// real deploy rather than on a scripted Sleep. It must exceed
// captureWaitTimeout, or the engine kills the take before VHS gives up.
const captureRenderTimeout = 15 * time.Minute

// DemoSpec is one [DEMO: description] or [CAPTURE: tool=…; description]
// marker from lesson.md.
type DemoSpec struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	// Kind decides where this recording may run: the sandbox with the network
	// off, or the host with a real credential. See capture.go.
	Kind CaptureKind `json:"kind"`
	// Tool is the allowlist key for a tool capture, empty for python.
	Tool string `json:"tool,omitempty"`
	// Fixture names a directory under the course's fixtures/ to seed the
	// scratch working directory with. Terminal captures only.
	Fixture string `json:"fixture,omitempty"`
	// Take names a file under the course's takes/ that drives a web capture.
	// Required for the web kind and meaningless for the others.
	Take string `json:"take,omitempty"`
}

// DemoEntry is one rendered demo in the manifest consumed by scenegraph.
type DemoEntry struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	// Path is relative to the lesson's generated dir, e.g. "demos/demo-1.mp4".
	Path       string      `json:"path"`
	DurationMs int         `json:"durationMs"`
	Kind       CaptureKind `json:"kind,omitempty"`
	Tool       string      `json:"tool,omitempty"`
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
			Kind:        CaptureKindPython,
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
// tapeTypingSpeedMs is the per-keystroke speed the header sets. footage.go's
// mark model reads the same constant, so a change here cannot silently
// mistime every mark in the library.
const tapeTypingSpeedMs = 80

// captureWaitTimeout is how long a tool tape's `Wait` may block. VHS defaults
// to a handful of seconds, which is right for a scripted demo and far too
// short for the commands worth recording — an agent editing files or a deploy
// building takes minutes, and the default turns exactly those into a failed
// take. Only the tool header sets it, so python tapes stay byte-identical.
const captureWaitTimeout = "10m"

func tapeHeader(id string, theme SceneTheme) string {
	return tapeHeaderTo(fmt.Sprintf("%s/%s.mp4", DemosDirName, id), theme)
}

// tapeHeaderTo is tapeHeader with an explicit output path, so a tool capture
// can run in a scratch directory while still writing its clip into the
// lesson's generated dir. extraSettings are appended to the engine-owned Set
// block.
func tapeHeaderTo(outputPath string, theme SceneTheme, extraSettings ...string) string {
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
		// Quoted, because VHS's parser splits an unquoted Output argument on
		// its path separators and reports the tail as an invalid command —
		// which is what a tool capture writing to an absolute path back in the
		// lesson dir hits on the very first line, at validate time.
		fmt.Sprintf("Output %q", outputPath),
		"Set FontSize 34",
		"Set Width 1440",
		"Set Height 640",
		"Set Padding 32",
		fmt.Sprintf("Set TypingSpeed %dms", tapeTypingSpeedMs),
		"Set Shell bash",
		"Set Theme " + vhsTheme,
	}, "\n") + "\n" + strings.Join(append(extraSettings, "", ""), "\n")
}

// lintTapeBody enforces the real-execution contract on an LLM-written tape
// body for the **python** kind, and returns the cleaned body. The tool kind's
// equivalent is lintToolTapeBody in capture.go: same iron rule, different
// engine, and a denylist this one does not need because the python path runs
// with the network off.
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

// demoPromptData feeds prompts/demo_tape.tmpl and prompts/capture_tape.tmpl.
type demoPromptData struct {
	Audience    string
	LessonTitle string
	Description string
	CodeContext string
	Critique    string
	// Tool and Fixture are the tool kind's own fields; the python prompt
	// ignores them.
	Tool       string
	Binary     string
	Fixture    string
	Invocation string
}

// generateTapeBody asks the content model for a tape body and lints it against
// the rules of the spec's kind.
func generateTapeBody(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec DemoSpec, critique string) (string, error) {
	data := demoPromptData{
		Audience:    cfg.Style.Audience,
		LessonTitle: l.FrontMatter.Title,
		Description: spec.Description,
		Critique:    critique,
		Fixture:     spec.Fixture,
	}
	tmpl := demoTapeTemplateName
	lint := func(content string) (string, error) { return lintTapeBody(content) }
	if spec.Kind == CaptureKindTool {
		tool, ok := captureTools[spec.Tool]
		if !ok {
			return "", fmt.Errorf("%s: %q is not a recordable tool", spec.ID, spec.Tool)
		}
		tmpl = captureTapeTemplateName
		data.Tool = tool.Display
		data.Binary = tool.Binary
		data.Invocation = tool.Invocation
		lint = func(content string) (string, error) { return lintToolTapeBody(content, tool) }
	} else {
		data.CodeContext = verifiedOutputsSummary(l)
	}
	system, user, err := e.renderPrompt(tmpl, data)
	if err != nil {
		return "", err
	}
	body, err := e.completeWithRepair(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.4, 2048, false, lint)
	if err != nil {
		return "", fmt.Errorf("generating tape for %s: %w", spec.ID, err)
	}
	return body, nil
}

// runDemosStage turns every [DEMO: ...] marker into a really-executed
// terminal recording: LLM-written VHS tape → vhs validate (one retry with
// the error) → vhs render in the sandbox → demos/manifest.json.
func runDemosStage(ctx context.Context, e *Env, c *project.Course, l *project.Lesson, cfg config.Config) error {
	specs := extractDemoMarkers(l.Body)
	if len(specs) == 0 {
		fmt.Fprintf(e.out(), "  → demos     no [DEMO] markers — nothing to record\n")
	}
	entries, err := recordAll(ctx, e, c, l, cfg, specs)
	if err != nil {
		return err
	}
	// The manifest the scene graph reads is both kinds together, in outline
	// order: captures first because they were recorded first. Writing it here
	// rather than in the capture stage keeps one file with one writer — two
	// stages appending to the same manifest is how it ends up half-overwritten.
	captures, err := loadCaptureManifest(l)
	if err != nil {
		return err
	}
	manifest := DemoManifest{Demos: append(captures.Demos, entries...)}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), DemosDirName, DemoManifestFileName), manifest); err != nil {
		return err
	}
	if len(specs) > 0 {
		fmt.Fprintf(e.out(), "    %d demo(s) recorded\n", len(entries))
	}
	return nil
}

// runCaptureStage records the real-tool markers, ahead of the script.
//
// It is a separate stage from demos for one reason, and it is the reason the
// whole no-code track exists: a script written before anything was recorded
// cannot say what the recording showed. Lesson five's narration could say
// "watch how much of the work is finding the problem" and could not say "it
// took under a minute", because at writing time nobody knew.
//
// Splitting by kind is what makes that possible without a cycle. A python demo
// needs the verified code, so it must run after verify, which runs after
// script. A capture needs none of that — its input is the marker and, for the
// web and desktop kinds, a checked-in take. So captures can run first, and the
// script stage takes them as input.
func runCaptureStage(ctx context.Context, e *Env, c *project.Course, l *project.Lesson, cfg config.Config) error {
	// A no-code piece has no lesson.md when this runs — the plan stage writes it
	// afterwards — so its recordings are declared by its segments instead. That
	// is also where somebody editing the piece expects to find them.
	var specs []DemoSpec
	if IsNoCode(l) {
		spec, err := LoadNoCodeSpec(l.Dir)
		if err != nil {
			return err
		}
		if err := spec.Validate(); err != nil {
			return err
		}
		specs = spec.CaptureSpecs()
	} else {
		var err error
		specs, err = extractCaptureMarkers(l.Body)
		if err != nil {
			return err
		}
	}
	if len(specs) == 0 {
		fmt.Fprintf(e.out(), "  → capture   no [CAPTURE] markers — nothing to record\n")
		// Written even when empty: its absence and its emptiness mean different
		// things to the script stage's staleness, and a missing file would make
		// every run look like the captures had just changed.
		return writeJSON(filepath.Join(l.GeneratedDir(), DemosDirName, CaptureManifestFileName), DemoManifest{})
	}
	entries, err := recordAll(ctx, e, c, l, cfg, specs)
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), DemosDirName, CaptureManifestFileName), DemoManifest{Demos: entries}); err != nil {
		return err
	}
	fmt.Fprintf(e.out(), "    %d capture(s)\n", len(entries))
	return nil
}

// stageLabel is which stage a spec belongs to, for progress lines. The two
// kinds record from different stages now, and a capture logging itself as
// "demos" sends anybody reading the output to the wrong stage name.
func stageLabel(k CaptureKind) string {
	if k == CaptureKindPython {
		return "demos  "
	}
	return "capture"
}

// recordAll records a list of specs into the lesson's demos dir.
func recordAll(ctx context.Context, e *Env, c *project.Course, l *project.Lesson, cfg config.Config, specs []DemoSpec) ([]DemoEntry, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	dir := filepath.Join(l.GeneratedDir(), DemosDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating %s: %w", dir, err)
	}
	entries := make([]DemoEntry, 0, len(specs))
	for _, spec := range specs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		entry, err := recordOneDemo(ctx, e, c, l, cfg, spec)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

// loadCaptureManifest reads what the capture stage recorded. A missing file is
// an empty manifest: a course with no captures never writes one.
func loadCaptureManifest(l *project.Lesson) (*DemoManifest, error) {
	path := filepath.Join(l.GeneratedDir(), DemosDirName, CaptureManifestFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &DemoManifest{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var m DemoManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing %s (delete it and re-run the capture stage): %w", path, err)
	}
	return &m, nil
}

// recordOneDemo writes a tape, validates it (one regeneration round on
// failure), renders it, and writes the clip's footage sidecar.
//
// The two kinds differ only in where the tape runs: the python kind records in
// the lesson's generated dir under whatever isolation ResolveTapeRunner found,
// and the tool kind records in a throwaway scratch dir on the host. Everything
// either side of that — generation, the validate-retry, measuring the clip,
// building footage.json — is shared, which is what stops the two paths drifting
// into two recorders.
func recordOneDemo(ctx context.Context, e *Env, c *project.Course, l *project.Lesson, cfg config.Config, spec DemoSpec) (DemoEntry, error) {
	// A web capture shares none of the tape machinery — no VHS, no shell, no
	// generated script. It is its own path from here.
	if spec.Kind == CaptureKindWeb {
		return recordWebCapture(ctx, e, c, l, spec)
	}
	if spec.Kind == CaptureKindDesktop {
		return recordDesktopCapture(ctx, e, c, l, spec)
	}
	runner, workDir, cleanup, err := resolveRecordingContext(ctx, e, c, l, spec)
	if err != nil {
		return DemoEntry{}, err
	}
	defer cleanup()

	theme := videoThemeForConfig(cfg, l.FrontMatter.Title)
	mp4Rel := filepath.Join(DemosDirName, spec.ID+".mp4")
	mp4Abs := filepath.Join(l.GeneratedDir(), mp4Rel)

	// A tool tape runs in a scratch dir, so its Output has to be the absolute
	// path back into the lesson; a python tape keeps the relative form it has
	// always had, so its tapes stay byte-identical to before.
	header := tapeHeader(spec.ID, theme)
	tapeRel := filepath.Join(DemosDirName, spec.ID+".tape")
	if spec.Kind == CaptureKindTool {
		header = tapeHeaderTo(mp4Abs, theme, "Set WaitTimeout "+captureWaitTimeout)
		// One level above the recorded directory, so the tape itself never
		// shows up in its own recording. See captureTapeRelPath.
		tapeRel = filepath.Join(captureTapeRelPath, spec.ID+".tape")
	}
	tapeAbs := filepath.Join(workDir, tapeRel)

	write := func(body string) error { return writeFileAtomic(tapeAbs, []byte(header+body)) }

	fmt.Fprintf(e.out(), "  → %s   %s: writing tape (%s)...\n", stageLabel(spec.Kind), spec.ID, cfg.Pipeline.LLMContent)
	body, err := generateTapeBody(ctx, e, l, cfg, spec, "")
	if err != nil {
		return DemoEntry{}, err
	}
	if err := write(body); err != nil {
		return DemoEntry{}, err
	}
	if err := runner.Validate(ctx, workDir, tapeRel); err != nil {
		fmt.Fprintf(e.out(), "    tape invalid — regenerating with the error\n")
		body, err = generateTapeBody(ctx, e, l, cfg, spec, err.Error())
		if err != nil {
			return DemoEntry{}, err
		}
		if err := write(body); err != nil {
			return DemoEntry{}, err
		}
		if err := runner.Validate(ctx, workDir, tapeRel); err != nil {
			return DemoEntry{}, fmt.Errorf("%s: tape still invalid after retry: %w", spec.ID, err)
		}
	}

	fmt.Fprintf(e.out(), "    recording %s (%s)...\n", spec.ID, runner.Name())
	budget := tapeRenderTimeout
	if spec.Kind == CaptureKindTool {
		budget = captureRenderTimeout
	}
	renderCtx, cancel := context.WithTimeout(ctx, budget)
	err = runner.RenderTape(renderCtx, workDir, tapeRel)
	cancel()
	if err != nil {
		return DemoEntry{}, fmt.Errorf("recording %s: %w", spec.ID, err)
	}

	durMs, err := mediaDurationMs(mp4Abs)
	if err != nil {
		return DemoEntry{}, fmt.Errorf("%s: %w", spec.ID, err)
	}

	// The scratch tape is the only record of what was run, so it is copied
	// back beside the clip before the scratch dir goes.
	if spec.Kind == CaptureKindTool {
		if err := writeFileAtomic(filepath.Join(l.GeneratedDir(), DemosDirName, spec.ID+".tape"), []byte(header+body)); err != nil {
			return DemoEntry{}, err
		}
	}

	var toolVersion string
	if spec.Kind == CaptureKindTool {
		toolVersion = observeToolVersion(ctx, captureTools[spec.Tool])
	}
	f := buildFootage(spec.ID, spec.Kind, spec.Tool, toolVersion, body, durMs, time.Now())
	if err := writeJSON(filepath.Join(l.GeneratedDir(), DemosDirName, spec.ID+FootageFileSuffix), f); err != nil {
		return DemoEntry{}, err
	}
	if len(f.Marks) > 0 && !f.Exact() {
		fmt.Fprintf(e.out(), "    ⚠ %s: %d Wait directives — marks after the first are approximate and will not be cut on\n", spec.ID, f.Waits)
	}

	return DemoEntry{
		ID:          spec.ID,
		Description: spec.Description,
		Path:        mp4Rel,
		DurationMs:  durMs,
		Kind:        spec.Kind,
		Tool:        spec.Tool,
	}, nil
}

// recordWebCapture drives a checked-in take against a real site and writes one
// still per shot, plus the footage sidecar carrying the origin it was really
// captured at.
//
// There is no clip here and no duration: the scene divides its own screen time
// across the frames. That is why the manifest entry's DurationMs is zero, and
// why nothing downstream should treat zero as an error for this kind.
func recordWebCapture(ctx context.Context, e *Env, c *project.Course, l *project.Lesson, spec DemoSpec) (DemoEntry, error) {
	site := captureSites[spec.Tool]
	courseDir := ""
	if c != nil {
		courseDir = c.Dir
	}
	takePath := filepath.Join(courseDir, "takes", spec.Take+".yaml")
	take, err := LoadWebTake(takePath)
	if err != nil {
		return DemoEntry{}, fmt.Errorf("%s: %w", spec.ID, err)
	}
	if take.Site != spec.Tool {
		return DemoEntry{}, fmt.Errorf("%s: the lesson asks for a %s capture but %s drives %s",
			spec.ID, spec.Tool, filepath.Base(takePath), take.Site)
	}

	demosDir := filepath.Join(l.GeneratedDir(), DemosDirName)
	fmt.Fprintf(e.out(), "  ⚠ capture   %s drives %s in a real browser, signed in as you — run `coursesmith footage login %s` first if it is not\n",
		spec.ID, site.Display, site.Key)
	fmt.Fprintf(e.out(), "  → capture   %s: capturing %d frame(s) from %s...\n", spec.ID, countShots(take), site.Origin)

	runCtx, cancel := context.WithTimeout(ctx, captureRenderTimeout)
	defer cancel()
	res, err := RunWebTake(runCtx, e, take, demosDir, spec.ID, true)
	if err != nil {
		return DemoEntry{}, fmt.Errorf("%s: %w", spec.ID, err)
	}

	f := Footage{
		ID:         spec.ID,
		Kind:       CaptureKindWeb,
		Tool:       spec.Tool,
		Origin:     res.Origin,
		Take:       spec.Take,
		CapturedAt: time.Now().UTC().Format(time.RFC3339),
		Frames:     res.Frames,
		DurationMs: res.ClipMs,
		Marks:      res.Marks,
	}
	if err := writeJSON(filepath.Join(demosDir, spec.ID+FootageFileSuffix), f); err != nil {
		return DemoEntry{}, err
	}
	if res.ClipPath != "" {
		fmt.Fprintf(e.out(), "    %s recorded (%.1fs, %d mark(s))\n", res.ClipPath, float64(res.ClipMs)/1000, len(res.Marks))
	}
	if len(res.Frames) > 0 {
		fmt.Fprintf(e.out(), "    %d still(s) captured\n", len(res.Frames))
	}

	path := filepath.Join(DemosDirName, spec.ID+FootageFileSuffix)
	if res.ClipPath != "" {
		path = filepath.Join(DemosDirName, res.ClipPath)
	}
	return DemoEntry{
		ID:          spec.ID,
		Description: spec.Description,
		Path:        path,
		DurationMs:  res.ClipMs,
		Kind:        CaptureKindWeb,
		Tool:        spec.Tool,
	}, nil
}

// recordDesktopCapture records an operator working through a take's beats in a
// native application.
//
// This is the only stage in the engine that stops and waits for a person, and
// it says so loudly before it starts: a run that silently blocks on stdin
// halfway through a course build is a hang as far as anybody watching is
// concerned.
func recordDesktopCapture(ctx context.Context, e *Env, c *project.Course, l *project.Lesson, spec DemoSpec) (DemoEntry, error) {
	app := captureApps[spec.Tool]
	courseDir := ""
	if c != nil {
		courseDir = c.Dir
	}
	// The console check comes first, ahead of even reading the take. Whether
	// anybody is at the keyboard is a property of the run rather than of the
	// take, and if nobody is then the take's validity is beside the point — a
	// missing-file error here would send somebody looking in the wrong place.
	if e.DesktopInput == nil {
		return DemoEntry{}, fmt.Errorf("%s: a %s capture needs somebody at the keyboard, and this run has no console attached. Record it with `coursesmith footage shoot <course>/%s`",
			spec.ID, app.Display, l.ID)
	}
	takePath := filepath.Join(courseDir, "takes", spec.Take+".yaml")
	take, err := LoadDesktopTake(takePath)
	if err != nil {
		return DemoEntry{}, fmt.Errorf("%s: %w", spec.ID, err)
	}
	if take.App != spec.Tool {
		return DemoEntry{}, fmt.Errorf("%s: the lesson asks for a %s capture but %s drives %s",
			spec.ID, spec.Tool, filepath.Base(takePath), take.App)
	}

	rec, err := NewDesktopRecorder(ctx, e)
	if err != nil {
		return DemoEntry{}, fmt.Errorf("%s: %w", spec.ID, err)
	}
	demosDir := filepath.Join(l.GeneratedDir(), DemosDirName)
	fmt.Fprintf(e.out(), "  ⚠ capture   %s records YOUR SCREEN while you work in %s — %d beat(s) to perform\n",
		spec.ID, app.Display, len(take.Beats))

	res, err := RunDesktopTake(ctx, e, take, rec, demosDir, spec.ID, e.DesktopInput, e.out())
	if err != nil {
		return DemoEntry{}, fmt.Errorf("%s: %w", spec.ID, err)
	}

	f := Footage{
		ID:         spec.ID,
		Kind:       CaptureKindDesktop,
		Tool:       spec.Tool,
		Take:       spec.Take,
		CapturedAt: time.Now().UTC().Format(time.RFC3339),
		DurationMs: res.ClipMs,
		Marks:      res.Marks,
	}
	if err := writeJSON(filepath.Join(demosDir, spec.ID+FootageFileSuffix), f); err != nil {
		return DemoEntry{}, err
	}
	fmt.Fprintf(e.out(), "    %s recorded (%.1fs, %d mark(s))\n", res.ClipPath, float64(res.ClipMs)/1000, len(res.Marks))

	return DemoEntry{
		ID:          spec.ID,
		Description: spec.Description,
		Path:        filepath.Join(DemosDirName, res.ClipPath),
		DurationMs:  res.ClipMs,
		Kind:        CaptureKindDesktop,
		Tool:        spec.Tool,
	}, nil
}

// countShots is how many frames a take will produce.
func countShots(t *WebTake) int {
	n := 0
	for _, s := range t.Steps {
		if s.Do == "shot" {
			n++
		}
	}
	return n
}

// resolveRecordingContext picks the runner and working directory for one spec,
// and returns the cleanup for whatever it allocated.
func resolveRecordingContext(ctx context.Context, e *Env, c *project.Course, l *project.Lesson, spec DemoSpec) (TapeRunner, string, func(), error) {
	noop := func() {}
	if spec.Kind != CaptureKindTool {
		if e.TapeRunner == nil {
			return nil, "", noop, fmt.Errorf(
				"no way to record terminal demos — install docker and run `%s` (or install vhs on the host)",
				sandboxBuildHelp,
			)
		}
		if strings.Contains(e.TapeRunner.Name(), "UNSANDBOXED") {
			fmt.Fprintf(e.out(), "  ⚠ demos     %s\n", "recording without isolation — build the sandbox: "+sandboxBuildHelp)
		}
		return e.TapeRunner, l.GeneratedDir(), noop, nil
	}

	tool, ok := captureTools[spec.Tool]
	if !ok {
		return nil, "", noop, fmt.Errorf("%s: %q is not a recordable tool", spec.ID, spec.Tool)
	}
	resolve := e.ToolTapeRunner
	if resolve == nil {
		resolve = resolveToolTapeRunner
	}
	runner, err := resolve(ctx, tool)
	if err != nil {
		return nil, "", noop, fmt.Errorf("%s: %w", spec.ID, err)
	}
	courseDir := ""
	if c != nil {
		courseDir = c.Dir
	}
	workDir, cleanup, err := prepareCaptureWorkdir(courseDir, spec.Fixture)
	if err != nil {
		return nil, "", noop, fmt.Errorf("%s: %w", spec.ID, err)
	}
	fmt.Fprintf(e.out(), "  ⚠ capture   %s records %s on the host, authenticated and online — the sandbox cannot run a network client\n",
		spec.ID, tool.Display)
	return runner, workDir, cleanup, nil
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
