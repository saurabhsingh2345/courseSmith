package pipeline

import (
	"strings"
	"testing"
)

const rlNarration = "Each stage finishes its one job and then hands control to the next one along."

func relayPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "relay",
		Title:    "From power button to login prompt",
		Relay: &RelaySpec{
			Stages: []RelayStage{
				{Label: "power", Does: "steadies the rails and releases reset", Hands: "a running CPU"},
				{Label: "firmware", Does: "counts memory and finds a boot device", Hands: "the boot sector"},
				{Label: "bootloader", Does: "picks a kernel and loads it into memory", Hands: "control at the entry point"},
				{Label: "kernel", Does: "starts drivers and mounts the root filesystem", Hands: ""},
			},
		},
		Beats: []SnippetBeat{
			{ID: "line", Heading: "The whole run", Narration: rlNarration, Relay: &RelayBeat{Show: "line"}},
			{ID: "power", Heading: "Power", Narration: rlNarration, Relay: &RelayBeat{Show: "ignite", At: 0}},
			{ID: "firmware", Heading: "Firmware", Narration: rlNarration, Relay: &RelayBeat{Show: "ignite", At: 1}},
			{ID: "boot", Heading: "Bootloader", Narration: rlNarration, Relay: &RelayBeat{Show: "ignite", At: 2}},
			{ID: "kernel", Heading: "Kernel", Narration: rlNarration, Relay: &RelayBeat{Show: "ignite", At: 3}},
			{ID: "chain", Heading: "The baton", Narration: rlNarration, Relay: &RelayBeat{Show: "chain"}},
		},
	}
	// The template's ideal is 28 words per beat, so the fixture budget is sized
	// against that — nBeats * 40 would demand more beats than it has.
	p.targetWords = 6 * 28
	return p
}

func TestRelayPlanAccepted(t *testing.T) {
	if err := validateRelayPlan(relayPlan()); err != nil {
		t.Fatalf("a well-formed relay was rejected: %v", err)
	}
}

func TestRelayRejectsTooFewStages(t *testing.T) {
	p := relayPlan()
	p.Relay.Stages = p.Relay.Stages[:3]
	err := validateRelayPlan(p)
	if err == nil {
		t.Fatal("a three-stage chain was accepted, and three stages are a pair of steps rather than a sequence")
	}
	if !strings.Contains(err.Error(), "3 stages") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestRelayRejectsTooManyStages(t *testing.T) {
	p := relayPlan()
	for i := 0; i < 5; i++ {
		p.Relay.Stages = append(p.Relay.Stages, RelayStage{Label: "extra", Does: "does something", Hands: "something"})
	}
	if err := validateRelayPlan(p); err == nil {
		t.Fatal("a nine-stage chain was accepted, and the capsules would shrink below a readable label")
	}
}

// A leg that hands nothing on is not a leg, it is a terminus.
func TestRelayRejectsAStageThatHandsNothingOn(t *testing.T) {
	p := relayPlan()
	p.Relay.Stages[1].Hands = ""
	err := validateRelayPlan(p)
	if err == nil {
		t.Fatal("a middle stage with no hand-off was accepted")
	}
	if !strings.Contains(err.Error(), "firmware") || !strings.Contains(err.Error(), "bootloader") {
		t.Fatalf("the error does not name both ends of the missing hand-off: %v", err)
	}
}

func TestRelayAllowsTheLastStageToHandNothingOn(t *testing.T) {
	p := relayPlan()
	if p.Relay.Stages[3].Hands != "" {
		t.Fatal("the fixture no longer exercises the final-stage exemption")
	}
	if err := validateRelayPlan(p); err != nil {
		t.Fatalf("the last stage was required to hand something on: %v", err)
	}
}

func TestRelayRejectsACapsuleThatIsADescription(t *testing.T) {
	p := relayPlan()
	p.Relay.Stages[2].Label = "the second stage bootloader program"
	if err := validateRelayPlan(p); err == nil {
		t.Fatal("a five-word capsule label was accepted")
	}
}

func TestRelayRequiresOpeningOnTheLine(t *testing.T) {
	p := relayPlan()
	p.Beats[0].Relay = &RelayBeat{Show: "ignite", At: 0}
	p.Beats[1].Relay = &RelayBeat{Show: "line"}
	if err := validateRelayPlan(p); err == nil {
		t.Fatal("a spark arriving before the line was drawn was accepted")
	}
}

func TestRelayRequiresClosingOnTheChain(t *testing.T) {
	p := relayPlan()
	p.Beats[len(p.Beats)-1].Relay = &RelayBeat{Show: "ignite", At: 3}
	if err := validateRelayPlan(p); err == nil {
		t.Fatal("a clip that never shows the whole chain was accepted")
	}
}

// The order IS the template, so a skip is rejected by NAME.
func TestRelayRejectsASkipAndNamesTheStageThatWasSkipped(t *testing.T) {
	p := relayPlan()
	p.Beats[3].Relay = &RelayBeat{Show: "ignite", At: 3}
	p.Beats[4].Relay = &RelayBeat{Show: "ignite", At: 3}
	err := validateRelayPlan(p)
	if err == nil {
		t.Fatal("a spark that jumped a stage was accepted")
	}
	if !strings.Contains(err.Error(), "bootloader") {
		t.Fatalf("the error does not name the skipped stage: %v", err)
	}
}

func TestRelayRejectsABatonThatGoesBackwards(t *testing.T) {
	p := relayPlan()
	p.Beats[4].Relay = &RelayBeat{Show: "ignite", At: 1}
	err := validateRelayPlan(p)
	if err == nil {
		t.Fatal("a spark travelling backwards was accepted")
	}
	if !strings.Contains(err.Error(), "firmware") || !strings.Contains(err.Error(), "bootloader") {
		t.Fatalf("the error does not name both stages: %v", err)
	}
}

func TestRelayRejectsARunThatStartsPastTheFirstStage(t *testing.T) {
	p := relayPlan()
	p.Beats[1].Relay = &RelayBeat{Show: "ignite", At: 1}
	err := validateRelayPlan(p)
	if err == nil {
		t.Fatal("a run whose first spark skipped stage zero was accepted")
	}
	if !strings.Contains(err.Error(), "power") {
		t.Fatalf("the error does not name the stage that never fires: %v", err)
	}
}

func TestRelayRejectsFewerThanThreeIgnites(t *testing.T) {
	p := relayPlan()
	p.Beats = []SnippetBeat{
		p.Beats[0],
		p.Beats[1],
		p.Beats[2],
		p.Beats[len(p.Beats)-1],
	}
	p.targetWords = 4 * 28
	err := validateRelayPlan(p)
	if err == nil {
		t.Fatal("a chain with two hand-offs was accepted, and two lit capsules is a pair rather than a chain")
	}
	if !strings.Contains(err.Error(), "2 stages ignite") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestRelayNormalizeClampsLabelsAndDropsNamelessStages(t *testing.T) {
	p := relayPlan()
	p.Relay.Stages = append(p.Relay.Stages, RelayStage{Label: "", Does: "nothing at all"})
	p.Relay.Stages[0].Label = "  the   power   supply   rails  "
	p.Beats[4].Relay.At = 12
	normalizeRelayPlan(p)

	if len(p.Relay.Stages) != 4 {
		t.Fatalf("the nameless stage survived normalize: %v", p.Relay.Stages)
	}
	if p.Relay.Stages[0].Label != "the power supply" {
		t.Fatalf("the label normalized to %q", p.Relay.Stages[0].Label)
	}
	if p.Beats[4].Relay.At != 3 {
		t.Fatalf("an index past the end clamped to %d, want 3", p.Beats[4].Relay.At)
	}
	if err := validateRelayPlan(p); err != nil {
		t.Fatalf("a repairable chain was rejected after normalize: %v", err)
	}
}

func TestRelayShowDefaultsToIgnite(t *testing.T) {
	b := RelayBeat{Show: "sparkle"}
	if got := b.ResolvedShow(); got != "ignite" {
		t.Fatalf("an unknown show resolved to %q, want ignite", got)
	}
}

// The component animates a spark between two named capsules; where it comes
// from and which capsules are already lit arrive precomputed.
func TestRelayScenesAccumulateTheLitStages(t *testing.T) {
	p := relayPlan()
	scenes, err := relayScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	stages, _ := props["stages"].([]map[string]any)
	if len(stages) != 4 {
		t.Fatalf("want 4 capsules, got %d", len(stages))
	}
	if stages[0]["label"] != "power" || stages[0]["hands"] != "a running CPU" {
		t.Fatalf("the first capsule is wrong: %v", stages[0])
	}

	steps, _ := props["steps"].([]map[string]any)
	if steps[0]["show"] != "line" {
		t.Fatalf("first step shows %v, want line", steps[0]["show"])
	}
	if lit, _ := steps[0]["lit"].([]int); len(lit) != 0 {
		t.Fatalf("the opener already has lit capsules: %v", lit)
	}
	if _, ok := steps[1]["from"]; ok {
		t.Fatal("the first ignite claims a spark origin, and nothing hands the baton to stage zero")
	}
	if steps[2]["from"] != 0 || steps[2]["at"] != 1 {
		t.Fatalf("the second ignite does not travel from stage zero: %v", steps[2])
	}

	last := steps[len(steps)-1]
	if last["show"] != "chain" {
		t.Fatalf("last step shows %v, want chain", last["show"])
	}
	lit, _ := last["lit"].([]int)
	if len(lit) != 4 || lit[0] != 0 || lit[3] != 3 {
		t.Fatalf("the closer does not light the whole chain: %v", lit)
	}
}
