package pipeline

// The data template: one dataset, read several ways.
//
// The distinguishing decision is that the chart is declared once for the whole
// clip and the beats only change what is *emphasised* in it. The obvious
// alternative — a chart per beat — was rejected for the reason the flow diagram
// was built the way it was: a viewer who has to re-read a new set of axes every
// eight seconds never gets past reading them. Holding one chart on screen and
// walking the narration around it is what lets numbers actually land, and it is
// also the only version where a chart is worth animating at all.
//
// So Chart lives on the plan rather than on a beat. That is a real departure
// from every other template here, and it is deliberate: the dataset is a
// property of the clip, not of a moment in it.
//
// The kind vocabulary is wide (thirteen) because the shape of a dataset is not
// a style choice — parts of a whole, a value over a sequence, and two variables
// against each other are three different claims, and drawing any of them as
// bars makes the wrong one. What every kind here has in common is that it can
// be read from a sofa in two seconds and that a highlight means something on
// it; anything that fails either test is not in the list, however common it is
// in a dashboard.

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "data",
		Category:    CatNumbers,
		Title:       "Data & maps",
		Description: "Real numbers on one chart or world map, with the narration walking around it.",
		Example:     "Where the world's undersea internet cables actually land",
		PromptFile:  snippetDataTemplateName,
		NeedsCode:   false,
		Owns:        beatFields{Data: true},
		OwnsPlan:    planFields{Chart: true},
		Normalize:   normalizeDataPlan,
		Validate:    validateDataPlan,
		Scenes:      dataScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"ChartKinds":    strings.Join(ChartKindNames(), ", "),
				"MinPoints":     minDataPoints,
				"MaxPoints":     maxDataPoints,
				"MaxLabelWords": maxDataLabelWords,
			}
		},
	})
}

const snippetDataTemplateName = "snippet_data.tmpl"

// How much data one chart can carry and still be read from a sofa. The floor
// is where a chart stops being a chart: two bars is a comparison you should
// have said in a sentence.
const (
	minDataPoints = 3
	maxDataPoints = 10
)

// maxDataLabelWords keeps an axis label an axis label rather than a caption.
const maxDataLabelWords = 4

// Series bounds, for the kinds that carry more than one number per label. Past
// four segments a stacked bar is a colour-matching exercise, and its legend is
// longer than the chart is tall.
const (
	minChartSeries = 2
	maxChartSeries = 4
)

// chartKind is one entry in the vocabulary: what it is for, and what the data
// has to look like for it to be honest.
type chartKind struct {
	// series is how many numbers each point carries; 0 means a single value.
	// -1 means "2 to maxChartSeries".
	series int
	// maxPoints overrides maxDataPoints where the drawing runs out of room
	// sooner than the reader does.
	maxPoints int
	// shares marks the kinds that claim their values are parts of one whole,
	// which is a claim worth checking rather than a style.
	shares bool
	// descending marks the kinds whose meaning depends on the order — a funnel
	// that widens is not a funnel.
	descending bool
}

// chartKindVocab mirrors the switch in renderer/src/components/DataScene.tsx.
var chartKindVocab = map[string]chartKind{
	// Horizontal bars. The default, and right for almost any comparison —
	// horizontal because it is the only orientation where a real-world label
	// fits without being rotated.
	"bars": {maxPoints: maxDataPoints},
	// One bar per label, split into named parts. For "what is this made of",
	// where the total matters as much as the split.
	"stackedbars": {series: -1, maxPoints: 7},
	// The same data side by side instead of stacked. For "compare the parts to
	// each other", where the total does not matter.
	"groupedbars": {series: -1, maxPoints: 6},
	// A value over an ordered sequence: time, versions, sizes.
	"line": {maxPoints: maxDataPoints},
	// The same, filled to the axis — for a quantity that accumulates or a
	// volume rather than a level.
	"area": {maxPoints: maxDataPoints},
	// Two variables against each other, to show whether they move together.
	"scatter": {series: 2, maxPoints: maxDataPoints},
	// Parts of one whole. Only honest when the values are shares of something.
	"donut": {shares: true, maxPoints: 7},
	// The same claim as a donut, in counted squares. Better than a donut
	// whenever the point is "how many in a hundred" rather than "which slice
	// is biggest", because people read areas badly and count squares well.
	"waffle": {shares: true, maxPoints: 5},
	// A row of dials, each filled against the largest. For a handful of rates
	// where the reader's question is "how full", not "how do these compare".
	"gauge": {maxPoints: 5},
	// Nested rectangles sized by value. For a breakdown with a wide spread,
	// where bars would leave the small entries invisible.
	"treemap": {maxPoints: 8},
	// Stages narrowing to a result. The values must fall.
	"funnel": {descending: true, maxPoints: 6},
	// Big numbers as cards, no axes at all. For a clip whose data is three
	// headline figures rather than a distribution.
	"kpi": {maxPoints: 5},
	// Countries shaded by value.
	"map": {maxPoints: maxDataPoints},
}

func ChartKindNames() []string {
	out := make([]string, 0, len(chartKindVocab))
	for k := range chartKindVocab {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func normalizeChartKind(s string) string {
	n := strings.ToLower(strings.TrimSpace(s))
	// Tolerate the spellings a model reaches for first. These are not aliases
	// we advertise; they are the near-misses that would otherwise silently
	// degrade a stacked bar chart to plain bars and drop every series.
	switch n {
	case "stacked", "stacked-bars", "stacked_bars", "stackedbar":
		n = "stackedbars"
	case "grouped", "grouped-bars", "grouped_bars", "groupedbar", "bar-group":
		n = "groupedbars"
	case "kpicards", "kpi-cards", "cards", "bignumbers":
		n = "kpi"
	}
	if _, ok := chartKindVocab[n]; ok {
		return n
	}
	return "bars"
}

// resolveCountry maps what a model wrote onto what the atlas calls it,
// returning "" when there is no such country.
func resolveCountry(name string) string {
	trimmed := strings.TrimSpace(name)
	if mapCountryVocab[trimmed] {
		return trimmed
	}
	if canonical, ok := mapCountryAliases[strings.ToLower(trimmed)]; ok {
		return canonical
	}
	// Last resort: a case-insensitive sweep, so "india" finds "India" without
	// needing an alias for every country in the atlas.
	for known := range mapCountryVocab {
		if strings.EqualFold(known, trimmed) {
			return known
		}
	}
	return ""
}

// dataScenes lays the clip out as one persistent chart whose highlight follows
// the narration — the same shape as the whiteboard and the flow diagram, and
// for the same reason: what accumulates is the viewer's understanding of one
// picture.
func dataScenes(in SnippetSceneInput) ([]Scene, error) {
	chart := in.Plan.Chart
	if chart == nil {
		return nil, fmt.Errorf("the plan has no chart")
	}
	kind := normalizeChartKind(chart.Kind)

	points := make([]map[string]any, 0, len(chart.Points))
	for _, p := range chart.Points {
		label := p.Label
		if kind == "map" {
			if c := resolveCountry(label); c != "" {
				label = c
			}
		}
		point := map[string]any{"label": label, "value": p.total()}
		// The parts ride alongside the total rather than replacing it, so a
		// renderer that only knows how to draw one number still draws the right
		// bar rather than nothing.
		if len(p.Values) > 0 {
			point["values"] = p.Values
		}
		points = append(points, point)
	}

	// Highlight windows, one per beat that names anything — the same mechanism
	// the flow diagram's focus uses, and it carries the same meaning: what is
	// lit is what is being talked about right now.
	windows := make([]map[string]any, 0, len(in.Plan.Beats))
	_, clipStart, _ := in.Beat(0)
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Data == nil || len(beat.Data.Highlight) == 0 {
			continue
		}
		labels := make([]string, 0, len(beat.Data.Highlight))
		for _, h := range beat.Data.Highlight {
			label := h
			if kind == "map" {
				if c := resolveCountry(h); c != "" {
					label = c
				}
			}
			labels = append(labels, label)
		}
		windows = append(windows, map[string]any{
			"startMs": startMs, "endMs": endMs, "labels": labels,
		})
	}

	captions := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Data == nil || strings.TrimSpace(beat.Data.Caption) == "" {
			continue
		}
		captions = append(captions, map[string]any{
			"startMs": startMs, "endMs": endMs, "text": beat.Data.Caption,
		})
	}

	_, _, lastEnd := in.Beat(len(in.Plan.Beats) - 1)
	props := map[string]any{
		"title":     in.Plan.Title,
		"kind":      kind,
		"unit":      chart.Unit,
		"points":    points,
		"highlight": windows,
		"captions":  captions,
	}
	if len(chart.Series) > 0 {
		props["series"] = chart.Series
	}
	return []Scene{{
		Type:    SceneData,
		StartMs: clipStart,
		EndMs:   lastEnd,
		Props:   props,
	}}, nil
}

// normalizeDataPlan tidies the dataset without touching the numbers.
//
// Labels are the one thing worth repairing here, because they are the chart's
// join key: a beat highlights "United States" against a point labelled "USA"
// and the highlight silently selects nothing, which reads on screen as a beat
// that talks about a bar while pointing at all of them. Values are never
// touched — a chart is a claim about the world, and a claim is not something to
// round into shape.
func normalizeDataPlan(p *SnippetPlan) {
	if p.Chart == nil {
		return
	}
	p.Chart.Kind = normalizeChartKind(p.Chart.Kind)
	p.Chart.Unit = strings.TrimSpace(p.Chart.Unit)

	labels := map[string]string{}
	points := make([]DataPoint, 0, len(p.Chart.Points))
	for _, pt := range p.Chart.Points {
		pt.Label = clampWords(collapseSpaces(pt.Label), maxDataLabelWords)
		if pt.Label == "" {
			continue
		}
		labels[strings.ToLower(pt.Label)] = pt.Label
		points = append(points, pt)
	}
	p.Chart.Points = points

	for i := range p.Beats {
		b := &p.Beats[i]
		if b.Data == nil {
			continue
		}
		b.Data.Caption = collapseSpaces(b.Data.Caption)
		hits := make([]string, 0, len(b.Data.Highlight))
		for _, h := range b.Data.Highlight {
			if label, ok := labels[strings.ToLower(collapseSpaces(h))]; ok && !slices.Contains(hits, label) {
				hits = append(hits, label)
			}
		}
		b.Data.Highlight = hits
	}
}

func validateDataPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Data: true}); err != nil {
		return err
	}
	if p.Chart == nil {
		return fmt.Errorf("the plan has no chart — this template is one dataset read several ways")
	}
	kindName := normalizeChartKind(p.Chart.Kind)
	kind := chartKindVocab[kindName]

	max := kind.maxPoints
	if max == 0 || max > maxDataPoints {
		max = maxDataPoints
	}
	if n := len(p.Chart.Points); n < minDataPoints || n > max {
		return fmt.Errorf("the chart has %d points, want %d-%d for kind %q", n, minDataPoints, max, kindName)
	}

	if err := checkChartSeries(p.Chart, kindName, kind); err != nil {
		return err
	}

	known := map[string]bool{}
	var prevTotal float64
	for i, pt := range p.Chart.Points {
		label := strings.TrimSpace(pt.Label)
		if label == "" {
			return fmt.Errorf("a data point has no label")
		}
		if n := len(strings.Fields(label)); n > maxDataLabelWords {
			return fmt.Errorf("data label %q is %d words; labels are at most %d", label, n, maxDataLabelWords)
		}
		// Negative values break every one of these charts differently — a bar
		// that grows downward, an arc with negative sweep, a treemap tile with
		// no area, a country the colour scale has no room for. Rejecting them
		// beats rendering nonsense.
		for _, v := range append([]float64{pt.Value}, pt.Values...) {
			if v < 0 {
				return fmt.Errorf("data point %q has a negative value (%v); these charts show magnitudes", label, v)
			}
		}
		if kindName == "map" {
			canonical := resolveCountry(label)
			if canonical == "" {
				return fmt.Errorf("the map has no country called %q — use the country's common English name", label)
			}
			label = canonical
		}
		key := strings.ToLower(label)
		if known[key] {
			return fmt.Errorf("data point %q appears twice", label)
		}
		known[key] = true

		// A funnel is a claim about attrition. Drawn from values that rise it
		// is a picture of a funnel with the numbers written on it, which is
		// worse than no chart because it looks like it means something.
		if kind.descending {
			if i > 0 && pt.total() > prevTotal {
				return fmt.Errorf("funnel stage %q (%v) is larger than the stage before it (%v) — a funnel narrows; put the stages in order or use bars",
					label, pt.total(), prevTotal)
			}
			prevTotal = pt.total()
		}
	}

	if kind.shares {
		total := 0.0
		for _, pt := range p.Chart.Points {
			total += pt.total()
		}
		if total <= 0 {
			return fmt.Errorf("the %s's values total zero — there are no shares to draw", kindName)
		}
	}

	highlighting := 0
	var prev string
	for _, b := range p.Beats {
		if b.Data == nil {
			return fmt.Errorf("beat %q has no data direction", b.ID)
		}
		if n := len(strings.Fields(b.Data.Caption)); n > maxCaptionWords {
			return fmt.Errorf("beat %q has a %d-word caption; at most %d", b.ID, n, maxCaptionWords)
		}
		if len(b.Data.Highlight) == 0 {
			prev = ""
			continue
		}
		highlighting++
		labels := make([]string, 0, len(b.Data.Highlight))
		for _, h := range b.Data.Highlight {
			label := strings.TrimSpace(h)
			if kindName == "map" {
				if c := resolveCountry(label); c != "" {
					label = c
				}
			}
			if !known[strings.ToLower(label)] {
				return fmt.Errorf("beat %q highlights %q, which is not one of the chart's points", b.ID, h)
			}
			labels = append(labels, strings.ToLower(label))
		}
		sort.Strings(labels)
		key := strings.Join(labels, "|")
		// Two beats in a row lighting the same thing is thirty seconds of one
		// unchanging picture. The whole reason the chart persists is that the
		// emphasis moves around it.
		if key == prev {
			return fmt.Errorf("beat %q highlights exactly what the previous beat did — move the emphasis or highlight nothing", b.ID)
		}
		prev = key
	}
	// A chart nobody points at is a screenshot with narration over it.
	if highlighting*2 < len(p.Beats) {
		return fmt.Errorf("only %d of %d beats highlight anything; at least half should point at part of the chart",
			highlighting, len(p.Beats))
	}
	return nil
}

// checkChartSeries enforces the shape the kind needs: the right number of
// dimensions declared, and every point carrying exactly that many numbers.
//
// This is checked hard rather than padded because a short row is not a drawing
// bug, it is a wrong chart — a stacked bar missing its third segment renders
// perfectly and states a total that is not the total.
func checkChartSeries(c *ChartSpec, name string, kind chartKind) error {
	if kind.series == 0 {
		if len(c.Series) > 0 {
			return fmt.Errorf("kind %q takes one number per point, but the chart declares series %v — drop them or pick stackedbars, groupedbars or scatter",
				name, c.Series)
		}
		for _, pt := range c.Points {
			if len(pt.Values) > 0 {
				return fmt.Errorf("point %q carries a values array, but kind %q takes a single value", pt.Label, name)
			}
		}
		return nil
	}

	want := kind.series
	if want < 0 {
		if n := len(c.Series); n < minChartSeries || n > maxChartSeries {
			return fmt.Errorf("kind %q needs %d-%d series, got %d", name, minChartSeries, maxChartSeries, len(c.Series))
		}
		want = len(c.Series)
	} else if len(c.Series) != want {
		return fmt.Errorf("kind %q needs exactly %d series (it plots one against the other), got %d", name, want, len(c.Series))
	}

	seen := map[string]bool{}
	for _, s := range c.Series {
		n := strings.TrimSpace(s)
		if n == "" {
			return fmt.Errorf("a series has no name")
		}
		if w := len(strings.Fields(n)); w > maxDataLabelWords {
			return fmt.Errorf("series name %q is %d words; names are at most %d", n, w, maxDataLabelWords)
		}
		if seen[strings.ToLower(n)] {
			return fmt.Errorf("series %q appears twice", n)
		}
		seen[strings.ToLower(n)] = true
	}
	for _, pt := range c.Points {
		if len(pt.Values) != want {
			return fmt.Errorf("point %q has %d values but the chart declares %d series (%s)",
				pt.Label, len(pt.Values), want, strings.Join(c.Series, ", "))
		}
	}
	return nil
}
