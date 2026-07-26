package pipeline

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// SandboxImage is the Docker image with python3.12 + vhs + ffmpeg.
const SandboxImage = "coursesmith-sandbox"

// sandboxBuildHelp tells the user how to create the sandbox image.
const sandboxBuildHelp = "docker build -t " + SandboxImage + " sandbox/"

// ExecResult is the outcome of running a code snippet.
type ExecResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	TimedOut bool   `json:"timed_out"`
}

// Ok reports whether the snippet ran to completion successfully.
func (r ExecResult) Ok() bool { return r.ExitCode == 0 && !r.TimedOut }

// CodeRunner executes Python snippets. Implementations: DockerRunner
// (isolated, no network), HostRunner (fallback, warns), fakes in tests.
type CodeRunner interface {
	// Name identifies the runner for progress output ("docker-sandbox").
	Name() string
	// RunPython executes the source of one Python program (via stdin) and
	// captures its output. A deadline exceeded sets TimedOut instead of err;
	// err is reserved for infrastructure failures (docker missing, etc.).
	RunPython(ctx context.Context, code string, timeout time.Duration) (ExecResult, error)
}

// DockerRunner executes code inside the coursesmith sandbox image with
// networking disabled.
type DockerRunner struct{}

func (DockerRunner) Name() string { return "docker-sandbox" }

func (DockerRunner) RunPython(ctx context.Context, code string, timeout time.Duration) (ExecResult, error) {
	args := []string{
		"run", "--rm", "-i",
		"--network", "none",
		"--memory", "256m",
		SandboxImage,
		"python3", "-",
	}
	return runWithStdin(ctx, timeout, code, "docker", args...)
}

// HostRunner executes code with the host python3 — no isolation. Used only
// when docker is unavailable; the verify stage prints a warning.
type HostRunner struct {
	// Python is the interpreter binary; "" resolves "python3".
	Python string
}

func (h HostRunner) Name() string { return "host-python3 (UNSANDBOXED)" }

func (h HostRunner) RunPython(ctx context.Context, code string, timeout time.Duration) (ExecResult, error) {
	python := h.Python
	if python == "" {
		python = "python3"
	}
	if _, err := exec.LookPath(python); err != nil {
		return ExecResult{}, fmt.Errorf(
			"neither docker nor %s is available to execute lesson code — install docker and run `%s`",
			python, sandboxBuildHelp,
		)
	}
	return runWithStdin(ctx, timeout, code, python, "-")
}

// runWithStdin runs bin with args, feeding code on stdin, enforcing timeout.
func runWithStdin(ctx context.Context, timeout time.Duration, code, bin string, args ...string) (ExecResult, error) {
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, bin, args...)
	cmd.Stdin = strings.NewReader(code)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	res := ExecResult{Stdout: stdout.String(), Stderr: stderr.String()}
	switch {
	case err == nil:
		return res, nil
	case errors.Is(runCtx.Err(), context.DeadlineExceeded):
		res.TimedOut = true
		res.ExitCode = -1
		return res, nil
	default:
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			res.ExitCode = exitErr.ExitCode()
			return res, nil
		}
		// The binary itself could not run (docker daemon down, image
		// missing, interpreter absent) — infrastructure failure.
		return res, fmt.Errorf("running %s: %w\n%s", bin, err, tailLines(stderr.String(), 6))
	}
}

// ResolveCodeRunner picks the best available runner: docker with the sandbox
// image, else the host interpreter. The returned note is non-empty when the
// choice degrades isolation and should be surfaced to the user.
func ResolveCodeRunner(ctx context.Context) (CodeRunner, string) {
	if _, err := exec.LookPath("docker"); err != nil {
		return HostRunner{}, "docker is not installed — executing lesson code UNSANDBOXED on the host; install docker and run `" + sandboxBuildHelp + "`"
	}
	probe, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := exec.CommandContext(probe, "docker", "image", "inspect", SandboxImage).Run(); err != nil {
		return HostRunner{}, "sandbox image missing — executing lesson code UNSANDBOXED on the host; run `" + sandboxBuildHelp + "`"
	}
	return DockerRunner{}, ""
}
