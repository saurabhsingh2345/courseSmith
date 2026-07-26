package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// DefaultSiteDir is the Hugo site skeleton at the project root.
const DefaultSiteDir = "site"

func (e *Env) siteDir() string {
	if e.SiteDir != "" {
		return e.SiteDir
	}
	return DefaultSiteDir
}

// runHugoStage is Stage 8: emit the lesson as a Hugo page bundle under
// site/content/courses/<course>/<lesson>/ — index.md with shortcodes plus
// copies of the generated assets (video, captions, diagrams, quiz).
func runHugoStage(ctx context.Context, e *Env, course *project.Course, l *project.Lesson, _ config.Config) error {
	script, err := loadScript(l)
	if err != nil {
		return err
	}
	// Everything embedded by the page must exist before we publish it.
	required := []struct{ name, stage string }{
		{QuizFileName, "quiz"},
		{CaptionsFileName, "captions"},
		{FinalVideoName, "video"},
	}
	for _, dep := range required {
		if _, err := os.Stat(filepath.Join(l.GeneratedDir(), dep.name)); err != nil {
			return fmt.Errorf("no %s yet — the %s stage must run first", dep.name, dep.stage)
		}
	}
	for _, d := range l.FrontMatter.Diagrams {
		svg := filepath.Join(l.GeneratedDir(), DiagramsDirName, d.ID+".svg")
		if _, err := os.Stat(svg); err != nil {
			return fmt.Errorf("diagram %s.svg missing — the visuals stage must run first", d.ID)
		}
	}

	bundleDir := filepath.Join(e.siteDir(), "content", "courses", course.Slug, l.ID)
	fmt.Fprintf(e.out(), "  → hugo      emitting page bundle %s...\n", bundleDir)
	if err := os.MkdirAll(filepath.Join(bundleDir, DiagramsDirName), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", bundleDir, err)
	}

	// Course section page (idempotent; shared by all lessons of the course).
	courseIndex := filepath.Join(e.siteDir(), "content", "courses", course.Slug, "_index.md")
	if err := writeFileAtomic(courseIndex, []byte(courseIndexMD(course))); err != nil {
		return err
	}

	// Copy generated assets into the page bundle. Required files hard-fail
	// above; the richer artifacts (chapters, transcript, alignment,
	// mistakes, exercises) copy when present so the site can use them.
	copies := map[string]string{
		filepath.Join(l.GeneratedDir(), FinalVideoName):   filepath.Join(bundleDir, FinalVideoName),
		filepath.Join(l.GeneratedDir(), CaptionsFileName): filepath.Join(bundleDir, CaptionsFileName),
	}
	for _, d := range l.FrontMatter.Diagrams {
		src := filepath.Join(l.GeneratedDir(), DiagramsDirName, d.ID+".svg")
		copies[src] = filepath.Join(bundleDir, DiagramsDirName, d.ID+".svg")
	}
	for _, optional := range []string{
		ChaptersJSONFileName, TranscriptFileName, AlignmentFileName, MistakesFileName,
	} {
		src := filepath.Join(l.GeneratedDir(), optional)
		if _, err := os.Stat(src); err == nil {
			copies[src] = filepath.Join(bundleDir, optional)
		}
	}
	for src, dst := range copies {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	// Exercise bundles (starter + README + tests; solutions stay out of the
	// public site).
	if manifest, err := os.ReadFile(filepath.Join(l.GeneratedDir(), ExercisesDirName, ExercisesManifestName)); err == nil {
		var doc ExercisesDoc
		if json.Unmarshal(manifest, &doc) == nil {
			public := ExercisesDoc{Runner: doc.Runner, GeneratedAt: doc.GeneratedAt}
			for _, ex := range doc.Exercises {
				ex.SolutionCode = "" // never publish reference solutions
				public.Exercises = append(public.Exercises, ex)
			}
			if err := writeJSON(filepath.Join(bundleDir, "exercises.json"), public); err != nil {
				return err
			}
		}
	}
	// Course-level concept map (produced by coursesmith analyze).
	if svg := filepath.Join(courseDirOf(l.Dir), CourseGeneratedDirName, ConceptsSVGFileName); fileExists(svg) {
		if err := copyFile(svg, filepath.Join(e.siteDir(), "content", "courses", course.Slug, ConceptsSVGFileName)); err != nil {
			return err
		}
	}

	// The published quiz is the generated one with human overrides
	// (quiz-overrides.yaml) merged on top — edits survive regeneration.
	mergedQuiz, _, overrides, err := LoadQuizWithOverrides(l)
	if err != nil {
		return err
	}
	if mergedQuiz == nil {
		return fmt.Errorf("no %s yet — the quiz stage must run first", QuizFileName)
	}
	if err := writeJSON(filepath.Join(bundleDir, QuizFileName), mergedQuiz); err != nil {
		return err
	}
	if overrides != nil && len(overrides.Questions) > 0 {
		fmt.Fprintf(e.out(), "    %d human quiz override(s) merged\n", len(overrides.Questions))
	}

	page := lessonPageMD(l, script)
	if err := writeFileAtomic(filepath.Join(bundleDir, "index.md"), []byte(page)); err != nil {
		return err
	}
	fmt.Fprintf(e.out(), "    index.md + %d asset(s) written\n", len(copies))
	return nil
}

// courseIndexMD renders the course's _index.md front-matter.
func courseIndexMD(course *project.Course) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %q\n", course.Name)
	fmt.Fprintf(&b, "description: %q\n", strings.TrimSpace(course.Description))
	b.WriteString("---\n")
	return b.String()
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// lessonDurationSec reads the real narration length from chapters.json
// (end of the last chapter); 0 when chapters have not been generated.
func lessonDurationSec(l *project.Lesson) int {
	data, err := os.ReadFile(filepath.Join(l.GeneratedDir(), ChaptersJSONFileName))
	if err != nil {
		return 0
	}
	var chapters []Chapter
	if json.Unmarshal(data, &chapters) != nil || len(chapters) == 0 {
		return 0
	}
	return chapters[len(chapters)-1].EndMs / 1000
}

// lessonPageMD renders the lesson's index.md: front-matter, the video with
// captions, per-section narration with inline diagrams at their cue
// positions, and the interactive quiz.
func lessonPageMD(l *project.Lesson, script *Script) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %q\n", l.FrontMatter.Title)
	fmt.Fprintf(&b, "weight: %d\n", lessonWeight(l.ID))
	if dur := lessonDurationSec(l); dur > 0 {
		fmt.Fprintf(&b, "duration_sec: %d\n", dur)
	}
	if len(l.FrontMatter.Outcomes) > 0 {
		b.WriteString("outcomes:\n")
		for _, o := range l.FrontMatter.Outcomes {
			fmt.Fprintf(&b, "  - %q\n", o)
		}
	}
	b.WriteString("---\n\n")

	fmt.Fprintf(&b, "{{< lesson-video src=%q captions=%q >}}\n", FinalVideoName, CaptionsFileName)

	diagramPrompts := make(map[string]string, len(l.FrontMatter.Diagrams))
	for _, d := range l.FrontMatter.Diagrams {
		diagramPrompts[d.ID] = d.Prompt
	}

	for _, sec := range script.Sections {
		fmt.Fprintf(&b, "\n## %s\n\n", humanizeSlug(sec.ID))
		fmt.Fprintf(&b, "%s\n", strings.TrimSpace(sec.Narration))
		for _, cue := range sec.Cues {
			if cue.Type != CueDiagram {
				continue
			}
			caption := shortCaption(diagramPrompts[cue.Ref])
			fmt.Fprintf(&b, "\n{{< diagram src=%q caption=%q >}}\n",
				DiagramsDirName+"/"+cue.Ref+".svg", caption)
		}
	}

	b.WriteString("\n## Check your understanding\n\n")
	fmt.Fprintf(&b, "{{< quiz src=%q >}}\n", QuizFileName)
	return b.String()
}

// lessonWeight orders lessons by their numeric directory prefix
// ("01-what-is-python" → 1); no prefix sorts last.
func lessonWeight(id string) int {
	digits, _, _ := strings.Cut(id, "-")
	n, err := strconv.Atoi(digits)
	if err != nil || n <= 0 {
		return 999
	}
	return n
}

// shortCaption trims a diagram prompt down to a one-line figure caption.
func shortCaption(prompt string) string {
	caption := strings.Join(strings.Fields(prompt), " ")
	if i := strings.IndexAny(caption, ".:;"); i > 0 {
		caption = caption[:i]
	}
	const max = 90
	if runes := []rune(caption); len(runes) > max {
		caption = string(runes[:max-1]) + "…"
	}
	return caption
}

// copyFile copies src to dst (atomic via the writeFileAtomic temp+rename).
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening %s: %w", src, err)
	}
	defer in.Close()
	data, err := io.ReadAll(in)
	if err != nil {
		return fmt.Errorf("reading %s: %w", src, err)
	}
	return writeFileAtomic(dst, data)
}
