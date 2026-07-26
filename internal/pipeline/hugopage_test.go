package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/enfec/coursesmith/internal/project"
)

// seedAllArtifacts writes every artifact the hugo stage embeds.
func seedAllArtifacts(t *testing.T, l *project.Lesson) {
	t.Helper()
	seedScript(t, l) // script.json with a diagram cue on memory-model
	files := map[string][]byte{
		QuizFileName:     []byte(quizBody()),
		CaptionsFileName: []byte("WEBVTT\n\n1\n00:00:00.000 --> 00:00:01.000\nHi.\n\n"),
		FinalVideoName:   []byte("fake mp4 bytes"),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(l.GeneratedDir(), name), content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	diagramsDir := filepath.Join(l.GeneratedDir(), DiagramsDirName)
	if err := os.MkdirAll(diagramsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(diagramsDir, "memory-model.svg"), []byte(svgBody("Memory")), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestHugoStageEmitsPageBundle(t *testing.T) {
	course, lesson := testCourse(t)
	seedAllArtifacts(t, lesson)

	env, _ := runEnv(t, &fakeRouter{})
	env.SiteDir = filepath.Join(t.TempDir(), "site")

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageHugo}); err != nil {
		t.Fatal(err)
	}

	bundle := filepath.Join(env.SiteDir, "content", "courses", "test-course", "01-test")
	indexMD, err := os.ReadFile(filepath.Join(bundle, "index.md"))
	if err != nil {
		t.Fatalf("index.md not written: %v", err)
	}
	page := string(indexMD)
	for _, want := range []string{
		`title: "Test Lesson"`,
		"weight: 1",
		`{{< lesson-video src="final.mp4" captions="captions.vtt" >}}`,
		"## First Idea",
		"Python reads code line by line.",
		`{{< diagram src="diagrams/memory-model.svg"`,
		"## Check your understanding",
		`{{< quiz src="quiz.json" >}}`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("index.md missing %q:\n%s", want, page)
		}
	}

	for _, asset := range []string{
		FinalVideoName, CaptionsFileName, QuizFileName,
		filepath.Join(DiagramsDirName, "memory-model.svg"),
	} {
		if _, err := os.Stat(filepath.Join(bundle, asset)); err != nil {
			t.Errorf("asset not copied into bundle: %v", err)
		}
	}

	courseIndex, err := os.ReadFile(filepath.Join(env.SiteDir, "content", "courses", "test-course", "_index.md"))
	if err != nil {
		t.Fatalf("course _index.md not written: %v", err)
	}
	if !strings.Contains(string(courseIndex), `title: "Test Course"`) {
		t.Errorf("course index = %q", courseIndex)
	}
}

func TestHugoStageRequiresArtifacts(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson) // script only — quiz/captions/video missing
	env, _ := runEnv(t, &fakeRouter{})
	env.SiteDir = filepath.Join(t.TempDir(), "site")

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageHugo})
	if err == nil || !strings.Contains(err.Error(), "stage must run first") {
		t.Errorf("error = %v, want missing-artifact error", err)
	}
}

func TestLessonWeight(t *testing.T) {
	tests := map[string]int{
		"01-what-is-python": 1,
		"10-final-project":  10,
		"7-shortcut":        7,
		"intro":             999,
		"00-zero":           999,
	}
	for id, want := range tests {
		if got := lessonWeight(id); got != want {
			t.Errorf("lessonWeight(%q) = %d, want %d", id, got, want)
		}
	}
}

func TestShortCaption(t *testing.T) {
	long := strings.Repeat("word ", 40)
	tests := []struct{ in, want string }{
		{"Python as a translator: person on the left", "Python as a translator"},
		{"Simple boxes.  With trailing detail.", "Simple boxes"},
		{"no punctuation here", "no punctuation here"},
	}
	for _, tt := range tests {
		if got := shortCaption(tt.in); got != tt.want {
			t.Errorf("shortCaption(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
	if got := shortCaption(long); utf8.RuneCountInString(got) > 90 {
		t.Errorf("shortCaption did not truncate: %d chars", utf8.RuneCountInString(got))
	}
}
