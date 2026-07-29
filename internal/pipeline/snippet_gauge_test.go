package pipeline

import (
	"strings"
	"testing"
)

const ggNarration = "Twenty-four gigabytes sounds like a lot until you try to put a thirteen-billion model in it."

func gaugePlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "gauge",
		Title:    "Which models actually fit in 24GB",
		Gauge: &GaugeSpec{
			Unit: "GB",
			Ceiling: GaugeCeiling{
				Value: 24,
				Label: "What a 4090 holds",
				Note:  "And some of that is already spoken for",
			},
			Bars: []GaugeBar{
				{Label: "7B at 16-bit", Value: 14, Note: "Comfortable, with room to spare"},
				{Label: "13B at 16-bit", Value: 26, Note: "Two over, which is the same as not fitting"},
				{Label: "13B quantised", Value: 8, Note: "Same model, a third of the memory"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "card", Heading: "The card", Narration: ggNarration, Gauge: &GaugeBeat{Show: "ceiling"}},
			{ID: "small", Heading: "The small one", Narration: ggNarration, Gauge: &GaugeBeat{Show: "bar", At: 0}},
			{ID: "bigger", Heading: "One size up", Narration: ggNarration, Gauge: &GaugeBeat{Show: "bar", At: 1}},
			{ID: "quant", Heading: "The trick", Narration: ggNarration, Gauge: &GaugeBeat{Show: "bar", At: 2}},
			{ID: "answer", Heading: "What fits", Narration: ggNarration, Gauge: &GaugeBeat{Show: "verdict"}},
		},
	}
	p.targetWords = 5 * 40
	return p
}

func TestGaugePlanAccepted(t *testing.T) {
	if err := validateGaugePlan(gaugePlan()); err != nil {
		t.Fatalf("a well-formed gauge was rejected: %v", err)
	}
}

// The rule this template exists for: until the viewer knows what the line is, a
// bar filling toward it means nothing.
func TestGaugeOpensOnTheCeiling(t *testing.T) {
	p := gaugePlan()
	p.Beats[0].Gauge = &GaugeBeat{Show: "bar", At: 0}
	p.Beats[1].Gauge = &GaugeBeat{Show: "ceiling"}
	err := validateGaugePlan(p)
	if err == nil {
		t.Fatal("a clip that measured before setting the line was accepted")
	}
	if !strings.Contains(err.Error(), "Establish the line first") {
		t.Errorf("the error does not say why: %v", err)
	}
}

// The second rule: past the ratio cap the ceiling is a hairline and the picture
// conveys nothing the bare number would not convey better.
func TestGaugeRejectsOffScaleBar(t *testing.T) {
	p := gaugePlan()
	p.Gauge.Bars[1].Value = 24 * 40
	err := validateGaugePlan(p)
	if err == nil {
		t.Fatal("a bar forty times the ceiling was accepted — the line would be a hairline")
	}
	if !strings.Contains(err.Error(), "metric template") {
		t.Errorf("the error does not point at the alternative: %v", err)
	}
	// Exactly at the cap is still drawable.
	p = gaugePlan()
	p.Gauge.Bars[1].Value = 24 * maxGaugeRatio
	if err := validateGaugePlan(p); err != nil {
		t.Errorf("a bar exactly at the cap was rejected: %v", err)
	}
}

func TestGaugeCeilingNeedsALabel(t *testing.T) {
	p := gaugePlan()
	p.Gauge.Ceiling.Label = ""
	err := validateGaugePlan(p)
	if err == nil {
		t.Fatal("an unlabelled ceiling was accepted")
	}
	if !strings.Contains(err.Error(), "no meaning") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGaugeEveryBarIsRun(t *testing.T) {
	p := gaugePlan()
	p.Beats = append(p.Beats[:3], p.Beats[4:]...)
	err := validateGaugePlan(p)
	if err == nil {
		t.Fatal("a bar nobody narrates was accepted")
	}
	if !strings.Contains(err.Error(), "never run") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGaugeVerdictIsLast(t *testing.T) {
	p := gaugePlan()
	p.Beats[1].Gauge = &GaugeBeat{Show: "verdict"}
	p.Beats[4].Gauge = &GaugeBeat{Show: "bar", At: 0}
	if err := validateGaugePlan(p); err == nil {
		t.Fatal("a verdict in the middle was accepted")
	}
}

func TestGaugeRejectsDuplicateBarLabels(t *testing.T) {
	p := gaugePlan()
	p.Gauge.Bars[2].Label = p.Gauge.Bars[0].Label
	if err := validateGaugePlan(p); err == nil {
		t.Fatal("two bars with the same label were accepted")
	}
}

func TestGaugeNormalizeRepairs(t *testing.T) {
	p := gaugePlan()
	// A sign error is a magnitude, not a negative quantity.
	p.Gauge.Bars[0].Value = -14
	p.Gauge.Ceiling.Value = -24
	p.Beats[2].Gauge.At = 99
	p.Beats[3].Gauge.Show = "nonsense"
	normalizeGaugePlan(p)

	if p.Gauge.Bars[0].Value != 14 {
		t.Errorf("a negative bar became %v, want its magnitude", p.Gauge.Bars[0].Value)
	}
	if p.Gauge.Ceiling.Value != 24 {
		t.Errorf("a negative ceiling became %v, want its magnitude", p.Gauge.Ceiling.Value)
	}
	if p.Beats[2].Gauge.At != len(p.Gauge.Bars)-1 {
		t.Errorf("an out-of-range beat points at %d", p.Beats[2].Gauge.At)
	}
	if p.Beats[3].Gauge.Show != "bar" {
		t.Errorf("an unknown show became %q, want bar", p.Beats[3].Gauge.Show)
	}
}

// A zero-valued bar is dropped rather than drawn as nothing.
func TestGaugeNormalizeDropsEmptyBars(t *testing.T) {
	p := gaugePlan()
	p.Gauge.Bars = append(p.Gauge.Bars, GaugeBar{Label: "Nothing", Value: 0})
	normalizeGaugePlan(p)
	for _, b := range p.Gauge.Bars {
		if b.Value <= 0 {
			t.Error("a zero-valued bar survived normalize")
		}
	}
}

func TestGaugeFitsIsInclusive(t *testing.T) {
	// A model that needs exactly the memory available does fit. The nuance
	// belongs in the narration, not in a rounding rule.
	if !(GaugeBar{Value: 24}).Fits(24) {
		t.Error("a bar exactly at the ceiling should fit")
	}
	if (GaugeBar{Value: 24.1}).Fits(24) {
		t.Error("a bar over the ceiling should not fit")
	}
}

func TestGaugeScenesShape(t *testing.T) {
	p := gaugePlan()
	scenes, err := gaugeScenes(sceneInput(t, p, 8000))
	if err != nil {
		t.Fatalf("gaugeScenes: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneGauge {
		t.Fatalf("want one gauge scene, got %d of %q", len(scenes), scenes[0].Type)
	}
	bars, ok := scenes[0].Props["bars"].([]map[string]any)
	if !ok || len(bars) != 3 {
		t.Fatalf("want three bars on the scene, got %v", scenes[0].Props["bars"])
	}
	// `fits` is decided in Go so the bar's colour and the validator's words can
	// never disagree about what fitting means.
	if bars[0]["fits"] != true || bars[1]["fits"] != false {
		t.Errorf("fits was computed wrong: %v, %v", bars[0]["fits"], bars[1]["fits"])
	}
	// Everything shares one scale, sized to the longest bar with air past it,
	// so an overrunning bar visibly overruns instead of hitting the frame edge.
	ceiling := scenes[0].Props["ceiling"].(map[string]any)
	if cf := ceiling["frac"].(float64); cf <= 0 || cf >= 1 {
		t.Errorf("the ceiling sits at %v of the track, want strictly inside it", cf)
	}
	if bf := bars[1]["frac"].(float64); bf <= ceiling["frac"].(float64) {
		t.Error("the overrunning bar does not reach past the ceiling on the track")
	}
	if bf := bars[1]["frac"].(float64); bf > 1 {
		t.Errorf("the longest bar runs off the track at %v", bf)
	}
}
