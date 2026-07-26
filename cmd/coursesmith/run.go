package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/pipeline"
)

// promptsDirName is where prompt templates live, relative to the project root.
const promptsDirName = "prompts"

func newRunCmd() *cobra.Command {
	var stage string
	var force bool
	var concurrency int
	cmd := &cobra.Command{
		Use:   "run <course>[/<lesson>]",
		Short: "Run the generation pipeline (resumable; up-to-date stages are skipped)",
		Long: "Runs the pipeline for one lesson (python-basics/01) or every lesson in a\n" +
			"course (python-basics). Stages whose inputs are unchanged are skipped, so\n" +
			"re-running is cheap and an interrupted run resumes where it stopped.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			courseArg, lessonArg, _ := strings.Cut(args[0], "/")
			course, err := resolveCourse(courseArg)
			if err != nil {
				return err
			}
			env := newEnv(cmd)
			if r, ok := env.Renderer.(*pipeline.RemotionRenderer); ok {
				r.Concurrency = concurrency
			}
			opts := pipeline.RunOptions{Stage: stage, Force: force}

			if lessonArg != "" {
				lesson, err := course.FindLesson(lessonArg)
				if err != nil {
					return err
				}
				return env.RunLesson(ctx, course, lesson, opts)
			}
			return env.RunCourse(ctx, course, opts)
		},
	}
	cmd.Flags().StringVar(&stage, "stage", "", "run only this stage (script, verify, review, visuals, quiz, mistakes, exercises, demos, audio, align, captions, chapters, scenegraph, render, hugo)")
	cmd.Flags().BoolVar(&force, "force", false, "re-run stages even if their inputs are unchanged")
	cmd.Flags().IntVar(&concurrency, "concurrency", 0, "parallel browser tabs for the Remotion render (0 = auto)")
	return cmd
}
