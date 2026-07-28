package pipeline

// The anatomy template: one thing, held still, taken apart.
//
// A function signature, a URL, a command line, a config line, a stack frame.
// The move it makes is "look at *this* bit of this thing", and no other template
// in the catalog can make it: the whiteboard and the flow diagram draw
// relationships between separate boxes, and everything else replaces the frame
// each beat. Here the artefact never moves, never re-animates, and is never
// replaced — the only thing that changes is which characters are lit and what
// is being said about them.
//
// **A part is a literal substring of the subject, and that is enforced.** The
// model does not describe where to point; it quotes the text it means, and Go
// finds it. That rules out the failure this template would otherwise have all
// the time — a callout landing on the wrong characters, or on none — and it is
// the same trade the illustration template makes by requiring its emphasis word
// to occur in its own heading. Spans are resolved here rather than in the
// renderer for the usual reason: it is string logic, it wants unit tests, and
// two implementations of "which characters did they mean" would disagree.

import (
	"fmt"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "anatomy",
		Title:       "Anatomy",
		Description: "One artefact taken apart — callouts reaching to each labelled piece in turn.",
		Example:     "Break down the parts of a Python function signature",
		PromptFile:  snippetAnatomyTemplateName,
		NeedsCode:   false,
		// Overview, then one beat per part, then a close: three parts is already
		// five beats, and rushing a part defeats the point of pausing on it.
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		Owns:             beatFields{Anatomy: true},
		OwnsPlan:         planFields{Anatomy: true},
		Normalize:        normalizeAnatomyPlan,
		Validate:         validateAnatomyPlan,
		Scenes:           anatomyScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"MinParts":        minAnatomyParts,
				"MaxParts":        maxAnatomyParts,
				"MaxSubjectChars": maxAnatomySubjectChars,
				"MaxLabelWords":   maxAnatomyLabelWords,
				"MaxNoteWords":    maxAnatomyNoteWords,
			}
		},
	})
}

const snippetAnatomyTemplateName = "snippet_anatomy.tmpl"

const (
	// Two parts is a comparison, not an anatomy; past five the subject is a
	// paragraph and the callouts have nowhere to go.
	minAnatomyParts = 3
	maxAnatomyParts = 5
	// One line, set large enough to read the individual characters a callout
	// points at. Past this it wraps, and a wrapped subject cannot carry
	// callouts under the right columns.
	maxAnatomySubjectChars = 58
	maxAnatomyLabelWords   = 4
	maxAnatomyNoteWords    = 16
)

// AnatomySpec is the artefact and its parts. On the plan for the same reason
// the quiz's question is: it is the subject of the clip.
type AnatomySpec struct {
	// Subject is the artefact, exactly as it should appear on screen.
	Subject string `json:"subject"`
	// Parts are the pieces, in the order they are read (left to right).
	Parts []AnatomyPart `json:"parts"`
}

// AnatomyPart is one labelled piece of the subject.
type AnatomyPart struct {
	// Text is the literal substring of Subject this part is. Quoted, not
	// described — see the file header.
	Text string `json:"text"`
	// Label names the part: "the parameter list", "the query string".
	Label string `json:"label"`
	// Note says what it does. One sentence.
	Note string `json:"note"`
}

// AnatomyBeat says which part this beat is on.
type AnatomyBeat struct {
	// Part indexes AnatomySpec.Parts. Omitted (or negative) means the whole
	// artefact is shown with nothing singled out — the overview, and the close.
	Part int `json:"part"`
	// Whole marks a beat that deliberately shows everything. Present because
	// `part` omitted decodes to 0, which is a real part index; without this the
	// opening beat would silently light the first piece.
	Whole bool `json:"whole,omitempty"`
}

// anatomySpan is one part's resolved position in the subject, in rune indices.
type anatomySpan struct {
	Start int
	End   int
}

// resolveAnatomySpans finds each part's characters in the subject.
//
// Parts are claimed in order and may not overlap: a subject where two callouts
// point at the same characters is a subject that has been described twice, and
// the renderer would draw two lines to one place. A part whose text appears
// more than once takes the first occurrence that is still free, which is what
// somebody reading left to right would mean.
func resolveAnatomySpans(subject string, parts []AnatomyPart) ([]anatomySpan, error) {
	runes := []rune(subject)
	claimed := make([]bool, len(runes))
	out := make([]anatomySpan, len(parts))

	for i, p := range parts {
		needle := []rune(p.Text)
		if len(needle) == 0 {
			return nil, fmt.Errorf("part %d (%q) quotes nothing — `text` is the literal characters of the subject this part is", i, p.Label)
		}
		found := -1
		for start := 0; start+len(needle) <= len(runes); start++ {
			match := true
			for k := range needle {
				if runes[start+k] != needle[k] || claimed[start+k] {
					match = false
					break
				}
			}
			if match {
				found = start
				break
			}
		}
		if found < 0 {
			if strings.Contains(subject, p.Text) {
				return nil, fmt.Errorf("part %d (%q) quotes %q, but those characters are already part of an earlier piece — the parts must not overlap",
					i, p.Label, p.Text)
			}
			return nil, fmt.Errorf("part %d (%q) quotes %q, which does not appear in the subject %q. `text` must be copied from the subject exactly",
				i, p.Label, p.Text, subject)
		}
		for k := range needle {
			claimed[found+k] = true
		}
		out[i] = anatomySpan{Start: found, End: found + len(needle)}
	}
	return out, nil
}

// normalizeAnatomyPlan repairs what has one repair: whitespace, over-long
// labels and notes, and the `whole` flag on a beat that named no part.
//
// It will not touch `text`. That field is a quotation, and a normalizer that
// trimmed or re-cased it would move the callout somewhere the model did not
// mean — the one thing this template must never do quietly.
func normalizeAnatomyPlan(p *SnippetPlan) {
	if a := p.Anatomy; a != nil {
		a.Subject = strings.Trim(a.Subject, " \t\n")
		for i := range a.Parts {
			a.Parts[i].Label = clampWords(collapseSpaces(a.Parts[i].Label), maxAnatomyLabelWords)
			a.Parts[i].Note = clampWords(collapseSpaces(a.Parts[i].Note), maxAnatomyNoteWords)
		}
	}
	for i := range p.Beats {
		b := p.Beats[i].Anatomy
		if b == nil {
			continue
		}
		// A negative index is how a model says "none of them"; the template
		// spells that `whole`, and the two mean the same thing.
		if b.Part < 0 {
			b.Part, b.Whole = 0, true
		}
	}
}

func validateAnatomyPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Anatomy: true}); err != nil {
		return err
	}

	a := p.Anatomy
	if a == nil {
		return fmt.Errorf("the plan has no subject — this template is one artefact taken apart")
	}
	subject := a.Subject
	if strings.TrimSpace(subject) == "" {
		return fmt.Errorf("the subject is empty — give it the line of text the clip pulls apart")
	}
	if strings.Contains(subject, "\n") {
		return fmt.Errorf("the subject spans several lines; it must be one line, so the callouts have columns to point at")
	}
	if n := len([]rune(subject)); n > maxAnatomySubjectChars {
		return fmt.Errorf("the subject is %d characters; at most %d — past that it wraps and the callouts lose their columns",
			n, maxAnatomySubjectChars)
	}
	if n := len(a.Parts); n < minAnatomyParts || n > maxAnatomyParts {
		return fmt.Errorf("there are %d parts, want %d-%d", n, minAnatomyParts, maxAnatomyParts)
	}
	for i, part := range a.Parts {
		if strings.TrimSpace(part.Label) == "" {
			return fmt.Errorf("part %d has no label", i)
		}
		if strings.TrimSpace(part.Note) == "" {
			return fmt.Errorf("part %d (%q) has no note — say what it does", i, part.Label)
		}
	}
	// The rule the template rests on.
	if _, err := resolveAnatomySpans(subject, a.Parts); err != nil {
		return err
	}

	explained := map[int]bool{}
	sawWhole := false
	for i, b := range p.Beats {
		if b.Anatomy == nil {
			return fmt.Errorf("beat %q has no anatomy direction — every beat is either on a part or on the whole thing", b.ID)
		}
		if b.Anatomy.Whole {
			sawWhole = true
			continue
		}
		if b.Anatomy.Part < 0 || b.Anatomy.Part >= len(a.Parts) {
			return fmt.Errorf("beat %q is on part %d, which does not exist", b.ID, b.Anatomy.Part)
		}
		if explained[b.Anatomy.Part] {
			return fmt.Errorf("beat %q is on part %d again; each piece gets one beat", b.ID, b.Anatomy.Part)
		}
		explained[b.Anatomy.Part] = true
		// Two beats in a row on the same part cannot happen (each is used
		// once), but a part *skipping* backwards can, and reading an artefact
		// right to left is not how anybody reads one.
		if i > 0 {
			prev := p.Beats[i-1].Anatomy
			if prev != nil && !prev.Whole && b.Anatomy.Part < prev.Part {
				return fmt.Errorf("beat %q goes back to part %d after part %d — the pieces are taken in the order they are read",
					b.ID, b.Anatomy.Part, prev.Part)
			}
		}
	}
	// The opening beat has to show the whole thing. Landing straight on a
	// highlighted fragment means the viewer has never seen the artefact intact,
	// and every callout after that points into something they are still
	// assembling in their head.
	if !sawWhole {
		return fmt.Errorf("no beat shows the whole artefact. Open with a beat carrying \"whole\": true — the viewer has to see the thing intact before any piece of it means anything")
	}
	if first := p.Beats[0].Anatomy; first != nil && !first.Whole {
		return fmt.Errorf("the clip opens on a single part; open on the whole artefact instead (\"whole\": true) and take it apart from there")
	}
	if len(explained) != len(a.Parts) {
		return fmt.Errorf("%d of the %d parts are never explained — a labelled piece nobody talks about is a callout with no sentence",
			len(a.Parts)-len(explained), len(a.Parts))
	}
	return nil
}

// anatomyScenes lays the clip out as ONE scene. The artefact is on screen from
// the first frame to the last and never re-mounts; the beats only move the
// highlight, which is the whole premise — a subject that re-animated each beat
// would be a subject the viewer has to re-read before every callout.
func anatomyScenes(in SnippetSceneInput) ([]Scene, error) {
	a := in.Plan.Anatomy
	if a == nil {
		return nil, fmt.Errorf("the plan has no subject")
	}
	spans, err := resolveAnatomySpans(a.Subject, a.Parts)
	if err != nil {
		return nil, err
	}

	parts := make([]map[string]any, len(a.Parts))
	for i, p := range a.Parts {
		parts[i] = map[string]any{
			"label": p.Label,
			"note":  p.Note,
			"start": spans[i].Start,
			"end":   spans[i].End,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Anatomy == nil {
			return nil, fmt.Errorf("beat %q has no anatomy direction", beat.ID)
		}
		step := map[string]any{"startMs": startMs, "endMs": endMs}
		if beat.Anatomy.Whole {
			step["whole"] = true
		} else {
			step["part"] = beat.Anatomy.Part
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneAnatomy,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":   in.Plan.Title,
			"subject": a.Subject,
			"parts":   parts,
			"steps":   steps,
		},
	}}, nil
}
