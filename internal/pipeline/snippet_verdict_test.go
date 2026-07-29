package pipeline

import (
	"strings"
	"testing"
)

const vdNarration = "Under about fifty gigabytes there is nothing a managed provider does that you cannot."

func verdictPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "verdict",
		Title:    "Should you actually self-host your database?",
		Verdict: &VerdictSpec{
			Subject: "Self-hosting Postgres",
			Call:    "Rent it until you have a platform team",
			Holds: []string{
				"Under about fifty gigabytes of data",
				"When downtime costs you real money",
				"If nobody runs backups weekly",
			},
			Breaks: []string{
				"When compliance forbids a managed provider",
				"Past a few terabytes, where the bill inverts",
			},
		},
		Beats: []SnippetBeat{
			{ID: "question", Heading: "The question", Narration: vdNarration, Verdict: &VerdictBeat{Show: "subject"}},
			{ID: "size", Heading: "How much data", Narration: vdNarration, Verdict: &VerdictBeat{Show: "holds", At: 0}},
			{ID: "downtime", Heading: "What downtime costs", Narration: vdNarration, Verdict: &VerdictBeat{Show: "holds", At: 1}},
			{ID: "backups", Heading: "Who runs backups", Narration: vdNarration, Verdict: &VerdictBeat{Show: "holds", At: 2}},
			{ID: "asterisk", Heading: "Where this is wrong", Narration: vdNarration, Verdict: &VerdictBeat{Show: "breaks", At: 0}},
			{ID: "scale", Heading: "The other exception", Narration: vdNarration, Verdict: &VerdictBeat{Show: "breaks", At: 1}},
			{ID: "call", Heading: "The call", Narration: vdNarration, Verdict: &VerdictBeat{Show: "call"}},
		},
	}
	p.targetWords = 7 * 40
	return p
}

func TestVerdictPlanAccepted(t *testing.T) {
	if err := validateVerdictPlan(verdictPlan()); err != nil {
		t.Fatalf("a well-formed verdict was rejected: %v", err)
	}
}

// The rule this template exists for: a recommendation with no asterisk is an
// advert, and the model will not write the awkward half unless required to.
func TestVerdictRequiresAnAsterisk(t *testing.T) {
	p := verdictPlan()
	p.Verdict.Breaks = nil
	err := validateVerdictPlan(p)
	if err == nil {
		t.Fatal("a ruling with no conditions against it was accepted")
	}
	if !strings.Contains(err.Error(), "advert") {
		t.Errorf("the error does not say why it matters: %v", err)
	}
}

// Written on screen but never spoken is the same as not said — which is exactly
// how a required asterisk gets quietly skipped.
func TestVerdictAsteriskMustBeNarrated(t *testing.T) {
	p := verdictPlan()
	// Keep the breaks on the card and keep the beat count intact — dropping
	// beats would trip the shared floor first and prove nothing about this
	// rule. Both asterisk beats are re-pointed at the subject instead, which is
	// exactly the shape a model produces when it writes the caveats down and
	// then talks about something else.
	p.Beats[4].Verdict = &VerdictBeat{Show: "subject"}
	p.Beats[5].Verdict = &VerdictBeat{Show: "subject"}
	err := validateVerdictPlan(p)
	if err == nil {
		t.Fatal("breaks written but never narrated were accepted")
	}
	if !strings.Contains(err.Error(), "nobody narrates") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerdictEndsOnTheCall(t *testing.T) {
	p := verdictPlan()
	p.Beats[6].Verdict = &VerdictBeat{Show: "holds", At: 0}
	err := validateVerdictPlan(p)
	if err == nil {
		t.Fatal("a clip that does not end on the call was accepted")
	}
	if !strings.Contains(err.Error(), "call") {
		t.Errorf("unexpected error: %v", err)
	}

	// And a call in the middle is rejected for the same reason from the other
	// direction.
	p = verdictPlan()
	p.Beats[1].Verdict = &VerdictBeat{Show: "call"}
	if err := validateVerdictPlan(p); err == nil {
		t.Fatal("a call with the clip carrying on afterwards was accepted")
	}
}

func TestVerdictNeedsACall(t *testing.T) {
	p := verdictPlan()
	p.Verdict.Call = ""
	if err := validateVerdictPlan(p); err == nil {
		t.Fatal("a verdict with no ruling was accepted")
	}
}

func TestVerdictRejectsRepeatedCondition(t *testing.T) {
	p := verdictPlan()
	p.Beats[2].Verdict.At = 0
	if err := validateVerdictPlan(p); err == nil {
		t.Fatal("the same condition walked twice was accepted")
	}
	// The two columns index independently: holds 0 and breaks 0 are different
	// lines and both may be walked.
	p = verdictPlan()
	p.Beats[1].Verdict = &VerdictBeat{Show: "holds", At: 0}
	p.Beats[4].Verdict = &VerdictBeat{Show: "breaks", At: 0}
	if err := validateVerdictPlan(p); err != nil {
		t.Errorf("holds 0 and breaks 0 collided: %v", err)
	}
}

func TestVerdictNormalizeRepairs(t *testing.T) {
	p := verdictPlan()
	p.Beats[3].Verdict.Show = "nonsense"
	p.Beats[4].Verdict.At = 99
	p.Verdict.Call = strings.Repeat("word ", 40)
	normalizeVerdictPlan(p)

	if p.Beats[3].Verdict.Show != "holds" {
		t.Errorf("an unknown show became %q, want holds", p.Beats[3].Verdict.Show)
	}
	if p.Beats[4].Verdict.At != len(p.Verdict.Breaks)-1 {
		t.Errorf("an out-of-range breaks beat points at %d", p.Beats[4].Verdict.At)
	}
	if n := len(strings.Fields(p.Verdict.Call)); n > maxVerdictCallWords {
		t.Errorf("the call is %d words after normalize, want <= %d — it has to fit the closing frame", n, maxVerdictCallWords)
	}
}

// An unlabelled beat is a blank field, not a claim: the shape says the clip
// opens on the subject and closes on the call.
func TestVerdictNormalizeInfersTheEnds(t *testing.T) {
	p := verdictPlan()
	p.Beats[0].Verdict.Show = ""
	p.Beats[6].Verdict.Show = ""
	normalizeVerdictPlan(p)
	if p.Beats[0].Verdict.Show != "subject" {
		t.Errorf("the first beat became %q, want subject", p.Beats[0].Verdict.Show)
	}
	if p.Beats[6].Verdict.Show != "call" {
		t.Errorf("the last beat became %q, want call", p.Beats[6].Verdict.Show)
	}
}

func TestVerdictScenesShape(t *testing.T) {
	p := verdictPlan()
	scenes, err := verdictScenes(sceneInput(t, p, 7000))
	if err != nil {
		t.Fatalf("verdictScenes: %v", err)
	}
	if len(scenes) != 1 || scenes[0].Type != SceneVerdict {
		t.Fatalf("want one verdict scene, got %d of %q", len(scenes), scenes[0].Type)
	}
	steps := scenes[0].Props["steps"].([]map[string]any)
	if len(steps) != len(p.Beats) {
		t.Fatalf("want one step per beat, got %d", len(steps))
	}
	if steps[len(steps)-1]["show"] != "call" {
		t.Error("the last step is not the call")
	}
	// The call is on the scene as its own prop: the closing frame is the whole
	// point of the template and the renderer must not have to dig it out of a
	// step.
	if scenes[0].Props["call"] != p.Verdict.Call {
		t.Errorf("call = %v, want %q", scenes[0].Props["call"], p.Verdict.Call)
	}
}
