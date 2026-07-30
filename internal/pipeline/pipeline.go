// Package pipeline orchestrates the per-lesson generation stages.
//
// Every stage reads inputs from the lesson dir, writes outputs to
// lessons/<id>/generated/, and records its input hashes in state.json.
// A stage whose recorded hashes match the current inputs is skipped, which
// makes runs idempotent and resumable.
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// Completer is the slice of llm.Router the pipeline needs; it is an
// interface so tests can substitute a fake.
type Completer interface {
	Complete(ctx context.Context, pcfg config.Pipeline, task llm.TaskType, req llm.Request) (*llm.Response, error)
}

// Env carries everything stages need to run. Router may be nil for
// operations that never call an LLM (e.g. status reporting).
type Env struct {
	Router     Completer
	PromptsDir string
	// CoursePromptsDir is an optional per-course prompt override dir
	// (courses/<slug>/prompts, populated by archetypes); templates found
	// there win over PromptsDir. RunLesson sets it from the course.
	CoursePromptsDir string
	// Out receives human-readable progress; nil discards it.
	Out io.Writer

	// TTSBaseURL is the Kokoro server's OpenAI-compatible API root;
	// "" uses DefaultTTSBaseURL.
	TTSBaseURL string
	// Transcriber provides speech-to-text; the align stage uses it as the
	// segment-level fallback when whisperX (Aligner) is unavailable.
	Transcriber Transcriber
	// Aligner provides word-level timestamps (whisperX); nil falls back to
	// Transcriber-based estimates with a quality warning.
	Aligner Aligner
	// CodeRunner executes lesson code blocks (verify stage, quiz checks).
	CodeRunner CodeRunner
	// TapeRunner records VHS terminal demos; nil makes the demos stage fail
	// with instructions when a lesson declares [DEMO] markers.
	TapeRunner TapeRunner
	// Renderer produces final.mp4 from the scene graph (Remotion); nil
	// falls back to the legacy ffmpeg assembly with a warning.
	Renderer VideoRenderer
	// Screenshotter rasterizes SVGs for vision QA; nil falls back to
	// source-text diagram review.
	Screenshotter Screenshotter
	// FFmpeg is the ffmpeg binary; "" resolves "ffmpeg" from PATH.
	FFmpeg string
	// HTTPClient is used for TTS requests; nil uses a long-timeout default.
	HTTPClient *http.Client
	// SiteDir is the Hugo site root the hugo stage emits into;
	// "" uses DefaultSiteDir.
	SiteDir string
}

func (e *Env) out() io.Writer {
	if e.Out == nil {
		return io.Discard
	}
	return e.Out
}

// artifactAbsent marks an input file that does not exist yet. Using a
// sentinel (instead of omitting the key) means the stage goes stale the
// moment the file appears.
const artifactAbsent = "absent"

// stageUpstream lists generated artifacts each stage consumes.
var stageUpstream = map[string][]string{
	// Plan consumes the fact sheet, so re-establishing the facts re-plans — and
	// through plan, re-renders. That chain is the point: correcting a figure in
	// substance.json must reach the screen, not just the file.
	project.StagePlan:   {SubstanceFileName},
	project.StageVerify: {ScriptFileName},
	// Trace re-runs when verified code changes; it reads blocks from lesson.md
	// (always a hashed input) but depends on verification to run after verify.
	project.StageTrace:        {VerificationFileName},
	project.StageReview:       {ScriptFileName, VerificationFileName},
	project.StageStoryboard:   {ScriptFileName},
	project.StageQuiz:         {ScriptFileName, VerificationFileName},
	project.StageQuizStrategy: {QuizFileName},
	project.StageMistakes:     {ScriptFileName, VerificationFileName},
	project.StageExercises:    {ScriptFileName, VerificationFileName},
	project.StageDemos:        {VerificationFileName},
	// tts_fixes.json is written by the align stage's WER gate; its
	// appearance makes audio stale so the next run re-synthesizes with the
	// pronunciation fixes applied.
	project.StageAudio:    {ScriptFileName, TTSFixesFileName, TTSSpeedFileName},
	project.StageAlign:    {VoiceoverFileName, ScriptFileName, TTSScriptFileName},
	project.StageCaptions: {AlignmentFileName},
	project.StageChapters: {AlignmentFileName, ScriptFileName},
	project.StageScenegraph: {
		ScriptFileName, AlignmentFileName, VerificationFileName,
		DemosDirName + "/" + DemoManifestFileName,
		CodeTracesDirName + "/" + TraceManifestFileName,
		CaptionEmphasisFileName,
		StoryboardFileName,
	},
	project.StageRender: {
		SceneGraphFileName, VoiceoverFileName, CaptionsFileName,
		DemosDirName + "/" + DemoManifestFileName,
	},
	project.StageHugo: {ScriptFileName, QuizFileName, CaptionsFileName, FinalVideoName},
}

// stageLessonFiles lists files in the lesson dir itself (not generated/)
// that a stage consumes. A recording appearing or changing re-triggers
// render; editing quiz overrides re-publishes the page.
var stageLessonFiles = map[string][]string{
	// A snippet's whole input is its request file; editing the prompt or
	// swapping the template re-plans (and so re-renders) the clip.
	project.StagePlan:       {SnippetFileName, ReelFileName},
	// The same input as plan: substance is established from the request, so
	// editing the prompt or the brief re-establishes the facts — and therefore
	// re-plans, since plan consumes substance.json.
	project.StageSubstance:  {SnippetFileName, ReelFileName},
	project.StageScenegraph: {VideoPlanFileName},
	project.StageRender:     {RecordingFileName},
	project.StageHugo:       {QuizOverridesFileName},
}

// stageTemplates lists prompt templates each stage renders; editing a
// template invalidates the stages that use it. Stages that gate their
// output through the critic also depend on the review rubric.
var stageTemplates = map[string][]string{
	project.StageSubstance: {substanceTemplateName, substanceSearchTemplateName},
	project.StageScript:    {scriptTemplateName},
	// Review runs three passes (claims+accuracy, pedagogy, tone) and
	// re-invokes the script generator on regeneration.
	project.StageReview: {
		reviewClaimsTemplateName, reviewAccuracyTemplateName,
		reviewPedagogyTemplateName, reviewToneTemplateName, scriptTemplateName,
	},
	project.StageVisuals:    {diagramTemplateName, d3DiagramTemplateName, d2LangTemplateName, mermaidDiagramTemplateName, excalidrawTemplateName, diagramVisualQATemplateName, reviewTemplateName},
	project.StageQuiz:       {quizTemplateName, reviewTemplateName, quizDistractorsTemplateName, quizDifficultyTemplateName},
	project.StageMistakes:   {mistakesTemplateName},
	project.StageExercises:  {exercisesTemplateName},
	project.StageDemos:      {demoTapeTemplateName},
	project.StageCaptions:   {captionEmphasisTemplateName},
	project.StageStoryboard: {storyboardTemplateName},
}

// stageFunc is one pipeline stage implementation.
type stageFunc func(ctx context.Context, e *Env, course *project.Course, l *project.Lesson, cfg config.Config) error

// stageFuncs maps implemented stages to their implementations. Stages absent
// from this map (still being built out in Phase 2) are reported as skipped.
var stageFuncs = map[string]stageFunc{
	project.StageSubstance:    runSubstanceStage,
	project.StagePlan:         runPlanStage,
	project.StageScript:       runScriptStage,
	project.StageVerify:       runVerifyStage,
	project.StageTrace:        runTraceStage,
	project.StageReview:       runReviewStage,
	project.StageStoryboard:   runStoryboardStage,
	project.StageVisuals:      runVisualsStage,
	project.StageQuiz:         runQuizStage,
	project.StageQuizStrategy: runQuizStrategyStage,
	project.StageMistakes:     runMistakesStage,
	project.StageExercises:    runExercisesStage,
	project.StageDemos:        runDemosStage,
	project.StageAudio:        runAudioStage,
	project.StageAlign:        runAlignStage,
	project.StageCaptions:     runCaptionsStage,
	project.StageChapters:     runChaptersStage,
	project.StageScenegraph:   runScenegraphStage,
	project.StageRender:       runRenderStage,
	project.StageHugo:         runHugoStage,
}

// StageInputs computes the current input-hash map for a stage: lesson.md,
// the resolved config, the stage's prompt templates, and any upstream
// artifacts it consumes.
func (e *Env) StageInputs(l *project.Lesson, cfg config.Config, stage string) (map[string]string, error) {
	inputs := map[string]string{}

	srcHash, err := project.HashFile(l.SourcePath())
	if err != nil {
		return nil, err
	}
	inputs["lesson.md"] = srcHash

	// VideoOnly selects which stages run; it never changes what any stage
	// produces. Excluding it from the fingerprint means toggling it doesn't
	// mark every already-rendered lesson stale.
	fp := cfg
	fp.Pipeline.VideoOnly = false
	cfgData, err := yaml.Marshal(fp)
	if err != nil {
		return nil, fmt.Errorf("fingerprinting config: %w", err)
	}
	inputs["config"] = project.HashBytes(cfgData)

	for _, name := range stageTemplates[stage] {
		inputs["prompts/"+name] = hashOrAbsent(filepath.Join(e.PromptsDir, name))
		if e.CoursePromptsDir != "" {
			inputs["course-prompts/"+name] = hashOrAbsent(filepath.Join(e.CoursePromptsDir, name))
		}
	}
	for _, name := range stageUpstream[stage] {
		inputs[name] = hashOrAbsent(filepath.Join(l.GeneratedDir(), name))
	}
	for _, name := range stageLessonFiles[stage] {
		inputs[name] = hashOrAbsent(filepath.Join(l.Dir, name))
	}
	if stage == project.StageScript {
		// Unresolved human review notes are a script input: new notes make
		// the stage stale; the stage marks them resolved before the runner
		// records post-run hashes, so state stabilizes again.
		notes, err := LoadReviewNotes(courseDirOf(l.Dir))
		if err != nil {
			return nil, err
		}
		if text := notes.UnresolvedText(l.ID); text != "" {
			inputs["review-notes"] = project.HashBytes([]byte(text))
		} else {
			inputs["review-notes"] = artifactAbsent
		}
	}
	if stage == project.StageHugo || stage == project.StageRender {
		// Both embed every declared diagram (page inline / DiagramScene).
		for _, d := range l.FrontMatter.Diagrams {
			rel := DiagramsDirName + "/" + d.ID + ".svg"
			inputs[rel] = hashOrAbsent(filepath.Join(l.GeneratedDir(), DiagramsDirName, d.ID+".svg"))
		}
	}
	if stage == project.StageVisuals {
		// Style exemplars are few-shot inputs to diagram generation.
		for i, ex := range loadExemplars(e.PromptsDir) {
			inputs[fmt.Sprintf("prompts/%s/%d", diagramStyleDirName, i)] = project.HashBytes([]byte(ex))
		}
	}
	if stage == project.StagePlan {
		// Which prompt template the plan stage renders depends on which
		// template the request names, so it cannot be listed statically:
		// editing snippet_vscode.tmpl must re-plan vscode snippets only.
		if spec, err := LoadSnippetSpec(l.Dir); err == nil {
			if tpl, ok := SnippetTemplates[spec.Template]; ok {
				inputs["prompts/"+tpl.PromptFile] = hashOrAbsent(filepath.Join(e.PromptsDir, tpl.PromptFile))
			}
		}
	}
	return inputs, nil
}

// hashOrAbsent hashes a file, mapping a missing file to the absent sentinel.
// Unreadable-but-present files also map to the sentinel; the stage that
// actually needs the file will surface the real error when it reads it.
func hashOrAbsent(path string) string {
	h, err := project.HashFile(path)
	if err != nil {
		return artifactAbsent
	}
	return h
}

// LessonStatus resolves the status of every lesson stage.
func (e *Env) LessonStatus(l *project.Lesson, cfg config.Config) (map[string]project.StageStatus, error) {
	return e.LessonStatusFor(l, cfg, project.StageOrder)
}

// LessonStatusFor resolves the status of an explicit stage list.
//
// Which stages are worth reporting depends on what the thing is: a snippet
// never runs the script or visuals stages, so reporting them as forever
// "pending" would be noise, and its own `plan` stage would be missing.
func (e *Env) LessonStatusFor(l *project.Lesson, cfg config.Config, stages []string) (map[string]project.StageStatus, error) {
	state, err := l.LoadState()
	if err != nil {
		return nil, err
	}
	out := make(map[string]project.StageStatus, len(stages))
	for _, stage := range stages {
		inputs, err := e.StageInputs(l, cfg, stage)
		if err != nil {
			return nil, fmt.Errorf("stage %s: %w", stage, err)
		}
		out[stage] = state.StageStatus(stage, inputs)
	}
	return out, nil
}

// RunOptions controls a pipeline run.
type RunOptions struct {
	// Stage restricts the run to a single named stage ("" runs all).
	Stage string
	// Force re-runs stages even when their inputs are unchanged.
	Force bool
}

// RunLesson executes the pipeline for one lesson, skipping up-to-date
// stages. State is saved after every completed stage, so an interrupted run
// resumes where it stopped.
func (e *Env) RunLesson(ctx context.Context, course *project.Course, l *project.Lesson, opts RunOptions) error {
	cfg := config.Resolve(course.Config, l.FrontMatter.Overrides(), config.Config{})

	// Course-level prompt overrides (archetypes): a copy keeps the caller's
	// Env reusable across courses.
	if e.CoursePromptsDir == "" && course != nil {
		if dir := filepath.Join(course.Dir, "prompts"); hashOrAbsent(filepath.Join(dir, scriptTemplateName)) != artifactAbsent || dirExists(dir) {
			scoped := *e
			scoped.CoursePromptsDir = dir
			e = &scoped
		}
	}

	stages := project.StagesFor(cfg.Pipeline.VideoOnly)
	fmt.Fprintf(e.out(), "%s — %q\n", l.ID, l.FrontMatter.Title)
	return e.runStages(ctx, course, l, cfg, stages, opts)
}

// runStages is the shared execution loop behind RunLesson and RunSnippet:
// walk the given stages in order, skip the ones whose inputs are unchanged,
// and record state after each so an interrupted run resumes cleanly.
func (e *Env) runStages(
	ctx context.Context,
	course *project.Course,
	l *project.Lesson,
	cfg config.Config,
	stages []string,
	opts RunOptions,
) error {
	if opts.Stage != "" {
		// An explicit --stage may name any stage, including one this run mode
		// would otherwise skip — asking for it by name is opting in.
		if !slices.Contains(project.AllStages(), opts.Stage) {
			return fmt.Errorf("unknown stage %q (stages: %s)", opts.Stage, strings.Join(project.AllStages(), ", "))
		}
		stages = []string{opts.Stage}
	}

	state, err := l.LoadState()
	if err != nil {
		return err
	}

	for _, name := range stages {
		if err := ctx.Err(); err != nil {
			return err
		}
		fn, implemented := stageFuncs[name]
		if !implemented {
			fmt.Fprintf(e.out(), "  ○ %-9s not implemented yet — skipped\n", name)
			continue
		}
		inputs, err := e.StageInputs(l, cfg, name)
		if err != nil {
			return fmt.Errorf("lesson %s, stage %s: %w", l.ID, name, err)
		}
		if !opts.Force && state.StageStatus(name, inputs) == project.StatusDone {
			fmt.Fprintf(e.out(), "  ✓ %-9s up to date — skipped\n", name)
			continue
		}

		if err := fn(ctx, e, course, l, cfg); err != nil {
			return fmt.Errorf("lesson %s, stage %s: %w", l.ID, name, err)
		}
		// Recompute inputs after the run: a stage may rewrite an artifact it
		// also consumes (review rewrites script.json), and the recorded
		// hashes must describe what it actually left on disk.
		post, err := e.StageInputs(l, cfg, name)
		if err != nil {
			return fmt.Errorf("lesson %s, stage %s: %w", l.ID, name, err)
		}
		state.MarkDone(name, post, time.Now())
		if err := l.SaveState(state); err != nil {
			return fmt.Errorf("lesson %s: %w", l.ID, err)
		}
		fmt.Fprintf(e.out(), "  ✓ %-9s done\n", name)
	}
	return nil
}

// RunCourse runs the pipeline for every lesson in the course, in order.
// A quota exhaustion stops the run cleanly; everything completed so far is
// recorded and a re-run resumes from the first unfinished stage.
func (e *Env) RunCourse(ctx context.Context, course *project.Course, opts RunOptions) error {
	lessons, err := course.Lessons()
	if err != nil {
		return err
	}
	for i, l := range lessons {
		if i > 0 {
			fmt.Fprintln(e.out())
		}
		if err := e.RunLesson(ctx, course, l, opts); err != nil {
			if quota, ok := errors.AsType[*llm.QuotaError](err); ok {
				fmt.Fprintf(e.out(), "\nStopping: %v\n", quota)
			}
			return err
		}
	}
	return nil
}
