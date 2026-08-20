package pipeline

// The titlecard template: a held card that names what is about to happen.
//
// The catalog already has `opener`, and the difference between them is the whole
// reason this exists. An opener sets its subject ENORMOUS at 12% ink, as texture
// — the words fill the frame and are deliberately not read, and the two small
// solid lines on top are what the viewer takes in. That is a title page, and it
// is right once, at the front of a piece.
//
// This is the other thing: a card that appears BETWEEN sections, four or five
// times in one video, and is read every time. So every decision inverts. The type
// is solid rather than ghosted, because it is the message. It is centred and
// optically sized rather than frame-filling, because a card that fills the frame
// cannot appear five times without becoming the video. And it is short — a
// section name, a command, a turn of the argument — where an opener wants length.
//
// Two treatments carry it, and both come from watching how a good tutorial cuts.
//
// THE SPLIT. A card can name two things at once and grade them: the first part
// solid, the second in the muted tone. "Scope" then "and steer" — the eye reads
// the first, understands the second is subordinate, and the card has made an
// argument rather than printed a label. One line of type doing two jobs.
//
// THE COMMAND. When the section is about a literal thing somebody types, the card
// sets it in mono at display size instead of the serif. That single face change
// is what tells a viewer the next thirty seconds are hands-on: the frame stops
// being editorial and becomes an instruction, without a word of narration
// spending itself on the transition.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "titlecard",
		Category: CatPresenting,
		Since:    SinceV10,
		Family:   FamilyAtelier,
		Title:    "The section card",
		Description: "A held card in display type that names the section about to start — solid, centred, optionally split into a subordinate second half or set as the literal command being taught. " +
			"Reach for it BETWEEN stretches of teaching; `opener` is the title page at the front.",
		Example:    "/compact",
		PromptFile: snippetTitleCardTemplateName,
		NeedsCode:  false,
		// A section card is a breath, not a scene. Eight seconds is enough to
		// read one line and understand the turn; past about twenty-five the
		// video has stopped to admire its own typography.
		MinTargetSec:     8,
		DefaultTargetSec: 14,
		MaxTargetSec:     30,
		// Three: the line, its subordinate half, the note under it. There is no
		// fourth thing on a card whose entire job is to be read in one glance.
		MaxBeats: 3,
		// Low, and lower than opener's. A card is held for a sentence.
		IdealWordsPerBeat: 18,
		Owns:              beatFields{TitleCard: true},
		OwnsPlan:          planFields{TitleCard: true},
		Normalize:         normalizeTitleCardPlan,
		Validate:          validateTitleCardPlan,
		Scenes:            titleCardScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":        strings.Join(TitleCardShows(), ", "),
				"MaxLineWords": maxTitleCardLineWords,
				"MaxTailWords": maxTitleCardTailWords,
				"MaxNoteWords": maxTitleCardNoteWords,
				"MaxLabelChars": maxTitleCardLabelChars,
			}
		},
	})
}

const snippetTitleCardTemplateName = "snippet_titlecard.tmpl"

const (
	// The line that is the card. Short by construction: this is set at display
	// size and centred, and past about seven words it wraps to three lines and
	// stops being a card.
	maxTitleCardLineWords = 7
	// The subordinate half of a split. Shorter than the line it hangs off —
	// a tail longer than its head reads as two competing statements.
	maxTitleCardTailWords = 5
	// The small line underneath, when the card needs one.
	maxTitleCardNoteWords = 12
	// The mono label above the line — a section mark, a part number. Characters
	// rather than words because "3 / 5" is not words.
	maxTitleCardLabelChars = 24
)

// titleCardShows is the closed vocabulary of what a beat does.
var titleCardShows = map[string]bool{
	// The line lands, alone on the card. The first beat.
	"line": true,
	// The subordinate second half arrives in the muted tone.
	"tail": true,
	// The small line under the card lands and the frame is complete.
	"note": true,
	// The precedence ladder lands under the line.
	"stack": true,
}

// TitleCardShows returns the beat vocabulary sorted.
func TitleCardShows() []string {
	out := make([]string, 0, len(titleCardShows))
	for k := range titleCardShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TitleCardSpec is the card.
type TitleCardSpec struct {
	// Line is the card: the section, the turn, the command. Required.
	Line string `json:"line"`
	// Mono sets Line in the mono face at display size rather than the serif.
	// True when the line IS something somebody types — see the file header.
	Mono bool `json:"mono,omitempty"`
	// Tail is the subordinate half of a split card, set in the muted tone.
	// Optional; a card is complete without one.
	Tail string `json:"tail,omitempty"`
	// Label is the small mono mark above the line. Optional.
	Label string `json:"label,omitempty"`
	// Note is the small line underneath. Optional.
	Note string `json:"note,omitempty"`
	// Stack is a precedence ladder set under the line: two to four levels where
	// the LAST one is solid and the ones above it are muted.
	//
	// It exists because "which of these wins" is a spatial claim and a sentence
	// makes a poor job of it. Two names stacked, one faded, is instantly legible
	// as an order of precedence — and it is the shape every layered-configuration
	// lesson needs and none of the diagram templates draw at card scale.
	Stack []string `json:"stack,omitempty"`
	// StackNote is the small mono annotation beside the solid level — usually
	// the path that level actually lives at.
	StackNote string `json:"stackNote,omitempty"`
}

// TitleCardBeat is one shot of the card.
type TitleCardBeat struct {
	// Show is a titleCardShows name.
	Show string `json:"show"`
}

// ResolvedShow defaults the unknown to the line landing, which is the beat this
// template is really about.
func (b TitleCardBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if titleCardShows[s] {
		return s
	}
	return "line"
}

func normalizeTitleCardPlan(p *SnippetPlan) {
	c := p.TitleCard
	if c == nil {
		return
	}
	c.Line = clampWords(collapseSpaces(c.Line), maxTitleCardLineWords)
	c.Tail = clampWords(collapseSpaces(c.Tail), maxTitleCardTailWords)
	c.Note = clampWords(collapseSpaces(c.Note), maxTitleCardNoteWords)
	c.Label = clampCodeLine(collapseSpaces(c.Label), maxTitleCardLabelChars)
	c.StackNote = clampCodeLine(collapseSpaces(c.StackNote), 34)
	stack := make([]string, 0, 4)
	for _, l := range c.Stack {
		if l = clampWords(collapseSpaces(l), 4); l != "" && len(stack) < 4 {
			stack = append(stack, l)
		}
	}
	c.Stack = stack
	// A line that is a slash command is set in mono whether or not the plan
	// remembered to say so. The face change is the signal, and losing it to a
	// missing boolean is the kind of silent downgrade nobody notices.
	if strings.HasPrefix(strings.TrimSpace(c.Line), "/") {
		c.Mono = true
	}

	for i := range p.Beats {
		if b := p.Beats[i].TitleCard; b != nil {
			b.Show = b.ResolvedShow()
		}
	}
}

func validateTitleCardPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{TitleCard: true}); err != nil {
		return err
	}

	c := p.TitleCard
	if c == nil {
		return fmt.Errorf("the plan has no card — this template is one held frame with one line on it, so the card IS the clip")
	}
	if strings.TrimSpace(c.Line) == "" {
		return fmt.Errorf("the card has no line. Everything else here is optional furniture; the line is the card")
	}

	shows := map[string]int{}
	for i, b := range p.Beats {
		if b.TitleCard == nil {
			return fmt.Errorf("beat %q has no titlecard direction", b.ID)
		}
		show := b.TitleCard.ResolvedShow()
		shows[show]++
		if i == 0 && show != "line" {
			return fmt.Errorf("the first beat is %q, but a card has to arrive before anything can be added to it — the opening beat is \"line\"", show)
		}
	}
	if shows["line"] != 1 {
		return fmt.Errorf("there are %d \"line\" beats and there must be exactly one: the card lands once", shows["line"])
	}
	// A beat that shows a part the card does not have draws nothing. This is the
	// commonest hand-authoring mistake on this template — a `tail` beat written
	// against a card whose tail was later cut — and it fails as a held frame with
	// narration over it rather than as an error.
	if shows["tail"] > 0 && strings.TrimSpace(c.Tail) == "" {
		return fmt.Errorf("a beat shows the tail but the card has none: either write `tail` on the card or drop the beat")
	}
	if shows["stack"] > 0 && len(c.Stack) < 2 {
		return fmt.Errorf("a beat shows the precedence ladder but the card has %d level(s): a ladder needs at least two, or there is no order to show", len(c.Stack))
	}
	if shows["note"] > 0 && strings.TrimSpace(c.Note) == "" {
		return fmt.Errorf("a beat shows the note but the card has none: either write `note` on the card or drop the beat")
	}
	if t := strings.TrimSpace(c.Tail); t != "" && strings.EqualFold(t, strings.TrimSpace(c.Line)) {
		return fmt.Errorf("the tail repeats the line. The split works because the second half is SUBORDINATE — say the other thing, or drop the tail")
	}
	return nil
}

func titleCardScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.TitleCard
	if c == nil {
		return nil, fmt.Errorf("the plan has no card")
	}

	// Latched, like the opener's lines: once a part is up it stays up, so the
	// card is whole at the cut rather than mid-assembly.
	tail, note, stack := false, false, false
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.TitleCard == nil {
			return nil, fmt.Errorf("beat %q has no titlecard direction", beat.ID)
		}
		show := beat.TitleCard.ResolvedShow()
		switch show {
		case "tail":
			tail = true
		case "note":
			note = true
		case "stack":
			stack = true
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"tail":    tail,
			"note":    note,
			"stack":   stack,
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneTitleCard,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"line":      c.Line,
			"stack":     c.Stack,
			"stackNote": c.StackNote,
			"mono":  c.Mono,
			"tail":  c.Tail,
			"label": c.Label,
			"note":  c.Note,
			"steps": steps,
		},
	}}, nil
}
