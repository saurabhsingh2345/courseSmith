package pipeline

import (
	"strings"
	"testing"
)

const plNarration = "Notice how the second ask names the exact thing that came back wrong."

func promptLoopPlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "promptloop",
		Title:    "Prompting your way to a landing page",
		Loop:     &PromptLoopSpec{Goal: "A landing page with a working signup form"},
		Beats: []SnippetBeat{
			{ID: "ask", Heading: "The first ask", Narration: plNarration, Loop: &PromptLoopBeat{
				Turn: "you", Text: "Build me a landing page with a signup form.",
			}},
			{ID: "try", Heading: "What came back", Narration: plNarration, Loop: &PromptLoopBeat{
				Turn: "ai", Text: "Built a hero and a form.", Status: "partial",
				Changes: []string{"Hero in place", "Form has no validation"},
			}},
			{ID: "sharper", Heading: "Naming the gap", Narration: plNarration, Loop: &PromptLoopBeat{
				Turn: "you", Text: "The form does nothing. Validate the email on submit.",
			}},
			{ID: "works", Heading: "Closer", Narration: plNarration, Loop: &PromptLoopBeat{
				Turn: "ai", Text: "Added validation and a confirmation.", Status: "ok",
				Changes: []string{"Email checked before submit"},
			}},
		},
	}
}

func TestPromptLoopPlanAccepted(t *testing.T) {
	if err := validatePromptLoopPlan(promptLoopPlan()); err != nil {
		t.Fatalf("a well-formed loop was rejected: %v", err)
	}
}

func TestPromptLoopNeedsAGoal(t *testing.T) {
	p := promptLoopPlan()
	p.Loop = nil
	err := validatePromptLoopPlan(p)
	if err == nil {
		t.Fatal("a loop with nothing to converge on was accepted")
	}
	if !strings.Contains(err.Error(), "converge") {
		t.Errorf("the error should say why the goal matters; got: %v", err)
	}
}

// A conversation alternates. Two prompts in a row is a thread where nothing
// answered the first one, and the layout stacks turns assuming otherwise.
func TestPromptLoopTurnsAlternate(t *testing.T) {
	p := promptLoopPlan()
	p.Beats[1].Loop = &PromptLoopBeat{Turn: "you", Text: "Also make it dark."}
	err := validatePromptLoopPlan(p)
	if err == nil {
		t.Fatal("two prompts in a row were accepted")
	}
	if !strings.Contains(err.Error(), "alternate") {
		t.Errorf("the error should name the rule; got: %v", err)
	}
}

func TestPromptLoopStartsWithAPrompt(t *testing.T) {
	p := promptLoopPlan()
	p.Beats[0].Loop.Turn = "ai"
	err := validatePromptLoopPlan(p)
	if err == nil {
		t.Fatal("a clip opening with the model speaking was accepted")
	}
	if !strings.Contains(err.Error(), "You start") {
		t.Errorf("the error should say who opens; got: %v", err)
	}
}

// The rule the template is named after.
func TestPromptLoopNeedsTwoPrompts(t *testing.T) {
	p := promptLoopPlan()
	p.Beats = p.Beats[:2]
	err := validatePromptLoopPlan(p)
	if err == nil {
		t.Fatal("one ask and one answer was accepted as a loop")
	}
	if !strings.Contains(err.Error(), "demo") {
		t.Errorf("the error should say what one round actually is; got: %v", err)
	}
}

func TestPromptLoopEndsOnAnAnswer(t *testing.T) {
	p := promptLoopPlan()
	p.Beats = append(p.Beats, SnippetBeat{
		ID: "again", Heading: "One more", Narration: plNarration,
		Loop: &PromptLoopBeat{Turn: "you", Text: "Now make the button bigger."},
	})
	err := validatePromptLoopPlan(p)
	if err == nil {
		t.Fatal("a clip ending on an unanswered prompt was accepted")
	}
	if !strings.Contains(err.Error(), "nobody answered") {
		t.Errorf("the error should say what is missing; got: %v", err)
	}
}

// An unlabelled turn is a blank field, not a claim — the position already says
// whose turn it is, so filling it in costs no correction round.
func TestPromptLoopNormalizeInfersTheTurn(t *testing.T) {
	p := promptLoopPlan()
	p.Beats[0].Loop.Turn = ""
	p.Beats[1].Loop.Turn = "assistant" // not in the vocabulary
	normalizePromptLoopPlan(p)
	if p.Beats[0].Loop.Turn != loopTurnYou {
		t.Errorf("beat 0 should have been inferred as %q, got %q", loopTurnYou, p.Beats[0].Loop.Turn)
	}
	if p.Beats[1].Loop.Turn != loopTurnAI {
		t.Errorf("beat 1 should have been inferred as %q, got %q", loopTurnAI, p.Beats[1].Loop.Turn)
	}
	if err := validatePromptLoopPlan(p); err != nil {
		t.Errorf("the repaired plan should validate: %v", err)
	}
}

// A prompt has no result. Left in place, a status on a "you" turn draws a
// verdict over a panel that is still showing the previous attempt.
func TestPromptLoopNormalizeStripsResultsFromPrompts(t *testing.T) {
	p := promptLoopPlan()
	p.Beats[0].Loop.Status = "ok"
	p.Beats[0].Loop.Changes = []string{"invented a result"}
	normalizePromptLoopPlan(p)
	if p.Beats[0].Loop.Status != "" || len(p.Beats[0].Loop.Changes) != 0 {
		t.Error("a prompt turn kept a result it cannot have")
	}
}

func TestPromptLoopNormalizeClamps(t *testing.T) {
	p := promptLoopPlan()
	p.Beats[1].Loop.Status = "sideways"
	p.Beats[1].Loop.Changes = []string{
		"a change whose description runs on well past the limit it was given",
		"two", "three", "four",
	}
	normalizePromptLoopPlan(p)
	if p.Beats[1].Loop.Status != "ok" {
		t.Errorf("an invented status should degrade to ok, got %q", p.Beats[1].Loop.Status)
	}
	if n := len(p.Beats[1].Loop.Changes); n > maxLoopChanges {
		t.Errorf("kept %d changes, want at most %d", n, maxLoopChanges)
	}
	if w := len(strings.Fields(p.Beats[1].Loop.Changes[0])); w > maxLoopChangeWords {
		t.Errorf("change still %d words, want at most %d", w, maxLoopChangeWords)
	}
}

func TestPromptLoopScenesNumberTheAttempts(t *testing.T) {
	p := promptLoopPlan()
	scenes, err := promptLoopScenes(sceneInput(t, p, 8000))
	if err != nil {
		t.Fatalf("laying out the loop failed: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != ScenePromptLoop {
		t.Fatalf("want one %s scene, got %v", ScenePromptLoop, scenes)
	}
	turns, ok := scenes[0].Props["turns"].([]map[string]any)
	if !ok || len(turns) != len(p.Beats) {
		t.Fatalf("want one turn per beat, got %v", scenes[0].Props["turns"])
	}
	// The counter is what makes the loop legible without drawing a loop.
	if turns[1]["attempt"] != 1 || turns[3]["attempt"] != 2 {
		t.Errorf("attempts should count up across answers, got %v and %v", turns[1]["attempt"], turns[3]["attempt"])
	}
	if _, set := turns[0]["attempt"]; set {
		t.Error("a prompt is not an attempt")
	}
}
