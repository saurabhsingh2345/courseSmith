package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

func TestSectionsFromOutline(t *testing.T) {
	body := "intro text\n## First Idea\ncontent a\n```python\nx=1\n```\n## Your First Line of Python!\ncontent b\n"
	sections := sectionsFromOutline(body)
	if len(sections) != 2 {
		t.Fatalf("sections = %v", sections)
	}
	if !strings.Contains(sections[sectionKey("first-idea")], "content a") {
		t.Errorf("first-idea = %q", sections[sectionKey("first-idea")])
	}
	if !strings.Contains(sections[sectionKey("your-first-line-of-python")], "content b") {
		t.Errorf("slug with punctuation missing: %v", sections)
	}
}

// The script's section ids come from the LLM, which slugifies punctuation
// differently than the outline heading does. Both spellings must resolve to
// the same section, or it loses its code blocks and its title.
func TestSectionKeyBridgesSlugDisagreement(t *testing.T) {
	body := "## What's next\ncontent\n"
	sections := sectionsFromOutline(body)
	headings := headingsFromOutline(body)

	for _, id := range []string{"whats-next", "what-s-next", "Whats Next"} {
		if !strings.Contains(sections[sectionKey(id)], "content") {
			t.Errorf("section id %q did not resolve: %v", id, sections)
		}
		if got := headings[sectionKey(id)]; got != "What's next" {
			t.Errorf("heading for %q = %q, want %q", id, got, "What's next")
		}
	}
}

func TestSectionFileNameFitsSidebar(t *testing.T) {
	tests := map[string]string{
		"names-without-quotes-text-with-quotes": "names_quotes_text.py",
		"your-first-line-of-python":             "first_line_python.py",
		"installing-python":                     "installing_python.py",
		// Every word is a stop word: fall back rather than return ".py".
		"a-the-of": "a_the_of.py",
	}
	for id, want := range tests {
		if got := sectionFileName(id); got != want {
			t.Errorf("sectionFileName(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestCueTimestamp(t *testing.T) {
	a := &Alignment{Words: wordSeq(1000, "a", "b", "c", "d", "e")}
	span := SectionSpan{WordStart: 1, WordEnd: 4, StartMs: 1250}

	tests := []struct {
		atWord int
		want   int
	}{
		{0, 1250},  // first word of the section
		{2, 1750},  // third word
		{99, 1750}, // past the end clamps to the last word
		{-1, 1250}, // negative clamps to the first
	}
	for _, tt := range tests {
		if got := cueTimestamp(a, span, tt.atWord); got != tt.want {
			t.Errorf("cueTimestamp(%d) = %d, want %d", tt.atWord, got, tt.want)
		}
	}

	empty := SectionSpan{WordStart: 3, WordEnd: 3, StartMs: 700}
	if got := cueTimestamp(a, empty, 0); got != 700 {
		t.Errorf("empty span = %d, want span start", got)
	}
}

func TestFindPhrase(t *testing.T) {
	a := &Alignment{Words: wordSeq(0, "Python", "reads", "your", "code,", "line", "by", "line.")}
	span := SectionSpan{WordStart: 0, WordEnd: 7}

	if got := findPhrase(a, span, "your code"); got != wordSeq(0, "a", "b", "c")[2].StartMs {
		t.Errorf("findPhrase = %d, want start of word 2 (%d)", got, 500)
	}
	if got := findPhrase(a, span, "LINE BY"); got < 0 {
		t.Error("case-insensitive phrase not found")
	}
	if got := findPhrase(a, span, "not present"); got != -1 {
		t.Errorf("phantom phrase at %d", got)
	}
}

// sceneGraphFixture builds a two-section lesson with a diagram cue, a demo
// cue, a code block, and a callout.
func sceneGraphFixture(t *testing.T) (*project.Course, *project.Lesson, config.Config, *Script, *Alignment, *DemoManifest, *VerificationReport) {
	t.Helper()
	course, lesson := testCourse(t)
	md := `---
title: Test Lesson
diagrams:
  - id: memory-model
    prompt: "3 variables"
outcomes:
  - Understand what Python is
  - Run your first line
callouts:
  - section: second-idea
    shape: circle
    x: 0.5
    y: 0.3
    label: watch this
    at: "try it live"
---

## First Idea
- a point

[DIAGRAM: memory-model]

## Second Idea

` + "```python\nprint(\"hi\")\n```" + `

[DEMO: run a greeting in the REPL]
`
	if err := os.WriteFile(lesson.SourcePath(), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	var err error
	lesson, err = project.LoadLesson(lesson.Dir)
	if err != nil {
		t.Fatal(err)
	}

	script := &Script{
		Title: "Test Lesson",
		Sections: []Section{
			{ID: "first-idea", Narration: "Python reads code line by line.", DurationEstSec: 5,
				Cues: []Cue{{Type: CueDiagram, Ref: "memory-model", AtWord: 3}}},
			{ID: "second-idea", Narration: "Now let us try it live.", DurationEstSec: 5,
				Cues: []Cue{{Type: CueDemo, Ref: "run a greeting in the REPL", AtWord: 2}}},
		},
	}
	words := append(
		wordSeq(0, "Python", "reads", "code", "line", "by", "line."), // 0..1450
		wordSeq(2000, "Now", "let", "us", "try", "it", "live.")...,   // 2000..3450
	)
	alignment := &Alignment{
		Source: AlignSourceWhisperX,
		Words:  words,
		Sections: []SectionSpan{
			{ID: "first-idea", StartMs: 0, EndMs: 1450, WordStart: 0, WordEnd: 6},
			{ID: "second-idea", StartMs: 2000, EndMs: 3450, WordStart: 6, WordEnd: 12},
		},
	}
	demos := &DemoManifest{Demos: []DemoEntry{
		{ID: "demo-1", Description: "run a greeting in the REPL", Path: "demos/demo-1.mp4", DurationMs: 8000},
	}}
	verification := &VerificationReport{Blocks: []VerifiedBlock{{
		CodeBlock: CodeBlock{Source: "lesson.md", Code: `print("hi")`},
		Hash:      project.HashBytes([]byte(`print("hi")`)),
		Stdout:    "hi\n",
	}}}
	cfg := configFor(course, lesson)
	return course, lesson, cfg, script, alignment, demos, verification
}

func TestBuildSceneGraph(t *testing.T) {
	course, lesson, cfg, script, alignment, demos, verification := sceneGraphFixture(t)

	graph, err := buildSceneGraph(course, lesson, cfg, script, alignment, demos, verification, nil, nil, 4000)
	if err != nil {
		t.Fatal(err)
	}

	if graph.Theme.Primary != "#306998" || graph.Theme.CourseName != "Test Course" {
		t.Errorf("theme = %+v", graph.Theme)
	}
	if graph.AudioFile != VoiceoverFileName || graph.DurationMs != 4250 {
		t.Errorf("audio %q duration %d (want 4250: last section end 3450 + %d tail... adjusted to max)", graph.AudioFile, graph.DurationMs, videoTailMs)
	}
	// On-screen captions are opt-in; the default scene graph carries none.
	if len(graph.Captions) != 0 {
		t.Errorf("captions embedded by default: %d words", len(graph.Captions))
	}
	cfgOn := cfg
	cfgOn.Style.Captions = "on"
	if withCaptions, err := buildSceneGraph(course, lesson, cfgOn, script, alignment, demos, verification, nil, nil, 4000); err != nil {
		t.Fatal(err)
	} else if len(withCaptions.Captions) != 12 {
		t.Errorf("captions with style.captions=on: %d words, want 12", len(withCaptions.Captions))
	}

	// Expected scenes:
	// 0: intro title card (section 1 base) 0 → diagram cue at word 3 (750ms)
	// 1: diagram 750 → 2000 (next section start)
	// 2: code scene (section 2 base) 2000 → demo cue at word 2 (2500ms)
	// 3: terminal 2500 → end
	if len(graph.Scenes) != 4 {
		t.Fatalf("scenes = %d: %+v", len(graph.Scenes), graph.Scenes)
	}

	intro := graph.Scenes[0]
	if intro.Type != SceneTitle || intro.StartMs != 0 || intro.EndMs != 750 {
		t.Errorf("intro = %+v", intro)
	}
	if intro.Props["intro"] != true || intro.Props["heading"] != "Test Lesson" {
		t.Errorf("intro props = %+v", intro.Props)
	}
	outcomes, _ := intro.Props["outcomes"].([]string)
	if len(outcomes) != 2 {
		t.Errorf("outcomes = %+v", intro.Props["outcomes"])
	}

	diagram := graph.Scenes[1]
	if diagram.Type != SceneDiagram || diagram.StartMs != 750 || diagram.EndMs != 2000 {
		t.Errorf("diagram = %+v", diagram)
	}
	if diagram.Props["src"] != "diagrams/memory-model.svg" {
		t.Errorf("diagram src = %v", diagram.Props["src"])
	}

	code := graph.Scenes[2]
	if code.Type != SceneCode || code.StartMs != 2000 || code.EndMs != 2500 {
		t.Errorf("code = %+v", code)
	}
	if code.Props["code"] != `print("hi")` || code.Props["output"] != "hi\n" {
		t.Errorf("code props = %+v", code.Props)
	}

	term := graph.Scenes[3]
	if term.Type != SceneTerminal || term.StartMs != 2500 {
		t.Errorf("terminal = %+v", term)
	}
	if term.Props["src"] != "demos/demo-1.mp4" || term.EndMs != graph.DurationMs {
		t.Errorf("terminal props/end = %+v", term)
	}

	// The callout lands on the scene covering "try it live" (word 9 @ 2750ms
	// → the terminal scene, which starts at 2500).
	if len(term.Callouts) != 1 {
		t.Fatalf("terminal callouts = %+v (code scene: %+v)", term.Callouts, code.Callouts)
	}
	callout := term.Callouts[0]
	if callout.AtMs != 2750 || callout.Shape != "circle" || callout.Label != "watch this" || callout.DurMs != defaultCalloutDurMs {
		t.Errorf("callout = %+v", callout)
	}
}

func TestBuildSceneGraphSectionMismatch(t *testing.T) {
	course, lesson, cfg, script, alignment, demos, verification := sceneGraphFixture(t)
	alignment.Sections = alignment.Sections[:1]
	_, err := buildSceneGraph(course, lesson, cfg, script, alignment, demos, verification, nil, nil, 4000)
	if err == nil || !strings.Contains(err.Error(), "re-run the align stage") {
		t.Errorf("error = %v", err)
	}
}

func TestScenegraphStageWritesGraph(t *testing.T) {
	course, lesson, _, script, alignment, demos, verification := sceneGraphFixture(t)

	// Persist the inputs the stage reads from disk.
	if err := os.MkdirAll(filepath.Join(lesson.GeneratedDir(), DemosDirName), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(lesson.GeneratedDir(), ScriptFileName), script); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(lesson.GeneratedDir(), AlignmentFileName), alignment); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(lesson.GeneratedDir(), DemosDirName, DemoManifestFileName), demos); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(lesson.GeneratedDir(), VerificationFileName), verification); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lesson.GeneratedDir(), VoiceoverFileName), makeWAV(4), 0o644); err != nil {
		t.Fatal(err)
	}

	env, out := runEnv(t, &fakeRouter{})
	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageScenegraph}); err != nil {
		t.Fatal(err)
	}
	graph, err := LoadSceneGraph(lesson)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Scenes) != 4 || graph.DurationMs < 4000 {
		t.Errorf("graph = %d scenes, %dms", len(graph.Scenes), graph.DurationMs)
	}
	if !strings.Contains(out.String(), "1 title, 0 points, 1 code, 0 walkthrough, 1 diagram, 1 terminal") {
		t.Errorf("output:\n%s", out.String())
	}
}

func TestScenegraphStageRequiresInputs(t *testing.T) {
	course, lesson := testCourse(t)
	env, _ := runEnv(t, &fakeRouter{})
	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageScenegraph})
	if err == nil || !strings.Contains(err.Error(), "script stage must run first") {
		t.Errorf("error = %v", err)
	}
}

// TestBuildSceneGraphWalkthroughAndPoints locks the W2/W4 contracts: a
// multi-block section becomes a VS Code walkthrough (one step per block),
// and a storyboard turns a heading-only section into a points scene.
func TestBuildSceneGraphWalkthroughAndPoints(t *testing.T) {
	course, lesson, cfg, script, alignment, demos, verification := sceneGraphFixture(t)

	// Rewrite the lesson so second-idea has TWO code blocks and no demo.
	md := `---
title: Test Lesson
---

## First Idea
- a point

## Second Idea

` + "```python\nx = 1\n```\n\nmore prose\n\n```python\nx = 1\nprint(x)\n```" + `
`
	if err := os.WriteFile(lesson.SourcePath(), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	var err error
	lesson, err = project.LoadLesson(lesson.Dir)
	if err != nil {
		t.Fatal(err)
	}
	script.Sections[0].Cues = nil
	script.Sections[1].Cues = nil

	storyboard := &Storyboard{Sections: []StoryboardSection{
		{ID: "first-idea", Points: []StoryPoint{{Text: "Line by line", Icon: "list", AtWord: 3}}},
	}}

	graph, err := buildSceneGraph(course, lesson, cfg, script, alignment, demos, verification, nil, storyboard, 4000)
	if err != nil {
		t.Fatal(err)
	}
	if len(graph.Scenes) != 2 {
		t.Fatalf("scenes = %d, want 2 (intro title + walkthrough)", len(graph.Scenes))
	}
	// Section 0 is the intro title card; the storyboard only upgrades
	// non-intro heading sections, so scene 0 stays a title.
	if graph.Scenes[0].Type != SceneTitle {
		t.Errorf("scene 0 = %s, want title (intro)", graph.Scenes[0].Type)
	}
	wt := graph.Scenes[1]
	if wt.Type != SceneWalkthrough {
		t.Fatalf("scene 1 = %s, want walkthrough", wt.Type)
	}
	steps, ok := wt.Props["steps"].([]map[string]any)
	if !ok || len(steps) != 2 {
		t.Fatalf("walkthrough steps = %#v, want 2", wt.Props["steps"])
	}
	if steps[0]["atMs"].(int) >= steps[1]["atMs"].(int) {
		t.Errorf("step times not increasing: %v then %v", steps[0]["atMs"], steps[1]["atMs"])
	}
	if wt.Props["file"] != "second_idea.py" {
		t.Errorf("file = %v, want second_idea.py", wt.Props["file"])
	}

	// Now prove the points upgrade on a heading-only non-intro section:
	// swap the sections so first-idea (storyboarded, no code) is second.
	script2 := &Script{Title: script.Title, Sections: []Section{script.Sections[1], script.Sections[0]}}
	alignment2 := &Alignment{
		Source: alignment.Source,
		Words:  alignment.Words,
		Sections: []SectionSpan{
			{ID: "second-idea", StartMs: 0, EndMs: 1450, WordStart: 0, WordEnd: 6},
			{ID: "first-idea", StartMs: 2000, EndMs: 3450, WordStart: 6, WordEnd: 12},
		},
	}
	graph2, err := buildSceneGraph(course, lesson, cfg, script2, alignment2, demos, verification, nil, storyboard, 4000)
	if err != nil {
		t.Fatal(err)
	}
	pts := graph2.Scenes[len(graph2.Scenes)-1]
	if pts.Type != ScenePoints {
		t.Fatalf("last scene = %s, want points", pts.Type)
	}
	items, ok := pts.Props["items"].([]map[string]any)
	if !ok || len(items) != 1 || items[0]["text"] != "Line by line" || items[0]["icon"] != "list" {
		t.Fatalf("points items = %#v", pts.Props["items"])
	}
	if at := items[0]["atMs"].(int); at < 2000 || at > 3450 {
		t.Errorf("point atMs = %d, want inside the section span", at)
	}
}
