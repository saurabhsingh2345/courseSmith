package pipeline

// The flow / systems diagram template.
//
// The clip for "how does this system actually work": labelled boxes in layered
// columns, edges curving between them, and — once the graph is up — traffic
// visibly moving along those edges while the narrator traces a path through it.
//
// That last part is why this is not the `whiteboard` template with square
// corners, and why it is not a static Mermaid render either. A systems diagram
// is about *movement*: which way requests go, what feeds what, which path is hot.
// A picture can only imply that; a clip can show it. Beats can focus a subset of
// the graph, and the renderer dims everything else and runs the tokens only on
// the focused path — so the same diagram carries five different explanations
// without ever being redrawn.
//
// As in every template, the model does not lay anything out. It declares a DAG —
// nodes with a kind and the nodes that feed them — and Go ranks it. Layout is a
// consequence of the graph's shape, not of the model's taste.

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "flow",
		Title:       "Systems flow",
		Description: "Layered boxes with traffic moving along the edges, focusing one path at a time.",
		Example:     "How a request flows through a rate-limited API gateway",
		PromptFile:  snippetFlowTemplateName,
		NeedsCode:   false,
		Validate:    validateFlowPlan,
		Scenes:      flowScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Kinds":       strings.Join(FlowNodeKinds(), ", "),
				"MinNodes":    minFlowNodes,
				"MaxNodes":    maxFlowNodes,
				"MaxRanks":    maxFlowRanks,
				"MaxPerRank":  maxFlowPerRank,
				"MaxLabelWds": maxFlowLabelWords,
			}
		},
	})
}

const snippetFlowTemplateName = "snippet_flow.tmpl"

// Graph capacity. The ceilings are layout constraints, not taste: four columns
// across the 1700px stage leaves ~340px per node, which is what a three-word
// label needs to stay readable at 1080p, and a fifth row in a column pushes the
// bottom node into the caption band.
const (
	minFlowNodes      = 4
	maxFlowNodes      = 9
	maxFlowRanks      = 4
	maxFlowPerRank    = 4
	maxFlowLabelWords = 4
)

// flowNodeKinds is the closed vocabulary of node types, mapped to the artwork
// figure each one is drawn with. Kind drives the accent, the figure and the
// silhouette.
//
// The values are `artFigureVocab` names, not icon names, and the change is the
// point: the figure set was built for exactly these concepts and it *animates*.
// A queue whose items visibly drain says what a queue is; the flat `layers`
// glyph it used to show says only that somebody picked an icon. Every one of
// these — server, database, queue, cache — already existed in the vocabulary and
// was going unused here.
//
// Kind still does not drive the *bounding box*: every silhouette is drawn inside
// the same rect the ranking algorithm placed, so layout stays entirely the
// layout's business. What changed is that the box is no longer always a
// rectangle — see nodeShape() in flow.ts, and note that horizontal edges anchor
// at mid-height and vertical ones at mid-width, which is exactly where every
// shape in that switch reaches its full extent.
var flowNodeKinds = map[string]string{
	"service":  "server",
	"store":    "database",
	"queue":    "queue",
	"external": "cloud",
	"cache":    "cache",
	"decision": "switch",
	"client":   "monitor",
}

// FlowNodeKinds returns the vocabulary sorted, for prompts and docs.
func FlowNodeKinds() []string {
	out := make([]string, 0, len(flowNodeKinds))
	for k := range flowNodeKinds {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// FlowNode is one box in the system.
type FlowNode struct {
	// ID is a slug, unique in the clip; From and Focus reference it.
	ID string `json:"id"`
	// Label is the box's caption, at most a few words.
	Label string `json:"label"`
	// Kind is a flowNodeKinds name; anything else degrades to "service".
	Kind string `json:"kind"`
	// From lists the ids of nodes that feed this one. They must already exist,
	// which is what makes the graph acyclic by construction — a cycle would
	// have no valid ranking and the layout would have nowhere to put it.
	From []string `json:"from,omitempty"`
}

// ResolvedKind returns the node's kind, defaulting the unknown to a service.
func (n FlowNode) ResolvedKind() string {
	k := strings.ToLower(strings.TrimSpace(n.Kind))
	if _, ok := flowNodeKinds[k]; ok {
		return k
	}
	return "service"
}

// flowScenes lays the clip out as an optional title card followed by one
// diagram that runs to the end: nodes and edges arrive on the beat that
// introduces them, and each beat's focus set steers what is lit.
func flowScenes(in SnippetSceneInput) ([]Scene, error) {
	firstNode := -1
	for i, b := range in.Plan.Beats {
		if len(b.Nodes) > 0 {
			firstNode = i
			break
		}
	}
	if firstNode < 0 {
		return nil, fmt.Errorf("no beat adds anything to the diagram")
	}

	_, clipStart, _ := in.Beat(0)
	_, firstNodeStart, _ := in.Beat(firstNode)

	var scenes []Scene
	graphStart := clipStart
	if firstNodeStart-clipStart >= minTitleCardMs {
		titleEnd := min(firstNodeStart, clipStart+maxTitleCardMs)
		scenes = append(scenes, Scene{
			Type:    SceneTitle,
			StartMs: clipStart,
			EndMs:   titleEnd,
			Props: map[string]any{
				"heading":  in.Plan.Title,
				"subtitle": in.Plan.Subtitle,
				"intro":    true,
			},
		})
		graphStart = titleEnd
	}

	// Collect every node in declaration order, with its arrival time.
	type placed struct {
		node  FlowNode
		index int
		atMs  int
		rank  int
	}
	nodes := make([]placed, 0, maxFlowNodes)
	indexOf := map[string]int{}
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if len(beat.Nodes) == 0 {
			continue
		}
		window := (endMs - startMs) * 2 / 3
		step := window / len(beat.Nodes)
		for j, n := range beat.Nodes {
			indexOf[n.ID] = len(nodes)
			nodes = append(nodes, placed{node: n, index: len(nodes), atMs: startMs + j*step})
		}
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no diagram nodes were produced")
	}

	// Longest-path layering. Because From only ever names an earlier node, one
	// forward pass is enough and no cycle can exist.
	perRank := map[int]int{}
	nodeProps := make([]map[string]any, len(nodes))
	var edges []map[string]any
	maxRank := 0
	for i := range nodes {
		rank := 0
		for _, from := range nodes[i].node.From {
			j, ok := indexOf[from]
			if !ok {
				return nil, fmt.Errorf("node %q is fed by unknown node %q", nodes[i].node.ID, from)
			}
			if nodes[j].rank+1 > rank {
				rank = nodes[j].rank + 1
			}
			edges = append(edges, map[string]any{
				"from": j,
				"to":   i,
				"atMs": nodes[i].atMs,
			})
		}
		nodes[i].rank = rank
		if rank > maxRank {
			maxRank = rank
		}
		nodeProps[i] = map[string]any{
			"id":    nodes[i].node.ID,
			"label": nodes[i].node.Label,
			"kind":  nodes[i].node.ResolvedKind(),
			"icon":  flowNodeKinds[nodes[i].node.ResolvedKind()],
			"rank":  rank,
			"order": perRank[rank],
			"atMs":  nodes[i].atMs,
		}
		perRank[rank]++
	}
	if maxRank+1 > maxFlowRanks {
		return nil, fmt.Errorf("the diagram is %d columns deep, at most %d fit the stage", maxRank+1, maxFlowRanks)
	}
	for rank, n := range perRank {
		if n > maxFlowPerRank {
			return nil, fmt.Errorf("column %d holds %d nodes, at most %d fit the stage", rank, n, maxFlowPerRank)
		}
	}

	// Focus windows: which nodes are lit, and when.
	var focus []map[string]any
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if len(beat.Focus) == 0 {
			continue
		}
		ids := make([]int, 0, len(beat.Focus))
		for _, id := range beat.Focus {
			if j, ok := indexOf[id]; ok {
				ids = append(ids, j)
			}
		}
		if len(ids) == 0 {
			continue
		}
		focus = append(focus, map[string]any{"startMs": startMs, "endMs": endMs, "nodes": ids})
	}

	_, _, graphEnd := in.Beat(len(in.Plan.Beats) - 1)
	props := map[string]any{
		"title": in.Plan.Title,
		"nodes": nodeProps,
		"edges": edges,
		"ranks": maxRank + 1,
	}
	if len(focus) > 0 {
		props["focus"] = focus
	}
	scenes = append(scenes, Scene{
		Type:    SceneFlow,
		StartMs: graphStart,
		EndMs:   graphEnd,
		Props:   props,
	})
	return scenes, nil
}

func validateFlowPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Nodes: true, Focus: true}); err != nil {
		return err
	}

	seen := map[string]bool{}
	total := 0
	for _, b := range p.Beats {
		for _, n := range b.Nodes {
			id := strings.TrimSpace(n.ID)
			if id == "" {
				return fmt.Errorf("beat %q has a node with no id", b.ID)
			}
			if seen[id] {
				return fmt.Errorf("node id %q appears twice — each node is drawn once and stays", id)
			}
			if strings.TrimSpace(n.Label) == "" {
				return fmt.Errorf("node %q has no label", id)
			}
			if w := len(strings.Fields(n.Label)); w > maxFlowLabelWords {
				return fmt.Errorf("node %q has a %d-word label; labels are at most %d words", id, w, maxFlowLabelWords)
			}
			// Edges must point forward. A node fed by something not yet
			// declared could form a cycle, which has no valid layering.
			for _, from := range n.From {
				if !seen[from] {
					return fmt.Errorf("node %q is fed by %q, which is not declared yet — the diagram must flow forwards", id, from)
				}
			}
			seen[id] = true
			total++
		}
	}
	if total < minFlowNodes || total > maxFlowNodes {
		return fmt.Errorf("the diagram has %d nodes, want %d-%d", total, minFlowNodes, maxFlowNodes)
	}
	// A graph with no edges is a row of unrelated boxes; the whole point of this
	// template is what connects to what.
	edges := 0
	children := map[string]int{}
	parents := map[string]int{}
	for _, b := range p.Beats {
		for _, n := range b.Nodes {
			edges += len(n.From)
			parents[n.ID] += len(n.From)
			for _, from := range n.From {
				children[from]++
			}
		}
	}
	if edges == 0 {
		return fmt.Errorf("no node is connected to any other — a flow diagram needs edges (set from)")
	}
	// The diagram has to fork or join somewhere. A straight chain is a worse fit
	// for this template than for `whiteboard`, which draws chains well — and a
	// layered layout of single-node columns wastes most of the frame. Requiring
	// a branch is what keeps the two templates distinct.
	forks := false
	for _, n := range children {
		if n >= 2 {
			forks = true
		}
	}
	for _, n := range parents {
		if n >= 2 {
			forks = true
		}
	}
	if !forks {
		return fmt.Errorf("the diagram is a straight chain — a flow diagram must fork or join somewhere (one node feeding two, or two feeding one)")
	}

	// Focus sets: each has to actually select something. A set covering every
	// node dims nothing, and two identical sets narrate the same picture twice.
	var focusSets []string
	for _, b := range p.Beats {
		if len(b.Focus) == 0 {
			continue
		}
		unique := map[string]bool{}
		for _, id := range b.Focus {
			if !seen[id] {
				return fmt.Errorf("beat %q focuses %q, which is not a node in the diagram", b.ID, id)
			}
			unique[id] = true
		}
		if len(unique) >= total {
			return fmt.Errorf("beat %q focuses every node, which highlights nothing — focus the path this beat is actually about", b.ID)
		}
		key := focusKey(unique)
		if slices.Contains(focusSets, key) {
			return fmt.Errorf("beat %q focuses the same nodes as an earlier beat — each focused beat should trace a different path", b.ID)
		}
		focusSets = append(focusSets, key)
	}
	return nil
}

// focusKey canonicalizes a focus set so duplicates across beats are detectable
// regardless of the order the model listed them in.
func focusKey(ids map[string]bool) string {
	out := make([]string, 0, len(ids))
	for id := range ids {
		out = append(out, id)
	}
	sort.Strings(out)
	return strings.Join(out, "|")
}
