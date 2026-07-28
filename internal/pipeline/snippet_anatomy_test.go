package pipeline

import (
	"strings"
	"testing"
)

const anatNarration = "This one line is doing four separate jobs at the same time."

func anatomyPlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "anatomy",
		Title:    "Every part of a function signature",
		Anatomy: &AnatomySpec{
			Subject: "def greet(name, excited=False) -> str:",
			Parts: []AnatomyPart{
				{Text: "def", Label: "the keyword", Note: "Tells Python a function is being defined."},
				{Text: "greet", Label: "the name", Note: "What you call it later."},
				{Text: "name, excited=False", Label: "the parameters", Note: "One required, one optional."},
				{Text: "-> str", Label: "the return type", Note: "A hint for readers and tools."},
			},
		},
		Beats: []SnippetBeat{
			{ID: "whole", Heading: "One line, four jobs", Narration: anatNarration, Anatomy: &AnatomyBeat{Whole: true}},
			{ID: "kw", Heading: "The keyword", Narration: anatNarration, Anatomy: &AnatomyBeat{Part: 0}},
			{ID: "name", Heading: "The name", Narration: anatNarration, Anatomy: &AnatomyBeat{Part: 1}},
			{ID: "params", Heading: "The parameters", Narration: anatNarration, Anatomy: &AnatomyBeat{Part: 2}},
			{ID: "ret", Heading: "The return type", Narration: anatNarration, Anatomy: &AnatomyBeat{Part: 3}},
		},
	}
}

func TestAnatomyPlanAccepted(t *testing.T) {
	if err := validateAnatomyPlan(anatomyPlan()); err != nil {
		t.Fatalf("a well-formed anatomy was rejected: %v", err)
	}
}

// The rule the template rests on: a part quotes the subject, it does not
// describe it. Anything else is a callout landing on the wrong characters.
func TestAnatomyPartsMustQuoteTheSubject(t *testing.T) {
	p := anatomyPlan()
	p.Anatomy.Parts[1].Text = "the function name"
	err := validateAnatomyPlan(p)
	if err == nil {
		t.Fatal("a part quoting text absent from the subject was accepted")
	}
	if !strings.Contains(err.Error(), "does not appear") {
		t.Errorf("the error should say the text is not in the subject; got: %v", err)
	}
}

func TestAnatomyPartsMayNotOverlap(t *testing.T) {
	p := anatomyPlan()
	// "greet" is already claimed by part 1; "greet(name" would re-use it.
	p.Anatomy.Parts[2].Text = "greet(name"
	err := validateAnatomyPlan(p)
	if err == nil {
		t.Fatal("two parts covering the same characters were accepted")
	}
	if !strings.Contains(err.Error(), "overlap") {
		t.Errorf("the error should say the parts overlap; got: %v", err)
	}
}

// A repeated substring takes the first occurrence still free, which is what
// somebody reading left to right means by it.
func TestAnatomySpansTakeTheFirstFreeOccurrence(t *testing.T) {
	spans, err := resolveAnatomySpans("a.b.c", []AnatomyPart{
		{Text: ".", Label: "first dot"},
		{Text: ".", Label: "second dot"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if spans[0].Start != 1 || spans[1].Start != 3 {
		t.Errorf("spans = %v, want the two dots at 1 and 3", spans)
	}
}

// The viewer has to see the thing intact before any piece of it means anything.
func TestAnatomyMustOpenOnTheWhole(t *testing.T) {
	p := anatomyPlan()
	p.Beats[0].Anatomy = &AnatomyBeat{Part: 0}
	p.Beats[1].Anatomy = &AnatomyBeat{Whole: true}
	if err := validateAnatomyPlan(p); err == nil {
		t.Error("a clip opening on a single part was accepted")
	}

	p = anatomyPlan()
	for i := range p.Beats {
		if p.Beats[i].Anatomy.Whole {
			p.Beats[i].Anatomy = &AnatomyBeat{Part: 0}
		}
	}
	if err := validateAnatomyPlan(p); err == nil {
		t.Error("a clip that never shows the whole artefact was accepted")
	}
}

func TestAnatomyTakesPartsInReadingOrder(t *testing.T) {
	p := anatomyPlan()
	p.Beats[2].Anatomy = &AnatomyBeat{Part: 3}
	p.Beats[4].Anatomy = &AnatomyBeat{Part: 1}
	err := validateAnatomyPlan(p)
	if err == nil {
		t.Fatal("parts taken out of reading order were accepted")
	}
	if !strings.Contains(err.Error(), "order they are read") {
		t.Errorf("the error should explain the ordering; got: %v", err)
	}
}

func TestAnatomyRequiresEveryPartExplained(t *testing.T) {
	p := anatomyPlan()
	p.Beats = p.Beats[:4] // the return type never gets a beat
	err := validateAnatomyPlan(p)
	if err == nil {
		t.Fatal("a part with no beat was accepted")
	}
	if !strings.Contains(err.Error(), "never explained") {
		t.Errorf("the error should name the unexplained part; got: %v", err)
	}
}

func TestAnatomyRejectsAMultiLineOrOversizedSubject(t *testing.T) {
	p := anatomyPlan()
	p.Anatomy.Subject = "def f():\n    pass"
	if err := validateAnatomyPlan(p); err == nil {
		t.Error("a multi-line subject was accepted — callouts need columns to point at")
	}

	p = anatomyPlan()
	p.Anatomy.Subject = strings.Repeat("x", maxAnatomySubjectChars+10)
	if err := validateAnatomyPlan(p); err == nil {
		t.Error("a subject far over the character cap was accepted")
	}
}

// `part` omitted decodes to 0, which is a real index — the `whole` flag is what
// keeps an overview beat from silently lighting the first piece.
func TestNormalizeAnatomyTreatsNegativePartAsWhole(t *testing.T) {
	p := anatomyPlan()
	p.Beats[0].Anatomy = &AnatomyBeat{Part: -1}
	normalizeAnatomyPlan(p)
	if !p.Beats[0].Anatomy.Whole || p.Beats[0].Anatomy.Part != 0 {
		t.Errorf("a negative part index should become whole; got %#v", p.Beats[0].Anatomy)
	}
	if err := validateAnatomyPlan(p); err != nil {
		t.Errorf("the normalized plan should validate: %v", err)
	}
}

// The normalizer must never touch `text`: it is a quotation, and trimming or
// re-casing it would move the callout somewhere the model did not mean.
func TestNormalizeAnatomyLeavesQuotedTextAlone(t *testing.T) {
	p := anatomyPlan()
	p.Anatomy.Parts[2].Text = "name, excited=False"
	p.Anatomy.Parts[2].Label = "  the    parameter   list here now  "
	normalizeAnatomyPlan(p)
	if p.Anatomy.Parts[2].Text != "name, excited=False" {
		t.Errorf("quoted text was modified: %q", p.Anatomy.Parts[2].Text)
	}
	if n := len(strings.Fields(p.Anatomy.Parts[2].Label)); n > maxAnatomyLabelWords {
		t.Errorf("label was not clamped: %d words", n)
	}
}

func TestAnatomyScenesEmitsResolvedSpans(t *testing.T) {
	plan := anatomyPlan()
	scenes, err := anatomyScenes(sceneInput(t, plan, 5000))
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneAnatomy {
		t.Fatalf("want one anatomy scene, got %d", len(scenes))
	}
	parts, ok := scenes[0].Props["parts"].([]map[string]any)
	if !ok || len(parts) != 4 {
		t.Fatalf("want four resolved parts, got %#v", scenes[0].Props["parts"])
	}
	// The renderer positions callouts from these and searches for nothing, so
	// a wrong span is a callout pointing at the wrong word.
	if parts[0]["start"] != 0 || parts[0]["end"] != 3 {
		t.Errorf("part 0 span = %v..%v, want 0..3 for \"def\"", parts[0]["start"], parts[0]["end"])
	}
	subject := scenes[0].Props["subject"].(string)
	start := parts[2]["start"].(int)
	end := parts[2]["end"].(int)
	if got := string([]rune(subject)[start:end]); got != "name, excited=False" {
		t.Errorf("part 2 span covers %q, want the parameter list", got)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if len(steps) != len(plan.Beats) {
		t.Fatalf("want one step per beat, got %d", len(steps))
	}
	if steps[0]["whole"] != true {
		t.Error("the opening step should be marked whole")
	}
	if _, present := steps[0]["part"]; present {
		t.Error("a whole step carries a part index it does not mean")
	}
}
