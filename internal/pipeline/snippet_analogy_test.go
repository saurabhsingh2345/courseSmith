package pipeline

import (
	"strings"
	"testing"
)

const anNarration = "Most of the librarian's day goes on walking to the shelf, not on reading the page."

func analogyPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "analogy",
		Title:    "A librarian and a library",
		Analogy: &AnalogySpec{
			Familiar:     "A library",
			FamiliarIcon: "book",
			Real:         "Running a model",
			RealIcon:     "server",
			Pairs: []AnalogyPair{
				{From: "The size of the room", To: "Memory capacity", Note: "A book that will not fit cannot be read"},
				{From: "The walk to the shelf", To: "Memory bandwidth", Note: "Most of the day goes on walking"},
				{From: "How fast they read", To: "Compute", Note: "Rarely the bottleneck"},
			},
			Breaks: "A librarian can skim; a machine reads every word",
		},
		Beats: []SnippetBeat{
			{ID: "library", Heading: "Picture a library", Narration: anNarration, Analogy: &AnalogyBeat{Show: "picture"}},
			{ID: "room", Heading: "The room", Narration: anNarration, Analogy: &AnalogyBeat{Show: "pair", At: 0}},
			{ID: "walk", Heading: "The walk", Narration: anNarration, Analogy: &AnalogyBeat{Show: "pair", At: 1}},
			{ID: "reading", Heading: "The reading", Narration: anNarration, Analogy: &AnalogyBeat{Show: "pair", At: 2}},
			{ID: "breaks", Heading: "Where it breaks", Narration: anNarration, Analogy: &AnalogyBeat{Show: "breaks"}},
		},
	}
	p.targetWords = 5 * 40
	return p
}

func TestAnalogyPlanAccepted(t *testing.T) {
	if err := validateAnalogyPlan(analogyPlan()); err != nil {
		t.Fatalf("a well-formed analogy was rejected: %v", err)
	}
}

// The rule this template exists for. An analogy that never admits its limits
// becomes the learner's actual mental model, wrong parts included.
func TestAnalogyMustSayWhereItBreaks(t *testing.T) {
	p := analogyPlan()
	p.Analogy.Breaks = ""
	err := validateAnalogyPlan(p)
	if err == nil {
		t.Fatal("an analogy with no limits was accepted")
	}
	if !strings.Contains(err.Error(), "actual mental model") {
		t.Errorf("the error does not say why it matters: %v", err)
	}
}

// Written down and never spoken is the same as not said.
func TestAnalogyBreaksMustBeNarrated(t *testing.T) {
	p := analogyPlan()
	p.Analogy.Pairs = append(p.Analogy.Pairs, AnalogyPair{From: "The catalogue", To: "The index"})
	p.Beats[4].Analogy = &AnalogyBeat{Show: "pair", At: 3}
	err := validateAnalogyPlan(p)
	if err == nil {
		t.Fatal("a written-but-unspoken limitation was accepted")
	}
	if !strings.Contains(err.Error(), "give it a beat") {
		t.Errorf("unexpected error: %v", err)
	}
}

// The other rule: nothing in the picture may point at nothing, and no two
// parts may point at the same thing.
func TestAnalogyRejectsBrokenMappings(t *testing.T) {
	p := analogyPlan()
	p.Analogy.Pairs[1].To = p.Analogy.Pairs[0].To
	err := validateAnalogyPlan(p)
	if err == nil {
		t.Fatal("two parts of the picture landing on one real thing was accepted")
	}
	if !strings.Contains(err.Error(), "starts lying") {
		t.Errorf("unexpected error: %v", err)
	}

	p = analogyPlan()
	p.Analogy.Pairs[2].From = p.Analogy.Pairs[0].From
	if err := validateAnalogyPlan(p); err == nil {
		t.Fatal("one part of the picture mapping to two things was accepted")
	}
}

// A half-written correspondence is dropped in normalize rather than reaching
// the validator, because inventing the missing half would be a claim about the
// subject.
func TestAnalogyNormalizeDropsHalfPairs(t *testing.T) {
	p := analogyPlan()
	p.Analogy.Pairs = append(p.Analogy.Pairs, AnalogyPair{From: "The quiet", To: ""})
	normalizeAnalogyPlan(p)
	for _, pr := range p.Analogy.Pairs {
		if strings.TrimSpace(pr.From) == "" || strings.TrimSpace(pr.To) == "" {
			t.Error("a half-written correspondence survived normalize")
		}
	}
}

func TestAnalogyOpensOnThePicture(t *testing.T) {
	p := analogyPlan()
	p.Beats[0].Analogy = &AnalogyBeat{Show: "pair", At: 0}
	p.Beats[1].Analogy = &AnalogyBeat{Show: "picture"}
	err := validateAnalogyPlan(p)
	if err == nil {
		t.Fatal("a clip that mapped before setting the picture up was accepted")
	}
	if !strings.Contains(err.Error(), "maps onto nothing") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAnalogyWalksEveryPair(t *testing.T) {
	p := analogyPlan()
	// Add a fourth pair and leave the beats alone: exactly one correspondence
	// goes unwalked while every other rule still passes.
	p.Analogy.Pairs = append(p.Analogy.Pairs, AnalogyPair{From: "The catalogue", To: "The index"})
	err := validateAnalogyPlan(p)
	if err == nil {
		t.Fatal("a correspondence nobody walks was accepted")
	}
	if !strings.Contains(err.Error(), "never spoken") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAnalogyNormalizeRepairs(t *testing.T) {
	p := analogyPlan()
	p.Analogy.FamiliarIcon = "not-an-icon"
	p.Beats[2].Analogy.At = 99
	p.Beats[3].Analogy.Show = "nonsense"
	normalizeAnalogyPlan(p)

	if p.Analogy.FamiliarIcon != "book" {
		t.Errorf("an unknown icon became %q, want the fallback", p.Analogy.FamiliarIcon)
	}
	if p.Beats[2].Analogy.At != len(p.Analogy.Pairs)-1 {
		t.Errorf("an out-of-range beat points at %d", p.Beats[2].Analogy.At)
	}
	if p.Beats[3].Analogy.Show != "pair" {
		t.Errorf("an unknown show became %q, want pair", p.Beats[3].Analogy.Show)
	}
}

// An unlabelled beat is a blank field, not a claim: the shape says the clip
// opens on the picture and closes on the admission.
func TestAnalogyNormalizeInfersTheEnds(t *testing.T) {
	p := analogyPlan()
	p.Beats[0].Analogy.Show = ""
	p.Beats[4].Analogy.Show = ""
	normalizeAnalogyPlan(p)
	if p.Beats[0].Analogy.Show != "picture" {
		t.Errorf("the first beat became %q, want picture", p.Beats[0].Analogy.Show)
	}
	if p.Beats[4].Analogy.Show != "breaks" {
		t.Errorf("the last beat became %q, want breaks", p.Beats[4].Analogy.Show)
	}
}

func TestAnalogyScenesShape(t *testing.T) {
	p := analogyPlan()
	scenes, err := analogyScenes(sceneInput(t, p, 8000))
	if err != nil {
		t.Fatalf("analogyScenes: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneAnalogy {
		t.Fatalf("want one analogy scene, got %d of %q", len(scenes), scenes[0].Type)
	}
	// Both columns and the admission are scene-level: everything is on screen
	// for the whole clip and only the lit row moves.
	for _, key := range []string{"familiar", "real", "breaks", "pairs"} {
		if scenes[0].Props[key] == nil {
			t.Errorf("%q is missing from the scene", key)
		}
	}
	pairs := scenes[0].Props["pairs"].([]map[string]any)
	if len(pairs) != 3 || pairs[1]["to"] != "Memory bandwidth" {
		t.Errorf("the mapping did not reach the scene intact: %v", pairs)
	}
}
