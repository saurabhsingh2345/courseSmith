package pipeline

import (
	"strings"
	"testing"
)

const tgNarration = "No, it is not replacing JavaScript, and the reason that answer is boring is where it actually won."

func togglePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "toggle",
		Title:    "No. But that is not the interesting part.",
		Toggle: &ToggleSpec{
			Question: "Is WebAssembly replacing JavaScript?",
			From:     "yes",
			To:       "no",
			Qualifiers: []ToggleQualifier{
				{Label: "in the browser", Note: "It cannot touch the DOM on its own", Role: "limit"},
				{Label: "outside it", Note: "On the edge it replaced containers instead", Role: "quantity"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "answer", Heading: "The short answer", Narration: tgNarration, Toggle: &ToggleBeat{Show: "answer"}},
			{ID: "browser", Heading: "In the browser", Narration: tgNarration, Toggle: &ToggleBeat{Show: "qualify", At: 0}},
			{ID: "outside", Heading: "Outside it", Narration: tgNarration, Toggle: &ToggleBeat{Show: "qualify", At: 1}},
			{ID: "settled", Heading: "So: no", Narration: tgNarration, Toggle: &ToggleBeat{Show: "settle"}},
		},
	}
	p.targetWords = 4 * 40
	return p
}

func TestTogglePlanAccepted(t *testing.T) {
	if err := validateTogglePlan(togglePlan()); err != nil {
		t.Fatalf("a well-formed toggle plan was rejected: %v", err)
	}
}

// The rule the template exists for, and the inversion of `verdict`. If the clip
// is worth this shape the viewer already arrived asking, so withholding is the
// one thing that could go wrong.
func TestToggleRequiresTheAnswerInTheFirstBeat(t *testing.T) {
	p := togglePlan()
	p.Beats[0].Toggle = &ToggleBeat{Show: "qualify", At: 0}
	p.Beats[1].Toggle = &ToggleBeat{Show: "answer"}
	err := validateTogglePlan(p)
	if err == nil {
		t.Fatal("a clip that withheld the answer was accepted")
	}
	if !strings.Contains(err.Error(), "verdict template") {
		t.Fatalf("the error does not name the template for the other shape: %v", err)
	}
}

// A switch whose two ends say the same thing does not switch.
func TestToggleRejectsAnAnswerThatDoesNotTurn(t *testing.T) {
	p := togglePlan()
	p.Toggle.From = "No"
	p.Toggle.To = "no"
	err := validateTogglePlan(p)
	if err == nil {
		t.Fatal("a switch from an answer to itself was accepted")
	}
	if !strings.Contains(err.Error(), "nothing to flick") {
		t.Fatalf("the error does not explain the picture: %v", err)
	}
}

// A bare answer is a tweet: the form works because the clip starts complicating
// what it just said.
func TestToggleRequiresAtLeastOneQualifier(t *testing.T) {
	p := togglePlan()
	p.Toggle.Qualifiers = nil
	p.Beats[1].Toggle = &ToggleBeat{Show: "settle"}
	p.Beats[2].Toggle = &ToggleBeat{Show: "settle"}
	err := validateTogglePlan(p)
	if err == nil {
		t.Fatal("an answer with no asterisks was accepted")
	}
	if !strings.Contains(err.Error(), "title card") {
		t.Fatalf("the error does not say what it would be instead: %v", err)
	}
}

// A switch that keeps moving is a clip that has not decided.
func TestToggleRejectsThrowingTheSwitchTwice(t *testing.T) {
	p := togglePlan()
	p.Beats[2].Toggle = &ToggleBeat{Show: "answer"}
	if err := validateTogglePlan(p); err == nil {
		t.Fatal("a clip that flicked the switch again was accepted")
	}
}

// Settling is the closing frame: the same answer, now carrying everything.
func TestToggleRequiresSettleToBeLast(t *testing.T) {
	p := togglePlan()
	p.Beats[1].Toggle = &ToggleBeat{Show: "settle"}
	p.Beats[3].Toggle = &ToggleBeat{Show: "qualify", At: 0}
	if err := validateTogglePlan(p); err == nil {
		t.Fatal("a settle part-way through the clip was accepted")
	}
}

func TestToggleRequiresTheQuestionInWords(t *testing.T) {
	p := togglePlan()
	p.Toggle.Question = ""
	err := validateTogglePlan(p)
	if err == nil {
		t.Fatal("a clip with no stated question was accepted")
	}
	if !strings.Contains(err.Error(), "implied") {
		t.Fatalf("the error does not say why stating it matters: %v", err)
	}
}

func TestToggleRequiresEveryQualifierToBeRaised(t *testing.T) {
	p := togglePlan()
	p.Beats[2].Toggle = &ToggleBeat{Show: "qualify", At: 0}
	if err := validateTogglePlan(p); err == nil {
		t.Fatal("a qualifier no beat ever raises was accepted")
	}
}

func TestToggleRejectsDuplicateConditions(t *testing.T) {
	p := togglePlan()
	p.Toggle.Qualifiers[1].Label = "in the browser"
	if err := validateTogglePlan(p); err == nil {
		t.Fatal("two qualifiers with the same condition were accepted")
	}
}

// The comparison ignores case and punctuation, so "No." and "no" are the same
// answer — which is exactly the near-miss a model produces.
func TestToggleComparesAnswersLoosely(t *testing.T) {
	p := togglePlan()
	p.Toggle.From = "Yes!"
	p.Toggle.To = "yes"
	if err := validateTogglePlan(p); err == nil {
		t.Fatal("a switch differing only in punctuation was accepted")
	}
}

// Each step carries the asterisks raised so far, so the renderer draws a whole
// frame from one step.
func TestToggleScenesAccumulateQualifiers(t *testing.T) {
	p := togglePlan()
	scenes, err := toggleScenes(sceneInput(t, p, 16000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if got, _ := steps[0]["raised"].([]int); len(got) != 0 {
		t.Fatalf("the answer beat already has asterisks: %v", got)
	}
	if got, _ := steps[2]["raised"].([]int); len(got) != 2 {
		t.Fatalf("the second qualifier beat carries %v asterisks, want 2", len(got))
	}
	// The closing beat keeps them: settling is the answer with everything it has
	// picked up, not a return to the bare switch.
	if got, _ := steps[3]["raised"].([]int); len(got) != 2 {
		t.Fatalf("the settle beat lost the asterisks: %v", got)
	}
}
