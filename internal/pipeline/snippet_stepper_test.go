package pipeline

import (
	"strings"
	"testing"
)

const stpNarration = "Half the row is gone with one look, and the half that is left is the only half worth checking."

func stepperPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "stepper",
		Title:    "Eight cells, three looks",
		Stepper: &StepperSpec{
			Values:   []int{3, 9, 14, 20, 42, 55, 71, 88},
			Pointers: []string{"low", "high", "mid"},
			Target:   42,
		},
		Beats: []SnippetBeat{
			{ID: "the-row", Heading: "The row", Narration: stpNarration, Stepper: &StepperBeat{Show: "array"}},
			{ID: "the-edges", Heading: "The edges", Narration: stpNarration, Stepper: &StepperBeat{Show: "point", Ptr: map[string]int{"low": 0, "high": 7}}},
			{ID: "first-look", Heading: "The first look", Narration: stpNarration, Stepper: &StepperBeat{Show: "compare", At: []int{3}, Ptr: map[string]int{"mid": 3}}},
			{ID: "move-low", Heading: "Half gone", Narration: stpNarration, Stepper: &StepperBeat{Show: "point", Ptr: map[string]int{"low": 4}}},
			{ID: "second-look", Heading: "The second look", Narration: stpNarration, Stepper: &StepperBeat{Show: "compare", At: []int{5}, Ptr: map[string]int{"mid": 5}}},
			{ID: "found-it", Heading: "Found", Narration: stpNarration, Stepper: &StepperBeat{Show: "found", At: []int{4}, Ptr: map[string]int{"high": 4, "mid": 4}}},
		},
	}
	// Against this template's own ideal of 28 words per beat rather than the
	// shared 40: at 40 the shared bounds demand more beats than the fixture
	// has, and it would be rejected for length before any rule under test ran.
	p.targetWords = 6 * 28
	return p
}

func TestStepperPlanAccepted(t *testing.T) {
	if err := validateStepperPlan(stepperPlan()); err != nil {
		t.Fatalf("a well-formed stepper plan was rejected: %v", err)
	}
}

// The family's signature rule: the array is tracked in Go, and a find that
// lands on a cell the value is not in is rejected with both numbers quoted.
func TestStepperRejectsAFindInACellThatDoesNotHoldTheTarget(t *testing.T) {
	p := stepperPlan()
	p.Beats[5].Stepper.At = []int{5}
	err := validateStepperPlan(p)
	if err == nil {
		t.Fatal("a find announced over the wrong cell was accepted")
	}
	if !strings.Contains(err.Error(), "42") || !strings.Contains(err.Error(), "holds 55") {
		t.Fatalf("the error does not quote both the target and what the cell really holds: %v", err)
	}
}

// The tracking is real: after a swap the target has MOVED, and the find has to
// follow it.
func TestStepperTracksTheArrayAcrossSwaps(t *testing.T) {
	p := stepperPlan()
	p.Beats[3].Stepper = &StepperBeat{Show: "swap", At: []int{0, 4}}
	p.Beats[5].Stepper.At = []int{0}
	if err := validateStepperPlan(p); err != nil {
		t.Fatalf("a find that followed the value through a swap was rejected: %v", err)
	}

	p.Beats[5].Stepper.At = []int{4}
	err := validateStepperPlan(p)
	if err == nil {
		t.Fatal("a find in the cell the target was swapped OUT of was accepted")
	}
	if !strings.Contains(err.Error(), "42") || !strings.Contains(err.Error(), "holds 3") {
		t.Fatalf("the error does not quote both numbers: %v", err)
	}
}

func TestStepperRejectsASwapOfOtherThanTwoCells(t *testing.T) {
	p := stepperPlan()
	p.Beats[3].Stepper = &StepperBeat{Show: "swap", At: []int{2}}
	err := validateStepperPlan(p)
	if err == nil {
		t.Fatal("a swap of one cell was accepted, and one cell has nowhere to go")
	}
	if !strings.Contains(err.Error(), "swaps 1 cells") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestStepperRejectsAComparisonOfThreeCells(t *testing.T) {
	p := stepperPlan()
	p.Beats[2].Stepper.At = []int{1, 2, 3}
	err := validateStepperPlan(p)
	if err == nil {
		t.Fatal("a three-way comparison was accepted, and this picture cannot draw one")
	}
	if !strings.Contains(err.Error(), "compares 3 cells") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestStepperRejectsAPointerNobodyDeclared(t *testing.T) {
	p := stepperPlan()
	p.Beats[1].Stepper.Ptr = map[string]int{"pivot": 2}
	err := validateStepperPlan(p)
	if err == nil {
		t.Fatal("a flag with a name the plan never declared was accepted")
	}
	if !strings.Contains(err.Error(), "pivot") {
		t.Fatalf("the error does not quote the offending name: %v", err)
	}
	if !strings.Contains(err.Error(), "low, high, mid") {
		t.Fatalf("the error does not list the declared pointers: %v", err)
	}
}

func TestStepperRejectsAPointerOffTheArray(t *testing.T) {
	p := stepperPlan()
	p.Beats[1].Stepper.Ptr = map[string]int{"high": 9}
	err := validateStepperPlan(p)
	if err == nil {
		t.Fatal("a flag standing over a cell that does not exist was accepted")
	}
	if !strings.Contains(err.Error(), "cell 9") {
		t.Fatalf("the error does not quote the cell: %v", err)
	}
}

func TestStepperRejectsACellIndexOffTheArray(t *testing.T) {
	p := stepperPlan()
	p.Beats[2].Stepper.At = []int{11}
	err := validateStepperPlan(p)
	if err == nil {
		t.Fatal("a beat acting on a cell past the end of the row was accepted")
	}
	if !strings.Contains(err.Error(), "cells 0-7") {
		t.Fatalf("the error does not say what the range is: %v", err)
	}
}

func TestStepperRejectsAFindWithNoTarget(t *testing.T) {
	p := stepperPlan()
	p.Stepper.Target = -1
	err := validateStepperPlan(p)
	if err == nil {
		t.Fatal("a clip that announces a find with nothing to find was accepted")
	}
	if !strings.Contains(err.Error(), "done") {
		t.Fatalf("the error does not offer the closer that fits instead: %v", err)
	}
}

func TestStepperRejectsAPointBeatThatMovesNothing(t *testing.T) {
	p := stepperPlan()
	p.Beats[3].Stepper.Ptr = nil
	if err := validateStepperPlan(p); err == nil {
		t.Fatal("a point beat with no pointer to move was accepted, and nothing on screen would change")
	}
}

func TestStepperRequiresOpeningOnTheArray(t *testing.T) {
	p := stepperPlan()
	p.Beats[0].Stepper = &StepperBeat{Show: "point", Ptr: map[string]int{"low": 0}}
	if err := validateStepperPlan(p); err == nil {
		t.Fatal("a clip that pointed at a row nobody had seen was accepted")
	}
}

func TestStepperRequiresClosingOnFoundOrDone(t *testing.T) {
	p := stepperPlan()
	p.Beats[5].Stepper = &StepperBeat{Show: "compare", At: []int{4}}
	if err := validateStepperPlan(p); err == nil {
		t.Fatal("a clip that stops mid-step was accepted")
	}
}

func TestStepperRejectsARowTooShortToNeedAnAlgorithm(t *testing.T) {
	p := stepperPlan()
	p.Stepper.Values = []int{3, 9, 42}
	err := validateStepperPlan(p)
	if err == nil {
		t.Fatal("a three-cell row was accepted, and a viewer solves three cells by looking")
	}
	if !strings.Contains(err.Error(), "3 cells") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestStepperRejectsAValueWiderThanATile(t *testing.T) {
	p := stepperPlan()
	p.Stepper.Values[2] = 1000
	err := validateStepperPlan(p)
	if err == nil {
		t.Fatal("a four-digit value was accepted, and it drops the whole row a type size")
	}
	if !strings.Contains(err.Error(), "1000") {
		t.Fatalf("the error does not quote the value: %v", err)
	}
}

func TestStepperRejectsARunWithNoPointers(t *testing.T) {
	p := stepperPlan()
	p.Stepper.Pointers = nil
	if err := validateStepperPlan(p); err == nil {
		t.Fatal("a run with no flags was accepted, and it is tiles changing colour with nothing over them")
	}
}

func TestStepperRejectsAPointerDeclaredTwice(t *testing.T) {
	p := stepperPlan()
	p.Stepper.Pointers = []string{"low", "low", "mid"}
	err := validateStepperPlan(p)
	if err == nil {
		t.Fatal("two flags with one name were accepted")
	}
	if !strings.Contains(err.Error(), "\"low\"") {
		t.Fatalf("the error does not quote the repeated name: %v", err)
	}
}

func TestStepperNormalizeClampsAndDeduplicatesCells(t *testing.T) {
	p := stepperPlan()
	p.Beats[2].Stepper.At = []int{3, 3, 99, -4}
	normalizeStepperPlan(p)
	got := p.Beats[2].Stepper.At
	if len(got) != 3 || got[0] != 3 || got[1] != 7 || got[2] != 0 {
		t.Fatalf("cells normalized to %v, want [3 7 0] after clamping and de-duplicating", got)
	}
}

func TestStepperNormalizeRepairsValuesAndPointers(t *testing.T) {
	p := stepperPlan()
	p.Stepper.Values[0] = -5
	p.Stepper.Values[1] = 1500
	p.Stepper.Pointers = []string{"low", " low ", "high", "the middle one", "spare"}
	normalizeStepperPlan(p)
	if p.Stepper.Values[0] != 0 || p.Stepper.Values[1] != maxStepperValue {
		t.Fatalf("values normalized to %v, want them clamped into 0-%d", p.Stepper.Values, maxStepperValue)
	}
	want := []string{"low", "high", "the middle"}
	if len(p.Stepper.Pointers) != len(want) {
		t.Fatalf("pointers normalized to %v, want three de-duplicated names", p.Stepper.Pointers)
	}
	for i, n := range want {
		if p.Stepper.Pointers[i] != n {
			t.Fatalf("pointer %d normalized to %q, want %q", i, p.Stepper.Pointers[i], n)
		}
	}
}

func TestStepperNormalizeCanonicalisesNoTarget(t *testing.T) {
	p := stepperPlan()
	p.Stepper.Target = -7
	normalizeStepperPlan(p)
	if p.Stepper.Target != -1 {
		t.Fatalf("a negative target normalized to %d, want the canonical -1", p.Stepper.Target)
	}
}

func TestStepperShowDefaultsToCompare(t *testing.T) {
	b := StepperBeat{Show: "shuffle"}
	if got := b.ResolvedShow(); got != "compare" {
		t.Fatalf("an unknown show resolved to %q, want compare", got)
	}
}

// The component draws the state it is handed. Every step carries the full
// contents of the row, the pointer positions and the op count.
func TestStepperScenesShipEveryArrayState(t *testing.T) {
	p := stepperPlan()
	scenes, err := stepperScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("want one scene spanning the clip, got %d", len(scenes))
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if len(steps) != 6 {
		t.Fatalf("want 6 steps, got %d", len(steps))
	}

	first := steps[0]
	if first["show"] != "array" || first["ops"] != 0 {
		t.Fatalf("the opener is wrong: %v", first)
	}
	if vals, _ := first["values"].([]int); len(vals) != 8 || vals[4] != 42 {
		t.Fatalf("the opener ships %v, want the starting row", first["values"])
	}
	if touched, _ := first["touched"].([]int); len(touched) != 0 {
		t.Fatalf("the opener has touched %v, want nothing", touched)
	}
	if ptr, _ := first["ptr"].(map[string]any); ptr["low"] != -1 || ptr["mid"] != -1 {
		t.Fatalf("the opener places flags at %v, want them all off the row", ptr)
	}

	last := steps[len(steps)-1]
	if last["show"] != "found" {
		t.Fatalf("last step shows %v, want found", last["show"])
	}
	if last["ops"] != 2 {
		t.Fatalf("the counter ends on %v, want the two comparisons", last["ops"])
	}
	touched, _ := last["touched"].([]int)
	if len(touched) != 3 || touched[0] != 3 || touched[1] != 4 || touched[2] != 5 {
		t.Fatalf("the closer has touched %v, want cells 3, 4 and 5", touched)
	}
	ptr, _ := last["ptr"].(map[string]any)
	if ptr["low"] != 4 || ptr["high"] != 4 || ptr["mid"] != 4 {
		t.Fatalf("the flags end at %v, want all three closed on cell 4", ptr)
	}
}

func TestStepperScenesApplyTheSwapsInGo(t *testing.T) {
	p := stepperPlan()
	p.Beats[3].Stepper = &StepperBeat{Show: "swap", At: []int{0, 4}}
	p.Beats[5].Stepper.At = []int{0}
	scenes, err := stepperScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)

	before, _ := steps[2]["values"].([]int)
	if before[0] != 3 || before[4] != 42 {
		t.Fatalf("the row before the swap is %v, want it untouched", before)
	}
	after, _ := steps[len(steps)-1]["values"].([]int)
	if after[0] != 42 || after[4] != 3 {
		t.Fatalf("the row after the swap is %v, want 42 and 3 to have traded places", after)
	}
	if steps[len(steps)-1]["ops"] != 3 {
		t.Fatalf("the counter ends on %v, want two comparisons and a swap", steps[len(steps)-1]["ops"])
	}
}
