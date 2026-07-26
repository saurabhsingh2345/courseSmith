package pipeline

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// DiagramsDirName is the Stage 3 output directory under generated/.
const DiagramsDirName = "diagrams"

// AttemptsDirName keeps every generation attempt (SVG + screenshot) under
// generated/diagrams/ for audit.
const AttemptsDirName = "attempts"

// diagramStyleDirName holds the few-shot exemplar SVGs, under prompts/.
const diagramStyleDirName = "diagram_style"

// maxVisualRounds is the initial render plus up to two regenerations driven
// by the vision QA critique.
const maxVisualRounds = 3

// diagramPromptData feeds prompts/diagram_svg.tmpl.
type diagramPromptData struct {
	DiagramStyle string
	Colors       config.Colors
	// Theme carries the derived dark video tokens the diagram must sit on.
	Theme SceneTheme
	Title        string
	ID           string
	Prompt       string
	Critique     string
	// Narration is the spoken text of the section that cues this diagram —
	// the context that makes the picture match the lesson.
	Narration string
	Audience  string
	// Exemplars are shared style examples injected as few-shot references.
	Exemplars []string
}

// loadExemplars reads prompts/diagram_style/*.svg in name order. Missing dir
// is fine — exemplars are an enhancement.
func loadExemplars(promptsDir string) []string {
	dir := filepath.Join(promptsDir, diagramStyleDirName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".svg") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	var out []string
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err == nil {
			out = append(out, strings.TrimSpace(string(data)))
		}
	}
	return out
}

// generateDiagram asks the content model for one standalone SVG and
// validates it (well-formed XML, <svg> root, viewBox, self-contained).
// Invalid SVG triggers one correction round carrying the exact XML error.
func generateDiagram(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec project.DiagramSpec, critique string) ([]byte, error) {
	data := diagramPromptData{
		DiagramStyle: cfg.Branding.DiagramStyle,
		Colors:       cfg.Branding.Colors,
		Theme:        videoThemeForConfig(cfg, l.FrontMatter.Title),
		Title:        l.FrontMatter.Title,
		ID:           spec.ID,
		Prompt:       spec.Prompt,
		Critique:     critique,
		Narration:    spec.Narration,
		Audience:     cfg.Style.Audience,
		Exemplars:    loadExemplars(e.PromptsDir),
	}
	system, user, err := e.renderPrompt(diagramTemplateName, data)
	if err != nil {
		return nil, err
	}
	svg, err := e.completeWithRepair(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.4, 8192, false, extractSVG)
	if err != nil {
		return nil, fmt.Errorf("generating diagram %q: %w", spec.ID, err)
	}
	return []byte(svg), nil
}

// extractSVG pulls the <svg>...</svg> element out of a model reply and
// validates it. It returns the cleaned SVG ready to write to disk.
func extractSVG(content string) (string, error) {
	s := stripFences(content)
	start := strings.Index(s, "<svg")
	if start < 0 {
		return "", fmt.Errorf("no <svg> element in the reply")
	}
	end := strings.LastIndex(s, "</svg>")
	if end < start {
		return "", fmt.Errorf("<svg> element is never closed")
	}
	s = s[start : end+len("</svg>")]
	if err := validateSVG(s); err != nil {
		return "", err
	}
	return s + "\n", nil
}

// validateSVG checks well-formedness with encoding/xml and enforces the
// standalone-SVG contract: svg root with a viewBox, no scripts, no external
// references.
func validateSVG(s string) error {
	dec := xml.NewDecoder(strings.NewReader(s))
	sawRoot := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("invalid XML: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if !sawRoot {
			sawRoot = true
			if start.Name.Local != "svg" {
				return fmt.Errorf("root element is <%s>, want <svg>", start.Name.Local)
			}
			if attrValue(start, "viewBox") == "" {
				return fmt.Errorf("<svg> has no viewBox attribute (want viewBox=\"0 0 800 H\")")
			}
		}
		switch start.Name.Local {
		case "script":
			return fmt.Errorf("<script> elements are not allowed in diagrams")
		case "image":
			return fmt.Errorf("<image> elements are not allowed in diagrams (SVG must be self-contained)")
		}
		for _, attr := range start.Attr {
			v := strings.ToLower(attr.Value)
			if strings.Contains(v, "http://") || strings.Contains(v, "https://") {
				// The xmlns namespace declarations are the only legitimate URLs.
				if attr.Name.Local == "xmlns" || attr.Name.Space == "xmlns" {
					continue
				}
				return fmt.Errorf("attribute %s references an external URL (%s) — the SVG must be self-contained", attr.Name.Local, attr.Value)
			}
		}
	}
	if !sawRoot {
		return fmt.Errorf("no elements found")
	}
	return nil
}

func attrValue(el xml.StartElement, name string) string {
	for _, attr := range el.Attr {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

// visualVerdict is the vision model's judgment of a rendered diagram.
type visualVerdict struct {
	Passed bool     `json:"passed"`
	Issues []string `json:"issues"`
}

func (v *visualVerdict) Validate() error {
	if !v.Passed && len(v.Issues) == 0 {
		return fmt.Errorf("a failing verdict must list the issues found")
	}
	return nil
}

// visualQAPromptData feeds prompts/diagram_visual_qa.tmpl.
type visualQAPromptData struct {
	ID     string
	Prompt string
}

// reviewDiagramVisually sends the rendered PNG to the review (vision) model
// with the visual checklist.
func reviewDiagramVisually(ctx context.Context, e *Env, cfg config.Config, spec project.DiagramSpec, png []byte) (*visualVerdict, error) {
	system, user, err := e.renderPrompt(diagramVisualQATemplateName, visualQAPromptData{ID: spec.ID, Prompt: spec.Prompt})
	if err != nil {
		return nil, err
	}
	var verdict visualVerdict
	err = e.completeJSONWithImages(ctx, cfg.Pipeline, llm.TaskVision, system, user,
		[]string{base64.StdEncoding.EncodeToString(png)}, 0, 1024, &verdict, verdict.Validate)
	if err != nil {
		return nil, fmt.Errorf("visual QA for %q: %w", spec.ID, err)
	}
	return &verdict, nil
}

// visualRecord is the persisted audit trail of one visual QA round.
type visualRecord struct {
	Kind      string   `json:"kind"`
	Round     int      `json:"round"`
	Model     string   `json:"model"`
	Passed    bool     `json:"passed"`
	Issues    []string `json:"issues,omitempty"`
	SVGPath   string   `json:"svg_path"`
	PNGPath   string   `json:"png_path"`
	CheckedAt string   `json:"checked_at"`
}

// runVisualsStage is Stage 3: generate each declared diagram, screenshot it
// headless, and gate it through the vision reviewer's checklist (max 2
// regenerations, every attempt kept under diagrams/attempts/). Without a
// screenshotter (or when screenshotting fails) it falls back to the
// text-based rubric gate.
func runVisualsStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	diagrams := l.FrontMatter.Diagrams
	if len(diagrams) == 0 {
		fmt.Fprintf(e.out(), "  → visuals   no diagrams declared — nothing to generate\n")
		return nil
	}
	if e.Screenshotter == nil {
		fmt.Fprintf(e.out(), "  ⚠ visuals   no headless browser — falling back to source-text review (visual QA disabled)\n")
	}

	// Context-aware generation: give each diagram the narration of the script
	// section that cues it, so the picture matches what is being said. The
	// script is an enhancement here — its absence just means no narration.
	narration := map[string]string{}
	if script, err := loadScript(l); err == nil {
		for _, sec := range script.Sections {
			for _, cue := range sec.Cues {
				if cue.Type == CueDiagram {
					narration[cue.Ref] = sec.Narration
				}
			}
		}
	}

	for _, spec := range diagrams {
		if err := ctx.Err(); err != nil {
			return err
		}
		spec.Narration = narration[spec.ID]
		fmt.Fprintf(e.out(), "  → visuals   drawing %q (%s, %s)...\n", spec.ID, spec.ResolvedKind(), cfg.Pipeline.LLMContent)

		var final []byte
		var err error
		switch spec.ResolvedKind() {
		case project.DiagramKindD3:
			// Structured node-link graph: validated on generation, laid out and
			// animated by the renderer (no PNG to vision-review here).
			final, err = produceD3Diagram(ctx, e, l, cfg, spec)
		case project.DiagramKindD2:
			// D2 source compiled in-process (pure Go), then vision-QA'd.
			final, err = produceD2Diagram(ctx, e, l, cfg, spec)
		case project.DiagramKindMermaid:
			// Mermaid syntax compiled to SVG, then vision-QA'd like any SVG.
			final, err = produceMermaidDiagram(ctx, e, l, cfg, spec)
		case project.DiagramKindExcalidraw:
			// Excalidraw elements compiled to a hand-drawn SVG, then vision-QA'd.
			final, err = produceExcalidrawDiagram(ctx, e, l, cfg, spec)
		default:
			final, err = produceDiagram(ctx, e, l, cfg, spec)
		}
		if err != nil {
			return err
		}
		path := filepath.Join(l.GeneratedDir(), DiagramsDirName, diagramArtifactName(spec))
		if err := writeFileAtomic(path, final); err != nil {
			return err
		}
	}
	fmt.Fprintf(e.out(), "    %d diagram(s) written to %s\n", len(diagrams), filepath.Join(project.GeneratedDirName, DiagramsDirName))
	return nil
}

// produceDiagram generates a freehand SVG and runs it through the vision-QA
// loop.
func produceDiagram(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec project.DiagramSpec) ([]byte, error) {
	return runDiagramVisionQA(ctx, e, l, cfg, spec, func(ctx context.Context, critique string) ([]byte, error) {
		return generateDiagram(ctx, e, l, cfg, spec, critique)
	})
}

// runDiagramVisionQA runs the generate → screenshot → vision-review loop for
// one diagram and returns the SVG to publish. gen produces (and regenerates,
// given a critique) the SVG to review — freehand for the svg kind, or a
// compiled source SVG for the mermaid/excalidraw kinds. Without a
// screenshotter (or when screenshotting fails) it falls back to the text-based
// rubric gate over the same gen.
func runDiagramVisionQA(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec project.DiagramSpec, gen func(ctx context.Context, critique string) ([]byte, error)) ([]byte, error) {
	textGate := func(svg []byte) ([]byte, error) {
		best, _, err := e.reviewGate(ctx, l, cfg, "diagram:"+spec.ID, svg, gen)
		return best, err
	}

	svg, err := gen(ctx, "")
	if err != nil {
		return nil, err
	}
	if e.Screenshotter == nil {
		return textGate(svg)
	}

	attemptsDir := filepath.Join(l.GeneratedDir(), DiagramsDirName, AttemptsDirName)
	if err := os.MkdirAll(attemptsDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating %s: %w", attemptsDir, err)
	}

	visionModel := cfg.Pipeline.LLMVision
	if visionModel == "" {
		visionModel = cfg.Pipeline.LLMReview
	}

	for round := 1; ; round++ {
		svgPath := filepath.Join(attemptsDir, fmt.Sprintf("%s-round-%d.svg", spec.ID, round))
		if err := writeFileAtomic(svgPath, svg); err != nil {
			return nil, err
		}

		png, err := e.Screenshotter.ScreenshotSVG(ctx, svg)
		if err != nil {
			fmt.Fprintf(e.out(), "  ⚠ visuals   screenshot failed (%v) — falling back to source-text review for %q\n", err, spec.ID)
			return textGate(svg)
		}
		pngPath := strings.TrimSuffix(svgPath, ".svg") + ".png"
		if err := writeFileAtomic(pngPath, png); err != nil {
			return nil, err
		}

		fmt.Fprintf(e.out(), "    visual QA round %d (%s)...\n", round, visionModel)
		verdict, err := reviewDiagramVisually(ctx, e, cfg, spec, png)
		if err != nil {
			return nil, err
		}
		record := visualRecord{
			Kind: "diagram:" + spec.ID, Round: round, Model: visionModel,
			Passed: verdict.Passed, Issues: verdict.Issues,
			SVGPath: svgPath, PNGPath: pngPath,
			CheckedAt: time.Now().UTC().Format(time.RFC3339),
		}
		recordPath := filepath.Join(l.GeneratedDir(), ReviewsDirName,
			fmt.Sprintf("%s-visual-round-%d.json", kindSlug("diagram:"+spec.ID), round))
		if err := writeJSON(recordPath, record); err != nil {
			return nil, err
		}

		if verdict.Passed {
			fmt.Fprintf(e.out(), "    visual QA passed\n")
			return svg, nil
		}
		fmt.Fprintf(e.out(), "    issues: %s\n", strings.Join(verdict.Issues, "; "))
		if round >= maxVisualRounds {
			fmt.Fprintf(e.out(), "  ⚠ visuals   %q still has visual issues after %d rounds — keeping the last attempt; see %s\n",
				spec.ID, round, filepath.Join(project.GeneratedDirName, DiagramsDirName, AttemptsDirName))
			return svg, nil
		}
		fmt.Fprintf(e.out(), "    regenerating with the critique...\n")
		prev := svg
		svg, err = gen(ctx, strings.Join(verdict.Issues, "; "))
		if err != nil {
			return nil, err
		}
		// A layout-engine kind (mermaid/excalidraw) that regenerates to the
		// exact same SVG cannot be improved by re-reviewing it — the critique
		// had no effect on the deterministic output. Keep it rather than
		// burning identical rounds (and repeating the same verdict).
		if bytes.Equal(svg, prev) {
			fmt.Fprintf(e.out(), "  ⚠ visuals   %q regenerated identically — layout is deterministic, so the critique cannot change it; keeping it\n", spec.ID)
			return svg, nil
		}
	}
}
