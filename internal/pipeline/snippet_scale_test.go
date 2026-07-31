package pipeline

import (
	"strings"
	"testing"
)

const scNarration = "A hundred bytes is about one sentence, which is the smallest thing on this whole ladder."

func scalePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "scale",
		Title:    "How much is a terabyte, really",
		Scale: &ScaleSpec{
			Unit: "MB",
			Levels: []ScaleLevel{
				{Label: "This sentence", Value: 0.0001, Display: "100 bytes", Icon: "file", Note: "A hundred characters"},
				{Label: "A phone photo", Value: 4, Display: "4 MB", Icon: "star", Note: "Forty thousand sentences"},
				{Label: "A feature film", Value: 4000, Display: "4 GB", Icon: "play", Note: "A thousand photos"},
				{Label: "A small library", Value: 4000000, Display: "4 TB", Icon: "city", Note: "A thousand films"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "sentence", Heading: "Start small", Narration: scNarration, Scale: &ScaleBeat{Show: "level", At: 0}},
			{ID: "photo", Heading: "One photo", Narration: scNarration, Scale: &ScaleBeat{Show: "level", At: 1}},
			{ID: "film", Heading: "A whole film", Narration: scNarration, Scale: &ScaleBeat{Show: "level", At: 2}},
			{ID: "library", Heading: "A library", Narration: scNarration, Scale: &ScaleBeat{Show: "level", At: 3}},
			{ID: "whole", Heading: "All of it", Narration: scNarration, Scale: &ScaleBeat{Show: "whole"}},
		},
	}
	p.targetWords = 5 * 40
	return p
}

func TestScalePlanAccepted(t *testing.T) {
	if err := validateScalePlan(scalePlan()); err != nil {
		t.Fatalf("a well-formed ladder was rejected: %v", err)
	}
}

// The rule this template exists for: a step under 4x is a box drawn inside
// another box that reads as a border, and gauge draws that range better.
func TestScaleRejectsStepsTooSmallToDraw(t *testing.T) {
	p := scalePlan()
	p.Scale.Levels[2].Value = 8 // twice the photo, not four thousand times
	err := validateScalePlan(p)
	if err == nil {
		t.Fatal("a two-times step was accepted, which nesting cannot show")
	}
	if !strings.Contains(err.Error(), "reads as a border") {
		t.Errorf("the error does not say why it cannot be drawn: %v", err)
	}
	// And it names the template that can.
	if !strings.Contains(err.Error(), "gauge") {
		t.Errorf("the error does not point at the template that handles this range: %v", err)
	}
}

func TestScaleRungsMustAscend(t *testing.T) {
	p := scalePlan()
	p.Scale.Levels[2].Value = 0.5
	err := validateScalePlan(p)
	if err == nil {
		t.Fatal("a ladder that went back down a rung was accepted")
	}
	if !strings.Contains(err.Error(), "only pulls back") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestScaleClimbsOneRungAtATime(t *testing.T) {
	p := scalePlan()
	p.Beats[1].Scale = &ScaleBeat{Show: "level", At: 2}
	p.Beats[2].Scale = &ScaleBeat{Show: "level", At: 1}
	err := validateScalePlan(p)
	if err == nil {
		t.Fatal("a clip that jumped a rung was accepted")
	}
	if !strings.Contains(err.Error(), "throws away the multiplier") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestScaleReachesEveryRung(t *testing.T) {
	p := scalePlan()
	// A fifth rung with the beats left alone: the camera never gets there.
	p.Scale.Levels = append(p.Scale.Levels, ScaleLevel{Label: "A data centre", Value: 4e10, Display: "40 PB"})
	err := validateScalePlan(p)
	if err == nil {
		t.Fatal("a rung the camera never stops at was accepted")
	}
	if !strings.Contains(err.Error(), "never gets to read") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestScaleLandsTheWholeLadderLast(t *testing.T) {
	p := scalePlan()
	p.Beats[4].Scale = &ScaleBeat{Show: "level", At: 3}
	p.Beats[3].Scale = &ScaleBeat{Show: "whole"}
	err := validateScalePlan(p)
	if err == nil {
		t.Fatal("a clip that carried on after showing the whole ladder was accepted")
	}
	if !strings.Contains(err.Error(), "nowhere further back") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestScaleNeedsAUnitAndLabels(t *testing.T) {
	p := scalePlan()
	p.Scale.Unit = ""
	if err := validateScalePlan(p); err == nil {
		t.Fatal("a ladder measured in nothing was accepted")
	}

	p = scalePlan()
	p.Scale.Levels[1].Label = ""
	err := validateScalePlan(p)
	if err == nil {
		t.Fatal("a rung that is a bare number was accepted")
	}
	if !strings.Contains(err.Error(), "nobody can picture") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNormalizeScaleWritesTheFigureNobodySupplied(t *testing.T) {
	p := scalePlan()
	p.Scale.Levels[2].Display = ""
	p.Scale.Levels[1].Value = -4 // a sign error, not a quantity
	p.Scale.Levels[0].Icon = "unicorn"
	normalizeScalePlan(p)

	if got := p.Scale.Levels[2].Display; got != "4K" {
		t.Errorf("a rung with no wording rendered as %q, want a compact figure rather than Go's float spelling", got)
	}
	if got := p.Scale.Levels[1].Value; got != 4 {
		t.Errorf("a negative size was not read as its magnitude: %v", got)
	}
	if got := p.Scale.Levels[0].Icon; got != "" {
		t.Errorf("an invented icon survived normalization: %q", got)
	}
}

func TestScaleTimesReadsAsSomebodyWouldSayIt(t *testing.T) {
	for _, c := range []struct {
		in   float64
		want string
	}{
		{10, "10"},
		{2.5, "2.5"},
		{1000, "1000"},
		// Past a hundred the decimal is noise, and 39.996 is the sort of figure
		// real division produces.
		{39.996, "40"},
	} {
		if got := scaleTimes(c.in); got != c.want {
			t.Errorf("scaleTimes(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestSpanWordsSaysTheBigNumberOutLoud(t *testing.T) {
	for _, c := range []struct {
		in   float64
		want string
	}{
		{4e10, "40 billion"},
		{1.2e6, "1.2 million"},
		{40000, "40 thousand"},
		{900, "900"},
	} {
		if got := spanWords(c.in); got != c.want {
			t.Errorf("spanWords(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestScaleScenesAreOneContinuousMove(t *testing.T) {
	p := scalePlan()
	scenes, err := scaleScenes(sceneInput(t, p, 10000))
	if err != nil {
		t.Fatalf("laying the ladder out: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("want one scene, got %d — the whole clip is one camera move, and a cut in the middle of it destroys the only thing it does", len(scenes))
	}
	props := scenes[0].Props
	if props["span"] != "40 billion" {
		t.Errorf("the end-to-end span is %v, want the figure the clip is actually about", props["span"])
	}
	levels, _ := props["levels"].([]map[string]any)
	if len(levels) != 4 {
		t.Fatalf("want four rungs, got %d", len(levels))
	}
	if _, set := levels[0]["times"]; set {
		t.Error("the first rung carries a multiplier; there is nothing below it to be a multiple of")
	}
	if got := levels[1]["times"]; got != "40000" {
		t.Errorf("the step to the second rung is %v, want the ratio the validator enforced", got)
	}
}
