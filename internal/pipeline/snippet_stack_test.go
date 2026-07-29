package pipeline

import (
	"strings"
	"testing"
)

const stNarration = "This is the tier where the records actually live, and everything else reads from it."

func stackPlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "stack",
		Title:    "The four tools behind a no-code job board",
		Stack: &StackSpec{
			Layers: []StackLayer{
				{Name: "Frontend", Role: "What the person actually looks at", Tools: []StackTool{
					{Name: "Softr", Icon: "monitor", Note: "Fastest if your data is Airtable"},
					{Name: "Framer", Icon: "layers", Note: "Better when design matters"},
				}},
				{Name: "Automation", Role: "The glue between everything else", Tools: []StackTool{
					{Name: "Make", Icon: "shuffle", Note: "Visual, cheap, forgiving"},
				}},
				{Name: "Data", Role: "Where the records actually live", Tools: []StackTool{
					{Name: "Airtable", Icon: "database", Note: "A spreadsheet-shaped database"},
				}},
			},
		},
		Beats: []SnippetBeat{
			{ID: "front", Heading: "What people see", Narration: stNarration, Stack: &StackBeat{At: 0}},
			{ID: "glue", Heading: "The glue layer", Narration: stNarration, Stack: &StackBeat{At: 1}},
			{ID: "data", Heading: "Where it lives", Narration: stNarration, Stack: &StackBeat{At: 2}},
			{ID: "whole", Heading: "All three together", Narration: stNarration, Stack: &StackBeat{Whole: true}},
		},
	}
}

func TestStackPlanAccepted(t *testing.T) {
	if err := validateStackPlan(stackPlan()); err != nil {
		t.Fatalf("a well-formed stack was rejected: %v", err)
	}
}

// The walk goes down. Hopping between tiers is describing a request's path, and
// the error names the template that draws paths.
func TestStackOnlyWalksDown(t *testing.T) {
	p := stackPlan()
	p.Beats[0].Stack = &StackBeat{At: 2}
	p.Beats[2].Stack = &StackBeat{At: 0}
	err := validateStackPlan(p)
	if err == nil {
		t.Fatal("a stack walked out of order was accepted")
	}
	if !strings.Contains(err.Error(), "walked downward") {
		t.Errorf("the error should name the rule; got: %v", err)
	}
	if !strings.Contains(err.Error(), "flow") {
		t.Errorf("the error should point at the template that fits; got: %v", err)
	}
}

// The whole claim of a stack is that a tool has a job.
func TestStackRejectsAToolInTwoLayers(t *testing.T) {
	p := stackPlan()
	p.Stack.Layers[2].Tools = append(p.Stack.Layers[2].Tools, StackTool{Name: "Softr", Icon: "monitor"})
	err := validateStackPlan(p)
	if err == nil {
		t.Fatal("the same product in two tiers was accepted")
	}
	if !strings.Contains(err.Error(), "one job") {
		t.Errorf("the error should say what a stack claims; got: %v", err)
	}
}

func TestStackNarratesEveryLayer(t *testing.T) {
	p := stackPlan()
	p.Beats = append(p.Beats[:2], p.Beats[3]) // "Data" never reached
	err := validateStackPlan(p)
	if err == nil {
		t.Fatal("a tier with no beat was accepted")
	}
	if !strings.Contains(err.Error(), "never narrated") {
		t.Errorf("the error should name the unexplained tier; got: %v", err)
	}
}

func TestStackMustEndOnTheWholeStack(t *testing.T) {
	p := stackPlan()
	p.Beats = p.Beats[:3]
	err := validateStackPlan(p)
	if err == nil {
		t.Fatal("a clip that never shows the whole stack was accepted")
	}
	if !strings.Contains(err.Error(), "tool reviews") {
		t.Errorf("the error should say what the clip degenerates into; got: %v", err)
	}
}

func TestStackNeedsARolePerLayer(t *testing.T) {
	p := stackPlan()
	p.Stack.Layers[1].Role = ""
	err := validateStackPlan(p)
	if err == nil {
		t.Fatal("a tier with no stated job was accepted")
	}
	if !strings.Contains(err.Error(), "logos") {
		t.Errorf("the error should say what a nameless tier is; got: %v", err)
	}
}

func TestStackRejectsAnEmptyLayer(t *testing.T) {
	p := stackPlan()
	p.Stack.Layers[1].Tools = nil
	if err := validateStackPlan(p); err == nil {
		t.Error("a tier with no tools in it was accepted")
	}
}

func TestStackNormalizeClampsAndDropsNamelessTools(t *testing.T) {
	p := stackPlan()
	p.Stack.Layers[0].Tools = append(p.Stack.Layers[0].Tools, StackTool{Name: "", Icon: "monitor"})
	p.Stack.Layers[0].Name = "Frontend and the whole presentation tier"
	p.Stack.Layers[2].Tools[0].Icon = "wobble"
	normalizeStackPlan(p)

	if n := len(p.Stack.Layers[0].Tools); n != 2 {
		t.Errorf("a card with no name should be dropped, got %d tools", n)
	}
	if w := len(strings.Fields(p.Stack.Layers[0].Name)); w > maxStackNameWords {
		t.Errorf("layer name still %d words, want at most %d", w, maxStackNameWords)
	}
	if p.Stack.Layers[2].Tools[0].Icon != "box" {
		t.Errorf("an invented icon should fall back to box, got %q", p.Stack.Layers[2].Tools[0].Icon)
	}
	if err := validateStackPlan(p); err != nil {
		t.Fatalf("the normalized plan should validate: %v", err)
	}
}

func TestStackScenesCoverTheWholeClip(t *testing.T) {
	p := stackPlan()
	scenes, err := stackScenes(sceneInput(t, p, 7000))
	if err != nil {
		t.Fatalf("laying out the stack failed: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneStack {
		t.Fatalf("want one %s scene, got %v", SceneStack, scenes)
	}
	layers, ok := scenes[0].Props["layers"].([]map[string]any)
	if !ok || len(layers) != 3 {
		t.Fatalf("want three layers in the props, got %v", scenes[0].Props["layers"])
	}
	steps, ok := scenes[0].Props["steps"].([]map[string]any)
	if !ok || len(steps) != len(p.Beats) {
		t.Fatalf("want one step per beat, got %v", scenes[0].Props["steps"])
	}
	if _, isWhole := steps[len(steps)-1]["whole"]; !isWhole {
		t.Error("the last step should show the whole stack")
	}
}
