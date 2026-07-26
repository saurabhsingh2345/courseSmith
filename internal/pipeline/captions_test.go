package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

type fakeTranscriber struct {
	gotPath  string
	gotModel string
	result   *llm.Transcription
	err      error
}

func (f *fakeTranscriber) Transcribe(_ context.Context, audioPath, model string) (*llm.Transcription, error) {
	f.gotPath = audioPath
	f.gotModel = model
	return f.result, f.err
}

func seedVoiceover(t *testing.T, l *project.Lesson) {
	t.Helper()
	if err := os.MkdirAll(l.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(l.GeneratedDir(), VoiceoverFileName), makeWAV(1), 0o644); err != nil {
		t.Fatal(err)
	}
}

func seedAlignment(t *testing.T, l *project.Lesson, a Alignment) {
	t.Helper()
	if err := os.MkdirAll(l.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), AlignmentFileName), a); err != nil {
		t.Fatal(err)
	}
}

func TestCaptionsStageFromAlignment(t *testing.T) {
	course, lesson := testCourse(t)
	seedAlignment(t, lesson, Alignment{
		Source: AlignSourceWhisperX,
		Words: []AlignedWord{
			{Word: "Hello", StartMs: 0, EndMs: 300},
			{Word: "world.", StartMs: 350, EndMs: 700},
			{Word: "This", StartMs: 900, EndMs: 1100},
			{Word: "is", StartMs: 1150, EndMs: 1250},
			{Word: "Python.", StartMs: 1300, EndMs: 1750},
		},
	})
	env, _ := runEnv(t, &fakeRouter{})

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageCaptions}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), CaptionsFileName))
	if err != nil {
		t.Fatal(err)
	}
	want := "WEBVTT\n\n" +
		"1\n00:00:00.000 --> 00:00:00.700\nHello world.\n\n" +
		"2\n00:00:00.900 --> 00:00:01.750\nThis is Python.\n\n"
	if string(data) != want {
		t.Errorf("captions.vtt =\n%q\nwant\n%q", data, want)
	}
}

func TestCaptionsStageRequiresAlignment(t *testing.T) {
	course, lesson := testCourse(t)
	env, _ := runEnv(t, &fakeRouter{})

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageCaptions})
	if err == nil || !strings.Contains(err.Error(), "align stage must run first") {
		t.Errorf("error = %v, want align-first error", err)
	}
}

func TestVTTFromWords(t *testing.T) {
	t.Run("sentence end breaks the cue", func(t *testing.T) {
		vtt, n := vttFromWords(wordSeq(0, "One.", "Two", "three"))
		if n != 2 {
			t.Fatalf("cues = %d, want 2:\n%s", n, vtt)
		}
		if !strings.Contains(vtt, "One.\n") || !strings.Contains(vtt, "Two three\n") {
			t.Errorf("vtt:\n%s", vtt)
		}
	})

	t.Run("long silence breaks the cue", func(t *testing.T) {
		words := []AlignedWord{
			{Word: "before", StartMs: 0, EndMs: 300},
			{Word: "after", StartMs: 1500, EndMs: 1800}, // 1200ms gap > 800ms
		}
		_, n := vttFromWords(words)
		if n != 2 {
			t.Errorf("cues = %d, want 2", n)
		}
	})

	t.Run("word cap breaks the cue", func(t *testing.T) {
		words := wordSeq(0, "a", "b", "c", "d", "e", "f", "g", "h", "i", "j")
		_, n := vttFromWords(words)
		if n != 2 {
			t.Errorf("cues = %d, want 2 (8-word cap)", n)
		}
	})

	t.Run("char cap breaks the cue", func(t *testing.T) {
		words := wordSeq(0, "supercalifragilistic", "expialidocious", "antidisestablishment")
		_, n := vttFromWords(words)
		if n < 2 {
			t.Errorf("cues = %d, want >= 2 (42-char cap)", n)
		}
	})

	t.Run("cue timing spans its words", func(t *testing.T) {
		vtt, _ := vttFromWords(wordSeq(1000, "hi", "there"))
		if !strings.Contains(vtt, "00:00:01.000 --> 00:00:01.450") {
			t.Errorf("vtt:\n%s", vtt)
		}
	})
}

func TestVTTTimestamp(t *testing.T) {
	tests := []struct {
		in   float64
		want string
	}{
		{0, "00:00:00.000"},
		{1.5, "00:00:01.500"},
		{59.9994, "00:00:59.999"},
		{61.25, "00:01:01.250"},
		{3661.007, "01:01:01.007"},
		{-2, "00:00:00.000"},
	}
	for _, tt := range tests {
		if got := vttTimestamp(tt.in); got != tt.want {
			t.Errorf("vttTimestamp(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
