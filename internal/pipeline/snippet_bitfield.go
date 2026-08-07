package pipeline

// The bitfield template: what the bits MEAN.
//
// By the time a career-switcher can read 01000001 as sixty-five, they have
// learned the least interesting half of the lesson. Almost no real bit pattern
// is one number. A float is three numbers glued together, an IP address is four,
// a permission mask is nine flags in three groups, and a two's-complement byte
// is a sign bit that changes what all the others are worth. The bits do not
// carry meaning; the LAYOUT does, and the layout is a fact you cannot derive by
// staring at the row.
//
// So the picture is a single bit row that gets divided. It arrives undivided —
// deliberately, because that is how the viewer has always seen it, an
// undifferentiated wall of ones and zeros — and then brackets drop in and cut
// it into named fields, and each field lifts in turn and says what its own bits
// decode to. The payoff beat reads the whole thing back as one sentence. That
// arc, wall to structure to meaning, is the template.
//
// The validators do interval arithmetic, because a diagram of a layout that
// does not add up is worse than no diagram.
//
// **The fields must TILE the row exactly.** Every bit belongs to exactly one
// field: no gap, no overlap. A model asked for IEEE 754 will write exponent
// 1-8 and mantissa 10-31 and lose bit 9, and the resulting picture has a cell
// with a bracket over nothing — which quietly teaches that some bits are spare.
// Go marks every position, counts its covers, and names the exact bit indices
// that are uncovered or covered twice.
//
// **The row is 8, 16 or 32 bits.** Those are the widths that exist: a byte, a
// short, a word. An arbitrary width is a model inventing a machine, and it is
// also the width at which the cells stop being legible.
//
// **Every field is focused exactly once, and the split comes first.** A field
// drawn but never explained is a mystery the clip planted on purpose, and a
// field lifting out of a row that has not been divided yet is a group of cells
// brightening for no stated reason.

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "bitfield",
		Category:    CatConcepts,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "What the bits mean",
		Description: "One bit row cut into labeled fields by brackets, each field lifting to say what its own bits decode to. Reach for it when a pattern is a LAYOUT rather than a number — a float's sign and exponent and mantissa, an IPv4 address, rwx permission bits, two's complement.",
		Example:     "The 32 bits of a float: sign, exponent, mantissa",
		PromptFile:  snippetBitfieldTemplateName,
		NeedsCode:   false,
		// The undivided row, the split, one lift per field and the read-back:
		// under thirty-five seconds those states cannot each hold long enough for
		// the bit cells under them to actually be read.
		MinTargetSec:     35,
		DefaultTargetSec: 50,
		// Opener + split + up to five field lifts + closer. Past that the row has
		// more brackets than a viewer tracks in one picture.
		MaxBeats: 8,
		// A beat is a SHOT — one group of cells lifting — not a step in an
		// argument. Twenty-eight words is about nine seconds, which is as long as
		// a highlighted field holds anybody.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Bitfield: true},
		OwnsPlan:          planFields{Bitfield: true},
		Normalize:         normalizeBitfieldPlan,
		Validate:          validateBitfieldPlan,
		Scenes:            bitfieldScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(BitfieldShows(), ", "),
				"Widths":        strings.Join(BitfieldWidths(), ", "),
				"MinFields":     minBitfieldFields,
				"MaxFields":     maxBitfieldFields,
				"MaxLabelWords": maxBitfieldLabelWords,
				"MaxMeansWords": maxBitfieldMeansWords,
			}
		},
	})
}

const snippetBitfieldTemplateName = "snippet_bitfield.tmpl"

const (
	// One field is a row with a bracket round the whole of it, which is the
	// picture the template exists to replace.
	minBitfieldFields = 2
	// Five brackets is where the labels under the row start colliding, and it is
	// also the most lifts the eight-beat ceiling can fund alongside an opener, a
	// split and a closer.
	maxBitfieldFields = 5

	// A field label sits under a bracket a few hundred pixels wide: "exponent",
	// "second octet". Three words is a name; more is a caption.
	maxBitfieldLabelWords = 3
	// The meaning is one line beside the lifted field. Twelve words is a
	// decoding; more is a paragraph the viewer gets nine seconds to read.
	maxBitfieldMeansWords = 12
)

// bitfieldWidths is the closed set of row widths. These are the widths that
// exist on a real machine — a byte, a short, a word — and they are also the
// only ones whose cells stay legible across the stage.
var bitfieldWidths = map[int]bool{8: true, 16: true, 32: true}

// BitfieldWidths returns the legal row widths, sorted, for prompts and docs.
func BitfieldWidths() []string {
	out := make([]int, 0, len(bitfieldWidths))
	for w := range bitfieldWidths {
		out = append(out, w)
	}
	sort.Ints(out)
	s := make([]string, len(out))
	for i, w := range out {
		s[i] = strconv.Itoa(w)
	}
	return s
}

// bitfieldShows is the closed vocabulary of what a beat does.
var bitfieldShows = map[string]bool{
	// The undivided bit row, exactly as the viewer has always met it. The opener.
	"row": true,
	// The field boundaries drop in as brackets and the row becomes a structure.
	"split": true,
	// Field At lifts: its cells brighten and its meaning lands beside them.
	"field": true,
	// The whole decoded meaning read back in one line. The closer.
	"read": true,
}

// BitfieldShows returns the beat vocabulary sorted.
func BitfieldShows() []string {
	out := make([]string, 0, len(bitfieldShows))
	for k := range bitfieldShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// BitfieldSpec is the row and its layout. On the plan because one row persists
// for the whole clip; the beats only change what is bracketed and lit.
type BitfieldSpec struct {
	// Bits is the row itself — 8, 16 or 32 characters of 0 and 1.
	Bits string `json:"bits"`
	// Fields are the named regions the row divides into. They must tile it.
	Fields []BitfieldField `json:"fields"`
}

// BitfieldField is one named region of the row.
type BitfieldField struct {
	// Label names the region — "sign", "exponent", "second octet".
	Label string `json:"label"`
	// From and To are inclusive bit indices with the MOST SIGNIFICANT bit at
	// zero, on the left, because that is how every specification numbers a
	// layout and how the row is drawn.
	From int `json:"from"`
	To   int `json:"to"`
	// Means is what this field's value decodes to — "exponent 128, so 2 to the
	// power of one".
	Means string `json:"means"`
}

// BitfieldBeat is one shot: which field this beat lifts.
type BitfieldBeat struct {
	// Show is a bitfieldShows name.
	Show string `json:"show"`
	// At indexes BitfieldSpec.Fields, for a "field" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a field lift,
// which is what most beats of this template do.
func (b BitfieldBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if bitfieldShows[s] {
		return s
	}
	return "field"
}

// cleanBitfieldBits strips the decoration a model puts on a bit string —
// spaces, underscores, a 0b prefix — so "0b0100_0001" and "01000001" are the
// same answer rather than a spurious rejection about width.
func cleanBitfieldBits(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "|", "")
	return strings.TrimPrefix(s, "0B")
}

func normalizeBitfieldPlan(p *SnippetPlan) {
	bf := p.Bitfield
	if bf == nil {
		return
	}
	bf.Bits = cleanBitfieldBits(bf.Bits)

	fields := make([]BitfieldField, 0, len(bf.Fields))
	for _, f := range bf.Fields {
		f.Label = clampWords(collapseSpaces(f.Label), maxBitfieldLabelWords)
		f.Means = clampWords(collapseSpaces(f.Means), maxBitfieldMeansWords)
		if len(fields) < maxBitfieldFields {
			fields = append(fields, f)
		}
	}
	bf.Fields = fields

	for i := range p.Beats {
		b := p.Beats[i].Bitfield
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "field" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(bf.Fields); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateBitfieldPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Bitfield: true}); err != nil {
		return err
	}

	bf := p.Bitfield
	if bf == nil {
		return fmt.Errorf("the plan has no bit row — this template is one row of bits cut into named fields, so the row IS the clip")
	}
	if strings.TrimSpace(bf.Bits) == "" {
		return fmt.Errorf("the plan gives no bits. The row is the whole picture: without an actual pattern there is nothing for the brackets to divide and nothing for a field to decode")
	}
	if !bitfieldWidths[len(bf.Bits)] {
		return fmt.Errorf("the row %q is %d bits wide, and this template takes %s. Those are the widths that exist on a real machine — a byte, a short, a word — and an arbitrary width is a model inventing a layout nobody has to read",
			bf.Bits, len(bf.Bits), strings.Join(BitfieldWidths(), ", "))
	}
	for i := 0; i < len(bf.Bits); i++ {
		if bf.Bits[i] != '0' && bf.Bits[i] != '1' {
			return fmt.Errorf("the row %q has %q at position %d, and every cell has to hold a 0 or a 1. One character is one bit cell, so anything else has no cell to sit in",
				bf.Bits, string(bf.Bits[i]), i)
		}
	}
	if n := len(bf.Fields); n < minBitfieldFields || n > maxBitfieldFields {
		return fmt.Errorf("the row is cut into %d field(s), want %d-%d. One field is a bracket round the whole row, which is the picture this template exists to replace; past %d the labels under the row collide and every field still needs its own beat",
			n, minBitfieldFields, maxBitfieldFields, maxBitfieldFields)
	}

	// THE INTERVAL ARITHMETIC. Every bit belongs to exactly one field, and the
	// exact offending positions are named — "bit 9 is in no field" is fixable,
	// "your fields do not tile" is not.
	covers := make([]int, len(bf.Bits))
	for i, f := range bf.Fields {
		if strings.TrimSpace(f.Label) == "" {
			return fmt.Errorf("field %d has no label. A bracket under the row with no name is a division the clip refuses to explain", i)
		}
		if strings.TrimSpace(f.Means) == "" {
			return fmt.Errorf("field %d (%q) says nothing about what its bits decode to. The lift IS the meaning: cells brightening with no sentence beside them is a colour change",
				i, f.Label)
		}
		if f.From > f.To {
			return fmt.Errorf("field %d (%q) runs from bit %d to bit %d, which is backwards. Bit 0 is the MOST significant bit, on the left, so \"from\" is the left edge and has to be the smaller index",
				i, f.Label, f.From, f.To)
		}
		if f.From < 0 || f.To >= len(bf.Bits) {
			return fmt.Errorf("field %d (%q) covers bits %d-%d, but the row only has bits 0-%d. Bit 0 is the leftmost, most significant cell",
				i, f.Label, f.From, f.To, len(bf.Bits)-1)
		}
		for b := f.From; b <= f.To; b++ {
			covers[b]++
		}
	}
	var gaps, overlaps []string
	for b, n := range covers {
		switch {
		case n == 0:
			gaps = append(gaps, strconv.Itoa(b))
		case n > 1:
			overlaps = append(overlaps, strconv.Itoa(b))
		}
	}
	if len(gaps) > 0 {
		return fmt.Errorf("bit(s) %s belong to no field. The fields have to TILE the row — every one of the %d bits inside exactly one bracket — because a cell with nothing bracketing it quietly teaches that some bits are spare, and no real layout has spare bits. Extend a neighbour or add a field",
			strings.Join(gaps, ", "), len(bf.Bits))
	}
	if len(overlaps) > 0 {
		return fmt.Errorf("bit(s) %s belong to more than one field. Two brackets over the same cell means the picture claims one bit is doing two jobs, and the renderer would draw one underline on top of the other. Move a boundary so the ranges meet without touching",
			strings.Join(overlaps, ", "))
	}

	// The shape: the wall, then the structure, then the meanings, then the read.
	if p.Beats[0].Bitfield == nil || p.Beats[0].Bitfield.ResolvedShow() != "row" {
		return fmt.Errorf("beat %q does not open on the undivided row. The whole arc of this clip is wall to structure to meaning, and it needs the wall first — the first beat shows the bits with no brackets on them",
			p.Beats[0].ID)
	}
	last := p.Beats[len(p.Beats)-1]
	if last.Bitfield == nil || last.Bitfield.ResolvedShow() != "read" {
		return fmt.Errorf("beat %q does not close on the read-back. The payoff is the whole pattern said as one sentence — end with {\"show\": \"read\"} or the clip stops at a set of labelled parts that never became a value",
			last.ID)
	}

	split := -1
	lifted := map[int]bool{}
	for i, b := range p.Beats {
		if b.Bitfield == nil {
			return fmt.Errorf("beat %q has no bitfield direction — every beat shows the plain row, drops the brackets in, lifts one field, or reads the whole thing back", b.ID)
		}
		switch b.Bitfield.ResolvedShow() {
		case "row":
			if i != 0 {
				return fmt.Errorf("beat %q goes back to the undivided row part-way through. Once the brackets are down they stay down; taking them off un-teaches the split the viewer just watched", b.ID)
			}
		case "read":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q reads the whole pattern back before every field has been explained. The read is the closer — it is the sentence the field beats earn", b.ID)
			}
		case "split":
			if split >= 0 {
				return fmt.Errorf("beat %q drops the field boundaries in again. The split happens once — a second one has nothing to divide", b.ID)
			}
			split = i
		case "field":
			if split < 0 {
				return fmt.Errorf("beat %q lifts a field before the row has been split. A group of cells brightening inside an undivided wall is a colour change with no stated reason — put the {\"show\": \"split\"} beat first", b.ID)
			}
			if b.Bitfield.At < 0 || b.Bitfield.At >= len(bf.Fields) {
				return fmt.Errorf("beat %q lifts field %d, which does not exist — the row is cut into fields 0-%d", b.ID, b.Bitfield.At, len(bf.Fields)-1)
			}
			if lifted[b.Bitfield.At] {
				return fmt.Errorf("beat %q lifts field %d (%q) again. Each field gets exactly one beat — a second one repeats a meaning the viewer already has and spends a beat some other field needed",
					b.ID, b.Bitfield.At, bf.Fields[b.Bitfield.At].Label)
			}
			lifted[b.Bitfield.At] = true
		}
	}
	if split < 0 {
		return fmt.Errorf("no beat splits the row. The brackets dropping in is the moment a wall of bits becomes a structure, and it is the one beat this template cannot do without")
	}
	if len(lifted) != len(bf.Fields) {
		return fmt.Errorf("%d of the %d fields are never lifted. A field bracketed under the row but never explained is a mystery the clip planted on purpose — give each one a beat, or cut the row into fewer fields",
			len(bf.Fields)-len(lifted), len(bf.Fields))
	}
	return nil
}

// bitfieldScenes lays the clip out as ONE scene. Each field arrives with its
// own bits already sliced out and their unsigned value already computed, so the
// component never reads a bit range or does base conversion.
func bitfieldScenes(in SnippetSceneInput) ([]Scene, error) {
	bf := in.Plan.Bitfield
	if bf == nil {
		return nil, fmt.Errorf("the plan has no bit row")
	}

	cells := make([]map[string]any, len(bf.Bits))
	for i := 0; i < len(bf.Bits); i++ {
		cells[i] = map[string]any{
			"bit":   string(bf.Bits[i]),
			"index": i,
		}
	}

	fields := make([]map[string]any, len(bf.Fields))
	for i, f := range bf.Fields {
		slice := ""
		if f.From >= 0 && f.To < len(bf.Bits) && f.From <= f.To {
			slice = bf.Bits[f.From : f.To+1]
		}
		value := 0
		if v, err := strconv.ParseUint(slice, 2, 64); err == nil {
			value = int(v)
		}
		fields[i] = map[string]any{
			"label": f.Label,
			"from":  f.From,
			"to":    f.To,
			"means": f.Means,
			"bits":  slice,
			"value": value,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	// Which fields have been explained by each beat, accumulated in Go so the
	// renderer never replays the beat list to find out.
	seen := map[int]bool{}
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Bitfield == nil {
			return nil, fmt.Errorf("beat %q has no bitfield direction", beat.ID)
		}
		show := beat.Bitfield.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "field" {
			seen[beat.Bitfield.At] = true
			step["at"] = beat.Bitfield.At
		}
		done := make([]int, 0, len(seen))
		for at := range seen {
			done = append(done, at)
		}
		sort.Ints(done)
		step["done"] = done
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneBitfield,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"bits":   bf.Bits,
			"cells":  cells,
			"fields": fields,
			"steps":  steps,
		}),
	}}, nil
}
