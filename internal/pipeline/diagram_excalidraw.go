package pipeline

// Excalidraw diagram kind. The content model authors a small, constrained list
// of Excalidraw-style elements (rectangles, ellipses, diamonds, arrows, lines,
// text) on a fixed canvas; the visuals stage renders them with the embedded
// Rough.js — the same hand-drawn engine Excalidraw uses — into a self-contained
// SVG, then gates it through the same vision-QA loop as the other kinds. We
// support a deliberate subset (we control the generation prompt), which keeps
// the spec trivial to validate and the output faithful to the Excalidraw look.

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// Canvas bounds the model draws within. Fixed so positions are easy to reason
// about and the compiled SVG has a stable aspect ratio (16:10-ish).
const (
	excalidrawCanvasW = 1000.0
	excalidrawCanvasH = 640.0
)

// hexColorRe validates #rgb / #rrggbb colours.
var hexColorRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// excalidrawShapeTypes are the closed shapes (have width/height).
var excalidrawShapeTypes = map[string]bool{"rectangle": true, "ellipse": true, "diamond": true}

// excalidrawLineTypes are the point-path types (have points).
var excalidrawLineTypes = map[string]bool{"arrow": true, "line": true}

// ExcalidrawElement is one drawn element. Shapes use x/y/width/height; lines
// and arrows use x/y as an origin plus points relative to it. Kept in sync
// with excalidrawRenderJS in diagram_compile.go.
type ExcalidrawElement struct {
	Type            string       `json:"type"`
	X               float64      `json:"x"`
	Y               float64      `json:"y"`
	Width           float64      `json:"width,omitempty"`
	Height          float64      `json:"height,omitempty"`
	StrokeColor     string       `json:"strokeColor,omitempty"`
	BackgroundColor string       `json:"backgroundColor,omitempty"`
	FillStyle       string       `json:"fillStyle,omitempty"`
	StrokeWidth     float64      `json:"strokeWidth,omitempty"`
	Roughness       float64      `json:"roughness,omitempty"`
	Points          [][2]float64 `json:"points,omitempty"`
	Text            string       `json:"text,omitempty"`  // for type "text"
	Label           string       `json:"label,omitempty"` // centred label inside a shape
	FontSize        float64      `json:"fontSize,omitempty"`
}

// ExcalidrawScene is the persisted diagrams/<id>.excalidraw.json source.
type ExcalidrawScene struct {
	Kind     string              `json:"kind"` // always "excalidraw"
	Width    float64             `json:"width"`
	Height   float64             `json:"height"`
	Elements []ExcalidrawElement `json:"elements"`
}

// Validate enforces the structural contract: known element types, sane
// geometry within the canvas, valid colours, and non-empty text/labels.
func (s *ExcalidrawScene) Validate() error {
	if s.Width <= 0 || s.Height <= 0 {
		return fmt.Errorf("scene width and height are required")
	}
	if len(s.Elements) == 0 {
		return fmt.Errorf("at least one element is required")
	}
	if len(s.Elements) > 60 {
		return fmt.Errorf("too many elements (%d) — keep diagrams under 60 for legibility", len(s.Elements))
	}
	for i, el := range s.Elements {
		where := fmt.Sprintf("elements[%d] (%s)", i, el.Type)
		switch {
		case excalidrawShapeTypes[el.Type]:
			if el.Width <= 0 || el.Height <= 0 {
				return fmt.Errorf("%s: width and height must be positive", where)
			}
		case excalidrawLineTypes[el.Type]:
			if len(el.Points) < 2 {
				return fmt.Errorf("%s: needs at least 2 points", where)
			}
		case el.Type == "text":
			if strings.TrimSpace(el.Text) == "" {
				return fmt.Errorf("%s: text is required", where)
			}
		default:
			return fmt.Errorf("%s: type must be one of rectangle, ellipse, diamond, arrow, line, text", where)
		}
		if el.StrokeColor != "" && !hexColorRe.MatchString(el.StrokeColor) {
			return fmt.Errorf("%s: strokeColor %q is not a hex colour", where, el.StrokeColor)
		}
		if el.BackgroundColor != "" && !hexColorRe.MatchString(el.BackgroundColor) {
			return fmt.Errorf("%s: backgroundColor %q is not a hex colour", where, el.BackgroundColor)
		}
		if el.FillStyle != "" && el.FillStyle != "hachure" && el.FillStyle != "cross-hatch" && el.FillStyle != "solid" {
			return fmt.Errorf("%s: fillStyle %q must be hachure, cross-hatch, or solid", where, el.FillStyle)
		}
		if el.Roughness < 0 || el.Roughness > 3 {
			return fmt.Errorf("%s: roughness %.2f must be between 0 and 3", where, el.Roughness)
		}
	}
	return nil
}

// excalidrawPromptData feeds prompts/diagram_excalidraw.tmpl.
type excalidrawPromptData struct {
	Title    string
	ID       string
	Prompt   string
	CanvasW  float64
	CanvasH  float64
	Colors   config.Colors
	Critique string
	// Narration is the spoken text of the cueing section (context).
	Narration string
	Audience  string
}

// generateExcalidrawSVG asks the content model for an ExcalidrawScene,
// validates it structurally, renders it to SVG with Rough.js, and returns the
// SVG. Both parse/validate and render happen inside the generation accept() so
// a malformed scene triggers the model's correction round. The scene JSON is
// written next to the SVG for auditing/editability.
func generateExcalidrawSVG(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec project.DiagramSpec, compiler DiagramCompiler, critique string) ([]byte, error) {
	data := excalidrawPromptData{
		Title:     l.FrontMatter.Title,
		ID:        spec.ID,
		Prompt:    spec.Prompt,
		CanvasW:   excalidrawCanvasW,
		CanvasH:   excalidrawCanvasH,
		Colors:    cfg.Branding.Colors,
		Critique:  critique,
		Narration: spec.Narration,
		Audience:  cfg.Style.Audience,
	}
	system, user, err := e.renderPrompt(excalidrawTemplateName, data)
	if err != nil {
		return nil, err
	}

	var (
		svg      []byte
		canonSrc []byte
	)
	accept := func(content string) (string, error) {
		var scene ExcalidrawScene
		dec := json.NewDecoder(strings.NewReader(stripFences(content)))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&scene); err != nil {
			return "", fmt.Errorf("invalid Excalidraw JSON: %v", err)
		}
		scene.Kind = project.DiagramKindExcalidraw
		if scene.Width == 0 {
			scene.Width = excalidrawCanvasW
		}
		if scene.Height == 0 {
			scene.Height = excalidrawCanvasH
		}
		if err := scene.Validate(); err != nil {
			return "", err
		}
		canonical, err := json.MarshalIndent(&scene, "", "  ")
		if err != nil {
			return "", err
		}
		out, rerr := compiler.RenderExcalidraw(ctx, canonical, videoThemeForConfig(cfg, l.FrontMatter.Title))
		if rerr != nil {
			return "", fmt.Errorf("excalidraw did not render: %v", rerr)
		}
		if verr := validateCompiledSVG(out); verr != nil {
			return "", fmt.Errorf("excalidraw rendered invalid SVG: %v", verr)
		}
		svg = out
		canonSrc = canonical
		return string(canonical), nil
	}

	if _, err := e.completeWithRepair(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.3, 4096, false, accept); err != nil {
		return nil, fmt.Errorf("generating excalidraw diagram %q: %w", spec.ID, err)
	}
	srcPath := filepath.Join(l.GeneratedDir(), DiagramsDirName, spec.ID+".excalidraw.json")
	if err := writeFileAtomic(srcPath, append(canonSrc, '\n')); err != nil {
		return nil, err
	}
	return svg, nil
}

// produceExcalidrawDiagram compiles an Excalidraw scene to a hand-drawn SVG and
// runs it through the shared vision-QA loop. With no browser to compile with,
// it degrades to a freehand SVG of the same request so the build still
// completes.
func produceExcalidrawDiagram(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec project.DiagramSpec) ([]byte, error) {
	compiler, ok := e.Screenshotter.(DiagramCompiler)
	if !ok {
		fmt.Fprintf(e.out(), "  ⚠ visuals   no headless browser — compiling %q as freehand SVG instead of excalidraw\n", spec.ID)
	}
	gen := func(ctx context.Context, critique string) ([]byte, error) {
		if !ok {
			return generateDiagram(ctx, e, l, cfg, spec, critique)
		}
		return generateExcalidrawSVG(ctx, e, l, cfg, spec, compiler, critique)
	}
	return runDiagramVisionQA(ctx, e, l, cfg, spec, gen)
}
