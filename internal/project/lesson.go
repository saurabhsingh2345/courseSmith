package project

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/enfec/coursesmith/internal/config"
)

// LessonFileName is the source file expected in every lesson directory.
const LessonFileName = "lesson.md"

// GeneratedDirName is where pipeline stages write their outputs,
// relative to the lesson directory.
const GeneratedDirName = "generated"

// Lesson is a parsed lesson directory: YAML front-matter plus markdown body.
type Lesson struct {
	// ID is the lesson directory name, e.g. "01-what-is-python".
	ID string
	// Dir is the absolute path of the lesson directory.
	Dir string

	FrontMatter FrontMatter
	// Body is the markdown outline below the front-matter.
	Body string
}

// FrontMatter is the YAML block at the top of lesson.md.
type FrontMatter struct {
	Title    string        `yaml:"title"`
	Diagrams []DiagramSpec `yaml:"diagrams"`
	// Outcomes are shown on the lesson's animated title card.
	Outcomes []string `yaml:"outcomes"`
	// Callouts are annotation overlays anchored to spoken phrases.
	Callouts []CalloutSpec `yaml:"callouts"`
	// Overrides for course-level config; merged in config.Resolve.
	Style    config.Style    `yaml:"style"`
	Branding config.Branding `yaml:"branding"`
	Pipeline config.Pipeline `yaml:"pipeline"`
}

// Diagram kinds. "svg" is a standalone generated SVG (vision-QA gated); "d3"
// is a structured node-link graph rendered and animated by the renderer.
// "mermaid" and "excalidraw" declare a structured source (Mermaid syntax /
// Excalidraw elements) that the visuals stage compiles to a self-contained
// SVG — so they flow through the same vision-QA gate and inline-SVG renderer
// as the "svg" kind, but the model authors an easy-to-validate spec instead of
// freehand markup.
const (
	DiagramKindSVG        = "svg"
	DiagramKindD3         = "d3"
	DiagramKindD2         = "d2"
	DiagramKindMermaid    = "mermaid"
	DiagramKindExcalidraw = "excalidraw"
)

// DiagramSpec declares one diagram for the visuals stage.
type DiagramSpec struct {
	ID     string `yaml:"id"`
	Prompt string `yaml:"prompt"`
	// Kind selects how the diagram is produced and rendered: "svg" (default),
	// "d3", "mermaid", or "excalidraw". Empty means svg, keeping older lessons
	// working unchanged.
	Kind string `yaml:"kind,omitempty"`
	// Narration is runtime-only context: the narration of the script section
	// that cues this diagram, injected into generation prompts so the picture
	// matches what the narrator is actually saying. Never parsed from YAML.
	Narration string `yaml:"-" json:"-"`
}

// ResolvedKind returns the diagram kind, defaulting empty to svg.
func (d DiagramSpec) ResolvedKind() string {
	if d.Kind == "" {
		return DiagramKindSVG
	}
	return d.Kind
}

// CalloutSpec declares one arrow/circle/label overlay. It appears when the
// narrator speaks the anchor phrase.
type CalloutSpec struct {
	// Section is the section id (heading slug) the phrase occurs in.
	Section string `yaml:"section"`
	// Shape is "arrow" or "circle".
	Shape string `yaml:"shape"`
	// X and Y position the callout as fractions of the frame (0-1).
	X float64 `yaml:"x"`
	Y float64 `yaml:"y"`
	// Label is the short text shown beside the shape.
	Label string `yaml:"label"`
	// At is the spoken phrase that triggers the callout.
	At string `yaml:"at"`
	// DurMs is how long it stays on screen (default 4000).
	DurMs int `yaml:"dur_ms"`
}

// Overrides returns the lesson's config layer for config.Resolve.
func (f FrontMatter) Overrides() config.Config {
	return config.Config{Style: f.Style, Branding: f.Branding, Pipeline: f.Pipeline}
}

// LoadLesson reads and validates <dir>/lesson.md.
func LoadLesson(dir string) (*Lesson, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolving lesson dir %s: %w", dir, err)
	}
	path := filepath.Join(abs, LessonFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	fm, body, err := splitFrontMatter(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	l := &Lesson{
		ID:   filepath.Base(abs),
		Dir:  abs,
		Body: body,
	}
	if len(fm) > 0 {
		dec := yaml.NewDecoder(bytes.NewReader(fm))
		dec.KnownFields(true)
		if err := dec.Decode(&l.FrontMatter); err != nil {
			return nil, fmt.Errorf("parsing front-matter of %s: %w", path, err)
		}
	}
	if err := l.Validate(); err != nil {
		return nil, fmt.Errorf("invalid %s: %w", path, err)
	}
	return l, nil
}

// Validate checks front-matter sanity.
func (l *Lesson) Validate() error {
	if l.FrontMatter.Title == "" {
		return fmt.Errorf("front-matter title is required")
	}
	if strings.TrimSpace(l.Body) == "" {
		return fmt.Errorf("lesson body is empty — add an outline below the front-matter")
	}
	seen := map[string]bool{}
	for i, d := range l.FrontMatter.Diagrams {
		if d.ID == "" {
			return fmt.Errorf("diagrams[%d]: id is required", i)
		}
		if !slugRe.MatchString(d.ID) {
			return fmt.Errorf("diagrams[%d]: id %q must be lowercase letters, digits, and hyphens", i, d.ID)
		}
		if d.Prompt == "" {
			return fmt.Errorf("diagram %q: prompt is required", d.ID)
		}
		switch d.ResolvedKind() {
		case DiagramKindSVG, DiagramKindD3, DiagramKindD2, DiagramKindMermaid, DiagramKindExcalidraw:
		default:
			return fmt.Errorf("diagram %q: kind %q is not one of svg, d3, d2, mermaid, excalidraw", d.ID, d.ResolvedKind())
		}
		if seen[d.ID] {
			return fmt.Errorf("duplicate diagram id %q", d.ID)
		}
		seen[d.ID] = true
	}
	for i, c := range l.FrontMatter.Callouts {
		if c.Shape != "arrow" && c.Shape != "circle" {
			return fmt.Errorf("callouts[%d]: shape %q must be arrow or circle", i, c.Shape)
		}
		if c.Section == "" || c.At == "" {
			return fmt.Errorf("callouts[%d]: section and at (the spoken anchor phrase) are required", i)
		}
		if c.X < 0 || c.X > 1 || c.Y < 0 || c.Y > 1 {
			return fmt.Errorf("callouts[%d]: x/y must be fractions of the frame (0-1)", i)
		}
	}
	return nil
}

// SourcePath returns the absolute path of the lesson's lesson.md.
func (l *Lesson) SourcePath() string {
	return filepath.Join(l.Dir, LessonFileName)
}

// GeneratedDir returns the absolute path of the lesson's output directory.
func (l *Lesson) GeneratedDir() string {
	return filepath.Join(l.Dir, GeneratedDirName)
}

// splitFrontMatter separates a leading "---\n...\n---\n" YAML block from the
// markdown body. A file without front-matter returns nil front-matter.
func splitFrontMatter(data []byte) (frontMatter []byte, body string, err error) {
	const delim = "---"
	s := string(data)
	// Normalize CRLF so the delimiter scan below only deals with \n.
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], " \t") != delim {
		return nil, s, nil
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], " \t") == delim {
			fm := strings.Join(lines[1:i], "\n")
			body = strings.Join(lines[i+1:], "\n")
			return []byte(fm), body, nil
		}
	}
	return nil, "", fmt.Errorf("front-matter opened with %q but never closed", delim)
}
