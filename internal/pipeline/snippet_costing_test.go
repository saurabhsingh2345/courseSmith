package pipeline

import (
	"strings"
	"testing"
)

const csNarration = "Four hundred watts running most of the day for a year is not a rounding error."

func costingPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "costing",
		Title:    "What a GPU box really costs",
		Costing: &CostingSpec{
			Subject: "A self-hosted GPU box",
			Unit:    "$",
			Lines: []CostLine{
				{Label: "The card", Amount: 1800, Note: "The number everyone quotes"},
				{Label: "The rest of the box", Amount: 1100, Note: "Power supply, board, case"},
				{Label: "Electricity", Amount: 620, Note: "Four hundred watts for a year", Hidden: true},
				{Label: "The noise fix", Amount: 340, Note: "Nobody keeps it in the room", Hidden: true},
			},
			Total:   3860,
			Verdict: "Twice the sticker price",
		},
		Beats: []SnippetBeat{
			{ID: "setup", Heading: "What we are pricing", Narration: csNarration, Costing: &CostingBeat{Show: "setup"}},
			{ID: "card", Heading: "The card", Narration: csNarration, Costing: &CostingBeat{Show: "line", At: 0}},
			{ID: "box", Heading: "The rest", Narration: csNarration, Costing: &CostingBeat{Show: "line", At: 1}},
			{ID: "power", Heading: "The power bill", Narration: csNarration, Costing: &CostingBeat{Show: "line", At: 2}},
			{ID: "noise", Heading: "The noise", Narration: csNarration, Costing: &CostingBeat{Show: "line", At: 3}},
			{ID: "total", Heading: "Year one", Narration: csNarration, Costing: &CostingBeat{Show: "total"}},
		},
	}
	p.targetWords = 6 * 40
	return p
}

func TestCostingPlanAccepted(t *testing.T) {
	if err := validateCostingPlan(costingPlan()); err != nil {
		t.Fatalf("a well-formed costing was rejected: %v", err)
	}
}

// The rule this template exists for: the arithmetic has to be right. It is the
// thing a language model gets wrong, because the total is generated from the
// shape of the list rather than from the list.
func TestCostingTotalMustMatchTheSum(t *testing.T) {
	p := costingPlan()
	p.Costing.Total = 2900 // forgot the two hidden lines
	err := validateCostingPlan(p)
	if err == nil {
		t.Fatal("a total that does not match the lines was accepted")
	}
	if !strings.Contains(err.Error(), "wrong sum") {
		t.Errorf("the error does not say why it matters: %v", err)
	}
}

// Rounding a bill to a readable figure is honest; being out by a tenth is
// arithmetic nobody did.
func TestCostingAllowsHonestRounding(t *testing.T) {
	p := costingPlan()
	p.Costing.Total = 3900 // 1% high — a rounded headline figure
	if err := validateCostingPlan(p); err != nil {
		t.Errorf("a total rounded within tolerance was rejected: %v", err)
	}
	p.Costing.Total = 4400 // 14% high
	if err := validateCostingPlan(p); err == nil {
		t.Error("a total well outside the tolerance was accepted")
	}
}

// The second rule: a bill made entirely of obvious items has told the viewer
// what they already knew.
func TestCostingNeedsAHiddenLine(t *testing.T) {
	p := costingPlan()
	for i := range p.Costing.Lines {
		p.Costing.Lines[i].Hidden = false
	}
	err := validateCostingPlan(p)
	if err == nil {
		t.Fatal("a bill with no surprising line was accepted")
	}
	if !strings.Contains(err.Error(), "sticker price is not the price") {
		t.Errorf("unexpected error: %v", err)
	}
}

// A blank or badly-rounded total is repaired rather than rejected: the sum is
// knowable from the lines, so failing the plan would burn a correction round on
// arithmetic the pipeline can simply do.
func TestCostingNormalizeFixesTheTotal(t *testing.T) {
	p := costingPlan()
	p.Costing.Total = 0
	normalizeCostingPlan(p)
	if p.Costing.Total != 3860 {
		t.Errorf("a blank total became %v, want the sum 3860", p.Costing.Total)
	}

	// But a total that is wildly wrong is left alone, so the validator can
	// reject it and the model gets told which line to look at.
	p = costingPlan()
	p.Costing.Total = 2900
	normalizeCostingPlan(p)
	if p.Costing.Total != 2900 {
		t.Errorf("a badly wrong total was silently rewritten to %v — that hides the model's mistake instead of correcting it", p.Costing.Total)
	}
}

func TestCostingNormalizeRepairsLines(t *testing.T) {
	p := costingPlan()
	p.Costing.Lines[0].Amount = -1800 // a sign error, not a credit
	p.Costing.Lines = append(p.Costing.Lines, CostLine{Label: "Nothing", Amount: 0})
	p.Beats[2].Costing.At = 99
	normalizeCostingPlan(p)

	if p.Costing.Lines[0].Amount != 1800 {
		t.Errorf("a negative line became %v, want its magnitude", p.Costing.Lines[0].Amount)
	}
	for _, l := range p.Costing.Lines {
		if l.Amount <= 0 {
			t.Error("a zero-amount line survived normalize")
		}
	}
	if p.Beats[2].Costing.At != len(p.Costing.Lines)-1 {
		t.Errorf("an out-of-range beat points at %d", p.Beats[2].Costing.At)
	}
}

func TestCostingOpensOnSetupAndEndsOnTotal(t *testing.T) {
	p := costingPlan()
	p.Beats[0].Costing = &CostingBeat{Show: "line", At: 0}
	if err := validateCostingPlan(p); err == nil {
		t.Fatal("a clip that priced before naming the subject was accepted")
	}

	p = costingPlan()
	p.Beats[1].Costing = &CostingBeat{Show: "total"}
	if err := validateCostingPlan(p); err == nil {
		t.Fatal("a total with the clip carrying on afterwards was accepted")
	}
}

func TestCostingSpeaksEveryLine(t *testing.T) {
	p := costingPlan()
	p.Beats[4].Costing = &CostingBeat{Show: "line", At: 2}
	err := validateCostingPlan(p)
	if err == nil {
		t.Fatal("a line nobody narrates was accepted")
	}
	if !strings.Contains(err.Error(), "again") && !strings.Contains(err.Error(), "never spoken") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCostingScenesCarryTheRunningTotal(t *testing.T) {
	p := costingPlan()
	scenes, err := costingScenes(sceneInput(t, p, 6000))
	if err != nil {
		t.Fatalf("costingScenes: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneCosting {
		t.Fatalf("want one costing scene, got %d of %q", len(scenes), scenes[0].Type)
	}
	lines := scenes[0].Props["lines"].([]map[string]any)
	// The running total is computed in Go, so the number on screen and the
	// number the validator checked are the same by construction.
	for i, want := range []float64{1800, 2900, 3520, 3860} {
		if lines[i]["running"] != want {
			t.Errorf("line %d running = %v, want %v", i, lines[i]["running"], want)
		}
	}
	// Bars scale against the biggest single line, not the total: against the
	// total the small lines vanish, and the small ones are the surprising ones.
	if lines[0]["frac"] != 1.0 {
		t.Errorf("the largest line's bar is %v, want a full 1.0", lines[0]["frac"])
	}
	if f := lines[2]["frac"].(float64); f <= 0.2 || f >= 0.5 {
		t.Errorf("the hidden line's bar is %v — scaled against the total it would have collapsed to nothing", f)
	}
}
