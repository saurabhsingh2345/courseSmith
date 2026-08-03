package pipeline

import (
	"strings"
	"testing"
)

const ltNarration = "A memory read and a table scan are not the same kind of slow, and the axis shows why."

func latencyPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "latency",
		Title:    "Not the same kind of slow",
		Latency: &LatencySpec{
			Operations: []LatencyOp{
				{Label: "a Redis GET", Ms: 0.12, Note: "Memory, one hop", Role: "quantity"},
				{Label: "an indexed SQL query", Ms: 12, Note: "A hundred times longer", Role: "neutral"},
				{Label: "the same query unindexed", Ms: 6479, Note: "Fifty thousand GETs in that time", Role: "limit"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "axis", Heading: "A time axis", Narration: ltNarration, Latency: &LatencyBeat{Show: "axis"}},
			{ID: "get", Heading: "The fast one", Narration: ltNarration, Latency: &LatencyBeat{Show: "place", At: 0}},
			{ID: "query", Heading: "A real query", Narration: ltNarration, Latency: &LatencyBeat{Show: "place", At: 1}},
			{ID: "scan", Heading: "Without the index", Narration: ltNarration, Latency: &LatencyBeat{Show: "place", At: 2}},
			{ID: "gap", Heading: "What that means", Narration: ltNarration, Latency: &LatencyBeat{Show: "read"}},
		},
	}
	p.targetWords = 5 * 40
	return p
}

func TestLatencyPlanAccepted(t *testing.T) {
	if err := validateLatencyPlan(latencyPlan()); err != nil {
		t.Fatalf("a well-formed latency plan was rejected: %v", err)
	}
}

// The rule the template exists for. Inside one decade a linear chart is honest,
// easier to read and already in the catalog — a log axis there is a scale chosen
// to make a small difference look structural.
func TestLatencyRejectsASpanInsideOneDecade(t *testing.T) {
	p := latencyPlan()
	p.Latency.Operations = []LatencyOp{
		{Label: "one", Ms: 12, Role: "quantity"},
		{Label: "two", Ms: 18, Role: "limit"},
		{Label: "three", Ms: 24, Role: "neutral"},
	}
	err := validateLatencyPlan(p)
	if err == nil {
		t.Fatal("three durations inside one decade were accepted")
	}
	if !strings.Contains(err.Error(), "data template") {
		t.Fatalf("the error does not name the template that fits: %v", err)
	}
}

// log10(0) is undefined, and "instant" is a claim no measurement supports.
func TestLatencyRejectsAZeroDuration(t *testing.T) {
	p := latencyPlan()
	p.Latency.Operations[0].Ms = 0
	err := validateLatencyPlan(p)
	if err == nil {
		t.Fatal("a zero duration was accepted")
	}
	if !strings.Contains(err.Error(), "instant") {
		t.Fatalf("the error does not address the likely intent: %v", err)
	}
}

func TestLatencyRequiresTheAxisFirst(t *testing.T) {
	p := latencyPlan()
	p.Beats[0].Latency = &LatencyBeat{Show: "place", At: 0}
	p.Beats[1].Latency = &LatencyBeat{Show: "axis"}
	err := validateLatencyPlan(p)
	if err == nil {
		t.Fatal("a clip that places before drawing the axis was accepted")
	}
	if !strings.Contains(err.Error(), "not a bar chart") {
		t.Fatalf("the error does not say why the axis matters: %v", err)
	}
}

func TestLatencyRequiresEveryOperationToLand(t *testing.T) {
	p := latencyPlan()
	p.Beats[3].Latency = &LatencyBeat{Show: "place", At: 1}
	if err := validateLatencyPlan(p); err == nil {
		t.Fatal("an operation no beat ever places was accepted")
	}
}

func TestLatencyRejectsADuplicateLabel(t *testing.T) {
	p := latencyPlan()
	p.Latency.Operations[1].Label = "a Redis GET"
	if err := validateLatencyPlan(p); err == nil {
		t.Fatal("two operations with the same label were accepted")
	}
}

// The axis pads out to whole decades so a tick means something to a viewer
// without arithmetic.
func TestLatencyAxisPadsToWholeDecades(t *testing.T) {
	lo, hi := latencyAxis([]LatencyOp{{Ms: 0.12}, {Ms: 6479}})
	if lo != -1 {
		t.Fatalf("low decade is 1e%d, want 1e-1 (below 0.12ms)", lo)
	}
	if hi != 4 {
		t.Fatalf("high decade is 1e%d, want 1e4 (above 6479ms)", hi)
	}
}

// A single value would otherwise give a zero-width axis and divide by zero.
func TestLatencyAxisNeverCollapses(t *testing.T) {
	lo, hi := latencyAxis([]LatencyOp{{Ms: 100}, {Ms: 100}})
	if hi <= lo {
		t.Fatalf("axis is 1e%d..1e%d, which has no width", lo, hi)
	}
}

// Placement is logarithmic: ten times the duration is one decade further along,
// which for a five-decade axis is exactly a fifth of the width.
func TestLatencyPlacementIsLogarithmic(t *testing.T) {
	lo, hi := -1, 4
	a := latencyFrac(1, lo, hi)
	b := latencyFrac(10, lo, hi)
	if d := b - a; d < 0.19 || d > 0.21 {
		t.Fatalf("a tenfold step moved %v of the axis, want ~0.2 across five decades", d)
	}
}

// A duration is rendered the way somebody would say it, not in the unit it was
// stored in: 6479ms is "6.5s" to a human and "6479ms" only to a log file.
func TestLatencyLabelsReadNaturally(t *testing.T) {
	cases := map[float64]string{
		0.12:   "120µs",
		12:     "12ms",
		6479:   "6.5s",
		120000: "2min",
	}
	for ms, want := range cases {
		if got := latencyLabel(ms); got != want {
			t.Errorf("latencyLabel(%v) = %q, want %q", ms, got, want)
		}
	}
}

func TestLatencyScenesCarryTicksAndPlacements(t *testing.T) {
	p := latencyPlan()
	scenes, err := latencyScenes(sceneInput(t, p, 20000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	ticks, _ := scenes[0].Props["ticks"].([]map[string]any)
	if len(ticks) != 6 {
		t.Fatalf("want 6 decade ticks for a 1e-1..1e4 axis, got %d", len(ticks))
	}
	if ticks[0]["frac"].(float64) != 0 || ticks[len(ticks)-1]["frac"].(float64) != 1 {
		t.Fatalf("the axis does not run 0..1: %v .. %v", ticks[0]["frac"], ticks[len(ticks)-1]["frac"])
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if got, _ := steps[0]["placed"].([]int); len(got) != 0 {
		t.Fatalf("the axis beat already has placements: %v", got)
	}
	if got, _ := steps[4]["placed"].([]int); len(got) != 3 {
		t.Fatalf("the closing beat carries %v placements, want 3", len(got))
	}
}
