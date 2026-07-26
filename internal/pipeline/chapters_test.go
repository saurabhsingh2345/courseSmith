package pipeline

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/project"
)

func TestYoutubeTimestamp(t *testing.T) {
	tests := []struct {
		ms   int
		want string
	}{
		{0, "0:00"},
		{62_000, "1:02"},
		{600_000, "10:00"},
		{3_726_000, "1:02:06"},
		{-5, "0:00"},
	}
	for _, tt := range tests {
		if got := youtubeTimestamp(tt.ms); got != tt.want {
			t.Errorf("youtubeTimestamp(%d) = %q, want %q", tt.ms, got, tt.want)
		}
	}
}

func TestChaptersStage(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson) // sections first-idea / second-idea
	seedAlignment(t, lesson, Alignment{
		Source: AlignSourceWhisperX,
		Words: append(
			wordSeq(500, "Python", "reads", "code", "line", "by", "line."),
			wordSeq(9000, "Now", "let", "us", "try", "it", "live.")...,
		),
		Sections: []SectionSpan{
			{ID: "first-idea", StartMs: 500, EndMs: 1950, WordStart: 0, WordEnd: 6},
			{ID: "second-idea", StartMs: 9000, EndMs: 10450, WordStart: 6, WordEnd: 12},
		},
	})

	env, out := runEnv(t, &fakeRouter{})
	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageChapters}); err != nil {
		t.Fatal(err)
	}

	var chapters []Chapter
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), ChaptersJSONFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &chapters); err != nil {
		t.Fatal(err)
	}
	if len(chapters) != 2 {
		t.Fatalf("chapters = %+v", chapters)
	}
	// Titles come from the outline's ## headings, matched by slug.
	if chapters[0].Title != "First idea" || chapters[1].Title != "Second idea" {
		t.Errorf("titles = %q, %q", chapters[0].Title, chapters[1].Title)
	}
	// First chapter starts at 0 (YouTube requirement); second at its span.
	if chapters[0].StartMs != 0 || chapters[0].EndMs != 9000 {
		t.Errorf("chapter 0 = %+v", chapters[0])
	}
	if chapters[1].StartMs != 9000 || chapters[1].EndMs != 10450 {
		t.Errorf("chapter 1 = %+v", chapters[1])
	}

	txt, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), ChaptersTxtFileName))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(txt); got != "0:00 First idea\n0:09 Second idea\n" {
		t.Errorf("chapters.txt = %q", got)
	}

	md, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), TranscriptFileName))
	if err != nil {
		t.Fatal(err)
	}
	transcript := string(md)
	for _, want := range []string{
		"# Test Lesson",
		"## First idea [0:00]",
		"## Second idea [0:09]",
		"Python reads code line by line.",
		"Now let us try it live.",
	} {
		if !strings.Contains(transcript, want) {
			t.Errorf("transcript.md missing %q:\n%s", want, transcript)
		}
	}
	if !strings.Contains(out.String(), "2 chapter(s)") {
		t.Errorf("output missing chapters notice:\n%s", out.String())
	}
}
