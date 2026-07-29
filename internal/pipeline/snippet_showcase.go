package pipeline

// The showcase template: introducing a tool, and handing off to its demo.
//
// A course that teaches tools has to introduce them, and doing it badly is the
// default. The bad version is a feature list: everything the product can do, in
// the order the marketing site put it, with no answer to the only question a
// learner actually has, which is whether to use this one. The catalog could
// already draw a comparison (`compare`) and a set of tiers (`stack`), and
// neither of those is an introduction — you cannot compare a thing you have not
// met, and a tier tells you where a tool sits, not what it is like to hold.
//
// So the card is organised around what people decide on. Use case first, then
// what it costs, then what happens if you want to leave, then how long before
// you are productive — and then, in two columns, what it is good at and what to
// watch out for.
//
// The second column is enforced, and it is the rule that earns this template its
// place. A showcase with no honest limitation is an advertisement, a course full
// of advertisements teaches nobody which tool to pick, and the model will not
// write the awkward half unless it is required to. This is the one place in the
// catalog where a validator is defending the *viewer* rather than the layout.
//
// The clip ends on a hand-off plate rather than on the card: the card recedes and
// a framed play glyph takes the screen. That is a designed out-point — a
// repeatable frame to cut a screen recording onto — and no other template has
// one. It is the difference between a clip that ends and a clip that hands over.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "showcase",
		Category:    CatDecisions,
		Title:       "Tool showcase",
		Description: "A product card — what it is, what it costs, what it is good and bad at. Reach for it the first time a course meets a tool.",
		Example:     "Introducing Airtable: what it is, what it costs, and when not to use it",
		PromptFile:  snippetShowcaseTemplateName,
		NeedsCode:   false,
		// Intro, three facts, the two columns and the hand-off is seven beats
		// before anything optional, so this cannot be a short clip. Four facts
		// makes it eight, which is why the ceiling goes up.
		MinTargetSec:     70,
		DefaultTargetSec: 95,
		MaxBeats:         10,
		Owns:             beatFields{Showcase: true},
		OwnsPlan:         planFields{Showcase: true},
		Normalize:        normalizeShowcasePlan,
		Validate:         validateShowcasePlan,
		Scenes:           showcaseScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":           strings.Join(ShowcaseShows(), ", "),
				"Icons":           strings.Join(PointIconNames(), ", "),
				"MinFacts":        minShowcaseFacts,
				"MaxFacts":        maxShowcaseFacts,
				"MinStrengths":    minShowcaseStrengths,
				"MaxStrengths":    maxShowcaseStrengths,
				"MinLimits":       minShowcaseLimits,
				"MaxLimits":       maxShowcaseLimits,
				"MaxNameWords":    maxShowcaseNameWords,
				"MaxTaglineWords": maxShowcaseTaglineWords,
				"MaxLabelWords":   maxShowcaseLabelWords,
				"MaxValueWords":   maxShowcaseValueWords,
				"MaxPointWords":   maxShowcasePointWords,
			}
		},
	})
}

const snippetShowcaseTemplateName = "snippet_showcase.tmpl"

// Card capacity. Four fact cells across the 1700px stage leaves 425px each,
// which is what a six-word value needs at 26px; a fifth cell turns the row into
// a strip nobody reads. Three is the floor because a tool you can describe in
// two facts is one the illustration template introduces in half the runtime.
const (
	minShowcaseFacts = 3
	maxShowcaseFacts = 4

	minShowcaseStrengths = 2
	maxShowcaseStrengths = 4
	// One honest limitation is the floor, and it is the whole point.
	minShowcaseLimits = 1
	maxShowcaseLimits = 3

	maxShowcaseNameWords    = 3
	maxShowcaseTaglineWords = 12
	maxShowcaseLabelWords   = 3
	maxShowcaseValueWords   = 6
	maxShowcasePointWords   = 8
)

// showcaseShows is the closed vocabulary of what a beat does to the card.
var showcaseShows = map[string]bool{
	"intro":     true,
	"fact":      true,
	"strengths": true,
	"limits":    true,
	"handoff":   true,
}

// ShowcaseShows returns the vocabulary sorted, for prompts and docs.
func ShowcaseShows() []string {
	out := make([]string, 0, len(showcaseShows))
	for k := range showcaseShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ShowcaseSpec is the tool being introduced. On the plan rather than per-beat
// because the card is the subject of the clip and the beats only light parts of
// it.
type ShowcaseSpec struct {
	// Name is the product.
	Name string `json:"name"`
	// Category is what kind of thing it is — "Database", "Test runner". It sits
	// in a chip beside the name.
	Category string `json:"category,omitempty"`
	// Tagline is what it is, in one line, in the words someone who has never
	// heard of it would use.
	Tagline string `json:"tagline"`
	// Icon is a PointIconNames name drawn in the product tile.
	Icon string `json:"icon,omitempty"`
	// Facts are the decision cells — price, lock-in, learning curve, best for.
	Facts []ShowcaseFact `json:"facts"`
	// Strengths are what it is genuinely good at.
	Strengths []string `json:"strengths"`
	// Limits are what to watch out for. Required, and the reason this template
	// is worth having: a showcase with no honest limitation is an advert.
	Limits []string `json:"limits"`
}

// ShowcaseFact is one cell of the decision grid.
type ShowcaseFact struct {
	// Label is what is being stated — "Price", "Lock-in", "Best for".
	Label string `json:"label"`
	// Value is the answer, short enough to read at a glance.
	Value string `json:"value"`
}

// ResolvedIcon returns the icon drawn in the product tile.
func (s ShowcaseSpec) ResolvedIcon() string {
	if icon := normalizePointIconName(s.Icon); icon != "" {
		return icon
	}
	return "box"
}

// ShowcaseBeat is one move on the card.
type ShowcaseBeat struct {
	// Show is the action: a showcaseShows name.
	Show string `json:"show"`
	// At indexes ShowcaseSpec.Facts, for a "fact" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to a fact.
func (b ShowcaseBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if showcaseShows[s] {
		return s
	}
	return "fact"
}

func normalizeShowcasePlan(p *SnippetPlan) {
	s := p.Showcase
	if s == nil {
		return
	}
	s.Name = clampWords(collapseSpaces(s.Name), maxShowcaseNameWords)
	s.Category = clampWords(collapseSpaces(s.Category), maxShowcaseLabelWords)
	s.Tagline = clampWords(collapseSpaces(s.Tagline), maxShowcaseTaglineWords)
	s.Icon = s.ResolvedIcon()

	facts := make([]ShowcaseFact, 0, len(s.Facts))
	for _, f := range s.Facts {
		f.Label = clampWords(collapseSpaces(f.Label), maxShowcaseLabelWords)
		f.Value = clampWords(collapseSpaces(f.Value), maxShowcaseValueWords)
		// A cell with only half of itself is a labelled blank. Dropping it is
		// the repair; inventing the missing half would be a claim about the
		// product.
		if f.Label != "" && f.Value != "" && len(facts) < maxShowcaseFacts {
			facts = append(facts, f)
		}
	}
	s.Facts = facts
	s.Strengths = clampPoints(s.Strengths, maxShowcaseStrengths)
	s.Limits = clampPoints(s.Limits, maxShowcaseLimits)

	for i := range p.Beats {
		b := p.Beats[i].Showcase
		if b == nil {
			continue
		}
		// An unlabelled beat is a blank field, not a claim. The card's shape says
		// what the ends are for: the first beat meets the tool and the last hands
		// over, so those are inferrable and everything between them is a fact.
		if !showcaseShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			switch {
			case i == 0:
				b.Show = "intro"
			case i == len(p.Beats)-1:
				b.Show = "handoff"
			default:
				b.Show = "fact"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "fact" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(s.Facts); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

// clampPoints trims a bullet list: each entry to the point-word limit, the list
// to its maximum, and empties dropped.
func clampPoints(in []string, max int) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = clampWords(collapseSpaces(s), maxShowcasePointWords)
		if s != "" && len(out) < max {
			out = append(out, s)
		}
	}
	return out
}

func validateShowcasePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Showcase: true}); err != nil {
		return err
	}

	s := p.Showcase
	if s == nil {
		return fmt.Errorf("the plan has no showcase — this template is one tool's card, walked and then handed off")
	}
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("the tool has no name")
	}
	if strings.TrimSpace(s.Tagline) == "" {
		return fmt.Errorf("%q has no tagline — say what it is in one line, in the words of somebody who has never heard of it", s.Name)
	}
	if n := len(s.Facts); n < minShowcaseFacts || n > maxShowcaseFacts {
		return fmt.Errorf("the card has %d facts, want %d-%d — a tool you can describe in two is one the illustration template introduces in half the runtime",
			n, minShowcaseFacts, maxShowcaseFacts)
	}
	seenLabel := map[string]bool{}
	for i, f := range s.Facts {
		if strings.TrimSpace(f.Label) == "" || strings.TrimSpace(f.Value) == "" {
			return fmt.Errorf("fact %d is half-written — every cell needs a label and a value", i)
		}
		key := strings.ToLower(strings.TrimSpace(f.Label))
		if seenLabel[key] {
			return fmt.Errorf("two facts are both labelled %q — each cell answers a different question", f.Label)
		}
		seenLabel[key] = true
	}
	if n := len(s.Strengths); n < minShowcaseStrengths || n > maxShowcaseStrengths {
		return fmt.Errorf("there are %d strengths, want %d-%d", n, minShowcaseStrengths, maxShowcaseStrengths)
	}
	// The rule this template exists for.
	if n := len(s.Limits); n < minShowcaseLimits || n > maxShowcaseLimits {
		return fmt.Errorf("there are %d things to watch out for, want %d-%d. A showcase with no honest limitation is an advert, and a course full of adverts teaches nobody which tool to pick — say what %s is genuinely bad at, or who should not use it",
			n, minShowcaseLimits, maxShowcaseLimits, s.Name)
	}

	factBeats := map[int]bool{}
	counts := map[string]int{}
	for i, b := range p.Beats {
		if b.Showcase == nil {
			return fmt.Errorf("beat %q has no showcase direction — every beat is lighting part of the card", b.ID)
		}
		show := b.Showcase.ResolvedShow()
		counts[show]++
		if i == 0 && show != "intro" {
			return fmt.Errorf("the clip opens on %q. Open by meeting the tool — the first beat is the intro", show)
		}
		if show == "handoff" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q hands off to the demo but the clip carries on afterwards. The hand-off is the cut point, so nothing follows it", b.ID)
		}
		if show != "fact" {
			continue
		}
		if b.Showcase.At < 0 || b.Showcase.At >= len(s.Facts) {
			return fmt.Errorf("beat %q states fact %d, which does not exist", b.ID, b.Showcase.At)
		}
		if factBeats[b.Showcase.At] {
			return fmt.Errorf("beat %q states fact %d again; each cell gets one beat", b.ID, b.Showcase.At)
		}
		factBeats[b.Showcase.At] = true
	}
	if counts["intro"] != 1 {
		return fmt.Errorf("there are %d intro beats; the tool is met once", counts["intro"])
	}
	if len(factBeats) != len(s.Facts) {
		return fmt.Errorf("%d of the %d facts are never said out loud — a cell nobody narrates is a value nobody read",
			len(s.Facts)-len(factBeats), len(s.Facts))
	}
	if counts["strengths"] == 0 {
		return fmt.Errorf("no beat covers what it is good at")
	}
	if counts["limits"] == 0 {
		return fmt.Errorf("no beat covers what to watch out for. The limits are written on the card but never spoken, which is the same as not saying them — give them a beat")
	}
	if counts["handoff"] != 1 {
		return fmt.Errorf("there are %d hand-off beats; the clip ends on exactly one, and it is the frame your demo cuts onto", counts["handoff"])
	}
	if p.Beats[len(p.Beats)-1].Showcase.ResolvedShow() != "handoff" {
		return fmt.Errorf("the clip does not end on the hand-off. Close with {\"show\": \"handoff\"} — that frame is the cut point for the demo recording")
	}
	return nil
}

// showcaseScenes lays the clip out as ONE scene: the card is on screen for the
// whole clip and the beats only move what is lit, including the last one, which
// pushes the card back and brings the hand-off plate forward.
func showcaseScenes(in SnippetSceneInput) ([]Scene, error) {
	s := in.Plan.Showcase
	if s == nil {
		return nil, fmt.Errorf("the plan has no showcase")
	}

	facts := make([]map[string]any, len(s.Facts))
	for i, f := range s.Facts {
		facts[i] = map[string]any{"label": f.Label, "value": f.Value}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Showcase == nil {
			return nil, fmt.Errorf("beat %q has no showcase direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Showcase.ResolvedShow(),
		}
		if beat.Showcase.ResolvedShow() == "fact" {
			step["at"] = beat.Showcase.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneShowcase,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":     in.Plan.Title,
			"name":      s.Name,
			"category":  s.Category,
			"tagline":   s.Tagline,
			"icon":      s.ResolvedIcon(),
			"facts":     facts,
			"strengths": s.Strengths,
			"limits":    s.Limits,
			"steps":     steps,
		},
	}}, nil
}
