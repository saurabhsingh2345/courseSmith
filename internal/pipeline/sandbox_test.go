package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"
)

// A program split across files is what the workspace template is about, and it
// cannot go down a pipe: `import greet` only resolves when greet.py is a real
// file beside the script importing it. Piped through stdin the same program is
// a ModuleNotFoundError — which is exactly the output a clip must never show.
func TestRunProjectResolvesImportsAcrossFiles(t *testing.T) {
	runner, note := ResolveCodeRunner(context.Background())
	if runner == nil {
		t.Skip("no python3 and no docker on this machine")
	}
	t.Logf("runner: %s %s", runner.Name(), note)

	files := map[string]string{
		"greet.py": "def hello(who):\n    return f\"Hello, {who}!\"\n",
		"main.py":  "from greet import hello\n\nprint(hello(\"Ada\"))\n",
	}
	res, err := runner.RunProject(context.Background(), files, "main.py", 20*time.Second)
	if err != nil {
		t.Fatalf("RunProject: %v", err)
	}
	if got := strings.TrimSpace(res.Stdout); got != "Hello, Ada!" {
		t.Errorf("stdout = %q, want %q (stderr: %s)", got, "Hello, Ada!", res.Stderr)
	}
	if res.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", res.ExitCode)
	}
}

// A failing project still comes back as a result rather than an error — a
// traceback is a legitimate thing for a clip to show, and it is the caller
// that decides whether it is fatal.
func TestRunProjectReportsATracebackAsOutput(t *testing.T) {
	runner, _ := ResolveCodeRunner(context.Background())
	if runner == nil {
		t.Skip("no python3 and no docker on this machine")
	}
	files := map[string]string{"main.py": "raise ValueError(\"boom\")\n"}
	res, err := runner.RunProject(context.Background(), files, "main.py", 20*time.Second)
	if err != nil {
		t.Fatalf("RunProject returned an infrastructure error for a failing program: %v", err)
	}
	if res.ExitCode == 0 {
		t.Error("a raising program exited 0")
	}
	if !strings.Contains(res.Stderr, "ValueError") {
		t.Errorf("stderr = %q, want it to carry the traceback", res.Stderr)
	}
}

// The file paths come from a model, so a path that climbs out of the working
// directory has to be refused rather than written wherever it leads. Docker
// would contain it; the host runner would not.
func TestRunProjectRefusesEscapingPaths(t *testing.T) {
	for _, bad := range []string{"../escape.py", "/etc/passwd", "a/../../b.py"} {
		if _, err := writeProject(map[string]string{bad: "x", "main.py": "y"}, "main.py"); err == nil {
			t.Errorf("writeProject accepted the escaping path %q", bad)
		}
	}
	if _, err := writeProject(map[string]string{"main.py": "x"}, "other.py"); err == nil {
		t.Error("writeProject accepted an entry that is not one of the files")
	}
	if _, err := writeProject(nil, "main.py"); err == nil {
		t.Error("writeProject accepted an empty file set")
	}
}
