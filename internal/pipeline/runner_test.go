package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// fakeRouter serves queued responses per task type and records requests.
type fakeRouter struct {
	content     []string
	review      []string
	contentReqs []llm.Request
	reviewReqs  []llm.Request
}

func (f *fakeRouter) Complete(_ context.Context, _ config.Pipeline, task llm.TaskType, req llm.Request) (*llm.Response, error) {
	switch task {
	case llm.TaskContent:
		f.contentReqs = append(f.contentReqs, req)
		if len(f.content) == 0 {
			return nil, fmt.Errorf("fake router: unexpected content request #%d", len(f.contentReqs))
		}
		body := f.content[0]
		f.content = f.content[1:]
		return &llm.Response{Content: body}, nil
	case llm.TaskReview, llm.TaskVision:
		// Vision QA shares the review queue here: the router falls back to the
		// review model when no dedicated vision model is set, and the tests
		// queue their QA verdicts as review responses.
		f.reviewReqs = append(f.reviewReqs, req)
		if len(f.review) == 0 {
			return nil, fmt.Errorf("fake router: unexpected review request #%d", len(f.reviewReqs))
		}
		body := f.review[0]
		f.review = f.review[1:]
		return &llm.Response{Content: body}, nil
	default:
		return nil, fmt.Errorf("fake router: unknown task %q", task)
	}
}

const testCourseYAML = `name: Test Course
slug: test-course
style:
  tone: warm teacher
  # Matches the fake aligner's ~250 wpm so auto-pace stays quiet and the
  # re-run no-op invariant holds (auto-pace has its own unit tests).
  pace_wpm: 250
branding:
  colors:
    primary: "#306998"
pipeline:
  review_threshold: 8
`

// testCourse builds a full course on disk with the standard test lesson.
func testCourse(t *testing.T) (*project.Course, *project.Lesson) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "test-course")
	lessonDir := filepath.Join(dir, "lessons", "01-test")
	if err := os.MkdirAll(lessonDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, project.CourseFileName), []byte(testCourseYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lessonDir, project.LessonFileName), []byte(testLessonMD), 0o644); err != nil {
		t.Fatal(err)
	}
	course, err := project.LoadCourse(dir)
	if err != nil {
		t.Fatal(err)
	}
	lesson, err := course.FindLesson("01")
	if err != nil {
		t.Fatal(err)
	}
	return course, lesson
}

func scriptJSON(narration string) string {
	return fmt.Sprintf(`{"title":"Test Lesson","sections":[`+
		`{"id":"first-idea","narration":%q,"duration_est_sec":10,"cues":[{"type":"diagram","ref":"memory-model","at_word":2}]},`+
		`{"id":"second-idea","narration":"Now let us try it live.","duration_est_sec":8,"cues":[{"type":"demo","ref":"showing the thing live","at_word":3}]}]}`,
		narration)
}


// storyboardBody is a minimal valid storyboard reply for the two-section
// test script.
func storyboardBody() string {
	return `{"sections":[{"id":"first-idea","points":[{"text":"One idea","icon":"idea","at_word":1},{"text":"Real output","icon":"terminal","at_word":2}]},{"id":"second-idea","points":[{"text":"Try it live","icon":"play","at_word":3}]}]}`
}

// emphasisBody is an empty caption-emphasis reply.
func emphasisBody() string {
	return `{"indices":[]}`
}

func reviewJSON(overall float64, critique string) string {
	return fmt.Sprintf(
		`{"scores":{"technical_accuracy":%[1]g,"clarity":%[1]g,"engagement":%[1]g,"pacing":%[1]g},"overall":%[1]g,"critique":%[2]q}`,
		overall, critique)
}

// Canned responses for the three-pass review stage, in consumption order:
// claims extraction, accuracy, pedagogy, tone.
func claimsJSON() string {
	return `{"claims":[
		{"claim":"Python reads code line by line","section":"first-idea","checkable":false},
		{"claim":"7 times 6 is 42","section":"first-idea","checkable":true,"code":"assert 7 * 6 == 42"}]}`
}

func accuracyJSON(score float64, critique string) string {
	return fmt.Sprintf(`{"score":%g,"verdicts":[{"claim":"Python reads code line by line","verdict":"correct","citation":"the Python tutorial describes sequential execution"}],"critique":%q}`, score, critique)
}

func pedagogyJSON(score float64, critique string) string {
	return fmt.Sprintf(`{"scores":{"concept_ordering":%[1]g,"cognitive_load":%[1]g,"concrete_before_abstract":%[1]g,"worked_examples":%[1]g},"score":%[1]g,"critique":%[2]q}`, score, critique)
}

func toneJSON(score float64, critique string) string {
	return fmt.Sprintf(`{"score":%g,"critique":%q}`, score, critique)
}

// multipassReview queues one full passing (or failing) review round.
func multipassReview(a, p, tn float64) []string {
	return []string{
		claimsJSON(),
		accuracyJSON(a, "accuracy critique"),
		pedagogyJSON(p, "pedagogy critique"),
		toneJSON(tn, "tone critique"),
	}
}

func svgBody(label string) string {
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 400">` +
		`<style>svg{--primary:#306998;--accent:#ffd43b;--bg:#ffffff;}</style>` +
		`<rect width="800" height="400" fill="var(--bg)"/>` +
		`<text x="40" y="60" font-size="20" font-family="sans-serif">` + label + `</text></svg>`
}

func quizBody() string {
	q := func(n int, qtype string) string {
		return fmt.Sprintf(`{"id":"q%d","type":%q,"prompt":"Question %d?","options":["a","b","c","d"],"answer_index":1,"explanation":"Because b."}`, n, qtype, n)
	}
	return `{"title":"Check your understanding","questions":[` +
		q(1, "recall") + `,` + q(2, "application") + `,` + q(3, "debugging") + `,` + q(4, "prediction") + `]}`
}

// distractorsJSON scores every wrong option of quizBody() at the given
// plausibility.
func distractorsJSON(plausibility float64) string {
	var qs []string
	for n := 1; n <= 4; n++ {
		var ds []string
		for _, idx := range []int{0, 2, 3} { // answer_index is 1
			ds = append(ds, fmt.Sprintf(`{"option_index":%d,"plausibility":%g,"misconception":"a common mix-up"}`, idx, plausibility))
		}
		qs = append(qs, fmt.Sprintf(`{"id":"q%d","distractors":[%s]}`, n, strings.Join(ds, ",")))
	}
	return `{"questions":[` + strings.Join(qs, ",") + `]}`
}

// difficultyJSON reports the same correct-count for all four questions.
func difficultyJSON(correct int) string {
	var qs []string
	for n := 1; n <= 4; n++ {
		qs = append(qs, fmt.Sprintf(`{"id":"q%d","correct":%d,"reasoning":"most get it"}`, n, correct))
	}
	return `{"questions":[` + strings.Join(qs, ",") + `]}`
}

// quizReviewQueue is the review responses one clean quiz stage run
// consumes: rubric gate, distractor scoring, difficulty simulation.
func quizReviewQueue() []string {
	return []string{reviewJSON(9, "Fair quiz."), distractorsJSON(8), difficultyJSON(7)}
}

// runStages runs a sequence of single stages, failing the test on error.
func runStages(t *testing.T, env *Env, course *project.Course, lesson *project.Lesson, stages ...string) {
	t.Helper()
	for _, s := range stages {
		if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: s}); err != nil {
			t.Fatalf("stage %s: %v", s, err)
		}
	}
}

// runEnv wires a fake router, test prompts, and an output buffer.
func runEnv(t *testing.T, fake *fakeRouter) (*Env, *bytes.Buffer) {
	t.Helper()
	out := &bytes.Buffer{}
	return &Env{Router: fake, PromptsDir: writeTestPrompts(t), Out: out}, out
}

func readScript(t *testing.T, l *project.Lesson) Script {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(l.GeneratedDir(), ScriptFileName))
	if err != nil {
		t.Fatal(err)
	}
	var s Script
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatalf("script.json does not parse: %v", err)
	}
	return s
}

func TestScriptStageWritesScript(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{content: []string{scriptJSON("Welcome to the lesson everyone.")}}
	env, _ := runEnv(t, fake)

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageScript})
	if err != nil {
		t.Fatal(err)
	}

	script := readScript(t, lesson)
	if script.Title != "Test Lesson" || len(script.Sections) != 2 {
		t.Errorf("script = %+v", script)
	}

	req := fake.contentReqs[0]
	if !req.JSONMode {
		t.Error("script request did not use JSON mode")
	}
	if req.Messages[0].Role != llm.RoleSystem || !strings.Contains(req.Messages[0].Content, "warm teacher") {
		t.Errorf("system prompt did not render course tone: %q", req.Messages[0].Content)
	}
	if !strings.Contains(req.Messages[1].Content, "memory-model") {
		t.Errorf("user prompt did not list declared diagrams: %q", req.Messages[1].Content)
	}

	cfg := config.Resolve(course.Config, lesson.FrontMatter.Overrides(), config.Config{})
	statuses, err := env.LessonStatus(lesson, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if statuses[project.StageScript] != project.StatusDone {
		t.Errorf("script status = %s, want done", statuses[project.StageScript])
	}
}

func TestScriptStageRetriesOnBadJSON(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{content: []string{"sorry, here is your script:", scriptJSON("Take two works fine.")}}
	env, _ := runEnv(t, fake)

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageScript})
	if err != nil {
		t.Fatal(err)
	}
	if len(fake.contentReqs) != 2 {
		t.Fatalf("content requests = %d, want 2 (original + correction)", len(fake.contentReqs))
	}
	retry := fake.contentReqs[1]
	if len(retry.Messages) != 4 {
		t.Fatalf("retry has %d messages, want 4 (system, user, assistant, correction)", len(retry.Messages))
	}
	if !strings.Contains(retry.Messages[3].Content, "rejected") {
		t.Errorf("correction message = %q, want the parse error surfaced", retry.Messages[3].Content)
	}
	if got := readScript(t, lesson).Sections[0].Narration; got != "Take two works fine." {
		t.Errorf("narration = %q, want the corrected draft", got)
	}
}

func TestScriptStageFailsAfterRetry(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{content: []string{"not json", "still not json"}}
	env, _ := runEnv(t, fake)

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageScript})
	if err == nil || !strings.Contains(err.Error(), "invalid after retry") {
		t.Fatalf("error = %v, want invalid-after-retry", err)
	}

	cfg := config.Resolve(course.Config, lesson.FrontMatter.Overrides(), config.Config{})
	statuses, err := env.LessonStatus(lesson, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if statuses[project.StageScript] != project.StatusPending {
		t.Errorf("failed stage recorded as %s, want pending", statuses[project.StageScript])
	}
}

func TestScriptStageRejectsUndeclaredDiagram(t *testing.T) {
	course, lesson := testCourse(t)
	bad := strings.ReplaceAll(scriptJSON("Hello there."), "memory-model", "not-a-diagram")
	fake := &fakeRouter{content: []string{bad, bad}}
	env, _ := runEnv(t, fake)

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageScript})
	if err == nil || !strings.Contains(err.Error(), "undeclared diagram") {
		t.Fatalf("error = %v, want undeclared-diagram validation failure", err)
	}
}

func TestReviewStagePassesFirstRound(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{
		content: []string{scriptJSON("Original draft narration.")},
		review:  multipassReview(9, 9, 9),
	}
	env, out := runEnv(t, fake)
	runStages(t, env, course, lesson, project.StageScript, project.StageReview)

	recordPath := filepath.Join(lesson.GeneratedDir(), ReviewsDirName, "script-multipass-round-1.json")
	data, err := os.ReadFile(recordPath)
	if err != nil {
		t.Fatalf("review record not persisted: %v", err)
	}
	var record MultiReviewRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	if !record.Passed || record.Round != 1 || record.Weighted != 9 {
		t.Errorf("record = %+v", record)
	}
	if record.Accuracy.Score != 9 || record.Pedagogy.Score != 9 || record.Tone.Score != 9 {
		t.Errorf("pass scores = %v %v %v", record.Accuracy.Score, record.Pedagogy.Score, record.Tone.Score)
	}
	if len(record.Claims) != 2 {
		t.Errorf("claims = %+v, want the 2 extracted", record.Claims)
	}
	if got := readScript(t, lesson).Sections[0].Narration; got != "Original draft narration." {
		t.Errorf("narration = %q, passing draft must be kept unchanged", got)
	}
	if !strings.Contains(out.String(), "passed") {
		t.Errorf("output missing pass notice:\n%s", out.String())
	}
	if len(fake.contentReqs) != 1 {
		t.Errorf("content requests = %d, want 1 (no regeneration on pass)", len(fake.contentReqs))
	}
	if len(fake.reviewReqs) != 4 {
		t.Errorf("review requests = %d, want 4 (claims + three passes)", len(fake.reviewReqs))
	}
}

func TestReviewStageExecutesCheckableClaims(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{
		content: []string{scriptJSON("Original draft narration.")},
		review:  multipassReview(9, 9, 9),
	}
	env, _ := runEnv(t, fake)
	runner := &fakeRunner{}
	env.CodeRunner = runner
	runStages(t, env, course, lesson, project.StageScript, project.StageReview)

	if runner.calls != 1 {
		t.Errorf("sandbox calls = %d, want 1 (the single checkable claim)", runner.calls)
	}
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), ReviewsDirName, "script-multipass-round-1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var record MultiReviewRecord
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	var executed *ClaimResult
	for i := range record.Claims {
		if record.Claims[i].Checkable {
			executed = &record.Claims[i]
		}
	}
	if executed == nil || !executed.Executed || executed.Held == nil || !*executed.Held {
		t.Errorf("checkable claim not executed: %+v", record.Claims)
	}
	// The scoring prompt saw the execution outcome.
	scoring := fake.reviewReqs[1]
	if !strings.Contains(scoring.Messages[1].Content, "EXECUTED") {
		t.Errorf("accuracy prompt missing execution results:\n%s", scoring.Messages[1].Content)
	}
}

func TestReviewLoopRegeneratesWithAllCritiques(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{
		content: []string{scriptJSON("Original draft narration."), scriptJSON("Improved draft with example.")},
		review: append(
			// Round 1: accuracy 6 → weighted 6×.5 + 9×.35 + 9×.15 = 7.5 < 8.
			multipassReview(6, 9, 9),
			// Round 2: all 9 → passes.
			multipassReview(9, 9, 9)...,
		),
	}
	env, _ := runEnv(t, fake)
	runStages(t, env, course, lesson, project.StageScript, project.StageReview)

	regen := fake.contentReqs[1]
	for _, critique := range []string{"accuracy critique", "pedagogy critique", "tone critique"} {
		if !strings.Contains(regen.Messages[1].Content, critique) {
			t.Errorf("regeneration prompt missing %q:\n%s", critique, regen.Messages[1].Content)
		}
	}
	if got := readScript(t, lesson).Sections[0].Narration; got != "Improved draft with example." {
		t.Errorf("narration = %q, want the regenerated draft", got)
	}
	for round := 1; round <= 2; round++ {
		p := filepath.Join(lesson.GeneratedDir(), ReviewsDirName, fmt.Sprintf("script-multipass-round-%d.json", round))
		if _, err := os.Stat(p); err != nil {
			t.Errorf("round %d record missing: %v", round, err)
		}
	}
}

func TestReviewKeepsBestDraftWhenAllRoundsFail(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{
		content: []string{
			scriptJSON("Best draft."),
			scriptJSON("Worse draft."),
			scriptJSON("Worst draft."),
		},
		review: append(append(
			multipassReview(7, 8, 8), // weighted 7.5
			multipassReview(6, 8, 8)...), // weighted 7.0
			multipassReview(5, 8, 8)..., // weighted 6.5
		),
	}
	env, out := runEnv(t, fake)
	runStages(t, env, course, lesson, project.StageScript, project.StageReview)

	if got := readScript(t, lesson).Sections[0].Narration; got != "Best draft." {
		t.Errorf("narration = %q, want the highest-scoring draft kept", got)
	}
	if !strings.Contains(out.String(), "⚠") {
		t.Errorf("output missing quality warning:\n%s", out.String())
	}

	cfg := config.Resolve(course.Config, lesson.FrontMatter.Overrides(), config.Config{})
	statuses, err := env.LessonStatus(lesson, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if statuses[project.StageReview] != project.StatusDone {
		t.Errorf("review status = %s, want done (soft gate)", statuses[project.StageReview])
	}
}

// TestRunnerFullPipelineAndSkip runs all seven implemented stages end-to-end
// (real ffmpeg, fake LLM/TTS/Whisper), then proves a re-run is a no-op.
func TestRunnerFullPipelineAndSkip(t *testing.T) {
	requireFFmpeg(t)
	if _, err := findFont(); err != nil {
		t.Skip("no system font available for slides mode")
	}
	course, lesson := testCourse(t)

	// Fake Kokoro: any speech request returns half a second of silence.
	tts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(makeWAV(0.5))
	}))
	defer tts.Close()

	fake := &fakeRouter{
		// script → diagram → quiz → mistakes → exercises → demo tape (the
		// test lesson has one [DEMO] marker), in stage order.
		content: []string{
			scriptJSON("Original draft narration."), storyboardBody(), svgBody("Memory"),
			quizBody(), mistakesBody(), exercisesBody(), tapeBody(),
		},
		review: append(append(append(
			multipassReview(9, 9, 9), // script three-pass review
			reviewJSON(9, "Clear diagram.")),
			quizReviewQueue()...), // quiz rubric + distractors + difficulty
			emphasisBody(), emphasisBody(), // caption emphasis, one per section
		),
	}
	env, _ := runEnv(t, fake)
	env.TTSBaseURL = tts.URL
	env.SiteDir = filepath.Join(t.TempDir(), "site")
	env.CodeRunner = failingRunner() // broken samples fail, solutions pass
	env.TapeRunner = &fakeTapeRunner{t: t}
	// Word-level alignment matching both sections' narration, within the 1s
	// of fake audio and with no compressible gaps.
	env.Aligner = &fakeAligner{words: append(
		wordSeq(0, "Original", "draft", "narration."),
		wordSeq(800, "Now", "let", "us", "try", "it", "live.")...,
	)}

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{}); err != nil {
		t.Fatal(err)
	}
	if len(fake.content) != 0 || len(fake.review) != 0 {
		t.Fatalf("full run left %d content / %d review responses unconsumed", len(fake.content), len(fake.review))
	}
	for _, artifact := range []string{ScriptFileName, QuizFileName, VoiceoverFileName, CaptionsFileName, FinalVideoName} {
		if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), artifact)); err != nil {
			t.Errorf("full run did not produce %s: %v", artifact, err)
		}
	}
	bundle := filepath.Join(env.SiteDir, "content", "courses", "test-course", "01-test")
	if _, err := os.Stat(filepath.Join(bundle, "index.md")); err != nil {
		t.Errorf("full run did not emit the Hugo page: %v", err)
	}

	// Re-run with an empty router and no TTS/aligner: any external call
	// would error, so success proves every stage was skipped.
	env2 := &Env{Router: &fakeRouter{}, PromptsDir: env.PromptsDir, Out: &bytes.Buffer{}, TTSBaseURL: "http://127.0.0.1:1", SiteDir: env.SiteDir}
	if err := env2.RunLesson(context.Background(), course, lesson, RunOptions{}); err != nil {
		t.Fatalf("re-run was not fully skipped: %v", err)
	}
	out := env2.Out.(*bytes.Buffer).String()
	if !strings.Contains(out, "up to date") {
		t.Errorf("output missing skip notices:\n%s", out)
	}
	if got := strings.Count(out, "up to date"); got != len(stageFuncs) {
		t.Errorf("skipped %d stages, want all %d implemented:\n%s", got, len(stageFuncs), out)
	}
}

func TestRunnerForceReruns(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{
		content: []string{scriptJSON("First run.")},
		review:  multipassReview(9, 9, 9),
	}
	env, _ := runEnv(t, fake)
	runStages(t, env, course, lesson, project.StageScript, project.StageReview)

	fake.content = []string{scriptJSON("Forced re-run.")}
	fake.review = multipassReview(9, 9, 9)
	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageScript, Force: true}); err != nil {
		t.Fatal(err)
	}
	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageReview, Force: true}); err != nil {
		t.Fatal(err)
	}
	if got := readScript(t, lesson).Sections[0].Narration; got != "Forced re-run." {
		t.Errorf("narration = %q, want the forced re-run output", got)
	}
}

func TestRunnerRejectsUnknownStage(t *testing.T) {
	course, lesson := testCourse(t)
	env, _ := runEnv(t, &fakeRouter{})
	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: "bogus"})
	if err == nil || !strings.Contains(err.Error(), `unknown stage "bogus"`) {
		t.Errorf("error = %v, want unknown-stage error", err)
	}
}

func TestReviewStageRequiresScript(t *testing.T) {
	course, lesson := testCourse(t)
	env, _ := runEnv(t, &fakeRouter{})
	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageReview})
	if err == nil || !strings.Contains(err.Error(), "script stage must run first") {
		t.Errorf("error = %v, want script-first error", err)
	}
}
