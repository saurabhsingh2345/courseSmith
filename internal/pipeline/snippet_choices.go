package pipeline

// The choices template: a question, then the answers, and nothing else.
//
// The catalog already had two ways to put named things in a row and neither of
// them can do this. `cards` requires a line under every name, deliberately and
// for a good reason — a card that is a logo and a name is a sticker. `quiz`
// requires an answer, a pause and a reveal, and requires a `why` for every
// option because saying what made a wrong answer tempting is where its teaching
// is. Both rules are right for those templates and both are fatal here.
//
// What was missing is the frame a recap actually needs: the voice asks which one
// it was, four names come up, and the clip stops. There is no line under the
// names because a line under the names is the answer — "the app we actually
// opened" hands over what the viewer was about to work out. There is no reveal
// because the answer is not in this clip at all: it is in the viewer's head, or
// it is in the next section. A template that insists on telling them cannot be
// used for asking.
//
// So three things are refused outright, and the refusals are the template.
//
// No note, and no room for one. Not "optional" — absent. An optional note is a
// note that gets filled in, and the first plan to fill it in has turned the
// frame back into `cards` with the answers written on it.
//
// No role, no tint, no accent on any single option. Every other card row in the
// catalog can single a card out, which is exactly what this frame must never do:
// a row of four where one is rimmed in the accent colour has been answered
// before it was read. The options render identically by construction, so no
// plan can lean on one of them.
//
// The question is asked before the answers exist. The opening beat is the
// question alone on the frame — the options have not arrived yet — because a row
// of four names on screen while the question is still being spoken is a viewer
// reading the answers and hearing nothing. One beat of question, then the row.
// That ordering is what buys the thinking, and it is the cheap version of the
// pause `quiz` enforces with a held beat.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "choices",
		Category: CatPresenting,
		Since:    SinceV11,
		// Showroom, and for the same reason `cards` is: the options in a recap
		// are usually named products wearing their own marks, and a real mark on
		// a near-black stage has to be recoloured to be seen.
		Family:      FamilyShowroom,
		Title:       "Pick one",
		Description: "A question, then the answers to choose from as bare cards — no explanation under any of them and no reveal. Reach for it for a recap question the viewer is meant to answer in their head.",
		Example:     "Which tool did we use to set up our web application: Cursor, Claude Desktop, OpenAI or Gemini",
		PromptFile:  snippetChoicesTemplateName,
		NeedsCode:   false,
		// Eight seconds is a question and a row, which is the whole shape. This
		// is the shortest floor in the catalog on purpose: a recap question that
		// takes half a minute has stopped being a recap.
		MinTargetSec:     8,
		DefaultTargetSec: 14,
		// Past this the row is held with nothing happening. There is no reveal
		// coming, so there is nothing for a longer runtime to spend itself on.
		MaxTargetSec: 40,
		MaxBeats:     4,
		// A beat is a SHOT and there are only ever two or three of them, so the
		// budget has to divide small. Fourteen words is about five seconds of
		// question, which is as long as a question can be and still be held in
		// mind while the options land.
		IdealWordsPerBeat: 14,
		Owns:              beatFields{Choices: true},
		OwnsPlan:          planFields{Choices: true},
		Normalize:         normalizeChoicesPlan,
		Validate:          validateChoicesPlan,
		Scenes:            choicesScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":          strings.Join(ChoicesShows(), ", "),
				"Icons":          strings.Join(PointIconNames(), ", "),
				"MinOptions":     minChoiceOptions,
				"MaxOptions":     maxChoiceOptions,
				"MaxLabelWords":  maxChoiceLabelWords,
				"MaxHoldsPerRow": maxChoiceHolds,
			}
		},
	})
}

const snippetChoicesTemplateName = "snippet_choices.tmpl"

const (
	// Two is a coin flip rather than a choice, so three is the floor — the same
	// arithmetic `quiz` uses for its options and for the same reason. Five is
	// the ceiling because the whole row has to be read in the seconds the
	// question is still ringing, and because past five the marks drop below the
	// size that lets a logo be recognised rather than read.
	minChoiceOptions = 3
	maxChoiceOptions = 5

	// A label is a NAME — "Cursor", "Claude Desktop", "Something else". Four
	// words is the most a name runs to, and anything longer is an explanation
	// wearing a label's clothes, which is the one thing this template refuses.
	maxChoiceLabelWords = 4

	// How many held beats may follow the row. One is a pause worth having, and a
	// second held frame is a clip that has run out of things to do.
	maxChoiceHolds = 1
)

// choicesShows is the closed vocabulary of what a beat does. Three words, and
// two of them are mandatory — this is the least stateful template in the catalog
// because the frame it builds has no state to move.
var choicesShows = map[string]bool{
	// The question alone, centred, with no options on the frame yet.
	"ask": true,
	// The options land, staggered left to right, under the question.
	"options": true,
	// Held. The row is up, nothing moves, the viewer is choosing. Optional, and
	// the only beat here that is.
	"hold": true,
}

// ChoicesShows returns the beat vocabulary sorted.
func ChoicesShows() []string {
	out := make([]string, 0, len(choicesShows))
	for k := range choicesShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ChoicesSpec is the row of answers. On the plan rather than on a beat because
// the row belongs to the question, not to a moment in the clip.
//
// There is no Answer field, and that is not an omission. The answer is what the
// viewer is supposed to supply, and a template that carries one grows a beat
// that shows it — see the file header.
type ChoicesSpec struct {
	// Options are the answers, left to right, in the order the voice reads them.
	Options []ChoiceOption `json:"options"`
}

// ChoiceOption is one answer: a name, a mark, and deliberately nothing else.
type ChoiceOption struct {
	// Label is the answer as the viewer would say it — "Claude Desktop".
	Label string `json:"label"`

	// == Where the mark comes from, resolved exactly as a card's is. ==

	// Brand is a Simple Icons slug — "cursor", "googlegemini".
	Brand string `json:"brand,omitempty"`
	// Site is the thing's domain, tried when there is no brand mark to be had.
	Site string `json:"site,omitempty"`
	// Icon is a PointIconNames name and it is the floor rather than a choice:
	// what the option wears when it never had a logo in the first place, which
	// is every option that is a word rather than a product.
	Icon string `json:"icon,omitempty"`

	// == Resolved by the pipeline, not written by the model. ==

	// Mark is SVG path data on a 0 0 24 24 viewBox.
	Mark string `json:"mark,omitempty"`
	// Tint is the brand's own hex, taken off the fetched mark.
	Tint string `json:"tint,omitempty"`
	// Image is a data: URI for a mark that has to keep its own colours.
	Image string `json:"image,omitempty"`
	// MarkFrom records where the art came from, so a wrong logo on a finished
	// frame can be traced rather than guessed at.
	MarkFrom string `json:"markFrom,omitempty"`
}

// ResolvedIcon returns the option's glyph, never empty.
func (o ChoiceOption) ResolvedIcon() string {
	if icon := normalizePointIconName(o.Icon); icon != "" {
		return icon
	}
	return "dot"
}

// asCard adapts one option to the card art resolver, which is the only place in
// the pipeline that knows how to turn a slug into a mark. Adapted rather than
// duplicated: two copies of that fetch order would drift, and the order is a
// quality judgement worth having in one place.
func (o ChoiceOption) asCard() Card {
	return Card{
		Title: o.Label,
		// The resolver never reads a note and this template has none. Set to the
		// label so nothing downstream sees an empty required field.
		Note:     o.Label,
		Brand:    o.Brand,
		Site:     o.Site,
		Icon:     o.ResolvedIcon(),
		Mark:     o.Mark,
		Tint:     o.Tint,
		Image:    o.Image,
		MarkFrom: o.MarkFrom,
	}
}

// ChoicesBeat is one move: ask, land the row, or hold it.
type ChoicesBeat struct {
	Show string `json:"show"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to a hold —
// the only one of the three that is safe to guess, because "ask" and "options"
// each happen exactly once and guessing either would duplicate it.
func (b ChoicesBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if choicesShows[s] {
		return s
	}
	return "hold"
}

func normalizeChoicesPlan(p *SnippetPlan) {
	c := p.Choices
	if c == nil {
		return
	}

	options := make([]ChoiceOption, 0, len(c.Options))
	for _, o := range c.Options {
		o.Label = clampWords(collapseSpaces(o.Label), maxChoiceLabelWords)
		o.Icon = o.ResolvedIcon()
		if o.Label != "" && len(options) < maxChoiceOptions {
			options = append(options, o)
		}
	}
	c.Options = options

	// The first beat asks and the second lands the row. Defaulted by position
	// rather than left to the model, because those two beats are the template
	// and a plan that names neither of them is a plan that meant the obvious.
	for i := range p.Beats {
		b := p.Beats[i].Choices
		if b == nil {
			continue
		}
		if choicesShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			b.Show = b.ResolvedShow()
			continue
		}
		switch i {
		case 0:
			b.Show = "ask"
		case 1:
			b.Show = "options"
		default:
			b.Show = "hold"
		}
	}
}

func validateChoicesPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Choices: true}); err != nil {
		return err
	}

	c := p.Choices
	if c == nil {
		return fmt.Errorf("the plan has no options — this template is a question and the answers to choose from, so the row is the clip")
	}
	if n := len(c.Options); n < minChoiceOptions || n > maxChoiceOptions {
		return fmt.Errorf("there are %d options, want %d-%d. Two is a coin flip rather than a choice, and past %d the whole row cannot be read in the seconds the question is still ringing",
			n, minChoiceOptions, maxChoiceOptions, maxChoiceOptions)
	}

	seen := map[string]bool{}
	for i, o := range c.Options {
		label := strings.TrimSpace(o.Label)
		if label == "" {
			return fmt.Errorf("option %d has no label. The label is the answer — an option with a mark and no name is a logo nobody can pick", i)
		}
		if n := len(strings.Fields(label)); n > maxChoiceLabelWords {
			return fmt.Errorf("option %d is %d words (%q); at most %d. A label is a NAME, and a longer one is an explanation wearing a label's clothes — which is the hint this template exists to withhold",
				i, n, label, maxChoiceLabelWords)
		}
		key := strings.ToLower(label)
		if seen[key] {
			return fmt.Errorf("option %d repeats %q, so the row offers the same answer twice", i, label)
		}
		seen[key] = true
	}

	counts := map[string]int{}
	askAt, optionsAt := -1, -1
	for i, b := range p.Beats {
		if b.Choices == nil {
			return fmt.Errorf("beat %q has no choices direction — every beat here asks the question, lands the row, or holds it", b.ID)
		}
		show := b.Choices.ResolvedShow()
		counts[show]++
		switch show {
		case "ask":
			askAt = i
		case "options":
			optionsAt = i
		}
	}
	if counts["ask"] != 1 {
		return fmt.Errorf("there are %d beats asking the question; it is asked once, and it is the first thing on the frame", counts["ask"])
	}
	if askAt != 0 {
		return fmt.Errorf("beat %q asks the question, but it is beat %d. The question comes first and alone: a row of answers already on screen while the question is still being spoken is a viewer reading instead of thinking",
			p.Beats[askAt].ID, askAt+1)
	}
	if counts["options"] != 1 {
		return fmt.Errorf("there are %d beats landing the row; the options arrive once. If the voice reads them twice, that second pass is a hold", counts["options"])
	}
	if optionsAt != 1 {
		return fmt.Errorf("the options land on beat %d. They land immediately after the question — everything between would be a held frame with nothing on it yet",
			optionsAt+1)
	}
	if n := counts["hold"]; n > maxChoiceHolds {
		return fmt.Errorf("there are %d held beats after the row; at most %d. There is no reveal coming, so a second held frame is a clip that has run out of things to do",
			n, maxChoiceHolds)
	}
	return nil
}

// choicesScenes lays the clip out as ONE scene: the question is on screen for
// the whole thing and the beats only decide whether the row is up.
func choicesScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.Choices
	if c == nil {
		return nil, fmt.Errorf("the plan has no options")
	}

	options := make([]map[string]any, len(c.Options))
	for i, o := range c.Options {
		options[i] = map[string]any{
			"label": o.Label,
			"icon":  o.ResolvedIcon(),
			"mark":  o.Mark,
			"tint":  o.Tint,
			"image": o.Image,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Choices == nil {
			return nil, fmt.Errorf("beat %q has no choices direction", beat.ID)
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Choices.ResolvedShow(),
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneChoices,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":   in.Plan.Title,
			"options": options,
			"steps":   steps,
		}),
	}}, nil
}
