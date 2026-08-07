package pipeline

import (
	"strings"
	"testing"
)

const ldNarration = "Each level down the ladder is bigger and slower, and the request only falls when the level above it misses."

func ladderPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "ladder",
		Title:    "How far away your data really is",
		Ladder: &LadderSpec{
			Rungs: []LadderRung{
				{Label: "registers", LatencyNs: 0.3, Capacity: "a few bytes"},
				{Label: "L1 cache", LatencyNs: 1, Capacity: "32 KB"},
				{Label: "L2 cache", LatencyNs: 4, Capacity: "256 KB"},
				{Label: "RAM", LatencyNs: 100, Capacity: "16 GB"},
				{Label: "SSD", LatencyNs: 100000, Capacity: "1 TB"},
				{Label: "network", LatencyNs: 150000000, Capacity: "everything else"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "ladder", Heading: "The whole ladder", Narration: ldNarration, Ladder: &LadderBeat{Show: "ladder"}},
			{ID: "l1", Heading: "The fast rung", Narration: ldNarration, Ladder: &LadderBeat{Show: "rung", At: 1}},
			{ID: "miss", Heading: "The fall", Narration: ldNarration, Ladder: &LadderBeat{Show: "miss", At: 3}},
			{ID: "spread", Heading: "Top to bottom", Narration: ldNarration, Ladder: &LadderBeat{Show: "spread"}},
		},
	}
	// The template's beat is a 28-word shot, so the fixture's budget is sized
	// against that ideal rather than the shared forty.
	p.targetWords = 4 * 28
	return p
}

func TestLadderPlanAccepted(t *testing.T) {
	if err := validateLadderPlan(ladderPlan()); err != nil {
		t.Fatalf("a well-formed ladder plan was rejected: %v", err)
	}
}

// The rule the template exists for: the ordering IS the subject, and the error
// quotes both out-of-order rungs by name and figure.
func TestLadderRejectsLatenciesOutOfOrder(t *testing.T) {
	p := ladderPlan()
	p.Ladder.Rungs[3].LatencyNs = 2 // RAM faster than L2
	err := validateLadderPlan(p)
	if err == nil {
		t.Fatal("a ladder that gets faster as it gets farther away was accepted")
	}
	if !strings.Contains(err.Error(), "RAM") || !strings.Contains(err.Error(), "L2 cache") {
		t.Fatalf("the error does not quote the two out-of-order rungs: %v", err)
	}
	if !strings.Contains(err.Error(), "2 ns") || !strings.Contains(err.Error(), "4 ns") {
		t.Fatalf("the error does not quote both latencies: %v", err)
	}
}

func TestLadderRejectsEqualLatencies(t *testing.T) {
	p := ladderPlan()
	p.Ladder.Rungs[2].LatencyNs = 1 // equal to L1
	if err := validateLadderPlan(p); err == nil {
		t.Fatal("two rungs at the same latency were accepted, and strictly increasing means strictly")
	}
}

func TestLadderRejectsNonPositiveLatency(t *testing.T) {
	p := ladderPlan()
	p.Ladder.Rungs[0].LatencyNs = 0
	err := validateLadderPlan(p)
	if err == nil {
		t.Fatal("a zero-latency rung was accepted")
	}
	if !strings.Contains(err.Error(), "log axis") {
		t.Fatalf("the error does not say why zero cannot be drawn: %v", err)
	}
}

func TestLadderRejectsTooFewRungs(t *testing.T) {
	p := ladderPlan()
	p.Ladder.Rungs = p.Ladder.Rungs[:3]
	if err := validateLadderPlan(p); err == nil {
		t.Fatal("a three-rung ladder was accepted, and three levels is a comparison rather than a hierarchy")
	}
}

func TestLadderRequiresOpeningOnTheWholeLadder(t *testing.T) {
	p := ladderPlan()
	p.Beats[0].Ladder = &LadderBeat{Show: "rung", At: 0}
	if err := validateLadderPlan(p); err == nil {
		t.Fatal("a clip that focuses a rung before showing the ladder whole was accepted")
	}
}

func TestLadderRequiresClosingOnTheSpread(t *testing.T) {
	p := ladderPlan()
	p.Beats[len(p.Beats)-1].Ladder = &LadderBeat{Show: "rung", At: 5}
	if err := validateLadderPlan(p); err == nil {
		t.Fatal("a clip that never states the top-to-bottom ratio was accepted")
	}
}

// A miss on the bottom rung has nowhere to fall.
func TestLadderRejectsAMissOnTheBottomRung(t *testing.T) {
	p := ladderPlan()
	p.Beats[2].Ladder = &LadderBeat{Show: "miss", At: 5}
	err := validateLadderPlan(p)
	if err == nil {
		t.Fatal("a miss on the last rung was accepted")
	}
	if !strings.Contains(err.Error(), "nowhere to fall") {
		t.Fatalf("the error does not say why the bottom cannot miss: %v", err)
	}
}

func TestLadderRejectsAFocusOffTheLadder(t *testing.T) {
	p := ladderPlan()
	p.Beats[1].Ladder = &LadderBeat{Show: "rung", At: 99}
	if err := validateLadderPlan(p); err == nil {
		t.Fatal("a focus on a rung that does not exist was accepted")
	}
}

func TestLadderNormalizeClampsBeatTargets(t *testing.T) {
	p := ladderPlan()
	p.Beats[1].Ladder.At = 99
	p.Beats[2].Ladder.At = 99
	normalizeLadderPlan(p)
	if got := p.Beats[1].Ladder.At; got != 5 {
		t.Fatalf("a rung focus off the end clamps to the last rung, got %d", got)
	}
	// A miss clamps to the second-to-last rung, because the fall needs a
	// rung to land on.
	if got := p.Beats[2].Ladder.At; got != 4 {
		t.Fatalf("a miss off the end clamps to the second-to-last rung, got %d", got)
	}
}

func TestLadderNormalizeClampsLabels(t *testing.T) {
	p := ladderPlan()
	p.Ladder.Rungs[0].Label = "the very fastest storage there is"
	p.Ladder.Rungs[0].Capacity = "just a handful of machine words"
	normalizeLadderPlan(p)
	if n := len(strings.Fields(p.Ladder.Rungs[0].Label)); n > 3 {
		t.Fatalf("a rung label survived at %d words", n)
	}
	if n := len(strings.Fields(p.Ladder.Rungs[0].Capacity)); n > 4 {
		t.Fatalf("a capacity chip survived at %d words", n)
	}
}

func TestLadderShowDefaultsToRung(t *testing.T) {
	b := LadderBeat{Show: "zoom"}
	if got := b.ResolvedShow(); got != "rung" {
		t.Fatalf("an unknown show resolves to %q, want rung", got)
	}
}

// The log positions and the ratio are the scene's arithmetic, so the test
// checks the numbers rather than the shapes.
func TestLadderScenesPrecomputeTheLogAxis(t *testing.T) {
	p := ladderPlan()
	scenes, err := ladderScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props
	rungs, _ := props["rungs"].([]map[string]any)
	if len(rungs) != 6 {
		t.Fatalf("want 6 rungs in the scene, got %d", len(rungs))
	}
	if pos, _ := rungs[0]["logPos"].(float64); pos != 0 {
		t.Fatalf("the top rung sits at log position %v, want 0", pos)
	}
	if pos, _ := rungs[5]["logPos"].(float64); pos != 1 {
		t.Fatalf("the bottom rung sits at log position %v, want 1", pos)
	}
	// 150000000 / 0.3 = 500,000,000 — the spread, grouped for reading.
	if ratio, _ := props["ratio"].(string); ratio != "×500,000,000" {
		t.Fatalf("the spread ratio is %q, want ×500,000,000", ratio)
	}
	steps, _ := props["steps"].([]map[string]any)
	if steps[0]["show"] != "ladder" {
		t.Fatalf("the first step is %v, want the whole ladder", steps[0]["show"])
	}
	if steps[len(steps)-1]["show"] != "spread" {
		t.Fatalf("the last step is %v, want the spread", steps[len(steps)-1]["show"])
	}
	// The miss carries its own cost: RAM to SSD is 100 -> 100000 ns.
	if cost, _ := steps[2]["cost"].(string); cost != "×1,000 slower" {
		t.Fatalf("the miss cost is %q, want ×1,000 slower", cost)
	}
	if to, _ := steps[2]["to"].(int); to != 4 {
		t.Fatalf("the miss lands on rung %v, want 4", steps[2]["to"])
	}
}
