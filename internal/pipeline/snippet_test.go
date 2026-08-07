package pipeline

import (
	"fmt"
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
		// Beat 0 is the opener, whose floor is the relaxed one — so this has to be
		// under THAT to be rejected, and the error should quote the floor it broke
		// rather than a single global number.
		p.Beats[0].Narration = "too short"
		err := p.Validate()
		if err == nil || !strings.Contains(err.Error(), "under their floor") {
			t.Fatalf("want beat-length error, got %v", err)
		}
		if !strings.Contains(err.Error(), "floor 6") {
			t.Errorf("the opener was judged against the wrong floor: %v", err)
		}
	})
	// The failure this catches: a beat hands back the lines it adds rather than
	// the whole file. Verify runs each buffer state on its own, so the second
	// one dies on a NameError and the clip cannot be published.
	t.Run("later beat sends a diff instead of the whole file", func(t *testing.T) {
		p := vscodePlan()
		p.Beats[1].Code = "integer_var = 42\nfloat_var = 3.14"
		p.Beats[2].Code = "print(integer_var)\nprint(float_var)"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "rewrites the file instead of editing it") {
			t.Fatalf("want whole-file error, got %v", err)
		}
	})
	// Editing a few lines is what the template is for; only replacing the file
	// wholesale is a mistake.
	t.Run("later beat edits a line and keeps the rest", func(t *testing.T) {
		p := vscodePlan()
		p.Beats[1].Code = "for i in range(3):\n    print(i)\nprint('done')"
		p.Beats[2].Code = "for i in range(5):\n    print(i)\nprint('done')"
		if err := p.Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
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

// A plan with several over-long beats must be told about all of them at once.
// Reporting one per round meant three correction rounds fixed three beats and
// the fourth still failed — the real failure behind a backpressure clip that
// died after three rounds with four separate "want 10-60" complaints.
func TestCheckBeatShapeReportsEveryOffender(t *testing.T) {
	plan := &SnippetPlan{Beats: []SnippetBeat{
		{ID: "the-problem", Narration: strings.Repeat("w ", 77)},
		{ID: "the-cause", Narration: strings.Repeat("w ", 62)},
		{ID: "the-effect", Narration: strings.Repeat("w ", 62)},
		{ID: "the-solution", Narration: strings.Repeat("w ", 72)},
	}}
	err := checkBeatShape(plan)
	if err == nil {
		t.Fatal("want an error")
	}
	for _, id := range []string{"the-problem", "the-cause", "the-effect", "the-solution"} {
		if !strings.Contains(err.Error(), id) {
			t.Errorf("error does not mention beat %q, so a correction round can only fix the ones it names:\n%s", id, err)
		}
	}
	// Told only the range, a model trims and fights its own content. With room
	// under the beat ceiling, splitting keeps the writing and satisfies the rule.
	if !strings.Contains(err.Error(), "SPLIT") {
		t.Errorf("error does not suggest splitting despite room for more beats:\n%s", err)
	}
}

// At the beat ceiling there is nowhere to split to, so the advice has to change.
func TestCheckBeatShapeAtBeatCeilingSaysTighten(t *testing.T) {
	beats := make([]SnippetBeat, maxSnippetBeats)
	for i := range beats {
		beats[i] = SnippetBeat{ID: fmt.Sprintf("b%d", i), Narration: strings.Repeat("w ", 70)}
	}
	err := checkBeatShape(&SnippetPlan{Beats: beats})
	if err == nil {
		t.Fatal("want an error")
	}
	if strings.Contains(err.Error(), "SPLIT") {
		t.Errorf("suggests splitting at the beat ceiling, which is impossible:\n%s", err)
	}
	if !strings.Contains(err.Error(), "tightened") {
		t.Errorf("want tighten advice at the ceiling:\n%s", err)
	}
}

// Under-long beats are the opposite problem and want the opposite advice.
func TestCheckBeatShapeReportsShortBeats(t *testing.T) {
	plan := &SnippetPlan{Beats: []SnippetBeat{
		{ID: "a", Narration: strings.Repeat("w ", 30)},
		{ID: "thin", Narration: "too short"},
		{ID: "c", Narration: strings.Repeat("w ", 30)},
	}}
	err := checkBeatShape(plan)
	if err == nil || !strings.Contains(err.Error(), "thin") {
		t.Fatalf("want the short beat named, got %v", err)
	}
	if !strings.Contains(err.Error(), "floor 10") {
		t.Errorf("a middle beat must be held to the develop floor, not the relaxed one:\n%s", err)
	}
}

// The beat count has to come from the runtime, because every snippet prompt
// also calibrates a beat at about forty words. A fixed three-beat floor is
// therefore a 120-word floor, which a short clip's ceiling is below — and a
// plan cannot satisfy both. This is the arithmetic that made 20-second clips
// fail in the field, so it is pinned rather than described.
func TestBeatBoundsFitTheWordBudget(t *testing.T) {
	for _, tc := range []struct{ sec, pace int }{
		{10, 150}, {10, 175}, {20, 150}, {20, 175},
		{45, 150}, {45, 175}, {75, 150}, {120, 175}, {180, 175},
	} {
		want, minWords, maxWords := wordBudget(tc.sec, tc.pace)
		minBeats, maxBeats, suggest, perBeat := beatBounds(want, 0, 0)

		// The suggested shape must fit inside the budget it was derived from,
		// or the prompt is asking for something it will then reject.
		if got := suggest * perBeat; got > maxWords {
			t.Errorf("%ds@%dwpm: %d beats x %d words = %d, over the %d ceiling",
				tc.sec, tc.pace, suggest, perBeat, got, maxWords)
		}
		// And the smallest legal plan must be able to reach the floor.
		if got := maxBeats * maxWordsPerBeat; got < minWords {
			t.Errorf("%ds@%dwpm: %d beats x %d words = %d, under the %d floor",
				tc.sec, tc.pace, maxBeats, maxWordsPerBeat, got, minWords)
		}
		// Every beat the budget suggests must clear the per-beat minimum, or
		// the plan is rejected for beats it was told to write.
		if perBeat < minWordsPerBeat {
			t.Errorf("%ds@%dwpm: budget affords %d words a beat, below the %d minimum",
				tc.sec, tc.pace, perBeat, minWordsPerBeat)
		}
		if minBeats < floorSnippetBeats || maxBeats > maxSnippetBeats || minBeats > maxBeats {
			t.Errorf("%ds@%dwpm: beat range %d-%d is not sane", tc.sec, tc.pace, minBeats, maxBeats)
		}
	}
}

// The same failure at the other end of the range, which went unnoticed until a
// template wanted three minutes: a 180-second clip was told to write seven beats
// of 75 words against a 60-word per-beat maximum. No plan satisfies both.
func TestLongClipsAreNotAskedForImpossiblePlans(t *testing.T) {
	for _, sec := range []int{90, 120, 150, 180} {
		want, minWords, _ := wordBudget(sec, 175)
		for _, ceiling := range []int{0, 10, 12} {
			_, maxBeats, _, perBeat := beatBounds(want, ceiling, 0)
			// The advice must be something the validator will accept.
			if perBeat > maxWordsPerBeat {
				t.Errorf("%ds (ceiling %d): advised %d words a beat, over the %d maximum",
					sec, ceiling, perBeat, maxWordsPerBeat)
			}
			// And the beat range must be able to fund the floor at all.
			if got := maxBeats * maxWordsPerBeat; got < minWords {
				t.Errorf("%ds (ceiling %d): %d beats can hold %d words, under the %d floor",
					sec, ceiling, maxBeats, got, minWords)
			}
		}
	}
}

// A raised ceiling has to actually raise the range, or the templates that set
// it are still capped at seven and the field does nothing.
func TestRaisedCeilingWidensTheBeatRange(t *testing.T) {
	want, _, _ := wordBudget(120, 175)
	_, defaultMax, _, _ := beatBounds(want, 0, 0)
	_, raisedMax, _, raisedPerBeat := beatBounds(want, 12, 0)
	if raisedMax <= defaultMax {
		t.Errorf("ceiling 12 gave %d beats, no more than the default %d", raisedMax, defaultMax)
	}
	if raisedPerBeat >= maxWordsPerBeat {
		t.Errorf("more beats should mean fewer words each; got %d", raisedPerBeat)
	}
}

// The regression that prompted all of this: a 20-second clip at the snippets
// course's pace was told to write three beats against an 89-word ceiling.
func TestShortClipsAreNotAskedForImpossiblePlans(t *testing.T) {
	want, _, maxWords := wordBudget(20, 175)
	_, _, suggest, perBeat := beatBounds(want, 0, 0)
	if suggest*perBeat > maxWords {
		t.Fatalf("a 20s clip is still asked for %d beats x %d words against a %d ceiling",
			suggest, perBeat, maxWords)
	}
	if suggest > 2 {
		t.Errorf("a 20s clip suggests %d beats; at ~%d words each that is more than the runtime holds", suggest, perBeat)
	}
}

// 10 seconds has to be a runtime a caller can actually ask for.
func TestTenSecondSnippetIsAccepted(t *testing.T) {
	spec := SnippetSpec{Prompt: "Why tabs beat spaces", Template: "illustration", TargetSec: 10}
	if err := spec.Validate(); err != nil {
		t.Fatalf("a 10s snippet should be allowed: %v", err)
	}
	if got := spec.ResolvedTargetSec(); got != 10 {
		t.Errorf("ResolvedTargetSec = %d, want 10", got)
	}
	// …but not for a template whose own floor is higher.
	story := SnippetSpec{Prompt: "Why tabs beat spaces", Template: "story", TargetSec: 10}
	if err := story.Validate(); err == nil {
		t.Error("the story template should still refuse 10 seconds — it needs eight beats")
	}
}

// The walkthrough carries a keystroke schedule: one entry per character of the
// first step's code, starting no earlier than typeAtMs.
//
// A length mismatch is the failure worth guarding. The renderer silently falls
// back to its own estimate when the counts disagree, so the typing still looks
// fine — while the click track, which is generated from these numbers, plays
// against a completely different rhythm.
func TestWalkthroughCarriesKeystrokeSchedule(t *testing.T) {
	scenes, err := vscodeScenes(sceneInput(t, vscodePlan(), 12000))
	if err != nil {
		t.Fatal(err)
	}
	var walk *Scene
	for i := range scenes {
		if scenes[i].Type == SceneWalkthrough {
			walk = &scenes[i]
		}
	}
	if walk == nil {
		t.Fatal("no walkthrough scene")
	}
	keys, ok := walk.Props["keystrokes"].([]int)
	if !ok {
		t.Fatalf("walkthrough has no keystrokes: %#v", walk.Props["keystrokes"])
	}
	steps, _ := walk.Props["steps"].([]map[string]any)
	if len(steps) == 0 {
		t.Fatal("walkthrough has no steps")
	}
	code, _ := steps[0]["code"].(string)
	if len(keys) != len([]rune(code)) {
		t.Fatalf("keystrokes has %d entries for %d characters of code", len(keys), len([]rune(code)))
	}
	typeAt, _ := walk.Props["typeAtMs"].(int)
	if keys[0] < typeAt {
		t.Errorf("first keystroke at %d is before typeAtMs %d", keys[0], typeAt)
	}
	for i := 1; i < len(keys); i++ {
		if keys[i] < keys[i-1] {
			t.Fatalf("keystrokes go backwards at %d: %d then %d", i, keys[i-1], keys[i])
		}
	}
	if keys[len(keys)-1] <= keys[0] {
		t.Error("every keystroke lands at the same moment — the schedule has no span")
	}
	// Typing occupies the front of the window, not all of it: the reader has to
	// get a moment with the finished file before the next step takes over.
	if next, ok := steps[1]["atMs"].(int); ok {
		if keys[len(keys)-1] >= next {
			t.Errorf("typing finishes at %d, at or after the next step at %d", keys[len(keys)-1], next)
		}
	}
}
