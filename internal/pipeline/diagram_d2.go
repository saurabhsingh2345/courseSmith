package pipeline

// D2 diagram kind: the model writes D2 source (a compact diagram DSL that is
// far harder to make ugly than raw SVG); the pipeline compiles it in-process
// with the pure-Go D2 library (MPL-2.0) — sketch aesthetic, deterministic
// dagre layout, no browser needed — and the compiled SVG flows through the
// same vision-QA gate and DiagramScene renderer as every other SVG kind.
// A compile failure is fed back to the model as its correction round.

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2target"
	d2log "oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// d2Ruler is the shared text-measuring ruler (expensive to build, safe to
// reuse).
var (
	d2RulerOnce sync.Once
	d2Ruler     *textmeasure.Ruler
	d2RulerErr  error
)

func getD2Ruler() (*textmeasure.Ruler, error) {
	d2RulerOnce.Do(func() {
		d2Ruler, d2RulerErr = textmeasure.NewRuler()
	})
	return d2Ruler, d2RulerErr
}

// compileD2 compiles D2 source to a standalone SVG. Dark Mauve (200) is the
// base; the background neutral is overridden to the course's derived stage
// colour so the compiled SVG sits seamlessly on the video background.
func compileD2(ctx context.Context, source string, theme SceneTheme) ([]byte, error) {
	ruler, err := getD2Ruler()
	if err != nil {
		return nil, fmt.Errorf("d2 text ruler: %w", err)
	}
	// D2 logs through a context slog; give it a discard-level default so
	// compile noise never reaches the pipeline output.
	ctx = d2log.WithDefault(ctx)

	sketch := true
	pad := int64(20)
	themeID := int64(200) // Dark Mauve
	bg := theme.BgTop
	overrides := &d2target.ThemeOverrides{N7: &bg}
	compileOpts := &d2lib.CompileOptions{
		Ruler: ruler,
		LayoutResolver: func(string) (d2graph.LayoutGraph, error) {
			return d2dagrelayout.DefaultLayout, nil
		},
	}
	renderOpts := &d2svg.RenderOpts{Pad: &pad, Sketch: &sketch, ThemeID: &themeID, ThemeOverrides: overrides}
	diagram, _, err := d2lib.Compile(ctx, source, compileOpts, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("d2 compile: %w", err)
	}
	svg, err := d2svg.Render(diagram, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("d2 render: %w", err)
	}
	return svg, nil
}

// d2PromptData feeds prompts/diagram_d2lang.tmpl.
type d2LangPromptData struct {
	Title    string
	ID       string
	Prompt   string
	Critique string
	// Narration is the spoken text of the cueing section (context).
	Narration string
	Audience  string
}

// generateD2SVG asks the content model for D2 source, compiles it in-process,
// and returns the SVG. The compile is folded into the generation accept() so
// invalid source triggers the model's correction round. The D2 source is
// written next to the SVG for auditing/editability.
func generateD2SVG(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec project.DiagramSpec, critique string) ([]byte, error) {
	data := d2LangPromptData{
		Title:     l.FrontMatter.Title,
		ID:        spec.ID,
		Prompt:    spec.Prompt,
		Critique:  critique,
		Narration: spec.Narration,
		Audience:  cfg.Style.Audience,
	}
	system, user, err := e.renderPrompt(d2LangTemplateName, data)
	if err != nil {
		return nil, err
	}

	theme := videoThemeForConfig(cfg, l.FrontMatter.Title)
	var svg []byte
	var source string
	accept := func(content string) (string, error) {
		src := strings.TrimSpace(stripFences(content))
		if src == "" {
			return "", fmt.Errorf("d2 source is empty")
		}
		compiled, err := compileD2(ctx, src, theme)
		if err != nil {
			return "", err
		}
		svg = compiled
		source = src
		return src, nil
	}
	if _, err := e.completeWithRepair(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.3, 4096, false, accept); err != nil {
		return nil, fmt.Errorf("generating d2 diagram %q: %w", spec.ID, err)
	}

	srcPath := filepath.Join(l.GeneratedDir(), DiagramsDirName, spec.ID+".d2")
	if err := writeFileAtomic(srcPath, []byte(source+"\n")); err != nil {
		return nil, err
	}
	return svg, nil
}

// produceD2Diagram generates D2 source, compiles it, and runs the compiled
// SVG through the shared vision-QA loop.
func produceD2Diagram(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec project.DiagramSpec) ([]byte, error) {
	return runDiagramVisionQA(ctx, e, l, cfg, spec, func(ctx context.Context, critique string) ([]byte, error) {
		return generateD2SVG(ctx, e, l, cfg, spec, critique)
	})
}
