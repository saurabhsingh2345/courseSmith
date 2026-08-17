package pipeline

// Planning a no-code piece.
//
// Structurally this is a combo: several segments, each planned through its own
// template's prompt, cut onto one timeline. So it reuses the combo's plan
// format, script assembly, markdown and scenegraph wholesale rather than
// growing a second copy of all of it — the differences are at the two ends.
//
// **Before planning**, each segment's evidence is resolved. A capture-backed
// segment has its clip found on disk and its real marks, duration and tool
// attached to the spec, so the writer is told what the recording contains and
// the validator can check the plan against it. A fact-backed segment has its
// claims attached the same way.
//
// **On failure**, nothing is recast. A combo rescues a miscast segment onto
// `illustration`, which is right for a combo and forbidden here — that template
// is excluded from this surface precisely because it fills a frame without
// evidence. A no-code segment that cannot be planned is a segment whose
// evidence did not survive contact with its template, and quietly replacing it
// with a drawn figure would break the one promise the surface makes.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
	"gopkg.in/yaml.v3"
)

// resolveCaptureEvidence finds a segment's recording and describes it.
//
// The clip is looked up by the id the segment named — `capture-1` — against the
// sidecars the capture stage wrote. Everything attached here is measured: the
// marks the recorder stamped, the real duration, the tool's display name. The
// writer gets those and nothing else, which is what makes "do not describe
// moments the recording does not have" an enforceable rule rather than a hope.
func resolveCaptureEvidence(l *project.Lesson, captureID string, spec *SnippetSpec) error {
	f, ok := loadFootageFor(l, captureID)
	if !ok {
		return fmt.Errorf("no recording called %q has been captured yet. Record it first (`coursesmith run` runs the capture stage; a desktop take needs `coursesmith footage shoot`), then plan", captureID)
	}
	if f.DurationMs == 0 && len(f.Frames) == 0 {
		return fmt.Errorf("the recording %q exists but is empty — re-shoot it", captureID)
	}

	marks := make([]string, 0, len(f.Marks)+len(f.Frames))
	for _, m := range f.Marks {
		marks = append(marks, m.Name)
	}
	for _, fr := range f.Frames {
		marks = append(marks, fr.Mark)
	}
	spec.FootageMarks = marks
	spec.FootageMs = f.DurationMs
	spec.FootageTool = captureDisplayName(f)
	return nil
}

// captureDisplayName is how the recorded tool is named to a reader, spelled the
// way its owner spells it.
func captureDisplayName(f Footage) string {
	switch f.Kind {
	case CaptureKindTool:
		if t, ok := captureTools[f.Tool]; ok {
			return t.Display
		}
	case CaptureKindWeb:
		if s, ok := captureSites[f.Tool]; ok {
			return s.Display
		}
	case CaptureKindDesktop:
		if a, ok := captureApps[f.Tool]; ok {
			return a.Display
		}
	}
	return f.Tool
}

// attachFootageScene fills in what the scene needs to play the clip.
//
// Kept apart from the writer's inputs on purpose: the model decides narration
// and which moments to anchor to, and these fields — where the file is, how
// long it is, what to put in the address bar — are resolved from the sidecar
// and never come from a plan.
func attachFootageScene(l *project.Lesson, captureID string, plan *SnippetPlan) error {
	f, ok := loadFootageFor(l, captureID)
	if !ok {
		return fmt.Errorf("recording %q vanished between planning and assembly", captureID)
	}
	plan.FootageMs = f.DurationMs
	plan.FootageOrigin = f.Origin
	plan.FootageKnownMarks = markNamesOf(f)
	plan.FootageIsTerminal = f.Kind == CaptureKindTool
	plan.FootageTitle = captureDisplayName(f)
	if f.Origin != "" {
		plan.FootageTitle = f.Origin
	}
	// The clip path, relative to the lesson's generated dir, is whatever the
	// recorder produced: an mp4 for the terminal, desktop and web-video kinds.
	plan.FootageSrc = filepath.Join(DemosDirName, captureID+".mp4")
	return nil
}

// markNamesOf lists a clip's moments in order.
func markNamesOf(f Footage) []string {
	out := make([]string, 0, len(f.Marks)+len(f.Frames))
	for _, m := range f.Marks {
		out = append(out, m.Name)
	}
	for _, fr := range f.Frames {
		out = append(out, fr.Mark)
	}
	return out
}

// runNoCodePlan plans every segment of a no-code piece.
func runNoCodePlan(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	spec, err := LoadNoCodeSpec(l.Dir)
	if err != nil {
		return err
	}
	if err := spec.Validate(); err != nil {
		return err
	}
	if e.Router == nil {
		return fmt.Errorf("planning a no-code piece needs an LLM — set OPENAI_API_KEY (or another provider) and retry")
	}
	sub, err := LoadSubstance(l)
	if err != nil {
		return err
	}

	live := spec.Live()
	fmt.Fprintf(e.out(), "  → plan      %d segments (%s)...\n", len(live), cfg.Pipeline.LLMContent)

	plan := &ComboPlan{Title: spec.Title}
	var priors []string
	for i, seg := range live {
		tpl, ok := SnippetTemplates[seg.Template]
		if !ok {
			return fmt.Errorf("segment %q: unknown template %q", seg.ID, seg.Template)
		}
		fmt.Fprintf(e.out(), "    %d/%d  %-14s %s\n", i+1, len(live), seg.Template, truncateForLog(seg.Prompt, 54))

		segSpec := SnippetSpec{
			ID:       seg.ID,
			Prompt:   seg.Prompt,
			Template: seg.Template,
			Config:   cfg,
			Brief:    spec.Brief,
			Priors:   priors,
		}
		segSpec.Substance = sub

		switch seg.Evidence.Kind {
		case EvidenceCapture:
			if err := resolveCaptureEvidence(l, seg.ID, &segSpec); err != nil {
				return fmt.Errorf("segment %q: %w", seg.ID, err)
			}
			fmt.Fprintf(e.out(), "         backed by %s (%s, %ds)\n",
				seg.ID, segSpec.FootageTool, (segSpec.FootageMs+500)/1000)
		case EvidenceFact:
			// The claims become the segment's material, which is the field every
			// template's prompt already reads. Written out in nocode.yaml rather
			// than referenced into substance.json, so a regenerated fact sheet
			// cannot silently change what this segment stands on.
			segSpec.Material = strings.Join(seg.Evidence.Facts, "; ")
			fmt.Fprintf(e.out(), "         backed by %d fact(s)\n", len(seg.Evidence.Facts))
		}

		planner := tpl.Plan
		if planner == nil {
			planner = planSnippetDefault
		}
		segSpec.Prompt = EnrichSnippetPrompt(ctx, e, segSpec, cfg)
		segPlan, err := planner(ctx, e, segSpec, cfg)
		if err != nil {
			// Nothing is recast here, unlike a combo. A combo rescues a miscast
			// segment onto `illustration`; that template is excluded from this
			// surface precisely because it fills a frame without evidence, so
			// using it as a rescue would break the one promise the surface makes.
			return fmt.Errorf("segment %q could not be planned as %s: %w\n"+
				"A no-code segment is not recast onto a drawn frame — that is what this surface exists to refuse. Either give it evidence its template can use, or change the template in nocode.yaml",
				seg.ID, seg.Template, err)
		}
		segPlan.Template = seg.Template
		if seg.Evidence.Kind == EvidenceCapture {
			if err := attachFootageScene(l, seg.ID, segPlan); err != nil {
				return fmt.Errorf("segment %q: %w", seg.ID, err)
			}
		}
		segPlan = e.gateSegmentPlan(ctx, l, cfg, segSpec, segPlan)

		plan.Segments = append(plan.Segments, ComboPlanSegment{
			ID:       seg.ID,
			Template: seg.Template,
			Plan:     segPlan,
		})
		priors = append(priors, seg.Prompt)
	}

	if plan.Title == "" && len(plan.Segments) > 0 && plan.Segments[0].Plan != nil {
		plan.Title = plan.Segments[0].Plan.Title
	}

	if err := writeJSON(filepath.Join(l.GeneratedDir(), ComboPlanFileName), plan); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), ScriptFileName), plan.Script(cfg.Style.PaceWPM)); err != nil {
		return err
	}
	md, err := plan.Markdown(ComboSpec{ID: spec.ID, Title: spec.Title, Brief: spec.Brief})
	if err != nil {
		return err
	}
	if err := writeFileAtomic(l.SourcePath(), []byte(md)); err != nil {
		return err
	}
	reloaded, err := project.LoadLesson(l.Dir)
	if err != nil {
		return fmt.Errorf("planned piece does not load (bug): %w", err)
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

// RunNoCode executes the no-code pipeline for one piece.
//
// A separate entry point for the same reason snippets and combos have one: the
// stage list is not the course's. Calling RunLesson here walks the lesson
// pipeline — trace, review, storyboard, visuals — spends real tokens on stages
// that have nothing to do, and then fails at scenegraph looking for a plan no
// stage in that list ever writes.
func (e *Env) RunNoCode(ctx context.Context, course *project.Course, l *project.Lesson, opts RunOptions) error {
	if !IsNoCode(l) {
		return fmt.Errorf("%s is not a no-code piece (no %s)", l.Dir, NoCodeFileName)
	}
	spec, err := LoadNoCodeSpec(l.Dir)
	if err != nil {
		return err
	}
	if err := spec.Validate(); err != nil {
		return err
	}
	stages, err := NoCodeStages(l)
	if err != nil {
		return err
	}
	cfg := config.Resolve(course.Config, l.FrontMatter.Overrides(), spec.Config)
	fmt.Fprintf(e.out(), "%s — %d segment(s), %d backed by a recording\n",
		l.ID, len(spec.Live()), len(spec.CaptureIDs()))
	return e.runStages(ctx, course, l, cfg, stages, opts)
}

// NewNoCodePiece scaffolds a piece on disk and returns its lesson.
func NewNoCodePiece(root string, spec NoCodeSpec) (*project.Course, *project.Lesson, error) {
	course, err := EnsureNoCodeCourse(root)
	if err != nil {
		return nil, nil, err
	}
	if spec.ID == "" {
		spec.ID = project.Slugify(spec.Title)
	}
	if spec.ID == "" {
		return nil, nil, fmt.Errorf("a piece needs a title or an id")
	}
	spec.assignIDs()
	if err := spec.Validate(); err != nil {
		return nil, nil, err
	}

	dir := filepath.Join(course.Dir, "lessons", spec.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, nil, err
	}
	if err := SaveNoCodeSpec(dir, &spec); err != nil {
		return nil, nil, err
	}

	title := spec.Title
	if title == "" {
		title = spec.ID
	}
	stub := fmt.Sprintf("---\ntitle: %q\n---\n\n## Pending\n\nThis piece has not been planned yet — run the plan stage.\n", title)
	if err := writeFileAtomic(filepath.Join(dir, project.LessonFileName), []byte(stub)); err != nil {
		return nil, nil, err
	}
	l, err := project.LoadLesson(dir)
	if err != nil {
		return nil, nil, err
	}
	return course, l, nil
}

// SaveNoCodeSpec writes nocode.yaml, pruned so a file people are told to edit
// is readable when they get there.
func SaveNoCodeSpec(dir string, spec *NoCodeSpec) error {
	spec.assignIDs()
	raw, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", NoCodeFileName, err)
	}
	var tree map[string]any
	if err := yaml.Unmarshal(raw, &tree); err != nil {
		return fmt.Errorf("encoding %s: %w", NoCodeFileName, err)
	}
	pruneEmpty(tree)
	out, err := yaml.Marshal(tree)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", NoCodeFileName, err)
	}
	return os.WriteFile(filepath.Join(dir, NoCodeFileName), out, 0o644)
}

// ListNoCodePieces returns every piece, newest first.
func ListNoCodePieces(root string) ([]*project.Lesson, error) {
	course, err := EnsureNoCodeCourse(root)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(course.Dir, "lessons")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []*project.Lesson
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		l, err := project.LoadLesson(filepath.Join(dir, e.Name()))
		if err != nil {
			continue // a half-written piece should not break the list
		}
		if IsNoCode(l) {
			out = append(out, l)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

// FindNoCodePiece resolves a piece by id.
func FindNoCodePiece(root, id string) (*project.Course, *project.Lesson, error) {
	course, err := EnsureNoCodeCourse(root)
	if err != nil {
		return nil, nil, err
	}
	l, err := project.LoadLesson(filepath.Join(course.Dir, "lessons", id))
	if err != nil {
		return nil, nil, fmt.Errorf("no-code piece %q not found: %w", id, err)
	}
	if !IsNoCode(l) {
		return nil, nil, fmt.Errorf("%q is not a no-code piece", id)
	}
	return course, l, nil
}
