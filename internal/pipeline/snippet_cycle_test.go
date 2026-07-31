package pipeline

import (
	"math"
	"strings"
	"testing"
)

const cyNarration = "Reproduce it first, because a bug you cannot make happen on demand is a bug you are only guessing at."

func cyclePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "cycle",
		Title:    "The debugging loop",
		Cycle: &CycleSpec{
			Name:    "The debugging loop",
			Changes: "The failing case gets smaller each lap",
			Stages: []CycleStage{
				{Label: "Reproduce", Icon: "repeat", Note: "Get it to fail on demand"},
				{Label: "Isolate", Icon: "search", Note: "Cut away what still fails without it"},
				{Label: "Fix", Icon: "wrench", Note: "Change one thing, and only one"},
				{Label: "Verify", Icon: "check", Note: "Run the case that failed"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "loop", Heading: "A loop", Narration: cyNarration, Cycle: &CycleBeat{Show: "ring"}},
			{ID: "reproduce", Heading: "Make it fail", Narration: cyNarration, Cycle: &CycleBeat{Show: "stage", At: 0}},
			{ID: "isolate", Heading: "Cut it down", Narration: cyNarration, Cycle: &CycleBeat{Show: "stage", At: 1}},
			{ID: "fix", Heading: "One change", Narration: cyNarration, Cycle: &CycleBeat{Show: "stage", At: 2}},
			{ID: "verify", Heading: "Prove it", Narration: cyNarration, Cycle: &CycleBeat{Show: "stage", At: 3}},
			{ID: "again", Heading: "Round again", Narration: cyNarration, Cycle: &CycleBeat{Show: "again"}},
		},
	}
	p.targetWords = 6 * 40
	return p
}

func TestCyclePlanAccepted(t *testing.T) {
	if err := validateCyclePlan(cyclePlan()); err != nil {
		t.Fatalf("a well-formed cycle was rejected: %v", err)
	}
}

// The rule this template exists for. A ring whose second pass is identical to
// its first is a wheel, and drawing one teaches nothing.
func TestCycleMustSayWhatChanges(t *testing.T) {
	p := cyclePlan()
	p.Cycle.Changes = ""
	err := validateCyclePlan(p)
	if err == nil {
		t.Fatal("a loop that never says what improves was accepted")
	}
	if !strings.Contains(err.Error(), "wheel spinning") {
		t.Errorf("the error does not say what is missing: %v", err)
	}
}

func TestCycleEndsOnTheReturn(t *testing.T) {
	p := cyclePlan()
	// Re-pointed rather than removed: dropping a beat trips the shared
	// beat-count floor first and proves nothing about this rule.
	p.Beats[5].Cycle = &CycleBeat{Show: "stage", At: 3}
	p.Beats[4].Cycle = &CycleBeat{Show: "again"}
	err := validateCyclePlan(p)
	if err == nil {
		t.Fatal("a cycle that carried on past its own return was accepted")
	}
	if !strings.Contains(err.Error(), "last frame") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCycleNeedsAReturn(t *testing.T) {
	p := cyclePlan()
	p.Beats[5].Cycle = &CycleBeat{Show: "stage", At: 3}
	p.Beats[4].Cycle = &CycleBeat{Show: "stage", At: 3}
	err := validateCyclePlan(p)
	if err == nil {
		t.Fatal("a cycle that stopped at its last stage was accepted")
	}
	// Two beats both running stage 3 fails the walk first, which is correct —
	// what must not happen is the plan passing.
	if !strings.Contains(err.Error(), "stage") {
		t.Errorf("unexpected error: %v", err)
	}
}

// On a ring, where a stage sits IS the claim about when it happens.
func TestCycleWalksTheRingInOrder(t *testing.T) {
	p := cyclePlan()
	p.Beats[2].Cycle = &CycleBeat{Show: "stage", At: 2}
	p.Beats[3].Cycle = &CycleBeat{Show: "stage", At: 1}
	err := validateCyclePlan(p)
	if err == nil {
		t.Fatal("a cycle narrated out of ring order was accepted")
	}
	if !strings.Contains(err.Error(), "disagrees with the words") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCycleRunsEveryStage(t *testing.T) {
	p := cyclePlan()
	// A fifth stage with the beats left alone: one is never reached while every
	// other rule still passes.
	p.Cycle.Stages = append(p.Cycle.Stages, CycleStage{Label: "Ship", Icon: "rocket"})
	err := validateCyclePlan(p)
	if err == nil {
		t.Fatal("a stage the light never reaches was accepted")
	}
	if !strings.Contains(err.Error(), "left to guess at") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCycleOpensOnTheRing(t *testing.T) {
	p := cyclePlan()
	p.Beats[0].Cycle = &CycleBeat{Show: "stage", At: 0}
	p.Beats[1].Cycle = &CycleBeat{Show: "ring"}
	err := validateCyclePlan(p)
	if err == nil {
		t.Fatal("a cycle that lit a stage before drawing the ring was accepted")
	}
	if !strings.Contains(err.Error(), "word next to it") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCycleRejectsTwoStages(t *testing.T) {
	p := cyclePlan()
	p.Cycle.Stages = p.Cycle.Stages[:2]
	p.Beats = append(p.Beats[:3], p.Beats[5])
	// Budgeted for the four beats a two-stage ring would have, so the shared
	// beat-count floor is not what rejects it.
	p.targetWords = 4 * 40
	err := validateCyclePlan(p)
	if err == nil {
		t.Fatal("a two-stage ring was accepted")
	}
	if !strings.Contains(err.Error(), "pendulum") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCycleRejectsDuplicateStages(t *testing.T) {
	p := cyclePlan()
	p.Cycle.Stages[3].Label = "fix"
	if err := validateCyclePlan(p); err == nil {
		t.Fatal("two stages with the same name were accepted")
	}
}

// The ring's geometry is a property of the plan, so the same stages land in the
// same places on every render.
func TestCycleAnglesStartAtTheTopAndRunClockwise(t *testing.T) {
	got := cyclePlan().Cycle.Angles()
	want := []float64{-90, 0, 90, 180}
	if len(got) != len(want) {
		t.Fatalf("got %d angles for %d stages", len(got), len(want))
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			t.Errorf("stage %d sits at %v, want %v", i, got[i], want[i])
		}
	}
}

func TestNormalizeCycleRepairsMechanicalMistakes(t *testing.T) {
	p := cyclePlan()
	p.Cycle.Changes = "  The failing   case gets smaller and smaller and smaller every single time around  "
	p.Cycle.Stages[0].Icon = "unicorn"
	p.Beats[2].Cycle.Show = "run"
	p.Beats[2].Cycle.At = 99
	normalizeCyclePlan(p)

	if got := len(strings.Fields(p.Cycle.Changes)); got > maxCycleChangesWords {
		t.Errorf("the changes line is %d words, over the %d that fit in the hub", got, maxCycleChangesWords)
	}
	if got := p.Cycle.Stages[0].Icon; got != "" {
		t.Errorf("an invented icon survived normalization: %q", got)
	}
	if got := p.Beats[2].Cycle.Show; got != "stage" {
		t.Errorf("an invented middle-beat action was not repaired: %q", got)
	}
	if got := p.Beats[2].Cycle.At; got >= len(p.Cycle.Stages) {
		t.Errorf("a run to a stage that does not exist was not clamped: %d", got)
	}
}

func TestCycleScenesAreOneStandingRing(t *testing.T) {
	p := cyclePlan()
	scenes, err := cycleScenes(sceneInput(t, p, 9000))
	if err != nil {
		t.Fatalf("laying the loop out: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("want one standing scene, got %d — a scene per beat would remount the ring, and a remount is a cut", len(scenes))
	}
	props := scenes[0].Props
	if props["changes"] != p.Cycle.Changes {
		t.Errorf("the line the template exists for did not reach the renderer: %v", props["changes"])
	}
	stages, _ := props["stages"].([]map[string]any)
	if len(stages) != 4 {
		t.Fatalf("want four stages on the ring, got %d", len(stages))
	}
	if stages[0]["angle"] != -90.0 {
		t.Errorf("the first stage is at %v, want the top of the ring", stages[0]["angle"])
	}
	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != len(p.Beats) {
		t.Fatalf("want a step per beat, got %d", len(steps))
	}
	if steps[len(steps)-1]["show"] != "again" {
		t.Errorf("the clip does not end on the return: %v", steps[len(steps)-1]["show"])
	}
}
