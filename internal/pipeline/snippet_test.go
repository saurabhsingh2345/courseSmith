package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// vscodePlan is a well-formed three-beat vscode plan: an intro with no code, a
// beat that writes it, and a beat that runs it.
func vscodePlan() *SnippetPlan {
	code := "for i in range(3):\n    print(i)"
	return &SnippetPlan{
		Template: "vscode",
		Title:    "For Loops in Python",
		Subtitle: "Repeat work without repeating yourself",
		Beats: []SnippetBeat{
			{ID: "the-idea", Heading: "The idea", Narration: strings.Repeat("idea ", 20)},
			{ID: "write-it", Heading: "Writing it", Narration: strings.Repeat("write ", 20), Code: code},
			{ID: "run-it", Heading: "Running it", Narration: strings.Repeat("run ", 20), Code: code, Run: true},
		},
	}
}

func TestSnippetSpecValidate(t *testing.T) {
	cases := []struct {
		name string
		spec SnippetSpec
		want string
	}{
		{"no prompt", SnippetSpec{Template: "vscode"}, "prompt is required"},
		{"no template", SnippetSpec{Prompt: "loops"}, "template is required"},
		{"unknown template", SnippetSpec{Prompt: "loops", Template: "nope"}, "unknown template"},
		{"target too short", SnippetSpec{Prompt: "loops", Template: "vscode", TargetSec: 3}, "out of range"},
		{"target too long", SnippetSpec{Prompt: "loops", Template: "vscode", TargetSec: 9000}, "out of range"},
		{"valid", SnippetSpec{Prompt: "loops", Template: "vscode"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate()
			if tc.want == "" {
				if err != nil {
					t.Fatalf("want valid, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestSnippetSpecDefaults(t *testing.T) {
	var spec SnippetSpec
	if got := spec.ResolvedTargetSec(); got != defaultSnippetTargetSec {
		t.Errorf("default target = %d, want %d", got, defaultSnippetTargetSec)
	}
	if got := (SnippetSpec{TargetSec: 1}).ResolvedTargetSec(); got != minSnippetTargetSec {
		t.Errorf("clamped target = %d, want %d", got, minSnippetTargetSec)
	}
	if got := spec.ResolvedCodeLanguage(); got != "python" {
		t.Errorf("default code language = %q, want python", got)
	}
}

func TestSnippetPlanValidate(t *testing.T) {
	t.Run("duplicate beat ids", func(t *testing.T) {
		p := vscodePlan()
		p.Beats[1].ID = p.Beats[0].ID
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "duplicate beat id") {
			t.Fatalf("want duplicate-id error, got %v", err)
		}
	})
	t.Run("empty narration", func(t *testing.T) {
		p := vscodePlan()
		p.Beats[1].Narration = "  "
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "empty narration") {
			t.Fatalf("want empty-narration error, got %v", err)
		}
	})
	t.Run("valid", func(t *testing.T) {
		if err := vscodePlan().Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
}

func TestValidateVSCodePlan(t *testing.T) {
	t.Run("no code at all", func(t *testing.T) {
		p := vscodePlan()
		for i := range p.Beats {
			p.Beats[i].Code = ""
			p.Beats[i].Run = false
		}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "at least one beat with code") {
			t.Fatalf("want missing-code error, got %v", err)
		}
	})
	t.Run("runs before any code exists", func(t *testing.T) {
		p := vscodePlan()
		p.Beats[0].Run = true
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "before any beat has written code") {
			t.Fatalf("want run-ordering error, got %v", err)
		}
	})
	t.Run("never runs", func(t *testing.T) {
		p := vscodePlan()
		p.Beats[2].Run = false
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "must end by running the code") {
			t.Fatalf("want missing-run error, got %v", err)
		}
	})
	t.Run("beat narration too thin", func(t *testing.T) {
		p := vscodePlan()
		p.Beats[0].Narration = "too short"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "words of narration") {
			t.Fatalf("want beat-length error, got %v", err)
		}
	})
}

// A snippet's lesson.md must be an ordinary lesson: it has to load with the
// same loader the pipeline uses, and the verify stage has to find its code.
func TestSnippetPlanMarkdownIsALesson(t *testing.T) {
	plan := vscodePlan()
	md, err := plan.Markdown(SnippetSpec{Template: "vscode"})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, project.LessonFileName), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	lesson, err := project.LoadLesson(dir)
	if err != nil {
		t.Fatalf("snippet markdown does not load as a lesson: %v", err)
	}
	if lesson.FrontMatter.Title != plan.Title {
		t.Errorf("title = %q, want %q", lesson.FrontMatter.Title, plan.Title)
	}
	blocks := extractCodeBlocks(lesson.Body, "lesson.md")
	if len(blocks) != 1 {
		t.Fatalf("got %d code blocks, want 1 (the two beats share one buffer state)", len(blocks))
	}
	if !strings.Contains(blocks[0].Code, "range(3)") {
		t.Errorf("code block does not carry the plan's code: %q", blocks[0].Code)
	}
	for _, b := range plan.Beats {
		if !strings.Contains(lesson.Body, "## "+b.Heading) {
			t.Errorf("body is missing heading for beat %q", b.ID)
		}
	}
}

func TestSnippetPlanScript(t *testing.T) {
	plan := vscodePlan()
	script := plan.Script(150)
	if script.Title != plan.Title {
		t.Errorf("script title = %q, want %q", script.Title, plan.Title)
	}
	if len(script.Sections) != len(plan.Beats) {
		t.Fatalf("got %d sections, want %d", len(script.Sections), len(plan.Beats))
	}
	for i, sec := range script.Sections {
		if sec.ID != plan.Beats[i].ID {
			t.Errorf("section %d id = %q, want %q", i, sec.ID, plan.Beats[i].ID)
		}
		if sec.DurationEstSec <= 0 {
			t.Errorf("section %q has non-positive duration estimate", sec.ID)
		}
	}
	// The script must satisfy the same validation the generated path applies,
	// or the audio and align stages would reject it.
	if err := script.Validate(map[string]bool{}); err != nil {
		t.Fatalf("derived script is invalid: %v", err)
	}
}

// sceneInput builds a timed SnippetSceneInput for a plan, with every beat the
// same length, and the plan's code marked verified.
func sceneInput(t *testing.T, plan *SnippetPlan, beatMs int) SnippetSceneInput {
	t.Helper()
	spans := make([]SectionSpan, len(plan.Beats))
	ends := make([]int, len(plan.Beats))
	for i, b := range plan.Beats {
		spans[i] = SectionSpan{ID: b.ID, StartMs: i * beatMs, EndMs: (i + 1) * beatMs}
		ends[i] = (i + 1) * beatMs
	}
	verified := map[string]string{}
	for _, b := range plan.Beats {
		if b.Code != "" {
			verified[project.HashBytes([]byte(b.Code))] = "0\n1\n2\n"
		}
	}
	return SnippetSceneInput{
		Spec:           SnippetSpec{Template: plan.Template},
		Plan:           plan,
		Course:         &project.Course{Name: "Snippets", Slug: "snippets"},
		Spans:          spans,
		BeatEndMs:      ends,
		VerifiedOutput: verified,
		DurationMs:     len(plan.Beats) * beatMs,
	}
}

func TestVSCodeScenes(t *testing.T) {
	plan := vscodePlan()
	scenes, err := vscodeScenes(sceneInput(t, plan, 12000))
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != 2 {
		t.Fatalf("got %d scenes, want a title card plus a walkthrough", len(scenes))
	}
	title, walk := scenes[0], scenes[1]
	if title.Type != SceneTitle || walk.Type != SceneWalkthrough {
		t.Fatalf("scene types = %s, %s; want title, walkthrough", title.Type, walk.Type)
	}

	// The title card is capped: a 12-second intro beat must not become a
	// 12-second held card.
	if got := title.EndMs - title.StartMs; got > maxTitleCardMs {
		t.Errorf("title card runs %dms, want at most %d", got, maxTitleCardMs)
	}
	// And the editor takes over the rest of the intro rather than leaving a gap.
	if walk.StartMs != title.EndMs {
		t.Errorf("walkthrough starts at %d but the title ends at %d — dead frames between", walk.StartMs, title.EndMs)
	}
	if walk.EndMs != 36000 {
		t.Errorf("walkthrough ends at %d, want the end of the clip (36000)", walk.EndMs)
	}

	// Typing must wait for the first code beat, not start under the title.
	typeAt, ok := walk.Props["typeAtMs"].(int)
	if !ok {
		t.Fatalf("walkthrough has no typeAtMs: %#v", walk.Props)
	}
	if typeAt != 12000 {
		t.Errorf("typeAtMs = %d, want the first code beat's start (12000)", typeAt)
	}

	steps, ok := walk.Props["steps"].([]map[string]any)
	if !ok {
		t.Fatalf("walkthrough steps have the wrong shape: %#v", walk.Props["steps"])
	}
	if len(steps) != 2 {
		t.Fatalf("got %d steps, want one for writing and one for running", len(steps))
	}
	if steps[0]["run"] != nil {
		t.Errorf("the writing step should not run anything: %#v", steps[0])
	}
	if steps[1]["run"] != true {
		t.Errorf("the running step is not marked run: %#v", steps[1])
	}
	if got := steps[1]["output"]; got != "0\n1\n2\n" {
		t.Errorf("run step output = %q, want the verified output", got)
	}
	if got, _ := steps[1]["command"].(string); !strings.HasPrefix(got, "python3 ") {
		t.Errorf("run command = %q, want a python3 invocation", got)
	}
	// The terminal opens inside the run beat, not at its edges.
	runAt, _ := steps[1]["runAtMs"].(int)
	if runAt <= 24000 || runAt >= 36000 {
		t.Errorf("runAtMs = %d, want it inside the run beat (24000-36000)", runAt)
	}
}

// A clip whose narration jumps straight into code gets no title card — there
// is nowhere to put one — and the editor still gets time to open.
func TestVSCodeScenesWithoutIntroBeat(t *testing.T) {
	plan := vscodePlan()
	plan.Beats = plan.Beats[1:] // drop the no-code intro
	plan.Beats = append(plan.Beats, SnippetBeat{
		ID: "wrap-up", Heading: "Wrap up", Narration: strings.Repeat("wrap ", 20),
	})
	scenes, err := vscodeScenes(sceneInput(t, plan, 10000))
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneWalkthrough {
		t.Fatalf("want a single walkthrough scene, got %d: %+v", len(scenes), scenes)
	}
	typeAt, _ := scenes[0].Props["typeAtMs"].(int)
	if typeAt < minTypingLeadMs {
		t.Errorf("typeAtMs = %d, want at least %d so the window can open first", typeAt, minTypingLeadMs)
	}
	// The trailing beat has no code and does not run: it must not add a step.
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if len(steps) != 2 {
		t.Errorf("got %d steps, want 2 — a beat that only talks over the buffer is not a step", len(steps))
	}
}

// The output shown in the terminal comes from the verify stage. If the code was
// never executed, the scene must refuse to build rather than show nothing.
func TestVSCodeScenesRequireVerifiedOutput(t *testing.T) {
	plan := vscodePlan()
	in := sceneInput(t, plan, 12000)
	in.VerifiedOutput = map[string]string{}
	_, err := vscodeScenes(in)
	if err == nil || !strings.Contains(err.Error(), "verify stage did not execute") {
		t.Fatalf("want a verify error, got %v", err)
	}
}

func TestBuildSnippetSceneGraph(t *testing.T) {
	plan := vscodePlan()
	spec := SnippetSpec{ID: "loops", Prompt: "loops", Template: "vscode"}
	alignment := &Alignment{}
	for i, b := range plan.Beats {
		alignment.Sections = append(alignment.Sections, SectionSpan{
			ID: b.ID, StartMs: i * 10000, EndMs: (i + 1) * 10000,
		})
		alignment.Words = append(alignment.Words, AlignedWord{
			Word: b.ID, StartMs: i * 10000, EndMs: i*10000 + 400,
		})
	}
	verification := &VerificationReport{Blocks: []VerifiedBlock{{
		CodeBlock: CodeBlock{Code: plan.Beats[1].Code},
		Stdout:    "0\n1\n2\n",
	}}}
	// Resolved the same way RunSnippet resolves it, so the branding defaults
	// the theme derives from are present.
	cfg := config.Resolve(config.Config{Style: config.Style{Captions: "on", PaceWPM: 175}}, config.Config{}, config.Config{})
	course := &project.Course{Name: "Snippets", Slug: "snippets"}

	graph, err := buildSnippetSceneGraph(course, &project.Lesson{ID: "loops"}, cfg, spec, plan, alignment, verification, 31000)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Scenes) == 0 {
		t.Fatal("no scenes")
	}
	// The last scene must reach the end of the graph: a gap there is a black
	// tail on the finished video.
	last := graph.Scenes[len(graph.Scenes)-1]
	if last.EndMs != graph.DurationMs {
		t.Errorf("last scene ends at %d but the graph runs to %d", last.EndMs, graph.DurationMs)
	}
	if graph.DurationMs < 31000 {
		t.Errorf("duration %d is shorter than the audio (31000) — the voice would be cut off", graph.DurationMs)
	}
	if graph.AudioFile != VoiceoverFileName {
		t.Errorf("audio file = %q, want %q", graph.AudioFile, VoiceoverFileName)
	}
	if len(graph.Captions) == 0 {
		t.Error("captions are on but the graph carries none")
	}
	if graph.Theme.Accent == "" || graph.Theme.BgTop == "" {
		t.Error("graph theme is missing derived design tokens")
	}
}

func TestBuildSnippetSceneGraphRejectsTimingMismatch(t *testing.T) {
	plan := vscodePlan()
	alignment := &Alignment{Sections: []SectionSpan{{ID: "the-idea", StartMs: 0, EndMs: 1000}}}
	_, err := buildSnippetSceneGraph(
		&project.Course{Name: "Snippets"}, &project.Lesson{ID: "x"}, config.Config{},
		SnippetSpec{Template: "vscode"}, plan, alignment, nil, 1000,
	)
	if err == nil || !strings.Contains(err.Error(), "re-run the align stage") {
		t.Fatalf("want an alignment-mismatch error, got %v", err)
	}
}

func TestCreateAndFindSnippet(t *testing.T) {
	root := t.TempDir()
	spec := SnippetSpec{Prompt: "How for loops work in Python", Template: "vscode"}

	course, lesson, err := CreateSnippet(root, spec)
	if err != nil {
		t.Fatal(err)
	}
	if course.Slug != SnippetsCourseSlug {
		t.Errorf("course slug = %q, want %q", course.Slug, SnippetsCourseSlug)
	}
	if !IsSnippet(lesson) {
		t.Error("created lesson is not recognized as a snippet")
	}
	if lesson.ID != "how-for-loops-work-in-python" {
		t.Errorf("id = %q, want a slug derived from the prompt", lesson.ID)
	}

	// A second snippet from the same prompt must not collide.
	_, second, err := CreateSnippet(root, spec)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID == lesson.ID {
		t.Errorf("second snippet reused id %q", second.ID)
	}

	loaded, err := LoadSnippetSpec(lesson.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Template != "vscode" || loaded.Prompt != spec.Prompt {
		t.Errorf("round-tripped spec = %+v", loaded)
	}
	if loaded.CreatedAt.IsZero() {
		t.Error("created_at was not stamped")
	}

	if _, found, err := FindSnippet(root, lesson.ID); err != nil {
		t.Fatal(err)
	} else if found.ID != lesson.ID {
		t.Errorf("FindSnippet returned %q, want %q", found.ID, lesson.ID)
	}
	if _, _, err := FindSnippet(root, "nope"); err == nil {
		t.Error("FindSnippet accepted a nonexistent id")
	}

	list, err := ListSnippets(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("listed %d snippets, want 2", len(list))
	}
}

func TestListSnippetsWithoutAny(t *testing.T) {
	list, err := ListSnippets(t.TempDir())
	if err != nil {
		t.Fatalf("listing an empty project should not error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("got %d snippets in an empty project", len(list))
	}
}

func TestSnippetFileNameAndWorkspace(t *testing.T) {
	cases := []struct{ title, lang, file, workspace string }{
		{"For Loops in Python", "python", "for_loops.py", "for-loops"},
		{"Async/await in JavaScript", "javascript", "async_await.js", "async-await"},
		{"Sorting", "go", "sorting.go", "sorting"},
		{"", "python", "main.py", "workspace"},
	}
	for _, tc := range cases {
		if got := snippetFileName(tc.title, tc.lang); got != tc.file {
			t.Errorf("snippetFileName(%q, %q) = %q, want %q", tc.title, tc.lang, got, tc.file)
		}
		if got := workspaceName(tc.title); got != tc.workspace {
			t.Errorf("workspaceName(%q) = %q, want %q", tc.title, got, tc.workspace)
		}
	}
}

func TestWordBudget(t *testing.T) {
	target, lo, hi := wordBudget(45, 175)
	if target != 131 {
		t.Errorf("target = %d, want 131 words for 45s at 175 wpm", target)
	}
	if lo >= target || hi <= target {
		t.Errorf("band %d-%d does not bracket the target %d", lo, hi, target)
	}
	// The band has to be wide enough that a good plan is not thrown away for
	// being a sentence off, and tight enough to catch a half-length draft.
	if lo > target*8/10 || hi < target*12/10 {
		t.Errorf("band %d-%d is too tight around %d", lo, hi, target)
	}
	// Asymmetric on purpose: short breaks the clip, long merely lengthens it.
	if target-lo >= hi-target {
		t.Errorf("band %d-%d should allow more slack above the target %d than below", lo, hi, target)
	}
	if got := narrationWords(vscodePlan()); got != 60 {
		t.Errorf("narrationWords = %d, want 60", got)
	}
}

// Every registered template must be wired up completely; a half-registered one
// would fail only at render time, on a real user's clip.
func TestSnippetTemplateCatalog(t *testing.T) {
	if len(SnippetTemplates) == 0 {
		t.Fatal("the template catalog is empty")
	}
	for _, tpl := range SnippetTemplateList() {
		if tpl.Name == "" || tpl.Title == "" || tpl.Description == "" || tpl.Example == "" {
			t.Errorf("template %q is missing gallery copy: %+v", tpl.Name, tpl)
		}
		if tpl.Scenes == nil {
			t.Errorf("template %q has no scene builder", tpl.Name)
		}
		if tpl.PromptFile == "" {
			t.Errorf("template %q names no prompt file", tpl.Name)
		}
		if _, err := os.Stat(filepath.Join("..", "..", "prompts", tpl.PromptFile)); err != nil {
			t.Errorf("template %q prompt file is missing: %v", tpl.Name, err)
		}
	}
	if !slicesContains(SnippetTemplateNames(), "vscode") {
		t.Error("the vscode template did not register itself")
	}
}

func slicesContains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
