package pipeline

// The occupancy template: a fixed set of things, and how much of it is spoken for.
//
// The catalog can draw a quantity against a threshold (`gauge`), a quantity on
// its own (`metric`), and a quantity against other quantities (`data`). None of
// them can draw the thing the reference clips keep reaching for: a *population*
// of identical units, all of it visible at once, with some part of it claimed.
//
// The difference is not decorative. "16 of 896 experts are active" as a bar is
// a bar at 1.8%, which is a hairline and says nothing. As a grid of 896 cells
// with 16 lit, it is instantly obvious that the number is tiny AND that the rest
// of the population still had to be paid for — two claims at once, from one
// picture, which is the whole argument that clip was making. Same shape for
// memory slots with some marked used, for seats on a plan, for cores on a die.
//
// Three rules earn it its place, and all three are validators.
//
// The population is established before anything claims it. A grid that arrives
// already lit has no "before", and the whole effect is the contrast between the
// full set and the claimed part. So the first beat draws the empty grid and says
// how many cells it is, exactly as `gauge` establishes its ceiling first.
//
// The bands cannot exceed the total. A grid whose claims overfill it is a
// picture that contradicts its own caption, and it fails silently — the renderer
// would clamp and the frame would look fine while saying something untrue.
//
// The total has to be drawable. Below a dozen cells a grid is a handful of boxes
// and `rundown` draws that better; past about twelve hundred the cells are
// smaller than the gaps between them and the picture is a texture. Refusing is
// better than rendering a grey rectangle and calling it a diagram.

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "occupancy",
		Category:    CatNumbers,
		Since:       SinceV4,
		Family:      FamilyReplica,
		Title:       "How much is spoken for",
		Description: "A grid of identical units with part of it claimed. Reach for it when the point is what fraction of a fixed population is in play — active experts, used slots, taken seats.",
		Example:     "Why a mixture-of-experts model only runs 16 of its 896 experts per token",
		PromptFile:  snippetOccupancyTemplateName,
		NeedsCode:   false,
		// The grid, then one beat per band, then what it means. Three bands is
		// already five beats, and a band that is not paused on is a band nobody
		// read.
		MinTargetSec:     30,
		DefaultTargetSec: 50,
		MaxBeats:         8,
		Owns:             beatFields{Occupancy: true},
		OwnsPlan:         planFields{Occupancy: true},
		Normalize:        normalizeOccupancyPlan,
		Validate:         validateOccupancyPlan,
		Scenes:           occupancyScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(OccupancyShows(), ", "),
				"MinTotal":      minOccupancyTotal,
				"MaxTotal":      maxOccupancyTotal,
				"MinBands":      minOccupancyBands,
				"MaxBands":      maxOccupancyBands,
				"MaxUnitWords":  maxOccupancyUnitWords,
				"MaxLabelWords": maxOccupancyLabelWords,
				"MaxNoteWords":  maxOccupancyNoteWords,
			}
		},
	})
}

const snippetOccupancyTemplateName = "snippet_occupancy.tmpl"

const (
	// Below a dozen this is a row of boxes, which `rundown` draws better and
	// with a contract attached. Past twelve hundred the cells are smaller than
	// the gaps and the grid reads as a texture rather than as a count.
	minOccupancyTotal = 12
	maxOccupancyTotal = 1200

	// One band is a fraction and needs no grid; four bands is a stacked bar
	// chart wearing a grid's clothes, and `data` draws that honestly.
	minOccupancyBands = 1
	maxOccupancyBands = 3

	maxOccupancyUnitWords  = 3
	maxOccupancyLabelWords = 5
	maxOccupancyNoteWords  = 16
)

// occupancyShows is the closed vocabulary of what a beat does.
var occupancyShows = map[string]bool{
	// Draw the empty population and say how big it is. The first beat, always.
	"grid": true,
	// Light one band's cells.
	"fill": true,
	// Hold the finished grid and say what the proportion means.
	"read": true,
}

// OccupancyShows returns the beat vocabulary sorted.
func OccupancyShows() []string {
	out := make([]string, 0, len(occupancyShows))
	for k := range occupancyShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// OccupancySpec is the population and the claims on it. On the plan rather than
// per-beat because the grid is one object that persists across every beat — the
// beats only change which cells are lit.
type OccupancySpec struct {
	// Total is how many units there are altogether.
	Total int `json:"total"`
	// Unit is what one cell is — "expert", "memory slot", "seat", "core".
	// Singular: the renderer sets it beside a count and pluralises nothing.
	Unit string `json:"unit"`
	// Label names the population — "the model's experts", "the key space".
	Label string `json:"label"`
	// Bands are the claims on the population, in the order they are made.
	Bands []OccupancyBand `json:"bands"`
}

// OccupancyBand is one claim on the population.
type OccupancyBand struct {
	// Count is how many units this band claims.
	Count int `json:"count"`
	// Label names the claim — "active this token", "already allocated".
	Label string `json:"label"`
	// Note is the one line that turns a fraction into an argument.
	Note string `json:"note,omitempty"`
	// Role is what this band is doing: a metricRoles name. It picks the
	// semantic accent, so it is a claim about meaning rather than about colour.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the band's role, defaulting to neutral — a claim whose
// job the model did not state should not be shouting in red.
func (b OccupancyBand) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(b.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// Claimed is how many units all the bands take together.
func (s *OccupancySpec) Claimed() int {
	n := 0
	for _, b := range s.Bands {
		n += b.Count
	}
	return n
}

// OccupancyBeat is one move: which band this beat lights.
type OccupancyBeat struct {
	// Show is an occupancyShows name.
	Show string `json:"show"`
	// At indexes OccupancySpec.Bands, for a "fill" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to lighting a
// band — which is what most beats of this template do.
func (b OccupancyBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if occupancyShows[s] {
		return s
	}
	return "fill"
}

func normalizeOccupancyPlan(p *SnippetPlan) {
	o := p.Occupancy
	if o == nil {
		return
	}
	o.Unit = clampWords(collapseSpaces(o.Unit), maxOccupancyUnitWords)
	o.Label = clampWords(collapseSpaces(o.Label), maxOccupancyLabelWords)

	bands := make([]OccupancyBand, 0, len(o.Bands))
	for _, b := range o.Bands {
		b.Label = clampWords(collapseSpaces(b.Label), maxOccupancyLabelWords)
		b.Note = clampWords(collapseSpaces(b.Note), maxOccupancyNoteWords)
		b.Role = b.ResolvedRole()
		// A band claiming nothing lights no cells, so it is not a band. Dropping
		// it is the repair; inventing a count would be a claim about the subject.
		if b.Count > 0 && b.Label != "" && len(bands) < maxOccupancyBands {
			bands = append(bands, b)
		}
	}
	o.Bands = bands

	for i := range p.Beats {
		b := p.Beats[i].Occupancy
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "fill" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(o.Bands); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateOccupancyPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Occupancy: true}); err != nil {
		return err
	}

	o := p.Occupancy
	if o == nil {
		return fmt.Errorf("the plan has no population — this template is one grid of identical units with part of it claimed")
	}
	if o.Total < minOccupancyTotal || o.Total > maxOccupancyTotal {
		return fmt.Errorf("the population is %d, want %d-%d. Below a dozen this is a row of boxes and the rundown template draws it better; past twelve hundred the cells are smaller than the gaps between them and the grid reads as a texture rather than as a count",
			o.Total, minOccupancyTotal, maxOccupancyTotal)
	}
	if strings.TrimSpace(o.Unit) == "" {
		return fmt.Errorf("the population has no unit — say what one cell is. A grid of %d unnamed squares is a pattern, not a measurement", o.Total)
	}
	if n := len(o.Bands); n < minOccupancyBands || n > maxOccupancyBands {
		return fmt.Errorf("there are %d bands, want %d-%d. Four claims on one population is a stacked bar chart wearing a grid's clothes, and the data template draws that honestly",
			n, minOccupancyBands, maxOccupancyBands)
	}

	// The rule that keeps the picture honest. A grid whose claims overfill it
	// contradicts its own caption, and it does so silently — the renderer would
	// clamp and the frame would still look fine.
	if claimed := o.Claimed(); claimed > o.Total {
		return fmt.Errorf("the bands claim %d %ss between them but the population is only %d. The grid cannot draw more cells than it has, so this is a picture that would contradict its own caption",
			claimed, strings.TrimSpace(o.Unit), o.Total)
	}

	seen := map[string]bool{}
	for i, b := range o.Bands {
		if b.Count <= 0 {
			return fmt.Errorf("band %d (%q) claims %d units — a band that lights no cells is not a band", i, b.Label, b.Count)
		}
		if strings.TrimSpace(b.Label) == "" {
			return fmt.Errorf("band %d has no label — say what those %d cells are", i, b.Count)
		}
		key := strings.ToLower(strings.TrimSpace(b.Label))
		if seen[key] {
			return fmt.Errorf("two bands are both %q — each claim on the population gets its own name", b.Label)
		}
		seen[key] = true
		if r := strings.ToLower(strings.TrimSpace(b.Role)); r != "" && !metricRoles[r] {
			return fmt.Errorf("band %d has role %q, which is not one of: %s", i, b.Role, strings.Join(MetricRoles(), ", "))
		}
	}

	// The population is established before anything claims it — the contrast
	// between the full set and the claimed part is the entire effect, and a grid
	// that arrives already lit has no "before" to contrast with.
	if p.Beats[0].Occupancy == nil || p.Beats[0].Occupancy.ResolvedShow() != "grid" {
		return fmt.Errorf("beat %q does not draw the empty grid. The first beat establishes the population and says how big it is: the whole picture is the contrast between the full set and the part claimed, and a grid that arrives already lit has no before",
			p.Beats[0].ID)
	}

	filled := map[int]bool{}
	grids := 0
	for i, b := range p.Beats {
		if b.Occupancy == nil {
			return fmt.Errorf("beat %q has no occupancy direction — every beat draws the grid, lights a band, or reads the result", b.ID)
		}
		switch b.Occupancy.ResolvedShow() {
		case "grid":
			grids++
			if i != 0 {
				return fmt.Errorf("beat %q draws the empty grid again part-way through. The population is established once, at the start", b.ID)
			}
		case "fill":
			if b.Occupancy.At < 0 || b.Occupancy.At >= len(o.Bands) {
				return fmt.Errorf("beat %q lights band %d, which does not exist", b.ID, b.Occupancy.At)
			}
			if filled[b.Occupancy.At] {
				return fmt.Errorf("beat %q lights band %d again; each claim gets one beat", b.ID, b.Occupancy.At)
			}
			filled[b.Occupancy.At] = true
		}
	}
	if grids != 1 {
		return fmt.Errorf("there are %d beats drawing the empty grid, want exactly 1", grids)
	}
	if len(filled) != len(o.Bands) {
		return fmt.Errorf("%d of the %d bands are never lit. A claim the narrator skips is a claim nobody read — give it a beat or cut it",
			len(o.Bands)-len(filled), len(o.Bands))
	}
	// At least one band has to be doing something in the argument, for the same
	// reason a metric clip cannot be all-neutral: a picture with no point of
	// view is a list of facts.
	allNeutral := true
	for _, b := range o.Bands {
		if b.ResolvedRole() != "neutral" {
			allNeutral = false
		}
	}
	if allNeutral {
		return fmt.Errorf("every band is role %q, so nothing in the clip is being argued. Mark the claim the subject is measured by as %q, and the one that runs it out of room as %q — the colours are how a viewer reads the argument before the narrator finishes the sentence",
			"neutral", "quantity", "limit")
	}
	return nil
}

// occupancyGridShape picks the grid's column count.
//
// Near-square, biased wide, because the stage is 16:9 and a square grid of 896
// cells leaves two thirds of the frame empty. Computed in Go rather than in CSS
// so the count is a fact the scene graph records — a grid that reflowed at a
// different viewport would not match its own baseline.
func occupancyGridShape(total int) (cols, rows int) {
	if total <= 0 {
		return 1, 1
	}
	// 16:9 of cells rather than of pixels: cols/rows ≈ 16/9 gives
	// cols ≈ sqrt(total * 16/9).
	cols = int(math.Round(math.Sqrt(float64(total) * 16.0 / 9.0)))
	if cols < 1 {
		cols = 1
	}
	if cols > total {
		cols = total
	}
	rows = int(math.Ceil(float64(total) / float64(cols)))
	return cols, rows
}

// occupancyScenes lays the clip out as ONE scene. The grid persists for the
// whole clip and the beats only change which cells are lit, so the layout is one
// component's business.
func occupancyScenes(in SnippetSceneInput) ([]Scene, error) {
	o := in.Plan.Occupancy
	if o == nil {
		return nil, fmt.Errorf("the plan has no population")
	}

	// Cells are assigned to bands in declaration order, so a band's cells are a
	// contiguous run. Contiguous rather than scattered because the eye reads a
	// block as a quantity and a sprinkle as noise — and the one clip that wanted
	// scattered (memory fragmentation) is a different template's job.
	bands := make([]map[string]any, len(o.Bands))
	at := 0
	for i, b := range o.Bands {
		bands[i] = map[string]any{
			"count": b.Count,
			"from":  at,
			"label": b.Label,
			"note":  b.Note,
			"role":  b.ResolvedRole(),
		}
		at += b.Count
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Occupancy == nil {
			return nil, fmt.Errorf("beat %q has no occupancy direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Occupancy.ResolvedShow(),
		}
		if beat.Occupancy.ResolvedShow() == "fill" {
			step["at"] = beat.Occupancy.At
		}
		steps = append(steps, step)
	}

	cols, rows := occupancyGridShape(o.Total)
	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneOccupancy,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title": in.Plan.Title,
			"total": o.Total,
			"unit":  o.Unit,
			"label": o.Label,
			"cols":  cols,
			"rows":  rows,
			"bands": bands,
			"steps": steps,
		}),
	}}, nil
}
