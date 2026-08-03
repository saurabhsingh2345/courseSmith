package pipeline

import (
	"strings"
	"testing"
)

const bgNarration = "Twenty-four gigabytes, and the weights alone take fourteen of them before anything runs."

func budgetPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "budget",
		Title:    "What is really left of 24GB",
		Budget: &BudgetSpec{
			Pot:            24,
			Unit:           "GB",
			PotLabel:       "what a 4090 holds",
			RemainderLabel: "left for your context",
			Claims: []BudgetClaim{
				{Amount: 14, Label: "the model weights", Note: "Before anything runs", Role: "neutral"},
				{Amount: 2.5, Label: "CUDA and the driver", Note: "Gone before your code starts", Role: "neutral"},
				{Amount: 6, Label: "the KV cache", Note: "And it grows with every token", Role: "limit"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "card", Heading: "The card", Narration: bgNarration, Budget: &BudgetBeat{Show: "pot"}},
			{ID: "weights", Heading: "The weights", Narration: bgNarration, Budget: &BudgetBeat{Show: "claim", At: 0}},
			{ID: "overhead", Heading: "The overhead", Narration: bgNarration, Budget: &BudgetBeat{Show: "claim", At: 1}},
			{ID: "cache", Heading: "The cache", Narration: bgNarration, Budget: &BudgetBeat{Show: "claim", At: 2}},
			{ID: "left", Heading: "What is left", Narration: bgNarration, Budget: &BudgetBeat{Show: "remainder"}},
		},
	}
	p.targetWords = 5 * 40
	return p
}

func TestBudgetPlanAccepted(t *testing.T) {
	if err := validateBudgetPlan(budgetPlan()); err != nil {
		t.Fatalf("a well-formed budget plan was rejected: %v", err)
	}
}

// The rule that keeps this from being a worse gauge. One claim against a pot is
// a single quantity measured against a limit, which gauge draws with a threshold.
func TestBudgetRejectsASingleClaim(t *testing.T) {
	p := budgetPlan()
	p.Budget.Claims = p.Budget.Claims[:1]
	p.Beats[2].Budget = &BudgetBeat{Show: "pot"}
	p.Beats[3].Budget = &BudgetBeat{Show: "pot"}
	err := validateBudgetPlan(p)
	if err == nil {
		t.Fatal("a budget with one claim was accepted")
	}
	if !strings.Contains(err.Error(), "gauge") {
		t.Fatalf("the error does not name the template that fits better: %v", err)
	}
}

// If the first bite busts the budget on its own, every claim after it is a beat
// spent on a figure that cannot matter.
func TestBudgetRejectsAClaimBiggerThanThePot(t *testing.T) {
	p := budgetPlan()
	p.Budget.Claims[0].Amount = 30
	err := validateBudgetPlan(p)
	if err == nil {
		t.Fatal("a claim larger than the whole pot was accepted")
	}
	if !strings.Contains(err.Error(), "gauge") {
		t.Fatalf("the error does not point at the right template: %v", err)
	}
}

// A budget that busts is the punchline, not an error: the reference clips keep
// landing on exactly that.
func TestBudgetAllowsTheRemainderToGoNegative(t *testing.T) {
	p := budgetPlan()
	p.Budget.Claims[2].Amount = 12
	if err := validateBudgetPlan(p); err != nil {
		t.Fatalf("an over-budget clip was rejected: %v", err)
	}
	if got := p.Budget.Remainder(); got >= 0 {
		t.Fatalf("remainder is %v, want negative for this fixture", got)
	}
	if got := p.Budget.ResolvedRemainderLabel(); got != "left for your context" {
		t.Fatalf("a stated remainder label was overridden: %q", got)
	}
	// With no stated label, an overrun says so rather than defaulting to "left".
	p.Budget.RemainderLabel = ""
	if got := p.Budget.ResolvedRemainderLabel(); got != "over budget" {
		t.Fatalf("an overrun defaulted to %q", got)
	}
}

func TestBudgetRequiresThePotFirst(t *testing.T) {
	p := budgetPlan()
	p.Beats[0].Budget = &BudgetBeat{Show: "claim", At: 0}
	p.Beats[1].Budget = &BudgetBeat{Show: "pot"}
	if err := validateBudgetPlan(p); err == nil {
		t.Fatal("a clip that spends before showing the pot was accepted")
	}
}

// The remainder is the number the whole clip is for, so it closes the clip.
func TestBudgetRequiresTheRemainderLast(t *testing.T) {
	p := budgetPlan()
	p.Beats[1].Budget = &BudgetBeat{Show: "remainder"}
	p.Beats[4].Budget = &BudgetBeat{Show: "claim", At: 0}
	err := validateBudgetPlan(p)
	if err == nil {
		t.Fatal("a remainder part-way through the clip was accepted")
	}
	if !strings.Contains(err.Error(), "closing frame") {
		t.Fatalf("the error does not say where it belongs: %v", err)
	}
}

func TestBudgetRequiresExactlyOneRemainder(t *testing.T) {
	p := budgetPlan()
	p.Beats[4].Budget = &BudgetBeat{Show: "claim", At: 2}
	p.Beats[3].Budget = &BudgetBeat{Show: "claim", At: 2}
	if err := validateBudgetPlan(p); err == nil {
		t.Fatal("a clip with no remainder beat was accepted")
	}
}

func TestBudgetRequiresEveryClaimToBeTaken(t *testing.T) {
	p := budgetPlan()
	// Re-point rather than delete, so the shared beat floor is not what rejects.
	p.Beats[3].Budget = &BudgetBeat{Show: "claim", At: 1}
	if err := validateBudgetPlan(p); err == nil {
		t.Fatal("a claim no beat ever takes was accepted")
	}
}

func TestBudgetRequiresAUnitAndALabel(t *testing.T) {
	for _, drop := range []string{"unit", "label"} {
		p := budgetPlan()
		if drop == "unit" {
			p.Budget.Unit = ""
		} else {
			p.Budget.PotLabel = ""
		}
		if err := validateBudgetPlan(p); err == nil {
			t.Fatalf("a pot with no %s was accepted", drop)
		}
	}
}

func TestBudgetNormalizeDropsClaimsOfNothing(t *testing.T) {
	p := budgetPlan()
	p.Budget.Claims = append(p.Budget.Claims, BudgetClaim{Amount: 0, Label: "nothing"})
	normalizeBudgetPlan(p)
	if len(p.Budget.Claims) != 3 {
		t.Fatalf("want 3 claims after normalize, got %d", len(p.Budget.Claims))
	}
}

// Segment widths are fractions of the POT, not of the claimed total: a bar that
// re-normalised would show the first claim shrinking as later ones arrived.
func TestBudgetScenesSizeSegmentsAgainstThePot(t *testing.T) {
	p := budgetPlan()
	scenes, err := budgetScenes(sceneInput(t, p, 20000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	claims, _ := scenes[0].Props["claims"].([]map[string]any)
	// 14 of 24.
	if got := claims[0]["frac"].(float64); got < 0.58 || got > 0.585 {
		t.Fatalf("first claim's share of the pot is %v, want ~0.5833", got)
	}
}

// Each step carries the claims taken as a SET, not a count: nothing enforces
// that beats take the claims in declaration order, and a count would have
// reported three landed with one on screen.
func TestBudgetScenesCarryTheTakenSetAndRunningLeft(t *testing.T) {
	p := budgetPlan()
	// Deliberately out of declaration order.
	p.Beats[1].Budget = &BudgetBeat{Show: "claim", At: 2}
	p.Beats[3].Budget = &BudgetBeat{Show: "claim", At: 0}
	scenes, err := budgetScenes(sceneInput(t, p, 20000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	taken, _ := steps[1]["taken"].([]int)
	if len(taken) != 1 || taken[0] != 2 {
		t.Fatalf("after taking claim 2 first, the step carries %v, want [2]", taken)
	}
	if got := steps[1]["left"].(float64); got != 18 {
		t.Fatalf("left after a 6GB claim is %v, want 18", got)
	}
}
