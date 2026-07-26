package pipeline

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/project"
)

// seedScript writes a script.json so the quiz stage has narration to work from.
func seedScript(t *testing.T, l *project.Lesson) {
	t.Helper()
	if err := os.MkdirAll(l.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(l.GeneratedDir(), ScriptFileName),
		[]byte(scriptJSON("Python reads code line by line.")), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestQuizStageWritesReviewedQuiz(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson)
	fake := &fakeRouter{
		content: []string{quizBody()},
		review:  quizReviewQueue(),
	}
	env, _ := runEnv(t, fake)

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageQuiz}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), QuizFileName))
	if err != nil {
		t.Fatalf("quiz not written: %v", err)
	}
	var quiz Quiz
	if err := json.Unmarshal(data, &quiz); err != nil {
		t.Fatal(err)
	}
	if len(quiz.Questions) != 4 || quiz.Questions[0].ID != "q1" {
		t.Errorf("quiz = %+v", quiz)
	}
	if quiz.Questions[0].Type != QRecall || quiz.Questions[3].Type != QPrediction {
		t.Errorf("taxonomy tags = %q, %q", quiz.Questions[0].Type, quiz.Questions[3].Type)
	}
	for _, record := range []string{"quiz-round-1.json", QuizDistractorsFileName, QuizDifficultyFileName} {
		if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), ReviewsDirName, record)); err != nil {
			t.Errorf("quiz QA record missing: %v", err)
		}
	}
	// The generator must have seen the narration, not just the outline.
	if !strings.Contains(fake.contentReqs[0].Messages[1].Content, "Python reads code line by line.") {
		t.Errorf("quiz prompt missing narration:\n%s", fake.contentReqs[0].Messages[1].Content)
	}
}

func TestQuizStageRegeneratesWeakDistractors(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson)
	fake := &fakeRouter{
		content: []string{quizBody(), quizBody()}, // initial + distractor regeneration
		review: []string{
			reviewJSON(9, "Fair quiz."),
			distractorsJSON(3), // all weak → regenerate
			distractorsJSON(8), // second round is strong
			difficultyJSON(7),
		},
	}
	env, out := runEnv(t, fake)

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageQuiz}); err != nil {
		t.Fatal(err)
	}
	regen := fake.contentReqs[1]
	if !strings.Contains(regen.Messages[1].Content, "plausible misconceptions") {
		t.Errorf("distractor critique not injected:\n%s", regen.Messages[1].Content)
	}
	if !strings.Contains(out.String(), "weak distractor") {
		t.Errorf("output missing weak-distractor notice:\n%s", out.String())
	}

	var report struct {
		Rounds []struct {
			Round int      `json:"round"`
			Weak  []string `json:"weak"`
		} `json:"rounds"`
	}
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), ReviewsDirName, QuizDistractorsFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Rounds) != 2 || len(report.Rounds[0].Weak) != 12 || len(report.Rounds[1].Weak) != 0 {
		t.Errorf("distractor rounds = %+v", report.Rounds)
	}
}

func TestQuizStageFlagsDifficultyExtremes(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson)
	fake := &fakeRouter{
		content: []string{quizBody()},
		review: []string{
			reviewJSON(9, "Fair quiz."),
			distractorsJSON(8),
			difficultyJSON(10), // 10/10 correct → too easy
		},
	}
	env, out := runEnv(t, fake)

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageQuiz}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "too_easy") {
		t.Errorf("output missing difficulty flag:\n%s", out.String())
	}
	var report struct {
		Flagged []string `json:"flagged"`
	}
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), ReviewsDirName, QuizDifficultyFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Flagged) != 4 {
		t.Errorf("flagged = %v, want all 4 questions", report.Flagged)
	}
}

func TestQuizStageRequiresScript(t *testing.T) {
	course, lesson := testCourse(t)
	env, _ := runEnv(t, &fakeRouter{})
	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageQuiz})
	if err == nil || !strings.Contains(err.Error(), "script stage must run first") {
		t.Errorf("error = %v, want script-first error", err)
	}
}

func TestQuizValidation(t *testing.T) {
	valid := func() Quiz {
		var q Quiz
		if err := json.Unmarshal([]byte(quizBody()), &q); err != nil {
			t.Fatal(err)
		}
		return q
	}
	tests := []struct {
		name    string
		mutate  func(*Quiz)
		wantErr string
	}{
		{name: "valid", mutate: func(q *Quiz) {}},
		{name: "no title", mutate: func(q *Quiz) { q.Title = " " }, wantErr: "title is empty"},
		{name: "too few questions", mutate: func(q *Quiz) { q.Questions = q.Questions[:2] }, wantErr: "want 4-10"},
		{
			name:    "unknown type",
			mutate:  func(q *Quiz) { q.Questions[0].Type = "trivia" },
			wantErr: `type "trivia"`,
		},
		{
			name:    "missing taxonomy coverage",
			mutate:  func(q *Quiz) { q.Questions[3].Type = QRecall }, // no prediction left
			wantErr: `no "prediction" question`,
		},
		{
			name:    "answer out of range",
			mutate:  func(q *Quiz) { q.Questions[1].AnswerIndex = 4 },
			wantErr: "answer_index 4 is out of range",
		},
		{
			name:    "duplicate ids",
			mutate:  func(q *Quiz) { q.Questions[2].ID = "q1" },
			wantErr: "duplicate question id",
		},
		{
			name:    "empty option",
			mutate:  func(q *Quiz) { q.Questions[0].Options[2] = "" },
			wantErr: "option 2 is empty",
		},
		{
			name:    "missing explanation",
			mutate:  func(q *Quiz) { q.Questions[0].Explanation = "" },
			wantErr: "empty explanation",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := valid()
			tt.mutate(&q)
			err := q.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Validate() = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestVerifyQuizCodeAllowsFailingDebuggingCode(t *testing.T) {
	q := &Quiz{Title: "T", Questions: []Question{
		{ID: "q1", Type: QDebugging, Prompt: "Why does this fail?\n```python\nBROKEN = oops\n```",
			Options: []string{"a", "b"}, AnswerIndex: 0, Explanation: "e"},
		{ID: "q2", Type: QPrediction, Prompt: "What prints?\n```python\nBROKEN = oops\n```",
			Options: []string{"a", "b"}, AnswerIndex: 0, Explanation: "e"},
	}}
	env := &Env{CodeRunner: failingRunner()} // BROKEN code exits non-zero

	// Debugging question with failing code passes...
	qDebugOnly := &Quiz{Title: "T", Questions: q.Questions[:1]}
	if err := verifyQuizCode(context.Background(), env, qDebugOnly); err != nil {
		t.Errorf("debugging question with failing code rejected: %v", err)
	}
	// ...but the same failing code in a prediction question is rejected.
	if err := verifyQuizCode(context.Background(), env, q); err == nil || !strings.Contains(err.Error(), "prediction") {
		t.Errorf("error = %v, want prediction-question rejection", err)
	}
}
