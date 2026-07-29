package pipeline

import (
	"strings"
	"testing"
)

const dcNarration = "Under eight gigabytes there is nothing a bigger card buys you, so stop reading reviews."

func decisionPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "decision",
		Title:    "Which GPU should you actually buy?",
		Decision: &DecisionSpec{
			Question: "How big is your model?",
			Unit:     "GB",
			Tiers: []DecisionTier{
				{UpTo: 8, Band: "Under 8GB", Answer: "A used 3060 is enough", Note: "Nothing bigger helps", Role: "quantity"},
				{UpTo: 24, Band: "8 to 24GB", Answer: "Buy the 4090", Note: "One card still does it", Role: "rival"},
				{Band: "Over 24GB", Answer: "Rent it by the hour", Note: "Cheaper than two cards", Role: "limit"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "question", Heading: "The only question", Narration: dcNarration, Decision: &DecisionBeat{Show: "question"}},
			{ID: "small", Heading: "Small models", Narration: dcNarration, Decision: &DecisionBeat{Show: "tier", At: 0}},
			{ID: "middle", Heading: "The sweet spot", Narration: dcNarration, Decision: &DecisionBeat{Show: "tier", At: 1}},
			{ID: "big", Heading: "Past one card", Narration: dcNarration, Decision: &DecisionBeat{Show: "tier", At: 2}},
			{ID: "rule", Heading: "The whole rule", Narration: dcNarration, Decision: &DecisionBeat{Show: "rule"}},
		},
	}
	p.targetWords = 5 * 40
	return p
}

func TestDecisionPlanAccepted(t *testing.T) {
	if err := validateDecisionPlan(decisionPlan()); err != nil {
		t.Fatalf("a well-formed decision was rejected: %v", err)
	}
}

// The rule this template exists for: the bands must partition the axis. A guide
// with a hole in it tells some fraction of the audience nothing.
func TestDecisionLastTierIsOpenEnded(t *testing.T) {
	p := decisionPlan()
	p.Decision.Tiers[2].UpTo = 48
	err := validateDecisionPlan(p)
	if err == nil {
		t.Fatal("a guide whose last band stops somewhere was accepted — anyone above it is told nothing")
	}
	if !strings.Contains(err.Error(), "open-ended") {
		t.Errorf("the error does not say what to do: %v", err)
	}
}

func TestDecisionBoundsMustAscend(t *testing.T) {
	p := decisionPlan()
	p.Decision.Tiers[1].UpTo = 4 // below the band before it
	err := validateDecisionPlan(p)
	if err == nil {
		t.Fatal("overlapping bands were accepted — a viewer in the overlap gets two answers")
	}
	if !strings.Contains(err.Error(), "overlap") {
		t.Errorf("unexpected error: %v", err)
	}

	// A non-final band with no bound at all leaves a hole the same way.
	p = decisionPlan()
	p.Decision.Tiers[0].UpTo = 0
	if err := validateDecisionPlan(p); err == nil {
		t.Fatal("a non-final band with no upper bound was accepted")
	}
}

// Every band ends in an instruction. A band that tells its viewer nothing is
// the exact failure the format exists to prevent.
func TestDecisionEveryTierHasAnAnswer(t *testing.T) {
	p := decisionPlan()
	p.Decision.Tiers[1].Answer = ""
	err := validateDecisionPlan(p)
	if err == nil {
		t.Fatal("a band with no answer was accepted")
	}
	if !strings.Contains(err.Error(), "instruction") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDecisionOpensOnTheQuestion(t *testing.T) {
	p := decisionPlan()
	p.Decision.Tiers = p.Decision.Tiers[:2]
	p.Decision.Tiers[1].UpTo = 0
	p.Beats = p.Beats[1:]
	p.Beats[0].Decision = &DecisionBeat{Show: "tier", At: 0}
	p.Beats = append(p.Beats, SnippetBeat{
		ID: "extra", Heading: "One more", Narration: dcNarration,
		Decision: &DecisionBeat{Show: "question"},
	})
	p.Beats[2].Decision = &DecisionBeat{Show: "rule"}
	p.Beats = p.Beats[:4]
	err := validateDecisionPlan(p)
	if err == nil {
		t.Fatal("a clip that lit a band before posing the question was accepted")
	}
	if !strings.Contains(err.Error(), "answer to nothing") {
		t.Errorf("the error does not say why: %v", err)
	}
}

func TestDecisionEveryBandIsNarrated(t *testing.T) {
	p := decisionPlan()
	// Re-point the third band's beat rather than deleting it, so the shared
	// beat-count floor does not fire first and prove nothing about this rule.
	p.Beats[3].Decision = &DecisionBeat{Show: "tier", At: 1}
	err := validateDecisionPlan(p)
	if err == nil {
		t.Fatal("a band nobody narrates was accepted")
	}
	// Landing on the same band twice is caught before the coverage check; both
	// are the same defect seen from two sides.
	if !strings.Contains(err.Error(), "again") && !strings.Contains(err.Error(), "never spoken") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDecisionRuleIsLast(t *testing.T) {
	p := decisionPlan()
	p.Beats[1].Decision = &DecisionBeat{Show: "rule"}
	p.Beats[4].Decision = &DecisionBeat{Show: "tier", At: 0}
	if err := validateDecisionPlan(p); err == nil {
		t.Fatal("a rule beat in the middle was accepted")
	}
}

func TestDecisionNormalizeRepairs(t *testing.T) {
	p := decisionPlan()
	p.Decision.Tiers[0].Role = "invented"
	p.Decision.Tiers[1].UpTo = -24
	p.Beats[2].Decision.At = 99
	p.Beats[3].Decision.Show = "nonsense"
	normalizeDecisionPlan(p)

	if p.Decision.Tiers[0].Role != "neutral" {
		t.Errorf("an invented role became %q, want neutral", p.Decision.Tiers[0].Role)
	}
	if p.Decision.Tiers[1].UpTo != 0 {
		t.Errorf("a negative bound became %v, want 0", p.Decision.Tiers[1].UpTo)
	}
	if p.Beats[2].Decision.At != len(p.Decision.Tiers)-1 {
		t.Errorf("an out-of-range beat points at %d", p.Beats[2].Decision.At)
	}
	if p.Beats[3].Decision.Show != "tier" {
		t.Errorf("an unknown show became %q, want tier", p.Beats[3].Decision.Show)
	}
}

// A band with no answer is dropped rather than given invented advice.
func TestDecisionNormalizeDropsAnswerlessTiers(t *testing.T) {
	p := decisionPlan()
	p.Decision.Tiers = append(p.Decision.Tiers, DecisionTier{Band: "Somewhere", Answer: ""})
	normalizeDecisionPlan(p)
	for _, tr := range p.Decision.Tiers {
		if strings.TrimSpace(tr.Answer) == "" {
			t.Error("a band with no answer survived normalize — inventing advice would be worse than saying nothing")
		}
	}
}

func TestDecisionScenesShape(t *testing.T) {
	p := decisionPlan()
	scenes, err := decisionScenes(sceneInput(t, p, 8000))
	if err != nil {
		t.Fatalf("decisionScenes: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneDecision {
		t.Fatalf("want one decision scene, got %d of %q", len(scenes), scenes[0].Type)
	}
	tiers, ok := scenes[0].Props["tiers"].([]map[string]any)
	if !ok || len(tiers) != 3 {
		t.Fatalf("want three tiers on the scene, got %v", scenes[0].Props["tiers"])
	}
	// The bounds deliberately do NOT reach the renderer: the segments are drawn
	// evenly and the arithmetic lives in the band labels, because a
	// proportional axis gives the open-ended band no width at all.
	if _, leaked := tiers[0]["upTo"]; leaked {
		t.Error("upTo reached the scene — the axis is a sequence of cases, not a measuring stick")
	}
	if tiers[2]["role"] != "limit" {
		t.Errorf("role = %v, want limit", tiers[2]["role"])
	}
}
