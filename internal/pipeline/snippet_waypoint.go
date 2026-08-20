package pipeline

// The waypoint template: where you are in a long piece.
//
// This is the card that replaces `titlecard` at a chapter break in anything over
// about ten minutes, and the difference between them is the difference between a
// divider and a spine.
//
// A titlecard names the section starting now. That is enough in a three-minute
// clip and it is actively bad in a thirty-minute one, because the viewer's
// question at minute nineteen is not "what is this section called", it is "how
// much of this is left and does the thing I am confused about get answered". A
// card that cannot answer that leaves them to decide for themselves whether to
// keep watching, and some of them decide no.
//
// So this card carries THE WHOLE ARC. Every chapter of the piece is listed down
// the right edge: the ones behind you ticked, the one starting now lit, the ones
// ahead of you faint but legible. It costs a third of the frame and it buys the
// two things a long video needs most — the viewer always knows where they are,
// and they can see that the thing they are waiting for is coming.
//
// The composition is asymmetric, and that is the second half of the fix. The
// centred-serif-with-air layout that titlecard uses reads as a slide, and ten
// slides in one video reads as a slideshow. Here the type sits on a hard left
// axis with the ordinal above it and a rule under it, the promise below that, and
// the spine down the right. It is an editorial page rather than a title, which is
// the register a chapter opening actually wants.
//
// What it is NOT: a table of contents. The spine is peripheral by construction —
// small, low ink, unreadable at a glance except for the lit row. A viewer who
// wants it can read it; a viewer who does not sees a texture that tells them how
// far along they are. Set it at full strength and the card becomes a contents
// page with a headline attached, which is `syllabus`.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "waypoint",
		Category: CatPresenting,
		Since:    SinceV10,
		Family:   FamilyAtelier,
		Title:    "Where you are",
		Description: "The chapter opening for a long piece: the ordinal and title on a hard left axis, the promise under it, and the whole arc down the right edge with what is behind you ticked and what is ahead still faint. " +
			"Reach for it at every section break in anything over ten minutes; `titlecard` is the short-form divider.",
		Example:    "chapter three of nine: how it actually works",
		PromptFile: snippetWaypointTemplateName,
		NeedsCode:  false,
		MinTargetSec:     8,
		DefaultTargetSec: 14,
		MaxTargetSec:     34,
		// The arrival, the promise, the spine. A chapter card holds three thoughts.
		MaxBeats:          3,
		IdealWordsPerBeat: 20,
		Owns:              beatFields{Waypoint: true},
		OwnsPlan:          planFields{Waypoint: true},
		Normalize:         normalizeWaypointPlan,
		Validate:          validateWaypointPlan,
		Scenes:            waypointScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":           strings.Join(WaypointShows(), ", "),
				"MaxLineWords":    maxWaypointLineWords,
				"MaxPromiseWords": maxWaypointPromiseWords,
				"MinStops":        minWaypointStops,
				"MaxStops":        maxWaypointStops,
				"MaxStopWords":    maxWaypointStopWords,
			}
		},
	})
}

const snippetWaypointTemplateName = "snippet_waypoint.tmpl"

const (
	maxWaypointLineWords    = 7
	maxWaypointPromiseWords = 14
	// Three stops is an arc; past twelve the spine is a wall of small type and
	// stops being peripheral, which is the one property it has to keep.
	minWaypointStops  = 3
	maxWaypointStops  = 12
	maxWaypointStopWords = 5
	maxWaypointOrdinalChars = 6
)

// waypointShows is the closed vocabulary of what a beat does.
var waypointShows = map[string]bool{
	// The ordinal and the chapter title land on the left axis. The first beat.
	"arrive": true,
	// The promise under the rule.
	"promise": true,
	// The arc lights down the right edge, with this stop marked.
	"spine": true,
}

// WaypointShows returns the beat vocabulary sorted.
func WaypointShows() []string {
	out := make([]string, 0, len(waypointShows))
	for k := range waypointShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// WaypointSpec is the chapter opening.
type WaypointSpec struct {
	// Ordinal is the chapter mark — "03", "3 / 9". Optional.
	Ordinal string `json:"ordinal,omitempty"`
	// Line is the chapter title. Required.
	Line string `json:"line"`
	// Promise is what the viewer will have after this chapter.
	Promise string `json:"promise,omitempty"`
	// Stops are every chapter of the piece, in order — including this one.
	Stops []string `json:"stops"`
	// At is which stop this chapter is, counting from 0.
	At int `json:"at"`
}

// WaypointBeat is one shot of the card.
type WaypointBeat struct {
	Show string `json:"show"`
}

// ResolvedShow defaults to the arrival.
func (b WaypointBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if waypointShows[s] {
		return s
	}
	return "arrive"
}

func normalizeWaypointPlan(p *SnippetPlan) {
	w := p.Waypoint
	if w == nil {
		return
	}
	w.Ordinal = clampCodeLine(collapseSpaces(w.Ordinal), maxWaypointOrdinalChars)
	w.Line = clampWords(collapseSpaces(w.Line), maxWaypointLineWords)
	w.Promise = clampWords(collapseSpaces(w.Promise), maxWaypointPromiseWords)

	stops := make([]string, 0, maxWaypointStops)
	for _, s := range w.Stops {
		if s = clampWords(collapseSpaces(s), maxWaypointStopWords); s != "" && len(stops) < maxWaypointStops {
			stops = append(stops, s)
		}
	}
	w.Stops = stops
	if w.At < 0 {
		w.At = 0
	}
	if len(stops) > 0 && w.At >= len(stops) {
		w.At = len(stops) - 1
	}

	for i := range p.Beats {
		if b := p.Beats[i].Waypoint; b != nil {
			b.Show = b.ResolvedShow()
		}
	}
}

func validateWaypointPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Waypoint: true}); err != nil {
		return err
	}

	w := p.Waypoint
	if w == nil {
		return fmt.Errorf("the plan has no waypoint — this template is one chapter card, so the card is the clip")
	}
	if strings.TrimSpace(w.Line) == "" {
		return fmt.Errorf("the chapter has no title")
	}
	if n := len(w.Stops); n < minWaypointStops || n > maxWaypointStops {
		return fmt.Errorf("the arc has %d stops and wants %d-%d. The spine is the whole point of this card: with fewer than %d there is no arc to show and the card should be a titlecard instead",
			n, minWaypointStops, maxWaypointStops, minWaypointStops)
	}
	if w.At < 0 || w.At >= len(w.Stops) {
		return fmt.Errorf("this chapter is stop %d of %d", w.At, len(w.Stops))
	}
	// The lit stop and the title naming different things is the mistake that
	// makes the spine worse than useless — a viewer trusts it to say where they
	// are, and a card whose spine points at chapter four while its headline reads
	// chapter five has taught them not to.
	if lit := strings.TrimSpace(w.Stops[w.At]); lit != "" && !waypointAgrees(lit, w.Line) {
		return fmt.Errorf("the spine's lit stop is %q but the chapter title is %q. They name the same chapter, so one of them is wrong — the stop is the short form of the title",
			lit, w.Line)
	}

	shows := map[string]int{}
	for i, b := range p.Beats {
		if b.Waypoint == nil {
			return fmt.Errorf("beat %q has no waypoint direction", b.ID)
		}
		show := b.Waypoint.ResolvedShow()
		shows[show]++
		if i == 0 && show != "arrive" {
			return fmt.Errorf("the first beat is %q, but the chapter has to land before anything is added to it — the opening beat is \"arrive\"", show)
		}
	}
	if shows["arrive"] != 1 {
		return fmt.Errorf("there are %d \"arrive\" beats and there must be exactly one", shows["arrive"])
	}
	if shows["promise"] > 0 && strings.TrimSpace(w.Promise) == "" {
		return fmt.Errorf("a beat shows the promise but the card has none")
	}
	return nil
}

// waypointAgrees reports whether a spine stop and a chapter title are plausibly
// the same chapter: one is the short form of the other, so a shared significant
// word is enough. Loose on purpose — this catches a spine pointing at the wrong
// row, not a wording difference.
func waypointAgrees(stop, line string) bool {
	norm := func(s string) []string {
		s = strings.ToLower(s)
		var out []string
		for _, f := range strings.Fields(s) {
			f = strings.Trim(f, ".,:;!?()[]`\"'")
			switch f {
			case "", "a", "an", "the", "of", "in", "to", "and", "or", "your", "you", "it", "is", "how", "what":
				continue
			}
			out = append(out, f)
		}
		return out
	}
	sw, lw := norm(stop), norm(line)
	if len(sw) == 0 || len(lw) == 0 {
		return true
	}
	for _, a := range sw {
		for _, b := range lw {
			if a == b || strings.HasPrefix(a, b) || strings.HasPrefix(b, a) {
				return true
			}
		}
	}
	return false
}

func waypointScenes(in SnippetSceneInput) ([]Scene, error) {
	w := in.Plan.Waypoint
	if w == nil {
		return nil, fmt.Errorf("the plan has no waypoint")
	}

	promise, spine := false, false
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Waypoint == nil {
			return nil, fmt.Errorf("beat %q has no waypoint direction", beat.ID)
		}
		show := beat.Waypoint.ResolvedShow()
		switch show {
		case "promise":
			promise = true
		case "spine":
			spine = true
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"promise": promise,
			"spine":   spine,
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneWaypoint,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"ordinal": w.Ordinal,
			"line":    w.Line,
			"promise": w.Promise,
			"stops":   w.Stops,
			"at":      w.At,
			"steps":   steps,
		},
	}}, nil
}
