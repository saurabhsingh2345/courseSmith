package pipeline

import (
	"strings"
	"testing"
)

const mxNarration = "Three sockets went ready at once, and the one thread took all three in a single pass."

func multiplexPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "multiplex",
		Title:    "One thread, a hundred thousand clients",
		Multiplex: &MultiplexSpec{
			SourceKind: "socket",
			Worker:     "epoll",
			WorkerNote: "1 thread",
			Sources: []MultiplexSource{
				{Label: "#00428"}, {Label: "#00429"}, {Label: "#00430"}, {Label: "#00431"},
				{Label: "#00432"}, {Label: "#00433"}, {Label: "#00434"}, {Label: "#00435"},
			},
			Rounds: []MultiplexRound{
				{Ready: []int{1}, Note: "One socket has data", Role: "neutral"},
				{Ready: []int{4, 6, 7}, Note: "Three woke together, one thread took all three", Role: "quantity"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "pool", Heading: "Eight sockets", Narration: mxNarration, Multiplex: &MultiplexBeat{Show: "pool"}},
			{ID: "one", Heading: "One wakes", Narration: mxNarration, Multiplex: &MultiplexBeat{Show: "round", At: 0}},
			{ID: "three", Heading: "Three at once", Narration: mxNarration, Multiplex: &MultiplexBeat{Show: "round", At: 1}},
			{ID: "why", Heading: "Why that is fast", Narration: mxNarration, Multiplex: &MultiplexBeat{Show: "read"}},
		},
	}
	p.targetWords = 4 * 40
	return p
}

func TestMultiplexPlanAccepted(t *testing.T) {
	if err := validateMultiplexPlan(multiplexPlan()); err != nil {
		t.Fatalf("a well-formed multiplex plan was rejected: %v", err)
	}
}

// The rule the template exists for. Every round waking exactly one source is a
// loop calling accept() in turn, which is the thing multiplexing replaces — so
// the clip would be teaching the opposite of its own claim.
func TestMultiplexRejectsAPoolThatOnlyEverWakesOne(t *testing.T) {
	p := multiplexPlan()
	p.Multiplex.Rounds[1].Ready = []int{4}
	err := validateMultiplexPlan(p)
	if err == nil {
		t.Fatal("a clip where every round wakes one source was accepted")
	}
	if !strings.Contains(err.Error(), "polling") {
		t.Fatalf("the error does not name what was drawn instead: %v", err)
	}
}

// A pass that finds the same things ready is a beat where the picture does not
// move, and the narrator is left describing a still frame.
func TestMultiplexRejectsTwoIdenticalConsecutiveRounds(t *testing.T) {
	p := multiplexPlan()
	p.Multiplex.Rounds[0].Ready = []int{4, 6, 7}
	if err := validateMultiplexPlan(p); err == nil {
		t.Fatal("two consecutive rounds waking the same set were accepted")
	}
}

// Order-insensitively: the same set listed in a different order is still the
// same set, and the picture still does not move.
func TestMultiplexRoundsAreComparedAsSets(t *testing.T) {
	p := multiplexPlan()
	p.Multiplex.Rounds[0].Ready = []int{7, 4, 6}
	if err := validateMultiplexPlan(p); err == nil {
		t.Fatal("the same ready set in a different order was accepted as a different round")
	}
}

func TestMultiplexRequiresTheWorker(t *testing.T) {
	p := multiplexPlan()
	p.Multiplex.Worker = ""
	err := validateMultiplexPlan(p)
	if err == nil {
		t.Fatal("a pool with nothing serving it was accepted")
	}
	if !strings.Contains(err.Error(), "ONE thing") {
		t.Fatalf("the error does not state the claim: %v", err)
	}
}

func TestMultiplexRequiresThePoolFirst(t *testing.T) {
	p := multiplexPlan()
	p.Beats[0].Multiplex = &MultiplexBeat{Show: "round", At: 0}
	p.Beats[1].Multiplex = &MultiplexBeat{Show: "pool"}
	if err := validateMultiplexPlan(p); err == nil {
		t.Fatal("a clip that wakes sources before drawing the pool was accepted")
	}
}

func TestMultiplexRejectsAPoolTooSmallToReadAsMany(t *testing.T) {
	p := multiplexPlan()
	p.Multiplex.Sources = p.Multiplex.Sources[:4]
	p.Multiplex.Rounds[0].Ready = []int{1}
	p.Multiplex.Rounds[1].Ready = []int{2, 3}
	if err := validateMultiplexPlan(p); err == nil {
		t.Fatal("a four-source pool was accepted")
	}
}

func TestMultiplexRejectsADuplicateHandle(t *testing.T) {
	p := multiplexPlan()
	p.Multiplex.Sources[3].Label = "#00428"
	if err := validateMultiplexPlan(p); err == nil {
		t.Fatal("two sources with the same handle were accepted")
	}
}

func TestMultiplexRequiresEveryRoundToRun(t *testing.T) {
	p := multiplexPlan()
	p.Multiplex.Rounds = append(p.Multiplex.Rounds, MultiplexRound{Ready: []int{0, 2}})
	if err := validateMultiplexPlan(p); err == nil {
		t.Fatal("a round that no beat ever runs was accepted")
	}
}

// Normalization drops what it can repair without guessing: an index pointing at
// no source, and a source listed twice in one round — which is not twice as
// ready.
func TestMultiplexNormalizeCleansTheReadySets(t *testing.T) {
	p := multiplexPlan()
	p.Multiplex.Rounds[1].Ready = []int{7, 99, 4, 4, -1, 6}
	normalizeMultiplexPlan(p)
	got := p.Multiplex.Rounds[1].Ready
	want := []int{4, 6, 7}
	if len(got) != len(want) {
		t.Fatalf("ready set is %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ready set is %v, want %v", got, want)
		}
	}
}

// A round where nothing is ready is a round where nothing happens, so it is
// dropped rather than drawn as an empty pass.
func TestMultiplexNormalizeDropsEmptyRounds(t *testing.T) {
	p := multiplexPlan()
	p.Multiplex.Rounds = append(p.Multiplex.Rounds, MultiplexRound{Ready: []int{}})
	normalizeMultiplexPlan(p)
	if len(p.Multiplex.Rounds) != 2 {
		t.Fatalf("want 2 rounds after normalize, got %d", len(p.Multiplex.Rounds))
	}
}

// Handles are cut by characters rather than words: "#00428" clamped to a word
// count is not a handle.
func TestMultiplexNormalizeClampsHandlesByCharacters(t *testing.T) {
	p := multiplexPlan()
	p.Multiplex.Sources[0].Label = strings.Repeat("x", maxMultiplexLabelChars+10)
	normalizeMultiplexPlan(p)
	if got := len(p.Multiplex.Sources[0].Label); got > maxMultiplexLabelChars {
		t.Fatalf("handle is %d chars, want at most %d", got, maxMultiplexLabelChars)
	}
}

func TestMultiplexScenesCarryTheReadySetPerRound(t *testing.T) {
	p := multiplexPlan()
	scenes, err := multiplexScenes(sceneInput(t, p, 16000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	ready, ok := steps[2]["ready"].([]int)
	if !ok || len(ready) != 3 {
		t.Fatalf("the wide round reached the scene as %v, want three ready sources", steps[2]["ready"])
	}
	// The opening beat carries no ready set: the pool is idle.
	if _, has := steps[0]["ready"]; has {
		t.Fatal("the pool beat carries a ready set")
	}
}
