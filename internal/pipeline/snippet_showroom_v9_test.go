package pipeline

import (
	"strings"
	"testing"
)

// The load-bearing rules of the four v9 showroom templates.
//
// One file rather than four, because what these tests are really covering is a
// pattern: each of the four rests on ONE rule that is unusual enough that a reader
// would assume it was a mistake, and each of those rules is the reason its template
// is not a restyling of something the catalog already had. Grouping them keeps that
// visible. The ordinary shape checks — open on the right beat, walk the list once,
// close on the closer — are the same code path in all four and are already covered
// by the cards and duel suites.

const v9Narration = "The arrangement holds still and the light moves across it, which is what lets a viewer keep their place in the frame."

// == opener: the big type needs LENGTH ==

func openerPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "opener",
		Title:    "Your first prompt in Claude Code",
		Opener: &OpenerSpec{
			Ground:  "Your first prompt in Claude Code",
			Kicker:  "Claude Code 101",
			Promise: "Write one prompt that gets a real change into your repo",
			Byline:  "Coursesmith",
		},
		Beats: []SnippetBeat{
			{ID: "the-title", Heading: "Where we are", Narration: v9Narration, Opener: &OpenerBeat{Show: "ground"}},
			{ID: "the-promise", Heading: "What you get", Narration: v9Narration, Opener: &OpenerBeat{Show: "promise"}},
			{ID: "the-mark", Heading: "Who this is", Narration: v9Narration, Opener: &OpenerBeat{Show: "mark"}},
		},
	}
	p.targetWords = 3 * 20
	return p
}

func TestOpenerPlanAccepted(t *testing.T) {
	if err := validateOpenerPlan(openerPlan()); err != nil {
		t.Fatalf("a well-formed title page was rejected: %v", err)
	}
}

// The rule that reads like a mistake: a MINIMUM length on a headline. The big type
// is the frame's texture, and texture needs area.
func TestOpenerRejectsAShortGround(t *testing.T) {
	p := openerPlan()
	p.Opener.Ground = "Claude Code"
	err := validateOpenerPlan(p)
	if err == nil {
		t.Fatal("a two-word ground was accepted, and at 250pt that is a logo rather than a title page")
	}
	if !strings.Contains(err.Error(), "the long way round") {
		t.Fatalf("the error does not tell the model what to do instead: %v", err)
	}
}

// The failure the template invites: the big words are right there, so restating them
// in the promise feels like reinforcement. It leaves nothing to read.
func TestOpenerRejectsAPromiseThatRepeatsTheGround(t *testing.T) {
	p := openerPlan()
	p.Opener.Promise = "your first prompt in claude code."
	if err := validateOpenerPlan(p); err == nil {
		t.Fatal("a promise that is the title again was accepted")
	}
}

// == changeplan: a summary must say what CHANGES ==

func changePlanPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "changeplan",
		Title:    "What the agent will touch",
		ChangePlan: &ChangePlanSpec{
			Closer: "Three files, one new dependency, nothing else moved",
			Files: []ChangeFile{
				{Path: "package.json", Summary: "add the sharp dependency", Verdict: "add", Edits: []string{"Add sharp to dependencies"}},
				{Path: "src/middleware/upload.ts", Summary: "switch to memory storage", Verdict: "edit", Edits: []string{"Replace diskStorage with memoryStorage"}},
				{Path: "src/index.ts", Summary: "already serves webp, nothing to do", Verdict: "unchanged"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "the-rail", Heading: "Three files", Narration: v9Narration, ChangePlan: &ChangePlanBeat{Show: "rail"}},
			{ID: "the-dep", Heading: "One dependency", Narration: v9Narration, ChangePlan: &ChangePlanBeat{Show: "file", At: 0}},
			{ID: "the-mw", Heading: "Memory not disk", Narration: v9Narration, ChangePlan: &ChangePlanBeat{Show: "file", At: 1}},
			{ID: "the-untouched", Heading: "Left alone", Narration: v9Narration, ChangePlan: &ChangePlanBeat{Show: "file", At: 2}},
			{ID: "the-total", Heading: "What it adds to", Narration: v9Narration, ChangePlan: &ChangePlanBeat{Show: "all"}},
		},
	}
	p.targetWords = 5 * 24
	return p
}

func TestChangePlanAccepted(t *testing.T) {
	if err := validateChangePlanPlan(changePlanPlan()); err != nil {
		t.Fatalf("a well-formed plan was rejected: %v", err)
	}
}

// An `unchanged` file needs no bullets, and that is the template's own argument:
// what a plan decided to leave alone is the most reassuring thing it contains.
func TestChangePlanAllowsAnUntouchedFileWithNoEdits(t *testing.T) {
	p := changePlanPlan()
	if p.ChangePlan.Files[2].Verdict != "unchanged" || len(p.ChangePlan.Files[2].Edits) != 0 {
		t.Fatal("fixture no longer covers the unchanged case")
	}
	if err := validateChangePlanPlan(p); err != nil {
		t.Fatalf("an unchanged file with no bullets was rejected: %v", err)
	}
}

// ...but a file the plan says it WILL change has to say how.
func TestChangePlanRejectsAChangedFileWithNoEdits(t *testing.T) {
	p := changePlanPlan()
	p.ChangePlan.Files[1].Edits = nil
	err := validateChangePlanPlan(p)
	if err == nil {
		t.Fatal("a file marked edit with nothing under it was accepted")
	}
	if !strings.Contains(err.Error(), "unchanged") {
		t.Fatalf("the error does not offer the honest alternative: %v", err)
	}
}

// The rule the header calls out: a row that names its own file twice.
func TestChangePlanRejectsASummaryThatIsJustTheFilename(t *testing.T) {
	p := changePlanPlan()
	p.ChangePlan.Files[0].Summary = "package"
	if err := validateChangePlanPlan(p); err == nil {
		t.Fatal("a summary that is the filename again was accepted")
	}
}

// == patch: a hunk is small, and the cap is on the CHANGE ==

func patchPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "patch",
		Title:    "Three lines made it async",
		Patch: &PatchSpec{
			Path: "src/routes/auth.ts",
			Lang: "ts",
			Hunks: []PatchHunk{
				{At: 58, Context: []string{"  try {"}, Before: []string{"    const p = req.file.path;"}, After: []string{"    const b = req.file.buffer;"}, Note: "The file never touches disk now"},
				{At: 69, Before: []string{"  } catch {"}, After: []string{"  } catch (err) {"}, Note: "Naming the error lets the 422 say something"},
			},
			Closer: "Two hunks, two lines added, two removed",
		},
		Beats: []SnippetBeat{
			{ID: "the-file", Heading: "The handler", Narration: v9Narration, Patch: &PatchBeat{Show: "file"}},
			{ID: "to-memory", Heading: "Off the disk", Narration: v9Narration, Patch: &PatchBeat{Show: "hunk", At: 0}},
			{ID: "the-catch", Heading: "Name the error", Narration: v9Narration, Patch: &PatchBeat{Show: "hunk", At: 1}},
			{ID: "the-total", Heading: "The whole change", Narration: v9Narration, Patch: &PatchBeat{Show: "tally"}},
		},
	}
	p.targetWords = 4 * 26
	return p
}

func TestPatchPlanAccepted(t *testing.T) {
	if err := validatePatchPlan(patchPlan()); err != nil {
		t.Fatalf("a well-formed patch was rejected: %v", err)
	}
}

// The premise, defended: one change readable at twice the usual size means the
// change has to be small, and the error has to say that rather than just refuse.
func TestPatchRejectsATooLargeHunk(t *testing.T) {
	p := patchPlan()
	p.Patch.Hunks[0].Before = []string{"a", "b", "c"}
	p.Patch.Hunks[0].After = []string{"d", "e", "f"}
	err := validatePatchPlan(p)
	if err == nil {
		t.Fatal("a six-line hunk was accepted")
	}
	if !strings.Contains(err.Error(), "code template") {
		t.Fatalf("the error does not point at the template that handles a rewrite: %v", err)
	}
}

// The note is why, and without it this is a slower way to read a patch file.
func TestPatchRequiresANote(t *testing.T) {
	p := patchPlan()
	p.Patch.Hunks[1].Note = ""
	if err := validatePatchPlan(p); err == nil {
		t.Fatal("a hunk with no note was accepted")
	}
}

// A long line is truncated rather than wrapped, because wrapping destroys the
// column alignment that makes a diff readable at all.
func TestPatchClampsLongLines(t *testing.T) {
	p := patchPlan()
	p.Patch.Hunks[0].After = []string{strings.Repeat("x", 200)}
	normalizePatchPlan(p)
	got := p.Patch.Hunks[0].After[0]
	if len([]rune(got)) != maxPatchLineChars {
		t.Errorf("line kept %d chars, want %d", len([]rune(got)), maxPatchLineChars)
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("a truncated line does not say so: %q", got)
	}
}

// The tally is arithmetic done once, in Go, so the number on the frame and the
// number the validator saw cannot disagree.
func TestPatchScenesCarryTheRunningTally(t *testing.T) {
	p := patchPlan()
	normalizePatchPlan(p)
	if err := validatePatchPlan(p); err != nil {
		t.Fatalf("fixture rejected: %v", err)
	}
	spans := make([]SectionSpan, len(p.Beats))
	ends := make([]int, len(p.Beats))
	for i, b := range p.Beats {
		spans[i] = SectionSpan{ID: b.ID, StartMs: i * 5000, EndMs: (i + 1) * 5000}
		ends[i] = (i + 1) * 5000
	}
	scenes, err := patchScenes(SnippetSceneInput{Plan: p, Spans: spans, BeatEndMs: ends, DurationMs: len(p.Beats) * 5000})
	if err != nil {
		t.Fatalf("patchScenes: %v", err)
	}
	steps := scenes[0].Props["steps"].([]map[string]any)
	for i, want := range []int{0, 1, 2, 2} {
		if got := steps[i]["added"]; got != want {
			t.Errorf("step %d added = %v, want %d", i, got, want)
		}
	}
}

// == approval: exactly one risk, and the pick may be it ==

func approvalPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "approval",
		Title:    "What auto-accept hands over",
		Approval: &ApprovalSpec{
			Tool:    "Claude Code",
			Context: "halfway through the avatar refactor",
			Ask:     "rm -rf uploads/avatars/",
			Pick:    0,
			Closer:  "Approve each edit until you trust it here",
			Answers: []ApprovalAnswer{
				{Label: "Yes", Consequence: "This one command runs; you are asked again next time"},
				{Label: "Yes, and stop asking", Consequence: "Every later command runs unasked this session", Risk: true},
				{Label: "No, and say why", Consequence: "It stops and re-plans with your reason"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "the-ask", Heading: "It wants to delete", Narration: v9Narration, Approval: &ApprovalBeat{Show: "ask"}},
			{ID: "once", Heading: "Yes, once", Narration: v9Narration, Approval: &ApprovalBeat{Show: "answer", At: 0}},
			{ID: "forever", Heading: "Yes, forever", Narration: v9Narration, Approval: &ApprovalBeat{Show: "answer", At: 1}},
			{ID: "no", Heading: "No, and why", Narration: v9Narration, Approval: &ApprovalBeat{Show: "answer", At: 2}},
			{ID: "the-call", Heading: "What to pick", Narration: v9Narration, Approval: &ApprovalBeat{Show: "pick"}},
		},
	}
	p.targetWords = 5 * 24
	return p
}

func TestApprovalPlanAccepted(t *testing.T) {
	if err := validateApprovalPlan(approvalPlan()); err != nil {
		t.Fatalf("a well-formed gate was rejected: %v", err)
	}
}

func TestApprovalRequiresExactlyOneRisk(t *testing.T) {
	none := approvalPlan()
	none.Approval.Answers[1].Risk = false
	err := validateApprovalPlan(none)
	if err == nil {
		t.Fatal("a gate with no risky answer was accepted — that is a menu")
	}
	if !strings.Contains(err.Error(), "menu") {
		t.Fatalf("the error does not say what the frame degrades into: %v", err)
	}

	two := approvalPlan()
	two.Approval.Answers[0].Risk = true
	if err := validateApprovalPlan(two); err == nil {
		t.Fatal("a gate with two risky answers was accepted — it has not helped anybody choose")
	}
}

// The combination that looks like a bug and is the honest answer: recommend the
// risky option. Deliberately allowed, so a clip can say "let it run, in a repo you
// can throw away".
func TestApprovalAllowsRecommendingTheRiskyAnswer(t *testing.T) {
	p := approvalPlan()
	p.Approval.Pick = 1
	p.Approval.Closer = "Turn it loose, but only in a repo you can throw away"
	if err := validateApprovalPlan(p); err != nil {
		t.Fatalf("recommending the risky answer was rejected, and that is the case the template is for: %v", err)
	}
}

// The consequence is the content of the whole frame.
func TestApprovalRequiresEveryConsequence(t *testing.T) {
	p := approvalPlan()
	p.Approval.Answers[2].Consequence = ""
	err := validateApprovalPlan(p)
	if err == nil {
		t.Fatal("an answer with no consequence was accepted")
	}
	if !strings.Contains(err.Error(), "one word apart") {
		t.Fatalf("the error does not explain why the consequence matters: %v", err)
	}
}

// A command is clipped by character, never by word: cutting "rm -rf build/" at a
// word boundary would silently change what it does.
func TestApprovalClampsTheAskByCharacter(t *testing.T) {
	p := approvalPlan()
	p.Approval.Ask = strings.Repeat("a", 200)
	normalizeApprovalPlan(p)
	if len([]rune(p.Approval.Ask)) != maxApprovalAskChars {
		t.Errorf("ask kept %d chars, want %d", len([]rune(p.Approval.Ask)), maxApprovalAskChars)
	}
}
