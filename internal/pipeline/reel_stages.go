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
	"time"

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

	sub, err := LoadSubstance(l)
	if err != nil {
		return err
	}

	active := spec.Active()
	fmt.Fprintf(e.out(), "  → plan      %d segments (%s)...\n", len(active), cfg.Pipeline.LLMContent)

	plan := &ReelPlan{Title: spec.Title}
	// What each planned segment ended up covering, in order. Handed to the next
	// segment so it advances the piece instead of restarting it: planned in
	// isolation, two `myth` segments six apart both argued that you need not know
	// everything before you start, and a `constellation` spent six beats
	// rephrasing its own first sentence.
	//
	// Built from the plan rather than from reel.yaml's prompts on purpose — the
	// prompt is what the segment was *asked* to cover, and after enrichment and a
	// possible recast, what it actually covered can be quite different. The next
	// segment needs the truth, not the instruction.
	var priors []string
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
		segSpec := seg.SnippetSpec(cfg, spec.Brief, priors)
		segSpec.Substance = sub
		segSpec.Prompt = EnrichSnippetPrompt(ctx, e, segSpec, cfg)
		segPlan, err := planner(ctx, e, segSpec, cfg)
		if err != nil {
			// A miscast segment must not kill the reel.
			//
			// The caster is asked to name the material each template needs, and
			// that catches most of it — but "non-empty" is all a validator can
			// check, so a look chosen for a part with nothing to put in it still
			// gets through. `gauge` on "how vibe coding lets people build with
			// AI" has no ceiling and no candidates, and no amount of correcting
			// will conjure them.
			//
			// Seven good segments and one that cannot be planned is a video with
			// a hole, not a failed video. So the segment is recast onto the
			// template with the fewest requirements — the one that needs a
			// subject and nothing else — and the run continues. The log says it
			// happened, and reel.yaml still holds the original choice, so the
			// fix is editing one line rather than starting over.
			fmt.Fprintf(e.out(), "    !  %s could not be planned as %s (%v)\n", seg.ID, seg.Template, errSummary(err))
			fmt.Fprintf(e.out(), "       recasting it as %s — edit reel.yaml if you want a different look\n", reelFallbackTemplate)
			fb := SnippetTemplates[reelFallbackTemplate]
			fbSpec := segSpec
			fbSpec.Template = reelFallbackTemplate
			fbPlanner := fb.Plan
			if fbPlanner == nil {
				fbPlanner = planSnippetDefault
			}
			segPlan, err = fbPlanner(ctx, e, fbSpec, cfg)
			if err != nil {
				return fmt.Errorf("segment %q could not be planned as %s or as %s: %w",
					seg.ID, seg.Template, reelFallbackTemplate, err)
			}
			seg.Template = reelFallbackTemplate
		}
		segPlan.Template = seg.Template
		plan.Segments = append(plan.Segments, ReelPlanSegment{
			ID:       seg.ID,
			Template: seg.Template,
			Plan:     segPlan,
		})
		priors = append(priors, segmentPrior(seg, segPlan))
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
	if spec.ID == "" {
		spec.ID, err = uniqueReelID(course.Dir, spec)
		if err != nil {
			return nil, nil, err
		}
	}
	if spec.CreatedAt.IsZero() {
		spec.CreatedAt = time.Now().UTC().Truncate(time.Second)
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

// uniqueReelID derives a free directory name, the same way a snippet's is:
// slugified from the title (or the brief), suffixed until nothing collides.
func uniqueReelID(courseDir string, spec ReelSpec) (string, error) {
	src := spec.Title
	if strings.TrimSpace(src) == "" {
		src = spec.Brief
	}
	base := slugify(snippetStubTitle(src))
	if base == "" {
		base = "reel"
	}
	if len(base) > 48 {
		base = strings.Trim(base[:48], "-")
	}
	for i := 0; i < 200; i++ {
		id := base
		if i > 0 {
			id = fmt.Sprintf("%s-%d", base, i+1)
		}
		if _, err := os.Stat(filepath.Join(courseDir, "lessons", id)); os.IsNotExist(err) {
			return id, nil
		}
	}
	return "", fmt.Errorf("could not find a free reel id for %q", base)
}

// reelFallbackTemplate is what a segment is recast as when its own template
// cannot be planned from the material. `illustration` is chosen because it is
// the one look in the catalog with no data requirement at all — a headline, a
// line under it, and a figure — so it can carry any subject. A fallback that
// could itself fail would just move the problem.
const reelFallbackTemplate = "illustration"

// maxPriorHeadings caps how much of a planned segment is described to the
// segments after it. The later ones carry the most priors and have the least
// prompt left to spend on them, and what the next writer needs is the ground
// already covered, not a transcript of how it was covered.
const maxPriorHeadings = 4

// segmentPrior summarises one planned segment in a line, for the segments that
// come after it.
//
// The headings rather than the narration: a heading is the beat's claim in a
// handful of words, which is exactly the granularity "do not cover this again"
// needs. Narration would be a dozen times longer and would tempt the next
// writer into matching its phrasing.
func segmentPrior(seg ReelSegment, plan *SnippetPlan) string {
	title := ""
	headings := []string{}
	if plan != nil {
		title = collapseSpaces(plan.Title)
		for _, b := range plan.Beats {
			if h := collapseSpaces(b.Heading); h != "" {
				headings = append(headings, h)
			}
			if len(headings) == maxPriorHeadings {
				break
			}
		}
	}
	// Fall back to what the segment was asked to cover. A segment whose plan
	// came back without a title or any headings is unusual, but a prior that
	// said only "myth:" would teach the next writer nothing.
	if title == "" && len(headings) == 0 {
		return fmt.Sprintf("%s (%s)", truncateForLog(collapseSpaces(seg.Prompt), 90), seg.Template)
	}
	line := fmt.Sprintf("%s (%s)", title, seg.Template)
	if len(headings) > 0 {
		line += ": " + strings.Join(headings, "; ")
	}
	return truncateForLog(line, 180)
}

// errSummary trims a wrapped planner error down to the part worth reading in a
// progress log.
func errSummary(err error) string {
	s := err.Error()
	if i := strings.Index(s, "\n"); i >= 0 {
		s = s[:i]
	}
	return truncateForLog(s, 90)
}
