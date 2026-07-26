package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/pipeline"
	"github.com/enfec/coursesmith/internal/project"
)

// newAuthorCmd drafts a complete lesson.md from a one-line prompt and files
// it into the course as the next numbered lesson — the CLI twin of the
// Studio's Compose flow, and the building block for scripting a whole course:
//
//	coursesmith author python-fundamentals "the print function"
//	coursesmith author python-fundamentals "data types"
//	coursesmith run python-fundamentals
func newAuthorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "author <course> <prompt>",
		Short: "Draft a complete lesson from a prompt and file it into the course",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			course, err := resolveCourse(args[0])
			if err != nil {
				return err
			}
			env := newEnv(cmd)
			cfg := config.Resolve(course.Config, config.Config{}, config.Config{})

			opts := pipeline.AuthorOptions{
				CourseName:        course.Name,
				CourseDescription: course.Description,
			}
			// Lessons() errors on an empty course; for authoring that is the
			// normal starting state, not a failure — the first authored lesson
			// simply has no titles to avoid duplicating.
			if lessons, err := course.Lessons(); err == nil {
				for _, l := range lessons {
					opts.ExistingLessons = append(opts.ExistingLessons, l.FrontMatter.Title)
				}
			}

			draft, err := env.AuthorLesson(cmd.Context(), cfg, args[1], opts)
			if err != nil {
				return err
			}

			lessonsDir := filepath.Join(course.Dir, "lessons")
			if err := os.MkdirAll(lessonsDir, 0o755); err != nil {
				return err
			}
			slug := project.Slugify(draft.Title)
			if slug == "" {
				slug = "lesson"
			}
			dir := filepath.Join(lessonsDir, fmt.Sprintf("%02d-%s", nextLessonNumber(lessonsDir), slug))
			if _, err := os.Stat(dir); err == nil {
				return fmt.Errorf("lesson %q already exists", filepath.Base(dir))
			}
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dir, project.LessonFileName), []byte(draft.Markdown()), 0o644); err != nil {
				return err
			}
			// Prove the write loads with the same loader the pipeline uses.
			if _, err := project.LoadLesson(dir); err != nil {
				return fmt.Errorf("drafted lesson does not load (bug): %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s — %q (%d sections)\n",
				filepath.Base(dir), draft.Title, len(draft.Sections))
			return nil
		},
	}
}

// nextLessonNumber returns 1 + the highest NN- prefix in lessonsDir.
func nextLessonNumber(lessonsDir string) int {
	entries, err := os.ReadDir(lessonsDir)
	if err != nil {
		return 1
	}
	max := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if i := strings.IndexByte(e.Name(), '-'); i > 0 {
			if n, err := strconv.Atoi(e.Name()[:i]); err == nil && n > max {
				max = n
			}
		}
	}
	return max + 1
}
