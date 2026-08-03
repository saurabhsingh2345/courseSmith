package pipeline

// The cycle template: the loop, and what is different next time round.
//
// Half of everything worth teaching is a loop — write, run, read the error,
// fix; plan, build, measure, learn; request, queue, work, respond. The catalog
// could draw all of them badly. `flow` is a DAG and its validator *requires* a
// fork or a join, so a ring is not expressible in it at all; drawn as a chain
// it becomes a list of four steps, which states the one thing that is not true
// about a loop, namely that it ends.
//
// So the frame is a closed ring with the stages on it and a light running
// round. Nothing else. The picture asserts exactly one thing — this comes back
// — and the whole design is in service of making the return visible rather than
// implied by an arrow nobody follows.
//
// The rule that earns it is `changes`, and it is the difference between a cycle
// and a wheel. A ring whose second lap is identical to its first is a machine
// idling: true of a diagram, useless as teaching, and the thing every badly
// drawn feedback loop in every deck actually shows. A loop is worth drawing
// when each pass leaves something behind — a better draft, a smaller error, one
// more test — so the template refuses to be planned without naming what that
// is, and spends its last beat on it.
//
// The stages are also walked in ring order, from the top, once each. That is
// not tidiness: the claim a ring makes is that position means sequence, and a
// narration that jumps from stage one to stage three has drawn a picture that
// contradicts itself.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "cycle",
		Category:    CatSystems,
		Since:       SinceV2,
		Title:       "It comes back round",
		Description: "A closed ring of stages with a light running round it, and a last beat on what is different next lap. Reach for it for anything iterative — a feedback loop, a workflow, a practice routine — where the point is that it repeats and improves.",
		Example:     "The debugging loop: reproduce, isolate, fix, verify — and what you learn each time round",
		PromptFile:  snippetCycleTemplateName,
		NeedsCode:   false,
		// The ring, one beat per stage, and the return. Three stages is five
		// beats before anything optional, which the shared 45s budget cannot
		// fund at a substantial beat.
		MinTargetSec:     45,
		DefaultTargetSec: 65,
		MaxBeats:         8,
		Owns:             beatFields{Cycle: true},
		OwnsPlan:         planFields{Cycle: true},
		Normalize:        normalizeCyclePlan,
		Validate:         validateCyclePlan,
		Scenes:           cycleScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":           strings.Join(CycleShows(), ", "),
				"Icons":           strings.Join(PointIconNames(), ", "),
				"MinStages":       minCycleStages,
				"MaxStages":       maxCycleStages,
				"MaxNameWords":    maxCycleNameWords,
				"MaxLabelWords":   maxCycleLabelWords,
				"MaxNoteWords":    maxCycleNoteWords,
				"MaxChangesWords": maxCycleChangesWords,
			}
		},
	})
}

const snippetCycleTemplateName = "snippet_cycle.tmpl"

const (
	// Two stages is a pendulum and reads as one arrow drawn twice. Past six the
	// labels on a ring start colliding at 1080p, and the stages stop being
	// distinguishable at a glance — which is the only reason to use a ring
	// rather than a list.
	minCycleStages = 3
	maxCycleStages = 6

	maxCycleNameWords  = 5
	maxCycleLabelWords = 4
	maxCycleNoteWords  = 16
	// The one line that says why the loop is worth running twice. It sits in
	// the hub, inside a circle, so it is short by geometry as well as by rule.
	maxCycleChangesWords = 10
)

// cycleShows is the closed vocabulary of what a beat does.
var cycleShows = map[string]bool{
	// The whole ring, unlit, with the light waiting at the top. The first beat.
	"ring": true,
	// Run to one stage: the arc draws and the stage lights.
	"stage": true,
	// The return. The light crosses the last arc back to the start, the ring
	// completes, and the hub says what is different this time. The last beat.
	"again": true,
}

// CycleShows returns the beat vocabulary sorted.
func CycleShows() []string {
	out := make([]string, 0, len(cycleShows))
	for k := range cycleShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// CycleSpec is the loop.
type CycleSpec struct {
	// Name is what the loop is called — "The debugging loop", "Deploy".
	Name string `json:"name"`
	// Changes is what is different on the next lap. Required, and the reason
	// this template exists: a ring whose second pass is identical to its first
	// is a wheel, and drawing one teaches nothing.
	Changes string `json:"changes"`
	// Stages are the steps, in the order the ring is walked.
	Stages []CycleStage `json:"stages"`
}

// CycleStage is one step on the ring.
type CycleStage struct {
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
	// Note is the line that arrives when the light reaches this stage.
	Note string `json:"note,omitempty"`
}

// CycleBeat is one move.
type CycleBeat struct {
	Show string `json:"show"`
	// At indexes the stage a `stage` beat runs to.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to running to
// a stage — which is what most beats of this template do.
func (b CycleBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if cycleShows[s] {
		return s
	}
	return "stage"
}

// Angles returns the compass angle of every stage, in degrees, with the first
// at the top and the rest running clockwise.
//
// Computed in Go for the same reason the constellation's are: a ring's geometry
// is a property of the plan, so the same set of stages lands in the same places
// on every render rather than wherever a layout pass happens to settle them.
// The renderer owns placement, not the map.
func (c CycleSpec) Angles() []float64 {
	n := len(c.Stages)
	out := make([]float64, n)
	for i := range out {
		out[i] = -90 + float64(i)*360/float64(n)
	}
	return out
}

func normalizeCyclePlan(p *SnippetPlan) {
	c := p.Cycle
	if c == nil {
		return
	}
	c.Name = clampWords(collapseSpaces(c.Name), maxCycleNameWords)
	c.Changes = clampWords(collapseSpaces(c.Changes), maxCycleChangesWords)

	stages := make([]CycleStage, 0, len(c.Stages))
	for _, s := range c.Stages {
		s.Label = clampWords(collapseSpaces(s.Label), maxCycleLabelWords)
		s.Note = clampWords(collapseSpaces(s.Note), maxCycleNoteWords)
		s.Icon = normalizePointIconName(s.Icon)
		if s.Label != "" && len(stages) < maxCycleStages {
			stages = append(stages, s)
		}
	}
	c.Stages = stages

	for i := range p.Beats {
		b := p.Beats[i].Cycle
		if b == nil {
			continue
		}
		if !cycleShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			if i == 0 {
				b.Show = "ring"
			} else if i == len(p.Beats)-1 {
				b.Show = "again"
			} else {
				b.Show = "stage"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "stage" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(c.Stages); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateCyclePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Cycle: true}); err != nil {
		return err
	}

	c := p.Cycle
	if c == nil {
		return fmt.Errorf("the plan has no cycle — this template is a closed ring of stages and the thing that improves each lap")
	}
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("the loop has no name — it sits in the middle of the ring, and a ring with an empty hub is a diagram of nothing in particular")
	}
	// The rule this template exists for.
	if strings.TrimSpace(c.Changes) == "" {
		return fmt.Errorf("the cycle never says what is different next time round. That is the whole reason to draw a ring rather than a list: a loop whose second pass is identical to its first is a wheel spinning, and nobody learns anything from watching one. Name what each lap leaves behind — a better draft, a smaller error, one more passing test")
	}
	if n := len(c.Stages); n < minCycleStages || n > maxCycleStages {
		return fmt.Errorf("there are %d stages, want %d-%d. Two is a pendulum and reads as one arrow drawn twice; past %d the labels collide on the ring and the stages stop being tellable apart, which is the only reason to draw a ring instead of a list",
			n, minCycleStages, maxCycleStages, maxCycleStages)
	}

	seen := map[string]bool{}
	for i, s := range c.Stages {
		if strings.TrimSpace(s.Label) == "" {
			return fmt.Errorf("stage %d has no label", i)
		}
		key := strings.ToLower(strings.TrimSpace(s.Label))
		if seen[key] {
			return fmt.Errorf("two stages are both called %q — on a ring that reads as the light passing the same place twice", s.Label)
		}
		seen[key] = true
	}

	counts := map[string]int{}
	walked := 0
	for i, b := range p.Beats {
		if b.Cycle == nil {
			return fmt.Errorf("beat %q has no cycle direction — every beat draws the ring, runs to a stage, or comes back round", b.ID)
		}
		show := b.Cycle.ResolvedShow()
		counts[show]++

		if i == 0 && show != "ring" {
			return fmt.Errorf("the clip opens on %q. Draw the ring first — a stage lighting up on a ring nobody has seen yet is a dot with a word next to it", show)
		}
		if show == "again" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q comes back round but the clip carries on afterwards. The return is the last frame — it is the moment the picture pays off, and anything after it is watched over the top of a completed ring", b.ID)
		}
		if show != "stage" {
			continue
		}
		// Position means sequence on a ring. A narration that jumps stages has
		// drawn a picture that contradicts itself.
		if b.Cycle.At != walked {
			return fmt.Errorf("beat %q runs to stage %d (%q) when the light is at stage %d. The ring is walked in order and once each — on a ring, where a stage SITS is the claim about when it happens, so skipping one draws a picture that disagrees with the words over it",
				b.ID, b.Cycle.At, c.Stages[min(b.Cycle.At, len(c.Stages)-1)].Label, walked)
		}
		walked++
	}

	if counts["ring"] != 1 {
		return fmt.Errorf("there are %d ring beats; the ring is drawn once and then stays", counts["ring"])
	}
	if counts["again"] != 1 {
		return fmt.Errorf("there are %d return beats, want exactly one — and it is the last. Without it the clip stops at the final stage, which is the one shape a cycle must never have", counts["again"])
	}
	if walked != len(c.Stages) {
		return fmt.Errorf("%d of the %d stages are never run. A stage on the ring that the light never reaches is a step the viewer is left to guess at — give it a beat or take it off the ring",
			len(c.Stages)-walked, len(c.Stages))
	}
	return nil
}

// cycleScenes lays the clip out as ONE scene: the ring is standing from the
// first frame and the light moves round it.
func cycleScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.Cycle
	if c == nil {
		return nil, fmt.Errorf("the plan has no cycle")
	}
	if len(c.Stages) == 0 {
		return nil, fmt.Errorf("the cycle has no stages")
	}

	angles := c.Angles()
	stages := make([]map[string]any, len(c.Stages))
	for i, s := range c.Stages {
		stages[i] = map[string]any{
			"label": s.Label,
			"icon":  s.Icon,
			"note":  s.Note,
			"angle": angles[i],
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Cycle == nil {
			return nil, fmt.Errorf("beat %q has no cycle direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Cycle.ResolvedShow(),
		}
		if beat.Cycle.ResolvedShow() == "stage" {
			step["at"] = beat.Cycle.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneCycle,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":   in.Plan.Title,
			"name":    c.Name,
			"changes": c.Changes,
			"stages":  stages,
			"steps":   steps,
		},
	}}, nil
}
