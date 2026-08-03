package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

func specFrom(t *testing.T, yaml string) (*NoCodeSpec, error) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, NoCodeFileName), []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	return LoadNoCodeSpec(dir)
}

const goodNoCode = `
title: One sentence, one app
brief: what these tools actually do, shown rather than described
segments:
  - template: footage
    prompt: a habit tracker built from one sentence
    evidence:
      kind: capture
      capture:
        tool: claude
        of: ask the agent to add a weekly summary
  - template: costing
    prompt: what running it actually costs
    evidence:
      kind: fact
      facts:
        - "Vercel's Hobby plan is free for personal projects"
        - "Supabase's free tier allows 500MB of database storage"
`

func TestLoadNoCodeSpec(t *testing.T) {
	spec, err := specFrom(t, goodNoCode)
	if err != nil {
		t.Fatal(err)
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("a good piece did not validate: %v", err)
	}
	if len(spec.Segments) != 2 {
		t.Fatalf("segments = %+v", spec.Segments)
	}
	// Ids are generated once from position, so an edit stays addressed to the
	// same segment when a neighbour moves.
	if spec.Segments[0].ID != "footage-1" || spec.Segments[1].ID != "costing-2" {
		t.Errorf("ids = %q, %q", spec.Segments[0].ID, spec.Segments[1].ID)
	}
	// A capture's id is its segment's id: one segment, one recording. Deriving
	// it removes the whole class of mistake where a segment references a
	// capture that was renamed or never declared.
	if got := spec.CaptureIDs(); len(got) != 1 || got[0] != "footage-1" {
		t.Errorf("captures = %v", got)
	}
	// And the capture stage gets a real request out of it, not a reference.
	reqs := spec.CaptureSpecs()
	if len(reqs) != 1 || reqs[0].Tool != "claude" || reqs[0].Kind != CaptureKindTool {
		t.Fatalf("capture specs = %+v", reqs)
	}
	if reqs[0].ID != "footage-1" || reqs[0].Description == "" {
		t.Errorf("capture spec = %+v", reqs[0])
	}
}

// The rule that earns the surface. A segment standing on nothing is refused
// before a single planning call is spent — a reel discovers that after the
// caster has gone; here it is a parse error.
func TestNoCodeRefusesASegmentWithNoEvidence(t *testing.T) {
	spec, err := specFrom(t, `
segments:
  - template: verdict
    prompt: whether this is worth doing
`)
	if err != nil {
		t.Fatal(err)
	}
	err = spec.Validate()
	if err == nil {
		t.Fatal("a segment standing on nothing was accepted")
	}
	for _, want := range []string{"names no evidence", "capture", "fact"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
}

func TestNoCodeEvidenceMustBeComplete(t *testing.T) {
	for _, tt := range []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			"a capture must be named",
			"segments:\n  - template: footage\n    prompt: x\n    evidence: {kind: capture}\n",
			"names no tool",
		},
		{
			"facts must be listed",
			"segments:\n  - template: verdict\n    prompt: x\n    evidence: {kind: fact}\n",
			"lists none",
		},
		{
			"an unknown kind is refused",
			"segments:\n  - template: verdict\n    prompt: x\n    evidence: {kind: vibes}\n",
			"not one of",
		},
		{
			"a prompt is still required",
			"segments:\n  - template: verdict\n    prompt: \"\"\n    evidence: {kind: capture, capture: {tool: claude, of: x}}\n",
			"no prompt",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := specFrom(t, tt.yaml)
			if err != nil {
				t.Fatal(err)
			}
			err = spec.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

// A face is the fastest way to fill a frame with no evidence behind it, and
// this surface exists to refuse that. The error has to say why, or it reads as
// an arbitrary restriction.
func TestNoCodeRefusesTheDrawnFigureTemplates(t *testing.T) {
	for _, name := range []string{"cast", "story", "illustration"} {
		spec, err := specFrom(t,
			"segments:\n  - template: "+name+"\n    prompt: x\n    evidence: {kind: capture, capture: {tool: claude, of: x}}\n")
		if err != nil {
			t.Fatal(err)
		}
		err = spec.Validate()
		if err == nil {
			t.Errorf("%s was accepted in a no-code piece", name)
			continue
		}
		if !strings.Contains(err.Error(), "drawn") && !strings.Contains(err.Error(), "evidence") {
			t.Errorf("%s was refused without saying why: %v", name, err)
		}
	}
	// And they really are excluded from the offered catalog, not just rejected
	// after somebody picks one.
	names := NoCodeTemplateNames()
	for _, banned := range []string{"cast", "story", "illustration"} {
		if slices.Contains(names, banned) {
			t.Errorf("%q is offered by NoCodeTemplateNames", banned)
		}
	}
	// The rest of the catalog is still there — this is a subset, not a new one.
	for _, want := range []string{"footage", "costing", "verdict", "mockup", "promptloop"} {
		if !slices.Contains(names, want) {
			t.Errorf("%q is missing from the no-code catalog", want)
		}
	}
}

// Captures must exist before anything decides what to say about them. That
// ordering is the surface's argument made mechanical.
func TestNoCodeRecordsBeforeItWrites(t *testing.T) {
	capture := slices.Index(NoCodeStageOrder, project.StageCapture)
	plan := slices.Index(NoCodeStageOrder, project.StagePlan)
	substance := slices.Index(NoCodeStageOrder, project.StageSubstance)
	if capture != 0 {
		t.Errorf("capture is stage %d, want first: %v", capture, NoCodeStageOrder)
	}
	if capture > substance || substance > plan {
		t.Errorf("order must be capture → substance → plan, got %v", NoCodeStageOrder)
	}
}

// Skipping is the commonest edit after a first watch, and a skipped segment
// must not be validated or cut — but its prompt stays on disk so it can come
// back.
func TestNoCodeSkipDropsASegmentFromTheCut(t *testing.T) {
	spec, err := specFrom(t, `
segments:
  - template: footage
    prompt: keep me
    evidence: {kind: capture, capture: {tool: claude, of: x}}
  - template: cast
    prompt: this one is skipped and would otherwise be refused
    skip: true
`)
	if err != nil {
		t.Fatal(err)
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("a skipped segment was validated: %v", err)
	}
	if len(spec.Live()) != 1 {
		t.Errorf("live = %+v", spec.Live())
	}
	if len(spec.Segments) != 2 {
		t.Errorf("the skipped segment was deleted rather than dropped: %+v", spec.Segments)
	}
}

func TestNoCodeIsRecognisedByItsSpecFile(t *testing.T) {
	dir := t.TempDir()
	l := &project.Lesson{Dir: dir}
	if IsNoCode(l) {
		t.Error("an empty directory was taken for a no-code piece")
	}
	if err := os.WriteFile(filepath.Join(dir, NoCodeFileName), []byte("segments: []\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !IsNoCode(l) {
		t.Error("a directory with nocode.yaml was not recognised")
	}
}

// An empty cut is a piece with nothing in it, and failing here is much cheaper
// than failing at render.
func TestNoCodeRefusesAnEmptyCut(t *testing.T) {
	spec, err := specFrom(t, "segments: []\n")
	if err != nil {
		t.Fatal(err)
	}
	if err := spec.Validate(); err == nil {
		t.Fatal("a piece with no segments was accepted")
	}
}

// The `footage` template cannot be filled from a prompt: its frame IS a clip.
// Picking it in the gallery with nothing recorded used to plan happily and
// render an empty browser window reading "capture unavailable" — a silent
// broken clip, which is worse than a build that stopped.
func TestFootageTemplateRefusesWithNoRecording(t *testing.T) {
	tpl, ok := SnippetTemplates["footage"]
	if !ok || tpl.Plan == nil {
		t.Fatal("the footage template has no plan guard")
	}
	_, err := tpl.Plan(context.Background(), &Env{}, SnippetSpec{Prompt: "show me lovable"}, config.Config{})
	if err == nil {
		t.Fatal("planned a footage clip with no recording attached")
	}
	// The message has to say what to do instead, or it reads as the template
	// being broken rather than misapplied.
	for _, want := range []string{"[CAPTURE]", "mockup"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not point anywhere useful (%q missing): %v", want, err)
		}
	}
}

// A web or desktop capture cannot be driven from a description — nobody can
// invent selectors for somebody else's page, or the beats for a native app —
// so it must name a take. Saying that here is much cheaper than discovering it
// when the browser opens.
func TestNoCodeWebCaptureNeedsATake(t *testing.T) {
	spec, err := specFrom(t,
		"segments:\n  - template: footage\n    prompt: x\n    evidence: {kind: capture, capture: {tool: lovable}}\n")
	if err != nil {
		t.Fatal(err)
	}
	err = spec.Validate()
	if err == nil || !strings.Contains(err.Error(), "take") {
		t.Errorf("error = %v, want it to demand a take", err)
	}
	// With one, it is fine.
	spec, err = specFrom(t,
		"segments:\n  - template: footage\n    prompt: x\n    evidence: {kind: capture, capture: {tool: lovable, take: first-build}}\n")
	if err != nil {
		t.Fatal(err)
	}
	if err := spec.Validate(); err != nil {
		t.Fatalf("a web capture naming a take was refused: %v", err)
	}
	if reqs := spec.CaptureSpecs(); len(reqs) != 1 || reqs[0].Kind != CaptureKindWeb || reqs[0].Take != "first-build" {
		t.Errorf("capture specs = %+v", reqs)
	}
}

// An unrecordable tool is refused with the whole vocabulary quoted back, so the
// fix is picking from a list rather than guessing.
func TestNoCodeRefusesAnUnrecordableTool(t *testing.T) {
	spec, err := specFrom(t,
		"segments:\n  - template: footage\n    prompt: x\n    evidence: {kind: capture, capture: {tool: photoshop, of: y}}\n")
	if err != nil {
		t.Fatal(err)
	}
	err = spec.Validate()
	if err == nil {
		t.Fatal("an unrecordable tool was accepted")
	}
	for _, want := range []string{"claude", "lovable", "cursor"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not offer %q: %v", want, err)
		}
	}
}

// The gate the whole surface rests on, written from the failure that produced
// it: a real run recorded a Claude Code session and narrated four beats about
// "the Vercel Agent", because a different segment's evidence mentioned Vercel
// and the fact sheet carried it across. Every other check passed.
func TestFootageMustNarrateTheRecordedTool(t *testing.T) {
	beats := func(lines ...string) []SnippetBeat {
		out := make([]SnippetBeat, 0, len(lines))
		for _, l := range lines {
			out = append(out, SnippetBeat{Narration: l})
		}
		return out
	}

	// The real failure, exactly as it came out: four beats about a tool that is
	// not in the clip, and the recorded one never named. The first rule catches
	// it, which is the cheaper and clearer of the two.
	p := &SnippetPlan{
		Footage:         &FootagePlan{Clip: "footage-1"},
		FootageToolName: "Claude Code",
		Beats: beats(
			"Introducing the Vercel Agent and what it does.",
			"The Vercel Agent analyses the project.",
			"Then the Vercel Agent writes the patch.",
		),
	}
	err := validateFootagePlan(p)
	if err == nil {
		t.Fatal("a plan narrating the wrong tool was accepted")
	}
	if !strings.Contains(err.Error(), "Claude Code") || !strings.Contains(err.Error(), "never names it") {
		t.Errorf("error = %v", err)
	}

	// The subtler shape: the recorded tool IS named, but the piece is about
	// something else. A mention is not the same as being the subject.
	p.Beats = beats(
		"Claude Code is one option here.",
		"The Vercel Agent analyses the project.",
		"Then the Vercel Agent writes the patch.",
		"The Vercel Agent pushes the change.",
	)
	err = validateFootagePlan(p)
	if err == nil {
		t.Fatal("a plan dominated by another tool was accepted")
	}
	if !strings.Contains(err.Error(), "Claude Code") || !strings.Contains(err.Error(), "Vercel") {
		t.Errorf("the error must name both tools: %v", err)
	}

	// Naming a neighbour in passing is fine — being about it is not.
	p.Beats = beats(
		"Claude Code reads the whole project before it writes anything.",
		"Unlike Cursor, it works from the terminal.",
		"Claude Code then shows you the files it changed.",
	)
	if err := validateFootagePlan(p); err != nil {
		t.Errorf("a passing mention of another tool was refused: %v", err)
	}
}

// Validation runs inside the correction loop, so anything a validator checks
// against has to be on the plan *there*. Attaching it after the planner returns
// is attaching it after every judgement has been made — which is how the mark
// check silently passed on an empty set.
func TestFootagePreValidateRunsBeforeJudgement(t *testing.T) {
	tpl := SnippetTemplates["footage"]
	if tpl.PreValidate == nil {
		t.Fatal("the footage template has no PreValidate hook")
	}
	p := &SnippetPlan{}
	tpl.PreValidate(SnippetSpec{
		FootageMarks: []string{"sent", "done"},
		FootageTool:  "Claude Code",
		FootageMs:    53280,
	}, p)
	if len(p.FootageKnownMarks) != 2 || p.FootageToolName != "Claude Code" || p.FootageMs != 53280 {
		t.Errorf("plan = %+v", p)
	}
}

// A footage segment occupies its own slot on the timeline, not all of it.
//
// footageScenes used to return {StartMs: 0, EndMs: in.DurationMs}. For a
// standalone snippet those are the same thing. On a multi-segment piece
// DurationMs is the whole finished video, so the recording claimed the entire
// timeline and every later segment drew itself *inside the terminal window*,
// over the top of the still-playing clip. Nothing errored: the plan was valid,
// the render reported done, and the video was unwatchable. That is the failure
// mode this asserts against, because no other check in the pipeline can see it.
func TestFootageSceneSpansOnlyItsOwnBeats(t *testing.T) {
	tpl, ok := SnippetTemplates["footage"]
	if !ok || tpl.Scenes == nil {
		t.Fatal("the footage template lays out no scenes")
	}
	// One segment's three beats, sitting in the middle of a 226s piece.
	in := SnippetSceneInput{
		Plan: &SnippetPlan{
			Footage:           &FootagePlan{Clip: "footage-1"},
			FootageSrc:        "demos/footage-1.mp4",
			FootageMs:         161_000,
			FootageIsTerminal: true,
			Beats: []SnippetBeat{{ID: "b1"}, {ID: "b2"}, {ID: "b3"}},
		},
		Spans: []SectionSpan{
			{StartMs: 4_000}, {StartMs: 22_000}, {StartMs: 40_000},
		},
		BeatEndMs: []int{22_000, 40_000, 61_300},
		// The whole piece, which is exactly the number that must NOT be used.
		DurationMs: 226_000,
	}
	in.Spec = SnippetSpec{ID: "footage-1", Template: "footage"}
	scenes, err := tpl.Scenes(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(scenes) != 1 {
		t.Fatalf("scenes = %d, want 1", len(scenes))
	}
	s := scenes[0]
	// The id both the pacing planner and the provenance chip look the sidecar
	// up by. Taken from the segment, never from the plan: one model wrote
	// "Claude Code recording" into footage.clip, both lookups missed, and the
	// clip played uncompressed with no credit and no payoff on screen.
	for _, key := range []string{"clipId", "provClipId"} {
		if got, _ := s.Props[key].(string); got != "footage-1" {
			t.Errorf("props[%q] = %q, want the segment id %q", key, got, "footage-1")
		}
	}
	if s.StartMs != 4_000 {
		t.Errorf("StartMs = %d, want 4000 (its first beat)", s.StartMs)
	}
	if s.EndMs != 61_300 {
		t.Errorf("EndMs = %d, want 61300 (its last beat)", s.EndMs)
	}
	if s.EndMs >= in.DurationMs {
		t.Errorf("the scene spans %d–%dms of a %dms piece — a footage segment that "+
			"reaches the end of the timeline draws underneath every segment after it",
			s.StartMs, s.EndMs, in.DurationMs)
	}
}
