package pipeline

// The radix template: the same number, three ways.
//
// Base conversion is the first wall in every CS foundations course, and the
// standard teaching aid — a conversion PROCEDURE, divide by two and keep the
// remainders — is exactly the wrong picture. It teaches a ritual and hides the
// fact, which is that 172, 10101100 and AC are not three numbers related by a
// recipe; they are one number written three ways. So this template never
// converts anything. It puts one value on screen in all three bases and spends
// its beats showing WHY they agree: the place-value columns light up and their
// weights sum, on screen, to the decimal.
//
// The whole clip is an equation, which is why the validator does the equation.
// A model writing from memory transposes digits — "10101100 is 204" is a
// mistake it makes fluently and confidently — and a base-conversion diagram
// with wrong digits is worse than no diagram: it teaches the viewer that the
// picture is decoration rather than arithmetic. So the binary is parsed as
// base 2, the hex as base 16, and both must equal the decimal exactly, with
// the real values quoted back in the rejection.
//
// The beat order is the pedagogy. The value comes first because the columns
// mean nothing until there is a number to explain. The weights appear before
// any bit lights, because a bit lighting under a column the viewer cannot
// price is just a cell changing colour. The hex comes after the sum, because
// hex is a shorthand FOR the bits and reads as arbitrary until the bits have
// earned their meaning. And the clip closes on all three lit at once — the
// frame somebody screenshots.

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "radix",
		Category:    CatNumbers,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The same number, three ways",
		Description: "One value shown in decimal, binary and hex at once, with the place-value columns summing to it on screen. Reach for it when the subject is number bases themselves — why 172 IS 10101100, what a hex digit is a shorthand for.",
		Example:     "Why 172 is 10101100 in binary and AC in hex",
		PromptFile:  snippetRadixTemplateName,
		NeedsCode:   false,
		// The value, the columns, the sum ticking up, the nibbles, the whole:
		// five distinct states, and under thirty-five seconds they cannot each
		// hold long enough to be read.
		MinTargetSec:     35,
		DefaultTargetSec: 50,
		// Opener + weights + a sum that may take two beats + hex + closer, with
		// a little slack. Past eight the same picture is being re-shown.
		MaxBeats: 8,
		// A beat here is a SHOT — a state change on one diagram — not a step in
		// an argument. Twenty-eight words is about nine seconds, which is as
		// long as one lighting-up holds anybody.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Radix: true},
		OwnsPlan:          planFields{Radix: true},
		Normalize:         normalizeRadixPlan,
		Validate:          validateRadixPlan,
		Scenes:            radixScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(RadixShows(), ", "),
				"MinDecimal":    minRadixDecimal,
				"MaxDecimal":    maxRadixDecimal,
				"MaxStoryWords": maxRadixStoryWords,
			}
		},
	})
}

const snippetRadixTemplateName = "snippet_radix.tmpl"

const (
	// Zero has no set bits, so the sum beat — the beat the template exists
	// for — would have nothing to light. One is the smallest number whose
	// columns do anything.
	minRadixDecimal = 1
	// Sixteen bits is where the cells stop being readable at the size the
	// weight labels need; 65535 is the largest value they can carry.
	maxRadixDecimal = 65535
	// The most bit cells the row can hold at a legible size — the string
	// bound that matches the value bound above, including leading zeros.
	maxRadixBits = 16

	// The story is a caption under a number, not a paragraph about it.
	maxRadixStoryWords = 10
)

// radixShows is the closed vocabulary of what a beat does.
var radixShows = map[string]bool{
	// The decimal alone, with its story. The opener.
	"value": true,
	// The binary row appears with its place-value weights above each cell.
	"weights": true,
	// The set bits light one by one and their weights tick up to the decimal.
	"sum": true,
	// The nibbles group under brackets and the hex digits land.
	"hex": true,
	// All three representations lit at once. The closer.
	"same": true,
}

// RadixShows returns the beat vocabulary sorted.
func RadixShows() []string {
	out := make([]string, 0, len(radixShows))
	for k := range radixShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// RadixSpec is one number and its three spellings. On the plan because the
// whole clip is about a single value that persists.
type RadixSpec struct {
	// Decimal is the value itself.
	Decimal int `json:"decimal"`
	// Binary is the same value in base 2 — "10101100". Checked, not trusted.
	Binary string `json:"binary"`
	// Hex is the same value in base 16 — "AC". Checked, not trusted.
	Hex string `json:"hex"`
	// Story says why this number — "the ASCII code for lowercase a".
	Story string `json:"story,omitempty"`
}

// RadixBeat is one shot: which state of the diagram this beat shows.
type RadixBeat struct {
	// Show is a radixShows name.
	Show string `json:"show"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to the sum —
// the workhorse state most beats of this template are in.
func (b RadixBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if radixShows[s] {
		return s
	}
	return "sum"
}

// cleanRadixDigits strips the decoration a model puts on a digit string —
// spaces, underscores, a 0b/0x prefix — and uppercases it, so "0b1010_1100"
// and "10101100" are the same answer rather than a spurious rejection.
func cleanRadixDigits(s, prefix string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "_", "")
	return strings.TrimPrefix(s, prefix)
}

func normalizeRadixPlan(p *SnippetPlan) {
	r := p.Radix
	if r == nil {
		return
	}
	r.Binary = cleanRadixDigits(r.Binary, "0B")
	r.Hex = cleanRadixDigits(r.Hex, "0X")
	r.Story = clampWords(collapseSpaces(r.Story), maxRadixStoryWords)
	for i := range p.Beats {
		if b := p.Beats[i].Radix; b != nil {
			b.Show = b.ResolvedShow()
		}
	}
}

func validateRadixPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Radix: true}); err != nil {
		return err
	}

	r := p.Radix
	if r == nil {
		return fmt.Errorf("the plan has no number — this template is one value spelled in decimal, binary and hex, so the value is the clip")
	}
	if r.Decimal < minRadixDecimal || r.Decimal > maxRadixDecimal {
		return fmt.Errorf("the decimal is %d, want %d-%d. Zero has no set bits so the sum beat has nothing to light, and past %d the bit row needs more than sixteen cells and stops being readable",
			r.Decimal, minRadixDecimal, maxRadixDecimal, maxRadixDecimal)
	}
	if strings.TrimSpace(r.Binary) == "" {
		return fmt.Errorf("the plan gives no binary spelling. The bit row is the middle of the picture — without it there are no columns to light")
	}
	if len(r.Binary) > maxRadixBits {
		return fmt.Errorf("the binary is %d digits; the row holds at most %d cells at a legible size. Drop the leading zeros", len(r.Binary), maxRadixBits)
	}
	binVal, err := strconv.ParseUint(r.Binary, 2, 64)
	if err != nil {
		return fmt.Errorf("the binary %q does not parse as base 2 — every digit must be 0 or 1", r.Binary)
	}
	// THE ARITHMETIC. A diagram of computing that is wrong is worse than none,
	// and transposed digits are the mistake a model makes fluently.
	if binVal != uint64(r.Decimal) {
		return fmt.Errorf("the binary %s is %d but the plan claims %d. This picture is an equation — the lit columns must sum to the decimal on screen — so the digits have to be the actual digits of %d",
			r.Binary, binVal, r.Decimal, r.Decimal)
	}
	if strings.TrimSpace(r.Hex) == "" {
		return fmt.Errorf("the plan gives no hex spelling. The nibble grouping is the payoff — hex as a shorthand for the bits — and without it the clip stops one base short")
	}
	hexVal, err := strconv.ParseUint(r.Hex, 16, 64)
	if err != nil {
		return fmt.Errorf("the hex %q does not parse as base 16", r.Hex)
	}
	if hexVal != uint64(r.Decimal) {
		return fmt.Errorf("the hex %s is %d but the plan claims %d. The nibble brackets will group the real bits, so a wrong hex digit would visibly disagree with its own bracket",
			r.Hex, hexVal, r.Decimal)
	}

	// The shape: value, then weights, then sum, then hex, closing on same.
	if p.Beats[0].Radix == nil || p.Beats[0].Radix.ResolvedShow() != "value" {
		return fmt.Errorf("beat %q does not open on the value alone. The columns mean nothing until there is a number to explain — open with {\"show\": \"value\"}",
			p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Radix == nil || last.Radix.ResolvedShow() != "same" {
		return fmt.Errorf("the clip does not close on all three representations lit. The final frame is the one the viewer keeps — end with {\"show\": \"same\"}")
	}
	firstAt := map[string]int{}
	for i, b := range p.Beats {
		if b.Radix == nil {
			return fmt.Errorf("beat %q has no radix direction — every beat shows one state of the diagram", b.ID)
		}
		show := b.Radix.ResolvedShow()
		if _, seen := firstAt[show]; !seen {
			firstAt[show] = i
		}
		if show == "value" && i != 0 {
			return fmt.Errorf("beat %q re-opens on the bare value part-way through. Once the columns are up they stay up", b.ID)
		}
		if show == "same" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q lights everything before the end. \"same\" is the closer — after it there is nothing left to show", b.ID)
		}
	}
	for _, need := range []string{"weights", "sum", "hex"} {
		if _, ok := firstAt[need]; !ok {
			return fmt.Errorf("no beat shows %q. The clip needs the columns to appear, the set bits to sum, and the nibbles to group — skip one and the equation on screen has a hole in it", need)
		}
	}
	if firstAt["weights"] > firstAt["sum"] {
		return fmt.Errorf("the bits start summing before the place-value weights appear. A bit lighting under a column the viewer cannot price is just a cell changing colour — put the weights beat first")
	}
	if firstAt["hex"] < firstAt["sum"] {
		return fmt.Errorf("the hex lands before the bits have summed. Hex is a shorthand FOR the bits and reads as arbitrary until they have earned their meaning — move the hex beat after the sum")
	}
	return nil
}

// radixScenes lays the clip out as ONE scene. All the arithmetic the renderer
// animates — cell weights, the running sum, the nibble grouping — is computed
// here, so the component never does base conversion in TypeScript.
func radixScenes(in SnippetSceneInput) ([]Scene, error) {
	r := in.Plan.Radix
	if r == nil {
		return nil, fmt.Errorf("the plan has no number")
	}

	bits := r.Binary
	// The bit cells, MSB first, each with its place value.
	cells := make([]map[string]any, len(bits))
	for i := range bits {
		cells[i] = map[string]any{
			"bit":    string(bits[i]),
			"weight": 1 << (len(bits) - 1 - i),
		}
	}
	// The sum, precomputed: each set bit's weight and the running total, in
	// the left-to-right order they light.
	sumSteps := make([]map[string]any, 0, len(bits))
	total := 0
	for i := range bits {
		if bits[i] != '1' {
			continue
		}
		total += 1 << (len(bits) - 1 - i)
		sumSteps = append(sumSteps, map[string]any{
			"at":     i,
			"weight": 1 << (len(bits) - 1 - i),
			"total":  total,
		})
	}
	// The nibbles: the bit row padded to a multiple of four and grouped, each
	// group with the hex digit it spells.
	pad := (4 - len(bits)%4) % 4
	padded := strings.Repeat("0", pad) + bits
	hexDigits := fmt.Sprintf("%0*X", len(padded)/4, r.Decimal)
	nibbles := make([]map[string]any, 0, len(padded)/4)
	for i := 0; i < len(padded); i += 4 {
		nibbles = append(nibbles, map[string]any{
			"bits": padded[i : i+4],
			"hex":  string(hexDigits[i/4]),
		})
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Radix == nil {
			return nil, fmt.Errorf("beat %q has no radix direction", beat.ID)
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Radix.ResolvedShow(),
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneRadix,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":    in.Plan.Title,
			"decimal":  r.Decimal,
			"story":    r.Story,
			"hex":      r.Hex,
			"cells":    cells,
			"sumSteps": sumSteps,
			"nibbles":  nibbles,
			"steps":    steps,
		}),
	}}, nil
}
