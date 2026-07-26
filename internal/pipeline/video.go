package pipeline

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// Stage 7 outputs, under the lesson's generated dir.
const (
	FinalVideoName  = "final.mp4"
	BurnedVideoName = "final-burned.mp4"
	// RecordingFileName, if present in the lesson dir (not generated/), is a
	// screen recording to assemble instead of slides mode.
	RecordingFileName = "recording.mp4"

	segmentsDirName = "video-segments"
)

// runVideoStage is Stage 7: assemble final.mp4 (soft subs) and
// final-burned.mp4 (subtitles rendered into the pixels).
//
// With lessons/<id>/recording.mp4 present, the recording's video is used and
// its audio replaced by the voiceover. Otherwise "slides mode" renders one
// 1920x1080 title card per script section, timed to the section's actual
// narration audio, so a lesson is publishable without a screen recording.
func runVideoStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	if _, err := e.CheckFFmpeg(); err != nil {
		return err
	}
	voiceover := filepath.Join(l.GeneratedDir(), VoiceoverFileName)
	captions := filepath.Join(l.GeneratedDir(), CaptionsFileName)
	for _, dep := range []struct{ path, stage string }{
		{voiceover, "audio"},
		{captions, "captions"},
	} {
		if _, err := os.Stat(dep.path); err != nil {
			return fmt.Errorf("no %s yet — the %s stage must run first", filepath.Base(dep.path), dep.stage)
		}
	}

	final := filepath.Join(l.GeneratedDir(), FinalVideoName)
	burned := filepath.Join(l.GeneratedDir(), BurnedVideoName)
	recording := filepath.Join(l.Dir, RecordingFileName)

	var videoSource string
	if _, err := os.Stat(recording); err == nil {
		fmt.Fprintf(e.out(), "  → video     assembling from %s...\n", RecordingFileName)
		videoSource = recording
	} else {
		fmt.Fprintf(e.out(), "  → video     no %s — rendering slides mode...\n", RecordingFileName)
		videoSource, err = renderSlides(ctx, e, l, cfg)
		if err != nil {
			return err
		}
	}

	if err := e.runFFmpeg(ctx, muxArgs(videoSource, voiceover, captions, final)...); err != nil {
		return err
	}
	fmt.Fprintf(e.out(), "    %s written (soft subtitles)\n", FinalVideoName)

	if err := e.runFFmpeg(ctx, burnArgs(videoSource, voiceover, captions, burned)...); err != nil {
		// The soft-sub deliverable exists; a missing libass build shouldn't
		// kill the pipeline over the optional burned variant.
		fmt.Fprintf(e.out(), "  ⚠ video     burned-subtitle variant failed (ffmpeg without libass?): %v\n", err)
		return nil
	}
	fmt.Fprintf(e.out(), "    %s written (burned-in subtitles)\n", BurnedVideoName)
	return nil
}

// muxArgs combines a video source, the voiceover, and soft mov_text
// subtitles into an MP4 (libx264 crf 20, aac).
func muxArgs(videoSrc, voiceover, captions, out string) []string {
	return []string{
		"-y",
		"-i", videoSrc,
		"-i", voiceover,
		"-i", captions,
		"-map", "0:v", "-map", "1:a", "-map", "2:s",
		"-c:v", "libx264", "-crf", "20", "-preset", "medium", "-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-c:s", "mov_text",
		"-metadata:s:s:0", "language=eng",
		"-shortest",
		out,
	}
}

// burnArgs renders the subtitles into the video pixels (needs libass).
func burnArgs(videoSrc, voiceover, captions, out string) []string {
	return []string{
		"-y",
		"-i", videoSrc,
		"-i", voiceover,
		"-vf", "subtitles=" + escapeFilterPath(captions),
		"-map", "0:v", "-map", "1:a",
		"-c:v", "libx264", "-crf", "20", "-preset", "medium", "-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-shortest",
		out,
	}
}

// renderSlides produces a silent 1920x1080 video: one branded title card per
// section (PNG rendered in Go), each lasting exactly as long as that
// section's narration WAV. It returns the path of the concatenated video.
func renderSlides(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config) (string, error) {
	script, err := loadScript(l)
	if err != nil {
		return "", err
	}
	fnt, err := findFont()
	if err != nil {
		return "", err
	}

	segDir := filepath.Join(l.GeneratedDir(), segmentsDirName)
	if err := os.MkdirAll(segDir, 0o755); err != nil {
		return "", fmt.Errorf("creating %s: %w", segDir, err)
	}

	segments := make([]string, 0, len(script.Sections))
	for i, sec := range script.Sections {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		dur, err := sectionDuration(l, sec)
		if err != nil {
			return "", err
		}
		card := filepath.Join(segDir, fmt.Sprintf("%03d-%s.png", i+1, sec.ID))
		spec := slideSpec{
			Heading:  humanizeSlug(sec.ID),
			Subtitle: script.Title,
			Colors:   cfg.Branding.Colors,
		}
		if err := renderTitleCard(fnt, spec, card); err != nil {
			return "", fmt.Errorf("rendering title card for section %q: %w", sec.ID, err)
		}
		seg := filepath.Join(segDir, fmt.Sprintf("%03d-%s.mp4", i+1, sec.ID))
		if err := e.runFFmpeg(ctx, slideArgs(card, dur, seg)...); err != nil {
			return "", fmt.Errorf("rendering slide for section %q: %w", sec.ID, err)
		}
		segments = append(segments, seg)
	}

	slides := filepath.Join(segDir, "slides.mp4")
	if err := concatVideos(ctx, e, segments, slides); err != nil {
		return "", err
	}
	return slides, nil
}

// sectionDuration prefers the section's actual synthesized audio length and
// falls back to the script's estimate.
func sectionDuration(l *project.Lesson, sec Section) (time.Duration, error) {
	wav := filepath.Join(l.GeneratedDir(), AudioDirName, sec.ID+".wav")
	if _, err := os.Stat(wav); err == nil {
		return wavDuration(wav)
	}
	return time.Duration(sec.DurationEstSec) * time.Second, nil
}

// slideArgs turns one still title card into a video segment of the given
// duration.
func slideArgs(cardPNG string, dur time.Duration, out string) []string {
	return []string{
		"-y",
		"-loop", "1",
		"-framerate", "30",
		"-i", cardPNG,
		"-t", fmt.Sprintf("%.3f", dur.Seconds()),
		"-c:v", "libx264", "-crf", "20", "-preset", "medium", "-pix_fmt", "yuv420p",
		out,
	}
}

// concatVideos joins same-codec MP4 segments via the concat demuxer.
func concatVideos(ctx context.Context, e *Env, paths []string, out string) error {
	if len(paths) == 0 {
		return fmt.Errorf("no video segments to concatenate")
	}
	var list strings.Builder
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("resolving %s: %w", p, err)
		}
		fmt.Fprintf(&list, "file '%s'\n", strings.ReplaceAll(abs, "'", `'\''`))
	}
	listPath := out + ".concat.txt"
	if err := writeFileAtomic(listPath, []byte(list.String())); err != nil {
		return err
	}
	defer os.Remove(listPath)
	return e.runFFmpeg(ctx, "-y", "-f", "concat", "-safe", "0", "-i", listPath, "-c", "copy", out)
}

// humanizeSlug turns "your-first-line-of-python" into "Your First Line Of Python".
func humanizeSlug(slug string) string {
	words := strings.Split(slug, "-")
	for i, w := range words {
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}

// escapeFilterPath escapes a file path for use inside a filtergraph value
// (the subtitles= argument).
func escapeFilterPath(p string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`:`, `\:`,
		`'`, `\'`,
	)
	return r.Replace(p)
}
