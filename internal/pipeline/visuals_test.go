package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/project"
)

func TestVisualsStageWritesReviewedSVG(t *testing.T) {
	course, lesson := testCourse(t)
	fake := &fakeRouter{
		content: []string{svgBody("Memory model")},
		review:  []string{reviewJSON(9, "Clean and legible.")},
	}
	env, _ := runEnv(t, fake)

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVisuals}); err != nil {
		t.Fatal(err)
	}

	svgPath := filepath.Join(lesson.GeneratedDir(), DiagramsDirName, "memory-model.svg")
	data, err := os.ReadFile(svgPath)
	if err != nil {
		t.Fatalf("diagram not written: %v", err)
	}
	if !strings.HasPrefix(string(data), "<svg") || !strings.Contains(string(data), "Memory model") {
		t.Errorf("svg content = %q", data)
	}
	if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), ReviewsDirName, "diagram-memory-model-round-1.json")); err != nil {
		t.Errorf("diagram review record missing: %v", err)
	}

	sys := fake.contentReqs[0].Messages[0].Content
	if !strings.Contains(sys, "#306998") {
		t.Errorf("diagram system prompt missing branding color: %q", sys)
	}
	if fake.contentReqs[0].JSONMode {
		t.Error("SVG generation must not use JSON mode")
	}
}

func TestVisualsStageRetriesInvalidXML(t *testing.T) {
	course, lesson := testCourse(t)
	broken := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 400"><rect width="800"</svg>`
	fake := &fakeRouter{
		content: []string{broken, svgBody("Fixed")},
		review:  []string{reviewJSON(9, "Fine.")},
	}
	env, _ := runEnv(t, fake)

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVisuals}); err != nil {
		t.Fatal(err)
	}
	if len(fake.contentReqs) != 2 {
		t.Fatalf("content requests = %d, want 2 (broken + correction)", len(fake.contentReqs))
	}
	correction := fake.contentReqs[1].Messages[3].Content
	if !strings.Contains(correction, "invalid XML") {
		t.Errorf("correction message = %q, want the XML error surfaced", correction)
	}
}

func TestVisualsStageNoDiagrams(t *testing.T) {
	course, lesson := testCourse(t)
	// Rewrite the lesson without diagram declarations.
	md := "---\ntitle: Test Lesson\n---\n\n## Only ideas\n- no diagrams here\n"
	if err := os.WriteFile(lesson.SourcePath(), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	var err error
	lesson, err = project.LoadLesson(lesson.Dir)
	if err != nil {
		t.Fatal(err)
	}

	env, out := runEnv(t, &fakeRouter{}) // any LLM call would fail the run
	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVisuals}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "no diagrams declared") {
		t.Errorf("output = %q", out.String())
	}
}

func TestVisualsReviewRegeneratesWithCritique(t *testing.T) {
	course, lesson := testCourse(t)
	critique := "Labels overlap; increase spacing."
	fake := &fakeRouter{
		content: []string{svgBody("v1"), svgBody("v2")},
		review:  []string{reviewJSON(6, critique), reviewJSON(9, "Fixed.")},
	}
	env, _ := runEnv(t, fake)

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageVisuals}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fake.contentReqs[1].Messages[1].Content, critique) {
		t.Errorf("regeneration prompt missing critique:\n%s", fake.contentReqs[1].Messages[1].Content)
	}
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), DiagramsDirName, "memory-model.svg"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "v2") {
		t.Errorf("svg = %q, want the regenerated version kept", data)
	}
}

func TestExtractSVG(t *testing.T) {
	valid := svgBody("ok")
	tests := []struct {
		name    string
		in      string
		wantErr string
	}{
		{name: "plain svg", in: valid},
		{name: "fenced svg", in: "```svg\n" + valid + "\n```"},
		{name: "chatty preamble", in: "Here is your diagram:\n" + valid + "\nHope it helps!"},
		{name: "no svg", in: "I cannot draw that.", wantErr: "no <svg> element"},
		{name: "unclosed", in: "<svg viewBox=\"0 0 800 400\"><rect/>", wantErr: "never closed"},
		{name: "malformed xml", in: `<svg viewBox="0 0 800 400"><rect width=</svg>`, wantErr: "invalid XML"},
		{name: "missing viewBox", in: `<svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>`, wantErr: "viewBox"},
		{name: "script element", in: `<svg viewBox="0 0 800 400"><script>alert(1)</script></svg>`, wantErr: "<script>"},
		{name: "image element", in: `<svg viewBox="0 0 800 400"><image href="x.png"/></svg>`, wantErr: "<image>"},
		{
			name:    "external url",
			in:      `<svg viewBox="0 0 800 400"><use href="https://evil.example/x.svg#a"/></svg>`,
			wantErr: "external URL",
		},
		{
			name: "xmlns urls are allowed",
			in:   `<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" viewBox="0 0 800 400"><rect/></svg>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractSVG(tt.in)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("extractSVG() error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(got, "<svg") || !strings.HasSuffix(strings.TrimSpace(got), "</svg>") {
				t.Errorf("extractSVG() = %q, want a bare svg element", got)
			}
		})
	}
}
