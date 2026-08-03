package pipeline

import (
	"bytes"
	"os"
	"testing"
)

// motionMirrors are the generated TS-side copies of DefaultMotion(). They must
// stay byte-identical to MotionDefaultsJSON(); this is the single-source-of-
// truth guarantee. Regenerate with: COURSESMITH_UPDATE_MOTION=1 go test ./internal/pipeline -run TestMotionDefaultsInSync
var motionMirrors = []string{
	"../../renderer/src/theme/motion.defaults.json",
	"../../studio/src/theme/motion.defaults.json",
}

func TestMotionDefaultsInSync(t *testing.T) {
	want, err := MotionDefaultsJSON()
	if err != nil {
		t.Fatalf("MotionDefaultsJSON: %v", err)
	}
	if os.Getenv("COURSESMITH_UPDATE_MOTION") == "1" {
		for _, p := range motionMirrors {
			if err := os.WriteFile(p, want, 0o644); err != nil {
				t.Fatalf("writing %s: %v", p, err)
			}
			t.Logf("regenerated %s", p)
		}
		return
	}
	for _, p := range motionMirrors {
		got, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("reading %s (regenerate with COURSESMITH_UPDATE_MOTION=1): %v", p, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s is out of sync with DefaultMotion() — regenerate with COURSESMITH_UPDATE_MOTION=1 go test ./internal/pipeline -run TestMotionDefaultsInSync", p)
		}
	}
}

func TestDefaultMotionValid(t *testing.T) {
	m := DefaultMotion()

	// Timing tiers must be strictly increasing — the semantic ordering the
	// tokens promise (fast < normal < slow < verySlow).
	tiers := []struct {
		name string
		v    float64
	}{
		{"fast", m.Timing.Fast},
		{"normal", m.Timing.Normal},
		{"slow", m.Timing.Slow},
		{"verySlow", m.Timing.VerySlow},
	}
	for i, tr := range tiers {
		if tr.v <= 0 {
			t.Errorf("timing.%s = %v, want > 0", tr.name, tr.v)
		}
		if i > 0 && tr.v <= tiers[i-1].v {
			t.Errorf("timing.%s (%v) must exceed timing.%s (%v)", tr.name, tr.v, tiers[i-1].name, tiers[i-1].v)
		}
	}

	// Every easing must be a parseable cubic-bezier with x-coords in [0,1]
	// (a valid CSS timing function; y may overshoot for bounce curves).
	for name, e := range map[string]string{"entrance": m.Easing.Entrance, "exit": m.Easing.Exit, "subtle": m.Easing.Subtle} {
		cp, err := parseBezier(e)
		if err != nil {
			t.Errorf("easing.%s: %v", name, err)
			continue
		}
		if cp[0] < 0 || cp[0] > 1 || cp[2] < 0 || cp[2] > 1 {
			t.Errorf("easing.%s: x control points must be in [0,1], got %v and %v", name, cp[0], cp[2])
		}
	}

	for name, s := range map[string]float64{"words": m.Stagger.Words, "items": m.Stagger.Items, "connections": m.Stagger.Connections} {
		if s <= 0 {
			t.Errorf("stagger.%s = %v, want > 0", name, s)
		}
	}

	// The camera has to move, and has to move imperceptibly. The whole effect
	// depends on being invisible frame to frame: past about 4% it stops being a
	// camera and becomes a zoom the viewer notices, and content starts leaving the
	// frame — the safe margins are 110px of 1920, which a 5% scale eats.
	if m.Camera.Push <= 0 {
		t.Error("camera.push is zero — every scene would hold perfectly still, which is what made the output read as slides")
	}
	if m.Camera.Push > 0.04 {
		t.Errorf("camera.push = %v; past 0.04 the move is visible as a zoom and pushes content past the safe margin", m.Camera.Push)
	}
	if m.Camera.Drift < 0 {
		t.Errorf("camera.drift = %v, want >= 0", m.Camera.Drift)
	}
	// Drift is in pixels of a 1920 frame and travels half its value either side of
	// centre, so it must stay well inside SAFE_X (110).
	if m.Camera.Drift > 60 {
		t.Errorf("camera.drift = %vpx; past 60 the lateral travel is noticeable and eats the horizontal safe margin", m.Camera.Drift)
	}
}

// An archetype must be able to retune the camera, and — the case that matters —
// to switch it off. A "minimal" philosophy that wanted stillness could not ask
// for it if zero meant "inherit", so this pins which way Merge treats it.
func TestMotionCameraMerges(t *testing.T) {
	base := DefaultMotion()

	louder := base.Merge(&Motion{Camera: MotionCamera{Push: 0.035, Drift: 40}})
	if louder.Camera.Push != 0.035 || louder.Camera.Drift != 40 {
		t.Errorf("override did not apply: %+v", louder.Camera)
	}
	// And it must not disturb the rest of the token set.
	if louder.Timing != base.Timing || louder.Easing != base.Easing || louder.Stagger != base.Stagger {
		t.Error("overriding the camera changed another token group")
	}

	// Zero means inherit, consistent with every other field in Merge. Switching
	// the camera off is therefore a renderer-side decision (a scene graph carrying
	// an explicit zero), not something Merge can express — which is worth knowing
	// rather than discovering.
	inherited := base.Merge(&Motion{})
	if inherited.Camera != base.Camera {
		t.Errorf("an empty override changed the camera: %+v", inherited.Camera)
	}
}

func TestMotionMerge(t *testing.T) {
	base := DefaultMotion()

	// Nil override is a no-op.
	if got := base.Merge(nil); got != base {
		t.Errorf("Merge(nil) changed the base")
	}

	// A partial "playful" override: faster stagger, bouncier — only the set
	// fields change, the rest inherit.
	playful := &Motion{
		Stagger: MotionStagger{Words: 0.09},
		Easing:  MotionEasing{Entrance: "cubic-bezier(0.68, -0.55, 0.27, 1.55)"},
	}
	got := base.Merge(playful)
	if got.Stagger.Words != 0.09 {
		t.Errorf("stagger.words = %v, want 0.09", got.Stagger.Words)
	}
	if got.Easing.Entrance != playful.Easing.Entrance {
		t.Errorf("easing.entrance not overridden")
	}
	// Untouched fields inherit the baseline.
	if got.Timing != base.Timing {
		t.Errorf("timing changed unexpectedly: %+v", got.Timing)
	}
	if got.Stagger.Items != base.Stagger.Items {
		t.Errorf("stagger.items = %v, want inherited %v", got.Stagger.Items, base.Stagger.Items)
	}
	if got.Easing.Exit != base.Easing.Exit {
		t.Errorf("easing.exit changed unexpectedly")
	}
}
