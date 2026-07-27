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

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "data",
		Title:       "Data & maps",
		Description: "Real numbers on one chart or world map, with the narration walking around it.",
		Example:     "Where the world's undersea internet cables actually land",
		PromptFile:  snippetDataTemplateName,
		NeedsCode:   false,
		Validate:    validateDataPlan,
		Scenes:      dataScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"ChartKinds":      strings.Join(ChartKindNames(), ", "),
				"MinPoints":       minDataPoints,
				"MaxPoints":       maxDataPoints,
				"MaxLabelWords":   maxDataLabelWords,
				"MaxCaptionWords": maxCaptionWords,
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

// chartKindVocab mirrors the switch in renderer/src/components/DataScene.tsx.
var chartKindVocab = map[string]bool{
	// Horizontal bars. The default, and right for almost any comparison —
	// horizontal because it is the only orientation where a real-world label
	// fits without being rotated.
	"bars": true,
	// A value over an ordered sequence: time, versions, sizes.
	"line": true,
	// Parts of one whole. Only honest when the values are shares of something.
	"donut": true,
	// Countries shaded by value.
	"map": true,
}

func ChartKindNames() []string { return sortedKeys(chartKindVocab) }

func normalizeChartKind(s string) string {
	n := strings.ToLower(strings.TrimSpace(s))
	if chartKindVocab[n] {
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
		points = append(points, map[string]any{"label": label, "value": p.Value})
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
	return []Scene{{
		Type:    SceneData,
		StartMs: clipStart,
		EndMs:   lastEnd,
		Props: map[string]any{
			"title":     in.Plan.Title,
			"kind":      kind,
			"unit":      chart.Unit,
			"points":    points,
			"highlight": windows,
			"captions":  captions,
		},
	}}, nil
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
	kind := normalizeChartKind(p.Chart.Kind)
	if n := len(p.Chart.Points); n < minDataPoints || n > maxDataPoints {
		return fmt.Errorf("the chart has %d points, want %d-%d", n, minDataPoints, maxDataPoints)
	}

	known := map[string]bool{}
	for _, pt := range p.Chart.Points {
		label := strings.TrimSpace(pt.Label)
		if label == "" {
			return fmt.Errorf("a data point has no label")
		}
		if n := len(strings.Fields(label)); n > maxDataLabelWords {
			return fmt.Errorf("data label %q is %d words; labels are at most %d", label, n, maxDataLabelWords)
		}
		// Negative values break every one of these charts differently — a bar
		// that grows downward, an arc with negative sweep, a country the colour
		// scale has no room for. Rejecting them beats rendering nonsense.
		if pt.Value < 0 {
			return fmt.Errorf("data point %q has a negative value (%v); these charts show magnitudes", label, pt.Value)
		}
		if kind == "map" {
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
	}
	if kind == "donut" {
		total := 0.0
		for _, pt := range p.Chart.Points {
			total += pt.Value
		}
		if total <= 0 {
			return fmt.Errorf("the donut's values total zero — there are no shares to draw")
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
			if kind == "map" {
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
