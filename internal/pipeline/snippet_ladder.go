package pipeline

// The ladder template: the memory hierarchy as rungs, and a request falling
// down them.
//
// Every CS foundations course has a lesson where the numbers stop being
// abstract: a register is touched in a fraction of a nanosecond, a disk read
// costs ten million times that, and the whole discipline of caching exists in
// the gap between those two figures. The catalog can show ONE latency well
// (`latency`) and a set of figures well (`metric`), but it has no picture of
// the hierarchy itself — the thing the numbers belong to, drawn as levels a
// request falls through.
//
// The picture is a ladder because the relationship is ordered: each rung is
// slower and bigger than the one above it, and a miss on any rung means
// falling to the next. Two properties make the drawing honest, and both are
// enforced in Go rather than trusted to the model.
//
// **The latencies must be strictly increasing down the ladder.** A hierarchy
// that gets faster as it gets farther away is drawn wrong — the ordering IS
// the subject, and a single inverted pair quietly teaches that caches are
// arbitrary rather than layered. The validator quotes the two out-of-order
// rungs by name and number, because "not sorted" tells a model nothing about
// which figure it misremembered.
//
// **The bars are log-scale, and the scale is computed here.** On a linear
// axis a 10ms disk next to a 1ns register is one visible bar and seven
// invisible ones, which is a picture of nothing. Each rung's position on the
// log axis is precomputed in Go and shipped as a 0-1 figure, so the renderer
// never re-derives arithmetic the validator has already vouched for — the
// same reason the top-to-bottom ratio ("×200,000") is computed and formatted
// here rather than in the component.

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "ladder",
		Category:    CatNumbers,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The memory ladder",
		Description: "The memory hierarchy as rungs — registers to cache to RAM to disk to network — with log-scale latency bars, capacity chips, and a request falling on a miss. Reach for it when the subject is why locality matters, what a cache miss costs, or how far away the data really is.",
		Example:     "Why a cache hit costs 1ns and a disk read costs 10ms",
		PromptFile:  snippetLadderTemplateName,
		NeedsCode:   false,
		// The whole ladder, a rung or two in focus, a miss, and the spread:
		// four to eight shots, so the clip needs room to cut.
		MinTargetSec:     35,
		DefaultTargetSec: 55,
		MaxBeats:         8,
		// A beat here is a SHOT — one rung, one fall — not an argument step.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Ladder: true},
		OwnsPlan:          planFields{Ladder: true},
		Normalize:         normalizeLadderPlan,
		Validate:          validateLadderPlan,
		Scenes:            ladderScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(LadderShows(), ", "),
				"MinRungs":      minLadderRungs,
				"MaxRungs":      maxLadderRungs,
				"MaxLabelWords": maxLadderLabelWords,
				"MaxCapWords":   maxLadderCapWords,
			}
		},
	})
}

const snippetLadderTemplateName = "snippet_ladder.tmpl"

const (
	// Below four rungs there is no hierarchy, only a comparison — `latency`
	// and `metric` already draw that. Past eight the rows are too shallow to
	// carry a latency figure and a capacity chip each.
	minLadderRungs = 4
	maxLadderRungs = 8

	// A rung's label is a name ("L1 cache"), not a description.
	maxLadderLabelWords = 3
	// A capacity is a figure and a unit ("32 KB"), free text so "a few MB"
	// and "effectively unlimited" both fit.
	maxLadderCapWords = 4
)

// ladderShows is the closed vocabulary of what a beat does.
var ladderShows = map[string]bool{
	// The whole ladder, every rung in place. The first beat.
	"ladder": true,
	// One rung in focus: its bar brightens, its latency and capacity read.
	"rung": true,
	// The request dot drops from rung At to rung At+1, with the cost stated.
	"miss": true,
	// The closer: the whole ladder with the top-to-bottom ratio stated.
	"spread": true,
}

// LadderShows returns the beat vocabulary sorted.
func LadderShows() []string {
	out := make([]string, 0, len(ladderShows))
	for k := range ladderShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// LadderSpec is the hierarchy: rungs from fastest to slowest.
type LadderSpec struct {
	// Rungs are the levels, top (fastest) first.
	Rungs []LadderRung `json:"rungs"`
}

// LadderRung is one level of the hierarchy.
type LadderRung struct {
	// Label names the level — "registers", "L1 cache".
	Label string `json:"label"`
	// LatencyNs is the access cost in nanoseconds. Must be positive, and
	// larger than the rung above.
	LatencyNs float64 `json:"latencyNs"`
	// Capacity is the chip at the right — "32 KB", "terabytes". Free text.
	Capacity string `json:"capacity,omitempty"`
}

// LadderBeat is one shot of the ladder.
type LadderBeat struct {
	// Show is a ladderShows name.
	Show string `json:"show"`
	// At indexes LadderSpec.Rungs, for "rung" (the focus) and "miss" (the
	// rung the request falls FROM).
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults the unknown to a rung in focus, which is what most
// beats of this template are.
func (b LadderBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if ladderShows[s] {
		return s
	}
	return "rung"
}

func normalizeLadderPlan(p *SnippetPlan) {
	l := p.Ladder
	if l == nil {
		return
	}
	rungs := make([]LadderRung, 0, len(l.Rungs))
	for _, r := range l.Rungs {
		r.Label = clampWords(collapseSpaces(r.Label), maxLadderLabelWords)
		r.Capacity = clampWords(collapseSpaces(r.Capacity), maxLadderCapWords)
		if len(rungs) < maxLadderRungs {
			rungs = append(rungs, r)
		}
	}
	l.Rungs = rungs

	for i := range p.Beats {
		b := p.Beats[i].Ladder
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		switch b.Show {
		case "rung":
			if b.At < 0 {
				b.At = 0
			}
			if n := len(l.Rungs); n > 0 && b.At >= n {
				b.At = n - 1
			}
		case "miss":
			// A miss falls from At to At+1, so the deepest legal At is the
			// second-to-last rung.
			if b.At < 0 {
				b.At = 0
			}
			if n := len(l.Rungs); n > 1 && b.At >= n-1 {
				b.At = n - 2
			}
		default:
			b.At = 0
		}
	}
}

func validateLadderPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Ladder: true}); err != nil {
		return err
	}

	l := p.Ladder
	if l == nil {
		return fmt.Errorf("the plan has no ladder — this template is the memory hierarchy drawn as rungs, so the rungs are the clip")
	}
	if n := len(l.Rungs); n < minLadderRungs || n > maxLadderRungs {
		return fmt.Errorf("the ladder has %d rungs, want %d-%d. Below four there is no hierarchy, only a comparison — latency or metric draws that better; past eight the rows are too shallow to carry a latency and a capacity each",
			n, minLadderRungs, maxLadderRungs)
	}
	for i, r := range l.Rungs {
		if strings.TrimSpace(r.Label) == "" {
			return fmt.Errorf("rung %d has no label. A level of the hierarchy the viewer cannot name is a bar with no meaning", i)
		}
		if r.LatencyNs <= 0 {
			return fmt.Errorf("rung %d (%q) has latency %v ns, which is not a positive number. Every level costs something to reach, and a zero or negative latency has no position on a log axis",
				i, r.Label, r.LatencyNs)
		}
	}
	// The ordering IS the subject. The two offenders are quoted by name and
	// number, because "not sorted" does not tell a model which figure it
	// misremembered.
	for i := 1; i < len(l.Rungs); i++ {
		prev, cur := l.Rungs[i-1], l.Rungs[i]
		if cur.LatencyNs <= prev.LatencyNs {
			return fmt.Errorf("rung %d (%q, %s) is not slower than the rung above it, %d (%q, %s). The latencies must strictly increase down the ladder — a hierarchy that gets faster as it gets farther away is drawn wrong, and the ordering is the thing this picture teaches",
				i, cur.Label, ladderNsLabel(cur.LatencyNs), i-1, prev.Label, ladderNsLabel(prev.LatencyNs))
		}
	}

	if p.Beats[0].Ladder == nil || p.Beats[0].Ladder.ResolvedShow() != "ladder" {
		return fmt.Errorf("beat %q does not open on the whole ladder. A rung brightening on a hierarchy the viewer has not seen whole is a bar with no context",
			p.Beats[0].ID)
	}
	last := p.Beats[len(p.Beats)-1]
	if last.Ladder == nil || last.Ladder.ResolvedShow() != "spread" {
		return fmt.Errorf("beat %q does not close on the spread. The clip's last word is the top-to-bottom ratio — the whole ladder with the multiplier stated — and ending anywhere else leaves the numbers uncompared",
			last.ID)
	}
	for _, b := range p.Beats {
		if b.Ladder == nil {
			return fmt.Errorf("beat %q has no ladder direction — every beat shows the ladder, focuses a rung, drops on a miss, or states the spread", b.ID)
		}
		switch b.Ladder.ResolvedShow() {
		case "rung":
			if b.Ladder.At < 0 || b.Ladder.At >= len(l.Rungs) {
				return fmt.Errorf("beat %q focuses rung %d, which does not exist — the ladder has rungs 0-%d", b.ID, b.Ladder.At, len(l.Rungs)-1)
			}
		case "miss":
			if b.Ladder.At < 0 || b.Ladder.At >= len(l.Rungs)-1 {
				return fmt.Errorf("beat %q misses at rung %d, but a miss drops the request to rung %d and the ladder has rungs 0-%d. A miss on the bottom rung has nowhere to fall — the last level is where the data finally is",
					b.ID, b.Ladder.At, b.Ladder.At+1, len(l.Rungs)-1)
			}
		}
	}
	return nil
}

// ladderScenes lays the clip out as ONE scene. The log positions, the axis
// ticks, and every formatted figure are computed here, so the renderer draws
// arithmetic the validator has already vouched for.
func ladderScenes(in SnippetSceneInput) ([]Scene, error) {
	l := in.Plan.Ladder
	if l == nil {
		return nil, fmt.Errorf("the plan has no ladder")
	}
	if len(l.Rungs) < 2 {
		return nil, fmt.Errorf("the ladder has %d rungs, which is not a hierarchy", len(l.Rungs))
	}

	lo := math.Log10(l.Rungs[0].LatencyNs)
	hi := math.Log10(l.Rungs[len(l.Rungs)-1].LatencyNs)
	span := hi - lo
	if span <= 0 {
		return nil, fmt.Errorf("the ladder's latencies do not increase, so it has no log axis")
	}

	rungs := make([]map[string]any, len(l.Rungs))
	for i, r := range l.Rungs {
		rungs[i] = map[string]any{
			"label":     r.Label,
			"capacity":  r.Capacity,
			"latencyNs": r.LatencyNs,
			"latency":   ladderNsLabel(r.LatencyNs),
			// The rung's position on the log axis, 0 at the top rung and 1 at
			// the bottom — precomputed so the component never does the log.
			//
			// Rounded, and not merely for tidiness. Logarithms do not come back
			// exactly: the top rung's own position works out as
			// (log(x) - log(x)) / span and still lands at -1.8e-18 rather than
			// zero, because the two logs are evaluated by different paths. That
			// is invisible in the picture and very visible in the scene JSON,
			// where it puts seventeen digits of float noise into every golden
			// file and makes the first and last rung miss the ends of their own
			// axis. Six places is finer than a pixel at any frame size.
			"logPos": roundTo((math.Log10(r.LatencyNs)-lo)/span, 6),
		}
	}

	// Decade ticks: one per power of ten the ladder crosses, positioned on
	// the same 0-1 axis the bars use.
	ticks := []map[string]any{}
	for d := math.Ceil(lo); d <= math.Floor(hi); d++ {
		pos := (d - lo) / span
		if pos < 0 || pos > 1 {
			continue
		}
		ticks = append(ticks, map[string]any{
			"pos":   roundTo(pos, 6),
			"label": ladderNsLabel(math.Pow(10, d)),
		})
	}

	ratio := ladderTimesLabel(l.Rungs[len(l.Rungs)-1].LatencyNs / l.Rungs[0].LatencyNs)

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Ladder == nil {
			return nil, fmt.Errorf("beat %q has no ladder direction", beat.ID)
		}
		show := beat.Ladder.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "rung":
			step["at"] = beat.Ladder.At
		case "miss":
			at := beat.Ladder.At
			if at < 0 || at+1 >= len(l.Rungs) {
				return nil, fmt.Errorf("beat %q misses at rung %d, which has no rung below it", beat.ID, at)
			}
			step["at"] = at
			step["to"] = at + 1
			// What the fall costs, computed here: the multiplier between the
			// rung missed and the rung the request lands on.
			step["cost"] = ladderTimesLabel(l.Rungs[at+1].LatencyNs/l.Rungs[at].LatencyNs) + " slower"
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneLadder,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title": in.Plan.Title,
			"rungs": rungs,
			"ticks": ticks,
			"ratio": ratio,
			"steps": steps,
		}),
	}}, nil
}

// ladderNsLabel formats a latency for the frame: the unit scales so "10 ms"
// never appears as "10000000 ns".
func ladderNsLabel(ns float64) string {
	v, unit := ns, "ns"
	switch {
	case ns >= 1e9:
		v, unit = ns/1e9, "s"
	case ns >= 1e6:
		v, unit = ns/1e6, "ms"
	case ns >= 1e3:
		v, unit = ns/1e3, "µs"
	}
	v = math.Round(v*100) / 100
	return strconv.FormatFloat(v, 'f', -1, 64) + " " + unit
}

// ladderTimesLabel formats a multiplier — "×200,000" — with the grouping that
// makes six zeros readable at a glance.
func ladderTimesLabel(r float64) string {
	if r < 10 {
		v := math.Round(r*10) / 10
		return "×" + strconv.FormatFloat(v, 'f', -1, 64)
	}
	s := strconv.FormatInt(int64(math.Round(r)), 10)
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteByte(s[i])
	}
	return "×" + b.String()
}
