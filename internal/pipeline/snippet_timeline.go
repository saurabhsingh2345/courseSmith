package pipeline

// The timeline template: a spine that fills in as the narrator walks it.
//
// The tenth template, and the one that covers "and then". A request's
// lifecycle, a release history, the stages of a build, the steps of a protocol
// — sequences where *position along the line* is the information, and where the
// thing you want at the end is the whole run visible at once.
//
// The whiteboard also accumulates, and the difference is worth stating because
// it decides which one a clip should use. A board is a *picture*: items land
// where the layout puts them and the arrows say how they relate. A timeline is
// an *axis*: the third milestone is after the second because it is further
// along, and that is the only claim it makes. When the subject is order, the
// axis says it without a single arrow; when the subject is structure, the board
// says it and the axis flattens it into a queue it never was.
//
// So the one rule that matters here is monotonicity. A clip that walks back to
// an earlier milestone is not narrating a timeline — it is narrating a diagram,
// and it should have been one.

import (
	"fmt"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "timeline",
		Title:       "Timeline",
		Description: "A spine of milestones filling in as you walk it — for anything whose subject is order.",
		Example:     "What happens between typing a URL and seeing the page",
		PromptFile:  snippetTimelineTemplateName,
		NeedsCode:   false,
		// One beat per milestone plus an opening is four at the floor, and a
		// milestone that gets less than a few seconds is a label nobody read.
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		Owns:             beatFields{Timeline: true},
		OwnsPlan:         planFields{Timeline: true},
		Normalize:        normalizeTimelinePlan,
		Validate:         validateTimelinePlan,
		Scenes:           timelineScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Figures":       strings.Join(ArtFigureNames(), ", "),
				"MinMilestones": minMilestones,
				"MaxMilestones": maxMilestones,
				"MaxMarkWords":  maxMilestoneMarkWords,
				"MaxTitleWords": maxMilestoneTitleWords,
				"MaxNoteWords":  maxMilestoneNoteWords,
			}
		},
	})
}

const snippetTimelineTemplateName = "snippet_timeline.tmpl"

const (
	// Three is the floor: two points is a before/after, which the compare
	// template does better. Six is the ceiling because the spine runs across
	// the frame and a seventh label has nowhere legible to sit.
	minMilestones = 3
	maxMilestones = 6

	maxMilestoneMarkWords  = 3
	maxMilestoneTitleWords = 5
	maxMilestoneNoteWords  = 16
)

// TimelineSpec is the run of milestones. On the plan rather than per-beat for
// the reason the quiz's question is: the sequence is the subject of the clip.
type TimelineSpec struct {
	Milestones []Milestone `json:"milestones"`
}

// Milestone is one stop on the spine.
type Milestone struct {
	// Mark is the position label — a year, a step number, a duration, a stage
	// name. Short: it sits under the dot.
	Mark string `json:"mark"`
	// Title is what happens here.
	Title string `json:"title"`
	// Note expands on it, and is shown only while this milestone is current.
	Note string `json:"note"`
	// Figure names an artwork figure drawn above the dot.
	Figure string `json:"figure,omitempty"`
}

// TimelineBeat says which milestone this beat is standing on.
type TimelineBeat struct {
	// At indexes TimelineSpec.Milestones.
	At int `json:"at"`
	// Whole marks the closing beat that shows the finished run with nothing
	// singled out. Present for the same reason the anatomy template has one:
	// `at` omitted decodes to 0, which is a real milestone.
	Whole bool `json:"whole,omitempty"`
}

func normalizeTimelinePlan(p *SnippetPlan) {
	if t := p.Timeline; t != nil {
		for i := range t.Milestones {
			m := &t.Milestones[i]
			m.Mark = clampWords(collapseSpaces(m.Mark), maxMilestoneMarkWords)
			m.Title = clampWords(collapseSpaces(m.Title), maxMilestoneTitleWords)
			m.Note = clampWords(collapseSpaces(m.Note), maxMilestoneNoteWords)
			if strings.TrimSpace(m.Figure) != "" {
				m.Figure = normalizeArtFigure(m.Figure)
			}
		}
	}
	for i := range p.Beats {
		if b := p.Beats[i].Timeline; b != nil && b.At < 0 {
			b.At, b.Whole = 0, true
		}
	}
}

func validateTimelinePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Timeline: true}); err != nil {
		return err
	}

	t := p.Timeline
	if t == nil {
		return fmt.Errorf("the plan has no timeline — this template is a run of milestones walked in order")
	}
	if n := len(t.Milestones); n < minMilestones || n > maxMilestones {
		return fmt.Errorf("there are %d milestones, want %d-%d — two points is a before-and-after, which the compare template does better",
			n, minMilestones, maxMilestones)
	}
	seen := map[string]bool{}
	for i, m := range t.Milestones {
		if strings.TrimSpace(m.Mark) == "" {
			return fmt.Errorf("milestone %d has no mark — give it the year, step or stage that places it on the line", i)
		}
		if strings.TrimSpace(m.Title) == "" {
			return fmt.Errorf("milestone %d (%q) has no title — say what happens there", i, m.Mark)
		}
		key := strings.ToLower(strings.TrimSpace(m.Title))
		if seen[key] {
			return fmt.Errorf("milestone %d repeats the title %q — each stop on the line is a different thing", i, m.Title)
		}
		seen[key] = true
	}

	visited := map[int]bool{}
	last := -1
	sawWhole := false
	for i, b := range p.Beats {
		if b.Timeline == nil {
			return fmt.Errorf("beat %q has no timeline direction — every beat is standing somewhere on the line", b.ID)
		}
		if b.Timeline.Whole {
			sawWhole = true
			continue
		}
		if b.Timeline.At < 0 || b.Timeline.At >= len(t.Milestones) {
			return fmt.Errorf("beat %q stands on milestone %d, which does not exist", b.ID, b.Timeline.At)
		}
		// The rule the template rests on. Walking back is not a timeline; it is
		// a diagram being narrated as though it were one.
		if b.Timeline.At < last {
			return fmt.Errorf("beat %q goes back to milestone %d after %d. A timeline only moves forward — if the story genuinely revisits an earlier point, it is a diagram and the flow or whiteboard template will tell it properly",
				b.ID, b.Timeline.At, last)
		}
		if visited[b.Timeline.At] {
			return fmt.Errorf("beat %q stands on milestone %d again; each stop gets one beat", b.ID, b.Timeline.At)
		}
		visited[b.Timeline.At] = true
		last = b.Timeline.At
		_ = i
	}
	if len(visited) != len(t.Milestones) {
		return fmt.Errorf("%d of the %d milestones are never reached — a stop nobody narrates is a dot with a label on it",
			len(t.Milestones)-len(visited), len(t.Milestones))
	}
	// The finished run, with everything visible at once, is what the viewer
	// takes away. Ending on the last milestone alone leaves the line still
	// looking like it is mid-walk.
	if !sawWhole {
		return fmt.Errorf("no beat shows the finished line. Close with a beat carrying \"whole\": true — the whole run visible at once is what the viewer leaves with")
	}
	if lastBeat := p.Beats[len(p.Beats)-1].Timeline; lastBeat != nil && !lastBeat.Whole {
		return fmt.Errorf("the clip ends standing on one milestone; end on the finished line instead (\"whole\": true)")
	}
	return nil
}

// timelineScenes lays the clip out as ONE scene: the spine is on screen for the
// whole clip and the beats only move where it has filled to.
func timelineScenes(in SnippetSceneInput) ([]Scene, error) {
	t := in.Plan.Timeline
	if t == nil {
		return nil, fmt.Errorf("the plan has no timeline")
	}

	milestones := make([]map[string]any, len(t.Milestones))
	for i, m := range t.Milestones {
		milestones[i] = map[string]any{
			"mark":   m.Mark,
			"title":  m.Title,
			"note":   m.Note,
			"figure": normalizeArtFigure(m.Figure),
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Timeline == nil {
			return nil, fmt.Errorf("beat %q has no timeline direction", beat.ID)
		}
		step := map[string]any{"startMs": startMs, "endMs": endMs}
		if beat.Timeline.Whole {
			step["whole"] = true
		} else {
			step["at"] = beat.Timeline.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneTimeline,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":      in.Plan.Title,
			"milestones": milestones,
			"steps":      steps,
		},
	}}, nil
}
