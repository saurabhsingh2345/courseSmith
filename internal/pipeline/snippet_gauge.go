package pipeline

// The gauge template: does it fit?
//
// This is the single most repeated device in the reference clips, and it is one
// the catalog had no way to draw. A bar fills toward a marked line and either
// clears it or runs past it, and that picture answers a question — will this
// model fit in that card, will this bundle fit in that budget, will this job
// finish inside that window — in less time than the narrator takes to ask it.
//
// `data` can draw a bar chart, and a bar chart is a different claim. A chart
// compares magnitudes to each other; a gauge compares them to a *threshold*,
// and the threshold is the whole point. The moment that teaches is the fill
// crossing the line, which a chart has no way to express.
//
// Two rules earn it its place, and both are validators.
//
// The first is that the ceiling is established before anything is measured
// against it. A bar filling toward an unexplained line is a bar with no
// meaning; the viewer has to know what the line *is* before they can care
// whether something clears it. So the first beat is the ceiling, always.
//
// The second is the ratio cap. If a bar is forty times the ceiling, the ceiling
// renders as a hairline against the left edge and the picture stops saying
// anything at all — the honest visual for "this is off the scale" is the number
// itself, which is what `metric` is for. Refusing to draw it is better than
// drawing something unreadable and calling it a diagram.

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "gauge",
		Category:    CatNumbers,
		Since:       SinceV1,
		Title:       "Does it fit?",
		Description: "A bar filling toward a marked ceiling. Reach for it whenever the question is whether something fits — memory, budget, a latency target.",
		Example:     "Which of these models actually fit in 24GB of VRAM",
		PromptFile:  snippetGaugeTemplateName,
		NeedsCode:   false,
		// The ceiling, three candidates and a verdict is five beats before
		// anything optional.
		MinTargetSec:     40,
		DefaultTargetSec: 55,
		MaxBeats:         8,
		Owns:             beatFields{Gauge: true},
		OwnsPlan:         planFields{Gauge: true},
		Normalize:        normalizeGaugePlan,
		Validate:         validateGaugePlan,
		Scenes:           gaugeScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":         strings.Join(GaugeShows(), ", "),
				"MinBars":       minGaugeBars,
				"MaxBars":       maxGaugeBars,
				"MaxRatio":      maxGaugeRatio,
				"MaxUnitWords":  maxGaugeUnitWords,
				"MaxLabelWords": maxGaugeLabelWords,
				"MaxNoteWords":  maxGaugeNoteWords,
			}
		},
	})
}

const snippetGaugeTemplateName = "snippet_gauge.tmpl"

const (
	// Two bars against a line is the smallest question worth asking; six is a
	// chart, and a chart wants the data template.
	minGaugeBars = 2
	maxGaugeBars = 5

	// How far past the ceiling a bar may reach before the picture stops
	// working. At 4x the ceiling still sits a quarter of the way across and is
	// legible as a line; at 40x it is a hairline against the left edge and the
	// diagram conveys nothing the number would not convey better.
	maxGaugeRatio = 4.0

	maxGaugeUnitWords  = 2
	maxGaugeLabelWords = 6
	maxGaugeNoteWords  = 14
)

// gaugeShows is the closed vocabulary of what a beat does.
var gaugeShows = map[string]bool{
	// Establish the line everything is measured against. The first beat.
	"ceiling": true,
	// Run one candidate against it: the bar fills and lands.
	"bar": true,
	// Every bar back at once, with the line through them.
	"verdict": true,
}

// GaugeShows returns the beat vocabulary sorted.
func GaugeShows() []string {
	out := make([]string, 0, len(gaugeShows))
	for k := range gaugeShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// GaugeSpec is the question the clip answers.
type GaugeSpec struct {
	// Unit is what everything here is measured in — "GB", "ms", "$". One unit
	// per clip: a gauge measuring two different things is two gauges.
	Unit string `json:"unit"`
	// Ceiling is the line. Establishing it is the first beat.
	Ceiling GaugeCeiling `json:"ceiling"`
	// Bars are the candidates run against it, in the order they are run.
	Bars []GaugeBar `json:"bars"`
}

// GaugeCeiling is the threshold everything is measured against.
type GaugeCeiling struct {
	Value float64 `json:"value"`
	// Label is what the line IS — "What a 4090 holds", "The 200ms budget".
	Label string `json:"label"`
	// Note is the one line explaining why that is the limit.
	Note string `json:"note,omitempty"`
}

// GaugeBar is one candidate measured against the ceiling.
type GaugeBar struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
	// Note is what this result means — the argument, not the arithmetic.
	Note string `json:"note,omitempty"`
}

// Fits reports whether this bar clears the ceiling. Equality counts as fitting:
// a model that needs exactly the memory available does fit, and the narrator's
// "just barely" is a better place for that nuance than a rounding rule.
func (b GaugeBar) Fits(ceiling float64) bool { return b.Value <= ceiling }

// GaugeBeat is one move.
type GaugeBeat struct {
	Show string `json:"show"`
	At   int    `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to running a
// bar — which is what most beats of this template do.
func (b GaugeBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if gaugeShows[s] {
		return s
	}
	return "bar"
}

func normalizeGaugePlan(p *SnippetPlan) {
	g := p.Gauge
	if g == nil {
		return
	}
	g.Unit = clampWords(collapseSpaces(g.Unit), maxGaugeUnitWords)
	g.Ceiling.Label = clampWords(collapseSpaces(g.Ceiling.Label), maxGaugeLabelWords)
	g.Ceiling.Note = clampWords(collapseSpaces(g.Ceiling.Note), maxGaugeNoteWords)
	g.Ceiling.Value = math.Abs(g.Ceiling.Value)

	bars := make([]GaugeBar, 0, len(g.Bars))
	for _, b := range g.Bars {
		b.Label = clampWords(collapseSpaces(b.Label), maxGaugeLabelWords)
		b.Note = clampWords(collapseSpaces(b.Note), maxGaugeNoteWords)
		// A negative measurement is a sign error, not a quantity; the magnitude
		// is what was meant.
		b.Value = math.Abs(b.Value)
		if b.Label != "" && b.Value > 0 && len(bars) < maxGaugeBars {
			bars = append(bars, b)
		}
	}
	g.Bars = bars

	for i := range p.Beats {
		b := p.Beats[i].Gauge
		if b == nil {
			continue
		}
		if !gaugeShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			// The shape says what the ends are for: the clip opens on the
			// ceiling and everything else runs a bar.
			if i == 0 {
				b.Show = "ceiling"
			} else {
				b.Show = "bar"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "bar" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(g.Bars); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateGaugePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Gauge: true}); err != nil {
		return err
	}

	g := p.Gauge
	if g == nil {
		return fmt.Errorf("the plan has no gauge — this template is one marked line and the things measured against it")
	}
	if strings.TrimSpace(g.Unit) == "" {
		return fmt.Errorf("the gauge has no unit — everything on it is measured in the same thing, and the viewer needs to know what")
	}
	if g.Ceiling.Value <= 0 {
		return fmt.Errorf("the ceiling is %v — the line has to be a real quantity, or nothing is being measured against anything", g.Ceiling.Value)
	}
	if strings.TrimSpace(g.Ceiling.Label) == "" {
		return fmt.Errorf("the ceiling has no label. A bar filling toward an unexplained line is a bar with no meaning — say what the line IS")
	}
	if n := len(g.Bars); n < minGaugeBars || n > maxGaugeBars {
		return fmt.Errorf("there are %d bars, want %d-%d. One bar against a line is a single fact and belongs in metric; six is a chart, and a chart wants the data template",
			n, minGaugeBars, maxGaugeBars)
	}

	seen := map[string]bool{}
	for i, b := range g.Bars {
		if strings.TrimSpace(b.Label) == "" {
			return fmt.Errorf("bar %d has no label", i)
		}
		if b.Value <= 0 {
			return fmt.Errorf("bar %q measures %v — every bar is a real quantity", b.Label, b.Value)
		}
		// The ratio cap: past this the ceiling is a hairline and the picture
		// stops being a picture.
		if r := b.Value / g.Ceiling.Value; r > maxGaugeRatio {
			return fmt.Errorf("bar %q is %.1fx the ceiling, past the %gx this can draw. At that ratio the line renders as a hairline against the left edge and the diagram says nothing the number would not say better — either raise the ceiling to something %q is actually close to, state it as a figure with the metric template, or draw the gap with the scale template, which nests each quantity inside the next and is built for exactly the ratios a bar cannot hold",
				b.Label, r, maxGaugeRatio, b.Label)
		}
		key := strings.ToLower(strings.TrimSpace(b.Label))
		if seen[key] {
			return fmt.Errorf("two bars are both labelled %q — each one measures a different thing", b.Label)
		}
		seen[key] = true
	}

	run := map[int]bool{}
	counts := map[string]int{}
	for i, b := range p.Beats {
		if b.Gauge == nil {
			return fmt.Errorf("beat %q has no gauge direction — every beat sets the line, runs a bar, or delivers the verdict", b.ID)
		}
		show := b.Gauge.ResolvedShow()
		counts[show]++
		// The rule this template exists for.
		if i == 0 && show != "ceiling" {
			return fmt.Errorf("the clip opens on %q. Establish the line first — until the viewer knows what the ceiling is, a bar filling toward it means nothing", show)
		}
		if show == "verdict" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q delivers the verdict but the clip carries on afterwards. The verdict is the last frame", b.ID)
		}
		if show != "bar" {
			continue
		}
		if b.Gauge.At < 0 || b.Gauge.At >= len(g.Bars) {
			return fmt.Errorf("beat %q runs bar %d, which does not exist", b.ID, b.Gauge.At)
		}
		if run[b.Gauge.At] {
			return fmt.Errorf("beat %q runs bar %d again; each one is measured once", b.ID, b.Gauge.At)
		}
		run[b.Gauge.At] = true
	}
	if counts["ceiling"] != 1 {
		return fmt.Errorf("there are %d ceiling beats; the line is established once", counts["ceiling"])
	}
	if len(run) != len(g.Bars) {
		return fmt.Errorf("%d of the %d bars are never run. A bar drawn but not narrated is a measurement nobody witnessed — give it a beat or cut it",
			len(g.Bars)-len(run), len(g.Bars))
	}
	if counts["verdict"] > 1 {
		return fmt.Errorf("there are %d verdict beats; the answer lands once", counts["verdict"])
	}
	return nil
}

// gaugeScenes lays the clip out as ONE scene: the ceiling is on screen from the
// moment it is set, and the bars fill under it in turn.
func gaugeScenes(in SnippetSceneInput) ([]Scene, error) {
	g := in.Plan.Gauge
	if g == nil {
		return nil, fmt.Errorf("the plan has no gauge")
	}

	// The axis. Everything is drawn against one scale, and it has to hold the
	// ceiling and the longest bar with a little air past it — a bar that ends
	// exactly at the frame edge reads as clipped rather than as measured.
	maxV := g.Ceiling.Value
	for _, b := range g.Bars {
		if b.Value > maxV {
			maxV = b.Value
		}
	}
	scale := maxV * 1.08

	bars := make([]map[string]any, len(g.Bars))
	for i, b := range g.Bars {
		bars[i] = map[string]any{
			"label": b.Label,
			"value": b.Value,
			"note":  b.Note,
			// Whether it clears is decided here rather than in the renderer, so
			// the colour of the bar and the words of the validator can never
			// disagree about what "fits" means.
			"fits": b.Fits(g.Ceiling.Value),
			"frac": b.Value / scale,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Gauge == nil {
			return nil, fmt.Errorf("beat %q has no gauge direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Gauge.ResolvedShow(),
		}
		if beat.Gauge.ResolvedShow() == "bar" {
			step["at"] = beat.Gauge.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneGauge,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title": in.Plan.Title,
			"unit":  g.Unit,
			"ceiling": map[string]any{
				"value": g.Ceiling.Value,
				"label": g.Ceiling.Label,
				"note":  g.Ceiling.Note,
				"frac":  g.Ceiling.Value / scale,
			},
			"bars":  bars,
			"steps": steps,
		},
	}}, nil
}
