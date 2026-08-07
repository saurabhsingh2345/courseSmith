package pipeline

// The gates template: one logic gate, wired up and fired.
//
// A logic gate is the smallest thing in computing that is genuinely a machine,
// and it is where a foundations course either lands or does not. The usual
// teaching aid is a truth table printed in a book: four rows of noughts and
// ones, correct, complete and dead. It states the gate's behaviour without ever
// showing the gate BEHAVING, so a viewer memorises four rows and still cannot
// say what the thing does.
//
// So the picture is the circuit, and the table is its receipt. The gate sits in
// the middle with its input wires on the left and its output on the right, and
// each beat sets the input pins to one combination and lets the signal
// propagate: the live wires brighten, the body lights, the output pin answers,
// and one row of the table on the right ticks in. By the last row the table has
// been FILLED IN by the circuit rather than asserted next to it, which is the
// entire difference between the two pictures.
//
// The clip closes on the gate's law — one line, "one when the inputs differ" —
// which is the sentence the table has just proved and the thing worth
// remembering.
//
// The validators compute, because a truth table that is wrong is worse than no
// truth table: it is a reference the viewer will trust.
//
// **Every row's output is recomputed in Go.** The gate function is evaluated
// over the row's own inputs and must agree, and a mismatch is rejected naming
// the row — "AND(1,0) is 0, the table claims 1". A model writing NAND from
// memory inverts one row confidently.
//
// **The table must be complete and non-repeating.** Exactly two-to-the-N rows,
// every input combination present exactly once, compared as bit patterns. A
// truth table missing a row is not a shorter truth table, it is a wrong one —
// the whole claim of the form is that it is exhaustive.
//
// **NOT takes one input; every other gate takes two.** An inverter with two
// wires is a drawing of a gate that does not exist.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "gates",
		Category:    CatConcepts,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The logic circuit",
		Description: "One gate drawn as a circuit, its inputs toggling and the signal glow propagating while the truth table fills in row by row. Reach for it when the subject is boolean logic itself — what AND, OR, NOT and XOR actually do, and why a truth table is a complete description.",
		Example:     "The XOR gate: different in, one out",
		PromptFile:  snippetGatesTemplateName,
		NeedsCode:   false,
		// The circuit, four firings and the law. Under thirty-five seconds a row
		// fires before the previous glow has finished crossing the wire.
		MinTargetSec:     35,
		DefaultTargetSec: 50,
		// Opener + up to four rows + closer is six; the ceiling leaves room for a
		// two-input gate to take a beat or two more over the rows that surprise.
		MaxBeats: 10,
		// A beat is a SHOT — one combination firing through the circuit — not a
		// step in an argument. Twenty-eight words is about nine seconds, which is
		// as long as one lit path holds anybody.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Gates: true},
		OwnsPlan:          planFields{Gates: true},
		Normalize:         normalizeGatesPlan,
		Validate:          validateGatesPlan,
		Scenes:            gatesScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(GatesShows(), ", "),
				"Kinds":         strings.Join(GatesKinds(), ", "),
				"MaxInputs":     maxGatesInputs,
				"MaxLabelWords": maxGatesLabelWords,
			}
		},
	})
}

const snippetGatesTemplateName = "snippet_gates.tmpl"

const (
	// Two inputs is every gate in this vocabulary except NOT, and it is also the
	// widest table the clip can fire exhaustively: three inputs is eight rows,
	// which is eight beats of firing before the law even lands.
	maxGatesInputs = 2
	// An input label is a letter on a pin — "A", "B", "carry in". Two words is
	// the most that fits beside a wire.
	maxGatesLabelWords = 2
)

// gatesKinds is the closed vocabulary of gate types, and it is closed because
// the renderer draws the distinctive IEEE body shape for each one. An invented
// gate has no silhouette, and a gate with no silhouette is a rectangle with a
// word in it, which is the picture this template exists to replace.
var gatesKinds = map[string]bool{
	"AND":  true,
	"OR":   true,
	"NOT":  true,
	"XOR":  true,
	"NAND": true,
	"NOR":  true,
}

// GatesKinds returns the gate vocabulary sorted.
func GatesKinds() []string {
	out := make([]string, 0, len(gatesKinds))
	for k := range gatesKinds {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// gatesLaws is each gate's one-line rule, stated in Go rather than asked for,
// because the closer is a FACT about the gate and not a thing a plan should be
// able to get wrong. It is also the sentence the truth table has just proved,
// so it has to agree with gateOutput by construction.
var gatesLaws = map[string]string{
	"AND":  "one only when both inputs are one",
	"OR":   "one when either input is one",
	"NOT":  "the output is always the opposite",
	"XOR":  "one when the inputs differ",
	"NAND": "zero only when both inputs are one",
	"NOR":  "one only when both inputs are zero",
}

// gatesShows is the closed vocabulary of what a beat does.
var gatesShows = map[string]bool{
	// The gate with its wires, nothing energised. The opener.
	"circuit": true,
	// Row At fires: the pins are set, the glow crosses the gate, the output
	// lights and the table row ticks in.
	"row": true,
	// The gate's one-line rule, with the finished table beside it. The closer.
	"law": true,
}

// GatesShows returns the beat vocabulary sorted.
func GatesShows() []string {
	out := make([]string, 0, len(gatesShows))
	for k := range gatesShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// GatesSpec is the circuit and its truth table. On the plan because one gate
// persists for the whole clip; the beats only change what is energised.
type GatesSpec struct {
	// Gate is the gate type, from gatesKinds. The star of the clip.
	Gate string `json:"gate"`
	// Inputs are the pin labels, left to right — ["A", "B"].
	Inputs []string `json:"inputs"`
	// Rows are the truth table, in the order the clip fires them.
	Rows []GatesRow `json:"rows"`
}

// GatesRow is one line of the truth table: one input combination and what the
// gate does with it. Recomputed in Go.
type GatesRow struct {
	// In is one 0 or 1 per input, in Inputs order.
	In []int `json:"in"`
	// Out is the claimed output.
	Out int `json:"out"`
}

// GatesBeat is one shot: which combination this beat fires.
type GatesBeat struct {
	// Show is a gatesShows name.
	Show string `json:"show"`
	// At indexes GatesSpec.Rows, for a "row" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a firing,
// which is what most beats of this template do.
func (b GatesBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if gatesShows[s] {
		return s
	}
	return "row"
}

// ResolvedGate returns the gate type uppercased and trimmed. Unlike a show,
// an unrecognised gate is NOT defaulted — the gate is the subject of the clip,
// and quietly turning an invented one into an AND would ship a video about a
// gate nobody asked for. Validation rejects it instead.
func (g GatesSpec) ResolvedGate() string {
	return strings.ToUpper(strings.TrimSpace(g.Gate))
}

// gateOutput evaluates the gate. This is the arithmetic the whole template
// rests on, and it lives here — in Go, once — so the validator, the law line
// and the table on screen cannot disagree.
func gateOutput(gate string, in []int) int {
	if gate == "NOT" {
		if len(in) < 1 {
			return 0
		}
		if in[0] == 1 {
			return 0
		}
		return 1
	}
	if len(in) < 2 {
		return 0
	}
	a, b := in[0], in[1]
	switch gate {
	case "AND":
		return a & b
	case "OR":
		return a | b
	case "XOR":
		return a ^ b
	case "NAND":
		return 1 - (a & b)
	case "NOR":
		return 1 - (a | b)
	}
	return 0
}

// gatesPattern packs a row's inputs into one integer, most significant input
// first, so "is every combination present exactly once" is a set comparison
// rather than a nested loop over slices.
func gatesPattern(in []int) int {
	n := 0
	for _, v := range in {
		n = n<<1 | (v & 1)
	}
	return n
}

// gatesCombination spells a pattern back out the way the table writes it, so a
// missing row can be named rather than merely counted.
func gatesCombination(labels []string, pattern int) string {
	parts := make([]string, len(labels))
	for i := range labels {
		bit := (pattern >> (len(labels) - 1 - i)) & 1
		parts[i] = fmt.Sprintf("%s=%d", labels[i], bit)
	}
	return strings.Join(parts, ", ")
}

// gatesCall spells one evaluation the way the rejection quotes it — "AND(1, 0)".
func gatesCall(gate string, in []int) string {
	parts := make([]string, len(in))
	for i, v := range in {
		parts[i] = fmt.Sprintf("%d", v)
	}
	return gate + "(" + strings.Join(parts, ", ") + ")"
}

func normalizeGatesPlan(p *SnippetPlan) {
	g := p.Gates
	if g == nil {
		return
	}
	g.Gate = g.ResolvedGate()

	inputs := make([]string, 0, len(g.Inputs))
	for _, label := range g.Inputs {
		if len(inputs) >= maxGatesInputs {
			break
		}
		inputs = append(inputs, clampWords(collapseSpaces(label), maxGatesLabelWords))
	}
	g.Inputs = inputs

	for i := range p.Beats {
		b := p.Beats[i].Gates
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "row" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(g.Rows); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateGatesPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Gates: true}); err != nil {
		return err
	}

	g := p.Gates
	if g == nil {
		return fmt.Errorf("the plan has no gate — this template is one logic gate wired up and fired, so the gate IS the clip")
	}
	gate := g.ResolvedGate()
	if !gatesKinds[gate] {
		return fmt.Errorf("the gate is %q, which is not one of: %s. The renderer draws each of those as its own IEEE silhouette — a D-shaped body for AND, a curved back for OR, a bubble for the inverting ones — and a gate outside the list has no shape to be drawn as",
			g.Gate, strings.Join(GatesKinds(), ", "))
	}

	wantInputs := 2
	if gate == "NOT" {
		wantInputs = 1
	}
	if len(g.Inputs) != wantInputs {
		return fmt.Errorf("the %s gate is drawn with %d input pin(s), and it takes exactly %d. An inverter with two wires, or an AND with one, is a drawing of a gate that does not exist",
			gate, len(g.Inputs), wantInputs)
	}
	seenLabel := map[string]bool{}
	for i, label := range g.Inputs {
		if strings.TrimSpace(label) == "" {
			return fmt.Errorf("input %d has no label. Every pin is named in the truth-table header, so an unlabeled wire leaves a column with no heading", i)
		}
		key := strings.ToLower(strings.TrimSpace(label))
		if seenLabel[key] {
			return fmt.Errorf("two input pins are both called %q. The table columns are told apart by their names, so a repeated label makes the rows unreadable", label)
		}
		seenLabel[key] = true
	}

	wantRows := 1 << len(g.Inputs)
	if len(g.Rows) != wantRows {
		return fmt.Errorf("the truth table has %d row(s), and a %d-input gate has exactly %d. A truth table missing a row is not a shorter table, it is a wrong one — the whole claim of the form is that it covers every case",
			len(g.Rows), len(g.Inputs), wantRows)
	}

	// THE TRUTH TABLE, recomputed. A wrong table is worse than none: it is a
	// reference the viewer will trust.
	seenPattern := map[int]int{}
	for i, row := range g.Rows {
		if len(row.In) != len(g.Inputs) {
			return fmt.Errorf("row %d gives %d input value(s) for a gate with %d pin(s). Every row sets every pin — that is what a row of a truth table is",
				i, len(row.In), len(g.Inputs))
		}
		for j, v := range row.In {
			if v != 0 && v != 1 {
				return fmt.Errorf("row %d sets pin %q to %d. A wire is either carrying a one or it is not, so every input value is 0 or 1", i, g.Inputs[j], v)
			}
		}
		if row.Out != 0 && row.Out != 1 {
			return fmt.Errorf("row %d claims the output is %d. The output pin lights or it does not, so it is 0 or 1", i, row.Out)
		}
		if got := gateOutput(gate, row.In); got != row.Out {
			return fmt.Errorf("row %d is wrong: %s is %d, and the table claims %d. The glow on screen is propagated from the inputs through the real gate function, so this row would light the output pin against its own number",
				i, gatesCall(gate, row.In), got, row.Out)
		}
		pattern := gatesPattern(row.In)
		if prev, dup := seenPattern[pattern]; dup {
			return fmt.Errorf("rows %d and %d are the same combination (%s). Firing it twice spends a beat on a case the viewer has already watched, and leaves another case with no beat at all",
				prev, i, gatesCombination(g.Inputs, pattern))
		}
		seenPattern[pattern] = i
	}
	for pattern := 0; pattern < wantRows; pattern++ {
		if _, ok := seenPattern[pattern]; !ok {
			return fmt.Errorf("the combination %s never appears in the table. A truth table is exhaustive by definition — every one of the %d combinations gets exactly one row",
				gatesCombination(g.Inputs, pattern), wantRows)
		}
	}

	// The shape: the circuit, then every row fired in table order, then the law.
	if p.Beats[0].Gates == nil || p.Beats[0].Gates.ResolvedShow() != "circuit" {
		return fmt.Errorf("beat %q does not open on the circuit. A signal glowing through a gate the viewer has not seen drawn is a shape flashing — the first beat shows the gate and its wires with nothing energised",
			p.Beats[0].ID)
	}
	last := p.Beats[len(p.Beats)-1]
	if last.Gates == nil || last.Gates.ResolvedShow() != "law" {
		return fmt.Errorf("beat %q does not close on the law. The rows prove a one-line rule and the clip has to say it — end with {\"show\": \"law\"}, which is the sentence the viewer keeps",
			last.ID)
	}

	next := 0
	for i, b := range p.Beats {
		if b.Gates == nil {
			return fmt.Errorf("beat %q has no gates direction — every beat shows the cold circuit, fires one row, or states the law", b.ID)
		}
		switch b.Gates.ResolvedShow() {
		case "circuit":
			if i != 0 {
				return fmt.Errorf("beat %q goes back to the cold circuit part-way through. The circuit is the opener; de-energising it mid-table throws away rows the viewer just watched fire", b.ID)
			}
		case "law":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q states the law before the table is filled. The law is the closer — it is the sentence the firings earn, and stating it early makes the rows a formality", b.ID)
			}
		case "row":
			if next >= len(g.Rows) {
				return fmt.Errorf("beat %q fires another row, but the table only has %d — rows 0 to %d", b.ID, len(g.Rows), len(g.Rows)-1)
			}
			if b.Gates.At != next {
				return fmt.Errorf("beat %q fires row %d when the table is at row %d. The rows fire in the order they are written, top to bottom, because the table fills in as they go and a row ticking in above an empty one reads as a skipped case",
					b.ID, b.Gates.At, next)
			}
			next++
		}
	}
	if next != len(g.Rows) {
		return fmt.Errorf("only %d of the %d rows are ever fired. A row printed in the table but never propagated through the circuit is back to being asserted rather than shown, which is the picture this template replaces — give each row its own beat",
			next, len(g.Rows))
	}
	return nil
}

// gatesScenes lays the clip out as ONE scene. The gate function, the law line
// and every row's bit pattern are resolved here, so the component only decides
// which wires are hot and how much of the table has arrived.
func gatesScenes(in SnippetSceneInput) ([]Scene, error) {
	g := in.Plan.Gates
	if g == nil {
		return nil, fmt.Errorf("the plan has no gate")
	}
	gate := g.ResolvedGate()

	rows := make([]map[string]any, len(g.Rows))
	for i, row := range g.Rows {
		values := make([]int, len(row.In))
		copy(values, row.In)
		rows[i] = map[string]any{
			"in":  values,
			"out": gateOutput(gate, row.In),
		}
	}
	inputs := make([]string, len(g.Inputs))
	copy(inputs, g.Inputs)

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	// Which rows have ticked into the table by each beat, accumulated in Go so
	// the renderer never replays the beat list to find out.
	fired := map[int]bool{}
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Gates == nil {
			return nil, fmt.Errorf("beat %q has no gates direction", beat.ID)
		}
		show := beat.Gates.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "row" {
			fired[beat.Gates.At] = true
			step["at"] = beat.Gates.At
		}
		done := make([]int, 0, len(fired))
		for at := range fired {
			done = append(done, at)
		}
		sort.Ints(done)
		step["done"] = done
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneGates,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"gate":   gate,
			"law":    gatesLaws[gate],
			"inputs": inputs,
			"rows":   rows,
			"steps":  steps,
		}),
	}}, nil
}
