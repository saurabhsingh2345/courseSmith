package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/project"
)

func exercisesBody() string {
	ex := func(slug string) string {
		return fmt.Sprintf(`{"slug":%q,"title":"Do %s","description":"Build the %s thing.",`+
			`"starter_code":"def f():\n    ...","solution_code":"def f():\n    return 1",`+
			`"test_code":"from exercise import f\n\ndef test_f():\n    assert f() == 1"}`, slug, slug, slug)
	}
	return `{"exercises":[` + ex("first-task") + `,` + ex("second-task") + `]}`
}

func TestExercisesStageWritesVerifiedBundles(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson)
	fake := &fakeRouter{content: []string{exercisesBody()}}
	env, out := runEnv(t, fake)
	runner := failingRunner() // solutions pass, starters fail
	env.CodeRunner = runner

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageExercises}); err != nil {
		t.Fatal(err)
	}

	for _, slug := range []string{"first-task", "second-task"} {
		dir := filepath.Join(lesson.GeneratedDir(), ExercisesDirName, slug)
		for _, name := range []string{"README.md", "starter.py", "solution.py", "test_exercise.py"} {
			if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
				t.Errorf("missing %s/%s: %v", slug, name, err)
			}
		}
	}
	var doc ExercisesDoc
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), ExercisesDirName, ExercisesManifestName))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Exercises) != 2 {
		t.Errorf("manifest exercises = %d, want 2", len(doc.Exercises))
	}
	// 2 exercises × (solution + starter) verification runs.
	if runner.calls != 4 {
		t.Errorf("sandbox calls = %d, want 4", runner.calls)
	}
	if !strings.Contains(out.String(), "verified solvable") {
		t.Errorf("output:\n%s", out.String())
	}
}

func TestExercisesStageRejectsTrivialStarter(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson)
	fake := &fakeRouter{content: []string{exercisesBody(), exercisesBody()}}
	env, _ := runEnv(t, fake)
	env.CodeRunner = &fakeRunner{} // everything passes — including starters

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageExercises})
	if err == nil || !strings.Contains(err.Error(), "STARTER code already passes") {
		t.Errorf("error = %v, want trivial-starter rejection", err)
	}
}

func TestExercisesStageRejectsFailingSolution(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson)
	fake := &fakeRouter{content: []string{exercisesBody(), exercisesBody()}}
	env, _ := runEnv(t, fake)
	env.CodeRunner = &fakeRunner{results: map[string]ExecResult{
		"coursesmith-exercise-check": {ExitCode: 1, Stdout: "2 failed"},
	}}

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageExercises})
	if err == nil || !strings.Contains(err.Error(), "FAILS its own tests") {
		t.Errorf("error = %v, want failing-solution rejection", err)
	}
}
