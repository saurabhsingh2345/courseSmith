package pipeline

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// traceRunner returns a HostRunner backed by the real python3, skipping the
// test when no interpreter is available. The trace stage is exercised through
// the exact production path (buildTraceProgram → runner → JSON).
func traceRunner(t *testing.T) CodeRunner {
	t.Helper()
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not available")
	}
	return HostRunner{}
}

// runTrace traces one code block via the real interpreter.
func runTrace(t *testing.T, code string) *CodeTrace {
	t.Helper()
	res, err := traceRunner(t).RunPython(context.Background(), buildTraceProgram(code), codeTimeout)
	if err != nil {
		t.Fatalf("running tracer: %v", err)
	}
	if res.TimedOut {
		t.Fatal("tracer timed out")
	}
	var tr CodeTrace
	if err := json.Unmarshal([]byte(res.Stdout), &tr); err != nil {
		t.Fatalf("tracer output is not valid JSON: %v\nstdout: %q\nstderr: %q", err, res.Stdout, res.Stderr)
	}
	return &tr
}

func TestTracerCapturesLoopAndCall(t *testing.T) {
	tr := runTrace(t, "nums = [3, 1, 4]\ntotal = 0\nfor n in nums:\n    total += n\n\ndef double(x):\n    return x * 2\n\nresult = double(total)\nprint(\"total\", total)\n")

	if tr.Error != "" {
		t.Fatalf("unexpected trace error: %s", tr.Error)
	}
	if len(tr.Steps) == 0 {
		t.Fatal("no steps recorded")
	}
	// The program's own output is captured out-of-band, not mixed into JSON.
	if tr.Stdout != "total 8\n" {
		t.Errorf("captured stdout = %q, want %q", tr.Stdout, "total 8\n")
	}

	// The accumulator must pass through its intermediate values in order.
	var totals []string
	for _, s := range tr.Steps {
		for _, v := range s.Vars {
			if v.Name == "total" {
				if n := len(totals); n == 0 || totals[n-1] != v.Value.Repr {
					totals = append(totals, v.Value.Repr)
				}
			}
		}
	}
	if got := lastValue(tr, "total"); got != "8" {
		t.Errorf("final total = %s, want 8", got)
	}
	if got := lastValue(tr, "result"); got != "16" {
		t.Errorf("result = %s, want 16 (double of 8)", got)
	}

	// The call into double() must show up as a nested stack frame.
	sawCall := false
	for _, s := range tr.Steps {
		if s.Func == "double" && len(s.Stack) == 2 && s.Stack[1] == "double" {
			sawCall = true
			if lastVarInStep(s, "x") != "8" {
				t.Errorf("inside double: x = %s, want 8", lastVarInStep(s, "x"))
			}
		}
	}
	if !sawCall {
		t.Error("no step recorded inside the double() call frame")
	}
}

func TestTracerRendersContainers(t *testing.T) {
	tr := runTrace(t, "d = {}\nd['a'] = [1, 2]\nnums = (10, 20, 30)\n")
	if tr.Error != "" {
		t.Fatalf("unexpected error: %s", tr.Error)
	}

	// Find the final step and inspect the structured container values.
	last := tr.Steps[len(tr.Steps)-1]
	byName := map[string]TraceValue{}
	for _, v := range last.Vars {
		byName[v.Name] = v.Value
	}
	d, ok := byName["d"]
	if !ok || d.Type != "dict" || len(d.Entries) != 1 {
		t.Fatalf("d = %+v, want a dict with 1 entry", d)
	}
	if d.Entries[0].Key != "'a'" || d.Entries[0].Value.Type != "list" {
		t.Errorf("d['a'] entry = %+v, want key 'a' → list", d.Entries[0])
	}
	nums, ok := byName["nums"]
	if !ok || nums.Type != "tuple" || len(nums.Items) != 3 {
		t.Errorf("nums = %+v, want a 3-item tuple", nums)
	}
}

func TestTracerCapturesRuntimeError(t *testing.T) {
	tr := runTrace(t, "x = 1\ny = x / 0\n")
	if tr.Error == "" {
		t.Fatal("expected a runtime error to be recorded")
	}
	if want := "ZeroDivisionError"; !contains(tr.Error, want) {
		t.Errorf("error = %q, want it to mention %s", tr.Error, want)
	}
	// It still produced a valid, non-empty trace up to the failure.
	if len(tr.Steps) == 0 {
		t.Error("expected steps before the error")
	}
}

func TestRunTraceStageWritesManifest(t *testing.T) {
	runner := traceRunner(t)
	dir := filepath.Join(t.TempDir(), "01-trace")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "---\ntitle: Trace me\n---\n\n## Adding up\n\n```python\ntotal = 0\nfor i in range(3):\n    total += i\nprint(total)\n```\n"
	if err := os.WriteFile(filepath.Join(dir, project.LessonFileName), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	l, err := project.LoadLesson(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(l.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}

	e := &Env{CodeRunner: runner, Out: io.Discard}
	if err := runTraceStage(context.Background(), e, nil, l, config.Defaults()); err != nil {
		t.Fatalf("runTraceStage: %v", err)
	}

	// Manifest exists with one entry, keyed by the block's code hash.
	traces, err := loadTraces(l)
	if err != nil {
		t.Fatalf("loadTraces: %v", err)
	}
	if len(traces) != 1 {
		t.Fatalf("loaded %d traces, want 1", len(traces))
	}
	block := extractCodeBlocks(l.Body, "lesson.md")[0]
	hash := project.HashBytes([]byte(block.Code))
	tr, ok := traces[hash]
	if !ok {
		t.Fatalf("no trace for block hash %s", hash)
	}
	if lastValue(tr, "total") != "3" {
		t.Errorf("final total = %s, want 3 (0+1+2)", lastValue(tr, "total"))
	}

	// Re-running reuses the cached trace file (no error, same result).
	if err := runTraceStage(context.Background(), e, nil, l, config.Defaults()); err != nil {
		t.Fatalf("re-run runTraceStage: %v", err)
	}
}

// lastValue returns the repr of the last-seen value of a module-level variable.
func lastValue(tr *CodeTrace, name string) string {
	val := ""
	for _, s := range tr.Steps {
		if v := lastVarInStep(s, name); v != "" {
			val = v
		}
	}
	return val
}

func lastVarInStep(s TraceStep, name string) string {
	for _, v := range s.Vars {
		if v.Name == name {
			return v.Value.Repr
		}
	}
	return ""
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
