package pipeline

// Exercises stage: two practice exercises per lesson, each with starter
// code, hidden pytest tests, and a reference solution that is PROVEN
// solvable by executing it against the tests in the sandbox. The starter
// code must fail the same tests, or the exercise is trivially solved.
//
// Files land in generated/exercises/<slug>/ structured for later embedding
// on the course site:
//
//	README.md         — the exercise prompt
//	starter.py        — what the learner begins from
//	test_exercise.py  — hidden pytest tests
//	solution.py       — verified reference solution

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// Exercises stage outputs, under the lesson's generated dir.
const (
	ExercisesDirName      = "exercises"
	ExercisesManifestName = "exercises.json"
)

// exercisesCount is how many practice exercises each lesson ships.
const exercisesCount = 2

// pytestTimeout bounds one pytest verification run (pytest startup is
// slower than a bare script).
const pytestTimeout = 30 * time.Second

// Exercise is one practice exercise.
type Exercise struct {
	Slug         string `json:"slug"`
	Title        string `json:"title"`
	Description  string `json:"description"` // markdown prompt for the learner
	StarterCode  string `json:"starter_code"`
	SolutionCode string `json:"solution_code"`
	TestCode     string `json:"test_code"` // pytest file, imports from exercise.py
}

// ExercisesDoc is the LLM response and persisted manifest shape.
type ExercisesDoc struct {
	Exercises   []Exercise `json:"exercises"`
	Runner      string     `json:"runner,omitempty"`
	GeneratedAt time.Time  `json:"generated_at,omitempty"`
}

var exerciseSlugRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// exercisesPromptData feeds prompts/exercises.tmpl.
type exercisesPromptData struct {
	Audience  string
	Title     string
	Outline   string
	Narration string
	Count     int
}

// pytestProgram builds a self-contained Python program that materializes
// exercise.py (from the given source) plus the hidden tests and runs
// pytest over them inside the sandbox. The marker comment identifies which
// variant is being checked in logs and tests.
func pytestProgram(marker, moduleSource, testSource string) string {
	enc := base64.StdEncoding.EncodeToString
	return fmt.Sprintf(`# coursesmith-exercise-check: %s
import base64, os, subprocess, sys, tempfile

d = tempfile.mkdtemp()
files = {
    "exercise.py": %q,
    "test_exercise.py": %q,
}
for name, b64 in files.items():
    with open(os.path.join(d, name), "wb") as f:
        f.write(base64.b64decode(b64))

r = subprocess.run(
    [sys.executable, "-m", "pytest", "-q", "-p", "no:cacheprovider", d],
    capture_output=True, text=True, cwd=d,
)
sys.stdout.write(r.stdout[-2000:])
sys.stderr.write(r.stderr[-2000:] or r.stdout[-2000:])
sys.exit(r.returncode)
`, marker, enc([]byte(moduleSource)), enc([]byte(testSource)))
}

// validateExercises checks shape and PROVES each exercise: the reference
// solution must pass the hidden tests, and the starter code must not.
func validateExercises(ctx context.Context, e *Env, doc *ExercisesDoc) error {
	if len(doc.Exercises) != exercisesCount {
		return fmt.Errorf("got %d exercises, want exactly %d", len(doc.Exercises), exercisesCount)
	}
	seen := map[string]bool{}
	for i := range doc.Exercises {
		ex := &doc.Exercises[i]
		if !exerciseSlugRe.MatchString(ex.Slug) {
			return fmt.Errorf("exercise %d slug %q must be lowercase letters, digits, and hyphens", i+1, ex.Slug)
		}
		if seen[ex.Slug] {
			return fmt.Errorf("duplicate exercise slug %q", ex.Slug)
		}
		seen[ex.Slug] = true
		if strings.TrimSpace(ex.Title) == "" || strings.TrimSpace(ex.Description) == "" {
			return fmt.Errorf("exercise %q is missing title or description", ex.Slug)
		}
		if strings.TrimSpace(ex.StarterCode) == "" || strings.TrimSpace(ex.SolutionCode) == "" || strings.TrimSpace(ex.TestCode) == "" {
			return fmt.Errorf("exercise %q is missing starter_code, solution_code, or test_code", ex.Slug)
		}
		if !strings.Contains(ex.TestCode, "exercise") {
			return fmt.Errorf("exercise %q: test_code must import from the exercise module (`from exercise import ...`)", ex.Slug)
		}

		solution, err := e.CodeRunner.RunPython(ctx, pytestProgram("solution "+ex.Slug, ex.SolutionCode, ex.TestCode), pytestTimeout)
		if err != nil {
			return err
		}
		if strings.Contains(solution.Stderr, "No module named pytest") {
			return fmt.Errorf("pytest is not installed in the sandbox — rebuild it: %s", sandboxBuildHelp)
		}
		if !solution.Ok() {
			return fmt.Errorf("exercise %q: the reference solution FAILS its own tests:\n%s",
				ex.Slug, tailLines(solution.Stdout+solution.Stderr, 6))
		}
		starter, err := e.CodeRunner.RunPython(ctx, pytestProgram("starter "+ex.Slug, ex.StarterCode, ex.TestCode), pytestTimeout)
		if err != nil {
			return err
		}
		if starter.Ok() {
			return fmt.Errorf("exercise %q: the STARTER code already passes the tests — there is nothing to solve; make the starter a genuine skeleton", ex.Slug)
		}
	}
	return nil
}

// exerciseFiles writes one exercise's file bundle for the course site.
func exerciseFiles(l *project.Lesson, ex *Exercise) error {
	dir := filepath.Join(l.GeneratedDir(), ExercisesDirName, ex.Slug)
	readme := fmt.Sprintf("# %s\n\n%s\n\nStart from `starter.py`. The hidden tests in `test_exercise.py` must pass.\n",
		ex.Title, strings.TrimSpace(ex.Description))
	for name, content := range map[string]string{
		"README.md":        readme,
		"starter.py":       strings.TrimSpace(ex.StarterCode) + "\n",
		"solution.py":      strings.TrimSpace(ex.SolutionCode) + "\n",
		"test_exercise.py": strings.TrimSpace(ex.TestCode) + "\n",
	} {
		if err := writeFileAtomic(filepath.Join(dir, name), []byte(content)); err != nil {
			return err
		}
	}
	return nil
}

// runExercisesStage generates the lesson's verified practice exercises.
func runExercisesStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	if e.CodeRunner == nil {
		return fmt.Errorf("no code runner available — exercises must be proven solvable; install docker and run `%s`", sandboxBuildHelp)
	}
	script, err := loadScript(l)
	if err != nil {
		return err
	}
	narrations := make([]string, 0, len(script.Sections))
	for _, sec := range script.Sections {
		narrations = append(narrations, sec.Narration)
	}

	data := exercisesPromptData{
		Audience:  cfg.Style.Audience,
		Title:     l.FrontMatter.Title,
		Outline:   l.Body,
		Narration: strings.Join(narrations, "\n\n"),
		Count:     exercisesCount,
	}
	system, user, err := e.renderPrompt(exercisesTemplateName, data)
	if err != nil {
		return err
	}

	fmt.Fprintf(e.out(), "  → exercises drafting %d exercise(s) (%s), proving solvable via %s...\n",
		exercisesCount, cfg.Pipeline.LLMContent, e.CodeRunner.Name())
	var doc ExercisesDoc
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.4, 8192, &doc,
		func() error { return validateExercises(ctx, e, &doc) })
	if err != nil {
		return fmt.Errorf("generating exercises: %w", err)
	}
	doc.Runner = e.CodeRunner.Name()
	doc.GeneratedAt = time.Now().UTC()

	// Clear stale bundles from earlier runs, then write fresh ones.
	exDir := filepath.Join(l.GeneratedDir(), ExercisesDirName)
	if err := os.RemoveAll(exDir); err != nil {
		return fmt.Errorf("clearing %s: %w", exDir, err)
	}
	for i := range doc.Exercises {
		if err := exerciseFiles(l, &doc.Exercises[i]); err != nil {
			return err
		}
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), ExercisesDirName, ExercisesManifestName), &doc); err != nil {
		return err
	}
	slugs := make([]string, 0, len(doc.Exercises))
	for _, ex := range doc.Exercises {
		slugs = append(slugs, ex.Slug)
	}
	fmt.Fprintf(e.out(), "    verified solvable: %s → %s/\n", strings.Join(slugs, ", "), ExercisesDirName)
	return nil
}
