package pipeline

import (
	"strings"
	"testing"
)

const bdNarration = "Block out the page before anybody starts arguing about which shade of blue."

func breakdownPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "breakdown",
		Title:    "Building a whole website with no code",
		Breakdown: &BreakdownSpec{
			Phases: []BreakdownPhase{
				{Title: "Design", Detail: "Decide what it looks like first", Items: []PhaseItem{
					{Name: "Figma", Note: "Fastest if you know it", Icon: "layers"},
					{Name: "Canva", Note: "Templates for non-designers", Icon: "star"},
				}},
				{Title: "Wireframe", Detail: "Block out the page before colour", Items: []PhaseItem{
					{Name: "Whimsical", Note: "Fast, ugly, on purpose", Icon: "box"},
					{Name: "Excalidraw", Note: "Hand-drawn feel", Icon: "idea"},
				}},
				{Title: "Front end", Detail: "Turn the design into real pages", Items: []PhaseItem{
					{Name: "Webflow", Note: "Most control, steepest curve", Icon: "monitor"},
					{Name: "Framer", Note: "Fastest from design to live", Icon: "zap"},
				}},
			},
		},
		Beats: []SnippetBeat{
			{ID: "design", Heading: "Start with design", Narration: bdNarration, Breakdown: &BreakdownBeat{Show: "phase", At: 0}},
			{ID: "figma", Heading: "Why Figma", Narration: bdNarration, Breakdown: &BreakdownBeat{Show: "item", At: 0, Item: 0}},
			{ID: "wire", Heading: "Block it out", Narration: bdNarration, Breakdown: &BreakdownBeat{Show: "phase", At: 1}},
			{ID: "front", Heading: "Building pages", Narration: bdNarration, Breakdown: &BreakdownBeat{Show: "phase", At: 2}},
			{ID: "webflow", Heading: "The powerful one", Narration: bdNarration, Breakdown: &BreakdownBeat{Show: "item", At: 2, Item: 0}},
			{ID: "whole", Heading: "The whole path", Narration: bdNarration, Breakdown: &BreakdownBeat{Show: "whole"}},
		},
	}
	p.targetWords = 6 * 40
	return p
}

func TestBreakdownPlanAccepted(t *testing.T) {
	if err := validateBreakdownPlan(breakdownPlan()); err != nil {
		t.Fatalf("a well-formed breakdown was rejected: %v", err)
	}
}

// The rule that keeps this a path rather than a catalogue: an item belongs to
// the phase you are standing in.
func TestBreakdownItemsBelongToTheOpenPhase(t *testing.T) {
	p := breakdownPlan()
	// Standing in phase 2, reaching back into phase 0.
	p.Beats[4].Breakdown = &BreakdownBeat{Show: "item", At: 0, Item: 1}
	err := validateBreakdownPlan(p)
	if err == nil {
		t.Fatal("an item beat reaching into an earlier phase was accepted")
	}
	if !strings.Contains(err.Error(), "reaches back") {
		t.Errorf("the error should name what went wrong; got: %v", err)
	}
	if !strings.Contains(err.Error(), "standing in") {
		t.Errorf("the error should say where items belong; got: %v", err)
	}
}

func TestBreakdownItemsNeedTheirPhaseOpenedFirst(t *testing.T) {
	p := breakdownPlan()
	p.Beats[1].Breakdown = &BreakdownBeat{Show: "item", At: 2, Item: 0}
	err := validateBreakdownPlan(p)
	if err == nil {
		t.Fatal("an item of an unopened phase was accepted")
	}
	if !strings.Contains(err.Error(), "before that phase has been opened") {
		t.Errorf("the error should say to open the stage first; got: %v", err)
	}
}

func TestBreakdownOnlyMovesForward(t *testing.T) {
	p := breakdownPlan()
	p.Beats[3].Breakdown = &BreakdownBeat{Show: "phase", At: 0}
	err := validateBreakdownPlan(p)
	if err == nil {
		t.Fatal("a path walked backwards was accepted")
	}
	// Opening phase 0 twice is caught first, which is also correct; either way
	// the clip is rejected. Force the other branch explicitly.
	p = breakdownPlan()
	p.Beats[0].Breakdown = &BreakdownBeat{Show: "phase", At: 2}
	p.Beats[1].Breakdown = &BreakdownBeat{Show: "item", At: 2, Item: 0}
	p.Beats[3].Breakdown = &BreakdownBeat{Show: "phase", At: 0}
	p.Beats[4].Breakdown = &BreakdownBeat{Show: "item", At: 0, Item: 0}
	err = validateBreakdownPlan(p)
	if err == nil {
		t.Fatal("a path walked backwards was accepted")
	}
	if !strings.Contains(err.Error(), "only moves forward") {
		t.Errorf("the error should name the rule; got: %v", err)
	}
}

func TestBreakdownOpensEveryPhase(t *testing.T) {
	p := breakdownPlan()
	p.Beats = append(p.Beats[:3], p.Beats[5]) // "Front end" never opened
	p.targetWords = 4 * 40
	err := validateBreakdownPlan(p)
	if err == nil {
		t.Fatal("a phase with no beat was accepted")
	}
	if !strings.Contains(err.Error(), "never opened") {
		t.Errorf("the error should name the unexplained stage; got: %v", err)
	}
}

func TestBreakdownMustEndOnTheWholePath(t *testing.T) {
	p := breakdownPlan()
	p.Beats = p.Beats[:5]
	p.targetWords = 5 * 40
	err := validateBreakdownPlan(p)
	if err == nil {
		t.Fatal("a clip that never shows the whole path was accepted")
	}
	if !strings.Contains(err.Error(), "whole") {
		t.Errorf("the error should say how to close; got: %v", err)
	}
}

// The items are the reason this is not a timeline.
func TestBreakdownPhasesNeedItems(t *testing.T) {
	p := breakdownPlan()
	p.Breakdown.Phases[1].Items = p.Breakdown.Phases[1].Items[:1]
	err := validateBreakdownPlan(p)
	if err == nil {
		t.Fatal("a phase with one item was accepted")
	}
	if !strings.Contains(err.Error(), "not a timeline") {
		t.Errorf("the error should say what the items are for; got: %v", err)
	}
}

func TestBreakdownRejectsTooFewPhases(t *testing.T) {
	p := breakdownPlan()
	p.Breakdown.Phases = p.Breakdown.Phases[:2]
	p.Beats = []SnippetBeat{
		{ID: "a", Heading: "One", Narration: bdNarration, Breakdown: &BreakdownBeat{Show: "phase", At: 0}},
		{ID: "b", Heading: "Two", Narration: bdNarration, Breakdown: &BreakdownBeat{Show: "phase", At: 1}},
		{ID: "c", Heading: "All", Narration: bdNarration, Breakdown: &BreakdownBeat{Show: "whole"}},
	}
	p.targetWords = 3 * 40
	err := validateBreakdownPlan(p)
	if err == nil {
		t.Fatal("a two-phase path was accepted")
	}
	if !strings.Contains(err.Error(), "compare") {
		t.Errorf("the error should point at the template that fits; got: %v", err)
	}
}

func TestBreakdownRejectsAnItemSpotlitTwice(t *testing.T) {
	p := breakdownPlan()
	p.Beats = append(p.Beats[:5], SnippetBeat{
		ID: "again", Heading: "Webflow again", Narration: bdNarration,
		Breakdown: &BreakdownBeat{Show: "item", At: 2, Item: 0},
	}, p.Beats[5])
	p.targetWords = 7 * 40
	if err := validateBreakdownPlan(p); err == nil {
		t.Error("the same item spotlit twice was accepted")
	}
}

func TestBreakdownNormalizeClampsAndInfers(t *testing.T) {
	p := breakdownPlan()
	p.Beats[5].Breakdown.Show = "" // the last beat is the overview by shape
	p.Breakdown.Phases[0].Detail = "decide what the whole thing is going to look like before a single line of anything gets built"
	p.Breakdown.Phases[0].Items = append(p.Breakdown.Phases[0].Items, PhaseItem{Name: "", Note: "orphan"})
	p.Breakdown.Phases[1].Items[0].Icon = "wobble"
	normalizeBreakdownPlan(p)

	if p.Beats[5].Breakdown.Show != "whole" {
		t.Errorf("the last beat should be inferred as whole, got %q", p.Beats[5].Breakdown.Show)
	}
	if w := len(strings.Fields(p.Breakdown.Phases[0].Detail)); w > maxPhaseDetailWords {
		t.Errorf("detail still %d words, want at most %d", w, maxPhaseDetailWords)
	}
	if n := len(p.Breakdown.Phases[0].Items); n != 2 {
		t.Errorf("an unnamed item should be dropped, got %d items", n)
	}
	if p.Breakdown.Phases[1].Items[0].Icon != "box" {
		t.Errorf("an invented icon should fall back to box, got %q", p.Breakdown.Phases[1].Items[0].Icon)
	}
	if err := validateBreakdownPlan(p); err != nil {
		t.Fatalf("the normalized plan should validate: %v", err)
	}
}

// The template is the longest in the catalog by design; its ceiling has to fund
// the shape its own prompt asks for.
func TestBreakdownCeilingFundsSixPhasesWithItems(t *testing.T) {
	tpl := SnippetTemplates["breakdown"]
	if tpl.MaxBeats < 12 {
		t.Fatalf("six phases plus item beats plus the overview needs 12; MaxBeats is %d", tpl.MaxBeats)
	}
	spec := SnippetSpec{Template: "breakdown", Prompt: "x", TargetSec: 150}
	want, _, _ := wordBudget(spec.ResolvedTargetSec(), 175)
	_, maxBeats, _, perBeat := beatBounds(want, templateBeatCeiling("breakdown"))
	if maxBeats < 11 {
		t.Errorf("a 150s breakdown may use only %d beats, too few for a deep path", maxBeats)
	}
	if perBeat > maxWordsPerBeat {
		t.Errorf("the prompt would advise %d words a beat, over the %d maximum", perBeat, maxWordsPerBeat)
	}
}

func TestBreakdownScenesCarryBothLevels(t *testing.T) {
	p := breakdownPlan()
	scenes, err := breakdownScenes(sceneInput(t, p, 6000))
	if err != nil {
		t.Fatalf("laying out the breakdown failed: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneBreakdown {
		t.Fatalf("want one %s scene, got %v", SceneBreakdown, scenes)
	}
	phases, ok := scenes[0].Props["phases"].([]map[string]any)
	if !ok || len(phases) != 3 {
		t.Fatalf("want three phases in the props, got %v", scenes[0].Props["phases"])
	}
	// The second level has to survive, or the template is a timeline with
	// bigger rows.
	items, ok := phases[0]["items"].([]map[string]any)
	if !ok || len(items) != 2 {
		t.Fatalf("want the phase's items in the props, got %v", phases[0]["items"])
	}
	steps, ok := scenes[0].Props["steps"].([]map[string]any)
	if !ok || len(steps) != len(p.Beats) {
		t.Fatalf("want one step per beat, got %v", scenes[0].Props["steps"])
	}
	if steps[1]["show"] != "item" || steps[1]["item"] != 0 {
		t.Errorf("an item beat should reach the renderer with its index, got %v", steps[1])
	}
	if steps[len(steps)-1]["show"] != "whole" {
		t.Error("the last step should show the whole path")
	}
}
