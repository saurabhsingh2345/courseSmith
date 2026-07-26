package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/pipeline"
)

func newBuildSiteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "build-site",
		Short: "Build the static course site (hugo + pagefind when configured)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			env := newEnv(cmd)
			publicDir, err := pipeline.BuildSite(cmd.Context(), env)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "site built → %s\n", publicDir)
			return nil
		},
	}
}

func newBundleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bundle <course>",
		Short: "Build the site and zip it (with videos) for offline file:// use",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			course, err := resolveCourse(args[0])
			if err != nil {
				return err
			}
			env := newEnv(cmd)
			zipPath, err := pipeline.BundleCourse(cmd.Context(), env, course)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "offline bundle: %s — unzip anywhere and open index.html\n", zipPath)
			return nil
		},
	}
}

func newEbookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ebook <course>",
		Short: "Render the companion PDF (transcripts + diagrams + quizzes + answer key)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			course, err := resolveCourse(args[0])
			if err != nil {
				return err
			}
			env := newEnv(cmd)
			pdfPath, err := pipeline.BuildEbook(cmd.Context(), env, course)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "ebook written: %s\n", pdfPath)
			return nil
		},
	}
}
