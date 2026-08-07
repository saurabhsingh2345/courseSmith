package pipeline

import (
	"fmt"
	"strings"
	"testing"
)

const caNarration = "Line the digits up and work one column at a time from the right."

func carryPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "carry",
		Title:    "Binary addition is the same addition",
		Carry: &CarrySpec{
			A:   "1011",
			B:   "0110",
			Sum: "10001",
		},
		Beats: []SnippetBeat{
			{ID: "problem", Heading: "The problem", Narration: caNarration, Carry: &CarryBeat{Show: "problem"}},
			{ID: "ones", Heading: "The ones column", Narration: caNarration, Carry: &CarryBeat{Show: "column", At: 0}},
			{ID: "twos", Heading: "The twos column", Narration: caNarration, Carry: &CarryBeat{Show: "column", At: 1}},
			{ID: "fours", Heading: "The fours column", Narration: caNarration, Carry: &CarryBeat{Show: "column", At: 2}},
			{ID: "eights", Heading: "The eights column", Narration: caNarration, Carry: &CarryBeat{Show: "column", At: 3}},
			{ID: "sixteens", Heading: "The last carry", Narration: caNarration, Carry: &CarryBeat{Show: "column", At: 4}},
			{ID: "answer", Heading: "The answer", Narration: caNarration, Carry: &CarryBeat{Show: "answer"}},
		},
	}
	// The template's ideal is 28 words per beat, so the fixture budget is sized
	// against that — nBeats * 40 would demand more beats than it has.
	p.targetWords = 7 * 28
	return p
}

func TestCarryPlanAccepted(t *testing.T) {
	if err := validateCarryPlan(carryPlan()); err != nil {
		t.Fatalf("a well-formed carry plan was rejected: %v", err)
	}
}

// The family's signature rule: the addition is done in Go, and a wrong bottom
// row is rejected with the true sum quoted in binary AND in decimal.
func TestCarryRejectsASumThatIsNotTheSum(t *testing.T) {
	p := carryPlan()
	p.Carry.Sum = "10011"
	err := validateCarryPlan(p)
	if err == nil {
		t.Fatal("an addition whose bottom row is wrong was accepted")
	}
	if !strings.Contains(err.Error(), "10001") {
		t.Fatalf("the error does not quote the true sum in binary: %v", err)
	}
	if !strings.Contains(err.Error(), "17") {
		t.Fatalf("the error does not quote the true sum in decimal: %v", err)
	}
}

func TestCarryRejectsOperandsWithForeignDigits(t *testing.T) {
	p := carryPlan()
	p.Carry.A = "1211"
	if err := validateCarryPlan(p); err == nil {
		t.Fatal("an operand with a 2 in it was accepted")
	}
}

func TestCarryRejectsAnOperandWiderThanAByte(t *testing.T) {
	p := carryPlan()
	p.Carry.A = "100000001"
	p.Carry.Sum = "100000111"
	if err := validateCarryPlan(p); err == nil {
		t.Fatal("a nine-bit operand was accepted")
	}
}

// A problem with no carries is a clip about carrying that never carries.
func TestCarryRejectsAnAdditionThatNeverCarries(t *testing.T) {
	p := carryPlan()
	p.Carry.A = "1010"
	p.Carry.B = "101"
	p.Carry.Sum = "1111"
	err := validateCarryPlan(p)
	if err == nil {
		t.Fatal("an addition with no carry at all was accepted")
	}
	if !strings.Contains(err.Error(), "carry") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCarryRejectsATooNarrowProblem(t *testing.T) {
	p := carryPlan()
	// One plus one carries, so this clears the carry rule and fails only on
	// being two columns wide — which is an example rather than a procedure.
	p.Carry.A = "1"
	p.Carry.B = "1"
	p.Carry.Sum = "10"
	if err := validateCarryPlan(p); err == nil {
		t.Fatal("a two-column problem was accepted")
	}
}

func TestCarryRequiresOpeningOnTheProblem(t *testing.T) {
	p := carryPlan()
	p.Beats[0].Carry = &CarryBeat{Show: "column", At: 0}
	if err := validateCarryPlan(p); err == nil {
		t.Fatal("a clip that works a column before laying the problem out was accepted")
	}
}

func TestCarryRequiresClosingOnTheAnswer(t *testing.T) {
	p := carryPlan()
	p.Beats[6].Carry = &CarryBeat{Show: "column", At: 5}
	if err := validateCarryPlan(p); err == nil {
		t.Fatal("a clip that stops mid-column was accepted")
	}
}

// Long addition has exactly one legal direction, because a column cannot be
// worked until the carry coming into it is known.
func TestCarryRejectsColumnsOutOfOrder(t *testing.T) {
	p := carryPlan()
	p.Beats[1].Carry = &CarryBeat{Show: "column", At: 1}
	p.Beats[2].Carry = &CarryBeat{Show: "column", At: 0}
	err := validateCarryPlan(p)
	if err == nil {
		t.Fatal("columns worked out of order were accepted")
	}
	if !strings.Contains(err.Error(), "column 1") {
		t.Fatalf("the error does not name the offending column: %v", err)
	}
}

func TestCarryRejectsASkippedColumn(t *testing.T) {
	p := carryPlan()
	p.Beats[3].Carry = &CarryBeat{Show: "column", At: 3}
	if err := validateCarryPlan(p); err == nil {
		t.Fatal("a walk that skipped column 2 was accepted")
	}
}

func TestCarryRequiresEveryColumnWorked(t *testing.T) {
	p := carryPlan()
	// Drop the last column beat, leaving the final carry unexplained.
	p.Beats = append(p.Beats[:5], p.Beats[6])
	p.targetWords = 6 * 28
	err := validateCarryPlan(p)
	if err == nil {
		t.Fatal("a walk that leaves a column unworked was accepted")
	}
	if !strings.Contains(err.Error(), "4 of the 5 columns") {
		t.Fatalf("the error does not count the unworked columns: %v", err)
	}
}

func TestCarryRejectsACarryChainBeforeAnyColumn(t *testing.T) {
	p := carryPlan()
	p.Beats[1].Carry = &CarryBeat{Show: "carrychain"}
	if err := validateCarryPlan(p); err == nil {
		t.Fatal("a carry chain highlighted before a single carry existed was accepted")
	}
}

// Decoration on a binary literal — a 0b prefix, underscores, leading zeros — is
// a phrasing habit, not a wrong answer, so it is repaired rather than argued.
func TestCarryNormalizeStripsBitDecoration(t *testing.T) {
	p := carryPlan()
	p.Carry.A = "0b1011"
	p.Carry.B = " 01_10 "
	p.Carry.Sum = "0B10001"
	normalizeCarryPlan(p)
	if p.Carry.A != "1011" {
		t.Fatalf("a normalized to %q, want 1011", p.Carry.A)
	}
	if p.Carry.B != "110" {
		t.Fatalf("b normalized to %q, want 110", p.Carry.B)
	}
	if p.Carry.Sum != "10001" {
		t.Fatalf("sum normalized to %q, want 10001", p.Carry.Sum)
	}
	if err := validateCarryPlan(p); err != nil {
		t.Fatalf("a decorated-but-correct plan was rejected after normalize: %v", err)
	}
}

func TestCarryNormalizeClampsAnOutOfRangeColumn(t *testing.T) {
	p := carryPlan()
	p.Beats[5].Carry.At = 99
	normalizeCarryPlan(p)
	if got := p.Beats[5].Carry.At; got != 4 {
		t.Fatalf("column index clamped to %d, want 4", got)
	}
}

func TestCarryShowDefaultsToColumn(t *testing.T) {
	b := CarryBeat{Show: "sparkle"}
	if got := b.ResolvedShow(); got != "column" {
		t.Fatalf("an unknown show resolved to %q, want column", got)
	}
}

// The renderer never adds anything: every digit, every carry and both decimal
// equivalents arrive precomputed.
func TestCarryScenesPrecomputeEveryColumn(t *testing.T) {
	p := carryPlan()
	scenes, err := carryScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	if props["aDecimal"] != 11 || props["bDecimal"] != 6 || props["sumDecimal"] != 17 {
		t.Fatalf("the decimal check is wrong: %v %v %v", props["aDecimal"], props["bDecimal"], props["sumDecimal"])
	}

	cols, _ := props["columns"].([]map[string]any)
	if len(cols) != 5 {
		t.Fatalf("want 5 columns, got %d", len(cols))
	}
	// Column 0 is the rightmost: 1 plus 0 with nothing coming in.
	if cols[0]["a"] != "1" || cols[0]["b"] != "0" || cols[0]["carryIn"] != 0 || cols[0]["digit"] != "1" || cols[0]["carryOut"] != 0 {
		t.Fatalf("the ones column is wrong: %v", cols[0])
	}
	// Column 1 is where the first carry is born: 1 plus 1 is 10.
	if cols[1]["digit"] != "0" || cols[1]["carryOut"] != 1 {
		t.Fatalf("the twos column does not carry: %v", cols[1])
	}
	// Column 4 exists only because the addition overflowed into it.
	if cols[4]["a"] != "0" || cols[4]["b"] != "0" || cols[4]["carryIn"] != 1 || cols[4]["digit"] != "1" {
		t.Fatalf("the overflow column is wrong: %v", cols[4])
	}

	steps, _ := props["steps"].([]map[string]any)
	if steps[0]["show"] != "problem" {
		t.Fatalf("first step shows %v, want problem", steps[0]["show"])
	}
	if got := fmt.Sprint(steps[0]["done"]); got != "[]" {
		t.Fatalf("the opener has already worked columns: %v", got)
	}
	last := steps[len(steps)-1]
	if last["show"] != "answer" {
		t.Fatalf("last step shows %v, want answer", last["show"])
	}
	if got := fmt.Sprint(last["done"]); got != "[0 1 2 3 4]" {
		t.Fatalf("the closer has worked %v, want every column", got)
	}
}
