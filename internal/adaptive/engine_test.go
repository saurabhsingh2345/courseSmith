package adaptive

import "testing"

func TestEstimateBKTMonotonic(t *testing.T) {
	p := DefaultBKTParams()
	// More correct answers must not decrease mastery; a wrong answer must not
	// increase it relative to a correct one at the same step.
	none := EstimateBKT(p, nil).PKnown
	oneRight := EstimateBKT(p, []bool{true}).PKnown
	fiveRight := EstimateBKT(p, []bool{true, true, true, true, true}).PKnown
	oneWrong := EstimateBKT(p, []bool{false}).PKnown

	if !(none < oneRight && oneRight < fiveRight) {
		t.Errorf("mastery not increasing with correct answers: none=%.3f one=%.3f five=%.3f", none, oneRight, fiveRight)
	}
	if oneWrong >= oneRight {
		t.Errorf("a wrong answer (%.3f) should give lower mastery than a right one (%.3f)", oneWrong, oneRight)
	}
	if none != round3(p.PInit) {
		t.Errorf("no responses should return the prior %.3f, got %.3f", p.PInit, none)
	}
}

func TestEstimateBKTDifficultyHints(t *testing.T) {
	p := DefaultBKTParams()
	// A long correct streak → strong → "harder"; a long wrong streak → "easier".
	strong := EstimateBKT(p, []bool{true, true, true, true, true, true, true, true})
	if strong.DifficultyHint != "harder" {
		t.Errorf("strong learner hint = %q, want harder (p_known=%.3f)", strong.DifficultyHint, strong.PKnown)
	}
	weak := EstimateBKT(p, []bool{false, false, false, false, false})
	if weak.DifficultyHint != "easier" {
		t.Errorf("weak learner hint = %q, want easier (p_known=%.3f)", weak.DifficultyHint, weak.PKnown)
	}
	if strong.PNextCorrect <= weak.PNextCorrect {
		t.Errorf("strong p_next_correct (%.3f) should exceed weak (%.3f)", strong.PNextCorrect, weak.PNextCorrect)
	}
}

func TestEstimateBKTMastered(t *testing.T) {
	p := DefaultBKTParams()
	long := make([]bool, 40)
	for i := range long {
		long[i] = true
	}
	est := EstimateBKT(p, long)
	if !est.Mastered || est.PKnown < MasteryThreshold {
		t.Errorf("40 correct answers should be mastered; got mastered=%v p_known=%.3f", est.Mastered, est.PKnown)
	}
}

func TestScheduleFSRS(t *testing.T) {
	// Ratings order the next interval: again < hard < good < easy.
	again := ScheduleFSRS(FSRSRequest{Rating: "again"})
	hard := ScheduleFSRS(FSRSRequest{Rating: "hard"})
	good := ScheduleFSRS(FSRSRequest{Rating: "good"})
	easy := ScheduleFSRS(FSRSRequest{Rating: "easy"})
	if !(again.Stability < hard.Stability && hard.Stability < good.Stability && good.Stability < easy.Stability) {
		t.Errorf("stability not ordered by rating: again=%.2f hard=%.2f good=%.2f easy=%.2f",
			again.Stability, hard.Stability, good.Stability, easy.Stability)
	}
	// Defaults fill in: stability 0 → 1, difficulty 0 → 5.
	if easy.Difficulty < 1 || easy.Difficulty > 10 {
		t.Errorf("difficulty out of range: %.2f", easy.Difficulty)
	}
	// An unknown rating is treated as "good".
	unknown := ScheduleFSRS(FSRSRequest{Rating: "banana"})
	if unknown.IntervalDays != good.IntervalDays {
		t.Errorf("unknown rating interval %d, want good's %d", unknown.IntervalDays, good.IntervalDays)
	}
}

func TestCalibrateIRT(t *testing.T) {
	res := CalibrateIRT([]IRTObservation{
		{QuestionID: "q1", Correct: true},
		{QuestionID: "q1", Correct: false},
		{QuestionID: "q2", Correct: true},
	})
	if len(res.Calibrated) != 2 {
		t.Fatalf("expected 2 distinct items, got %d", len(res.Calibrated))
	}
	for _, it := range res.Calibrated {
		if it.Difficulty != 0 || it.Discrimination != 1 {
			t.Errorf("item %q should have neutral priors, got d=%.2f a=%.2f", it.QuestionID, it.Difficulty, it.Discrimination)
		}
	}
}
