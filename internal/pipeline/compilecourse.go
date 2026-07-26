package pipeline

// compile-course: stitch every rendered lesson video into one long course
// video ("editing the template into a bigger video"). All lesson finals come
// from the same Remotion settings, so the ffmpeg concat demuxer joins them
// losslessly (stream copy, no re-encode). A YouTube-format chapter file with
// one entry per lesson is written next to it.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/enfec/coursesmith/internal/project"
)

// Course-level compile outputs, in the course directory.
const (
	CourseVideoName    = "course.mp4"
	CourseChaptersName = "course-chapters.txt"
)

// CompileCourse concatenates every lesson's final.mp4 (in lesson order) into
// <course>/course.mp4 and writes course-chapters.txt. Lessons without a
// rendered final are skipped with a warning.
func (e *Env) CompileCourse(ctx context.Context, course *project.Course) error {
	if _, err := e.CheckFFmpeg(); err != nil {
		return err
	}
	lessons, err := course.Lessons()
	if err != nil {
		return err
	}

	type part struct {
		lesson *project.Lesson
		path   string
		durMs  int
	}
	var parts []part
	for _, l := range lessons {
		path := filepath.Join(l.GeneratedDir(), FinalVideoName)
		if _, err := os.Stat(path); err != nil {
			fmt.Fprintf(e.out(), "  ⚠ compile   %s has no %s yet — skipping (run the pipeline first)\n", l.ID, FinalVideoName)
			continue
		}
		durMs, err := mediaDurationMs(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		parts = append(parts, part{lesson: l, path: path, durMs: durMs})
	}
	if len(parts) == 0 {
		return fmt.Errorf("no rendered lessons found — run `coursesmith run %s` first", course.Slug)
	}

	// The concat demuxer needs a list file; single-quote paths per its rules.
	var list strings.Builder
	for _, p := range parts {
		fmt.Fprintf(&list, "file '%s'\n", strings.ReplaceAll(p.path, "'", `'\''`))
	}
	listPath := filepath.Join(course.Dir, ".concat-list.txt")
	if err := writeFileAtomic(listPath, []byte(list.String())); err != nil {
		return err
	}
	defer os.Remove(listPath)

	outPath := filepath.Join(course.Dir, CourseVideoName)
	fmt.Fprintf(e.out(), "  → compile   joining %d lesson video(s) into %s...\n", len(parts), CourseVideoName)
	if err := e.runFFmpeg(ctx, "-y", "-f", "concat", "-safe", "0", "-i", listPath, "-c", "copy", outPath); err != nil {
		return fmt.Errorf("concatenating lesson videos: %w", err)
	}

	// YouTube chapters: MM:SS (or HH:MM:SS past an hour) + title per lesson.
	var chapters strings.Builder
	atMs := 0
	for _, p := range parts {
		chapters.WriteString(chapterStamp(atMs) + " " + p.lesson.FrontMatter.Title + "\n")
		atMs += p.durMs
	}
	if err := writeFileAtomic(filepath.Join(course.Dir, CourseChaptersName), []byte(chapters.String())); err != nil {
		return err
	}

	total := atMs / 1000
	fmt.Fprintf(e.out(), "    %s written (%d lessons, %dm%02ds) + %s\n",
		CourseVideoName, len(parts), total/60, total%60, CourseChaptersName)
	return nil
}

// chapterStamp formats milliseconds as a YouTube chapter timestamp.
func chapterStamp(ms int) string {
	s := ms / 1000
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s%3600/60, s%60)
	}
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}
