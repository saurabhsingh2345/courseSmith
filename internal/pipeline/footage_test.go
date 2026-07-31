package pipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/enfec/coursesmith/internal/project"
)

// A tape with no Wait is fully modelled: every mark should land exactly where
// the arithmetic says, and none should be flagged.
func TestMarksAreExactWithoutAWait(t *testing.T) {
	body := `Type "ls"
Enter
# MARK listed
Sleep 2s
Type "git status"
Enter
# MARK status
Sleep 1s
`
	// "ls" = 2 chars, Enter = 1 key → 3 * 80ms = 240ms
	// + Sleep 2s → 2240ms
	// + "git status" = 10 chars + Enter = 11 * 80ms = 880ms → 3120ms
	scan := scanTape(body, tapeTypingSpeedMs)
	if scan.waits != 0 {
		t.Fatalf("waits = %d", scan.waits)
	}
	if scan.totalMs != 3120+1000 {
		t.Errorf("totalMs = %d, want %d", scan.totalMs, 4120)
	}

	marks := resolveMarks(scan, scan.totalMs)
	if len(marks) != 2 {
		t.Fatalf("marks = %+v", marks)
	}
	if marks[0].Name != "listed" || marks[0].AtMs != 240 || marks[0].Approximate {
		t.Errorf("mark 0 = %+v, want listed @240ms exact", marks[0])
	}
	if marks[1].Name != "status" || marks[1].AtMs != 3120 || marks[1].Approximate {
		t.Errorf("mark 1 = %+v, want status @3120ms exact", marks[1])
	}
}

// The case the whole design is for: one Wait of unknown length. The clip runs
// longer than its tape time, and the entire discrepancy belongs to that Wait —
// so marks before it do not move, marks after it move by exactly the measured
// amount, and nothing is approximate.
func TestOneWaitIsMeasuredAndNothingIsApproximate(t *testing.T) {
	body := `Type "claude go"
Enter
# MARK sent
Wait
# MARK done
Sleep 1s
`
	scan := scanTape(body, tapeTypingSpeedMs)
	if scan.waits != 1 {
		t.Fatalf("waits = %d", scan.waits)
	}
	// "claude go" = 9 chars + Enter = 10 * 80 = 800ms, then Sleep 1s.
	if scan.totalMs != 1800 {
		t.Fatalf("totalMs = %d, want 1800", scan.totalMs)
	}

	// The agent really took 30 extra seconds.
	marks := resolveMarks(scan, 31800)
	if marks[0].AtMs != 800 || marks[0].Approximate {
		t.Errorf("mark before the wait moved or was flagged: %+v", marks[0])
	}
	if marks[1].AtMs != 30800 {
		t.Errorf("mark after the wait = %dms, want 30800 (800 + 30000 measured drift)", marks[1].AtMs)
	}
	if marks[1].Approximate {
		t.Error("a single wait is measurable, so the mark after it must not be approximate")
	}
}

// Two Waits and the drift cannot be attributed to either, so marks after the
// first must say so rather than be silently placed. A mark that is quietly
// eight seconds out is worse than no mark, because nothing downstream can tell.
func TestTwoWaitsFlagEverythingAfterTheFirst(t *testing.T) {
	body := `Type "a"
Enter
# MARK one
Wait
# MARK two
Wait
# MARK three
`
	scan := scanTape(body, tapeTypingSpeedMs)
	if scan.waits != 2 {
		t.Fatalf("waits = %d", scan.waits)
	}
	marks := resolveMarks(scan, 60000)
	if marks[0].Approximate {
		t.Error("the mark before any wait is still exact")
	}
	if !marks[1].Approximate || !marks[2].Approximate {
		t.Errorf("marks after the first of two waits must be approximate: %+v", marks)
	}
}

// The case a real run turned up: a model told to wait for a slow command also
// puts `Wait` after the fast ones, and `Wait` after `git init` returns
// instantly. Three waits that between them cost nothing must not cost the clip
// its marks — the total drift bounds the sum of every wait's overrun, so a
// small total is a proof that each one was small.
func TestCheapWaitsDoNotFlagAnything(t *testing.T) {
	body := `Type "git init"
Enter
Wait
# MARK repo
Sleep 1s
Type "git status"
Enter
Wait
# MARK status
Sleep 1s
Wait
# MARK done
Sleep 3s
`
	scan := scanTape(body, tapeTypingSpeedMs)
	if scan.waits != 3 {
		t.Fatalf("waits = %d", scan.waits)
	}
	// The clip came out 200ms shorter than its tape time, which is VHS
	// trimming the tail — nothing waited on anything.
	marks := resolveMarks(scan, scan.totalMs-200)
	for _, m := range marks {
		if m.Approximate {
			t.Errorf("mark %q flagged despite the clip running to its tape time: %+v", m.Name, marks)
		}
	}
	if marks[0].AtMs != 720 { // "git init" = 8 chars + Enter = 9 * 80ms
		t.Errorf("mark 0 at %dms, want 720", marks[0].AtMs)
	}
}

// The same three waits, but this time one of them really blocked. Now the
// drift cannot be attributed and the marks after the first wait must say so.
func TestExpensiveWaitsStillFlag(t *testing.T) {
	body := "Type \"a\"\nEnter\n# MARK typed\nWait\n# MARK one\nWait\n# MARK two\nWait\n# MARK three\n"
	scan := scanTape(body, tapeTypingSpeedMs)
	marks := resolveMarks(scan, scan.totalMs+45000)
	if marks[0].Name != "typed" || marks[0].Approximate {
		t.Errorf("the mark before any wait is still exact: %+v", marks[0])
	}
	for _, m := range marks[1:] {
		if !m.Approximate {
			t.Errorf("mark %q after the first of three expensive waits must be approximate: %+v", m.Name, marks)
		}
	}
}

// A clip that fits its slot is played exactly as it always was. Every python
// demo in the existing course is this case, and it must not change.
func TestPacingLeavesAFittingClipAlone(t *testing.T) {
	segs := PlanTerminalPacing(8000, 20000, []FootageMark{{Name: "a", AtMs: 4000}})
	if len(segs) != 1 || segs[0].Rate != 1 || segs[0].FromMs != 0 || segs[0].ToMs != 8000 {
		t.Errorf("segments = %+v, want one real-time segment", segs)
	}
}

// The real case that exposed this: lesson five's capture was 53.3s of recording
// in a 20.8s slot, so the video cut away ten seconds in and the viewer never
// saw the agent's output — the entire point of the shot.
func TestPacingFitsTheRealAgentCapture(t *testing.T) {
	const clip, slot = 53280, 20792
	marks := []FootageMark{
		{Name: "project-open", AtMs: 2240},
		{Name: "feature-added", AtMs: 47760},
		{Name: "files-changed", AtMs: 50000},
		{Name: "readme-content", AtMs: 53120},
	}
	segs := PlanTerminalPacing(clip, slot, marks)

	// The whole recording is covered, start to finish, in order — the point is
	// that nothing is lost, so a plan that drops the tail is the bug itself.
	if segs[0].FromMs != 0 {
		t.Errorf("plan starts at %dms, not the beginning", segs[0].FromMs)
	}
	if last := segs[len(segs)-1].ToMs; last != clip {
		t.Errorf("plan ends at %dms, want the whole %dms clip", last, clip)
	}
	for i := 1; i < len(segs); i++ {
		if segs[i].FromMs != segs[i-1].ToMs {
			t.Errorf("gap or overlap between segment %d and %d: %+v", i-1, i, segs)
		}
	}

	if played := PlayedMs(segs); played > slot+100 {
		t.Errorf("plan plays for %dms in a %dms slot", played, slot)
	}

	// The long think is what gets compressed; the moments around it stay at
	// real time, which is the whole reason marks exist.
	var think, opening *PacingSegment
	for i := range segs {
		if segs[i].FromMs == 2240 {
			think = &segs[i]
		}
		if segs[i].FromMs == 0 {
			opening = &segs[i]
		}
	}
	if think == nil || opening == nil {
		t.Fatalf("segments = %+v", segs)
	}
	if think.Rate <= 1.5 {
		t.Errorf("the 45s think was not compressed: rate %.2f", think.Rate)
	}
	if opening.Rate != 1 {
		t.Errorf("the opening typing was sped up to %.2f; it should stay real time", opening.Rate)
	}
}

// No marks means nothing says which part is dead air. Too fast everywhere is
// still better than losing the ending, so it degrades rather than truncating.
func TestPacingWithoutMarksIsUniform(t *testing.T) {
	segs := PlanTerminalPacing(40000, 10000, nil)
	if len(segs) != 1 {
		t.Fatalf("segments = %+v", segs)
	}
	if segs[0].Rate < 3.9 || segs[0].Rate > 4.1 {
		t.Errorf("rate = %.2f, want ~4 (40s into 10s)", segs[0].Rate)
	}
}

// Nothing may be sped up past the point where it reads as a glitch, even when
// that means the plan cannot fit. Overrunning a slot slightly is recoverable;
// an unwatchable blur is not.
func TestPacingNeverExceedsTheRateCeiling(t *testing.T) {
	segs := PlanTerminalPacing(600000, 1000, []FootageMark{{Name: "a", AtMs: 300000}})
	for _, s := range segs {
		if s.Rate > maxPacingRate+0.001 {
			t.Errorf("segment %+v exceeds the ceiling of %.0f", s, maxPacingRate)
		}
	}
}

// writeFootage puts a sidecar where applyTerminalPacing will look for it.
func writeFootage(t *testing.T, l *project.Lesson, f Footage) {
	t.Helper()
	dir := filepath.Join(l.GeneratedDir(), DemosDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(dir, f.ID+FootageFileSuffix), f); err != nil {
		t.Fatal(err)
	}
}

func terminalGraph(clipMs, slotMs int) *SceneGraph {
	return &SceneGraph{Scenes: []Scene{{
		Type: SceneTerminal, StartMs: 0, EndMs: slotMs,
		Props: map[string]any{"src": "demos/capture-1.mp4", "durationMs": clipMs, "clipId": "capture-1"},
	}}}
}

func TestApplyTerminalPacingUsesExactMarks(t *testing.T) {
	_, lesson := testCourse(t)
	writeFootage(t, lesson, Footage{
		ID: "capture-1", Kind: CaptureKindTool, DurationMs: 50000,
		Marks: []FootageMark{{Name: "sent", AtMs: 2000}, {Name: "done", AtMs: 45000}},
	})
	g := terminalGraph(50000, 20000)
	applyTerminalPacing(g, lesson)

	segs, ok := g.Scenes[0].Props["segments"].([]PacingSegment)
	if !ok {
		t.Fatalf("no segments planned: %+v", g.Scenes[0].Props)
	}
	if len(segs) != 3 {
		t.Fatalf("segments = %+v, want one per mark boundary", segs)
	}
	if segs[0].Rate != 1 {
		t.Errorf("the opening should stay real time, got %.2f", segs[0].Rate)
	}
	if segs[1].Rate <= 1 {
		t.Errorf("the long gap should be compressed, got %.2f", segs[1].Rate)
	}
	// clipId is plumbing between the two passes and must not survive into the
	// scene graph — it is not something the renderer has any use for.
	if _, leaked := g.Scenes[0].Props["clipId"]; leaked {
		t.Error("clipId leaked into the scene graph")
	}
}

// An approximate mark is not a cut point. A clip whose timing could not be
// attributed gets a uniform speed-up rather than a confident cut in the wrong
// place — this is the one consumer that makes the flag worth carrying.
func TestApplyTerminalPacingIgnoresApproximateMarks(t *testing.T) {
	_, lesson := testCourse(t)
	writeFootage(t, lesson, Footage{
		ID: "capture-1", Kind: CaptureKindTool, DurationMs: 50000,
		Marks: []FootageMark{
			{Name: "sent", AtMs: 2000},
			{Name: "done", AtMs: 45000, Approximate: true},
		},
	})
	g := terminalGraph(50000, 20000)
	applyTerminalPacing(g, lesson)

	segs := g.Scenes[0].Props["segments"].([]PacingSegment)
	if len(segs) != 1 {
		t.Fatalf("segments = %+v, want a single uniform stretch when marks cannot be trusted", segs)
	}
}

// A clip that fits its slot gets no plan at all, so every existing python demo
// renders through exactly the path it always did.
func TestApplyTerminalPacingSkipsAFittingClip(t *testing.T) {
	_, lesson := testCourse(t)
	g := terminalGraph(8000, 20000)
	applyTerminalPacing(g, lesson)
	if _, planned := g.Scenes[0].Props["segments"]; planned {
		t.Error("a clip that fits was given a pacing plan")
	}
	if _, leaked := g.Scenes[0].Props["clipId"]; leaked {
		t.Error("clipId leaked into the scene graph")
	}
}

// Compressing a recording and saying nothing about it is a quiet claim that
// the tool was faster than it was. In a course whose moat is "the tool really
// did that", the length of the recording cannot be the one claim that drifts.
func TestCaptureCreditStatesBothDurationsWhenCondensed(t *testing.T) {
	f := Footage{Kind: CaptureKindTool, Tool: "claude", ToolVersion: "2.1.220"}
	c := captureCreditFor(f, "Claude Code", 53280, 20792)
	if c.Tool != "Claude Code" || c.Version != "2.1.220" {
		t.Errorf("credit = %+v", c)
	}
	// Tools print their name beside their version, so the raw banner would
	// render as "Claude Code 2.1.220 (Claude Code)". The banner stays intact in
	// footage.json — that file is evidence — and only the display is shortened.
	for banner, want := range map[string]string{
		"2.1.220 (Claude Code)":              "2.1.220",
		"git version 2.50.1 (Apple Git-155)": "2.50.1",
		"Vercel CLI 39.1.1":                  "39.1.1",
		"unversioned":                        "unversioned",
	} {
		if got := shortVersion(banner); got != want {
			t.Errorf("shortVersion(%q) = %q, want %q", banner, got, want)
		}
	}
	if c.RealMs != 53280 || c.ShownMs != 20792 {
		t.Errorf("credit must carry both durations: %+v", c)
	}
}

// A clip that plays at real time makes no claim about its own speed. "12s real,
// shown in 12s" is noise wearing the costume of rigour.
func TestCaptureCreditClaimsNoCompressionWhenThereIsNone(t *testing.T) {
	f := Footage{Kind: CaptureKindTool, Tool: "git", ToolVersion: "2.50.1"}
	if c := captureCreditFor(f, "git", 12000, 20000); c.ShownMs != 0 {
		t.Errorf("a clip shorter than its slot claimed compression: %+v", c)
	}
	// Sub-second differences are rounding, not a fast-forward.
	if c := captureCreditFor(f, "git", 12400, 12000); c.ShownMs != 0 {
		t.Errorf("a 400ms difference was reported as compression: %+v", c)
	}
}

// The credit is assembled from measured values only, and a python demo — our
// own code, not somebody else's product — makes no claim at all.
func TestCaptureCreditIsAttachedOnlyToRealProducts(t *testing.T) {
	_, lesson := testCourse(t)
	writeFootage(t, lesson, Footage{
		ID: "demo-1", Kind: CaptureKindPython, DurationMs: 40000,
	})
	writeFootage(t, lesson, Footage{
		ID: "capture-1", Kind: CaptureKindTool, Tool: "claude",
		ToolVersion: "2.1.220", DurationMs: 40000,
	})
	g := &SceneGraph{Scenes: []Scene{
		{Type: SceneTerminal, StartMs: 0, EndMs: 10000,
			Props: map[string]any{"durationMs": 40000, "provClipId": "demo-1"}},
		{Type: SceneTerminal, StartMs: 10000, EndMs: 20000,
			Props: map[string]any{"durationMs": 40000, "provClipId": "capture-1"}},
	}}
	applyCaptureProvenance(g, lesson)

	if _, has := g.Scenes[0].Props["provenance"]; has {
		t.Error("a python demo claimed provenance it has no business claiming")
	}
	c, ok := g.Scenes[1].Props["provenance"].(CaptureCredit)
	if !ok {
		t.Fatalf("no credit on the tool capture: %+v", g.Scenes[1].Props)
	}
	if c.Tool != "Claude Code" || c.Version != "2.1.220" || c.ShownMs != 10000 {
		t.Errorf("credit = %+v", c)
	}
	// provClipId is plumbing between passes; it must not reach the renderer.
	for i := range g.Scenes {
		if _, leaked := g.Scenes[i].Props["provClipId"]; leaked {
			t.Errorf("scene %d leaked provClipId", i)
		}
	}
}

// What the stage split buys: the script stage is told what the recordings
// actually caught, so narration can lead into them and state how long the tool
// took. Before the split it was writing about footage that did not exist yet.
func TestCaptureBriefingDescribesWhatWasRecorded(t *testing.T) {
	_, lesson := testCourse(t)
	writeFootage(t, lesson, Footage{
		ID: "capture-1", Kind: CaptureKindTool, Tool: "claude",
		ToolVersion: "2.1.220 (Claude Code)", DurationMs: 53280,
		Marks: []FootageMark{{Name: "project-open"}, {Name: "feature-added"}},
	})
	lines := CaptureBriefing(lesson)
	if len(lines) != 1 {
		t.Fatalf("briefing = %+v", lines)
	}
	for _, want := range []string{"Claude Code", "2.1.220", "53 seconds", "project open", "feature added"} {
		if !strings.Contains(lines[0], want) {
			t.Errorf("briefing %q does not mention %q", lines[0], want)
		}
	}
	// The banner is shortened for reading, not repeated whole.
	if strings.Contains(lines[0], "(Claude Code)") {
		t.Errorf("briefing repeats the version banner: %q", lines[0])
	}
}

// A python demo is our own code running and the outline already describes it,
// so it contributes nothing — the briefing is about somebody else's product.
func TestCaptureBriefingIgnoresPythonDemos(t *testing.T) {
	_, lesson := testCourse(t)
	writeFootage(t, lesson, Footage{ID: "demo-1", Kind: CaptureKindPython, DurationMs: 9000})
	if lines := CaptureBriefing(lesson); len(lines) != 0 {
		t.Errorf("briefing = %+v, want nothing for a python demo", lines)
	}
}

// Before anything has been recorded there is nothing to say, and the prompt
// treats the briefing as optional — a first run must not claim footage exists.
func TestCaptureBriefingIsEmptyBeforeAnythingIsRecorded(t *testing.T) {
	_, lesson := testCourse(t)
	if lines := CaptureBriefing(lesson); len(lines) != 0 {
		t.Errorf("briefing = %+v, want nothing", lines)
	}
}

func TestFootageExactReportsPerClip(t *testing.T) {
	f := Footage{Marks: []FootageMark{{Name: "a"}, {Name: "b", Approximate: true}}}
	if f.Exact() {
		t.Error("a clip with one approximate mark is not exact")
	}
	if _, ok := f.Mark("b"); !ok {
		t.Error("Mark did not find b")
	}
	if _, ok := f.Mark("nope"); ok {
		t.Error("Mark invented a mark")
	}
}

// A mark can never sit past the end of the clip it is in: if the model of the
// tape overshot the real recording, the mark is clamped and flagged rather than
// handed downstream as a cut point beyond the last frame.
func TestMarksAreClampedToTheRealDuration(t *testing.T) {
	body := "Type \"aaaa\"\nEnter\nWait\n# MARK late\n"
	scan := scanTape(body, tapeTypingSpeedMs)
	marks := resolveMarks(scan, 100)
	if marks[0].AtMs != 100 || !marks[0].Approximate {
		t.Errorf("mark = %+v, want clamped to 100ms and flagged", marks[0])
	}
}

func TestTypedTextLenHandlesQuoting(t *testing.T) {
	for _, tt := range []struct {
		arg  string
		want int
	}{
		{`"hello"`, 5},
		{`'hello'`, 5},
		{"`hello`", 5},
		{`hello`, 5},
		{`"it's"`, 4},
		{`""`, 0},
	} {
		if got := typedTextLen(tt.arg); got != tt.want {
			t.Errorf("typedTextLen(%s) = %d, want %d", tt.arg, got, tt.want)
		}
	}
}

func TestScanTapeHonoursSpeedOverridesAndRepeats(t *testing.T) {
	scan := scanTape("Type@10ms \"abc\"\nEnter 4\n# MARK m\n", tapeTypingSpeedMs)
	// 3 chars at 10ms, then 4 Enters at the default 80ms.
	want := 30 + 320
	if scan.totalMs != want {
		t.Errorf("totalMs = %d, want %d", scan.totalMs, want)
	}
	if scan.marks[0].AtMs != want {
		t.Errorf("mark at %d, want %d", scan.marks[0].AtMs, want)
	}
}

// An unfamiliar directive must not take the marks down with it. VHS already
// validated the tape; a timing model that refuses to run because it met a new
// keyword loses every mark in the clip to protect against a small drift.
func TestScanTapeSurvivesUnknownDirectives(t *testing.T) {
	scan := scanTape("Hide\nType \"a\"\nShow\n# MARK m\n", tapeTypingSpeedMs)
	if len(scan.marks) != 1 {
		t.Fatalf("marks = %+v", scan.marks)
	}
	if scan.marks[0].AtMs != 80 {
		t.Errorf("mark at %d, want 80 (only the Type counted)", scan.marks[0].AtMs)
	}
}

func TestBuildFootageRecordsWhatWasObserved(t *testing.T) {
	now := time.Date(2026, 7, 31, 12, 0, 0, 0, time.UTC)
	f := buildFootage("capture-1", CaptureKindTool, "claude", "1.2.3",
		"Type \"claude go\"\nEnter\n# MARK sent\n", 5000, now)
	if f.ID != "capture-1" || f.Kind != CaptureKindTool || f.Tool != "claude" {
		t.Errorf("footage = %+v", f)
	}
	if f.ToolVersion != "1.2.3" {
		t.Errorf("toolVersion = %q", f.ToolVersion)
	}
	if f.CapturedAt != "2026-07-31T12:00:00Z" {
		t.Errorf("capturedAt = %q", f.CapturedAt)
	}
	if f.DurationMs != 5000 || f.TapeTimeMs != 800 {
		t.Errorf("durations = %d / %d", f.DurationMs, f.TapeTimeMs)
	}
	if len(f.Marks) != 1 || f.Marks[0].Name != "sent" {
		t.Errorf("marks = %+v", f.Marks)
	}
}
