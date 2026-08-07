package pipeline

import (
	"strings"
	"testing"
)

const swNarration = "It is free until about a thousand rows, and per-seat pricing after that."

func showcasePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "showcase",
		Title:    "Airtable, and when not to use it",
		Showcase: &ShowcaseSpec{
			Name:     "Airtable",
			Category: "Database",
			Tagline:  "A database that looks and feels like a spreadsheet",
			Icon:     "database",
			Facts: []ShowcaseFact{
				{Label: "Best for", Value: "Small structured datasets"},
				{Label: "Price", Value: "Free to 1,000 rows"},
				{Label: "Lock-in", Value: "CSV out, formulas stay"},
			},
			Strengths: []string{"Non-technical people can edit it", "Views without queries"},
			Limits:    []string{"Slows badly past fifty thousand rows"},
		},
		Beats: []SnippetBeat{
			{ID: "meet", Heading: "Meet Airtable", Narration: swNarration, Showcase: &ShowcaseBeat{Show: "intro"}},
			{ID: "best", Heading: "What it is for", Narration: swNarration, Showcase: &ShowcaseBeat{Show: "fact", At: 0}},
			{ID: "price", Heading: "What it costs", Narration: swNarration, Showcase: &ShowcaseBeat{Show: "fact", At: 1}},
			{ID: "exit", Heading: "Getting data out", Narration: swNarration, Showcase: &ShowcaseBeat{Show: "fact", At: 2}},
			{ID: "good", Heading: "Where it shines", Narration: swNarration, Showcase: &ShowcaseBeat{Show: "strengths"}},
			{ID: "bad", Heading: "Where it hurts", Narration: swNarration, Showcase: &ShowcaseBeat{Show: "limits"}},
			{ID: "demo", Heading: "See it working", Narration: swNarration, Showcase: &ShowcaseBeat{Show: "handoff"}},
		},
	}
	// Seven beats only fit once the budget is big enough to fund them, which is
	// what this template raises MaxBeats for.
	p.targetWords = 7 * 40
	return p
}

func TestShowcasePlanAccepted(t *testing.T) {
	if err := validateShowcasePlan(showcasePlan()); err != nil {
		t.Fatalf("a well-formed showcase was rejected: %v", err)
	}
}

// The rule this template exists for: a showcase with no honest limitation is an
// advert, and the model will not write the awkward half unless required to.
func TestShowcaseRequiresAnHonestLimitation(t *testing.T) {
	p := showcasePlan()
	p.Showcase.Limits = nil
	err := validateShowcasePlan(p)
	if err == nil {
		t.Fatal("a showcase with nothing to watch out for was accepted")
	}
	if !strings.Contains(err.Error(), "advert") {
		t.Errorf("the error should say what such a clip actually is; got: %v", err)
	}
	if !strings.Contains(err.Error(), "Airtable") {
		t.Errorf("the error should name the tool it is asking about; got: %v", err)
	}
}

// Limits written on the card but never spoken is the same as not saying them.
func TestShowcaseRequiresALimitsBeat(t *testing.T) {
	p := showcasePlan()
	p.Beats[5].Showcase = &ShowcaseBeat{Show: "strengths"}
	err := validateShowcasePlan(p)
	if err == nil {
		t.Fatal("a clip that never speaks its limits was accepted")
	}
	if !strings.Contains(err.Error(), "never spoken") {
		t.Errorf("the error should say why the card alone is not enough; got: %v", err)
	}
}

// The hand-off is the cut point, so nothing may follow it.
func TestShowcaseMustEndOnTheHandoff(t *testing.T) {
	p := showcasePlan()
	p.Beats[6].Showcase = &ShowcaseBeat{Show: "strengths"}
	err := validateShowcasePlan(p)
	if err == nil {
		t.Fatal("a clip with no hand-off was accepted")
	}
	if !strings.Contains(err.Error(), "hand-off") {
		t.Errorf("the error should name what is missing; got: %v", err)
	}

	p = showcasePlan()
	p.Beats = append(p.Beats, SnippetBeat{
		ID: "extra", Heading: "One more thing", Narration: swNarration,
		Showcase: &ShowcaseBeat{Show: "strengths"},
	})
	p.targetWords = 8 * 40
	err = validateShowcasePlan(p)
	if err == nil {
		t.Fatal("a beat after the hand-off was accepted")
	}
	if !strings.Contains(err.Error(), "cut point") {
		t.Errorf("the error should say why nothing follows it; got: %v", err)
	}
}

func TestShowcaseOpensOnTheIntro(t *testing.T) {
	p := showcasePlan()
	p.Beats[0].Showcase = &ShowcaseBeat{Show: "fact", At: 0}
	p.Beats[1].Showcase = &ShowcaseBeat{Show: "intro"}
	err := validateShowcasePlan(p)
	if err == nil {
		t.Fatal("a clip that opens on a fact was accepted")
	}
	if !strings.Contains(err.Error(), "meeting the tool") {
		t.Errorf("the error should say how to open; got: %v", err)
	}
}

func TestShowcaseStatesEveryFact(t *testing.T) {
	p := showcasePlan()
	p.Beats = append(p.Beats[:3], p.Beats[4:]...) // "Lock-in" never stated
	p.targetWords = 6 * 40
	err := validateShowcasePlan(p)
	if err == nil {
		t.Fatal("a fact with no beat was accepted")
	}
	if !strings.Contains(err.Error(), "never said out loud") {
		t.Errorf("the error should name the silent cell; got: %v", err)
	}
}

func TestShowcaseRejectsDuplicateFactLabels(t *testing.T) {
	p := showcasePlan()
	p.Showcase.Facts[2].Label = "Price"
	if err := validateShowcasePlan(p); err == nil {
		t.Error("two cells answering the same question were accepted")
	}
}

// The card's shape says what its ends are for, so an unlabelled beat is a blank
// field rather than a claim — inferring it costs no correction round.
func TestShowcaseNormalizeInfersTheShow(t *testing.T) {
	p := showcasePlan()
	p.Beats[0].Showcase.Show = ""
	p.Beats[6].Showcase.Show = "wrap-up" // not in the vocabulary
	normalizeShowcasePlan(p)
	if p.Beats[0].Showcase.Show != "intro" {
		t.Errorf("the first beat should be inferred as intro, got %q", p.Beats[0].Showcase.Show)
	}
	if p.Beats[6].Showcase.Show != "handoff" {
		t.Errorf("the last beat should be inferred as handoff, got %q", p.Beats[6].Showcase.Show)
	}
	if err := validateShowcasePlan(p); err != nil {
		t.Errorf("the repaired plan should validate: %v", err)
	}
}

func TestShowcaseNormalizeDropsHalfWrittenFacts(t *testing.T) {
	p := showcasePlan()
	p.Showcase.Facts = append(p.Showcase.Facts, ShowcaseFact{Label: "Support", Value: ""})
	p.Showcase.Tagline = "a database that looks and feels exactly like a spreadsheet you already know"
	normalizeShowcasePlan(p)

	if n := len(p.Showcase.Facts); n != 3 {
		t.Errorf("a labelled blank should be dropped, got %d facts", n)
	}
	if w := len(strings.Fields(p.Showcase.Tagline)); w > maxShowcaseTaglineWords {
		t.Errorf("tagline still %d words, want at most %d", w, maxShowcaseTaglineWords)
	}
	if err := validateShowcasePlan(p); err != nil {
		t.Fatalf("the normalized plan should validate: %v", err)
	}
}

// The template raises the shared seven-beat ceiling, and its own minimum shape
// is seven beats — so the two have to agree or nothing it plans can validate.
func TestShowcaseBeatCeilingFundsItsOwnShape(t *testing.T) {
	tpl := SnippetTemplates["showcase"]
	if tpl.MaxBeats < 8 {
		t.Fatalf("showcase needs room for intro + 4 facts + 2 columns + handoff; MaxBeats is %d", tpl.MaxBeats)
	}
	spec := SnippetSpec{Template: "showcase", Prompt: "x"}
	want, _, _ := wordBudget(spec.ResolvedTargetSec(), 175)
	minBeats, maxBeats, _, perBeat := beatBounds(want, templateBeatCeiling("showcase"), 0)
	if minBeats > 7 || maxBeats < 8 {
		t.Errorf("at the default runtime the range is %d-%d, which cannot hold a 7- or 8-beat card", minBeats, maxBeats)
	}
	if perBeat > maxWordsPerBeat {
		t.Errorf("the prompt would advise %d words a beat, over the %d maximum", perBeat, maxWordsPerBeat)
	}
}

func TestShowcaseScenesCoverTheWholeClip(t *testing.T) {
	p := showcasePlan()
	scenes, err := showcaseScenes(sceneInput(t, p, 6000))
	if err != nil {
		t.Fatalf("laying out the showcase failed: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneShowcase {
		t.Fatalf("want one %s scene, got %v", SceneShowcase, scenes)
	}
	steps, ok := scenes[0].Props["steps"].([]map[string]any)
	if !ok || len(steps) != len(p.Beats) {
		t.Fatalf("want one step per beat, got %v", scenes[0].Props["steps"])
	}
	if steps[len(steps)-1]["show"] != "handoff" {
		t.Error("the last step should be the hand-off the demo cuts onto")
	}
	// The limits have to reach the renderer or the enforced half is invisible.
	if limits, ok := scenes[0].Props["limits"].([]string); !ok || len(limits) == 0 {
		t.Error("the limits should reach the renderer")
	}
}
