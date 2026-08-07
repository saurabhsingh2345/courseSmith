package pipeline

import (
	"strings"
	"testing"
)

const lcNarration = "You will install the tools first and then boot the machine and check the screen."

func labCardPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "labcard",
		Title:    "Boot a whole computer inside yours",
		LabCard: &LabCardSpec{
			Task:  "install VirtualBox and boot Ubuntu",
			Tools: []string{"VirtualBox", "Ubuntu ISO"},
			Steps: []string{
				"Download the Ubuntu desktop ISO",
				"Create a new virtual machine",
				"Attach the ISO as a disc",
				"Start the machine and pick Try Ubuntu",
			},
			Expect: "a purple Ubuntu desktop with no errors",
		},
		Beats: []SnippetBeat{
			{ID: "brief", Heading: "The task", Narration: lcNarration, LabCard: &LabCardBeat{Show: "task"}},
			{ID: "s1", Heading: "Get the image", Narration: lcNarration, LabCard: &LabCardBeat{Show: "step", At: 0}},
			{ID: "s2", Heading: "Make the machine", Narration: lcNarration, LabCard: &LabCardBeat{Show: "step", At: 1}},
			{ID: "s3", Heading: "Attach the disc", Narration: lcNarration, LabCard: &LabCardBeat{Show: "step", At: 2}},
			{ID: "s4", Heading: "Boot it", Narration: lcNarration, LabCard: &LabCardBeat{Show: "step", At: 3}},
			{ID: "done", Heading: "What you should see", Narration: lcNarration, LabCard: &LabCardBeat{Show: "expect"}},
		},
	}
	// A card template declares no IdealWordsPerBeat, so its budget is sized
	// against the shared forty words a beat.
	p.targetWords = 6 * 40
	return p
}

func TestLabCardPlanAccepted(t *testing.T) {
	if err := validateLabCardPlan(labCardPlan()); err != nil {
		t.Fatalf("a well-formed lab card was rejected: %v", err)
	}
}

func TestLabCardRejectsTooFewSteps(t *testing.T) {
	p := labCardPlan()
	p.LabCard.Steps = p.LabCard.Steps[:2]
	err := validateLabCardPlan(p)
	if err == nil {
		t.Fatal("a two-step lab was accepted, and two steps have no order to get wrong")
	}
	if !strings.Contains(err.Error(), "2 steps") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestLabCardRejectsTooManySteps(t *testing.T) {
	p := labCardPlan()
	p.LabCard.Steps = append(p.LabCard.Steps, "One", "Two", "Three")
	if err := validateLabCardPlan(p); err == nil {
		t.Fatal("a seven-step lab was accepted, and seven rows do not read at the size they are set")
	}
}

func TestLabCardRejectsAStepThatIsASentence(t *testing.T) {
	p := labCardPlan()
	p.LabCard.Steps[1] = "Open the VirtualBox window and click the new button and then give the machine a name and pick Linux"
	err := validateLabCardPlan(p)
	if err == nil {
		t.Fatal("a step long enough to wrap was accepted")
	}
	if !strings.Contains(err.Error(), "step 2") {
		t.Fatalf("the error does not identify which step: %v", err)
	}
}

func TestLabCardRejectsNoTools(t *testing.T) {
	p := labCardPlan()
	p.LabCard.Tools = nil
	if err := validateLabCardPlan(p); err == nil {
		t.Fatal("a lab with no prerequisites was accepted, and the tool row would be an empty box")
	}
}

func TestLabCardRejectsTooManyTools(t *testing.T) {
	p := labCardPlan()
	p.LabCard.Tools = []string{"one", "two", "three", "four", "five"}
	if err := validateLabCardPlan(p); err == nil {
		t.Fatal("a lab needing five separate installs was accepted")
	}
}

func TestLabCardRejectsAMissingExpectedResult(t *testing.T) {
	p := labCardPlan()
	p.LabCard.Expect = ""
	err := validateLabCardPlan(p)
	if err == nil {
		t.Fatal("a card with no success criterion was accepted")
	}
	if !strings.Contains(err.Error(), "worked") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLabCardRequiresOpeningOnTheTask(t *testing.T) {
	p := labCardPlan()
	p.Beats[0].LabCard = &LabCardBeat{Show: "step", At: 0}
	p.Beats[1].LabCard = &LabCardBeat{Show: "task"}
	if err := validateLabCardPlan(p); err == nil {
		t.Fatal("a clip that walks a step before naming the task was accepted")
	}
}

func TestLabCardRequiresClosingOnTheExpectedResult(t *testing.T) {
	p := labCardPlan()
	p.Beats[len(p.Beats)-1].LabCard = &LabCardBeat{Show: "step", At: 3}
	if err := validateLabCardPlan(p); err == nil {
		t.Fatal("a clip that never says what success looks like was accepted")
	}
}

// The order is the point: a highlight that doubles back teaches the viewer that
// the numbers are decoration.
func TestLabCardRejectsAHighlightThatDoublesBack(t *testing.T) {
	p := labCardPlan()
	p.Beats[3].LabCard = &LabCardBeat{Show: "step", At: 1}
	err := validateLabCardPlan(p)
	if err == nil {
		t.Fatal("a step lit twice was accepted")
	}
	if !strings.Contains(err.Error(), "step 1 after step 1") {
		t.Fatalf("the error does not quote the two indices: %v", err)
	}
}

func TestLabCardRejectsASkippedStepByName(t *testing.T) {
	p := labCardPlan()
	p.Beats[2].LabCard = &LabCardBeat{Show: "step", At: 2}
	err := validateLabCardPlan(p)
	if err == nil {
		t.Fatal("a walk that jumped over a step was accepted")
	}
	if !strings.Contains(err.Error(), "Create a new virtual machine") {
		t.Fatalf("the error does not name the skipped step: %v", err)
	}
}

func TestLabCardRejectsAWalkThatStartsPastTheFirstStep(t *testing.T) {
	p := labCardPlan()
	p.Beats[1].LabCard = &LabCardBeat{Show: "step", At: 1}
	err := validateLabCardPlan(p)
	if err == nil {
		t.Fatal("a walk that skipped step zero was accepted")
	}
	if !strings.Contains(err.Error(), "Download the Ubuntu desktop ISO") {
		t.Fatalf("the error does not name the skipped step: %v", err)
	}
}

// The counting rule, with both numbers quoted: a step nobody says out loud is a
// step the learner does not do.
func TestLabCardRejectsStepsNeverWalked(t *testing.T) {
	p := labCardPlan()
	p.LabCard.Steps = append(p.LabCard.Steps, "Shut the machine down cleanly")
	err := validateLabCardPlan(p)
	if err == nil {
		t.Fatal("a card with a step no beat covers was accepted")
	}
	if !strings.Contains(err.Error(), "4 of the card's 5") {
		t.Fatalf("the error does not quote both counts: %v", err)
	}
	if !strings.Contains(err.Error(), "Shut the machine down cleanly") {
		t.Fatalf("the error does not name the uncovered step: %v", err)
	}
}

func TestLabCardRejectsAStepIndexOffTheCard(t *testing.T) {
	p := labCardPlan()
	p.Beats[4].LabCard = &LabCardBeat{Show: "step", At: 9}
	if err := validateLabCardPlan(p); err == nil {
		t.Fatal("a beat pointing past the end of the step list was accepted")
	}
}

// Over-long labels and a stray index are phrasing habits, not wrong answers, so
// they are repaired rather than argued.
func TestLabCardNormalizeTrimsAndClampsTheCard(t *testing.T) {
	p := labCardPlan()
	p.LabCard.Task = "  install   VirtualBox and boot Ubuntu on your own laptop this afternoon  "
	p.LabCard.Tools = []string{"VirtualBox", "", "the Ubuntu desktop ISO image"}
	p.Beats[4].LabCard.At = 9
	normalizeLabCardPlan(p)

	if got := len(strings.Fields(p.LabCard.Task)); got > maxLabCardTaskWords {
		t.Fatalf("the task normalized to %d words, want at most %d: %q", got, maxLabCardTaskWords, p.LabCard.Task)
	}
	if len(p.LabCard.Tools) != 2 {
		t.Fatalf("the blank tool survived normalize: %v", p.LabCard.Tools)
	}
	if p.LabCard.Tools[1] != "the Ubuntu desktop" {
		t.Fatalf("the long tool normalized to %q", p.LabCard.Tools[1])
	}
	if p.Beats[4].LabCard.At != 3 {
		t.Fatalf("an index past the end clamped to %d, want 3", p.Beats[4].LabCard.At)
	}
	if err := validateLabCardPlan(p); err != nil {
		t.Fatalf("a repairable card was rejected after normalize: %v", err)
	}
}

func TestLabCardShowDefaultsToStep(t *testing.T) {
	b := LabCardBeat{Show: "briefing"}
	if got := b.ResolvedShow(); got != "step" {
		t.Fatalf("an unknown show resolved to %q, want step", got)
	}
}

// The component draws a card and a highlight; which rows are behind the learner
// at any moment arrives precomputed.
func TestLabCardScenesAccumulateTheWalkedSteps(t *testing.T) {
	p := labCardPlan()
	scenes, err := labCardScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	list, _ := props["stepList"].([]map[string]any)
	if len(list) != 4 {
		t.Fatalf("want 4 numbered rows, got %d", len(list))
	}
	if list[0]["n"] != 1 || list[0]["text"] != "Download the Ubuntu desktop ISO" {
		t.Fatalf("the first row is wrong: %v", list[0])
	}
	tools, _ := props["tools"].([]map[string]any)
	if len(tools) != 2 || tools[0]["name"] != "VirtualBox" {
		t.Fatalf("the tool chips are wrong: %v", tools)
	}

	steps, _ := props["steps"].([]map[string]any)
	if steps[0]["show"] != "task" {
		t.Fatalf("first step shows %v, want task", steps[0]["show"])
	}
	if got, _ := steps[0]["reached"].([]int); len(got) != 0 {
		t.Fatalf("the opener has already walked steps: %v", got)
	}

	last := steps[len(steps)-1]
	if last["show"] != "expect" {
		t.Fatalf("last step shows %v, want expect", last["show"])
	}
	reached, _ := last["reached"].([]int)
	if len(reached) != 4 || reached[0] != 0 || reached[3] != 3 {
		t.Fatalf("the closer has not accumulated every step: %v", reached)
	}
}
