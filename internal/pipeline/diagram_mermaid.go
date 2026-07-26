package pipeline

// Mermaid diagram kind. The content model authors Mermaid syntax (flowchart,
// sequence, state, class, er, …); the visuals stage compiles it to SVG with
// the embedded Mermaid.js in the headless browser and gates the result through
// the same vision-QA loop as freehand SVG. Compiling inside the generation
// loop's accept() means invalid syntax is caught and fed back to the model as
// a correction, so what reaches disk always renders.

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// mermaidPromptData feeds prompts/diagram_mermaid.tmpl.
type mermaidPromptData struct {
	Title    string
	ID       string
	Prompt   string
	Critique string
	// Narration is the spoken text of the cueing section (context).
	Narration string
	Audience  string
}

// mermaidHeaderRe matches the diagram-type keyword that must open valid
// Mermaid source (after any %%{init}%% directive and comments).
var mermaidHeaderRe = regexp.MustCompile(`^(?:flowchart|graph|sequenceDiagram|stateDiagram(?:-v2)?|classDiagram|erDiagram|mindmap|journey|gantt|pie|gitGraph|timeline|quadrantChart)\b`)

// validateMermaidSyntax rejects obvious non-diagrams (prose, empty replies)
// before the browser is asked to render. It is a cheap pre-filter; the real
// gate is that the syntax actually renders.
func validateMermaidSyntax(syntax string) error {
	if strings.TrimSpace(syntax) == "" {
		return fmt.Errorf("mermaid source is empty")
	}
	for _, line := range strings.Split(syntax, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "%%") { // skip blanks and %% directives/comments
			continue
		}
		if !mermaidHeaderRe.MatchString(t) {
			return fmt.Errorf("mermaid source must open with a diagram type (flowchart, sequenceDiagram, stateDiagram, classDiagram, erDiagram, …); got %q", firstWords(t, 6))
		}
		return nil
	}
	return fmt.Errorf("mermaid source has no diagram declaration")
}

// firstWords returns up to n space-separated words of s, for error messages.
func firstWords(s string, n int) string {
	f := strings.Fields(s)
	if len(f) > n {
		f = f[:n]
	}
	return strings.Join(f, " ")
}

// validateCompiledSVG is the lightweight contract for browser-compiled SVG
// (mermaid/excalidraw): well-formed XML, an <svg> root, and no <script>. It is
// deliberately looser than validateSVG (no viewBox requirement, external-URL
// check skipped) because the source is trusted and compiled by our own
// pinned libraries — Mermaid, for instance, emits a self-contained <style>
// block and omits viewBox.
func validateCompiledSVG(s []byte) error {
	dec := xml.NewDecoder(bytes.NewReader(s))
	sawRoot := false
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("compiled SVG is not well-formed XML: %w", err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}
		if !sawRoot {
			sawRoot = true
			if start.Name.Local != "svg" {
				return fmt.Errorf("compiled output root is <%s>, want <svg>", start.Name.Local)
			}
		}
		if start.Name.Local == "script" {
			return fmt.Errorf("<script> elements are not allowed in diagrams")
		}
	}
	if !sawRoot {
		return fmt.Errorf("no <svg> element was produced")
	}
	return nil
}

// generateMermaidSVG asks the content model for Mermaid syntax, compiles it to
// SVG in the browser, and returns the SVG. The compile is folded into the
// generation accept() so malformed syntax triggers the model's one-shot
// correction round. The Mermaid source is also written next to the SVG for
// auditing/editability.
func generateMermaidSVG(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec project.DiagramSpec, compiler DiagramCompiler, critique string) ([]byte, error) {
	data := mermaidPromptData{
		Title:     l.FrontMatter.Title,
		ID:        spec.ID,
		Prompt:    spec.Prompt,
		Critique:  critique,
		Narration: spec.Narration,
		Audience:  cfg.Style.Audience,
	}
	system, user, err := e.renderPrompt(mermaidDiagramTemplateName, data)
	if err != nil {
		return nil, err
	}

	theme := videoThemeForConfig(cfg, l.FrontMatter.Title)
	var svg []byte
	accept := func(content string) (string, error) {
		syntax := strings.TrimSpace(stripFences(content))
		if err := validateMermaidSyntax(syntax); err != nil {
			return "", err
		}
		out, rerr := compiler.RenderMermaid(ctx, syntax, theme)
		if rerr != nil {
			return "", fmt.Errorf("mermaid did not render: %v", rerr)
		}
		if verr := validateCompiledSVG(out); verr != nil {
			return "", fmt.Errorf("mermaid rendered invalid SVG: %v", verr)
		}
		svg = out
		return syntax, nil
	}

	syntax, err := e.completeWithRepair(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.3, 2048, false, accept)
	if err != nil {
		return nil, fmt.Errorf("generating mermaid diagram %q: %w", spec.ID, err)
	}
	srcPath := filepath.Join(l.GeneratedDir(), DiagramsDirName, spec.ID+".mmd")
	if err := writeFileAtomic(srcPath, []byte(syntax+"\n")); err != nil {
		return nil, err
	}
	return svg, nil
}

// produceMermaidDiagram compiles a Mermaid diagram to SVG and runs it through
// the shared vision-QA loop. With no browser to compile with, it degrades to a
// freehand SVG of the same request so the build still completes.
func produceMermaidDiagram(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec project.DiagramSpec) ([]byte, error) {
	compiler, ok := e.Screenshotter.(DiagramCompiler)
	if !ok {
		fmt.Fprintf(e.out(), "  ⚠ visuals   no headless browser — compiling %q as freehand SVG instead of mermaid\n", spec.ID)
	}
	gen := func(ctx context.Context, critique string) ([]byte, error) {
		if !ok {
			return generateDiagram(ctx, e, l, cfg, spec, critique)
		}
		return generateMermaidSVG(ctx, e, l, cfg, spec, compiler, critique)
	}
	return runDiagramVisionQA(ctx, e, l, cfg, spec, gen)
}
