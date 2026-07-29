package pipeline

// The breakdown template: a path with phases, and each phase opens.
//
// "How do I actually build one of these" is the question a course exists to
// answer, and the answer is always the same shape: a handful of phases, in
// order, each of which is a small world of its own. Design, wireframe, front
// end, back end. Or: read the spec, model the data, write the handler, test it,
// ship it. The list is the map, and the substance is inside the stops.
//
// Nothing in the catalog could hold both halves. `timeline` is the closest and
// it is deliberately thin — its claim is *order*, it gives a milestone one note,
// and a note is not a set of tools. `stack` has the depth but no sequence: its
// tiers coexist, and you do not "do" the frontend tier before the data tier.
// `whiteboard` accumulates but flattens everything onto one plane. What was
// missing is a structure with two levels, where the top level is a sequence and
// the second level is a bag of things worth naming.
//
// So a beat can stand on a phase or on one item inside it, and that is what
// makes this the longest template in the catalog — the beat count is a property
// of the content rather than of how long one picture stays interesting, which is
// exactly the case SnippetTemplate.MaxBeats exists for.
//
// An item beat may only spotlight something in the phase that is currently open.
// That is not a layout rule: a clip that reaches back into an earlier phase to
// pull out a tool has stopped walking a path and started browsing a catalogue,
// and the whole value of this shape is that it goes somewhere.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "breakdown",
		Category:    CatSystems,
		Title:       "Phase breakdown",
		Description: "A path of phases where each one opens into its own detail — the description, and the tools or techniques inside it.",
		Example:     "Everything it takes to build a complete website with no-code tools",
		PromptFile:  snippetBreakdownTemplateName,
		NeedsCode:   false,
		// Three phases and a closing overview is four beats, which a 40-second
		// clip can fund. The interesting version is six phases with items
		// pulled out of them, and that is what the raised ceiling is for.
		MinTargetSec:     40,
		DefaultTargetSec: 95,
		MaxBeats:         12,
		Owns:             beatFields{Breakdown: true},
		OwnsPlan:         planFields{Breakdown: true},
		Normalize:        normalizeBreakdownPlan,
		Validate:         validateBreakdownPlan,
		Scenes:           breakdownScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":          strings.Join(BreakdownShows(), ", "),
				"Icons":          strings.Join(PointIconNames(), ", "),
				"MinPhases":      minBreakdownPhases,
				"MaxPhases":      maxBreakdownPhases,
				"MinItems":       minPhaseItems,
				"MaxItems":       maxPhaseItems,
				"MaxTitleWords":  maxPhaseTitleWords,
				"MaxDetailWords": maxPhaseDetailWords,
				"MaxItemWords":   maxPhaseItemWords,
				"MaxItemNoteWds": maxPhaseItemNoteWords,
			}
		},
	})
}

const snippetBreakdownTemplateName = "snippet_breakdown.tmpl"

// Path capacity. Six phases is what the column holds once one of them is open
// at full height; three is the floor because a two-phase path is a before and an
// after, and `compare` weighs those properly. Five items across an open phase
// leaves 300px each, which is what a name plus a qualifier needs.
const (
	minBreakdownPhases = 3
	maxBreakdownPhases = 6
	minPhaseItems      = 2
	maxPhaseItems      = 5

	maxPhaseTitleWords    = 4
	maxPhaseDetailWords   = 14
	maxPhaseItemWords     = 2
	maxPhaseItemNoteWords = 6
)

// breakdownShows is the closed vocabulary of what a beat is standing on.
var breakdownShows = map[string]bool{
	"phase": true,
	"item":  true,
	"whole": true,
}

// BreakdownShows returns the vocabulary sorted, for prompts and docs.
func BreakdownShows() []string {
	out := make([]string, 0, len(breakdownShows))
	for k := range breakdownShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// BreakdownSpec is the whole path. On the plan rather than per-beat for the
// reason the timeline's milestones are: the path is the subject of the clip and
// the beats only walk it.
type BreakdownSpec struct {
	// Phases are the stages, in the order they are done.
	Phases []BreakdownPhase `json:"phases"`
}

// BreakdownPhase is one stage, and everything inside it.
type BreakdownPhase struct {
	// Title is the stage — "Design", "Wireframe", "Back end".
	Title string `json:"title"`
	// Detail is what this stage actually involves, in one line. Shown when the
	// phase is open.
	Detail string `json:"detail"`
	// Items are the things worth naming inside it: the tools you could use, the
	// techniques, the checks. This is the second level, and it is the reason
	// this template exists rather than a timeline.
	Items []PhaseItem `json:"items"`
}

// PhaseItem is one thing inside a phase.
type PhaseItem struct {
	// Name is the tool or technique.
	Name string `json:"name"`
	// Note is the qualifier that makes it a recommendation rather than a list
	// entry — when to reach for this one instead of its neighbour.
	Note string `json:"note,omitempty"`
	// Icon is a PointIconNames name drawn in the item's chip.
	Icon string `json:"icon,omitempty"`
}

// ResolvedIcon returns the icon drawn in the item's chip.
func (i PhaseItem) ResolvedIcon() string {
	if icon := normalizePointIconName(i.Icon); icon != "" {
		return icon
	}
	return "box"
}

// BreakdownBeat says where in the two-level structure this beat is standing.
type BreakdownBeat struct {
	// Show is one of breakdownShows.
	Show string `json:"show"`
	// At indexes BreakdownSpec.Phases, for a "phase" or "item" beat.
	At int `json:"at,omitempty"`
	// Item indexes that phase's Items, for an "item" beat.
	Item int `json:"item,omitempty"`
}

// ResolvedShow returns the beat's level, defaulting the unknown to a phase.
func (b BreakdownBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if breakdownShows[s] {
		return s
	}
	return "phase"
}

func normalizeBreakdownPlan(p *SnippetPlan) {
	bd := p.Breakdown
	if bd == nil {
		return
	}
	for i := range bd.Phases {
		ph := &bd.Phases[i]
		ph.Title = clampWords(collapseSpaces(ph.Title), maxPhaseTitleWords)
		ph.Detail = clampWords(collapseSpaces(ph.Detail), maxPhaseDetailWords)
		items := make([]PhaseItem, 0, len(ph.Items))
		for _, it := range ph.Items {
			it.Name = clampWords(collapseSpaces(it.Name), maxPhaseItemWords)
			it.Note = clampWords(collapseSpaces(it.Note), maxPhaseItemNoteWords)
			it.Icon = it.ResolvedIcon()
			// An unnamed item is a chip with a blank line under it. Dropping it
			// is the repair; naming it would be inventing a recommendation.
			if it.Name != "" && len(items) < maxPhaseItems {
				items = append(items, it)
			}
		}
		ph.Items = items
	}

	for i := range p.Beats {
		b := p.Beats[i].Breakdown
		if b == nil {
			continue
		}
		// The last beat is the overview by the shape of the template, so an
		// unlabelled one there is inferrable; anything else defaults to a phase,
		// which is the level that must exist for an item to hang off.
		if !breakdownShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			if i == len(p.Beats)-1 {
				b.Show = "whole"
			} else {
				b.Show = "phase"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show == "whole" {
			b.At, b.Item = 0, 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(bd.Phases); n > 0 && b.At >= n {
			b.At = n - 1
		}
		if b.Show == "phase" {
			b.Item = 0
			continue
		}
		if b.Item < 0 {
			b.Item = 0
		}
		if n := len(bd.Phases[b.At].Items); n > 0 && b.Item >= n {
			b.Item = n - 1
		}
	}
}

func validateBreakdownPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Breakdown: true}); err != nil {
		return err
	}

	bd := p.Breakdown
	if bd == nil {
		return fmt.Errorf("the plan has no breakdown — this template is a path of phases, each of which opens")
	}
	if n := len(bd.Phases); n < minBreakdownPhases || n > maxBreakdownPhases {
		return fmt.Errorf("the path has %d phases, want %d-%d — a two-phase path is a before and an after, and compare weighs those properly",
			n, minBreakdownPhases, maxBreakdownPhases)
	}
	seenPhase := map[string]bool{}
	for i, ph := range bd.Phases {
		if strings.TrimSpace(ph.Title) == "" {
			return fmt.Errorf("phase %d has no title — name the stage", i)
		}
		key := strings.ToLower(strings.TrimSpace(ph.Title))
		if seenPhase[key] {
			return fmt.Errorf("phase %d repeats the title %q — each stage is a different piece of work", i, ph.Title)
		}
		seenPhase[key] = true
		if strings.TrimSpace(ph.Detail) == "" {
			return fmt.Errorf("phase %q has no detail — say what the stage actually involves, in one line", ph.Title)
		}
		if n := len(ph.Items); n < minPhaseItems || n > maxPhaseItems {
			return fmt.Errorf("phase %q holds %d items, want %d-%d. The items are the reason this is not a timeline — name the tools or techniques somebody would actually reach for",
				ph.Title, n, minPhaseItems, maxPhaseItems)
		}
		seenItem := map[string]bool{}
		for _, it := range ph.Items {
			if strings.TrimSpace(it.Name) == "" {
				return fmt.Errorf("phase %q has an item with no name", ph.Title)
			}
			k := strings.ToLower(strings.TrimSpace(it.Name))
			if seenItem[k] {
				return fmt.Errorf("phase %q lists %q twice", ph.Title, it.Name)
			}
			seenItem[k] = true
		}
	}

	opened := map[int]bool{}
	spotlit := map[[2]int]bool{}
	current := -1
	sawWhole := false
	for _, b := range p.Beats {
		if b.Breakdown == nil {
			return fmt.Errorf("beat %q has no breakdown direction — every beat is on a phase, on an item, or showing the whole path", b.ID)
		}
		switch b.Breakdown.ResolvedShow() {
		case "whole":
			sawWhole = true
		case "phase":
			at := b.Breakdown.At
			if at < 0 || at >= len(bd.Phases) {
				return fmt.Errorf("beat %q opens phase %d, which does not exist", b.ID, at)
			}
			if opened[at] {
				return fmt.Errorf("beat %q opens phase %d again; each stage is opened once and then its items are walked", b.ID, at)
			}
			// The path goes forward. A clip that returns to an earlier stage is
			// not walking a path, it is browsing a list of stages.
			if at < current {
				return fmt.Errorf("beat %q goes back to phase %d after %d. The path only moves forward — if the clip genuinely revisits an earlier stage, these are not phases of one job and the whiteboard will draw them better",
					b.ID, at, current)
			}
			opened[at] = true
			current = at
		case "item":
			at, item := b.Breakdown.At, b.Breakdown.Item
			if at < 0 || at >= len(bd.Phases) {
				return fmt.Errorf("beat %q spotlights an item of phase %d, which does not exist", b.ID, at)
			}
			// An item belongs to the phase that is open. Reaching back into an
			// earlier one is browsing a catalogue rather than walking a path.
			if at != current {
				if !opened[at] {
					return fmt.Errorf("beat %q spotlights an item of phase %d before that phase has been opened — open the stage first, then walk what is inside it", b.ID, at)
				}
				return fmt.Errorf("beat %q reaches back into phase %d while phase %d is open. Finish a stage before moving on; the items belong to the phase you are standing in",
					b.ID, at, current)
			}
			if item < 0 || item >= len(bd.Phases[at].Items) {
				return fmt.Errorf("beat %q spotlights item %d of phase %q, which does not exist", b.ID, item, bd.Phases[at].Title)
			}
			if spotlit[[2]int{at, item}] {
				return fmt.Errorf("beat %q spotlights %q again; each item gets at most one beat", b.ID, bd.Phases[at].Items[item].Name)
			}
			spotlit[[2]int{at, item}] = true
		}
	}
	if len(opened) != len(bd.Phases) {
		return fmt.Errorf("%d of the %d phases are never opened — a stage nobody explains is a row with a number on it",
			len(bd.Phases)-len(opened), len(bd.Phases))
	}
	if !sawWhole {
		return fmt.Errorf("no beat shows the whole path. Close with a beat carrying {\"show\": \"whole\"} — the finished list, all of it visible at once, is what the viewer leaves with")
	}
	if last := p.Beats[len(p.Beats)-1].Breakdown; last.ResolvedShow() != "whole" {
		return fmt.Errorf("the clip ends inside one phase; end on the whole path instead ({\"show\": \"whole\"})")
	}
	return nil
}

// breakdownScenes lays the clip out as ONE scene: the path is on screen for the
// whole clip and the beats only move which phase is open and what is spotlit.
func breakdownScenes(in SnippetSceneInput) ([]Scene, error) {
	bd := in.Plan.Breakdown
	if bd == nil {
		return nil, fmt.Errorf("the plan has no breakdown")
	}

	phases := make([]map[string]any, len(bd.Phases))
	for i, ph := range bd.Phases {
		items := make([]map[string]any, len(ph.Items))
		for j, it := range ph.Items {
			items[j] = map[string]any{
				"name": it.Name,
				"note": it.Note,
				"icon": it.ResolvedIcon(),
			}
		}
		phases[i] = map[string]any{
			"title":  ph.Title,
			"detail": ph.Detail,
			"items":  items,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Breakdown == nil {
			return nil, fmt.Errorf("beat %q has no breakdown direction", beat.ID)
		}
		show := beat.Breakdown.ResolvedShow()
		step := map[string]any{"startMs": startMs, "endMs": endMs, "show": show}
		if show != "whole" {
			step["at"] = beat.Breakdown.At
		}
		if show == "item" {
			step["item"] = beat.Breakdown.Item
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneBreakdown,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":  in.Plan.Title,
			"phases": phases,
			"steps":  steps,
		},
	}}, nil
}
