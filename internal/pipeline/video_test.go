package pipeline

import (
	"context"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

func TestHumanizeSlug(t *testing.T) {
	tests := map[string]string{
		"your-first-line-of-python": "Your First Line Of Python",
		"intro":                     "Intro",
		"a--b":                      "A  B",
	}
	for in, want := range tests {
		if got := humanizeSlug(in); got != want {
			t.Errorf("humanizeSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderTitleCard(t *testing.T) {
	fnt, err := findFont()
	if err != nil {
		t.Skip("no system font available")
	}
	out := filepath.Join(t.TempDir(), "card.png")
	spec := slideSpec{
		Heading:  "Your First Line Of Python",
		Subtitle: "What is Python?",
		Colors:   config.Colors{Primary: "#306998", Accent: "#ffd43b", Background: "#ffffff"},
	}
	if err := renderTitleCard(fnt, spec, out); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("card is not a valid PNG: %v", err)
	}
	if img.Bounds().Dx() != 1920 || img.Bounds().Dy() != 1080 {
		t.Errorf("card size = %v, want 1920x1080", img.Bounds())
	}
	// The heading must actually have been drawn: some pixel near the center
	// should differ from the white background.
	painted := false
	for x := 0; x < 1920 && !painted; x += 4 {
		r, g, b, _ := img.At(x, 500).RGBA()
		if r != 0xffff || g != 0xffff || b != 0xffff {
			painted = true
		}
	}
	if !painted {
		t.Error("title card appears blank")
	}
}

func TestParseHexColor(t *testing.T) {
	fallback := color.RGBA{1, 2, 3, 255}
	tests := []struct {
		in   string
		want color.RGBA
	}{
		{"#306998", color.RGBA{0x30, 0x69, 0x98, 255}},
		{"#fff", color.RGBA{255, 255, 255, 255}},
		{"", fallback},
		{"blue", fallback},
		{"#12345", fallback},
	}
	for _, tt := range tests {
		if got := parseHexColor(tt.in, fallback); got != tt.want {
			t.Errorf("parseHexColor(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestMuxAndBurnArgs(t *testing.T) {
	mux := muxArgs("slides.mp4", "voice.wav", "caps.vtt", "final.mp4")
	joined := strings.Join(mux, " ")
	for _, want := range []string{"-c:v libx264", "-crf 20", "-c:s mov_text", "-c:a aac", "-shortest", "final.mp4"} {
		if !strings.Contains(joined, want) {
			t.Errorf("muxArgs missing %q: %s", want, joined)
		}
	}

	burn := burnArgs("slides.mp4", "voice.wav", "caps.vtt", "burned.mp4")
	joinedBurn := strings.Join(burn, " ")
	if !strings.Contains(joinedBurn, "subtitles=caps.vtt") {
		t.Errorf("burnArgs missing subtitles filter: %s", joinedBurn)
	}
	if slices.Contains(burn, "-c:s") {
		t.Error("burned variant must not also carry a subtitle stream")
	}
}

func TestVideoStageRequiresUpstream(t *testing.T) {
	requireFFmpeg(t)
	course, lesson := testCourse(t)
	env, _ := runEnv(t, &fakeRouter{})

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageRender})
	if err == nil || !strings.Contains(err.Error(), "audio stage must run first") {
		t.Errorf("error = %v, want audio-first error", err)
	}

	seedVoiceover(t, lesson)
	err = env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageRender})
	if err == nil || !strings.Contains(err.Error(), "captions stage must run first") {
		t.Errorf("error = %v, want captions-first error", err)
	}
}

// TestVideoStageSlidesMode runs real ffmpeg end-to-end in slides mode.
func TestVideoStageSlidesMode(t *testing.T) {
	requireFFmpeg(t)
	if _, err := findFont(); err != nil {
		t.Skip("no system font available for drawtext slides")
	}
	course, lesson := testCourse(t)
	seedScript(t, lesson)

	// Per-section audio drives slide timing; voiceover + captions feed the mux.
	audioDir := filepath.Join(lesson.GeneratedDir(), AudioDirName)
	if err := os.MkdirAll(audioDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"first-idea", "second-idea"} {
		if err := os.WriteFile(filepath.Join(audioDir, id+".wav"), makeWAV(0.4), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(lesson.GeneratedDir(), VoiceoverFileName), makeWAV(0.8), 0o644); err != nil {
		t.Fatal(err)
	}
	vtt := "WEBVTT\n\n1\n00:00:00.000 --> 00:00:00.800\nHello world.\n\n"
	if err := os.WriteFile(filepath.Join(lesson.GeneratedDir(), CaptionsFileName), []byte(vtt), 0o644); err != nil {
		t.Fatal(err)
	}

	env, out := runEnv(t, &fakeRouter{})
	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageRender}); err != nil {
		t.Fatal(err)
	}

	final := filepath.Join(lesson.GeneratedDir(), FinalVideoName)
	info, err := os.Stat(final)
	if err != nil {
		t.Fatalf("final.mp4 not written: %v", err)
	}
	if info.Size() == 0 {
		t.Error("final.mp4 is empty")
	}
	if !strings.Contains(out.String(), "slides mode") {
		t.Errorf("output missing slides-mode notice:\n%s", out.String())
	}
	// The burned variant depends on libass; it either exists or was warned about.
	if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), BurnedVideoName)); err != nil {
		if !strings.Contains(out.String(), "burned-subtitle variant failed") {
			t.Errorf("no burned video and no warning:\n%s", out.String())
		}
	}
}
