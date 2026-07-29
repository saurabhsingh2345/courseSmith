package pipeline

import (
	"strings"
	"testing"
)

const clNarration = "Everything Redis holds lives in memory, which is the whole reason it answers so fast."

func constellationPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "constellation",
		Title:    "Everything that makes Redis Redis",
		Constellation: &ConstellationSpec{
			Centre:     "Redis",
			CentreIcon: "database",
			Spokes: []ConstellationSpoke{
				{Rel: "is", Label: "In-memory", Note: "Every value lives in RAM", Icon: "zap"},
				{Rel: "is", Label: "Single-threaded", Note: "One command at a time", Icon: "clock"},
				{Rel: "gives you", Label: "Data structures", Note: "Not just opaque blobs", Icon: "layers"},
				{Rel: "survives", Label: "Restarts", Note: "Snapshots and a log", Icon: "shield"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "redis", Heading: "One picture", Narration: clNarration, Constellation: &ConstellationBeat{Show: "centre"}},
			{ID: "memory", Heading: "In memory", Narration: clNarration, Constellation: &ConstellationBeat{Show: "spoke", At: 0}},
			{ID: "threaded", Heading: "One at a time", Narration: clNarration, Constellation: &ConstellationBeat{Show: "spoke", At: 1}},
			{ID: "structures", Heading: "Real structures", Narration: clNarration, Constellation: &ConstellationBeat{Show: "spoke", At: 2}},
			{ID: "durable", Heading: "It survives", Narration: clNarration, Constellation: &ConstellationBeat{Show: "spoke", At: 3}},
			{ID: "whole", Heading: "The whole picture", Narration: clNarration, Constellation: &ConstellationBeat{Show: "whole"}},
		},
	}
	p.targetWords = 6 * 40
	return p
}

func TestConstellationPlanAccepted(t *testing.T) {
	if err := validateConstellationPlan(constellationPlan()); err != nil {
		t.Fatalf("a well-formed constellation was rejected: %v", err)
	}
}

// The rule this template exists for: a spoke with no relation to the centre is
// a sub-topic, and a map of sub-topics is a table of contents.
func TestConstellationSpokesNeedARelation(t *testing.T) {
	p := constellationPlan()
	p.Constellation.Spokes[2].Rel = ""
	err := validateConstellationPlan(p)
	if err == nil {
		t.Fatal("a spoke with no relation to the centre was accepted")
	}
	if !strings.Contains(err.Error(), "table of contents") {
		t.Errorf("the error does not say what goes wrong: %v", err)
	}
}

func TestConstellationOpensOnTheCentre(t *testing.T) {
	p := constellationPlan()
	p.Beats[0].Constellation = &ConstellationBeat{Show: "spoke", At: 0}
	p.Beats[1].Constellation = &ConstellationBeat{Show: "centre"}
	err := validateConstellationPlan(p)
	if err == nil {
		t.Fatal("a clip that lit a property before naming the centre was accepted")
	}
	if !strings.Contains(err.Error(), "floating in space") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConstellationLightsEverySpoke(t *testing.T) {
	p := constellationPlan()
	// Add a fifth spoke and leave the beats alone: one goes unlit while every
	// other rule still passes.
	p.Constellation.Spokes = append(p.Constellation.Spokes, ConstellationSpoke{Rel: "speaks", Label: "One protocol"})
	err := validateConstellationPlan(p)
	if err == nil {
		t.Fatal("a spoke nobody narrates was accepted")
	}
	if !strings.Contains(err.Error(), "never spoken") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConstellationRejectsDuplicateLabels(t *testing.T) {
	p := constellationPlan()
	p.Constellation.Spokes[3].Label = p.Constellation.Spokes[0].Label
	if err := validateConstellationPlan(p); err == nil {
		t.Fatal("two spokes with the same label were accepted")
	}
}

func TestConstellationWholeIsLast(t *testing.T) {
	p := constellationPlan()
	p.Beats[1].Constellation = &ConstellationBeat{Show: "whole"}
	p.Beats[5].Constellation = &ConstellationBeat{Show: "spoke", At: 0}
	if err := validateConstellationPlan(p); err == nil {
		t.Fatal("a closing beat in the middle was accepted")
	}
}

func TestConstellationNormalizeRepairs(t *testing.T) {
	p := constellationPlan()
	p.Constellation.CentreIcon = "not-an-icon"
	p.Constellation.Spokes[0].Icon = "also-not-an-icon"
	p.Beats[2].Constellation.At = 99
	p.Beats[3].Constellation.Show = "nonsense"
	normalizeConstellationPlan(p)

	if p.Constellation.CentreIcon != "sparkles" {
		t.Errorf("an unknown centre icon became %q, want the fallback", p.Constellation.CentreIcon)
	}
	if p.Constellation.Spokes[0].Icon != "dot" {
		t.Errorf("an unknown spoke icon became %q, want the neutral dot", p.Constellation.Spokes[0].Icon)
	}
	if p.Beats[2].Constellation.At != len(p.Constellation.Spokes)-1 {
		t.Errorf("an out-of-range beat points at %d", p.Beats[2].Constellation.At)
	}
	if p.Beats[3].Constellation.Show != "spoke" {
		t.Errorf("an unknown show became %q, want spoke", p.Beats[3].Constellation.Show)
	}
}

// The layout is computed in Go so the same map is drawn every render, and four
// spokes land on the compass points rather than wherever a layout pass settles.
func TestConstellationAnglesAreDeterministic(t *testing.T) {
	p := constellationPlan()
	scenes, err := constellationScenes(sceneInput(t, p, 7000))
	if err != nil {
		t.Fatalf("constellationScenes: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneConstellation {
		t.Fatalf("want one constellation scene, got %d of %q", len(scenes), scenes[0].Type)
	}
	spokes := scenes[0].Props["spokes"].([]map[string]any)
	// Four spokes, starting at the top and going clockwise.
	for i, want := range []float64{-90, 0, 90, 180} {
		if spokes[i]["angle"] != want {
			t.Errorf("spoke %d sits at %v degrees, want %v", i, spokes[i]["angle"], want)
		}
	}
	// The relation word reaches the renderer: it rides on the connector and is
	// what makes the spoke a property rather than a neighbouring topic.
	if spokes[2]["rel"] != "gives you" {
		t.Errorf("rel = %v, want the relation to reach the scene", spokes[2]["rel"])
	}
}
