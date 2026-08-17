package pipeline

import (
	"strings"
	"testing"
)

const spotlightNarration = "The card holds still on the left while each claim lands beside it, one at a time, in the order the voice reaches them."

func spotlightPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "spotlight",
		Title:    "What Claude Code is actually for",
		Spotlight: &SpotlightSpec{
			Title: "Claude Code",
			Note:  "An agent that works in your terminal",
			Role:  "quantity",
			Brand: "claude",
			Site:  "anthropic.com",
			Icon:  "terminal",
			Points: []SpotlightPoint{
				{Text: "Refactors across files you never opened", Icon: "code"},
				{Text: "Reads the whole repo before it edits", Icon: "search"},
				{Text: "Runs the tests and fixes what broke", Icon: "check"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "the-tool", Heading: "The agent", Narration: spotlightNarration, Spotlight: &SpotlightBeat{Show: "card"}},
			{ID: "refactor", Heading: "Across files", Narration: spotlightNarration, Spotlight: &SpotlightBeat{Show: "point", At: 0}},
			{ID: "reads", Heading: "Reads first", Narration: spotlightNarration, Spotlight: &SpotlightBeat{Show: "point", At: 1}},
			{ID: "tests", Heading: "Runs the tests", Narration: spotlightNarration, Spotlight: &SpotlightBeat{Show: "point", At: 2}},
			{ID: "all-three", Heading: "What it is for", Narration: spotlightNarration, Spotlight: &SpotlightBeat{Show: "all"}},
		},
	}
	p.targetWords = 5 * 22
	return p
}

func TestSpotlightPlanAccepted(t *testing.T) {
	if err := validateSpotlightPlan(spotlightPlan()); err != nil {
		t.Fatalf("a well-formed spotlight was rejected: %v", err)
	}
}

// The anti-sticker rule, inherited from cards: a logo and a name teaches nobody
// anything, and here the claims beside it need something to attach to.
func TestSpotlightRequiresALineUnderTheName(t *testing.T) {
	p := spotlightPlan()
	p.Spotlight.Note = ""
	err := validateSpotlightPlan(p)
	if err == nil {
		t.Fatal("a card with no line under the name was accepted")
	}
	if !strings.Contains(err.Error(), "Claude Code") {
		t.Fatalf("the error does not name the subject: %v", err)
	}
}

// A claim with no beat is not a dim row, it is a row that never gets drawn — and
// the message has to say that, because the cards template's equivalent failure
// leaves a visible card and this one leaves nothing.
func TestSpotlightRejectsAClaimWithNoBeat(t *testing.T) {
	p := spotlightPlan()
	p.Beats = p.Beats[:4] // drops the "all" beat's predecessor pairing
	p.Beats[3] = SnippetBeat{ID: "all-three", Heading: "What it is for", Narration: spotlightNarration, Spotlight: &SpotlightBeat{Show: "all"}}
	p.targetWords = 4 * 22
	err := validateSpotlightPlan(p)
	if err == nil {
		t.Fatal("a claim with no beat was accepted")
	}
	if !strings.Contains(err.Error(), "Runs the tests and fixes what broke") {
		t.Fatalf("the error does not name the claim that never appears: %v", err)
	}
}

// The rows stack downward, so landing them out of order makes the stack jump.
func TestSpotlightRejectsClaimsOutOfOrder(t *testing.T) {
	p := spotlightPlan()
	p.Beats[1].Spotlight = &SpotlightBeat{Show: "point", At: 1}
	p.Beats[2].Spotlight = &SpotlightBeat{Show: "point", At: 0}
	err := validateSpotlightPlan(p)
	if err == nil {
		t.Fatal("claims landing out of order were accepted")
	}
	if !strings.Contains(err.Error(), "next") {
		t.Fatalf("the error does not say which claim was due: %v", err)
	}
}

func TestSpotlightRejectsASingleClaim(t *testing.T) {
	p := spotlightPlan()
	p.Spotlight.Points = p.Spotlight.Points[:1]
	p.Beats = []SnippetBeat{
		{ID: "the-tool", Heading: "The agent", Narration: spotlightNarration, Spotlight: &SpotlightBeat{Show: "card"}},
		{ID: "refactor", Heading: "Across files", Narration: spotlightNarration, Spotlight: &SpotlightBeat{Show: "point", At: 0}},
		{ID: "all-of-it", Heading: "What it is for", Narration: spotlightNarration, Spotlight: &SpotlightBeat{Show: "all"}},
	}
	p.targetWords = 3 * 22
	if err := validateSpotlightPlan(p); err == nil {
		t.Fatal("one claim was accepted — that is a tagline with extra steps")
	}
}

// The nine-word ceiling is what keeps a claim from being a spec, so normalize has
// to actually enforce it rather than leaving it to the prompt.
func TestSpotlightClampsALongClaim(t *testing.T) {
	p := spotlightPlan()
	p.Spotlight.Points[0].Text = "Refactors across every file in the repository including the ones you have never opened yourself"
	normalizeSpotlightPlan(p)
	if got := len(strings.Fields(p.Spotlight.Points[0].Text)); got > maxSpotlightPointWords {
		t.Errorf("claim kept %d words, want at most %d", got, maxSpotlightPointWords)
	}
}

// The scene graph carries how many rows are on screen, so the component paints a
// frame from one step instead of replaying the clip to count them.
func TestSpotlightScenesCountTheRows(t *testing.T) {
	p := spotlightPlan()
	normalizeSpotlightPlan(p)
	if err := validateSpotlightPlan(p); err != nil {
		t.Fatalf("fixture rejected: %v", err)
	}
	spans := make([]SectionSpan, len(p.Beats))
	ends := make([]int, len(p.Beats))
	for i, b := range p.Beats {
		spans[i] = SectionSpan{ID: b.ID, StartMs: i * 5000, EndMs: (i + 1) * 5000}
		ends[i] = (i + 1) * 5000
	}
	scenes, err := spotlightScenes(SnippetSceneInput{Plan: p, Spans: spans, BeatEndMs: ends, DurationMs: len(p.Beats) * 5000})
	if err != nil {
		t.Fatalf("spotlightScenes: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("got %d scenes, want 1 — the card persists for the whole clip", len(scenes))
	}
	steps, ok := scenes[0].Props["steps"].([]map[string]any)
	if !ok {
		t.Fatalf("steps are %T, want []map[string]any", scenes[0].Props["steps"])
	}
	for i, want := range []int{0, 1, 2, 3, 3} {
		if got := steps[i]["shown"]; got != want {
			t.Errorf("step %d shows %v rows, want %d", i, got, want)
		}
	}
}
