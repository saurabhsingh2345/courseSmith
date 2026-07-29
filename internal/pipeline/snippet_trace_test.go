package pipeline

import (
	"strings"
	"testing"
)

const trNarration = "Both of them read one in stock, and neither has written anything back yet."

func tracePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "trace",
		Title:    "Two users, one item left",
		Trace: &TraceSpec{
			Actors:   []string{"User A", "User B"},
			Resource: "Inventory",
			Start:    "1",
			Steps: []TraceStepSpec{
				{By: 0, Op: "read inv", Becomes: "1", Note: "A reads one in stock"},
				{By: 1, Op: "read inv", Becomes: "1", Note: "B reads the same one"},
				{By: 0, Op: "write 0", Becomes: "0", Note: "A takes the item"},
				{By: 1, Op: "write 0", Becomes: "0", Note: "B writes from a stale read"},
			},
			Outcome: "Two customers, one item, both charged",
			Broken:  true,
		},
		Beats: []SnippetBeat{
			{ID: "setup", Heading: "One item left", Narration: trNarration, Trace: &TraceBeat{Show: "setup"}},
			{ID: "both", Heading: "Both arrive", Narration: trNarration, Trace: &TraceBeat{Show: "queue"}},
			{ID: "a-reads", Heading: "A reads", Narration: trNarration, Trace: &TraceBeat{Show: "step", At: 0}},
			{ID: "b-reads", Heading: "B reads", Narration: trNarration, Trace: &TraceBeat{Show: "step", At: 1}},
			{ID: "a-writes", Heading: "A writes", Narration: trNarration, Trace: &TraceBeat{Show: "step", At: 2}},
			{ID: "b-writes", Heading: "B writes", Narration: trNarration, Trace: &TraceBeat{Show: "step", At: 3}},
			{ID: "damage", Heading: "The damage", Narration: trNarration, Trace: &TraceBeat{Show: "outcome"}},
		},
	}
	p.targetWords = 7 * 40
	return p
}

func TestTracePlanAccepted(t *testing.T) {
	if err := validateTracePlan(tracePlan()); err != nil {
		t.Fatalf("a well-formed trace was rejected: %v", err)
	}
}

// The rule this template exists for: the state has to add up, which starts with
// every step saying what the value becomes.
func TestTraceEveryStepStatesTheValue(t *testing.T) {
	p := tracePlan()
	p.Trace.Steps[2].Becomes = ""
	err := validateTracePlan(p)
	if err == nil {
		t.Fatal("a step that does not say what the value becomes was accepted")
	}
	if !strings.Contains(err.Error(), "repeats the previous value") {
		t.Errorf("the error does not explain the no-op case: %v", err)
	}
}

// A value that never moves means nothing is contended for, which makes this the
// wrong template rather than a badly written one.
func TestTraceRejectsUnchangingState(t *testing.T) {
	p := tracePlan()
	for i := range p.Trace.Steps {
		p.Trace.Steps[i].Becomes = "1"
	}
	err := validateTracePlan(p)
	if err == nil {
		t.Fatal("a trace whose value never changes was accepted")
	}
	if !strings.Contains(err.Error(), "flow template") {
		t.Errorf("the error does not point at the right template: %v", err)
	}
}

// An actor drawn but never acting is a column of empty space implying a
// participant who is not in the story.
func TestTraceEveryActorActs(t *testing.T) {
	p := tracePlan()
	p.Trace.Actors = append(p.Trace.Actors, "User C")
	err := validateTracePlan(p)
	if err == nil {
		t.Fatal("an actor who issues nothing was accepted")
	}
	if !strings.Contains(err.Error(), "User C") {
		t.Errorf("the error does not name the idle actor: %v", err)
	}
}

func TestTraceOpensOnSetup(t *testing.T) {
	p := tracePlan()
	p.Beats[0].Trace = &TraceBeat{Show: "queue"}
	p.Beats[1].Trace = &TraceBeat{Show: "setup"}
	err := validateTracePlan(p)
	if err == nil {
		t.Fatal("a clip that queued before showing the value at rest was accepted")
	}
	if !strings.Contains(err.Error(), "change from nothing") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTraceEndsOnOutcome(t *testing.T) {
	p := tracePlan()
	p.Beats[1].Trace = &TraceBeat{Show: "outcome"}
	if err := validateTracePlan(p); err == nil {
		t.Fatal("an outcome with the clip carrying on afterwards was accepted")
	}
}

func TestTraceDrainsEveryOperation(t *testing.T) {
	p := tracePlan()
	// Re-point rather than delete, so the shared beat floor does not fire.
	p.Beats[5].Trace = &TraceBeat{Show: "queue"}
	err := validateTracePlan(p)
	if err == nil {
		t.Fatal("an operation nobody narrates was accepted")
	}
	if !strings.Contains(err.Error(), "cannot account for") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTraceRejectsUnknownActor(t *testing.T) {
	p := tracePlan()
	p.Trace.Steps[1].By = 7
	// Normalize would clamp it, so this asserts the validator's own guard for
	// a plan built by hand or arriving past the repair pass.
	if err := validateTracePlan(p); err == nil {
		t.Fatal("an operation from an actor who does not exist was accepted")
	}
}

func TestTraceNormalizeRepairs(t *testing.T) {
	p := tracePlan()
	p.Trace.Steps[1].By = 7
	p.Beats[3].Trace.At = 99
	p.Beats[4].Trace.Show = "nonsense"
	p.Trace.Start = strings.Repeat("9", 40)
	normalizeTracePlan(p)

	if p.Trace.Steps[1].By != len(p.Trace.Actors)-1 {
		t.Errorf("an out-of-range actor became %d", p.Trace.Steps[1].By)
	}
	if p.Beats[3].Trace.At != len(p.Trace.Steps)-1 {
		t.Errorf("an out-of-range beat points at %d", p.Beats[3].Trace.At)
	}
	if p.Beats[4].Trace.Show != "step" {
		t.Errorf("an unknown show became %q, want step", p.Beats[4].Trace.Show)
	}
	if n := len([]rune(p.Trace.Start)); n > maxTraceValueChars {
		t.Errorf("the start value is %d chars after normalize, want <= %d", n, maxTraceValueChars)
	}
}

func TestTraceScenesMarkNoOps(t *testing.T) {
	p := tracePlan()
	scenes, err := traceScenes(sceneInput(t, p, 6000))
	if err != nil {
		t.Fatalf("traceScenes: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneTrace {
		t.Fatalf("want one trace scene, got %d of %q", len(scenes), scenes[0].Type)
	}
	ops := scenes[0].Props["ops"].([]map[string]any)
	if len(ops) != 4 {
		t.Fatalf("want four operations on the scene, got %d", len(ops))
	}
	// Whether an operation moved the value is decided in Go, so the renderer's
	// "no change" mark can never disagree with the state chain. Both reads and
	// B's stale write are no-ops; only A's write moves it.
	want := []bool{false, false, true, false}
	for i, w := range want {
		if ops[i]["changes"] != w {
			t.Errorf("op %d changes = %v, want %v — the no-op mark is what makes the bug visible", i, ops[i]["changes"], w)
		}
	}
}
