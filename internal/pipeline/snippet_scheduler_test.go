package pipeline

import (
	"strings"
	"testing"
)

const schedNarration = "Only one of these lanes can be filled at any instant, and the rest are simply waiting their turn."

func schedulerPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "scheduler",
		Title:    "Three programs, one processor",
		Scheduler: &SchedulerSpec{
			Policy: "Round Robin",
			Procs: []SchedulerProc{
				{Label: "P1"},
				{Label: "P2"},
				{Label: "P3"},
			},
			Slots: []SchedulerSlot{
				{Proc: 0, Len: 2},
				{Proc: 1, Len: 2},
				{Proc: 2, Len: 2},
				{Proc: 0, Len: 2},
				{Proc: 1, Len: 1},
				{Proc: 2, Len: 1},
			},
		},
		Beats: []SnippetBeat{
			{ID: "queue", Heading: "Three waiting", Narration: schedNarration, Scheduler: &SchedulerBeat{Show: "queue"}},
			{ID: "first", Heading: "P1 runs", Narration: schedNarration, Scheduler: &SchedulerBeat{Show: "run", At: 0}},
			{ID: "second", Heading: "P2 runs", Narration: schedNarration, Scheduler: &SchedulerBeat{Show: "run", At: 1}},
			{ID: "switch", Heading: "The changeover", Narration: schedNarration, Scheduler: &SchedulerBeat{Show: "switch", At: 1}},
			{ID: "third", Heading: "P3 runs", Narration: schedNarration, Scheduler: &SchedulerBeat{Show: "run", At: 2}},
			{ID: "round-two", Heading: "Round again", Narration: schedNarration, Scheduler: &SchedulerBeat{Show: "run", At: 3}},
			{ID: "fair", Heading: "The tally", Narration: schedNarration, Scheduler: &SchedulerBeat{Show: "fair"}},
		},
	}
	// A beat here is a shot, so the fixture budget is sized at the template's
	// own 28-word ideal — nBeats * 40 would make beatBounds demand more beats
	// than the fixture has.
	p.targetWords = 7 * 28
	return p
}

func TestSchedulerPlanAccepted(t *testing.T) {
	if err := validateSchedulerPlan(schedulerPlan()); err != nil {
		t.Fatalf("a well-formed scheduler plan was rejected: %v", err)
	}
}

// The family's signature rule: the axis is summed in Go, and a schedule longer
// than the chart can draw is rejected with both numbers quoted.
func TestSchedulerRejectsAScheduleLongerThanTheAxis(t *testing.T) {
	p := schedulerPlan()
	for i := range p.Scheduler.Slots {
		p.Scheduler.Slots[i].Len = 6
	}
	err := validateSchedulerPlan(p)
	if err == nil {
		t.Fatal("a thirty-six unit schedule was accepted onto a twenty-four unit axis")
	}
	if !strings.Contains(err.Error(), "36") || !strings.Contains(err.Error(), "24") {
		t.Fatalf("the error does not quote the real total and the limit: %v", err)
	}
}

func TestSchedulerRejectsAProcessThatNeverRuns(t *testing.T) {
	p := schedulerPlan()
	p.Scheduler.Slots[2].Proc = 0
	p.Scheduler.Slots[5].Proc = 0
	err := validateSchedulerPlan(p)
	if err == nil {
		t.Fatal("a lane that never runs was accepted")
	}
	if !strings.Contains(err.Error(), "never run") || !strings.Contains(err.Error(), "P3") {
		t.Fatalf("the error does not say which lane sat empty: %v", err)
	}
}

func TestSchedulerRejectsATurnForAProcessThatDoesNotExist(t *testing.T) {
	p := schedulerPlan()
	p.Scheduler.Slots[1].Proc = 7
	err := validateSchedulerPlan(p)
	if err == nil {
		t.Fatal("a turn belonging to a process off the chart was accepted")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSchedulerRejectsAnOutOfRangeTurnLength(t *testing.T) {
	p := schedulerPlan()
	p.Scheduler.Slots[0].Len = 9
	err := validateSchedulerPlan(p)
	if err == nil {
		t.Fatal("a nine-unit turn was accepted")
	}
	if !strings.Contains(err.Error(), "9") {
		t.Fatalf("the error does not quote the offending length: %v", err)
	}
}

func TestSchedulerRequiresOpeningOnTheQueue(t *testing.T) {
	p := schedulerPlan()
	p.Beats[0].Scheduler = &SchedulerBeat{Show: "run", At: 0}
	err := validateSchedulerPlan(p)
	if err == nil {
		t.Fatal("a clip that lays a block before showing the queue was accepted")
	}
	if !strings.Contains(err.Error(), "open on the queue") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSchedulerRequiresClosingOnTheTally(t *testing.T) {
	p := schedulerPlan()
	p.Beats[6].Scheduler = &SchedulerBeat{Show: "run", At: 4}
	err := validateSchedulerPlan(p)
	if err == nil {
		t.Fatal("a clip that never totals the lanes was accepted")
	}
	if !strings.Contains(err.Error(), "close on the finished timeline") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSchedulerRejectsATallyBeforeTheEnd(t *testing.T) {
	p := schedulerPlan()
	p.Beats[1].Scheduler = &SchedulerBeat{Show: "fair"}
	err := validateSchedulerPlan(p)
	if err == nil {
		t.Fatal("totals struck with turns still to come were accepted")
	}
	if !strings.Contains(err.Error(), "closer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSchedulerRejectsATurnLaidOutOfOrder(t *testing.T) {
	p := schedulerPlan()
	p.Beats[2].Scheduler = &SchedulerBeat{Show: "run", At: 3}
	err := validateSchedulerPlan(p)
	if err == nil {
		t.Fatal("a block laid to the right of a hole was accepted")
	}
	if !strings.Contains(err.Error(), "turn 3") || !strings.Contains(err.Error(), "turn 1") {
		t.Fatalf("the error does not quote both the given and the expected turn: %v", err)
	}
}

func TestSchedulerRejectsASwitchWhoseNeighboursHaveNotRun(t *testing.T) {
	p := schedulerPlan()
	p.Beats[3].Scheduler = &SchedulerBeat{Show: "switch", At: 4}
	err := validateSchedulerPlan(p)
	if err == nil {
		t.Fatal("a switch zoomed at a boundary with nothing on one side was accepted")
	}
	if !strings.Contains(err.Error(), "no changeover to charge for") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSchedulerRejectsASwitchBeforeTheFirstTurn(t *testing.T) {
	p := schedulerPlan()
	p.Beats[3].Scheduler = &SchedulerBeat{Show: "switch", At: 0}
	err := validateSchedulerPlan(p)
	if err == nil {
		t.Fatal("a switch before the first turn was accepted")
	}
	if !strings.Contains(err.Error(), "a turn on each side") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSchedulerRejectsAChartWithNoPolicy(t *testing.T) {
	p := schedulerPlan()
	p.Scheduler.Policy = ""
	err := validateSchedulerPlan(p)
	if err == nil {
		t.Fatal("a Gantt chart with no policy on it was accepted")
	}
	if !strings.Contains(err.Error(), "no policy") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSchedulerRejectsASingleProcess(t *testing.T) {
	p := schedulerPlan()
	p.Scheduler.Procs = p.Scheduler.Procs[:1]
	if err := validateSchedulerPlan(p); err == nil {
		t.Fatal("a one-process chart was accepted, and one process shares nothing")
	}
}

func TestSchedulerNormalizeClampsTurnLengths(t *testing.T) {
	p := schedulerPlan()
	p.Scheduler.Slots[0].Len = 0
	p.Scheduler.Slots[1].Len = 99
	normalizeSchedulerPlan(p)
	if got := p.Scheduler.Slots[0].Len; got != minSchedulerSlotLen {
		t.Fatalf("a zero-length turn clamped to %d, want %d", got, minSchedulerSlotLen)
	}
	if got := p.Scheduler.Slots[1].Len; got != maxSchedulerSlotLen {
		t.Fatalf("an over-long turn clamped to %d, want %d", got, maxSchedulerSlotLen)
	}
}

func TestSchedulerNormalizeClampsAnOutOfRangeProcess(t *testing.T) {
	p := schedulerPlan()
	p.Scheduler.Slots[1].Proc = 9
	normalizeSchedulerPlan(p)
	if got := p.Scheduler.Slots[1].Proc; got != 2 {
		t.Fatalf("an out-of-range process clamped to %d, want 2", got)
	}
}

func TestSchedulerShowDefaultsToRun(t *testing.T) {
	b := SchedulerBeat{Show: "sparkle"}
	if got := b.ResolvedShow(); got != "run" {
		t.Fatalf("an unknown show resolved to %q, want run", got)
	}
}

// The component measures nothing and adds nothing up: block offsets, the axis
// length and each lane's total all arrive precomputed.
func TestSchedulerScenesComputeTheTimeline(t *testing.T) {
	p := schedulerPlan()
	scenes, err := schedulerScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	if props["units"] != 10 {
		t.Fatalf("the axis is %v units, want 10", props["units"])
	}
	procs, _ := props["procs"].([]map[string]any)
	if len(procs) != 3 {
		t.Fatalf("want 3 lanes, got %d", len(procs))
	}
	if procs[0]["total"] != 4 || procs[1]["total"] != 3 || procs[2]["total"] != 3 {
		t.Fatalf("the lane totals are wrong: %v, %v, %v", procs[0]["total"], procs[1]["total"], procs[2]["total"])
	}
	slots, _ := props["slots"].([]map[string]any)
	if slots[0]["start"] != 0 || slots[1]["start"] != 2 || slots[3]["start"] != 6 || slots[5]["start"] != 9 {
		t.Fatalf("the block offsets are wrong: %v", slots)
	}

	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != 7 {
		t.Fatalf("want 7 steps, got %d", len(steps))
	}
	if steps[0]["laid"] != 0 || steps[0]["show"] != "queue" {
		t.Fatalf("the opening beat already has blocks down: %v", steps[0])
	}
	if steps[3]["boundary"] != 2 {
		t.Fatalf("the switch beat's boundary is %v, want 2", steps[3]["boundary"])
	}
	// The closer completes the schedule, including the turns no beat narrated.
	last := steps[len(steps)-1]
	if last["laid"] != 6 {
		t.Fatalf("the closing beat has %v turns down, want all 6", last["laid"])
	}
	if last["show"] != "fair" {
		t.Fatalf("the last step shows %v, want fair", last["show"])
	}
}
