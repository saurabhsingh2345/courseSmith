package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/pipeline"
)

func newExportReviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "export-review <course>",
		Short: "Export one markdown review document per lesson for a human SME",
		Long: "Writes courses/<slug>/review-export/<lesson>.md for every lesson:\n" +
			"script + diagrams (inline) + quiz + common mistakes + exercises +\n" +
			"the flags from every automated pass. The reviewer answers in\n" +
			"review-notes.yaml; the next pipeline run applies the notes.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			course, err := resolveCourse(args[0])
			if err != nil {
				return err
			}
			env := newEnv(cmd)
			outDir, err := pipeline.ExportReview(env, course)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nreview documents in %s — send them to your SME\n", outDir)
			return nil
		},
	}
}
