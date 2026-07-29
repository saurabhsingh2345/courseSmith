package pipeline

// The stack template: which tool does what, and where the handoff is.
//
// Every no-code course has to answer "so which of these do I actually use for
// what" at least three times, and until now the catalog answered it badly.
// `compare` weighs two things against each other, which is the wrong question —
// Airtable and Make are not competing, they are doing different jobs. `flow`
// draws a graph with traffic on it, which says how a request travels and not
// which *tier* a tool belongs to. And a stack rendered as a flow comes out as a
// straight chain, which that template rejects for good reasons of its own.
//
// So the structure here is layers, because the claim a stack makes is vertical:
// this tool is your data, that one is your glue, that one is what people
// actually look at. Position in the band IS the information, the way position
// along the spine is the information in a timeline — and a tool's neighbours in
// its band are its alternatives, which is the comparison a viewer actually wants
// and gets for free from the layout.
//
// The walk goes down the stack and only down. A course that hops between tiers
// while explaining them is describing a request's path, and `flow` draws paths.

import (
	"fmt"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:             "stack",
		Category:         CatDecisions,
		Title:            "Tool stack",
		Description:      "The layers of a build — what each tier is for and which tools live there — walked top to bottom.",
		Example:          "The four tools behind a no-code job board, and what each one is actually for",
		PromptFile:       snippetStackTemplateName,
		NeedsCode:        false,
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		Owns:             beatFields{Stack: true},
		OwnsPlan:         planFields{Stack: true},
		Normalize:        normalizeStackPlan,
		Validate:         validateStackPlan,
		Scenes:           stackScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Icons":         strings.Join(PointIconNames(), ", "),
				"MinLayers":     minStackLayers,
				"MaxLayers":     maxStackLayers,
				"MaxPerLayer":   maxToolsPerLayer,
				"MaxNameWords":  maxStackNameWords,
				"MaxRoleWords":  maxStackRoleWords,
				"MaxToolWords":  maxStackToolWords,
				"MaxToolNoteWd": maxStackToolNoteWords,
			}
		},
	})
}

const snippetStackTemplateName = "snippet_stack.tmpl"

// Stack capacity. Four bands across the stage leaves 150px each, which is what a
// tool card with a name and a note needs; a fifth turns the diagram into a set
// of strips. Three tools in a band is the point at which the cards stop being
// wide enough to name a tool and say what it is for.
const (
	minStackLayers   = 3
	maxStackLayers   = 4
	maxToolsPerLayer = 3

	maxStackNameWords     = 3
	maxStackRoleWords     = 9
	maxStackToolWords     = 2
	maxStackToolNoteWords = 6
)

// StackSpec is the whole stack. On the plan rather than per-beat for the reason
// the timeline's milestones are: the arrangement is the subject of the clip and
// the beats only walk it.
type StackSpec struct {
	// Layers are the tiers, top of the stack first — the one closest to the
	// person using the thing, down to the one furthest from them.
	Layers []StackLayer `json:"layers"`
}

// StackLayer is one tier.
type StackLayer struct {
	// Name is the tier — "Frontend", "Automation", "Data", "AI".
	Name string `json:"name"`
	// Role is what this tier is FOR, in one short line. Shown while the layer
	// is the current one.
	Role string `json:"role"`
	// Tools are the products that live in this tier. More than one reads as
	// alternatives, which is what a viewer wants and what the layout gives for
	// free.
	Tools []StackTool `json:"tools"`
}

// StackTool is one product in a tier.
type StackTool struct {
	// Name is the product — "Airtable", "Make", "Softr".
	Name string `json:"name"`
	// Icon is a PointIconNames name drawn in the card's chip.
	Icon string `json:"icon,omitempty"`
	// Note is what this one in particular is good at. Optional.
	Note string `json:"note,omitempty"`
}

// ResolvedIcon returns the icon drawn in the tool's chip.
func (t StackTool) ResolvedIcon() string {
	if icon := normalizePointIconName(t.Icon); icon != "" {
		return icon
	}
	return "box"
}

// StackBeat says which layer this beat is standing on, or that it is showing
// the whole stack.
type StackBeat struct {
	// At indexes StackSpec.Layers.
	At int `json:"at"`
	// Whole marks the closing beat showing the assembled stack with nothing
	// singled out.
	Whole bool `json:"whole,omitempty"`
}

func normalizeStackPlan(p *SnippetPlan) {
	s := p.Stack
	if s == nil {
		return
	}
	for i := range s.Layers {
		l := &s.Layers[i]
		l.Name = clampWords(collapseSpaces(l.Name), maxStackNameWords)
		l.Role = clampWords(collapseSpaces(l.Role), maxStackRoleWords)
		tools := make([]StackTool, 0, len(l.Tools))
		for _, t := range l.Tools {
			t.Name = clampWords(collapseSpaces(t.Name), maxStackToolWords)
			t.Note = clampWords(collapseSpaces(t.Note), maxStackToolNoteWords)
			t.Icon = t.ResolvedIcon()
			// A card with no name is a chip and a blank line. Dropping it is
			// the only repair; inventing a product name would be a claim.
			if t.Name != "" && len(tools) < maxToolsPerLayer {
				tools = append(tools, t)
			}
		}
		l.Tools = tools
	}
	for i := range p.Beats {
		if b := p.Beats[i].Stack; b != nil && b.At < 0 {
			b.At, b.Whole = 0, true
		}
	}
}

func validateStackPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Stack: true}); err != nil {
		return err
	}

	s := p.Stack
	if s == nil {
		return fmt.Errorf("the plan has no stack — this template is a set of tiers, walked from the top down")
	}
	if n := len(s.Layers); n < minStackLayers || n > maxStackLayers {
		return fmt.Errorf("the stack has %d layers, want %d-%d — two tiers is a before-and-after of one handoff, and compare tells that better",
			n, minStackLayers, maxStackLayers)
	}
	seenLayer := map[string]bool{}
	seenTool := map[string]bool{}
	for i, l := range s.Layers {
		if strings.TrimSpace(l.Name) == "" {
			return fmt.Errorf("layer %d has no name — say which tier it is", i)
		}
		key := strings.ToLower(strings.TrimSpace(l.Name))
		if seenLayer[key] {
			return fmt.Errorf("layer %d repeats the name %q — each tier is a different job", i, l.Name)
		}
		seenLayer[key] = true
		if strings.TrimSpace(l.Role) == "" {
			return fmt.Errorf("layer %q has no role — a tier nobody explains is a band with logos in it", l.Name)
		}
		if n := len(l.Tools); n < 1 || n > maxToolsPerLayer {
			return fmt.Errorf("layer %q holds %d tools, want 1-%d", l.Name, n, maxToolsPerLayer)
		}
		for _, t := range l.Tools {
			if strings.TrimSpace(t.Name) == "" {
				return fmt.Errorf("layer %q has a tool with no name", l.Name)
			}
			tk := strings.ToLower(strings.TrimSpace(t.Name))
			// The same product in two tiers is the mistake this template exists
			// to prevent — the whole claim is that a tool has *a* job.
			if seenTool[tk] {
				return fmt.Errorf("%q appears in two layers. A stack says each tool has one job; if it genuinely spans two tiers, put it in the one it is mostly used for and say so in the narration", t.Name)
			}
			seenTool[tk] = true
		}
	}

	visited := map[int]bool{}
	last := -1
	sawWhole := false
	for _, b := range p.Beats {
		if b.Stack == nil {
			return fmt.Errorf("beat %q has no stack direction — every beat is standing on a layer or showing the whole stack", b.ID)
		}
		if b.Stack.Whole {
			sawWhole = true
			continue
		}
		if b.Stack.At < 0 || b.Stack.At >= len(s.Layers) {
			return fmt.Errorf("beat %q stands on layer %d, which does not exist", b.ID, b.Stack.At)
		}
		if b.Stack.At < last {
			return fmt.Errorf("beat %q goes back up to layer %d after %d. The stack is walked downward — a clip that hops between tiers is describing a request's path, and the flow template draws paths",
				b.ID, b.Stack.At, last)
		}
		if visited[b.Stack.At] {
			return fmt.Errorf("beat %q stands on layer %d again; each tier gets one beat", b.ID, b.Stack.At)
		}
		visited[b.Stack.At] = true
		last = b.Stack.At
	}
	if len(visited) != len(s.Layers) {
		return fmt.Errorf("%d of the %d layers are never narrated — a tier nobody explains is a band with logos in it",
			len(s.Layers)-len(visited), len(s.Layers))
	}
	if !sawWhole {
		return fmt.Errorf("no beat shows the whole stack. Close with a beat carrying \"whole\": true — seeing the tiers together is what makes it a stack rather than four separate tool reviews")
	}
	if lastBeat := p.Beats[len(p.Beats)-1].Stack; lastBeat != nil && !lastBeat.Whole {
		return fmt.Errorf("the clip ends on one tier; end on the whole stack instead (\"whole\": true)")
	}
	return nil
}

// stackScenes lays the clip out as ONE scene: every tier is on screen for the
// whole clip and the beats only move which one is lit.
func stackScenes(in SnippetSceneInput) ([]Scene, error) {
	s := in.Plan.Stack
	if s == nil {
		return nil, fmt.Errorf("the plan has no stack")
	}

	layers := make([]map[string]any, len(s.Layers))
	for i, l := range s.Layers {
		tools := make([]map[string]any, len(l.Tools))
		for j, t := range l.Tools {
			tools[j] = map[string]any{
				"name": t.Name,
				"icon": t.ResolvedIcon(),
				"note": t.Note,
			}
		}
		layers[i] = map[string]any{
			"name":  l.Name,
			"role":  l.Role,
			"tools": tools,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Stack == nil {
			return nil, fmt.Errorf("beat %q has no stack direction", beat.ID)
		}
		step := map[string]any{"startMs": startMs, "endMs": endMs}
		if beat.Stack.Whole {
			step["whole"] = true
		} else {
			step["at"] = beat.Stack.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneStack,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":  in.Plan.Title,
			"layers": layers,
			"steps":  steps,
		},
	}}, nil
}
