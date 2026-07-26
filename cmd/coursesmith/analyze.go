package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/pipeline"
)

func newAnalyzeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "analyze <course>",
		Short: "Build the course concept graph and check dependencies, terminology, and narrative",
		Long: "Extracts the concepts every lesson introduces and uses, builds the\n" +
			"course-wide concept DAG (concepts.json + concepts.svg), and errors on\n" +
			"dependency violations — any concept used before it is taught. Also\n" +
			"flags terminology drift and scores the narrative bridge between\n" +
			"consecutive lessons. Outputs land in courses/<slug>/generated/.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			course, err := resolveCourse(args[0])
			if err != nil {
				return err
			}
			env := newEnv(cmd)
			report, err := pipeline.AnalyzeCourse(cmd.Context(), env, course)
			if report != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "\nartifacts: %s\n",
					filepath.Join(course.Dir, pipeline.CourseGeneratedDirName))
			}
			return err
		},
	}
}
