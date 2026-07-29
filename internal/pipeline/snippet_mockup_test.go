package pipeline

import (
	"strings"
	"testing"
)

const mkNarration = "One field is all you need here, because every extra one costs you signups."

func mockupPlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "mockup",
		Title:    "A signup page, one block at a time",
		Mockup: &MockupSpec{
			Device: "browser",
			Screen: "Signup",
			Blocks: []MockupBlock{
				{Kind: "header", Label: "Nav bar", Text: "Notely", Note: "A logo and two links."},
				{Kind: "hero", Label: "Hero", Text: "Notes that find themselves", Note: "One promise, above the fold."},
				{Kind: "input", Label: "Email field", Text: "you@work.com", Note: "One field only."},
				{Kind: "button", Label: "Signup button", Text: "Start free", Note: "The verb says what happens next."},
			},
		},
		Beats: []SnippetBeat{
			{ID: "nav", Heading: "The bar at the top", Narration: mkNarration, Mockup: &MockupBeat{At: 0}},
			{ID: "hero", Heading: "One big promise", Narration: mkNarration, Mockup: &MockupBeat{At: 1}},
			{ID: "field", Heading: "Asking for one thing", Narration: mkNarration, Mockup: &MockupBeat{At: 2}},
			{ID: "cta", Heading: "The button", Narration: mkNarration, Mockup: &MockupBeat{At: 3}},
			{ID: "whole", Heading: "The finished page", Narration: mkNarration, Mockup: &MockupBeat{Whole: true}},
		},
	}
}

func TestMockupPlanAccepted(t *testing.T) {
	if err := validateMockupPlan(mockupPlan()); err != nil {
		t.Fatalf("a well-formed mockup was rejected: %v", err)
	}
}

// A page is read and built downward. A clip that jumps back up is describing a
// layout, and the error names the template that draws layouts properly.
func TestMockupOnlyBuildsDownward(t *testing.T) {
	p := mockupPlan()
	p.Beats[1].Mockup = &MockupBeat{At: 2}
	p.Beats[2].Mockup = &MockupBeat{At: 1}
	err := validateMockupPlan(p)
	if err == nil {
		t.Fatal("a page built out of order was accepted")
	}
	if !strings.Contains(err.Error(), "built downward") {
		t.Errorf("the error should name the rule; got: %v", err)
	}
	if !strings.Contains(err.Error(), "whiteboard") {
		t.Errorf("the error should point at the template that fits; got: %v", err)
	}
}

func TestMockupNarratesEveryBlock(t *testing.T) {
	p := mockupPlan()
	p.Beats = append(p.Beats[:3], p.Beats[4]) // the button is never added
	err := validateMockupPlan(p)
	if err == nil {
		t.Fatal("a block with no beat was accepted")
	}
	if !strings.Contains(err.Error(), "never narrated") {
		t.Errorf("the error should name the unexplained block; got: %v", err)
	}
}

func TestMockupMustEndOnTheFinishedScreen(t *testing.T) {
	p := mockupPlan()
	p.Beats = p.Beats[:4]
	err := validateMockupPlan(p)
	if err == nil {
		t.Fatal("a clip that never shows the finished page was accepted")
	}
	if !strings.Contains(err.Error(), "whole") {
		t.Errorf("the error should say how to close; got: %v", err)
	}
}

// A page has one top and one bottom. Two of either is usually a model listing
// sections rather than stacking them.
func TestMockupRejectsTwoHeaders(t *testing.T) {
	p := mockupPlan()
	p.Mockup.Blocks[2].Kind = "header"
	err := validateMockupPlan(p)
	if err == nil {
		t.Fatal("a page with two headers was accepted")
	}
	if !strings.Contains(err.Error(), "A page has one") {
		t.Errorf("the error should say why; got: %v", err)
	}
}

func TestMockupRejectsTooFewBlocks(t *testing.T) {
	p := mockupPlan()
	p.Mockup.Blocks = p.Mockup.Blocks[:2]
	p.Beats = append(p.Beats[:2], p.Beats[4])
	err := validateMockupPlan(p)
	if err == nil {
		t.Fatal("a two-block page was accepted")
	}
	if !strings.Contains(err.Error(), "component") {
		t.Errorf("the error should say what two blocks actually is; got: %v", err)
	}
}

func TestMockupRejectsDuplicateLabels(t *testing.T) {
	p := mockupPlan()
	p.Mockup.Blocks[2].Label = "Nav bar"
	if err := validateMockupPlan(p); err == nil {
		t.Error("two layers with the same name were accepted")
	}
}

func TestMockupNormalizeDefaultsAndClamps(t *testing.T) {
	p := mockupPlan()
	p.Mockup.Device = "tablet" // not in the vocabulary
	p.Mockup.Blocks[1].Kind = "banner"
	p.Mockup.Blocks[1].Label = ""
	p.Mockup.Blocks[2].Text = "a placeholder that runs on far past what a field can show"
	normalizeMockupPlan(p)

	if p.Mockup.Device != "browser" {
		t.Errorf("an invented device should degrade to browser, got %q", p.Mockup.Device)
	}
	if p.Mockup.Blocks[1].Kind != "text" {
		t.Errorf("an invented kind should degrade to text, got %q", p.Mockup.Blocks[1].Kind)
	}
	// The layer list is the one place a block must be nameable.
	if p.Mockup.Blocks[1].Label == "" {
		t.Error("a block with no label should get one rather than a blank row")
	}
	if w := len(strings.Fields(p.Mockup.Blocks[2].Text)); w > maxMockupTextWords {
		t.Errorf("text still %d words, want at most %d", w, maxMockupTextWords)
	}
	if err := validateMockupPlan(p); err != nil {
		t.Fatalf("the normalized plan should validate: %v", err)
	}
}

func TestMockupScenesCoverTheWholeClip(t *testing.T) {
	p := mockupPlan()
	scenes, err := mockupScenes(sceneInput(t, p, 6000))
	if err != nil {
		t.Fatalf("laying out the mockup failed: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneMockup {
		t.Fatalf("want one %s scene, got %v", SceneMockup, scenes)
	}
	if scenes[0].Props["device"] != "browser" {
		t.Errorf("device should reach the renderer, got %v", scenes[0].Props["device"])
	}
	steps, ok := scenes[0].Props["steps"].([]map[string]any)
	if !ok || len(steps) != len(p.Beats) {
		t.Fatalf("want one step per beat, got %v", scenes[0].Props["steps"])
	}
	if _, isWhole := steps[len(steps)-1]["whole"]; !isWhole {
		t.Error("the last step should show the finished screen")
	}
}
