package pipeline

// Quiz QA beyond the rubric gate: distractor plausibility scoring and a
// cold-answer difficulty simulation, both persisted to generated/reviews/.

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// Quiz QA reports under generated/reviews/.
const (
	QuizDistractorsFileName = "quiz_distractors.json"
	QuizDifficultyFileName  = "quiz_difficulty.json"
)

// weakDistractorThreshold: plausibility below this marks a distractor as a
// throwaway that tests nothing.
const weakDistractorThreshold = 6

// Difficulty simulation bounds (fractions of simulated students correct).
const (
	tooEasyAbove = 0.9
	tooHardBelow = 0.3
)

// simulatedStudents is how many audience members the model role-plays.
const simulatedStudents = 10

// DistractorScore is the reviewer's ruling on one wrong option.
type DistractorScore struct {
	OptionIndex int `json:"option_index"`
	// Plausibility 1-10: does a real learner with a real misconception
	// pick this?
	Plausibility  float64 `json:"plausibility"`
	Misconception string  `json:"misconception"`
}

type questionDistractors struct {
	ID          string            `json:"id"`
	Distractors []DistractorScore `json:"distractors"`
}

type distractorReview struct {
	Questions []questionDistractors `json:"questions"`
}

// distractorReport is the persisted quiz_distractors.json.
type distractorReport struct {
	Model     string                `json:"model"`
	Threshold float64               `json:"threshold"`
	Rounds    []distractorRoundInfo `json:"rounds"`
	CheckedAt time.Time             `json:"checked_at"`
}

type distractorRoundInfo struct {
	Round     int                   `json:"round"`
	Questions []questionDistractors `json:"questions"`
	Weak      []string              `json:"weak"` // "qid option N" labels
}

// distractorPromptData feeds prompts/quiz_distractors.tmpl.
type distractorPromptData struct {
	Audience string
	Quiz     string
}

// scoreDistractors asks the review model to rate every wrong option.
func scoreDistractors(ctx context.Context, e *Env, cfg config.Config, quizJSON []byte) (*distractorReview, error) {
	system, user, err := e.renderPrompt(quizDistractorsTemplateName,
		distractorPromptData{Audience: cfg.Style.Audience, Quiz: string(quizJSON)})
	if err != nil {
		return nil, err
	}
	var review distractorReview
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskReview, system, user, 0, 4096, &review, nil)
	if err != nil {
		return nil, fmt.Errorf("scoring distractors: %w", err)
	}
	return &review, nil
}

// weakDistractors lists sub-threshold distractors as human-readable labels.
func weakDistractors(review *distractorReview) []string {
	var weak []string
	for _, q := range review.Questions {
		for _, d := range q.Distractors {
			if d.Plausibility < weakDistractorThreshold {
				weak = append(weak, fmt.Sprintf("%s option %d (%s)", q.ID, d.OptionIndex, d.Misconception))
			}
		}
	}
	return weak
}

// distractorGate scores the quiz's distractors; weak ones trigger one
// regeneration with the weaknesses injected, and the round with fewer weak
// distractors wins. The full audit trail lands in quiz_distractors.json.
func distractorGate(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, quizJSON []byte, regenerate func(ctx context.Context, critique string) ([]byte, error)) ([]byte, error) {
	report := distractorReport{
		Model:     cfg.Pipeline.LLMReview,
		Threshold: weakDistractorThreshold,
		CheckedAt: time.Now().UTC(),
	}
	fmt.Fprintf(e.out(), "  → quiz      scoring distractor plausibility (%s)...\n", cfg.Pipeline.LLMReview)
	review, err := scoreDistractors(ctx, e, cfg, quizJSON)
	if err != nil {
		return nil, err
	}
	weak := weakDistractors(review)
	report.Rounds = append(report.Rounds, distractorRoundInfo{Round: 1, Questions: review.Questions, Weak: weak})

	best := quizJSON
	if len(weak) > 0 {
		fmt.Fprintf(e.out(), "    %d weak distractor(s) — regenerating quiz with stronger misconceptions\n", len(weak))
		critique := "These wrong options are not plausible misconceptions — a learner would never pick them. " +
			"Replace each with a wrong answer produced by a REAL beginner misunderstanding:\n  " +
			strings.Join(weak, "\n  ")
		next, err := regenerate(ctx, critique)
		if err != nil {
			return nil, err
		}
		review2, err := scoreDistractors(ctx, e, cfg, next)
		if err != nil {
			return nil, err
		}
		weak2 := weakDistractors(review2)
		report.Rounds = append(report.Rounds, distractorRoundInfo{Round: 2, Questions: review2.Questions, Weak: weak2})
		if len(weak2) <= len(weak) {
			best = next
			weak = weak2
		}
	}
	if len(weak) > 0 {
		fmt.Fprintf(e.out(), "  ⚠ quiz      %d distractor(s) still weak — see %s\n",
			len(weak), filepath.Join(project.GeneratedDirName, ReviewsDirName, QuizDistractorsFileName))
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), ReviewsDirName, QuizDistractorsFileName), report); err != nil {
		return nil, err
	}
	return best, nil
}

// questionDifficulty is one row of the difficulty simulation.
type questionDifficulty struct {
	ID        string  `json:"id"`
	CorrectOf int     `json:"correct_of"`
	Rate      float64 `json:"rate"`
	Reasoning string  `json:"reasoning"`
	Verdict   string  `json:"verdict"` // ok | too_easy | too_hard
}

// difficultyReport is the persisted quiz_difficulty.json.
type difficultyReport struct {
	Model     string               `json:"model"`
	Students  int                  `json:"students"`
	Questions []questionDifficulty `json:"questions"`
	Flagged   []string             `json:"flagged"`
	CheckedAt time.Time            `json:"checked_at"`
}

// difficultySim is the LLM response shape.
type difficultySim struct {
	Questions []struct {
		ID        string `json:"id"`
		Correct   int    `json:"correct"`
		Reasoning string `json:"reasoning"`
	} `json:"questions"`
}

// difficultyPromptData feeds prompts/quiz_difficulty.tmpl.
type difficultyPromptData struct {
	Audience string
	Students int
	Quiz     string
}

// simulateDifficulty has the review model role-play the target audience
// answering each question cold; extreme success rates are flagged in
// quiz_difficulty.json and on the console.
func simulateDifficulty(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, quiz *Quiz) error {
	quizJSON, err := json.MarshalIndent(quiz, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding quiz for simulation: %w", err)
	}
	fmt.Fprintf(e.out(), "  → quiz      simulating %d student(s) answering cold (%s)...\n", simulatedStudents, cfg.Pipeline.LLMReview)
	system, user, err := e.renderPrompt(quizDifficultyTemplateName,
		difficultyPromptData{Audience: cfg.Style.Audience, Students: simulatedStudents, Quiz: string(quizJSON)})
	if err != nil {
		return err
	}
	var sim difficultySim
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskReview, system, user, 0.3, 4096, &sim, func() error {
		if len(sim.Questions) == 0 {
			return fmt.Errorf("simulation returned no questions")
		}
		for i, q := range sim.Questions {
			if q.Correct < 0 || q.Correct > simulatedStudents {
				return fmt.Errorf("questions[%d] correct=%d is outside 0-%d", i, q.Correct, simulatedStudents)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("difficulty simulation: %w", err)
	}

	report := difficultyReport{
		Model:     cfg.Pipeline.LLMReview,
		Students:  simulatedStudents,
		CheckedAt: time.Now().UTC(),
	}
	for _, q := range sim.Questions {
		rate := float64(q.Correct) / float64(simulatedStudents)
		row := questionDifficulty{
			ID: q.ID, CorrectOf: q.Correct, Rate: rate, Reasoning: q.Reasoning, Verdict: "ok",
		}
		switch {
		case rate > tooEasyAbove:
			row.Verdict = "too_easy"
		case rate < tooHardBelow:
			row.Verdict = "too_hard"
		}
		if row.Verdict != "ok" {
			report.Flagged = append(report.Flagged, fmt.Sprintf("%s (%s: %d/%d correct)", q.ID, row.Verdict, q.Correct, simulatedStudents))
		}
		report.Questions = append(report.Questions, row)
	}
	if len(report.Flagged) > 0 {
		fmt.Fprintf(e.out(), "  ⚠ quiz      difficulty flags: %s — see %s\n",
			strings.Join(report.Flagged, "; "),
			filepath.Join(project.GeneratedDirName, ReviewsDirName, QuizDifficultyFileName))
	}
	return writeJSON(filepath.Join(l.GeneratedDir(), ReviewsDirName, QuizDifficultyFileName), report)
}
