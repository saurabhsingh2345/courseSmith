package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// QuizFileName is the Stage 4 output in the lesson's generated dir.
const QuizFileName = "quiz.json"

// Quiz is the end-of-lesson assessment.
type Quiz struct {
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
}

// Question taxonomy: every quiz needs at least one of each.
const (
	QRecall      = "recall"      // remember a fact from the lesson
	QApplication = "application" // apply the concept to a new case
	QDebugging   = "debugging"   // spot what is wrong with given code
	QPrediction  = "prediction"  // predict what code prints/does
)

var questionTypes = []string{QRecall, QApplication, QDebugging, QPrediction}

// Question is one multiple-choice question.
type Question struct {
	ID     string `json:"id"`
	Type   string `json:"type"` // recall | application | debugging | prediction
	Prompt string `json:"prompt"`
	// Review marks a spaced-repetition question drawn from an earlier
	// lesson's concepts (lessons 4+).
	Review      bool     `json:"review,omitempty"`
	Options     []string `json:"options"`
	AnswerIndex int      `json:"answer_index"`
	Explanation string   `json:"explanation"`
}

// Validate checks structural soundness of a quiz: 4-10 questions, valid
// taxonomy tags, and at least one question of every type.
func (q *Quiz) Validate() error {
	if strings.TrimSpace(q.Title) == "" {
		return fmt.Errorf("quiz title is empty")
	}
	if len(q.Questions) < 4 || len(q.Questions) > 10 {
		return fmt.Errorf("quiz has %d questions, want 4-10", len(q.Questions))
	}
	seen := map[string]bool{}
	typeCount := map[string]int{}
	for i, question := range q.Questions {
		if strings.TrimSpace(question.ID) == "" {
			return fmt.Errorf("question %d has an empty id", i)
		}
		if seen[question.ID] {
			return fmt.Errorf("duplicate question id %q", question.ID)
		}
		seen[question.ID] = true
		if !slices.Contains(questionTypes, question.Type) {
			return fmt.Errorf("question %q has type %q (want one of: %s)",
				question.ID, question.Type, strings.Join(questionTypes, ", "))
		}
		typeCount[question.Type]++
		if strings.TrimSpace(question.Prompt) == "" {
			return fmt.Errorf("question %q has an empty prompt", question.ID)
		}
		if len(question.Options) < 2 || len(question.Options) > 5 {
			return fmt.Errorf("question %q has %d options, want 2-5", question.ID, len(question.Options))
		}
		for j, opt := range question.Options {
			if strings.TrimSpace(opt) == "" {
				return fmt.Errorf("question %q option %d is empty", question.ID, j)
			}
		}
		if question.AnswerIndex < 0 || question.AnswerIndex >= len(question.Options) {
			return fmt.Errorf("question %q answer_index %d is out of range (0-%d)", question.ID, question.AnswerIndex, len(question.Options)-1)
		}
		if strings.TrimSpace(question.Explanation) == "" {
			return fmt.Errorf("question %q has an empty explanation", question.ID)
		}
	}
	for _, qt := range questionTypes {
		if typeCount[qt] == 0 {
			return fmt.Errorf("quiz has no %q question — every quiz needs at least one of each type (%s)",
				qt, strings.Join(questionTypes, ", "))
		}
	}
	return nil
}

// quizPromptData feeds prompts/quiz.tmpl.
type quizPromptData struct {
	Audience  string
	Language  string
	Title     string
	Outline   string
	Narration string
	// EarlierConcepts lists concepts from previous lessons for spaced-
	// repetition review questions ("" before lesson 4 or when unknown).
	EarlierConcepts string
	Critique        string
}

// spacedRepetitionFromLesson is the lesson ordinal (numeric prefix) from
// which quizzes include review questions on earlier material.
const spacedRepetitionFromLesson = 4

// lessonOrdinal parses the numeric prefix of a lesson id ("04-loops" → 4);
// 0 when there is none.
func lessonOrdinal(id string) int {
	digits, _, _ := strings.Cut(id, "-")
	n, err := strconv.Atoi(digits)
	if err != nil {
		return 0
	}
	return n
}

// earlierConceptsSummary collects concepts introduced by lessons before l,
// from their cached concept extractions (coursesmith analyze). Missing
// caches are skipped — spaced repetition degrades gracefully.
func earlierConceptsSummary(course *project.Course, l *project.Lesson) string {
	if course == nil || lessonOrdinal(l.ID) < spacedRepetitionFromLesson {
		return ""
	}
	lessons, err := course.Lessons()
	if err != nil {
		return ""
	}
	var b strings.Builder
	for _, prev := range lessons {
		if prev.ID >= l.ID {
			break
		}
		data, err := os.ReadFile(filepath.Join(prev.GeneratedDir(), ConceptsFileName))
		if err != nil {
			continue
		}
		var lc LessonConcepts
		if json.Unmarshal(data, &lc) != nil {
			continue
		}
		var names []string
		for _, ref := range lc.Introduced {
			names = append(names, ref.Name)
		}
		if len(names) > 0 {
			fmt.Fprintf(&b, "%s: %s\n", prev.ID, strings.Join(names, ", "))
		}
	}
	return strings.TrimSpace(b.String())
}

// loadScript reads and parses the lesson's generated script.json.
func loadScript(l *project.Lesson) (*Script, error) {
	path := filepath.Join(l.GeneratedDir(), ScriptFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no %s yet — the script stage must run first", ScriptFileName)
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var script Script
	if err := json.Unmarshal(data, &script); err != nil {
		return nil, fmt.Errorf("parsing %s (delete it and re-run the script stage): %w", path, err)
	}
	return &script, nil
}

// generateQuiz asks the content model for a quiz grounded in the outline and
// the (already quality-gated) narration.
func generateQuiz(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, narration, earlierConcepts, critique string) (*Quiz, error) {
	data := quizPromptData{
		Audience:        cfg.Style.Audience,
		Language:        cfg.Style.Language,
		Title:           l.FrontMatter.Title,
		Outline:         l.Body,
		Narration:       narration,
		EarlierConcepts: earlierConcepts,
		Critique:        critique,
	}
	system, user, err := e.renderPrompt(quizTemplateName, data)
	if err != nil {
		return nil, err
	}
	var quiz Quiz
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.4, 4096, &quiz, func() error {
		if err := quiz.Validate(); err != nil {
			return err
		}
		// Execution check: code in questions must run, and when it prints
		// one of the options, the answer must point at it.
		return verifyQuizCode(ctx, e, &quiz)
	})
	if err != nil {
		return nil, fmt.Errorf("generating quiz: %w", err)
	}
	return &quiz, nil
}

// runQuizStage is Stage 4: script + outline → quality-gated generated/quiz.json,
// with distractor scoring and a difficulty simulation on the final quiz.
func runQuizStage(ctx context.Context, e *Env, course *project.Course, l *project.Lesson, cfg config.Config) error {
	script, err := loadScript(l)
	if err != nil {
		return err
	}
	narrations := make([]string, 0, len(script.Sections))
	for _, sec := range script.Sections {
		narrations = append(narrations, sec.Narration)
	}
	narration := strings.Join(narrations, "\n\n")
	earlierConcepts := earlierConceptsSummary(course, l)
	if earlierConcepts != "" {
		fmt.Fprintf(e.out(), "  → quiz      including spaced-repetition review of earlier concepts\n")
	}

	fmt.Fprintf(e.out(), "  → quiz      writing questions (%s)...\n", cfg.Pipeline.LLMContent)
	quiz, err := generateQuiz(ctx, e, l, cfg, narration, earlierConcepts, "")
	if err != nil {
		return err
	}
	draft, err := json.MarshalIndent(quiz, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding quiz: %w", err)
	}
	regenerate := func(ctx context.Context, critique string) ([]byte, error) {
		next, err := generateQuiz(ctx, e, l, cfg, narration, earlierConcepts, critique)
		if err != nil {
			return nil, err
		}
		data, err := json.MarshalIndent(next, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("encoding regenerated quiz: %w", err)
		}
		return append(data, '\n'), nil
	}
	best, _, err := e.reviewGate(ctx, l, cfg, "quiz", append(draft, '\n'), regenerate)
	if err != nil {
		return err
	}

	// Distractor quality: wrong options must be plausible misconceptions.
	// One regeneration round when weak distractors are found.
	best, err = distractorGate(ctx, e, l, cfg, best, regenerate)
	if err != nil {
		return err
	}

	if err := writeFileAtomic(filepath.Join(l.GeneratedDir(), QuizFileName), best); err != nil {
		return err
	}

	var final Quiz
	if err := json.Unmarshal(best, &final); err == nil {
		fmt.Fprintf(e.out(), "    %d questions written to %s\n", len(final.Questions), QuizFileName)
		// Difficulty simulation: the review model role-plays the audience
		// answering cold; extremes are flagged, not fatal.
		if err := simulateDifficulty(ctx, e, l, cfg, &final); err != nil {
			return err
		}
	}
	return nil
}
