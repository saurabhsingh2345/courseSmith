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

// Motion is the complete animation token set embedded in every scene graph.
type Motion struct {
	Timing  MotionTiming  `json:"timing"`
	Easing  MotionEasing  `json:"easing"`
	Stagger MotionStagger `json:"stagger"`
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
