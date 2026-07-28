package pipeline

import (
	"strings"
	"testing"
)

const cmpNarration = "Here is the way most people write this when they first meet it."

func comparePlan() *SnippetPlan {
	return &SnippetPlan{
		Template: "compare",
		Title:    "Two ways to build a list",
		Compare: &CompareSpec{
			Left:    CompareSide{Label: "A for loop", Code: "out = []\nfor x in xs:\n    out.append(x * 2)", Note: "3 lines, one mutation"},
			Right:   CompareSide{Label: "A comprehension", Code: "out = [x * 2 for x in xs]", Note: "1 line, no mutation"},
			Winner:  "right",
			Verdict: "When the loop only builds a list, the comprehension says so in one line.",
		},
		Beats: []SnippetBeat{
			{ID: "loop", Heading: "The familiar way", Narration: cmpNarration, Compare: &CompareBeat{Show: "left"}},
			{ID: "other", Heading: "The other way", Narration: cmpNarration, Compare: &CompareBeat{Show: "right"}},
			{ID: "together", Heading: "Side by side", Narration: cmpNarration, Compare: &CompareBeat{Show: "both"}},
			{ID: "verdict", Heading: "Which to reach for", Narration: cmpNarration, Compare: &CompareBeat{Show: "verdict"}},
		},
	}
}

func TestComparePlanAccepted(t *testing.T) {
	if err := validateComparePlan(comparePlan()); err != nil {
		t.Fatalf("a well-formed comparison was rejected: %v", err)
	}
}

// The `both` beat is the one the template exists for. A clip that describes two
// things separately and then announces a winner has compared nothing — and it
// is the beat a model will drop, because each half reads fine on its own.
func TestCompareRequiresBothLitTogether(t *testing.T) {
	p := comparePlan()
	p.Beats = []SnippetBeat{
		{ID: "loop", Heading: "h", Narration: cmpNarration, Compare: &CompareBeat{Show: "left"}},
		{ID: "other", Heading: "h", Narration: cmpNarration, Compare: &CompareBeat{Show: "right"}},
		{ID: "verdict", Heading: "h", Narration: cmpNarration, Compare: &CompareBeat{Show: "verdict"}},
	}
	err := validateComparePlan(p)
	if err == nil {
		t.Fatal("a comparison with no `both` beat was accepted")
	}
	if !strings.Contains(err.Error(), "both") {
		t.Errorf("the error should name the missing beat; got: %v", err)
	}
}

func TestCompareRequiresLeftBeforeRight(t *testing.T) {
	p := comparePlan()
	p.Beats[0].Compare = &CompareBeat{Show: "right"}
	p.Beats[1].Compare = &CompareBeat{Show: "left"}
	if err := validateComparePlan(p); err == nil {
		t.Error("the right column was introduced first and accepted")
	}
}

func TestCompareRequiresAVerdictLast(t *testing.T) {
	p := comparePlan()
	p.Beats = p.Beats[:3] // drop the verdict
	if err := validateComparePlan(p); err == nil {
		t.Error("a comparison with no verdict was accepted")
	}

	p = comparePlan()
	// A verdict before the two have been seen together decides nothing.
	p.Beats[2].Compare = &CompareBeat{Show: "verdict"}
	p.Beats[3].Compare = &CompareBeat{Show: "both"}
	if err := validateComparePlan(p); err == nil {
		t.Error("a verdict before the `both` beat was accepted")
	}
}

func TestCompareColumnHoldsOneThing(t *testing.T) {
	p := comparePlan()
	p.Compare.Left.Figure = "gears" // it already has code
	err := validateComparePlan(p)
	if err == nil {
		t.Fatal("a column with both code and a figure was accepted")
	}

	p = comparePlan()
	p.Compare.Right.Code = ""
	if err := validateComparePlan(p); err == nil {
		t.Error("an empty column was accepted")
	}
}

func TestCompareRejectsIdenticalLabels(t *testing.T) {
	p := comparePlan()
	p.Compare.Right.Label = "a for LOOP"
	if err := validateComparePlan(p); err == nil {
		t.Error("two columns with the same label were accepted — they are meant to differ")
	}
}

// A tie is a first-class answer. Most honest comparisons are ties, and forcing
// a winner would make the template lie in exactly the cases where the teaching
// is most useful.
func TestCompareAcceptsATie(t *testing.T) {
	p := comparePlan()
	p.Compare.Winner = "tie"
	if err := validateComparePlan(p); err != nil {
		t.Errorf("a tie was rejected: %v", err)
	}
	p.Compare.Winner = "whichever"
	if err := validateComparePlan(p); err == nil {
		t.Error("an invented winner value was accepted")
	}
}

func TestCompareRejectsOversizedColumns(t *testing.T) {
	p := comparePlan()
	p.Compare.Left.Code = strings.Repeat("x = 1\n", maxCompareCodeLines+3)
	if err := validateComparePlan(p); err == nil {
		t.Error("a column far over the line cap was accepted")
	}

	p = comparePlan()
	p.Compare.Left.Note = strings.Repeat("word ", maxCompareNoteWords+4)
	if err := validateComparePlan(p); err == nil {
		t.Error("a note far over the word cap was accepted — it is a measurement, not a sentence")
	}
}

func TestCompareScenesIsOneSceneWithBothColumns(t *testing.T) {
	plan := comparePlan()
	scenes, err := compareScenes(sceneInput(t, plan, 6000))
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneCompare {
		t.Fatalf("want one compare scene, got %d scenes", len(scenes))
	}
	// Both columns must be in the props from the first frame: the scene lights
	// them, it does not mount them, or the layout would move mid-comparison.
	for _, key := range []string{"left", "right", "winner", "verdict"} {
		if _, ok := scenes[0].Props[key]; !ok {
			t.Errorf("the scene is missing %q", key)
		}
	}
	steps, ok := scenes[0].Props["steps"].([]map[string]any)
	if !ok || len(steps) != len(plan.Beats) {
		t.Fatalf("want one step per beat, got %#v", scenes[0].Props["steps"])
	}
	if steps[3]["show"] != "verdict" {
		t.Errorf("the last step is %v, want the verdict", steps[3]["show"])
	}
}

// A column with a figure carries a name the renderer can draw. figureFor falls
// back to `spark`, so an unnormalized name renders a burst that has nothing to
// do with the side it is arguing for.
func TestCompareFigureSideIsNormalized(t *testing.T) {
	plan := comparePlan()
	plan.Compare.Left.Code = ""
	plan.Compare.Left.Figure = "not-a-real-figure"
	scenes, err := compareScenes(sceneInput(t, plan, 6000))
	if err != nil {
		t.Fatal(err)
	}
	left, _ := scenes[0].Props["left"].(map[string]any)
	if got := left["figure"]; got != "spark" {
		t.Errorf("left figure = %v, want the spark fallback", got)
	}
	if _, hasCode := left["code"]; hasCode {
		t.Error("a figure column also carried code")
	}
}

func TestNormalizeCompareShow(t *testing.T) {
	for _, s := range CompareShowNames() {
		if got := normalizeCompareShow(s); got != s {
			t.Errorf("normalizeCompareShow(%q) = %q, want it preserved", s, got)
		}
	}
	// `both` is the fallback: an invented name becoming a second verdict or a
	// stray introduction would break the shape the validation depends on.
	for _, bad := range []string{"", "winner", "  "} {
		if got := normalizeCompareShow(bad); got != "both" {
			t.Errorf("normalizeCompareShow(%q) = %q, want the both fallback", bad, got)
		}
	}
}

// The normalizer repairs what has one sensible repair, so those never cost a
// correction round.
func TestNormalizeComparePlanRepairsTheMechanical(t *testing.T) {
	p := comparePlan()
	p.Compare.Left.Label = "  A   really quite long   loop label  "
	p.Compare.Left.Note = strings.Repeat("word ", maxCompareNoteWords+4)
	p.Compare.Winner = "  RIGHT "
	normalizeComparePlan(p)

	if n := len(strings.Fields(p.Compare.Left.Label)); n > maxCompareLabelWords {
		t.Errorf("label was not clamped: %d words", n)
	}
	if n := len(strings.Fields(p.Compare.Left.Note)); n > maxCompareNoteWords {
		t.Errorf("note was not clamped: %d words", n)
	}
	if p.Compare.Winner != "right" {
		t.Errorf("winner = %q, want it lowercased and trimmed", p.Compare.Winner)
	}
	if err := validateComparePlan(p); err != nil {
		t.Errorf("a normalized plan should validate: %v", err)
	}
}

// Models answer with the column's label about as often as with the side. That
// is the same answer in the other vocabulary, not a mistake about the clip.
func TestNormalizeCompareMatchesWinnerByLabel(t *testing.T) {
	p := comparePlan()
	p.Compare.Winner = "A comprehension" // the right column's label
	normalizeComparePlan(p)
	if p.Compare.Winner != "right" {
		t.Errorf("winner = %q, want it matched back to the right column", p.Compare.Winner)
	}

	for _, synonym := range []string{"neither", "both", "draw"} {
		p = comparePlan()
		p.Compare.Winner = synonym
		normalizeComparePlan(p)
		if p.Compare.Winner != "tie" {
			t.Errorf("winner %q normalized to %q, want tie", synonym, p.Compare.Winner)
		}
	}
}

// A winner that resolves to neither column is a claim, and quietly rewriting a
// claim is a different act from tidying a label — validation still rejects it.
func TestNormalizeCompareLeavesAnUnresolvableWinnerAlone(t *testing.T) {
	p := comparePlan()
	p.Compare.Winner = "the fast one"
	normalizeComparePlan(p)
	if err := validateComparePlan(p); err == nil {
		t.Error("an unresolvable winner survived normalization and validation")
	}
}

// A figure column names something the renderer can draw; figureFor otherwise
// falls back to a burst that argues for nothing.
func TestNormalizeCompareFixesFigureNames(t *testing.T) {
	p := comparePlan()
	p.Compare.Left.Code = ""
	p.Compare.Left.Figure = "  GEARS "
	normalizeComparePlan(p)
	if p.Compare.Left.Figure != "gears" {
		t.Errorf("figure = %q, want the normalized name", p.Compare.Left.Figure)
	}
}
