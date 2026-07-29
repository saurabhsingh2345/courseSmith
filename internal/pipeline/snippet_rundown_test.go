package pipeline

import (
	"strings"
	"testing"
)

const rdNarration = "Bandwidth is the one nobody reads off the box, and it is the one that decides your speed."

func rundownPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "rundown",
		Title:    "The three numbers that decide everything",
		Rundown: &RundownSpec{
			Promise: "Three numbers decide everything",
			Items: []RundownItem{
				{Label: "Memory capacity", Detail: "How much the model needs", Icon: "database"},
				{Label: "Memory bandwidth", Detail: "The hidden boss", Icon: "zap"},
				{Label: "Compute", Detail: "The overrated one", Icon: "gear"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "three", Heading: "Three numbers", Narration: rdNarration, Rundown: &RundownBeat{Show: "promise"}},
			{ID: "capacity", Heading: "How big", Narration: rdNarration, Rundown: &RundownBeat{Show: "item", At: 0}},
			{ID: "bandwidth", Heading: "The hidden boss", Narration: rdNarration, Rundown: &RundownBeat{Show: "item", At: 1}},
			{ID: "compute", Heading: "The overrated one", Narration: rdNarration, Rundown: &RundownBeat{Show: "item", At: 2}},
			{ID: "all", Heading: "All three", Narration: rdNarration, Rundown: &RundownBeat{Show: "all"}},
		},
	}
	p.targetWords = 5 * 40
	return p
}

func TestRundownPlanAccepted(t *testing.T) {
	if err := validateRundownPlan(rundownPlan()); err != nil {
		t.Fatalf("a well-formed rundown was rejected: %v", err)
	}
}

// The rule this template exists for: the promise has to be true.
func TestRundownPromiseMustMatchTheCount(t *testing.T) {
	p := rundownPlan()
	p.Rundown.Promise = "Four numbers decide everything"
	err := validateRundownPlan(p)
	if err == nil {
		t.Fatal("a promise of four with three cards was accepted")
	}
	if !strings.Contains(err.Error(), "only agreement") {
		t.Errorf("the error does not say why it matters: %v", err)
	}
}

// The count can be spelled or written in digits, and can sit anywhere in the
// line — a promise is a sentence, not a field.
func TestRundownReadsTheCountAnywhere(t *testing.T) {
	for _, promise := range []string{
		"Three numbers decide everything",
		"It comes down to 3 things",
		"THREE things break at scale",
		"Everything hinges on three numbers.",
	} {
		p := rundownPlan()
		p.Rundown.Promise = promise
		if err := validateRundownPlan(p); err != nil {
			t.Errorf("the promise %q was rejected against three cards: %v", promise, err)
		}
	}
	// And a promise with no number in it is simply not checked — "It comes down
	// to a handful of things" is a legitimate opening.
	p := rundownPlan()
	p.Rundown.Promise = "It comes down to a handful"
	if err := validateRundownPlan(p); err != nil {
		t.Errorf("a promise with no count was rejected: %v", err)
	}
}

func TestRundownOpensOnThePromise(t *testing.T) {
	p := rundownPlan()
	p.Beats[0].Rundown = &RundownBeat{Show: "item", At: 0}
	p.Beats[1].Rundown = &RundownBeat{Show: "promise"}
	err := validateRundownPlan(p)
	if err == nil {
		t.Fatal("a clip that lit a card before promising was accepted")
	}
	if !strings.Contains(err.Error(), "nobody agreed to watch") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRundownCoversEveryCard(t *testing.T) {
	p := rundownPlan()
	// Add a fourth card and leave the beats alone, so exactly one card goes
	// uncovered while every other rule still passes — the promise matches the
	// new count, no card repeats, and the beat count is untouched. Deleting a
	// beat instead would trip the shared beat floor and prove nothing.
	p.Rundown.Promise = "Four numbers decide everything"
	p.Rundown.Items = append(p.Rundown.Items, RundownItem{Label: "Power draw", Detail: "The one nobody budgets for"})

	err := validateRundownPlan(p)
	if err == nil {
		t.Fatal("a card nobody covers was accepted")
	}
	if !strings.Contains(err.Error(), "never covered") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRundownRejectsRepeatedLabels(t *testing.T) {
	p := rundownPlan()
	p.Rundown.Items[2].Label = p.Rundown.Items[0].Label
	err := validateRundownPlan(p)
	if err == nil {
		t.Fatal("two cards with the same label were accepted")
	}
	if !strings.Contains(err.Error(), "padding the count") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRundownAllIsLast(t *testing.T) {
	p := rundownPlan()
	p.Beats[1].Rundown = &RundownBeat{Show: "all"}
	p.Beats[4].Rundown = &RundownBeat{Show: "item", At: 0}
	if err := validateRundownPlan(p); err == nil {
		t.Fatal("a closing beat in the middle was accepted")
	}
}

func TestRundownNormalizeRepairs(t *testing.T) {
	p := rundownPlan()
	p.Rundown.Items[0].Icon = "not-an-icon"
	p.Beats[2].Rundown.At = 99
	p.Beats[3].Rundown.Show = "nonsense"
	normalizeRundownPlan(p)

	if p.Rundown.Items[0].Icon != "box" {
		t.Errorf("an unknown icon became %q, want the neutral fallback", p.Rundown.Items[0].Icon)
	}
	if p.Beats[2].Rundown.At != len(p.Rundown.Items)-1 {
		t.Errorf("an out-of-range beat points at %d", p.Beats[2].Rundown.At)
	}
	if p.Beats[3].Rundown.Show != "item" {
		t.Errorf("an unknown show became %q, want item", p.Beats[3].Rundown.Show)
	}
}

func TestPromisedCount(t *testing.T) {
	for in, want := range map[string]int{
		"Three numbers decide everything": 3,
		"It comes down to 3 things":       3,
		"Four things break at scale":      4,
		"FIVE stages, in order":           5,
		"A handful of things":             0,
		"":                                0,
	} {
		got, ok := promisedCount(in)
		if want == 0 {
			if ok {
				t.Errorf("promisedCount(%q) found %d, want no count", in, got)
			}
			continue
		}
		if !ok || got != want {
			t.Errorf("promisedCount(%q) = %d, %v; want %d, true", in, got, ok, want)
		}
	}
}

func TestRundownScenesShape(t *testing.T) {
	p := rundownPlan()
	scenes, err := rundownScenes(sceneInput(t, p, 8000))
	if err != nil {
		t.Fatalf("rundownScenes: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneRundown {
		t.Fatalf("want one rundown scene, got %d of %q", len(scenes), scenes[0].Type)
	}
	items, ok := scenes[0].Props["items"].([]map[string]any)
	if !ok || len(items) != 3 {
		t.Fatalf("want three cards on the scene, got %v", scenes[0].Props["items"])
	}
	// Every card reaches the renderer, not just the covered one: the row is on
	// screen from the first frame and only brightness moves.
	if items[2]["label"] != "Compute" {
		t.Errorf("card 2 = %v, want the third item", items[2]["label"])
	}
}
