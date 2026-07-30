package pipeline

// Reels: one video cut from several templates.
//
// A snippet is one template start to finish, which is the right shape for
// thirty seconds and the wrong one for ten minutes — nothing holds attention
// through ten minutes of the same picture. A reel is an ordered run of
// segments, each with its own template, rendered onto ONE timeline.
//
// The decision that matters is that a reel is not several clips stitched
// together. There is one narration, one TTS pass and one alignment across the
// whole piece. Stitching would give seams at every join, loudness that drifts
// between segments, and a supercut rather than a video; more practically, it
// would mean the boundaries could never move without re-rendering both sides.
//
// What makes this cheap is that alignment spans are absolute milliseconds. A
// segment's template does not know it is part of a reel and does not need to:
// it is handed the slice of spans covering its own beats, already timed against
// the finished audio, and lays its scenes out exactly as it would in a snippet.
// Assembly is slicing, not arithmetic. See buildReelSceneGraph.
//
// On disk a reel is an ordinary lesson directory inside a synthetic course,
// exactly like a snippet, so the stage machinery, state tracking and artifact
// serving work with no special cases:
//
//	.coursesmith/reels/
//	  course.yaml
//	  lessons/<id>/
//	    reel.yaml            the request: brief, and the ordered segments
//	    lesson.md            synthesized by the plan stage
//	    generated/           reel-plan.json, script.json, …, final.mp4
//
// Phase 1 is deliberately manual: you name the template for each segment. The
// casting stage that picks them is the next piece, and it writes exactly the
// same reel.yaml a person would — which is what makes it possible to hand-edit
// the result afterwards rather than being stuck with whatever the model chose.

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

const (
	// ReelFileName is the request file in a reel's directory.
	ReelFileName = "reel.yaml"
	// ReelPlanFileName is the plan stage output in generated/.
	ReelPlanFileName = "reel-plan.json"
	// ReelsCourseSlug is the slug of the synthetic course reels live in.
	ReelsCourseSlug = "reels"
)

// ReelsRoot is where reel courses live, relative to the project root.
var ReelsRoot = filepath.Join(".coursesmith", "reels")

// Reel length bounds. The floor is two segments because one is a snippet and
// should be made as one — the whole point of a reel is the cut between looks.
// The ceiling is not about rendering but about attention and cost: twelve
// segments is roughly a fifteen-minute piece, and every segment is its own
// planning call.
const (
	minReelSegments = 2
	maxReelSegments = 12
)

// ReelSpec is reel.yaml: the request, and the edit surface.
//
// This file is what a person changes after seeing the first cut — retitle a
// segment, rewrite its prompt, swap its template, drop it. Editing it re-stales
// the stages that depend on it, the same way editing a lesson does, which is
// why the segments carry stable ids: an edit has to be addressable to something
// that does not move when its neighbours change.
type ReelSpec struct {
	// ID is the reel's directory name, and its stable handle everywhere.
	ID string `yaml:"id"`
	// Title is the finished piece's title.
	Title string `yaml:"title,omitempty"`
	// Brief is the creator's description of the whole piece in their own words.
	//
	// Unused by the Phase 1 pipeline, and deliberately carried anyway: it is
	// the input the casting stage will read, and recording it now means a reel
	// authored by hand today can be re-cast later without the intent having
	// been thrown away.
	Brief string `yaml:"brief,omitempty"`
	// Segments are the parts, in order.
	Segments []ReelSegment `yaml:"segments"`
	// Config overrides the course defaults for this reel alone.
	Config config.Config `yaml:",inline"`

	CreatedAt time.Time `yaml:"created_at,omitempty"`
}

// ReelSegment is one part of the reel: a template, and what it should cover.
type ReelSegment struct {
	// ID is stable across edits and is how the editor addresses this segment.
	// Generated from the position when absent, but only once — renumbering on
	// every edit would make every id a lie the moment a segment moved.
	ID string `yaml:"id"`
	// Template names a registered template (see snippet_templates.go).
	Template string `yaml:"template"`
	// Prompt is what this segment should cover, in the creator's words. It is
	// planned through the template's own prompt, so a segment is exactly as
	// good as the equivalent snippet would have been.
	Prompt string `yaml:"prompt"`
	// Material is the concrete facts this template will be filled with — the
	// ceiling and its candidates, the line items that add up, the belief and the
	// truth. The caster names it to prove the template can be filled at all
	// (CastReel), and it is written down here because the segment's *writer*
	// needs it more than the validator did.
	//
	// Persisted, unlike the rest of the planning context, because it is a
	// per-segment choice a creator will want to edit — correcting a wrong figure
	// here is the difference between re-running one segment and re-casting the
	// reel. Empty is legal: a hand-authored reel need not supply it, and a
	// segment without it is planned exactly as it was before this field existed.
	Material string `yaml:"material,omitempty"`
	// TargetSec is the runtime to aim for (0 = the template's own default).
	TargetSec int `yaml:"target_sec,omitempty"`
	// Skip drops the segment from the cut without deleting it from the file.
	//
	// A separate flag rather than asking people to delete lines, because the
	// commonest edit after watching a first cut is "lose that bit" and the
	// second commonest is "actually put it back". Deleting loses the prompt
	// that produced it.
	Skip bool `yaml:"skip,omitempty"`
}

// Active returns the segments that are actually in the cut, in order.
func (r ReelSpec) Active() []ReelSegment {
	out := make([]ReelSegment, 0, len(r.Segments))
	for _, s := range r.Segments {
		if !s.Skip {
			out = append(out, s)
		}
	}
	return out
}

// ResolvedTargetSec returns the runtime this segment aims for, defaulted from
// the template the same way a snippet's is.
func (s ReelSegment) ResolvedTargetSec() int {
	spec := SnippetSpec{Template: s.Template, TargetSec: s.TargetSec}
	return spec.ResolvedTargetSec()
}

// SnippetSpec projects a segment onto the request shape the template planners
// already take, so a segment is planned by exactly the same code a snippet is.
// Nothing in a template knows about reels.
//
// brief and priors come from the reel rather than the segment, so they are
// parameters: a segment cannot know the piece it is part of, and asking the
// caller to remember to set two fields afterwards is how one of them ends up
// unset on the path nobody re-read.
func (s ReelSegment) SnippetSpec(cfg config.Config, brief string, priors []string) SnippetSpec {
	return SnippetSpec{
		ID:        s.ID,
		Prompt:    s.Prompt,
		Template:  s.Template,
		TargetSec: s.TargetSec,
		Config:    cfg,
		Brief:     brief,
		Material:  s.Material,
		Priors:    priors,
	}
}

// Validate checks a request before anything is written to disk.
func (r ReelSpec) Validate() error {
	active := r.Active()
	if len(active) < minReelSegments {
		return fmt.Errorf("a reel needs at least %d segments in the cut (found %d). One template start to finish is a snippet, and `coursesmith snippet` makes a better job of it",
			minReelSegments, len(active))
	}
	if len(r.Segments) > maxReelSegments {
		return fmt.Errorf("a reel takes at most %d segments (found %d) — every segment is its own planning call, and past this it is a course",
			maxReelSegments, len(r.Segments))
	}
	seen := map[string]bool{}
	for i, s := range r.Segments {
		if strings.TrimSpace(s.Template) == "" {
			return fmt.Errorf("segment %d has no template (templates: %s)", i, strings.Join(SnippetTemplateNames(), ", "))
		}
		if _, ok := SnippetTemplates[s.Template]; !ok {
			return fmt.Errorf("segment %d: unknown template %q (templates: %s)", i, s.Template, strings.Join(SnippetTemplateNames(), ", "))
		}
		if strings.TrimSpace(s.Prompt) == "" && !s.Skip {
			return fmt.Errorf("segment %d (%s) has no prompt — say what it should cover", i, s.Template)
		}
		if s.ID != "" {
			if seen[s.ID] {
				return fmt.Errorf("two segments share the id %q — ids are how an edit addresses one of them", s.ID)
			}
			seen[s.ID] = true
		}
		// Caught here rather than three correction rounds later, exactly as the
		// snippet spec does: a runtime below a template's floor is not a
		// segment that plans badly, it is one whose rules contradict.
		if tpl, ok := SnippetTemplates[s.Template]; ok && s.TargetSec != 0 &&
			tpl.MinTargetSec > 0 && s.TargetSec < tpl.MinTargetSec {
			return fmt.Errorf("segment %d: the %s template needs at least %d seconds (asked for %d)",
				i, s.Template, tpl.MinTargetSec, s.TargetSec)
		}
	}
	return nil
}

// EnsureSegmentIDs fills in any missing segment ids, in place.
//
// Derived from the template name plus an ordinal, and only for segments that
// have none: an id that changed when a neighbour moved would break every edit
// addressed to it, which is the one thing ids exist to prevent.
func (r *ReelSpec) EnsureSegmentIDs() {
	used := map[string]bool{}
	for _, s := range r.Segments {
		if s.ID != "" {
			used[s.ID] = true
		}
	}
	for i := range r.Segments {
		if r.Segments[i].ID != "" {
			continue
		}
		base := r.Segments[i].Template
		if base == "" {
			base = "segment"
		}
		id := fmt.Sprintf("%s-%d", base, i+1)
		for n := 2; used[id]; n++ {
			id = fmt.Sprintf("%s-%d-%d", base, i+1, n)
		}
		used[id] = true
		r.Segments[i].ID = id
	}
}

// IsReel reports whether a lesson directory is a reel.
func IsReel(l *project.Lesson) bool {
	_, err := os.Stat(filepath.Join(l.Dir, ReelFileName))
	return err == nil
}

// LoadReelSpec reads reel.yaml from a lesson directory.
func LoadReelSpec(dir string) (*ReelSpec, error) {
	raw, err := os.ReadFile(filepath.Join(dir, ReelFileName))
	if err != nil {
		return nil, err
	}
	var spec ReelSpec
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", ReelFileName, err)
	}
	if spec.ID == "" {
		spec.ID = filepath.Base(dir)
	}
	spec.EnsureSegmentIDs()
	return &spec, nil
}

// SaveReelSpec writes reel.yaml, filling in segment ids first so the file on
// disk is the addressable one.
//
// The inline config is pruned of everything left at its zero value before
// writing. config.Config has no omitempty tags — it cannot, since a course
// manifest legitimately records an explicit zero — so marshalling the spec
// straight out buries the twelve lines that matter under forty lines of empty
// strings. That is tolerable for a file nobody opens; reel.yaml is the
// documented edit surface, and a surface people are told to edit has to be
// readable when they get there.
func SaveReelSpec(dir string, spec *ReelSpec) error {
	spec.EnsureSegmentIDs()
	raw, err := yaml.Marshal(spec)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", ReelFileName, err)
	}
	// Round-trip through a generic tree so the pruning is structural rather
	// than a set of string rules about which keys to drop.
	var tree map[string]any
	if err := yaml.Unmarshal(raw, &tree); err != nil {
		return fmt.Errorf("encoding %s: %w", ReelFileName, err)
	}
	pruneEmpty(tree)
	out, err := yaml.Marshal(tree)
	if err != nil {
		return fmt.Errorf("encoding %s: %w", ReelFileName, err)
	}
	return os.WriteFile(filepath.Join(dir, ReelFileName), out, 0o644)
}

// pruneEmpty removes keys whose value is a zero scalar, or a map that is empty
// once its own zero keys are gone.
//
// Deliberately does NOT touch lists: a segment is a map inside a list and its
// own empty keys are pruned by the recursion, but an empty list stays, because
// "segments: []" is a meaningful thing for a person to see and fill in.
func pruneEmpty(m map[string]any) {
	for k, v := range m {
		switch t := v.(type) {
		case map[string]any:
			pruneEmpty(t)
			if len(t) == 0 {
				delete(m, k)
			}
		case []any:
			for _, item := range t {
				if child, ok := item.(map[string]any); ok {
					pruneEmpty(child)
				}
			}
		case string:
			if t == "" {
				delete(m, k)
			}
		case int, int64:
			if fmt.Sprint(t) == "0" {
				delete(m, k)
			}
		case float64:
			if t == 0 {
				delete(m, k)
			}
		case bool:
			if !t {
				delete(m, k)
			}
		case nil:
			delete(m, k)
		}
	}
}

// ReelPlan is reel-plan.json: every segment planned, in order.
//
// One SnippetPlan per segment rather than a merged plan, because merging would
// throw away the thing the scenegraph stage needs — which beats belong to which
// template. The templates' validators also only make sense against their own
// plan.
type ReelPlan struct {
	Title    string            `json:"title"`
	Segments []ReelPlanSegment `json:"segments"`
}

// ReelPlanSegment is one planned segment.
type ReelPlanSegment struct {
	// ID matches the ReelSegment it came from, so an edit to reel.yaml can be
	// traced to the plan it produced.
	ID       string       `json:"id"`
	Template string       `json:"template"`
	Plan     *SnippetPlan `json:"plan"`
}

// Beats returns the total beat count across the reel — the number of aligned
// sections the scenegraph stage expects.
func (p *ReelPlan) Beats() int {
	n := 0
	for _, s := range p.Segments {
		if s.Plan != nil {
			n += len(s.Plan.Beats)
		}
	}
	return n
}

// Script converts every segment's narration into the one script.json the
// audio, align and chapters stages already consume.
//
// The sections are the beats of every segment in order, which is what makes the
// whole reel a single continuous read: the TTS stage never learns where one
// segment ends, so there is no seam to hear.
//
// Section ids are prefixed with the segment id. Beat ids are only unique within
// a template's plan — two segments that both open on a beat called "intro"
// would otherwise collide, and the aligner keys sections by id.
func (p *ReelPlan) Script(paceWPM int) *Script {
	if paceWPM <= 0 {
		paceWPM = 150
	}
	script := &Script{Title: p.Title}
	for _, seg := range p.Segments {
		if seg.Plan == nil {
			continue
		}
		for _, b := range seg.Plan.Beats {
			words := len(strings.Fields(b.Narration))
			est := max(1, int(float64(words)/float64(paceWPM)*60+0.5))
			script.Sections = append(script.Sections, Section{
				ID:             seg.ID + "--" + b.ID,
				Narration:      b.Narration,
				DurationEstSec: est,
			})
		}
	}
	return script
}

// Markdown renders the reel as an ordinary lesson.md, so nothing downstream
// needs to know the piece was assembled rather than written.
func (p *ReelPlan) Markdown(spec ReelSpec) (string, error) {
	fm := project.FrontMatter{
		Title:    p.Title,
		Style:    spec.Config.Style,
		Branding: spec.Config.Branding,
		Pipeline: spec.Config.Pipeline,
	}
	fmData, err := yaml.Marshal(fm)
	if err != nil {
		return "", fmt.Errorf("encoding reel front-matter: %w", err)
	}
	var sb strings.Builder
	sb.WriteString("---\n")
	sb.Write(fmData)
	sb.WriteString("---\n\n")
	for _, seg := range p.Segments {
		if seg.Plan == nil {
			continue
		}
		for _, b := range seg.Plan.Beats {
			sb.WriteString("## " + b.Heading + "\n\n")
			sb.WriteString(b.Narration + "\n\n")
			if b.Code != "" {
				sb.WriteString("```" + "\n" + b.Code + "\n```\n\n")
			}
		}
	}
	return sb.String(), nil
}

// LoadReelPlan reads generated/reel-plan.json.
func LoadReelPlan(l *project.Lesson) (*ReelPlan, error) {
	path := filepath.Join(l.GeneratedDir(), ReelPlanFileName)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no %s yet — the plan stage must run first", ReelPlanFileName)
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var plan ReelPlan
	// Read back the way it was written, and for the same reason LoadSnippetPlan
	// gives: the plan stage already applied every template's standards with the
	// model in the room to answer for them, so re-judging the file here can only
	// reject the best answer anyone is going to get.
	if err := parseJSONLenient(string(raw), &plan, nil); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", ReelPlanFileName, err)
	}
	return &plan, nil
}

// ReelStages returns the stage list a reel runs: the same pipeline a snippet
// walks, keeping the verify stage only when some segment actually shows code.
//
// Per-reel rather than per-template because a reel is a mixture: one `vscode`
// segment among nine that show no code still has to have its code executed, and
// dropping verify because most segments do not need it would ship a clip
// claiming output nothing produced.
func ReelStages(l *project.Lesson) ([]string, error) {
	spec, err := LoadReelSpec(l.Dir)
	if err != nil {
		return nil, err
	}
	needsCode := false
	for _, s := range spec.Active() {
		tpl, ok := SnippetTemplates[s.Template]
		if !ok {
			return nil, fmt.Errorf("reel %s: unknown template %q (templates: %s)",
				l.ID, s.Template, strings.Join(SnippetTemplateNames(), ", "))
		}
		if tpl.NeedsCode {
			needsCode = true
		}
	}
	stages := slices.Clone(project.SnippetStageOrder)
	if !needsCode {
		stages = slices.DeleteFunc(stages, func(s string) bool { return s == project.StageVerify })
	}
	return stages, nil
}

// EnsureReelsCourse creates (or loads) the synthetic course reels live in.
func EnsureReelsCourse(root string) (*project.Course, error) {
	dir := filepath.Join(root, ReelsRoot)
	manifest := filepath.Join(dir, project.CourseFileName)
	if _, err := os.Stat(manifest); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Join(dir, "lessons"), 0o755); err != nil {
			return nil, fmt.Errorf("creating reels course: %w", err)
		}
		if err := writeFileAtomic(manifest, []byte(reelsCourseYAML)); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, fmt.Errorf("checking %s: %w", manifest, err)
	}
	return project.LoadCourse(dir)
}

// ListReels returns every reel, newest first.
func ListReels(root string) ([]*project.Lesson, error) {
	course, err := EnsureReelsCourse(root)
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
			continue // a half-written reel should not break the list
		}
		if IsReel(l) {
			out = append(out, l)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

// FindReel resolves a reel by id.
func FindReel(root, id string) (*project.Course, *project.Lesson, error) {
	course, err := EnsureReelsCourse(root)
	if err != nil {
		return nil, nil, err
	}
	l, err := project.LoadLesson(filepath.Join(course.Dir, "lessons", id))
	if err != nil {
		return nil, nil, fmt.Errorf("reel %q not found: %w", id, err)
	}
	return course, l, nil
}

// buildReelSceneGraph lays every segment onto one timeline.
//
// The whole assembly is the slice: each template is handed the spans covering
// its own beats — already absolute, already measured against the finished
// audio — and lays out scenes exactly as it would for a snippet. No template
// knows it is in a reel, nothing is offset, and a segment's internal timing is
// identical to what the same prompt would have produced on its own.
func buildReelSceneGraph(
	course *project.Course,
	l *project.Lesson,
	cfg config.Config,
	spec ReelSpec,
	plan *ReelPlan,
	alignment *Alignment,
	verification *VerificationReport,
	audioDurMs int,
) (*SceneGraph, error) {
	arch, err := ResolveArchetype(cfg.Style)
	if err != nil {
		return nil, err
	}

	graph := &SceneGraph{
		Theme:     videoThemeForConfig(cfg, course.Name),
		Motion:    arch.Motion,
		AudioFile: VoiceoverFileName,
	}
	if cfg.Style.Captions == "on" {
		graph.Captions = alignment.CaptionWords()
	}

	spans := alignment.CaptionSections()
	if want := plan.Beats(); len(spans) != want {
		return nil, fmt.Errorf("alignment has %d sections but the reel has %d beats across %d segments — re-run the align stage",
			len(spans), want, len(plan.Segments))
	}

	// A beat's visual holds until the next beat starts — including across a
	// segment boundary, which is what makes the cut between two templates land
	// on a word rather than on a gap. Only the very last beat of the reel runs
	// past its span, so the piece does not cut on the final syllable.
	ends := make([]int, len(spans))
	for i := range spans {
		if i+1 < len(spans) {
			ends[i] = spans[i+1].StartMs
		} else {
			ends[i] = spans[i].EndMs + videoTailMs
		}
	}
	if audioDurMs > 0 && len(ends) > 0 && ends[len(ends)-1] > audioDurMs+videoTailMs {
		ends[len(ends)-1] = audioDurMs + videoTailMs
	}

	verified := map[string]string{}
	if verification != nil {
		for _, b := range verification.Blocks {
			verified[project.HashBytes([]byte(b.Code))] = b.Stdout
		}
	}

	at := 0
	for _, seg := range plan.Segments {
		if seg.Plan == nil || len(seg.Plan.Beats) == 0 {
			continue
		}
		tpl, ok := SnippetTemplates[seg.Template]
		if !ok {
			return nil, fmt.Errorf("segment %q: unknown template %q", seg.ID, seg.Template)
		}
		if tpl.Scenes == nil {
			return nil, fmt.Errorf("segment %q: template %q lays out no scenes", seg.ID, seg.Template)
		}
		n := len(seg.Plan.Beats)
		scenes, err := tpl.Scenes(SnippetSceneInput{
			Spec:           seg.SegmentSpec(spec, cfg),
			Plan:           seg.Plan,
			Cfg:            cfg,
			Course:         course,
			Spans:          spans[at : at+n],
			BeatEndMs:      ends[at : at+n],
			VerifiedOutput: verified,
			DurationMs:     audioDurMs,
		})
		if err != nil {
			return nil, fmt.Errorf("segment %q (%s): %w", seg.ID, seg.Template, err)
		}
		graph.Scenes = append(graph.Scenes, scenes...)
		at += n
	}
	if len(graph.Scenes) == 0 {
		return nil, fmt.Errorf("the reel produced no scenes — every segment is skipped or empty")
	}
	graph.DurationMs = audioDurMs + videoTailMs
	return graph, nil
}

// SegmentSpec rebuilds the request shape a template's Scenes call expects.
func (s ReelPlanSegment) SegmentSpec(spec ReelSpec, cfg config.Config) SnippetSpec {
	return SnippetSpec{
		ID:       s.ID,
		Template: s.Template,
		Title:    spec.Title,
		Config:   cfg,
	}
}

// reelsCourseYAML is the synthetic course reels live in. Deliberately a
// separate course from snippets rather than a folder inside it: the two have
// different default runtimes (a reel segment is planned to its own template's
// default, but the piece as a whole is minutes long), and sharing a course
// would mean one config had to serve both.
const reelsCourseYAML = `name: Reels
slug: reels
description: Multi-template videos assembled from one narration.
style:
  tone: crisp, confident, direct
  audience: developers and learners watching a short video
  language: en
  # Matched to the snippets course: measured Kokoro af_heart delivery at
  # speed 1.0, so the pace target is where the voice already lives.
  pace_wpm: 174
pipeline:
  # Pinned rather than inherited, the same way the snippets course pins it.
  # A reel is one planning call per segment plus the cast, so the model is a
  # cost and quality decision this course should state rather than take from
  # whatever the global default happens to be.
  llm_content: openai/gpt-4o-mini
`
