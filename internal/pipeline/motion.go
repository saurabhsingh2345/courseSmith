package pipeline

// The motion design system — the single canonical source of the course's
// animation language (timing, easing, stagger). Go owns these tokens because
// the scene graph is Go-built and the archetype system (workstream F) overrides
// them per course, so every consumer downstream must honour whatever the scene
// graph carries. The renderer and studio import a generated JSON mirror of
// DefaultMotion() (renderer/src/theme/motion.defaults.json and the studio copy);
// TestMotionDefaultsInSync guards those mirrors against drift.
//
// Tuning one set of numbers here retunes every animation in the whole course.

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// MotionTiming holds scene/element durations in seconds.
type MotionTiming struct {
	Fast     float64 `json:"fast"`     // micro-interactions
	Normal   float64 `json:"normal"`   // default transition
	Slow     float64 `json:"slow"`     // dramatic reveal
	VerySlow float64 `json:"verySlow"` // intro/outro
}

// MotionEasing holds cubic-bezier easing curves as CSS "cubic-bezier(a,b,c,d)"
// strings — the one representation both CSS (studio) and Remotion's
// Easing.bezier (renderer) can consume.
type MotionEasing struct {
	Entrance string `json:"entrance"` // bounce-like settle
	Exit     string `json:"exit"`     // decelerate out
	Subtle   string `json:"subtle"`   // standard material curve
}

// MotionStagger holds inter-element delays in seconds.
type MotionStagger struct {
	Words       float64 `json:"words"`       // per word in a text reveal
	Items       float64 `json:"items"`       // per list item / outcome
	Connections float64 `json:"connections"` // per network edge
}

// MotionCamera is the slow move the frame makes underneath a scene.
//
// The catalog had no camera at all, and its absence is most of why a finished
// clip read as a slide deck. Every scene animated its *entrances* — things faded
// and rose into place over the first half-second — and then held perfectly still
// for the remaining twenty or fifty seconds while only opacity changed to mark
// which item was being spoken about. A held frame with a voice over it is a
// slide, however well it is set.
//
// This is deliberately not a "camera" in the sense of the story template's shot
// list, which chooses a framing per beat. It is the continuous drift that makes a
// frame feel photographed rather than printed, and it applies to every scene at
// once rather than being something a template opts into and twenty-six others
// forget.
type MotionCamera struct {
	// Push is the MOST the frame scales, as a fraction: 0.04 ends at most 4%
	// closer than it began. A ceiling, not a distance travelled.
	//
	// Past about 0.04 the move becomes a zoom the viewer notices and then
	// resents, and content starts crossing the safe margin — 110px of 1920, which
	// a 5% scale eats.
	Push float64 `json:"push"`
	// Drift is how far the frame travels sideways, in pixels of a 1920-wide frame.
	// Paired with Push so the move has a direction and not only a magnitude: a
	// pure scale reads as a zoom, a scale plus a little lateral travel reads as a
	// camera.
	Drift float64 `json:"drift"`
	// SettleSec is how long the move takes to reach Push, after which the frame
	// holds.
	//
	// This exists because pacing the move across the whole scene — the obvious
	// implementation, and the first one written here — produced no motion at all
	// on the scenes that needed it most. A reel segment runs sixty to a hundred
	// seconds, so spreading 2% over it gives 0.03% a second: measured on a
	// rendered frame the content moved twenty-eight pixels across seventy
	// seconds, which is not a camera, it is a rounding error. Documentary-style
	// pushes run nearer 1% a second.
	//
	// So the move has a RATE (Push/SettleSec) and a duration of its own, and a
	// long scene finishes it and then holds. A push that settles is a real move —
	// it is what a locked-off shot with a slow creep in actually does — and it is
	// strictly better than a crawl nobody can see. The alternative, re-framing per
	// beat so the motion never stops, needs each template to publish its beat
	// boundaries up to the timeline; that is the better design and a bigger one.
	SettleSec float64 `json:"settleSec"`
}

// Motion is the complete animation token set embedded in every scene graph.
type Motion struct {
	Timing  MotionTiming  `json:"timing"`
	Easing  MotionEasing  `json:"easing"`
	Stagger MotionStagger `json:"stagger"`
	Camera  MotionCamera  `json:"camera"`
}

// DefaultMotion is the canonical baseline animation language. Archetypes
// override individual fields via Merge; unspecified fields inherit these.
func DefaultMotion() Motion {
	return Motion{
		Timing: MotionTiming{
			Fast:     0.2,
			Normal:   0.6,
			Slow:     1.0,
			VerySlow: 2.0,
		},
		Easing: MotionEasing{
			Entrance: "cubic-bezier(0.34, 1.56, 0.64, 1)",
			Exit:     "cubic-bezier(0.16, 1, 0.3, 1)",
			Subtle:   "cubic-bezier(0.4, 0, 0.2, 1)",
		},
		Stagger: MotionStagger{
			Words:       0.05,
			Items:       0.08,
			Connections: 0.12,
		},
		Camera: MotionCamera{
			// 3.2% over eighteen seconds — about 0.18% a second. Slow enough that
			// no single second reads as movement, fast enough that a still from
			// the top of a beat and one from the bottom are different shots.
			Push:      0.032,
			Drift:     22,
			SettleSec: 18,
		},
	}
}

// Merge returns a copy of m with every non-zero field of override applied on
// top. This is how the archetype system layers a philosophy (minimal / smooth /
// playful) over the baseline without restating every token.
func (m Motion) Merge(override *Motion) Motion {
	if override == nil {
		return m
	}
	out := m
	if override.Timing.Fast != 0 {
		out.Timing.Fast = override.Timing.Fast
	}
	if override.Timing.Normal != 0 {
		out.Timing.Normal = override.Timing.Normal
	}
	if override.Timing.Slow != 0 {
		out.Timing.Slow = override.Timing.Slow
	}
	if override.Timing.VerySlow != 0 {
		out.Timing.VerySlow = override.Timing.VerySlow
	}
	if override.Easing.Entrance != "" {
		out.Easing.Entrance = override.Easing.Entrance
	}
	if override.Easing.Exit != "" {
		out.Easing.Exit = override.Easing.Exit
	}
	if override.Easing.Subtle != "" {
		out.Easing.Subtle = override.Easing.Subtle
	}
	if override.Stagger.Words != 0 {
		out.Stagger.Words = override.Stagger.Words
	}
	if override.Stagger.Items != 0 {
		out.Stagger.Items = override.Stagger.Items
	}
	if override.Stagger.Connections != 0 {
		out.Stagger.Connections = override.Stagger.Connections
	}
	if override.Camera.Push != 0 {
		out.Camera.Push = override.Camera.Push
	}
	if override.Camera.Drift != 0 {
		out.Camera.Drift = override.Camera.Drift
	}
	if override.Camera.SettleSec != 0 {
		out.Camera.SettleSec = override.Camera.SettleSec
	}
	return out
}

// MotionDefaultsJSON is the exact byte content the TS mirrors must contain:
// DefaultMotion() as 2-space-indented JSON with a trailing newline (matching
// Prettier's default so the committed files stay stable).
func MotionDefaultsJSON() ([]byte, error) {
	b, err := json.MarshalIndent(DefaultMotion(), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// parseBezier extracts the four control-point numbers from a
// "cubic-bezier(a, b, c, d)" string. Used by validation; the renderer has its
// own TS parser for the same format.
func parseBezier(s string) ([4]float64, error) {
	var out [4]float64
	s = strings.TrimSpace(s)
	inner, ok := strings.CutPrefix(s, "cubic-bezier(")
	if !ok {
		return out, fmt.Errorf("not a cubic-bezier(): %q", s)
	}
	inner, ok = strings.CutSuffix(inner, ")")
	if !ok {
		return out, fmt.Errorf("missing closing paren: %q", s)
	}
	parts := strings.Split(inner, ",")
	if len(parts) != 4 {
		return out, fmt.Errorf("want 4 control points, got %d: %q", len(parts), s)
	}
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return out, fmt.Errorf("control point %d (%q): %w", i, p, err)
		}
		out[i] = v
	}
	return out, nil
}
