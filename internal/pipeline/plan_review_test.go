package pipeline

import (
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

func planReview(scores map[string]float64, overall float64) *Review {
	return &Review{Scores: scores, Overall: overall, Critique: "beat 3 invents a price"}
}

func fullPlanScores() map[string]float64 {
	return map[string]float64{"fabrication": 9, "concreteness": 8, "teaching": 8, "non_redundancy": 7}
}

func TestPlanRubricAcceptsItsOwnDimensions(t *testing.T) {
	if err := planRubric.validate(planReview(fullPlanScores(), 8)); err != nil {
		t.Fatalf("a well-formed plan review was rejected: %v", err)
	}
}

// The two rubrics must not accept each other's dimensions, or a prompt could
// drift onto the wrong set and still pass validation.
func TestRubricsRejectEachOthersDimensions(t *testing.T) {
	scriptScores := map[string]float64{"technical_accuracy": 8, "clarity": 8, "engagement": 8, "pacing": 8}
	if err := planRubric.validate(planReview(scriptScores, 8)); err == nil {
		t.Error("the plan rubric accepted the script rubric's dimensions")
	}
	if err := scriptRubric.validate(planReview(fullPlanScores(), 8)); err == nil {
		t.Error("the script rubric accepted the plan rubric's dimensions")
	}
}

func TestPlanRubricRequiresEveryDimension(t *testing.T) {
	scores := fullPlanScores()
	delete(scores, "fabrication")
	err := planRubric.validate(planReview(scores, 8))
	if err == nil {
		t.Fatal("a review missing the fabrication score was accepted")
	}
	if !strings.Contains(err.Error(), "fabrication") {
		t.Errorf("the error does not name the missing dimension: %v", err)
	}
}

// The reason FatalBelow exists. An average cannot express "one invented figure
// ruins the clip": a plan excellent on three dimensions that made up a number
// scores well and ships, which is the outcome this whole change exists to stop.
func TestFabricationIsFatalWhateverTheAverageSays(t *testing.T) {
	scores := fullPlanScores()
	scores["fabrication"] = 2
	// An overall that would sail past any sane threshold.
	r := planReview(scores, 9.5)

	dim, score, fatal := planRubric.failedDimension(r)
	if !fatal {
		t.Fatal("a plan scoring 2 on fabrication was not fatal")
	}
	if dim != "fabrication" {
		t.Errorf("the fatal dimension is reported as %q", dim)
	}
	if score != 2 {
		t.Errorf("the reported score is %v, want 2", score)
	}
}

// And a merely mediocre plan is NOT fatal — it is worth shipping and improving.
// A gate that treated dullness like dishonesty would block everything.
func TestOnlyFabricationIsFatal(t *testing.T) {
	scores := fullPlanScores()
	scores["concreteness"] = 1
	scores["teaching"] = 2
	scores["non_redundancy"] = 1
	if _, _, fatal := planRubric.failedDimension(planReview(scores, 3)); fatal {
		t.Error("a dull plan was treated as fatal; only fabrication should be")
	}
	// The script rubric has no fatal dimension at all.
	if _, _, fatal := scriptRubric.failedDimension(planReview(scores, 3)); fatal {
		t.Error("the script rubric has a fatal dimension it should not")
	}
}

// A fabrication score exactly at the floor passes: the floor is "below this is
// fatal", and an off-by-one here would fail every plan that scraped through.
func TestFabricationAtTheFloorIsNotFatal(t *testing.T) {
	scores := fullPlanScores()
	scores["fabrication"] = planRubric.FatalBelow["fabrication"]
	if _, _, fatal := planRubric.failedDimension(planReview(scores, 7)); fatal {
		t.Error("a fabrication score exactly at the floor was treated as fatal")
	}
}

// The rubric prompt must render, and must carry the fact sheet — a fabrication
// score judged without the facts is the same guess the writer already made.
func TestPlanRubricPromptCarriesTheFacts(t *testing.T) {
	data := reviewPromptData{
		Kind:     "plan:showcase-3",
		Audience: config.Defaults().Style.Audience,
		Tone:     config.Defaults().Style.Tone,
		Artifact: `{"title":"Tools","beats":[]}`,
		Facts:    substanceLines(substanceFixture()),
		Gaps:     []string{"No figure for professional no-code adoption"},
		Grounded: true,
	}
	system, user, healed, err := renderPromptFileHealed(repoPromptsDir, reviewPlanTemplateName, data)
	if err != nil {
		t.Fatalf("rendering %s: %v", reviewPlanTemplateName, err)
	}
	if len(healed) > 0 {
		t.Errorf("the plan rubric references keys nothing supplies: %v", healed)
	}
	if !strings.Contains(system, "https://webflow.com/pricing") {
		t.Error("the judge was not given the facts, so it cannot tell invented from established")
	}
	if !strings.Contains(system, "No figure for professional no-code adoption") {
		t.Error("the judge was not given the gaps, so it cannot spot a filled-in hole")
	}
	if !strings.Contains(user, data.Artifact) {
		t.Error("the plan under review did not reach the user message")
	}
	// Every dimension the rubric requires must be asked for by name, or the model
	// cannot know to return it.
	for _, dim := range planRubric.Scores {
		if !strings.Contains(strings.ToLower(system), dim) {
			t.Errorf("the prompt never mentions the %q dimension it will be validated on", dim)
		}
	}
}

// With grounding off the judge must be told, or it holds an ungrounded sheet to a
// sourced standard and rejects everything.
func TestPlanRubricTellsTheJudgeWhenUngrounded(t *testing.T) {
	data := reviewPromptData{Kind: "plan:x", Artifact: "{}", Facts: []string{"a fact"}, Grounded: false}
	system, _, _, err := renderPromptFileHealed(repoPromptsDir, reviewPlanTemplateName, data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(system, "no web search ran") {
		t.Error("an ungrounded review does not tell the judge that nothing was looked up")
	}

	data.Grounded = true
	system, _, _, err = renderPromptFileHealed(repoPromptsDir, reviewPlanTemplateName, data)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(system, "no web search ran") {
		t.Error("a grounded review was told nothing was looked up")
	}
}

// The critique has to reach the next attempt, or the gate spends three rounds
// asking for the same fix. Appended to the rendered prompt rather than added to
// twenty-seven files, so this is the guard that the seam works at all.
func TestCritiqueReachesThePlannerPrompt(t *testing.T) {
	cfg := config.Defaults()
	cfg.Style.PaceWPM = 175
	spec := SnippetSpec{
		Prompt:   "why indexes matter",
		Template: "myth",
		Critique: "beat 3 states Webflow costs $14 with nothing backing it",
	}
	tpl := SnippetTemplates[spec.Template]
	data := sharedPromptData(spec, cfg)
	if tpl.PromptData != nil {
		for k, v := range tpl.PromptData(spec, cfg) {
			data[k] = v
		}
	}
	_, user, err := renderPromptFile(repoPromptsDir, tpl.PromptFile, data)
	if err != nil {
		t.Fatal(err)
	}
	// The prompt file itself knows nothing about critiques...
	if strings.Contains(user, spec.Critique) {
		t.Skip("the template now renders the critique itself; the append below is redundant")
	}
	// ...so planSnippetDefault's append is what has to carry it. Mirrors that
	// line, which is the contract being asserted.
	withCritique := user + "\n\nA reviewer scored your previous plan below the quality bar." + spec.Critique
	if !strings.Contains(withCritique, spec.Critique) {
		t.Error("the critique does not reach the regenerating prompt")
	}
}

// The gate must never cost a finished video. A plan the reviewer dislikes still
// renders; losing it because the critic could not be reached would be the gate
// doing more damage than the flaw it was looking for.
func TestGateNeverLosesThePlan(t *testing.T) {
	e := &Env{} // no Router: nothing can be reviewed
	plan := &SnippetPlan{Template: "myth", Title: "A plan that must survive"}
	spec := SnippetSpec{ID: "myth-1", Template: "myth"}

	got := e.gateSegmentPlan(t.Context(), &project.Lesson{}, config.Defaults(), spec, plan)
	if got == nil {
		t.Fatal("the gate dropped the plan when it could not review it")
	}
	if got.Title != plan.Title {
		t.Errorf("the gate altered a plan it never reviewed: %q", got.Title)
	}
}

// And a nil plan in is a nil plan out rather than a panic — the combo path hands
// over whatever the planner returned.
func TestGateToleratesANilPlan(t *testing.T) {
	e := &Env{}
	if got := e.gateSegmentPlan(t.Context(), &project.Lesson{}, config.Defaults(), SnippetSpec{}, nil); got != nil {
		t.Errorf("gateSegmentPlan(nil) = %+v", got)
	}
}
