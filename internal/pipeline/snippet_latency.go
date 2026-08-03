package pipeline

// The latency template: operations placed on a logarithmic time axis.
//
// The honest boundary with `data` is the axis, and it is worth stating plainly
// because a careless version of this template would just be a bar chart. A bar
// chart is linear, and linear is the right scale when the quantities are within
// an order of magnitude of each other. It is the wrong scale — actively
// misleading — the moment they are not: one millisecond next to three hundred
// draws as an invisible bar beside a full one, which says "the first is nothing"
// when what is true is "the first is in a different category of fast".
//
// So this template does the thing a chart cannot: a log axis with its decades
// NAMED — 1ms, 10ms, 100ms, 1s — and operations placed along it. The reading is
// positional and categorical rather than proportional. A viewer does not learn
// "Docker is three hundred times slower"; they learn "these two are not in the
// same decade", which is the claim the reference clips are actually making when
// they put `<1ms` next to `~hundreds of ms`.
//
// Two rules earn it its place, and both are validators.
//
// The axis is established before anything is placed on it. A mark on an unlabelled
// line is a mark with no scale, and the decades are the whole reason this is not
// a chart.
//
// **The operations must span at least two decade boundaries.** This is the rule
// the template exists for. Inside one decade a linear chart is honest, easier to
// read and already in the catalog — a log axis there is a scale chosen to make a
// small difference look structural, which is the exact dishonesty this template
// is otherwise built to avoid. The error names `data` as the right tool.

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:             "latency",
		Category:         CatNumbers,
		Since:            SinceV4,
		Family:           FamilyReplica,
		Title:            "Different categories of fast",
		Description:      "Operations placed on a logarithmic time axis with its decades named. Reach for it when two things are orders of magnitude apart and a bar chart would draw one of them as nothing.",
		Example:          "Why an in-memory read and a disk query are not the same kind of slow",
		PromptFile:       snippetLatencyTemplateName,
		NeedsCode:        false,
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		MaxBeats:         8,
		Owns:             beatFields{Latency: true},
		OwnsPlan:         planFields{Latency: true},
		Normalize:        normalizeLatencyPlan,
		Validate:         validateLatencyPlan,
		Scenes:           latencyScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(LatencyShows(), ", "),
				"MinOps":        minLatencyOps,
				"MaxOps":        maxLatencyOps,
				"MinDecades":    minLatencyDecades,
				"MaxLabelWords": maxLatencyLabelWords,
				"MaxNoteWords":  maxLatencyNoteWords,
			}
		},
	})
}

const snippetLatencyTemplateName = "snippet_latency.tmpl"

const (
	// Two operations is the comparison this exists for; five rows on one axis is
	// as many as can be labelled without the labels colliding.
	minLatencyOps = 2
	maxLatencyOps = 5

	// The span, in decade boundaries crossed. Two is the floor — see the header.
	minLatencyDecades = 2

	maxLatencyLabelWords = 5
	maxLatencyNoteWords  = 16
)

// latencyShows is the closed vocabulary of what a beat does.
var latencyShows = map[string]bool{
	// The axis and its named decades, nothing on it. The first beat, always.
	"axis": true,
	// One operation lands on the axis.
	"place": true,
	// Hold the axis and say what the gap means.
	"read": true,
}

// LatencyShows returns the beat vocabulary sorted.
func LatencyShows() []string {
	out := make([]string, 0, len(latencyShows))
	for k := range latencyShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// LatencySpec is the set of operations being timed. On the plan because they all
// share one axis and the axis is derived from all of them at once.
type LatencySpec struct {
	// Operations are the things being timed, in the order they are placed.
	Operations []LatencyOp `json:"operations"`
}

// LatencyOp is one timed operation.
type LatencyOp struct {
	// Label is what the operation is.
	Label string `json:"label"`
	// Ms is how long it takes, in milliseconds. Milliseconds always, whatever
	// the display unit ends up being: one unit for the whole axis is what makes
	// the placements comparable, and letting each row pick its own is how a
	// clip ends up putting a microsecond and a minute on the same tick.
	Ms float64 `json:"ms"`
	// Note is what this duration means. One sentence.
	Note string `json:"note,omitempty"`
	// Role is what this operation is doing: a metricRoles name.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the operation's role, defaulting to neutral.
func (o LatencyOp) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(o.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// LatencyBeat is one move: which operation this beat places.
type LatencyBeat struct {
	// Show is a latencyShows name.
	Show string `json:"show"`
	// At indexes LatencySpec.Operations, for a "place" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to a placement.
func (b LatencyBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if latencyShows[s] {
		return s
	}
	return "place"
}

// latencyAxis is the decade range the axis covers: the powers of ten just below
// the fastest operation and just above the slowest.
//
// Padding out to whole decades rather than fitting the data exactly is what
// makes the axis readable: a tick at 1ms means something to a viewer, and a tick
// at 0.7ms means they have to do arithmetic before the picture says anything.
func latencyAxis(ops []LatencyOp) (loExp, hiExp int) {
	if len(ops) == 0 {
		return 0, 1
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, o := range ops {
		if o.Ms > 0 {
			lo = math.Min(lo, o.Ms)
			hi = math.Max(hi, o.Ms)
		}
	}
	if math.IsInf(lo, 1) || math.IsInf(hi, -1) {
		return 0, 1
	}
	loExp = int(math.Floor(math.Log10(lo)))
	hiExp = int(math.Ceil(math.Log10(hi)))
	// A value sitting exactly on a decade would otherwise land on the very edge
	// of the axis, where its label has nowhere to go.
	if hiExp == loExp {
		hiExp = loExp + 1
	}
	return loExp, hiExp
}

// latencyDecades is how many decade boundaries the operations span.
func latencyDecades(ops []LatencyOp) int {
	lo, hi := latencyAxis(ops)
	return hi - lo
}

// latencyFrac is where a duration sits along the axis, 0 at the left edge and 1
// at the right. Computed in Go so the placement is a fact the scene graph
// records rather than something the renderer re-derives with its own log.
func latencyFrac(ms float64, loExp, hiExp int) float64 {
	if ms <= 0 || hiExp <= loExp {
		return 0
	}
	f := (math.Log10(ms) - float64(loExp)) / float64(hiExp-loExp)
	return math.Max(0, math.Min(1, f))
}

// latencyLabel renders a duration the way somebody would say it out loud, which
// is not always in the unit it was given in. 6479ms is "6.5s" to a human and
// "6479ms" only to a log file.
func latencyLabel(ms float64) string {
	switch {
	case ms < 1:
		return fmt.Sprintf("%gµs", roundTo(ms*1000, 0))
	case ms < 1000:
		return fmt.Sprintf("%gms", roundTo(ms, 1))
	case ms < 60000:
		return fmt.Sprintf("%gs", roundTo(ms/1000, 1))
	default:
		return fmt.Sprintf("%gmin", roundTo(ms/60000, 1))
	}
}

func normalizeLatencyPlan(p *SnippetPlan) {
	l := p.Latency
	if l == nil {
		return
	}
	ops := make([]LatencyOp, 0, len(l.Operations))
	for _, o := range l.Operations {
		o.Label = clampWords(collapseSpaces(o.Label), maxLatencyLabelWords)
		o.Note = clampWords(collapseSpaces(o.Note), maxLatencyNoteWords)
		o.Role = o.ResolvedRole()
		// A duration of zero has no place on a log axis — log10(0) is undefined,
		// and "instant" is a claim no measurement supports.
		if o.Label != "" && o.Ms > 0 && len(ops) < maxLatencyOps {
			ops = append(ops, o)
		}
	}
	l.Operations = ops

	for i := range p.Beats {
		b := p.Beats[i].Latency
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "place" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(l.Operations); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateLatencyPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Latency: true}); err != nil {
		return err
	}

	l := p.Latency
	if l == nil {
		return fmt.Errorf("the plan has no operations — this template is a time axis with things placed along it")
	}
	if n := len(l.Operations); n < minLatencyOps || n > maxLatencyOps {
		return fmt.Errorf("there are %d operations, want %d-%d. One is a figure and belongs to metric; six rows on one axis cannot be labelled without the labels colliding",
			n, minLatencyOps, maxLatencyOps)
	}

	seen := map[string]bool{}
	for i, o := range l.Operations {
		if strings.TrimSpace(o.Label) == "" {
			return fmt.Errorf("operation %d has no label", i)
		}
		if o.Ms <= 0 {
			return fmt.Errorf("operation %q takes %v ms. A log axis has no place for zero, and \"instant\" is a claim no measurement supports — give it the number you actually measured",
				o.Label, o.Ms)
		}
		key := strings.ToLower(strings.TrimSpace(o.Label))
		if seen[key] {
			return fmt.Errorf("two operations are both %q — each row on the axis is a different thing being timed", o.Label)
		}
		seen[key] = true
		if r := strings.ToLower(strings.TrimSpace(o.Role)); r != "" && !metricRoles[r] {
			return fmt.Errorf("operation %d has role %q, which is not one of: %s", i, o.Role, strings.Join(MetricRoles(), ", "))
		}
	}

	// The rule the template exists for. Inside one decade a linear chart is
	// honest, easier to read, and already in the catalog.
	if d := latencyDecades(l.Operations); d < minLatencyDecades {
		return fmt.Errorf("every operation sits inside %d decade(s) of each other, so a logarithmic axis is a scale chosen to make a small difference look structural. Use the data template for a linear chart, or find the operation that really is orders of magnitude away — this template exists for the case where a bar chart would draw the fast one as nothing",
			d)
	}

	// The axis is named before anything is placed on it.
	if p.Beats[0].Latency == nil || p.Beats[0].Latency.ResolvedShow() != "axis" {
		return fmt.Errorf("beat %q does not establish the axis. A mark on an unlabelled line has no scale, and the named decades are the whole reason this is not a bar chart",
			p.Beats[0].ID)
	}

	placed := map[int]bool{}
	axes := 0
	for i, b := range p.Beats {
		if b.Latency == nil {
			return fmt.Errorf("beat %q has no latency direction — every beat draws the axis, places an operation, or reads the gap", b.ID)
		}
		switch b.Latency.ResolvedShow() {
		case "axis":
			axes++
			if i != 0 {
				return fmt.Errorf("beat %q draws the axis again part-way through. It is established once, at the start", b.ID)
			}
		case "place":
			if b.Latency.At < 0 || b.Latency.At >= len(l.Operations) {
				return fmt.Errorf("beat %q places operation %d, which does not exist", b.ID, b.Latency.At)
			}
			if placed[b.Latency.At] {
				return fmt.Errorf("beat %q places operation %d again; each one lands once", b.ID, b.Latency.At)
			}
			placed[b.Latency.At] = true
		}
	}
	if axes != 1 {
		return fmt.Errorf("there are %d beats establishing the axis, want exactly 1", axes)
	}
	if len(placed) != len(l.Operations) {
		return fmt.Errorf("%d of the %d operations never land. A row the narrator skips is one nobody read — give it a beat or cut it",
			len(l.Operations)-len(placed), len(l.Operations))
	}
	return nil
}

// latencyScenes lays the clip out as ONE scene. The axis persists and the beats
// only say what has been placed on it.
func latencyScenes(in SnippetSceneInput) ([]Scene, error) {
	l := in.Plan.Latency
	if l == nil {
		return nil, fmt.Errorf("the plan has no operations")
	}

	loExp, hiExp := latencyAxis(l.Operations)
	ticks := make([]map[string]any, 0, hiExp-loExp+1)
	for e := loExp; e <= hiExp; e++ {
		ms := math.Pow(10, float64(e))
		ticks = append(ticks, map[string]any{
			"label": latencyLabel(ms),
			"frac":  roundTo(float64(e-loExp)/float64(hiExp-loExp), 4),
		})
	}

	ops := make([]map[string]any, len(l.Operations))
	for i, o := range l.Operations {
		ops[i] = map[string]any{
			"label": o.Label,
			"value": latencyLabel(o.Ms),
			"note":  o.Note,
			"role":  o.ResolvedRole(),
			"frac":  roundTo(latencyFrac(o.Ms, loExp, hiExp), 4),
		}
	}

	placed := map[int]bool{}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Latency == nil {
			return nil, fmt.Errorf("beat %q has no latency direction", beat.ID)
		}
		show := beat.Latency.ResolvedShow()
		if show == "place" {
			placed[beat.Latency.At] = true
		}
		shown := make([]int, 0, len(placed))
		for at := range placed {
			shown = append(shown, at)
		}
		sort.Ints(shown)

		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"placed":  shown,
		}
		if show == "place" {
			step["at"] = beat.Latency.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneLatency,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":      in.Plan.Title,
			"ticks":      ticks,
			"operations": ops,
			"steps":      steps,
		}),
	}}, nil
}
