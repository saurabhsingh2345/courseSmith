package pipeline

// Archetype system (workstream F): a course declares an archetype, an animation
// philosophy, and a colour palette in course.yaml; these resolve to a Motion
// override, a colour palette, and prompt hints, so one line of config restyles
// pacing, motion, and visuals together instead of tuning dozens of knobs.
//
// This is the scaffold: the registry and resolution are real (and drive
// SceneGraph.Motion via Motion.Merge), while the prompt-hint injection into the
// script/quiz generators is left as a documented TODO.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

// Archetype is one course shape. PaceWPM/MotionStyle/Prompt hints capture how
// this kind of course should feel; the concrete motion comes from the animation
// philosophy layered on top.
type Archetype struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// DefaultAnimation is the philosophy used when the course doesn't pick one.
	DefaultAnimation string `json:"default_animation"`
	// PaceHint is an advisory narration pace (wpm) typical for the archetype.
	PaceHint int `json:"pace_hint"`
	// PromptHint is injected into content-generation prompts (TODO: wire into
	// script.go / quiz.go). It nudges structure, not wording.
	PromptHint string `json:"prompt_hint"`
}

// Archetypes is the registry. Keys are the course.yaml `archetype` values.
var Archetypes = map[string]Archetype{
	"project-based": {
		Name: "project-based", Description: "Build software incrementally.",
		DefaultAnimation: "playful", PaceHint: 160,
		PromptHint: "Structure each section around building one runnable piece; end sections with a 'run it now' beat. Favor short code cycles over long exposition.",
	},
	"concept-first": {
		Name: "concept-first", Description: "Deep conceptual understanding.",
		DefaultAnimation: "smooth", PaceHint: 135,
		PromptHint: "Introduce one idea at a time, concrete example before formal definition. Prefer diagrams and analogies; allow longer reveals.",
	},
	"practical-skills": {
		Name: "practical-skills", Description: "Learn a specific tool or workflow.",
		DefaultAnimation: "minimal", PaceHint: 150,
		PromptHint: "Step-by-step, action-oriented. Each step is a concrete thing the learner does; minimize theory.",
	},
	"story-driven": {
		Name: "story-driven", Description: "Narrative arc with characters and metaphor.",
		DefaultAnimation: "playful", PaceHint: 145,
		PromptHint: "Carry a running narrative or metaphor across sections. Use characters and stakes; humor is welcome.",
	},
	"reference": {
		Name: "reference", Description: "Lookup-first, modular, minimal fluff.",
		DefaultAnimation: "minimal", PaceHint: 155,
		PromptHint: "Modular, self-contained sections optimized for lookup. State facts plainly; no narrative connective tissue.",
	},
}

// animationMotions maps an animation philosophy to a Motion override applied
// over DefaultMotion(). Only the fields that differ from the baseline are set.
var animationMotions = map[string]Motion{
	"minimal": {
		Timing:  MotionTiming{Fast: 0.15, Normal: 0.35},
		Easing:  MotionEasing{Entrance: "cubic-bezier(0.4, 0, 0.2, 1)"}, // no overshoot
		Stagger: MotionStagger{Words: 0.03, Items: 0.05, Connections: 0.06},
	},
	"smooth": {
		Timing:  MotionTiming{Normal: 0.7, Slow: 1.2},
		Easing:  MotionEasing{Entrance: "cubic-bezier(0.22, 1, 0.36, 1)"},
		Stagger: MotionStagger{Items: 0.08, Connections: 0.12},
	},
	"playful": {
		Timing:  MotionTiming{Fast: 0.25, Normal: 0.6},
		Easing:  MotionEasing{Entrance: "cubic-bezier(0.68, -0.55, 0.27, 1.55)"}, // bouncy
		Stagger: MotionStagger{Words: 0.06, Items: 0.1, Connections: 0.14},
	},
}

// colorPalettes maps a palette name to brand colours. "colorblind" uses an
// Okabe-Ito-derived pairing safe for deuteranopia/protanopia.
var colorPalettes = map[string]config.Colors{
	"corporate":  {Primary: "#1f4e79", Accent: "#2e86de", Background: "#ffffff"},
	"warm":       {Primary: "#c0392b", Accent: "#e67e22", Background: "#fffaf5"},
	"cool":       {Primary: "#5b4b8a", Accent: "#0aa3a3", Background: "#f7fbfb"},
	"colorblind": {Primary: "#0072b2", Accent: "#e69f00", Background: "#ffffff"},
}

// sceneTemplates maps an animation philosophy to the scene-template variant
// each scene type renders with. This is the "template format": one line of
// course config (animation_style / archetype) selects a coherent set of
// component variants, and video-plan.yaml can override any single scene.
// The renderer treats an unknown/absent template as its default variant, so
// new variants can ship renderer-first.
var sceneTemplates = map[string]map[string]string{
	"minimal": {
		SceneTitle:  "clean",
		ScenePoints: "rows",
	},
	"smooth": {
		SceneTitle:  "hero",
		ScenePoints: "rows",
	},
	"playful": {
		SceneTitle:  "hero",
		ScenePoints: "grid",
	},
}

// TemplateFor returns the scene-template variant for a scene type under this
// archetype's animation philosophy ("" = renderer default).
func (r ResolvedArchetype) TemplateFor(sceneType string) string {
	if m, ok := sceneTemplates[r.AnimationStyle]; ok {
		return m[sceneType]
	}
	return ""
}

// ResolvedArchetype is the outcome of applying a course's archetype selection.
type ResolvedArchetype struct {
	Archetype      Archetype
	AnimationStyle string
	Motion         Motion        // DefaultMotion().Merge(philosophy override)
	Palette        config.Colors // zero value when no palette selected
	HasPalette     bool
}

// ResolveArchetype interprets the style selection. Any field may be empty:
// empty archetype → a neutral default; empty animation → the archetype's
// default; empty palette → keep the course's own branding colours. Unknown
// names are an error so typos surface at config time.
func ResolveArchetype(style config.Style) (ResolvedArchetype, error) {
	out := ResolvedArchetype{Motion: DefaultMotion()}

	if style.Archetype != "" {
		a, ok := Archetypes[style.Archetype]
		if !ok {
			return out, fmt.Errorf("unknown archetype %q (options: %s)", style.Archetype, knownKeys(Archetypes))
		}
		out.Archetype = a
	}

	anim := style.AnimationStyle
	if anim == "" {
		anim = out.Archetype.DefaultAnimation
	}
	if anim != "" {
		override, ok := animationMotions[anim]
		if !ok {
			return out, fmt.Errorf("unknown animation_style %q (options: minimal, smooth, playful)", anim)
		}
		out.AnimationStyle = anim
		out.Motion = DefaultMotion().Merge(&override)
	}

	if style.ColorPalette != "" {
		p, ok := colorPalettes[style.ColorPalette]
		if !ok {
			return out, fmt.Errorf("unknown color_palette %q (options: corporate, warm, cool, colorblind)", style.ColorPalette)
		}
		out.Palette = p
		out.HasPalette = true
	}
	return out, nil
}

// AnimationStyleNames returns the available animation philosophies, sorted.
func AnimationStyleNames() []string {
	names := make([]string, 0, len(animationMotions))
	for k := range animationMotions {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// PaletteColors is one named colour palette exposed to the studio.
type PaletteColors struct {
	Name   string        `json:"name"`
	Colors config.Colors `json:"colors"`
}

// ColorPaletteList returns the named brand palettes, sorted by name.
func ColorPaletteList() []PaletteColors {
	out := make([]PaletteColors, 0, len(colorPalettes))
	for name, colors := range colorPalettes {
		out = append(out, PaletteColors{Name: name, Colors: colors})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ArchetypeList returns the archetype registry as a stable, sorted slice so the
// studio can present a catalog without reaching into the map's iteration order.
func ArchetypeList() []Archetype {
	out := make([]Archetype, 0, len(Archetypes))
	for _, a := range Archetypes {
		out = append(out, a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// knownKeys renders a sorted, comma-joined key list for error messages.
func knownKeys(m map[string]Archetype) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
