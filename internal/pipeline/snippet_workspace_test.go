package pipeline

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func workspacePlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "workspace",
		Title:    "Share one function across files",
		Project: &ProjectSpec{
			Entry: "main.py",
			Files: []ProjectFile{
				{Path: "greet.py", Code: "def hello(who):\n    return f\"Hello, {who}!\"\n"},
				{Path: "main.py", Code: "from greet import hello\n\nprint(hello(\"Ada\"))\n"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "the-need", Heading: "One greeting, two places", Narration: strings.Repeat("need ", 22),
				Work: &WorkspaceBeat{File: "greet.py", Focus: "tree"}},
			{ID: "the-import", Heading: "Import it next door", Narration: strings.Repeat("import ", 22),
				Work: &WorkspaceBeat{File: "main.py", Focus: "code"}},
			{ID: "the-proof", Heading: "Run it", Narration: strings.Repeat("proof ", 22),
				Work: &WorkspaceBeat{File: "main.py", Focus: "terminal", Run: true}},
		},
	}
}

func TestValidateWorkspacePlan(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := workspacePlan().Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	// One file is a legitimate clip: this template is chosen for how it looks,
	// and requiring a second only ever produced one invented to satisfy the
	// rule.
	t.Run("a single file is allowed", func(t *testing.T) {
		p := workspacePlan()
		p.Project.Files = p.Project.Files[1:2]
		p.Project.Files[0].Code = "print(\"Hello, Ada!\")\n"
		p.Project.Entry = "main.py"
		for i := range p.Beats {
			p.Beats[i].Work.File = "main.py"
		}
		if err := p.Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	t.Run("entry must be one of the files", func(t *testing.T) {
		p := workspacePlan()
		p.Project.Entry = "nope.py"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "not one of the project") {
			t.Fatalf("want entry error, got %v", err)
		}
	})
	// The paths are written to a real directory before the program runs.
	t.Run("paths may not escape", func(t *testing.T) {
		p := workspacePlan()
		p.Project.Files[0].Path = "../greet.py"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "plain relative path") {
			t.Fatalf("want path error, got %v", err)
		}
	})
	t.Run("a beat must name a real file", func(t *testing.T) {
		p := workspacePlan()
		p.Beats[1].Work.File = "ghost.py"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "which the project does not have") {
			t.Fatalf("want unknown-file error, got %v", err)
		}
	})
	// A project nobody runs is a project nobody can believe.
	t.Run("something must run", func(t *testing.T) {
		p := workspacePlan()
		p.Beats[2].Work.Run = false
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "no beat runs the project") {
			t.Fatalf("want no-run error, got %v", err)
		}
	})
	t.Run("not on the first beat", func(t *testing.T) {
		p := workspacePlan()
		p.Beats[0].Work.Run = true
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "before any of it is on screen") {
			t.Fatalf("want early-run error, got %v", err)
		}
	})
}

func TestWorkspaceScenes(t *testing.T) {
	plan := workspacePlan()
	plan.Project.Output = "Hello, Ada!"
	scenes, err := workspaceScenes(sceneInput(t, plan, 6000))
	if err != nil {
		t.Fatal(err)
	}
	// One scene for the whole clip: a scene per beat would remount the editor,
	// and a remount is a cut in the middle of a screen recording.
	if len(scenes) != 1 {
		t.Fatalf("got %d scenes, want exactly 1", len(scenes))
	}
	s := scenes[0]
	if s.Type != SceneWorkspace {
		t.Errorf("scene type = %q, want %q", s.Type, SceneWorkspace)
	}
	if got := s.Props["output"]; got != "Hello, Ada!" {
		t.Errorf("output = %v, want the executed program's real stdout", got)
	}
	if got := s.Props["command"]; got != "python3 main.py" {
		t.Errorf("command = %v, want it derived from the entry", got)
	}
	steps, _ := s.Props["steps"].([]map[string]any)
	if len(steps) != len(plan.Beats) {
		t.Fatalf("got %d steps, want one per beat", len(steps))
	}
	if steps[0]["focus"] != "tree" || steps[2]["run"] != true {
		t.Errorf("steps did not carry focus/run: %+v", steps)
	}
}

// The camera vocabulary lives in two places: Go validates the names a model may
// emit and the renderer turns each into a camera. A focus Go allows and the
// scene does not know silently becomes `wide`, so the clip renders — just
// never looking at the thing the beat is about.
func TestWorkspaceFocusVocabularyInSync(t *testing.T) {
	const mirror = "../../renderer/src/components/WorkspaceScene.tsx"
	src, err := os.ReadFile(mirror)
	if err != nil {
		t.Fatalf("reading %s: %v", mirror, err)
	}
	block := regexp.MustCompile(`(?s)const camFor =.*?\n};`).Find(src)
	if block == nil {
		t.Fatalf("no camFor found in %s — has its shape changed?", mirror)
	}
	drawn := map[string]bool{"wide": true} // the default arm
	for _, m := range regexp.MustCompile(`case '([a-z]+)':`).FindAllSubmatch(block, -1) {
		drawn[string(m[1])] = true
	}
	var missing []string
	for name := range workspaceFocusVocab {
		if !drawn[name] {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Errorf("workspaceFocusVocab allows %v, which camFor does not aim at — those beats would sit wide", missing)
	}
	for name := range drawn {
		if _, ok := workspaceFocusVocab[name]; !ok {
			t.Errorf("camFor aims at %q, which workspaceFocusVocab rejects — nobody can ask for it", name)
		}
	}
}
