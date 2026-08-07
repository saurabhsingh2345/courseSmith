package pipeline

import (
	"strings"
	"testing"
)

const rxNarration = "Each lit column adds its weight, and the running total climbs to exactly the same value."

func radixPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "radix",
		Title:    "One number wearing three costumes",
		Radix: &RadixSpec{
			Decimal: 172,
			Binary:  "10101100",
			Hex:     "AC",
			Story:   "a byte you will meet in every subnet mask",
		},
		Beats: []SnippetBeat{
			{ID: "value", Heading: "The number", Narration: rxNarration, Radix: &RadixBeat{Show: "value"}},
			{ID: "weights", Heading: "The columns", Narration: rxNarration, Radix: &RadixBeat{Show: "weights"}},
			{ID: "sum", Heading: "The sum", Narration: rxNarration, Radix: &RadixBeat{Show: "sum"}},
			{ID: "hex", Heading: "The shorthand", Narration: rxNarration, Radix: &RadixBeat{Show: "hex"}},
			{ID: "same", Heading: "All three", Narration: rxNarration, Radix: &RadixBeat{Show: "same"}},
		},
	}
	// The template's ideal is 28 words per beat, so the fixture budget is
	// sized against that — nBeats * 40 would demand more beats than it has.
	p.targetWords = 5 * 28
	return p
}

func TestRadixPlanAccepted(t *testing.T) {
	if err := validateRadixPlan(radixPlan()); err != nil {
		t.Fatalf("a well-formed radix plan was rejected: %v", err)
	}
}

// The family's signature rule: the arithmetic is done in Go, and a wrong
// spelling is rejected with both numbers quoted.
func TestRadixRejectsBinaryThatDoesNotEqualTheDecimal(t *testing.T) {
	p := radixPlan()
	p.Radix.Binary = "11001100"
	err := validateRadixPlan(p)
	if err == nil {
		t.Fatal("a binary spelling of a different number was accepted")
	}
	if !strings.Contains(err.Error(), "204") || !strings.Contains(err.Error(), "172") {
		t.Fatalf("the error does not quote both values: %v", err)
	}
}

func TestRadixRejectsHexThatDoesNotEqualTheDecimal(t *testing.T) {
	p := radixPlan()
	p.Radix.Hex = "CA"
	err := validateRadixPlan(p)
	if err == nil {
		t.Fatal("a hex spelling of a different number was accepted")
	}
	if !strings.Contains(err.Error(), "202") || !strings.Contains(err.Error(), "172") {
		t.Fatalf("the error does not quote both values: %v", err)
	}
}

func TestRadixRejectsBinaryWithForeignDigits(t *testing.T) {
	p := radixPlan()
	p.Radix.Binary = "10102100"
	if err := validateRadixPlan(p); err == nil {
		t.Fatal("a binary string with a 2 in it was accepted")
	}
}

func TestRadixRejectsAValueOutOfRange(t *testing.T) {
	p := radixPlan()
	p.Radix.Decimal = 0
	p.Radix.Binary = "0"
	p.Radix.Hex = "0"
	if err := validateRadixPlan(p); err == nil {
		t.Fatal("zero was accepted, and zero has no set bits for the sum beat to light")
	}
}

func TestRadixRequiresOpeningOnTheValue(t *testing.T) {
	p := radixPlan()
	p.Beats[0].Radix = &RadixBeat{Show: "weights"}
	p.Beats[1].Radix = &RadixBeat{Show: "value"}
	if err := validateRadixPlan(p); err == nil {
		t.Fatal("a clip that raises the columns before showing the number was accepted")
	}
}

func TestRadixRequiresWeightsBeforeSum(t *testing.T) {
	p := radixPlan()
	p.Beats[1].Radix = &RadixBeat{Show: "sum"}
	p.Beats[2].Radix = &RadixBeat{Show: "weights"}
	err := validateRadixPlan(p)
	if err == nil {
		t.Fatal("a sum before the place-value weights was accepted")
	}
	if !strings.Contains(err.Error(), "weights") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRadixRequiresHexAfterSum(t *testing.T) {
	p := radixPlan()
	p.Beats[2].Radix = &RadixBeat{Show: "hex"}
	p.Beats[3].Radix = &RadixBeat{Show: "sum"}
	if err := validateRadixPlan(p); err == nil {
		t.Fatal("hex landing before the bits have summed was accepted")
	}
}

func TestRadixRequiresClosingOnSame(t *testing.T) {
	p := radixPlan()
	p.Beats[4].Radix = &RadixBeat{Show: "sum"}
	if err := validateRadixPlan(p); err == nil {
		t.Fatal("a clip that does not close on all three representations was accepted")
	}
}

// Decoration on a digit string — a 0b prefix, underscores, spaces — is a
// phrasing habit, not a wrong answer, so it is repaired rather than argued.
func TestRadixNormalizeStripsDigitDecoration(t *testing.T) {
	p := radixPlan()
	p.Radix.Binary = "0b1010_1100"
	p.Radix.Hex = "0xac"
	normalizeRadixPlan(p)
	if p.Radix.Binary != "10101100" {
		t.Fatalf("binary normalized to %q, want 10101100", p.Radix.Binary)
	}
	if p.Radix.Hex != "AC" {
		t.Fatalf("hex normalized to %q, want AC", p.Radix.Hex)
	}
	if err := validateRadixPlan(p); err != nil {
		t.Fatalf("a decorated-but-correct plan was rejected after normalize: %v", err)
	}
}

func TestRadixShowDefaultsToSum(t *testing.T) {
	b := RadixBeat{Show: "sparkle"}
	if got := b.ResolvedShow(); got != "sum" {
		t.Fatalf("an unknown show resolved to %q, want sum", got)
	}
}

// The renderer never does base conversion: the weights, the running total and
// the nibble grouping all arrive precomputed.
func TestRadixScenesPrecomputeTheArithmetic(t *testing.T) {
	p := radixPlan()
	scenes, err := radixScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	cells, _ := props["cells"].([]map[string]any)
	if len(cells) != 8 {
		t.Fatalf("want 8 bit cells, got %d", len(cells))
	}
	if cells[0]["weight"] != 128 || cells[0]["bit"] != "1" {
		t.Fatalf("the MSB cell is wrong: %v", cells[0])
	}

	sum, _ := props["sumSteps"].([]map[string]any)
	if len(sum) != 4 {
		t.Fatalf("172 has 4 set bits, got %d sum steps", len(sum))
	}
	if sum[0]["weight"] != 128 || sum[0]["total"] != 128 {
		t.Fatalf("first sum step is wrong: %v", sum[0])
	}
	if sum[3]["total"] != 172 {
		t.Fatalf("the running total does not end on the decimal: %v", sum[3])
	}

	nibbles, _ := props["nibbles"].([]map[string]any)
	if len(nibbles) != 2 {
		t.Fatalf("want 2 nibbles, got %d", len(nibbles))
	}
	if nibbles[0]["bits"] != "1010" || nibbles[0]["hex"] != "A" {
		t.Fatalf("first nibble is wrong: %v", nibbles[0])
	}
	if nibbles[1]["bits"] != "1100" || nibbles[1]["hex"] != "C" {
		t.Fatalf("second nibble is wrong: %v", nibbles[1])
	}

	steps, _ := props["steps"].([]map[string]any)
	if steps[0]["show"] != "value" {
		t.Fatalf("first step shows %v, want value", steps[0]["show"])
	}
	if steps[len(steps)-1]["show"] != "same" {
		t.Fatalf("last step shows %v, want same", steps[len(steps)-1]["show"])
	}
}
