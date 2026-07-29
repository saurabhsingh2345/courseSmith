package pipeline

// The rundown template: N things, promised and then delivered.
//
// This is the oldest shape in explanatory video and the catalog had no honest
// version of it. `points` throws bullets on a slide, which is a list; a rundown
// is a *contract*. The clip opens by saying how many things there are, the
// viewer decides on the spot whether to stay, and every card is a promise
// discharged. That opening number is the whole reason the format works, and it
// is also the thing that makes it fail when it is a lie.
//
// So the count is validated. If the promise says three numbers decide it, there
// are exactly three cards and each gets exactly one beat. A clip that announces
// four and shows five has broken the only agreement it made with the viewer,
// and — because the number is generated separately from the cards — that is
// precisely the failure a model produces by default.
//
// The other decision worth stating is that all the cards are on screen from the
// first frame, numbered and dim, and light one at a time. A rundown that reveals
// its items one by one is a list again: the viewer cannot see how far through
// they are, which is the single piece of information the format exists to give
// them.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "rundown",
		Category:    CatPresenting,
		Since:       SinceV1,
		Title:       "N things",
		Description: "A numbered row that promises how many there are and then delivers exactly that many.",
		Example:     "The three numbers that decide whether a model runs on your machine",
		PromptFile:  snippetRundownTemplateName,
		NeedsCode:   false,
		// The promise plus five items is six beats, which the shared budget
		// cannot fund at 45s.
		MinTargetSec:     35,
		DefaultTargetSec: 60,
		// The beat count here is a property of the content — a subject with five
		// items has five beats — so the ceiling goes above the shared one.
		MaxBeats:  9,
		Owns:      beatFields{Rundown: true},
		OwnsPlan:  planFields{Rundown: true},
		Normalize: normalizeRundownPlan,
		Validate:  validateRundownPlan,
		Scenes:    rundownScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":           strings.Join(RundownShows(), ", "),
				"Icons":           strings.Join(PointIconNames(), ", "),
				"MinItems":        minRundownItems,
				"MaxItems":        maxRundownItems,
				"MaxPromiseWords": maxRundownPromiseWords,
				"MaxLabelWords":   maxRundownLabelWords,
				"MaxDetailWords":  maxRundownDetailWords,
			}
		},
	})
}

const snippetRundownTemplateName = "snippet_rundown.tmpl"

const (
	// Two things is not a rundown, it is a comparison. Six cards across the
	// stage leaves each one 240px, which cannot hold a label and a detail.
	minRundownItems = 3
	maxRundownItems = 5

	maxRundownPromiseWords = 10
	maxRundownLabelWords   = 5
	maxRundownDetailWords  = 14
)

// rundownShows is the closed vocabulary of what a beat does.
var rundownShows = map[string]bool{
	// State the promise with every card numbered and dim. The first beat.
	"promise": true,
	// Light one card and open its detail.
	"item": true,
	// All of them lit at once. The closing frame.
	"all": true,
}

// RundownShows returns the beat vocabulary sorted.
func RundownShows() []string {
	out := make([]string, 0, len(rundownShows))
	for k := range rundownShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// RundownSpec is the promise and the things that discharge it.
type RundownSpec struct {
	// Promise is the opening line — "Three numbers decide everything". It is
	// set at headline size over the numbered row.
	Promise string        `json:"promise"`
	Items   []RundownItem `json:"items"`
}

// RundownItem is one numbered card.
type RundownItem struct {
	// Label is the thing itself, short enough to sit on a card.
	Label string `json:"label"`
	// Detail is the line that arrives when the card lights.
	Detail string `json:"detail,omitempty"`
	// Icon is a PointIconNames name drawn above the number.
	Icon string `json:"icon,omitempty"`
}

// ResolvedIcon returns the card's icon, defaulting to a neutral mark.
func (i RundownItem) ResolvedIcon() string {
	if icon := normalizePointIconName(i.Icon); icon != "" {
		return icon
	}
	return "box"
}

// RundownBeat is one move.
type RundownBeat struct {
	Show string `json:"show"`
	At   int    `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to lighting a
// card — which is what most beats of this template do.
func (b RundownBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if rundownShows[s] {
		return s
	}
	return "item"
}

// spelledNumbers maps the counts this template can promise onto the words a
// promise is actually written with. Only the range the item bounds allow, since
// a promise of "seven" cannot be honoured by a template that caps at five.
var spelledNumbers = map[string]int{
	"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
	"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
	"1": 1, "2": 2, "3": 3, "4": 4, "5": 5,
	"6": 6, "7": 7, "8": 8, "9": 9, "10": 10,
}

// promisedCount reads the number a promise announces, and whether it announced
// one at all.
//
// Written as a scan over the words rather than a regexp because the number can
// be spelled or written in digits and can sit anywhere in the line — "Three
// numbers decide it", "It comes down to 3 things".
func promisedCount(promise string) (int, bool) {
	for _, w := range strings.Fields(strings.ToLower(promise)) {
		w = strings.Trim(w, ",.;:!?\"'()")
		if n, ok := spelledNumbers[w]; ok {
			return n, true
		}
	}
	return 0, false
}

func normalizeRundownPlan(p *SnippetPlan) {
	r := p.Rundown
	if r == nil {
		return
	}
	r.Promise = clampWords(collapseSpaces(r.Promise), maxRundownPromiseWords)

	items := make([]RundownItem, 0, len(r.Items))
	for _, it := range r.Items {
		it.Label = clampWords(collapseSpaces(it.Label), maxRundownLabelWords)
		it.Detail = clampWords(collapseSpaces(it.Detail), maxRundownDetailWords)
		it.Icon = it.ResolvedIcon()
		if it.Label != "" && len(items) < maxRundownItems {
			items = append(items, it)
		}
	}
	r.Items = items

	for i := range p.Beats {
		b := p.Beats[i].Rundown
		if b == nil {
			continue
		}
		if !rundownShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			if i == 0 {
				b.Show = "promise"
			} else {
				b.Show = "item"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "item" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(r.Items); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateRundownPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Rundown: true}); err != nil {
		return err
	}

	r := p.Rundown
	if r == nil {
		return fmt.Errorf("the plan has no rundown — this template is a numbered promise and the things that discharge it")
	}
	if strings.TrimSpace(r.Promise) == "" {
		return fmt.Errorf("there is no promise. The clip opens by saying how many things there are; that number is what makes the viewer stay")
	}
	if n := len(r.Items); n < minRundownItems || n > maxRundownItems {
		return fmt.Errorf("there are %d items, want %d-%d. Two things is a comparison, and six cards across the stage leaves each one too narrow to carry a label and a line",
			n, minRundownItems, maxRundownItems)
	}

	// The rule this template exists for: the promise has to be true.
	if want, ok := promisedCount(r.Promise); ok && want != len(r.Items) {
		return fmt.Errorf("the promise %q announces %d, but there are %d cards. That number is the only agreement the clip makes with its viewer — either write %d cards or change the promise",
			r.Promise, want, len(r.Items), want)
	}

	seen := map[string]bool{}
	for i, it := range r.Items {
		if strings.TrimSpace(it.Label) == "" {
			return fmt.Errorf("item %d has no label", i)
		}
		key := strings.ToLower(strings.TrimSpace(it.Label))
		if seen[key] {
			return fmt.Errorf("two items are both %q — a rundown that repeats itself is padding the count", it.Label)
		}
		seen[key] = true
	}

	covered := map[int]bool{}
	counts := map[string]int{}
	for i, b := range p.Beats {
		if b.Rundown == nil {
			return fmt.Errorf("beat %q has no rundown direction — every beat makes the promise, covers one item, or brings them all back", b.ID)
		}
		show := b.Rundown.ResolvedShow()
		counts[show]++
		if i == 0 && show != "promise" {
			return fmt.Errorf("the clip opens on %q. Make the promise first — the count is what the viewer decides on, and a card lighting before it has been made is an item in a list nobody agreed to watch", show)
		}
		if show == "all" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q brings every card back but the clip carries on afterwards. That frame is the close", b.ID)
		}
		if show != "item" {
			continue
		}
		if b.Rundown.At < 0 || b.Rundown.At >= len(r.Items) {
			return fmt.Errorf("beat %q covers item %d, which does not exist", b.ID, b.Rundown.At)
		}
		if covered[b.Rundown.At] {
			return fmt.Errorf("beat %q covers item %d again; each card gets one beat", b.ID, b.Rundown.At)
		}
		covered[b.Rundown.At] = true
	}
	if counts["promise"] != 1 {
		return fmt.Errorf("there are %d promise beats; the count is announced once", counts["promise"])
	}
	if len(covered) != len(r.Items) {
		return fmt.Errorf("%d of the %d cards are never covered. The clip promised %d things and delivers %d, which is the one way this format loses a viewer's trust",
			len(r.Items)-len(covered), len(r.Items), len(r.Items), len(covered))
	}
	if counts["all"] > 1 {
		return fmt.Errorf("there are %d closing beats; the row comes back once", counts["all"])
	}
	return nil
}

// rundownScenes lays the clip out as ONE scene: the numbered row is on screen
// throughout and the beats only move which card is lit.
func rundownScenes(in SnippetSceneInput) ([]Scene, error) {
	r := in.Plan.Rundown
	if r == nil {
		return nil, fmt.Errorf("the plan has no rundown")
	}

	items := make([]map[string]any, len(r.Items))
	for i, it := range r.Items {
		items[i] = map[string]any{
			"label":  it.Label,
			"detail": it.Detail,
			"icon":   it.ResolvedIcon(),
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Rundown == nil {
			return nil, fmt.Errorf("beat %q has no rundown direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Rundown.ResolvedShow(),
		}
		if beat.Rundown.ResolvedShow() == "item" {
			step["at"] = beat.Rundown.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneRundown,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":   in.Plan.Title,
			"promise": r.Promise,
			"items":   items,
			"steps":   steps,
		},
	}}, nil
}
