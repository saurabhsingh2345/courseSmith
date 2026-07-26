// Package adaptive is the learning-science layer (workstream D): Bayesian
// Knowledge Tracing for mastery, a simplified spaced-repetition scheduler for
// review timing, and (stubbed) IRT calibration. The pure functions in
// engine.go are the single source of truth — the coursesmith-tutor HTTP
// service, the in-process Mock, and any Go caller all run the same math.
package adaptive

// BKTParams are the four classic Bayesian Knowledge Tracing parameters.
type BKTParams struct {
	PInit  float64 `json:"p_init"`  // P(known) before any practice
	PLearn float64 `json:"p_learn"` // P(unknown→known) per opportunity
	PSlip  float64 `json:"p_slip"`  // P(wrong | known)
	PGuess float64 `json:"p_guess"` // P(right | unknown)
}

// BKTEstimate is the posterior mastery state after a response sequence.
type BKTEstimate struct {
	PKnown         float64 `json:"p_known"`         // posterior mastery after all responses
	PNextCorrect   float64 `json:"p_next_correct"`  // predicted P(correct) on the next item
	Mastered       bool    `json:"mastered"`        // p_known >= 0.95
	DifficultyHint string  `json:"difficulty_hint"` // easier | medium | harder
	Recommendation string  `json:"recommendation"`  // human-facing next step
}

// FSRSRequest is one review outcome to schedule the next review from.
type FSRSRequest struct {
	Rating     string  `json:"rating"`     // again | hard | good | easy
	Stability  float64 `json:"stability"`  // days; 0 → first review
	Difficulty float64 `json:"difficulty"` // 1..10; 0 → default 5
}

// FSRSResult is the updated memory state and the next interval.
type FSRSResult struct {
	Stability    float64 `json:"stability"`
	Difficulty   float64 `json:"difficulty"`
	IntervalDays int     `json:"interval_days"`
	Note         string  `json:"note"`
}

// IRTObservation is one student's correct/incorrect answer to a question.
type IRTObservation struct {
	QuestionID string `json:"question_id"`
	Correct    bool   `json:"correct"`
}

// IRTItem is a calibrated question: difficulty and discrimination.
type IRTItem struct {
	QuestionID     string  `json:"question_id"`
	Difficulty     float64 `json:"difficulty"`     // [-2, 2]
	Discrimination float64 `json:"discrimination"` // [0, 2]
}

// IRTResult is the calibration output for a pool of items.
type IRTResult struct {
	Calibrated []IRTItem `json:"calibrated"`
	Note       string    `json:"note"`
}
