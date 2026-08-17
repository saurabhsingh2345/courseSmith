package pipeline

import (
	"strings"
	"testing"
)

const faceOffNarration = "Both cards are up from the first frame, and the bars only fill once both of the things on them have been introduced."

func duelPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "duel",
		Title:    "Free ChatGPT or paid Gemini",
		Duel: &DuelSpec{
			Axis: "capability",
			Pick: 1,
			Sides: []Contender{
				{Title: "ChatGPT", Tag: "Free", Note: "The most mature chatbot, on its older model", Score: 42, Role: "neutral", Site: "openai.com", Icon: "message"},
				{Title: "Gemini", Tag: "$20 a month", Note: "Reads a whole repository in one pass", Score: 88, Role: "rival", Brand: "googlegemini", Icon: "sparkles"},
			},
			Verdict: "Pay for one model if you use it daily; free tiers are for trying it out",
		},
		Beats: []SnippetBeat{
			{ID: "two-tiers", Heading: "Free against paid", Narration: faceOffNarration, Duel: &DuelBeat{Show: "pair"}},
			{ID: "the-free-one", Heading: "What free gives", Narration: faceOffNarration, Duel: &DuelBeat{Show: "card", At: 0}},
			{ID: "the-paid-one", Heading: "What twenty buys", Narration: faceOffNarration, Duel: &DuelBeat{Show: "card", At: 1}},
			{ID: "the-gap", Heading: "The size of it", Narration: faceOffNarration, Duel: &DuelBeat{Show: "bars"}},
			{ID: "the-call", Heading: "Which one", Narration: faceOffNarration, Duel: &DuelBeat{Show: "call"}},
		},
	}
	p.targetWords = 5 * 26
	return p
}

func TestDuelPlanAccepted(t *testing.T) {
	if err := validateDuelPlan(duelPlan()); err != nil {
		t.Fatalf("a well-formed face-off was rejected: %v", err)
	}
}

// The rule the whole template rests on. Two bars of the same length is a picture
// that says nothing, and the message has to say what to do instead rather than
// just refusing.
func TestDuelRejectsTwoBarsTheSameLength(t *testing.T) {
	p := duelPlan()
	p.Duel.Sides[1].Score = 48 // six apart, under the floor
	err := validateDuelPlan(p)
	if err == nil {
		t.Fatal("a face-off whose two bars are the same length was accepted")
	}
	if !strings.Contains(err.Error(), "capability") || !strings.Contains(err.Error(), "versus") {
		t.Fatalf("the error does not name the axis or point at the template that handles a genuine tie: %v", err)
	}
}

// The pick is allowed to be the SHORTER bar, and that is the most useful thing
// this template can say. A validator that quietly forced the pick to follow the
// measurement would make the case unsayable.
func TestDuelAllowsPickingTheShorterBar(t *testing.T) {
	p := duelPlan()
	p.Duel.Pick = 0
	p.Duel.Verdict = "Start on the free tier; only pay once you are using it every day"
	if err := validateDuelPlan(p); err != nil {
		t.Fatalf("picking the shorter bar was rejected, which is the case the template is for: %v", err)
	}
}

// A bar drawn for something the viewer has not met is a number about a stranger.
func TestDuelRejectsBarsBeforeBothSidesAreIntroduced(t *testing.T) {
	p := duelPlan()
	p.Beats = []SnippetBeat{
		{ID: "two-tiers", Heading: "Free against paid", Narration: faceOffNarration, Duel: &DuelBeat{Show: "pair"}},
		{ID: "the-free-one", Heading: "What free gives", Narration: faceOffNarration, Duel: &DuelBeat{Show: "card", At: 0}},
		{ID: "the-gap", Heading: "The size of it", Narration: faceOffNarration, Duel: &DuelBeat{Show: "bars"}},
		{ID: "the-paid-one", Heading: "What twenty buys", Narration: faceOffNarration, Duel: &DuelBeat{Show: "card", At: 1}},
		{ID: "the-call", Heading: "Which one", Narration: faceOffNarration, Duel: &DuelBeat{Show: "call"}},
	}
	err := validateDuelPlan(p)
	if err == nil {
		t.Fatal("bars filled before both sides were introduced")
	}
	if !strings.Contains(err.Error(), "Gemini") {
		t.Fatalf("the error does not name the side that had not been introduced yet: %v", err)
	}
}

// The pill is required because it is where the reason the shorter bar wins tends
// to live — "Free" against "$20 a month" is most of the argument.
func TestDuelRequiresATag(t *testing.T) {
	p := duelPlan()
	p.Duel.Sides[0].Tag = ""
	err := validateDuelPlan(p)
	if err == nil {
		t.Fatal("a side with no tag was accepted")
	}
	if !strings.Contains(err.Error(), "ChatGPT") {
		t.Fatalf("the error does not name the side missing its tag: %v", err)
	}
}

func TestDuelRequiresAnAxis(t *testing.T) {
	p := duelPlan()
	p.Duel.Axis = ""
	if err := validateDuelPlan(p); err == nil {
		t.Fatal("two unlabelled bars were accepted — a chart with no units")
	}
}

// Same rule as the cards closer and the versus verdict: a name is a preference,
// not a call.
func TestDuelRejectsAVerdictThatIsJustAName(t *testing.T) {
	p := duelPlan()
	p.Duel.Verdict = "Gemini"
	err := validateDuelPlan(p)
	if err == nil {
		t.Fatal("a verdict that is just one side's name was accepted")
	}
	if !strings.Contains(err.Error(), "shorter bar") {
		t.Fatalf("the error does not tell the model what a call needs to contain: %v", err)
	}
}

// The scene graph has to carry the bars as a latched state: once they have filled
// they stay filled, so the component paints a frame from one step rather than
// replaying the timeline.
func TestDuelScenesLatchTheBars(t *testing.T) {
	p := duelPlan()
	normalizeDuelPlan(p)
	if err := validateDuelPlan(p); err != nil {
		t.Fatalf("fixture rejected: %v", err)
	}
	spans := make([]SectionSpan, len(p.Beats))
	ends := make([]int, len(p.Beats))
	for i, b := range p.Beats {
		spans[i] = SectionSpan{ID: b.ID, StartMs: i * 5000, EndMs: (i + 1) * 5000}
		ends[i] = (i + 1) * 5000
	}
	scenes, err := duelScenes(SnippetSceneInput{Plan: p, Spans: spans, BeatEndMs: ends, DurationMs: len(p.Beats) * 5000})
	if err != nil {
		t.Fatalf("duelScenes: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("got %d scenes, want 1 — the pair persists for the whole clip", len(scenes))
	}
	steps, ok := scenes[0].Props["steps"].([]map[string]any)
	if !ok {
		t.Fatalf("steps are %T, want []map[string]any", scenes[0].Props["steps"])
	}
	want := []bool{false, false, false, true, true}
	for i, w := range want {
		if got := steps[i]["bars"]; got != w {
			t.Errorf("step %d bars = %v, want %v — the fill latches on its beat and holds", i, got, w)
		}
	}
}
