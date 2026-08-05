package pipeline

// The objective template: what you will be able to DO, and how you will know.
//
// This is the first frame of a lesson, and the catalog had nothing for it. Forty-
// four templates could each explain something; none could say what the next ten
// minutes are for. A lesson that opens by explaining has skipped the only part
// the viewer needs to decide whether to keep watching.
//
// One rule earns it its place, and it is the whole template.
//
// **An outcome names something observable.** "Understand how retries work" is
// not an outcome, it is a feeling, and nobody — the learner least of all — can
// tell whether it happened. "Write a retry that gives up" can be watched. The
// verb is checked against a closed list of things a person visibly does, and
// the error names the belief verbs it refuses, because the failure is never
// carelessness: "understand", "learn about" and "get familiar with" are what
// every syllabus in the world says, so a model writing a syllabus writes them.
//
// The closed list is deliberately short. A longer one would admit "explore" and
// "consider", which are belief verbs wearing an action's clothes.
//
// **Each outcome carries its evidence.** What would you point at to show it
// happened? A file that runs, a number that comes out right, a thing that
// deploys. This is the field that makes the rest honest: an outcome whose
// evidence cannot be named was a belief verb that got past the vocabulary.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "objective",
		Category:    CatPresenting,
		Since:       SinceV5,
		Title:       "What you'll be able to do",
		Description: "The outcome contract a lesson opens on: two to four things the viewer will be able to do afterwards, each with the evidence that would show it. Reach for it as the first clip of a lesson, or of a course.",
		Example:     "Open a lesson on retries and backoff for people who have shipped an API client",
		PromptFile:  snippetObjectiveTemplateName,
		MinTargetSec:     20,
		DefaultTargetSec: 40,
		MaxBeats:         7,
		Owns:             beatFields{Objective: true},
		OwnsPlan:         planFields{Objective: true},
		Normalize:        normalizeObjectivePlan,
		Validate:         validateObjectivePlan,
		Scenes:           objectiveScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				// The emphasis roles every headline picks from.
				"Roles":            strings.Join(MetricRoles(), ", "),
				"Verbs":            strings.Join(ObjectiveVerbs(), ", "),
				"BeliefVerbs":      strings.Join(objectiveBeliefVerbList(), ", "),
				"MinOutcomes":      minObjectiveOutcomes,
				"MaxOutcomes":      maxObjectiveOutcomes,
				"MaxOutcomeWords":  maxObjectiveOutcomeWords,
				"MaxEvidenceWords": maxObjectiveEvidenceWords,
				"MaxAudienceWords": maxObjectiveAudienceWords,
			}
		},
	})
}

const snippetObjectiveTemplateName = "snippet_objective.tmpl"

const (
	// Two is the floor because one outcome is a title, not a contract. Four is
	// the ceiling because a lesson promising five things is a lesson that will
	// deliver two of them.
	minObjectiveOutcomes = 2
	maxObjectiveOutcomes = 4

	maxObjectiveOutcomeWords  = 9
	maxObjectiveEvidenceWords = 12
	maxObjectiveAudienceWords = 14
)

// objectiveVerbs is the closed vocabulary an outcome may open with: things a
// person visibly does. Kept short on purpose — see the file comment.
var objectiveVerbs = map[string]bool{
	"build": true, "write": true, "ship": true, "deploy": true, "debug": true,
	"fix": true, "read": true, "trace": true, "measure": true, "choose": true,
	"configure": true, "test": true, "refactor": true, "run": true, "replace": true,
	"spot": true, "size": true, "wire": true, "cut": true, "prove": true,
}

// objectiveBeliefVerbs are the ones that name a state of mind. Refused by name
// rather than by absence from the list above, so the error can say WHY — a
// model that gets "not in the vocabulary" back writes a synonym; one that is
// told "nobody can watch you understand something" rewrites the outcome.
var objectiveBeliefVerbs = map[string]string{
	"understand": "nobody can watch somebody understand something",
	"learn":      "learning is the lesson's job, not its outcome",
	"know":       "knowing is a state, and states are not observable",
	"grasp":      "same as understand, in a nicer jumper",
	"appreciate": "an outcome nobody has ever failed",
	"explore":    "an activity with no finish line",
	"consider":   "thinking about a thing is not doing it",
	"familiar":   "familiarity is the absence of a claim",
	"comprehend": "same as understand",
	"discover":   "names a feeling about the lesson, not a capability",
}

// ObjectiveVerbs returns the allowed opening verbs, sorted, for the prompt.
func ObjectiveVerbs() []string {
	out := make([]string, 0, len(objectiveVerbs))
	for v := range objectiveVerbs {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func objectiveBeliefVerbList() []string {
	out := make([]string, 0, len(objectiveBeliefVerbs))
	for v := range objectiveBeliefVerbs {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

// ObjectiveSpec is the lesson's contract with the viewer.
type ObjectiveSpec struct {
	// Audience is who this lesson assumes you are. Short — it sets the level,
	// it does not describe a person.
	Audience string `json:"audience,omitempty"`
	// Outcomes are what the viewer will be able to do, in the order the lesson
	// delivers them.
	Outcomes []ObjectiveOutcome `json:"outcomes"`
}

// ObjectiveOutcome is one thing the viewer will be able to do.
type ObjectiveOutcome struct {
	// Action is the outcome itself, opening with an observable verb.
	Action string `json:"action"`
	// Evidence is what you would point at to show it happened.
	Evidence string `json:"evidence"`
}

// Verb returns the outcome's opening word, lowercased and stripped.
func (o ObjectiveOutcome) Verb() string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(o.Action)))
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(fields[0], ".,:;\"'")
}

// ObjectiveBeat says which outcome this beat is on.
type ObjectiveBeat struct {
	// At indexes ObjectiveSpec.Outcomes. A beat that is setting up rather than
	// landing an outcome leaves Show as "frame".
	At int `json:"at,omitempty"`
	// Show is an objectiveShows name.
	Show string `json:"show"`
}

// objectiveShows is the closed vocabulary of what a beat does.
var objectiveShows = map[string]bool{
	// The level this lesson is pitched at — who it assumes you are.
	"frame": true,
	// One outcome, landing.
	"outcome": true,
	// Every outcome at once, at the end: the contract, whole.
	"contract": true,
}

// ObjectiveShows returns the beat vocabulary, sorted, for the prompt.
func ObjectiveShows() []string {
	out := make([]string, 0, len(objectiveShows))
	for s := range objectiveShows {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// ResolvedShow defaults an unknown direction to landing an outcome, which is
// what most beats of this template do.
func (b ObjectiveBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if objectiveShows[s] {
		return s
	}
	return "outcome"
}

func normalizeObjectivePlan(p *SnippetPlan) {
	if p.Objective == nil {
		return
	}
	o := p.Objective
	o.Audience = strings.TrimSpace(o.Audience)
	for i := range o.Outcomes {
		out := &o.Outcomes[i]
		// A model asked for an imperative writes "You will be able to write a
		// retry" as often as "Write a retry". Both are the same outcome and only
		// one of them starts with a verb the vocabulary can see, so the lead-in
		// is stripped rather than argued about — it is a phrasing habit, not a
		// wrong answer, and spending a correction round on it teaches nothing.
		out.Action = strings.TrimSpace(out.Action)
		for _, lead := range []string{
			"you will be able to ", "you'll be able to ", "you will ", "you'll ",
			"be able to ", "the viewer will ", "learners will ",
		} {
			if len(out.Action) > len(lead) && strings.EqualFold(out.Action[:len(lead)], lead) {
				out.Action = strings.TrimSpace(out.Action[len(lead):])
				break
			}
		}
		out.Evidence = strings.TrimSpace(out.Evidence)
	}
	for i := range p.Beats {
		if b := p.Beats[i].Objective; b != nil {
			b.Show = b.ResolvedShow()
			if b.At < 0 || b.At >= len(o.Outcomes) {
				b.At = 0
			}
		}
	}
}

func validateObjectivePlan(p *SnippetPlan) error {
	o := p.Objective
	if o == nil {
		return fmt.Errorf("the plan has no objective — this template opens a lesson by saying what the viewer will be able to do afterwards, so the outcomes are the clip")
	}
	if n := len(o.Outcomes); n < minObjectiveOutcomes || n > maxObjectiveOutcomes {
		return fmt.Errorf("the lesson promises %d outcome(s); this template takes %d to %d. One is a title rather than a contract, and a lesson promising more than %d will deliver two of them",
			n, minObjectiveOutcomes, maxObjectiveOutcomes, maxObjectiveOutcomes)
	}
	if n := len(strings.Fields(o.Audience)); n > maxObjectiveAudienceWords {
		return fmt.Errorf("the audience line is %d words; keep it to %d. It sets the level, it does not describe a person", n, maxObjectiveAudienceWords)
	}

	seen := map[string]bool{}
	for i, out := range o.Outcomes {
		verb := out.Verb()
		if verb == "" {
			return fmt.Errorf("outcome %d has no action", i+1)
		}
		// The belief check comes first, because its error is the useful one.
		for belief, why := range objectiveBeliefVerbs {
			if verb == belief || strings.HasPrefix(verb, belief) {
				return fmt.Errorf("outcome %d opens with %q: %s. An outcome has to name something you could WATCH somebody do, so that the viewer — and the checkpoint clip at the end of this lesson — can tell whether it happened. Rewrite it around one of: %s",
					i+1, verb, why, strings.Join(ObjectiveVerbs(), ", "))
			}
		}
		if !objectiveVerbs[verb] {
			return fmt.Errorf("outcome %d opens with %q, which is not in this template's vocabulary. Use one of: %s. If none of them fits, the outcome is probably a belief in disguise — say what the viewer would DO with the idea",
				i+1, verb, strings.Join(ObjectiveVerbs(), ", "))
		}
		if n := len(strings.Fields(out.Action)); n > maxObjectiveOutcomeWords {
			return fmt.Errorf("outcome %d is %d words; keep it to %d. It goes on screen as a line, not a paragraph", i+1, n, maxObjectiveOutcomeWords)
		}
		if strings.TrimSpace(out.Evidence) == "" {
			return fmt.Errorf("outcome %d (%q) names no evidence. What would you point at to show it happened — a file that runs, a number that comes out right, a thing that deploys? An outcome whose evidence cannot be named is a belief that got past the verb list",
				i+1, out.Action)
		}
		if n := len(strings.Fields(out.Evidence)); n > maxObjectiveEvidenceWords {
			return fmt.Errorf("the evidence for outcome %d is %d words; keep it to %d", i+1, n, maxObjectiveEvidenceWords)
		}
		key := strings.ToLower(out.Action)
		if seen[key] {
			return fmt.Errorf("outcome %d repeats an earlier one (%q). Two outcomes that are the same promise made twice is a contract for one thing", i+1, out.Action)
		}
		seen[key] = true
	}

	// Every outcome has to be landed by some beat, or the clip promised
	// something it never says out loud.
	landed := map[int]bool{}
	for _, b := range p.Beats {
		if b.Objective == nil {
			return fmt.Errorf("beat %q has no objective direction — say whether it frames the level, lands an outcome, or shows the whole contract", b.ID)
		}
		if b.Objective.ResolvedShow() == "outcome" {
			landed[b.Objective.At] = true
		}
	}
	for i, out := range o.Outcomes {
		if !landed[i] {
			return fmt.Errorf("outcome %d (%q) is never landed by a beat. Every promise on the contract has to be said out loud, or it is a line on a slide nobody read",
				i+1, out.Action)
		}
	}
	return nil
}

func objectiveScenes(in SnippetSceneInput) ([]Scene, error) {
	o := in.Plan.Objective
	if o == nil {
		return nil, fmt.Errorf("the plan has no objective")
	}

	outcomes := make([]map[string]any, len(o.Outcomes))
	for i, out := range o.Outcomes {
		outcomes[i] = map[string]any{"action": out.Action, "evidence": out.Evidence}
	}

	// Outcomes accumulate rather than replace: by the closing frame the whole
	// contract is on screen at once, which is the frame somebody screenshots.
	shown := map[int]bool{}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Objective == nil {
			return nil, fmt.Errorf("beat %q has no objective direction", beat.ID)
		}
		show := beat.Objective.ResolvedShow()
		if show == "outcome" {
			shown[beat.Objective.At] = true
		}
		if show == "contract" {
			for j := range o.Outcomes {
				shown[j] = true
			}
		}
		lit := make([]int, 0, len(shown))
		for at := range shown {
			lit = append(lit, at)
		}
		sort.Ints(lit)

		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"lit":     lit,
		}
		if show == "outcome" {
			step["at"] = beat.Objective.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneObjective,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":    in.Plan.Title,
			"audience": o.Audience,
			"outcomes": outcomes,
			"steps":    steps,
		}),
	}}, nil
}
