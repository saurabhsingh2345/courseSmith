package pipeline

// The carry template: addition, worked the way the machine works it.
//
// A career-switcher meets binary addition as a rule to memorise — "one plus one
// is zero carry one" — and memorises it as a separate arithmetic, a second
// system that happens to look like the one they learned at seven. It is not.
// It is the SAME procedure: line the digits up, add a column, keep what fits in
// one digit, hand the overflow to the column on the left. Decimal has ten
// symbols and binary has two, and that is the entire difference. So this
// template does not teach a rule. It works one addition, column by column, in
// exactly the layout a child does long addition in, and lets the viewer notice
// that they already know how.
//
// Which is why a beat is a COLUMN. The clip is a walk right to left, one
// column at a time, and the carry chip hopping to the next column is the only
// thing that moves — because the carry is the whole idea and everything else is
// bookkeeping. The decimal equivalents sit in muted type at the ends of the
// rows: not the subject, but the proof that the two arithmetics agree.
//
// The validators exist because this picture is an equation and a wrong equation
// is worse than no picture at all.
//
// **A + B must equal Sum, checked in Go.** A model writing binary from memory
// drops a carry fluently and confidently, and a column addition whose bottom
// row is wrong teaches the viewer that the diagram is decoration. Both operands
// and the claimed sum are parsed as base 2 and added, and a mismatch is
// rejected with the true sum quoted in binary AND in decimal.
//
// **At least one column must carry.** An addition with no carries — 1010 plus
// 0101 — is four independent one-bit sums, and the chip never hops. That is a
// clip about the template's own subject that never shows its subject.
//
// **The columns are walked right to left with no skips.** Long addition has
// exactly one legal direction, because a column cannot be worked until the
// carry coming into it is known. A clip that jumps to the fourth column has
// shown a digit it could not have computed yet.

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "carry",
		Category:    CatNumbers,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "Column arithmetic",
		Description: "One binary addition worked column by column, with the carry chip hopping left and the decimal check underneath. Reach for it when the point is that binary arithmetic IS decimal arithmetic with two symbols — adding, carrying, overflow.",
		Example:     "1011 + 0110 in binary, one column at a time",
		PromptFile:  snippetCarryTemplateName,
		NeedsCode:   false,
		// The problem, a column a beat, and the answer. Under thirty-five seconds
		// the word budget cannot fund an opener, three columns and a closer at a
		// spoken pace, and a column that flashes past is a column nobody worked.
		MinTargetSec: 35,
		// Longer than the family's usual default because the beat count here is
		// set by the arithmetic rather than by taste: a five-column addition needs
		// seven beats whatever the runtime, and 55 seconds is what funds them.
		DefaultTargetSec: 55,
		// Opener + up to nine columns + an optional carry-chain beat + closer. Nine
		// is the widest problem this template accepts (eight-bit operands plus the
		// carry-out column), so the ceiling is exactly what the widest legal
		// addition needs and not one beat more.
		MaxBeats: 12,
		// A beat is a SHOT — one column lighting up — not a step in an argument.
		// Twenty-eight words is about nine seconds, which is as long as a single
		// column holds anybody's attention.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Carry: true},
		OwnsPlan:          planFields{Carry: true},
		Normalize:         normalizeCarryPlan,
		Validate:          validateCarryPlan,
		Scenes:            carryScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":      strings.Join(MetricRoles(), ", "),
				"Shows":      strings.Join(CarryShows(), ", "),
				"MaxBits":    maxCarryBits,
				"MinColumns": minCarryColumns,
				"MaxColumns": maxCarryColumns,
			}
		},
	})
}

const snippetCarryTemplateName = "snippet_carry.tmpl"

const (
	// Eight bits is a byte, which is the unit the rest of the course is about,
	// and it is also the widest operand whose digits stay legible in a grid that
	// has to carry a carry chip above every column.
	maxCarryBits = 8
	// Three columns is the shortest walk that is still a walk: two columns and a
	// carry is a single example, not a procedure.
	minCarryColumns = 3
	// Eight-bit operands can sum to nine bits, so nine is the arithmetic answer
	// rather than a taste judgement — and nine columns plus an opener and a
	// closer is exactly the twelve-beat ceiling above.
	maxCarryColumns = maxCarryBits + 1
)

// carryShows is the closed vocabulary of what a beat does.
var carryShows = map[string]bool{
	// A over B with the underline, nothing worked yet. The opener.
	"problem": true,
	// Column At is worked: its two bits, the carry coming in, the result digit
	// landing, and the carry chip hopping to the column on the left.
	"column": true,
	// The whole carry path lights at once. Optional, and only ever a review of
	// columns already worked.
	"carrychain": true,
	// The sum row complete, with its decimal check beneath. The closer.
	"answer": true,
}

// CarryShows returns the beat vocabulary sorted.
func CarryShows() []string {
	out := make([]string, 0, len(carryShows))
	for k := range carryShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// CarrySpec is the one addition the clip works. On the plan because the problem
// persists for the whole clip — the beats only move down it.
type CarrySpec struct {
	// A is the first operand in binary — "1011". Checked, not trusted.
	A string `json:"a"`
	// B is the second operand in binary — "0110". Checked, not trusted.
	B string `json:"b"`
	// Sum is the claimed result in binary — "10001". Recomputed in Go.
	Sum string `json:"sum"`
}

// CarryBeat is one shot: which column this beat works.
type CarryBeat struct {
	// Show is a carryShows name.
	Show string `json:"show"`
	// At is the column, counting from the RIGHT starting at zero, for a
	// "column" beat. Right-handed because that is the order addition happens in.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a column —
// the workhorse state nearly every beat of this template is in.
func (b CarryBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if carryShows[s] {
		return s
	}
	return "column"
}

// carryColumn is one worked column, computed in Go so the renderer never adds
// anything.
type carryColumn struct {
	a, b     string
	carryIn  int
	digit    string
	carryOut int
}

// carryBitAt reads the i-th bit from the RIGHT of s, treating anything past the
// left-hand end as a leading zero, which is what makes ragged operands line up.
func carryBitAt(s string, i int) int {
	if i >= len(s) {
		return 0
	}
	if s[len(s)-1-i] == '1' {
		return 1
	}
	return 0
}

// carryColumnCount is the width of the worked grid: the widest of the three
// rows, so an addition that overflows into a new digit still has a column for
// that digit to land in.
func carryColumnCount(a, b, sum string) int {
	w := len(a)
	if len(b) > w {
		w = len(b)
	}
	if len(sum) > w {
		w = len(sum)
	}
	return w
}

// carryColumns works the addition right to left and returns every column's
// state. This is the arithmetic the whole template is made of, and it lives
// here — in Go, once — so the validator and the renderer cannot disagree.
func carryColumns(a, b, sum string) []carryColumn {
	width := carryColumnCount(a, b, sum)
	cols := make([]carryColumn, width)
	carry := 0
	for i := 0; i < width; i++ {
		ab := carryBitAt(a, i)
		bb := carryBitAt(b, i)
		total := ab + bb + carry
		cols[i] = carryColumn{
			a:        strconv.Itoa(ab),
			b:        strconv.Itoa(bb),
			carryIn:  carry,
			digit:    strconv.Itoa(total % 2),
			carryOut: total / 2,
		}
		carry = total / 2
	}
	return cols
}

// cleanCarryBits strips the decoration a model puts on a binary literal — a 0b
// prefix, underscores, spaces — and drops leading zeros, which would otherwise
// become dead columns the walk still has to visit.
func cleanCarryBits(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.TrimPrefix(s, "0B")
	for len(s) > 1 && s[0] == '0' {
		s = s[1:]
	}
	return s
}

func normalizeCarryPlan(p *SnippetPlan) {
	c := p.Carry
	if c == nil {
		return
	}
	c.A = cleanCarryBits(c.A)
	c.B = cleanCarryBits(c.B)
	c.Sum = cleanCarryBits(c.Sum)

	width := carryColumnCount(c.A, c.B, c.Sum)
	for i := range p.Beats {
		b := p.Beats[i].Carry
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "column" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if width > 0 && b.At >= width {
			b.At = width - 1
		}
	}
}

func validateCarryPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Carry: true}); err != nil {
		return err
	}

	c := p.Carry
	if c == nil {
		return fmt.Errorf("the plan has no addition — this template is one sum worked column by column, so the problem IS the clip")
	}
	if strings.TrimSpace(c.A) == "" || strings.TrimSpace(c.B) == "" {
		return fmt.Errorf("the addition is missing an operand. Column arithmetic needs two numbers stacked over the line, so both \"a\" and \"b\" have to be there")
	}
	if len(c.A) > maxCarryBits {
		return fmt.Errorf("the first operand %s is %d bits; this template takes at most %d. A byte is the unit the rest of the course is about, and past %d bits the digit grid stops being legible above a row of carry chips",
			c.A, len(c.A), maxCarryBits, maxCarryBits)
	}
	if len(c.B) > maxCarryBits {
		return fmt.Errorf("the second operand %s is %d bits; this template takes at most %d. Drop to numbers that fit in a byte",
			c.B, len(c.B), maxCarryBits)
	}
	av, err := strconv.ParseUint(c.A, 2, 64)
	if err != nil {
		return fmt.Errorf("the first operand %q does not parse as base 2 — every digit has to be 0 or 1, because the picture puts one digit in one cell of a binary column", c.A)
	}
	bv, err := strconv.ParseUint(c.B, 2, 64)
	if err != nil {
		return fmt.Errorf("the second operand %q does not parse as base 2 — every digit has to be 0 or 1", c.B)
	}
	if strings.TrimSpace(c.Sum) == "" {
		return fmt.Errorf("the addition has no sum. The bottom row is where every column's result digit lands, so without it the walk has nowhere to write")
	}
	sv, err := strconv.ParseUint(c.Sum, 2, 64)
	if err != nil {
		return fmt.Errorf("the sum %q does not parse as base 2 — every digit has to be 0 or 1", c.Sum)
	}

	// THE ARITHMETIC. This picture is an equation, and a column addition whose
	// bottom row is wrong teaches the viewer that the diagram is decoration.
	if av+bv != sv {
		want := av + bv
		return fmt.Errorf("the plan adds %s and %s and claims %s, but the real sum is %s — that is %d in decimal, and the plan's %s is %d. Every result digit and every carry chip is recomputed from your operands, so a wrong bottom row would visibly disagree with the columns above it",
			c.A, c.B, c.Sum, strconv.FormatUint(want, 2), want, c.Sum, sv)
	}

	cols := carryColumns(c.A, c.B, c.Sum)
	if len(cols) < minCarryColumns {
		return fmt.Errorf("the problem is only %d column(s) wide, and this template needs at least %d. Two columns and a carry is a single example rather than a procedure — pick larger numbers",
			len(cols), minCarryColumns)
	}
	if len(cols) > maxCarryColumns {
		return fmt.Errorf("the problem is %d columns wide, and the grid holds at most %d. Each column gets its own beat, so a wider addition needs more beats than any runtime funds",
			len(cols), maxCarryColumns)
	}
	carries := 0
	for _, col := range cols {
		if col.carryOut == 1 {
			carries++
		}
	}
	if carries == 0 {
		return fmt.Errorf("adding %s and %s never carries: every column sums to 0 or 1 on its own. The carry chip hopping left IS this template's subject, so a clip built on this problem never shows the thing it is about — pick operands that set the same bit at least once",
			c.A, c.B)
	}

	// The shape: open on the problem, walk right to left, close on the answer.
	if p.Beats[0].Carry == nil || p.Beats[0].Carry.ResolvedShow() != "problem" {
		return fmt.Errorf("beat %q does not open on the problem. A column lighting up in a grid the viewer has not seen laid out is a cell changing colour — the first beat puts A over B with the underline and works nothing",
			p.Beats[0].ID)
	}
	last := p.Beats[len(p.Beats)-1]
	if last.Carry == nil || last.Carry.ResolvedShow() != "answer" {
		return fmt.Errorf("beat %q does not close on the answer. The final frame is the completed sum row with its decimal check beneath — that is the frame somebody screenshots, and a clip that stops mid-column never gets there",
			last.ID)
	}

	next := 0
	for i, b := range p.Beats {
		if b.Carry == nil {
			return fmt.Errorf("beat %q has no carry direction — every beat shows the problem, works one column, reviews the carry path, or lands the answer", b.ID)
		}
		switch b.Carry.ResolvedShow() {
		case "problem":
			if i != 0 {
				return fmt.Errorf("beat %q goes back to the unworked problem part-way through. The problem is the opener; returning to it erases columns the viewer just watched being worked", b.ID)
			}
		case "answer":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q lands the answer before the columns are worked. The answer is the closer — after it there is nothing left to compute", b.ID)
			}
		case "carrychain":
			if next == 0 {
				return fmt.Errorf("beat %q highlights the carry path before a single column has been worked. There is no chain yet: the chain is made of carries the viewer has already watched hop", b.ID)
			}
		case "column":
			if next >= len(cols) {
				return fmt.Errorf("beat %q works another column, but %s plus %s is only %d columns wide — columns 0 to %d, counting from the right",
					b.ID, c.A, c.B, len(cols), len(cols)-1)
			}
			if b.Carry.At != next {
				return fmt.Errorf("beat %q works column %d when the walk is at column %d. Long addition has exactly one legal direction: right to left, no skips, no repeats, because a column cannot be worked until the carry coming into it is known",
					b.ID, b.Carry.At, next)
			}
			next++
		}
	}
	if next != len(cols) {
		return fmt.Errorf("the walk stops after %d of the %d columns. A column left unworked is a digit in the answer the clip never explains — give each column its own beat, counting 0, 1, 2 from the right",
			next, len(cols))
	}
	return nil
}

// carryScenes lays the clip out as ONE scene. Every digit, every carry and both
// decimal equivalents are computed here, so the component never adds anything:
// it only decides what is visible yet.
func carryScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.Carry
	if c == nil {
		return nil, fmt.Errorf("the plan has no addition")
	}

	cols := carryColumns(c.A, c.B, c.Sum)
	// Indexed by significance: entry 0 is the RIGHTMOST column. The renderer
	// draws the grid reversed, which is one line there and keeps the arithmetic
	// order honest here.
	columns := make([]map[string]any, len(cols))
	for i, col := range cols {
		columns[i] = map[string]any{
			"a":        col.a,
			"b":        col.b,
			"carryIn":  col.carryIn,
			"digit":    col.digit,
			"carryOut": col.carryOut,
		}
	}

	av, _ := strconv.ParseUint(c.A, 2, 64)
	bv, _ := strconv.ParseUint(c.B, 2, 64)
	sv, _ := strconv.ParseUint(c.Sum, 2, 64)

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	// Which columns have been worked by each beat, accumulated in Go so the
	// renderer never replays the beat list to find out.
	worked := map[int]bool{}
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Carry == nil {
			return nil, fmt.Errorf("beat %q has no carry direction", beat.ID)
		}
		show := beat.Carry.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "column" {
			worked[beat.Carry.At] = true
			step["at"] = beat.Carry.At
		}
		done := make([]int, 0, len(worked))
		for at := range worked {
			done = append(done, at)
		}
		sort.Ints(done)
		step["done"] = done
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneCarry,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":      in.Plan.Title,
			"a":          c.A,
			"b":          c.B,
			"sum":        c.Sum,
			"aDecimal":   int(av),
			"bDecimal":   int(bv),
			"sumDecimal": int(sv),
			"columns":    columns,
			"steps":      steps,
		}),
	}}, nil
}
