package pipeline

// The toggle template: the answer first, and then the reasons it is not simple.
//
// This is `verdict` inverted, and the inversion is the whole template. Verdict
// spends a clip building an argument and lands the ruling on the closing frame —
// its validator requires the call to be last, because a recommendation given
// before its grounds is a recommendation nobody trusts. That is right for a
// buying guide and wrong for the other kind of clip entirely: the one answering
// a question the viewer arrived already asking.
//
// "Is WebAssembly replacing JavaScript?" has a one-word answer, and a clip that
// makes somebody wait four minutes for it has mistaken suspense for teaching.
// The reference clips answer in the first three seconds — a switch flicking from
// NO to YES — and then spend the rest of the runtime earning it. That shape is
// the inverted pyramid, it is how every news desk has opened for a century, and
// the catalog had no way to draw it.
//
// Three rules earn it its place, and all three are validators.
//
// **The answer lands in the first beat.** Not the second, not after a setup. If
// the clip is worth this template the viewer already has the question, and
// withholding is the failure mode the template exists to prevent. The error says
// so and names `verdict` for the other shape.
//
// **The answer must be a real turn, not a restatement.** A toggle whose two
// positions say the same thing in different words — "no" and "not really" — has
// drawn a switch that does not switch. The two labels have to differ.
//
// **There is at least one qualifier.** A bare answer is a tweet. What makes the
// form work is that the clip immediately starts complicating what it just said,
// and a toggle with nothing after it has used a switch as a title card.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:             "toggle",
		Category:         CatDecisions,
		Since:            SinceV4,
		Family:           FamilyReplica,
		Title:            "The answer, then the asterisks",
		Description:      "A question answered in the first three seconds by a switch, and then complicated. Reach for it when the viewer arrived already asking and making them wait is the only thing that could go wrong.",
		Example:          "Is WebAssembly actually replacing JavaScript?",
		PromptFile:       snippetToggleTemplateName,
		NeedsCode:        false,
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		MaxBeats:         8,
		Owns:             beatFields{Toggle: true},
		OwnsPlan:         planFields{Toggle: true},
		Normalize:        normalizeTogglePlan,
		Validate:         validateTogglePlan,
		Scenes:           toggleScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":          strings.Join(MetricRoles(), ", "),
				"Shows":          strings.Join(ToggleShows(), ", "),
				"MinQualifiers":  minToggleQualifiers,
				"MaxQualifiers":  maxToggleQualifiers,
				"MaxAnswerWords": maxToggleAnswerWords,
				"MaxLabelWords":  maxToggleLabelWords,
				"MaxNoteWords":   maxToggleNoteWords,
			}
		},
	})
}

const snippetToggleTemplateName = "snippet_toggle.tmpl"

const (
	// One qualifier is the minimum that stops this being a title card; four is
	// as many asterisks as an answer can carry before the answer stops being one.
	minToggleQualifiers = 1
	maxToggleQualifiers = 4

	// The answer is a word or two. "No" and "Yes, for anything CPU-bound" are
	// both fine; a sentence is not an answer, it is the qualifier.
	maxToggleAnswerWords = 5
	maxToggleLabelWords  = 5
	maxToggleNoteWords   = 18
)

// toggleShows is the closed vocabulary of what a beat does.
var toggleShows = map[string]bool{
	// The question, and the switch flicking to the answer. The first beat.
	"answer": true,
	// One qualifier: the ground the answer does not cover.
	"qualify": true,
	// The answer again, with everything it now carries.
	"settle": true,
}

// ToggleShows returns the beat vocabulary sorted.
func ToggleShows() []string {
	out := make([]string, 0, len(toggleShows))
	for k := range toggleShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ToggleSpec is the question, the answer, and the asterisks. On the plan because
// the switch is one object the whole clip is arguing with.
type ToggleSpec struct {
	// Question is what the viewer arrived asking. Stated, not implied.
	Question string `json:"question"`
	// From is where the switch starts — the answer people assume, or the naive
	// one. To is where it lands: the real answer.
	From string `json:"from"`
	To   string `json:"to"`
	// Qualifiers are what the answer does not cover, in the order they are made.
	Qualifiers []ToggleQualifier `json:"qualifiers"`
}

// ToggleQualifier is one asterisk on the answer.
type ToggleQualifier struct {
	// Label is the condition — "in the browser", "for CPU-bound work".
	Label string `json:"label"`
	// Note is what changes under that condition. One sentence.
	Note string `json:"note,omitempty"`
	// Role is what this qualifier is doing: a metricRoles name.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the qualifier's role, defaulting to neutral.
func (q ToggleQualifier) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(q.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// ToggleBeat is one move.
type ToggleBeat struct {
	// Show is a toggleShows name.
	Show string `json:"show"`
	// At indexes ToggleSpec.Qualifiers, for a "qualify" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to a qualifier
// — which is what most beats of this template do.
func (b ToggleBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if toggleShows[s] {
		return s
	}
	return "qualify"
}

func normalizeTogglePlan(p *SnippetPlan) {
	t := p.Toggle
	if t == nil {
		return
	}
	t.Question = collapseSpaces(t.Question)
	t.From = clampWords(collapseSpaces(t.From), maxToggleAnswerWords)
	t.To = clampWords(collapseSpaces(t.To), maxToggleAnswerWords)

	qs := make([]ToggleQualifier, 0, len(t.Qualifiers))
	for _, q := range t.Qualifiers {
		q.Label = clampWords(collapseSpaces(q.Label), maxToggleLabelWords)
		q.Note = clampWords(collapseSpaces(q.Note), maxToggleNoteWords)
		q.Role = q.ResolvedRole()
		if q.Label != "" && len(qs) < maxToggleQualifiers {
			qs = append(qs, q)
		}
	}
	t.Qualifiers = qs

	for i := range p.Beats {
		b := p.Beats[i].Toggle
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "qualify" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(t.Qualifiers); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateTogglePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Toggle: true}); err != nil {
		return err
	}

	t := p.Toggle
	if t == nil {
		return fmt.Errorf("the plan has no question — this template is one question answered immediately and then complicated")
	}
	if strings.TrimSpace(t.Question) == "" {
		return fmt.Errorf("there is no question. The whole form depends on the viewer having arrived with one, so it has to be on screen in their words rather than implied by the answer")
	}
	if strings.TrimSpace(t.From) == "" || strings.TrimSpace(t.To) == "" {
		return fmt.Errorf("the switch needs both positions: where it starts and where it lands")
	}
	// A toggle whose two positions say the same thing has drawn a switch that
	// does not switch.
	if phraseKey(t.From) == phraseKey(t.To) {
		return fmt.Errorf("the switch goes from %q to %q, which is the same answer twice. The picture is a position CHANGING — if the honest answer is the one everybody already assumed, there is nothing to flick and the verdict template states it better",
			t.From, t.To)
	}
	if n := len(t.Qualifiers); n < minToggleQualifiers || n > maxToggleQualifiers {
		return fmt.Errorf("there are %d qualifiers, want %d-%d. A bare answer is a tweet: what makes this form work is that the clip starts complicating what it just said, and one with nothing after it has used a switch as a title card. Past four the answer stops being one",
			n, minToggleQualifiers, maxToggleQualifiers)
	}
	seen := map[string]bool{}
	for i, q := range t.Qualifiers {
		if strings.TrimSpace(q.Label) == "" {
			return fmt.Errorf("qualifier %d has no label — say which ground the answer does not cover", i)
		}
		key := strings.ToLower(strings.TrimSpace(q.Label))
		if seen[key] {
			return fmt.Errorf("two qualifiers are both %q — each asterisk is a different condition", q.Label)
		}
		seen[key] = true
		if r := strings.ToLower(strings.TrimSpace(q.Role)); r != "" && !metricRoles[r] {
			return fmt.Errorf("qualifier %d has role %q, which is not one of: %s", i, q.Role, strings.Join(MetricRoles(), ", "))
		}
	}

	// The rule the template exists for.
	if p.Beats[0].Toggle == nil || p.Beats[0].Toggle.ResolvedShow() != "answer" {
		return fmt.Errorf("beat %q does not answer the question. This template is the inverted pyramid: if the clip is worth it the viewer already arrived asking, and making them wait is the only thing that could go wrong. If the answer genuinely has to be earned first, that is the verdict template",
			p.Beats[0].ID)
	}

	qualified := map[int]bool{}
	answers := 0
	for i, b := range p.Beats {
		if b.Toggle == nil {
			return fmt.Errorf("beat %q has no toggle direction — every beat answers, qualifies, or settles", b.ID)
		}
		switch b.Toggle.ResolvedShow() {
		case "answer":
			answers++
			if i != 0 {
				return fmt.Errorf("beat %q flicks the switch again part-way through. It moves once, at the start — a switch that keeps moving is a clip that has not decided", b.ID)
			}
		case "qualify":
			at := b.Toggle.At
			if at < 0 || at >= len(t.Qualifiers) {
				return fmt.Errorf("beat %q raises qualifier %d, which does not exist", b.ID, at)
			}
			if qualified[at] {
				return fmt.Errorf("beat %q raises qualifier %d again; each asterisk gets one beat", b.ID, at)
			}
			qualified[at] = true
		case "settle":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q settles the answer but the clip carries on afterwards. Settling is the closing frame: the same answer, now carrying everything the clip has added to it", b.ID)
			}
		}
	}
	if answers != 1 {
		return fmt.Errorf("there are %d beats answering, want exactly 1", answers)
	}
	if len(qualified) != len(t.Qualifiers) {
		return fmt.Errorf("%d of the %d qualifiers are never raised. An asterisk the narrator skips is one nobody heard — give it a beat or cut it",
			len(t.Qualifiers)-len(qualified), len(t.Qualifiers))
	}
	return nil
}

// toggleScenes lays the clip out as ONE scene. The switch is thrown once and the
// qualifiers accumulate under it.
func toggleScenes(in SnippetSceneInput) ([]Scene, error) {
	t := in.Plan.Toggle
	if t == nil {
		return nil, fmt.Errorf("the plan has no question")
	}

	qs := make([]map[string]any, len(t.Qualifiers))
	for i, q := range t.Qualifiers {
		qs[i] = map[string]any{
			"label": q.Label,
			"note":  q.Note,
			"role":  q.ResolvedRole(),
		}
	}

	raised := map[int]bool{}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Toggle == nil {
			return nil, fmt.Errorf("beat %q has no toggle direction", beat.ID)
		}
		show := beat.Toggle.ResolvedShow()
		if show == "qualify" {
			raised[beat.Toggle.At] = true
		}
		shown := make([]int, 0, len(raised))
		for at := range raised {
			shown = append(shown, at)
		}
		sort.Ints(shown)

		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"raised":  shown,
		}
		if show == "qualify" {
			step["at"] = beat.Toggle.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneToggle,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":      in.Plan.Title,
			"question":   t.Question,
			"from":       t.From,
			"to":         t.To,
			"qualifiers": qs,
			"steps":      steps,
		}),
	}}, nil
}
