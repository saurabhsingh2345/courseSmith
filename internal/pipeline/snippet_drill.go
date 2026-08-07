package pipeline

// The drill template: one sharp question, and the wrong answers dying one at
// a time.
//
// The catalog already has a quiz card, and its shape is the problem this one
// exists to fix: question up, pause, answer lights, done. What that skips is
// the part where learning happens — the moment a plausible wrong answer is
// looked at squarely and struck. For CS foundations the wrong answers ARE the
// misconceptions ("floats are imprecise because computers round randomly"),
// and a card that never engages them leaves them alive. So this drill
// eliminates one option at a time, each strike a beat of narration about WHY
// that one is wrong, before the survivor lights.
//
// The rules keep the elimination honest.
//
// **The clip opens on the full board.** Striking an option the viewer has not
// read is theatre; the ask beat puts the question and every option up first,
// and it is not repeated — re-asking mid-clip un-strikes the board.
//
// **An eliminate may not strike the answer.** Obvious, and exactly the index
// arithmetic a model gets wrong: it shuffles options between drafts and the
// answer index goes stale. Checked in Go because on screen this mistake is
// the clip telling the viewer the right answer is wrong.
//
// **No option is eliminated twice.** A struck plate stays struck; a second
// strike is a beat spent on a corpse.
//
// **Exactly one reveal, and the why comes after it.** Two reveals is two
// right answers. A why before the reveal explains a choice the viewer has not
// seen made — the reason attaches to the answer, so the answer lands first.
//
// Three or four options is the band: two is a coin flip, and five means at
// least one distractor nobody would pick, which teaches nothing when struck.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "drill",
		Category:    CatPresenting,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "Quick check",
		Description: "One sharp question with three or four options, the wrong ones struck one at a time, then the answer and a one-line why. Reach for it after a concept lands, when the wrong answers are the misconceptions worth killing.",
		Example:     "Why does 0.1 + 0.2 not equal 0.3 in floating point?",
		PromptFile:  snippetDrillTemplateName,
		NeedsCode:   false,
		// Ask, a strike or three, the reveal, the why. Below ~25s the strikes
		// have no narration to explain themselves with.
		MinTargetSec:     25,
		DefaultTargetSec: 40,
		// List-shaped: the ask, up to six acting beats, and the close.
		MaxBeats:  8,
		Owns:      beatFields{Drill: true},
		OwnsPlan:  planFields{Drill: true},
		Normalize: normalizeDrillPlan,
		Validate:  validateDrillPlan,
		Scenes:    drillScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":            strings.Join(MetricRoles(), ", "),
				"Shows":            strings.Join(DrillShows(), ", "),
				"MinOptions":       minDrillOptions,
				"MaxOptions":       maxDrillOptions,
				"MaxQuestionWords": maxDrillQuestionWords,
				"MaxOptionWords":   maxDrillOptionWords,
				"MaxWhyWords":      maxDrillWhyWords,
			}
		},
	})
}

const snippetDrillTemplateName = "snippet_drill.tmpl"

const (
	// Two options is a coin flip; five means at least one distractor nobody
	// would pick, which teaches nothing when struck.
	minDrillOptions = 3
	maxDrillOptions = 4

	// The question is set large; sixteen words is where large type wraps past
	// two lines and stops being a question at a glance.
	maxDrillQuestionWords = 16
	// An option is one lettered plate; eight words keeps four plates on one
	// screen without shrinking the type.
	maxDrillOptionWords = 8
	// The why is one line under the winning plate, not a second lesson.
	maxDrillWhyWords = 16
)

// drillShows is the closed vocabulary of what a beat does.
var drillShows = map[string]bool{
	// The question and every option, whole. The opener.
	"ask": true,
	// Option At grays with a strike through it.
	"eliminate": true,
	// The answer plate lights and lifts.
	"reveal": true,
	// The one-line reason lands under the answer. The closer.
	"why": true,
}

// DrillShows returns the beat vocabulary sorted, for the prompt.
func DrillShows() []string {
	out := make([]string, 0, len(drillShows))
	for k := range drillShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// DrillSpec is the check-question: the board and its resolution.
type DrillSpec struct {
	// Question is the check itself.
	Question string `json:"question"`
	// Options are the lettered plates, top to bottom.
	Options []string `json:"options"`
	// Answer indexes Options: the plate that survives.
	Answer int `json:"answer"`
	// Why is the one-line reason the answer is right.
	Why string `json:"why"`
}

// DrillBeat is one move on the board.
type DrillBeat struct {
	// Show is a drillShows name.
	Show string `json:"show"`
	// At indexes DrillSpec.Options, for an "eliminate" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults an unknown direction to an elimination, which is what
// most beats of this template do.
func (b DrillBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if drillShows[s] {
		return s
	}
	return "eliminate"
}

func normalizeDrillPlan(p *SnippetPlan) {
	d := p.Drill
	if d == nil {
		return
	}
	d.Question = clampWords(collapseSpaces(d.Question), maxDrillQuestionWords)
	d.Why = clampWords(collapseSpaces(d.Why), maxDrillWhyWords)
	options := make([]string, 0, len(d.Options))
	for _, opt := range d.Options {
		opt = clampWords(collapseSpaces(opt), maxDrillOptionWords)
		if opt == "" {
			continue
		}
		if len(options) < maxDrillOptions {
			options = append(options, opt)
		}
	}
	d.Options = options
	// Answer is NOT clamped: moving it to a different option would silently
	// change which plate the clip calls correct, and that is a content error
	// only the validator's message can fix.

	for i := range p.Beats {
		b := p.Beats[i].Drill
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "eliminate" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(d.Options); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateDrillPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Drill: true}); err != nil {
		return err
	}

	d := p.Drill
	if d == nil {
		return fmt.Errorf("the plan has no drill — this template is one check-question and its board, so the question, options and answer are the clip")
	}
	if strings.TrimSpace(d.Question) == "" {
		return fmt.Errorf("the drill has no question. The board is options to a question; without one the plates are answers to nothing")
	}
	if n := len(d.Options); n < minDrillOptions || n > maxDrillOptions {
		return fmt.Errorf("the board has %d option(s), want %d-%d. Two is a coin flip, and past four at least one distractor is one nobody would pick — striking it teaches nothing",
			n, minDrillOptions, maxDrillOptions)
	}
	if d.Answer < 0 || d.Answer >= len(d.Options) {
		return fmt.Errorf("the answer is option %d, but the board has options 0-%d. The reveal has to light a plate that exists",
			d.Answer, len(d.Options)-1)
	}

	if p.Beats[0].Drill == nil || p.Beats[0].Drill.ResolvedShow() != "ask" {
		return fmt.Errorf("beat %q does not open on the ask. Striking an option the viewer has not read is theatre — put the question and the whole board up first",
			p.Beats[0].ID)
	}

	struck := map[int]bool{}
	reveals := 0
	whys := 0
	for i, b := range p.Beats {
		if b.Drill == nil {
			return fmt.Errorf("beat %q has no drill direction — every beat asks, eliminates an option, reveals the answer, or lands the why", b.ID)
		}
		switch b.Drill.ResolvedShow() {
		case "ask":
			if i != 0 {
				return fmt.Errorf("beat %q asks the question again part-way through. The ask is the opener; re-asking un-strikes a board the viewer watched get struck", b.ID)
			}
		case "eliminate":
			at := b.Drill.At
			if at < 0 || at >= len(d.Options) {
				return fmt.Errorf("beat %q eliminates option %d, which does not exist", b.ID, at)
			}
			if at == d.Answer {
				return fmt.Errorf("beat %q strikes option %d (%q), which is the ANSWER. On screen this is the clip telling the viewer the right answer is wrong — if the answer moved between drafts, update `answer` to match the options",
					b.ID, at, d.Options[at])
			}
			if struck[at] {
				return fmt.Errorf("beat %q eliminates option %d (%q) a second time. A struck plate stays struck; the second strike is a beat spent on a corpse",
					b.ID, at, d.Options[at])
			}
			struck[at] = true
		case "reveal":
			reveals++
		case "why":
			whys++
			if reveals == 0 {
				return fmt.Errorf("beat %q lands the why before any answer is revealed. The reason attaches to the answer, so the reveal comes first — a why on an unresolved board explains a choice the viewer has not seen made",
					b.ID)
			}
		}
	}
	if reveals != 1 {
		return fmt.Errorf("there are %d reveal beats, want exactly 1. One plate survives; two reveals is two right answers and zero is a question the clip never answers", reveals)
	}
	if whys > 1 {
		return fmt.Errorf("there are %d why beats; the reason is one line and it lands once", whys)
	}
	if whys == 1 && strings.TrimSpace(d.Why) == "" {
		return fmt.Errorf("a beat lands the why, but the drill's why line is empty. The one-line reason is what turns a reveal into teaching — write it, or drop the beat")
	}
	return nil
}

// drillScenes lays the clip out as ONE scene. The board persists; the steps
// only say which plates are struck and whether the answer is lit.
func drillScenes(in SnippetSceneInput) ([]Scene, error) {
	d := in.Plan.Drill
	if d == nil {
		return nil, fmt.Errorf("the plan has no drill")
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	struck := map[int]bool{}
	revealed := false
	whyOn := false
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Drill == nil {
			return nil, fmt.Errorf("beat %q has no drill direction", beat.ID)
		}
		show := beat.Drill.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "eliminate":
			struck[beat.Drill.At] = true
			step["at"] = beat.Drill.At
		case "reveal":
			revealed = true
		case "why":
			whyOn = true
		}
		strikes := make([]int, 0, len(struck))
		for at := range struck {
			strikes = append(strikes, at)
		}
		sort.Ints(strikes)
		step["struck"] = strikes
		step["revealed"] = revealed
		step["whyOn"] = whyOn
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneDrill,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":    in.Plan.Title,
			"question": d.Question,
			"options":  append([]string{}, d.Options...),
			"answer":   d.Answer,
			"why":      d.Why,
			"steps":    steps,
		}),
	}}, nil
}
