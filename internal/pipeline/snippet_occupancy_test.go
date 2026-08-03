package pipeline

import (
	"strings"
	"testing"
)

const ocNarration = "Eight hundred and ninety-six experts, and sixteen of them do the work."

func occupancyPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "occupancy",
		Title:    "16 of 896 experts do the work",
		Occupancy: &OccupancySpec{
			Total: 896,
			Unit:  "expert",
			Label: "the model's experts",
			Bands: []OccupancyBand{
				{Count: 16, Label: "Active this token", Note: "The rest sit idle holding memory", Role: "quantity"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "all", Heading: "The whole model", Narration: ocNarration, Occupancy: &OccupancyBeat{Show: "grid"}},
			{ID: "active", Heading: "What runs", Narration: ocNarration, Occupancy: &OccupancyBeat{Show: "fill", At: 0}},
			{ID: "bill", Heading: "What you pay", Narration: ocNarration, Occupancy: &OccupancyBeat{Show: "read"}},
		},
	}
	p.targetWords = 3 * 40
	return p
}

func TestOccupancyPlanAccepted(t *testing.T) {
	if err := validateOccupancyPlan(occupancyPlan()); err != nil {
		t.Fatalf("a well-formed occupancy plan was rejected: %v", err)
	}
}

// The rule the picture depends on. Claims that overfill the grid would be
// clamped by the renderer, so the frame would look fine while saying something
// untrue — which is exactly the failure a validator is for.
func TestOccupancyRejectsBandsThatOverfillTheGrid(t *testing.T) {
	p := occupancyPlan()
	p.Occupancy.Bands[0].Count = 900
	err := validateOccupancyPlan(p)
	if err == nil {
		t.Fatal("bands claiming more cells than the population has were accepted")
	}
	if !strings.Contains(err.Error(), "896") {
		t.Fatalf("the error does not name the population: %v", err)
	}
}

// The contrast between the full set and the claimed part is the whole effect,
// so a grid that arrives already lit has no before to contrast with.
func TestOccupancyRequiresTheEmptyGridFirst(t *testing.T) {
	p := occupancyPlan()
	p.Beats[0].Occupancy = &OccupancyBeat{Show: "fill", At: 0}
	p.Beats[1].Occupancy = &OccupancyBeat{Show: "grid"}
	if err := validateOccupancyPlan(p); err == nil {
		t.Fatal("a clip that lights cells before drawing the grid was accepted")
	}
}

func TestOccupancyRejectsASecondGridBeat(t *testing.T) {
	p := occupancyPlan()
	p.Beats[2].Occupancy = &OccupancyBeat{Show: "grid"}
	if err := validateOccupancyPlan(p); err == nil {
		t.Fatal("a clip that re-establishes the population part-way through was accepted")
	}
}

// Below a dozen cells this is a row of boxes; past twelve hundred the cells are
// smaller than the gaps and the grid reads as a texture.
func TestOccupancyRejectsAnUndrawablePopulation(t *testing.T) {
	for _, total := range []int{4, 5000} {
		p := occupancyPlan()
		p.Occupancy.Total = total
		p.Occupancy.Bands[0].Count = 2
		if err := validateOccupancyPlan(p); err == nil {
			t.Fatalf("a population of %d was accepted", total)
		}
	}
}

func TestOccupancyRequiresEveryBandToBeLit(t *testing.T) {
	p := occupancyPlan()
	p.Occupancy.Bands = append(p.Occupancy.Bands, OccupancyBand{
		Count: 40, Label: "Warm in cache", Role: "neutral",
	})
	if err := validateOccupancyPlan(p); err == nil {
		t.Fatal("a band that no beat ever lights was accepted")
	}
}

// A clip where every band is neutral has no point of view, which is a list of
// facts rather than an explanation.
func TestOccupancyRejectsAnAllNeutralArgument(t *testing.T) {
	p := occupancyPlan()
	p.Occupancy.Bands[0].Role = "neutral"
	err := validateOccupancyPlan(p)
	if err == nil {
		t.Fatal("a clip with nothing being argued was accepted")
	}
	if !strings.Contains(err.Error(), "quantity") {
		t.Fatalf("the error does not say how to fix it: %v", err)
	}
}

func TestOccupancyRequiresAUnit(t *testing.T) {
	p := occupancyPlan()
	p.Occupancy.Unit = ""
	if err := validateOccupancyPlan(p); err == nil {
		t.Fatal("a grid of unnamed squares was accepted")
	}
}

// Normalization drops what it can repair without guessing: a band claiming no
// cells lights nothing, so it is not a band.
func TestOccupancyNormalizeDropsEmptyBands(t *testing.T) {
	p := occupancyPlan()
	p.Occupancy.Bands = append(p.Occupancy.Bands, OccupancyBand{Count: 0, Label: "Nothing"})
	normalizeOccupancyPlan(p)
	if len(p.Occupancy.Bands) != 1 {
		t.Fatalf("want 1 band after normalize, got %d", len(p.Occupancy.Bands))
	}
}

// The grid is biased wide because the stage is 16:9 — a square grid of 896
// cells leaves two thirds of the frame empty.
func TestOccupancyGridShapeIsWiderThanTall(t *testing.T) {
	cols, rows := occupancyGridShape(896)
	if cols <= rows {
		t.Fatalf("grid is %dx%d, want wider than tall", cols, rows)
	}
	if cols*rows < 896 {
		t.Fatalf("grid %dx%d cannot hold 896 cells", cols, rows)
	}
}

func TestOccupancyGridShapeHoldsEveryPopulation(t *testing.T) {
	for _, n := range []int{12, 13, 64, 100, 897, 1200} {
		cols, rows := occupancyGridShape(n)
		if cols*rows < n {
			t.Fatalf("grid %dx%d cannot hold %d cells", cols, rows, n)
		}
	}
}

// The scene carries the bands as contiguous runs, so the renderer never has to
// decide which cells a band owns.
func TestOccupancyScenesAssignContiguousRuns(t *testing.T) {
	p := occupancyPlan()
	p.Occupancy.Bands = append(p.Occupancy.Bands, OccupancyBand{
		Count: 30, Label: "Warm in cache", Role: "neutral",
	})
	p.Beats = append(p.Beats, SnippetBeat{
		ID: "warm", Heading: "Still resident", Narration: ocNarration,
		Occupancy: &OccupancyBeat{Show: "fill", At: 1},
	})
	scenes, err := occupancyScenes(sceneInput(t, p, 12000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	bands, _ := scenes[0].Props["bands"].([]map[string]any)
	if len(bands) != 2 {
		t.Fatalf("want 2 bands in the scene, got %d", len(bands))
	}
	if bands[0]["from"] != 0 {
		t.Fatalf("first band starts at %v, want 0", bands[0]["from"])
	}
	if bands[1]["from"] != 16 {
		t.Fatalf("second band starts at %v, want 16 (after the first band's cells)", bands[1]["from"])
	}
}
