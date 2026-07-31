package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/project"
)

// The device index is parsed rather than assumed, because it counts *all* video
// devices: on a machine with no webcam the screen is 0, and on one with a
// FaceTime camera it is 1. Hard-coding it records somebody's face instead of
// their screen.
func TestParseAVScreenIndex(t *testing.T) {
	withCamera := `[AVFoundation indev @ 0x1] AVFoundation video devices:
[AVFoundation indev @ 0x1] [0] FaceTime HD Camera
[AVFoundation indev @ 0x1] [1] Capture screen 0
[AVFoundation indev @ 0x1] AVFoundation audio devices:
[AVFoundation indev @ 0x1] [0] MacBook Pro Microphone`
	if idx, err := parseAVScreenIndex(withCamera); err != nil || idx != 1 {
		t.Errorf("idx = %d, err = %v; want 1", idx, err)
	}

	noCamera := `[AVFoundation indev @ 0x1] AVFoundation video devices:
[AVFoundation indev @ 0x1] [0] Capture screen 0`
	if idx, err := parseAVScreenIndex(noCamera); err != nil || idx != 0 {
		t.Errorf("idx = %d, err = %v; want 0", idx, err)
	}
}

// No screen device almost always means the permission was never granted, and
// saying so is the difference between a five-second fix and an afternoon.
func TestParseAVScreenIndexExplainsTheUsualCause(t *testing.T) {
	_, err := parseAVScreenIndex("[AVFoundation indev @ 0x1] [0] FaceTime HD Camera")
	if err == nil {
		t.Fatal("a listing with no screen device was accepted")
	}
	if !strings.Contains(err.Error(), "Screen Recording") {
		t.Errorf("error does not point at the permission: %v", err)
	}
}

// The crop is in pixels and the window is in points, so the measured Retina
// scale is what stands between a correct clip and one cropped to a quarter of
// the window with nothing on screen to say why.
func TestCropFilterScalesPointsToPixels(t *testing.T) {
	// Retina: a 1440x900 window at (40,60) doubles.
	if got, want := cropFilter(40, 60, 1440, 900, 2), "crop=2880:1800:80:120"; got != want {
		t.Errorf("retina crop = %q, want %q", got, want)
	}
	// External monitor: no scaling at all.
	if got, want := cropFilter(40, 60, 1440, 900, 1), "crop=1440:900:40:60"; got != want {
		t.Errorf("1x crop = %q, want %q", got, want)
	}
}

// h264 refuses odd dimensions, and a fractional scale is the way to get them.
func TestCropFilterKeepsDimensionsEven(t *testing.T) {
	got := cropFilter(0, 0, 801, 601, 1.5)
	for _, part := range strings.Split(strings.TrimPrefix(got, "crop="), ":") {
		n, err := strconv.Atoi(part)
		if err != nil {
			t.Fatalf("crop %q is not numeric: %v", got, err)
		}
		if n%2 != 0 {
			t.Errorf("crop %q has an odd dimension %d; h264 will refuse it", got, n)
		}
	}
}

const goodDesktopTake = `
app: cursor
window: {width: 1440, height: 900}
beats:
  - mark: project-open
    prompt: Open the habit-tracker folder
  - mark: agent-asked
    prompt: Ask the agent to add a weekly summary
`

func TestLoadDesktopTake(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "t.yaml")
	if err := os.WriteFile(p, []byte(goodDesktopTake), 0o644); err != nil {
		t.Fatal(err)
	}
	take, err := LoadDesktopTake(p)
	if err != nil {
		t.Fatal(err)
	}
	if take.App != "cursor" || len(take.Beats) != 2 {
		t.Fatalf("take = %+v", take)
	}
	w, h := take.size()
	if w != 1440 || h != 900 {
		t.Errorf("size = %dx%d", w, h)
	}
}

func TestDesktopTakeValidation(t *testing.T) {
	for _, tt := range []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{"accepts a real take", goodDesktopTake, ""},
		{"rejects an unknown app", "app: notepad\nbeats:\n  - mark: a\n    prompt: do\n", "not recordable"},
		{"rejects no beats", "app: cursor\n", "no beats"},
		{"rejects a beat with no mark", "app: cursor\nbeats:\n  - prompt: do\n", "no mark"},
		{"rejects a beat with no prompt", "app: cursor\nbeats:\n  - mark: a\n", "tells the operator nothing"},
		{"rejects a duplicate mark", "app: cursor\nbeats:\n  - mark: a\n    prompt: x\n  - mark: a\n    prompt: y\n", "used twice"},
		{"rejects an upper-case mark", "app: cursor\nbeats:\n  - mark: Open\n    prompt: x\n", "lowercase"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			p := filepath.Join(dir, "t.yaml")
			if err := os.WriteFile(p, []byte(tt.yaml), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := LoadDesktopTake(p)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %v, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

// A desktop capture must never appear in a batch run: it blocks on a keypress,
// and a build that stopped waiting for somebody is indistinguishable from a
// hang. The error has to name the command that does work.
func TestDesktopCaptureRefusesAHeadlessRun(t *testing.T) {
	course, lesson := testCourse(t)
	lesson = lessonWithCode(t, lesson, "[CAPTURE: tool=cursor, take=agent-edit; the agent editing files]\n")
	env, _ := runEnv(t, &fakeRouter{}) // no DesktopInput

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageCapture})
	if err == nil {
		t.Fatal("a desktop capture ran with nobody at the keyboard")
	}
	if !strings.Contains(err.Error(), "footage shoot") {
		t.Errorf("the error should name the command that works: %v", err)
	}
}

// The two attribute mistakes, named rather than ignored.
func TestDesktopCaptureMarkerRules(t *testing.T) {
	if _, err := extractCaptureMarkers("[CAPTURE: tool=cursor; no take]\n"); err == nil {
		t.Error("a desktop capture with no take was accepted")
	}
	_, err := extractCaptureMarkers("[CAPTURE: tool=cursor, take=x, fixture=y; no workdir]\n")
	if err == nil || !strings.Contains(err.Error(), "no working directory") {
		t.Errorf("a desktop capture with a fixture= was accepted or misreported: %v", err)
	}
	specs, err := extractCaptureMarkers("[CAPTURE: tool=cursor, take=agent-edit; the agent editing]\n")
	if err != nil {
		t.Fatal(err)
	}
	if specs[0].Kind != CaptureKindDesktop || specs[0].Take != "agent-edit" {
		t.Errorf("spec = %+v", specs[0])
	}
}

// Keys must be unique across all three registries, since `tool=` looks in all
// of them and a collision makes the answer depend on lookup order.
func TestCaptureKeysAreUniqueAcrossRegistries(t *testing.T) {
	seen := map[string]string{}
	for k := range captureTools {
		seen[k] = "tool"
	}
	for k := range captureSites {
		if where, clash := seen[k]; clash {
			t.Errorf("%q is both a %s and a site", k, where)
		}
		seen[k] = "site"
	}
	for k := range captureApps {
		if where, clash := seen[k]; clash {
			t.Errorf("%q is both a %s and a desktop app", k, where)
		}
	}
}
