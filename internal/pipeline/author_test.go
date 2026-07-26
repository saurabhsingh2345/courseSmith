package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/project"
)

func goodDraft() *Draft {
	return &Draft{
		Title:    "What's a variable?",
		Summary:  "Store a value and use it later.",
		Outcomes: []string{"Store a value", "Reuse it later"},
		Diagrams: []DraftDiagram{{
			ID:     "box-and-label",
			Kind:   project.DiagramKindMermaid,
			Prompt: `A flowchart: "name" points to a box holding 'Ada'.`,
		}},
		Sections: []DraftSection{
			{
				Heading: "Giving a value a name",
				Bullets: []string{"A variable is a label", "The label points at a value"},
				Code:    "name = \"Ada\"\nprint(name)",
				Output:  "Ada",
				Diagram: "box-and-label",
			},
			{
				Heading: "What's next",
				Bullets: []string{"Next up: changing what's inside"},
				Demo:    "open the REPL and assign a variable",
			},
		},
	}
}

// The drafted markdown must be parseable by the same loader that reads a
// hand-written lesson — otherwise a draft is only nominally a lesson.
func TestDraftMarkdownLoadsAsLesson(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, project.LessonFileName), []byte(goodDraft().Markdown()), 0o644); err != nil {
		t.Fatal(err)
	}
	lesson, err := project.LoadLesson(dir)
	if err != nil {
		t.Fatalf("LoadLesson: %v", err)
	}
	if err := lesson.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if lesson.FrontMatter.Title != "What's a variable?" {
		t.Errorf("title = %q", lesson.FrontMatter.Title)
	}
	if n := len(lesson.FrontMatter.Diagrams); n != 1 {
		t.Fatalf("diagrams = %d", n)
	}
	// The prompt contains both a colon and double quotes — the case that
	// breaks front-matter written without escaping.
	if got := lesson.FrontMatter.Diagrams[0].Prompt; !strings.Contains(got, `"name"`) {
		t.Errorf("diagram prompt lost its quoting: %q", got)
	}
	if got := lesson.FrontMatter.Diagrams[0].ResolvedKind(); got != project.DiagramKindMermaid {
		t.Errorf("kind = %q", got)
	}
}

// The markers and code fences the downstream stages key off must survive.
func TestDraftMarkdownEmitsPipelineMarkers(t *testing.T) {
	md := goodDraft().Markdown()
	for _, want := range []string{
		"## Giving a value a name",
		"## What's next",
		"```python\nname = \"Ada\"\nprint(name)\n```",
		"```output\nAda\n```",
		"[DIAGRAM: box-and-label]",
		"[DEMO: open the REPL and assign a variable]",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("markdown missing %q\n---\n%s", want, md)
		}
	}

	// Sections and their headings must round-trip through the same lookup the
	// scene graph uses, apostrophes included.
	sections := sectionsFromOutline(md)
	if !strings.Contains(sections[sectionKey("whats-next")], "changing what's inside") {
		t.Errorf("apostrophe heading did not round-trip: %v", sections)
	}
}

func TestDraftValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Draft)
		wantErr string
	}{
		{"ok", func(*Draft) {}, ""},
		{"no title", func(d *Draft) { d.Title = " " }, "no title"},
		{"too few outcomes", func(d *Draft) { d.Outcomes = d.Outcomes[:1] }, "want 2-4"},
		{"no sections", func(d *Draft) { d.Sections = nil }, "no sections"},
		{"bad diagram id", func(d *Draft) { d.Diagrams[0].ID = "Box Label" }, "kebab-case"},
		{"undeclared diagram", func(d *Draft) { d.Sections[0].Diagram = "nope" }, "undeclared diagram"},
		{"unused diagram", func(d *Draft) { d.Sections[0].Diagram = "" }, "never referenced"},
		{"empty bullets", func(d *Draft) { d.Sections[1].Bullets = nil }, "no bullets"},
		{"output without code", func(d *Draft) { d.Sections[1].Output = "hi" }, "output with no code"},
		{
			"duplicate headings",
			// Differs only by punctuation — still the same scene-graph key.
			func(d *Draft) { d.Sections[1].Heading = "Giving a value, a name!" },
			"share the heading",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := goodDraft()
			tt.mutate(d)
			err := d.validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}
