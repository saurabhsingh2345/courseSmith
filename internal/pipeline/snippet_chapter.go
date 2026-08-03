package pipeline

// The chapter template: where you are, and what opens next.
//
// A course is not a run of clips, it is a journey somebody is part-way through,
// and the catalog had nothing that said so. Every other template in here is
// about the *subject*; this one is about the viewer. It is the punctuation
// between two stretches of teaching — the frame that says three of seven are
// behind you, this is the one starting now, and here is why it follows from
// what you just did.
//
// `timeline` walks the milestones of a subject and `breakdown` opens the phases
// of a process. Both draw a path and both are content. A chapter break is
// furniture: the path is small, the ordinal is enormous, and the picture is a
// card rather than a diagram. It is the difference between a map of the Roman
// empire and the "PART THREE" title in the middle of a documentary about it.
//
// Two rules earn it, and they are the same rule seen from either end.
//
// The marker only moves FORWARD. A break that walks back through the whole
// course is a recap, which is a different clip with a different job — the whole
// value of this one is that it is short, and the moment it starts re-teaching
// it stops being punctuation.
//
// And the clip ENDS on what opens next. A break that closes by summarising what
// just finished leaves the viewer at a stopping point: the words that carry
// somebody across a gap are the ones about the thing on the other side of it.
// So `here` is the last beat, always, and everything looking backwards happens
// before it.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "chapter",
		Category:    CatPresenting,
		Since:       SinceV2,
		Title:       "Where you are",
		Description: "A huge ordinal, the section that is starting, and the path of everything else — behind you ticked off, ahead of you still faint. Reach for it between two stretches of teaching, when the viewer needs to know how far in they are and what opens now.",
		Example:     "Part three of a Python course: variables and print are done, now we start loops",
		PromptFile:  snippetChapterTemplateName,
		NeedsCode:   false,
		// Punctuation, not a lesson. The whole shape is the path, one look
		// back and the handover, which is three beats and about half a minute.
		MinTargetSec:     15,
		DefaultTargetSec: 30,
		Owns:             beatFields{Chapter: true},
		OwnsPlan:         planFields{Chapter: true},
		Normalize:        normalizeChapterPlan,
		Validate:         validateChapterPlan,
		Scenes:           chapterScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":         strings.Join(ChapterShows(), ", "),
				"Icons":         strings.Join(PointIconNames(), ", "),
				"MinStops":      minChapterStops,
				"MaxStops":      maxChapterStops,
				"MaxPathWords":  maxChapterPathWords,
				"MaxLabelWords": maxChapterLabelWords,
				"MaxNoteWords":  maxChapterNoteWords,
			}
		},
	})
}

const snippetChapterTemplateName = "snippet_chapter.tmpl"

const (
	// Three stops is the least that has a behind, a here and an ahead — which
	// is the whole claim this template makes. Past seven the path is a row of
	// dots nobody can count at a glance, and counting them is the point.
	minChapterStops = 3
	maxChapterStops = 7

	maxChapterPathWords  = 5
	maxChapterLabelWords = 5
	maxChapterNoteWords  = 16
)

// chapterShows is the closed vocabulary of what a beat does.
var chapterShows = map[string]bool{
	// Draw the whole path and land the marker on where we are. The first beat.
	"path": true,
	// Look back at one stop already behind us. Only backwards, and only before
	// the handover.
	"done": true,
	// The section that is starting. The last beat, always.
	"here": true,
}

// ChapterShows returns the beat vocabulary sorted.
func ChapterShows() []string {
	out := make([]string, 0, len(chapterShows))
	for k := range chapterShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ChapterSpec is the run this break sits inside.
type ChapterSpec struct {
	// Path is what the whole run is called — "The Python course", "Building the
	// API". It is the eyebrow, and without it the break is a title card for a
	// section of nothing.
	Path string `json:"path"`
	// Stops are every part of the run, in order.
	Stops []ChapterStop `json:"stops"`
	// At indexes the stop this break sits at: the one that is starting now.
	At int `json:"at"`
}

// ChapterStop is one part of the run.
type ChapterStop struct {
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
	// Note is the one line about this stop, shown when it is looked back at or
	// when it is the one starting.
	Note string `json:"note,omitempty"`
}

// ChapterBeat is one move.
type ChapterBeat struct {
	Show string `json:"show"`
	// At indexes the stop a `done` beat looks back at.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to the
// handover — which is the beat every one of these clips has.
func (b ChapterBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if chapterShows[s] {
		return s
	}
	return "here"
}

// Ordinal is the number this stop wears on screen, counting from one. The
// picture says "part three", not "part index two".
func (s ChapterSpec) Ordinal() int { return s.At + 1 }

func normalizeChapterPlan(p *SnippetPlan) {
	c := p.Chapter
	if c == nil {
		return
	}
	c.Path = clampWords(collapseSpaces(c.Path), maxChapterPathWords)

	stops := make([]ChapterStop, 0, len(c.Stops))
	for _, s := range c.Stops {
		s.Label = clampWords(collapseSpaces(s.Label), maxChapterLabelWords)
		s.Note = clampWords(collapseSpaces(s.Note), maxChapterNoteWords)
		s.Icon = normalizePointIconName(s.Icon)
		if s.Label != "" && len(stops) < maxChapterStops {
			stops = append(stops, s)
		}
	}
	c.Stops = stops

	// A break has to sit *somewhere* on its own path. Out of range is an index
	// mistake rather than a claim, so it is clamped rather than rejected.
	if c.At < 0 {
		c.At = 0
	}
	if n := len(c.Stops); n > 0 && c.At >= n {
		c.At = n - 1
	}

	for i := range p.Beats {
		b := p.Beats[i].Chapter
		if b == nil {
			continue
		}
		if !chapterShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			// The shape says what the ends are for: the clip opens on the path
			// and closes on the handover.
			if i == 0 {
				b.Show = "path"
			} else if i == len(p.Beats)-1 {
				b.Show = "here"
			} else {
				b.Show = "done"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "done" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(c.Stops); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateChapterPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Chapter: true}); err != nil {
		return err
	}

	c := p.Chapter
	if c == nil {
		return fmt.Errorf("the plan has no chapter — this template is a path, a position on it, and the section that opens next")
	}
	if strings.TrimSpace(c.Path) == "" {
		return fmt.Errorf("the path has no name. A break has to say what it is a break IN — without it this is a title card for a section of nothing")
	}
	if n := len(c.Stops); n < minChapterStops || n > maxChapterStops {
		return fmt.Errorf("there are %d stops, want %d-%d. Fewer than %d has no behind, here and ahead — which is the whole claim this picture makes; more than %d is a row of dots nobody can count at a glance",
			n, minChapterStops, maxChapterStops, minChapterStops, maxChapterStops)
	}
	if c.At < 0 || c.At >= len(c.Stops) {
		return fmt.Errorf("the break sits at stop %d, which is not on the path", c.At)
	}

	seen := map[string]bool{}
	for i, s := range c.Stops {
		if strings.TrimSpace(s.Label) == "" {
			return fmt.Errorf("stop %d has no label", i)
		}
		key := strings.ToLower(strings.TrimSpace(s.Label))
		if seen[key] {
			return fmt.Errorf("two stops are both called %q — a viewer counting their way along the path cannot tell which one they are at", s.Label)
		}
		seen[key] = true
	}

	counts := map[string]int{}
	looked := map[int]bool{}
	last := -1
	for i, b := range p.Beats {
		if b.Chapter == nil {
			return fmt.Errorf("beat %q has no chapter direction — every beat draws the path, looks back at one stop, or opens the next section", b.ID)
		}
		show := b.Chapter.ResolvedShow()
		counts[show]++

		if i == 0 && show != "path" {
			return fmt.Errorf("the clip opens on %q. Draw the path first — a position means nothing until the viewer can see what it is a position on", show)
		}
		if show == "here" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q opens the next section but the clip carries on afterwards. The handover is the LAST thing said: a break that ends by summarising what just finished leaves the viewer at a stopping point, and the words that carry somebody across a gap are the ones about what is on the other side of it", b.ID)
		}
		if show != "done" {
			continue
		}
		// The rule this template exists for, in its backwards half.
		if b.Chapter.At >= c.At {
			target := c.Stops[b.Chapter.At].Label
			return fmt.Errorf("beat %q looks back at %q, which is stop %d — at or ahead of the break at stop %d. The marker only moves forward: everything this beat can talk about is already behind the viewer",
				b.ID, target, b.Chapter.At, c.At)
		}
		if looked[b.Chapter.At] {
			return fmt.Errorf("beat %q looks back at %q again; each stop is recalled once", b.ID, c.Stops[b.Chapter.At].Label)
		}
		if b.Chapter.At < last {
			return fmt.Errorf("beat %q looks back at stop %d after stop %d. Even the looking back runs forwards — walking the path backwards is a recap, which is a different clip with a different job",
				b.ID, b.Chapter.At, last)
		}
		looked[b.Chapter.At] = true
		last = b.Chapter.At
	}

	if counts["path"] != 1 {
		return fmt.Errorf("there are %d path beats; the path is drawn once and then stays", counts["path"])
	}
	if counts["here"] != 1 {
		return fmt.Errorf("there are %d beats opening the next section, want exactly one — and it is the last", counts["here"])
	}
	// A break in the middle of a course that never mentions what closed is a
	// title card. At the very start there is nothing behind, and requiring a
	// look back there would be asking the model to invent one.
	if c.At > 0 && counts["done"] == 0 {
		return fmt.Errorf("the break sits at stop %d and never looks back at anything. %d stop(s) are behind the viewer; naming what closed is what makes this a break rather than a title card",
			c.At+1, c.At)
	}
	return nil
}

// chapterScenes lays the clip out as ONE scene. The card is standing furniture
// — the path, the ordinal and the title are on screen from the first frame —
// and the beats only move the light around it.
func chapterScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.Chapter
	if c == nil {
		return nil, fmt.Errorf("the plan has no chapter")
	}

	stops := make([]map[string]any, len(c.Stops))
	for i, s := range c.Stops {
		stops[i] = map[string]any{
			"label": s.Label,
			"icon":  s.Icon,
			"note":  s.Note,
			// Decided here rather than in the renderer so the tick a stop wears
			// and the rule the validator enforces can never disagree about what
			// counts as behind you.
			"state": chapterStopState(i, c.At),
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Chapter == nil {
			return nil, fmt.Errorf("beat %q has no chapter direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Chapter.ResolvedShow(),
		}
		if beat.Chapter.ResolvedShow() == "done" {
			step["at"] = beat.Chapter.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneChapter,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":   in.Plan.Title,
			"path":    c.Path,
			"stops":   stops,
			"at":      c.At,
			"ordinal": c.Ordinal(),
			"total":   len(c.Stops),
			"steps":   steps,
		},
	}}, nil
}

// chapterStopState is where a stop sits relative to the break.
func chapterStopState(i, at int) string {
	switch {
	case i < at:
		return "done"
	case i == at:
		return "here"
	default:
		return "ahead"
	}
}
