package pipeline

// Multi-format diagrams (workstream A): beyond standalone SVG, a diagram can be
// a structured node-link graph ("d3" kind). The content model emits a compact
// JSON spec (nodes + edges + layout) which the Go stage validates structurally,
// and the renderer lays out and animates with D3 — nodes popping in, edges
// drawing on. Structured specs are cheaper to get right than freehand SVG and
// animate far better in the video.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// D3 layout kinds the renderer understands.
const (
	D3LayoutTree   = "tree"   // rooted hierarchy (d3.tree)
	D3LayoutRadial = "radial" // rooted hierarchy on concentric rings
	D3LayoutForce  = "force"  // general graph, deterministic ring placement
)

// D3Node is one vertex.
type D3Node struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	// Group buckets nodes for colouring (0-based); optional.
	Group int `json:"group,omitempty"`
}

// D3Edge is one directed connection between node ids.
type D3Edge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Label string `json:"label,omitempty"`
}

// D3Spec is the persisted diagrams/<id>.json for a d3-kind diagram.
type D3Spec struct {
	Kind   string   `json:"kind"` // always "d3" — lets the renderer tell specs apart
	Layout string   `json:"layout"`
	Title  string   `json:"title,omitempty"`
	Nodes  []D3Node `json:"nodes"`
	Edges  []D3Edge `json:"edges"`
}

// Validate enforces the structural contract: a supported layout, unique
// non-empty node ids, edges that reference declared nodes, and — for the
// hierarchy layouts — a single root with no cycles so the tree is well-formed.
func (s *D3Spec) Validate() error {
	switch s.Layout {
	case D3LayoutTree, D3LayoutRadial, D3LayoutForce:
	case "":
		return fmt.Errorf("layout is required (one of tree, radial, force)")
	default:
		return fmt.Errorf("layout %q is not one of tree, radial, force", s.Layout)
	}
	if len(s.Nodes) == 0 {
		return fmt.Errorf("at least one node is required")
	}
	if len(s.Nodes) > 40 {
		return fmt.Errorf("too many nodes (%d) — keep diagrams under 40 for legibility", len(s.Nodes))
	}
	ids := make(map[string]bool, len(s.Nodes))
	for i, n := range s.Nodes {
		if n.ID == "" {
			return fmt.Errorf("nodes[%d]: id is required", i)
		}
		if n.Label == "" {
			return fmt.Errorf("node %q: label is required", n.ID)
		}
		if ids[n.ID] {
			return fmt.Errorf("duplicate node id %q", n.ID)
		}
		ids[n.ID] = true
	}
	indeg := make(map[string]int, len(s.Nodes))
	for i, e := range s.Edges {
		if !ids[e.From] {
			return fmt.Errorf("edges[%d]: from %q is not a declared node", i, e.From)
		}
		if !ids[e.To] {
			return fmt.Errorf("edges[%d]: to %q is not a declared node", i, e.To)
		}
		if e.From == e.To {
			return fmt.Errorf("edges[%d]: self-loop on %q is not allowed", i, e.From)
		}
		indeg[e.To]++
	}
	if s.Layout == D3LayoutTree || s.Layout == D3LayoutRadial {
		if err := validateHierarchy(s, indeg); err != nil {
			return err
		}
	}
	return nil
}

// validateHierarchy checks a tree/radial spec forms a single rooted tree:
// exactly one root (in-degree 0), every non-root has in-degree 1, and every
// node is reachable from the root (no cycles, no forest).
func validateHierarchy(s *D3Spec, indeg map[string]int) error {
	var roots []string
	for _, n := range s.Nodes {
		switch indeg[n.ID] {
		case 0:
			roots = append(roots, n.ID)
		case 1:
		default:
			return fmt.Errorf("node %q has %d parents — a %s layout needs a tree (one parent each)", n.ID, indeg[n.ID], s.Layout)
		}
	}
	if len(roots) != 1 {
		return fmt.Errorf("a %s layout needs exactly one root (node with no incoming edge), found %d", s.Layout, len(roots))
	}
	children := map[string][]string{}
	for _, e := range s.Edges {
		children[e.From] = append(children[e.From], e.To)
	}
	seen := map[string]bool{}
	stack := []string{roots[0]}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[id] {
			return fmt.Errorf("cycle detected at node %q — a %s layout must be acyclic", id, s.Layout)
		}
		seen[id] = true
		stack = append(stack, children[id]...)
	}
	if len(seen) != len(s.Nodes) {
		return fmt.Errorf("%d node(s) are unreachable from the root — a %s layout must be fully connected", len(s.Nodes)-len(seen), s.Layout)
	}
	return nil
}

// d3PromptData feeds prompts/diagram_d3.tmpl.
type d3PromptData struct {
	Title    string
	ID       string
	Prompt   string
	Colors   config.Colors
	Critique string
	// Narration is the spoken text of the cueing section (context).
	Narration string
	Audience  string
}

// generateD3Diagram asks the content model for a structured node-link spec and
// validates it; the completeJSON repair loop re-prompts on malformed or
// structurally invalid JSON.
func generateD3Diagram(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec project.DiagramSpec, critique string) (*D3Spec, error) {
	data := d3PromptData{
		Title:     l.FrontMatter.Title,
		ID:        spec.ID,
		Prompt:    spec.Prompt,
		Colors:    cfg.Branding.Colors,
		Critique:  critique,
		Narration: spec.Narration,
		Audience:  cfg.Style.Audience,
	}
	system, user, err := e.renderPrompt(d3DiagramTemplateName, data)
	if err != nil {
		return nil, err
	}
	out := &D3Spec{}
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.3, 4096, out, out.Validate)
	if err != nil {
		return nil, fmt.Errorf("generating d3 diagram %q: %w", spec.ID, err)
	}
	out.Kind = project.DiagramKindD3
	return out, nil
}

// produceD3Diagram generates and marshals a d3 spec. The structural Validate in
// the generation loop is the quality gate (no PNG to vision-review at this
// stage — the graph is laid out and drawn client-side in the renderer).
func produceD3Diagram(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, spec project.DiagramSpec) ([]byte, error) {
	d3, err := generateD3Diagram(ctx, e, l, cfg, spec, "")
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(d3, "", "  ")
}

// diagramArtifactName is the on-disk file for a diagram, by kind.
func diagramArtifactName(spec project.DiagramSpec) string {
	if spec.ResolvedKind() == project.DiagramKindD3 {
		return spec.ID + ".json"
	}
	return spec.ID + ".svg"
}

// diagramSceneSrc is the scene-graph src (relative to the diagrams dir) for a
// diagram id of a given kind.
func diagramSceneSrc(id, kind string) string {
	if kind == project.DiagramKindD3 {
		return DiagramsDirName + "/" + id + ".json"
	}
	return DiagramsDirName + "/" + id + ".svg"
}

// diagramKindByID indexes a lesson's declared diagram kinds for the scenegraph.
func diagramKindByID(l *project.Lesson) map[string]string {
	out := make(map[string]string, len(l.FrontMatter.Diagrams))
	for _, d := range l.FrontMatter.Diagrams {
		out[d.ID] = d.ResolvedKind()
	}
	return out
}
