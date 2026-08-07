package pipeline

import (
	"strings"
	"testing"
)

const stackNarration = "The plate lifts onto the pile with its own argument, and it waits there for the call above it."

func callStackPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "callstack",
		Title:    "The stack breathes in and out",
		CallStack: &CallStackSpec{
			Fn:   "factorial",
			Base: "zero is where it stops and answers one",
			Frames: []CallStackFrame{
				{Args: "n=3", Returns: "6"},
				{Args: "n=2", Returns: "2"},
				{Args: "n=1", Returns: "1"},
				{Args: "n=0", Returns: "1"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "call-3", Heading: "The first call", Narration: stackNarration, CallStack: &CallStackBeat{Show: "call", At: 0}},
			{ID: "call-2", Heading: "Deeper", Narration: stackNarration, CallStack: &CallStackBeat{Show: "call", At: 1}},
			{ID: "call-1", Heading: "Deeper still", Narration: stackNarration, CallStack: &CallStackBeat{Show: "call", At: 2}},
			{ID: "call-0", Heading: "The last call", Narration: stackNarration, CallStack: &CallStackBeat{Show: "call", At: 3}},
			{ID: "the-floor", Heading: "The base case", Narration: stackNarration, CallStack: &CallStackBeat{Show: "base", At: 3}},
			{ID: "back-0", Heading: "One comes back", Narration: stackNarration, CallStack: &CallStackBeat{Show: "return", At: 3}},
			{ID: "back-1", Heading: "And again", Narration: stackNarration, CallStack: &CallStackBeat{Show: "return", At: 2}},
			{ID: "back-2", Heading: "Two falls down", Narration: stackNarration, CallStack: &CallStackBeat{Show: "return", At: 1}},
			{ID: "back-3", Heading: "The outermost frame", Narration: stackNarration, CallStack: &CallStackBeat{Show: "return", At: 0}},
			{ID: "the-answer", Heading: "Nothing left", Narration: stackNarration, CallStack: &CallStackBeat{Show: "empty"}},
		},
	}
	// The template's ideal is 28 words per beat, so the fixture budget is sized
	// against that rather than the shared 40.
	p.targetWords = 10 * 28
	return p
}

func TestCallStackPlanAccepted(t *testing.T) {
	if err := validateCallStackPlan(callStackPlan()); err != nil {
		t.Fatalf("a well-formed call stack plan was rejected: %v", err)
	}
}

// The family's signature rule, applied to a stack: the plan is EXECUTED in Go,
// and a pop out of order is rejected with both frames quoted by their args.
func TestCallStackRejectsAReturnOutOfOrder(t *testing.T) {
	p := callStackPlan()
	p.Beats[5].CallStack = &CallStackBeat{Show: "return", At: 2}
	err := validateCallStackPlan(p)
	if err == nil {
		t.Fatal("a frame returning while another sat above it was accepted, which draws a queue and calls it a stack")
	}
	if !strings.Contains(err.Error(), "n=1") || !strings.Contains(err.Error(), "n=0") {
		t.Fatalf("the error does not quote both frames: %v", err)
	}
	if !strings.Contains(err.Error(), "above it") {
		t.Fatalf("the error does not say which frame is in the way: %v", err)
	}
}

func TestCallStackRejectsAPushOutOfOrder(t *testing.T) {
	p := callStackPlan()
	p.Beats[1].CallStack = &CallStackBeat{Show: "call", At: 2}
	err := validateCallStackPlan(p)
	if err == nil {
		t.Fatal("a call that skipped a frame was accepted")
	}
	if !strings.Contains(err.Error(), "n=1") || !strings.Contains(err.Error(), "n=2") {
		t.Fatalf("the error does not quote the frame pushed and the frame expected: %v", err)
	}
}

func TestCallStackRejectsTheBaseCaseLandingEarly(t *testing.T) {
	p := callStackPlan()
	p.Beats[3].CallStack = &CallStackBeat{Show: "base", At: 3}
	err := validateCallStackPlan(p)
	if err == nil {
		t.Fatal("a base case landing before every frame was pushed was accepted")
	}
	if !strings.Contains(err.Error(), "3 of 4") {
		t.Fatalf("the error does not count the frames that are up: %v", err)
	}
}

func TestCallStackRejectsTheBaseCaseOnANonDeepestFrame(t *testing.T) {
	p := callStackPlan()
	p.Beats[4].CallStack = &CallStackBeat{Show: "base", At: 2}
	err := validateCallStackPlan(p)
	if err == nil {
		t.Fatal("a base case pointed at a frame that is not the deepest was accepted")
	}
	if !strings.Contains(err.Error(), "n=1") || !strings.Contains(err.Error(), "n=0") {
		t.Fatalf("the error does not quote the claimed frame and the real one: %v", err)
	}
}

func TestCallStackRejectsAnEmptyStackWithFramesStillLive(t *testing.T) {
	p := callStackPlan()
	p.Beats[8].CallStack = &CallStackBeat{Show: "base", At: 3}
	err := validateCallStackPlan(p)
	if err == nil {
		t.Fatal("an empty-stack closer with a frame still on the stack was accepted")
	}
	if !strings.Contains(err.Error(), "n=3") {
		t.Fatalf("the error does not name the frame still live: %v", err)
	}
}

func TestCallStackRequiresClosingOnEmpty(t *testing.T) {
	p := callStackPlan()
	p.Beats[9].CallStack = &CallStackBeat{Show: "return", At: 0}
	err := validateCallStackPlan(p)
	if err == nil {
		t.Fatal("a clip that never shows the stack gone was accepted")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCallStackRequiresABaseBeat(t *testing.T) {
	p := callStackPlan()
	p.Beats = append(p.Beats[:4], p.Beats[5:]...)
	p.targetWords = 9 * 28
	err := validateCallStackPlan(p)
	if err == nil {
		t.Fatal("a recursion with no base case beat was accepted, which is a descent with no floor")
	}
	if !strings.Contains(err.Error(), "zero is where it stops") {
		t.Fatalf("the error does not quote the base case the plan already wrote: %v", err)
	}
}

func TestCallStackRejectsAFrameCountOutOfRange(t *testing.T) {
	p := callStackPlan()
	p.CallStack.Frames = p.CallStack.Frames[:1]
	if err := validateCallStackPlan(p); err == nil {
		t.Fatal("a single-frame recursion was accepted, and one call is not a descent")
	}
}

func TestCallStackRejectsAFrameWithNoReturnValue(t *testing.T) {
	p := callStackPlan()
	p.CallStack.Frames[2].Returns = ""
	err := validateCallStackPlan(p)
	if err == nil {
		t.Fatal("a frame with nothing to hand down was accepted")
	}
	if !strings.Contains(err.Error(), "n=1") {
		t.Fatalf("the error does not name the frame: %v", err)
	}
}

// Over-long labels and a stray index are phrasing and bookkeeping, not wrong
// answers, so they are repaired rather than argued with.
func TestCallStackNormalizeRepairsLabelsAndIndices(t *testing.T) {
	p := callStackPlan()
	p.CallStack.Fn = "  the   factorial   function  "
	p.CallStack.Frames[0].Args = "n=3141592653589"
	p.CallStack.Base = "zero is the floor and it returns one without calling itself again ever"
	p.Beats[4].CallStack.At = 0
	p.Beats[0].CallStack.At = 99
	normalizeCallStackPlan(p)

	if p.CallStack.Fn != "the factorial" {
		t.Fatalf("fn normalized to %q, want the two-word clamp", p.CallStack.Fn)
	}
	if len(p.CallStack.Frames[0].Args) != maxCallStackArgChars {
		t.Fatalf("args normalized to %q, want %d characters", p.CallStack.Frames[0].Args, maxCallStackArgChars)
	}
	if n := len(strings.Fields(p.CallStack.Base)); n != maxCallStackBaseWords {
		t.Fatalf("base normalized to %d words, want %d", n, maxCallStackBaseWords)
	}
	if p.Beats[4].CallStack.At != 3 {
		t.Fatalf("the base beat kept index %d; the base case is the deepest frame by definition", p.Beats[4].CallStack.At)
	}
	if p.Beats[0].CallStack.At != 3 {
		t.Fatalf("an out-of-range call index normalized to %d, want the last frame", p.Beats[0].CallStack.At)
	}
}

func TestCallStackShowDefaultsToCall(t *testing.T) {
	b := CallStackBeat{Show: "unwind"}
	if got := b.ResolvedShow(); got != "call" {
		t.Fatalf("an unknown show resolved to %q, want call", got)
	}
}

// The component never reasons about push and pop order: every step arrives
// carrying the exact set of live frames and the exact set that have returned.
func TestCallStackScenesAccumulateTheStack(t *testing.T) {
	p := callStackPlan()
	scenes, err := callStackScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	if props["answer"] != "6" {
		t.Fatalf("the clip ends on %v, want the outermost frame's return value", props["answer"])
	}
	frames, _ := props["frames"].([]map[string]any)
	if len(frames) != 4 {
		t.Fatalf("want 4 frames, got %d", len(frames))
	}
	if frames[3]["base"] != true || frames[0]["base"] != false {
		t.Fatalf("the base flag is on the wrong plate: %v", frames)
	}

	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != 10 {
		t.Fatalf("want 10 steps, got %d", len(steps))
	}

	first := steps[0]
	if first["show"] != "call" || first["at"] != 0 {
		t.Fatalf("the first step is wrong: %v", first)
	}
	if live, _ := first["onStack"].([]int); len(live) != 1 || live[0] != 0 {
		t.Fatalf("after the first call the stack holds %v, want just frame 0", first["onStack"])
	}
	if done, _ := first["returned"].([]int); len(done) != 0 {
		t.Fatalf("nothing has returned yet, but the first step reports %v", done)
	}

	// The first pop carries its value and the plate it drops into.
	pop := steps[5]
	if pop["value"] != "1" {
		t.Fatalf("the first return carries %v, want the base case's value", pop["value"])
	}
	if pop["into"] != 2 {
		t.Fatalf("the first return drops into %v, want the frame below it", pop["into"])
	}

	last := steps[len(steps)-1]
	if last["show"] != "empty" {
		t.Fatalf("the last step shows %v, want empty", last["show"])
	}
	if live, _ := last["onStack"].([]int); len(live) != 0 {
		t.Fatalf("the stack is not gone at the end: %v", live)
	}
	done, _ := last["returned"].([]int)
	if len(done) != 4 || done[0] != 0 || done[3] != 3 {
		t.Fatalf("the returned set at the end is %v, want every frame sorted", done)
	}
}
