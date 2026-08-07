package pipeline

import (
	"strings"
	"testing"
)

const bandNarration = "The segment lights up and the one thing that made this generation different rises inside it."

func erasPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "eras",
		Title:    "From glass tubes to the cloud",
		Eras: &ErasSpec{
			Eras: []ErasEra{
				{Label: "vacuum tubes", When: "1940s", Mark: "rooms full of glass and heat", Carry: "the idea of a stored program"},
				{Label: "transistors", When: "1950s", Mark: "the same logic in a solid sliver", Carry: "cheap and reliable switching"},
				{Label: "integrated circuits", When: "1960s", Mark: "thousands of gates on one wafer", Carry: "everything shrinks together"},
				{Label: "the microprocessor", When: "1970s", Mark: "a whole processor on a single chip", Carry: "computing becomes personal"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "the-band", Heading: "Four generations", Narration: bandNarration, Eras: &ErasBeat{Show: "band"}},
			{ID: "tubes", Heading: "Glass and heat", Narration: bandNarration, Eras: &ErasBeat{Show: "era", At: 0}},
			{ID: "transistors", Heading: "Solid state", Narration: bandNarration, Eras: &ErasBeat{Show: "era", At: 1}},
			{ID: "chips", Heading: "Many gates", Narration: bandNarration, Eras: &ErasBeat{Show: "era", At: 2}},
			{ID: "micros", Heading: "One chip", Narration: bandNarration, Eras: &ErasBeat{Show: "era", At: 3}},
			{ID: "the-thread", Heading: "What carried over", Narration: bandNarration, Eras: &ErasBeat{Show: "thread"}},
			{ID: "today", Heading: "Where that leaves us", Narration: bandNarration, Eras: &ErasBeat{Show: "now"}},
		},
	}
	p.targetWords = 7 * 40
	return p
}

func TestErasPlanAccepted(t *testing.T) {
	if err := validateErasPlan(erasPlan()); err != nil {
		t.Fatalf("a well-formed era band was rejected: %v", err)
	}
}

// History is walked forward, once through. An era beat out of order is the way
// this template quietly stops arguing for causation.
func TestErasRejectsAnEraOutOfOrder(t *testing.T) {
	p := erasPlan()
	p.Beats[2].Eras = &ErasBeat{Show: "era", At: 3}
	err := validateErasPlan(p)
	if err == nil {
		t.Fatal("an era lit out of order was accepted, and history does not run backwards")
	}
	if !strings.Contains(err.Error(), "1970s") || !strings.Contains(err.Error(), "1950s") {
		t.Fatalf("the error does not quote the era lit and the era that was due: %v", err)
	}
	if !strings.Contains(err.Error(), "forward") {
		t.Fatalf("the error does not say why the order matters: %v", err)
	}
}

func TestErasRejectsAnEraLitTwice(t *testing.T) {
	p := erasPlan()
	p.Beats[3].Eras = &ErasBeat{Show: "era", At: 1}
	err := validateErasPlan(p)
	if err == nil {
		t.Fatal("an era lit twice was accepted")
	}
	if !strings.Contains(err.Error(), "transistors") {
		t.Fatalf("the error does not name the repeated era: %v", err)
	}
}

func TestErasRejectsLeavingAnEraUnlit(t *testing.T) {
	p := erasPlan()
	p.Eras.Eras = append(p.Eras.Eras, ErasEra{Label: "the network", When: "1990s", Mark: "the machines start talking to each other"})
	err := validateErasPlan(p)
	if err == nil {
		t.Fatal("an era nobody narrated was accepted")
	}
	if !strings.Contains(err.Error(), "the network") || !strings.Contains(err.Error(), "4 of 5") {
		t.Fatalf("the error does not name and count the era left dark: %v", err)
	}
}

// One hand-off is an anecdote, not a thread, and the count is done in Go.
func TestErasRejectsAThreadWithOneCarry(t *testing.T) {
	p := erasPlan()
	for i := range p.Eras.Eras {
		if i > 0 {
			p.Eras.Eras[i].Carry = ""
		}
	}
	err := validateErasPlan(p)
	if err == nil {
		t.Fatal("a thread beat drawn over a single hand-off was accepted")
	}
	if !strings.Contains(err.Error(), "only 1 era") {
		t.Fatalf("the error does not count the carries: %v", err)
	}
}

func TestErasAcceptsATwoCarryThread(t *testing.T) {
	p := erasPlan()
	for i := range p.Eras.Eras {
		if i > 1 {
			p.Eras.Eras[i].Carry = ""
		}
	}
	if err := validateErasPlan(p); err != nil {
		t.Fatalf("two hand-offs are a thread, but the plan was rejected: %v", err)
	}
}

func TestErasRequiresOpeningOnTheBand(t *testing.T) {
	p := erasPlan()
	p.Beats[0].Eras = &ErasBeat{Show: "era", At: 0}
	if err := validateErasPlan(p); err == nil {
		t.Fatal("an era lighting before the band was shown was accepted")
	}
}

func TestErasRequiresClosingOnNow(t *testing.T) {
	p := erasPlan()
	p.Beats[6].Eras = &ErasBeat{Show: "thread"}
	err := validateErasPlan(p)
	if err == nil {
		t.Fatal("a history that never arrives at the present was accepted")
	}
	if !strings.Contains(err.Error(), "now") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestErasRejectsAnEraCountOutOfRange(t *testing.T) {
	p := erasPlan()
	p.Eras.Eras = p.Eras.Eras[:2]
	if err := validateErasPlan(p); err == nil {
		t.Fatal("a two-era band was accepted, and two eras is a before and an after")
	}
}

func TestErasRejectsAnEraWithNoMark(t *testing.T) {
	p := erasPlan()
	p.Eras.Eras[2].Mark = ""
	err := validateErasPlan(p)
	if err == nil {
		t.Fatal("an era with no defining thing was accepted")
	}
	if !strings.Contains(err.Error(), "integrated circuits") {
		t.Fatalf("the error does not name the era: %v", err)
	}
}

// Over-long phrases and a stray index are phrasing, not wrong answers, so they
// are repaired rather than argued with.
func TestErasNormalizeRepairsLabelsAndIndices(t *testing.T) {
	p := erasPlan()
	p.Eras.Eras[0].Label = "the great big vacuum tube age"
	p.Eras.Eras[0].When = "the   late   1940s"
	p.Eras.Eras[1].Mark = "the very same logic implemented in a small solid sliver of doped silicon"
	p.Eras.Eras[2].Carry = "everything shrinks together and it keeps on shrinking every single year"
	p.Beats[1].Eras.At = 88
	normalizeErasPlan(p)

	if n := len(strings.Fields(p.Eras.Eras[0].Label)); n != maxErasLabelWords {
		t.Fatalf("the label normalized to %d words, want %d", n, maxErasLabelWords)
	}
	if p.Eras.Eras[0].When != "the late" {
		t.Fatalf("the date normalized to %q, want the two-word clamp", p.Eras.Eras[0].When)
	}
	if n := len(strings.Fields(p.Eras.Eras[1].Mark)); n != maxErasMarkWords {
		t.Fatalf("the mark normalized to %d words, want %d", n, maxErasMarkWords)
	}
	if n := len(strings.Fields(p.Eras.Eras[2].Carry)); n != maxErasCarryWords {
		t.Fatalf("the carry normalized to %d words, want %d", n, maxErasCarryWords)
	}
	if p.Beats[1].Eras.At != 3 {
		t.Fatalf("an out-of-range era index normalized to %d, want the last era", p.Beats[1].Eras.At)
	}
}

func TestErasShowDefaultsToEra(t *testing.T) {
	b := ErasBeat{Show: "epoch"}
	if got := b.ResolvedShow(); got != "era" {
		t.Fatalf("an unknown show resolved to %q, want era", got)
	}
}

// The component draws a thread it was handed: the arcs are paired in Go, and
// the last era's carry becomes the line to today rather than a hop to nowhere.
func TestErasScenesPairTheHandOffs(t *testing.T) {
	p := erasPlan()
	scenes, err := erasScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	threads, _ := props["threads"].([]map[string]any)
	if len(threads) != 3 {
		t.Fatalf("four eras with a carry each give 3 arcs between them, got %d", len(threads))
	}
	if threads[0]["from"] != 0 || threads[0]["to"] != 1 {
		t.Fatalf("the first arc runs %v to %v, want 0 to 1", threads[0]["from"], threads[0]["to"])
	}
	if props["carryNow"] != "computing becomes personal" {
		t.Fatalf("the last era's carry became %v, want the line to today", props["carryNow"])
	}

	steps, _ := props["steps"].([]map[string]any)
	first := steps[0]
	if first["show"] != "band" {
		t.Fatalf("the first step shows %v, want band", first["show"])
	}
	if up, _ := first["lit"].([]int); len(up) != 0 {
		t.Fatalf("no era is lit on the opener, but the step reports %v", up)
	}

	last := steps[len(steps)-1]
	if last["show"] != "now" {
		t.Fatalf("the last step shows %v, want now", last["show"])
	}
	if last["at"] != 3 {
		t.Fatalf("the closer focuses %v, want the last era", last["at"])
	}
	up, _ := last["lit"].([]int)
	if len(up) != 4 || up[0] != 0 || up[3] != 3 {
		t.Fatalf("the closer lights %v, want every era sorted", up)
	}
}
