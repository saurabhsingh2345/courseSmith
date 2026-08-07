package pipeline

import (
	"fmt"
	"strings"
	"testing"
)

const gtNarration = "Set both pins the same way and the output stays dark, which is the whole rule."

func gatesPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "gates",
		Title:    "The gate that spots a difference",
		Gates: &GatesSpec{
			Gate:   "XOR",
			Inputs: []string{"A", "B"},
			Rows: []GatesRow{
				{In: []int{0, 0}, Out: 0},
				{In: []int{0, 1}, Out: 1},
				{In: []int{1, 0}, Out: 1},
				{In: []int{1, 1}, Out: 0},
			},
		},
		Beats: []SnippetBeat{
			{ID: "circuit", Heading: "The circuit", Narration: gtNarration, Gates: &GatesBeat{Show: "circuit"}},
			{ID: "both-off", Heading: "Both off", Narration: gtNarration, Gates: &GatesBeat{Show: "row", At: 0}},
			{ID: "b-on", Heading: "One on", Narration: gtNarration, Gates: &GatesBeat{Show: "row", At: 1}},
			{ID: "a-on", Heading: "The other on", Narration: gtNarration, Gates: &GatesBeat{Show: "row", At: 2}},
			{ID: "both-on", Heading: "Both on", Narration: gtNarration, Gates: &GatesBeat{Show: "row", At: 3}},
			{ID: "law", Heading: "The rule", Narration: gtNarration, Gates: &GatesBeat{Show: "law"}},
		},
	}
	// The template's ideal is 28 words per beat, so the fixture budget is sized
	// against that — nBeats * 40 would demand more beats than it has.
	p.targetWords = 6 * 28
	return p
}

func TestGatesPlanAccepted(t *testing.T) {
	if err := validateGatesPlan(gatesPlan()); err != nil {
		t.Fatalf("a well-formed gates plan was rejected: %v", err)
	}
}

// A one-input gate is a different circuit and a different table, so it gets its
// own acceptance test rather than being assumed.
func TestGatesAcceptsAOneInputGate(t *testing.T) {
	p := gatesPlan()
	p.Gates.Gate = "NOT"
	p.Gates.Inputs = []string{"A"}
	p.Gates.Rows = []GatesRow{
		{In: []int{0}, Out: 1},
		{In: []int{1}, Out: 0},
	}
	p.Beats = []SnippetBeat{
		{ID: "circuit", Heading: "The circuit", Narration: gtNarration, Gates: &GatesBeat{Show: "circuit"}},
		{ID: "low", Heading: "Low in", Narration: gtNarration, Gates: &GatesBeat{Show: "row", At: 0}},
		{ID: "high", Heading: "High in", Narration: gtNarration, Gates: &GatesBeat{Show: "row", At: 1}},
		{ID: "law", Heading: "The rule", Narration: gtNarration, Gates: &GatesBeat{Show: "law"}},
	}
	p.targetWords = 4 * 28
	if err := validateGatesPlan(p); err != nil {
		t.Fatalf("a well-formed inverter plan was rejected: %v", err)
	}
}

// The family's signature rule: the gate function is evaluated in Go over every
// row's own inputs, and a wrong row is rejected naming the row.
func TestGatesRejectsARowTheGateDoesNotProduce(t *testing.T) {
	p := gatesPlan()
	p.Gates.Rows[2].Out = 0
	err := validateGatesPlan(p)
	if err == nil {
		t.Fatal("a truth table with a wrong output was accepted")
	}
	if !strings.Contains(err.Error(), "XOR(1, 0)") {
		t.Fatalf("the error does not quote the offending row: %v", err)
	}
	if !strings.Contains(err.Error(), "is 1") || !strings.Contains(err.Error(), "claims 0") {
		t.Fatalf("the error does not quote both outputs: %v", err)
	}
}

func TestGatesRejectsARepeatedCombination(t *testing.T) {
	p := gatesPlan()
	p.Gates.Rows[3] = GatesRow{In: []int{1, 0}, Out: 1}
	err := validateGatesPlan(p)
	if err == nil {
		t.Fatal("a table with the same combination twice was accepted")
	}
	if !strings.Contains(err.Error(), "A=1, B=0") {
		t.Fatalf("the error does not name the duplicated combination: %v", err)
	}
}

func TestGatesRejectsAnIncompleteTable(t *testing.T) {
	p := gatesPlan()
	p.Gates.Rows = p.Gates.Rows[:3]
	p.Beats = append(p.Beats[:4], p.Beats[5])
	p.targetWords = 5 * 28
	err := validateGatesPlan(p)
	if err == nil {
		t.Fatal("a three-row table for a two-input gate was accepted")
	}
	if !strings.Contains(err.Error(), "exactly 4") {
		t.Fatalf("the error does not say how many rows a two-input gate has: %v", err)
	}
}

func TestGatesRejectsAnUnknownGate(t *testing.T) {
	p := gatesPlan()
	p.Gates.Gate = "XNOR"
	err := validateGatesPlan(p)
	if err == nil {
		t.Fatal("a gate outside the drawn vocabulary was accepted")
	}
	if !strings.Contains(err.Error(), "NAND") {
		t.Fatalf("the error does not list the gates that can be drawn: %v", err)
	}
}

func TestGatesRejectsAnInverterWithTwoInputs(t *testing.T) {
	p := gatesPlan()
	p.Gates.Gate = "NOT"
	err := validateGatesPlan(p)
	if err == nil {
		t.Fatal("an inverter drawn with two input wires was accepted")
	}
	if !strings.Contains(err.Error(), "exactly 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGatesRejectsANonBinaryInput(t *testing.T) {
	p := gatesPlan()
	p.Gates.Rows[1].In = []int{0, 2}
	if err := validateGatesPlan(p); err == nil {
		t.Fatal("a wire carrying a 2 was accepted")
	}
}

func TestGatesRequiresOpeningOnTheCircuit(t *testing.T) {
	p := gatesPlan()
	p.Beats[0].Gates = &GatesBeat{Show: "row", At: 0}
	if err := validateGatesPlan(p); err == nil {
		t.Fatal("a clip that fires a row before drawing the gate was accepted")
	}
}

func TestGatesRequiresClosingOnTheLaw(t *testing.T) {
	p := gatesPlan()
	p.Beats[5].Gates = &GatesBeat{Show: "row", At: 3}
	if err := validateGatesPlan(p); err == nil {
		t.Fatal("a clip that never states the gate's rule was accepted")
	}
}

func TestGatesRequiresRowsFiredInTableOrder(t *testing.T) {
	p := gatesPlan()
	p.Beats[1].Gates = &GatesBeat{Show: "row", At: 1}
	p.Beats[2].Gates = &GatesBeat{Show: "row", At: 0}
	err := validateGatesPlan(p)
	if err == nil {
		t.Fatal("rows fired out of table order were accepted")
	}
	if !strings.Contains(err.Error(), "row 1") {
		t.Fatalf("the error does not name the offending row: %v", err)
	}
}

func TestGatesRequiresEveryRowFired(t *testing.T) {
	p := gatesPlan()
	p.Beats = append(p.Beats[:4], p.Beats[5])
	p.targetWords = 5 * 28
	err := validateGatesPlan(p)
	if err == nil {
		t.Fatal("a table with a row that never fires was accepted")
	}
	if !strings.Contains(err.Error(), "3 of the 4 rows") {
		t.Fatalf("the error does not count the unfired rows: %v", err)
	}
}

// Case and stray spacing on a gate name are phrasing habits, not wrong answers.
func TestGatesNormalizeUppercasesTheGate(t *testing.T) {
	p := gatesPlan()
	p.Gates.Gate = " xor "
	p.Gates.Inputs = []string{"input  A", "the  second  input  wire"}
	normalizeGatesPlan(p)
	if p.Gates.Gate != "XOR" {
		t.Fatalf("gate normalized to %q, want XOR", p.Gates.Gate)
	}
	if p.Gates.Inputs[0] != "input A" || p.Gates.Inputs[1] != "the second" {
		t.Fatalf("input labels normalized to %v", p.Gates.Inputs)
	}
	if err := validateGatesPlan(p); err != nil {
		t.Fatalf("a lowercased-but-correct plan was rejected after normalize: %v", err)
	}
}

func TestGatesNormalizeClampsAnOutOfRangeRow(t *testing.T) {
	p := gatesPlan()
	p.Beats[4].Gates.At = 99
	normalizeGatesPlan(p)
	if got := p.Beats[4].Gates.At; got != 3 {
		t.Fatalf("row index clamped to %d, want 3", got)
	}
}

func TestGatesShowDefaultsToRow(t *testing.T) {
	b := GatesBeat{Show: "sparkle"}
	if got := b.ResolvedShow(); got != "row" {
		t.Fatalf("an unknown show resolved to %q, want row", got)
	}
}

// The law line is a fact about the gate, stated in Go, so a plan cannot get it
// wrong and it cannot disagree with the table beside it.
func TestGatesScenesShipTheComputedTableAndLaw(t *testing.T) {
	p := gatesPlan()
	scenes, err := gatesScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	if props["gate"] != "XOR" {
		t.Fatalf("the gate is %v, want XOR", props["gate"])
	}
	if props["law"] != "one when the inputs differ" {
		t.Fatalf("the law line is %v", props["law"])
	}

	rows, _ := props["rows"].([]map[string]any)
	if len(rows) != 4 {
		t.Fatalf("want 4 rows, got %d", len(rows))
	}
	if rows[0]["out"] != 0 || rows[1]["out"] != 1 || rows[2]["out"] != 1 || rows[3]["out"] != 0 {
		t.Fatalf("the computed outputs are wrong: %v", rows)
	}

	steps, _ := props["steps"].([]map[string]any)
	if steps[0]["show"] != "circuit" {
		t.Fatalf("first step shows %v, want circuit", steps[0]["show"])
	}
	if got := fmt.Sprint(steps[0]["done"]); got != "[]" {
		t.Fatalf("the opener has already fired rows: %v", got)
	}
	last := steps[len(steps)-1]
	if last["show"] != "law" {
		t.Fatalf("last step shows %v, want law", last["show"])
	}
	if got := fmt.Sprint(last["done"]); got != "[0 1 2 3]" {
		t.Fatalf("the closer has fired %v, want every row", got)
	}
}
