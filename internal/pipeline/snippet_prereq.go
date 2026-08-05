package pipeline

// The prereq template: what this lesson assumes, and how to get out of it.
//
// A course loses people in the gap between lessons, and it loses them silently:
// the viewer does not know they are missing something, they just find lesson six
// harder than lesson five and conclude they are bad at this. The fix is not a
// harder lesson six, it is saying out loud what lesson six stands on.
//
// Two rules earn it its place.
//
// **Every assumption resolves.** An assumption is either something this course
// already taught — and then it names WHERE, so the viewer can go back — or it is
// external, and then it says so plainly. A bare list of nouns is the failure
// mode: "you'll need to know about promises, streams and backpressure" tells
// somebody who does not know those things nothing except that they are behind.
//
// **At least one assumption is skippable.** This is the rule that makes the
// template worth having rather than a wall with a list on it. Most lessons have
// a genuine floor and a lot of nice-to-have, and a viewer who cannot tell them
// apart treats all of it as the floor and leaves. Marking what can be skipped —
// and what breaks if you skip it — is the difference between a gate and a map.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "prereq",
		Category:    CatPresenting,
		Since:       SinceV5,
		Title:       "What this one stands on",
		Description: "The floor a lesson assumes: each thing it takes for granted, whether the course already taught it or you bring it, and which of them you can skip anyway. Reach for it when a lesson is about to get steep.",
		Example:     "Set up a lesson on backpressure for people who have seen streams but never a slow consumer",
		PromptFile:  snippetPrereqTemplateName,
		MinTargetSec:     20,
		DefaultTargetSec: 40,
		MaxBeats:         7,
		Owns:             beatFields{Prereq: true},
		OwnsPlan:         planFields{Prereq: true},
		Normalize:        normalizePrereqPlan,
		Validate:         validatePrereqPlan,
		Scenes:           prereqScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				// The emphasis roles every headline picks from.
				"Roles":          strings.Join(MetricRoles(), ", "),
				"Sources":        strings.Join(PrereqSources(), ", "),
				"MinAssumptions": minPrereqAssumptions,
				"MaxAssumptions": maxPrereqAssumptions,
				"MaxItemWords":   maxPrereqItemWords,
				"MaxWhereWords":  maxPrereqWhereWords,
				"MaxBreaksWords": maxPrereqBreaksWords,
			}
		},
	})
}

const snippetPrereqTemplateName = "snippet_prereq.tmpl"

const (
	minPrereqAssumptions = 2
	maxPrereqAssumptions = 5

	maxPrereqItemWords   = 8
	maxPrereqWhereWords  = 12
	maxPrereqBreaksWords = 16
)

// prereqSources say where an assumption is supposed to have come from.
var prereqSources = map[string]bool{
	// Taught earlier in this course. Where must name the lesson.
	"taught": true,
	// Brought from outside. Honest, and the viewer can go and get it.
	"external": true,
}

// PrereqSources returns the vocabulary, sorted, for the prompt.
func PrereqSources() []string {
	out := make([]string, 0, len(prereqSources))
	for s := range prereqSources {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// PrereqSpec is the floor a lesson stands on.
type PrereqSpec struct {
	// Assumptions are what the lesson takes for granted, hardest first.
	Assumptions []PrereqItem `json:"assumptions"`
}

// PrereqItem is one thing the lesson assumes.
type PrereqItem struct {
	// Item is the thing itself — a capability, not a topic. "Reading a stack
	// trace", not "stack traces".
	Item string `json:"item"`
	// Source is a prereqSources name.
	Source string `json:"source"`
	// Where names the lesson that taught it, for a "taught" item, or where to
	// get it, for an external one.
	Where string `json:"where,omitempty"`
	// Skippable marks an assumption the viewer can proceed without.
	Skippable bool `json:"skippable,omitempty"`
	// Breaks is what stops working if a skippable item is skipped. Required when
	// Skippable — "you can skip this" with no consequence is not permission, it
	// is a shrug.
	Breaks string `json:"breaks,omitempty"`
}

// ResolvedSource defaults an unknown source to external, which is the honest
// answer when the model could not place it in the course.
func (i PrereqItem) ResolvedSource() string {
	s := strings.ToLower(strings.TrimSpace(i.Source))
	if prereqSources[s] {
		return s
	}
	return "external"
}

// PrereqBeat says which assumption this beat is on.
type PrereqBeat struct {
	At   int    `json:"at,omitempty"`
	Show string `json:"show"`
}

var prereqShows = map[string]bool{
	// One assumption, named and placed.
	"assume": true,
	// The skippable ones, marked as such.
	"skip": true,
	// The whole floor at once.
	"floor": true,
}

// PrereqShows returns the beat vocabulary, sorted, for the prompt.
func PrereqShows() []string {
	out := make([]string, 0, len(prereqShows))
	for s := range prereqShows {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// ResolvedShow defaults to naming an assumption.
func (b PrereqBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if prereqShows[s] {
		return s
	}
	return "assume"
}

func normalizePrereqPlan(p *SnippetPlan) {
	if p.Prereq == nil {
		return
	}
	for i := range p.Prereq.Assumptions {
		a := &p.Prereq.Assumptions[i]
		a.Item = strings.TrimSpace(a.Item)
		a.Source = a.ResolvedSource()
		a.Where = strings.TrimSpace(a.Where)
		a.Breaks = strings.TrimSpace(a.Breaks)
		// A non-skippable item with a consequence attached is a model hedging;
		// the consequence has nowhere to render, so drop it rather than reject.
		if !a.Skippable {
			a.Breaks = ""
		}
	}
	for i := range p.Beats {
		if b := p.Beats[i].Prereq; b != nil {
			b.Show = b.ResolvedShow()
			if b.At < 0 || b.At >= len(p.Prereq.Assumptions) {
				b.At = 0
			}
		}
	}
}

func validatePrereqPlan(p *SnippetPlan) error {
	pr := p.Prereq
	if pr == nil {
		return fmt.Errorf("the plan has no prerequisites — this template says what a lesson stands on, so the assumptions are the clip")
	}
	if n := len(pr.Assumptions); n < minPrereqAssumptions || n > maxPrereqAssumptions {
		return fmt.Errorf("the lesson names %d assumption(s); this template takes %d to %d. Fewer than %d is not a floor, more than %d is a syllabus",
			n, minPrereqAssumptions, maxPrereqAssumptions, minPrereqAssumptions, maxPrereqAssumptions)
	}

	skippable := 0
	for i, a := range pr.Assumptions {
		if strings.TrimSpace(a.Item) == "" {
			return fmt.Errorf("assumption %d is empty", i+1)
		}
		if n := len(strings.Fields(a.Item)); n > maxPrereqItemWords {
			return fmt.Errorf("assumption %d is %d words; keep it to %d. It is a capability on a line, not a description", i+1, n, maxPrereqItemWords)
		}
		if strings.TrimSpace(a.Where) == "" {
			return fmt.Errorf("assumption %d (%q) does not say where it comes from. A %s assumption has to name the lesson that taught it, or where to go and get it — a bare list of nouns tells somebody who lacks them only that they are behind",
				i+1, a.Item, a.ResolvedSource())
		}
		if n := len(strings.Fields(a.Where)); n > maxPrereqWhereWords {
			return fmt.Errorf("the source for assumption %d is %d words; keep it to %d", i+1, n, maxPrereqWhereWords)
		}
		if a.Skippable {
			skippable++
			if strings.TrimSpace(a.Breaks) == "" {
				return fmt.Errorf("assumption %d (%q) is marked skippable but does not say what breaks. \"You can skip this\" with no consequence is not permission, it is a shrug — say what stops working so the viewer can decide",
					i+1, a.Item)
			}
			if n := len(strings.Fields(a.Breaks)); n > maxPrereqBreaksWords {
				return fmt.Errorf("the consequence for assumption %d is %d words; keep it to %d", i+1, n, maxPrereqBreaksWords)
			}
		}
	}
	if skippable == 0 {
		return fmt.Errorf("every assumption is required. A floor with no way through it is a wall, and a viewer who cannot tell the real floor from the nice-to-have treats all of it as the floor and leaves. Mark at least one assumption skippable and say what breaks without it")
	}
	if skippable == len(pr.Assumptions) {
		return fmt.Errorf("every assumption is skippable, which means the lesson assumes nothing and this clip has no work to do. At least one thing has to be the actual floor")
	}

	named := map[int]bool{}
	for _, b := range p.Beats {
		if b.Prereq == nil {
			return fmt.Errorf("beat %q has no prereq direction — say whether it names an assumption, marks what can be skipped, or shows the whole floor", b.ID)
		}
		if b.Prereq.ResolvedShow() == "assume" {
			named[b.Prereq.At] = true
		}
	}
	for i, a := range pr.Assumptions {
		if !named[i] {
			return fmt.Errorf("assumption %d (%q) is never said out loud. An assumption the clip lists but never speaks is one the viewer will not register", i+1, a.Item)
		}
	}
	return nil
}

func prereqScenes(in SnippetSceneInput) ([]Scene, error) {
	pr := in.Plan.Prereq
	if pr == nil {
		return nil, fmt.Errorf("the plan has no prerequisites")
	}

	items := make([]map[string]any, len(pr.Assumptions))
	for i, a := range pr.Assumptions {
		items[i] = map[string]any{
			"item":      a.Item,
			"source":    a.ResolvedSource(),
			"where":     a.Where,
			"skippable": a.Skippable,
			"breaks":    a.Breaks,
		}
	}

	shown := map[int]bool{}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Prereq == nil {
			return nil, fmt.Errorf("beat %q has no prereq direction", beat.ID)
		}
		show := beat.Prereq.ResolvedShow()
		if show == "assume" {
			shown[beat.Prereq.At] = true
		}
		if show == "floor" || show == "skip" {
			for j := range pr.Assumptions {
				shown[j] = true
			}
		}
		lit := make([]int, 0, len(shown))
		for at := range shown {
			lit = append(lit, at)
		}
		sort.Ints(lit)

		step := map[string]any{"startMs": startMs, "endMs": endMs, "show": show, "lit": lit}
		if show == "assume" {
			step["at"] = beat.Prereq.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    ScenePrereq,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":       in.Plan.Title,
			"assumptions": items,
			"steps":       steps,
		}),
	}}, nil
}
