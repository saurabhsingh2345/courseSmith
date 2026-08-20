package pipeline

// The wiring template: named parts, the path between them, one hop at a time.
//
// Three chapters of a course about an agent need the same picture and the catalog
// could not draw any of them in this house style: the loop the tool runs in
// (prompt, read, propose, approve, round again), the way it reaches a system
// outside the repo, and the chain an event travels down before a command runs.
// All three are the same shape — a handful of named boxes and the labelled hops
// between them — and all three are ruined by the same two mistakes.
//
// FIRST MISTAKE: drawing all the arrows at once. A diagram with every edge live
// is a map, and a map is something a viewer studies rather than watches. Here
// exactly one hop is lit at a time and the narration walks them, so the picture
// is a sequence rather than a state. The unlit edges stay visible at low ink,
// because the shape has to be legible for the lit hop to mean anything.
//
// SECOND MISTAKE: unlabelled edges. "A points at B" is not a claim. What travels
// down the wire is the entire content of the diagram — a prompt, a file read, a
// tool call, an approval — and an arrow without that word on it has drawn the
// architecture while withholding the mechanism. So a hop carries a label and the
// validator asks for it.
//
// The row is horizontal and it stays horizontal even for a loop, with the return
// drawn as an arc underneath. A ring would be prettier and it would cost the one
// thing this shape is good at: left-to-right is time, so the eye reads the order
// before it reads a single word. On a ring there is no first box.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "wiring",
		Category: CatSystems,
		Since:    SinceV10,
		Family:   FamilyAtelier,
		Title:    "The path between the parts",
		Description: "Named blocks in a row with labelled hops between them, lit one hop at a time, and an optional return arc when the path comes back round. " +
			"Reach for it when the lesson is what travels between the parts — a loop, a connection to an outside system, an event reaching a command.",
		Example:    "the loop Claude Code runs in",
		PromptFile: snippetWiringTemplateName,
		NeedsCode:  false,
		MinTargetSec:     20,
		DefaultTargetSec: 45,
		MaxTargetSec:     140,
		// The shape, up to five hops, the whole path.
		MaxBeats:          8,
		IdealWordsPerBeat: 26,
		Owns:              beatFields{Wiring: true},
		OwnsPlan:          planFields{Wiring: true},
		Normalize:         normalizeWiringPlan,
		Validate:          validateWiringPlan,
		Scenes:            wiringScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":        strings.Join(WiringShows(), ", "),
				"Kinds":        strings.Join(WiringKinds(), ", "),
				"MinNodes":     minWiringNodes,
				"MaxNodes":     maxWiringNodes,
				"MaxLabelWords": maxWiringLabelWords,
				"MaxNoteWords": maxWiringNoteWords,
				"MaxHopWords":  maxWiringHopWords,
			}
		},
	})
}

const snippetWiringTemplateName = "snippet_wiring.tmpl"

const (
	// Three blocks is a path; past five the row runs out of frame at a size
	// anybody can read.
	minWiringNodes = 3
	maxWiringNodes = 5

	maxWiringLabelWords = 4
	maxWiringNoteWords  = 10
	maxWiringHopWords   = 6
	maxWiringReturnWords = 8
)

// wiringShows is the closed vocabulary of what a beat does.
var wiringShows = map[string]bool{
	// The whole shape at low ink, nothing lit. The first beat.
	"shape": true,
	// Hop At lights: the edge into node At+1, and that node with it.
	"hop": true,
	// The return arc, for a path that comes back round.
	"round": true,
	// Every hop up at once — the finished path. The last beat.
	"path": true,
}

// wiringKinds is what a block IS, which decides how it is drawn.
var wiringKinds = map[string]bool{
	// Where the work comes from: a person, a prompt, an event.
	"in": true,
	// The thing doing the work.
	"work": true,
	// Something that holds state — a file, a database, a repo.
	"store": true,
	// Where the work ends up.
	"out": true,
}

// WiringShows returns the beat vocabulary sorted.
func WiringShows() []string {
	out := make([]string, 0, len(wiringShows))
	for k := range wiringShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// WiringKinds returns the block vocabulary sorted.
func WiringKinds() []string {
	out := make([]string, 0, len(wiringKinds))
	for k := range wiringKinds {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// WiringSpec is the row and the hops along it.
type WiringSpec struct {
	// Nodes are the blocks, left to right.
	Nodes []WiringNode `json:"nodes"`
	// Hops are the labels on the edges between them. Hop i is the edge from
	// node i to node i+1, so there is one fewer hop than there are nodes.
	Hops []string `json:"hops"`
	// Return is the label on the arc back to the start. Set it and the path is
	// a loop; leave it empty and the path ends at the last block.
	Return string `json:"return,omitempty"`
}

// WiringNode is one block in the row.
type WiringNode struct {
	Label string `json:"label"`
	Kind  string `json:"kind,omitempty"`
	// Note is the line under the block saying what it does.
	Note string `json:"note,omitempty"`
}

// ResolvedKind defaults to the thing doing the work.
func (n WiringNode) ResolvedKind() string {
	k := strings.ToLower(strings.TrimSpace(n.Kind))
	if wiringKinds[k] {
		return k
	}
	return "work"
}

// WiringBeat is one shot of the path.
type WiringBeat struct {
	Show string `json:"show"`
	At   int    `json:"at,omitempty"`
}

// ResolvedShow defaults to a hop lighting.
func (b WiringBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if wiringShows[s] {
		return s
	}
	return "hop"
}

func normalizeWiringPlan(p *SnippetPlan) {
	w := p.Wiring
	if w == nil {
		return
	}
	nodes := make([]WiringNode, 0, maxWiringNodes)
	for _, n := range w.Nodes {
		n.Label = clampWords(collapseSpaces(n.Label), maxWiringLabelWords)
		n.Note = clampWords(collapseSpaces(n.Note), maxWiringNoteWords)
		n.Kind = n.ResolvedKind()
		if n.Label != "" && len(nodes) < maxWiringNodes {
			nodes = append(nodes, n)
		}
	}
	w.Nodes = nodes

	hops := make([]string, 0, maxWiringNodes)
	for _, h := range w.Hops {
		if h = clampWords(collapseSpaces(h), maxWiringHopWords); h != "" && len(hops) < maxWiringNodes {
			hops = append(hops, h)
		}
	}
	w.Hops = hops
	w.Return = clampWords(collapseSpaces(w.Return), maxWiringReturnWords)

	for i := range p.Beats {
		if b := p.Beats[i].Wiring; b != nil {
			b.Show = b.ResolvedShow()
			if b.Show != "hop" {
				b.At = 0
			}
		}
	}
}

func validateWiringPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Wiring: true}); err != nil {
		return err
	}

	w := p.Wiring
	if w == nil {
		return fmt.Errorf("the plan has no wiring — this template is one path, so the path is the clip")
	}
	if n := len(w.Nodes); n < minWiringNodes || n > maxWiringNodes {
		return fmt.Errorf("the path has %d block(s) and wants %d-%d", n, minWiringNodes, maxWiringNodes)
	}
	// One fewer hop than nodes, and every one labelled. An unlabelled edge is
	// the failure this template exists to prevent — see the file header.
	if want := len(w.Nodes) - 1; len(w.Hops) != want {
		return fmt.Errorf("there are %d blocks and %d hop label(s); a row of %d blocks has exactly %d edges, and each one needs the word for what travels down it",
			len(w.Nodes), len(w.Hops), len(w.Nodes), want)
	}

	var (
		next   int
		shapes int
		paths  int
		rounds int
	)
	for i, b := range p.Beats {
		if b.Wiring == nil {
			return fmt.Errorf("beat %q has no wiring direction", b.ID)
		}
		switch b.Wiring.ResolvedShow() {
		case "shape":
			shapes++
			if i != 0 {
				return fmt.Errorf("beat %q draws the shape, but it is beat %d — the shape arrives once, first", b.ID, i+1)
			}
		case "hop":
			if b.Wiring.At != next {
				return fmt.Errorf("beat %q lights hop %d but hop %d has not been walked yet. The path is read left to right, so the hops light in order with none skipped",
					b.ID, b.Wiring.At, next)
			}
			next++
			if next > len(w.Hops) {
				return fmt.Errorf("beat %q lights a hop past the end of the path", b.ID)
			}
		case "round":
			rounds++
			if strings.TrimSpace(w.Return) == "" {
				return fmt.Errorf("beat %q shows the return arc but the path has no return label — a loop has to say what comes back round", b.ID)
			}
		case "path":
			paths++
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q lights the whole path, but %d beat(s) follow it — that is the closing frame", b.ID, len(p.Beats)-1-i)
			}
		}
	}
	if shapes != 1 {
		return fmt.Errorf("there are %d \"shape\" beats and there must be exactly one", shapes)
	}
	if paths != 1 {
		return fmt.Errorf("there are %d \"path\" beats and there must be exactly one, last", paths)
	}
	if rounds > 1 {
		return fmt.Errorf("the return arc is shown %d times; it comes back round once", rounds)
	}
	if next != len(w.Hops) {
		return fmt.Errorf("%d of the %d hops are never walked. An edge nobody narrates is a line the viewer counted and heard nothing about",
			len(w.Hops)-next, len(w.Hops))
	}
	return nil
}

func wiringScenes(in SnippetSceneInput) ([]Scene, error) {
	w := in.Plan.Wiring
	if w == nil {
		return nil, fmt.Errorf("the plan has no wiring")
	}

	nodes := make([]map[string]any, 0, len(w.Nodes))
	for _, n := range w.Nodes {
		nodes = append(nodes, map[string]any{
			"label": n.Label,
			"kind":  n.ResolvedKind(),
			"note":  n.Note,
		})
	}

	// walked is latched: a hop that has been narrated stays lit, so the path
	// builds up across the clip rather than a single spark moving along it. The
	// difference matters — the built-up version lets a viewer see how far through
	// the mechanism they are, which is the same reason the waypoint has a spine.
	walked, round := 0, false
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Wiring == nil {
			return nil, fmt.Errorf("beat %q has no wiring direction", beat.ID)
		}
		show := beat.Wiring.ResolvedShow()
		switch show {
		case "hop":
			walked = beat.Wiring.At + 1
		case "round":
			round = true
		case "path":
			walked = len(w.Hops)
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"at":      beat.Wiring.At,
			"walked":  walked,
			"round":   round,
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneWiring,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"nodes":  nodes,
			"hops":   w.Hops,
			"return": w.Return,
			"steps":  steps,
		},
	}}, nil
}
