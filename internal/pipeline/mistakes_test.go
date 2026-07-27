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

// mistakesBody is a canned LLM response: broken code carries the BROKEN
// marker so fakeRunner can be scripted to fail it.
func mistakesBody() string {
	m := func(n int) string {
		return fmt.Sprintf(`{"title":"Mistake %d","explanation":"A misunderstanding.","broken_code":"BROKEN_%d = oops","fix":"Do it right.","fixed_code":"x = %d"}`, n, n, n)
	}
	return `{"mistakes":[` + m(1) + `,` + m(2) + `,` + m(3) + `]}`
}

// failingRunner scripts the sandbox: BROKEN code and starter exercises
// fail, everything else passes.
func failingRunner() *fakeRunner {
	return &fakeRunner{results: map[string]ExecResult{
		"BROKEN":                              {ExitCode: 1, Stderr: "Traceback (most recent call last):\n  File \"<stdin>\", line 1\nNameError: name 'oops' is not defined"},
		"coursesmith-exercise-check: starter": {ExitCode: 1, Stdout: "1 failed, 1 passed"},
	}}
}

func TestMistakesStageCapturesRealTracebacks(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson)
	fake := &fakeRouter{content: []string{mistakesBody()}}
	env, out := runEnv(t, fake)
	runner := failingRunner()
	env.CodeRunner = runner

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageMistakes}); err != nil {
		t.Fatal(err)
	}

	var doc MistakesDoc
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), MistakesFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Mistakes) != 3 {
		t.Fatalf("mistakes = %d, want 3", len(doc.Mistakes))
	}
	for i, m := range doc.Mistakes {
		if !strings.Contains(m.Traceback, "NameError") {
			t.Errorf("mistake %d traceback = %q, want the real NameError", i, m.Traceback)
		}
	}
	// 3 broken + 3 fixed executions.
	if runner.calls != 6 {
		t.Errorf("sandbox calls = %d, want 6", runner.calls)
	}
	if !strings.Contains(out.String(), "real tracebacks") {
		t.Errorf("output:\n%s", out.String())
	}
}

func TestMistakesStageRejectsCodeThatDoesNotFail(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson)
	// Two identical drafts (original + repair round); the runner never
	// fails anything, so both drafts are rejected.
	fake := &fakeRouter{content: []string{mistakesBody(), mistakesBody()}}
	env, _ := runEnv(t, fake)
	env.CodeRunner = &fakeRunner{} // everything exits 0

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageMistakes})
	if err == nil || !strings.Contains(err.Error(), "must actually fail") {
		t.Errorf("error = %v, want broken-code-must-fail rejection", err)
	}
}

func TestExceptionType(t *testing.T) {
	cases := []struct {
		name      string
		traceback string
		want      string
	}{
		{
			name: "ordinary traceback",
			traceback: `Traceback (most recent call last):
  File "<stdin>", line 1, in <module>
NameError: name 'x' is not defined`,
			want: "NameError",
		},
		{
			name: "syntax error caret form",
			traceback: `  File "<stdin>", line 1
    if x = 5:
         ^
SyntaxError: invalid syntax`,
			want: "SyntaxError",
		},
		{
			name:      "dotted class reduces to final segment",
			traceback: "json.decoder.JSONDecodeError: Expecting value",
			want:      "JSONDecodeError",
		},
		{
			name:      "trailing blank lines ignored",
			traceback: "IndentationError: expected an indented block\n\n",
			want:      "IndentationError",
		},
		{
			name:      "prose is not an exception line",
			traceback: "something went wrong somewhere",
			want:      "",
		},
		{
			name:      "empty",
			traceback: "",
			want:      "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := exceptionType(tc.traceback); got != tc.want {
				t.Errorf("exceptionType() = %q, want %q", got, tc.want)
			}
		})
	}
}
