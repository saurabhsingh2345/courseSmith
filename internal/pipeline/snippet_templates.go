package pipeline

// The snippet template catalog.
//
// A template is the answer to "what does this clip look like?" — it owns the
// prompt that plans the clip, the rules that plan must satisfy, and the
// mapping from a planned-and-timed clip onto renderer scenes. Adding a
// template means adding one file here and one Remotion component; nothing in
// the pipeline needs to know the catalog grew.
//
// The split of responsibility is deliberate: the *shared* code owns timing,
// theming, captions, and the scene-graph envelope, so no template can get
// those wrong or drift from the design system. A template only decides what
// fills the frame.

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// SnippetTemplate is one entry in the catalog.
type SnippetTemplate struct {
	// Name is the id used on the CLI and in snippet.yaml.
	Name string
	// Title and Description are the gallery copy shown in the studio.
	Title       string
	Description string
	// Example is a prompt that shows this template at its best; the studio
	// offers it as a starting point.
	Example string
	// PromptFile is the prompt template rendered to plan a clip.
	PromptFile string
	// NeedsCode makes the verify stage part of this template's pipeline, so
	// any code shown is code that really ran.
	NeedsCode bool

	// Plan produces the clip's design. Nil uses planSnippetDefault, which
	// renders PromptFile and decodes a SnippetPlan — enough for every
	// template whose plan fits the standard shape.
	Plan func(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config) (*SnippetPlan, error)
	// Validate enforces this template's own rules on a plan, and rejects
	// beat fields the template does not own.
	Validate func(p *SnippetPlan) error
	// Scenes maps the planned, timed clip onto renderer scenes.
	Scenes func(in SnippetSceneInput) ([]Scene, error)
	// PromptData contributes extra top-level fields to the prompt beyond the
	// shared set — a template's own vocabularies and bounds. Keys collide at
	// the template author's peril; the shared keys are listed in
	// sharedPromptData.
	PromptData func(spec SnippetSpec, cfg config.Config) map[string]any
}

// SnippetTemplates is the catalog, keyed by name. Templates register
// themselves from their own files' init().
var SnippetTemplates = map[string]*SnippetTemplate{}

// registerSnippetTemplate adds a template to the catalog.
func registerSnippetTemplate(t *SnippetTemplate) {
	if _, dup := SnippetTemplates[t.Name]; dup {
		panic("duplicate snippet template " + t.Name)
	}
	SnippetTemplates[t.Name] = t
}

// SnippetTemplateNames returns the catalog's names, sorted.
func SnippetTemplateNames() []string {
	out := make([]string, 0, len(SnippetTemplates))
	for name := range SnippetTemplates {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// SnippetTemplateList returns the catalog sorted by name, for the studio
// gallery and `coursesmith snippet --list`.
func SnippetTemplateList() []*SnippetTemplate {
	out := make([]*SnippetTemplate, 0, len(SnippetTemplates))
	for _, name := range SnippetTemplateNames() {
		out = append(out, SnippetTemplates[name])
	}
	return out
}

// SnippetSceneInput is everything a template needs to lay its clip out on the
// timeline. Timing is already resolved: spans are real, measured word
// timings from the aligner, not estimates.
type SnippetSceneInput struct {
	Spec SnippetSpec
	Plan *SnippetPlan
	Cfg  config.Config
	// Course is the synthetic snippets course (branding, name).
	Course *project.Course
	// Spans are the aligned section spans, one per beat, in plan order.
	Spans []SectionSpan
	// BeatEndMs is the ms at which each beat's visual should give way to the
	// next — the next beat's start, or the padded end of the audio.
	BeatEndMs []int
	// Verification maps a code block's hash to what running it really
	// printed; empty for templates that do not show code.
	VerifiedOutput map[string]string
	// DurationMs is the finished clip's length.
	DurationMs int
}

// Beat returns beat i with its resolved timing.
func (in SnippetSceneInput) Beat(i int) (b SnippetBeat, startMs, endMs int) {
	return in.Plan.Beats[i], in.Spans[i].StartMs, in.BeatEndMs[i]
}

// sharedPromptData is the field set every template's prompt can rely on. A
// template adds its own with SnippetTemplate.PromptData.
//
// It is a map rather than a struct so templates can extend it; prompts still
// render with missingkey=error, so a typo in a template is a build failure
// rather than a silently empty instruction.
func sharedPromptData(spec SnippetSpec, cfg config.Config) map[string]any {
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	target := spec.ResolvedTargetSec()
	wantWords, minWords, maxWords := wordBudget(target, pace)
	return map[string]any{
		"Prompt":          spec.Prompt,
		"Title":           spec.Title,
		"TargetSec":       target,
		"TargetWords":     wantWords,
		"MinWords":        minWords,
		"MaxWords":        maxWords,
		"MinWordsPerBeat": minWordsPerBeat,
		"MaxWordsPerBeat": maxWordsPerBeat,
		"Tone":            cfg.Style.Tone,
		"Audience":        cfg.Style.Audience,
		"Language":        cfg.Style.Language,
		"CodeLanguage":    spec.ResolvedCodeLanguage(),
		"PaceWPM":         pace,
		"MinBeats":        minSnippetBeats,
		"MaxBeats":        maxSnippetBeats,
	}
}

// Beat-count bounds. Fewer than three and the clip is a single held shot;
// more than seven in under three minutes and nothing lands.
const (
	minSnippetBeats = 3
	maxSnippetBeats = 7
)

// Per-beat narration bounds. Under ten words a beat is a caption, not a
// thought; over sixty it outlasts any single visual.
const (
	minWordsPerBeat = 10
	maxWordsPerBeat = 60
)

// snippetPlanRepairRounds is how many correction attempts a plan gets. Replies
// are cached, so the cost lands once per distinct prompt.
const snippetPlanRepairRounds = 3

// wordBudget is how much narration a clip of the requested length needs.
//
// Models systematically under-write to a seconds target — they have no clock —
// so the budget is enforced, not suggested: a plan outside the band is
// rejected and regenerated with the miss quoted back.
//
// The band is asymmetric because the two failures are not symmetric. Coming in
// short breaks the clip: the visuals are timed to the voice, so half the
// narration is half a video. Running long only makes it longer than asked —
// every beat still lands, the viewer just gets more of them.
//
// The ceiling is deliberately loose. A tighter one (135%) was tried and did not
// hold: on a topic the model judged to need ~185 words it produced 184-185 in
// three consecutive correction rounds, ignoring the stated target entirely.
// Models write to what the content seems to need, not to a word count, and no
// amount of restating the budget changed that. Rejecting those plans bought a
// failed generation rather than a shorter clip.
//
// So the target is treated as a target: it steers the draft, the floor is
// enforced because undershooting is fatal, and the runtime a creator picks is
// documented as approximate. The finished duration is always reported.
func wordBudget(targetSec, paceWPM int) (target, minWords, maxWords int) {
	target = targetSec * paceWPM / 60
	return target, target * 75 / 100, target * 155 / 100
}

// narrationWords counts the words the voice will actually speak.
func narrationWords(p *SnippetPlan) int {
	n := 0
	for _, b := range p.Beats {
		n += len(strings.Fields(b.Narration))
	}
	return n
}

// planSnippetDefault renders the template's prompt and decodes the reply.
func planSnippetDefault(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config) (*SnippetPlan, error) {
	tpl := SnippetTemplates[spec.Template]
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	target := spec.ResolvedTargetSec()
	wantWords, minWords, maxWords := wordBudget(target, pace)
	data := sharedPromptData(spec, cfg)
	if tpl.PromptData != nil {
		for k, v := range tpl.PromptData(spec, cfg) {
			data[k] = v
		}
	}
	system, user, err := e.renderPrompt(tpl.PromptFile, data)
	if err != nil {
		return nil, err
	}
	var plan SnippetPlan
	// A plan has more independent numeric rules than anything else the pipeline
	// asks for — beat count, per-beat words, total words, and whatever the
	// template adds on top. One correction round is not enough to land them all.
	err = e.completeJSONRounds(ctx, cfg.Pipeline, llm.TaskContent, system, user, nil, 0.5, 6144, snippetPlanRepairRounds, &plan, func() error {
		plan.Template = spec.Template // so Validate dispatches to this template
		if err := plan.Validate(); err != nil {
			return err
		}
		if n := narrationWords(&plan); n < minWords || n > maxWords {
			return fmt.Errorf(
				"narration totals %d words but a %ds clip needs %d-%d (aim for %d) — rewrite with fuller sentences, do not add beats past %d",
				n, target, minWords, maxWords, wantWords, maxSnippetBeats)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("planning %s snippet: %w", spec.Template, err)
	}
	return &plan, nil
}

// checkBeatShape is the shared structural rule every template applies: how
// many beats, and how much narration each one carries.
func checkBeatShape(p *SnippetPlan) error {
	if n := len(p.Beats); n < minSnippetBeats || n > maxSnippetBeats {
		return fmt.Errorf("plan has %d beats, want %d-%d", n, minSnippetBeats, maxSnippetBeats)
	}
	for _, b := range p.Beats {
		n := len(strings.Fields(b.Narration))
		if n < minWordsPerBeat || n > maxWordsPerBeat {
			return fmt.Errorf("beat %q has %d words of narration, want %d-%d",
				b.ID, n, minWordsPerBeat, maxWordsPerBeat)
		}
	}
	return nil
}

// beatFields names the optional SnippetBeat fields a template consumes.
//
// SnippetBeat is the union of what every template needs, so a field one
// template owns is meaningless to the others. Declaring ownership once per
// template — rather than each template hand-checking the others' fields, which
// is quadratic and rots as the catalog grows — means a model that puts a
// whiteboard sketch on a flow diagram gets a loud error instead of silence.
type beatFields struct {
	Code   bool
	Run    bool
	Sketch bool
	Nodes  bool
	Focus  bool
	Art    bool
}

// rejectForeignBeatFields fails when a beat sets a field its template does not
// own. Adding a field to SnippetBeat means adding one case here.
func rejectForeignBeatFields(p *SnippetPlan, owned beatFields) error {
	for _, b := range p.Beats {
		var set string
		switch {
		case !owned.Code && b.Code != "":
			set = "code"
		case !owned.Run && b.Run:
			set = "run"
		case !owned.Sketch && len(b.Sketch) > 0:
			set = "sketch"
		case !owned.Nodes && len(b.Nodes) > 0:
			set = "nodes"
		case !owned.Focus && len(b.Focus) > 0:
			set = "focus"
		case !owned.Art && b.Art != nil:
			set = "art"
		default:
			continue
		}
		return fmt.Errorf("beat %q sets %s, which the %s template does not use", b.ID, set, p.Template)
	}
	return nil
}

// buildSnippetSceneGraph assembles the renderer input for a snippet: the
// shared envelope (theme, motion, captions, duration) plus whatever scenes the
// template lays out.
func buildSnippetSceneGraph(
	course *project.Course,
	l *project.Lesson,
	cfg config.Config,
	spec SnippetSpec,
	plan *SnippetPlan,
	alignment *Alignment,
	verification *VerificationReport,
	audioDurMs int,
) (*SceneGraph, error) {
	tpl, ok := SnippetTemplates[spec.Template]
	if !ok {
		return nil, fmt.Errorf("unknown template %q", spec.Template)
	}
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
	if len(spans) != len(plan.Beats) {
		return nil, fmt.Errorf("alignment has %d sections but the plan has %d beats — re-run the align stage", len(spans), len(plan.Beats))
	}

	// A beat's visual holds until the next beat starts; the last one runs to
	// the padded end of the audio so the clip does not cut on the final word.
	ends := make([]int, len(spans))
	for i := range spans {
		if i+1 < len(spans) {
			ends[i] = spans[i+1].StartMs
		} else {
			ends[i] = spans[i].EndMs + videoTailMs
		}
	}

	verified := map[string]string{}
	if verification != nil {
		for _, b := range verification.Blocks {
			verified[project.HashBytes([]byte(b.Code))] = b.Stdout
		}
	}

	scenes, err := tpl.Scenes(SnippetSceneInput{
		Spec:           spec,
		Plan:           plan,
		Cfg:            cfg,
		Course:         course,
		Spans:          spans,
		BeatEndMs:      ends,
		VerifiedOutput: verified,
		DurationMs:     max(audioDurMs, ends[len(ends)-1]),
	})
	if err != nil {
		return nil, fmt.Errorf("template %s: %w", tpl.Name, err)
	}
	if len(scenes) == 0 {
		return nil, fmt.Errorf("template %s produced no scenes", tpl.Name)
	}

	graph.Scenes = scenes
	graph.DurationMs = max(audioDurMs, scenes[len(scenes)-1].EndMs)
	graph.Scenes[len(graph.Scenes)-1].EndMs = graph.DurationMs
	for i := range graph.Scenes {
		if graph.Scenes[i].Props == nil {
			graph.Scenes[i].Props = map[string]any{}
		}
		if _, set := graph.Scenes[i].Props["template"]; !set {
			if v := arch.TemplateFor(graph.Scenes[i].Type); v != "" {
				graph.Scenes[i].Props["template"] = v
			}
		}
	}
	return graph, nil
}

// snippetFileName is the filename shown in a code-bearing snippet's editor
// chrome: derived from the title so it reads like a real file in a real
// project rather than "main.py" every time.
func snippetFileName(title, language string) string {
	ext := map[string]string{
		"python": "py", "javascript": "js", "typescript": "ts", "go": "go",
		"rust": "rs", "java": "java", "ruby": "rb", "sql": "sql", "bash": "sh",
	}[strings.ToLower(language)]
	if ext == "" {
		ext = "txt"
	}
	// Short enough to sit in the editor's file tree without ellipsis: the
	// first couple of words of the title, not the whole thing.
	words := strings.Split(slugify(title), "-")
	if len(words) > 2 {
		words = words[:2]
	}
	base := strings.Join(words, "_")
	if len(base) > 16 {
		base = strings.Trim(base[:16], "_")
	}
	if base == "" {
		base = "main"
	}
	return base + "." + ext
}
