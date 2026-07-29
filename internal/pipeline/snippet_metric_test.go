package pipeline

import (
	"strings"
	"testing"
)

const mtNarration = "Seventy billion parameters, and every one of them has to be sitting in memory."

func metricPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "metric",
		Title:    "What a 70B model actually costs to run",
		Metric: &MetricSpec{
			Figures: []Metric{
				{Value: "70", Unit: "B params", Label: "The model you want", Note: "Every parameter has to be in memory", Role: "quantity"},
				{Value: "140", Unit: "GB", Label: "Memory it needs", Note: "Two bytes a parameter", Role: "quantity"},
				{Value: "24", Unit: "GB", Label: "What a 4090 has", Note: "A sixth of what you need", Role: "limit"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "model", Heading: "The model", Narration: mtNarration, Metric: &MetricBeat{Show: "state", At: 0}},
			{ID: "memory", Heading: "What it needs", Narration: mtNarration, Metric: &MetricBeat{Show: "state", At: 1}},
			{ID: "card", Heading: "What you have", Narration: mtNarration, Metric: &MetricBeat{Show: "state", At: 2}},
			{ID: "recap", Heading: "All three", Narration: mtNarration, Metric: &MetricBeat{Show: "recap"}},
		},
	}
	p.targetWords = 4 * 40
	return p
}

func TestMetricPlanAccepted(t *testing.T) {
	if err := validateMetricPlan(metricPlan()); err != nil {
		t.Fatalf("a well-formed metric plan was rejected: %v", err)
	}
}

// The rule the template exists for. A number with no unit is not a
// measurement, and one with no label is trivia.
func TestMetricRequiresUnitAndLabel(t *testing.T) {
	p := metricPlan()
	p.Metric.Figures[1].Unit = ""
	err := validateMetricPlan(p)
	if err == nil {
		t.Fatal("a plain figure with no unit was accepted — 140 of what?")
	}
	if !strings.Contains(err.Error(), "unit") {
		t.Errorf("the error does not mention the unit: %v", err)
	}

	p = metricPlan()
	p.Metric.Figures[0].Label = ""
	if err := validateMetricPlan(p); err == nil {
		t.Fatal("a figure with no label was accepted — a number nobody labels is trivia")
	}
}

// A value with no single number to run to — a range, a threshold — is exempt
// from the unit rule, because "<1" and "313K–577K" carry their own reading.
func TestMetricNonNumericValueNeedsNoUnit(t *testing.T) {
	p := metricPlan()
	p.Metric.Figures[2] = Metric{Value: "313K–577K", Label: "Annual rental", Role: "limit"}
	if err := validateMetricPlan(p); err != nil {
		t.Fatalf("a range with no unit was rejected: %v", err)
	}
	if p.Metric.Figures[2].countsUp() {
		t.Error("a range should not count up — animating it would state a false number on the way")
	}
	if !(Metric{Value: "2.8"}).countsUp() {
		t.Error("a plain decimal should count up")
	}
}

// The second rule: a clip where every figure is context has no point of view.
func TestMetricRejectsAllNeutral(t *testing.T) {
	p := metricPlan()
	for i := range p.Metric.Figures {
		p.Metric.Figures[i].Role = "neutral"
	}
	err := validateMetricPlan(p)
	if err == nil {
		t.Fatal("a clip with no quantity and no limit was accepted — that is a list of facts, not an argument")
	}
	if !strings.Contains(err.Error(), "neutral") {
		t.Errorf("the error does not name the problem: %v", err)
	}
}

func TestMetricEveryFigureIsNarrated(t *testing.T) {
	p := metricPlan()
	// Drop the beat that states the third figure.
	p.Beats = append(p.Beats[:2], p.Beats[3:]...)
	err := validateMetricPlan(p)
	if err == nil {
		t.Fatal("a figure nobody states was accepted")
	}
	if !strings.Contains(err.Error(), "never said out loud") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMetricRejectsRepeatedFigure(t *testing.T) {
	p := metricPlan()
	p.Beats[1].Metric.At = 0
	if err := validateMetricPlan(p); err == nil {
		t.Fatal("the same figure stated twice was accepted")
	}
}

func TestMetricRecapIsLast(t *testing.T) {
	p := metricPlan()
	p.Beats[0].Metric = &MetricBeat{Show: "recap"}
	p.Beats[3].Metric = &MetricBeat{Show: "state", At: 0}
	err := validateMetricPlan(p)
	if err == nil {
		t.Fatal("a recap in the middle of the clip was accepted")
	}
	if !strings.Contains(err.Error(), "closing frame") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestMetricRejectsDuplicateValues(t *testing.T) {
	p := metricPlan()
	p.Metric.Figures[1].Value = "70"
	p.Metric.Figures[1].Unit = "B params"
	if err := validateMetricPlan(p); err == nil {
		t.Fatal("two identical figures were accepted")
	}
}

// Normalize repairs mechanical mistakes rather than rejecting them: an
// over-long value, an invented role, a beat pointing past the end.
func TestMetricNormalizeRepairs(t *testing.T) {
	p := metricPlan()
	p.Metric.Figures[0].Value = "1234567890123"
	p.Metric.Figures[0].Role = "invented"
	p.Beats[0].Metric.At = 99
	p.Beats[1].Metric.Show = "nonsense"
	normalizeMetricPlan(p)

	if n := len([]rune(p.Metric.Figures[0].Value)); n > maxMetricValueChars {
		t.Errorf("value is %d chars after normalize, want <= %d", n, maxMetricValueChars)
	}
	if p.Metric.Figures[0].Role != "neutral" {
		t.Errorf("an invented role became %q, want neutral — a figure whose job nobody stated should not be shouting", p.Metric.Figures[0].Role)
	}
	if p.Beats[0].Metric.At != len(p.Metric.Figures)-1 {
		t.Errorf("an out-of-range beat points at %d, want it clamped to %d", p.Beats[0].Metric.At, len(p.Metric.Figures)-1)
	}
	if p.Beats[1].Metric.Show != "state" {
		t.Errorf("an unknown show became %q, want state", p.Beats[1].Metric.Show)
	}
}

// A figure with no value at all is dropped rather than invented.
func TestMetricNormalizeDropsEmptyFigures(t *testing.T) {
	p := metricPlan()
	p.Metric.Figures = append(p.Metric.Figures, Metric{Label: "Nothing", Role: "limit"})
	normalizeMetricPlan(p)
	for _, f := range p.Metric.Figures {
		if strings.TrimSpace(f.Value) == "" {
			t.Error("a figure with no number survived normalize")
		}
	}
}

func TestMetricScenesShape(t *testing.T) {
	p := metricPlan()
	scenes, err := metricScenes(sceneInput(t, p, 9000))
	if err != nil {
		t.Fatalf("metricScenes: %v", err)
	}
	if len(scenes) != 1 {
		t.Fatalf("got %d scenes, want 1 — the figures share one frame", len(scenes))
	}
	if scenes[0].Type != SceneMetric {
		t.Errorf("scene type = %q, want %q", scenes[0].Type, SceneMetric)
	}
	steps, ok := scenes[0].Props["steps"].([]map[string]any)
	if !ok || len(steps) != len(p.Beats) {
		t.Fatalf("want one step per beat, got %v", scenes[0].Props["steps"])
	}
	figures, ok := scenes[0].Props["figures"].([]map[string]any)
	if !ok || len(figures) != len(p.Metric.Figures) {
		t.Fatalf("want every figure on the scene, got %v", scenes[0].Props["figures"])
	}
	// countsUp is decided in Go so the renderer never guesses whether a value
	// is a number.
	if figures[0]["countsUp"] != true {
		t.Error("a plain figure is not marked countsUp")
	}
}
