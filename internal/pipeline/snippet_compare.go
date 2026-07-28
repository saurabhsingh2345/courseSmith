package pipeline

// The compare template: two things, side by side, and a verdict.
//
// The catalog could show one subject at a time and nothing else. Every template
// before this one frames a single thing — a board, a diagram, a figure, an
// editor — and teaching contrasts constantly: the naive way and the good way,
// v1 and v2, ours and theirs, before and after. Doing that with any existing
// template means cutting between two shots, which is exactly the thing that
// does not work: the viewer has to hold the first in memory to compare it to
// the second, and holding it is the work you were trying to save them.
//
// So both are on screen at once, and the shape is enforced: introduce one, then
// the other, then a beat with both lit, then the verdict. The `both` beat is
// the one that matters and the one a model will skip — a clip that shows two
// things separately and then announces a winner has not compared anything.
//
// The verdict may be a tie. Most honest comparisons are: the two sides have
// different costs and the answer is "it depends". Forcing a winner would make
// the template lie in exactly the cases where the teaching is most useful.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "compare",
		Title:       "Side by side",
		Description: "Two approaches in one frame, introduced in turn, then judged.",
		Example:     "A for loop building a list versus a list comprehension",
		PromptFile:  snippetCompareTemplateName,
		// The columns can hold code, and code that is shown is code that ran.
		NeedsCode:        true,
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		Owns:             beatFields{Compare: true},
		OwnsPlan:         planFields{Compare: true},
		Normalize:        normalizeComparePlan,
		Validate:         validateComparePlan,
		Scenes:           compareScenes,
		PromptData: func(spec SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":         strings.Join(CompareShowNames(), ", "),
				"Figures":       strings.Join(ArtFigureNames(), ", "),
				"Language":      spec.ResolvedCodeLanguage(),
				"MaxLabelWords": maxCompareLabelWords,
				"MaxNoteWords":  maxCompareNoteWords,
				"MaxCodeLines":  maxCompareCodeLines,
			}
		},
	})
}

const snippetCompareTemplateName = "snippet_compare.tmpl"

const (
	maxCompareLabelWords = 4
	maxCompareNoteWords  = 8
	// Half the frame, at a size two columns of code stay legible at. Past this
	// the reader is scrolling their eyes rather than comparing.
	maxCompareCodeLines = 9
)

// compareShows is what a beat does to the two columns.
var compareShows = map[string]bool{
	// The left column arrives and is the only thing lit.
	"left": true,
	// The right column arrives beside it.
	"right": true,
	// Both lit, neither favoured. The comparison itself.
	"both": true,
	// The verdict lands on whichever side won (or on both, for a tie).
	"verdict": true,
}

// CompareShowNames returns the vocabulary sorted, for the prompt.
func CompareShowNames() []string {
	out := make([]string, 0, len(compareShows))
	for s := range compareShows {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func normalizeCompareShow(s string) string {
	n := strings.ToLower(strings.TrimSpace(s))
	if compareShows[n] {
		return n
	}
	return "both"
}

// CompareSpec is the pair being compared. On the plan rather than on a beat for
// the reason the quiz's question is: the pair is the subject of the clip.
type CompareSpec struct {
	Left  CompareSide `json:"left"`
	Right CompareSide `json:"right"`
	// Winner is "left", "right", or "tie". A tie is a first-class answer, not a
	// failure to decide — see the file header.
	Winner string `json:"winner"`
	// Verdict is the one line the clip lands on. Shown across both columns
	// when the verdict beat arrives.
	Verdict string `json:"verdict"`
}

// CompareSide is one column.
type CompareSide struct {
	// Label heads the column: "A for loop", "A comprehension".
	Label string `json:"label"`
	// Code fills the column when this side is code. Mutually exclusive with
	// Figure; a column showing both would be two subjects in one column, which
	// is the thing this template exists to avoid.
	Code string `json:"code,omitempty"`
	// Figure names an artwork figure when this side is not code.
	Figure string `json:"figure,omitempty"`
	// Note is the short measured claim under the column — "6 lines",
	// "two passes over the data". It is what makes the comparison concrete.
	Note string `json:"note,omitempty"`
}

// CompareBeat is one move in the comparison.
type CompareBeat struct {
	Show string `json:"show"`
}

// normalizeComparePlan repairs everything about the pair that has one sensible
// repair, so those never cost a correction round.
//
// The line it draws: trimming a five-word label to four is arithmetic, and
// lowercasing "Right" is spelling. Deciding what a column *is* when the model
// gave it both code and a figure is not — that is a judgement about the clip's
// content, and guessing it would silently ship the wrong comparison.
//
// The winner gets one repair worth naming. Models answer with the column's
// label ("A comprehension") about as often as with the side ("right"), and that
// is not a mistake about the clip — it is the same answer in the other
// vocabulary, so it is matched back. An answer that resolves to neither column
// is left alone for validation, because the winner is a *claim* and quietly
// rewriting a claim is a different act from tidying a label.
func normalizeComparePlan(p *SnippetPlan) {
	c := p.Compare
	if c == nil {
		return
	}
	side := func(s *CompareSide) {
		s.Label = clampWords(collapseSpaces(s.Label), maxCompareLabelWords)
		s.Note = clampWords(collapseSpaces(s.Note), maxCompareNoteWords)
		// A figure name only means anything if the renderer can draw it, and
		// figureFor falls back to a burst that argues for nothing.
		if strings.TrimSpace(s.Code) == "" && strings.TrimSpace(s.Figure) != "" {
			s.Figure = normalizeArtFigure(s.Figure)
		}
	}
	side(&c.Left)
	side(&c.Right)
	c.Verdict = collapseSpaces(c.Verdict)

	w := strings.ToLower(strings.TrimSpace(c.Winner))
	switch w {
	case "left", "right", "tie":
	case strings.ToLower(c.Left.Label):
		w = "left"
	case strings.ToLower(c.Right.Label):
		w = "right"
	case "neither", "both", "draw", "either":
		// Four ways of saying the same thing the template already has a word
		// for, and none of them changes what the clip concludes.
		w = "tie"
	}
	c.Winner = w

	for i := range p.Beats {
		if b := p.Beats[i].Compare; b != nil {
			b.Show = normalizeCompareShow(b.Show)
		}
	}
}

func validateCompareSide(name string, s CompareSide) error {
	if strings.TrimSpace(s.Label) == "" {
		return fmt.Errorf("the %s column has no label", name)
	}
	if n := len(strings.Fields(s.Label)); n > maxCompareLabelWords {
		return fmt.Errorf("the %s column's label is %d words; at most %d", name, n, maxCompareLabelWords)
	}
	hasCode := strings.TrimSpace(s.Code) != ""
	hasFigure := strings.TrimSpace(s.Figure) != ""
	if hasCode && hasFigure {
		return fmt.Errorf("the %s column has both code and a figure — a column shows one thing, or it is two comparisons in one", name)
	}
	if !hasCode && !hasFigure {
		return fmt.Errorf("the %s column is empty — give it code or a figure", name)
	}
	if hasCode {
		if n := len(strings.Split(strings.TrimRight(s.Code, "\n"), "\n")); n > maxCompareCodeLines {
			return fmt.Errorf("the %s column is %d lines of code; at most %d — two columns at that size stop being readable",
				name, n, maxCompareCodeLines)
		}
	}
	if n := len(strings.Fields(s.Note)); n > maxCompareNoteWords {
		return fmt.Errorf("the %s column's note is %d words; at most %d — it is a measurement, not a sentence",
			name, n, maxCompareNoteWords)
	}
	return nil
}

func validateComparePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Compare: true}); err != nil {
		return err
	}

	c := p.Compare
	if c == nil {
		return fmt.Errorf("the plan has no comparison — this template is two things judged against each other")
	}
	if err := validateCompareSide("left", c.Left); err != nil {
		return err
	}
	if err := validateCompareSide("right", c.Right); err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(c.Left.Label), strings.TrimSpace(c.Right.Label)) {
		return fmt.Errorf("both columns are labelled %q — they are meant to be different things", c.Left.Label)
	}
	switch strings.ToLower(strings.TrimSpace(c.Winner)) {
	case "left", "right", "tie":
	default:
		return fmt.Errorf("winner is %q; it must be \"left\", \"right\" or \"tie\"", c.Winner)
	}
	if strings.TrimSpace(c.Verdict) == "" {
		return fmt.Errorf("the comparison has no verdict line — say what the clip concludes")
	}

	leftAt, rightAt, bothAt, verdictAt := -1, -1, -1, -1
	for i, b := range p.Beats {
		if b.Compare == nil {
			return fmt.Errorf("beat %q has no compare direction — every beat is doing something to the two columns", b.ID)
		}
		switch normalizeCompareShow(b.Compare.Show) {
		case "left":
			if leftAt < 0 {
				leftAt = i
			}
		case "right":
			if rightAt < 0 {
				rightAt = i
			}
		case "both":
			if bothAt < 0 {
				bothAt = i
			}
		case "verdict":
			if verdictAt >= 0 {
				return fmt.Errorf("beat %q delivers a second verdict; the clip lands once", b.ID)
			}
			verdictAt = i
		}
	}
	if leftAt < 0 || rightAt < 0 {
		return fmt.Errorf("both columns must be introduced — one beat with \"show\": \"left\" and one with \"show\": \"right\"")
	}
	if rightAt < leftAt {
		return fmt.Errorf("the right column is introduced before the left one; introduce them in the order they are read")
	}
	if verdictAt < 0 {
		return fmt.Errorf("no beat delivers the verdict — mark the last one with \"show\": \"verdict\"")
	}
	// The beat that matters. A clip that shows two things separately and then
	// announces a winner has not compared anything: the whole reason both
	// columns are on screen at once is that there is a moment where the viewer
	// looks at them together.
	if bothAt < 0 {
		return fmt.Errorf("no beat shows both columns lit together. Put a \"both\" beat between introducing the second column and the verdict — looking at the two side by side IS the comparison, and without it this is two descriptions and an announcement")
	}
	if bothAt < rightAt {
		return fmt.Errorf("both columns are lit before the right one has been introduced")
	}
	if verdictAt < bothAt {
		return fmt.Errorf("the verdict lands before the two have been seen together")
	}
	return nil
}

// compareScenes lays the clip out as ONE scene: both columns are in the frame
// from the start and the beats only change what is lit, which is the same shape
// the quiz and data templates use and for the same reason — a column that
// re-mounts is a column the eye has to find again.
func compareScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.Compare
	if c == nil {
		return nil, fmt.Errorf("the plan has no comparison")
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Compare == nil {
			return nil, fmt.Errorf("beat %q has no compare direction", beat.ID)
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    normalizeCompareShow(beat.Compare.Show),
		})
	}

	side := func(s CompareSide) map[string]any {
		out := map[string]any{"label": s.Label, "note": s.Note}
		if strings.TrimSpace(s.Code) != "" {
			out["code"] = s.Code
		} else {
			out["figure"] = normalizeArtFigure(s.Figure)
		}
		return out
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneCompare,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":    in.Plan.Title,
			"language": in.Spec.ResolvedCodeLanguage(),
			"left":     side(c.Left),
			"right":    side(c.Right),
			"winner":   strings.ToLower(strings.TrimSpace(c.Winner)),
			"verdict":  c.Verdict,
			"steps":    steps,
		},
	}}, nil
}
