package pipeline

import (
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// repoPromptsDir points at the real templates shipped with the project, so
// tests catch template syntax errors and data-shape drift.
const repoPromptsDir = "../../prompts"

func TestRepoScriptTemplateRenders(t *testing.T) {
	data := scriptPromptData{
		Tone:     "warm teacher",
		Audience: "absolute beginners",
		Language: "en",
		PaceWPM:  145,
		Title:    "What is Python?",
		Diagrams: []project.DiagramSpec{{ID: "memory-model", Prompt: "3 variables in memory"}},
		Outline:  "## A heading\n- a point\n[DIAGRAM: memory-model]",
	}
	system, user, err := renderPromptFile(repoPromptsDir, scriptTemplateName, data)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"warm teacher", "145", "absolute beginners"} {
		if !strings.Contains(system, want) {
			t.Errorf("system prompt missing %q", want)
		}
	}
	for _, want := range []string{"What is Python?", "memory-model", "[DIAGRAM: memory-model]"} {
		if !strings.Contains(user, want) {
			t.Errorf("user prompt missing %q", want)
		}
	}
	if strings.Contains(user, "quality bar") {
		t.Error("critique block rendered without a critique")
	}

	data.Critique = "Section two is inaccurate about ints."
	_, userWithCritique, err := renderPromptFile(repoPromptsDir, scriptTemplateName, data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(userWithCritique, data.Critique) {
		t.Error("critique not rendered on regeneration")
	}
}

func TestRepoReviewTemplateRenders(t *testing.T) {
	data := reviewPromptData{
		Kind:     "script",
		Audience: "absolute beginners",
		Tone:     "warm teacher",
		PaceWPM:  145,
		Outline:  "## A heading",
		Artifact: `{"title":"x"}`,
	}
	system, user, err := renderPromptFile(repoPromptsDir, reviewTemplateName, data)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range requiredScores {
		if !strings.Contains(system, want) {
			t.Errorf("rubric system prompt missing dimension %q", want)
		}
	}
	if !strings.Contains(user, `{"title":"x"}`) {
		t.Error("artifact not embedded in user prompt")
	}
}

func TestRepoDiagramTemplateRenders(t *testing.T) {
	colors := config.Colors{Primary: "#306998", Accent: "#ffd43b", Background: "#ffffff"}
	theme := deriveVideoTheme(colors, config.Fonts{}, "What is Python?", "")
	data := diagramPromptData{
		DiagramStyle: "clean, flat",
		Colors:       colors,
		Theme:        theme,
		Title:        "What is Python?",
		ID:           "memory-model",
		Prompt:       "3 variables in memory",
	}
	system, user, err := renderPromptFile(repoPromptsDir, diagramTemplateName, data)
	if err != nil {
		t.Fatal(err)
	}
	// The prompt hands the model the derived dark-stage tokens, keeps the
	// course accent, and demands a transparent canvas.
	for _, want := range []string{theme.Surface, theme.Text, theme.Accent, theme.BgTop, "TRANSPARENT", "viewBox", "self-contained"} {
		if !strings.Contains(system, want) {
			t.Errorf("diagram system prompt missing %q", want)
		}
	}
	if !strings.Contains(user, "memory-model") || !strings.Contains(user, "3 variables in memory") {
		t.Errorf("diagram user prompt missing spec: %q", user)
	}
}

func TestRepoQuizTemplateRenders(t *testing.T) {
	data := quizPromptData{
		Audience:  "absolute beginners",
		Language:  "en",
		Title:     "What is Python?",
		Outline:   "## A heading",
		Narration: "Python reads code line by line.",
	}
	system, user, err := renderPromptFile(repoPromptsDir, quizTemplateName, data)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(system, "answer_index") {
		t.Errorf("quiz system prompt missing schema: %q", system)
	}
	if !strings.Contains(user, "Python reads code line by line.") {
		t.Errorf("quiz user prompt missing narration: %q", user)
	}
}

func TestRepoDemoTapeTemplateRenders(t *testing.T) {
	data := demoPromptData{
		Audience:    "absolute beginners",
		LessonTitle: "What is Python?",
		Description: "open the Python REPL and print a greeting",
		CodeContext: "Code (lesson.md):\nprint('hi')\nActual output:\nhi",
	}
	system, user, err := renderPromptFile(repoPromptsDir, demoTapeTemplateName, data)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"NEVER fake output", "echo", "python3", "tape body"} {
		if !strings.Contains(system, want) {
			t.Errorf("tape system prompt missing %q", want)
		}
	}
	if !strings.Contains(user, "REPL") || !strings.Contains(user, "print('hi')") {
		t.Errorf("tape user prompt missing demo/code context: %q", user)
	}
}

func TestStripFences(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{name: "plain", in: `{"a":1}`, want: `{"a":1}`},
		{name: "json fence", in: "```json\n{\"a\":1}\n```", want: `{"a":1}`},
		{name: "bare fence", in: "```\n{\"a\":1}\n```", want: `{"a":1}`},
		{name: "surrounding whitespace", in: "\n\n  {\"a\":1}  \n", want: `{"a":1}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripFences(tt.in); got != tt.want {
				t.Errorf("stripFences(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
