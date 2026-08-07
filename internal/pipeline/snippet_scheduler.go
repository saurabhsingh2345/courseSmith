package pipeline

// The scheduler template: who gets the CPU, and for how long.
//
// "The operating system shares the processor between programs" is a sentence
// every beginner can repeat and almost none can draw, because the sharing
// happens on an axis nothing in their experience exposes: time. They picture
// programs running side by side, which is the one thing that is not happening
// on a single core. What fixes it is the picture every operating systems
// course eventually draws on a whiteboard and no video ever animates — lanes,
// one per process, and blocks laid along a shared time axis so that at every
// instant exactly one lane is filled.
//
// So the clip lays the schedule down in real time, slot by slot, and the empty
// stretches in the other lanes are as much of the content as the filled ones:
// waiting is what the picture is FOR. The policy is named on screen because
// Round Robin, FCFS and priority produce visibly different pictures out of the
// same processes, and a Gantt chart with no policy on it is a chart of one
// arbitrary outcome.
//
// The validators are arithmetic because the picture is arithmetic. Total time
// is summed in Go and capped, since a timeline past two dozen units draws its
// unit ticks closer together than the eye separates them and the whole reason
// for a tick — that time is discrete here — is lost. Every process must run at
// least once, because a lane that stays empty for the whole clip is a row of
// nothing the viewer spends the whole clip waiting to see used. The slots are
// laid down in order with no skips, since a block appearing to the right of a
// hole says time passed with nobody on the CPU. And a context switch can only
// be zoomed at a boundary where both neighbours are already on screen: the
// cost of a switch is the cost of the changeover between two blocks, and with
// one of them missing there is no changeover to charge for.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "scheduler",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "Who gets the CPU",
		Description: "Processes as lanes and their turns as blocks along one shared time axis, laid down slot by slot with the context switches marked. Reach for it when the subject is time-sharing itself — round robin, first come first served, priority, why a busy machine still feels responsive.",
		Example:     "Round Robin: three processes share one CPU",
		PromptFile:  snippetSchedulerTemplateName,
		NeedsCode:   false,
		// The queue, three or four turns, a switch and the tally: under
		// thirty-five seconds the blocks land faster than the viewer can check
		// which lane they went into.
		MinTargetSec:     35,
		DefaultTargetSec: 55,
		// Opener + up to six turns and switches + closer. Past nine the
		// timeline has been narrated for longer than it would take to run.
		MaxBeats: 9,
		// A beat here is a SHOT — one block landing on one timeline — not a
		// step in an argument. Twenty-eight words is about nine seconds, which
		// is as long as one turn holds anybody.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Scheduler: true},
		OwnsPlan:          planFields{Scheduler: true},
		Normalize:         normalizeSchedulerPlan,
		Validate:          validateSchedulerPlan,
		Scenes:            schedulerScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":          strings.Join(MetricRoles(), ", "),
				"Shows":          strings.Join(SchedulerShows(), ", "),
				"MinProcs":       minSchedulerProcs,
				"MaxProcs":       maxSchedulerProcs,
				"MinSlots":       minSchedulerSlots,
				"MaxSlots":       maxSchedulerSlots,
				"MinSlotLen":     minSchedulerSlotLen,
				"MaxSlotLen":     maxSchedulerSlotLen,
				"MaxUnits":       maxSchedulerUnits,
				"MaxPolicyWords": maxSchedulerPolicyWords,
				"MaxLabelWords":  maxSchedulerLabelWords,
			}
		},
	})
}

const snippetSchedulerTemplateName = "snippet_scheduler.tmpl"

const (
	// One process does not share anything, so two is the floor by definition.
	minSchedulerProcs = 2
	// Past four the lanes are shorter than their own labels once the axis and
	// the totals have taken their share of the height.
	maxSchedulerProcs = 4

	// Two turns is an alternation, not a schedule — the pattern a policy makes
	// only becomes visible on the third.
	minSchedulerSlots = 3
	// Past ten the blocks are narrower than the switch band drawn between them.
	maxSchedulerSlots = 10

	// A zero-length turn is not a turn, and past six units one slot eats a
	// quarter of the axis and the others become slivers.
	minSchedulerSlotLen = 1
	maxSchedulerSlotLen = 6

	// The whole axis. Past twenty-four the unit ticks are closer together than
	// the eye separates them, and a tick nobody can count is not a tick.
	maxSchedulerUnits = 24

	// The policy is a nameplate over the chart: "Round Robin", "FCFS".
	maxSchedulerPolicyWords = 4
	// A lane label sits at the left of its row in a narrow gutter.
	maxSchedulerLabelWords = 2
)

// schedulerShows is the closed vocabulary of what a beat does.
var schedulerShows = map[string]bool{
	// The processes waiting, the timeline empty. The opener.
	"queue": true,
	// Slot At lays down: that process's block extends while the others wait.
	"run": true,
	// The boundary between slot At-1 and slot At, zoomed, with its cost.
	"switch": true,
	// The finished timeline with each lane's total. The closer.
	"fair": true,
}

// SchedulerShows returns the beat vocabulary sorted.
func SchedulerShows() []string {
	out := make([]string, 0, len(schedulerShows))
	for k := range schedulerShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// SchedulerSpec is the chart: who is waiting, under what policy, and the turns
// they actually get. On the plan because one timeline stands for the clip.
type SchedulerSpec struct {
	// Policy names the rule being illustrated — "Round Robin".
	Policy string `json:"policy"`
	// Procs are the lanes, drawn top to bottom in this order.
	Procs []SchedulerProc `json:"procs"`
	// Slots are the turns, in the order they run.
	Slots []SchedulerSlot `json:"slots"`
}

// SchedulerProc is one lane.
type SchedulerProc struct {
	// Label names the process — "P1", "editor".
	Label string `json:"label"`
}

// SchedulerSlot is one turn on the CPU.
type SchedulerSlot struct {
	// Proc indexes SchedulerSpec.Procs — whose turn this is.
	Proc int `json:"proc"`
	// Len is how many time units the turn lasts.
	Len int `json:"len"`
}

// SchedulerBeat is one shot: which state of the timeline this beat shows.
type SchedulerBeat struct {
	// Show is a schedulerShows name.
	Show string `json:"show"`
	// At indexes SchedulerSpec.Slots: the slot being laid down for a "run",
	// and the slot on the RIGHT of the boundary for a "switch".
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a run — the
// workhorse state most beats of this template are in.
func (b SchedulerBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if schedulerShows[s] {
		return s
	}
	return "run"
}

func normalizeSchedulerPlan(p *SnippetPlan) {
	s := p.Scheduler
	if s == nil {
		return
	}
	s.Policy = clampWords(collapseSpaces(s.Policy), maxSchedulerPolicyWords)

	procs := make([]SchedulerProc, 0, len(s.Procs))
	for _, pr := range s.Procs {
		pr.Label = clampWords(collapseSpaces(pr.Label), maxSchedulerLabelWords)
		if len(procs) < maxSchedulerProcs {
			procs = append(procs, pr)
		}
	}
	s.Procs = procs

	slots := make([]SchedulerSlot, 0, len(s.Slots))
	for _, sl := range s.Slots {
		if sl.Proc < 0 {
			sl.Proc = 0
		}
		if n := len(s.Procs); n > 0 && sl.Proc >= n {
			sl.Proc = n - 1
		}
		if sl.Len < minSchedulerSlotLen {
			sl.Len = minSchedulerSlotLen
		}
		if sl.Len > maxSchedulerSlotLen {
			sl.Len = maxSchedulerSlotLen
		}
		if len(slots) < maxSchedulerSlots {
			slots = append(slots, sl)
		}
	}
	s.Slots = slots

	for i := range p.Beats {
		b := p.Beats[i].Scheduler
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "run" && b.Show != "switch" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(s.Slots); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateSchedulerPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Scheduler: true}); err != nil {
		return err
	}

	s := p.Scheduler
	if s == nil {
		return fmt.Errorf("the plan has no schedule — this template is lanes of processes over one time axis, so the chart is the clip")
	}
	if strings.TrimSpace(s.Policy) == "" {
		return fmt.Errorf("the chart has no policy. Round Robin, first come first served and priority make visibly different pictures out of the same processes, so a Gantt chart with no policy on it is a chart of one arbitrary outcome")
	}
	if n := len(s.Procs); n < minSchedulerProcs || n > maxSchedulerProcs {
		return fmt.Errorf("the chart has %d process(es), want %d-%d. One process shares nothing with anybody, and past %d the lanes are shorter than their own labels once the axis and the totals have taken their share of the height",
			n, minSchedulerProcs, maxSchedulerProcs, maxSchedulerProcs)
	}
	for i, pr := range s.Procs {
		if strings.TrimSpace(pr.Label) == "" {
			return fmt.Errorf("process %d has no label — an unnamed lane is a row the viewer cannot attribute a block to", i)
		}
	}
	if n := len(s.Slots); n < minSchedulerSlots || n > maxSchedulerSlots {
		return fmt.Errorf("the schedule has %d turn(s), want %d-%d. Two turns is an alternation rather than a schedule — the pattern a policy makes only becomes visible on the third — and past %d the blocks are narrower than the switch band drawn between them",
			n, minSchedulerSlots, maxSchedulerSlots, maxSchedulerSlots)
	}

	// THE ARITHMETIC. The axis is summed in Go, and every lane has to be used.
	ran := make([]int, len(s.Procs))
	total := 0
	for i, sl := range s.Slots {
		if sl.Proc < 0 || sl.Proc >= len(s.Procs) {
			return fmt.Errorf("turn %d belongs to process %d, which does not exist — the chart holds processes 0-%d", i, sl.Proc, len(s.Procs)-1)
		}
		if sl.Len < minSchedulerSlotLen || sl.Len > maxSchedulerSlotLen {
			return fmt.Errorf("turn %d (process %q) is %d unit(s) long, want %d-%d. A zero-length turn is not a turn, and past %d one slot eats a quarter of the axis and the rest become slivers",
				i, s.Procs[sl.Proc].Label, sl.Len, minSchedulerSlotLen, maxSchedulerSlotLen, maxSchedulerSlotLen)
		}
		ran[sl.Proc]++
		total += sl.Len
	}
	var idle []string
	for i, n := range ran {
		if n == 0 {
			idle = append(idle, fmt.Sprintf("%q (process %d)", s.Procs[i].Label, i))
		}
	}
	if len(idle) > 0 {
		return fmt.Errorf("these processes never run: %s. A lane that stays empty for the whole clip is a row of nothing the viewer spends the entire chart waiting to see used — give each one a turn, or take it out of the queue",
			strings.Join(idle, ", "))
	}
	if total > maxSchedulerUnits {
		return fmt.Errorf("the schedule runs %d time units, and the axis holds %d. Past %d the unit ticks are closer together than the eye separates them, and a tick nobody can count is not a tick — shorten the turns or use fewer of them so the lengths sum to %d or less",
			total, maxSchedulerUnits, maxSchedulerUnits, maxSchedulerUnits)
	}

	// The shape. The queue is seen before anything is scheduled.
	if p.Beats[0].Scheduler == nil || p.Beats[0].Scheduler.ResolvedShow() != "queue" {
		return fmt.Errorf("beat %q does not open on the queue. A block landing on a timeline nobody has been shown is a rectangle appearing — the first beat is {\"show\": \"queue\"}",
			p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Scheduler == nil || last.Scheduler.ResolvedShow() != "fair" {
		return fmt.Errorf("the clip does not close on the finished timeline. The last frame is the whole schedule with each lane's total beside it — end with {\"show\": \"fair\"}")
	}

	next := 0
	for i, b := range p.Beats {
		if b.Scheduler == nil {
			return fmt.Errorf("beat %q has no scheduler direction — every beat shows one state of the chart", b.ID)
		}
		switch b.Scheduler.ResolvedShow() {
		case "queue":
			if i != 0 {
				return fmt.Errorf("beat %q empties the timeline part-way through. The queue is the opener; going back to it takes blocks off a chart the clip already laid down", b.ID)
			}
		case "fair":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q totals the lanes before the schedule is over. \"fair\" is the closer — totals struck while turns are still to come are the wrong totals", b.ID)
			}
		case "run":
			at := b.Scheduler.At
			if at < 0 || at >= len(s.Slots) {
				return fmt.Errorf("beat %q runs turn %d, which does not exist — the schedule holds turns 0-%d", b.ID, at, len(s.Slots)-1)
			}
			if at != next {
				return fmt.Errorf("beat %q runs turn %d, but turn %d is the next one on the axis. Turns are laid down left to right with no gaps: a block appearing to the right of a hole says time passed with nobody on the CPU",
					b.ID, at, next)
			}
			next++
		case "switch":
			at := b.Scheduler.At
			if at < 1 || at >= len(s.Slots) {
				return fmt.Errorf("beat %q zooms the boundary before turn %d. A context switch is the changeover BETWEEN two turns, so it needs a turn on each side — the boundaries are 1 to %d",
					b.ID, at, len(s.Slots)-1)
			}
			if next <= at {
				return fmt.Errorf("beat %q zooms the boundary between turns %d and %d, but only %d turn(s) have been laid down. The cost of a switch is the cost of a changeover, and with one side of it not yet on screen there is no changeover to charge for — run both turns first",
					b.ID, at-1, at, next)
			}
		}
	}
	return nil
}

// schedulerScenes lays the clip out as ONE scene. Every number the chart shows
// — where each block starts, how long the axis is, what each lane totals — is
// computed here, so the component measures nothing and adds nothing up.
func schedulerScenes(in SnippetSceneInput) ([]Scene, error) {
	s := in.Plan.Scheduler
	if s == nil {
		return nil, fmt.Errorf("the plan has no schedule")
	}

	totals := make([]int, len(s.Procs))
	slots := make([]map[string]any, len(s.Slots))
	at := 0
	for i, sl := range s.Slots {
		if sl.Proc < 0 || sl.Proc >= len(s.Procs) {
			return nil, fmt.Errorf("turn %d belongs to process %d, which does not exist", i, sl.Proc)
		}
		slots[i] = map[string]any{
			"proc":  sl.Proc,
			"len":   sl.Len,
			"start": at,
		}
		totals[sl.Proc] += sl.Len
		at += sl.Len
	}
	procs := make([]map[string]any, len(s.Procs))
	for i, pr := range s.Procs {
		procs[i] = map[string]any{
			"label": pr.Label,
			"total": totals[i],
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	laid := 0
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Scheduler == nil {
			return nil, fmt.Errorf("beat %q has no scheduler direction", beat.ID)
		}
		show := beat.Scheduler.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "run":
			laid = beat.Scheduler.At + 1
			step["at"] = beat.Scheduler.At
		case "switch":
			step["at"] = beat.Scheduler.At
			// The unit on the axis where the changeover happens, so the band is
			// drawn from a number rather than measured off the blocks.
			step["boundary"] = slots[beat.Scheduler.At]["start"]
		case "fair":
			// The closer completes the schedule: whatever the beats did not get
			// to still belongs on the finished chart.
			laid = len(s.Slots)
		}
		step["laid"] = laid
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneScheduler,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"policy": s.Policy,
			"procs":  procs,
			"slots":  slots,
			"units":  at,
			"steps":  steps,
		}),
	}}, nil
}
