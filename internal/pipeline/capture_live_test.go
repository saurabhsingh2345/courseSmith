package pipeline

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// The mark model in footage.go is arithmetic about a program we do not control.
// Every unit test around it asserts the arithmetic against itself, which cannot
// notice VHS changing what a directive costs, or refusing a tape shape the
// engine emits.
//
// This test records for real and checks the model against the clip that comes
// out. It is what would have caught the header bug found on the first real run:
// VHS's parser splits an unquoted `Output` argument on its path separators and
// reports the tail as an invalid command, so every tool capture — which writes
// to an absolute path back in the lesson dir — died at validate, while every
// fake-runner test passed.
//
// Skipped unless vhs is installed. It costs about twenty seconds.
func requireVHS(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("vhs"); err != nil {
		t.Skip("vhs is not installed — brew install vhs")
	}
}

func TestTapeHeaderIsAcceptedByRealVHS(t *testing.T) {
	requireVHS(t)
	dir := t.TempDir()

	// The absolute-path form, which is what a tool capture uses.
	out := filepath.Join(dir, "sub dir", "capture-1.mp4")
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		t.Fatal(err)
	}
	tape := filepath.Join(dir, "capture-1.tape")
	body := "Type \"true\"\nEnter\n# MARK ran\nSleep 1s\n"
	if err := os.WriteFile(tape, []byte(tapeHeaderTo(out, SceneTheme{})+body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := (HostTapeRunner{}).Validate(context.Background(), dir, "capture-1.tape"); err != nil {
		t.Fatalf("real vhs rejected the engine's own header: %v", err)
	}
}

// The timing model, measured against a real recording.
//

// Tolerance is 400ms rather than something tighter because VHS trims a little
// off the tail of every clip (a bare `Sleep 5s` records as ~4.76s), and that
// trim is an encoder artifact rather than something the model should chase with
// a magic constant. What the assertion is really defending is the *shape* of
// the model — that a character costs the typing speed, that a keypress costs
// the same, and that a Sleep costs its face value.
func TestMarkModelMatchesARealRecording(t *testing.T) {
	// Behind a flag, unlike the other live tests here, because this one
	// measures WALL-CLOCK behaviour of an external tool and the ordinary suite
	// is a hostile place to do that: run alongside the rest of the package it
	// has come back 17% and 30% short, and run alone it passes every time.
	//
	// The note used to say "re-run on a quiet machine before believing it",
	// which is true and not good enough — a test that fails half the time
	// teaches people to skip past red, and then the one real failure goes past
	// too. Deliberate is better than noisy:
	//
	//	COURSESMITH_LIVE=1 go test ./internal/pipeline -run MarkModel
	if os.Getenv("COURSESMITH_LIVE") != "1" {
		t.Skip("set COURSESMITH_LIVE=1 to measure VHS timing (needs a quiet machine)")
	}
	requireVHS(t)
	requireFFmpeg(t)
	dir := t.TempDir()
	out := filepath.Join(dir, "clip.mp4")

	// 10 chars + Enter = 11 keystrokes at 80ms = 880ms, then 2s, then the mark,
	// then 1s. Nothing here waits on anything real, so the model should land.
	body := "Type \"echo-nothing\"\nEnter\nSleep 2s\n# MARK settled\nSleep 1s\n"
	tape := filepath.Join(dir, "clip.tape")
	if err := os.WriteFile(tape, []byte(tapeHeaderTo(out, SceneTheme{})+body), 0o644); err != nil {
		t.Fatal(err)
	}
	runner := HostTapeRunner{}
	if err := runner.Validate(context.Background(), dir, "clip.tape"); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := runner.RenderTape(ctx, dir, "clip.tape"); err != nil {
		t.Fatal(err)
	}

	real, err := mediaDurationMs(out)
	if err != nil {
		t.Fatal(err)
	}
	f := buildFootage("clip", CaptureKindTool, "git", "", body, real, time.Now())

	// "echo-nothing" is 12 chars; + Enter = 13 * 80ms = 1040ms; + 2s = 3040ms.
	const wantMark = 3040
	const wantTotal = 4040
	if diff := abs(f.TapeTimeMs - wantTotal); diff != 0 {
		t.Fatalf("the model itself moved: tape time %dms, want %dms", f.TapeTimeMs, wantTotal)
	}
	if diff := abs(real - wantTotal); diff > 400 {
		t.Errorf("real clip is %dms, model says %dms — the cost of a directive has changed under us", real, wantTotal)
	}
	if len(f.Marks) != 1 || f.Marks[0].Name != "settled" {
		t.Fatalf("marks = %+v", f.Marks)
	}
	if f.Marks[0].AtMs != wantMark {
		t.Errorf("mark at %dms, want %dms", f.Marks[0].AtMs, wantMark)
	}
	if f.Marks[0].Approximate {
		t.Error("a tape with no Wait must produce an exact mark")
	}
	if f.Waits != 0 {
		t.Errorf("waits = %d", f.Waits)
	}
}
