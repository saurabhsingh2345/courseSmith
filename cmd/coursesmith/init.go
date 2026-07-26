package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/project"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <course-name>",
		Short: "Scaffold a new course: course.yaml plus an example lesson",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args[0])
		},
	}
}

func runInit(cmd *cobra.Command, name string) error {
	res, err := project.ScaffoldCourse(coursesDirName, name, "")
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Created course %q:\n", res.Slug)
	fmt.Fprintf(out, "  %s\n", res.CourseFile)
	fmt.Fprintf(out, "  %s\n", res.LessonFile)
	fmt.Fprintf(out, "\nNext steps:\n")
	fmt.Fprintf(out, "  1. Edit %s with your course details\n", res.CourseFile)
	fmt.Fprintf(out, "  2. Outline your first lesson in %s\n", res.LessonFile)
	fmt.Fprintf(out, "  3. Run: coursesmith status %s\n", res.Slug)
	return nil
}
