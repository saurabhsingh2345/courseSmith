package pipeline

// The outcome template: by the end of this lesson, you will be able to.
//
// `objective` already opens a lesson on promises, and it pairs each one with
// evidence — the thing you would point at to show it happened. That is the
// right opener for a practitioner audience. A career-switcher learning CS
// foundations needs the other pairing: each ability with why it matters ON THE
// JOB. "Convert a number to binary by hand" is a promise; "so hex dumps and
// permission bits stop being noise" is the reason to sit through the lesson.
// Payoff, not proof, is what keeps someone in week three of a subject they
// cannot yet use.
//
// The rules are inherited from the observable-verb idea and re-argued here.
//
// **An ability opens with a verb somebody could watch.** "Understand how RAM
// works" is a feeling, and nobody — the learner least of all — can tell when
// it has happened. "Trace a byte from RAM to disk" can be checked by watching.
// The belief verbs are refused BY NAME with a reason each, because a model
// told only "not in the vocabulary" reaches for a synonym, while one told
// "nobody has ever watched someone understand" rewrites the outcome.
//
// **Each ability carries its payoff.** The payoff is what makes the verb rule
// honest: an ability whose job-relevance cannot be stated in ten words was a
// syllabus line, not a skill, and it got past the verb list on a technicality.
//
// **Every ability is landed exactly once, promise first, contract last.** Two
// promises is a title; five is a lesson that will deliver two. The count chip
// filling as the cards stamp in IS the picture, so a card that never stamps —
// or stamps twice — breaks the arithmetic on screen.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "outcome",
		Category:    CatPresenting,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "By the end of this lesson",
		Description: "The lesson opener for a foundations course: two to four observable abilities, each with why it matters on the job. Reach for it as a lesson's first clip when the audience needs a reason to care, not just a contract.",
		Example:     "Open the Memory & Storage lesson: reading a hex dump, tracing a byte from RAM to disk, sizing a cache",
		PromptFile:  snippetOutcomeTemplateName,
		NeedsCode:   false,
		// A promise, two to four cards, and the full ledger: 20s carries the
		// two-ability version, 35s the usual three.
		MinTargetSec:     20,
		DefaultTargetSec: 35,
		MaxBeats:         7,
		Owns:             beatFields{Outcome: true},
		OwnsPlan:         planFields{Outcome: true},
		Normalize:        normalizeOutcomePlan,
		Validate:         validateOutcomePlan,
		Scenes:           outcomeScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":          strings.Join(MetricRoles(), ", "),
				"Shows":          strings.Join(OutcomeShows(), ", "),
				"Verbs":          strings.Join(OutcomeVerbs(), ", "),
				"BeliefVerbs":    strings.Join(outcomeBeliefVerbList(), ", "),
				"MinAbilities":   minOutcomeAbilities,
				"MaxAbilities":   maxOutcomeAbilities,
				"MaxLessonWords": maxOutcomeLessonWords,
				"MaxSkillWords":  maxOutcomeSkillWords,
				"MaxPayoffWords": maxOutcomePayoffWords,
			}
		},
	})
}

const snippetOutcomeTemplateName = "snippet_outcome.tmpl"

const (
	// Two is the floor because one promise is a title, not an opener. Four is
	// the ceiling because a lesson promising five things will deliver two of
	// them, and the viewer knows it.
	minOutcomeAbilities = 2
	maxOutcomeAbilities = 4

	// The lesson name is a chip, not a sentence.
	maxOutcomeLessonWords = 6
	// A skill is one line on a ledger; eight words is where a line becomes a
	// paragraph and the card stops reading at a glance.
	maxOutcomeSkillWords = 8
	// The payoff is a muted sub-line; ten words states a job-relevance, more
	// argues for it, and the argument is the lesson's job.
	maxOutcomePayoffWords = 10
)

// outcomeVerbs is the closed vocabulary a skill may open with: things a person
// visibly does. Deliberately short — a longer list admits "explore" and
// "review", which are belief verbs wearing an action's clothes. Weighted
// toward what a foundations course actually asks for: conversions, traces,
// predictions, and reading the machine's own artifacts.
var outcomeVerbs = map[string]bool{
	"build": true, "write": true, "run": true, "read": true, "trace": true,
	"convert": true, "calculate": true, "predict": true, "explain": true,
	"name": true, "spot": true, "choose": true, "measure": true, "debug": true,
	"fix": true, "draw": true, "compare": true, "test": true, "size": true,
	"decode": true,
}

// outcomeBeliefVerbs name states of mind, refused with the reason so the model
// rewrites the outcome instead of reaching for a synonym.
var outcomeBeliefVerbs = map[string]string{
	"understand": "nobody has ever watched someone understand something",
	"know":       "knowing is a state, and a state cannot be observed happening",
	"learn":      "learning is what the lesson does, not what the viewer walks out able to do",
	"appreciate": "an outcome no learner in history has failed",
	"grasp":      "understand, wearing a different coat",
	"comprehend": "understand again, with more syllables",
	"familiar":   "familiarity claims nothing anyone could check",
	"explore":    "an activity with no finish line is not an ability",
	"absorb":     "describes the lesson washing over someone, not a capability",
	"master":     "a boast about degree, hiding that the doing was never named",
}

// OutcomeVerbs returns the allowed opening verbs, sorted, for the prompt.
func OutcomeVerbs() []string {
	out := make([]string, 0, len(outcomeVerbs))
	for v := range outcomeVerbs {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func outcomeBeliefVerbList() []string {
	out := make([]string, 0, len(outcomeBeliefVerbs))
	for v := range outcomeBeliefVerbs {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

// OutcomeSpec is the lesson's opening promise.
type OutcomeSpec struct {
	// Lesson names the lesson the promise belongs to — "Memory & Storage".
	Lesson string `json:"lesson,omitempty"`
	// Abilities are what the viewer will be able to do, in the order the
	// lesson delivers them.
	Abilities []OutcomeAbility `json:"abilities"`
}

// OutcomeAbility is one promise with its reason.
type OutcomeAbility struct {
	// Skill is the ability itself, opening with an observable verb.
	Skill string `json:"skill"`
	// Payoff is why this matters on the job.
	Payoff string `json:"payoff"`
}

// Verb returns the skill's opening word, lowercased and stripped.
func (a OutcomeAbility) Verb() string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(a.Skill)))
	if len(fields) == 0 {
		return ""
	}
	return strings.Trim(fields[0], ".,:;\"'")
}

// OutcomeBeat is one move: which ability this beat lands.
type OutcomeBeat struct {
	// Show is an outcomeShows name.
	Show string `json:"show"`
	// At indexes OutcomeSpec.Abilities, for an "ability" beat.
	At int `json:"at,omitempty"`
}

// outcomeShows is the closed vocabulary of what a beat does.
var outcomeShows = map[string]bool{
	// The lesson name and the count of promises. The opener.
	"promise": true,
	// One ability card stamps in with its payoff.
	"ability": true,
	// Every card lit at once: the whole ledger. The closer.
	"contract": true,
}

// OutcomeShows returns the beat vocabulary sorted, for the prompt.
func OutcomeShows() []string {
	out := make([]string, 0, len(outcomeShows))
	for k := range outcomeShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ResolvedShow defaults an unknown direction to landing an ability, which is
// what most beats of this template do.
func (b OutcomeBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if outcomeShows[s] {
		return s
	}
	return "ability"
}

func normalizeOutcomePlan(p *SnippetPlan) {
	o := p.Outcome
	if o == nil {
		return
	}
	o.Lesson = clampWords(collapseSpaces(o.Lesson), maxOutcomeLessonWords)
	abilities := make([]OutcomeAbility, 0, len(o.Abilities))
	for _, a := range o.Abilities {
		// The lead-in strip from objective's lesson: "You will be able to
		// trace..." is the same outcome as "Trace...", and only one of them
		// opens with a verb the vocabulary can see. A phrasing habit, not a
		// wrong answer — repair it rather than spend a correction round.
		a.Skill = collapseSpaces(a.Skill)
		for _, lead := range []string{
			"you will be able to ", "you'll be able to ", "you will ", "you'll ",
			"be able to ", "the viewer will ", "learners will ",
		} {
			if len(a.Skill) > len(lead) && strings.EqualFold(a.Skill[:len(lead)], lead) {
				a.Skill = strings.TrimSpace(a.Skill[len(lead):])
				break
			}
		}
		a.Skill = clampWords(a.Skill, maxOutcomeSkillWords)
		a.Payoff = clampWords(collapseSpaces(a.Payoff), maxOutcomePayoffWords)
		if len(abilities) < maxOutcomeAbilities {
			abilities = append(abilities, a)
		}
	}
	o.Abilities = abilities

	for i := range p.Beats {
		b := p.Beats[i].Outcome
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "ability" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(o.Abilities); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateOutcomePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Outcome: true}); err != nil {
		return err
	}

	o := p.Outcome
	if o == nil {
		return fmt.Errorf("the plan has no outcome — this template opens a lesson on what the viewer will be able to do and why that matters on the job, so the abilities are the clip")
	}
	if strings.TrimSpace(o.Lesson) == "" {
		return fmt.Errorf("the plan does not name the lesson. The opening beat is the lesson's name over the count of what it promises, so without one the clip opens on a number attached to nothing — and this template's whole job is to say what THIS lesson is for")
	}
	if n := len(o.Abilities); n < minOutcomeAbilities || n > maxOutcomeAbilities {
		return fmt.Errorf("the lesson promises %d abilit(ies), want %d-%d. One is a title rather than an opener, and a lesson promising more than %d will deliver two of them",
			n, minOutcomeAbilities, maxOutcomeAbilities, maxOutcomeAbilities)
	}

	seen := map[string]bool{}
	for i, a := range o.Abilities {
		verb := a.Verb()
		if verb == "" {
			return fmt.Errorf("ability %d names no skill", i+1)
		}
		// The belief check first, because its error is the one that teaches.
		//
		// Matched whole, not by prefix. A prefix test reads "understand" inside
		// any longer word that happens to start the same way, so the day this
		// vocabulary grows a legitimate verb sharing an opening with a refused
		// one, that verb is rejected and told it is a belief — an error whose
		// reason is simply untrue, about a word the template itself allows.
		for belief, why := range outcomeBeliefVerbs {
			if verb == belief {
				return fmt.Errorf("ability %d opens with %q: %s. An ability has to name something a person could be WATCHED doing, or neither the viewer nor the drill at the end of the lesson can tell whether it arrived. Rewrite it around one of: %s",
					i+1, verb, why, strings.Join(OutcomeVerbs(), ", "))
			}
		}
		if !outcomeVerbs[verb] {
			return fmt.Errorf("ability %d opens with %q, which is not in this template's vocabulary. Use one of: %s. If none fits, the ability is probably a belief in disguise — say what the viewer would DO with the idea",
				i+1, verb, strings.Join(OutcomeVerbs(), ", "))
		}
		if strings.TrimSpace(a.Payoff) == "" {
			return fmt.Errorf("ability %d (%q) has no payoff. The payoff is why this matters on the job — hex dumps stop being noise, the bug report starts making sense. An ability whose job-relevance cannot be stated was a syllabus line, not a skill",
				i+1, a.Skill)
		}
		key := strings.ToLower(a.Skill)
		if seen[key] {
			return fmt.Errorf("ability %d repeats an earlier one (%q). The same promise made twice is a ledger for one thing fewer than the count chip claims", i+1, a.Skill)
		}
		seen[key] = true
	}

	if p.Beats[0].Outcome == nil || p.Beats[0].Outcome.ResolvedShow() != "promise" {
		return fmt.Errorf("beat %q does not open on the promise. The lesson name and the count come first — cards stamping onto a ledger nobody has been promised is a list with no reason to watch it fill",
			p.Beats[0].ID)
	}

	landed := map[int]bool{}
	for i, b := range p.Beats {
		if b.Outcome == nil {
			return fmt.Errorf("beat %q has no outcome direction — every beat makes the promise, lands an ability, or shows the whole contract", b.ID)
		}
		switch b.Outcome.ResolvedShow() {
		case "promise":
			if i != 0 {
				return fmt.Errorf("beat %q makes the promise again part-way through. The promise is the opener; repeating it resets a ledger that was mid-fill", b.ID)
			}
		case "ability":
			at := b.Outcome.At
			if at < 0 || at >= len(o.Abilities) {
				return fmt.Errorf("beat %q lands ability %d, which does not exist", b.ID, at)
			}
			if landed[at] {
				return fmt.Errorf("beat %q lands ability %d (%q) a second time. A card stamps once; stamping it again breaks the count the chip is filling on screen",
					b.ID, at, o.Abilities[at].Skill)
			}
			landed[at] = true
		case "contract":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q shows the whole contract before the clip ends. The full ledger is the closer — showing it early leaves the remaining cards stamping onto a picture that already claimed to be complete", b.ID)
			}
		}
	}
	if p.Beats[len(p.Beats)-1].Outcome.ResolvedShow() != "contract" {
		return fmt.Errorf("the clip does not close on the contract. The last beat lights every card at once — that frame, the whole promise visible, is the one a viewer screenshots")
	}
	for i, a := range o.Abilities {
		if !landed[i] {
			return fmt.Errorf("ability %d (%q) is never landed by a beat. Every promise on the ledger has to be said out loud, or the count chip fills past what the narration delivered",
				i+1, a.Skill)
		}
	}
	return nil
}

// outcomeScenes lays the clip out as ONE scene. The ledger accumulates: by the
// closer every card is lit, which is the frame that gets screenshotted.
func outcomeScenes(in SnippetSceneInput) ([]Scene, error) {
	o := in.Plan.Outcome
	if o == nil {
		return nil, fmt.Errorf("the plan has no outcome")
	}

	abilities := make([]map[string]any, len(o.Abilities))
	for i, a := range o.Abilities {
		abilities[i] = map[string]any{"skill": a.Skill, "payoff": a.Payoff}
	}

	lit := map[int]bool{}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Outcome == nil {
			return nil, fmt.Errorf("beat %q has no outcome direction", beat.ID)
		}
		show := beat.Outcome.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "ability" {
			lit[beat.Outcome.At] = true
			step["at"] = beat.Outcome.At
		}
		if show == "contract" {
			for j := range o.Abilities {
				lit[j] = true
			}
		}
		on := make([]int, 0, len(lit))
		for at := range lit {
			on = append(on, at)
		}
		sort.Ints(on)
		step["lit"] = on
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneOutcome,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":     in.Plan.Title,
			"lesson":    o.Lesson,
			"abilities": abilities,
			"steps":     steps,
		}),
	}}, nil
}
