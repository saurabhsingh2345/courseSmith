package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Template file names under the prompts directory.
const (
	scriptTemplateName          = "script.tmpl"
	reviewTemplateName          = "review_rubric.tmpl"
	reviewClaimsTemplateName    = "review_claims.tmpl"
	reviewAccuracyTemplateName  = "review_accuracy.tmpl"
	reviewPedagogyTemplateName  = "review_pedagogy.tmpl"
	reviewToneTemplateName      = "review_tone.tmpl"
	diagramTemplateName         = "diagram_svg.tmpl"
	d3DiagramTemplateName       = "diagram_d3.tmpl"
	mermaidDiagramTemplateName  = "diagram_mermaid.tmpl"
	excalidrawTemplateName      = "diagram_excalidraw.tmpl"
	diagramVisualQATemplateName = "diagram_visual_qa.tmpl"
	quizTemplateName            = "quiz.tmpl"
	demoTapeTemplateName        = "demo_tape.tmpl"
	conceptsTemplateName        = "concepts.tmpl"
	terminologyTemplateName     = "terminology.tmpl"
	bridgeTemplateName          = "bridge.tmpl"
	quizDistractorsTemplateName = "quiz_distractors.tmpl"
	quizDifficultyTemplateName  = "quiz_difficulty.tmpl"
	mistakesTemplateName        = "mistakes.tmpl"
	exercisesTemplateName       = "exercises.tmpl"
	captionEmphasisTemplateName = "caption_emphasis.tmpl"
	storyboardTemplateName      = "storyboard.tmpl"
	d2LangTemplateName          = "diagram_d2lang.tmpl"
)

// renderPrompt resolves a template through the Env's prompt search path —
// the course's own prompts/ dir (archetype/course overrides) first, the
// project prompts/ dir second — and renders it.
func (e *Env) renderPrompt(file string, data any) (system, user string, err error) {
	if dir := e.CoursePromptsDir; dir != "" {
		if _, statErr := os.Stat(filepath.Join(dir, file)); statErr == nil {
			return renderPromptFile(dir, file, data)
		}
	}
	return renderPromptFile(e.PromptsDir, file, data)
}

// renderPromptFile loads a prompt template file and renders its "system"
// and "user" sections with data. Templates live on disk (prompts/*.tmpl)
// so they can be tuned without recompiling; each must contain
// {{define "system"}}...{{end}} and {{define "user"}}...{{end}}.
func renderPromptFile(promptsDir, file string, data any) (system, user string, err error) {
	path := filepath.Join(promptsDir, file)
	if _, statErr := os.Stat(path); statErr != nil {
		return "", "", fmt.Errorf(
			"prompt template %s not found — run coursesmith from the project root (the directory containing %s/): %w",
			path, promptsDir, statErr,
		)
	}
	tmpl, err := template.New(file).Option("missingkey=error").ParseFiles(path)
	if err != nil {
		return "", "", fmt.Errorf("parsing prompt template %s: %w", path, err)
	}
	system, err = renderSection(tmpl, path, "system", data)
	if err != nil {
		return "", "", err
	}
	user, err = renderSection(tmpl, path, "user", data)
	if err != nil {
		return "", "", err
	}
	return system, user, nil
}

func renderSection(tmpl *template.Template, path, section string, data any) (string, error) {
	if tmpl.Lookup(section) == nil {
		return "", fmt.Errorf(`prompt template %s must contain {{define %q}}...{{end}}`, path, section)
	}
	var sb strings.Builder
	if err := tmpl.ExecuteTemplate(&sb, section, data); err != nil {
		return "", fmt.Errorf("rendering %q section of %s: %w", section, path, err)
	}
	out := strings.TrimSpace(sb.String())
	if out == "" {
		return "", fmt.Errorf("%q section of %s rendered empty", section, path)
	}
	return out, nil
}
