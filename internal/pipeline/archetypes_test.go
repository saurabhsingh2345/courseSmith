package pipeline

import (
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

func TestResolveArchetypeDefaults(t *testing.T) {
	// Empty selection → baseline motion, no palette.
	r, err := ResolveArchetype(config.Style{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Motion != DefaultMotion() {
		t.Errorf("empty style should yield DefaultMotion, got %+v", r.Motion)
	}
	if r.HasPalette {
		t.Error("empty style should not set a palette")
	}
}

func TestResolveArchetypeAppliesPhilosophyAndPalette(t *testing.T) {
	r, err := ResolveArchetype(config.Style{Archetype: "concept-first", ColorPalette: "cool"})
	if err != nil {
		t.Fatal(err)
	}
	// concept-first defaults to the "smooth" philosophy, which slows Normal.
	if r.AnimationStyle != "smooth" {
		t.Errorf("animation style = %q, want smooth (archetype default)", r.AnimationStyle)
	}
	if r.Motion.Timing.Normal <= DefaultMotion().Timing.Normal {
		t.Errorf("smooth motion should lengthen Normal timing, got %v", r.Motion.Timing.Normal)
	}
	// Untouched tokens still inherit the baseline.
	if r.Motion.Timing.Fast != DefaultMotion().Timing.Fast {
		t.Errorf("Fast timing changed unexpectedly: %v", r.Motion.Timing.Fast)
	}
	if !r.HasPalette || r.Palette.Primary != colorPalettes["cool"].Primary {
		t.Errorf("cool palette not applied: %+v", r.Palette)
	}
}

func TestResolveArchetypeExplicitAnimationWins(t *testing.T) {
	r, err := ResolveArchetype(config.Style{Archetype: "concept-first", AnimationStyle: "playful"})
	if err != nil {
		t.Fatal(err)
	}
	if r.AnimationStyle != "playful" {
		t.Errorf("explicit animation_style should win, got %q", r.AnimationStyle)
	}
}

func TestResolveArchetypeUnknownErrors(t *testing.T) {
	for _, s := range []config.Style{
		{Archetype: "nope"},
		{AnimationStyle: "zoomy"},
		{ColorPalette: "neon"},
	} {
		if _, err := ResolveArchetype(s); err == nil {
			t.Errorf("expected error for %+v", s)
		}
	}
}

func TestInterleaveQuestionsSpreadsTypes(t *testing.T) {
	// A front-loaded quiz: 3 recall then 3 application.
	qs := []Question{
		{ID: "r1", Type: QRecall}, {ID: "r2", Type: QRecall}, {ID: "r3", Type: QRecall},
		{ID: "a1", Type: QApplication}, {ID: "a2", Type: QApplication}, {ID: "a3", Type: QApplication},
	}
	order := interleaveQuestions(qs)
	if len(order) != 6 {
		t.Fatalf("got %d ids, want 6", len(order))
	}
	// Every question appears exactly once.
	seen := map[string]bool{}
	for _, id := range order {
		if seen[id] {
			t.Fatalf("duplicate id %q in order", id)
		}
		seen[id] = true
	}
	// With two balanced types, a perfect interleave has zero adjacent repeats.
	typeOf := map[string]string{}
	for _, q := range qs {
		typeOf[q.ID] = q.Type
	}
	repeats := 0
	for i := 1; i < len(order); i++ {
		if typeOf[order[i]] == typeOf[order[i-1]] {
			repeats++
		}
	}
	if repeats != 0 {
		t.Errorf("expected 0 adjacent same-type pairs, got %d (%v)", repeats, order)
	}
}

func TestDifficultyTargetsSumToN(t *testing.T) {
	for _, n := range []int{5, 7, 10, 13} {
		d := difficultyTargets(n)
		if d["easy"]+d["medium"]+d["hard"] != n {
			t.Errorf("targets for n=%d sum to %d, want %d (%v)", n, d["easy"]+d["medium"]+d["hard"], n, d)
		}
		if d["medium"] < d["easy"] || d["medium"] < d["hard"] {
			t.Errorf("n=%d: medium should dominate, got %v", n, d)
		}
	}
}
