package pipeline

// The trace template: a system caught in the act.
//
// `flow` draws a system's shape — boxes in ranks with traffic moving along the
// edges — and that answers "what is connected to what". It cannot answer "what
// happens when two of these arrive at once", which is where almost every real
// bug and every interesting design decision lives. Concurrency, atomicity,
// queueing, contention: none of them are visible in a topology, because they
// are properties of *time*.
//
// So this template draws state instead of structure. Actors put operations into
// a queue, the queue drains one at a time, and a piece of shared state changes
// value as each one lands. The frame the whole thing exists for is the one
// where two operations are pending against the same value and the viewer can
// see, before the narrator says it, why the order matters.
//
// The rule that earns it its place is that the state has to add up. Each step
// declares what the shared value becomes, and a clip whose value jumps to
// something the operations could not have produced is a clip teaching a wrong
// model of the very thing it is about. That is checked: the declared values
// must form a chain where every change is attributable to the step that made
// it, and a step that changes nothing must say so rather than silently holding.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "trace",
		Category:    CatSystems,
		Since:       SinceV1,
		Title:       "Watch it run",
		Description: "Actors, a queue and one shared value — drained a step at a time, so the order becomes visible.",
		Example:     "Why two users buying the last item at once oversells your inventory",
		PromptFile:  snippetTraceTemplateName,
		NeedsCode:   false,
		// Setting the actors up, showing the contention and draining three
		// operations is five beats before the payoff.
		MinTargetSec:     45,
		DefaultTargetSec: 65,
		MaxBeats:         9,
		Owns:             beatFields{Trace: true},
		OwnsPlan:         planFields{Trace: true},
		Normalize:        normalizeTracePlan,
		Validate:         validateTracePlan,
		Scenes:           traceScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":         strings.Join(TraceShows(), ", "),
				"MinActors":     minTraceActors,
				"MaxActors":     maxTraceActors,
				"MinSteps":      minTraceSteps,
				"MaxSteps":      maxTraceSteps,
				"MaxNameWords":  maxTraceNameWords,
				"MaxOpWords":    maxTraceOpWords,
				"MaxValueChars": maxTraceValueChars,
				"MaxNoteWords":  maxTraceNoteWords,
			}
		},
	})
}

const snippetTraceTemplateName = "snippet_trace.tmpl"

const (
	// One actor cannot contend with anything. Four columns of actors across the
	// stage leaves each too narrow to name.
	minTraceActors = 2
	maxTraceActors = 3

	// Two operations is the smallest race worth drawing; six is a log.
	minTraceSteps = 3
	maxTraceSteps = 6

	maxTraceNameWords  = 3
	maxTraceOpWords    = 4
	maxTraceValueChars = 12
	maxTraceNoteWords  = 14
)

// traceShows is the closed vocabulary of what a beat does.
var traceShows = map[string]bool{
	// The actors and the shared state, at rest, before anything is sent.
	"setup": true,
	// Everything queued at once, nothing drained. The contention frame.
	"queue": true,
	// Drain one operation: the state takes its new value.
	"step": true,
	// What went wrong, or what the order proved. The closing frame.
	"outcome": true,
}

// TraceShows returns the beat vocabulary sorted.
func TraceShows() []string {
	out := make([]string, 0, len(traceShows))
	for k := range traceShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TraceSpec is the system being run.
type TraceSpec struct {
	// Actors are who is sending work — "User A", "Thread 1".
	Actors []string `json:"actors"`
	// Resource names the shared thing everything contends for — "Inventory".
	Resource string `json:"resource"`
	// Start is the shared value before anything runs.
	Start string `json:"start"`
	// Steps are the operations, in the order they actually execute.
	Steps []TraceStepSpec `json:"steps"`
	// Outcome is what the run proved. It takes the closing frame.
	Outcome string `json:"outcome"`
	// Broken says whether the outcome is the bug rather than the happy path.
	// It picks the colour of the closing frame, so it is a claim about what the
	// clip just showed rather than a styling choice.
	Broken bool `json:"broken,omitempty"`
}

// TraceStepSpec is one operation draining out of the queue.
type TraceStepSpec struct {
	// By indexes TraceSpec.Actors — who issued this operation.
	By int `json:"by"`
	// Op is the operation as it would be written — "DECR inv", "read balance".
	Op string `json:"op"`
	// Becomes is the shared value AFTER this operation lands. A step that does
	// not change the value repeats the previous one, which is a claim worth
	// making explicitly: "this read changed nothing" is usually the whole
	// point of the bug.
	Becomes string `json:"becomes"`
	// Note is the line that arrives with this step.
	Note string `json:"note,omitempty"`
}

// TraceBeat is one move.
type TraceBeat struct {
	Show string `json:"show"`
	At   int    `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to draining a
// step — the bulk of the clip.
func (b TraceBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if traceShows[s] {
		return s
	}
	return "step"
}

func normalizeTracePlan(p *SnippetPlan) {
	tr := p.Trace
	if tr == nil {
		return
	}
	tr.Resource = clampWords(collapseSpaces(tr.Resource), maxTraceNameWords)
	tr.Start = clampChars(collapseSpaces(tr.Start), maxTraceValueChars)
	tr.Outcome = clampWords(collapseSpaces(tr.Outcome), maxTraceNoteWords)

	actors := make([]string, 0, len(tr.Actors))
	for _, a := range tr.Actors {
		a = clampWords(collapseSpaces(a), maxTraceNameWords)
		if a != "" && len(actors) < maxTraceActors {
			actors = append(actors, a)
		}
	}
	tr.Actors = actors

	steps := make([]TraceStepSpec, 0, len(tr.Steps))
	for _, s := range tr.Steps {
		s.Op = clampWords(collapseSpaces(s.Op), maxTraceOpWords)
		s.Becomes = clampChars(collapseSpaces(s.Becomes), maxTraceValueChars)
		s.Note = clampWords(collapseSpaces(s.Note), maxTraceNoteWords)
		if s.By < 0 {
			s.By = 0
		}
		if n := len(tr.Actors); n > 0 && s.By >= n {
			s.By = n - 1
		}
		// An operation with no name is not an operation. Dropping it is the
		// repair; the state chain is re-checked by the validator afterwards.
		if s.Op != "" && len(steps) < maxTraceSteps {
			steps = append(steps, s)
		}
	}
	tr.Steps = steps

	for i := range p.Beats {
		b := p.Beats[i].Trace
		if b == nil {
			continue
		}
		if !traceShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			switch {
			case i == 0:
				b.Show = "setup"
			case i == len(p.Beats)-1:
				b.Show = "outcome"
			default:
				b.Show = "step"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "step" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(tr.Steps); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateTracePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Trace: true}); err != nil {
		return err
	}

	tr := p.Trace
	if tr == nil {
		return fmt.Errorf("the plan has no trace — this template is a system caught in the act, with one shared value changing as work drains")
	}
	if n := len(tr.Actors); n < minTraceActors || n > maxTraceActors {
		return fmt.Errorf("there are %d actors, want %d-%d. One actor cannot contend with anything, and contention is the only reason to draw a system running rather than its shape",
			n, minTraceActors, maxTraceActors)
	}
	if strings.TrimSpace(tr.Resource) == "" {
		return fmt.Errorf("nothing is named as the shared resource — say what everyone is contending for")
	}
	if strings.TrimSpace(tr.Start) == "" {
		return fmt.Errorf("the shared value has no starting state. Without it nothing that follows is a change")
	}
	if n := len(tr.Steps); n < minTraceSteps || n > maxTraceSteps {
		return fmt.Errorf("there are %d operations, want %d-%d. Two is not yet a race, and seven is a log rather than a diagram",
			n, minTraceSteps, maxTraceSteps)
	}
	if strings.TrimSpace(tr.Outcome) == "" {
		return fmt.Errorf("the run has no outcome. Say what the order proved — that line is the reason the clip exists")
	}

	// The rule this template exists for: the state has to add up.
	usedActors := map[int]bool{}
	for i, s := range tr.Steps {
		if strings.TrimSpace(s.Op) == "" {
			return fmt.Errorf("operation %d has no name", i)
		}
		if s.By < 0 || s.By >= len(tr.Actors) {
			return fmt.Errorf("operation %q is issued by actor %d, who does not exist", s.Op, s.By)
		}
		usedActors[s.By] = true
		if strings.TrimSpace(s.Becomes) == "" {
			return fmt.Errorf("operation %q does not say what %s becomes. Every step states the value afterwards — a step that changes nothing repeats the previous value, and saying so explicitly is usually the whole point of the bug",
				s.Op, tr.Resource)
		}
	}
	// Every actor has to actually do something. An actor drawn but never
	// issuing an operation is a column of empty space that implies a
	// participant who never participates.
	if len(usedActors) != len(tr.Actors) {
		for i, a := range tr.Actors {
			if !usedActors[i] {
				return fmt.Errorf("%q never issues an operation. An actor who does nothing is a column of empty space implying a participant that is not in the story — give them a step or cut them", a)
			}
		}
	}
	// A trace whose value never changes at all is a sequence diagram, not a
	// trace: nothing is contended for and the template is the wrong one.
	changed := false
	last := tr.Start
	for _, s := range tr.Steps {
		if s.Becomes != last {
			changed = true
		}
		last = s.Becomes
	}
	if !changed {
		return fmt.Errorf("%s is %q from beginning to end, so nothing is actually contended for. If the point is which components talk to each other rather than how their order interacts, that is the flow template",
			tr.Resource, tr.Start)
	}

	drained := map[int]bool{}
	counts := map[string]int{}
	for i, b := range p.Beats {
		if b.Trace == nil {
			return fmt.Errorf("beat %q has no trace direction — every beat sets the system up, queues the work, drains one operation, or delivers the outcome", b.ID)
		}
		show := b.Trace.ResolvedShow()
		counts[show]++
		if i == 0 && show != "setup" {
			return fmt.Errorf("the clip opens on %q. Show the actors and the value at rest first — an operation landing on a value the viewer has not seen is a change from nothing", show)
		}
		if show == "outcome" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q delivers the outcome but the clip carries on afterwards. The outcome is the closing frame", b.ID)
		}
		if show != "step" {
			continue
		}
		if b.Trace.At < 0 || b.Trace.At >= len(tr.Steps) {
			return fmt.Errorf("beat %q drains operation %d, which does not exist", b.ID, b.Trace.At)
		}
		if drained[b.Trace.At] {
			return fmt.Errorf("beat %q drains operation %d again; each one lands once", b.ID, b.Trace.At)
		}
		drained[b.Trace.At] = true
	}
	if counts["setup"] != 1 {
		return fmt.Errorf("there are %d setup beats; the system is introduced once", counts["setup"])
	}
	if counts["outcome"] != 1 {
		return fmt.Errorf("there are %d outcome beats; the run resolves exactly once", counts["outcome"])
	}
	if len(drained) != len(tr.Steps) {
		return fmt.Errorf("%d of the %d operations never get a beat. An operation that lands without narration is a state change the viewer sees but cannot account for, which is the exact confusion this template exists to remove",
			len(tr.Steps)-len(drained), len(tr.Steps))
	}
	return nil
}

// traceScenes lays the clip out as ONE scene: the actors, the queue and the
// shared value are the frame throughout, and the beats drain it.
func traceScenes(in SnippetSceneInput) ([]Scene, error) {
	tr := in.Plan.Trace
	if tr == nil {
		return nil, fmt.Errorf("the plan has no trace")
	}

	steps := make([]map[string]any, len(tr.Steps))
	prev := tr.Start
	for i, s := range tr.Steps {
		steps[i] = map[string]any{
			"by":      s.By,
			"op":      s.Op,
			"becomes": s.Becomes,
			"note":    s.Note,
			// Whether this operation actually moved the value is decided here,
			// so the renderer can mark a no-op without re-deriving the chain —
			// and so "this read changed nothing" is a fact in the scene graph
			// rather than a coincidence of two strings matching in a component.
			"changes": s.Becomes != prev,
		}
		prev = s.Becomes
	}

	beatSteps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Trace == nil {
			return nil, fmt.Errorf("beat %q has no trace direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Trace.ResolvedShow(),
		}
		if beat.Trace.ResolvedShow() == "step" {
			step["at"] = beat.Trace.At
		}
		beatSteps = append(beatSteps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneTrace,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":    in.Plan.Title,
			"actors":   tr.Actors,
			"resource": tr.Resource,
			"start":    tr.Start,
			"ops":      steps,
			"outcome":  tr.Outcome,
			"broken":   tr.Broken,
			"steps":    beatSteps,
		},
	}}, nil
}
