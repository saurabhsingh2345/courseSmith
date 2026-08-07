package pipeline

// The layers template: the layer cake, and the line something crosses.
//
// Computing is taught in strata — hardware, firmware, kernel, userland; the
// seven OSI layers; user space and kernel space — and the stack drawing is so
// common that it has stopped meaning anything. Every course draws four boxes on
// top of each other and calls it an explanation. It is not, because the drawing
// leaves out the only part that matters: there is usually ONE line in the stack
// that is different from all the others, and crossing it costs something. A
// syscall is not "calling down a layer", it is a trap. Encapsulation is not
// "going down the stack", it is a payload getting wrapped.
//
// So this template draws the strata as full-width bars, and it draws the
// BOUNDARY as a brighter dashed rule with a name on it, and it spends a beat
// sending a payload chip through that rule. Everything else in the picture
// exists to make that one crossing legible.
//
// The validator therefore guards the boundary above all. A cross beat with no
// boundary declared is the failure that produces a chip sliding through empty
// space, so it is rejected — and rejected by telling the model to drop the
// cross beat, not to invent a boundary. That direction is deliberate: a model
// told "set a boundary" will pick one at random and ship a picture claiming a
// privilege line exists where it does not, which is exactly the confident wrong
// diagram this family was built to prevent. A missing beat is a smaller lie
// than an invented boundary.
//
// Each stratum may be focused at most once. Coming back to a layer means the
// clip is arguing rather than describing, and the stack picture is a
// description; a template that wanted an argument is bridge or versus.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "layers",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The layer cake",
		Description: "Strata stacked full-width with one named boundary running through them, and a payload that crosses it on screen. Reach for it when the subject is what each level holds and what it costs to cross the line between two of them — syscalls, encapsulation, user space against kernel space.",
		Example:     "The syscall line: what user space cannot do for itself",
		PromptFile:  snippetLayersTemplateName,
		NeedsCode:   false,
		// The stack, two or three focused strata, the crossing and the whole:
		// under thirty-five seconds the crossing is over before it registers.
		MinTargetSec:     35,
		DefaultTargetSec: 50,
		// Opener + up to five focused strata + a crossing + closer. Past eight
		// the same bars are being re-lit.
		MaxBeats: 8,
		// A beat here is a SHOT — one bar lighting, or one chip crossing — not
		// a step in an argument.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Layers: true},
		OwnsPlan:          planFields{Layers: true},
		Normalize:         normalizeLayersPlan,
		Validate:          validateLayersPlan,
		Scenes:            layersScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":                 strings.Join(MetricRoles(), ", "),
				"Shows":                 strings.Join(LayersShows(), ", "),
				"MinStrata":             minLayersStrata,
				"MaxStrata":             maxLayersStrata,
				"MaxLabelWords":         maxLayersLabelWords,
				"MaxHoldsWords":         maxLayersHoldsWords,
				"MaxBoundaryLabelWords": maxLayersBoundaryLabelWords,
			}
		},
	})
}

const snippetLayersTemplateName = "snippet_layers.tmpl"

const (
	// Two bars is an over-and-under, which needs no diagram. Three is the
	// least that reads as a stack with a middle.
	minLayersStrata = 3
	// Seven full-width bars with 2px gaps fill the stage at a height that
	// still holds a label and its contents on one line; an eighth does not.
	maxLayersStrata = 7
	// A stratum is named — "kernel", "user space", "physical layer" — and
	// three words covers the longest of the standard names.
	maxLayersLabelWords = 3
	// Holds sits right-aligned inside the bar, opposite the label. Eight
	// words is what fits there before the two collide.
	maxLayersHoldsWords = 8
	// The boundary's name rides at the right-hand end of the dashed rule —
	// "the syscall line", "privilege boundary". Four words is the rule's
	// width at the type size it is set in.
	maxLayersBoundaryLabelWords = 4

	// noLayersBoundary is the sentinel for a stack that has no special line
	// in it. Some stacks genuinely do not, and forcing one is worse.
	noLayersBoundary = -1
)

// layersShows is the closed vocabulary of what a beat does.
var layersShows = map[string]bool{
	// Every stratum as a bar, the shape of the whole stack. The opener.
	"stack": true,
	// The stratum at At comes forward with what it holds.
	"stratum": true,
	// A payload chip crosses the boundary, shrinking as it passes through.
	"cross": true,
	// The whole stack lit with the boundary named. The closer.
	"whole": true,
}

// LayersShows returns the beat vocabulary sorted.
func LayersShows() []string {
	out := make([]string, 0, len(layersShows))
	for k := range layersShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// LayersStratum is one band of the stack.
type LayersStratum struct {
	// Label is the level's name — "kernel", "firmware".
	Label string `json:"label"`
	// Holds is what lives at this level.
	Holds string `json:"holds"`
}

// LayersSpec is the stack and its one special line. On the plan because the
// bars are up for the whole clip and the beats only light parts of them.
type LayersSpec struct {
	// Strata are the bands, TOP FIRST — index 0 is drawn at the top.
	Strata []LayersStratum `json:"strata"`
	// Boundary is the index above which is one world and below another: the
	// rule is drawn under stratum Boundary. -1 for a stack with no such line.
	Boundary int `json:"boundary"`
	// BoundaryLabel names the line — "the syscall line".
	BoundaryLabel string `json:"boundaryLabel,omitempty"`
}

// LayersBeat is one shot: which state of the stack this beat shows.
type LayersBeat struct {
	// Show is a layersShows name.
	Show string `json:"show"`
	// At is the zero-based stratum index, used only when Show is "stratum".
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a stratum —
// the workhorse state most beats of this template are in.
func (b LayersBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if layersShows[s] {
		return s
	}
	return "stratum"
}

func normalizeLayersPlan(p *SnippetPlan) {
	l := p.Layers
	if l == nil {
		return
	}
	strata := make([]LayersStratum, 0, len(l.Strata))
	for _, s := range l.Strata {
		s.Label = clampWords(collapseSpaces(s.Label), maxLayersLabelWords)
		s.Holds = clampWords(collapseSpaces(s.Holds), maxLayersHoldsWords)
		if s.Label == "" {
			continue
		}
		strata = append(strata, s)
	}
	if len(strata) > maxLayersStrata {
		strata = strata[:maxLayersStrata]
	}
	l.Strata = strata
	l.BoundaryLabel = clampWords(collapseSpaces(l.BoundaryLabel), maxLayersBoundaryLabelWords)

	// Anything that is not a usable index becomes "no boundary" rather than a
	// guess: the whole point of the sentinel is that inventing a privilege
	// line is the worse error.
	if l.Boundary < noLayersBoundary || l.Boundary > len(l.Strata)-2 {
		l.Boundary = noLayersBoundary
	}

	for i := range p.Beats {
		b := p.Beats[i].Layers
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.At < 0 {
			b.At = 0
		}
		if n := len(l.Strata); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateLayersPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Layers: true}); err != nil {
		return err
	}

	l := p.Layers
	if l == nil {
		return fmt.Errorf("the plan has no stack — this template is a set of strata with a line running through them, so without strata there is nothing to draw")
	}
	if n := len(l.Strata); n < minLayersStrata || n > maxLayersStrata {
		return fmt.Errorf("the stack has %d strata, want %d-%d. Two bars is an over-and-under that needs no diagram, and past %d the bars lose the height their labels need",
			n, minLayersStrata, maxLayersStrata, maxLayersStrata)
	}
	for i, s := range l.Strata {
		if strings.TrimSpace(s.Label) == "" {
			return fmt.Errorf("stratum %d has no label. Every band is named, top first — index 0 is the top of the stack", i)
		}
		if n := len(strings.Fields(s.Label)); n > maxLayersLabelWords {
			return fmt.Errorf("the stratum label %q is %d words and a bar holds %d. A level has a name — what lives there goes in \"holds\"", s.Label, n, maxLayersLabelWords)
		}
		if strings.TrimSpace(s.Holds) == "" {
			return fmt.Errorf("stratum %q says nothing about what it holds. A named empty bar is the stack drawing this template exists to replace — say what lives at this level", s.Label)
		}
		if n := len(strings.Fields(s.Holds)); n > maxLayersHoldsWords {
			return fmt.Errorf("stratum %q holds %d words' worth, and the space opposite its label fits %d. Name the inhabitants, not their behaviour", s.Label, n, maxLayersHoldsWords)
		}
	}

	// The boundary is drawn UNDER stratum Boundary, so the last index cannot
	// carry one: there would be nothing beneath the line.
	if l.Boundary != noLayersBoundary && (l.Boundary < 0 || l.Boundary > len(l.Strata)-2) {
		return fmt.Errorf("the boundary index is %d, and with %d strata it has to be 0 to %d, or -1 for a stack with no special line. The rule is drawn UNDER the stratum you name, so the bottom stratum cannot carry one — there would be nothing beneath it",
			l.Boundary, len(l.Strata), len(l.Strata)-2)
	}
	if l.Boundary != noLayersBoundary && strings.TrimSpace(l.BoundaryLabel) == "" {
		return fmt.Errorf("the boundary between %q and %q has no name. An unnamed brighter rule reads as a styling accident — call it what it is, in at most %d words",
			l.Strata[l.Boundary].Label, l.Strata[l.Boundary+1].Label, maxLayersBoundaryLabelWords)
	}
	if n := len(strings.Fields(l.BoundaryLabel)); n > maxLayersBoundaryLabelWords {
		return fmt.Errorf("the boundary label %q is %d words and the rule carries %d at the end of its run", l.BoundaryLabel, n, maxLayersBoundaryLabelWords)
	}

	for _, b := range p.Beats {
		if b.Layers == nil {
			return fmt.Errorf("beat %q has no layers direction — every beat shows one state of the stack, so every beat needs a {\"show\": ...}", b.ID)
		}
	}
	if p.Beats[0].Layers.ResolvedShow() != "stack" {
		return fmt.Errorf("beat %q does not open on the stack. A bar coming forward before the viewer has seen the shape it belongs to is a rectangle with a word on it — open with {\"show\": \"stack\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Layers.ResolvedShow() != "whole" {
		return fmt.Errorf("the clip does not close on the whole stack. The final frame is the one the viewer keeps, and it has to show every level and the line at once — end with {\"show\": \"whole\"}")
	}

	focused := map[int]string{}
	for i, b := range p.Beats {
		show := b.Layers.ResolvedShow()
		switch show {
		case "stack":
			if i != 0 {
				return fmt.Errorf("beat %q returns to the bare stack part-way through. The bars never go away, so there is nothing to come back to — this beat is a stratum, a cross or the closer", b.ID)
			}
		case "whole":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q lights the whole stack before the end. \"whole\" is the closer, and after every level is lit at once there is nothing a later beat can add", b.ID)
			}
		case "stratum":
			at := b.Layers.At
			if at < 0 || at >= len(l.Strata) {
				return fmt.Errorf("beat %q focuses stratum %d, and the stack has %d strata numbered 0 to %d, top first. The index is zero-based",
					b.ID, at, len(l.Strata), len(l.Strata)-1)
			}
			if prev, seen := focused[at]; seen {
				return fmt.Errorf("beat %q focuses %q again, after beat %q already did. Each level gets one beat — coming back to a layer means the clip is arguing rather than describing, and the stack picture is a description",
					b.ID, l.Strata[at].Label, prev)
			}
			focused[at] = b.ID
		case "cross":
			// THE BOUNDARY RULE. Told to invent one, a model picks a line at
			// random and ships a picture claiming a privilege boundary where
			// none exists. A missing beat is the smaller lie.
			if l.Boundary == noLayersBoundary {
				return fmt.Errorf("beat %q sends a payload across the boundary, and the stack declares boundary -1, meaning it has no such line. Do NOT invent one to satisfy this beat — a made-up privilege line is a confidently wrong diagram. Drop the \"cross\" beat, or rebuild the stack around a line that genuinely exists in the subject", b.ID)
			}
		}
	}
	return nil
}

// layersScenes lays the clip out as ONE scene. Which side of the line each
// stratum sits on, which bars have been focused and whether the crossing has
// happened are all settled here; the component draws the state it is handed.
func layersScenes(in SnippetSceneInput) ([]Scene, error) {
	l := in.Plan.Layers
	if l == nil {
		return nil, fmt.Errorf("the plan has no stack")
	}

	strata := make([]map[string]any, len(l.Strata))
	for i, s := range l.Strata {
		strata[i] = map[string]any{
			"label": s.Label,
			"holds": s.Holds,
			// Which world this band belongs to. Precomputed so the component
			// never has to reason about the sentinel.
			"above": l.Boundary != noLayersBoundary && i <= l.Boundary,
		}
	}

	seen := map[int]bool{}
	crossed := false
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Layers == nil {
			return nil, fmt.Errorf("beat %q has no layers direction", beat.ID)
		}
		show := beat.Layers.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "stratum":
			at := beat.Layers.At
			if at >= 0 && at < len(l.Strata) {
				seen[at] = true
				step["at"] = at
			}
		case "cross":
			crossed = true
		}
		lit := make([]int, 0, len(seen))
		for k := range seen {
			lit = append(lit, k)
		}
		sort.Ints(lit)
		step["lit"] = lit
		step["crossed"] = crossed
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneLayers,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":         in.Plan.Title,
			"strata":        strata,
			"boundary":      l.Boundary,
			"boundaryLabel": l.BoundaryLabel,
			"steps":         steps,
		}),
	}}, nil
}
