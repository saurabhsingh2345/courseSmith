package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/enfec/coursesmith/internal/pipeline"
)

func newPreviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "preview <course>/<lesson>",
		Short: "Open the lesson in Remotion Studio for hot-reload visual editing",
		Long: "Stages the lesson's scene graph and assets into the renderer and opens\n" +
			"Remotion Studio. Requires the pipeline to have run at least through the\n" +
			"scenegraph stage for this lesson.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			courseArg, lessonArg, ok := strings.Cut(args[0], "/")
			if !ok {
				return fmt.Errorf("preview needs <course>/<lesson>, e.g. python-basics/01")
			}
			course, err := resolveCourse(courseArg)
			if err != nil {
				return err
			}
			lesson, err := course.FindLesson(lessonArg)
			if err != nil {
				return err
			}
			renderer, _ := resolveRenderer().(*pipeline.RemotionRenderer)
			if renderer == nil {
				return fmt.Errorf("preview needs Node 18+ and the renderer project — install node, then: cd %s && npm install", pipeline.DefaultRendererDir)
			}

			graph, err := pipeline.LoadSceneGraph(lesson)
			if err != nil {
				return fmt.Errorf("%w\n(run: coursesmith run %s/%s first)", err, course.Slug, lesson.ID)
			}
			staged, err := renderer.StageAssets(lesson, graph, "preview")
			if err != nil {
				return err
			}
			data, err := json.MarshalIndent(staged, "", "  ")
			if err != nil {
				return fmt.Errorf("encoding preview props: %w", err)
			}
			previewJSON := filepath.Join(pipeline.DefaultRendererDir, "public", "preview", "lesson-video.json")
			if err := os.WriteFile(previewJSON, data, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", previewJSON, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Lesson %s staged for preview — opening Remotion Studio (Ctrl-C to stop)...\n", lesson.ID)
			studio := exec.CommandContext(ctx, "npx", "remotion", "studio")
			studio.Dir = pipeline.DefaultRendererDir
			studio.Stdout = cmd.OutOrStdout()
			studio.Stderr = cmd.ErrOrStderr()
			return studio.Run()
		},
	}
}
