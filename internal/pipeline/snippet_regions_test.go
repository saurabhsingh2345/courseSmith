package pipeline

import (
	"strings"
	"testing"
)

const rgNarration = "Low addresses sit at the bottom and every band above them is claimed by somebody with a job."

func regionsPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "regions",
		Title:    "Two ends of one address space",
		Regions: &RegionsSpec{
			Regions: []RegionsRegion{
				{Label: "code", Role: "code", Note: "the instructions themselves, loaded once and never resized"},
				{Label: "static data", Role: "static", Note: "globals and constants sized before the program ever runs"},
				{Label: "the heap", Role: "heap", Note: "whatever you allocate at runtime lives up here"},
				{Label: "free space", Role: "gap", Note: "unclaimed room both neighbours are quietly spending"},
				{Label: "the stack", Role: "stack", Note: "one frame per call, pushed and popped again"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "map", Heading: "The whole space", Narration: rgNarration, Regions: &RegionsBeat{Show: "map"}},
			{ID: "heap", Heading: "The heap", Narration: rgNarration, Regions: &RegionsBeat{Show: "region", At: 2}},
			{ID: "grow-heap", Heading: "Allocating", Narration: rgNarration, Regions: &RegionsBeat{Show: "grow", At: 2}},
			{ID: "grow-stack", Heading: "Calling", Narration: rgNarration, Regions: &RegionsBeat{Show: "grow", At: 4}},
			{ID: "collide", Heading: "They meet", Narration: rgNarration, Regions: &RegionsBeat{Show: "collide"}},
			{ID: "whole", Heading: "One space", Narration: rgNarration, Regions: &RegionsBeat{Show: "whole"}},
		},
	}
	// A beat here is a shot, so the fixture budget is sized at the template's
	// own 28-word ideal — nBeats * 40 would make beatBounds demand more beats
	// than the fixture has.
	p.targetWords = 6 * 28
	return p
}

func TestRegionsPlanAccepted(t *testing.T) {
	if err := validateRegionsPlan(regionsPlan()); err != nil {
		t.Fatalf("a well-formed regions plan was rejected: %v", err)
	}
}

// The family's signature rule: the adjacency IS the picture, so the positions
// are computed in Go and a gap outside the two fronts is rejected with the
// real positions quoted back.
func TestRegionsRejectsAGapOutsideTheHeapAndStack(t *testing.T) {
	p := regionsPlan()
	// code, free space, the heap, static data, the stack — the gap now sits
	// below the heap instead of between the two growing segments.
	p.Regions.Regions[1], p.Regions.Regions[3] = p.Regions.Regions[3], p.Regions.Regions[1]
	err := validateRegionsPlan(p)
	if err == nil {
		t.Fatal("a gap outside the heap and the stack was accepted")
	}
	if !strings.Contains(err.Error(), "position 2") || !strings.Contains(err.Error(), "position 4") {
		t.Fatalf("the error does not quote the heap and stack positions: %v", err)
	}
	if !strings.Contains(err.Error(), "position 1") {
		t.Fatalf("the error does not quote where the gap actually sits: %v", err)
	}
}

func TestRegionsRejectsAHeapAndStackWithNoGap(t *testing.T) {
	p := regionsPlan()
	p.Regions.Regions[3].Role = "static"
	err := validateRegionsPlan(p)
	if err == nil {
		t.Fatal("a heap and a stack with no gap between them was accepted")
	}
	if !strings.Contains(err.Error(), "no gap") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "position 2") || !strings.Contains(err.Error(), "position 4") {
		t.Fatalf("the error does not quote both fronts: %v", err)
	}
}

func TestRegionsRejectsTwoHeaps(t *testing.T) {
	p := regionsPlan()
	p.Regions.Regions[4].Role = "heap"
	err := validateRegionsPlan(p)
	if err == nil {
		t.Fatal("a column with two heaps was accepted")
	}
	if !strings.Contains(err.Error(), "one heap") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegionsRejectsGrowingASegmentThatCannotGrow(t *testing.T) {
	p := regionsPlan()
	p.Beats[2].Regions = &RegionsBeat{Show: "grow", At: 0}
	err := validateRegionsPlan(p)
	if err == nil {
		t.Fatal("a growing code segment was accepted")
	}
	if !strings.Contains(err.Error(), "code") || !strings.Contains(err.Error(), "no operating system") {
		t.Fatalf("the error does not name the role or say why: %v", err)
	}
}

func TestRegionsRejectsACollisionWithOnlyOneFront(t *testing.T) {
	p := regionsPlan()
	p.Regions.Regions[4].Role = "static"
	p.Beats[3].Regions = &RegionsBeat{Show: "grow", At: 2}
	err := validateRegionsPlan(p)
	if err == nil {
		t.Fatal("a collision in a column with no stack was accepted")
	}
	if !strings.Contains(err.Error(), "no stack") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegionsRejectsACollisionBeforeAnyGrowth(t *testing.T) {
	p := regionsPlan()
	p.Beats[1].Regions = &RegionsBeat{Show: "collide"}
	p.Beats[4].Regions = &RegionsBeat{Show: "region", At: 4}
	err := validateRegionsPlan(p)
	if err == nil {
		t.Fatal("a collision before any edge advanced was accepted")
	}
	if !strings.Contains(err.Error(), "before anything has grown") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegionsRejectsASecondCollision(t *testing.T) {
	p := regionsPlan()
	p.Beats[1].Regions = &RegionsBeat{Show: "grow", At: 4}
	p.Beats[2].Regions = &RegionsBeat{Show: "collide"}
	err := validateRegionsPlan(p)
	if err == nil {
		t.Fatal("two collisions in one clip were accepted")
	}
	if !strings.Contains(err.Error(), "collides a second time") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegionsRequiresOpeningOnTheMap(t *testing.T) {
	p := regionsPlan()
	p.Beats[0].Regions = &RegionsBeat{Show: "region", At: 0}
	err := validateRegionsPlan(p)
	if err == nil {
		t.Fatal("a clip that lights a block before showing the column was accepted")
	}
	if !strings.Contains(err.Error(), "open on the whole map") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegionsRejectsTheWholeBeforeTheEnd(t *testing.T) {
	p := regionsPlan()
	p.Beats[1].Regions = &RegionsBeat{Show: "whole"}
	err := validateRegionsPlan(p)
	if err == nil {
		t.Fatal("a closer part-way through the clip was accepted")
	}
	if !strings.Contains(err.Error(), "closer") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRegionsRejectsAnInventedRole(t *testing.T) {
	p := regionsPlan()
	p.Regions.Regions[1].Role = "cache"
	err := validateRegionsPlan(p)
	if err == nil {
		t.Fatal("an invented segment role was accepted")
	}
	if !strings.Contains(err.Error(), "no drawing") {
		t.Fatalf("the error does not explain why: %v", err)
	}
}

func TestRegionsRejectsTooFewSegments(t *testing.T) {
	p := regionsPlan()
	p.Regions.Regions = p.Regions.Regions[:2]
	if err := validateRegionsPlan(p); err == nil {
		t.Fatal("a two-segment column was accepted")
	}
}

func TestRegionsNormalizeCoercesRoles(t *testing.T) {
	p := regionsPlan()
	p.Regions.Regions[2].Role = "HEAP"
	p.Regions.Regions[1].Role = ""
	normalizeRegionsPlan(p)
	if got := p.Regions.Regions[2].Role; got != "heap" {
		t.Fatalf("a shouted role resolved to %q, want heap", got)
	}
	if got := p.Regions.Regions[1].Role; got != "static" {
		t.Fatalf("an empty role resolved to %q, want static", got)
	}
	if err := validateRegionsPlan(p); err != nil {
		t.Fatalf("a sloppy-but-correct plan was rejected after normalize: %v", err)
	}
}

func TestRegionsNormalizeClampsAnOutOfRangeFocus(t *testing.T) {
	p := regionsPlan()
	p.Beats[1].Regions.At = 99
	normalizeRegionsPlan(p)
	if got := p.Beats[1].Regions.At; got != 4 {
		t.Fatalf("an out-of-range focus clamped to %d, want 4", got)
	}
}

func TestRegionsNormalizeClampsALongNote(t *testing.T) {
	p := regionsPlan()
	p.Regions.Regions[0].Note = "one two three four five six seven eight nine ten eleven twelve"
	normalizeRegionsPlan(p)
	if got := len(strings.Fields(p.Regions.Regions[0].Note)); got != maxRegionsNoteWords {
		t.Fatalf("the note kept %d words, want %d", got, maxRegionsNoteWords)
	}
}

func TestRegionsRoleDefaultsToStatic(t *testing.T) {
	r := RegionsRegion{Label: "some band", Role: "sparkle"}
	if got := r.ResolvedRole(); got != "static" {
		t.Fatalf("an unknown role resolved to %q, want static", got)
	}
}

func TestRegionsShowDefaultsToRegion(t *testing.T) {
	b := RegionsBeat{Show: "sparkle"}
	if got := b.ResolvedShow(); got != "region" {
		t.Fatalf("an unknown show resolved to %q, want region", got)
	}
}

// Each step carries the growth so far, so the renderer draws a whole frame
// from one step rather than replaying the beat list, and which way a segment
// grows is decided in Go from its role.
func TestRegionsScenesAccumulateGrowth(t *testing.T) {
	p := regionsPlan()
	scenes, err := regionsScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	if props["heapAt"] != 2 || props["stackAt"] != 4 || props["gapAt"] != 3 {
		t.Fatalf("the landmark positions are wrong: heap=%v stack=%v gap=%v", props["heapAt"], props["stackAt"], props["gapAt"])
	}
	regions, _ := props["regions"].([]map[string]any)
	if len(regions) != 5 {
		t.Fatalf("want 5 segments, got %d", len(regions))
	}
	if regions[2]["grows"] != "up" || regions[4]["grows"] != "down" {
		t.Fatalf("growth directions are wrong: heap=%v stack=%v", regions[2]["grows"], regions[4]["grows"])
	}
	if regions[0]["grows"] != "" {
		t.Fatalf("the code segment was given a growth direction: %v", regions[0]["grows"])
	}

	steps, _ := props["steps"].([]map[string]any)
	if len(steps) != 6 {
		t.Fatalf("want 6 steps, got %d", len(steps))
	}
	first := steps[0]
	if grown, _ := first["grown"].([]int); len(grown) != 0 {
		t.Fatalf("the opening beat already has grown segments: %v", grown)
	}
	if first["collided"] != false {
		t.Fatalf("the opening beat is already collided: %v", first["collided"])
	}
	if first["show"] != "map" {
		t.Fatalf("the first step shows %v, want map", first["show"])
	}

	last := steps[len(steps)-1]
	grown, _ := last["grown"].([]int)
	if len(grown) != 2 || grown[0] != 2 || grown[1] != 4 {
		t.Fatalf("the closing beat's grown set is %v, want [2 4]", grown)
	}
	if last["collided"] != true {
		t.Fatalf("the closing beat has not recorded the collision: %v", last["collided"])
	}
	if last["show"] != "whole" {
		t.Fatalf("the last step shows %v, want whole", last["show"])
	}
}
