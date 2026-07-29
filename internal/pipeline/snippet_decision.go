package pipeline

// The decision template: one question, and the answer for wherever you land.
//
// This is the shape a good buying guide takes and a bad one never does. The bad
// one is a list of options with pros and cons, which leaves the viewer holding
// exactly the problem they arrived with. The good one finds the ONE question
// that actually separates the choices, puts it on an axis, and gives every
// point on that axis its own answer — so the viewer locates themselves and
// leaves with an instruction rather than a summary.
//
// It is not `gauge`, which measures things against a threshold to ask whether
// they fit. Here nothing is being measured: the axis is a question the *viewer*
// answers about themselves, and the tiers are answers rather than results.
//
// It is not `verdict` either. A verdict is one ruling with conditions attached;
// this is several rulings, one per band, and which one applies is not the
// author's call to make.
//
// The rule that earns it its place is coverage. The tiers must partition the
// axis with no gaps and no overlaps, which is enforced by construction: each
// tier carries an exclusive upper bound, the bounds must strictly increase, and
// the last tier is open-ended. A decision guide with a hole in it is one where
// some fraction of the audience watches to the end and is told nothing, and
// that is precisely the failure the format exists to prevent.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "decision",
		Category:    CatDecisions,
		Since:       SinceV1,
		Title:       "Which one do you buy?",
		Description: "One question on an axis, split into tiers — and the answer for whichever tier you land in.",
		Example:     "Which database should you actually pick for a new project?",
		PromptFile:  snippetDecisionTemplateName,
		NeedsCode:   false,
		// The question, three tiers and the closing rule is five beats.
		MinTargetSec:     40,
		DefaultTargetSec: 60,
		MaxBeats:         8,
		Owns:             beatFields{Decision: true},
		OwnsPlan:         planFields{Decision: true},
		Normalize:        normalizeDecisionPlan,
		Validate:         validateDecisionPlan,
		Scenes:           decisionScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":            strings.Join(DecisionShows(), ", "),
				"Roles":            strings.Join(MetricRoles(), ", "),
				"MinTiers":         minDecisionTiers,
				"MaxTiers":         maxDecisionTiers,
				"MaxQuestionWords": maxDecisionQuestionWords,
				"MaxUnitWords":     maxDecisionUnitWords,
				"MaxBandWords":     maxDecisionBandWords,
				"MaxAnswerWords":   maxDecisionAnswerWords,
				"MaxNoteWords":     maxDecisionNoteWords,
			}
		},
	})
}

const snippetDecisionTemplateName = "snippet_decision.tmpl"

const (
	// Two tiers is a fork and belongs to compare; five bands across the stage
	// leaves each one too narrow to carry a legible answer.
	minDecisionTiers = 2
	maxDecisionTiers = 4

	// The question is set at headline size above the axis.
	maxDecisionQuestionWords = 8
	maxDecisionUnitWords     = 2
	// The band label sits under its segment of the axis — "Under 8GB".
	maxDecisionBandWords = 4
	// The answer takes the frame under the axis, so it is an instruction and
	// not a paragraph.
	maxDecisionAnswerWords = 8
	maxDecisionNoteWords   = 16
)

// decisionShows is the closed vocabulary of what a beat does.
var decisionShows = map[string]bool{
	// Put the question on screen with the axis empty. The first beat.
	"question": true,
	// Land on one tier: the band lights and its answer comes up.
	"tier": true,
	// Every band at once, each with its answer. The closing frame.
	"rule": true,
}

// DecisionShows returns the beat vocabulary sorted.
func DecisionShows() []string {
	out := make([]string, 0, len(decisionShows))
	for k := range decisionShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// DecisionSpec is the one question and the answers along it.
type DecisionSpec struct {
	// Question is what the viewer has to answer about themselves before any
	// recommendation applies — "How big is your model?", "How many writes a
	// second?".
	Question string `json:"question"`
	// Unit is what the axis is measured in — "GB", "users", "$/mo". Empty for
	// an axis whose bands are not numeric.
	Unit string `json:"unit,omitempty"`
	// Tiers partition the axis, in ascending order.
	Tiers []DecisionTier `json:"tiers"`
}

// DecisionTier is one band of the axis and the answer for landing in it.
type DecisionTier struct {
	// UpTo is the exclusive upper bound of this band. Zero on the LAST tier
	// means open-ended, which is what makes the partition total: every possible
	// answer to the question falls in exactly one band.
	UpTo float64 `json:"upTo,omitempty"`
	// Band is how the range reads on screen — "Under 8GB", "8 to 32GB".
	Band string `json:"band"`
	// Answer is the instruction for this band. It takes the frame.
	Answer string `json:"answer"`
	// Note is the one line of reasoning under the answer.
	Note string `json:"note,omitempty"`
	// Role picks the semantic accent, so the bands read as a gradient of
	// consequence rather than as four arbitrary colours. See MetricRoles.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the tier's role, defaulting to neutral.
func (t DecisionTier) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(t.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// DecisionBeat is one move.
type DecisionBeat struct {
	Show string `json:"show"`
	At   int    `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to landing on
// a tier — which is what most beats of this template do.
func (b DecisionBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if decisionShows[s] {
		return s
	}
	return "tier"
}

func normalizeDecisionPlan(p *SnippetPlan) {
	d := p.Decision
	if d == nil {
		return
	}
	d.Question = clampWords(collapseSpaces(d.Question), maxDecisionQuestionWords)
	d.Unit = clampWords(collapseSpaces(d.Unit), maxDecisionUnitWords)

	tiers := make([]DecisionTier, 0, len(d.Tiers))
	for _, t := range d.Tiers {
		t.Band = clampWords(collapseSpaces(t.Band), maxDecisionBandWords)
		t.Answer = clampWords(collapseSpaces(t.Answer), maxDecisionAnswerWords)
		t.Note = clampWords(collapseSpaces(t.Note), maxDecisionNoteWords)
		t.Role = t.ResolvedRole()
		if t.UpTo < 0 {
			t.UpTo = 0
		}
		// A band with no answer is a band that tells its viewer nothing, which
		// is the one thing this template must not ship. Dropping it is the
		// repair; inventing advice would be worse than saying nothing.
		if t.Band != "" && t.Answer != "" && len(tiers) < maxDecisionTiers {
			tiers = append(tiers, t)
		}
	}
	d.Tiers = tiers

	for i := range p.Beats {
		b := p.Beats[i].Decision
		if b == nil {
			continue
		}
		if !decisionShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			if i == 0 {
				b.Show = "question"
			} else {
				b.Show = "tier"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "tier" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(d.Tiers); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateDecisionPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Decision: true}); err != nil {
		return err
	}

	d := p.Decision
	if d == nil {
		return fmt.Errorf("the plan has no decision — this template is one question and the answer for wherever the viewer lands")
	}
	if strings.TrimSpace(d.Question) == "" {
		return fmt.Errorf("there is no question. The whole template hangs on finding the ONE thing that separates the options — write it as a question the viewer answers about themselves")
	}
	if n := len(d.Tiers); n < minDecisionTiers || n > maxDecisionTiers {
		return fmt.Errorf("there are %d tiers, want %d-%d. Two branches is a fork and belongs in compare; five bands across the stage leaves each too narrow to carry a legible answer",
			n, minDecisionTiers, maxDecisionTiers)
	}

	// The rule this template exists for: the bands must partition the axis.
	last := len(d.Tiers) - 1
	prev := 0.0
	for i, t := range d.Tiers {
		if strings.TrimSpace(t.Band) == "" {
			return fmt.Errorf("tier %d has no band label — say what range it covers", i)
		}
		if strings.TrimSpace(t.Answer) == "" {
			return fmt.Errorf("tier %q has no answer. A band that tells its viewer nothing is the exact failure this format exists to prevent — every band ends in an instruction", t.Band)
		}
		if i == last {
			if t.UpTo != 0 {
				return fmt.Errorf("the last tier %q stops at %v, so anyone above that is told nothing. The final band is open-ended — leave upTo out of it",
					t.Band, t.UpTo)
			}
			continue
		}
		if t.UpTo <= 0 {
			return fmt.Errorf("tier %q has no upper bound, but it is not the last one. Every band except the final one ends somewhere, or the axis has a hole in it", t.Band)
		}
		if t.UpTo <= prev {
			return fmt.Errorf("tier %q ends at %v, which is not past the %v the band before it ended at. The tiers have to ascend, or they overlap and a viewer in the overlap gets two different answers",
				t.Band, t.UpTo, prev)
		}
		prev = t.UpTo
	}

	landed := map[int]bool{}
	counts := map[string]int{}
	for i, b := range p.Beats {
		if b.Decision == nil {
			return fmt.Errorf("beat %q has no decision direction — every beat poses the question, lands on a tier, or states the rule", b.ID)
		}
		show := b.Decision.ResolvedShow()
		counts[show]++
		if i == 0 && show != "question" {
			return fmt.Errorf("the clip opens on %q. Pose the question first — a tier lighting up before the viewer knows what is being asked is an answer to nothing", show)
		}
		if show == "rule" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q states the rule but the clip carries on afterwards. The rule is the closing frame", b.ID)
		}
		if show != "tier" {
			continue
		}
		if b.Decision.At < 0 || b.Decision.At >= len(d.Tiers) {
			return fmt.Errorf("beat %q lands on tier %d, which does not exist", b.ID, b.Decision.At)
		}
		if landed[b.Decision.At] {
			return fmt.Errorf("beat %q lands on tier %d again; each band gets one beat", b.ID, b.Decision.At)
		}
		landed[b.Decision.At] = true
	}
	if counts["question"] != 1 {
		return fmt.Errorf("there are %d question beats; the question is posed once", counts["question"])
	}
	if len(landed) != len(d.Tiers) {
		return fmt.Errorf("%d of the %d bands are never spoken. A band drawn but not narrated is an answer some fraction of the audience never hears — give it a beat or merge it",
			len(d.Tiers)-len(landed), len(d.Tiers))
	}
	if counts["rule"] > 1 {
		return fmt.Errorf("there are %d rule beats; the guide closes once", counts["rule"])
	}
	return nil
}

// decisionScenes lays the clip out as ONE scene: the axis is on screen from the
// question onward and the beats only move which band is lit.
func decisionScenes(in SnippetSceneInput) ([]Scene, error) {
	d := in.Plan.Decision
	if d == nil {
		return nil, fmt.Errorf("the plan has no decision")
	}

	// Segment widths. A purely proportional axis would give an open-ended final
	// band no width at all and squeeze a "0-8" band next to a "32-512" one into
	// a sliver — so the bands are drawn evenly and the *labels* carry the
	// arithmetic. The axis here is a sequence of cases, not a measuring stick;
	// `gauge` is the template where distance along the track is the claim.
	tiers := make([]map[string]any, len(d.Tiers))
	for i, t := range d.Tiers {
		tiers[i] = map[string]any{
			"band":   t.Band,
			"answer": t.Answer,
			"note":   t.Note,
			"role":   t.ResolvedRole(),
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Decision == nil {
			return nil, fmt.Errorf("beat %q has no decision direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Decision.ResolvedShow(),
		}
		if beat.Decision.ResolvedShow() == "tier" {
			step["at"] = beat.Decision.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneDecision,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":    in.Plan.Title,
			"question": d.Question,
			"unit":     d.Unit,
			"tiers":    tiers,
			"steps":    steps,
		},
	}}, nil
}
