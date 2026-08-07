package pipeline

// The mission template: build this — the mini-project brief.
//
// A foundations course that is all watching is a course nobody finishes. The
// point where a career-switcher starts believing they can do this work is the
// first brief that treats them like someone who ships: a goal, a spec
// checklist, a named artifact, and a definition of done. The catalog could
// present, compare and quiz; it could not ASSIGN. This template is the
// assignment card, and it sits in the decisions category because its job is
// to end deliberation — after this clip the viewer is either building or has
// decided not to, and both are decisions the brief forced.
//
// Every field earns its place by being the thing beginners are never given.
//
// **The deliverable is an artifact, not an activity.** "A CLI script", "a
// report" — a noun you could attach to an email. Briefs that assign an
// activity ("practice loops") produce practice, which nobody can show anyone.
//
// **Done is observable.** "The script lists every drive with its size" can be
// checked by running it; "you feel comfortable with files" cannot, and a
// mission whose end cannot be recognised does not end. Done stamps LAST, over
// everything, because a definition of done read before the specs is a test
// for a project that does not exist yet.
//
// **Every spec is landed exactly once, and the brief opens.** A checklist row
// the narrator never reads is a requirement the builder will discover at the
// end; one read twice suggests the list is padding. The goal comes first and
// alone, because a mission is a sentence before it is a list — a viewer who
// cannot hold the goal has nothing to hang six specs on.
//
// Three to six specs is the band: under three the checklist is the goal
// restated, past six the mission is a sprint, and this card is a brief.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "mission",
		Category:    CatDecisions,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "Build this",
		Description: "The mini-project brief: a goal, a spec checklist, the artifact to deliver, and an observable definition of done. Reach for it to close a module with an assignment the viewer can actually start.",
		Example:     "Build a Hardware Inventory Report script",
		PromptFile:  snippetMissionTemplateName,
		NeedsCode:   false,
		// The brief, three-plus specs, the artifact, the stamp: the shortest
		// honest mission is six beats, which ~25s of narration just funds.
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		// List-shaped: the brief, up to six landing items, and the stamp.
		MaxBeats:  8,
		Owns:      beatFields{Mission: true},
		OwnsPlan:  planFields{Mission: true},
		Normalize: normalizeMissionPlan,
		Validate:  validateMissionPlan,
		Scenes:    missionScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":               strings.Join(MetricRoles(), ", "),
				"Shows":               strings.Join(MissionShows(), ", "),
				"MinSpecs":            minMissionSpecs,
				"MaxSpecs":            maxMissionSpecs,
				"MaxGoalWords":        maxMissionGoalWords,
				"MaxSpecWords":        maxMissionSpecWords,
				"MaxDeliverableWords": maxMissionDeliverableWords,
				"MaxDoneWords":        maxMissionDoneWords,
			}
		},
	})
}

const snippetMissionTemplateName = "snippet_mission.tmpl"

const (
	// Under three specs the checklist is the goal restated; past six the
	// mission is a sprint, and this card is a brief.
	minMissionSpecs = 3
	maxMissionSpecs = 6

	// The goal is one sentence at the top of the card; ten words is a goal,
	// more is a description.
	maxMissionGoalWords = 10
	// A spec is one checklist row; eight words keeps six rows on one card.
	maxMissionSpecWords = 8
	// The deliverable is a chip — a noun phrase, not a plan.
	maxMissionDeliverableWords = 8
	// Done is one observable sentence on the DONE-WHEN strip.
	maxMissionDoneWords = 12
)

// missionShows is the closed vocabulary of what a beat does.
var missionShows = map[string]bool{
	// The goal, alone. The opener.
	"brief": true,
	// Checklist item At lands.
	"spec": true,
	// The artifact chip appears.
	"deliverable": true,
	// The definition of done stamps over everything. The closer.
	"done": true,
}

// MissionShows returns the beat vocabulary sorted, for the prompt.
func MissionShows() []string {
	out := make([]string, 0, len(missionShows))
	for k := range missionShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// MissionSpec is the project brief.
type MissionSpec struct {
	// Goal is what gets built, as one sentence.
	Goal string `json:"goal"`
	// Specs are the checklist, top to bottom.
	Specs []string `json:"specs"`
	// Deliverable is the artifact — "a CLI script", "a report".
	Deliverable string `json:"deliverable"`
	// Done is the observable definition of done.
	Done string `json:"done"`
}

// MissionBeat is one move on the card.
type MissionBeat struct {
	// Show is a missionShows name.
	Show string `json:"show"`
	// At indexes MissionSpec.Specs, for a "spec" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults an unknown direction to landing a spec, which is what
// most beats of this template do.
func (b MissionBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if missionShows[s] {
		return s
	}
	return "spec"
}

func normalizeMissionPlan(p *SnippetPlan) {
	m := p.Mission
	if m == nil {
		return
	}
	m.Goal = clampWords(collapseSpaces(m.Goal), maxMissionGoalWords)
	m.Deliverable = clampWords(collapseSpaces(m.Deliverable), maxMissionDeliverableWords)
	m.Done = clampWords(collapseSpaces(m.Done), maxMissionDoneWords)
	specs := make([]string, 0, len(m.Specs))
	for _, s := range m.Specs {
		s = clampWords(collapseSpaces(s), maxMissionSpecWords)
		if s == "" {
			continue
		}
		if len(specs) < maxMissionSpecs {
			specs = append(specs, s)
		}
	}
	m.Specs = specs

	for i := range p.Beats {
		b := p.Beats[i].Mission
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "spec" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(m.Specs); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateMissionPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Mission: true}); err != nil {
		return err
	}

	m := p.Mission
	if m == nil {
		return fmt.Errorf("the plan has no mission — this template is a project brief, so the goal, specs and definition of done are the clip")
	}
	if strings.TrimSpace(m.Goal) == "" {
		return fmt.Errorf("the mission has no goal. The brief opens on one sentence saying what gets built — without it the checklist is requirements for nothing")
	}
	if n := len(m.Specs); n < minMissionSpecs || n > maxMissionSpecs {
		return fmt.Errorf("the checklist has %d spec(s), want %d-%d. Under three it is the goal restated; past six the mission is a sprint, and this card is a brief",
			n, minMissionSpecs, maxMissionSpecs)
	}
	if strings.TrimSpace(m.Deliverable) == "" {
		return fmt.Errorf("the mission names no deliverable. The artifact is a noun you could attach to an email — \"a CLI script\", \"a report\". A brief that assigns an activity produces practice, which nobody can show anyone")
	}
	if strings.TrimSpace(m.Done) == "" {
		return fmt.Errorf("the mission has no definition of done. Done is the observable check — \"the script lists every drive with its size\" — and a mission whose end cannot be recognised does not end")
	}

	if p.Beats[0].Mission == nil || p.Beats[0].Mission.ResolvedShow() != "brief" {
		return fmt.Errorf("beat %q does not open on the brief. The goal comes first and alone — a mission is a sentence before it is a list, and the viewer needs something to hang the specs on",
			p.Beats[0].ID)
	}

	landed := map[int]bool{}
	deliverables := 0
	dones := 0
	doneAt := -1
	for i, b := range p.Beats {
		if b.Mission == nil {
			return fmt.Errorf("beat %q has no mission direction — every beat states the goal, lands a spec, shows the deliverable, or stamps done", b.ID)
		}
		switch b.Mission.ResolvedShow() {
		case "brief":
			if i != 0 {
				return fmt.Errorf("beat %q re-states the brief part-way through. The goal is the opener; repeating it wipes the checklist the card has been building", b.ID)
			}
		case "spec":
			at := b.Mission.At
			if at < 0 || at >= len(m.Specs) {
				return fmt.Errorf("beat %q lands spec %d, which does not exist", b.ID, at)
			}
			if landed[at] {
				return fmt.Errorf("beat %q lands spec %d (%q) a second time. A checklist row lands once — reading it twice suggests the list is padding",
					b.ID, at, m.Specs[at])
			}
			landed[at] = true
		case "deliverable":
			deliverables++
		case "done":
			dones++
			doneAt = i
		}
	}
	if deliverables != 1 {
		return fmt.Errorf("there are %d deliverable beats, want exactly 1. The artifact is named once — zero leaves the builder shipping nothing in particular, and two is the same chip landing twice", deliverables)
	}
	if dones != 1 {
		return fmt.Errorf("there are %d done beats, want exactly 1. One stamp; a mission that is done twice was never clearly done at all", dones)
	}
	if doneAt != len(p.Beats)-1 {
		return fmt.Errorf("beat %q stamps done and then the clip keeps going. Done is the closer — it stamps over everything, because a definition of done read before the specs is a test for a project that does not exist yet",
			p.Beats[doneAt].ID)
	}
	if len(landed) != len(m.Specs) {
		return fmt.Errorf("%d of the %d specs are never landed. A checklist row the narrator skips is a requirement the builder discovers at the end — give it a beat or cut it",
			len(m.Specs)-len(landed), len(m.Specs))
	}
	return nil
}

// missionScenes lays the clip out as ONE scene. The card builds downward; the
// steps only say which rows have landed and what has stamped.
func missionScenes(in SnippetSceneInput) ([]Scene, error) {
	m := in.Plan.Mission
	if m == nil {
		return nil, fmt.Errorf("the plan has no mission")
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	landed := map[int]bool{}
	deliverableOn := false
	doneOn := false
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Mission == nil {
			return nil, fmt.Errorf("beat %q has no mission direction", beat.ID)
		}
		show := beat.Mission.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "spec":
			landed[beat.Mission.At] = true
			step["at"] = beat.Mission.At
		case "deliverable":
			deliverableOn = true
		case "done":
			doneOn = true
		}
		rows := make([]int, 0, len(landed))
		for at := range landed {
			rows = append(rows, at)
		}
		sort.Ints(rows)
		step["landed"] = rows
		step["deliverableOn"] = deliverableOn
		step["doneOn"] = doneOn
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneMission,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":       in.Plan.Title,
			"goal":        m.Goal,
			"specs":       append([]string{}, m.Specs...),
			"deliverable": m.Deliverable,
			"done":        m.Done,
			"steps":       steps,
		}),
	}}, nil
}
