package pipeline

// Learning-science gate (workstream I): deterministic Go unit tests over the
// quiz-strategy logic (workstream G). These encode the cognitive-science
// invariants the pipeline is supposed to enforce so a regression in the reorder
// or the difficulty targets fails CI — no browser, no network, no LLM.
//
// The checks are written against the REAL contracts in quiz_strategy.go
// (interleaveQuestions / difficultyTargets / interleavingSummary /
// runQuizStrategyStage over the real Question and QuizSequence types), not the
// idealized `ValidateInterleaving` API sketched in the phase notes.

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

// countAdjacentSameType returns how many neighbouring pairs in the delivered
// order share a question type — the thing interleaving is meant to minimize.
func countAdjacentSameType(order []string, typeOf map[string]string) int {
	repeats := 0
	for i := 1; i < len(order); i++ {
		if typeOf[order[i]] == typeOf[order[i-1]] {
			repeats++
		}
	}
	return repeats
}

// minAdjacentSameType is the information-theoretic floor on adjacent same-type
// pairs for a multiset of types: you can always interleave to zero repeats
// unless one type dominates more than half the items, in which case the excess
// is forced to sit adjacent. Greedy interleaving should hit this floor.
func minAdjacentSameType(counts map[string]int, n int) int {
	max := 0
	for _, c := range counts {
		if c > max {
			max = c
		}
	}
	if floor := 2*max - n - 1; floor > 0 {
		return floor
	}
	return 0
}

func typeIndex(questions []Question) map[string]string {
	m := make(map[string]string, len(questions))
	for _, q := range questions {
		m[q.ID] = q.Type
	}
	return m
}

func typeCounts(questions []Question) map[string]int {
	m := map[string]int{}
	for _, q := range questions {
		m[q.Type]++
	}
	return m
}

// assertPermutation fails unless order is exactly the set of question ids, once
// each — the reorder must never drop, duplicate, or invent a question.
func assertPermutation(t *testing.T, questions []Question, order []string) {
	t.Helper()
	if len(order) != len(questions) {
		t.Fatalf("order has %d ids, want %d", len(order), len(questions))
	}
	want := map[string]bool{}
	for _, q := range questions {
		want[q.ID] = true
	}
	seen := map[string]bool{}
	for _, id := range order {
		if !want[id] {
			t.Errorf("order references unknown id %q", id)
		}
		if seen[id] {
			t.Errorf("order duplicates id %q", id)
		}
		seen[id] = true
	}
}

// TestInterleavingSpreadsTypes: when no single type dominates, consecutive
// questions should never share a type (the core interleaving principle —
// varying the retrieval demand so consecutive items don't feel fluent).
func TestInterleavingSpreadsTypes(t *testing.T) {
	questions := []Question{
		{ID: "q1", Type: QRecall},
		{ID: "q2", Type: QApplication},
		{ID: "q3", Type: QRecall},
		{ID: "q4", Type: QApplication},
		{ID: "q5", Type: QDebugging},
		{ID: "q6", Type: QPrediction},
	}
	order := interleaveQuestions(questions)
	assertPermutation(t, questions, order)

	repeats := countAdjacentSameType(order, typeIndex(questions))
	if repeats != 0 {
		t.Errorf("interleaving left %d adjacent same-type pair(s), want 0 for %v", repeats, order)
	}
}

// TestInterleavingMinimizesUnavoidableRepeats: when one type dominates the quiz
// it is impossible to fully separate them, but greedy interleaving must still
// hit the theoretical floor rather than clumping the whole run together.
func TestInterleavingMinimizesUnavoidableRepeats(t *testing.T) {
	cases := [][]Question{
		{ // 5 recall + 1 application -> floor 3
			{ID: "a1", Type: QRecall}, {ID: "a2", Type: QRecall},
			{ID: "a3", Type: QRecall}, {ID: "a4", Type: QRecall},
			{ID: "a5", Type: QRecall}, {ID: "a6", Type: QApplication},
		},
		{ // 3 recall + 3 application -> floor 0
			{ID: "b1", Type: QRecall}, {ID: "b2", Type: QRecall},
			{ID: "b3", Type: QRecall}, {ID: "b4", Type: QApplication},
			{ID: "b5", Type: QApplication}, {ID: "b6", Type: QApplication},
		},
		{ // 4 recall + 2 application + 1 debugging -> floor 1
			{ID: "c1", Type: QRecall}, {ID: "c2", Type: QRecall},
			{ID: "c3", Type: QRecall}, {ID: "c4", Type: QRecall},
			{ID: "c5", Type: QApplication}, {ID: "c6", Type: QApplication},
			{ID: "c7", Type: QDebugging},
		},
	}
	for _, questions := range cases {
		order := interleaveQuestions(questions)
		assertPermutation(t, questions, order)

		got := countAdjacentSameType(order, typeIndex(questions))
		want := minAdjacentSameType(typeCounts(questions), len(questions))
		if got != want {
			t.Errorf("adjacent same-type = %d, want floor %d for order %v", got, want, order)
		}
	}
}

// TestInterleavingIsDeterministic: the reorder must be stable given the same
// input — the video pipeline is idempotent on input hash, so a nondeterministic
// sequence would spuriously invalidate cached downstream stages.
func TestInterleavingIsDeterministic(t *testing.T) {
	questions := []Question{
		{ID: "q1", Type: QRecall}, {ID: "q2", Type: QApplication},
		{ID: "q3", Type: QDebugging}, {ID: "q4", Type: QRecall},
		{ID: "q5", Type: QPrediction},
	}
	first := interleaveQuestions(questions)
	for i := 0; i < 5; i++ {
		if got := interleaveQuestions(questions); !slicesEqual(got, first) {
			t.Fatalf("run %d = %v, want %v (nondeterministic)", i, got, first)
		}
	}
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestDifficultyDistribution: the intended easy/medium/hard split must cover
// every question exactly once and skew toward medium — the productive-difficulty
// sweet spot for retrieval practice (not all-easy fluency, not all-hard defeat).
func TestDifficultyDistribution(t *testing.T) {
	for n := 4; n <= 10; n++ { // Quiz.Validate() bounds quizzes to 4-10 questions
		targets := difficultyTargets(n)
		easy, medium, hard := targets["easy"], targets["medium"], targets["hard"]

		if easy < 0 || medium < 0 || hard < 0 {
			t.Errorf("n=%d: negative bucket in %v", n, targets)
		}
		if easy+medium+hard != n {
			t.Errorf("n=%d: buckets sum to %d, want %d (%v)", n, easy+medium+hard, n, targets)
		}
		if medium < easy || medium < hard {
			t.Errorf("n=%d: distribution should skew medium, got %v", n, targets)
		}
	}
}

// TestInterleavingSummaryCountsRepeats: the human-readable audit string must
// report the true adjacent-repeat count, since that's what the .mjs gate and
// reviewers read.
func TestInterleavingSummaryCountsRepeats(t *testing.T) {
	questions := []Question{
		{ID: "q1", Type: QRecall}, {ID: "q2", Type: QRecall}, // one repeat
		{ID: "q3", Type: QApplication},
	}
	if got, want := interleavingSummary(questions, []string{"q1", "q2", "q3"}), "3 questions, 1 adjacent same-type pair(s)"; got != want {
		t.Errorf("summary = %q, want %q", got, want)
	}
	if got, want := interleavingSummary(questions, []string{"q1", "q3", "q2"}), "3 questions, 0 adjacent same-type pair(s)"; got != want {
		t.Errorf("summary = %q, want %q", got, want)
	}
}

// TestRunQuizStrategyStageProducesValidSequence exercises the whole stage over a
// seeded quiz.json and asserts the emitted quiz_sequence.json satisfies every
// invariant the learning_science.mjs CI gate checks: permutation, minimized
// interleaving, and a difficulty split that sums to the question count.
func TestRunQuizStrategyStageProducesValidSequence(t *testing.T) {
	course, lesson := testCourse(t)
	if err := os.MkdirAll(lesson.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	quiz := Quiz{
		Title: "Sequencing check",
		Questions: []Question{
			{ID: "q1", Type: QRecall}, {ID: "q2", Type: QRecall},
			{ID: "q3", Type: QApplication}, {ID: "q4", Type: QApplication},
			{ID: "q5", Type: QDebugging},
		},
	}
	if err := writeJSON(filepath.Join(lesson.GeneratedDir(), QuizFileName), quiz); err != nil {
		t.Fatal(err)
	}

	env, _ := runEnv(t, &fakeRouter{})
	if err := runQuizStrategyStage(context.Background(), env, course, lesson, config.Config{}); err != nil {
		t.Fatalf("quiz-strategy stage: %v", err)
	}

	var seq QuizSequence
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), QuizSequenceFileName))
	if err != nil {
		t.Fatalf("quiz_sequence.json not written: %v", err)
	}
	if err := json.Unmarshal(data, &seq); err != nil {
		t.Fatal(err)
	}

	assertPermutation(t, quiz.Questions, seq.Order)

	repeats := countAdjacentSameType(seq.Order, typeIndex(quiz.Questions))
	if floor := minAdjacentSameType(typeCounts(quiz.Questions), len(quiz.Questions)); repeats != floor {
		t.Errorf("sequence has %d adjacent same-type pair(s), want floor %d (%v)", repeats, floor, seq.Order)
	}
	sum := seq.DifficultyTargets["easy"] + seq.DifficultyTargets["medium"] + seq.DifficultyTargets["hard"]
	if sum != len(quiz.Questions) {
		t.Errorf("difficulty targets sum to %d, want %d (%v)", sum, len(quiz.Questions), seq.DifficultyTargets)
	}
}

// TestRunQuizStrategyStageNoQuiz: with no quiz to sequence the stage must
// degrade softly to an empty sequence rather than error — it's an optional
// enhancement stage.
func TestRunQuizStrategyStageNoQuiz(t *testing.T) {
	course, lesson := testCourse(t)
	if err := os.MkdirAll(lesson.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	env, _ := runEnv(t, &fakeRouter{})
	if err := runQuizStrategyStage(context.Background(), env, course, lesson, config.Config{}); err != nil {
		t.Fatalf("quiz-strategy stage with no quiz: %v", err)
	}
	var seq QuizSequence
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), QuizSequenceFileName))
	if err != nil {
		t.Fatalf("quiz_sequence.json not written: %v", err)
	}
	if err := json.Unmarshal(data, &seq); err != nil {
		t.Fatal(err)
	}
	if len(seq.Order) != 0 {
		t.Errorf("expected empty order for no quiz, got %v", seq.Order)
	}
}
