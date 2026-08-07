package pipeline

import (
	"strings"
	"testing"
)

const brNarration = "The last lesson left you holding something real, and this one starts exactly where that one stopped."

func bridgePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "bridge",
		Title:    "You know the bits, now find them",
		Bridge: &BridgeSpec{
			From:        "Binary & Data",
			To:          "Memory & Storage",
			Established: []string{"a bit is one two-way switch", "a byte spells a character or a number"},
			Gap:         "So where do those bytes actually live?",
		},
		Beats: []SnippetBeat{
			{ID: "back", Heading: "Where we left off", Narration: brNarration, Bridge: &BridgeBeat{Show: "back"}},
			{ID: "carry-bits", Heading: "Bits still count", Narration: brNarration, Bridge: &BridgeBeat{Show: "carry", At: 0}},
			{ID: "carry-bytes", Heading: "Bytes still count", Narration: brNarration, Bridge: &BridgeBeat{Show: "carry", At: 1}},
			{ID: "the-hole", Heading: "The open question", Narration: brNarration, Bridge: &BridgeBeat{Show: "gap"}},
			{ID: "ahead", Heading: "Memory and storage", Narration: brNarration, Bridge: &BridgeBeat{Show: "ahead"}},
		},
	}
	p.targetWords = 5 * 40
	return p
}

func TestBridgePlanAccepted(t *testing.T) {
	if err := validateBridgePlan(bridgePlan()); err != nil {
		t.Fatalf("a well-formed bridge plan was rejected: %v", err)
	}
}

func TestBridgeRejectsAnUnnamedSide(t *testing.T) {
	p := bridgePlan()
	p.Bridge.To = "  "
	err := validateBridgePlan(p)
	if err == nil {
		t.Fatal("a hand-off with no destination lesson was accepted")
	}
	if !strings.Contains(err.Error(), "both lessons") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBridgeRejectsASingleEstablishedItem(t *testing.T) {
	p := bridgePlan()
	p.Bridge.Established = p.Bridge.Established[:1]
	err := validateBridgePlan(p)
	if err == nil {
		t.Fatal("one established item was accepted, and one item is a fact rather than ground to stand on")
	}
	if !strings.Contains(err.Error(), "recap") {
		t.Fatalf("the error does not name the template that draws the other shape: %v", err)
	}
}

func TestBridgeRejectsARecapsWorthOfGround(t *testing.T) {
	p := bridgePlan()
	p.Bridge.Established = append(p.Bridge.Established, "encodings map bytes onto characters", "hex is four bits at a time")
	if err := validateBridgePlan(p); err == nil {
		t.Fatal("four established items were accepted, and four is a recap")
	}
}

func TestBridgeRejectsAMissingGap(t *testing.T) {
	p := bridgePlan()
	p.Bridge.Gap = ""
	err := validateBridgePlan(p)
	if err == nil {
		t.Fatal("a hand-off with no gap was accepted, and without one the clip is a recap with a chevron on it")
	}
	if !strings.Contains(err.Error(), "gap") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBridgeRequiresOpeningOnTheBackSide(t *testing.T) {
	p := bridgePlan()
	p.Beats[0].Bridge = &BridgeBeat{Show: "carry", At: 0}
	p.Beats[1].Bridge = &BridgeBeat{Show: "back"}
	if err := validateBridgePlan(p); err == nil {
		t.Fatal("a clip that ticks an item before showing the ground it stands on was accepted")
	}
}

func TestBridgeRejectsLookingBackTwice(t *testing.T) {
	p := bridgePlan()
	p.Beats[2].Bridge = &BridgeBeat{Show: "back"}
	err := validateBridgePlan(p)
	if err == nil {
		t.Fatal("a clip that walks back across the bridge was accepted")
	}
	if !strings.Contains(err.Error(), "walks the viewer back") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBridgeRejectsCarryingAnItemTwice(t *testing.T) {
	p := bridgePlan()
	p.Beats[2].Bridge = &BridgeBeat{Show: "carry", At: 0}
	err := validateBridgePlan(p)
	if err == nil {
		t.Fatal("an item ticked twice was accepted, and the second tick spends a beat on nothing")
	}
	if !strings.Contains(err.Error(), "second time") {
		t.Fatalf("the error does not name the repeat: %v", err)
	}
}

func TestBridgeRejectsACarryOffTheList(t *testing.T) {
	p := bridgePlan()
	p.Beats[1].Bridge = &BridgeBeat{Show: "carry", At: 9}
	if err := validateBridgePlan(p); err == nil {
		t.Fatal("a carry of an item that does not exist was accepted")
	}
}

func TestBridgeRejectsTwoGaps(t *testing.T) {
	p := bridgePlan()
	p.Beats[2].Bridge = &BridgeBeat{Show: "gap"}
	err := validateBridgePlan(p)
	if err == nil {
		t.Fatal("two gap beats were accepted, and a clip with two has not decided what the lesson is for")
	}
	if !strings.Contains(err.Error(), "want exactly 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBridgeRejectsNoGapAtAll(t *testing.T) {
	p := bridgePlan()
	p.Bridge.Established = append(p.Bridge.Established, "encodings map bytes onto characters")
	p.Beats[3].Bridge = &BridgeBeat{Show: "carry", At: 2}
	err := validateBridgePlan(p)
	if err == nil {
		t.Fatal("a hand-off with no gap beat was accepted")
	}
	if !strings.Contains(err.Error(), "want exactly 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBridgeRequiresAheadToBeTheLastBeat(t *testing.T) {
	p := bridgePlan()
	p.Beats[3].Bridge = &BridgeBeat{Show: "ahead"}
	p.Beats[4].Bridge = &BridgeBeat{Show: "gap"}
	err := validateBridgePlan(p)
	if err == nil {
		t.Fatal("a clip that arrives and then keeps talking was accepted")
	}
	if !strings.Contains(err.Error(), "wrong side of the bridge") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBridgeRejectsAnItemNeverCarried(t *testing.T) {
	p := bridgePlan()
	p.Bridge.Established = append(p.Bridge.Established, "encodings map bytes onto characters")
	err := validateBridgePlan(p)
	if err == nil {
		t.Fatal("an established item with no beat was accepted, and that is knowledge silently dropped")
	}
	if !strings.Contains(err.Error(), "never carried") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBridgeNormalizeClampsACarryOffTheList(t *testing.T) {
	p := bridgePlan()
	p.Beats[1].Bridge.At = 99
	p.Beats[3].Bridge.At = 7
	normalizeBridgePlan(p)
	if at := p.Beats[1].Bridge.At; at != len(p.Bridge.Established)-1 {
		t.Fatalf("want the carry clamped to the last item, got %d", at)
	}
	// A gap beat does not index anything, so any index it arrived with is noise.
	if at := p.Beats[3].Bridge.At; at != 0 {
		t.Fatalf("a gap beat kept its index %d", at)
	}
}

func TestBridgeNormalizeClampsTheWording(t *testing.T) {
	p := bridgePlan()
	p.Bridge.From = "the previous lesson about binary and data representation"
	p.Bridge.Gap = "so where exactly do all of those bytes end up living once the power is on"
	p.Bridge.Established[0] = "a bit is one single two way switch that is either on or off"
	normalizeBridgePlan(p)
	if n := len(strings.Fields(p.Bridge.From)); n != maxBridgeNameWords {
		t.Fatalf("the lesson name survived at %d words", n)
	}
	if n := len(strings.Fields(p.Bridge.Gap)); n != maxBridgeGapWords {
		t.Fatalf("the gap survived at %d words", n)
	}
	if n := len(strings.Fields(p.Bridge.Established[0])); n != maxBridgeItemWords {
		t.Fatalf("an established item survived at %d words", n)
	}
}

func TestBridgeNormalizeCapsAndDropsEmptyItems(t *testing.T) {
	p := bridgePlan()
	p.Bridge.Established = []string{"a bit is a switch", "   ", "a byte is eight of them", "hex groups four bits", "encodings map bytes to characters"}
	normalizeBridgePlan(p)
	if n := len(p.Bridge.Established); n != maxBridgeEstablished {
		t.Fatalf("want %d established items after normalize, got %d", maxBridgeEstablished, n)
	}
	for _, it := range p.Bridge.Established {
		if strings.TrimSpace(it) == "" {
			t.Fatal("an empty item survived normalize")
		}
	}
}

func TestBridgeShowDefaultsToCarry(t *testing.T) {
	b := BridgeBeat{Show: "wander"}
	if got := b.ResolvedShow(); got != "carry" {
		t.Fatalf("an unknown show resolved to %q, want carry", got)
	}
	b = BridgeBeat{Show: " AHEAD "}
	if got := b.ResolvedShow(); got != "ahead" {
		t.Fatalf("a shouted ahead resolved to %q", got)
	}
}

// The slot is what the clip is for: it is shut until the gap beat, and the
// arrival is the only thing that fills it.
func TestBridgeScenesOpenTheSlotAndFillItAtTheEnd(t *testing.T) {
	p := bridgePlan()
	scenes, err := bridgeScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if len(steps) != len(p.Beats) {
		t.Fatalf("want %d steps, got %d", len(p.Beats), len(steps))
	}

	first := steps[0]
	carried, _ := first["carried"].([]int)
	if first["show"] != "back" || len(carried) != 0 {
		t.Fatalf("the opening beat already carries items: %v", first)
	}
	if first["gapOpen"] != false || first["arrived"] != false {
		t.Fatalf("the opening beat has already opened the slot: %v", first)
	}

	last := steps[len(steps)-1]
	ticks, _ := last["carried"].([]int)
	if last["show"] != "ahead" || len(ticks) != len(p.Bridge.Established) {
		t.Fatalf("the arrival does not carry every established item: %v", last)
	}
	for i, at := range ticks {
		if at != i {
			t.Fatalf("the tick set is not sorted or complete: %v", ticks)
		}
	}
	if last["gapOpen"] != true || last["arrived"] != true {
		t.Fatalf("the arrival does not fill the slot: %v", last)
	}
}
