package pipeline

import (
	"strings"
	"testing"
)

const cvNarration = "Nothing at all runs until somebody actually fills in that form and presses send."

func canvasPlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "canvas",
		Title:    "From form to spreadsheet to Slack",
		Canvas: &CanvasSpec{
			Payload: "New signup",
			Nodes: []CanvasNode{
				{App: "Typeform", Title: "Someone submits the form", Kind: "trigger", Icon: "zap", Note: "The chain waits here."},
				{App: "Make", Title: "Check the plan field", Kind: "filter", Icon: "filter", Note: "Free signups stop here."},
				{App: "Sheets", Title: "Append a row", Kind: "action", Icon: "database", Note: "One row per signup."},
				{App: "Slack", Title: "Post to sales", Kind: "output", Icon: "message", Note: "The team sees it immediately."},
			},
		},
		Beats: []SnippetBeat{
			{ID: "trigger", Heading: "Something starts it", Narration: cvNarration, Canvas: &CanvasBeat{At: 0}},
			{ID: "filter", Heading: "What carries on", Narration: cvNarration, Canvas: &CanvasBeat{At: 1}},
			{ID: "row", Heading: "Writing it down", Narration: cvNarration, Canvas: &CanvasBeat{At: 2}},
			{ID: "notify", Heading: "Telling the team", Narration: cvNarration, Canvas: &CanvasBeat{At: 3}},
			{ID: "run", Heading: "Watch it fire", Narration: cvNarration, Canvas: &CanvasBeat{Run: true}},
		},
	}
}

func TestCanvasPlanAccepted(t *testing.T) {
	if err := validateCanvasPlan(canvasPlan()); err != nil {
		t.Fatalf("a well-formed canvas was rejected: %v", err)
	}
}

// The claim an automation makes that a picture of boxes does not: something
// starts it, and only one thing does.
func TestCanvasMustBeginWithATrigger(t *testing.T) {
	p := canvasPlan()
	p.Canvas.Nodes[0].Kind = "action"
	err := validateCanvasPlan(p)
	if err == nil {
		t.Fatal("a canvas that starts with an action was accepted")
	}
	if !strings.Contains(err.Error(), "trigger") {
		t.Errorf("the error should name what is missing; got: %v", err)
	}
}

func TestCanvasRejectsASecondTrigger(t *testing.T) {
	p := canvasPlan()
	p.Canvas.Nodes[2].Kind = "trigger"
	err := validateCanvasPlan(p)
	if err == nil {
		t.Fatal("a canvas with two triggers was accepted")
	}
	if !strings.Contains(err.Error(), "second trigger") {
		t.Errorf("the error should say which rule broke; got: %v", err)
	}
}

// An automation read backwards is not an automation. The error names the
// template that would tell that story properly, because the fix is to change
// template rather than to reorder the beats.
func TestCanvasOnlyMovesForward(t *testing.T) {
	p := canvasPlan()
	p.Beats[1].Canvas = &CanvasBeat{At: 2}
	p.Beats[2].Canvas = &CanvasBeat{At: 1}
	err := validateCanvasPlan(p)
	if err == nil {
		t.Fatal("a canvas that walks backwards was accepted")
	}
	if !strings.Contains(err.Error(), "only moves forward") {
		t.Errorf("the error should name the rule; got: %v", err)
	}
	if !strings.Contains(err.Error(), "flow") {
		t.Errorf("the error should point at the template that fits; got: %v", err)
	}
}

func TestCanvasNarratesEveryCard(t *testing.T) {
	p := canvasPlan()
	p.Beats = append(p.Beats[:3], p.Beats[4]) // "Post to sales" never reached
	err := validateCanvasPlan(p)
	if err == nil {
		t.Fatal("a card with no beat was accepted")
	}
	if !strings.Contains(err.Error(), "never narrated") {
		t.Errorf("the error should name the unexplained card; got: %v", err)
	}
}

// The payoff beat. A workflow nobody watched fire is a diagram of a workflow.
func TestCanvasMustEndByRunning(t *testing.T) {
	p := canvasPlan()
	p.Beats = p.Beats[:4]
	err := validateCanvasPlan(p)
	if err == nil {
		t.Fatal("a canvas that is never run was accepted")
	}
	if !strings.Contains(err.Error(), "run") {
		t.Errorf("the error should say to add the run beat; got: %v", err)
	}
}

func TestCanvasRunMustBeLast(t *testing.T) {
	p := canvasPlan()
	p.Beats[4].Canvas = &CanvasBeat{At: 3}
	p.Beats[3].Canvas = &CanvasBeat{Run: true}
	if err := validateCanvasPlan(p); err == nil {
		t.Error("a clip that runs the automation and then goes back to a card was accepted")
	}
}

func TestCanvasRejectsTooFewCards(t *testing.T) {
	p := canvasPlan()
	p.Canvas.Nodes = p.Canvas.Nodes[:2]
	p.Beats = append(p.Beats[:2], p.Beats[4])
	err := validateCanvasPlan(p)
	if err == nil {
		t.Fatal("a two-card automation was accepted")
	}
	if !strings.Contains(err.Error(), "workflow") {
		t.Errorf("the error should say why two is not enough; got: %v", err)
	}
}

// Clamping a long title is arithmetic, not a correction round.
func TestCanvasNormalizeClampsAndDefaults(t *testing.T) {
	p := canvasPlan()
	p.Canvas.Payload = ""
	p.Canvas.Nodes[1].Title = "check the plan field before doing anything else at all"
	p.Canvas.Nodes[1].Kind = "sieve" // not in the vocabulary
	p.Canvas.Nodes[1].Icon = "wobble"
	normalizeCanvasPlan(p)

	if p.Canvas.Payload == "" {
		t.Error("an empty payload should get a default — the token has to say something")
	}
	if w := len(strings.Fields(p.Canvas.Nodes[1].Title)); w > maxCanvasTitleWords {
		t.Errorf("title still %d words, want at most %d", w, maxCanvasTitleWords)
	}
	if p.Canvas.Nodes[1].Kind != "action" {
		t.Errorf("an invented kind should degrade to action, got %q", p.Canvas.Nodes[1].Kind)
	}
	if p.Canvas.Nodes[1].Icon != canvasNodeKinds["action"] {
		t.Errorf("an invented icon should fall back to the kind's, got %q", p.Canvas.Nodes[1].Icon)
	}
	if err := validateCanvasPlan(p); err != nil {
		t.Fatalf("the normalized plan should validate: %v", err)
	}
}

// A model that describes a trigger and files it as an action has understood the
// workflow and mislabelled the card. That is a repair, not a round.
func TestCanvasNormalizePromotesTheFirstCard(t *testing.T) {
	p := canvasPlan()
	p.Canvas.Nodes[0].Kind = "action"
	p.Canvas.Nodes[0].Icon = ""
	normalizeCanvasPlan(p)
	if p.Canvas.Nodes[0].Kind != "trigger" {
		t.Fatalf("the first card should have been promoted to a trigger, got %q", p.Canvas.Nodes[0].Kind)
	}
	if err := validateCanvasPlan(p); err != nil {
		t.Errorf("the repaired plan should validate: %v", err)
	}
}

// But only when there is no trigger anywhere. A model that put the trigger
// second has said something about the workflow, and quietly adding a second one
// would ship a different automation than the one it described.
func TestCanvasNormalizeLeavesAMisplacedTriggerToValidation(t *testing.T) {
	p := canvasPlan()
	p.Canvas.Nodes[0].Kind = "action"
	p.Canvas.Nodes[1].Kind = "trigger"
	normalizeCanvasPlan(p)
	if p.Canvas.Nodes[0].Kind == "trigger" {
		t.Fatal("normalizing invented a second trigger instead of leaving the claim alone")
	}
	if err := validateCanvasPlan(p); err == nil {
		t.Error("a canvas whose trigger is not first should be rejected")
	}
}

func TestCanvasScenesCoverTheWholeClip(t *testing.T) {
	p := canvasPlan()
	scenes, err := canvasScenes(sceneInput(t, p, 6000))
	if err != nil {
		t.Fatalf("laying out the canvas failed: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("the canvas is one continuous scene, got %d", len(scenes))
	}
	if scenes[0].Type != SceneCanvas {
		t.Errorf("scene type is %q, want %q", scenes[0].Type, SceneCanvas)
	}
	steps, ok := scenes[0].Props["steps"].([]map[string]any)
	if !ok || len(steps) != len(p.Beats) {
		t.Fatalf("want one step per beat, got %v", scenes[0].Props["steps"])
	}
	if _, isRun := steps[len(steps)-1]["run"]; !isRun {
		t.Error("the last step should be the run")
	}
}
