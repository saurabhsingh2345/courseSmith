package pipeline

import (
	"strings"
	"testing"
)

const lyNarration = "Every level below this one holds something the level above it cannot touch directly."

func layersPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "layers",
		Title:    "The line your program cannot cross",
		Layers: &LayersSpec{
			Strata: []LayersStratum{
				{Label: "your program", Holds: "variables, loops, library calls"},
				{Label: "the C library", Holds: "printf, malloc, the syscall wrappers"},
				{Label: "kernel", Holds: "schedulers, drivers, the page tables"},
				{Label: "hardware", Holds: "registers, memory, the disc itself"},
			},
			Boundary:      1,
			BoundaryLabel: "the syscall line",
		},
		Beats: []SnippetBeat{
			{ID: "stack", Heading: "Four levels", Narration: lyNarration, Layers: &LayersBeat{Show: "stack"}},
			{ID: "top", Heading: "Your code", Narration: lyNarration, Layers: &LayersBeat{Show: "stratum", At: 0}},
			{ID: "libc", Heading: "The library", Narration: lyNarration, Layers: &LayersBeat{Show: "stratum", At: 1}},
			{ID: "kernel", Heading: "The kernel", Narration: lyNarration, Layers: &LayersBeat{Show: "stratum", At: 2}},
			{ID: "metal", Heading: "The metal", Narration: lyNarration, Layers: &LayersBeat{Show: "stratum", At: 3}},
			{ID: "cross", Heading: "Crossing over", Narration: lyNarration, Layers: &LayersBeat{Show: "cross"}},
			{ID: "whole", Heading: "The whole stack", Narration: lyNarration, Layers: &LayersBeat{Show: "whole"}},
		},
	}
	// The template's ideal is 28 words per beat, so the fixture budget is sized
	// against that — nBeats * 40 would demand more beats than it has.
	p.targetWords = 7 * 28
	return p
}

func TestLayersPlanAccepted(t *testing.T) {
	if err := validateLayersPlan(layersPlan()); err != nil {
		t.Fatalf("a well-formed stack was rejected: %v", err)
	}
}

func TestLayersRejectsTooFewStrata(t *testing.T) {
	p := layersPlan()
	p.Layers.Strata = p.Layers.Strata[:2]
	err := validateLayersPlan(p)
	if err == nil {
		t.Fatal("a two-band stack was accepted, and two bands are an over-and-under that needs no diagram")
	}
	if !strings.Contains(err.Error(), "2 strata") {
		t.Fatalf("the error does not quote the count: %v", err)
	}
}

func TestLayersRejectsTooManyStrata(t *testing.T) {
	p := layersPlan()
	for i := 0; i < 5; i++ {
		p.Layers.Strata = append(p.Layers.Strata, LayersStratum{Label: "extra", Holds: "something else"})
	}
	if err := validateLayersPlan(p); err == nil {
		t.Fatal("a nine-band stack was accepted, and the bars would lose the height their labels need")
	}
}

func TestLayersRejectsABandThatSaysNothing(t *testing.T) {
	p := layersPlan()
	p.Layers.Strata[2].Holds = ""
	err := validateLayersPlan(p)
	if err == nil {
		t.Fatal("a named empty bar was accepted")
	}
	if !strings.Contains(err.Error(), "kernel") {
		t.Fatalf("the error does not name the band: %v", err)
	}
}

// The rule is drawn UNDER the stratum it names, so the bottom band cannot
// carry one — there would be nothing beneath the line.
func TestLayersRejectsABoundaryOnTheBottomBand(t *testing.T) {
	p := layersPlan()
	p.Layers.Boundary = 3
	err := validateLayersPlan(p)
	if err == nil {
		t.Fatal("a boundary under the bottom band was accepted")
	}
	if !strings.Contains(err.Error(), "0 to 2") {
		t.Fatalf("the error does not quote the usable range: %v", err)
	}
}

func TestLayersRejectsAnUnnamedBoundary(t *testing.T) {
	p := layersPlan()
	p.Layers.BoundaryLabel = ""
	err := validateLayersPlan(p)
	if err == nil {
		t.Fatal("an unnamed brighter rule was accepted, and it reads as a styling accident")
	}
	if !strings.Contains(err.Error(), "the C library") || !strings.Contains(err.Error(), "kernel") {
		t.Fatalf("the error does not name the two bands the line runs between: %v", err)
	}
}

func TestLayersAcceptsAStackWithNoBoundary(t *testing.T) {
	p := layersPlan()
	p.Layers.Boundary = noLayersBoundary
	p.Layers.BoundaryLabel = ""
	// Some stacks genuinely have no special line, and forcing one is the worse
	// error, so the crossing beat goes rather than the boundary appearing.
	p.Beats = append(p.Beats[:5], p.Beats[6:]...)
	p.targetWords = len(p.Beats) * 28
	if err := validateLayersPlan(p); err != nil {
		t.Fatalf("a stack with no special line was rejected: %v", err)
	}
}

// The one rule this template exists for: told to invent a boundary, a model
// picks one at random and ships a picture claiming a privilege line where none
// exists. So the rejection says to drop the beat, not to set a boundary.
func TestLayersRejectsACrossingWithNoBoundary(t *testing.T) {
	p := layersPlan()
	p.Layers.Boundary = noLayersBoundary
	p.Layers.BoundaryLabel = ""
	err := validateLayersPlan(p)
	if err == nil {
		t.Fatal("a payload crossing a line that does not exist was accepted")
	}
	if !strings.Contains(err.Error(), "invent") {
		t.Fatalf("the error tells the model to set a boundary instead of dropping the beat: %v", err)
	}
	if !strings.Contains(err.Error(), "Drop the \"cross\" beat") {
		t.Fatalf("the error does not give the repair: %v", err)
	}
}

func TestLayersRejectsABandFocusedTwice(t *testing.T) {
	p := layersPlan()
	p.Beats[4].Layers = &LayersBeat{Show: "stratum", At: 1}
	err := validateLayersPlan(p)
	if err == nil {
		t.Fatal("a band focused twice was accepted")
	}
	if !strings.Contains(err.Error(), "the C library") {
		t.Fatalf("the error does not name the repeated band: %v", err)
	}
}

func TestLayersRequiresOpeningOnTheStack(t *testing.T) {
	p := layersPlan()
	p.Beats[0].Layers = &LayersBeat{Show: "stratum", At: 0}
	p.Beats[1].Layers = &LayersBeat{Show: "stack"}
	if err := validateLayersPlan(p); err == nil {
		t.Fatal("a band coming forward before the stack was drawn was accepted")
	}
}

func TestLayersRequiresClosingOnTheWholeStack(t *testing.T) {
	p := layersPlan()
	p.Beats[len(p.Beats)-1].Layers = &LayersBeat{Show: "stratum", At: 0}
	if err := validateLayersPlan(p); err == nil {
		t.Fatal("a clip that never lights the whole stack was accepted")
	}
}

// An index that cannot name a line becomes "no line" rather than a guess, for
// the same reason the cross rule exists.
func TestLayersNormalizeDropsAnUnusableBoundary(t *testing.T) {
	p := layersPlan()
	p.Layers.Boundary = 6
	normalizeLayersPlan(p)
	if p.Layers.Boundary != noLayersBoundary {
		t.Fatalf("an out-of-range boundary normalized to %d, want %d", p.Layers.Boundary, noLayersBoundary)
	}
}

func TestLayersNormalizeClampsLabelsAndDropsNamelessBands(t *testing.T) {
	p := layersPlan()
	p.Layers.Strata = append(p.Layers.Strata, LayersStratum{Label: "  ", Holds: "nothing"})
	p.Layers.Strata[0].Holds = "  variables   and loops and library calls and everything else your code touches "
	p.Beats[4].Layers.At = 11
	normalizeLayersPlan(p)

	if len(p.Layers.Strata) != 4 {
		t.Fatalf("the nameless band survived normalize: %v", p.Layers.Strata)
	}
	if got := len(strings.Fields(p.Layers.Strata[0].Holds)); got > maxLayersHoldsWords {
		t.Fatalf("holds normalized to %d words, want at most %d", got, maxLayersHoldsWords)
	}
	if p.Beats[4].Layers.At != 3 {
		t.Fatalf("an index past the end clamped to %d, want 3", p.Beats[4].Layers.At)
	}
	if err := validateLayersPlan(p); err != nil {
		t.Fatalf("a repairable stack was rejected after normalize: %v", err)
	}
}

func TestLayersShowDefaultsToStratum(t *testing.T) {
	b := LayersBeat{Show: "sediment"}
	if got := b.ResolvedShow(); got != "stratum" {
		t.Fatalf("an unknown show resolved to %q, want stratum", got)
	}
}

// Which side of the line each band sits on, and whether the crossing has
// happened, arrive precomputed.
func TestLayersScenesSideTheBandsAndAccumulate(t *testing.T) {
	p := layersPlan()
	scenes, err := layersScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	if props["boundary"] != 1 || props["boundaryLabel"] != "the syscall line" {
		t.Fatalf("the boundary props are wrong: %v %v", props["boundary"], props["boundaryLabel"])
	}
	strata, _ := props["strata"].([]map[string]any)
	if len(strata) != 4 {
		t.Fatalf("want 4 bands, got %d", len(strata))
	}
	if strata[0]["above"] != true || strata[1]["above"] != true {
		t.Fatalf("the bands over the line are not marked as above: %v", strata)
	}
	if strata[2]["above"] != false || strata[3]["above"] != false {
		t.Fatalf("the bands under the line are not marked as below: %v", strata)
	}

	steps, _ := props["steps"].([]map[string]any)
	if steps[0]["show"] != "stack" {
		t.Fatalf("first step shows %v, want stack", steps[0]["show"])
	}
	if lit, _ := steps[0]["lit"].([]int); len(lit) != 0 {
		t.Fatalf("the opener already has focused bands: %v", lit)
	}
	if steps[0]["crossed"] != false {
		t.Fatal("the opener claims the crossing has already happened")
	}

	last := steps[len(steps)-1]
	if last["show"] != "whole" {
		t.Fatalf("last step shows %v, want whole", last["show"])
	}
	lit, _ := last["lit"].([]int)
	if len(lit) != 4 || lit[0] != 0 || lit[3] != 3 {
		t.Fatalf("the closer has not accumulated every band: %v", lit)
	}
	if last["crossed"] != true {
		t.Fatal("the closer does not remember the crossing")
	}
}
