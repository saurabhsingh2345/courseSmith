package pipeline

// The progress template: what is behind you, what is in front of you, and the
// line between them.
//
// This is the frame a teacher draws at the end of a stretch of work, and the
// catalog had no way to draw it. Four templates come close and all four are
// about a COURSE rather than about work done. `chapter` and `waypoint` are
// openers: a huge ordinal, the section that is starting, and the syllabus down
// the side with sections ticked. `recap` brings back what earlier lessons
// established, every claim tagged with the lesson it came from. `checkpoint` is
// a task the viewer performs. All of them take their items from the table of
// contents, which is exactly what this frame does not do — "we installed the app
// and set up a Python project" is not a section title and it is not a claim with
// a source, and "authorization, getting data" is not a syllabus. They are things
// accomplished and things coming, and the units are different.
//
// So the composition is a single axis with NOW on it. Left of the line is
// finished and ticked. Right of the line is coming and not yet inked. That
// arrangement is the argument the narration is making — you are further along
// than you think, and here is what the rest looks like — and it is one no
// comparison can make, because a comparison has no direction.
//
// Two decisions earn the shape, and both are about what a beat moves.
//
// A beat moves a SIDE, not an item. Every other row template in this catalog
// lights one card per beat, because those templates are walking a list. This one
// is summarising, and a summary is spoken in whole sentences: "we installed
// Claude Desktop and we set up our first Python project" is one breath covering
// two accomplishments. Forcing a beat per item would either shred that sentence
// or pad it, so the done side arrives together and the next side arrives
// together. Two beats is the whole clip, and that is why this template can be
// cut to a narration somebody else wrote.
//
// The next side is present from the first frame, unlit. It is not withheld and
// it is not a spoiler: the viewer seeing that there IS a right-hand side, before
// hearing what is on it, is what makes the first beat feel like progress rather
// than like a stopping point. What the second beat does is ink it.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "progress",
		Category: CatPresenting,
		Since:    SinceV11,
		// Showroom: it is a paper frame with ticks on it, and the whole read
		// depends on inked-versus-not, which needs a ground that holds ink.
		Family:      FamilyShowroom,
		Title:       "So far, and next",
		Description: "One axis with NOW on it: what has been finished, ticked off on the left, and what is coming, still faint on the right. Reach for it at the end of a stretch of teaching, when the voice is summarising what got built and saying what follows.",
		Example:     "We installed the app and shipped a first project, and next comes customizing it, authorization and data",
		PromptFile:  snippetProgressTemplateName,
		NeedsCode:   false,
		// Two beats of a summarising sentence each. Eighteen seconds is the least
		// that holds "here is what we did" and "here is what is coming" without
		// either of them being a caption.
		MinTargetSec:     18,
		DefaultTargetSec: 30,
		// Past a minute the frame is being held with nothing moving: there are
		// only ever two sides to bring in.
		MaxTargetSec: 60,
		MaxBeats:     4,
		// A beat here is a whole SIDE of the frame read out in one breath, which
		// makes it the longest shot in the catalog. Thirty words is about ten
		// seconds of summary.
		IdealWordsPerBeat: 30,
		Owns:              beatFields{Progress: true},
		OwnsPlan:          planFields{Progress: true},
		Normalize:         normalizeProgressPlan,
		Validate:          validateProgressPlan,
		Scenes:            progressScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":        strings.Join(ProgressShows(), ", "),
				"Icons":        strings.Join(PointIconNames(), ", "),
				"MinItems":     minProgressItems,
				"MaxItems":     maxProgressItems,
				"MaxItemWords": maxProgressItemWords,
				"MaxNowWords":  maxProgressNowWords,
			}
		},
	})
}

const snippetProgressTemplateName = "snippet_progress.tmpl"

const (
	// One item a side is not a summary, it is a sentence. Four is where the rows
	// stop being readable at a glance, which is the only way this frame gets
	// read at all — it is on screen while a summary is being spoken over it.
	minProgressItems = 2
	maxProgressItems = 4

	// An item is a THING DONE or a THING COMING, named in the fewest words that
	// still name it: "Claude Desktop installed", "Authorization". Six is a
	// generous ceiling and most items want three.
	maxProgressItemWords = 6

	// The label on the line between the two sides — "you are here", "now".
	maxProgressNowWords = 3
)

// progressShows is the closed vocabulary of what a beat does. Two of the three
// are mandatory and the third is a held frame — see the file header on why a
// beat moves a side rather than an item.
var progressShows = map[string]bool{
	// The done side inks and its ticks strike in. The next side is on the frame,
	// unlit. The opener, always.
	"sofar": true,
	// The axis carries on past NOW and the next side inks.
	"ahead": true,
	// Both sides at full, the line complete. Optional, and the frame worth
	// pausing on.
	"all": true,
}

// ProgressShows returns the beat vocabulary sorted.
func ProgressShows() []string {
	out := make([]string, 0, len(progressShows))
	for k := range progressShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ProgressSpec is the axis: what is done, what is next, and the label on the
// line between them.
type ProgressSpec struct {
	// Done are the finished things, in the order they happened.
	Done []ProgressItem `json:"done"`
	// Next are the coming things, in the order they will be taught.
	Next []ProgressItem `json:"next"`
	// Now is the label on the line between the two sides. Optional — it defaults
	// to something true of every clip of this shape.
	Now string `json:"now,omitempty"`
}

// ResolvedNow returns the label on the line, never empty. A line with no label
// is a divider, and a divider says the two sides are different kinds of thing
// rather than one axis at two points in time.
func (s ProgressSpec) ResolvedNow() string {
	if n := strings.TrimSpace(s.Now); n != "" {
		return n
	}
	return "you are here"
}

// ProgressItem is one thing, done or coming.
type ProgressItem struct {
	// Label is the thing, named. Required.
	Label string `json:"label"`
	// Icon is a PointIconNames name, drawn beside a coming item. A finished item
	// wears a tick instead — that is what done looks like, and it is the
	// renderer's to draw rather than the plan's to choose.
	Icon string `json:"icon,omitempty"`
}

// ResolvedIcon returns the item's glyph, never empty.
func (i ProgressItem) ResolvedIcon() string {
	if icon := normalizePointIconName(i.Icon); icon != "" {
		return icon
	}
	return "arrow"
}

// ProgressBeat is one move: ink the done side, ink the next side, or hold both.
type ProgressBeat struct {
	Show string `json:"show"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to holding both
// sides — the only one of the three that is safe to guess, because the other two
// happen exactly once each and guessing either would duplicate it.
func (b ProgressBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if progressShows[s] {
		return s
	}
	return "all"
}

func normalizeProgressPlan(p *SnippetPlan) {
	g := p.Progress
	if g == nil {
		return
	}
	g.Now = clampWords(collapseSpaces(g.Now), maxProgressNowWords)

	clean := func(in []ProgressItem) []ProgressItem {
		out := make([]ProgressItem, 0, len(in))
		for _, it := range in {
			it.Label = clampWords(collapseSpaces(it.Label), maxProgressItemWords)
			it.Icon = it.ResolvedIcon()
			if it.Label != "" && len(out) < maxProgressItems {
				out = append(out, it)
			}
		}
		return out
	}
	g.Done = clean(g.Done)
	g.Next = clean(g.Next)

	// Defaulted by position: the first beat inks what is done and the second inks
	// what is coming. A plan that names neither meant the obvious.
	for i := range p.Beats {
		b := p.Beats[i].Progress
		if b == nil {
			continue
		}
		if progressShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			b.Show = b.ResolvedShow()
			continue
		}
		switch i {
		case 0:
			b.Show = "sofar"
		case 1:
			b.Show = "ahead"
		default:
			b.Show = "all"
		}
	}
}

func validateProgressPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Progress: true}); err != nil {
		return err
	}

	g := p.Progress
	if g == nil {
		return fmt.Errorf("the plan has no progress — this template is one axis with what is done on the left of NOW and what is coming on the right")
	}

	sides := []struct {
		name  string
		items []ProgressItem
		why   string
	}{
		{"done", g.Done, "what has been finished"},
		{"next", g.Next, "what is coming"},
	}
	for _, side := range sides {
		if n := len(side.items); n < minProgressItems || n > maxProgressItems {
			return fmt.Errorf("the %s side has %d items (%s), want %d-%d. One item is a sentence rather than a summary, and past %d the rows stop being readable at a glance — which is all this frame gets, because a summary is being spoken over it",
				side.name, n, side.why, minProgressItems, maxProgressItems, maxProgressItems)
		}
		seen := map[string]bool{}
		for i, it := range side.items {
			label := strings.TrimSpace(it.Label)
			if label == "" {
				return fmt.Errorf("%s item %d has no label", side.name, i)
			}
			if n := len(strings.Fields(label)); n > maxProgressItemWords {
				return fmt.Errorf("%s item %d is %d words (%q); at most %d. Name the thing — a row that runs to a sentence is being read instead of the voice being listened to",
					side.name, i, n, label, maxProgressItemWords)
			}
			key := strings.ToLower(label)
			if seen[key] {
				return fmt.Errorf("the %s side lists %q twice", side.name, label)
			}
			seen[key] = true
		}
	}

	// The same thing cannot be both finished and coming. Worth its own check
	// rather than folding into the per-side one: a plan with "getting data" on
	// both sides has not merely repeated itself, it has told the viewer that the
	// work they just did is still ahead of them, which is the one thing this
	// frame exists to deny.
	done := map[string]bool{}
	for _, it := range g.Done {
		done[strings.ToLower(strings.TrimSpace(it.Label))] = true
	}
	for _, it := range g.Next {
		if done[strings.ToLower(strings.TrimSpace(it.Label))] {
			return fmt.Errorf("%q is on both sides of NOW, so the frame says the thing that was just finished is still ahead. Put it on one side", it.Label)
		}
	}

	counts := map[string]int{}
	sofarAt, aheadAt := -1, -1
	for i, b := range p.Beats {
		if b.Progress == nil {
			return fmt.Errorf("beat %q has no progress direction — every beat here inks the done side, inks the next side, or holds both", b.ID)
		}
		show := b.Progress.ResolvedShow()
		counts[show]++
		switch show {
		case "sofar":
			sofarAt = i
		case "ahead":
			aheadAt = i
		case "all":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q holds the finished frame but the clip carries on afterwards. That frame is the close", b.ID)
			}
		}
	}
	if counts["sofar"] != 1 {
		return fmt.Errorf("there are %d beats inking what is done; the done side arrives once, and it is the first thing on the frame", counts["sofar"])
	}
	if sofarAt != 0 {
		return fmt.Errorf("beat %q inks what is done, but it is beat %d. What has been finished comes first: this frame's argument is that the viewer is further along than they think, and it cannot be made after the work still to come is already on screen",
			p.Beats[sofarAt].ID, sofarAt+1)
	}
	if counts["ahead"] != 1 {
		return fmt.Errorf("there are %d beats inking what is coming; the next side arrives once", counts["ahead"])
	}
	if aheadAt != 1 {
		return fmt.Errorf("what is coming inks on beat %d. It follows the done side immediately — a held frame in between is the summary stalling at the halfway point", aheadAt+1)
	}
	return nil
}

// progressScenes lays the clip out as ONE scene: the axis is on screen
// throughout and the beats only decide which side of it is inked.
func progressScenes(in SnippetSceneInput) ([]Scene, error) {
	g := in.Plan.Progress
	if g == nil {
		return nil, fmt.Errorf("the plan has no progress")
	}

	items := func(src []ProgressItem) []map[string]any {
		out := make([]map[string]any, len(src))
		for i, it := range src {
			out[i] = map[string]any{"label": it.Label, "icon": it.ResolvedIcon()}
		}
		return out
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Progress == nil {
			return nil, fmt.Errorf("beat %q has no progress direction", beat.ID)
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Progress.ResolvedShow(),
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneProgress,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title": in.Plan.Title,
			"done":  items(g.Done),
			"next":  items(g.Next),
			"now":   g.ResolvedNow(),
			"steps": steps,
		}),
	}}, nil
}
