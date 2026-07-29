package pipeline

// The costing template: what it really adds up to.
//
// The reference clips keep landing on a number nobody expected — four thousand
// for the box, three hundred thousand for the cluster — and the number only
// lands because the clip built it in front of you. A total asserted at the end
// is a claim; a total that accumulates line by line is an argument the viewer
// has already checked by the time it finishes.
//
// `metric` can state a big number and `data` can chart several, but neither can
// draw *accumulation*, which is the one thing that makes a cost estimate
// persuasive rather than merely alarming.
//
// Two rules earn it its place.
//
// The arithmetic has to be right. The stated total must equal the sum of the
// lines, to a tolerance that allows for honest rounding and nothing else. This
// is checked because it is exactly what a language model gets wrong — the
// number it writes in the total field is generated from the *vibe* of the list
// rather than from the list, and a course that teaches a wrong sum is worse
// than one that teaches nothing.
//
// And at least one line has to be a cost the viewer would not have thought of.
// The whole reason to draw a bill of materials is that the sticker price is not
// the price; a costing made entirely of obvious line items has told the viewer
// what they already knew, expensively.

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "costing",
		Title:       "What it really costs",
		Description: "Line items stacking into a running total — including the ones nobody budgets for.",
		Example:     "What a self-hosted GPU box actually costs in year one",
		PromptFile:  snippetCostingTemplateName,
		NeedsCode:   false,
		// The setup, three lines and the total is five beats.
		MinTargetSec:     40,
		DefaultTargetSec: 60,
		MaxBeats:         9,
		Owns:             beatFields{Costing: true},
		OwnsPlan:         planFields{Costing: true},
		Normalize:        normalizeCostingPlan,
		Validate:         validateCostingPlan,
		Scenes:           costingScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":         strings.Join(CostingShows(), ", "),
				"MinLines":      minCostingLines,
				"MaxLines":      maxCostingLines,
				"MaxUnitWords":  maxCostingUnitWords,
				"MaxLabelWords": maxCostingLabelWords,
				"MaxNoteWords":  maxCostingNoteWords,
				"Tolerance":     costingTolerancePct,
			}
		},
	})
}

const snippetCostingTemplateName = "snippet_costing.tmpl"

const (
	// Two lines is a sum anybody does in their head. Seven rows down the stage
	// leaves each too short to carry its note.
	minCostingLines = 3
	maxCostingLines = 6

	maxCostingUnitWords  = 2
	maxCostingLabelWords = 5
	maxCostingNoteWords  = 14

	// How far the stated total may sit from the sum of the lines, as a
	// percentage. Rounding a bill to a readable figure is honest; being out by
	// a tenth is arithmetic nobody did.
	costingTolerancePct = 2.0
)

// costingShows is the closed vocabulary of what a beat does.
var costingShows = map[string]bool{
	// Name what is being priced, with the sheet empty.
	"setup": true,
	// Add one line: it lands and the running total moves.
	"line": true,
	// The total, alone, with the sheet behind it. The closing frame.
	"total": true,
}

// CostingShows returns the beat vocabulary sorted.
func CostingShows() []string {
	out := make([]string, 0, len(costingShows))
	for k := range costingShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// CostingSpec is the bill being built.
type CostingSpec struct {
	// Subject is what is being priced — "A self-hosted GPU box, year one".
	Subject string `json:"subject"`
	// Unit is the currency or measure every line is in — "$", "$/mo", "hours".
	Unit string `json:"unit"`
	// Lines are the items, in the order they are added.
	Lines []CostLine `json:"lines"`
	// Total is the figure the clip lands on. Validated against the sum.
	Total float64 `json:"total"`
	// Verdict is the one line that says what the total means.
	Verdict string `json:"verdict,omitempty"`
}

// CostLine is one item on the bill.
type CostLine struct {
	Label  string  `json:"label"`
	Amount float64 `json:"amount"`
	// Note is why this line is what it is.
	Note string `json:"note,omitempty"`
	// Hidden marks a cost the viewer would not have thought of. At least one
	// line must be hidden, and they are drawn in the limit colour — the whole
	// reason to draw a bill is that the sticker price is not the price.
	Hidden bool `json:"hidden,omitempty"`
}

// CostingBeat is one move.
type CostingBeat struct {
	Show string `json:"show"`
	At   int    `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to adding a
// line — the bulk of the clip.
func (b CostingBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if costingShows[s] {
		return s
	}
	return "line"
}

// Sum returns what the lines actually add up to.
func (c CostingSpec) Sum() float64 {
	var total float64
	for _, l := range c.Lines {
		total += l.Amount
	}
	return total
}

func normalizeCostingPlan(p *SnippetPlan) {
	c := p.Costing
	if c == nil {
		return
	}
	c.Subject = clampWords(collapseSpaces(c.Subject), maxCostingLabelWords+3)
	c.Unit = clampWords(collapseSpaces(c.Unit), maxCostingUnitWords)
	c.Verdict = clampWords(collapseSpaces(c.Verdict), maxCostingNoteWords)

	lines := make([]CostLine, 0, len(c.Lines))
	for _, l := range c.Lines {
		l.Label = clampWords(collapseSpaces(l.Label), maxCostingLabelWords)
		l.Note = clampWords(collapseSpaces(l.Note), maxCostingNoteWords)
		// A negative line is a sign error rather than a credit; this template
		// draws costs stacking up, and a bar that grows downward would need a
		// grammar it does not have.
		l.Amount = math.Abs(l.Amount)
		if l.Label != "" && l.Amount > 0 && len(lines) < maxCostingLines {
			lines = append(lines, l)
		}
	}
	c.Lines = lines

	// A total the model left blank, or one it got wrong by a rounding, is
	// repaired rather than rejected — the sum is knowable from the lines, so
	// failing the plan over it would burn a correction round on arithmetic the
	// pipeline can simply do.
	if sum := c.Sum(); c.Total <= 0 || nearlyEqual(c.Total, sum, costingTolerancePct) {
		c.Total = sum
	}

	for i := range p.Beats {
		b := p.Beats[i].Costing
		if b == nil {
			continue
		}
		if !costingShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			switch {
			case i == 0:
				b.Show = "setup"
			case i == len(p.Beats)-1:
				b.Show = "total"
			default:
				b.Show = "line"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "line" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(c.Lines); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

// nearlyEqual reports whether a is within tolerance percent of b.
func nearlyEqual(a, b, tolerancePct float64) bool {
	if b == 0 {
		return a == 0
	}
	return math.Abs(a-b)/math.Abs(b)*100 <= tolerancePct
}

func validateCostingPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Costing: true}); err != nil {
		return err
	}

	c := p.Costing
	if c == nil {
		return fmt.Errorf("the plan has no costing — this template is a bill built line by line")
	}
	if strings.TrimSpace(c.Subject) == "" {
		return fmt.Errorf("nothing is named as the thing being priced")
	}
	if strings.TrimSpace(c.Unit) == "" {
		return fmt.Errorf("the bill has no unit — say what every line is measured in")
	}
	if n := len(c.Lines); n < minCostingLines || n > maxCostingLines {
		return fmt.Errorf("there are %d lines, want %d-%d. Two is a sum anybody does in their head, and seven rows leaves each too short to carry its note",
			n, minCostingLines, maxCostingLines)
	}

	hidden := 0
	seen := map[string]bool{}
	for i, l := range c.Lines {
		if strings.TrimSpace(l.Label) == "" {
			return fmt.Errorf("line %d has no label", i)
		}
		if l.Amount <= 0 {
			return fmt.Errorf("line %q costs %v — every line is a real amount", l.Label, l.Amount)
		}
		key := strings.ToLower(strings.TrimSpace(l.Label))
		if seen[key] {
			return fmt.Errorf("two lines are both %q — each one is a different cost", l.Label)
		}
		seen[key] = true
		if l.Hidden {
			hidden++
		}
	}

	// The first rule: the arithmetic has to be right.
	if sum := c.Sum(); !nearlyEqual(c.Total, sum, costingTolerancePct) {
		return fmt.Errorf("the total says %.2f but the lines add up to %.2f. A course that teaches a wrong sum is worse than one that teaches nothing — either fix the total or fix the line that is wrong",
			c.Total, sum)
	}
	// The second: at least one line the viewer would not have thought of.
	if hidden == 0 {
		return fmt.Errorf("no line is marked hidden. The whole reason to build a bill in front of somebody is that the sticker price is not the price — mark the cost they would not have budgeted for, or this is a list of things they already knew")
	}

	added := map[int]bool{}
	counts := map[string]int{}
	for i, b := range p.Beats {
		if b.Costing == nil {
			return fmt.Errorf("beat %q has no costing direction — every beat names the subject, adds a line, or lands the total", b.ID)
		}
		show := b.Costing.ResolvedShow()
		counts[show]++
		if i == 0 && show != "setup" {
			return fmt.Errorf("the clip opens on %q. Say what is being priced first — a line item landing before the viewer knows what the bill is for prices nothing", show)
		}
		if show == "total" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q lands the total but the clip carries on afterwards. The total is the closing frame", b.ID)
		}
		if show != "line" {
			continue
		}
		if b.Costing.At < 0 || b.Costing.At >= len(c.Lines) {
			return fmt.Errorf("beat %q adds line %d, which does not exist", b.ID, b.Costing.At)
		}
		if added[b.Costing.At] {
			return fmt.Errorf("beat %q adds line %d again; each cost lands once", b.ID, b.Costing.At)
		}
		added[b.Costing.At] = true
	}
	if counts["setup"] != 1 {
		return fmt.Errorf("there are %d setup beats; the subject is named once", counts["setup"])
	}
	if counts["total"] != 1 {
		return fmt.Errorf("there are %d total beats; the bill lands exactly once", counts["total"])
	}
	if len(added) != len(c.Lines) {
		return fmt.Errorf("%d of the %d lines are never spoken. A cost that appears on the sheet without narration is a number the viewer cannot check, and an unchecked total is one they will not believe",
			len(c.Lines)-len(added), len(c.Lines))
	}
	return nil
}

// costingScenes lays the clip out as ONE scene: the sheet accumulates and the
// running total moves with it.
func costingScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.Costing
	if c == nil {
		return nil, fmt.Errorf("the plan has no costing")
	}

	// The running total after each line, computed here so the renderer never
	// re-adds the bill — the number on screen and the number the validator
	// checked are then the same number by construction.
	lines := make([]map[string]any, len(c.Lines))
	var running float64
	widest := 0.0
	for _, l := range c.Lines {
		if l.Amount > widest {
			widest = l.Amount
		}
	}
	for i, l := range c.Lines {
		running += l.Amount
		lines[i] = map[string]any{
			"label":   l.Label,
			"amount":  l.Amount,
			"note":    l.Note,
			"hidden":  l.Hidden,
			"running": running,
			// Each line's bar is drawn against the biggest single line, not
			// against the total: against the total the small lines vanish, and
			// the small ones are usually the surprising ones.
			"frac": l.Amount / widest,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Costing == nil {
			return nil, fmt.Errorf("beat %q has no costing direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Costing.ResolvedShow(),
		}
		if beat.Costing.ResolvedShow() == "line" {
			step["at"] = beat.Costing.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneCosting,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":   in.Plan.Title,
			"subject": c.Subject,
			"unit":    c.Unit,
			"lines":   lines,
			"total":   c.Total,
			"verdict": c.Verdict,
			"steps":   steps,
		},
	}}, nil
}
