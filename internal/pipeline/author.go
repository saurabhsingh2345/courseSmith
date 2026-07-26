package pipeline

// Author stage: turns a one-line request into a complete lesson.md.
//
// Everything downstream of lesson.md was already generated, but lesson.md
// itself was hand-written — front-matter outcomes, per-diagram prompts,
// sectioned bullets, code blocks with their expected output, demo markers.
// That is ~100 lines of exacting authoring before the pipeline can do
// anything, and it is the single biggest cost of adding a lesson.
//
// AuthorLesson closes that gap: prompt in, lesson.md out. The result is
// ordinary lesson.md source — editable, reviewable, and identical in shape to
// a hand-written one — so nothing downstream knows the difference.

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

const outlineTemplateName = "outline.tmpl"

// Shape bounds for a drafted lesson. These are prompt inputs, not hard
// validation limits — except where noted in validate().
const (
	minOutlineSections = 5
	maxOutlineSections = 8
	minOutlineCode     = 2
	maxOutlineCode     = 4
	outlineDiagrams    = "1-2"
	minOutlineOutcomes = 2
	maxOutlineOutcomes = 4
)

// DraftSection is one section of a drafted lesson.
type DraftSection struct {
	Heading string   `json:"heading"`
	Bullets []string `json:"bullets"`
	Code    string   `json:"code,omitempty"`
	Output  string   `json:"output,omitempty"`
	Diagram string   `json:"diagram,omitempty"`
	Demo    string   `json:"demo,omitempty"`
}

// DraftDiagram is one declared diagram of a drafted lesson.
type DraftDiagram struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Prompt string `json:"prompt"`
}

// Draft is the model's structured lesson outline, before it becomes markdown.
type Draft struct {
	Title    string         `json:"title"`
	Summary  string         `json:"summary"`
	Outcomes []string       `json:"outcomes"`
	Diagrams []DraftDiagram `json:"diagrams"`
	Sections []DraftSection `json:"sections"`
}

// outlinePromptData feeds prompts/outline.tmpl.
type outlinePromptData struct {
	Prompt            string
	Tone              string
	Audience          string
	Language          string
	CourseName        string
	CourseDescription string
	ExistingLessons   []string
	Critique          string
	MinSections       int
	MaxSections       int
	MinCode           int
	MaxCode           int
	DiagramCount      string
}

// AuthorOptions carries the optional course context for a draft. A draft can
// be written with none of it — that is the point of the unfiled-draft flow —
// but supplying it produces a lesson that fits the course it will join.
type AuthorOptions struct {
	CourseName        string
	CourseDescription string
	// ExistingLessons are lesson titles already in the course, so the draft
	// neither duplicates nor contradicts them.
	ExistingLessons []string
}

var draftIDRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// validate rejects drafts the pipeline could not process, and only those. It
// deliberately does not enforce the prompt's soft ranges (section count, code
// count) — a good six-section lesson should not be thrown away for asking.
func (d *Draft) validate() error {
	if strings.TrimSpace(d.Title) == "" {
		return fmt.Errorf("draft has no title")
	}
	if n := len(d.Outcomes); n < minOutlineOutcomes || n > maxOutlineOutcomes {
		return fmt.Errorf("got %d outcomes, want %d-%d", n, minOutlineOutcomes, maxOutlineOutcomes)
	}
	if len(d.Sections) == 0 {
		return fmt.Errorf("draft has no sections")
	}

	declared := make(map[string]bool, len(d.Diagrams))
	for i, dg := range d.Diagrams {
		if !draftIDRe.MatchString(dg.ID) {
			return fmt.Errorf("diagram %d: id %q is not a kebab-case slug", i+1, dg.ID)
		}
		if declared[dg.ID] {
			return fmt.Errorf("diagram id %q declared twice", dg.ID)
		}
		if strings.TrimSpace(dg.Prompt) == "" {
			return fmt.Errorf("diagram %q has no prompt", dg.ID)
		}
		declared[dg.ID] = true
	}

	// Headings must be distinct after normalization, or two sections collapse
	// onto the same scene-graph key and one silently loses its content.
	seen := make(map[string]bool, len(d.Sections))
	referenced := make(map[string]bool, len(d.Diagrams))
	for i, s := range d.Sections {
		if strings.TrimSpace(s.Heading) == "" {
			return fmt.Errorf("section %d has no heading", i+1)
		}
		key := sectionKey(s.Heading)
		if key == "" {
			return fmt.Errorf("section %d heading %q has no alphanumeric content", i+1, s.Heading)
		}
		if seen[key] {
			return fmt.Errorf("two sections share the heading %q", s.Heading)
		}
		seen[key] = true
		if len(s.Bullets) == 0 {
			return fmt.Errorf("section %q has no bullets", s.Heading)
		}
		if s.Diagram != "" {
			if !declared[s.Diagram] {
				return fmt.Errorf("section %q references undeclared diagram %q", s.Heading, s.Diagram)
			}
			if referenced[s.Diagram] {
				return fmt.Errorf("diagram %q is referenced by more than one section", s.Diagram)
			}
			referenced[s.Diagram] = true
		}
		if s.Output != "" && s.Code == "" {
			return fmt.Errorf("section %q declares output with no code", s.Heading)
		}
	}
	for id := range declared {
		if !referenced[id] {
			return fmt.Errorf("diagram %q is declared but never referenced by a section", id)
		}
	}
	return nil
}

// Markdown renders the draft as lesson.md source: YAML front-matter followed
// by the sectioned outline, in exactly the dialect LoadLesson parses and the
// script/visuals/demos stages expect.
func (d *Draft) Markdown() string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %s\n", yamlScalar(d.Title))
	if len(d.Outcomes) > 0 {
		b.WriteString("outcomes:\n")
		for _, o := range d.Outcomes {
			fmt.Fprintf(&b, "  - %s\n", yamlScalar(o))
		}
	}
	if len(d.Diagrams) > 0 {
		b.WriteString("diagrams:\n")
		for _, dg := range d.Diagrams {
			fmt.Fprintf(&b, "  - id: %s\n", dg.ID)
			kind := dg.Kind
			if kind == "" {
				kind = project.DiagramKindMermaid
			}
			fmt.Fprintf(&b, "    kind: %s\n", kind)
			fmt.Fprintf(&b, "    prompt: %s\n", yamlScalar(dg.Prompt))
		}
	}
	b.WriteString("---\n\n")

	fmt.Fprintf(&b, "# %s\n", d.Title)
	for _, s := range d.Sections {
		fmt.Fprintf(&b, "\n## %s\n", s.Heading)
		for _, bullet := range s.Bullets {
			fmt.Fprintf(&b, "- %s\n", bullet)
		}
		if s.Code != "" {
			fmt.Fprintf(&b, "\n```python\n%s\n```\n", strings.TrimRight(s.Code, "\n"))
			if s.Output != "" {
				fmt.Fprintf(&b, "\n```output\n%s\n```\n", strings.TrimRight(s.Output, "\n"))
			}
		}
		if s.Diagram != "" {
			fmt.Fprintf(&b, "\n[DIAGRAM: %s]\n", s.Diagram)
		}
		if s.Demo != "" {
			fmt.Fprintf(&b, "\n[DEMO: %s]\n", s.Demo)
		}
	}
	return b.String()
}

// yamlScalar quotes a string for YAML, escaping what a double-quoted scalar
// must escape. Diagram prompts routinely contain colons and quotes, which is
// exactly what breaks naive front-matter writing.
func yamlScalar(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\t", `\t`)
	return `"` + r.Replace(strings.TrimSpace(s)) + `"`
}

// AuthorLesson drafts a complete lesson from a free-text prompt.
func (e *Env) AuthorLesson(ctx context.Context, cfg config.Config, prompt string, opts AuthorOptions) (*Draft, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return nil, fmt.Errorf("a lesson prompt is required")
	}

	data := outlinePromptData{
		Prompt:            prompt,
		Tone:              cfg.Style.Tone,
		Audience:          cfg.Style.Audience,
		Language:          cfg.Style.Language,
		CourseName:        opts.CourseName,
		CourseDescription: opts.CourseDescription,
		ExistingLessons:   opts.ExistingLessons,
		MinSections:       minOutlineSections,
		MaxSections:       maxOutlineSections,
		MinCode:           minOutlineCode,
		MaxCode:           maxOutlineCode,
		DiagramCount:      outlineDiagrams,
	}
	system, user, err := e.renderPrompt(outlineTemplateName, data)
	if err != nil {
		return nil, err
	}

	var draft Draft
	if err := e.completeJSON(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.5, 4000, &draft,
		func() error { return draft.validate() }); err != nil {
		return nil, fmt.Errorf("drafting lesson: %w", err)
	}
	return &draft, nil
}
