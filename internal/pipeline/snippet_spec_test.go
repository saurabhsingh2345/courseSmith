package pipeline

import (
	"strings"
	"testing"
)

const spNarration = "Loads fast is a wish; first paint under two seconds is something you can check."

func specPlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "spec",
		Title:    "Write the test before you write the prompt",
		Spec: &SpecSheet{
			Goal:        "A signup form that actually converts",
			Constraints: []string{"No backend", "Ship today"},
			Criteria: []SpecCriterion{
				{Text: "One field, nothing else", Note: "Every extra field costs signups."},
				{Text: "Invalid email caught before submit", Note: "Caught in the browser."},
				{Text: "Success message replaces the form", Status: "missed", Note: "Nobody should wonder."},
				{Text: "Readable on a phone", Note: "Most traffic never sees a laptop."},
			},
		},
		Beats: []SnippetBeat{
			{ID: "one-field", Heading: "Count the fields", Narration: spNarration, Spec: &SpecBeat{At: 0}},
			{ID: "validate", Heading: "Catch it early", Narration: spNarration, Spec: &SpecBeat{At: 1}},
			{ID: "confirm", Heading: "Say it worked", Narration: spNarration, Spec: &SpecBeat{At: 2}},
			{ID: "phone", Heading: "On a phone", Narration: spNarration, Spec: &SpecBeat{At: 3}},
			{ID: "check", Heading: "Now check it", Narration: spNarration, Spec: &SpecBeat{Check: true}},
		},
	}
}

func TestSpecPlanAccepted(t *testing.T) {
	if err := validateSpecPlan(specPlan()); err != nil {
		t.Fatalf("a well-formed spec was rejected: %v", err)
	}
}

func TestSpecNeedsAGoal(t *testing.T) {
	p := specPlan()
	p.Spec.Goal = ""
	err := validateSpecPlan(p)
	if err == nil {
		t.Fatal("criteria for nothing in particular were accepted")
	}
	if !strings.Contains(err.Error(), "criteria for something") {
		t.Errorf("the error should say what is missing; got: %v", err)
	}
}

// The payoff, and the reason the two halves of the clip are separate.
func TestSpecMustEndByChecking(t *testing.T) {
	p := specPlan()
	p.Beats = p.Beats[:4]
	err := validateSpecPlan(p)
	if err == nil {
		t.Fatal("a spec that is never checked was accepted")
	}
	if !strings.Contains(err.Error(), "go green") {
		t.Errorf("the error should say where the idea lands; got: %v", err)
	}
}

func TestSpecCheckMustBeLast(t *testing.T) {
	p := specPlan()
	p.Beats[4].Spec = &SpecBeat{At: 3}
	p.Beats[3].Spec = &SpecBeat{Check: true}
	if err := validateSpecPlan(p); err == nil {
		t.Error("a clip that checks the sheet and then keeps writing was accepted")
	}
}

func TestSpecWritesTheListInOrder(t *testing.T) {
	p := specPlan()
	p.Beats[1].Spec = &SpecBeat{At: 2}
	p.Beats[2].Spec = &SpecBeat{At: 1}
	err := validateSpecPlan(p)
	if err == nil {
		t.Fatal("a list written out of order was accepted")
	}
	if !strings.Contains(err.Error(), "written in order") {
		t.Errorf("the error should name the rule; got: %v", err)
	}
}

func TestSpecWritesEveryCriterion(t *testing.T) {
	p := specPlan()
	p.Beats = append(p.Beats[:3], p.Beats[4]) // "Readable on a phone" never written
	err := validateSpecPlan(p)
	if err == nil {
		t.Fatal("a criterion with no beat was accepted")
	}
	if !strings.Contains(err.Error(), "never written") {
		t.Errorf("the error should name the unexplained line; got: %v", err)
	}
}

// A miss is a first-class outcome — most honest specs have one — but a sheet
// where nothing was met is a failure being described, and promptloop tells that
// story with somewhere for it to go next.
func TestSpecAllowsAMissButNotATotalFailure(t *testing.T) {
	if err := validateSpecPlan(specPlan()); err != nil {
		t.Fatalf("a spec with one missed line should be accepted: %v", err)
	}
	p := specPlan()
	for i := range p.Spec.Criteria {
		p.Spec.Criteria[i].Status = "missed"
	}
	err := validateSpecPlan(p)
	if err == nil {
		t.Fatal("a spec where nothing was met was accepted")
	}
	if !strings.Contains(err.Error(), "promptloop") {
		t.Errorf("the error should name the template that fits; got: %v", err)
	}
}

func TestSpecRejectsTooFewCriteria(t *testing.T) {
	p := specPlan()
	p.Spec.Criteria = p.Spec.Criteria[:2]
	p.Beats = append(p.Beats[:2], p.Beats[4])
	err := validateSpecPlan(p)
	if err == nil {
		t.Fatal("a two-line spec was accepted")
	}
	if !strings.Contains(err.Error(), "inconvenient") {
		t.Errorf("the error should say what a spec is for; got: %v", err)
	}
}

func TestSpecNormalizeClampsAndDefaults(t *testing.T) {
	p := specPlan()
	p.Spec.Criteria[0].Status = "" // the ordinary case: omitted means met
	p.Spec.Criteria[1].Status = "maybe"
	p.Spec.Criteria[1].Text = "invalid email is caught before the form is ever submitted anywhere"
	p.Spec.Constraints = []string{"No backend", "Ship today", "Free tier only", "One more"}
	normalizeSpecPlan(p)

	if p.Spec.Criteria[0].Status != "met" {
		t.Errorf("an omitted status should mean met, got %q", p.Spec.Criteria[0].Status)
	}
	if p.Spec.Criteria[1].Status != "met" {
		t.Errorf("an invented status should degrade to met, got %q", p.Spec.Criteria[1].Status)
	}
	if w := len(strings.Fields(p.Spec.Criteria[1].Text)); w > maxSpecCriterionWords {
		t.Errorf("criterion still %d words, want at most %d", w, maxSpecCriterionWords)
	}
	if n := len(p.Spec.Constraints); n > maxSpecConstraints {
		t.Errorf("kept %d constraints, want at most %d", n, maxSpecConstraints)
	}
	if err := validateSpecPlan(p); err != nil {
		t.Fatalf("the normalized plan should validate: %v", err)
	}
}

func TestSpecScenesCoverTheWholeClip(t *testing.T) {
	p := specPlan()
	scenes, err := specScenes(sceneInput(t, p, 6000))
	if err != nil {
		t.Fatalf("laying out the spec failed: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneSpec {
		t.Fatalf("want one %s scene, got %v", SceneSpec, scenes)
	}
	criteria, ok := scenes[0].Props["criteria"].([]map[string]any)
	if !ok || len(criteria) != 4 {
		t.Fatalf("want four criteria in the props, got %v", scenes[0].Props["criteria"])
	}
	// The renderer draws a miss by its shape, so the status has to survive.
	if criteria[2]["status"] != "missed" {
		t.Errorf("the missed line should reach the renderer as missed, got %v", criteria[2]["status"])
	}
	steps, ok := scenes[0].Props["steps"].([]map[string]any)
	if !ok || len(steps) != len(p.Beats) {
		t.Fatalf("want one step per beat, got %v", scenes[0].Props["steps"])
	}
	if _, isCheck := steps[len(steps)-1]["check"]; !isCheck {
		t.Error("the last step should check the sheet")
	}
}
