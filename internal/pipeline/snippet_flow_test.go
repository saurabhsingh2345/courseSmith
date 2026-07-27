package pipeline

import (
	"strings"
	"testing"
)

// flowPlan is a well-formed six-beat plan: a framing beat, three beats that
// build a forking graph, and two beats that focus different paths through it.
func flowPlan() *SnippetPlan {
	n := func(s string) string { return strings.Repeat(s+" ", 22) }
	return &SnippetPlan{
		Template: "flow",
		Title:    "How a request gets rate limited",
		Subtitle: "Two paths through the gateway",
		Beats: []SnippetBeat{
			{ID: "the-problem", Heading: "The problem", Narration: n("problem")},
			{ID: "the-gateway", Heading: "The gateway", Narration: n("gateway"), Nodes: []FlowNode{
				{ID: "client", Label: "Client", Kind: "client"},
				{ID: "gateway", Label: "API gateway", Kind: "service", From: []string{"client"}},
			}},
			{ID: "the-counter", Heading: "The counter", Narration: n("counter"), Nodes: []FlowNode{
				{ID: "counter", Label: "Rate counter", Kind: "cache", From: []string{"gateway"}},
			}},
			{ID: "the-upstream", Heading: "The upstream", Narration: n("upstream"), Nodes: []FlowNode{
				{ID: "upstream", Label: "Orders service", Kind: "service", From: []string{"gateway"}},
				{ID: "db", Label: "Postgres", Kind: "store", From: []string{"upstream"}},
			}},
			{ID: "over-the-limit", Heading: "Over the limit", Narration: n("over"),
				Focus: []string{"client", "gateway", "counter"}},
			{ID: "under-the-limit", Heading: "Under the limit", Narration: n("under"),
				Focus: []string{"gateway", "upstream", "db"}},
		},
	}
}

func TestValidateFlowPlan(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		if err := flowPlan().Validate(); err != nil {
			t.Fatalf("want valid, got %v", err)
		}
	})
	t.Run("too few nodes", func(t *testing.T) {
		p := flowPlan()
		p.Beats[3].Nodes = nil
		p.Beats[5].Focus = []string{"client"}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "diagram has 3 nodes") {
			t.Fatalf("want node-count error, got %v", err)
		}
	})
	t.Run("duplicate node id", func(t *testing.T) {
		p := flowPlan()
		p.Beats[2].Nodes[0].ID = "gateway"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "appears twice") {
			t.Fatalf("want duplicate-id error, got %v", err)
		}
	})
	// Edges must point backwards in declaration order. That is what makes the
	// graph acyclic, and a cycle has no valid layering at all.
	t.Run("forward reference", func(t *testing.T) {
		p := flowPlan()
		p.Beats[1].Nodes[0].From = []string{"db"}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "not declared yet") {
			t.Fatalf("want forward-reference error, got %v", err)
		}
	})
	// A straight chain wastes the layered layout; the whiteboard template draws
	// chains better, so this one insists on a branch.
	t.Run("straight chain", func(t *testing.T) {
		p := flowPlan()
		p.Beats[3].Nodes = []FlowNode{
			{ID: "upstream", Label: "Orders service", Kind: "service", From: []string{"counter"}},
			{ID: "db", Label: "Postgres", Kind: "store", From: []string{"upstream"}},
		}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "straight chain") {
			t.Fatalf("want no-branch error, got %v", err)
		}
	})
	t.Run("join counts as a branch", func(t *testing.T) {
		p := flowPlan()
		// One node fed by two: a join, not a fork, but still a branch.
		p.Beats[3].Nodes = []FlowNode{
			{ID: "upstream", Label: "Orders service", Kind: "service", From: []string{"gateway"}},
			{ID: "db", Label: "Postgres", Kind: "store", From: []string{"upstream", "counter"}},
		}
		if err := p.Validate(); err != nil {
			t.Fatalf("a join should satisfy the branch rule: %v", err)
		}
	})
	t.Run("focus covering everything", func(t *testing.T) {
		p := flowPlan()
		p.Beats[4].Focus = []string{"client", "gateway", "counter", "upstream", "db"}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "focuses every node") {
			t.Fatalf("want focus-everything error, got %v", err)
		}
	})
	t.Run("duplicate focus sets", func(t *testing.T) {
		p := flowPlan()
		// Same set, listed in a different order — still the same picture.
		p.Beats[5].Focus = []string{"counter", "client", "gateway"}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "same nodes as an earlier beat") {
			t.Fatalf("want duplicate-focus error, got %v", err)
		}
	})
	t.Run("focus on unknown node", func(t *testing.T) {
		p := flowPlan()
		p.Beats[4].Focus = []string{"nope"}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "not a node in the diagram") {
			t.Fatalf("want unknown-focus error, got %v", err)
		}
	})
	t.Run("rejects whiteboard fields", func(t *testing.T) {
		p := flowPlan()
		p.Beats[1].Sketch = []SketchItem{{Label: "Nope", Icon: "dot"}}
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "does not use") {
			t.Fatalf("want wrong-template-field error, got %v", err)
		}
	})
	t.Run("label too long", func(t *testing.T) {
		p := flowPlan()
		p.Beats[2].Nodes[0].Label = "a label with far too many words"
		if err := p.Validate(); err == nil || !strings.Contains(err.Error(), "labels are at most") {
			t.Fatalf("want label-length error, got %v", err)
		}
	})
}

func TestFlowScenes(t *testing.T) {
	plan := flowPlan()
	scenes, err := flowScenes(sceneInput(t, plan, 9000))
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != 2 {
		t.Fatalf("got %d scenes, want a title card plus a diagram", len(scenes))
	}
	title, graph := scenes[0], scenes[1]
	if title.Type != SceneTitle || graph.Type != SceneFlow {
		t.Fatalf("scene types = %s, %s; want title, flow", title.Type, graph.Type)
	}
	if got := title.EndMs - title.StartMs; got > maxTitleCardMs {
		t.Errorf("title card runs %dms, want at most %d", got, maxTitleCardMs)
	}
	if graph.StartMs != title.EndMs {
		t.Errorf("diagram starts at %d but the title ends at %d", graph.StartMs, title.EndMs)
	}
	if graph.EndMs != 54000 {
		t.Errorf("diagram ends at %d, want the end of the clip (54000)", graph.EndMs)
	}

	if got := graph.Props["ranks"]; got != 4 {
		t.Errorf("ranks = %v, want 4 (client → gateway → {counter, upstream} → db)", got)
	}
	nodes, ok := graph.Props["nodes"].([]map[string]any)
	if !ok {
		t.Fatalf("nodes have the wrong shape: %#v", graph.Props["nodes"])
	}
	if len(nodes) != 5 {
		t.Fatalf("got %d nodes, want 5", len(nodes))
	}

	// Longest-path layering, and the fork's two children share a column with
	// distinct row orders — that is what stops them overlapping.
	wantRank := map[string]int{"client": 0, "gateway": 1, "counter": 2, "upstream": 2, "db": 3}
	byID := map[string]map[string]any{}
	for _, n := range nodes {
		id, _ := n["id"].(string)
		byID[id] = n
		if got, _ := n["rank"].(int); got != wantRank[id] {
			t.Errorf("node %q rank = %d, want %d", id, got, wantRank[id])
		}
	}
	if a, b := byID["counter"]["order"], byID["upstream"]["order"]; a == b {
		t.Errorf("the fork's two children share row order %v — they would be drawn on top of each other", a)
	}
	// Kind drives the icon, and the icon must be a real vocabulary name.
	if got, _ := byID["db"]["icon"].(string); got != "database" {
		t.Errorf("store node icon = %q, want database", got)
	}

	edges, ok := graph.Props["edges"].([]map[string]any)
	if !ok || len(edges) != 4 {
		t.Fatalf("got %v edges, want 4", graph.Props["edges"])
	}
	for _, e := range edges {
		from, _ := e["from"].(int)
		to, _ := e["to"].(int)
		if from >= to {
			t.Errorf("edge %d→%d does not point forwards", from, to)
		}
	}

	focus, ok := graph.Props["focus"].([]map[string]any)
	if !ok || len(focus) != 2 {
		t.Fatalf("got %v focus windows, want 2", graph.Props["focus"])
	}
	for i, w := range focus {
		start, _ := w["startMs"].(int)
		end, _ := w["endMs"].(int)
		ids, _ := w["nodes"].([]int)
		if end <= start {
			t.Errorf("focus window %d spans %d→%d", i, start, end)
		}
		if len(ids) != 3 {
			t.Errorf("focus window %d selects %v, want 3 nodes", i, ids)
		}
	}
}

// Depth is bounded by the stage: five columns would not leave a readable label
// width, so a chain that deep is rejected rather than squeezed.
func TestFlowScenesRejectsTooDeep(t *testing.T) {
	plan := flowPlan()
	plan.Beats[3].Nodes = []FlowNode{
		{ID: "upstream", Label: "Orders service", Kind: "service", From: []string{"counter"}},
		{ID: "db", Label: "Postgres", Kind: "store", From: []string{"upstream"}},
		{ID: "replica", Label: "Read replica", Kind: "store", From: []string{"db"}},
	}
	plan.Beats[5].Focus = []string{"gateway", "counter"}
	if _, err := flowScenes(sceneInput(t, plan, 9000)); err == nil ||
		!strings.Contains(err.Error(), "columns deep") {
		t.Fatalf("want a depth error, got %v", err)
	}
}

func TestFlowScenesRejectsTooWide(t *testing.T) {
	plan := flowPlan()
	// Five siblings all fed by the gateway: one column, five rows.
	plan.Beats[2].Nodes = []FlowNode{
		{ID: "a", Label: "A", Kind: "service", From: []string{"gateway"}},
		{ID: "b", Label: "B", Kind: "service", From: []string{"gateway"}},
		{ID: "c", Label: "C", Kind: "service", From: []string{"gateway"}},
	}
	plan.Beats[3].Nodes = []FlowNode{
		{ID: "d", Label: "D", Kind: "service", From: []string{"gateway"}},
		{ID: "e", Label: "E", Kind: "service", From: []string{"gateway"}},
	}
	plan.Beats[4].Focus = []string{"client", "gateway"}
	plan.Beats[5].Focus = []string{"gateway", "a"}
	if _, err := flowScenes(sceneInput(t, plan, 9000)); err == nil ||
		!strings.Contains(err.Error(), "fit the stage") {
		t.Fatalf("want a column-width error, got %v", err)
	}
}

func TestFlowNodeResolvedKind(t *testing.T) {
	if got := (FlowNode{Kind: "  Store "}).ResolvedKind(); got != "store" {
		t.Errorf("ResolvedKind(%q) = %q, want store", "  Store ", got)
	}
	if got := (FlowNode{Kind: "invented"}).ResolvedKind(); got != "service" {
		t.Errorf("unknown kind = %q, want the service fallback", got)
	}
	if got := (FlowNode{}).ResolvedKind(); got != "service" {
		t.Errorf("empty kind = %q, want the service fallback", got)
	}
}

// Every kind's figure has to be a name the renderer can actually draw. A miss
// here is silent and ugly: figureFor() falls back to `spark`, so the node still
// renders — as a burst that has nothing to do with what it is meant to be.
func TestFlowKindFiguresAreInVocabulary(t *testing.T) {
	for kind, figure := range flowNodeKinds {
		if !artFigureVocab[figure] {
			t.Errorf("kind %q maps to figure %q, which is not in the figure vocabulary", kind, figure)
		}
	}
	if len(FlowNodeKinds()) != len(flowNodeKinds) {
		t.Error("FlowNodeKinds() dropped an entry")
	}
}
