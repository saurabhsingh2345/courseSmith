package pipeline

// Quiz strategy (workstream G): reorder a generated quiz using cognitive-science
// principles — interleaving (vary question type so consecutive items don't feel
// fluent), plus a difficulty distribution target. Produces quiz_sequence.json
// alongside quiz.json.
//
// Scaffold status: the interleaving reorder is real and tested. Difficulty is
// currently target-only — the per-question difficulty estimates come from
// workstream D's IRT model (see tools/tutor), which is stubbed, so the
// distribution is reported as a goal rather than enforced.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// loadQuiz reads the generated quiz.json; a missing file returns nil.
func loadQuiz(l *project.Lesson) (*Quiz, error) {
	data, err := os.ReadFile(filepath.Join(l.GeneratedDir(), QuizFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", QuizFileName, err)
	}
	var q Quiz
	if err := json.Unmarshal(data, &q); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", QuizFileName, err)
	}
	return &q, nil
}

// QuizSequenceFileName is the quiz-strategy stage output.
const QuizSequenceFileName = "quiz_sequence.json"

// QuizSequence is the delivered order plus the reasoning behind it.
type QuizSequence struct {
	// Order is question ids in delivery order (interleaved by type).
	Order []string `json:"order"`
	// Interleaving summarizes the type pattern actually achieved.
	Interleaving string `json:"interleaving"`
	// DifficultyTargets is the intended easy/medium/hard split for a quiz of
	// this size (retrieval-practice sweet spot skews toward medium).
	DifficultyTargets map[string]int `json:"difficulty_targets"`
	// TODO records what still needs IRT data from the tutor service (D).
	TODO string `json:"_todo,omitempty"`
}

// interleaveQuestions reorders questions so that, as far as possible, no two
// consecutive questions share a type — the core interleaving principle. Greedy:
// at each step pick the type with the most remaining items that isn't the type
// just placed. Deterministic given the input order.
func interleaveQuestions(questions []Question) []string {
	buckets := map[string][]string{}
	var order []string // stable type order for determinism
	for _, q := range questions {
		if _, ok := buckets[q.Type]; !ok {
			order = append(order, q.Type)
		}
		buckets[q.Type] = append(buckets[q.Type], q.ID)
	}

	var out []string
	last := ""
	for len(out) < len(questions) {
		best := ""
		for _, typ := range order {
			if len(buckets[typ]) == 0 || typ == last {
				continue
			}
			if best == "" || len(buckets[typ]) > len(buckets[best]) {
				best = typ
			}
		}
		if best == "" {
			// Everything remaining is the same type as `last`; unavoidable repeat.
			for _, typ := range order {
				if len(buckets[typ]) > 0 {
					best = typ
					break
				}
			}
		}
		out = append(out, buckets[best][0])
		buckets[best] = buckets[best][1:]
		last = best
	}
	return out
}

// difficultyTargets returns the intended split for n questions: ~30% easy,
// ~45% medium (the productive-difficulty sweet spot), ~25% hard.
func difficultyTargets(n int) map[string]int {
	easy := n * 3 / 10
	hard := n * 25 / 100
	return map[string]int{"easy": easy, "medium": n - easy - hard, "hard": hard}
}

// interleavingSummary describes the achieved type pattern for the audit trail.
func interleavingSummary(questions []Question, order []string) string {
	byID := make(map[string]string, len(questions))
	for _, q := range questions {
		byID[q.ID] = q.Type
	}
	adjacentRepeats := 0
	prev := ""
	for _, id := range order {
		if byID[id] == prev {
			adjacentRepeats++
		}
		prev = byID[id]
	}
	return fmt.Sprintf("%d questions, %d adjacent same-type pair(s)", len(order), adjacentRepeats)
}

// runQuizStrategyStage reads quiz.json and writes quiz_sequence.json. It never
// fails on content — it only reorders — so it's a soft enhancement stage.
func runQuizStrategyStage(_ context.Context, e *Env, _ *project.Course, l *project.Lesson, _ config.Config) error {
	quiz, err := loadQuiz(l)
	if err != nil {
		return err
	}
	if quiz == nil || len(quiz.Questions) == 0 {
		fmt.Fprintf(e.out(), "  → quiz-strategy no quiz to sequence\n")
		return writeJSON(filepath.Join(l.GeneratedDir(), QuizSequenceFileName), QuizSequence{})
	}

	order := interleaveQuestions(quiz.Questions)
	seq := QuizSequence{
		Order:             order,
		Interleaving:      interleavingSummary(quiz.Questions, order),
		DifficultyTargets: difficultyTargets(len(quiz.Questions)),
		TODO:              "difficulty distribution is target-only until the IRT model (workstream D) supplies per-question estimates",
	}
	fmt.Fprintf(e.out(), "  → quiz-strategy interleaved %d questions (%s)\n", len(order), seq.Interleaving)
	return writeJSON(filepath.Join(l.GeneratedDir(), QuizSequenceFileName), seq)
}
