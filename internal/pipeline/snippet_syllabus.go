package pipeline

// The syllabus template: the whole journey, and where on it you are standing.
//
// A course is the one thing the catalog kept explaining pieces of without ever
// drawing. `chapter` opens a section, `objective` opens a lesson — both point
// at the next ten minutes. Nothing could put the whole route on screen: eight
// modules as stops on a line, the ones behind you ticked, the one you are on
// lit, the ones ahead dimmed but visible. For a career-switcher working
// through CS foundations that picture is not decoration, it is the answer to
// the question that makes people quit — "how much of this is left, and does
// what I just did count for anything?"
//
// The frame is a route, not a list, and the rules keep it one.
//
// **The clip opens on the map.** A stop lighting up on a route the viewer has
// not seen whole is a location without a journey — "module three" means
// nothing until modules one through eight have been on screen together. And
// the map is shown once: re-showing it mid-clip resets the very picture the
// stops have been building on.
//
// **Exactly one "here", and it is the last beat.** The template's job is to
// land the viewer somewhere. A clip that lands twice has two presents, and a
// clip that lands early then keeps touring has told the viewer where they are
// and then wandered off — the arrival is the ending or it is not an arrival.
//
// **A stop is visited once.** The route already orders the modules; focusing
// the same one twice spends a beat saying nothing new, and on an eight-beat
// budget that beat came out of some module's only appearance.
//
// The bounds: below four modules the route is a before/after, not a journey,
// and `bridge` draws that hand-off better; past eight the stops compress until
// their labels collide and the map stops being readable at a glance.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "syllabus",
		Category:    CatPresenting,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The whole journey",
		Description: "The course map: four to eight modules as stops on a route, the current one lit with its sub-line, everything behind it ticked. Reach for it to open a course, to open a module, or any time the viewer needs to see how far they have come.",
		Example:     "The 8 modules of CS Foundations, we are on Memory & Storage",
		PromptFile:  snippetSyllabusTemplateName,
		NeedsCode:   false,
		// The map, a few stops worth pausing on, and the landing. Below ~25s
		// there is no room to visit anything between the map and the arrival.
		MinTargetSec:     25,
		DefaultTargetSec: 40,
		// List-shaped: the map, up to six visited stops, and the landing.
		MaxBeats:  8,
		Owns:      beatFields{Syllabus: true},
		OwnsPlan:  planFields{Syllabus: true},
		Normalize: normalizeSyllabusPlan,
		Validate:  validateSyllabusPlan,
		Scenes:    syllabusScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(SyllabusShows(), ", "),
				"MinModules":    minSyllabusModules,
				"MaxModules":    maxSyllabusModules,
				"MaxLabelWords": maxSyllabusLabelWords,
				"MaxSubWords":   maxSyllabusSubWords,
			}
		},
	})
}

const snippetSyllabusTemplateName = "snippet_syllabus.tmpl"

const (
	// Below four modules the route is a before/after rather than a journey —
	// that hand-off is bridge's picture. Past eight the stops compress until
	// their labels collide and the map is no longer readable at a glance.
	minSyllabusModules = 4
	maxSyllabusModules = 8

	// A stop's label is a name on a node, not a sentence; four words is where
	// names end and descriptions begin.
	maxSyllabusLabelWords = 4
	// The sub-line under the lit stop gets one clause of what the module
	// covers. Eight words fills the card without wrapping it.
	maxSyllabusSubWords = 8
)

// syllabusShows is the closed vocabulary of what a beat does.
var syllabusShows = map[string]bool{
	// The whole route, dimmed. The first beat, and only the first.
	"map": true,
	// One stop lifts, with its sub-line.
	"stop": true,
	// Land on the current module, everything before it ticked. The last beat.
	"here": true,
}

// SyllabusShows returns the beat vocabulary sorted, for the prompt.
func SyllabusShows() []string {
	out := make([]string, 0, len(syllabusShows))
	for k := range syllabusShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// SyllabusSpec is the course map: the modules in order, and where we are.
type SyllabusSpec struct {
	// Modules are the stops, drawn left to right in course order.
	Modules []SyllabusModule `json:"modules"`
	// Current indexes Modules: the stop the viewer is standing on. Everything
	// before it is done; everything after is still ahead.
	Current int `json:"current"`
}

// SyllabusModule is one stop on the route.
type SyllabusModule struct {
	// Label names the module — "Memory & Storage".
	Label string `json:"label"`
	// Sub is the one-line summary shown when this stop is lit. Optional: an
	// unannotated stop is still a stop.
	Sub string `json:"sub,omitempty"`
}

// SyllabusBeat is one move on the map.
type SyllabusBeat struct {
	// Show is a syllabusShows name.
	Show string `json:"show"`
	// At indexes SyllabusSpec.Modules, for a "stop" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults an unknown direction to visiting a stop, which is what
// most beats of this template do.
func (b SyllabusBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if syllabusShows[s] {
		return s
	}
	return "stop"
}

func normalizeSyllabusPlan(p *SnippetPlan) {
	s := p.Syllabus
	if s == nil {
		return
	}
	modules := make([]SyllabusModule, 0, len(s.Modules))
	for _, m := range s.Modules {
		m.Label = clampWords(collapseSpaces(m.Label), maxSyllabusLabelWords)
		m.Sub = clampWords(collapseSpaces(m.Sub), maxSyllabusSubWords)
		if len(modules) < maxSyllabusModules {
			modules = append(modules, m)
		}
	}
	s.Modules = modules

	for i := range p.Beats {
		b := p.Beats[i].Syllabus
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "stop" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(s.Modules); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateSyllabusPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Syllabus: true}); err != nil {
		return err
	}

	s := p.Syllabus
	if s == nil {
		return fmt.Errorf("the plan has no syllabus — this template is the course map, so the modules and the current position are the clip")
	}
	if n := len(s.Modules); n < minSyllabusModules || n > maxSyllabusModules {
		return fmt.Errorf("the map has %d module(s), want %d-%d. Below four the route is a before/after rather than a journey — that hand-off is bridge's picture; past eight the stops compress until their labels collide",
			n, minSyllabusModules, maxSyllabusModules)
	}
	if s.Current < 0 || s.Current >= len(s.Modules) {
		return fmt.Errorf("the current position is module %d, but the map has modules 0-%d. The landing beat has to arrive somewhere that exists",
			s.Current, len(s.Modules)-1)
	}
	for i, m := range s.Modules {
		if strings.TrimSpace(m.Label) == "" {
			return fmt.Errorf("module %d has no label. A stop with no name is a dot on a line, and the viewer cannot tell what part of the course it stands for", i)
		}
	}

	if p.Beats[0].Syllabus == nil || p.Beats[0].Syllabus.ResolvedShow() != "map" {
		return fmt.Errorf("beat %q does not open on the map. A stop lighting up on a route the viewer has not seen whole is a location without a journey — show the whole route dimmed first",
			p.Beats[0].ID)
	}

	visited := map[int]bool{}
	heres := 0
	hereAt := -1
	for i, b := range p.Beats {
		if b.Syllabus == nil {
			return fmt.Errorf("beat %q has no syllabus direction — every beat shows the map, visits a stop, or lands on the current module", b.ID)
		}
		switch b.Syllabus.ResolvedShow() {
		case "map":
			if i != 0 {
				return fmt.Errorf("beat %q shows the whole map again part-way through. The map is the opener; re-showing it resets the picture the stops have been building on", b.ID)
			}
		case "stop":
			at := b.Syllabus.At
			if at < 0 || at >= len(s.Modules) {
				return fmt.Errorf("beat %q visits module %d, which does not exist", b.ID, at)
			}
			if visited[at] {
				return fmt.Errorf("beat %q visits module %d (%q) a second time. The route already orders the modules, so a repeat visit spends a beat saying nothing new — and on this budget that beat came out of some other module's only appearance",
					b.ID, at, s.Modules[at].Label)
			}
			visited[at] = true
		case "here":
			heres++
			hereAt = i
		}
	}
	if heres != 1 {
		return fmt.Errorf("there are %d \"here\" beats, want exactly 1. The template's job is to land the viewer somewhere, and a clip that lands twice has two presents", heres)
	}
	if hereAt != len(p.Beats)-1 {
		return fmt.Errorf("beat %q lands on the current module and then the clip keeps going. The arrival is the ending or it is not an arrival — make \"here\" the last beat",
			p.Beats[hereAt].ID)
	}
	return nil
}

// syllabusScenes lays the clip out as ONE scene. The route persists; the steps
// only say which stop is lifted and, at the landing, which are ticked.
func syllabusScenes(in SnippetSceneInput) ([]Scene, error) {
	s := in.Plan.Syllabus
	if s == nil {
		return nil, fmt.Errorf("the plan has no syllabus")
	}

	modules := make([]map[string]any, len(s.Modules))
	for i, m := range s.Modules {
		modules[i] = map[string]any{"label": m.Label, "sub": m.Sub}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Syllabus == nil {
			return nil, fmt.Errorf("beat %q has no syllabus direction", beat.ID)
		}
		show := beat.Syllabus.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "stop" {
			step["at"] = beat.Syllabus.At
		}
		// The ticks belong to the landing: everything before the current stop
		// stamps done. Computed here so the renderer never re-derives "before"
		// from the current index.
		ticked := []int{}
		if show == "here" {
			for j := 0; j < s.Current && j < len(s.Modules); j++ {
				ticked = append(ticked, j)
			}
		}
		sort.Ints(ticked)
		step["ticked"] = ticked
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneSyllabus,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":   in.Plan.Title,
			"modules": modules,
			"current": s.Current,
			"steps":   steps,
		}),
	}}, nil
}
