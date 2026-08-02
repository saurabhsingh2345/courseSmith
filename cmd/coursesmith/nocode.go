package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/pipeline"
)

func newNoCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nocode",
		Short: "Build a piece where every segment is backed by a real recording",
		Long: "A no-code piece is several segments cut onto one timeline, like a reel —\n" +
			"with one extra rule that is the whole point: every segment must stand on\n" +
			"evidence that exists on disk. A real capture, or facts written out.\n\n" +
			"A segment that can name neither is refused before a single planning call is\n" +
			"spent, and nothing is ever recast onto a drawn figure to fill the gap. That\n" +
			"is what separates this from a reel about no-code tools.",
	}
	cmd.AddCommand(newNoCodeNewCmd())
	cmd.AddCommand(newNoCodeListCmd())
	cmd.AddCommand(newNoCodeRunCmd())
	cmd.AddCommand(newNoCodeTemplatesCmd())
	return cmd
}

func newNoCodeNewCmd() *cobra.Command {
	var title, brief string
	c := &cobra.Command{
		Use:   "new",
		Short: "Scaffold a piece, ready for you to fill in its segments",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(title) == "" {
				return fmt.Errorf("a piece needs a --title")
			}
			// A scaffold with one worked segment of each evidence kind, because
			// the shape of the evidence block is the one thing somebody opening
			// this file for the first time will not guess.
			spec := pipeline.NoCodeSpec{
				Title: title,
				Brief: brief,
				Segments: []pipeline.NoCodeSegment{
					{
						Template: "footage",
						Prompt:   "what the recording shows, in your own words",
						Evidence: pipeline.NoCodeEvidence{
							Kind: pipeline.EvidenceCapture,
							Capture: &pipeline.NoCodeCapture{
								Tool: "claude",
								Of:   "ask the agent to add a small feature, and show the files it changed",
							},
						},
					},
					{
						Template: "verdict",
						Prompt:   "what you conclude from it",
						Evidence: pipeline.NoCodeEvidence{
							Kind:  pipeline.EvidenceFact,
							Facts: []string{"a claim you can point at a source for"},
						},
					},
				},
			}
			_, l, err := pipeline.NewNoCodePiece(".", spec)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Created %s\n\n", l.Dir)
			fmt.Fprintf(out, "Edit %s/%s — every segment needs evidence:\n", l.Dir, pipeline.NoCodeFileName)
			fmt.Fprintf(out, "  a capture   evidence: {kind: capture, capture: {tool: claude, of: what to record}}\n")
			fmt.Fprintf(out, "  facts       evidence: {kind: fact, facts: [\"...\"]}\n\n")
			fmt.Fprintf(out, "Recordable: %s\n\n", strings.Join(pipeline.CaptureToolNames(), ", "))
			fmt.Fprintf(out, "Then: coursesmith nocode run %s\n", l.ID)
			return nil
		},
	}
	c.Flags().StringVar(&title, "title", "", "the piece's title")
	c.Flags().StringVar(&brief, "brief", "", "what the whole piece is about, in your words")
	return c
}

func newNoCodeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List every no-code piece",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pieces, err := pipeline.ListNoCodePieces(".")
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if len(pieces) == 0 {
				fmt.Fprintln(out, "No pieces yet. Start one: coursesmith nocode new --title \"...\"")
				return nil
			}
			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tSEGMENTS\tEVIDENCE\tTITLE")
			for _, l := range pieces {
				spec, err := pipeline.LoadNoCodeSpec(l.Dir)
				if err != nil {
					continue
				}
				live := spec.Live()
				fmt.Fprintf(w, "%s\t%d\t%d capture(s)\t%s\n",
					l.ID, len(live), len(spec.CaptureIDs()), spec.Title)
			}
			return w.Flush()
		},
	}
}

func newNoCodeRunCmd() *cobra.Command {
	var stage string
	var force bool
	c := &cobra.Command{
		Use:   "run <id>",
		Short: "Record, plan and render a piece",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			course, lesson, err := pipeline.FindNoCodePiece(".", args[0])
			if err != nil {
				return err
			}
			// Validate before anything is spent. The whole argument for this
			// surface is that a hollow segment fails immediately rather than
			// after the planning calls.
			spec, err := pipeline.LoadNoCodeSpec(lesson.Dir)
			if err != nil {
				return err
			}
			if err := spec.Validate(); err != nil {
				return err
			}
			env := newEnv(cmd)
			return env.RunNoCode(ctx, course, lesson, pipeline.RunOptions{Stage: stage, Force: force})
		},
	}
	c.Flags().StringVar(&stage, "stage", "", "run only this stage")
	c.Flags().BoolVar(&force, "force", false, "re-run stages even if their inputs are unchanged")
	return c
}

func newNoCodeTemplatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "templates",
		Short: "List the templates a no-code segment may use",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			for _, name := range pipeline.NoCodeTemplateNames() {
				tpl := pipeline.SnippetTemplates[name]
				fmt.Fprintf(w, "%s\t%s\n", name, tpl.Title)
			}
			if err := w.Flush(); err != nil {
				return err
			}
			fmt.Fprintln(out, "\nExcluded here: cast, story, illustration — they put a drawn figure")
			fmt.Fprintln(out, "on screen, which is the fastest way to fill a frame with no evidence")
			fmt.Fprintln(out, "behind it. Use them in a reel, where that is a fair trade.")
			return nil
		},
	}
}
