package pipeline

// Importing a plan somebody wrote by hand.
//
// Every other way into this pipeline goes through a planning call: a prompt and
// a template become a plan, and the plan becomes a video. That is the right
// default and it is the only door, which means the pipeline cannot be used at
// all by somebody who has already decided what the video says.
//
// Two cases want that, and they are not exotic. The first is a remake — a
// reference clip in front of you, a shot list you have already worked out, and
// nothing left for a writer to decide. The second is no key: the renderer, the
// voice and the aligner all run on this machine, and the plan stage is the only
// thing in the run that reaches for a provider. Refusing to start because of one
// stage, whose output is a small JSON file a person can write, is a hard
// dependency on a model to type something you already know.
//
// So this is the same door with the writer taken out. The plan arrives finished;
// everything that would have judged the model's draft still runs — each
// segment's Normalize, its Validate, the word budget, and a dry run of its scene
// layout — because a hand-written plan is exactly as capable of being wrong as a
// generated one, and it deserves the same error messages rather than a crash six
// stages later in the renderer.
//
// The substance and plan stages are then marked done with the hashes the ordinary
// path would have recorded, so the run that follows starts at audio and every
// later re-run skips them exactly as it would after a real planning call.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// ComboImportSegment is one segment of an import document: the request fields
// a combo.yaml segment carries, and the finished plan for it.
type ComboImportSegment struct {
	ID       string `json:"id,omitempty"`
	Template string `json:"template"`
	// Prompt is what this segment covers. Not read by anything on this path —
	// nothing is being planned — and required anyway, because it is what a
	// re-plan of one segment would be made from and what a person reads six
	// months later to find out why the segment exists.
	Prompt    string       `json:"prompt,omitempty"`
	Heading   string       `json:"heading,omitempty"`
	Role      string       `json:"role,omitempty"`
	Material  string       `json:"material,omitempty"`
	TargetSec int          `json:"target_sec,omitempty"`
	Plan      *SnippetPlan `json:"plan"`
}

// ComboImport is one file holding a whole hand-authored piece: the request and
// every segment's finished plan.
//
// One document rather than a combo.yaml plus a plan file, because the two are
// written in one sitting and matched by position — splitting them means the
// commonest authoring mistake is a silent misalignment between two files
// instead of an error inside one.
type ComboImport struct {
	ID       string               `json:"id,omitempty"`
	Title    string               `json:"title,omitempty"`
	Brief    string               `json:"brief,omitempty"`
	Angle    string               `json:"angle,omitempty"`
	Segments []ComboImportSegment `json:"segments"`
}

// LoadComboImport reads an import document.
func LoadComboImport(path string) (*ComboImport, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading the plan: %w", err)
	}
	var doc ComboImport
	dec := json.NewDecoder(bytes.NewReader(data))
	// Strict, unlike the plan the model's reply is decoded from. A model that
	// invents a field needs a correction round; a person who misspells one
	// needs to be told, or the field they meant to set stays at its default and
	// the video is quietly not the one they wrote.
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filepath.Base(path), err)
	}
	if len(doc.Segments) == 0 {
		return nil, fmt.Errorf("%s has no segments", filepath.Base(path))
	}
	return &doc, nil
}

// Spec is the combo.yaml this document describes.
func (d *ComboImport) Spec() ComboSpec {
	spec := ComboSpec{ID: d.ID, Title: d.Title, Brief: d.Brief, Angle: d.Angle}
	for _, s := range d.Segments {
		prompt := strings.TrimSpace(s.Prompt)
		if prompt == "" && s.Plan != nil {
			// The spec validator requires one, and a plan always carries
			// something better than an empty string.
			prompt = s.Plan.Title
		}
		spec.Segments = append(spec.Segments, ComboSegment{
			ID:        s.ID,
			Template:  s.Template,
			Prompt:    prompt,
			Heading:   s.Heading,
			Role:      s.Role,
			Material:  s.Material,
			TargetSec: s.TargetSec,
		})
	}
	return spec
}

// Plan is the combo-plan.json this document describes.
func (d *ComboImport) Plan() *ComboPlan {
	plan := &ComboPlan{Title: d.Title}
	for _, s := range d.Segments {
		plan.Segments = append(plan.Segments, ComboPlanSegment{
			ID:       s.ID,
			Template: s.Template,
			Plan:     s.Plan,
		})
	}
	return plan
}

// ImportComboPlan writes a hand-authored plan into a combo's directory and
// leaves it as though the plan stage had produced it: combo-plan.json,
// script.json, lesson.md, and a state file that says planning is done.
//
// It never calls a model. Every check the planner applies to a model's reply is
// applied here instead, so an imported plan that renders badly fails now, while
// the fix is one edit to a JSON file, rather than in the scenegraph stage where
// the message is about a nil field.
func (e *Env) ImportComboPlan(course *project.Course, l *project.Lesson, spec *ComboSpec, plan *ComboPlan) error {
	// Resolved here rather than passed in: every stage that follows resolves it
	// the same way, and a caller that assembles its own is a caller whose plan is
	// checked against a different pace than it is spoken at.
	cfg := config.Resolve(course.Config, l.FrontMatter.Overrides(), spec.Config)
	if plan == nil || len(plan.Segments) == 0 {
		return fmt.Errorf("the imported plan has no segments")
	}
	active := spec.Active()
	if len(plan.Segments) != len(active) {
		return fmt.Errorf("the plan has %d segments and %s has %d in the cut — a segment with no plan is a hole in the timeline",
			len(plan.Segments), ComboFileName, len(active))
	}

	if plan.Title == "" {
		plan.Title = spec.Title
	}
	for i := range plan.Segments {
		seg := active[i]
		ps := &plan.Segments[i]
		// The id and template come from combo.yaml rather than from the plan
		// file. Both places carrying them is how they disagree, and combo.yaml
		// is the one every later edit addresses.
		if ps.ID != "" && ps.ID != seg.ID {
			return fmt.Errorf("plan segment %d is %q but %s calls it %q — they are matched in order, so a mismatch means the file is out of step",
				i, ps.ID, ComboFileName, seg.ID)
		}
		if ps.Template != "" && ps.Template != seg.Template {
			return fmt.Errorf("plan segment %q is %s but %s casts it as %s",
				seg.ID, ps.Template, ComboFileName, seg.Template)
		}
		ps.ID, ps.Template = seg.ID, seg.Template
		if ps.Plan == nil {
			return fmt.Errorf("plan segment %q has no plan", seg.ID)
		}
		if err := e.checkImportedSegment(cfg, spec, seg, ps.Plan); err != nil {
			return fmt.Errorf("segment %q (%s): %w", seg.ID, seg.Template, err)
		}
	}
	if plan.Title == "" && plan.Segments[0].Plan != nil {
		plan.Title = plan.Segments[0].Plan.Title
	}

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
	reloaded, err := project.LoadLesson(l.Dir)
	if err != nil {
		return fmt.Errorf("the imported combo does not load (bug): %w", err)
	}
	*l = *reloaded

	// Substance as well as plan. Substance is the stage ahead of planning and it
	// is the other one that needs a provider; an imported plan carries its own
	// facts, so leaving it pending would stop the very next command for the
	// reason this path exists to avoid.
	if err := e.markStagesDone(l, cfg, project.StageSubstance, project.StagePlan); err != nil {
		return err
	}

	words := 0
	for _, seg := range plan.Segments {
		words += narrationWords(seg.Plan)
	}
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	fmt.Fprintf(e.out(), "  ✓ plan      imported — %q, %d segments, %d beats, %d words (~%ds at %d wpm)\n",
		plan.Title, len(plan.Segments), plan.Beats(), words, words*60/pace, pace)
	return nil
}

// checkImportedSegment puts one hand-written segment through everything the
// planning loop would have put a model's reply through.
func (e *Env) checkImportedSegment(cfg config.Config, spec *ComboSpec, seg ComboSegment, p *SnippetPlan) error {
	tpl, ok := SnippetTemplates[seg.Template]
	if !ok {
		return fmt.Errorf("unknown template (templates: %s)", strings.Join(SnippetTemplateNames(), ", "))
	}
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	target := seg.ResolvedTargetSec()
	wantWords, minWords, maxWords := wordBudget(target, pace)

	p.Template = seg.Template
	// The budget the segment is scored against, exactly as the planner sets it
	// before validating a reply. Without it beatBounds falls back to the
	// fixture range and a segment written to a 60-second budget is judged
	// against nothing in particular.
	p.targetWords = wantWords
	normalizeSnippetPlan(p)

	segSpec := seg.SnippetSpec(cfg, spec.Brief, nil)
	if tpl.PreValidate != nil {
		tpl.PreValidate(segSpec, p)
	}
	if err := p.Validate(); err != nil {
		return err
	}
	if n := narrationWords(p); n < minWords || n > maxWords {
		return fmt.Errorf("narration totals %d words but a %ds segment needs %d-%d (aim for %d) — change the words or change target_sec in %s",
			n, target, minWords, maxWords, wantWords, ComboFileName)
	}
	// The layout, run for real against fabricated timings. This is what catches
	// a plan that satisfies every rule and still cannot be drawn.
	if err := dryRunSnippetScenes(segSpec, cfg, p); err != nil {
		return fmt.Errorf("the segment validates but does not lay out: %w", err)
	}
	return nil
}

// markStagesDone records stages as complete with the input hashes they have
// right now, which is what runStages does after running one for real.
func (e *Env) markStagesDone(l *project.Lesson, cfg config.Config, stages ...string) error {
	state, err := l.LoadState()
	if err != nil {
		return err
	}
	now := time.Now()
	for _, name := range stages {
		inputs, err := e.StageInputs(l, cfg, name)
		if err != nil {
			return fmt.Errorf("stage %s: %w", name, err)
		}
		state.MarkDone(name, inputs, now)
	}
	return l.SaveState(state)
}
