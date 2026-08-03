package pipeline

import (
	"strings"
	"testing"
)

const mpNarration = "One node draws about what an oven does, and eight of them draw more than the floor is wired for."

func multiplyPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "multiply",
		Title:    "One node is fine. Eight is a substation.",
		Multiply: &MultiplySpec{
			UnitValue:  14.5,
			Unit:       "kW",
			UnitLabel:  "one B200 node",
			UnitNote:   "About what a domestic oven pulls",
			Count:      8,
			CountLabel: "nodes in one rack",
			Total:      116,
			TotalLabel: "before cooling",
			TotalNote:  "More than most office floors are wired for",
			Caveat:     "And cooling adds roughly half again",
			Role:       "limit",
		},
		Beats: []SnippetBeat{
			{ID: "one", Heading: "One node", Narration: mpNarration, Multiply: &MultiplyBeat{Show: "unit"}},
			{ID: "rack", Heading: "A full rack", Narration: mpNarration, Multiply: &MultiplyBeat{Show: "count"}},
			{ID: "draw", Heading: "What that draws", Narration: mpNarration, Multiply: &MultiplyBeat{Show: "total"}},
			{ID: "cooling", Heading: "Before cooling", Narration: mpNarration, Multiply: &MultiplyBeat{Show: "caveat"}},
		},
	}
	p.targetWords = 4 * 40
	return p
}

func TestMultiplyPlanAccepted(t *testing.T) {
	if err := validateMultiplyPlan(multiplyPlan()); err != nil {
		t.Fatalf("a well-formed multiply plan was rejected: %v", err)
	}
}

// The rule the whole template exists for, and the only arithmetic check in the
// catalog. A clip that says "14.5 times 8, so a hundred" has taught the viewer
// to distrust everything else it said.
func TestMultiplyRejectsArithmeticThatDoesNotCheckOut(t *testing.T) {
	p := multiplyPlan()
	p.Multiply.Total = 100
	// Far enough out that normalize will not quietly correct it.
	err := validateMultiplyPlan(p)
	if err == nil {
		t.Fatal("a total that is not the product was accepted")
	}
	if !strings.Contains(err.Error(), "116") {
		t.Fatalf("the error does not state the right answer: %v", err)
	}
}

// Rounding for speech is right, not wrong: 14.5 x 8 stated as "116" and as
// "116.0" are the same number.
func TestMultiplyAcceptsATotalRoundedForSpeech(t *testing.T) {
	p := multiplyPlan()
	p.Multiply.UnitValue = 14.52
	p.Multiply.Total = 116 // 116.16 rounded
	if err := validateMultiplyPlan(p); err != nil {
		t.Fatalf("a total rounded for speech was rejected: %v", err)
	}
}

// A near-miss is arithmetic rather than a claim about the subject, and Go can do
// arithmetic — so it is corrected rather than costing a round.
func TestMultiplyNormalizeFixesANearMiss(t *testing.T) {
	p := multiplyPlan()
	p.Multiply.Total = 120 // out by 3.4%, within the repair band
	normalizeMultiplyPlan(p)
	if p.Multiply.Total != 116 {
		t.Fatalf("total is %v after normalize, want 116", p.Multiply.Total)
	}
}

// A total that is wildly out is a different intention, not a slip, so it is left
// for the validator to reject with an explanation.
func TestMultiplyNormalizeLeavesAWildTotalAlone(t *testing.T) {
	p := multiplyPlan()
	p.Multiply.Total = 4000
	normalizeMultiplyPlan(p)
	if p.Multiply.Total != 4000 {
		t.Fatalf("normalize silently rewrote a wildly wrong total to %v", p.Multiply.Total)
	}
	if err := validateMultiplyPlan(p); err == nil {
		t.Fatal("a wildly wrong total was accepted")
	}
}

// The move only works if the viewer has accepted the small number first.
func TestMultiplyRequiresTheUnitFigureFirst(t *testing.T) {
	p := multiplyPlan()
	p.Beats[0].Multiply = &MultiplyBeat{Show: "count"}
	p.Beats[1].Multiply = &MultiplyBeat{Show: "unit"}
	err := validateMultiplyPlan(p)
	if err == nil {
		t.Fatal("a clip that states the count before the figure was accepted")
	}
	if !strings.Contains(err.Error(), "sounds reasonable") {
		t.Fatalf("the error does not explain the move: %v", err)
	}
}

// The order is the argument: any other sequence gives the product away before
// the multiplication happens.
func TestMultiplyRejectsTheTotalBeforeTheCount(t *testing.T) {
	p := multiplyPlan()
	p.Beats[1].Multiply = &MultiplyBeat{Show: "total"}
	p.Beats[2].Multiply = &MultiplyBeat{Show: "count"}
	if err := validateMultiplyPlan(p); err == nil {
		t.Fatal("a clip that states the product before the count was accepted")
	}
}

func TestMultiplyRequiresEachPartExactlyOnce(t *testing.T) {
	p := multiplyPlan()
	p.Beats[3].Multiply = &MultiplyBeat{Show: "total"}
	if err := validateMultiplyPlan(p); err == nil {
		t.Fatal("a clip stating the product twice was accepted")
	}
}

func TestMultiplyRejectsAnUndrawableCount(t *testing.T) {
	for _, n := range []int{2, 200} {
		p := multiplyPlan()
		p.Multiply.Count = n
		p.Multiply.Total = 14.5 * float64(n)
		if err := validateMultiplyPlan(p); err == nil {
			t.Fatalf("a count of %d was accepted", n)
		}
	}
}

func TestMultiplyRequiresLabelsOnBothSides(t *testing.T) {
	for _, drop := range []string{"unit", "count"} {
		p := multiplyPlan()
		if drop == "unit" {
			p.Multiply.UnitLabel = ""
		} else {
			p.Multiply.CountLabel = ""
		}
		if err := validateMultiplyPlan(p); err == nil {
			t.Fatalf("a plan with no %s label was accepted", drop)
		}
	}
}

// The product's role defaults to the limit: the point of a clip like this is
// almost always that the total is a problem.
func TestMultiplyRoleDefaultsToLimit(t *testing.T) {
	m := &MultiplySpec{}
	if m.ResolvedRole() != "limit" {
		t.Fatalf("default role is %q, want limit", m.ResolvedRole())
	}
	m.Role = "quantity"
	if m.ResolvedRole() != "quantity" {
		t.Fatalf("a stated role was ignored: %q", m.ResolvedRole())
	}
}

// A caveat beat with no caveat text would render an empty chip.
func TestMultiplyRejectsACaveatBeatWithNoCaveat(t *testing.T) {
	p := multiplyPlan()
	p.Multiply.Caveat = ""
	if err := validateMultiplyPlan(p); err == nil {
		t.Fatal("a caveat beat with nothing to show was accepted")
	}
}

func TestMultiplyScenesCarryTheWholeStatement(t *testing.T) {
	p := multiplyPlan()
	scenes, err := multiplyScenes(sceneInput(t, p, 16000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props
	if props["unitValue"] != 14.5 || props["count"] != 8 || props["total"] != float64(116) {
		t.Fatalf("the statement did not reach the scene: %v / %v / %v",
			props["unitValue"], props["count"], props["total"])
	}
	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != 4 || steps[0]["show"] != "unit" || steps[3]["show"] != "caveat" {
		t.Fatalf("the beat sequence did not reach the scene: %v", steps)
	}
}
