package pipeline

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// fakeRunner scripts code execution results by substring match on the code.
type fakeRunner struct {
	calls   int
	results map[string]ExecResult // substring of code → result
	err     error
}

func (f *fakeRunner) Name() string { return "fake-runner" }

func (f *fakeRunner) RunPython(_ context.Context, code string, _ time.Duration) (ExecResult, error) {
	f.calls++
	if f.err != nil {
		return ExecResult{}, f.err
	}
	for needle, res := range f.results {
		if strings.Contains(code, needle) {
			return res, nil
		}
	}
	return ExecResult{Stdout: "default\n"}, nil
}

// The multi-file path scripts on the entry file's source, the same way the
// single-file path scripts on the code.
func (f *fakeRunner) RunProject(_ context.Context, files map[string]string, entry string, _ time.Duration) (ExecResult, error) {
	f.calls++
	if f.err != nil {
		return ExecResult{}, f.err
	}
	for needle, res := range f.results {
		if strings.Contains(files[entry], needle) {
			return res, nil
		}
	}
	return ExecResult{Stdout: "default\n"}, nil
}

func TestExtractCodeBlocks(t *testing.T) {
	md := "# Lesson\n" +
		"```python\nprint(\"hi\")\n```\n" +
		"\n```output\nhi\n```\n" +
		"Some prose.\n" +
		"```bash\nls\n```\n" + // not python — skipped
		"```py\nx = 1\nprint(x)\n```\n" + // no output claim
		"```text\nnot a claim\n```\n"
	blocks := extractCodeBlocks(md, "lesson.md")
	if len(blocks) != 2 {
		t.Fatalf("blocks = %d, want 2: %+v", len(blocks), blocks)
	}
	if blocks[0].Code != `print("hi")` || blocks[0].ClaimedOutput != "hi" {
		t.Errorf("block 0 = %+v", blocks[0])
	}
	if blocks[1].Code != "x = 1\nprint(x)" || blocks[1].ClaimedOutput != "" {
		t.Errorf("block 1 = %+v", blocks[1])
	}
	if blocks[1].Index != 1 || blocks[1].Source != "lesson.md" {
		t.Errorf("block 1 metadata = %+v", blocks[1])
	}
}

func TestNormalizeOutput(t *testing.T) {
	tests := []struct{ in, want string }{
		{"hi\n", "hi"},
		{"hi  \nthere\t\n\n\n", "hi\nthere"},
		{"a\r\nb\r\n", "a\nb"},
	}
	for _, tt := range tests {
		if got := normalizeOutput(tt.in); got != tt.want {
			t.Errorf("normalizeOutput(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// lessonWithCode swaps in a lesson.md containing code blocks.
func lessonWithCode(t *testing.T, lesson *project.Lesson, body string) *project.Lesson {
	t.Helper()
	md := "---\ntitle: Test Lesson\ndiagrams:\n  - id: memory-model\n    prompt: \"3 variables\"\n---\n\n" + body
	if err := os.WriteFile(lesson.SourcePath(), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	reloaded, err := project.LoadLesson(lesson.Dir)
	if err != nil {
		t.Fatal(err)
	}
	return reloaded
}

func TestVerifyStagePassesAndRecords(t *testing.T) {
	course, lesson := testCourse(t)
	lesson = lessonWithCode(t, lesson,
		"## Code\n```python\nprint(\"hello\")\n```\n\n```output\nwrong claim\n```\n")
	seedScript(t, lesson)

	runner := &fakeRunner{results: map[string]ExecResult{
		`print("hello")`: {Stdout: "hello\n"},
	}}
	env, out := runEnv(t, &fakeRouter{})
	env.CodeRunner = runner

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVerify}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), VerificationFileName))
	if err != nil {
		t.Fatal(err)
	}
	var report VerificationReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Blocks) != 1 || report.Blocks[0].Stdout != "hello\n" {
		t.Fatalf("report = %+v", report)
	}
	if report.Blocks[0].OutputMatches == nil || *report.Blocks[0].OutputMatches {
		t.Error("claimed output mismatch not recorded")
	}
	if !strings.Contains(out.String(), "claimed output differs") {
		t.Errorf("output missing mismatch warning:\n%s", out.String())
	}
}

func TestVerifyStageFailsOnBrokenCode(t *testing.T) {
	course, lesson := testCourse(t)
	lesson = lessonWithCode(t, lesson,
		"## Code\n```python\nprint(undefined_name)\n```\n")
	seedScript(t, lesson)

	runner := &fakeRunner{results: map[string]ExecResult{
		"undefined_name": {Stderr: "NameError: name 'undefined_name' is not defined\n", ExitCode: 1},
	}}
	env, _ := runEnv(t, &fakeRouter{})
	env.CodeRunner = runner

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVerify})
	if err == nil {
		t.Fatal("broken code did not fail the pipeline")
	}
	for _, want := range []string{"lesson.md block 1", "NameError", "can be published"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q missing %q", err, want)
		}
	}

	// Failed blocks are still recorded for inspection.
	if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), VerificationFileName)); err != nil {
		t.Errorf("verification.json missing after failure: %v", err)
	}
	// And the stage is not marked done.
	cfg := configFor(course, lesson)
	env2 := &Env{PromptsDir: env.PromptsDir}
	statuses, err := env2.LessonStatus(lesson, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if statuses[project.StageVerify] == project.StatusDone {
		t.Error("failed verify stage recorded as done")
	}
}

func TestVerifyStageSkipsUnchangedBlocks(t *testing.T) {
	course, lesson := testCourse(t)
	lesson = lessonWithCode(t, lesson, "## Code\n```python\nprint(\"cached\")\n```\n")
	seedScript(t, lesson)

	runner := &fakeRunner{results: map[string]ExecResult{"cached": {Stdout: "cached\n"}}}
	env, _ := runEnv(t, &fakeRouter{})
	env.CodeRunner = runner

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVerify}); err != nil {
		t.Fatal(err)
	}
	if runner.calls != 1 {
		t.Fatalf("calls = %d, want 1", runner.calls)
	}

	// Force a re-run: the unchanged block must be served from verification.json.
	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVerify, Force: true}); err != nil {
		t.Fatal(err)
	}
	if runner.calls != 1 {
		t.Errorf("calls = %d after forced re-run, want 1 (hash cache)", runner.calls)
	}
}

func TestVerifyStageTimeout(t *testing.T) {
	course, lesson := testCourse(t)
	lesson = lessonWithCode(t, lesson, "## Code\n```python\nwhile True: pass\n```\n")
	seedScript(t, lesson)

	runner := &fakeRunner{results: map[string]ExecResult{
		"while True": {TimedOut: true, ExitCode: -1},
	}}
	env, _ := runEnv(t, &fakeRouter{})
	env.CodeRunner = runner

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVerify})
	if err == nil || !strings.Contains(err.Error(), "timed out") {
		t.Errorf("error = %v, want timeout failure", err)
	}
}

func TestVerifyQuizCode(t *testing.T) {
	quizWithCode := func(answerIndex int) *Quiz {
		return &Quiz{
			Title: "Q",
			Questions: []Question{{
				ID:          "q1",
				Prompt:      "What does this print?\n```python\nprint(2 + 2)\n```",
				Options:     []string{"3", "4", "22", "error"},
				AnswerIndex: answerIndex,
				Explanation: "Addition.",
			}},
		}
	}
	env := &Env{CodeRunner: &fakeRunner{results: map[string]ExecResult{
		"2 + 2": {Stdout: "4\n"},
	}}}

	if err := verifyQuizCode(context.Background(), env, quizWithCode(1)); err != nil {
		t.Errorf("correct answer rejected: %v", err)
	}
	err := verifyQuizCode(context.Background(), env, quizWithCode(0))
	if err == nil || !strings.Contains(err.Error(), `actually prints "4"`) {
		t.Errorf("wrong answer accepted: %v", err)
	}

	env.CodeRunner = &fakeRunner{results: map[string]ExecResult{
		"2 + 2": {Stderr: "SyntaxError\n", ExitCode: 1},
	}}
	err = verifyQuizCode(context.Background(), env, quizWithCode(1))
	if err == nil || !strings.Contains(err.Error(), "fails when executed") {
		t.Errorf("broken quiz code accepted: %v", err)
	}
}

func TestHostRunnerExecutesRealPython(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	runner := HostRunner{}
	res, err := runner.RunPython(context.Background(), `print("real output", 6*7)`, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok() || res.Stdout != "real output 42\n" {
		t.Errorf("result = %+v", res)
	}

	res, err = runner.RunPython(context.Background(), `raise ValueError("boom")`, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if res.Ok() || !strings.Contains(res.Stderr, "ValueError: boom") {
		t.Errorf("result = %+v", res)
	}

	res, err = runner.RunPython(context.Background(), "import time\ntime.sleep(30)", 500*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if !res.TimedOut {
		t.Errorf("timeout not detected: %+v", res)
	}
}

// configFor resolves the effective config the same way RunLesson does.
func configFor(course *project.Course, lesson *project.Lesson) config.Config {
	return config.Resolve(course.Config, lesson.FrontMatter.Overrides(), config.Config{})
}
