package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// fakeTapeRunner scripts validate results and renders tiny real MP4s (so
// ffprobe can measure them).
type fakeTapeRunner struct {
	t            *testing.T
	validateErrs []error // popped per Validate call; empty → nil
	validated    []string
	rendered     []string
	// durationSec is how long the rendered clip is; 0 means 0.3s. A tool
	// capture's marks are placed against the *real* clip length, so a test
	// exercising them needs a clip at least as long as its tape's own time.
	durationSec float64
}

func (f *fakeTapeRunner) Name() string { return "fake-tapes" }

func (f *fakeTapeRunner) Validate(_ context.Context, workDir, tapePath string) error {
	f.validated = append(f.validated, tapePath)
	if _, err := os.Stat(filepath.Join(workDir, tapePath)); err != nil {
		return fmt.Errorf("tape not written: %v", err)
	}
	if len(f.validateErrs) == 0 {
		return nil
	}
	err := f.validateErrs[0]
	f.validateErrs = f.validateErrs[1:]
	return err
}

func (f *fakeTapeRunner) RenderTape(_ context.Context, workDir, tapePath string) error {
	f.rendered = append(f.rendered, tapePath)
	// Honor the tape's Output directive like vhs would.
	data, err := os.ReadFile(filepath.Join(workDir, tapePath))
	if err != nil {
		return err
	}
	var outRel string
	for _, line := range strings.Split(string(data), "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "Output "); ok {
			// vhs takes a quoted path, which is what the engine emits: an
			// unquoted absolute path is split on its separators and the tail
			// reported as an invalid command.
			outRel = strings.Trim(strings.TrimSpace(rest), `"`)
			break
		}
	}
	if outRel == "" {
		return fmt.Errorf("tape has no Output directive")
	}
	// Like vhs: a relative Output resolves against the working directory, an
	// absolute one is taken as written. A tool capture runs in a scratch dir
	// and writes its clip back into the lesson, so it uses the absolute form.
	out := outRel
	if !filepath.IsAbs(outRel) {
		out = filepath.Join(workDir, outRel)
	}
	dur := f.durationSec
	if dur == 0 {
		dur = 0.3
	}
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi",
		"-i", fmt.Sprintf("color=c=black:s=64x64:d=%g:r=10", dur),
		"-pix_fmt", "yuv420p", out)
	if b, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("fake render: %v\n%s", err, b)
	}
	return nil
}

const demoLessonBody = "## Try it\n- run things\n\n[DEMO: open the Python REPL and print a greeting]\n\n[DEMO: run a two-line script with python3]\n"

func TestExtractDemoMarkers(t *testing.T) {
	specs := extractDemoMarkers(demoLessonBody)
	if len(specs) != 2 {
		t.Fatalf("specs = %+v", specs)
	}
	if specs[0].ID != "demo-1" || specs[0].Description != "open the Python REPL and print a greeting" {
		t.Errorf("spec 0 = %+v", specs[0])
	}
	if specs[1].ID != "demo-2" {
		t.Errorf("spec 1 = %+v", specs[1])
	}
	if got := extractDemoMarkers("no markers here"); len(got) != 0 {
		t.Errorf("phantom markers: %+v", got)
	}
}

func TestLintTapeBody(t *testing.T) {
	valid := "Type \"python3\"\nEnter\nSleep 1s\nType \"print(1)\"\nEnter\nSleep 2s\n"
	tests := []struct {
		name    string
		body    string
		wantErr string
	}{
		{name: "valid body", body: valid},
		{name: "fenced body is cleaned", body: "```tape\n" + valid + "```"},
		{name: "echo forbidden", body: "Type \"echo 'fake output'\"\nEnter\nType \"python3\"\n", wantErr: "echo is forbidden"},
		{name: "no python", body: "Type \"ls\"\nEnter\n", wantErr: "never runs python3"},
		{name: "output directive", body: "Output foo.mp4\n" + valid, wantErr: "engine-owned"},
		{name: "set directive", body: "Set FontSize 12\n" + valid, wantErr: "engine-owned"},
		{name: "empty", body: "   ", wantErr: "never runs python3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lintTapeBody(tt.body)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(got, "```") || !strings.HasSuffix(got, "\n") {
				t.Errorf("cleaned body = %q", got)
			}
		})
	}
}

func TestTapeHeader(t *testing.T) {
	theme := deriveVideoTheme(config.Colors{Primary: "#306998", Accent: "#ffd43b", Background: "#ffffff"}, config.Fonts{}, "t", "")
	h := tapeHeader("demo-1", theme)
	for _, want := range []string{
		`Output "demos/demo-1.mp4"`,
		"Set FontSize 34",
		"Set Width 1440",
		"Set Height 640",
		"Set TypingSpeed 80ms",
		// The recording is themed to the derived dark tokens, never the
		// (light) course background.
		`"background": "` + theme.BgBottom + `"`,
		`"foreground": "` + theme.Text + `"`,
		`"cursor": "` + theme.Accent + `"`,
	} {
		if !strings.Contains(h, want) {
			t.Errorf("header missing %q:\n%s", want, h)
		}
	}
	// A zero theme still produces a usable dark header.
	bare := tapeHeader("demo-2", SceneTheme{})
	for _, want := range []string{`"background": "#11151c"`, `"foreground": "#f9fafb"`} {
		if !strings.Contains(bare, want) {
			t.Errorf("zero-theme header missing %q:\n%s", want, bare)
		}
	}
}

// tapeBody returns a lint-clean tape body.
func tapeBody() string {
	return "Type \"python3\"\nEnter\nSleep 1s\nType \"print('hi')\"\nEnter\nSleep 2s\nType \"exit()\"\nEnter\nSleep 3s"
}

// The command the fixture tape runs. Both flags are load-bearing: -p keeps it
// out of the interactive trust prompt, and --allowedTools is what lets it write
// a file at all. Its length feeds the mark-timing assertion below.
const toolTapeCommand = `claude -p 'add streaks' --allowedTools Write`

func toolTapeBody() string {
	return "Type \"" + toolTapeCommand + "\"\nEnter\n# MARK sent\nWait\n# MARK done\nSleep 3s"
}

// The tool path end to end: the tape runs in a scratch dir on the host, the
// clip lands back in the lesson, the tape is copied beside it for audit, and
// the footage sidecar records what really ran.
func TestCaptureStageRecordsAToolSession(t *testing.T) {
	requireFFmpeg(t)
	course, lesson := testCourse(t)
	lesson = lessonWithCode(t, lesson, "[CAPTURE: tool=claude; ask the agent to add streaks]\n")

	fake := &fakeRouter{content: []string{toolTapeBody()}}
	// The tape's own time is ~6.6s (44 characters plus Enter at 80ms, then the
	// closing Sleep 3s); the agent really took 9s, and the extra ~2.4s is the
	// Wait. That is the case the mark model exists to measure.
	tapes := &fakeTapeRunner{t: t, durationSec: 9}
	env, out := runEnv(t, fake)
	var sawTool captureTool
	var ranIn string
	env.ToolTapeRunner = func(_ context.Context, tool captureTool) (TapeRunner, error) {
		sawTool = tool
		return tapeRunnerFunc{tapes, func(workDir string) { ranIn = workDir }}, nil
	}

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageCapture}); err != nil {
		t.Fatal(err)
	}
	if sawTool.Binary != "claude" {
		t.Fatalf("resolved tool = %+v", sawTool)
	}

	// The recording must not happen inside the course tree — that is the
	// containment for an LLM-written tape running unsandboxed, and also why an
	// agent session does not see the course's own files.
	if ranIn == "" || strings.HasPrefix(ranIn, course.Dir) {
		t.Errorf("tool tape ran in %q, which is inside the course at %q", ranIn, course.Dir)
	}

	// The clip and the tape both land beside the lesson's other demos.
	genDir := filepath.Join(lesson.GeneratedDir(), DemosDirName)
	if _, err := os.Stat(filepath.Join(genDir, "capture-1.mp4")); err != nil {
		t.Errorf("clip not written back into the lesson: %v", err)
	}
	tape, err := os.ReadFile(filepath.Join(genDir, "capture-1.tape"))
	if err != nil {
		t.Fatalf("tape not kept for audit: %v", err)
	}
	if !strings.Contains(string(tape), toolTapeCommand) {
		t.Errorf("tape content:\n%s", tape)
	}
	if !strings.Contains(string(tape), filepath.Join(genDir, "capture-1.mp4")) {
		t.Errorf("a scratch-dir tape must write its Output to an absolute path:\n%s", tape)
	}

	var f Footage
	if !readJSONFile(filepath.Join(genDir, "capture-1"+FootageFileSuffix), &f) {
		t.Fatal("no footage sidecar was written beside the clip")
	}
	if f.Kind != CaptureKindTool || f.Tool != "claude" {
		t.Errorf("footage = %+v", f)
	}
	if f.CapturedAt == "" {
		t.Error("footage records no capture time, so nothing can tell when this clip goes stale")
	}
	if len(f.Marks) != 2 || f.Marks[0].Name != "sent" || f.Marks[1].Name != "done" {
		t.Errorf("marks = %+v", f.Marks)
	}
	if !f.Exact() {
		t.Errorf("a single Wait is measurable, so these marks should be exact: %+v", f.Marks)
	}
	// The mark before the Wait sits at its computed tape time; the one after it
	// carries the whole measured overrun, which is the point of the model.
	// The command is len(toolTapeCommand) characters, plus Enter, at 80ms each.
	wantBefore := (len(toolTapeCommand) + 1) * tapeTypingSpeedMs
	if f.Marks[0].AtMs != wantBefore {
		t.Errorf("mark before the wait = %dms, want %dms (%d chars + Enter at 80ms)", f.Marks[0].AtMs, wantBefore, len(toolTapeCommand))
	}
	wantAfter := wantBefore + (f.DurationMs - f.TapeTimeMs)
	if f.Marks[1].AtMs != wantAfter {
		t.Errorf("mark after the wait = %dms, want %dms (tape time + measured overrun)", f.Marks[1].AtMs, wantAfter)
	}
	if f.Marks[1].AtMs <= f.Marks[0].AtMs {
		t.Errorf("the wait moved nothing: marks = %+v", f.Marks)
	}

	manifest, err := loadCaptureManifest(lesson)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Demos) != 1 || manifest.Demos[0].Tool != "claude" {
		t.Errorf("manifest = %+v", manifest)
	}
	if !strings.Contains(out.String(), "on the host, authenticated and online") {
		t.Errorf("the unsandboxed warning was not printed:\n%s", out.String())
	}
}

// A lesson may mix both kinds, and they must not fight over ids or manifests.
func TestDemosStageMixesBothKinds(t *testing.T) {
	requireFFmpeg(t)
	course, lesson := testCourse(t)
	lesson = lessonWithCode(t, lesson, "[DEMO: run a script]\n\n[CAPTURE: tool=gh; list the pull requests]\n")

	// Capture runs first now, so its tape is the first reply.
	fake := &fakeRouter{content: []string{"Type \"gh pr list\"\nEnter\n# MARK listed\nSleep 2s", tapeBody()}}
	tapes := &fakeTapeRunner{t: t}
	env, _ := runEnv(t, fake)
	env.TapeRunner = tapes
	env.ToolTapeRunner = func(context.Context, captureTool) (TapeRunner, error) { return tapes, nil }

	// Captures record first, ahead of the script; demos record later, after
	// verify. The manifest the scene graph reads is both, in that order.
	for _, stage := range []string{project.StageCapture, project.StageDemos} {
		if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: stage}); err != nil {
			t.Fatalf("stage %s: %v", stage, err)
		}
	}
	manifest, err := loadDemoManifest(lesson)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Demos) != 2 {
		t.Fatalf("manifest = %+v", manifest)
	}
	if manifest.Demos[0].ID != "capture-1" || manifest.Demos[0].Kind != CaptureKindTool {
		t.Errorf("entry 0 = %+v", manifest.Demos[0])
	}
	if manifest.Demos[1].ID != "demo-1" || manifest.Demos[1].Kind != CaptureKindPython {
		t.Errorf("entry 1 = %+v", manifest.Demos[1])
	}
}

// tapeRunnerFunc wraps a runner to observe the working directory it was handed.
type tapeRunnerFunc struct {
	inner  TapeRunner
	record func(workDir string)
}

func (f tapeRunnerFunc) Name() string { return f.inner.Name() }
func (f tapeRunnerFunc) Validate(ctx context.Context, workDir, tapePath string) error {
	return f.inner.Validate(ctx, workDir, tapePath)
}
func (f tapeRunnerFunc) RenderTape(ctx context.Context, workDir, tapePath string) error {
	f.record(workDir)
	return f.inner.RenderTape(ctx, workDir, tapePath)
}

func TestDemosStageRecordsAndManifests(t *testing.T) {
	requireFFmpeg(t) // the fake runner renders real (tiny) MP4s
	course, lesson := testCourse(t)
	lesson = lessonWithCode(t, lesson, demoLessonBody)

	fake := &fakeRouter{content: []string{tapeBody(), tapeBody()}}
	tapes := &fakeTapeRunner{t: t}
	env, _ := runEnv(t, fake)
	env.TapeRunner = tapes

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageDemos}); err != nil {
		t.Fatal(err)
	}

	if len(tapes.rendered) != 2 {
		t.Fatalf("rendered = %v", tapes.rendered)
	}
	// The written tape carries the engine header + LLM body.
	tape, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), DemosDirName, "demo-1.tape"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(tape), `Output "demos/demo-1.mp4"`) || !strings.Contains(string(tape), `Type "python3"`) {
		t.Errorf("tape content:\n%s", tape)
	}

	manifest, err := loadDemoManifest(lesson)
	if err != nil {
		t.Fatal(err)
	}
	if len(manifest.Demos) != 2 {
		t.Fatalf("manifest = %+v", manifest)
	}
	entry := manifest.Demos[0]
	if entry.ID != "demo-1" || entry.Path != filepath.Join(DemosDirName, "demo-1.mp4") {
		t.Errorf("entry = %+v", entry)
	}
	if entry.DurationMs < 200 || entry.DurationMs > 500 {
		t.Errorf("duration = %dms, want ~300", entry.DurationMs)
	}
	if entry.Description != "open the Python REPL and print a greeting" {
		t.Errorf("description = %q", entry.Description)
	}
}

func TestDemosStageValidateRetry(t *testing.T) {
	requireFFmpeg(t)
	course, lesson := testCourse(t)
	lesson = lessonWithCode(t, lesson, "[DEMO: one demo]\n")

	fake := &fakeRouter{content: []string{tapeBody(), tapeBody()}}
	tapes := &fakeTapeRunner{t: t, validateErrs: []error{fmt.Errorf("syntax error on line 3")}}
	env, out := runEnv(t, fake)
	env.TapeRunner = tapes

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageDemos}); err != nil {
		t.Fatal(err)
	}
	if len(fake.contentReqs) != 2 {
		t.Fatalf("content requests = %d, want 2 (regeneration after invalid tape)", len(fake.contentReqs))
	}
	if !strings.Contains(fake.contentReqs[1].Messages[1].Content, "syntax error on line 3") {
		t.Errorf("regeneration prompt missing validator error:\n%s", fake.contentReqs[1].Messages[1].Content)
	}
	if !strings.Contains(out.String(), "regenerating") {
		t.Errorf("output:\n%s", out.String())
	}
}

func TestDemosStageNoMarkers(t *testing.T) {
	course, lesson := testCourse(t)
	lesson = lessonWithCode(t, lesson, "## Only ideas\n- no demos declared\n")
	env, out := runEnv(t, &fakeRouter{}) // no runner needed, no LLM calls
	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageDemos}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "no [DEMO] markers") {
		t.Errorf("output:\n%s", out.String())
	}
}

func TestDemosStageNeedsRunner(t *testing.T) {
	course, lesson := testCourse(t)
	lesson = lessonWithCode(t, lesson, "[DEMO: something]\n")
	env, _ := runEnv(t, &fakeRouter{})

	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageDemos})
	if err == nil || !strings.Contains(err.Error(), "docker build") {
		t.Errorf("error = %v, want sandbox instructions", err)
	}
}

func TestDemoManifestRoundTrip(t *testing.T) {
	_, lesson := testCourse(t)
	m, err := loadDemoManifest(lesson)
	if err != nil || len(m.Demos) != 0 {
		t.Fatalf("missing manifest should be empty: %+v, %v", m, err)
	}
	want := DemoManifest{Demos: []DemoEntry{{ID: "demo-1", Description: "d", Path: "demos/demo-1.mp4", DurationMs: 1234}}}
	if err := os.MkdirAll(filepath.Join(lesson.GeneratedDir(), DemosDirName), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(lesson.GeneratedDir(), DemosDirName, DemoManifestFileName), want); err != nil {
		t.Fatal(err)
	}
	got, err := loadDemoManifest(lesson)
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := json.Marshal(got); string(b) != mustJSON(t, want) {
		t.Errorf("round trip = %+v", got)
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
