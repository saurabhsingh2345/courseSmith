package pipeline

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestKeystrokeTrackIsAReadableWav(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "keys.wav")

	times := KeystrokeTimesMs(KeystrokeSchedule("print('hi')\nprint('there')", 3000), 500)
	n, err := WriteKeystrokeTrack(path, times, map[int]bool{11: true}, 4000)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(times) {
		t.Errorf("wrote %d clicks for %d keystrokes", n, len(times))
	}

	// Read it back with the pipeline's own WAV reader rather than trusting the
	// writer's arithmetic. The two were written independently, so agreeing is
	// evidence; a header this file alone believes in is not.
	got, err := wavDuration(path)
	if err != nil {
		t.Fatalf("the generated track is not readable as a WAV: %v", err)
	}
	want := 4 * time.Second
	if diff := got - want; diff > 200*time.Millisecond || diff < -200*time.Millisecond {
		t.Errorf("track runs %v, want about %v", got, want)
	}
}

func TestKeystrokeTrackStaysUnderTheVoice(t *testing.T) {
	// The number that decides whether this feature is pleasant or unbearable.
	// A click loud enough to sit *on* a −16 LUFS voice makes the clip
	// unwatchable, and it is the kind of thing nobody notices in review
	// because reviews are read, not listened to.
	dir := t.TempDir()
	path := filepath.Join(dir, "keys.wav")

	// Worst case for level: every keystroke at once, so their decays pile up.
	times := make([]int, 60)
	for i := range times {
		times[i] = 100 + i // 1ms apart — far faster than anyone types
	}
	if _, err := WriteKeystrokeTrack(path, times, nil, 1000); err != nil {
		t.Fatal(err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	peak := 0.0
	for i := 44; i+1 < len(raw); i += 2 {
		s := float64(int16(uint16(raw[i])|uint16(raw[i+1])<<8)) / 32767
		peak = math.Max(peak, math.Abs(s))
	}
	if peak > 0.5 {
		t.Errorf("peak amplitude %.3f — the typing track is competing with the narration", peak)
	}
	if peak == 0 {
		t.Error("the track is silent")
	}
}

func TestKeystrokeTrackIsDeterministic(t *testing.T) {
	dir := t.TempDir()
	times := KeystrokeTimesMs(KeystrokeSchedule("x = 1\ny = 2", 2000), 0)

	first := filepath.Join(dir, "a.wav")
	second := filepath.Join(dir, "b.wav")
	if _, err := WriteKeystrokeTrack(first, times, map[int]bool{5: true}, 3000); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteKeystrokeTrack(second, times, map[int]bool{5: true}, 3000); err != nil {
		t.Fatal(err)
	}
	a, _ := os.ReadFile(first)
	b, _ := os.ReadFile(second)
	if len(a) != len(b) {
		t.Fatalf("two runs produced %d and %d bytes", len(a), len(b))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("two runs differ at byte %d — the track is not reproducible", i)
		}
	}
}

func TestNoKeystrokesNoTrack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.wav")
	n, err := WriteKeystrokeTrack(path, nil, nil, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("wrote %d clicks for an empty schedule", n)
	}
	// A template that types nothing must not leave an empty file behind for the
	// renderer to load and the render to stage.
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("an empty schedule still created a file")
	}
}

func TestCollectKeystrokesFindsSchedulesAndNewlines(t *testing.T) {
	code := "a\nb"
	graph := &SceneGraph{
		Scenes: []Scene{
			{Type: SceneTitle, Props: map[string]any{"heading": "no typing here"}},
			{
				Type: SceneWalkthrough,
				Props: map[string]any{
					"keystrokes": []int{100, 200, 300},
					"steps":      []map[string]any{{"code": code}},
				},
			},
		},
	}
	times, newlines := collectKeystrokes(graph)
	if len(times) != 3 {
		t.Fatalf("collected %d keystrokes, want 3", len(times))
	}
	if !newlines[1] {
		t.Error("index 1 is the newline in \"a\\nb\" and was not marked")
	}
	if newlines[0] || newlines[2] {
		t.Error("an ordinary character was marked as a newline")
	}
}

func TestIntSlicePropSurvivesJSONShapes(t *testing.T) {
	// A scene graph that has been through JSON, or through a video-plan patch,
	// delivers []any of float64 rather than []int. Dropping it there would be a
	// silent track.
	if got := intSliceProp([]any{float64(10), float64(20)}); len(got) != 2 || got[1] != 20 {
		t.Errorf("float64 slice not coerced: %v", got)
	}
	if got := intSliceProp([]int{1, 2}); len(got) != 2 {
		t.Errorf("int slice not passed through: %v", got)
	}
	if got := intSliceProp("nonsense"); got != nil {
		t.Errorf("a non-slice should yield nil, got %v", got)
	}
}
