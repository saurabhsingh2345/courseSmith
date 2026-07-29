package pipeline

// The spec template: write down how you will know it is done, then check.
//
// This is the skill underneath both of the courses this batch of templates was
// written for, and it is the one that transfers. In vibe coding it is
// spec-driven prompting — goals, constraints, acceptance criteria, written
// before the ask rather than discovered by looking at what came back. In no-code
// it is the same move under a different name: deciding what the automation is
// supposed to do before wiring anything, so you can tell whether it does.
//
// Nothing in the catalog could show it. `quiz` asks the viewer a question.
// `promptloop` shows the conversation, and it teaches the other half of the
// lesson — that the first attempt falls short — but it has no way to say what
// "short" was measured against. `mockup` shows the artefact. A checklist being
// satisfied is a different thing from all three, and the satisfying part of it
// is temporal: the gap between writing a criterion and watching it go green is
// where the idea lands, exactly as the gap between a quiz's question and its
// answer is where that one lands.
//
// So the clip is in two halves, and the shape is the argument. The beats write
// the criteria, one at a time, with nothing ticked. Then one closing beat checks
// them all at once. A clip that ticked each criterion as it wrote it would be
// showing a build that happened to be described afterwards, which is precisely
// the habit this template exists to argue against.
//
// A criterion may be missed, and the closing beat says so rather than lying.
// Most honest specs have one — that is what a spec is FOR — and a template that
// could only show a clean sweep would be useless in exactly the lessons where
// the idea matters most.

import (
	"fmt"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:             "spec",
		Category:         CatCode,
		Title:            "Spec checklist",
		Description:      "Acceptance criteria written down first, then checked off — the skill that transfers across every tool.",
		Example:          "How to write a prompt that tells you whether the result is any good",
		PromptFile:       snippetSpecTemplateName,
		NeedsCode:        false,
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		Owns:             beatFields{Spec: true},
		OwnsPlan:         planFields{Spec: true},
		Normalize:        normalizeSpecPlan,
		Validate:         validateSpecPlan,
		Scenes:           specScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"MinCriteria":       minSpecCriteria,
				"MaxCriteria":       maxSpecCriteria,
				"MaxConstraints":    maxSpecConstraints,
				"MaxGoalWords":      maxSpecGoalWords,
				"MaxCriterionWords": maxSpecCriterionWords,
				"MaxConstraintWds":  maxSpecConstraintWords,
				"MaxNoteWords":      maxSpecNoteWords,
			}
		},
	})
}

const snippetSpecTemplateName = "snippet_spec.tmpl"

// Sheet capacity. Five criteria at 78px each is what the card holds before the
// list stops being readable at 1080p; three is the floor because a spec with two
// lines in it has not yet done the thing a spec does, which is to be specific
// enough to be inconvenient.
const (
	minSpecCriteria    = 3
	maxSpecCriteria    = 5
	maxSpecConstraints = 3

	maxSpecGoalWords       = 12
	maxSpecCriterionWords  = 10
	maxSpecConstraintWords = 4
	maxSpecNoteWords       = 16
)

// specStatuses is how a criterion came out when it was checked.
var specStatuses = map[string]bool{"met": true, "missed": true}

// SpecSheet is the whole spec. On the plan rather than per-beat for the same
// reason the quiz's question is: it is the subject of the clip, not a property
// of one moment in it.
type SpecSheet struct {
	// Goal is the ask, in one line. It heads the card for the whole clip.
	Goal string `json:"goal"`
	// Constraints are the boundaries the answer has to respect — "No backend",
	// "Ship today", "Free tier only". Shown as pills under the goal. Optional,
	// and usually where the interesting part of a spec actually is.
	Constraints []string `json:"constraints,omitempty"`
	// Criteria are how you will know it is done, in the order they are written.
	Criteria []SpecCriterion `json:"criteria"`
}

// SpecCriterion is one line of the checklist.
type SpecCriterion struct {
	// Text is the criterion — a testable statement, not a wish.
	Text string `json:"text"`
	// Status is how it came out when checked: a specStatuses name. Empty means
	// "met", because a spec whose lines are all satisfiable is the ordinary
	// case and a model should not have to say so five times.
	Status string `json:"status,omitempty"`
	// Note expands on it, and is shown only while this line is being written.
	Note string `json:"note,omitempty"`
}

// ResolvedStatus returns the criterion's outcome, defaulting to met.
func (c SpecCriterion) ResolvedStatus() string {
	s := strings.ToLower(strings.TrimSpace(c.Status))
	if specStatuses[s] {
		return s
	}
	return "met"
}

// SpecBeat says which criterion this beat is writing, or that it is checking
// the whole sheet.
type SpecBeat struct {
	// At indexes SpecSheet.Criteria.
	At int `json:"at"`
	// Check marks the closing beat that runs down the list and marks every
	// line. A flag rather than an out-of-range index for the same reason the
	// canvas's Run is: `at` omitted decodes to 0, which is a real criterion.
	Check bool `json:"check,omitempty"`
}

func normalizeSpecPlan(p *SnippetPlan) {
	s := p.Spec
	if s == nil {
		return
	}
	s.Goal = clampWords(collapseSpaces(s.Goal), maxSpecGoalWords)
	constraints := make([]string, 0, len(s.Constraints))
	for _, c := range s.Constraints {
		c = clampWords(collapseSpaces(c), maxSpecConstraintWords)
		if c != "" && len(constraints) < maxSpecConstraints {
			constraints = append(constraints, c)
		}
	}
	s.Constraints = constraints
	for i := range s.Criteria {
		c := &s.Criteria[i]
		c.Text = clampWords(collapseSpaces(c.Text), maxSpecCriterionWords)
		c.Note = clampWords(collapseSpaces(c.Note), maxSpecNoteWords)
		c.Status = c.ResolvedStatus()
	}
	for i := range p.Beats {
		if b := p.Beats[i].Spec; b != nil && b.At < 0 {
			b.At = 0
		}
	}
}

func validateSpecPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Spec: true}); err != nil {
		return err
	}

	s := p.Spec
	if s == nil || strings.TrimSpace(s.Goal) == "" {
		return fmt.Errorf("the plan has no goal — a spec is criteria for something, and the something goes at the top of the card")
	}
	if n := len(s.Criteria); n < minSpecCriteria || n > maxSpecCriteria {
		return fmt.Errorf("the spec has %d criteria, want %d-%d — two lines has not yet done the thing a spec does, which is to be specific enough to be inconvenient",
			n, minSpecCriteria, maxSpecCriteria)
	}
	seen := map[string]bool{}
	met := 0
	for i, c := range s.Criteria {
		if strings.TrimSpace(c.Text) == "" {
			return fmt.Errorf("criterion %d is empty — write what would have to be true", i)
		}
		key := strings.ToLower(strings.TrimSpace(c.Text))
		if seen[key] {
			return fmt.Errorf("criterion %d repeats %q — each line is a different thing to check", i, c.Text)
		}
		seen[key] = true
		if c.ResolvedStatus() == "met" {
			met++
		}
	}
	// A sheet where nothing was met is not a spec being checked, it is a
	// failure being described, and the promptloop template tells that story with
	// somewhere for it to go next.
	if met == 0 {
		return fmt.Errorf("every criterion is missed. That is not a spec being checked, it is an attempt that failed — the promptloop template tells that story, because it has somewhere for the next prompt to go")
	}

	visited := map[int]bool{}
	last := -1
	sawCheck := false
	for _, b := range p.Beats {
		if b.Spec == nil {
			return fmt.Errorf("beat %q has no spec direction — every beat is writing a criterion or checking the sheet", b.ID)
		}
		if b.Spec.Check {
			sawCheck = true
			continue
		}
		if b.Spec.At < 0 || b.Spec.At >= len(s.Criteria) {
			return fmt.Errorf("beat %q writes criterion %d, which does not exist", b.ID, b.Spec.At)
		}
		if b.Spec.At < last {
			return fmt.Errorf("beat %q goes back to criterion %d after %d. The list is written in order — if the clip genuinely revisits an earlier line, it is arguing about the spec rather than writing one",
				b.ID, b.Spec.At, last)
		}
		if visited[b.Spec.At] {
			return fmt.Errorf("beat %q writes criterion %d again; each line gets one beat", b.ID, b.Spec.At)
		}
		visited[b.Spec.At] = true
		last = b.Spec.At
	}
	if len(visited) != len(s.Criteria) {
		return fmt.Errorf("%d of the %d criteria are never written — a line nobody explains is a sentence with a box next to it",
			len(s.Criteria)-len(visited), len(s.Criteria))
	}
	// The payoff, and the reason the two halves are separate.
	if !sawCheck {
		return fmt.Errorf("no beat checks the spec. Close with a beat carrying \"check\": true — the gap between writing a criterion and watching it go green is where the idea lands")
	}
	if lastBeat := p.Beats[len(p.Beats)-1].Spec; lastBeat != nil && !lastBeat.Check {
		return fmt.Errorf("the clip ends still writing the list; end by checking it instead (\"check\": true)")
	}
	return nil
}

// specScenes lays the clip out as ONE scene: the card is on screen throughout
// and the beats only move how much of the list has been written.
func specScenes(in SnippetSceneInput) ([]Scene, error) {
	s := in.Plan.Spec
	if s == nil {
		return nil, fmt.Errorf("the plan has no spec")
	}

	criteria := make([]map[string]any, len(s.Criteria))
	for i, c := range s.Criteria {
		criteria[i] = map[string]any{
			"text":   c.Text,
			"status": c.ResolvedStatus(),
			"note":   c.Note,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Spec == nil {
			return nil, fmt.Errorf("beat %q has no spec direction", beat.ID)
		}
		step := map[string]any{"startMs": startMs, "endMs": endMs}
		if beat.Spec.Check {
			step["check"] = true
		} else {
			step["at"] = beat.Spec.At
		}
		steps = append(steps, step)
	}

	props := map[string]any{
		"title":    in.Plan.Title,
		"goal":     s.Goal,
		"criteria": criteria,
		"steps":    steps,
	}
	if len(s.Constraints) > 0 {
		props["constraints"] = s.Constraints
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneSpec,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props:   props,
	}}, nil
}
