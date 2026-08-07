package pipeline

// The machine template: inside the box, one part at a time.
//
// A career-switcher's first real gap is not an algorithm, it is the object on
// their own desk. They have used a computer for twenty years and could not say
// what is inside the case or what any of it is for — and every explanation
// they have met either dumps a motherboard photo (all detail, no order) or a
// bulleted list (all order, no object). The catalog could draw flows, stacks
// and traces; it had nothing that draws a THING and walks its parts.
//
// So the frame is a chassis holding stylized part blocks, all visible from the
// first beat, and each beat lifts exactly one part toward the camera with its
// one-line job. The whole stays on screen the entire clip: a part explained
// out of context is a vocabulary word, a part lifted out of a machine the
// viewer has already seen whole is a component.
//
// The validators keep the walk honest.
//
// **The box is seen whole before any part lifts.** A part rising out of a
// machine nobody has seen assembled is a floating rectangle, so the first beat
// is always "whole" and only the first.
//
// **Every part gets exactly one lift.** A part drawn in the chassis but never
// spoken about is a mystery the clip planted on purpose; a part lifted twice
// is a clip that forgot where it was. One beat per part, no skips, no repeats.
//
// **The clip closes on "fit".** The parts go back and the box lights as one
// object — the point of the walk is that these jobs compose into the machine
// the viewer started with, and a clip that ends mid-lift never says so.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "machine",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "Inside the box",
		Description: "A labeled hardware diagram: a chassis holding stylized part blocks, each beat lifting one part toward the camera with its one-line job. Reach for it when the subject is a physical machine and its components — what is inside a PC, a phone, a hard drive, a router.",
		Example:     "What is actually inside a desktop PC, part by part",
		PromptFile:  snippetMachineTemplateName,
		NeedsCode:   false,
		// The whole, one lift per part, and the refit: below 35 seconds the word
		// budget cannot fund an opener, four parts and a closer at a spoken pace.
		MinTargetSec:     35,
		DefaultTargetSec: 60,
		// Opener + up to six lifts + closer. A beat here is a SHOT — one part
		// forward — so the cap is about how many lifts a viewer tracks, not how
		// long an argument runs.
		MaxBeats: 8,
		// A beat is a shot, not an argument step: 28 words is nine seconds of one
		// part held forward, which is as long as a lifted block stays interesting.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Machine: true},
		OwnsPlan:          planFields{Machine: true},
		Normalize:         normalizeMachinePlan,
		Validate:          validateMachinePlan,
		Scenes:            machineScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":           strings.Join(MetricRoles(), ", "),
				"Shows":           strings.Join(MachineShows(), ", "),
				"Sizes":           strings.Join(MachineSizes(), ", "),
				"MinParts":        minMachineParts,
				"MaxParts":        maxMachineParts,
				"MaxChassisWords": maxMachineChassisWords,
				"MaxLabelWords":   maxMachineLabelWords,
				"MaxJobWords":     maxMachineJobWords,
			}
		},
	})
}

const snippetMachineTemplateName = "snippet_machine.tmpl"

const (
	// Below four parts the box is not a machine, it is a diagram of a couple of
	// words; past seven the blocks shrink until a lifted one has nowhere to
	// stand out from, and the one-beat-per-part rule would need more beats than
	// any runtime funds.
	minMachineParts = 4
	maxMachineParts = 7

	// The chassis name is set large over the box; four words is a name, more is
	// a sentence wearing a nameplate.
	maxMachineChassisWords = 4
	// A part label sits inside a block a few hundred pixels wide.
	maxMachineLabelWords = 3
	// The job is one line under the lifted part. Ten words is a job; more is a
	// paragraph the viewer is given nine seconds to read.
	maxMachineJobWords = 10
)

// machineShows is the closed vocabulary of what a beat does.
var machineShows = map[string]bool{
	// Every part in place, dimmed. The opener.
	"whole": true,
	// Part At lifts toward the camera with its job line.
	"part": true,
	// Everything back in place, lit. The closer.
	"fit": true,
}

// MachineShows returns the beat vocabulary sorted.
func MachineShows() []string {
	out := make([]string, 0, len(machineShows))
	for k := range machineShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// machineSizes is a part's relative visual weight. A closed vocabulary because
// the renderer packs blocks by it, and an invented size has no width.
var machineSizes = map[string]bool{"large": true, "medium": true, "small": true}

// MachineSizes returns the size vocabulary sorted.
func MachineSizes() []string {
	out := make([]string, 0, len(machineSizes))
	for k := range machineSizes {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// MachineSpec is the box and what is inside it. On the plan because the
// chassis is one object that persists for the whole clip.
type MachineSpec struct {
	// Chassis is what the box is — "a desktop PC", "a phone".
	Chassis string `json:"chassis"`
	// Parts are the blocks inside, drawn in this order.
	Parts []MachinePart `json:"parts"`
}

// MachinePart is one block in the chassis.
type MachinePart struct {
	// Label names the part — "CPU", "power supply".
	Label string `json:"label"`
	// Job is the one-line answer to "what is this for".
	Job string `json:"job"`
	// Size is the part's relative visual weight: a machineSizes name.
	Size string `json:"size,omitempty"`
}

// ResolvedSize returns the part's visual weight, defaulting to medium — the
// honest weight for a part the model made no claim about.
func (p MachinePart) ResolvedSize() string {
	s := strings.ToLower(strings.TrimSpace(p.Size))
	if machineSizes[s] {
		return s
	}
	return "medium"
}

// MachineBeat is one move: which part this beat lifts.
type MachineBeat struct {
	// Show is a machineShows name.
	Show string `json:"show"`
	// At indexes MachineSpec.Parts, for a "part" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to a lift,
// which is what most beats of this template do.
func (b MachineBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if machineShows[s] {
		return s
	}
	return "part"
}

func normalizeMachinePlan(p *SnippetPlan) {
	m := p.Machine
	if m == nil {
		return
	}
	m.Chassis = clampWords(collapseSpaces(m.Chassis), maxMachineChassisWords)

	parts := make([]MachinePart, 0, len(m.Parts))
	for _, part := range m.Parts {
		part.Label = clampWords(collapseSpaces(part.Label), maxMachineLabelWords)
		part.Job = clampWords(collapseSpaces(part.Job), maxMachineJobWords)
		part.Size = part.ResolvedSize()
		if len(parts) < maxMachineParts {
			parts = append(parts, part)
		}
	}
	m.Parts = parts

	for i := range p.Beats {
		b := p.Beats[i].Machine
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "part" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(m.Parts); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateMachinePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Machine: true}); err != nil {
		return err
	}

	m := p.Machine
	if m == nil {
		return fmt.Errorf("the plan has no machine — this template is a chassis full of parts and a walk through their jobs, so the box is the clip")
	}
	if strings.TrimSpace(m.Chassis) == "" {
		return fmt.Errorf("the machine has no chassis name. The box has to say what it is — \"a desktop PC\", \"a phone\" — or the parts are floating in an anonymous rectangle")
	}
	if n := len(m.Parts); n < minMachineParts || n > maxMachineParts {
		return fmt.Errorf("the chassis holds %d part(s), want %d-%d. Below %d the box is a couple of labels rather than a machine; past %d the blocks shrink until a lifted one has nothing to stand out from — and every part needs its own beat, which no runtime funds past %d",
			n, minMachineParts, maxMachineParts, minMachineParts, maxMachineParts, maxMachineParts)
	}
	for i, part := range m.Parts {
		if strings.TrimSpace(part.Label) == "" {
			return fmt.Errorf("part %d has no label — an unlabeled block inside a machine is a mystery the clip planted on purpose", i)
		}
		if strings.TrimSpace(part.Job) == "" {
			return fmt.Errorf("part %d (%q) has no job. The lift IS the job line: a part raised toward the camera with nothing to say is a rectangle doing a bow", i, part.Label)
		}
		if s := strings.ToLower(strings.TrimSpace(part.Size)); s != "" && !machineSizes[s] {
			return fmt.Errorf("part %d (%q) has size %q, which is not one of: %s. The renderer packs blocks by their weight, so an invented size has no width",
				i, part.Label, part.Size, strings.Join(MachineSizes(), ", "))
		}
	}

	// The box is seen whole before any part lifts, and only then.
	if p.Beats[0].Machine == nil || p.Beats[0].Machine.ResolvedShow() != "whole" {
		return fmt.Errorf("beat %q does not open on the whole machine. A part rising out of a box nobody has seen assembled is a floating rectangle — the first beat shows every part in place",
			p.Beats[0].ID)
	}
	last := p.Beats[len(p.Beats)-1]
	if last.Machine == nil || last.Machine.ResolvedShow() != "fit" {
		return fmt.Errorf("beat %q does not close on the refit. The point of the walk is that the jobs compose back into one machine, so the last beat puts everything back in place, lit",
			last.ID)
	}

	lifted := map[int]bool{}
	for i, b := range p.Beats {
		if b.Machine == nil {
			return fmt.Errorf("beat %q has no machine direction — every beat shows the whole box, lifts one part, or fits everything back", b.ID)
		}
		switch b.Machine.ResolvedShow() {
		case "whole":
			if i != 0 {
				return fmt.Errorf("beat %q shows the machine whole again part-way through. The whole is the opener; going back to it mid-walk un-lifts a part nobody put down", b.ID)
			}
		case "fit":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q fits the parts back before the walk is over. The refit is the closer — a box reassembled with lifts still to come has to be taken apart again", b.ID)
			}
		case "part":
			if b.Machine.At < 0 || b.Machine.At >= len(m.Parts) {
				return fmt.Errorf("beat %q lifts part %d, which does not exist — the chassis holds parts 0-%d", b.ID, b.Machine.At, len(m.Parts)-1)
			}
			if lifted[b.Machine.At] {
				return fmt.Errorf("beat %q lifts part %d (%q) again. Each part gets exactly one lift — a second one says the clip forgot where it was",
					b.ID, b.Machine.At, m.Parts[b.Machine.At].Label)
			}
			lifted[b.Machine.At] = true
		}
	}
	if len(lifted) != len(m.Parts) {
		return fmt.Errorf("%d of the %d parts are never lifted. A part drawn in the chassis but never spoken about is a mystery the clip planted on purpose — give each one a beat, or use fewer parts",
			len(m.Parts)-len(lifted), len(m.Parts))
	}
	return nil
}

// machineScenes lays the clip out as ONE scene. The chassis persists; the
// steps only say which part is forward and which have already been visited.
func machineScenes(in SnippetSceneInput) ([]Scene, error) {
	m := in.Plan.Machine
	if m == nil {
		return nil, fmt.Errorf("the plan has no machine")
	}

	parts := make([]map[string]any, len(m.Parts))
	for i, part := range m.Parts {
		parts[i] = map[string]any{
			"label": part.Label,
			"job":   part.Job,
			"size":  part.ResolvedSize(),
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	// Which parts have been lifted by each beat, accumulated in Go so the
	// renderer never replays the beat list to find out.
	visited := map[int]bool{}
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Machine == nil {
			return nil, fmt.Errorf("beat %q has no machine direction", beat.ID)
		}
		show := beat.Machine.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "part" {
			visited[beat.Machine.At] = true
			step["at"] = beat.Machine.At
		}
		seen := make([]int, 0, len(visited))
		for at := range visited {
			seen = append(seen, at)
		}
		sort.Ints(seen)
		step["visited"] = seen
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneMachine,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":   in.Plan.Title,
			"chassis": m.Chassis,
			"parts":   parts,
			"steps":   steps,
		}),
	}}, nil
}
