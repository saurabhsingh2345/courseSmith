package pipeline

import (
	"strings"
	"testing"
)

const pipeNarration = "One tick of the clock moves every chip along and lets another instruction walk in."

func pipelinePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "pipeline",
		Title:    "Five instructions in flight at once",
		Pipeline: &PipelineSpec{
			StageNames: []string{"IF", "ID", "EX", "MEM", "WB"},
			Items:      []string{"load", "add", "store", "branch"},
			Stall:      "the add is waiting on the load result",
		},
		Beats: []SnippetBeat{
			{ID: "grid", Heading: "Five stages", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "empty"}},
			{ID: "t1", Heading: "First tick", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "fill"}},
			{ID: "t2", Heading: "Second tick", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "fill"}},
			{ID: "t3", Heading: "Third tick", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "fill"}},
			{ID: "bubble", Heading: "The bubble", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "stall"}},
			{ID: "flow", Heading: "One per cycle", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "flow"}},
		},
	}
	// The template's ideal is 28 words per beat, so the fixture budget is sized
	// against that — nBeats * 40 would demand more beats than it has.
	p.targetWords = 6 * 28
	return p
}

func TestPipelinePlanAccepted(t *testing.T) {
	if err := validatePipelinePlan(pipelinePlan()); err != nil {
		t.Fatalf("a well-formed pipeline was rejected: %v", err)
	}
}

func TestPipelineRejectsTooFewStages(t *testing.T) {
	p := pipelinePlan()
	p.Pipeline.StageNames = []string{"IF", "WB"}
	err := validatePipelinePlan(p)
	if err == nil {
		t.Fatal("a two-column pipeline was accepted, and nothing can be in flight behind anything else")
	}
	if !strings.Contains(err.Error(), "2 stages") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestPipelineRejectsTooManyStages(t *testing.T) {
	p := pipelinePlan()
	p.Pipeline.StageNames = append(p.Pipeline.StageNames, "COMMIT")
	if err := validatePipelinePlan(p); err == nil {
		t.Fatal("a six-column pipeline was accepted, and the chips lose the width their labels need")
	}
}

func TestPipelineRejectsASingleItem(t *testing.T) {
	p := pipelinePlan()
	p.Pipeline.Items = []string{"load"}
	err := validatePipelinePlan(p)
	if err == nil {
		t.Fatal("a one-item stream was accepted, and one chip never overlaps with anything")
	}
	if !strings.Contains(err.Error(), "1 items") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestPipelineRejectsAChipThatIsAnExplanation(t *testing.T) {
	p := pipelinePlan()
	p.Pipeline.Items[1] = "add the two register values together"
	if err := validatePipelinePlan(p); err == nil {
		t.Fatal("a six-word chip label was accepted")
	}
}

func TestPipelineRejectsAStallWithNoNamedHazard(t *testing.T) {
	p := pipelinePlan()
	p.Pipeline.Stall = ""
	err := validatePipelinePlan(p)
	if err == nil {
		t.Fatal("a bubble with no cause was accepted, and it reads as a rendering glitch")
	}
	if !strings.Contains(err.Error(), "bubble") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPipelineRejectsASecondStall(t *testing.T) {
	p := pipelinePlan()
	p.Beats[3].Pipeline = &PipelineBeat{Show: "stall"}
	err := validatePipelinePlan(p)
	if err == nil {
		t.Fatal("a pipeline that stutters twice was accepted")
	}
	if !strings.Contains(err.Error(), "second stall") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPipelineRejectsAStallBeforeAnythingIsInFlight(t *testing.T) {
	p := pipelinePlan()
	p.Beats[1].Pipeline = &PipelineBeat{Show: "stall"}
	p.Beats[4].Pipeline = &PipelineBeat{Show: "fill"}
	err := validatePipelinePlan(p)
	if err == nil {
		t.Fatal("a stall on an empty grid was accepted")
	}
	if !strings.Contains(err.Error(), "before anything is in flight") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// One tick shows a chip moving, and a chip moving is not pipelining.
func TestPipelineRejectsASingleTick(t *testing.T) {
	p := pipelinePlan()
	p.Beats = []SnippetBeat{p.Beats[0], p.Beats[1], p.Beats[len(p.Beats)-1]}
	p.targetWords = 3 * 28
	err := validatePipelinePlan(p)
	if err == nil {
		t.Fatal("a one-tick clip was accepted, and one tick shows nothing about pipelining")
	}
	if !strings.Contains(err.Error(), "1 fill beats") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

// THE ARITHMETIC. The stream drains after items + stages - 1 ticks, and the
// rejection quotes both the ticks asked for and the ticks available.
func TestPipelineRejectsMoreTicksThanTheStreamCanFill(t *testing.T) {
	p := pipelinePlan()
	p.Pipeline.StageNames = []string{"fetch", "run", "retire"}
	p.Pipeline.Items = []string{"first", "second"}
	p.Pipeline.Stall = ""
	p.Beats = []SnippetBeat{
		{ID: "grid", Heading: "Three stages", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "empty"}},
		{ID: "t1", Heading: "Tick one", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "fill"}},
		{ID: "t2", Heading: "Tick two", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "fill"}},
		{ID: "t3", Heading: "Tick three", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "fill"}},
		{ID: "t4", Heading: "Tick four", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "fill"}},
		{ID: "t5", Heading: "Tick five", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "fill"}},
		{ID: "flow", Heading: "Steady state", Narration: pipeNarration, Pipeline: &PipelineBeat{Show: "flow"}},
	}
	p.targetWords = 7 * 28
	err := validatePipelinePlan(p)
	if err == nil {
		t.Fatal("a clip that spends its last ticks animating an empty grid was accepted")
	}
	if !strings.Contains(err.Error(), "runs 5 ticks") {
		t.Fatalf("the error does not quote the ticks asked for: %v", err)
	}
	if !strings.Contains(err.Error(), "drains after 4") {
		t.Fatalf("the error does not quote the ticks the stream can fill: %v", err)
	}
}

func TestPipelineRequiresOpeningOnTheEmptyGrid(t *testing.T) {
	p := pipelinePlan()
	p.Beats[0].Pipeline = &PipelineBeat{Show: "fill"}
	if err := validatePipelinePlan(p); err == nil {
		t.Fatal("a clip whose first tick lands in unlabelled space was accepted")
	}
}

func TestPipelineRequiresClosingOnTheFlow(t *testing.T) {
	p := pipelinePlan()
	p.Beats[len(p.Beats)-1].Pipeline = &PipelineBeat{Show: "fill"}
	if err := validatePipelinePlan(p); err == nil {
		t.Fatal("a clip that never lands the throughput point was accepted")
	}
}

func TestPipelineRejectsAGridEmptiedPartWayThrough(t *testing.T) {
	p := pipelinePlan()
	p.Beats[3].Pipeline = &PipelineBeat{Show: "empty"}
	err := validatePipelinePlan(p)
	if err == nil {
		t.Fatal("a clip that un-ran its own stream was accepted")
	}
	if !strings.Contains(err.Error(), "empties the grid") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPipelineNormalizeClampsNamesAndCaption(t *testing.T) {
	p := pipelinePlan()
	p.Pipeline.StageNames = []string{"  instruction   fetch  ", "", "instruction decode and register read", "EX", "MEM"}
	p.Pipeline.Stall = "  the add instruction is waiting on the load result to come back from memory  "
	normalizePipelinePlan(p)

	if len(p.Pipeline.StageNames) != 4 {
		t.Fatalf("the blank column survived normalize: %v", p.Pipeline.StageNames)
	}
	if p.Pipeline.StageNames[0] != "instruction fetch" {
		t.Fatalf("the first column normalized to %q", p.Pipeline.StageNames[0])
	}
	if got := len(strings.Fields(p.Pipeline.StageNames[1])); got > maxPipelineStageWords {
		t.Fatalf("a column header normalized to %d words, want at most %d", got, maxPipelineStageWords)
	}
	if got := len(strings.Fields(p.Pipeline.Stall)); got > maxPipelineStallWords {
		t.Fatalf("the stall caption normalized to %d words, want at most %d", got, maxPipelineStallWords)
	}
	if err := validatePipelinePlan(p); err != nil {
		t.Fatalf("a repairable pipeline was rejected after normalize: %v", err)
	}
}

func TestPipelineShowDefaultsToFill(t *testing.T) {
	b := PipelineBeat{Show: "advance"}
	if got := b.ResolvedShow(); got != "fill" {
		t.Fatalf("an unknown show resolved to %q, want fill", got)
	}
}

// The machine runs in Go. The component draws the grid it is handed and never
// works out for itself where a chip should be.
func TestPipelineScenesSimulateTheOccupancy(t *testing.T) {
	p := pipelinePlan()
	scenes, err := pipelineScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	if props["sequentialTicks"] != 20 {
		t.Fatalf("four items through five stages one at a time is 20 ticks, got %v", props["sequentialTicks"])
	}
	if props["pipelinedTicks"] != 8 {
		t.Fatalf("four items overlapped through five stages is 8 ticks, got %v", props["pipelinedTicks"])
	}

	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != 6 {
		t.Fatalf("want 6 steps, got %d", len(steps))
	}

	first, _ := steps[0]["occ"].([]int)
	if len(first) != 5 {
		t.Fatalf("the grid is %d wide, want 5", len(first))
	}
	for i, v := range first {
		if v != pipelineEmptyCell {
			t.Fatalf("the opening grid has item %d in stage %d, and it should be empty", v, i)
		}
	}
	if steps[0]["tick"] != 0 {
		t.Fatalf("the opener has already ticked: %v", steps[0]["tick"])
	}

	// Three ticks in, the stream has walked three chips into the first three
	// columns and nothing has retired.
	third, _ := steps[3]["occ"].([]int)
	want := []int{2, 1, 0, -1, -1}
	for i := range want {
		if third[i] != want[i] {
			t.Fatalf("after three ticks the grid is %v, want %v", third, want)
		}
	}
	if steps[3]["retired"] != 0 {
		t.Fatalf("something retired before reaching the last stage: %v", steps[3]["retired"])
	}

	// The stall: the chip in ID holds, the stages ahead of it advance, and the
	// gap that opens behind them is the bubble.
	stall, _ := steps[4]["occ"].([]int)
	wantStall := []int{2, 1, -1, 0, -1}
	for i := range wantStall {
		if stall[i] != wantStall[i] {
			t.Fatalf("the stalled grid is %v, want %v", stall, wantStall)
		}
	}
	if steps[4]["bubble"] != 2 {
		t.Fatalf("the bubble opened in stage %v, want stage 2", steps[4]["bubble"])
	}

	last := steps[len(steps)-1]
	if last["show"] != "flow" {
		t.Fatalf("last step shows %v, want flow", last["show"])
	}
	if last["bubble"] != pipelineEmptyCell {
		t.Fatalf("the closer still shows a bubble: %v", last["bubble"])
	}
	held, _ := last["occ"].([]int)
	for i := range wantStall {
		if held[i] != wantStall[i] {
			t.Fatalf("the closer moved the grid to %v, and flow holds steady state at %v", held, wantStall)
		}
	}
	inFlight, _ := last["inFlight"].([]int)
	if len(inFlight) != 3 || inFlight[0] != 0 || inFlight[2] != 2 {
		t.Fatalf("the closer reports %v in flight, want items 0, 1 and 2", inFlight)
	}
}
