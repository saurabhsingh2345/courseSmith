package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/pipeline"
)

// newSnippetCmd is the short-form product surface: one prompt plus one visual
// template becomes a finished clip.
//
//	coursesmith snippet templates
//	coursesmith snippet new --template vscode "how for loops work in python"
//	coursesmith snippet run for-loops-in-python
//	coursesmith snippet list
func newSnippetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snippet",
		Short: "Create and render short standalone clips from a prompt and a template",
		Long: "A snippet is a short video built from one prompt and one visual template —\n" +
			"no lesson to write, no course to belong to. `snippet new` plans and renders\n" +
			"one end to end; `snippet templates` lists what it can look like.",
	}
	cmd.AddCommand(newSnippetTemplatesCmd())
	cmd.AddCommand(newSnippetNewCmd())
	cmd.AddCommand(newSnippetRunCmd())
	cmd.AddCommand(newSnippetListCmd())
	return cmd
}

func newSnippetTemplatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "templates",
		Short: "List the visual templates a snippet can use",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Grouped rather than one alphabetical table. Past twenty entries a
			// flat A-to-Z list stops being a catalog and becomes a wall — and
			// the thing somebody has when they run this is a job to do, not a
			// template name, so the headings are what they can actually scan.
			out := cmd.OutOrStdout()
			for _, g := range pipeline.SnippetTemplatesByCategory() {
				fmt.Fprintf(out, "\n%s\n%s\n\n", strings.ToUpper(g.Title), g.Blurb)
				w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
				for _, t := range g.Templates {
					name := t.Name
					if t.Since != "" {
						name += " (" + t.Since + ")"
					}
					fmt.Fprintf(w, "  %s\t%s\t%s\n", name, t.Title, t.Description)
				}
				if err := w.Flush(); err != nil {
					return err
				}
			}
			// Counted off the list actually printed above, not off the whole
			// registry — a shelved template is not on offer, and a footer
			// claiming twenty-nine over a table of twenty-seven is the kind of
			// small lie that makes somebody go looking for the missing two.
			fmt.Fprintf(out, "\n%d templates in %d groups. Start one with:\n\n",
				len(pipeline.SnippetTemplateList()), len(pipeline.SnippetTemplatesByCategory()))
			fmt.Fprintf(out, "  coursesmith snippet new --template <name> \"what it should teach\"\n\n")
			return nil
		},
	}
}

func newSnippetNewCmd() *cobra.Command {
	var (
		template    string
		title       string
		targetSec   int
		codeLang    string
		voice       string
		model       string
		captions    string
		mode        string
		planOnly    bool
		concurrency int
	)
	cmd := &cobra.Command{
		Use:   "new <prompt>",
		Short: "Create a snippet from a prompt and render it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			spec := pipeline.SnippetSpec{
				Prompt:       args[0],
				Template:     template,
				Title:        title,
				TargetSec:    targetSec,
				CodeLanguage: codeLang,
			}
			if voice != "" {
				spec.Config.Style.Voice = voice
			}
			if model != "" {
				spec.Config.Pipeline.LLMContent = model
			}
			// Both ride the ordinary per-snippet config override, so they layer
			// over the snippets course the same way the voice does.
			if captions != "" {
				spec.Config.Style.Captions = captions
			}
			if mode != "" {
				spec.Config.Style.Mode = mode
			}
			course, lesson, err := pipeline.CreateSnippet(".", spec)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "created %s\n", relOrAbs(lesson.Dir))

			env := newEnv(cmd)
			if r, ok := env.Renderer.(*pipeline.RemotionRenderer); ok {
				r.Concurrency = concurrency
			}
			opts := pipeline.RunOptions{}
			if planOnly {
				opts.Stage = "plan"
			}
			if err := env.RunSnippet(ctx, course, lesson, opts); err != nil {
				return err
			}
			if !planOnly {
				fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n",
					relOrAbs(filepath.Join(lesson.GeneratedDir(), pipeline.FinalVideoName)))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&template, "template", "", "visual template (see `coursesmith snippet templates`)")
	cmd.Flags().StringVar(&title, "title", "", "override the title the model would write")
	cmd.Flags().IntVar(&targetSec, "seconds", 0, "approximate runtime to aim for, in seconds (default 45)")
	cmd.Flags().StringVar(&codeLang, "code-language", "", "programming language for code templates (default python)")
	cmd.Flags().StringVar(&voice, "voice", "", "TTS voice id (default: the snippets course voice)")
	cmd.Flags().StringVar(&model, "model", "", "planning model as provider/model, e.g. openai/gpt-4o-mini (default: the course's llm_content)")
	cmd.Flags().StringVar(&captions, "captions", "", "burn the caption track into the video: on | off (default: the snippets course setting)")
	cmd.Flags().StringVar(&mode, "mode", "", "light or dark video (default dark)")
	cmd.Flags().BoolVar(&planOnly, "plan-only", false, "stop after planning; do not synthesize or render")
	cmd.Flags().IntVar(&concurrency, "concurrency", 0, "parallel browser tabs for the Remotion render (0 = auto)")
	_ = cmd.MarkFlagRequired("template")
	return cmd
}

func newSnippetRunCmd() *cobra.Command {
	var (
		stage       string
		force       bool
		concurrency int
	)
	cmd := &cobra.Command{
		Use:   "run <id>",
		Short: "Re-run an existing snippet (up-to-date stages are skipped)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			course, lesson, err := pipeline.FindSnippet(".", args[0])
			if err != nil {
				return err
			}
			env := newEnv(cmd)
			if r, ok := env.Renderer.(*pipeline.RemotionRenderer); ok {
				r.Concurrency = concurrency
			}
			return env.RunSnippet(ctx, course, lesson, pipeline.RunOptions{Stage: stage, Force: force})
		},
	}
	cmd.Flags().StringVar(&stage, "stage", "", "run only this stage (plan, verify, audio, align, captions, chapters, scenegraph, render)")
	cmd.Flags().BoolVar(&force, "force", false, "re-run stages even if their inputs are unchanged")
	cmd.Flags().IntVar(&concurrency, "concurrency", 0, "parallel browser tabs for the Remotion render (0 = auto)")
	return cmd
}

func newSnippetListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the snippets in this project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			snippets, err := pipeline.ListSnippets(".")
			if err != nil {
				return err
			}
			if len(snippets) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no snippets yet — try `coursesmith snippet templates`")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTEMPLATE\tVIDEO\tTITLE")
			for _, l := range snippets {
				spec, err := pipeline.LoadSnippetSpec(l.Dir)
				if err != nil {
					continue
				}
				video := "—"
				if _, err := os.Stat(filepath.Join(l.GeneratedDir(), pipeline.FinalVideoName)); err == nil {
					video = "ready"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", l.ID, spec.Template, video, l.FrontMatter.Title)
			}
			return w.Flush()
		},
	}
}

// relOrAbs shortens a path against the working directory when it is below it.
func relOrAbs(path string) string {
	wd, err := os.Getwd()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(wd, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	return rel
}
