package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/pipeline"
	"github.com/enfec/coursesmith/internal/project"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status <course>",
		Short: "Show each lesson's pipeline progress (done / stale / pending)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, args[0])
		},
	}
}

func runStatus(cmd *cobra.Command, courseArg string) error {
	course, err := resolveCourse(courseArg)
	if err != nil {
		return err
	}
	lessons, err := course.Lessons()
	if err != nil {
		return err
	}

	env := &pipeline.Env{PromptsDir: promptsDirName} // no router: status never calls an LLM
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "%s (%s) — %d lesson(s)\n\n", course.Name, course.Slug, len(lessons))

	// A video-only course reports only the stages it actually runs.
	courseCfg := config.Resolve(course.Config, config.Config{}, config.Config{})
	stages := project.StagesFor(courseCfg.Pipeline.VideoOnly)

	w := tabwriter.NewWriter(out, 2, 4, 2, ' ', 0)
	fmt.Fprint(w, "LESSON")
	for _, stage := range stages {
		fmt.Fprintf(w, "\t%s", stage)
	}
	fmt.Fprintln(w)

	for _, l := range lessons {
		cfg := config.Resolve(course.Config, l.FrontMatter.Overrides(), config.Config{})
		statuses, err := env.LessonStatus(l, cfg)
		if err != nil {
			return fmt.Errorf("lesson %s: %w", l.ID, err)
		}
		fmt.Fprint(w, l.ID)
		for _, stage := range stages {
			fmt.Fprintf(w, "\t%s", statusGlyph(statuses[stage]))
		}
		fmt.Fprintln(w)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	fmt.Fprintln(out, "\n✓ done   ~ stale (inputs changed)   · pending")
	return nil
}

func statusGlyph(s project.StageStatus) string {
	switch s {
	case project.StatusDone:
		return "✓ done"
	case project.StatusStale:
		return "~ stale"
	default:
		return "· pending"
	}
}
