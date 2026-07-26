package pipeline

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// ffmpegBin returns the ffmpeg binary to run.
func (e *Env) ffmpegBin() string {
	if e.FFmpeg != "" {
		return e.FFmpeg
	}
	return "ffmpeg"
}

// CheckFFmpeg verifies ffmpeg is runnable and returns its version line.
// Stages that shell out call it first so users get one clear error instead
// of an exec failure mid-pipeline.
func (e *Env) CheckFFmpeg() (string, error) {
	bin := e.ffmpegBin()
	if _, err := exec.LookPath(bin); err != nil {
		return "", fmt.Errorf(
			"ffmpeg not found (looked for %q) — install it (macOS: brew install ffmpeg, Debian/Ubuntu: apt install ffmpeg) and re-run; `coursesmith doctor` checks your whole setup",
			bin,
		)
	}
	out, err := exec.Command(bin, "-version").Output()
	if err != nil {
		return "", fmt.Errorf("running %s -version: %w", bin, err)
	}
	version, _, _ := strings.Cut(string(out), "\n")
	return version, nil
}

// runFFmpeg executes ffmpeg with args, surfacing the tail of stderr on
// failure (ffmpeg writes its diagnostics there).
func (e *Env) runFFmpeg(ctx context.Context, args ...string) error {
	if _, err := e.CheckFFmpeg(); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, e.ffmpegBin(), args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg %s: %w\n%s", strings.Join(args, " "), err, tailLines(stderr.String(), 12))
	}
	return nil
}

// runFFmpegCapture executes ffmpeg and returns its stderr even on success
// (loudnorm prints its measurement report there).
func (e *Env) runFFmpegCapture(ctx context.Context, args ...string) (string, error) {
	if _, err := e.CheckFFmpeg(); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, e.ffmpegBin(), args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stderr.String(), fmt.Errorf("ffmpeg %s: %w\n%s", strings.Join(args, " "), err, tailLines(stderr.String(), 12))
	}
	return stderr.String(), nil
}

// tailLines returns the last n non-empty lines of s.
func tailLines(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// mediaDurationMs probes a media file's duration with ffprobe (ships with
// ffmpeg).
func mediaDurationMs(path string) (int, error) {
	out, err := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		path,
	).Output()
	if err != nil {
		return 0, fmt.Errorf("probing duration of %s (is ffprobe installed alongside ffmpeg?): %w", path, err)
	}
	var seconds float64
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%f", &seconds); err != nil {
		return 0, fmt.Errorf("parsing ffprobe duration for %s: %w", path, err)
	}
	return int(seconds*1000 + 0.5), nil
}
