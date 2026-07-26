package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/project"
)

// lessonWithMarkdown writes lesson.md to a temp dir and loads it (running the
// lesson validator). Returns the error so callers can assert on rejection.
func lessonWithMarkdown(t *testing.T, md string) (*project.Lesson, error) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "01-diag")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, project.LessonFileName), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	return project.LoadLesson(dir)
}

func lessonWithDiagrams(t *testing.T, md string) *project.Lesson {
	t.Helper()
	l, err := lessonWithMarkdown(t, md)
	if err != nil {
		t.Fatalf("loading lesson: %v", err)
	}
	return l
}

func treeSpec() *D3Spec {
	return &D3Spec{
		Kind:   project.DiagramKindD3,
		Layout: D3LayoutTree,
		Nodes:  []D3Node{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}, {ID: "c", Label: "C"}},
		Edges:  []D3Edge{{From: "a", To: "b"}, {From: "a", To: "c"}},
	}
}

func TestD3SpecValidateAcceptsGoodSpecs(t *testing.T) {
	if err := treeSpec().Validate(); err != nil {
		t.Errorf("valid tree rejected: %v", err)
	}
	force := &D3Spec{
		Layout: D3LayoutForce,
		Nodes:  []D3Node{{ID: "x", Label: "X"}, {ID: "y", Label: "Y"}},
		Edges:  []D3Edge{{From: "x", To: "y"}, {From: "y", To: "x"}}, // cycles are fine for force
	}
	if err := force.Validate(); err != nil {
		t.Errorf("valid force graph rejected: %v", err)
	}
}

func TestD3SpecValidateRejectsBadSpecs(t *testing.T) {
	cases := []struct {
		name string
		spec *D3Spec
		want string
	}{
		{"no layout", &D3Spec{Nodes: []D3Node{{ID: "a", Label: "A"}}}, "layout is required"},
		{"bad layout", &D3Spec{Layout: "grid", Nodes: []D3Node{{ID: "a", Label: "A"}}}, "not one of"},
		{"no nodes", &D3Spec{Layout: D3LayoutForce}, "at least one node"},
		{"missing id", &D3Spec{Layout: D3LayoutForce, Nodes: []D3Node{{Label: "A"}}}, "id is required"},
		{"missing label", &D3Spec{Layout: D3LayoutForce, Nodes: []D3Node{{ID: "a"}}}, "label is required"},
		{"dup id", &D3Spec{Layout: D3LayoutForce, Nodes: []D3Node{{ID: "a", Label: "A"}, {ID: "a", Label: "B"}}}, "duplicate node id"},
		{"edge to unknown", &D3Spec{Layout: D3LayoutForce, Nodes: []D3Node{{ID: "a", Label: "A"}}, Edges: []D3Edge{{From: "a", To: "z"}}}, "not a declared node"},
		{"self loop", &D3Spec{Layout: D3LayoutForce, Nodes: []D3Node{{ID: "a", Label: "A"}}, Edges: []D3Edge{{From: "a", To: "a"}}}, "self-loop"},
		{
			"tree with cycle",
			&D3Spec{Layout: D3LayoutTree, Nodes: []D3Node{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}},
				Edges: []D3Edge{{From: "a", To: "b"}, {From: "b", To: "a"}}},
			"one root",
		},
		{
			"tree with two roots",
			&D3Spec{Layout: D3LayoutTree, Nodes: []D3Node{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}, {ID: "c", Label: "C"}},
				Edges: []D3Edge{{From: "a", To: "c"}}},
			"exactly one root",
		},
		{
			"tree with two parents",
			&D3Spec{Layout: D3LayoutTree, Nodes: []D3Node{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}, {ID: "c", Label: "C"}},
				Edges: []D3Edge{{From: "a", To: "c"}, {From: "b", To: "c"}}},
			"parents",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate()
			if err == nil {
				t.Fatalf("expected an error mentioning %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestDiagramArtifactAndSceneNames(t *testing.T) {
	svg := project.DiagramSpec{ID: "mem", Prompt: "p"}
	d3 := project.DiagramSpec{ID: "graph", Prompt: "p", Kind: project.DiagramKindD3}

	if got := diagramArtifactName(svg); got != "mem.svg" {
		t.Errorf("svg artifact = %q, want mem.svg", got)
	}
	if got := diagramArtifactName(d3); got != "graph.json" {
		t.Errorf("d3 artifact = %q, want graph.json", got)
	}
	if got := diagramSceneSrc("mem", project.DiagramKindSVG); got != "diagrams/mem.svg" {
		t.Errorf("svg src = %q", got)
	}
	if got := diagramSceneSrc("graph", project.DiagramKindD3); got != "diagrams/graph.json" {
		t.Errorf("d3 src = %q", got)
	}
}

func TestDiagramKindByID(t *testing.T) {
	l := lessonWithDiagrams(t, "---\ntitle: T\ndiagrams:\n  - id: a\n    prompt: p\n  - id: b\n    prompt: p\n    kind: d3\n---\n\n## S\n\ntext\n")
	kinds := diagramKindByID(l)
	if kinds["a"] != project.DiagramKindSVG {
		t.Errorf("a kind = %q, want svg (default)", kinds["a"])
	}
	if kinds["b"] != project.DiagramKindD3 {
		t.Errorf("b kind = %q, want d3", kinds["b"])
	}
}

func TestLessonRejectsUnknownDiagramKind(t *testing.T) {
	_, err := lessonWithMarkdown(t, "---\ntitle: T\ndiagrams:\n  - id: a\n    prompt: p\n    kind: banana\n---\n\n## S\n\ntext\n")
	if err == nil || !strings.Contains(err.Error(), "kind") {
		t.Fatalf("expected a kind validation error, got %v", err)
	}
}

func TestLessonAcceptsCompiledDiagramKinds(t *testing.T) {
	l := lessonWithDiagrams(t, "---\ntitle: T\ndiagrams:\n  - id: a\n    prompt: p\n    kind: mermaid\n  - id: b\n    prompt: p\n    kind: excalidraw\n---\n\n## S\n\ntext\n")
	// Both compile to a self-contained SVG, so they publish and are referenced
	// as .svg — flowing through the same inline-SVG renderer as the svg kind.
	for _, d := range l.FrontMatter.Diagrams {
		if got := diagramArtifactName(d); got != d.ID+".svg" {
			t.Errorf("%s artifact = %q, want %s.svg", d.Kind, got, d.ID)
		}
		if got := diagramSceneSrc(d.ID, d.ResolvedKind()); got != "diagrams/"+d.ID+".svg" {
			t.Errorf("%s src = %q, want diagrams/%s.svg", d.Kind, got, d.ID)
		}
	}
}
