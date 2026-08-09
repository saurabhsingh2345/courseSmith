package pipeline

// The combo branches of the plan and scenegraph stages.
//
// Everything between them — verify, audio, align, captions, chapters, render —
// is the ordinary video path, reused unchanged. A combo only has to answer two
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

// runComboPlan plans every segment, in order, and writes the one script the
// audio stage reads.
//
// Segments are planned SEPARATELY, each through its own template's prompt.
// Asking one call for the whole combo would have been fewer round trips and
// worse in every other way: each template's prompt carries its own vocabulary,
// bounds and enforced shape, and a single merged prompt would either drop those
// or become a document no model follows to the end. Planning per segment also
// means a segment that fails its template's validator fails alone, and can be
// re-planned without disturbing the ones that came out well.
func runComboPlan(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	spec, err := LoadComboSpec(l.Dir)
	if err != nil {
		return err
	}
	if err := spec.Validate(); err != nil {
		return err
	}
	if e.Router == nil {
		return fmt.Errorf("planning a combo needs an LLM — set GROQ_API_KEY (or an OpenAI-compatible provider) and retry")
	}

	sub, err := LoadSubstance(l)
	if err != nil {
		return err
	}

	active := spec.Active()
	fmt.Fprintf(e.out(), "  → plan      %d segments (%s)...\n", len(active), cfg.Pipeline.LLMContent)

	plan := &ComboPlan{Title: spec.Title}
	// What each planned segment ended up covering, in order. Handed to the next
	// segment so it advances the piece instead of restarting it: planned in
	// isolation, two `myth` segments six apart both argued that you need not know
	// everything before you start, and a `constellation` spent six beats
	// rephrasing its own first sentence.
	//
	// Built from the plan rather than from combo.yaml's prompts on purpose — the
	// prompt is what the segment was *asked* to cover, and after enrichment and a
	// possible recast, what it actually covered can be quite different. The next
	// segment needs the truth, not the instruction.
	var priors []string
	// The spec each segment was planned from, kept so the critic can re-plan one
	// through exactly the same request with a criticism attached.
	//
	// Rebuilding it afterwards from combo.yaml would not be the same request: the
	// prompt handed to the planner has been through enrichment, and the priors it
	// carried were whatever had been planned before it. A repair against a
	// reconstructed spec is a repair of a segment that was never made.
	segSpecs := make(map[string]SnippetSpec, len(active))
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
			// A miscast segment must not kill the combo.
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
			// happened, and combo.yaml still holds the original choice, so the
			// fix is editing one line rather than starting over.
			fmt.Fprintf(e.out(), "    !  %s could not be planned as %s (%v)\n", seg.ID, seg.Template, errSummary(err))
			fmt.Fprintf(e.out(), "       recasting it as %s — edit combo.yaml if you want a different look\n", comboFallbackTemplate)
			fb := SnippetTemplates[comboFallbackTemplate]
			fbSpec := segSpec
			fbSpec.Template = comboFallbackTemplate
			fbPlanner := fb.Plan
			if fbPlanner == nil {
				fbPlanner = planSnippetDefault
			}
			segPlan, err = fbPlanner(ctx, e, fbSpec, cfg)
			if err != nil {
				return fmt.Errorf("segment %q could not be planned as %s or as %s: %w",
					seg.ID, seg.Template, comboFallbackTemplate, err)
			}
			seg.Template = comboFallbackTemplate
		}
		segPlan.Template = seg.Template
		// Gate each segment as it is planned, not the finished combo.
		//
		// A critique of segment three has nothing to say about the other four, and
		// gating the whole plan would re-plan all of them to fix one — five calls
		// to repair one, and the four that were already good get a fresh chance to
		// come back worse.
		segSpec.Template = seg.Template
		segPlan = e.gateSegmentPlan(ctx, l, cfg, segSpec, segPlan)

		segSpecs[seg.ID] = segSpec
		plan.Segments = append(plan.Segments, ComboPlanSegment{
			ID:       seg.ID,
			Template: seg.Template,
			Plan:     segPlan,
		})
		priors = append(priors, segmentPrior(seg, segPlan))
	}

	// A combo with no title of its own takes the first segment's, which is
	// almost always the hook — better than an empty heading on the page.
	if plan.Title == "" && len(plan.Segments) > 0 && plan.Segments[0].Plan != nil {
		plan.Title = plan.Segments[0].Plan.Title
	}

	// And now read the whole thing as a viewer would.
	//
	// This is the only pass that sees every segment at once, and it catches the
	// class of defect that is invisible from inside one: a part that repeats what
	// segment three established, contradicts it, or is simply true and does not
	// move the argument along. Each of those segments passed its own validator and
	// scored well on its own rubric, because from where those stand there is
	// nothing wrong with it.
	//
	// After the priors chain rather than instead of it. Priors stop a segment
	// restarting the piece, which is a forward-looking fix and only knows about
	// what came BEFORE; the critic is the backward-looking one, and it is the only
	// thing that can tell segment three it should have left room for segment nine.
	fmt.Fprintf(e.out(), "  → critique  reading %d segments as one piece...\n", len(plan.Segments))
	plan = e.criticiseCombo(ctx, spec, plan, cfg, segSpecs)

	if err := writeJSON(filepath.Join(l.GeneratedDir(), ComboPlanFileName), plan); err != nil {
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
		return fmt.Errorf("planned combo does not load (bug): %w", err)
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

	// The compromises, gathered at the end rather than only warned about as they
	// happened. On a ten-segment combo the per-segment warnings have scrolled well
	// off the screen by the time the run finishes, which is how three
	// non-compliant segments went unnoticed in a finished video.
	var loose []string
	for _, seg := range plan.Segments {
		if seg.Plan != nil && len(seg.Plan.Compromises) > 0 {
			loose = append(loose, fmt.Sprintf("%s (%s): %s",
				seg.ID, seg.Template, strings.Join(seg.Plan.Compromises, "; ")))
		}
	}
	if len(loose) > 0 {
		fmt.Fprintf(e.out(), "    %d of %d segments shipped looser than asked, recorded in %s:\n",
			len(loose), len(plan.Segments), ComboPlanFileName)
		for _, line := range loose {
			fmt.Fprintf(e.out(), "      %s\n", truncateForLog(line, 110))
		}
	}
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

// runComboScenegraph builds lesson-video.json for a combo: every segment's
// template lays out its own slice of the timeline, and the shared finishing
// pass writes it exactly as it would for a lesson.
func runComboScenegraph(_ context.Context, e *Env, course *project.Course, l *project.Lesson, cfg config.Config) error {
	// A no-code piece assembles through here too and has no combo.yaml. The
	// assembler only reads the spec for the piece's identity, so a synthesised
	// one carries everything it needs — the alternative was a second copy of
	// this function differing in one line.
	var spec *ComboSpec
	if IsNoCode(l) {
		nc, err := LoadNoCodeSpec(l.Dir)
		if err != nil {
			return err
		}
		spec = &ComboSpec{ID: nc.ID, Title: nc.Title, Brief: nc.Brief}
	} else {
		var err error
		spec, err = LoadComboSpec(l.Dir)
		if err != nil {
			return err
		}
	}
	plan, err := LoadComboPlan(l)
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
	graph, err := buildComboSceneGraph(course, l, cfg, *spec, plan, alignment, verification, int(audioDur.Milliseconds()))
	if err != nil {
		return err
	}
	return finishSceneGraph(e, l, graph)
}

// CreateCombo writes a new combo directory and returns its lesson handle.
//
// The stub lesson.md exists only so the directory is a valid lesson from the
// moment it is created — the plan stage overwrites it with the real thing.
func CreateCombo(root string, spec ComboSpec) (*project.Course, *project.Lesson, error) {
	spec.EnsureSegmentIDs()
	if err := spec.Validate(); err != nil {
		return nil, nil, err
	}
	course, err := EnsureCombosCourse(root)
	if err != nil {
		return nil, nil, err
	}
	if spec.ID == "" {
		spec.ID, err = uniqueComboID(course.Dir, spec)
		if err != nil {
			return nil, nil, err
		}
	}
	if spec.CreatedAt.IsZero() {
		spec.CreatedAt = time.Now().UTC().Truncate(time.Second)
	}
	dir := filepath.Join(course.Dir, "lessons", spec.ID)
	if _, err := os.Stat(dir); err == nil {
		return nil, nil, fmt.Errorf("combo %q already exists at %s", spec.ID, dir)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, fmt.Errorf("creating %s: %w", dir, err)
	}
	if err := SaveComboSpec(dir, &spec); err != nil {
		return nil, nil, err
	}
	title := spec.Title
	if title == "" {
		title = snippetStubTitle(spec.Brief)
	}
	stub := fmt.Sprintf("---\ntitle: %q\n---\n\n## Pending\n\n%s\n",
		title, "This combo has not been planned yet — run the plan stage.")
	if err := writeFileAtomic(filepath.Join(dir, project.LessonFileName), []byte(stub)); err != nil {
		return nil, nil, err
	}
	l, err := project.LoadLesson(dir)
	if err != nil {
		return nil, nil, err
	}
	return course, l, nil
}

// RunCombo executes the combo pipeline, skipping up-to-date stages exactly like
// RunLesson and RunSnippet.
func (e *Env) RunCombo(ctx context.Context, course *project.Course, l *project.Lesson, opts RunOptions) error {
	if !IsCombo(l) {
		return fmt.Errorf("%s is not a combo (no %s)", l.Dir, ComboFileName)
	}
	spec, err := LoadComboSpec(l.Dir)
	if err != nil {
		return err
	}
	stages, err := ComboStages(l)
	if err != nil {
		return err
	}
	cfg := config.Resolve(course.Config, l.FrontMatter.Overrides(), spec.Config)
	return e.runStages(ctx, course, l, cfg, stages, opts)
}

// uniqueComboID derives a free directory name, the same way a snippet's is:
// slugified from the title (or the brief), suffixed until nothing collides.
func uniqueComboID(courseDir string, spec ComboSpec) (string, error) {
	src := spec.Title
	if strings.TrimSpace(src) == "" {
		src = spec.Brief
	}
	base := slugify(snippetStubTitle(src))
	if base == "" {
		base = "combo"
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
	return "", fmt.Errorf("could not find a free combo id for %q", base)
}

// comboFallbackTemplate is what a segment is recast as when its own template
// cannot be planned from the material. `illustration` is chosen because it is
// the one look in the catalog with no data requirement at all — a headline, a
// line under it, and a figure — so it can carry any subject. A fallback that
// could itself fail would just move the problem.
const comboFallbackTemplate = "illustration"

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
func segmentPrior(seg ComboSegment, plan *SnippetPlan) string {
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
