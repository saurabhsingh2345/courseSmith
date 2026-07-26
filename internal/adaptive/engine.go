package adaptive

import "math"

// DefaultBKTParams are reasonable priors for a beginner programming concept; a
// calibrated set would come from IRT/EM over collected student data.
func DefaultBKTParams() BKTParams {
	return BKTParams{PInit: 0.25, PLearn: 0.15, PSlip: 0.1, PGuess: 0.25}
}

// MasteryThreshold is the posterior P(known) at which a concept is "mastered".
const MasteryThreshold = 0.95

// EstimateBKT runs the standard BKT recurrence over a sequence of correct/
// incorrect responses and returns the posterior mastery plus a difficulty
// target that keeps the learner in the productive zone.
func EstimateBKT(p BKTParams, corrects []bool) BKTEstimate {
	pk := p.PInit
	for _, correct := range corrects {
		var posterior float64
		if correct {
			num := pk * (1 - p.PSlip)
			posterior = num / (num + (1-pk)*p.PGuess)
		} else {
			num := pk * p.PSlip
			posterior = num / (num + (1-pk)*(1-p.PGuess))
		}
		// Apply the learning transition for the next opportunity.
		pk = posterior + (1-posterior)*p.PLearn
	}
	pNext := pk*(1-p.PSlip) + (1-pk)*p.PGuess

	hint, rec := "medium", "Keep practicing at the current level."
	switch {
	case pk < 0.3:
		hint, rec = "easier", "Struggling — serve an easier item or review the prerequisite."
	case pk > 0.8:
		hint, rec = "harder", "Strong — advance to a harder item or the next concept."
	}
	return BKTEstimate{
		PKnown:         round3(pk),
		PNextCorrect:   round3(pNext),
		Mastered:       pk >= MasteryThreshold,
		DifficultyHint: hint,
		Recommendation: rec,
	}
}

// fsrsMultipliers scale stability per rating.
var fsrsMultipliers = map[string]float64{"again": 0.5, "hard": 1.2, "good": 2.5, "easy": 4.0}

// fsrsDeltas adjust difficulty per rating.
var fsrsDeltas = map[string]float64{"again": +1.0, "hard": +0.3, "good": -0.1, "easy": -0.5}

// ScheduleFSRS is a simplified spaced-repetition step (not full FSRS-5): the
// rating adjusts difficulty and multiplies stability, and the next interval is
// the new stability in days. Good enough to drive a review queue; swap for the
// real FSRS-5 weights (tools/tutor) once response history exists.
func ScheduleFSRS(req FSRSRequest) FSRSResult {
	d := req.Difficulty
	if d == 0 {
		d = 5
	}
	s := req.Stability
	if s == 0 {
		s = 1
	}
	m, ok := fsrsMultipliers[req.Rating]
	delta := fsrsDeltas[req.Rating]
	if !ok {
		m, delta = 2.5, 0 // unknown rating → treat as "good"
	}
	d = clamp(d+delta, 1, 10)
	s = math.Max(0.1, s*m)
	return FSRSResult{
		Stability:    round3(s),
		Difficulty:   round3(d),
		IntervalDays: int(math.Round(s)),
		Note:         "simplified scheduler — replace with FSRS-5 weights (tools/tutor) once response history is collected",
	}
}

// CalibrateIRT is a stub: real 2PL/3PL calibration (difficulty + discrimination
// per item) needs many students' responses and is delegated to py-irt in
// tools/tutor. This returns a neutral prior per distinct item so callers can
// integrate against the shape now.
func CalibrateIRT(obs []IRTObservation) IRTResult {
	seen := map[string]bool{}
	items := []IRTItem{}
	for _, o := range obs {
		if seen[o.QuestionID] {
			continue
		}
		seen[o.QuestionID] = true
		items = append(items, IRTItem{QuestionID: o.QuestionID, Difficulty: 0, Discrimination: 1})
	}
	return IRTResult{
		Calibrated: items,
		Note:       "STUB: neutral priors (difficulty 0, discrimination 1). Real calibration runs py-irt over pooled student data — see tools/tutor.",
	}
}

func clamp(x, lo, hi float64) float64 { return math.Max(lo, math.Min(hi, x)) }

func round3(x float64) float64 { return math.Round(x*1000) / 1000 }
