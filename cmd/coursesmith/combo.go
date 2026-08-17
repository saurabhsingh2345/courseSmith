package main

// The combo surface: one video cut from several templates.
//
// The verbs split into two groups, and the split is the point. `new` and `run`
// make a combo; `segment` edits one after you have watched it. Everything in the
// second group writes combo.yaml and nothing else — the pipeline then re-stales
// exactly the stages that file feeds, so the cost of an edit is a property of
// what you changed rather than of which command you typed.
//
// Every edit this command can make moves the words, and that is worth stating
// plainly rather than implying otherwise. Swapping a template re-plans the
// segment through a different prompt and gets different beats; skipping one
// takes its narration out of the read. So all of them re-run the voice track
// and every timing after it, and `segment` says so before you find out.
//
// A genuinely cheap edit — fixing a wrong label, nudging a value, leaving the
// narration untouched — is a different operation at a different level: it
// patches the generated scene's props rather than the request. That is what
// video-plan.yaml already does for lessons, and it is the shape the studio's
// editor should take. It is deliberately not here, because a flag that looked
// like the others and cost a hundredth as much would teach the wrong model of
// what this tool does.

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/pipeline"
	"github.com/enfec/coursesmith/internal/project"
)

func newComboCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "combo",
		// `reel` was this surface's name until the director landed. Kept as an
		// alias rather than removed, because the rename is not a reason for
		// somebody's script or muscle memory to start erroring — and it costs one
		// line. It is undocumented on purpose: the help should teach the current
		// name, not preserve the old one.
		Aliases: []string{"reel"},
		Short:   "Build a longer video from several templates on one timeline",
		Long: "A combo is an ordered run of segments, each with its own visual template,\n" +
			"narrated as one continuous read and rendered onto one timeline.\n\n" +
			"Use a snippet for one idea in one look. Use a combo when the piece is long\n" +
			"enough that one look would not hold it.\n\n" +
			"`coursesmith combo direct \"<subject>\"` is the whole surface in one command:\n" +
			"it decides what the piece argues, how it divides, which look carries each\n" +
			"part and how long each runs.",
	}
	cmd.AddCommand(newComboDirectCmd())
	cmd.AddCommand(newComboNewCmd())
	cmd.AddCommand(newComboRunCmd())
	cmd.AddCommand(newComboListCmd())
	cmd.AddCommand(newComboShowCmd())
	cmd.AddCommand(newComboSegmentCmd())
	return cmd
}

// parseSegmentFlag reads a --segment value: "template:what it should cover".
//
// One flag repeated rather than a pair of parallel lists, because parallel
// lists silently mis-pair the moment one has an entry the other does not, and
// the failure shows up as a segment covering the wrong thing rather than as an
// error.
func parseSegmentFlag(v string) (pipeline.ComboSegment, error) {
	name, prompt, ok := strings.Cut(v, ":")
	if !ok {
		return pipeline.ComboSegment{}, fmt.Errorf("segment %q is not template:prompt — try --segment 'gauge:which models fit in 24GB'", v)
	}
	name = strings.TrimSpace(name)
	prompt = strings.TrimSpace(prompt)
	if name == "" || prompt == "" {
		return pipeline.ComboSegment{}, fmt.Errorf("segment %q needs both a template and a prompt", v)
	}
	return pipeline.ComboSegment{Template: name, Prompt: prompt}, nil
}

func newComboNewCmd() *cobra.Command {
	var (
		segments    []string
		id          string
		title       string
		brief       string
		voice       string
		model       string
		captions    string
		mode        string
		skin        string
		planOnly    bool
		concurrency int
	)
	cmd := &cobra.Command{
		Use:   "new",
		Short: "Create a combo from a list of segments and render it",
		Long: "Each --segment is 'template:what it should cover', in order.\n\n" +
			"  coursesmith combo new --title \"What decides whether a model runs\" \\\n" +
			"    --segment 'rundown:the three numbers that decide it' \\\n" +
			"    --segment 'gauge:which models fit in 24GB' \\\n" +
			"    --segment 'verdict:what to actually buy'\n\n" +
			"See `coursesmith snippet templates` for the catalog, grouped by what\n" +
			"each group is for.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			spec := pipeline.ComboSpec{
				ID:        id,
				Title:     title,
				Brief:     brief,
				CreatedAt: time.Now().UTC(),
			}
			for _, s := range segments {
				seg, err := parseSegmentFlag(s)
				if err != nil {
					return err
				}
				spec.Segments = append(spec.Segments, seg)
			}
			if voice != "" {
				spec.Config.Style.Voice = voice
			}
			if model != "" {
				spec.Config.Pipeline.LLMContent = model
			}
			if captions != "" {
				spec.Config.Style.Captions = captions
			}
			if mode != "" {
				spec.Config.Style.Mode = mode
			}
			if skin != "" {
				spec.Config.Style.Skin = skin
			}

			course, lesson, err := pipeline.CreateCombo(".", spec)
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
			if err := env.RunCombo(ctx, course, lesson, opts); err != nil {
				return err
			}
			if !planOnly {
				fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n",
					relOrAbs(filepath.Join(lesson.GeneratedDir(), pipeline.FinalVideoName)))
			}
			return nil
		},
	}
	cmd.Flags().StringArrayVar(&segments, "segment", nil, "a segment, as template:prompt (repeat, in order)")
	cmd.Flags().StringVar(&id, "id", "", "combo id (default: derived from the title)")
	cmd.Flags().StringVar(&title, "title", "", "the finished piece's title")
	cmd.Flags().StringVar(&brief, "brief", "", "what the whole piece is about, in your words")
	cmd.Flags().StringVar(&voice, "voice", "", "TTS voice id (default: the combos course voice)")
	cmd.Flags().StringVar(&model, "model", "", "planning model as provider/model (default: the course's llm_content)")
	cmd.Flags().StringVar(&captions, "captions", "", "burn the caption track in: on | off")
	cmd.Flags().StringVar(&mode, "mode", "", "light or dark video (default dark)")
	cmd.Flags().StringVar(&skin, "skin", "", "house style: default | broadcast | minimal | editorial | showroom")
	cmd.Flags().BoolVar(&planOnly, "plan-only", false, "stop after planning; do not synthesize or render")
	cmd.Flags().IntVar(&concurrency, "concurrency", 0, "parallel browser tabs for the Remotion render (0 = auto)")
	_ = cmd.MarkFlagRequired("segment")
	return cmd
}

func newComboRunCmd() *cobra.Command {
	var (
		stage       string
		force       bool
		concurrency int
	)
	cmd := &cobra.Command{
		Use:   "run <id>",
		Short: "Re-run an existing combo (up-to-date stages are skipped)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			course, lesson, err := pipeline.FindCombo(".", args[0])
			if err != nil {
				return err
			}
			env := newEnv(cmd)
			if r, ok := env.Renderer.(*pipeline.RemotionRenderer); ok {
				r.Concurrency = concurrency
			}
			return env.RunCombo(ctx, course, lesson, pipeline.RunOptions{Stage: stage, Force: force})
		},
	}
	// Generated from the real stage list rather than spelled out. The hand-written
	// version was already stale the moment `substance` was inserted ahead of
	// `plan`, and a help string that omits a stage is a stage nobody runs directly.
	cmd.Flags().StringVar(&stage, "stage", "", "run only this stage ("+strings.Join(project.SnippetStageOrder, ", ")+")")
	cmd.Flags().BoolVar(&force, "force", false, "re-run stages even if their inputs are unchanged")
	cmd.Flags().IntVar(&concurrency, "concurrency", 0, "parallel browser tabs for the Remotion render (0 = auto)")
	return cmd
}

func newComboListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the combos in this project",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			combos, err := pipeline.ListCombos(".")
			if err != nil {
				return err
			}
			if len(combos) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no combos yet — try `coursesmith combo new --help`")
				return nil
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tSEGMENTS\tVIDEO\tTITLE")
			for _, l := range combos {
				spec, err := pipeline.LoadComboSpec(l.Dir)
				if err != nil {
					continue
				}
				video := "—"
				if _, err := os.Stat(filepath.Join(l.GeneratedDir(), pipeline.FinalVideoName)); err == nil {
					video = "ready"
				}
				skipped := len(spec.Segments) - len(spec.Active())
				count := fmt.Sprintf("%d", len(spec.Active()))
				if skipped > 0 {
					count += fmt.Sprintf(" (+%d skipped)", skipped)
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", l.ID, count, video, spec.Title)
			}
			return w.Flush()
		},
	}
}

func newComboShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a combo's segments, in order",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, lesson, err := pipeline.FindCombo(".", args[0])
			if err != nil {
				return err
			}
			spec, err := pipeline.LoadComboSpec(lesson.Dir)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if spec.Title != "" {
				fmt.Fprintf(out, "%s\n", spec.Title)
			}
			fmt.Fprintf(out, "%s\n\n", relOrAbs(lesson.Dir))

			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "  #\tID\tTEMPLATE\tCOVERS")
			for i, s := range spec.Segments {
				mark := " "
				if s.Skip {
					mark = "-" // out of the cut, still in the file
				}
				fmt.Fprintf(w, "%s %d\t%s\t%s\t%s\n", mark, i+1, s.ID, s.Template, truncate(s.Prompt, 52))
			}
			if err := w.Flush(); err != nil {
				return err
			}
			fmt.Fprintf(out, "\nEdit with `coursesmith combo segment %s <segment-id> --…`, then re-run.\n", spec.ID)
			return nil
		},
	}
}

// newComboSegmentCmd is the customisation surface: change one segment of an
// already-generated combo and re-run only what that actually invalidates.
func newComboSegmentCmd() *cobra.Command {
	var (
		template string
		prompt   string
		seconds  int
		skip     bool
		unskip   bool
	)
	cmd := &cobra.Command{
		Use:   "segment <combo-id> <segment-id>",
		Short: "Change one segment of an existing combo",
		Long: "Edits combo.yaml in place. The pipeline then re-stales exactly the\n" +
			"stages that file feeds.\n\n" +
			"Every edit here moves the words: swapping a template re-plans the segment\n" +
			"and gets different beats, and skipping one takes its narration out of the\n" +
			"read. So all of them re-run the plan, the voice track and every timing\n" +
			"after it — expect minutes, not seconds.\n\n" +
			"Nothing is re-run by this command; it reports what became stale and leaves\n" +
			"the run to you.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, lesson, err := pipeline.FindCombo(".", args[0])
			if err != nil {
				return err
			}
			spec, err := pipeline.LoadComboSpec(lesson.Dir)
			if err != nil {
				return err
			}
			if skip && unskip {
				return fmt.Errorf("--skip and --unskip contradict each other")
			}

			idx := -1
			for i, s := range spec.Segments {
				if s.ID == args[1] {
					idx = i
					break
				}
			}
			if idx < 0 {
				ids := make([]string, 0, len(spec.Segments))
				for _, s := range spec.Segments {
					ids = append(ids, s.ID)
				}
				return fmt.Errorf("combo %q has no segment %q (segments: %s)",
					args[0], args[1], strings.Join(ids, ", "))
			}

			// Whether the words change is the whole cost model, so it is decided
			// explicitly rather than inferred later from which fields differ.
			narrationChanged := false
			seg := &spec.Segments[idx]
			if template != "" {
				if _, ok := pipeline.SnippetTemplates[template]; !ok {
					return fmt.Errorf("unknown template %q (see `coursesmith snippet templates`)", template)
				}
				seg.Template = template
				// A different template plans different beats from the same
				// prompt, so this does move the words after all.
				narrationChanged = true
			}
			if prompt != "" {
				seg.Prompt = prompt
				narrationChanged = true
			}
			if seconds != 0 {
				seg.TargetSec = seconds
				narrationChanged = true
			}
			if skip {
				seg.Skip = true
				narrationChanged = true
			}
			if unskip {
				seg.Skip = false
				narrationChanged = true
			}
			if err := spec.Validate(); err != nil {
				return err
			}
			if err := pipeline.SaveComboSpec(lesson.Dir, spec); err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "updated segment %q (%s)\n", seg.ID, seg.Template)
			// Honest about the cost. Every edit this command can make alters what
			// is spoken — a template swap re-plans, and skipping removes a
			// segment's words from the read — so all of them re-run the voice.
			// The cheap path exists at the props level and is the studio's job.
			if narrationChanged {
				fmt.Fprintf(out, "\nThis changes what is said, so the plan, the voice track and every\n"+
					"timing after it are now stale. Re-run with:\n\n  coursesmith combo run %s\n", spec.ID)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&template, "template", "", "render this segment with a different template")
	cmd.Flags().StringVar(&prompt, "prompt", "", "change what this segment covers")
	cmd.Flags().IntVar(&seconds, "seconds", 0, "change this segment's target runtime")
	cmd.Flags().BoolVar(&skip, "skip", false, "drop this segment from the cut (keeps it in the file)")
	cmd.Flags().BoolVar(&unskip, "unskip", false, "put a skipped segment back in the cut")
	return cmd
}

// truncate keeps a prompt on one line of a table.
func truncate(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

// newComboDirectCmd is the one-input path: a subject, and the machine decides
// everything else.
//
// It writes combo.yaml and stops by default, because the structure is worth
// reading before a dozen planning calls are spent on it — and because the file
// it writes is the same one a person would have written, so overriding a
// decision is editing a line rather than starting again.
func newComboDirectCmd() *cobra.Command {
	var (
		title       string
		minutes     int
		run         bool
		skin        string
		captions    string
		mode        string
		concurrency int
	)
	cmd := &cobra.Command{
		Use:   "direct <subject>",
		Short: "Direct a whole piece from one line: outline it, cast the looks, write it",
		Long: "Directing reads the subject and decides what the piece argues, how it\n" +
			"divides into parts, which template carries each one and how long each runs.\n" +
			"It writes combo.yaml and stops, so you can read the structure — and change\n" +
			"any of it — before paying for the plan.\n\n" +
			"  coursesmith combo direct \"What is artificial intelligence?\" --minutes 6\n" +
			"  coursesmith combo show <id>          # read what it decided\n" +
			"  coursesmith combo segment <id> …     # override a pick\n" +
			"  coursesmith combo run <id>           # then build it\n\n" +
			"The theme decides which templates can be cast: default and minimal draw\n" +
			"from the core catalog, broadcast adds the replica batch, editorial adds\n" +
			"the foundations batch. A piece is cut in one house style throughout.\n\n" +
			"Pass --run to go straight through.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			env := newEnv(cmd)
			course, err := pipeline.EnsureCombosCourse(".")
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()

			// Layered over the defaults, not the bare course manifest: the combos
			// course records only what it overrides, so course.Config alone has
			// no model configured and directing would fail before it began.
			cfg := config.Resolve(course.Config, config.Config{}, config.Config{})
			result, err := pipeline.DirectCombo(ctx, env, pipeline.ComboRequest{
				Subject:  args[0],
				Title:    title,
				Minutes:  minutes,
				Skin:     skin,
				Captions: captions,
				Mode:     mode,
			}, cfg)
			if err != nil {
				return err
			}
			spec := result.Spec

			_, lesson, err := pipeline.CreateCombo(".", *spec)
			if err != nil {
				return err
			}

			// The ANGLE first, because it is the decision everything else serves
			// and the one most worth disagreeing with. A piece can be about the
			// right subject and be making the wrong point, and this is the only
			// place that is visible before the render.
			fmt.Fprintf(out, "\n%s\n", spec.Title)
			fmt.Fprintf(out, "  the argument: %s\n\n", spec.Angle)

			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			// ROLE alongside the template, because the arc is a decision worth
			// reading and it used to be invisible: an enforced shape that leaves no
			// trace on the page cannot be corrected by the person looking at it.
			fmt.Fprintln(w, "  #\tROLE\tTEMPLATE\tESTABLISHES")
			for i, seg := range spec.Segments {
				fmt.Fprintf(w, "  %d\t%s\t%s\t%s\n", i+1, seg.Role, seg.Template, truncate(seg.Prompt, 46))
				// The material on its own line under the segment it belongs to.
				//
				// This command stops before planning so the structure can be read
				// and corrected, and the material is the part actually worth
				// reading: the template names the look, but these are the facts the
				// piece will state as true. A wrong figure spotted here costs one
				// edit; the same figure spotted in the rendered video costs the
				// render.
				if m := strings.TrimSpace(seg.Material); m != "" {
					fmt.Fprintf(w, "  \t\t\t%s\n", truncate(m, 46))
				}
			}
			if err := w.Flush(); err != nil {
				return err
			}
			fmt.Fprintf(out, "\n%s\n", relOrAbs(lesson.Dir))

			if !run {
				// lesson.ID, not spec.ID: CreateCombo takes the spec by value and
				// derives the id on its own copy, so the caller's is still empty.
				fmt.Fprintf(out, "\nRead it, change anything, then build:\n\n  coursesmith combo run %s\n", lesson.ID)
				return nil
			}
			if r, ok := env.Renderer.(*pipeline.RemotionRenderer); ok {
				r.Concurrency = concurrency
			}
			if err := env.RunCombo(ctx, course, lesson, pipeline.RunOptions{}); err != nil {
				return err
			}
			fmt.Fprintf(out, "\n%s\n", relOrAbs(filepath.Join(lesson.GeneratedDir(), pipeline.FinalVideoName)))
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "the finished piece's title (default: the director writes one)")
	cmd.Flags().IntVar(&minutes, "minutes", 0, "how long the piece should run (0 = read it out of the subject, else the default)")
	cmd.Flags().BoolVar(&run, "run", false, "build it immediately instead of stopping at the structure")
	cmd.Flags().StringVar(&skin, "skin", "", "theme, which also decides the template pool: default | broadcast | minimal | editorial | showroom")
	cmd.Flags().StringVar(&captions, "captions", "", "burn the caption track in: on | off")
	cmd.Flags().StringVar(&mode, "mode", "", "light or dark video (default dark)")
	cmd.Flags().IntVar(&concurrency, "concurrency", 0, "parallel browser tabs for the Remotion render (0 = auto)")
	return cmd
}
