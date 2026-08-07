package pipeline

import (
	"strings"
	"testing"
)

const gwNarration = "On ten items nobody could tell these two apart, and on a million only one of them finishes."

func growthPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "growth",
		Title:    "Why the nested loop dies",
		Growth: &GrowthSpec{
			Curves: []GrowthCurve{
				{Class: "n", Label: "the single pass"},
				{Class: "n2", Label: "the nested loop"},
			},
			Probe: 1000000,
		},
		Beats: []SnippetBeat{
			{ID: "the-axes", Heading: "The chart", Narration: gwNarration, Growth: &GrowthBeat{Show: "axes"}},
			{ID: "linear", Heading: "One pass", Narration: gwNarration, Growth: &GrowthBeat{Show: "curve", At: 0}},
			{ID: "quadratic", Heading: "A loop inside a loop", Narration: gwNarration, Growth: &GrowthBeat{Show: "curve", At: 1}},
			{ID: "the-probe", Heading: "At a million", Narration: gwNarration, Growth: &GrowthBeat{Show: "probe"}},
			{ID: "the-moral", Heading: "The moral", Narration: gwNarration, Growth: &GrowthBeat{Show: "moral"}},
		},
	}
	// Against this template's own ideal of 28 words per beat rather than the
	// shared 40: at 40 the shared bounds demand more beats than the fixture
	// has, and it would be rejected for length before any rule under test ran.
	p.targetWords = 5 * 28
	return p
}

func TestGrowthPlanAccepted(t *testing.T) {
	if err := validateGrowthPlan(growthPlan()); err != nil {
		t.Fatalf("a well-formed growth plan was rejected: %v", err)
	}
}

// The family's signature rule here: the hierarchy is known in Go, and a legend
// written out of order is rejected with both offending classes named.
func TestGrowthRejectsClassesOutOfGrowthOrder(t *testing.T) {
	p := growthPlan()
	p.Growth.Curves[0].Class = "n2"
	p.Growth.Curves[1].Class = "n"
	err := validateGrowthPlan(p)
	if err == nil {
		t.Fatal("a legend that runs fast to slow was accepted")
	}
	if !strings.Contains(err.Error(), "O(n²)") || !strings.Contains(err.Error(), "O(n)") {
		t.Fatalf("the error does not name both classes that are out of order: %v", err)
	}
}

func TestGrowthRejectsTheSameClassTwice(t *testing.T) {
	p := growthPlan()
	p.Growth.Curves[1].Class = "n"
	err := validateGrowthPlan(p)
	if err == nil {
		t.Fatal("one line drawn twice was accepted")
	}
	if !strings.Contains(err.Error(), "curves 0 and 1") {
		t.Fatalf("the error does not point at both curves: %v", err)
	}
}

func TestGrowthRejectsASingleCurve(t *testing.T) {
	p := growthPlan()
	p.Growth.Curves = p.Growth.Curves[:1]
	p.Beats = append(p.Beats[:2], p.Beats[3:]...)
	p.targetWords = 4 * 28
	err := validateGrowthPlan(p)
	if err == nil {
		t.Fatal("a chart with one curve was accepted, and one line has no story")
	}
	if !strings.Contains(err.Error(), "1 curves") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestGrowthRejectsACurveWithNoLabel(t *testing.T) {
	p := growthPlan()
	p.Growth.Curves[1].Label = ""
	err := validateGrowthPlan(p)
	if err == nil {
		t.Fatal("a curve with only its notation was accepted")
	}
	if !strings.Contains(err.Error(), "O(n²)") {
		t.Fatalf("the error does not name the unlabelled curve: %v", err)
	}
}

func TestGrowthRejectsAProbeBeforeEveryCurveIsDrawn(t *testing.T) {
	p := growthPlan()
	p.Beats[2].Growth = &GrowthBeat{Show: "probe"}
	p.Beats[3].Growth = &GrowthBeat{Show: "curve", At: 1}
	err := validateGrowthPlan(p)
	if err == nil {
		t.Fatal("a probe that read a chart with a curve missing was accepted")
	}
	if !strings.Contains(err.Error(), "the nested loop") {
		t.Fatalf("the error does not name the curve that gets no reading: %v", err)
	}
}

func TestGrowthRejectsAProbeWithNoInputSize(t *testing.T) {
	p := growthPlan()
	p.Growth.Probe = 0
	err := validateGrowthPlan(p)
	if err == nil {
		t.Fatal("a drop-line at no particular n was accepted")
	}
	if !strings.Contains(err.Error(), "probe is 0") {
		t.Fatalf("the error does not quote the value: %v", err)
	}
}

func TestGrowthRejectsDrawingACurveTwice(t *testing.T) {
	p := growthPlan()
	p.Beats[2].Growth = &GrowthBeat{Show: "curve", At: 0}
	if err := validateGrowthPlan(p); err == nil {
		t.Fatal("a curve redrawn on top of itself was accepted")
	}
}

func TestGrowthRejectsACurveThatNeverAppears(t *testing.T) {
	p := growthPlan()
	p.Growth.Curves = append(p.Growth.Curves, GrowthCurve{Class: "2n", Label: "the brute force"})
	p.Beats[3].Growth = &GrowthBeat{Show: "curve", At: 2}
	p.Growth.Probe = 0
	if err := validateGrowthPlan(p); err != nil {
		t.Fatalf("a three-curve chart with no probe was rejected: %v", err)
	}
	// Take the beat that drew the third curve away and leave the curve in the
	// plan.
	p.Beats = append(p.Beats[:3], p.Beats[4])
	p.targetWords = 4 * 28
	err := validateGrowthPlan(p)
	if err == nil {
		t.Fatal("a chart with a curve in the plan and never on screen was accepted")
	}
	if !strings.Contains(err.Error(), "the brute force") {
		t.Fatalf("the error does not name the missing curve: %v", err)
	}
}

func TestGrowthRequiresOpeningOnTheAxes(t *testing.T) {
	p := growthPlan()
	p.Beats[0].Growth = &GrowthBeat{Show: "curve", At: 0}
	p.Beats[1].Growth = &GrowthBeat{Show: "axes"}
	if err := validateGrowthPlan(p); err == nil {
		t.Fatal("a curve arriving on axes nobody has read was accepted")
	}
}

func TestGrowthRequiresClosingOnTheMoral(t *testing.T) {
	p := growthPlan()
	p.Beats[4].Growth = &GrowthBeat{Show: "probe"}
	p.Beats[3].Growth = &GrowthBeat{Show: "moral"}
	if err := validateGrowthPlan(p); err == nil {
		t.Fatal("a chart that ends without its verdict was accepted")
	}
}

func TestGrowthNormalizeCoercesNotationIntoTheVocabulary(t *testing.T) {
	p := growthPlan()
	p.Growth.Curves[0].Class = "O(log n)"
	p.Growth.Curves[1].Class = "quadratic"
	p.Growth.Curves[1].Label = "the nested loop over every single pair"
	p.Growth.Probe = maxGrowthProbe * 4
	p.Beats[2].Growth.At = 17
	normalizeGrowthPlan(p)
	if got := p.Growth.Curves[0].ResolvedClass(); got != "logn" {
		t.Fatalf("O(log n) normalized to %q, want logn", got)
	}
	if got := p.Growth.Curves[1].ResolvedClass(); got != "n2" {
		t.Fatalf("quadratic normalized to %q, want n2", got)
	}
	if got := len(strings.Fields(p.Growth.Curves[1].Label)); got != maxGrowthLabelWords {
		t.Fatalf("the label is %d words, want it clamped to %d", got, maxGrowthLabelWords)
	}
	if p.Growth.Probe != maxGrowthProbe {
		t.Fatalf("the probe normalized to %d, want it capped at %d", p.Growth.Probe, maxGrowthProbe)
	}
	if p.Beats[2].Growth.At != 1 {
		t.Fatalf("an index off the end normalized to %d, want the last curve 1", p.Beats[2].Growth.At)
	}
	if err := validateGrowthPlan(p); err != nil {
		t.Fatalf("a loosely-spelled but sound plan was rejected after normalize: %v", err)
	}
}

func TestGrowthClassDefaultsToLinear(t *testing.T) {
	c := GrowthCurve{Class: "polynomialish"}
	if got := c.ResolvedClass(); got != "n" {
		t.Fatalf("an unknown class resolved to %q, want n", got)
	}
}

func TestGrowthShowDefaultsToCurve(t *testing.T) {
	b := GrowthBeat{Show: "zoom"}
	if got := b.ResolvedShow(); got != "curve" {
		t.Fatalf("an unknown show resolved to %q, want curve", got)
	}
}

// The readings are computed and formatted in Go, including the ones no float64
// can hold.
func TestGrowthReadingFormatsTheDamage(t *testing.T) {
	cases := []struct {
		class string
		n     int
		want  string
	}{
		{"1", 1000000, "1"},
		{"logn", 1000000, "20"},
		{"n", 1000000, "1,000,000"},
		{"n2", 1000000, "1,000,000,000,000"},
	}
	for _, c := range cases {
		if got := growthReading(c.class, c.n); got != c.want {
			t.Errorf("growthReading(%q, %d) = %q, want %q", c.class, c.n, got, c.want)
		}
	}
	// 2^1000000 overflows a float64 entirely, so the exponent has to be
	// computed rather than measured.
	exp := growthReading("2n", 1000000)
	if !strings.Contains(exp, "e+301029") {
		t.Fatalf("growthReading(2n, 1000000) = %q, want scientific notation around 10^301030", exp)
	}
}

// The component draws polylines and prints strings. Everything numeric is
// resolved here.
func TestGrowthScenesPrecomputeTheGeometry(t *testing.T) {
	p := growthPlan()
	scenes, err := growthScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("want one scene spanning the clip, got %d", len(scenes))
	}
	props := scenes[0].Props

	curves, _ := props["curves"].([]map[string]any)
	if len(curves) != 2 {
		t.Fatalf("want 2 curves, got %d", len(curves))
	}
	if curves[1]["notation"] != "O(n²)" {
		t.Fatalf("the second curve is set as %v, want O(n²)", curves[1]["notation"])
	}
	linear, _ := curves[0]["points"].([]float64)
	if len(linear) != growthSamples {
		t.Fatalf("the linear curve has %d samples, want %d", len(linear), growthSamples)
	}
	if linear[0] >= linear[len(linear)-1] || linear[len(linear)-1] > 1 {
		t.Fatalf("the linear curve runs %v to %v, want a rise that stays inside the frame", linear[0], linear[len(linear)-1])
	}
	quad, _ := curves[1]["points"].([]float64)
	if quad[len(quad)-1] != 1 {
		t.Fatalf("the quadratic curve ends at %v, want it clamped at the frame top", quad[len(quad)-1])
	}
	if curves[0]["reading"] != "1,000,000" || curves[1]["reading"] != "1,000,000,000,000" {
		t.Fatalf("the probe readings are %v and %v", curves[0]["reading"], curves[1]["reading"])
	}
	if props["probeLabel"] != "1,000,000" || props["worst"] != 1 {
		t.Fatalf("the probe chrome is wrong: %v / %v", props["probeLabel"], props["worst"])
	}

	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != 5 {
		t.Fatalf("want 5 steps, got %d", len(steps))
	}
	if steps[0]["show"] != "axes" {
		t.Fatalf("first step shows %v, want axes", steps[0]["show"])
	}
	if first, _ := steps[0]["drawn"].([]int); len(first) != 0 {
		t.Fatalf("the opener has drawn %v, want an empty chart", first)
	}
	last := steps[len(steps)-1]
	if last["show"] != "moral" {
		t.Fatalf("last step shows %v, want moral", last["show"])
	}
	drawn, _ := last["drawn"].([]int)
	if len(drawn) != 2 || drawn[0] != 0 || drawn[1] != 1 {
		t.Fatalf("the closer has drawn %v, want both curves", drawn)
	}
}
