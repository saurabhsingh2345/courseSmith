package pipeline

// A stage whose lesson is deleted while it runs used to fail with a bare
// "hashing …/lesson.md: no such file or directory" from the post-run
// bookkeeping — an error about a file the reader has every reason to believe is
// still there, naming neither the deletion nor the work that did survive.

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

func TestStageWhoseLessonVanishedMidRunSaysSo(t *testing.T) {
	l := testLesson(t)
	e := statusEnv(t)
	course := &project.Course{Slug: "test", Dir: l.Dir}

	// Stand in for the real thing: any stage long enough to overlap a delete
	// does this — writes its artifact, into a generated/ dir that gets silently
	// recreated, and leaves the source gone.
	const stage = project.StageScript
	original := stageFuncs[stage]
	stageFuncs[stage] = func(_ context.Context, _ *Env, _ *project.Course, l *project.Lesson, _ config.Config) error {
		if err := os.RemoveAll(l.Dir); err != nil {
			return err
		}
		return writeJSON(l.GeneratedDir()+"/script.json", map[string]string{"title": "written after the delete"})
	}
	t.Cleanup(func() { stageFuncs[stage] = original })

	err := e.runStages(context.Background(), course, l, config.Defaults(), []string{stage}, RunOptions{})
	if err == nil {
		t.Fatal("a run whose lesson was deleted mid-stage should fail")
	}
	for _, want := range []string{"removed while the stage was running", project.LessonFileName, l.GeneratedDir()} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not mention %q:\n%v", want, err)
		}
	}
	if strings.Contains(err.Error(), "hashing") {
		t.Errorf("the deletion is still being reported as a hashing failure:\n%v", err)
	}
}
