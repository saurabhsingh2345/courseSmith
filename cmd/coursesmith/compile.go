package main

import (
	"github.com/spf13/cobra"
)

func newCompileCourseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "compile-course <course>",
		Short: "Join every rendered lesson video into one course video + chapter file",
		Long: "Concatenates each lesson's final.mp4, in lesson order, into\n" +
			"courses/<slug>/course.mp4 (lossless stream copy — the lessons share\n" +
			"one render pipeline) and writes course-chapters.txt in YouTube's\n" +
			"chapter format. Lessons that haven't rendered yet are skipped with\n" +
			"a warning.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			course, err := resolveCourse(args[0])
			if err != nil {
				return err
			}
			env := newEnv(cmd)
			return env.CompileCourse(cmd.Context(), course)
		},
	}
}
