package pipeline

// The bridge template: from there to here — the hand-off between lessons.
//
// Every lesson after the first begins in the middle of something, and the
// catalog had no picture for that moment. `recap` looks backward only;
// `prereq` demands rather than acknowledges; `outcome` promises without
// connecting. A career-switcher three weeks into CS foundations does not need
// another list — they need to see that what they finished still counts, and
// that this lesson exists because of a specific hole the last one left. The
// picture is two columns and a chevron: the previous lesson's ground sliding
// in already ticked, and the gap this lesson fills arriving as an open slot.
//
// The open slot is the template's one idea. Established knowledge is drawn
// solid and ticked; the gap is drawn as a dashed outline holding a QUESTION,
// because that is what a gap is — a thing the viewer can now ask and cannot
// yet answer. The slot filling when the new lesson names itself is the whole
// emotional payload: this lesson is the missing piece, visibly.
//
// The rules keep the hand-off honest.
//
// **The clip opens on the back side.** A gap only reads as a gap against
// ground the viewer recognises; an open slot with nothing established around
// it is just an empty box.
//
// **Every established item is carried exactly once.** The tick is the point —
// "this survives into the new lesson" said item by item. An item never carried
// is knowledge silently dropped; one carried twice spends a beat re-ticking a
// box the viewer watched tick.
//
// **The clip closes on the ahead side.** Naming the destination and then
// continuing to talk about the old lesson walks back across the bridge. Two
// or three established items is the band: one is not ground, it is a fact,
// and four belongs to recap.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "bridge",
		Category:    CatPresenting,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "From there to here",
		Description: "The hand-off between lessons: what the previous one established slides in ticked, and the gap this one fills arrives as an open slot. Reach for it as the first clip of any lesson that builds on the one before.",
		Example:     "From Binary & Data into Memory & Storage: bits, bytes and encodings are done, but where do the bytes live?",
		PromptFile:  snippetBridgeTemplateName,
		NeedsCode:   false,
		// Back, two or three carries, the gap, and ahead: 20s carries the tight
		// version, 35s the usual one.
		MinTargetSec:     20,
		DefaultTargetSec: 35,
		MaxBeats:         7,
		Owns:             beatFields{Bridge: true},
		OwnsPlan:         planFields{Bridge: true},
		Normalize:        normalizeBridgePlan,
		Validate:         validateBridgePlan,
		Scenes:           bridgeScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":          strings.Join(MetricRoles(), ", "),
				"Shows":          strings.Join(BridgeShows(), ", "),
				"MinEstablished": minBridgeEstablished,
				"MaxEstablished": maxBridgeEstablished,
				"MaxNameWords":   maxBridgeNameWords,
				"MaxItemWords":   maxBridgeItemWords,
				"MaxGapWords":    maxBridgeGapWords,
			}
		},
	})
}

const snippetBridgeTemplateName = "snippet_bridge.tmpl"

const (
	// One established item is a fact, not ground to stand on; four is a recap,
	// and recap already draws that. Two or three is a hand-off.
	minBridgeEstablished = 2
	maxBridgeEstablished = 3

	// Lesson names are chips on the two columns, not sentences.
	maxBridgeNameWords = 6
	// An established item is one line in the left column; ten words states a
	// thing learned, more re-teaches it.
	maxBridgeItemWords = 10
	// The gap is a question the viewer can now ask; twelve words is a question,
	// more is a paragraph in a dashed box.
	maxBridgeGapWords = 12
)

// bridgeShows is the closed vocabulary of what a beat does.
var bridgeShows = map[string]bool{
	// The From side with its established items. The opener.
	"back": true,
	// Item At ticks: it survives into this lesson.
	"carry": true,
	// The open slot appears, holding the question.
	"gap": true,
	// The To side names itself and the slot fills. The closer.
	"ahead": true,
}

// BridgeShows returns the beat vocabulary sorted, for the prompt.
func BridgeShows() []string {
	out := make([]string, 0, len(bridgeShows))
	for k := range bridgeShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// BridgeSpec is the hand-off: the two lessons and what passes between them.
type BridgeSpec struct {
	// From names the previous lesson.
	From string `json:"from"`
	// To names this lesson.
	To string `json:"to"`
	// Established is what the previous lesson left behind, drawn top to bottom.
	Established []string `json:"established"`
	// Gap is the question this lesson answers — what the viewer can now ask
	// and cannot yet answer.
	Gap string `json:"gap"`
}

// BridgeBeat is one move across the hand-off.
type BridgeBeat struct {
	// Show is a bridgeShows name.
	Show string `json:"show"`
	// At indexes BridgeSpec.Established, for a "carry" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults an unknown direction to carrying an item, which is
// what most beats of this template do.
func (b BridgeBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if bridgeShows[s] {
		return s
	}
	return "carry"
}

func normalizeBridgePlan(p *SnippetPlan) {
	br := p.Bridge
	if br == nil {
		return
	}
	br.From = clampWords(collapseSpaces(br.From), maxBridgeNameWords)
	br.To = clampWords(collapseSpaces(br.To), maxBridgeNameWords)
	br.Gap = clampWords(collapseSpaces(br.Gap), maxBridgeGapWords)
	items := make([]string, 0, len(br.Established))
	for _, it := range br.Established {
		it = clampWords(collapseSpaces(it), maxBridgeItemWords)
		if it == "" {
			continue
		}
		if len(items) < maxBridgeEstablished {
			items = append(items, it)
		}
	}
	br.Established = items

	for i := range p.Beats {
		b := p.Beats[i].Bridge
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "carry" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(br.Established); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateBridgePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Bridge: true}); err != nil {
		return err
	}

	br := p.Bridge
	if br == nil {
		return fmt.Errorf("the plan has no bridge — this template is the hand-off between two lessons, so the established ground and the gap are the clip")
	}
	if strings.TrimSpace(br.From) == "" || strings.TrimSpace(br.To) == "" {
		return fmt.Errorf("the bridge does not name both lessons. A hand-off needs a From and a To — an unnamed side is a column the viewer cannot place in the course")
	}
	if n := len(br.Established); n < minBridgeEstablished || n > maxBridgeEstablished {
		return fmt.Errorf("the previous lesson establishes %d item(s), want %d-%d. One is a fact rather than ground to stand on, and four is a recap — which recap already draws",
			n, minBridgeEstablished, maxBridgeEstablished)
	}
	if strings.TrimSpace(br.Gap) == "" {
		return fmt.Errorf("the bridge has no gap. The gap is the question this lesson answers — the thing the viewer can now ask and cannot yet answer — and without it the clip is a recap with a chevron on it")
	}

	if p.Beats[0].Bridge == nil || p.Beats[0].Bridge.ResolvedShow() != "back" {
		return fmt.Errorf("beat %q does not open on the back side. A gap only reads as a gap against ground the viewer recognises — show what the last lesson established first",
			p.Beats[0].ID)
	}

	carried := map[int]bool{}
	gaps := 0
	for i, b := range p.Beats {
		if b.Bridge == nil {
			return fmt.Errorf("beat %q has no bridge direction — every beat looks back, carries an item, opens the gap, or names the lesson ahead", b.ID)
		}
		switch b.Bridge.ResolvedShow() {
		case "back":
			if i != 0 {
				return fmt.Errorf("beat %q looks back again part-way through. The back side is the opener; returning to it walks the viewer back across the bridge", b.ID)
			}
		case "carry":
			at := b.Bridge.At
			if at < 0 || at >= len(br.Established) {
				return fmt.Errorf("beat %q carries item %d, which does not exist", b.ID, at)
			}
			if carried[at] {
				return fmt.Errorf("beat %q carries item %d (%q) a second time. The tick already said it survives — re-ticking a ticked box spends a beat on nothing",
					b.ID, at, br.Established[at])
			}
			carried[at] = true
		case "gap":
			gaps++
		case "ahead":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q names the lesson ahead and then the clip keeps going. Arriving is the ending — anything said after it is said from the wrong side of the bridge", b.ID)
			}
		}
	}
	if gaps != 1 {
		return fmt.Errorf("there are %d gap beats, want exactly 1. The gap is the reason this lesson exists; a hand-off without it is a recap, and one with two has not decided what the lesson is for", gaps)
	}
	if p.Beats[len(p.Beats)-1].Bridge.ResolvedShow() != "ahead" {
		return fmt.Errorf("the clip does not close on the lesson ahead. The open slot filling as the new lesson names itself is the payoff — end there")
	}
	if len(carried) != len(br.Established) {
		return fmt.Errorf("%d of the %d established items are never carried. An item the narrator skips is knowledge silently dropped at the hand-off — give it a beat or cut it",
			len(br.Established)-len(carried), len(br.Established))
	}
	return nil
}

// bridgeScenes lays the clip out as ONE scene. The two columns persist; the
// steps only say which items have ticked and whether the slot is open or full.
func bridgeScenes(in SnippetSceneInput) ([]Scene, error) {
	br := in.Plan.Bridge
	if br == nil {
		return nil, fmt.Errorf("the plan has no bridge")
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	carried := map[int]bool{}
	gapOpen := false
	arrived := false
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Bridge == nil {
			return nil, fmt.Errorf("beat %q has no bridge direction", beat.ID)
		}
		show := beat.Bridge.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "carry":
			carried[beat.Bridge.At] = true
			step["at"] = beat.Bridge.At
		case "gap":
			gapOpen = true
		case "ahead":
			gapOpen = true
			arrived = true
		}
		ticks := make([]int, 0, len(carried))
		for at := range carried {
			ticks = append(ticks, at)
		}
		sort.Ints(ticks)
		step["carried"] = ticks
		step["gapOpen"] = gapOpen
		step["arrived"] = arrived
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneBridge,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":       in.Plan.Title,
			"from":        br.From,
			"to":          br.To,
			"established": append([]string{}, br.Established...),
			"gap":         br.Gap,
			"steps":       steps,
		}),
	}}, nil
}
