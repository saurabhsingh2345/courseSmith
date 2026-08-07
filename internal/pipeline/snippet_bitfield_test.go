package pipeline

import (
	"fmt"
	"strings"
	"testing"
)

const bfNarration = "These thirty two bits are not one number, they are three fields glued together."

// The IEEE 754 single-precision bits of pi: one sign bit, eight exponent bits
// and twenty three mantissa bits, which is the layout this template exists for.
const bfBits = "01000000010010010000111111011011"

func bitfieldPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "bitfield",
		Title:    "A float is three numbers",
		Bitfield: &BitfieldSpec{
			Bits: bfBits,
			Fields: []BitfieldField{
				{Label: "sign", From: 0, To: 0, Means: "zero, so the number is positive"},
				{Label: "exponent", From: 1, To: 8, Means: "128, which after the bias means times two"},
				{Label: "mantissa", From: 9, To: 31, Means: "the fraction, roughly one point five seven"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "row", Heading: "The row", Narration: bfNarration, Bitfield: &BitfieldBeat{Show: "row"}},
			{ID: "split", Heading: "The boundaries", Narration: bfNarration, Bitfield: &BitfieldBeat{Show: "split"}},
			{ID: "sign", Heading: "The sign bit", Narration: bfNarration, Bitfield: &BitfieldBeat{Show: "field", At: 0}},
			{ID: "exponent", Heading: "The exponent", Narration: bfNarration, Bitfield: &BitfieldBeat{Show: "field", At: 1}},
			{ID: "mantissa", Heading: "The mantissa", Narration: bfNarration, Bitfield: &BitfieldBeat{Show: "field", At: 2}},
			{ID: "read", Heading: "Reading it back", Narration: bfNarration, Bitfield: &BitfieldBeat{Show: "read"}},
		},
	}
	// The template's ideal is 28 words per beat, so the fixture budget is sized
	// against that — nBeats * 40 would demand more beats than it has.
	p.targetWords = 6 * 28
	return p
}

func TestBitfieldPlanAccepted(t *testing.T) {
	if err := validateBitfieldPlan(bitfieldPlan()); err != nil {
		t.Fatalf("a well-formed bitfield plan was rejected: %v", err)
	}
}

// The family's signature rule, in interval form: the fields are covered bit by
// bit in Go and the exact uncovered positions are named.
func TestBitfieldRejectsAGapBetweenFields(t *testing.T) {
	p := bitfieldPlan()
	p.Bitfield.Fields[2].From = 10
	err := validateBitfieldPlan(p)
	if err == nil {
		t.Fatal("a layout that leaves bit 9 in no field was accepted")
	}
	if !strings.Contains(err.Error(), "bit(s) 9 belong to no field") {
		t.Fatalf("the error does not name the uncovered bit: %v", err)
	}
}

func TestBitfieldRejectsOverlappingFields(t *testing.T) {
	p := bitfieldPlan()
	p.Bitfield.Fields[2].From = 8
	err := validateBitfieldPlan(p)
	if err == nil {
		t.Fatal("a layout with bit 8 in two fields was accepted")
	}
	if !strings.Contains(err.Error(), "bit(s) 8 belong to more than one field") {
		t.Fatalf("the error does not name the doubly covered bit: %v", err)
	}
}

func TestBitfieldRejectsARowThatIsNotAMachineWidth(t *testing.T) {
	p := bitfieldPlan()
	p.Bitfield.Bits = bfBits[:31]
	p.Bitfield.Fields[2].To = 30
	err := validateBitfieldPlan(p)
	if err == nil {
		t.Fatal("a thirty-one bit row was accepted")
	}
	if !strings.Contains(err.Error(), "31 bits wide") {
		t.Fatalf("the error does not quote the width: %v", err)
	}
}

func TestBitfieldRejectsANonBinaryCell(t *testing.T) {
	p := bitfieldPlan()
	p.Bitfield.Bits = "2" + bfBits[1:]
	if err := validateBitfieldPlan(p); err == nil {
		t.Fatal("a row with a 2 in it was accepted")
	}
}

func TestBitfieldRejectsABackwardsField(t *testing.T) {
	p := bitfieldPlan()
	p.Bitfield.Fields[1].From = 8
	p.Bitfield.Fields[1].To = 1
	if err := validateBitfieldPlan(p); err == nil {
		t.Fatal("a field running from bit 8 back to bit 1 was accepted")
	}
}

func TestBitfieldRejectsAFieldPastTheEndOfTheRow(t *testing.T) {
	p := bitfieldPlan()
	p.Bitfield.Fields[2].To = 32
	if err := validateBitfieldPlan(p); err == nil {
		t.Fatal("a field running past the last bit was accepted")
	}
}

func TestBitfieldRequiresOpeningOnThePlainRow(t *testing.T) {
	p := bitfieldPlan()
	p.Beats[0].Bitfield = &BitfieldBeat{Show: "split"}
	p.Beats[1].Bitfield = &BitfieldBeat{Show: "row"}
	if err := validateBitfieldPlan(p); err == nil {
		t.Fatal("a clip that splits the row before showing it undivided was accepted")
	}
}

func TestBitfieldRequiresTheSplitBeforeAnyField(t *testing.T) {
	p := bitfieldPlan()
	p.Beats[1].Bitfield = &BitfieldBeat{Show: "field", At: 0}
	p.Beats[2].Bitfield = &BitfieldBeat{Show: "split"}
	err := validateBitfieldPlan(p)
	if err == nil {
		t.Fatal("a field lifting out of an undivided row was accepted")
	}
	if !strings.Contains(err.Error(), "split") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBitfieldRequiresEveryFieldLiftedOnce(t *testing.T) {
	p := bitfieldPlan()
	p.Beats[4].Bitfield = &BitfieldBeat{Show: "field", At: 1}
	err := validateBitfieldPlan(p)
	if err == nil {
		t.Fatal("a clip that lifts the exponent twice and never the mantissa was accepted")
	}
	if !strings.Contains(err.Error(), "exponent") {
		t.Fatalf("the error does not name the repeated field: %v", err)
	}
}

func TestBitfieldRequiresClosingOnTheRead(t *testing.T) {
	p := bitfieldPlan()
	p.Beats[5].Bitfield = &BitfieldBeat{Show: "split"}
	if err := validateBitfieldPlan(p); err == nil {
		t.Fatal("a clip that never reads the pattern back was accepted")
	}
}

// Decoration on a bit string — a 0b prefix, the pipes and spaces a model uses
// to show the field boundaries in text — is a phrasing habit, not a wrong
// answer, so it is repaired rather than argued.
func TestBitfieldNormalizeStripsBitDecoration(t *testing.T) {
	p := bitfieldPlan()
	p.Bitfield.Bits = "0b0|10000000|10010010000111111011011"
	normalizeBitfieldPlan(p)
	if p.Bitfield.Bits != bfBits {
		t.Fatalf("bits normalized to %q, want %q", p.Bitfield.Bits, bfBits)
	}
	if err := validateBitfieldPlan(p); err != nil {
		t.Fatalf("a decorated-but-correct plan was rejected after normalize: %v", err)
	}
}

func TestBitfieldNormalizeClampsLabelsAndIndices(t *testing.T) {
	p := bitfieldPlan()
	p.Bitfield.Fields[0].Label = "the  sign  bit  of  the  float"
	p.Beats[4].Bitfield.At = 99
	normalizeBitfieldPlan(p)
	if got := p.Bitfield.Fields[0].Label; got != "the sign bit" {
		t.Fatalf("label normalized to %q, want three words", got)
	}
	if got := p.Beats[4].Bitfield.At; got != 2 {
		t.Fatalf("field index clamped to %d, want 2", got)
	}
}

func TestBitfieldShowDefaultsToField(t *testing.T) {
	b := BitfieldBeat{Show: "sparkle"}
	if got := b.ResolvedShow(); got != "field" {
		t.Fatalf("an unknown show resolved to %q, want field", got)
	}
}

// The renderer never reads a bit range: each field arrives with its own bits
// already sliced out and their unsigned value already computed.
func TestBitfieldScenesSliceEveryField(t *testing.T) {
	p := bitfieldPlan()
	scenes, err := bitfieldScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	cells, _ := props["cells"].([]map[string]any)
	if len(cells) != 32 {
		t.Fatalf("want 32 bit cells, got %d", len(cells))
	}

	fields, _ := props["fields"].([]map[string]any)
	if len(fields) != 3 {
		t.Fatalf("want 3 fields, got %d", len(fields))
	}
	if fields[0]["bits"] != "0" || fields[0]["value"] != 0 {
		t.Fatalf("the sign field is wrong: %v", fields[0])
	}
	if fields[1]["bits"] != "10000000" || fields[1]["value"] != 128 {
		t.Fatalf("the exponent field is wrong: %v", fields[1])
	}
	if fields[2]["bits"] != "10010010000111111011011" {
		t.Fatalf("the mantissa field is wrong: %v", fields[2])
	}

	steps, _ := props["steps"].([]map[string]any)
	if steps[0]["show"] != "row" {
		t.Fatalf("first step shows %v, want row", steps[0]["show"])
	}
	if got := fmt.Sprint(steps[0]["done"]); got != "[]" {
		t.Fatalf("the opener has already explained fields: %v", got)
	}
	last := steps[len(steps)-1]
	if last["show"] != "read" {
		t.Fatalf("last step shows %v, want read", last["show"])
	}
	if got := fmt.Sprint(last["done"]); got != "[0 1 2]" {
		t.Fatalf("the closer has explained %v, want every field", got)
	}
}
