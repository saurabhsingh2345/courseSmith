package pipeline

import (
	"strings"
	"testing"
)

const lkNarration = "Nobody in this chain holds the whole answer, so each one hands the question a little further along."

func lookupPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "lookup",
		Title:    "Nobody knows the whole answer",
		Lookup: &LookupSpec{
			Key: "www.example.com",
			Hops: []LookupHop{
				{Where: "your resolver", Gives: "nothing cached, so it asks upward", Miss: "it starts again from the root"},
				{Where: "a root server", Gives: "go ask the dot com servers", Miss: "there is nothing above the root"},
				{Where: "the com servers", Gives: "the address for example dot com", Miss: "the name does not exist"},
			},
			Answer: "93.184.216.34",
		},
		Beats: []SnippetBeat{
			{ID: "ask", Heading: "The question", Narration: lkNarration, Lookup: &LookupBeat{Show: "ask"}},
			{ID: "resolver", Heading: "Your resolver", Narration: lkNarration, Lookup: &LookupBeat{Show: "hop", At: 0}},
			{ID: "root", Heading: "The root", Narration: lkNarration, Lookup: &LookupBeat{Show: "hop", At: 1}},
			{ID: "tld", Heading: "The com servers", Narration: lkNarration, Lookup: &LookupBeat{Show: "hop", At: 2}},
			{ID: "hit", Heading: "The answer", Narration: lkNarration, Lookup: &LookupBeat{Show: "hit"}},
			{ID: "cache", Heading: "Next time", Narration: lkNarration, Lookup: &LookupBeat{Show: "cache"}},
		},
	}
	// A beat here is a shot, so the fixture budget is sized at the template's
	// own 28-word ideal — nBeats * 40 would make beatBounds demand more beats
	// than the fixture has.
	p.targetWords = 6 * 28
	return p
}

func TestLookupPlanAccepted(t *testing.T) {
	if err := validateLookupPlan(lookupPlan()); err != nil {
		t.Fatalf("a well-formed lookup plan was rejected: %v", err)
	}
}

// The family's signature rule: the walk is computed in Go, so a card that
// teleports past a station is rejected with both positions quoted.
func TestLookupRejectsASkippedHop(t *testing.T) {
	p := lookupPlan()
	p.Beats[2].Lookup = &LookupBeat{Show: "hop", At: 2}
	err := validateLookupPlan(p)
	if err == nil {
		t.Fatal("a chain that skips a station was accepted")
	}
	if !strings.Contains(err.Error(), "hop 2") || !strings.Contains(err.Error(), "hop 1") {
		t.Fatalf("the error does not quote both the visited and the expected hop: %v", err)
	}
}

func TestLookupRejectsAHitBeforeTheLastHop(t *testing.T) {
	p := lookupPlan()
	p.Beats[3], p.Beats[4] = p.Beats[4], p.Beats[3]
	err := validateLookupPlan(p)
	if err == nil {
		t.Fatal("an answer returning mid-chain was accepted")
	}
	if !strings.Contains(err.Error(), "decoration") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLookupRejectsTwoHits(t *testing.T) {
	p := lookupPlan()
	p.Beats[5].Lookup = &LookupBeat{Show: "hit"}
	err := validateLookupPlan(p)
	if err == nil {
		t.Fatal("two answers coming back were accepted")
	}
	if !strings.Contains(err.Error(), "a second time") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLookupRejectsAClipWithNoHit(t *testing.T) {
	p := lookupPlan()
	p.Beats = p.Beats[:5]
	p.Beats[4].Lookup = &LookupBeat{Show: "cache"}
	p.targetWords = 5 * 28
	err := validateLookupPlan(p)
	if err == nil {
		t.Fatal("a clip whose answer never comes back was accepted")
	}
	if !strings.Contains(err.Error(), "no beat returns the answer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLookupRejectsAnUnvisitedHop(t *testing.T) {
	p := lookupPlan()
	// Drop the third station's beat, so the chain stops one short.
	p.Beats = append(p.Beats[:3], p.Beats[4:]...)
	p.targetWords = 5 * 28
	err := validateLookupPlan(p)
	if err == nil {
		t.Fatal("a chain that stops before its last station was accepted")
	}
	if !strings.Contains(err.Error(), "never visited") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLookupRequiresOpeningOnTheAsk(t *testing.T) {
	p := lookupPlan()
	p.Beats[0].Lookup = &LookupBeat{Show: "hop", At: 0}
	err := validateLookupPlan(p)
	if err == nil {
		t.Fatal("a clip that starts travelling before showing the key was accepted")
	}
	if !strings.Contains(err.Error(), "open on the key alone") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLookupRejectsACacheBeforeTheEnd(t *testing.T) {
	p := lookupPlan()
	p.Beats[1].Lookup = &LookupBeat{Show: "cache"}
	err := validateLookupPlan(p)
	if err == nil {
		t.Fatal("a shortcut drawn before the road was accepted")
	}
	if !strings.Contains(err.Error(), "closer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLookupRejectsAHopThatAnswersNothing(t *testing.T) {
	p := lookupPlan()
	p.Lookup.Hops[1].Gives = ""
	err := validateLookupPlan(p)
	if err == nil {
		t.Fatal("a station that leaves the card unchanged was accepted")
	}
	if !strings.Contains(err.Error(), "answers nothing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLookupRejectsTooFewHops(t *testing.T) {
	p := lookupPlan()
	p.Lookup.Hops = p.Lookup.Hops[:1]
	if err := validateLookupPlan(p); err == nil {
		t.Fatal("a one-station chain was accepted")
	}
}

func TestLookupRejectsAMissingAnswer(t *testing.T) {
	p := lookupPlan()
	p.Lookup.Answer = ""
	err := validateLookupPlan(p)
	if err == nil {
		t.Fatal("a lookup that resolves to nothing was accepted")
	}
	if !strings.Contains(err.Error(), "no answer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLookupNormalizeClampsLongFields(t *testing.T) {
	p := lookupPlan()
	p.Lookup.Hops[0].Gives = "one two three four five six seven eight nine ten"
	p.Lookup.Answer = "one two three four five six seven eight"
	normalizeLookupPlan(p)
	if got := len(strings.Fields(p.Lookup.Hops[0].Gives)); got != maxLookupGivesWords {
		t.Fatalf("the answer line kept %d words, want %d", got, maxLookupGivesWords)
	}
	if got := len(strings.Fields(p.Lookup.Answer)); got != maxLookupAnswerWords {
		t.Fatalf("the answer kept %d words, want %d", got, maxLookupAnswerWords)
	}
}

func TestLookupNormalizeClampsAnOutOfRangeHop(t *testing.T) {
	p := lookupPlan()
	p.Beats[1].Lookup.At = 99
	normalizeLookupPlan(p)
	if got := p.Beats[1].Lookup.At; got != 2 {
		t.Fatalf("an out-of-range hop clamped to %d, want 2", got)
	}
}

func TestLookupShowDefaultsToHop(t *testing.T) {
	b := LookupBeat{Show: "sparkle"}
	if got := b.ResolvedShow(); got != "hop" {
		t.Fatalf("an unknown show resolved to %q, want hop", got)
	}
}

// Each step carries the stamps collected so far, so the renderer draws a whole
// frame from one step rather than replaying the beat list.
func TestLookupScenesAccumulateStamps(t *testing.T) {
	p := lookupPlan()
	scenes, err := lookupScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	hops, _ := props["hops"].([]map[string]any)
	if len(hops) != 3 {
		t.Fatalf("want 3 stations, got %d", len(hops))
	}
	if hops[1]["where"] != "a root server" {
		t.Fatalf("the second station is wrong: %v", hops[1])
	}
	if props["key"] != "www.example.com" || props["answer"] != "93.184.216.34" {
		t.Fatalf("the key or the answer did not reach the props: %v / %v", props["key"], props["answer"])
	}

	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != 6 {
		t.Fatalf("want 6 steps, got %d", len(steps))
	}
	first := steps[0]
	if visited, _ := first["visited"].([]int); len(visited) != 0 {
		t.Fatalf("the opening beat already has stamps: %v", visited)
	}
	if first["answered"] != false || first["cached"] != false {
		t.Fatalf("the opening beat is already resolved: %v", first)
	}

	last := steps[len(steps)-1]
	visited, _ := last["visited"].([]int)
	if len(visited) != 3 || visited[0] != 0 || visited[2] != 2 {
		t.Fatalf("the closing beat's stamps are %v, want [0 1 2]", visited)
	}
	if last["answered"] != true || last["cached"] != true {
		t.Fatalf("the closing beat has not recorded the hit and the cache: %v", last)
	}
	if last["show"] != "cache" {
		t.Fatalf("the last step shows %v, want cache", last["show"])
	}
}
