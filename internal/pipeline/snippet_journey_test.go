package pipeline

import (
	"strings"
	"testing"
)

const jyNarration = "The request leaves your machine and every stop along the way changes it a little."

func journeyPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "journey",
		Title:    "What happens when you press enter",
		Journey: &JourneySpec{
			Stops: []JourneyStop{
				{Label: "your laptop", Kind: "device", Adds: "the address you typed into the bar"},
				{Label: "the hall router", Kind: "router", Adds: "a public address the reply can come back to"},
				{Label: "the resolver", Kind: "dns", Adds: "the numeric address hiding behind the name"},
				{Label: "the web server", Kind: "server", Adds: "the page itself, finally"},
			},
			Return: "the HTML for the page you asked for",
		},
		Beats: []SnippetBeat{
			{ID: "the-route", Heading: "The route", Narration: jyNarration, Journey: &JourneyBeat{Show: "map"}},
			{ID: "first-hop", Heading: "Out of the house", Narration: jyNarration, Journey: &JourneyBeat{Show: "hop", At: 1}},
			{ID: "the-name", Heading: "Finding the number", Narration: jyNarration, Journey: &JourneyBeat{Show: "hop", At: 2}},
			{ID: "arrival", Heading: "Arrival", Narration: jyNarration, Journey: &JourneyBeat{Show: "reach", At: 3}},
			{ID: "the-answer", Heading: "The answer", Narration: jyNarration, Journey: &JourneyBeat{Show: "return"}},
		},
	}
	// Sized against this template's own ideal of 28 words per beat rather than
	// the shared 40: at 40 a five-beat fixture is below the minimum beat count
	// the shared bounds derive from its own budget, and the fixture would be
	// rejected for its length before any rule under test ran.
	p.targetWords = 5 * 28
	return p
}

func TestJourneyPlanAccepted(t *testing.T) {
	if err := validateJourneyPlan(journeyPlan()); err != nil {
		t.Fatalf("a well-formed journey plan was rejected: %v", err)
	}
}

// The family's signature rule for a route: the walk is checked leg by leg, and
// a skip is rejected with both stop numbers quoted.
func TestJourneyRejectsAHopThatSkipsAStop(t *testing.T) {
	p := journeyPlan()
	p.Beats[1].Journey = &JourneyBeat{Show: "hop", At: 2}
	err := validateJourneyPlan(p)
	if err == nil {
		t.Fatal("a packet that appeared two stops along was accepted")
	}
	if !strings.Contains(err.Error(), "stop 2") || !strings.Contains(err.Error(), "the resolver") {
		t.Fatalf("the error does not quote the stop it jumped to: %v", err)
	}
	if !strings.Contains(err.Error(), "the hall router") {
		t.Fatalf("the error does not name the stop that was skipped: %v", err)
	}
}

func TestJourneyRejectsHoppingBackwards(t *testing.T) {
	p := journeyPlan()
	p.Beats[2].Journey = &JourneyBeat{Show: "hop", At: 1}
	err := validateJourneyPlan(p)
	if err == nil {
		t.Fatal("a packet that went back a stop was accepted")
	}
	if !strings.Contains(err.Error(), "one way") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJourneyRejectsHoppingToTheDestination(t *testing.T) {
	p := journeyPlan()
	p.Beats[3].Journey = &JourneyBeat{Show: "hop", At: 3}
	p.Beats[4].Journey = &JourneyBeat{Show: "reach", At: 3}
	err := validateJourneyPlan(p)
	if err == nil {
		t.Fatal("a hop onto the destination was accepted, and arriving is its own shot")
	}
	if !strings.Contains(err.Error(), "reach") {
		t.Fatalf("the error does not point at the reach beat: %v", err)
	}
}

func TestJourneyRejectsReachingSomewhereOtherThanTheLastStop(t *testing.T) {
	p := journeyPlan()
	p.Beats[2].Journey = &JourneyBeat{Show: "reach", At: 2}
	p.Beats[3].Journey = &JourneyBeat{Show: "hop", At: 3}
	err := validateJourneyPlan(p)
	if err == nil {
		t.Fatal("an arrival short of the end of the map was accepted")
	}
	if !strings.Contains(err.Error(), "the web server") {
		t.Fatalf("the error does not name the real destination: %v", err)
	}
}

func TestJourneyRejectsTwoArrivals(t *testing.T) {
	p := journeyPlan()
	p.Beats[4].Journey = &JourneyBeat{Show: "reach", At: 3}
	if err := validateJourneyPlan(p); err == nil {
		t.Fatal("a route with two arrivals was accepted")
	}
}

func TestJourneyRejectsAResponseBeforeTheRequestArrives(t *testing.T) {
	p := journeyPlan()
	p.Beats[3].Journey = &JourneyBeat{Show: "return"}
	p.Beats[4].Journey = &JourneyBeat{Show: "reach", At: 3}
	err := validateJourneyPlan(p)
	if err == nil {
		t.Fatal("a response that swept home before the packet arrived was accepted")
	}
	if !strings.Contains(err.Error(), "latency") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestJourneyRequiresOpeningOnTheMap(t *testing.T) {
	p := journeyPlan()
	p.Beats[0].Journey = &JourneyBeat{Show: "hop", At: 1}
	p.Beats[1].Journey = &JourneyBeat{Show: "map"}
	if err := validateJourneyPlan(p); err == nil {
		t.Fatal("a clip that moved the packet before drawing the route was accepted")
	}
}

func TestJourneyRejectsARouteWithNoMiddle(t *testing.T) {
	p := journeyPlan()
	p.Journey.Stops = p.Journey.Stops[:2]
	p.Beats[1].Journey = &JourneyBeat{Show: "reach", At: 1}
	p.Beats[2].Journey = &JourneyBeat{Show: "return"}
	p.Beats = p.Beats[:3]
	p.targetWords = 3 * 28
	err := validateJourneyPlan(p)
	if err == nil {
		t.Fatal("a two-stop route was accepted, and two stops is the arrow the viewer already believes in")
	}
	if !strings.Contains(err.Error(), "2 stops") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestJourneyRequiresEveryStopToSayWhatItAdded(t *testing.T) {
	p := journeyPlan()
	p.Journey.Stops[2].Adds = ""
	err := validateJourneyPlan(p)
	if err == nil {
		t.Fatal("a stop that does nothing to the request was accepted")
	}
	if !strings.Contains(err.Error(), "the resolver") {
		t.Fatalf("the error does not name the silent stop: %v", err)
	}
}

func TestJourneyRequiresAResponseWhenTheClipSweepsHome(t *testing.T) {
	p := journeyPlan()
	p.Journey.Return = ""
	if err := validateJourneyPlan(p); err == nil {
		t.Fatal("a return sweep with nothing coming back was accepted")
	}
}

func TestJourneyNormalizeRepairsLabelsKindsAndIndices(t *testing.T) {
	p := journeyPlan()
	p.Journey.Stops[0].Label = "your   own  personal laptop computer"
	p.Journey.Stops[1].Kind = "satellite uplink"
	p.Journey.Return = "the HTML for the page you asked for, all of it, every byte"
	p.Beats[3].Journey.At = 99
	normalizeJourneyPlan(p)
	if got := p.Journey.Stops[0].Label; got != "your own personal" {
		t.Fatalf("label normalized to %q, want three collapsed words", got)
	}
	if got := p.Journey.Stops[1].Kind; got != "device" {
		t.Fatalf("an unknown kind normalized to %q, want device", got)
	}
	if got := len(strings.Fields(p.Journey.Return)); got != maxJourneyReturnWords {
		t.Fatalf("the response is %d words, want it clamped to %d", got, maxJourneyReturnWords)
	}
	if got := p.Beats[3].Journey.At; got != 3 {
		t.Fatalf("an index off the end normalized to %d, want the last stop 3", got)
	}
	if err := validateJourneyPlan(p); err != nil {
		t.Fatalf("a wordy-but-sound plan was rejected after normalize: %v", err)
	}
}

func TestJourneyNormalizeCapsTheRoute(t *testing.T) {
	p := journeyPlan()
	for i := 0; i < 5; i++ {
		p.Journey.Stops = append(p.Journey.Stops, JourneyStop{Label: "another hop", Kind: "router", Adds: "nothing much at all"})
	}
	normalizeJourneyPlan(p)
	if got := len(p.Journey.Stops); got != maxJourneyStops {
		t.Fatalf("the route normalized to %d stops, want it capped at %d", got, maxJourneyStops)
	}
}

func TestJourneyShowDefaultsToHop(t *testing.T) {
	b := JourneyBeat{Show: "teleport"}
	if got := b.ResolvedShow(); got != "hop" {
		t.Fatalf("an unknown show resolved to %q, want hop", got)
	}
}

func TestJourneyKindDefaultsToDevice(t *testing.T) {
	s := JourneyStop{Kind: "carrier pigeon"}
	if got := s.ResolvedKind(); got != "device" {
		t.Fatalf("an unknown kind resolved to %q, want device", got)
	}
}

// The component is handed the state of the map, never asked to work out where
// the packet has been.
func TestJourneyScenesAccumulateTheRoute(t *testing.T) {
	p := journeyPlan()
	scenes, err := journeyScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("want one scene spanning the clip, got %d", len(scenes))
	}
	props := scenes[0].Props

	stops, _ := props["stops"].([]map[string]any)
	if len(stops) != 4 {
		t.Fatalf("want 4 stops, got %d", len(stops))
	}
	if stops[2]["kind"] != "dns" || stops[2]["label"] != "the resolver" {
		t.Fatalf("the resolver stop is wrong: %v", stops[2])
	}

	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != 5 {
		t.Fatalf("want 5 steps, got %d", len(steps))
	}
	if steps[0]["show"] != "map" {
		t.Fatalf("first step shows %v, want map", steps[0]["show"])
	}
	if first, _ := steps[0]["legs"].([]int); len(first) != 0 {
		t.Fatalf("the opener has travelled %v, want nothing", first)
	}
	if steps[0]["reached"] != 0 {
		t.Fatalf("the opener has the packet at %v, want the origin", steps[0]["reached"])
	}
	last := steps[len(steps)-1]
	if last["show"] != "return" {
		t.Fatalf("last step shows %v, want return", last["show"])
	}
	legs, _ := last["legs"].([]int)
	if len(legs) != 3 || legs[0] != 1 || legs[2] != 3 {
		t.Fatalf("the closer has travelled %v, want every leg 1 through 3", legs)
	}
	if last["reached"] != 3 {
		t.Fatalf("the closer has the packet at %v, want the destination", last["reached"])
	}
}
