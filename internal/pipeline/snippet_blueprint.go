package pipeline

// The blueprint template: the block diagram, and the paths data takes across it.
//
// Von Neumann architecture is the first diagram every CS course draws and the
// first one every textbook gets away with murdering: five boxes, some lines,
// and no way to tell which line matters when. A static block diagram states
// that things are CONNECTED; what a learner needs to see is things TALKING —
// an address going out on one bus while data comes back on another. The
// catalog's flow template draws a process moving forward; it had nothing that
// draws a standing structure whose edges light up one at a time.
//
// So the frame is the whole board from the first beat — blocks with ports,
// wires between them — and the beats do two different things to it: focus one
// block to say what it is, or light a PATH, pulsing along one wire from its
// source to its destination. The board never rearranges; only attention moves.
//
// The validators keep the wiring honest.
//
// **Every wire ends on real blocks.** A wire is stored as a pair of block IDs,
// and an ID that resolves to nothing is a connection to a component that is
// not on the board — the error quotes the bad ID because the model usually
// misspelled a slug it defined three lines earlier.
//
// **A path repeats a wire only when every wire has had its turn.** The wires
// ARE the content: a clip that pulses the same bus twice while another sits
// dark has explained one connection and merely drawn the other, and drawing a
// connection nobody explains is how the textbook diagram failed in the first
// place.
//
// **The clip opens on the whole board.** A wire lighting between blocks the
// viewer has not seen in place is a line segment, not a bus.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "blueprint",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The block diagram",
		Description: "Labeled blocks with ports and the wires between them, where beats light a path along one wire at a time to show data flowing. Reach for it for Von Neumann architecture, CPU internals, bus systems — any standing structure whose point is who talks to whom.",
		Example:     "Von Neumann architecture: how the CPU, memory and I/O talk over buses",
		PromptFile:  snippetBlueprintTemplateName,
		NeedsCode:   false,
		// A board, a few block focuses, a path per wire and a closer need more
		// than half a minute of voice; below that the wires go unexplained.
		MinTargetSec:     35,
		DefaultTargetSec: 60,
		// One more than the shot templates around it: a board beat, a handful of
		// blocks and paths, and a closer. Nine is where a viewer stops holding
		// the board in mind between pulses.
		MaxBeats: 9,
		// A beat is a shot — one block forward or one pulse along a wire — and
		// 28 words is nine seconds, which is as long as one lit edge holds.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Blueprint: true},
		OwnsPlan:          planFields{Blueprint: true},
		Normalize:         normalizeBlueprintPlan,
		Validate:          validateBlueprintPlan,
		Scenes:            blueprintScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":             strings.Join(MetricRoles(), ", "),
				"Shows":             strings.Join(BlueprintShows(), ", "),
				"BlockRoles":        strings.Join(BlueprintRoles(), ", "),
				"MinBlocks":         minBlueprintBlocks,
				"MaxBlocks":         maxBlueprintBlocks,
				"MinWires":          minBlueprintWires,
				"MaxWires":          maxBlueprintWires,
				"MaxLabelWords":     maxBlueprintLabelWords,
				"MaxWireLabelWords": maxBlueprintWireLabelWords,
			}
		},
	})
}

const snippetBlueprintTemplateName = "snippet_blueprint.tmpl"

const (
	// Below three blocks there is nothing for a wire to be BETWEEN that a
	// sentence would not say better; past six the grid rows shrink until a
	// pulse crosses the frame too fast to follow.
	minBlueprintBlocks = 3
	maxBlueprintBlocks = 6

	// Two wires is the least board that has a "which path" question in it at
	// all; past eight the picture is a net, and a pulse on a net reads as
	// decoration rather than as an answer.
	minBlueprintWires = 2
	maxBlueprintWires = 8

	maxBlueprintLabelWords     = 3
	maxBlueprintWireLabelWords = 3
)

// blueprintShows is the closed vocabulary of what a beat does.
var blueprintShows = map[string]bool{
	// The whole diagram, everything in place. The opener.
	"board": true,
	// Focus block At: it brightens and its neighbours dim.
	"block": true,
	// Animate a pulse along wire At, from its From to its To.
	"path": true,
	// Everything lit at once. The closer.
	"whole": true,
}

// BlueprintShows returns the beat vocabulary sorted.
func BlueprintShows() []string {
	out := make([]string, 0, len(blueprintShows))
	for k := range blueprintShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// blueprintRoles is what kind of component a block is. A closed vocabulary
// because the renderer styles a store differently from a unit, and an invented
// kind has no look.
var blueprintRoles = map[string]bool{"unit": true, "store": true, "io": true}

// BlueprintRoles returns the block role vocabulary sorted.
func BlueprintRoles() []string {
	out := make([]string, 0, len(blueprintRoles))
	for k := range blueprintRoles {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// BlueprintSpec is the board: blocks and the wires between them. On the plan
// because the structure persists for the whole clip.
type BlueprintSpec struct {
	// Blocks are the components, laid out on a grid in this order.
	Blocks []BlueprintBlock `json:"blocks"`
	// Wires are the connections, each drawn from one block to another.
	Wires []BlueprintWire `json:"wires"`
}

// BlueprintBlock is one component on the board.
type BlueprintBlock struct {
	// ID is a short slug other fields point at — "cpu", "ram".
	ID string `json:"id"`
	// Label is the block's on-screen name.
	Label string `json:"label"`
	// Role is what kind of component this is: a blueprintRoles name.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the block's kind, defaulting to a processing unit.
func (b BlueprintBlock) ResolvedRole() string {
	s := strings.ToLower(strings.TrimSpace(b.Role))
	if blueprintRoles[s] {
		return s
	}
	return "unit"
}

// BlueprintWire is one connection between two blocks.
type BlueprintWire struct {
	// From and To are block IDs.
	From string `json:"from"`
	To   string `json:"to"`
	// Label names the wire — "address bus". Optional.
	Label string `json:"label,omitempty"`
}

// BlueprintBeat is one move: which block or wire this beat is on.
type BlueprintBeat struct {
	// Show is a blueprintShows name.
	Show string `json:"show"`
	// At indexes Blocks for a "block" beat and Wires for a "path" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to a path —
// the move this template exists for.
func (b BlueprintBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if blueprintShows[s] {
		return s
	}
	return "path"
}

func normalizeBlueprintPlan(p *SnippetPlan) {
	bp := p.Blueprint
	if bp == nil {
		return
	}

	blocks := make([]BlueprintBlock, 0, len(bp.Blocks))
	for _, b := range bp.Blocks {
		// IDs are matched, not displayed, so they are normalized hard; a wire
		// naming "CPU " and a block naming "cpu" meant the same component.
		b.ID = slugify(collapseSpaces(b.ID))
		b.Label = clampWords(collapseSpaces(b.Label), maxBlueprintLabelWords)
		b.Role = b.ResolvedRole()
		if len(blocks) < maxBlueprintBlocks {
			blocks = append(blocks, b)
		}
	}
	bp.Blocks = blocks

	// Which blocks actually survived the cap above. A wire is only meaningful
	// if both its ends are still on the board.
	onBoard := make(map[string]bool, len(bp.Blocks))
	for _, b := range bp.Blocks {
		onBoard[b.ID] = true
	}

	wires := make([]BlueprintWire, 0, len(bp.Wires))
	for _, w := range bp.Wires {
		w.From = slugify(collapseSpaces(w.From))
		w.To = slugify(collapseSpaces(w.To))
		w.Label = clampWords(collapseSpaces(w.Label), maxBlueprintWireLabelWords)
		// A wire onto a block the cap just removed is damage this function
		// did, not something the model wrote. Left in place it reaches the
		// validator as "no block on the board has that id" — an error that
		// describes a board the model never sent, and which it therefore
		// cannot act on. It would spend a correction round re-pointing a wire
		// that was already correct when it arrived. Dropping it is the honest
		// repair: the diagram loses an edge it had no room for either way.
		if !onBoard[w.From] || !onBoard[w.To] {
			continue
		}
		if len(wires) < maxBlueprintWires {
			wires = append(wires, w)
		}
	}
	bp.Wires = wires

	for i := range p.Beats {
		b := p.Beats[i].Blueprint
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.At < 0 {
			b.At = 0
		}
		switch b.Show {
		case "block":
			if n := len(bp.Blocks); n > 0 && b.At >= n {
				b.At = n - 1
			}
		case "path":
			if n := len(bp.Wires); n > 0 && b.At >= n {
				b.At = n - 1
			}
		default:
			b.At = 0
		}
	}
}

func validateBlueprintPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Blueprint: true}); err != nil {
		return err
	}

	bp := p.Blueprint
	if bp == nil {
		return fmt.Errorf("the plan has no blueprint — this template is blocks, wires, and the paths data takes between them, so the board is the clip")
	}
	if n := len(bp.Blocks); n < minBlueprintBlocks || n > maxBlueprintBlocks {
		return fmt.Errorf("the board has %d block(s), want %d-%d. Below %d there is nothing for a wire to be between that a sentence would not say better; past %d the grid shrinks until a pulse crosses the frame too fast to follow",
			n, minBlueprintBlocks, maxBlueprintBlocks, minBlueprintBlocks, maxBlueprintBlocks)
	}
	if n := len(bp.Wires); n < minBlueprintWires || n > maxBlueprintWires {
		return fmt.Errorf("the board has %d wire(s), want %d-%d. Below %d there is no \"which path\" question in the picture at all; past %d it is a net, and a pulse on a net reads as decoration",
			n, minBlueprintWires, maxBlueprintWires, minBlueprintWires, maxBlueprintWires)
	}

	ids := map[string]bool{}
	for i, b := range bp.Blocks {
		id := strings.TrimSpace(b.ID)
		if id == "" {
			return fmt.Errorf("block %d has no id — wires point at blocks by id, so a block without one cannot be wired to anything", i)
		}
		if strings.TrimSpace(b.Label) == "" {
			return fmt.Errorf("block %q has no label. The board is the diagram the viewer keeps; an unlabeled block on it is a box that never gets explained", id)
		}
		if ids[id] {
			return fmt.Errorf("two blocks share the id %q, so a wire pointing at it could mean either — give every block its own id", id)
		}
		ids[id] = true
		if r := strings.ToLower(strings.TrimSpace(b.Role)); r != "" && !blueprintRoles[r] {
			return fmt.Errorf("block %q has role %q, which is not one of: %s. The renderer styles a store differently from a unit, so an invented kind has no look",
				id, b.Role, strings.Join(BlueprintRoles(), ", "))
		}
	}
	for i, w := range bp.Wires {
		if !ids[w.From] {
			return fmt.Errorf("wire %d runs from %q, and no block on the board has that id. A wire ending on a component that is not drawn is a connection to nowhere — the block ids are: %s",
				i, w.From, strings.Join(blueprintBlockIDs(bp), ", "))
		}
		if !ids[w.To] {
			return fmt.Errorf("wire %d runs to %q, and no block on the board has that id. A wire ending on a component that is not drawn is a connection to nowhere — the block ids are: %s",
				i, w.To, strings.Join(blueprintBlockIDs(bp), ", "))
		}
		if w.From == w.To {
			return fmt.Errorf("wire %d runs from %q to itself. A pulse along it would leave and arrive in the same place, which shows nothing moving", i, w.From)
		}
	}

	// The whole board before any of it lights: a wire pulsing between blocks
	// nobody has seen in place is a line segment, not a bus.
	if p.Beats[0].Blueprint == nil || p.Beats[0].Blueprint.ResolvedShow() != "board" {
		return fmt.Errorf("beat %q does not open on the board. The viewer has to see the whole structure before any one edge of it lights — make the first beat {\"show\": \"board\"}",
			p.Beats[0].ID)
	}

	// And the whole board lit at the end. Every template in this batch closes
	// on its subject entire, because the last frame is the one a viewer holds
	// after the voice stops — and on a diagram whose parts have been lit one at
	// a time, the only frame that shows the machine rather than a tour of it is
	// the one where all of it is on at once.
	last := p.Beats[len(p.Beats)-1]
	if last.Blueprint == nil || last.Blueprint.ResolvedShow() != "whole" {
		return fmt.Errorf("the clip does not close on the whole board. The viewer has watched one edge light at a time and never seen the machine working as a machine — end with {\"show\": \"whole\"}")
	}

	pulsed := map[int]bool{}
	for i, b := range p.Beats {
		if b.Blueprint == nil {
			return fmt.Errorf("beat %q has no blueprint direction — every beat shows the board, focuses a block, lights a path, or lights the whole", b.ID)
		}
		switch b.Blueprint.ResolvedShow() {
		case "board":
			if i != 0 {
				return fmt.Errorf("beat %q re-opens on the board part-way through. Attention has moved on; going back to the establishing shot un-explains everything since", b.ID)
			}
		case "whole":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q lights the whole board part-way through. Once everything is on there is nothing left to reveal, so any beat after it is a step backwards — the whole is the last shot", b.ID)
			}
		case "block":
			if b.Blueprint.At < 0 || b.Blueprint.At >= len(bp.Blocks) {
				return fmt.Errorf("beat %q focuses block %d, which does not exist — the board has blocks 0-%d", b.ID, b.Blueprint.At, len(bp.Blocks)-1)
			}
		case "path":
			at := b.Blueprint.At
			if at < 0 || at >= len(bp.Wires) {
				return fmt.Errorf("beat %q lights wire %d, which does not exist — the board has wires 0-%d", b.ID, at, len(bp.Wires)-1)
			}
			// A wire may carry a second pulse only once every wire has carried
			// one: the wires are the content, and repeating a lit one while
			// another sits dark leaves a drawn connection unexplained.
			if pulsed[at] && len(pulsed) < len(bp.Wires) {
				return fmt.Errorf("beat %q lights wire %d again while %d of the %d wires have never lit. A drawn connection nobody explains is the textbook diagram this template exists to fix — pulse the dark wires before repeating a lit one",
					b.ID, at, len(bp.Wires)-len(pulsed), len(bp.Wires))
			}
			pulsed[at] = true
		}
	}
	return nil
}

// blueprintBlockIDs lists the board's block ids for error messages.
func blueprintBlockIDs(bp *BlueprintSpec) []string {
	out := make([]string, 0, len(bp.Blocks))
	for _, b := range bp.Blocks {
		out = append(out, b.ID)
	}
	return out
}

// blueprintScenes lays the clip out as ONE scene. The board persists; the
// steps only say where attention is and which wires have carried a pulse.
func blueprintScenes(in SnippetSceneInput) ([]Scene, error) {
	bp := in.Plan.Blueprint
	if bp == nil {
		return nil, fmt.Errorf("the plan has no blueprint")
	}

	indexOf := map[string]int{}
	for i, b := range bp.Blocks {
		indexOf[b.ID] = i
	}
	blocks := make([]map[string]any, len(bp.Blocks))
	for i, b := range bp.Blocks {
		blocks[i] = map[string]any{"id": b.ID, "label": b.Label, "role": b.ResolvedRole()}
	}
	wires := make([]map[string]any, len(bp.Wires))
	for i, w := range bp.Wires {
		from, ok := indexOf[w.From]
		if !ok {
			return nil, fmt.Errorf("wire %d runs from %q, which is not a block on the board", i, w.From)
		}
		to, ok := indexOf[w.To]
		if !ok {
			return nil, fmt.Errorf("wire %d runs to %q, which is not a block on the board", i, w.To)
		}
		// Endpoints resolved to indices here, once, so the renderer never does
		// string matching to find where a pulse starts.
		wires[i] = map[string]any{"from": from, "to": to, "label": w.Label}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	pulsed := map[int]bool{}
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Blueprint == nil {
			return nil, fmt.Errorf("beat %q has no blueprint direction", beat.ID)
		}
		show := beat.Blueprint.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "block" || show == "path" {
			step["at"] = beat.Blueprint.At
		}
		if show == "path" {
			pulsed[beat.Blueprint.At] = true
		}
		// The closer lights everything. Without this the final frame is the
		// board with whichever wires happened to have pulsed — which is the
		// same picture as the beat before it, and leaves the one shot of the
		// machine entire looking like a step that failed to do anything.
		if show == "whole" {
			for at := range bp.Wires {
				pulsed[at] = true
			}
		}
		lit := make([]int, 0, len(pulsed))
		for at := range pulsed {
			lit = append(lit, at)
		}
		sort.Ints(lit)
		step["lit"] = lit
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneBlueprint,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"blocks": blocks,
			"wires":  wires,
			"steps":  steps,
		}),
	}}, nil
}
