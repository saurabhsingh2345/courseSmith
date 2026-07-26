package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

const testNotesYAML = `lessons:
  01-test:
    notes:
      - note: "Open with the kitchen-recipe analogy"
    sections:
      first-idea:
        - note: "Slow down over the diagram"
`

func TestReviewNotesInjectionAndResolution(t *testing.T) {
	course, lesson := testCourse(t)
	notesPath := filepath.Join(course.Dir, ReviewNotesFileName)
	if err := os.WriteFile(notesPath, []byte(testNotesYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	fake := &fakeRouter{content: []string{scriptJSON("Draft with notes applied.")}}
	env, out := runEnv(t, fake)
	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageScript}); err != nil {
		t.Fatal(err)
	}

	prompt := fake.contentReqs[0].Messages[1].Content
	for _, want := range []string{"kitchen-recipe analogy", "[section first-idea] Slow down over the diagram"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("script prompt missing note %q:\n%s", want, prompt)
		}
	}
	if !strings.Contains(out.String(), "review note(s) applied and marked resolved") {
		t.Errorf("output missing resolution notice:\n%s", out.String())
	}

	// Notes are now resolved on disk...
	notes, err := LoadReviewNotes(course.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := notes.UnresolvedText(lesson.ID); got != "" {
		t.Errorf("unresolved after run = %q, want none", got)
	}
	// ...and the applied set is recorded for the audit trail.
	if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), ReviewsDirName, NotesAppliedFileName)); err != nil {
		t.Errorf("notes-applied record missing: %v", err)
	}

	// The stage is stable: a re-run skips (resolution happened before the
	// runner recorded post-run hashes). The empty router proves no LLM call.
	env2, out2 := runEnv(t, &fakeRouter{})
	env2.PromptsDir = env.PromptsDir
	if err := env2.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageScript}); err != nil {
		t.Fatalf("re-run should have been skipped: %v", err)
	}
	if !strings.Contains(out2.String(), "up to date") {
		t.Errorf("re-run was not skipped:\n%s", out2.String())
	}

	// A NEW note makes the stage stale again.
	appended := strings.Replace(testNotesYAML, `- note: "Open with the kitchen-recipe analogy"`,
		`- note: "Open with the kitchen-recipe analogy"
        resolved: true
      - note: "Mention the snake joke"`, 1)
	if err := os.WriteFile(notesPath, []byte(appended), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := config.Resolve(course.Config, lesson.FrontMatter.Overrides(), config.Config{})
	statuses, err := env.LessonStatus(lesson, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if statuses[project.StageScript] != project.StatusStale {
		t.Errorf("script status after new note = %s, want stale", statuses[project.StageScript])
	}
}

func TestExportReview(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson)
	if err := os.WriteFile(filepath.Join(lesson.GeneratedDir(), QuizFileName), []byte(quizBody()), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(lesson.GeneratedDir(), ReviewsDirName), 0o755); err != nil {
		t.Fatal(err)
	}
	mistake := `{"mistakes":[{"title":"Missing quotes","explanation":"why","broken_code":"print(Hi)","traceback":"NameError: name 'Hi' is not defined","fix":"quote it","fixed_code":"print(\"Hi\")"}]}`
	if err := os.WriteFile(filepath.Join(lesson.GeneratedDir(), MistakesFileName), []byte(mistake), 0o644); err != nil {
		t.Fatal(err)
	}

	env, _ := runEnv(t, &fakeRouter{})
	outDir, err := ExportReview(env, course)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := os.ReadFile(filepath.Join(outDir, lesson.ID+".md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(doc)
	for _, want := range []string{
		"# 01-test — Test Lesson",
		"## Narration script",
		"### first-idea",
		"## Quiz",
		"[recall]",
		"- [x] b", // the marked answer
		"## Common mistakes",
		"NameError: name 'Hi' is not defined",
		"## Leaving feedback",
		"review-notes.yaml",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("review doc missing %q", want)
		}
	}
}
