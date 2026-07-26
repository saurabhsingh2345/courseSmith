package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/project"
)

type fakeScreenshotter struct {
	calls int
	err   error
}

func (f *fakeScreenshotter) Name() string { return "fake-screens" }

func (f *fakeScreenshotter) ScreenshotSVG(_ context.Context, svg []byte) ([]byte, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return []byte("PNG-of-" + string(svg[:20])), nil
}

func visualVerdictJSON(passed bool, issues ...string) string {
	b, _ := json.Marshal(visualVerdict{Passed: passed, Issues: issues})
	return string(b)
}

func TestVisualQAPassesFirstRound(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{
		content: []string{svgBody("v1")},
		review:  []string{visualVerdictJSON(true)},
	}
	shots := &fakeScreenshotter{}
	env, out := runEnv(t, fake)
	env.Screenshotter = shots

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVisuals}); err != nil {
		t.Fatal(err)
	}
	if shots.calls != 1 {
		t.Errorf("screenshots = %d, want 1", shots.calls)
	}
	// The review request must carry the screenshot as an image.
	if len(fake.reviewReqs) != 1 || len(fake.reviewReqs[0].Images) != 1 {
		t.Fatalf("review requests = %+v", fake.reviewReqs)
	}
	// Attempt artifacts are kept for audit.
	for _, name := range []string{"memory-model-round-1.svg", "memory-model-round-1.png"} {
		if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), DiagramsDirName, AttemptsDirName, name)); err != nil {
			t.Errorf("attempt artifact missing: %v", err)
		}
	}
	if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), ReviewsDirName, "diagram-memory-model-visual-round-1.json")); err != nil {
		t.Errorf("visual record missing: %v", err)
	}
	if !strings.Contains(out.String(), "visual QA passed") {
		t.Errorf("output:\n%s", out.String())
	}
}

func TestVisualQARegeneratesOnIssues(t *testing.T) {
	course, lesson := testCourse(t)
	issue := "label overlaps the box outline"
	fake := &fakeRouter{
		content: []string{svgBody("v1"), svgBody("v2")},
		review:  []string{visualVerdictJSON(false, issue), visualVerdictJSON(true)},
	}
	env, _ := runEnv(t, fake)
	env.Screenshotter = &fakeScreenshotter{}

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVisuals}); err != nil {
		t.Fatal(err)
	}
	if len(fake.contentReqs) != 2 {
		t.Fatalf("content requests = %d, want 2", len(fake.contentReqs))
	}
	if !strings.Contains(fake.contentReqs[1].Messages[1].Content, issue) {
		t.Errorf("regeneration prompt missing visual critique:\n%s", fake.contentReqs[1].Messages[1].Content)
	}
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), DiagramsDirName, "memory-model.svg"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "v2") {
		t.Errorf("final svg = %q, want regenerated version", data)
	}
	// Both rounds' attempts audited.
	for round := 1; round <= 2; round++ {
		p := filepath.Join(lesson.GeneratedDir(), DiagramsDirName, AttemptsDirName, fmt.Sprintf("memory-model-round-%d.svg", round))
		if _, err := os.Stat(p); err != nil {
			t.Errorf("round %d attempt missing: %v", round, err)
		}
	}
}

func TestVisualQAKeepsLastAfterMaxRounds(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{
		content: []string{svgBody("v1"), svgBody("v2"), svgBody("v3")},
		review: []string{
			visualVerdictJSON(false, "issue a"),
			visualVerdictJSON(false, "issue b"),
			visualVerdictJSON(false, "issue c"),
		},
	}
	env, out := runEnv(t, fake)
	env.Screenshotter = &fakeScreenshotter{}

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVisuals}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), DiagramsDirName, "memory-model.svg"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "v3") {
		t.Errorf("final svg = %q, want the last attempt kept", data)
	}
	if !strings.Contains(out.String(), "still has visual issues after 3 rounds") {
		t.Errorf("output:\n%s", out.String())
	}
}

func TestVisualQAFallsBackWhenScreenshotFails(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{
		content: []string{svgBody("v1")},
		review:  []string{reviewJSON(9, "Fine.")}, // text rubric review
	}
	env, out := runEnv(t, fake)
	env.Screenshotter = &fakeScreenshotter{err: fmt.Errorf("no chromium")}

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVisuals}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "falling back to source-text review") {
		t.Errorf("output:\n%s", out.String())
	}
	if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), DiagramsDirName, "memory-model.svg")); err != nil {
		t.Errorf("diagram not written on fallback: %v", err)
	}
}

func TestExemplarsInjectedIntoDiagramPrompt(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{
		content: []string{svgBody("v1")},
		review:  []string{visualVerdictJSON(true)},
	}
	env, _ := runEnv(t, fake)
	env.Screenshotter = &fakeScreenshotter{}

	// Add an exemplar dir + template support to the test prompts.
	styleDir := filepath.Join(env.PromptsDir, diagramStyleDirName)
	if err := os.MkdirAll(styleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(styleDir, "a.svg"), []byte("<svg>EXEMPLAR-A</svg>"), 0o644); err != nil {
		t.Fatal(err)
	}
	tmpl := `{{define "system"}}Draw. {{range .Exemplars}}EX:{{.}} {{end}}{{end}}
{{define "user"}}Diagram {{.ID}}: {{.Prompt}}{{if .Critique}} CRITIQUE: {{.Critique}}{{end}}{{end}}`
	if err := os.WriteFile(filepath.Join(env.PromptsDir, diagramTemplateName), []byte(tmpl), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVisuals}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fake.contentReqs[0].Messages[0].Content, "EXEMPLAR-A") {
		t.Errorf("exemplar not injected:\n%s", fake.contentReqs[0].Messages[0].Content)
	}
}
