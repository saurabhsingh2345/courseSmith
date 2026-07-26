package main

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/pipeline"
)

func newAuditionCmd() *cobra.Command {
	var choose string
	cmd := &cobra.Command{
		Use:   "audition <course>",
		Short: "Render a sample paragraph in every matching Kokoro voice, or pick one",
		Long: "Renders one fixed paragraph in every Kokoro voice matching the course\n" +
			"language into courses/<slug>/auditions/, with an index.html page for\n" +
			"side-by-side listening. When you've picked a favourite:\n\n" +
			"  coursesmith audition <course> --choose af_bella\n\n" +
			"writes it to course.yaml (style.voice).",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			course, err := resolveCourse(args[0])
			if err != nil {
				return err
			}
			env := newEnv(cmd)
			cfg := config.Resolve(course.Config, config.Config{}, config.Config{})

			if choose != "" {
				if voices, err := env.ListVoices(cmd.Context()); err == nil && !slices.Contains(voices, choose) {
					return fmt.Errorf("voice %q is not on the Kokoro server (%d voices available — run audition without --choose to hear them)", choose, len(voices))
				}
				if err := pipeline.ChooseVoice(course, choose); err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "course.yaml updated: style.voice = %s\n", choose)
				fmt.Fprintln(cmd.OutOrStdout(), "re-run the pipeline to synthesize lessons with the new voice")
				return nil
			}

			index, err := pipeline.RunAudition(cmd.Context(), env, course, cfg)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nopen %s to compare voices side by side\n", index)
			return nil
		},
	}
	cmd.Flags().StringVar(&choose, "choose", "", "write this voice id to course.yaml instead of rendering auditions")
	return cmd
}
