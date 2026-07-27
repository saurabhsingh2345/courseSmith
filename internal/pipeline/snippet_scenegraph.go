package pipeline

// The snippet branch of the scenegraph stage.

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// runSnippetScenegraph builds lesson-video.json for a snippet: the template
// lays out the scenes, the aligner supplies the timing, and the shared
// finishing pass writes it exactly as it would for a lesson.
func runSnippetScenegraph(_ context.Context, e *Env, course *project.Course, l *project.Lesson, cfg config.Config) error {
	spec, err := LoadSnippetSpec(l.Dir)
	if err != nil {
		return err
	}
	plan, err := LoadSnippetPlan(l)
	if err != nil {
		return err
	}
	alignment, err := loadAlignment(l)
	if err != nil {
		return err
	}
	verification, err := loadVerification(l)
	if err != nil {
		return err
	}
	audioDur, err := wavDuration(filepath.Join(l.GeneratedDir(), VoiceoverFileName))
	if err != nil {
		return fmt.Errorf("no usable %s — the audio stage must run first: %w", VoiceoverFileName, err)
	}

	fmt.Fprintf(e.out(), "  → scenegraph building %s from the %s template...\n", SceneGraphFileName, spec.Template)
	graph, err := buildSnippetSceneGraph(course, l, cfg, *spec, plan, alignment, verification, int(audioDur.Milliseconds()))
	if err != nil {
		return err
	}
	return finishSceneGraph(e, l, graph)
}
