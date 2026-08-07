package pipeline

import (
	"strings"
	"testing"
)

const statesNarration = "A process is only ever in one of these at a time, and something named has to move it."

func statesPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "states",
		Title:    "One state at a time",
		States: &StatesSpec{
			Nodes: []StatesNode{
				{ID: "new", Label: "new"},
				{ID: "ready", Label: "ready"},
				{ID: "running", Label: "running"},
				{ID: "waiting", Label: "waiting"},
				{ID: "terminated", Label: "terminated"},
			},
			Arcs: []StatesArc{
				{From: "new", To: "ready", On: "the loader admits it"},
				{From: "ready", To: "running", On: "the scheduler picks it"},
				{From: "running", To: "waiting", On: "it asks for input"},
				{From: "waiting", To: "ready", On: "the input arrives"},
				{From: "running", To: "terminated", On: "it returns from main"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "machine", Heading: "The whole machine", Narration: statesNarration, States: &StatesBeat{Show: "machine", At: 0}},
			{ID: "admit", Heading: "Admitted", Narration: statesNarration, States: &StatesBeat{Show: "fire", At: 0}},
			{ID: "dispatch", Heading: "Picked", Narration: statesNarration, States: &StatesBeat{Show: "fire", At: 1}},
			{ID: "on-cpu", Heading: "On the CPU", Narration: statesNarration, States: &StatesBeat{Show: "state", At: 2}},
			{ID: "block", Heading: "Blocked", Narration: statesNarration, States: &StatesBeat{Show: "fire", At: 2}},
			{ID: "wake", Heading: "Woken", Narration: statesNarration, States: &StatesBeat{Show: "fire", At: 3}},
			{ID: "walk", Heading: "The route", Narration: statesNarration, States: &StatesBeat{Show: "walk"}},
		},
	}
	// A beat here is a shot, so the fixture budget is sized at the template's
	// own 28-word ideal — nBeats * 40 would make beatBounds demand more beats
	// than the fixture has.
	p.targetWords = 7 * 28
	return p
}

func TestStatesPlanAccepted(t *testing.T) {
	if err := validateStatesPlan(statesPlan()); err != nil {
		t.Fatalf("a well-formed states plan was rejected: %v", err)
	}
}

// The family's signature rule: the token's position is walked in Go, and a
// transition that does not start where the dot is standing is rejected with
// both the arc's origin and the token's real state quoted.
func TestStatesRejectsATeleportingToken(t *testing.T) {
	p := statesPlan()
	// The token is on "ready" here, and this arc leaves "waiting".
	p.Beats[2].States = &StatesBeat{Show: "fire", At: 3}
	err := validateStatesPlan(p)
	if err == nil {
		t.Fatal("a token that jumps a gap with no arrow across it was accepted")
	}
	if !strings.Contains(err.Error(), "waiting") || !strings.Contains(err.Error(), "ready") {
		t.Fatalf("the error does not quote the arc and where the token actually sits: %v", err)
	}
	if !strings.Contains(err.Error(), "telling a lie") {
		t.Fatalf("the error does not say why it matters: %v", err)
	}
}

func TestStatesRejectsAnArcToAStateThatDoesNotExist(t *testing.T) {
	p := statesPlan()
	p.States.Arcs[1].To = "sleeping"
	err := validateStatesPlan(p)
	if err == nil {
		t.Fatal("an arrow to a state that is not on screen was accepted")
	}
	if !strings.Contains(err.Error(), "sleeping") {
		t.Fatalf("the error does not quote the bad endpoint: %v", err)
	}
}

func TestStatesRejectsAnArcFromAStateThatDoesNotExist(t *testing.T) {
	p := statesPlan()
	p.States.Arcs[0].From = "ghost"
	err := validateStatesPlan(p)
	if err == nil {
		t.Fatal("an arrow out of a state that is not on screen was accepted")
	}
	if !strings.Contains(err.Error(), "ghost") {
		t.Fatalf("the error does not quote the bad endpoint: %v", err)
	}
}

func TestStatesRejectsADuplicateTransition(t *testing.T) {
	p := statesPlan()
	p.States.Arcs = append(p.States.Arcs, StatesArc{From: "new", To: "ready", On: "the loader admits it"})
	err := validateStatesPlan(p)
	if err == nil {
		t.Fatal("the same transition drawn twice was accepted")
	}
	if !strings.Contains(err.Error(), "repeats") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatesRejectsADuplicateStateID(t *testing.T) {
	p := statesPlan()
	p.States.Nodes[1].ID = "new"
	err := validateStatesPlan(p)
	if err == nil {
		t.Fatal("two states sharing an id were accepted")
	}
	if !strings.Contains(err.Error(), "repeats the id") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatesRejectsAnUnlabelledTransition(t *testing.T) {
	p := statesPlan()
	p.States.Arcs[2].On = ""
	err := validateStatesPlan(p)
	if err == nil {
		t.Fatal("an arrow with no event on it was accepted")
	}
	if !strings.Contains(err.Error(), "no event") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatesRequiresOpeningOnTheMachine(t *testing.T) {
	p := statesPlan()
	p.Beats[0].States = &StatesBeat{Show: "fire", At: 0}
	err := validateStatesPlan(p)
	if err == nil {
		t.Fatal("a clip that slides the token before showing the graph was accepted")
	}
	if !strings.Contains(err.Error(), "open on the whole machine") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatesRequiresClosingOnTheWalk(t *testing.T) {
	p := statesPlan()
	p.Beats[6].States = &StatesBeat{Show: "state", At: 1}
	err := validateStatesPlan(p)
	if err == nil {
		t.Fatal("a clip that does not close on the route was accepted")
	}
	if !strings.Contains(err.Error(), "close on the route") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatesRejectsAWalkBeforeTheEnd(t *testing.T) {
	p := statesPlan()
	p.Beats[1].States = &StatesBeat{Show: "walk"}
	err := validateStatesPlan(p)
	if err == nil {
		t.Fatal("the route drawn before the walk finished was accepted")
	}
	if !strings.Contains(err.Error(), "closer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatesRejectsATokenStartingNowhere(t *testing.T) {
	p := statesPlan()
	p.Beats[0].States.At = 9
	err := validateStatesPlan(p)
	if err == nil {
		t.Fatal("a token placed on a state that does not exist was accepted")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatesRejectsTooFewStates(t *testing.T) {
	p := statesPlan()
	p.States.Nodes = p.States.Nodes[:2]
	if err := validateStatesPlan(p); err == nil {
		t.Fatal("a two-state machine was accepted")
	}
}

func TestStatesNormalizeSlugifiesIdentifiers(t *testing.T) {
	p := statesPlan()
	p.States.Nodes[3].ID = "Waiting For IO"
	p.States.Arcs[2].To = "Waiting For IO"
	p.States.Arcs[3].From = "Waiting For IO"
	normalizeStatesPlan(p)
	if got := p.States.Nodes[3].ID; got != "waiting-for-io" {
		t.Fatalf("the state id normalized to %q, want waiting-for-io", got)
	}
	if got := p.States.Arcs[2].To; got != "waiting-for-io" {
		t.Fatalf("the arc endpoint normalized to %q, want waiting-for-io", got)
	}
	if err := validateStatesPlan(p); err != nil {
		t.Fatalf("a sloppy-but-correct plan was rejected after normalize: %v", err)
	}
}

func TestStatesNormalizeClampsAnOutOfRangeFire(t *testing.T) {
	p := statesPlan()
	p.Beats[1].States.At = 99
	normalizeStatesPlan(p)
	if got := p.Beats[1].States.At; got != 4 {
		t.Fatalf("an out-of-range transition clamped to %d, want 4", got)
	}
}

func TestStatesShowDefaultsToFire(t *testing.T) {
	b := StatesBeat{Show: "sparkle"}
	if got := b.ResolvedShow(); got != "fire" {
		t.Fatalf("an unknown show resolved to %q, want fire", got)
	}
}

// The token's node is precomputed per step, so the renderer never replays the
// beat list to find out where the dot is.
func TestStatesScenesWalkTheToken(t *testing.T) {
	p := statesPlan()
	scenes, err := statesScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	arcs, _ := props["arcs"].([]map[string]any)
	if len(arcs) != 5 {
		t.Fatalf("want 5 transitions, got %d", len(arcs))
	}
	// The endpoints arrive as indices: the component never looks up an id.
	if arcs[1]["from"] != 1 || arcs[1]["to"] != 2 {
		t.Fatalf("transition 1 resolved to %v, want from 1 to 2", arcs[1])
	}

	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != 7 {
		t.Fatalf("want 7 steps, got %d", len(steps))
	}
	first := steps[0]
	if first["token"] != 0 || first["from"] != 0 {
		t.Fatalf("the opening beat puts the token at %v (from %v), want 0", first["token"], first["from"])
	}
	if lit, _ := first["lit"].([]int); len(lit) != 0 {
		t.Fatalf("the opening beat already has lit transitions: %v", lit)
	}
	if steps[1]["from"] != 0 || steps[1]["token"] != 1 {
		t.Fatalf("the first fire slides from %v to %v, want 0 to 1", steps[1]["from"], steps[1]["token"])
	}

	last := steps[len(steps)-1]
	// new, ready, running, waiting, back to ready.
	if last["token"] != 1 {
		t.Fatalf("the walk ends with the token on %v, want 1", last["token"])
	}
	lit, _ := last["lit"].([]int)
	if len(lit) != 4 || lit[0] != 0 || lit[3] != 3 {
		t.Fatalf("the closing beat's lit set is %v, want [0 1 2 3]", lit)
	}
	if last["show"] != "walk" {
		t.Fatalf("the last step shows %v, want walk", last["show"])
	}
}
