package pipeline

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// illustrationPlan is a well-formed four-shot clip: a problem, an idea, a
// catch and a payoff, each on its own figure.
func illustrationPlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "illustration",
		Title:    "Why caching wins",
		Subtitle: "The cheapest performance fix you own",
		Beats: []SnippetBeat{
			{ID: "the-waste", Heading: "Your server answers the same question", Narration: strings.Repeat("waste ", 22),
				Art: &ArtBeat{Figure: "gears", Emphasis: "same question", Caption: "Every request recomputes a result that has not changed."}},
			{ID: "the-idea", Heading: "A cache remembers the answer", Narration: strings.Repeat("idea ", 22),
				Art: &ArtBeat{Figure: "lightbulb", Emphasis: "remembers", Caption: "Store it once, hand it back until it stops being true."}},
			{ID: "the-catch", Heading: "Stale data is the real cost", Narration: strings.Repeat("catch ", 22),
				Art: &ArtBeat{Figure: "clock", Emphasis: "Stale", Caption: "The hard part is deciding when the answer expires."}},
			{ID: "the-payoff", Heading: "Ninety percent fewer queries", Narration: strings.Repeat("payoff ", 22),
				Art: &ArtBeat{Figure: "chart", Emphasis: "Ninety percent", Caption: "The database only sees work that actually changed."}},
		},
	}
}

func TestValidateIllustrationPlan(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := illustrationPlan().Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	t.Run("beat without art", func(t *testing.T) {
		p := illustrationPlan()
		p.Beats[2].Art = nil
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "has no art") {
			t.Fatalf("want missing-art error, got %v", err)
		}
	})
	t.Run("heading too long", func(t *testing.T) {
		p := illustrationPlan()
		p.Beats[0].Heading = "this heading runs on and on and on for far too many words"
		p.Beats[0].Art.Emphasis = ""
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "headlines are") {
			t.Fatalf("want headline-length error, got %v", err)
		}
	})
	t.Run("heading too short", func(t *testing.T) {
		p := illustrationPlan()
		p.Beats[0].Heading = "Caching"
		p.Beats[0].Art.Emphasis = ""
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "headlines are") {
			t.Fatalf("want headline-length error, got %v", err)
		}
	})
	// The emphasis is a marker stroke drawn under part of the headline. A
	// phrase that is not in the headline has nothing to underline, and the shot
	// would quietly lose its accent rather than fail.
	t.Run("emphasis not in heading", func(t *testing.T) {
		p := illustrationPlan()
		p.Beats[1].Art.Emphasis = "forgets"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "does not appear in its heading") {
			t.Fatalf("want emphasis error, got %v", err)
		}
	})
	t.Run("caption too long", func(t *testing.T) {
		p := illustrationPlan()
		p.Beats[1].Art.Caption = strings.Repeat("word ", maxCaptionWords+1)
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "caption") {
			t.Fatalf("want caption-length error, got %v", err)
		}
	})
	// A run of shots on one drawing is a still image with the text changing,
	// which is the one thing this template exists to avoid.
	t.Run("one figure over-used", func(t *testing.T) {
		p := illustrationPlan()
		p.Beats[1].Art.Figure = "gears"
		p.Beats[2].Art.Figure = "gears"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "use a different figure") {
			t.Fatalf("want figure-variety error, got %v", err)
		}
	})
	// Two beats sharing a figure is a callback, not a rut.
	t.Run("one figure repeated once is fine", func(t *testing.T) {
		p := illustrationPlan()
		p.Beats[1].Art.Figure = "gears"
		if err := p.Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	t.Run("foreign beat field", func(t *testing.T) {
		p := illustrationPlan()
		p.Beats[0].Sketch = []SketchItem{{Label: "Browser", Icon: "monitor"}}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "does not use") {
			t.Fatalf("want foreign-field error, got %v", err)
		}
	})
}

// The emphasis and the heading arrive in different fields of the same reply and
// models are inconsistent about echoing case and punctuation. Rejecting a pair
// that differs only in those buys a correction round that teaches nothing.
func TestEmphasisMatchingIsLoose(t *testing.T) {
	cases := []struct {
		heading  string
		emphasis string
		want     bool
	}{
		{"Ship it — twice.", "Twice", true},
		{"A cache remembers the answer", "REMEMBERS", true},
		{"Ninety percent fewer queries", "ninety percent", true},
		{"Stale data is the real cost", "expensive", false},
		{"Your server answers the same question", "same  question", true},
	}
	for _, c := range cases {
		if got := containsPhrase(c.heading, c.emphasis); got != c.want {
			t.Errorf("containsPhrase(%q, %q) = %v, want %v", c.heading, c.emphasis, got, c.want)
		}
	}
}

func TestIllustrationScenes(t *testing.T) {
	plan := illustrationPlan()
	scenes, err := illustrationScenes(sceneInput(t, plan, 6000))
	if err != nil {
		t.Fatal(err)
	}
	// One shot per beat, and no title card: the first headline is already big
	// type on an empty stage, so a card in front of it says it twice.
	if len(scenes) != len(plan.Beats) {
		t.Fatalf("got %d scenes, want one per beat (%d)", len(scenes), len(plan.Beats))
	}
	for i, s := range scenes {
		if s.Type != SceneIllustration {
			t.Errorf("scene %d is %q, want %q", i, s.Type, SceneIllustration)
		}
		if s.StartMs != i*6000 || s.EndMs != (i+1)*6000 {
			t.Errorf("scene %d spans %d-%d, want %d-%d", i, s.StartMs, s.EndMs, i*6000, (i+1)*6000)
		}
		if got := s.Props["headline"]; got != plan.Beats[i].Heading {
			t.Errorf("scene %d headline = %v, want %q", i, got, plan.Beats[i].Heading)
		}
		// The figure alternates sides so a run of cuts does not read as one
		// slide with the words swapped out.
		if got, want := s.Props["flip"], i%2 == 1; got != want {
			t.Errorf("scene %d flip = %v, want %v", i, got, want)
		}
	}
}

// An invented figure name has to reach the renderer as the fallback, not as
// itself — the renderer would draw nothing at all.
func TestIllustrationScenesNormalizeFigure(t *testing.T) {
	plan := illustrationPlan()
	plan.Beats[0].Art.Figure = "unicorn"
	scenes, err := illustrationScenes(sceneInput(t, plan, 6000))
	if err != nil {
		t.Fatal(err)
	}
	if got := scenes[0].Props["figure"]; got != "spark" {
		t.Errorf("figure = %v, want spark", got)
	}
}

// The prompt ships an example reply, and the example has to satisfy the rules
// the same prompt states.
//
// This is not hypothetical tidiness. The flow template's prompt demanded a
// graph that forks, and its own example was a straight chain; the model
// dutifully copied the example and every generation failed validation. The
// prompt was the last place anyone thought to look. Now the example is checked
// against the validator it is supposed to demonstrate.
func TestIllustrationPromptExampleIsValid(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("../../prompts", snippetIllustrationTemplateName))
	if err != nil {
		t.Fatalf("reading prompt: %v", err)
	}
	at := bytes.Index(src, []byte(`{"title":`))
	if at < 0 {
		t.Fatalf("no example reply found in %s", snippetIllustrationTemplateName)
	}
	line := src[at:]
	if end := bytes.IndexByte(line, '\n'); end >= 0 {
		line = line[:end]
	}

	var plan SnippetPlan
	if err := json.Unmarshal(line, &plan); err != nil {
		t.Fatalf("the example reply in %s is not valid JSON: %v", snippetIllustrationTemplateName, err)
	}
	plan.Template = "illustration"
	// The example elides its narration to "...", so the shared word-budget rule
	// cannot apply to it. Everything the example actually demonstrates — the
	// headline lengths, the emphasis matching its heading, the figure variety,
	// the beat count — is what this fills in to reach.
	for i := range plan.Beats {
		plan.Beats[i].Narration = strings.Repeat("narration ", 22)
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("the example reply in %s does not satisfy the rules that same prompt states: %v",
			snippetIllustrationTemplateName, err)
	}
	// The floor is the absolute one now: how many beats a prompt asks for
	// depends on the runtime it is rendered at (beatBounds), and the example is
	// rendered without one.
	if len(plan.Beats) < floorSnippetBeats {
		t.Errorf("the example shows %d beats, below the floor of %d", len(plan.Beats), floorSnippetBeats)
	}
}
