package pipeline

// The multiply template: one unit's figure, the count, and what that comes to.
//
// The boundary with `costing` is that a bill has heterogeneous lines and this has
// exactly one, repeated. `costing` answers "what does all of this add up to" and
// its picture is a list with a running total. This answers a narrower and more
// common question — "that seems fine, how many do you need?" — and its picture is
// one unremarkable figure, a count that multiplies it, and a product nobody
// expected. The reference clips land this repeatedly: fourteen and a half
// kilowatts is a rounding error until it is eight of them.
//
// The template is narrow on purpose. It earns its place on one thing, and it is
// not a layout.
//
// **The arithmetic is checked.** unit × count must equal the total the clip
// states, and nothing else in the catalog validates a number against another
// number. That matters more here than anywhere else because multiplication in
// front of an audience is the whole content: a clip that says "fourteen and a
// half kilowatts, times eight, so a hundred kilowatts" has taught a viewer to
// distrust it, and it fails in the one way a renderer cannot catch and a reviewer
// skims past. Rounding is allowed — 14.5 × 8 stated as "116" is right and stated
// as "116.0" is the same number — so the check is a tolerance rather than an
// equality.
//
// Two smaller rules. The per-unit figure is stated before the count, because the
// move only works if the viewer has already accepted the small number as
// unremarkable. And the count has to be drawable as a row of glyphs, which is
// what makes the multiplication visible rather than merely asserted.

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:             "multiply",
		Category:         CatNumbers,
		Since:            SinceV4,
		Family:           FamilyReplica,
		Title:            "One of them, times how many",
		Description:      "A per-unit figure, the count that multiplies it, and the product. Reach for it when the single number sounds reasonable and the total does not.",
		Example:          "What eight GPU nodes actually draw before you cool them",
		PromptFile:       snippetMultiplyTemplateName,
		NeedsCode:        false,
		MinTargetSec:     20,
		DefaultTargetSec: 40,
		MaxBeats:         7,
		Owns:             beatFields{Multiply: true},
		OwnsPlan:         planFields{Multiply: true},
		Normalize:        normalizeMultiplyPlan,
		Validate:         validateMultiplyPlan,
		Scenes:           multiplyScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(MultiplyShows(), ", "),
				"MinCount":      minMultiplyCount,
				"MaxCount":      maxMultiplyCount,
				"MaxLabelWords": maxMultiplyLabelWords,
				"MaxNoteWords":  maxMultiplyNoteWords,
				"MaxUnitWords":  maxMultiplyUnitWords,
			}
		},
	})
}

const snippetMultiplyTemplateName = "snippet_multiply.tmpl"

const (
	// Below three, "times two" is not a revelation and `compare` states it
	// better. Past sixty-four the glyph row is a texture and the count has to be
	// read from the number alone — at which point this is `metric`.
	minMultiplyCount = 3
	maxMultiplyCount = 64

	maxMultiplyLabelWords = 5
	maxMultiplyNoteWords  = 16
	maxMultiplyUnitWords  = 3

	// The arithmetic tolerance, as a fraction of the product. Half a percent
	// covers a total rounded for speech — "116" for 116.0, "1.4" for 1.44 — and
	// nothing else. A clip that is out by more than that is out by a mistake.
	multiplyTolerance = 0.005
)

// multiplyShows is the closed vocabulary of what a beat does.
var multiplyShows = map[string]bool{
	// The per-unit figure alone. The first beat, always.
	"unit": true,
	// The count arrives and the glyph row fills.
	"count": true,
	// The product lands.
	"total": true,
	// The caveat: what the product is still not counting.
	"caveat": true,
}

// MultiplyShows returns the beat vocabulary sorted.
func MultiplyShows() []string {
	out := make([]string, 0, len(multiplyShows))
	for k := range multiplyShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// MultiplySpec is the multiplication. On the plan because it is one statement
// the whole clip builds, not a per-beat visual.
type MultiplySpec struct {
	// UnitValue is what ONE of them is, and Unit is what that is counted in.
	UnitValue float64 `json:"unitValue"`
	Unit      string  `json:"unit"`
	// UnitLabel names the thing — "one B200 node", "a single request".
	UnitLabel string `json:"unitLabel"`
	// UnitNote is why that figure sounds reasonable. Optional, and the clip is
	// weaker without it: the move depends on the viewer accepting it first.
	UnitNote string `json:"unitNote,omitempty"`
	// Count is how many of them there are, and CountLabel says why that many.
	Count      int    `json:"count"`
	CountLabel string `json:"countLabel"`
	// Total is the product as the clip states it. Checked against
	// UnitValue × Count — see the file header.
	Total float64 `json:"total"`
	// TotalLabel names the product — "before cooling", "a month".
	TotalLabel string `json:"totalLabel,omitempty"`
	// TotalNote is what the product means. Optional.
	TotalNote string `json:"totalNote,omitempty"`
	// Caveat is what the product still does not count — "and that is before
	// cooling". Deliberately NOT a number: adding a second figure would make
	// this a bill, and a bill is `costing`.
	Caveat string `json:"caveat,omitempty"`
	// Role is what the product is doing: a metricRoles name.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the product's role, defaulting to the limit — the point of
// a clip like this is almost always that the total is a problem.
func (s *MultiplySpec) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(s.Role))
	if metricRoles[r] {
		return r
	}
	return "limit"
}

// Product is the arithmetic the clip is doing.
func (s *MultiplySpec) Product() float64 { return s.UnitValue * float64(s.Count) }

// ArithmeticOK reports whether the stated total matches the product, within a
// tolerance for a figure rounded for speech.
func (s *MultiplySpec) ArithmeticOK() bool {
	want := s.Product()
	if want == 0 {
		return s.Total == 0
	}
	return math.Abs(s.Total-want)/math.Abs(want) <= multiplyTolerance
}

// MultiplyBeat is one move.
type MultiplyBeat struct {
	// Show is a multiplyShows name.
	Show string `json:"show"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to the unit —
// which is where every clip of this shape starts.
func (b MultiplyBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if multiplyShows[s] {
		return s
	}
	return "unit"
}

func normalizeMultiplyPlan(p *SnippetPlan) {
	m := p.Multiply
	if m == nil {
		return
	}
	m.Unit = clampWords(collapseSpaces(m.Unit), maxMultiplyUnitWords)
	m.UnitLabel = clampWords(collapseSpaces(m.UnitLabel), maxMultiplyLabelWords)
	m.UnitNote = clampWords(collapseSpaces(m.UnitNote), maxMultiplyNoteWords)
	m.CountLabel = clampWords(collapseSpaces(m.CountLabel), maxMultiplyLabelWords)
	m.TotalLabel = clampWords(collapseSpaces(m.TotalLabel), maxMultiplyLabelWords)
	m.TotalNote = clampWords(collapseSpaces(m.TotalNote), maxMultiplyNoteWords)
	m.Caveat = clampWords(collapseSpaces(m.Caveat), maxMultiplyNoteWords)
	m.Role = m.ResolvedRole()

	// A total the model got slightly wrong is corrected rather than rejected.
	// The arithmetic is not a claim about the subject — it is arithmetic, and Go
	// can do it. Rejecting would spend a correction round teaching a model its
	// eight times table; the validator below still catches a total that is out
	// by enough to be a different intention rather than a slip.
	if m.Count > 0 && m.UnitValue > 0 && !m.ArithmeticOK() {
		if want := m.Product(); want != 0 && math.Abs(m.Total-want)/math.Abs(want) <= 0.25 {
			m.Total = roundTo(want, 2)
		}
	}

	for i := range p.Beats {
		b := p.Beats[i].Multiply
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
	}
}

func validateMultiplyPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Multiply: true}); err != nil {
		return err
	}

	m := p.Multiply
	if m == nil {
		return fmt.Errorf("the plan has no multiplication — this template is one per-unit figure, a count, and the product")
	}
	if m.UnitValue <= 0 {
		return fmt.Errorf("the per-unit figure is %v; there has to be something to multiply", m.UnitValue)
	}
	if strings.TrimSpace(m.Unit) == "" {
		return fmt.Errorf("the figure has no unit — %v of what? Every number in this clip is in the same unit", m.UnitValue)
	}
	if strings.TrimSpace(m.UnitLabel) == "" {
		return fmt.Errorf("the unit has no label — say what ONE of them is. \"14.5 kW\" is a reading; \"one B200 node draws 14.5 kW\" is the start of an argument")
	}
	if m.Count < minMultiplyCount || m.Count > maxMultiplyCount {
		return fmt.Errorf("the count is %d, want %d-%d. Below three, \"times two\" is not a revelation and the compare template states it better; past sixty-four the row of glyphs is a texture and the count can only be read from the number, which is metric",
			m.Count, minMultiplyCount, maxMultiplyCount)
	}
	if strings.TrimSpace(m.CountLabel) == "" {
		return fmt.Errorf("the count has no label — say why there are %d of them. A multiplier with no reason behind it is a number the viewer has to take on trust", m.Count)
	}

	// The rule the template exists for.
	if !m.ArithmeticOK() {
		return fmt.Errorf("the clip states a total of %v but %v %s times %d is %v. This template's whole content is doing multiplication in front of an audience, so arithmetic that does not check out is the one mistake it cannot ship — fix the total, or the count, or the per-unit figure",
			m.Total, m.UnitValue, m.Unit, m.Count, roundTo(m.Product(), 2))
	}
	if r := strings.ToLower(strings.TrimSpace(m.Role)); r != "" && !metricRoles[r] {
		return fmt.Errorf("the product has role %q, which is not one of: %s", m.Role, strings.Join(MetricRoles(), ", "))
	}

	// The per-unit figure is stated before the count: the move only works if the
	// viewer has already accepted the small number as unremarkable.
	if p.Beats[0].Multiply == nil || p.Beats[0].Multiply.ResolvedShow() != "unit" {
		return fmt.Errorf("beat %q does not state the per-unit figure. The whole move is that one of them sounds reasonable, so the viewer has to accept the small number before the count arrives to ruin it",
			p.Beats[0].ID)
	}

	counts := map[string]int{}
	order := make([]string, 0, len(p.Beats))
	for _, b := range p.Beats {
		if b.Multiply == nil {
			return fmt.Errorf("beat %q has no multiply direction — every beat states the unit, the count, the total, or the caveat", b.ID)
		}
		show := b.Multiply.ResolvedShow()
		counts[show]++
		order = append(order, show)
	}
	for _, once := range []string{"unit", "count", "total"} {
		if counts[once] != 1 {
			return fmt.Errorf("there are %d beats showing %q, want exactly 1. The clip states the figure once, the count once, and the product once", counts[once], once)
		}
	}
	if counts["caveat"] > 1 {
		return fmt.Errorf("there are %d caveat beats, want at most 1", counts["caveat"])
	}
	if m.Caveat == "" && counts["caveat"] > 0 {
		return fmt.Errorf("a beat shows the caveat but the plan has no caveat text to show")
	}
	// The order is the argument: figure, then count, then product.
	iUnit, iCount, iTotal := indexOfShow(order, "unit"), indexOfShow(order, "count"), indexOfShow(order, "total")
	if !(iUnit < iCount && iCount < iTotal) {
		return fmt.Errorf("the beats run %v. The order is the argument — the figure, then how many, then what that comes to — and any other sequence gives the product away before the multiplication happens",
			order)
	}
	return nil
}

// indexOfShow is the position of the first beat with this action, or -1.
func indexOfShow(order []string, show string) int {
	for i, s := range order {
		if s == show {
			return i
		}
	}
	return -1
}

// multiplyScenes lays the clip out as ONE scene: the statement builds across the
// beats and nothing is ever replaced.
func multiplyScenes(in SnippetSceneInput) ([]Scene, error) {
	m := in.Plan.Multiply
	if m == nil {
		return nil, fmt.Errorf("the plan has no multiplication")
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Multiply == nil {
			return nil, fmt.Errorf("beat %q has no multiply direction", beat.ID)
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Multiply.ResolvedShow(),
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneMultiply,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":      in.Plan.Title,
			"unitValue":  m.UnitValue,
			"unit":       m.Unit,
			"unitLabel":  m.UnitLabel,
			"unitNote":   m.UnitNote,
			"count":      m.Count,
			"countLabel": m.CountLabel,
			"total":      m.Total,
			"totalLabel": m.TotalLabel,
			"totalNote":  m.TotalNote,
			"caveat":     m.Caveat,
			"role":       m.ResolvedRole(),
			"steps":      steps,
		}),
	}}, nil
}
