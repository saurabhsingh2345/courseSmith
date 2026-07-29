package pipeline

// The constellation template: one idea, and everything that hangs off it.
//
// The Redis reference clip ends on exactly this frame — the name in the middle
// with four properties radiating out — and it is doing something none of the
// other closing shapes can. A list recaps in the order things were said, which
// is the order of the *video*. A radial map recaps in no order at all, which is
// how the idea will actually be stored: not as a sequence but as a thing with
// properties attached.
//
// `flow` is the closest existing template and it makes the opposite claim. A
// flow has direction — this passes to that — and drawing an idea's properties
// as a flow would assert a sequence between them that does not exist.
//
// The rule that earns it its place is that every spoke has to complete a
// sentence about the centre. "Redis / is in-memory" works; "Redis / Chapter 3"
// does not, and a map whose spokes are sub-topics rather than properties is a
// table of contents wearing a diagram's clothes — which is the exact shape a
// model produces when asked to summarise. So each spoke carries the relation
// word that joins it to the centre, and a spoke that cannot supply one is
// rejected.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "constellation",
		Category:    CatConcepts,
		Since:       SinceV1,
		Title:       "The whole picture",
		Description: "One idea in the middle with everything that hangs off it, lighting one spoke at a time.",
		Example:     "Everything that makes Redis Redis, in one picture",
		PromptFile:  snippetConstellationTemplateName,
		NeedsCode:   false,
		// The centre, three spokes and the closing frame is five beats.
		MinTargetSec:     35,
		DefaultTargetSec: 55,
		MaxBeats:         8,
		Owns:             beatFields{Constellation: true},
		OwnsPlan:         planFields{Constellation: true},
		Normalize:        normalizeConstellationPlan,
		Validate:         validateConstellationPlan,
		Scenes:           constellationScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":          strings.Join(ConstellationShows(), ", "),
				"Icons":          strings.Join(PointIconNames(), ", "),
				"MinSpokes":      minConstellationSpokes,
				"MaxSpokes":      maxConstellationSpokes,
				"MaxCentreWords": maxConstellationCentreWords,
				"MaxRelWords":    maxConstellationRelWords,
				"MaxLabelWords":  maxConstellationLabelWords,
				"MaxNoteWords":   maxConstellationNoteWords,
			}
		},
	})
}

const snippetConstellationTemplateName = "snippet_constellation.tmpl"

const (
	// Three spokes is the fewest that reads as a map rather than a pair of
	// labels; six around a centre at this size start colliding.
	minConstellationSpokes = 3
	maxConstellationSpokes = 6

	maxConstellationCentreWords = 3
	// The relation word or two that joins a spoke to the centre — "is",
	// "runs on", "gives you". Short by design: it sits on the connector.
	maxConstellationRelWords   = 3
	maxConstellationLabelWords = 4
	maxConstellationNoteWords  = 12
)

// constellationShows is the closed vocabulary of what a beat does.
var constellationShows = map[string]bool{
	// The centre alone, spokes not yet drawn.
	"centre": true,
	// Light one spoke and draw its connector.
	"spoke": true,
	// Everything lit at once. The closing frame.
	"whole": true,
}

// ConstellationShows returns the beat vocabulary sorted.
func ConstellationShows() []string {
	out := make([]string, 0, len(constellationShows))
	for k := range constellationShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ConstellationSpec is the idea and what hangs off it.
type ConstellationSpec struct {
	// Centre is the idea itself — "Redis".
	Centre string `json:"centre"`
	// CentreIcon is a PointIconNames name drawn in the middle node.
	CentreIcon string `json:"centreIcon,omitempty"`
	// Spokes are the properties, in the order they are lit.
	Spokes []ConstellationSpoke `json:"spokes"`
}

// ResolvedCentreIcon returns the middle node's glyph.
func (c ConstellationSpec) ResolvedCentreIcon() string {
	if icon := normalizePointIconName(c.CentreIcon); icon != "" {
		return icon
	}
	return "sparkles"
}

// ConstellationSpoke is one property of the centre.
type ConstellationSpoke struct {
	// Rel is the relation word that joins this spoke to the centre — "is",
	// "runs on", "gives you". It sits on the connector, and it is what makes
	// the spoke a property rather than a sub-topic.
	Rel string `json:"rel"`
	// Label is the property itself — "in-memory", "single-threaded".
	Label string `json:"label"`
	// Note is the line that arrives when the spoke lights.
	Note string `json:"note,omitempty"`
	// Icon is a PointIconNames name drawn in the spoke node.
	Icon string `json:"icon,omitempty"`
}

// ResolvedIcon returns the spoke node's glyph.
func (s ConstellationSpoke) ResolvedIcon() string {
	if icon := normalizePointIconName(s.Icon); icon != "" {
		return icon
	}
	return "dot"
}

// ConstellationBeat is one move.
type ConstellationBeat struct {
	Show string `json:"show"`
	At   int    `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to lighting a
// spoke — the bulk of the clip.
func (b ConstellationBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if constellationShows[s] {
		return s
	}
	return "spoke"
}

func normalizeConstellationPlan(p *SnippetPlan) {
	c := p.Constellation
	if c == nil {
		return
	}
	c.Centre = clampWords(collapseSpaces(c.Centre), maxConstellationCentreWords)
	c.CentreIcon = c.ResolvedCentreIcon()

	spokes := make([]ConstellationSpoke, 0, len(c.Spokes))
	for _, s := range c.Spokes {
		s.Rel = clampWords(collapseSpaces(s.Rel), maxConstellationRelWords)
		s.Label = clampWords(collapseSpaces(s.Label), maxConstellationLabelWords)
		s.Note = clampWords(collapseSpaces(s.Note), maxConstellationNoteWords)
		s.Icon = s.ResolvedIcon()
		if s.Label != "" && len(spokes) < maxConstellationSpokes {
			spokes = append(spokes, s)
		}
	}
	c.Spokes = spokes

	for i := range p.Beats {
		b := p.Beats[i].Constellation
		if b == nil {
			continue
		}
		if !constellationShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			switch {
			case i == 0:
				b.Show = "centre"
			case i == len(p.Beats)-1:
				b.Show = "whole"
			default:
				b.Show = "spoke"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "spoke" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(c.Spokes); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateConstellationPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Constellation: true}); err != nil {
		return err
	}

	c := p.Constellation
	if c == nil {
		return fmt.Errorf("the plan has no constellation — this template is one idea with everything that hangs off it")
	}
	if strings.TrimSpace(c.Centre) == "" {
		return fmt.Errorf("there is nothing in the middle — name the idea everything else is a property of")
	}
	if n := len(c.Spokes); n < minConstellationSpokes || n > maxConstellationSpokes {
		return fmt.Errorf("there are %d spokes, want %d-%d. Two is a pair of labels rather than a map, and seven around a centre start colliding",
			n, minConstellationSpokes, maxConstellationSpokes)
	}

	seen := map[string]bool{}
	for i, s := range c.Spokes {
		if strings.TrimSpace(s.Label) == "" {
			return fmt.Errorf("spoke %d has no label", i)
		}
		// The rule this template exists for.
		if strings.TrimSpace(s.Rel) == "" {
			return fmt.Errorf("spoke %q has no relation to %q. Every spoke completes a sentence about the centre — %q IS in-memory, %q GIVES YOU sorted sets. A spoke with no relation is a sub-topic, and a map of sub-topics is a table of contents wearing a diagram's clothes",
				s.Label, c.Centre, c.Centre, c.Centre)
		}
		key := strings.ToLower(strings.TrimSpace(s.Label))
		if seen[key] {
			return fmt.Errorf("two spokes are both %q — each one is a different property", s.Label)
		}
		seen[key] = true
	}

	lit := map[int]bool{}
	counts := map[string]int{}
	for i, b := range p.Beats {
		if b.Constellation == nil {
			return fmt.Errorf("beat %q has no constellation direction — every beat names the centre, lights one spoke, or shows the whole picture", b.ID)
		}
		show := b.Constellation.ResolvedShow()
		counts[show]++
		if i == 0 && show != "centre" {
			return fmt.Errorf("the clip opens on %q. Name the centre first — a property drawn before the thing it belongs to is a label floating in space", show)
		}
		if show == "whole" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q shows the whole picture but the clip carries on afterwards. That frame is the close", b.ID)
		}
		if show != "spoke" {
			continue
		}
		if b.Constellation.At < 0 || b.Constellation.At >= len(c.Spokes) {
			return fmt.Errorf("beat %q lights spoke %d, which does not exist", b.ID, b.Constellation.At)
		}
		if lit[b.Constellation.At] {
			return fmt.Errorf("beat %q lights spoke %d again; each property gets one beat", b.ID, b.Constellation.At)
		}
		lit[b.Constellation.At] = true
	}
	if counts["centre"] != 1 {
		return fmt.Errorf("there are %d centre beats; the idea is named once", counts["centre"])
	}
	if len(lit) != len(c.Spokes) {
		return fmt.Errorf("%d of the %d spokes are never spoken. A property drawn but not narrated is one the viewer will not recall, which is the only job a closing map has",
			len(c.Spokes)-len(lit), len(c.Spokes))
	}
	if counts["whole"] > 1 {
		return fmt.Errorf("there are %d closing beats; the picture assembles once", counts["whole"])
	}
	return nil
}

// constellationScenes lays the clip out as ONE scene. The spoke positions are
// computed here rather than in the renderer so the layout is deterministic and
// the same map is drawn every render.
func constellationScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.Constellation
	if c == nil {
		return nil, fmt.Errorf("the plan has no constellation")
	}

	spokes := make([]map[string]any, len(c.Spokes))
	for i, s := range c.Spokes {
		spokes[i] = map[string]any{
			"rel":   s.Rel,
			"label": s.Label,
			"note":  s.Note,
			"icon":  s.ResolvedIcon(),
			// Angle around the centre, in degrees, starting at the top and
			// going clockwise. Computed in Go so a map with four spokes always
			// puts them at the compass points rather than wherever a layout
			// pass happened to land them.
			"angle": float64(i)*(360.0/float64(len(c.Spokes))) - 90,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Constellation == nil {
			return nil, fmt.Errorf("beat %q has no constellation direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Constellation.ResolvedShow(),
		}
		if beat.Constellation.ResolvedShow() == "spoke" {
			step["at"] = beat.Constellation.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneConstellation,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":      in.Plan.Title,
			"centre":     c.Centre,
			"centreIcon": c.ResolvedCentreIcon(),
			"spokes":     spokes,
			"steps":      steps,
		},
	}}, nil
}
