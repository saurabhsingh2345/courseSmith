package pipeline

import "testing"

func planGraph() *SceneGraph {
	return &SceneGraph{Scenes: []Scene{
		{Type: SceneTitle, StartMs: 0, EndMs: 1000, Props: map[string]any{"heading": "A"}},
		{Type: ScenePoints, StartMs: 1000, EndMs: 3000, Props: map[string]any{"title": "B"}},
		{Type: SceneCode, StartMs: 3000, EndMs: 5000, Props: map[string]any{"code": "x"}},
	}}
}

func TestApplyVideoPlanOverrides(t *testing.T) {
	g := planGraph()
	err := applyVideoPlan(g, &VideoPlan{Edits: []VideoPlanEdit{
		{Scene: 1, Template: "grid", Props: map[string]any{"title": "Better title"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if g.Scenes[1].Props["template"] != "grid" || g.Scenes[1].Props["title"] != "Better title" {
		t.Errorf("edit not applied: %#v", g.Scenes[1].Props)
	}
}

func TestApplyVideoPlanSkipExtendsPrevious(t *testing.T) {
	g := planGraph()
	if err := applyVideoPlan(g, &VideoPlan{Edits: []VideoPlanEdit{{Scene: 1, Skip: true}}}); err != nil {
		t.Fatal(err)
	}
	if len(g.Scenes) != 2 {
		t.Fatalf("scenes = %d, want 2", len(g.Scenes))
	}
	if g.Scenes[0].EndMs != 3000 {
		t.Errorf("previous scene EndMs = %d, want 3000 (absorbed span)", g.Scenes[0].EndMs)
	}
	if g.Scenes[1].StartMs != 3000 {
		t.Errorf("next scene must be untouched, StartMs = %d", g.Scenes[1].StartMs)
	}
}

func TestApplyVideoPlanSkipHead(t *testing.T) {
	g := planGraph()
	if err := applyVideoPlan(g, &VideoPlan{Edits: []VideoPlanEdit{{Scene: 0, Skip: true}}}); err != nil {
		t.Fatal(err)
	}
	if len(g.Scenes) != 2 || g.Scenes[0].StartMs != 0 {
		t.Fatalf("head skip: next scene must start at 0, got %+v", g.Scenes[0])
	}
}

func TestApplyVideoPlanRejectsBadIndexAndFullSkip(t *testing.T) {
	if err := applyVideoPlan(planGraph(), &VideoPlan{Edits: []VideoPlanEdit{{Scene: 7}}}); err == nil {
		t.Error("out-of-range scene index must error")
	}
	if err := applyVideoPlan(planGraph(), &VideoPlan{Edits: []VideoPlanEdit{
		{Scene: 0, Skip: true}, {Scene: 1, Skip: true}, {Scene: 2, Skip: true},
	}}); err == nil {
		t.Error("skipping every scene must error")
	}
}
