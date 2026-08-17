package pipeline

import (
	"fmt"
	"testing"
)

// The guard against a template that cannot satisfy its own instructions.
//
// This is the most expensive class of bug this codebase has, and it has now been
// introduced three times: once at short runtimes, once at long ones (both recorded
// in beatBounds' comments), and once more by the v9 showroom batch, where `duel`
// declared MaxBeats: 7 while its validator accepts exactly five beats. At 90
// seconds the shared beat budget demanded at least six and the validator refused
// the sixth, so no plan could ever pass.
//
// What makes it expensive rather than merely broken is the correction loop. A
// rejection is the signal that escalates reasoning effort (see llmjson.go), and
// reasoning tokens bill as completion — so an unsatisfiable plan does not fail
// fast and cheap, it burns every correction round at the most expensive setting
// available and then salvages a draft that breaks its own rules. From the outside
// it looks like a run wedged in `plan` for minutes, which is exactly what was
// reported.
//
// So: for every template whose beat count is a property of its SHAPE, the number
// of beats its validator can accept is written down here, and the test asserts
// that the beat budget never asks for more than that at any runtime the template
// admits. A template that grows a new beat kind, or raises an item ceiling, has to
// come back to this table — and if it forgets, this fails instead of the invoice.
//
// The table is hand-maintained on purpose. Deriving it would mean asking each
// validator how many beats it would accept, which is the same question by a
// harder route; the value here is that a human wrote down what the shape IS, next
// to the arithmetic that has to agree with it.
var beatShapeMax = map[string]struct {
	max int
	why string
}{
	"opener":     {3, "ground + promise + mark"},
	"duel":       {5, "pair + one beat per side + bars + call"},
	"approval":   {5, "ask + one beat per answer (max 3) + pick"},
	"spotlight":  {6, "card + one beat per claim (max 4) + all"},
	"patch":      {6, "file + one beat per hunk (max 4) + tally"},
	"cards":      {7, "row + one beat per card (max 5) + all"},
	"changeplan": {8, "rail + one beat per file (max 6) + all"},
}

// The runtimes a caller can actually ask for. `snippet new --seconds` is free-form,
// so the range is wide on purpose: three minutes is not a silly request and it is
// where the last two versions of this bug lived.
var probeRuntimes = []int{10, 15, 20, 25, 30, 35, 45, 50, 60, 75, 90, 120, 150, 180, 240}

func TestBeatBudgetNeverExceedsTheShape(t *testing.T) {
	for name, shape := range beatShapeMax {
		tpl := SnippetTemplates[name]
		if tpl == nil {
			t.Errorf("beatShapeMax names %q, which is not a registered template", name)
			continue
		}
		// The declared ceiling must BE the shape's maximum. This is the actual fix
		// for the duel bug: beatBounds clamps to MaxBeats, so a ceiling equal to
		// the shape makes the contradiction unrepresentable rather than merely
		// unlikely.
		if tpl.MaxBeats != shape.max {
			t.Errorf("%s declares MaxBeats %d but its shape is %d beats (%s). beatBounds clamps the beat range to MaxBeats, so any ceiling above the shape lets a long runtime demand a beat the validator will refuse — forever, at escalating reasoning cost",
				name, tpl.MaxBeats, shape.max, shape.why)
		}
		for _, sec := range probeRuntimes {
			if sec < tpl.MinTargetSec {
				continue
			}
			want, _, _ := wordBudget(sec, 150)
			minBeats, maxBeats, _, _ := beatBounds(want, templateBeatCeiling(name), templateIdealWords(name))
			if minBeats > shape.max {
				t.Errorf("%s at %ds: the budget demands at least %d beats, but the validator accepts at most %d (%s). No plan can satisfy both, and every correction round pays for reasoning to discover that",
					name, sec, minBeats, shape.max, shape.why)
			}
			if maxBeats > shape.max {
				t.Errorf("%s at %ds: the budget allows up to %d beats, above the %d the validator accepts (%s) — a model that takes the offer gets rejected",
					name, sec, maxBeats, shape.max, shape.why)
			}
		}
	}
}

// The other half of the same contradiction: the beats the shape REQUIRES have to
// be fundable without blowing the per-beat word ceiling.
//
// duel needs five beats whatever the runtime. At a long enough runtime the word
// budget divided over five beats exceeds what one beat may hold, and the plan is
// unsatisfiable from the other direction — the model is told to write 90 words in
// a beat that may hold 60. Nothing had checked this, and it is the failure mode a
// future "why not a five-minute duel" request walks straight into.
func TestRequiredBeatsCanHoldTheWordBudget(t *testing.T) {
	// The minimum beats each shape forces, which is not the same as its maximum:
	// duel always needs five, but cards needs only four (row + 2 cards + all).
	required := map[string]int{
		"opener":     2, // ground + promise; the mark is optional
		"duel":       5,
		"approval":   4, // ask + 2 answers + pick
		"spotlight":  4, // card + 2 claims + all
		"patch":      3, // file + 1 hunk + tally
		"cards":      4, // row + 2 cards + all
		"changeplan": 4, // rail + 2 files + all
	}
	for name, need := range required {
		tpl := SnippetTemplates[name]
		if tpl == nil {
			continue
		}
		if tpl.MaxTargetSec == 0 {
			t.Errorf("%s has a fixed shape (%d beats) but declares no MaxTargetSec, so a long enough request reaches the correction loop instead of being refused at the door",
				name, beatShapeMax[name].max)
			continue
		}
		for _, sec := range probeRuntimes {
			// Only runtimes the spec validator ACCEPTS have to be satisfiable. That
			// is the fix: a request outside the range now fails before any token is
			// spent, so the arithmetic only has to hold inside it.
			if sec < tpl.MinTargetSec || sec > tpl.MaxTargetSec {
				continue
			}
			want, _, _ := wordBudget(sec, 150)
			// The most beats the shape can offer is what shares the load.
			perBeat := want / beatShapeMax[name].max
			if perBeat > maxWordsPerBeat {
				t.Errorf("%s at %ds (inside its %d-%ds range): %d words over its %d-beat maximum is %d words a beat, above the %d ceiling — MaxTargetSec is too high",
					name, sec, tpl.MinTargetSec, tpl.MaxTargetSec, want, beatShapeMax[name].max, perBeat, maxWordsPerBeat)
			}
			_ = need
		}
		// And the ceiling has to be honest in the other direction: set it too low
		// and the template silently refuses runtimes it could have handled.
		if over := tpl.MaxTargetSec + 30; true {
			want, _, _ := wordBudget(over, 150)
			if want/beatShapeMax[name].max <= maxWordsPerBeat {
				t.Errorf("%s could satisfy %ds (%d words over %d beats) but MaxTargetSec refuses anything past %ds — the ceiling is lower than the arithmetic requires",
					name, over, want, beatShapeMax[name].max, tpl.MaxTargetSec)
			}
		}
	}
}

// A sanity print, so a failure above can be read against the whole picture rather
// than one line of it. Skipped unless -v.
func TestShowBeatBudgets(t *testing.T) {
	if !testing.Verbose() {
		t.Skip("run with -v to see the table")
	}
	for name, shape := range beatShapeMax {
		tpl := SnippetTemplates[name]
		fmt.Printf("\n%-11s shape=%d (%s)  MaxBeats=%d\n", name, shape.max, shape.why, tpl.MaxBeats)
		for _, sec := range probeRuntimes {
			if sec < tpl.MinTargetSec {
				continue
			}
			want, _, _ := wordBudget(sec, 150)
			mn, mx, sug, wpb := beatBounds(want, templateBeatCeiling(name), templateIdealWords(name))
			fmt.Printf("   %3ds words=%3d beats %d-%d (suggest %d @ %d/beat)\n", sec, want, mn, mx, sug, wpb)
		}
	}
}
