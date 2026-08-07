package pipeline

// The relay template: the baton pass.
//
// Some things in computing are not a system of parts, they are a SEQUENCE of
// hand-offs, and the difference matters. Power, BIOS, POST, the bootloader, the
// kernel, the login prompt: each one does a small job and then gives control
// away, permanently, to the next. Nothing loops back. Nothing runs alongside.
// The whole answer to "what happens between the power button and the login
// screen" is that list, in that order, with the baton visible.
//
// Drawn as a generic box-and-arrow diagram this reads as an architecture, and
// an architecture is a picture of things that coexist. So the picture here is a
// single horizontal line of capsules with chevrons between them and one spark
// that travels, stage to stage, in real time. A viewer who sees the spark
// arrive knows the previous stage is FINISHED, which is the fact a static
// arrow cannot carry.
//
// The order is the template, so the order is what the validator defends. A
// model writing a boot sequence from memory will cheerfully ignite the kernel
// before the bootloader, or narrate POST and then jump to login — and a relay
// whose baton teleports is worse than a bullet list, because it looks
// authoritative while teaching the wrong sequence. So ignites must start at the
// first stage and advance by exactly one, and a skip is rejected by NAME: the
// error says which stage was jumped over, because "out of order" sends a model
// hunting and "you skipped POST" does not.
//
// Hands is required on every stage but the last, because a stage with nothing
// to hand on has broken the metaphor: it is not a relay leg, it is a terminus.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "relay",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The baton pass",
		Description: "A strict ordered chain where each stage does one job and hands control to the next, with a spark travelling down the line. Reach for it when the answer is a sequence of hand-offs — boot, request lifecycle, compile to run — and the ORDER is the lesson.",
		Example:     "What happens between pressing the power button and the login screen",
		PromptFile:  snippetRelayTemplateName,
		NeedsCode:   false,
		// The line, three or four ignites and the whole chain: five distinct
		// states, and under thirty-five seconds the spark crosses faster than
		// the eye follows it.
		MinTargetSec:     35,
		DefaultTargetSec: 55,
		// Opener + up to six ignites + closer. Past eight the same line is
		// being re-lit.
		MaxBeats: 8,
		// A beat here is a SHOT — one stage lighting and one spark crossing —
		// not a step in an argument.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Relay: true},
		OwnsPlan:          planFields{Relay: true},
		Normalize:         normalizeRelayPlan,
		Validate:          validateRelayPlan,
		Scenes:            relayScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(RelayShows(), ", "),
				"MinStages":     minRelayStages,
				"MaxStages":     maxRelayStages,
				"MaxLabelWords": maxRelayLabelWords,
				"MaxDoesWords":  maxRelayDoesWords,
				"MaxHandsWords": maxRelayHandsWords,
				"MinIgnites":    minRelayIgnites,
			}
		},
	})
}

const snippetRelayTemplateName = "snippet_relay.tmpl"

const (
	// Three hand-offs is where a sequence stops being "A then B" and starts
	// being a chain the viewer has to hold in their head, which is the thing
	// this picture is for.
	minRelayStages = 4
	// Seven capsules with chevrons between them is exactly the width of the
	// stage at a label size that reads; an eighth has to shrink all of them.
	maxRelayStages = 7
	// A capsule label is a name — "bootloader", "POST" — and three words is
	// already two more than most stages need.
	maxRelayLabelWords = 3
	// The Does line sits under the label inside the capsule. Ten words is one
	// short clause, which is all a capsule holds without becoming a card.
	maxRelayDoesWords = 10
	// The hand-off rides under the chevron, in the gap between two capsules.
	// Six words is what fits there before it collides with its neighbours.
	maxRelayHandsWords = 6

	// Two lit stages is a hand-off; three is a chain. Below three the picture
	// never establishes that this is a sequence rather than a pair.
	minRelayIgnites = 3
)

// relayShows is the closed vocabulary of what a beat does.
var relayShows = map[string]bool{
	// Every stage as a dim capsule, the shape of the whole run. The opener.
	"line": true,
	// The stage at At lights, its Does appears, and a spark travels into it
	// from the stage before.
	"ignite": true,
	// The whole line lit with every hand-off visible. The closer.
	"chain": true,
}

// RelayShows returns the beat vocabulary sorted.
func RelayShows() []string {
	out := make([]string, 0, len(relayShows))
	for k := range relayShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// RelayStage is one leg of the run.
type RelayStage struct {
	// Label is the stage's name — "BIOS", "bootloader".
	Label string `json:"label"`
	// Does is the one job this stage performs before letting go.
	Does string `json:"does"`
	// Hands is what it passes on — "control at 0x7c00". Empty on the last
	// stage, which hands on to the human.
	Hands string `json:"hands,omitempty"`
}

// RelaySpec is the whole run. On the plan because the line is up for the
// entire clip and the beats only light parts of it.
type RelaySpec struct {
	// Stages are the legs, in the order they happen. The order IS the clip.
	Stages []RelayStage `json:"stages"`
}

// RelayBeat is one shot: which state of the line this beat shows.
type RelayBeat struct {
	// Show is a relayShows name.
	Show string `json:"show"`
	// At is the zero-based stage index, used only when Show is "ignite".
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to an ignite —
// the workhorse state most beats of this template are in.
func (b RelayBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if relayShows[s] {
		return s
	}
	return "ignite"
}

func normalizeRelayPlan(p *SnippetPlan) {
	r := p.Relay
	if r == nil {
		return
	}
	stages := make([]RelayStage, 0, len(r.Stages))
	for _, s := range r.Stages {
		s.Label = clampWords(collapseSpaces(s.Label), maxRelayLabelWords)
		s.Does = clampWords(collapseSpaces(s.Does), maxRelayDoesWords)
		s.Hands = clampWords(collapseSpaces(s.Hands), maxRelayHandsWords)
		if s.Label == "" {
			continue
		}
		stages = append(stages, s)
	}
	if len(stages) > maxRelayStages {
		stages = stages[:maxRelayStages]
	}
	r.Stages = stages

	for i := range p.Beats {
		b := p.Beats[i].Relay
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		// Clamp rather than drop: an index one past the end is a model that
		// counted from one, and the beat still means "the last stage".
		if b.At < 0 {
			b.At = 0
		}
		if n := len(r.Stages); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateRelayPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Relay: true}); err != nil {
		return err
	}

	r := p.Relay
	if r == nil {
		return fmt.Errorf("the plan has no chain — this template is a line of stages handing control down it, so without stages there is no line")
	}
	if n := len(r.Stages); n < minRelayStages || n > maxRelayStages {
		return fmt.Errorf("the chain has %d stages, want %d-%d. Under %d it is a pair of steps rather than a sequence anybody has to hold in their head, and over %d the capsules shrink below a readable label",
			n, minRelayStages, maxRelayStages, minRelayStages, maxRelayStages)
	}
	for i, s := range r.Stages {
		if strings.TrimSpace(s.Label) == "" {
			return fmt.Errorf("stage %d has no label. Every capsule is named — the name is what the viewer will repeat back", i+1)
		}
		if n := len(strings.Fields(s.Label)); n > maxRelayLabelWords {
			return fmt.Errorf("the stage label %q is %d words and a capsule holds %d. A stage has a name, not a description — what it does goes in \"does\"", s.Label, n, maxRelayLabelWords)
		}
		if strings.TrimSpace(s.Does) == "" {
			return fmt.Errorf("stage %q says nothing about what it does. A capsule with only a name is a box, and a row of boxes is the diagram this template exists to replace", s.Label)
		}
		if n := len(strings.Fields(s.Does)); n > maxRelayDoesWords {
			return fmt.Errorf("stage %q does %d words' worth, and the capsule holds %d. One clause: the single job it finishes before letting go", s.Label, n, maxRelayDoesWords)
		}
		if n := len(strings.Fields(s.Hands)); n > maxRelayHandsWords {
			return fmt.Errorf("the hand-off from %q is %d words and it rides under the chevron, in the gap between two capsules, which holds %d. Name the baton, not the ceremony", s.Label, n, maxRelayHandsWords)
		}
		// The last stage hands on to the human, so it alone may be silent.
		if i < len(r.Stages)-1 && strings.TrimSpace(s.Hands) == "" {
			return fmt.Errorf("stage %q hands nothing to %q. Every leg but the last passes something on — control, an address, a loaded image — and a stage with nothing to hand over is not a relay leg, it is a terminus",
				s.Label, r.Stages[i+1].Label)
		}
	}

	for _, b := range p.Beats {
		if b.Relay == nil {
			return fmt.Errorf("beat %q has no relay direction — every beat shows one state of the line, so every beat needs a {\"show\": ...}", b.ID)
		}
	}
	if p.Beats[0].Relay.ResolvedShow() != "line" {
		return fmt.Errorf("beat %q does not open on the line. A spark arriving at a capsule the viewer has not seen yet is a light with no map behind it — open with {\"show\": \"line\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Relay.ResolvedShow() != "chain" {
		return fmt.Errorf("the clip does not close on the whole chain. The final frame is the one the viewer keeps, and it has to show every hand-off at once — end with {\"show\": \"chain\"}")
	}

	// THE ORDER. This is the template's entire claim, so a skip is rejected by
	// name: "out of order" sends a model hunting, "you skipped POST" does not.
	prev := -1
	ignites := 0
	for i, b := range p.Beats {
		show := b.Relay.ResolvedShow()
		switch show {
		case "line":
			if i != 0 {
				return fmt.Errorf("beat %q dims back to the bare line part-way through. Once a stage has fired it stays lit — the line only exists un-lit at the start", b.ID)
			}
		case "chain":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q lights the whole chain before the end. \"chain\" is the closer, and after every hand-off is visible there is nothing left for a later beat to reveal", b.ID)
			}
		case "ignite":
			at := b.Relay.At
			if at < 0 || at >= len(r.Stages) {
				return fmt.Errorf("beat %q ignites stage %d, and the chain has %d stages numbered 0 to %d. The index is zero-based",
					b.ID, at, len(r.Stages), len(r.Stages)-1)
			}
			if prev < 0 && at != 0 {
				return fmt.Errorf("beat %q starts the run at stage %d, %q, so stage 0, %q, never fires. The chain starts where it starts — the first ignite is stage 0",
					b.ID, at, r.Stages[at].Label, r.Stages[0].Label)
			}
			if prev >= 0 && at <= prev {
				return fmt.Errorf("beat %q ignites stage %d, %q, after stage %d, %q, has already fired. The baton never goes backwards, and a relay whose spark doubles back teaches a sequence that does not happen",
					b.ID, at, r.Stages[at].Label, prev, r.Stages[prev].Label)
			}
			if prev >= 0 && at != prev+1 {
				return fmt.Errorf("beat %q jumps from stage %d, %q, to stage %d, %q, skipping %q. The order IS this template — every stage in between has to fire, because the whole claim of the picture is that control passes hand to hand with nothing in between",
					b.ID, prev, r.Stages[prev].Label, at, r.Stages[at].Label, r.Stages[prev+1].Label)
			}
			prev = at
			ignites++
		}
	}
	if ignites < minRelayIgnites {
		return fmt.Errorf("only %d stages ignite, and the clip needs at least %d. Two lit capsules is a hand-off; the picture does not become a chain until a third one fires", ignites, minRelayIgnites)
	}
	return nil
}

// relayScenes lays the clip out as ONE scene. Which capsules are lit, which
// hand-offs are visible and where the spark is travelling from are all decided
// here, so the component animates a given state rather than inferring one.
func relayScenes(in SnippetSceneInput) ([]Scene, error) {
	r := in.Plan.Relay
	if r == nil {
		return nil, fmt.Errorf("the plan has no chain")
	}

	stages := make([]map[string]any, len(r.Stages))
	for i, s := range r.Stages {
		stages[i] = map[string]any{
			"label": s.Label,
			"does":  s.Does,
			"hands": s.Hands,
		}
	}

	fired := map[int]bool{}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Relay == nil {
			return nil, fmt.Errorf("beat %q has no relay direction", beat.ID)
		}
		show := beat.Relay.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "ignite" {
			at := beat.Relay.At
			if at >= 0 && at < len(r.Stages) {
				fired[at] = true
				step["at"] = at
				// The spark's origin. Absent on the first stage, which is lit
				// by the power button rather than by a predecessor.
				if at > 0 {
					step["from"] = at - 1
				}
			}
		}
		lit := make([]int, 0, len(fired))
		for k := range fired {
			lit = append(lit, k)
		}
		sort.Ints(lit)
		if show == "chain" {
			// The closer lights everything, including any tail the ignites
			// did not have beats to reach.
			lit = make([]int, len(r.Stages))
			for k := range lit {
				lit[k] = k
			}
		}
		step["lit"] = lit
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneRelay,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"stages": stages,
			"steps":  steps,
		}),
	}}, nil
}
