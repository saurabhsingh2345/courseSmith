package pipeline

// The journey template: the packet's trip.
//
// "You type google.com and press enter" is the question every foundations
// course opens with and almost none of them answers, because the honest answer
// is a route — your laptop, the router in the hall, a DNS resolver, somebody's
// server — and a route is a thing you draw, not a thing you list. A bulleted
// list of the seven steps is what most courses ship, and it teaches the words
// without teaching the shape: the viewer learns that DNS comes before TCP and
// still cannot say where DNS *is*.
//
// So this template is a map with one moving object on it. The stops are drawn
// first and dim, the whole route visible before anything travels, because the
// point of the clip is the distance — the viewer should see how far the packet
// has to go before it goes anywhere. Then it hops, one leg at a time, and each
// stop says what it ADDED as the packet passed through. That is the real
// lesson: nothing on the route is a relay, every stop does something to the
// request, and the request that arrives is not the one that left.
//
// The validators are about the route being a route. Hops must visit the stops
// in order with no skips, because a packet that appears two stops along has
// silently taught the viewer that the missing stop is optional. There is
// exactly one arrival and it is at the last stop, because a map with two
// destinations is not a round trip. And the return sweeps the WHOLE way back,
// after the arrival, because the round trip is the only part of this picture
// that explains latency: the answer has to walk home too.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "journey",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The packet's trip",
		Description: "A route drawn end to end with one packet travelling it, each stop saying what it added, and the response sweeping all the way back. Reach for it when the subject is what happens between an action and its answer — a URL, a request, a round trip.",
		Example:     "You type google.com and press enter — the full round trip",
		PromptFile:  snippetJourneyTemplateName,
		NeedsCode:   false,
		// The dim route, three or four hops, the arrival and the sweep home:
		// six states minimum, and under thirty-five seconds none of them holds
		// long enough for the viewer to read what the stop added.
		MinTargetSec: 35,
		// Longer than the family default because the beat count here is a
		// property of the ROUTE, not of how long attention lasts: five stops is
		// five beats before the closer, and a budget that cannot fund them
		// makes the validator reject every plan the prompt asks for.
		DefaultTargetSec: 55,
		// Six stops is the ceiling, so map + five hops + return is eight. Past
		// that the same packet is being moved across the same curve again.
		MaxBeats: 8,
		// A beat here is a SHOT — one leg of the trip — not a step in an
		// argument. Twenty-eight words is about nine seconds, which is how long
		// one hop and its caption stay interesting.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Journey: true},
		OwnsPlan:          planFields{Journey: true},
		Normalize:         normalizeJourneyPlan,
		Validate:          validateJourneyPlan,
		Scenes:            journeyScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":          strings.Join(MetricRoles(), ", "),
				"Shows":          strings.Join(JourneyShows(), ", "),
				"Kinds":          strings.Join(JourneyKinds(), ", "),
				"MinStops":       minJourneyStops,
				"MaxStops":       maxJourneyStops,
				"MaxLabelWords":  maxJourneyLabelWords,
				"MaxAddsWords":   maxJourneyAddsWords,
				"MaxReturnWords": maxJourneyReturnWords,
			}
		},
	})
}

const snippetJourneyTemplateName = "snippet_journey.tmpl"

const (
	// Two stops is an arrow, not a journey — and an arrow is what the viewer
	// already believes the internet is. Three is the smallest route with a
	// middle, which is the whole point.
	minJourneyStops = 3
	// The curve carries six nodes with their labels at a legible size across
	// the drawing box; a seventh puts the captions on top of each other.
	maxJourneyStops = 6

	// A stop is a place, and a place has a name — "your laptop", "the DNS
	// resolver". Three words is a name; four is a sentence.
	maxJourneyLabelWords = 3
	// What the stop added, as a caption under a node rather than a paragraph
	// beside it.
	maxJourneyAddsWords = 10
	// The response, in the same breath as the sweep home.
	maxJourneyReturnWords = 10
)

// journeyKinds is the closed vocabulary of what a stop IS. It picks the glyph
// drawn on the node, so an invented kind would be a node with no picture.
var journeyKinds = map[string]bool{
	// Something a person is holding or sitting at. The origin, usually.
	"device": true,
	// A box that forwards: the hall router, the ISP's edge, a gateway.
	"router": true,
	// A resolver — the stop that turns a name into an address.
	"dns": true,
	// A machine that answers rather than forwards. The destination, usually.
	"server": true,
	// Somebody else's infrastructure taken as one thing: a CDN, a region.
	"cloud": true,
}

// JourneyKinds returns the stop vocabulary sorted.
func JourneyKinds() []string {
	out := make([]string, 0, len(journeyKinds))
	for k := range journeyKinds {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// journeyShows is the closed vocabulary of what a beat does.
var journeyShows = map[string]bool{
	// The whole route, dim, before anything moves. The opener.
	"map": true,
	// The packet travels the leg into stop At, and that stop's Adds lands.
	"hop": true,
	// The packet arrives at the final stop, which lights fully.
	"reach": true,
	// The response sweeps back the whole way with the Return caption. The
	// closer.
	"return": true,
}

// JourneyShows returns the beat vocabulary sorted.
func JourneyShows() []string {
	out := make([]string, 0, len(journeyShows))
	for k := range journeyShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// JourneyStop is one place on the route.
type JourneyStop struct {
	// Label names the place — "your laptop", "the DNS resolver".
	Label string `json:"label"`
	// Kind picks the glyph, from journeyKinds.
	Kind string `json:"kind,omitempty"`
	// Adds says what this stop did to the request as it passed through.
	Adds string `json:"adds,omitempty"`
}

// ResolvedKind returns the stop's glyph, defaulting the unknown to a device —
// the one kind that is never wrong as a placeholder, because every route has
// one at the near end and it reads as "a thing", not as a claim.
func (s JourneyStop) ResolvedKind() string {
	k := strings.ToLower(strings.TrimSpace(s.Kind))
	if journeyKinds[k] {
		return k
	}
	return "device"
}

// JourneySpec is the route and what comes back down it. On the plan because
// the map is up for the whole clip; only the packet moves.
type JourneySpec struct {
	// Stops are the places, in travel order, origin first.
	Stops []JourneyStop `json:"stops"`
	// Return is what comes back — "the HTML for the page".
	Return string `json:"return,omitempty"`
}

// JourneyBeat is one shot: which state of the map this beat shows.
type JourneyBeat struct {
	// Show is a journeyShows name.
	Show string `json:"show"`
	// At is the stop this beat travels to, for "hop" and "reach".
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a hop — the
// workhorse state most beats of this template are in.
func (b JourneyBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if journeyShows[s] {
		return s
	}
	return "hop"
}

// journeyStopName quotes a stop for an error message. It takes an index that
// may be off the end because a validator can be handed a plan that never went
// through normalize — a test, a hand-written snippet.yaml — and a rejection
// that panics is worse than the plan it was rejecting.
func journeyStopName(j *JourneySpec, i int) string {
	if i < 0 || i >= len(j.Stops) {
		return "no such stop"
	}
	return fmt.Sprintf("%q", j.Stops[i].Label)
}

func normalizeJourneyPlan(p *SnippetPlan) {
	j := p.Journey
	if j == nil {
		return
	}
	if len(j.Stops) > maxJourneyStops {
		j.Stops = j.Stops[:maxJourneyStops]
	}
	for i := range j.Stops {
		j.Stops[i].Label = clampWords(collapseSpaces(j.Stops[i].Label), maxJourneyLabelWords)
		j.Stops[i].Kind = j.Stops[i].ResolvedKind()
		j.Stops[i].Adds = clampWords(collapseSpaces(j.Stops[i].Adds), maxJourneyAddsWords)
	}
	j.Return = clampWords(collapseSpaces(j.Return), maxJourneyReturnWords)

	last := len(j.Stops) - 1
	for i := range p.Beats {
		b := p.Beats[i].Journey
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		// Clamped rather than dropped: an index off the end is a counting slip,
		// and the ordering rules below will still catch a route that skips.
		if b.At < 0 {
			b.At = 0
		}
		if last >= 0 && b.At > last {
			b.At = last
		}
	}
}

func validateJourneyPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Journey: true}); err != nil {
		return err
	}

	j := p.Journey
	if j == nil {
		return fmt.Errorf("the plan has no route — this template is a map with one packet crossing it, so without stops there is nothing to cross")
	}
	if n := len(j.Stops); n < minJourneyStops || n > maxJourneyStops {
		return fmt.Errorf("the route has %d stops, want %d-%d. Two stops is an arrow, which is the picture the viewer already has in their head, and past %d the labels on the curve land on top of each other",
			n, minJourneyStops, maxJourneyStops, maxJourneyStops)
	}
	for i, s := range j.Stops {
		if strings.TrimSpace(s.Label) == "" {
			return fmt.Errorf("stop %d has no label. Every node on the map is a place with a name — \"your laptop\", \"the DNS resolver\" — and an unnamed circle on a route teaches nothing", i)
		}
		if strings.TrimSpace(s.Adds) == "" {
			return fmt.Errorf("stop %d (%q) never says what it added. The lesson of this picture is that nothing on the route is a relay: give each stop the one thing it does to the request as it passes through",
				i, s.Label)
		}
	}
	last := len(j.Stops) - 1

	if p.Beats[0].Journey == nil || p.Beats[0].Journey.ResolvedShow() != "map" {
		return fmt.Errorf("beat %q does not open on the route. The whole point is the DISTANCE — the viewer has to see how far the packet must go before it goes anywhere — so open with {\"show\": \"map\"}",
			p.Beats[0].ID)
	}

	// The traversal, checked as one contiguous walk. want is the next stop the
	// packet is allowed to be at.
	want := 1
	reachAt, returnAt := -1, -1
	for i, b := range p.Beats {
		jb := b.Journey
		if jb == nil {
			return fmt.Errorf("beat %q has no journey direction — every beat shows one state of the map", b.ID)
		}
		switch jb.ResolvedShow() {
		case "map":
			if i != 0 {
				return fmt.Errorf("beat %q goes back to the dim route part-way through. Once the packet is moving the map stays up behind it — a second {\"show\": \"map\"} reads as the clip starting over", b.ID)
			}
		case "hop":
			if jb.At > want {
				return fmt.Errorf("beat %q hops to stop %d (%s) but the packet is standing at stop %d, so stops %d through %d are drawn and never travelled to. Hops move one leg at a time — a packet that appears two stops along has quietly taught the viewer that %s is optional",
					b.ID, jb.At, journeyStopName(j, jb.At), want-1, want, jb.At-1, journeyStopName(j, want))
			}
			if jb.At < want {
				return fmt.Errorf("beat %q hops back to stop %d (%s), which the packet passed through already. The route runs one way — if that stop needs a second beat, give it another \"hop\" further along or say it in the narration",
					b.ID, jb.At, journeyStopName(j, jb.At))
			}
			if jb.At == last {
				return fmt.Errorf("beat %q hops to stop %d (%s), which is the destination. Arriving is its own shot — the node lights fully and the request is finally answered — so use {\"show\": \"reach\", \"at\": %d}",
					b.ID, jb.At, journeyStopName(j, jb.At), last)
			}
			want++
		case "reach":
			if reachAt >= 0 {
				return fmt.Errorf("beat %q arrives a second time, after beat %q already did. There is one destination on this map; a route with two arrivals is two clips",
					b.ID, p.Beats[reachAt].ID)
			}
			if jb.At > want {
				return fmt.Errorf("beat %q arrives at stop %d (%s) but the packet is standing at stop %d, so stops %d through %d are drawn and never travelled to. Hop through them first",
					b.ID, jb.At, journeyStopName(j, jb.At), want-1, want, jb.At-1)
			}
			if jb.At != last {
				return fmt.Errorf("beat %q arrives at stop %d (%s), but the last stop on the route is %d (%s). \"reach\" is the destination lighting up — anything before the end of the map is a hop",
					b.ID, jb.At, journeyStopName(j, jb.At), last, journeyStopName(j, last))
			}
			reachAt = i
			want++
		case "return":
			if returnAt >= 0 {
				return fmt.Errorf("beat %q sweeps the response home a second time, after beat %q. The answer only walks back once", b.ID, p.Beats[returnAt].ID)
			}
			if reachAt < 0 {
				return fmt.Errorf("beat %q sends the response home before the request has arrived anywhere. The round trip is the only part of this picture that explains latency, and it only reads if the packet is seen to get there first — put {\"show\": \"reach\"} ahead of it",
					b.ID)
			}
			if strings.TrimSpace(j.Return) == "" {
				return fmt.Errorf("beat %q sweeps a response home but the route says nothing about what comes back. Give journey.return the answer in at most %d words — \"the HTML for the page\"",
					b.ID, maxJourneyReturnWords)
			}
			returnAt = i
		}
	}
	if reachAt < 0 {
		return fmt.Errorf("no beat reaches the destination. The map ends at %q and the packet never gets there, so the clip is a trip with no arrival — close it with {\"show\": \"reach\", \"at\": %d}, and better still with a {\"show\": \"return\"} after that",
			j.Stops[last].Label, last)
	}
	return nil
}

// journeyScenes lays the clip out as ONE scene. The route geometry, which stop
// is lit at each moment, and which legs have been travelled are all resolved
// here, so the component draws a state rather than working one out.
func journeyScenes(in SnippetSceneInput) ([]Scene, error) {
	j := in.Plan.Journey
	if j == nil {
		return nil, fmt.Errorf("the plan has no route")
	}

	stops := make([]map[string]any, 0, len(j.Stops))
	for _, s := range j.Stops {
		stops = append(stops, map[string]any{
			"label": s.Label,
			"kind":  s.ResolvedKind(),
			"adds":  s.Adds,
		})
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	legs := make([]int, 0, len(j.Stops))
	at := 0
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Journey == nil {
			return nil, fmt.Errorf("beat %q has no journey direction", beat.ID)
		}
		show := beat.Journey.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "hop" || show == "reach" {
			at = beat.Journey.At
			step["at"] = at
			// Leg n is the stretch from stop n-1 to stop n, so the leg the
			// packet just travelled shares the index of the stop it landed on.
			legs = append(legs, at)
		}
		// The furthest the packet has ever been, so the map accumulates: a lit
		// stop stays lit for the rest of the clip.
		step["reached"] = at
		travelled := append([]int(nil), legs...)
		sort.Ints(travelled)
		step["legs"] = travelled
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneJourney,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"stops":  stops,
			"return": j.Return,
			"steps":  steps,
		}),
	}}, nil
}
