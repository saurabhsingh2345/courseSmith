package pipeline

// The reel branches of the plan and scenegraph stages.
//
// Everything between them — verify, audio, align, captions, chapters, render —
// is the ordinary video path, reused unchanged. A reel only has to answer two
// questions differently from a snippet: what gets planned (every segment, not
// one) and how the plan becomes scenes (each template over its own slice).

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// runReelPlan plans every segment, in order, and writes the one script the
// audio stage reads.
//
// Segments are planned SEPARATELY, each through its own template's prompt.
// Asking one call for the whole reel would have been fewer round trips and
// worse in every other way: each template's prompt carries its own vocabulary,
// bounds and enforced shape, and a single merged prompt would either drop those
// or become a document no model follows to the end. Planning per segment also
// means a segment that fails its template's validator fails alone, and can be
// re-planned without disturbing the ones that came out well.
func runReelPlan(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	spec, err := LoadReelSpec(l.Dir)
	if err != nil {
		return err
	}
	if err := spec.Validate(); err != nil {
		return err
	}
	if e.Router == nil {
		return fmt.Errorf("planning a reel needs an LLM — set GROQ_API_KEY (or an OpenAI-compatible provider) and retry")
	}

	active := spec.Active()
	fmt.Fprintf(e.out(), "  → plan      %d segments (%s)...\n", len(active), cfg.Pipeline.LLMContent)

	plan := &ReelPlan{Title: spec.Title}
	for i, seg := range active {
		tpl, ok := SnippetTemplates[seg.Template]
		if !ok {
			return fmt.Errorf("segment %q: unknown template %q", seg.ID, seg.Template)
		}
		fmt.Fprintf(e.out(), "    %d/%d  %-14s %s\n", i+1, len(active), seg.Template, truncateForLog(seg.Prompt, 54))

		planner := tpl.Plan
		if planner == nil {
			planner = planSnippetDefault
		}
		segPlan, err := planner(ctx, e, seg.SnippetSpec(cfg), cfg)
		if err != nil {
			// Named, because "segment 4 of 9 failed" is the difference between
			// re-running one prompt and re-running the whole reel.
			return fmt.Errorf("segment %q (%s): %w", seg.ID, seg.Template, err)
		}
		segPlan.Template = seg.Template
		plan.Segments = append(plan.Segments, ReelPlanSegment{
			ID:       seg.ID,
			Template: seg.Template,
			Plan:     segPlan,
		})
	}

	// A reel with no title of its own takes the first segment's, which is
	// almost always the hook — better than an empty heading on the page.
	if plan.Title == "" && len(plan.Segments) > 0 && plan.Segments[0].Plan != nil {
		plan.Title = plan.Segments[0].Plan.Title
	}

	if err := writeJSON(filepath.Join(l.GeneratedDir(), ReelPlanFileName), plan); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), ScriptFileName), plan.Script(cfg.Style.PaceWPM)); err != nil {
		return err
	}
	md, err := plan.Markdown(*spec)
	if err != nil {
		return err
	}
	if err := writeFileAtomic(l.SourcePath(), []byte(md)); err != nil {
		return err
	}
	// Same reason the snippet stage reloads: the caller holds this lesson from
	// before the plan existed, and verify reads the struct rather than the file.
	reloaded, err := project.LoadLesson(l.Dir)
	if err != nil {
		return fmt.Errorf("planned reel does not load (bug): %w", err)
	}
	*l = *reloaded

	words := 0
	for _, seg := range plan.Segments {
		for _, b := range seg.Plan.Beats {
			words += len(strings.Fields(b.Narration))
		}
	}
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	fmt.Fprintf(e.out(), "    %q — %d segments, %d beats, %d words (~%ds at %d wpm)\n",
		plan.Title, len(plan.Segments), plan.Beats(), words, words*60/pace, pace)
	return nil
}

// truncateForLog keeps a prompt on one line of the progress output.
func truncateForLog(s string, n int) string {
	s = collapseSpaces(s)
	if len([]rune(s)) <= n {
		return s
	}
	return string([]rune(s)[:n-1]) + "…"
}

// runReelScenegraph builds lesson-video.json for a reel: every segment's
// template lays out its own slice of the timeline, and the shared finishing
// pass writes it exactly as it would for a lesson.
func runReelScenegraph(_ context.Context, e *Env, course *project.Course, l *project.Lesson, cfg config.Config) error {
	spec, err := LoadReelSpec(l.Dir)
	if err != nil {
		return err
	}
	plan, err := LoadReelPlan(l)
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

	fmt.Fprintf(e.out(), "  → scenegraph building %s from %d segments...\n", SceneGraphFileName, len(plan.Segments))
	graph, err := buildReelSceneGraph(course, l, cfg, *spec, plan, alignment, verification, int(audioDur.Milliseconds()))
	if err != nil {
		return err
	}
	return finishSceneGraph(e, l, graph)
}

// CreateReel writes a new reel directory and returns its lesson handle.
//
// The stub lesson.md exists only so the directory is a valid lesson from the
// moment it is created — the plan stage overwrites it with the real thing.
func CreateReel(root string, spec ReelSpec) (*project.Course, *project.Lesson, error) {
	spec.EnsureSegmentIDs()
	if err := spec.Validate(); err != nil {
		return nil, nil, err
	}
	course, err := EnsureReelsCourse(root)
	if err != nil {
		return nil, nil, err
	}
	dir := filepath.Join(course.Dir, "lessons", spec.ID)
	if _, err := os.Stat(dir); err == nil {
		return nil, nil, fmt.Errorf("reel %q already exists at %s", spec.ID, dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("creating %s: %w", dir, err)
	}
	if err := SaveReelSpec(dir, &spec); err != nil {
		return nil, nil, err
	}
	title := spec.Title
	if title == "" {
		title = snippetStubTitle(spec.Brief)
	}
	stub := fmt.Sprintf("---\ntitle: %q\n---\n\n## Pending\n\n%s\n",
		title, "This reel has not been planned yet — run the plan stage.")
	if err := writeFileAtomic(filepath.Join(dir, project.LessonFileName), []byte(stub)); err != nil {
		return nil, nil, err
	}
	l, err := project.LoadLesson(dir)
	if err != nil {
		return nil, nil, err
	}
	return course, l, nil
}

// RunReel executes the reel pipeline, skipping up-to-date stages exactly like
// RunLesson and RunSnippet.
func (e *Env) RunReel(ctx context.Context, course *project.Course, l *project.Lesson, opts RunOptions) error {
	if !IsReel(l) {
		return fmt.Errorf("%s is not a reel (no %s)", l.Dir, ReelFileName)
	}
	spec, err := LoadReelSpec(l.Dir)
	if err != nil {
		return err
	}
	stages, err := ReelStages(l)
	if err != nil {
		return err
	}
	cfg := config.Resolve(course.Config, l.FrontMatter.Overrides(), spec.Config)
	return e.runStages(ctx, course, l, cfg, stages, opts)
}
