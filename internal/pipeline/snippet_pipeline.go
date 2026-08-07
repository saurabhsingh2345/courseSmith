package pipeline

// The pipeline template: five instructions in flight.
//
// "A five-stage pipeline finishes one instruction every cycle" is the sentence
// every architecture course says and almost nobody believes on first hearing,
// because it sounds like a claim that instructions got five times faster. They
// did not. Each one still takes five cycles end to end. What changed is that
// the five stages are busy with five DIFFERENT instructions at once, and the
// only way to see that is to watch the grid: columns for the stages, chips for
// the instructions, and one tick of the clock moving every chip one column
// right while a new one walks in at the left.
//
// That is why the occupancy is simulated in Go rather than animated in the
// component. A renderer that computes its own grid is a second implementation
// of the machine, drifting quietly from the one the validator checked, and the
// failure mode is a chip in the wrong column on a diagram that claims to teach
// what column a chip is in. So this template walks its own beats, advances the
// stream on every fill, holds on a stall, and ships the FULL grid for every
// tick — which item sits in which stage, -1 for empty. The component draws the
// occupancy it is handed and animates between consecutive grids. It never
// simulates.
//
// Two fill beats is the floor because one tick shows a chip moving, and a chip
// moving is not pipelining: pipelining is the second chip entering while the
// first is still in flight, which does not exist until tick two. The stall gets
// at most one beat, and only when the plan says what causes it, because a
// bubble with no named hazard is a gap the viewer reads as a rendering glitch.
// And the tick count is checked against the length of the run — the stream
// drains after items + stages - 1 ticks, so asking for more fills than that
// spends the end of the clip animating an empty grid.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "pipeline",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "Five instructions in flight",
		Description: "Stage columns with work streaming through them one tick at a time, a bubble where it stalls, and the throughput arithmetic at the end. Reach for it when the lesson is that overlapping stages raise throughput without making any single item faster — CPU pipelines, assembly lines, CI stages.",
		Example:     "Why a 5-stage pipeline finishes one instruction every cycle",
		PromptFile:  snippetPipelineTemplateName,
		NeedsCode:   false,
		// The empty grid, at least two ticks, usually a stall and the payoff:
		// under thirty-five seconds the chips slide faster than the eye tracks
		// them and the grid reads as flicker.
		MinTargetSec:     35,
		DefaultTargetSec: 55,
		// Opener + up to six ticks + closer. Past nine the grid is repeating a
		// motion the viewer understood four ticks ago.
		MaxBeats: 9,
		// A beat here is a SHOT — one tick of the clock — not a step in an
		// argument.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Pipeline: true},
		OwnsPlan:          planFields{Pipeline: true},
		Normalize:         normalizePipelinePlan,
		Validate:          validatePipelinePlan,
		Scenes:            pipelineScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(PipelineShows(), ", "),
				"MinStages":     minPipelineStages,
				"MaxStages":     maxPipelineStages,
				"MinItems":      minPipelineItems,
				"MaxItems":      maxPipelineItems,
				"MaxStageWords": maxPipelineStageWords,
				"MaxItemWords":  maxPipelineItemWords,
				"MaxStallWords": maxPipelineStallWords,
				"MinFills":      minPipelineFills,
			}
		},
	})
}

const snippetPipelineTemplateName = "snippet_pipeline.tmpl"

const (
	// Two columns is a hand-off, not a pipeline: nothing can be in flight
	// behind something else. Three is the least that overlaps.
	minPipelineStages = 3
	// Five columns is the classic RISC pipeline and the widest grid that
	// still leaves a chip room for a readable label.
	maxPipelineStages = 5
	// One item never overlaps with anything, so the picture would show a
	// single chip walking a corridor.
	minPipelineItems = 2
	// Five chips is enough to fill the deepest grid and show one retiring;
	// a sixth only repeats a motion already understood.
	maxPipelineItems = 5
	// A column header is an abbreviation — "IF", "decode", "write back".
	maxPipelineStageWords = 3
	// A chip label is a name — "load", "add r1", "branch".
	maxPipelineItemWords = 3
	// The stall caption sits under the bubble. Ten words names the hazard
	// without becoming a paragraph in the gap.
	maxPipelineStallWords = 10

	// One tick shows a chip moving, and a chip moving is not pipelining.
	// Pipelining is the second chip entering while the first is still in
	// flight, which does not exist until tick two.
	minPipelineFills = 2

	// pipelineEmptyCell is what sits in a stage nothing occupies.
	pipelineEmptyCell = -1
)

// pipelineShows is the closed vocabulary of what a beat does.
var pipelineShows = map[string]bool{
	// The stage columns alone, nothing in flight. The opener.
	"empty": true,
	// One tick of the clock: every item advances one stage, the next enters.
	"fill": true,
	// A bubble: one item holds, a gap opens behind it, the hazard is named.
	"stall": true,
	// Steady state, and the throughput arithmetic. The closer.
	"flow": true,
}

// PipelineShows returns the beat vocabulary sorted.
func PipelineShows() []string {
	out := make([]string, 0, len(pipelineShows))
	for k := range pipelineShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// PipelineSpec is the machine and the work it is given. On the plan because
// the columns stand for the whole clip; the beats only move what is in them.
type PipelineSpec struct {
	// StageNames are the column headers, left to right.
	StageNames []string `json:"stageNames"`
	// Items are the things streaming through, in the order they enter.
	Items []string `json:"items"`
	// Stall names the hazard that opens the bubble. Empty when the clip has
	// no stall beat.
	Stall string `json:"stall,omitempty"`
}

// PipelineBeat is one shot: which state of the grid this beat shows. There is
// no index here — a tick moves everything, so the beat says what the clock
// does and Go works out where every chip lands.
type PipelineBeat struct {
	// Show is a pipelineShows name.
	Show string `json:"show"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a fill —
// the workhorse state most beats of this template are in.
func (b PipelineBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if pipelineShows[s] {
		return s
	}
	return "fill"
}

func normalizePipelinePlan(p *SnippetPlan) {
	pl := p.Pipeline
	if pl == nil {
		return
	}
	names := make([]string, 0, len(pl.StageNames))
	for _, n := range pl.StageNames {
		n = clampWords(collapseSpaces(n), maxPipelineStageWords)
		if n != "" {
			names = append(names, n)
		}
	}
	if len(names) > maxPipelineStages {
		names = names[:maxPipelineStages]
	}
	pl.StageNames = names

	items := make([]string, 0, len(pl.Items))
	for _, it := range pl.Items {
		it = clampWords(collapseSpaces(it), maxPipelineItemWords)
		if it != "" {
			items = append(items, it)
		}
	}
	if len(items) > maxPipelineItems {
		items = items[:maxPipelineItems]
	}
	pl.Items = items

	pl.Stall = clampWords(collapseSpaces(pl.Stall), maxPipelineStallWords)

	for i := range p.Beats {
		if b := p.Beats[i].Pipeline; b != nil {
			b.Show = b.ResolvedShow()
		}
	}
}

func validatePipelinePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Pipeline: true}); err != nil {
		return err
	}

	pl := p.Pipeline
	if pl == nil {
		return fmt.Errorf("the plan has no pipeline — this template is a grid of stage columns with work moving through it, so without stages and items there is no grid")
	}
	if n := len(pl.StageNames); n < minPipelineStages || n > maxPipelineStages {
		return fmt.Errorf("the pipeline has %d stages, want %d-%d. Two columns is a hand-off — nothing can be in flight behind anything else — and past %d the chips lose the width their labels need",
			n, minPipelineStages, maxPipelineStages, maxPipelineStages)
	}
	for i, n := range pl.StageNames {
		if strings.TrimSpace(n) == "" {
			return fmt.Errorf("stage %d has no name. Every column is headed — IF, ID, EX and so on", i+1)
		}
		if w := len(strings.Fields(n)); w > maxPipelineStageWords {
			return fmt.Errorf("the stage name %q is %d words and a column header holds %d. Use the abbreviation the field uses", n, w, maxPipelineStageWords)
		}
	}
	if n := len(pl.Items); n < minPipelineItems || n > maxPipelineItems {
		return fmt.Errorf("the pipeline carries %d items, want %d-%d. One item never overlaps with anything, so the picture would be a single chip walking a corridor, and past %d the chips repeat a motion the viewer already read",
			n, minPipelineItems, maxPipelineItems, maxPipelineItems)
	}
	for i, it := range pl.Items {
		if strings.TrimSpace(it) == "" {
			return fmt.Errorf("item %d has no label. Every chip is named — the name is what makes it the SAME chip two columns later", i+1)
		}
		if w := len(strings.Fields(it)); w > maxPipelineItemWords {
			return fmt.Errorf("the item %q is %d words and a chip holds %d. A chip is a name, not an explanation", it, w, maxPipelineItemWords)
		}
	}
	if w := len(strings.Fields(pl.Stall)); w > maxPipelineStallWords {
		return fmt.Errorf("the stall caption %q is %d words and the gap under the bubble holds %d. Name the hazard, not its remedy", pl.Stall, w, maxPipelineStallWords)
	}

	for _, b := range p.Beats {
		if b.Pipeline == nil {
			return fmt.Errorf("beat %q has no pipeline direction — every beat is one state of the clock, so every beat needs a {\"show\": ...}", b.ID)
		}
	}
	if p.Beats[0].Pipeline.ResolvedShow() != "empty" {
		return fmt.Errorf("beat %q does not open on the empty grid. The columns have to be read before anything moves through them, or the first tick is chips arriving in unlabelled space — open with {\"show\": \"empty\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Pipeline.ResolvedShow() != "flow" {
		return fmt.Errorf("the clip does not close on the flow. The last beat is where the point lands — every stage busy, one item finishing per tick — so end with {\"show\": \"flow\"}")
	}

	fills, stalls := 0, 0
	for i, b := range p.Beats {
		switch b.Pipeline.ResolvedShow() {
		case "empty":
			if i != 0 {
				return fmt.Errorf("beat %q empties the grid part-way through. The stream never un-runs — once work is in flight it stays in flight, so this beat is a fill, a stall or the closer", b.ID)
			}
		case "flow":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q lands the throughput point before the end. \"flow\" is the closer, and after the arithmetic is on screen another tick is just motion", b.ID)
			}
		case "fill":
			fills++
		case "stall":
			if strings.TrimSpace(pl.Stall) == "" {
				return fmt.Errorf("beat %q opens a bubble, and the plan gives no stall text. A gap with no named hazard reads as a rendering glitch — either say what causes the bubble in \"stall\" or drop this beat", b.ID)
			}
			if stalls > 0 {
				return fmt.Errorf("beat %q is a second stall, and the clip may have one. A pipeline that stutters twice is teaching hazards, which is a different clip — this one is teaching throughput, and one bubble is enough to show what a bubble costs", b.ID)
			}
			if fills == 0 {
				return fmt.Errorf("beat %q stalls before anything is in flight. There is nothing in the grid to hold and nothing behind it to open a gap — put at least one fill before the stall", b.ID)
			}
			stalls++
		}
	}
	if fills < minPipelineFills {
		return fmt.Errorf("the clip has %d fill beats and needs at least %d. One tick shows a chip moving, and a chip moving is not pipelining — pipelining is the second chip entering while the first is still in flight, which does not exist until the second tick",
			fills, minPipelineFills)
	}
	// THE ARITHMETIC. The stream drains after items + stages - 1 ticks, plus
	// one for the bubble; ticks beyond that animate an empty grid.
	ticks := fills + stalls
	drain := len(pl.Items) + len(pl.StageNames) - 1 + stalls
	if ticks > drain {
		return fmt.Errorf("the clip runs %d ticks, and %d items through %d stages drains after %d. The last %d tick(s) would animate an emptying grid with the point already made — cut fills, or add items so there is still work entering",
			ticks, len(pl.Items), len(pl.StageNames), drain, ticks-drain)
	}
	return nil
}

// pipelineScenes lays the clip out as ONE scene AND runs the machine.
//
// This is the whole reason the template exists in Go: the occupancy grid for
// every tick is simulated here, so there is exactly one implementation of what
// a pipeline does and the renderer cannot drift from it.
func pipelineScenes(in SnippetSceneInput) ([]Scene, error) {
	pl := in.Plan.Pipeline
	if pl == nil {
		return nil, fmt.Errorf("the plan has no pipeline")
	}
	nStages := len(pl.StageNames)
	if nStages == 0 {
		return nil, fmt.Errorf("the pipeline has no stages")
	}

	stages := make([]map[string]any, nStages)
	for i, n := range pl.StageNames {
		stages[i] = map[string]any{"name": n}
	}
	items := make([]map[string]any, len(pl.Items))
	for i, it := range pl.Items {
		items[i] = map[string]any{"label": it}
	}

	// The machine. occ[s] is the index of the item in stage s, or -1.
	occ := make([]int, nStages)
	for i := range occ {
		occ[i] = pipelineEmptyCell
	}
	next, tick, retired := 0, 0, 0

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Pipeline == nil {
			return nil, fmt.Errorf("beat %q has no pipeline direction", beat.ID)
		}
		show := beat.Pipeline.ResolvedShow()
		bubble := pipelineEmptyCell

		switch show {
		case "fill":
			// The clock ticks: whatever is in the last stage leaves, every
			// other chip moves one column right, and the next item walks in.
			if occ[nStages-1] != pipelineEmptyCell {
				retired++
			}
			for s := nStages - 1; s >= 1; s-- {
				occ[s] = occ[s-1]
			}
			if next < len(pl.Items) {
				occ[0] = next
				next++
			} else {
				occ[0] = pipelineEmptyCell
			}
			tick++
		case "stall":
			// One chip holds and everything behind it holds with it, while the
			// stages ahead drain — which is exactly what opens the gap. The
			// held chip is the one in the second stage, where a real machine
			// notices the hazard; if that stage is empty, the earliest
			// occupied stage that has somewhere to hand on to.
			held := pipelineEmptyCell
			if nStages >= 3 && occ[1] != pipelineEmptyCell {
				held = 1
			} else {
				for s := 0; s < nStages-1; s++ {
					if occ[s] != pipelineEmptyCell {
						held = s
						break
					}
				}
			}
			if held != pipelineEmptyCell {
				if occ[nStages-1] != pipelineEmptyCell {
					retired++
				}
				for s := nStages - 1; s >= held+2; s-- {
					occ[s] = occ[s-1]
				}
				occ[held+1] = pipelineEmptyCell
				bubble = held + 1
			}
			tick++
		}

		grid := make([]int, nStages)
		copy(grid, occ)
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"occ":     grid,
			"bubble":  bubble,
			"tick":    tick,
			"retired": retired,
		}
		// Which items are on screen at this tick, sorted, so the component can
		// dim the ones that have already retired without scanning the grid.
		inFlight := make([]int, 0, nStages)
		for _, v := range grid {
			if v != pipelineEmptyCell {
				inFlight = append(inFlight, v)
			}
		}
		sort.Ints(inFlight)
		step["inFlight"] = inFlight
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    ScenePipeline,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"stages": stages,
			"items":  items,
			"stall":  pl.Stall,
			// The payoff, computed once: the same work done one at a time
			// against the same work overlapped.
			"sequentialTicks": len(pl.Items) * nStages,
			"pipelinedTicks":  nStages + len(pl.Items) - 1,
			"steps":           steps,
		}),
	}}, nil
}
