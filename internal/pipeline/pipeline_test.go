package pipeline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// testLessonMD declares one diagram so scripts can reference it in cues.
const testLessonMD = `---
title: Test Lesson
diagrams:
  - id: memory-model
    prompt: "3 variables in memory"
---

## First idea
- a point

[DIAGRAM: memory-model]

## Second idea
- another point

[DEMO: showing the thing live]
`

func testLesson(t *testing.T) *project.Lesson {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "01-test")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, project.LessonFileName), []byte(testLessonMD), 0o644); err != nil {
		t.Fatal(err)
	}
	l, err := project.LoadLesson(dir)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

// writeTestPrompts creates a minimal but well-formed prompts dir.
func writeTestPrompts(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "prompts")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		storyboardTemplateName: `{{define "system"}}Plan beats. Max {{.MaxPoints}}. Icons: {{.Icons}}{{end}}
{{define "user"}}Lesson: {{.Title}}
{{- range .Sections}}
[{{.ID}}] {{.Narration}}
{{- end}}{{end}}`,
		captionEmphasisTemplateName: `{{define "system"}}Pick {{.Budget}} keywords.{{end}}
{{define "user"}}{{.Words}}{{end}}`,
		scriptTemplateName: `{{define "system"}}Write a script. Tone: {{.Tone}}. Pace: {{.PaceWPM}} wpm.{{end}}
{{define "user"}}Title: {{.Title}}
Diagrams:{{range .Diagrams}} {{.ID}}{{end}}
Outline:
{{.Outline}}
{{- if .Notes}}
NOTES: {{.Notes}}
{{- end}}
{{- if .Critique}}
CRITIQUE: {{.Critique}}
{{- end}}{{end}}`,
		reviewTemplateName: `{{define "system"}}Review the artifact for {{.Audience}}.{{end}}
{{define "user"}}Kind: {{.Kind}}
Artifact:
{{.Artifact}}{{end}}`,
		reviewClaimsTemplateName: `{{define "system"}}Extract claims.{{end}}
{{define "user"}}Outline: {{.Outline}}
Artifact: {{.Artifact}}{{end}}`,
		reviewAccuracyTemplateName: `{{define "system"}}Score accuracy for {{.Audience}}.{{end}}
{{define "user"}}Artifact: {{.Artifact}}
Claims: {{.ClaimResults}}
{{- if .VerifiedOutputs}}
Verified: {{.VerifiedOutputs}}
{{- end}}{{end}}`,
		reviewPedagogyTemplateName: `{{define "system"}}Score pedagogy for {{.Audience}} at {{.PaceWPM}} wpm.{{end}}
{{define "user"}}Outline: {{.Outline}}
Artifact: {{.Artifact}}{{end}}`,
		reviewToneTemplateName: `{{define "system"}}Score tone: {{.Tone}} for {{.Audience}}.{{end}}
{{define "user"}}Artifact: {{.Artifact}}{{end}}`,
		conceptsTemplateName: `{{define "system"}}Extract concepts.{{end}}
{{define "user"}}Lesson {{.LessonID}} — {{.Title}}
Outline: {{.Outline}}
{{- if .Narration}}
Narration: {{.Narration}}
{{- end}}{{end}}`,
		terminologyTemplateName: `{{define "system"}}Check terminology.{{end}}
{{define "user"}}Terms: {{.TermsDump}}{{end}}`,
		bridgeTemplateName: `{{define "system"}}Score the bridge.{{end}}
{{define "user"}}{{.PrevID}} ({{.PrevTitle}}): {{.PrevClosing}}
{{.NextID}} ({{.NextTitle}}): {{.NextOpening}}{{end}}`,
		mistakesTemplateName: `{{define "system"}}List {{.Count}} mistakes for {{.Audience}}.{{end}}
{{define "user"}}Title: {{.Title}}
Outline: {{.Outline}}
Narration: {{.Narration}}{{end}}`,
		exercisesTemplateName: `{{define "system"}}Write {{.Count}} exercises for {{.Audience}}.{{end}}
{{define "user"}}Title: {{.Title}}
Outline: {{.Outline}}
Narration: {{.Narration}}{{end}}`,
		diagramTemplateName: `{{define "system"}}Draw an SVG. Style: {{.DiagramStyle}}. Primary: {{.Colors.Primary}}.{{end}}
{{define "user"}}Diagram {{.ID}}: {{.Prompt}}
{{- if .Critique}}
CRITIQUE: {{.Critique}}
{{- end}}{{end}}`,
		quizTemplateName: `{{define "system"}}Write a quiz for {{.Audience}}.{{end}}
{{define "user"}}Title: {{.Title}}
Narration: {{.Narration}}
{{- if .EarlierConcepts}}
EARLIER: {{.EarlierConcepts}}
{{- end}}
{{- if .Critique}}
CRITIQUE: {{.Critique}}
{{- end}}{{end}}`,
		quizDistractorsTemplateName: `{{define "system"}}Score distractors for {{.Audience}}.{{end}}
{{define "user"}}Quiz: {{.Quiz}}{{end}}`,
		quizDifficultyTemplateName: `{{define "system"}}Simulate {{.Students}} students for {{.Audience}}.{{end}}
{{define "user"}}Quiz: {{.Quiz}}{{end}}`,
		diagramVisualQATemplateName: `{{define "system"}}Inspect the diagram screenshot.{{end}}
{{define "user"}}Diagram {{.ID}} should show: {{.Prompt}}{{end}}`,
		demoTapeTemplateName: `{{define "system"}}Write a VHS tape for {{.Audience}}.{{end}}
{{define "user"}}Demo: {{.Description}}
{{- if .CodeContext}}
Code: {{.CodeContext}}
{{- end}}
{{- if .Critique}}
CRITIQUE: {{.Critique}}
{{- end}}{{end}}`,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func statusEnv(t *testing.T) *Env {
	t.Helper()
	return &Env{PromptsDir: writeTestPrompts(t)}
}

func TestStageInputsChangeWithSourceConfigAndTemplates(t *testing.T) {
	l := testLesson(t)
	cfg := config.Defaults()
	e := statusEnv(t)

	base, err := e.StageInputs(l, cfg, project.StageScript)
	if err != nil {
		t.Fatal(err)
	}
	if base["lesson.md"] == "" || base["config"] == "" {
		t.Fatalf("StageInputs missing core keys: %+v", base)
	}
	if base["prompts/"+scriptTemplateName] == artifactAbsent {
		t.Fatalf("script template not hashed: %+v", base)
	}

	// Editing lesson.md must change only the lesson.md hash.
	if err := os.WriteFile(l.SourcePath(), []byte(testLessonMD+"\n- edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	edited, err := e.StageInputs(l, cfg, project.StageScript)
	if err != nil {
		t.Fatal(err)
	}
	if edited["lesson.md"] == base["lesson.md"] {
		t.Error("lesson.md hash unchanged after edit")
	}
	if edited["config"] != base["config"] {
		t.Error("config hash changed without a config change")
	}

	// Changing config must change the config fingerprint.
	cfg2 := cfg
	cfg2.Style.Tone = "completely different"
	reconfigured, err := e.StageInputs(l, cfg2, project.StageScript)
	if err != nil {
		t.Fatal(err)
	}
	if reconfigured["config"] == base["config"] {
		t.Error("config hash unchanged after config edit")
	}

	// Editing the prompt template must change its hash.
	if err := os.WriteFile(filepath.Join(e.PromptsDir, scriptTemplateName),
		[]byte(`{{define "system"}}new{{end}}{{define "user"}}new {{.Title}}{{end}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	retempled, err := e.StageInputs(l, cfg, project.StageScript)
	if err != nil {
		t.Fatal(err)
	}
	if retempled["prompts/"+scriptTemplateName] == base["prompts/"+scriptTemplateName] {
		t.Error("template hash unchanged after template edit")
	}
}

func TestStageInputsUpstreamArtifacts(t *testing.T) {
	l := testLesson(t)
	cfg := config.Defaults()
	e := statusEnv(t)

	before, err := e.StageInputs(l, cfg, project.StageReview)
	if err != nil {
		t.Fatal(err)
	}
	if before[ScriptFileName] != artifactAbsent {
		t.Fatalf("%s = %q, want %q while artifact is missing", ScriptFileName, before[ScriptFileName], artifactAbsent)
	}

	if err := os.MkdirAll(l.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(l.GeneratedDir(), ScriptFileName), []byte(`{"title":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	after, err := e.StageInputs(l, cfg, project.StageReview)
	if err != nil {
		t.Fatal(err)
	}
	if after[ScriptFileName] == artifactAbsent {
		t.Errorf("%s still reads as absent after being written", ScriptFileName)
	}
}

func TestLessonStatusLifecycle(t *testing.T) {
	l := testLesson(t)
	cfg := config.Defaults()
	e := statusEnv(t)

	statuses, err := e.LessonStatus(l, cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, stage := range project.StageOrder {
		if statuses[stage] != project.StatusPending {
			t.Errorf("fresh lesson: %s = %s, want pending", stage, statuses[stage])
		}
	}

	// Simulate the script stage completing with exactly its current inputs.
	inputs, err := e.StageInputs(l, cfg, project.StageScript)
	if err != nil {
		t.Fatal(err)
	}
	state, err := l.LoadState()
	if err != nil {
		t.Fatal(err)
	}
	state.MarkDone(project.StageScript, inputs, time.Now())
	if err := l.SaveState(state); err != nil {
		t.Fatal(err)
	}

	statuses, err = e.LessonStatus(l, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if statuses[project.StageScript] != project.StatusDone {
		t.Errorf("script = %s, want done", statuses[project.StageScript])
	}

	// Editing the source makes the completed stage stale.
	if err := os.WriteFile(l.SourcePath(), []byte(testLessonMD+"\n- edited\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	statuses, err = e.LessonStatus(l, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if statuses[project.StageScript] != project.StatusStale {
		t.Errorf("script after edit = %s, want stale", statuses[project.StageScript])
	}
}
