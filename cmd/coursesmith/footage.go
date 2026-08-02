package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/pipeline"
	"github.com/enfec/coursesmith/internal/project"
)

func newFootageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "footage",
		Short: "Manage captured recordings of real tools",
		Long: "Captures are real recordings of real tools — a terminal session, or frames\n" +
			"of somebody else's web product. They are the no-code course's answer to the\n" +
			"python course's executed code blocks: the tool really did that.\n\n" +
			"These products redesign themselves regularly, so a capture ages and nothing\n" +
			"in an ordinary build can tell. `footage list` is what tells.",
	}
	cmd.AddCommand(newFootageLoginCmd())
	cmd.AddCommand(newFootageListCmd())
	cmd.AddCommand(newFootageShootCmd())
	return cmd
}

func newFootageShootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shoot <course>/<lesson>",
		Short: "Record this lesson's desktop captures, with you driving the app",
		Long: "A desktop capture is the one recording with a person in it: Cursor and\n" +
			"Figma have no selectors to drive, so the engine frames the window, records,\n" +
			"crops and times, and you do the work. It shows one beat at a time and stamps\n" +
			"a mark when you press Enter.\n\n" +
			"It is a separate command from `run` because it blocks on a keypress. A batch\n" +
			"build that stopped halfway waiting for somebody is indistinguishable from a\n" +
			"hang, so `run` refuses these and points here instead.\n\n" +
			"macOS only. Needs Screen Recording and Accessibility permission for this\n" +
			"terminal, and the app open with a window.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFootageShoot(cmd, args[0])
		},
	}
}

func runFootageShoot(cmd *cobra.Command, arg string) error {
	courseArg, lessonArg, ok := strings.Cut(arg, "/")
	if !ok || lessonArg == "" {
		return fmt.Errorf("name a lesson, as <course>/<lesson> — a shoot is one lesson's captures, with you at the keyboard")
	}
	course, err := resolveCourse(courseArg)
	if err != nil {
		return err
	}
	lesson, err := course.FindLesson(lessonArg)
	if err != nil {
		return err
	}
	env := newEnv(cmd)
	// The attached console is what makes this command different from `run`.
	env.DesktopInput = cmd.InOrStdin()
	return env.RunLesson(cmd.Context(), course, lesson, pipeline.RunOptions{
		Stage: project.StageDemos,
		Force: true,
	})
}

func newFootageLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login <site>",
		Short: "Sign in to a site once, so captures can run headless afterwards",
		Long: "Opens a real browser window against the site, using the same profile the\n" +
			"capture will use. Sign in, then press Enter here. The session persists, so\n" +
			"every later capture of that site runs headless with no credentials anywhere\n" +
			"near the repository.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pipeline.WebLogin(cmd.Context(), cmd.OutOrStdout(), cmd.InOrStdin(), args[0])
		},
	}
}

func newFootageListCmd() *cobra.Command {
	var staleDays int
	c := &cobra.Command{
		Use:   "list [course]",
		Short: "List captured clips, when they were shot, and which have aged",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg := ""
			if len(args) == 1 {
				arg = args[0]
			}
			return runFootageList(cmd, arg, staleDays)
		},
	}
	c.Flags().IntVar(&staleDays, "stale-after", 90,
		"days after which a capture is flagged for re-shooting")
	return c
}

// runFootageList walks a course's lessons and reports every capture sidecar.
//
// The interesting column is age. A course full of recordings of other people's
// products is wrong the quarter after it ships, and there is no compiler for
// that — this is the compiler for that.
func runFootageList(cmd *cobra.Command, courseArg string, staleDays int) error {
	var courses []*project.Course
	if courseArg != "" {
		c, err := resolveCourse(courseArg)
		if err != nil {
			return err
		}
		courses = append(courses, c)
	} else {
		entries, err := os.ReadDir(coursesDirName)
		if err != nil {
			return fmt.Errorf("reading %s: %w", coursesDirName, err)
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			c, err := project.LoadCourse(filepath.Join(coursesDirName, e.Name()))
			if err != nil {
				continue // not every directory is a course
			}
			courses = append(courses, c)
		}
	}

	out := cmd.OutOrStdout()
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "COURSE\tLESSON\tCLIP\tTOOL\tCAPTURED\tAGE\tSOURCE")
	total, stale := 0, 0
	for _, c := range courses {
		lessons, err := c.Lessons()
		if err != nil {
			continue
		}
		for _, l := range lessons {
			for _, f := range pipeline.LessonFootage(l) {
				total++
				age, aged := footageAge(f, staleDays)
				if aged {
					stale++
				}
				src := f.Origin
				if src == "" {
					src = f.ToolVersion
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
					c.Slug, l.ID, f.ID, f.Tool, shortDate(f.CapturedAt), age, src)
			}
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if total == 0 {
		fmt.Fprintln(out, "\nNo captures yet.")
		return nil
	}
	fmt.Fprintf(out, "\n%d capture(s); %d older than %d days.\n", total, stale, staleDays)
	if stale > 0 {
		fmt.Fprintln(out, "Re-shoot with: coursesmith run <course>/<lesson> --stage demos --force")
	}
	return nil
}

func shortDate(rfc3339 string) string {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return "unknown"
	}
	return t.Format("2006-01-02")
}

// footageAge renders the clip's age and says whether it has passed the window.
func footageAge(f pipeline.Footage, staleDays int) (string, bool) {
	t, err := time.Parse(time.RFC3339, f.CapturedAt)
	if err != nil {
		return "?", false
	}
	days := int(time.Since(t).Hours() / 24)
	label := fmt.Sprintf("%dd", days)
	if days >= staleDays {
		return label + " STALE", true
	}
	return label, false
}
