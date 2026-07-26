package pipeline

// The code-visualisation stage (workstream C): for every Python block in the
// lesson outline, run it under the execution tracer (tracer.py) and persist a
// step-by-step trace — line, call stack, local variables (structured), and the
// stdout captured at that point. The scenegraph embeds these into code scenes
// so the renderer can show variable state changing in real time, Python
// Tutor-style. Runs after verify, so the code is already known to execute.

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// tracerBody is the instrumentation harness; buildTraceProgram prepends a
// USER_CODE definition before running it.
//
//go:embed tracer.py
var tracerBody string

// CodeTracesDirName holds one <blockhash>.json trace per code block plus the
// manifest the scenegraph depends on.
const CodeTracesDirName = "code_traces"

// TraceManifestFileName is the single stable artifact downstream stages watch;
// it changes whenever any block's trace changes.
const TraceManifestFileName = "manifest.json"

// TraceValue is a rendered runtime value: a scalar (repr only), or a container
// carrying its structured contents for the variable panel.
type TraceValue struct {
	Type    string       `json:"type"`
	Repr    string       `json:"repr"`
	Items   []TraceValue `json:"items,omitempty"`   // list/tuple/set
	Entries []TraceEntry `json:"entries,omitempty"` // dict
	Fields  []TraceField `json:"fields,omitempty"`  // object attributes
}

// TraceEntry is one dict key→value pair.
type TraceEntry struct {
	Key   string     `json:"key"`
	Value TraceValue `json:"value"`
}

// TraceField is one object attribute.
type TraceField struct {
	Key   string     `json:"key"`
	Value TraceValue `json:"value"`
}

// TraceVar is one named local variable at a step.
type TraceVar struct {
	Name  string     `json:"name"`
	Value TraceValue `json:"value"`
}

// TraceStep is one execution step (one source line executed).
type TraceStep struct {
	Step   int        `json:"step"`
	Line   int        `json:"line"`  // 1-based line in Code
	Event  string     `json:"event"` // line|return|exception
	Func   string     `json:"func"`  // enclosing function ("<module>" at top level)
	Vars   []TraceVar `json:"vars"`
	Stack  []string   `json:"stack"`  // call stack, outermost first
	Stdout string     `json:"stdout"` // cumulative program output at this step
}

// CodeTrace is the full trace of one code block.
type CodeTrace struct {
	Code      string      `json:"code"`
	Lines     []string    `json:"lines"`
	Steps     []TraceStep `json:"steps"`
	Truncated bool        `json:"truncated"`       // hit the tracer's step cap
	Error     string      `json:"error,omitempty"` // uncaught exception summary
	Stdout    string      `json:"stdout"`
}

// TraceManifestEntry links a code block to its trace file.
type TraceManifestEntry struct {
	Hash      string `json:"hash"`
	File      string `json:"file"`
	Source    string `json:"source"`
	Index     int    `json:"index"`
	Steps     int    `json:"steps"`
	Truncated bool   `json:"truncated,omitempty"`
	Error     string `json:"error,omitempty"`
}

// TraceManifest is code_traces/manifest.json.
type TraceManifest struct {
	Runner string               `json:"runner"`
	Traces []TraceManifestEntry `json:"traces"`
}

// buildTraceProgram wraps a code block into a standalone program: it decodes
// the block from base64 into USER_CODE, then runs the tracer body. Base64 keeps
// arbitrary block content (quotes, backslashes, newlines) from breaking the
// generated source.
func buildTraceProgram(code string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(code))
	return "import base64 as _cs_b64\n" +
		"USER_CODE = _cs_b64.b64decode(\"" + b64 + "\").decode(\"utf-8\")\n" +
		tracerBody
}

// loadTraces reads all persisted traces for a lesson keyed by block hash, via
// the manifest. A missing manifest returns an empty map (no trace stage yet).
func loadTraces(l *project.Lesson) (map[string]*CodeTrace, error) {
	dir := filepath.Join(l.GeneratedDir(), CodeTracesDirName)
	manifest, err := os.ReadFile(filepath.Join(dir, TraceManifestFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]*CodeTrace{}, nil
		}
		return nil, fmt.Errorf("reading trace manifest: %w", err)
	}
	var m TraceManifest
	if err := json.Unmarshal(manifest, &m); err != nil {
		return nil, fmt.Errorf("parsing trace manifest (delete %s to re-trace): %w", dir, err)
	}
	out := make(map[string]*CodeTrace, len(m.Traces))
	for _, e := range m.Traces {
		data, err := os.ReadFile(filepath.Join(dir, e.File))
		if err != nil {
			return nil, fmt.Errorf("reading trace %s: %w", e.File, err)
		}
		var tr CodeTrace
		if err := json.Unmarshal(data, &tr); err != nil {
			return nil, fmt.Errorf("parsing trace %s: %w", e.File, err)
		}
		out[e.Hash] = &tr
	}
	return out, nil
}

// runTraceStage traces every lesson-outline Python block. It never fails on
// user code that raises — the tracer records the exception in the trace — but
// it does fail if the tracer itself can't run or emits invalid JSON.
func runTraceStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, _ config.Config) error {
	if e.CodeRunner == nil {
		return fmt.Errorf("no code runner available — install docker and run `%s`", sandboxBuildHelp)
	}

	blocks := extractCodeBlocks(l.Body, "lesson.md")
	dir := filepath.Join(l.GeneratedDir(), CodeTracesDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	manifest := TraceManifest{Runner: e.CodeRunner.Name()}
	if len(blocks) == 0 {
		fmt.Fprintf(e.out(), "  → trace     no Python code blocks — nothing to trace\n")
		return writeJSON(filepath.Join(dir, TraceManifestFileName), manifest)
	}

	fmt.Fprintf(e.out(), "  → trace     instrumenting %d code block(s) via %s...\n", len(blocks), e.CodeRunner.Name())
	seen := map[string]bool{}
	totalSteps := 0
	for _, block := range blocks {
		if err := ctx.Err(); err != nil {
			return err
		}
		hash := project.HashBytes([]byte(block.Code))
		if seen[hash] {
			continue // identical block already traced this run
		}
		seen[hash] = true
		file := hash + ".json"
		path := filepath.Join(dir, file)

		tr, cached := reuseTrace(path, block.Code)
		if !cached {
			res, err := e.CodeRunner.RunPython(ctx, buildTraceProgram(block.Code), codeTimeout)
			if err != nil {
				return fmt.Errorf("trace %s block %d: %w", block.Source, block.Index+1, err)
			}
			if res.TimedOut {
				return fmt.Errorf("trace %s block %d: timed out after %s", block.Source, block.Index+1, codeTimeout)
			}
			tr = &CodeTrace{}
			if err := json.Unmarshal([]byte(res.Stdout), tr); err != nil {
				return fmt.Errorf("trace %s block %d: tracer produced invalid JSON: %w\n%s",
					block.Source, block.Index+1, err, tailLines(res.Stderr, 4))
			}
			if err := writeJSON(path, tr); err != nil {
				return err
			}
		}
		totalSteps += len(tr.Steps)
		manifest.Traces = append(manifest.Traces, TraceManifestEntry{
			Hash: hash, File: file, Source: block.Source, Index: block.Index,
			Steps: len(tr.Steps), Truncated: tr.Truncated, Error: tr.Error,
		})
	}

	if err := writeJSON(filepath.Join(dir, TraceManifestFileName), manifest); err != nil {
		return err
	}
	fmt.Fprintf(e.out(), "    %d block(s) traced, %d total step(s)\n", len(manifest.Traces), totalSteps)
	return nil
}

// reuseTrace returns a previously-written trace when it still matches the block
// (so unchanged blocks skip re-execution), else (nil, false).
func reuseTrace(path, code string) (*CodeTrace, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	var tr CodeTrace
	if json.Unmarshal(data, &tr) != nil || tr.Code != code {
		return nil, false
	}
	return &tr, true
}
