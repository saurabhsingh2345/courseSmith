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
			outRel = strings.TrimSpace(rest)
			break
		}
	}
	if outRel == "" {
		return fmt.Errorf("tape has no Output directive")
	}
	out := filepath.Join(workDir, outRel)
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "color=c=black:s=64x64:d=0.3:r=10",
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
	theme := deriveVideoTheme(config.Colors{Primary: "#306998", Accent: "#ffd43b", Background: "#ffffff"}, config.Fonts{}, "t")
	h := tapeHeader("demo-1", theme)
	for _, want := range []string{
		"Output demos/demo-1.mp4",
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
	if !strings.Contains(string(tape), "Output demos/demo-1.mp4") || !strings.Contains(string(tape), `Type "python3"`) {
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
