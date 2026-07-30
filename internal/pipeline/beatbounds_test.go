package pipeline

import "testing"

// A template that declares MaxBeats is saying its beat count is a property of
// its content — a list of five things needs five beats plus an opener and a
// summary, whatever the runtime. That declaration was unreachable: suggest+2
// capped constellation at 6 beats when its shape needs 7, so every run failed
// its own validator, spent three correction rounds, and shipped the closest
// draft with a card missing.
func TestDeclaredCeilingIsReachableWhenFundable(t *testing.T) {
	cases := []struct {
		name        string
		target      int
		ceiling     int
		wantAtLeast int
	}{
		// centre + five spokes + whole
		{"constellation at its default", 55, 8, 7},
		// opener + five cards + summary
		{"rundown at its default", 60, 9, 7},
		// the question + four conditions + the call
		{"verdict at its default", 55, 9, 7},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			words, _, _ := wordBudget(c.target, 174)
			_, maxBeats, _, _ := beatBounds(words, c.ceiling)
			if maxBeats < c.wantAtLeast {
				t.Errorf("maxBeats = %d at %ds, want at least %d — the template's own shape is illegal",
					maxBeats, c.target, c.wantAtLeast)
			}
			if maxBeats > c.ceiling {
				t.Errorf("maxBeats = %d exceeds the declared ceiling %d", maxBeats, c.ceiling)
			}
		})
	}
}

// The widening must not reach the ceiling by starving every beat. Eight beats of
// ten words is a slideshow, which is the failure suggest+2 was protecting
// against — so fundability is judged at a substantial beat, not the hard minimum.
func TestWideningKeepsBeatsSubstantial(t *testing.T) {
	for _, target := range []int{45, 55, 60, 95, 120, 180} {
		words, _, _ := wordBudget(target, 174)
		// A generous ceiling, so only the budget limits the count.
		_, maxBeats, _, _ := beatBounds(words, 12)
		if maxBeats == 0 {
			t.Fatalf("%ds produced no beats", target)
		}
		if perBeat := words / maxBeats; perBeat < leanWordsPerBeat {
			t.Errorf("%ds allows %d beats, which is %d words each — below the %d-word floor for a beat that is a thought",
				target, maxBeats, perBeat, leanWordsPerBeat)
		}
	}
}

// Only ever widens. Every runtime that worked before must keep at least the range
// it had, or this fix trades one contradiction for another at the short end —
// which is exactly the history beatBounds already carries.
func TestWideningNeverNarrows(t *testing.T) {
	for _, target := range []int{10, 15, 20, 30, 45, 55, 60, 90, 120, 180} {
		for _, ceiling := range []int{0, 7, 8, 9, 10, 12} {
			words, _, _ := wordBudget(target, 174)
			minBeats, maxBeats, suggest, _ := beatBounds(words, ceiling)

			// The pre-fix maximum, recomputed here so the guarantee is asserted
			// rather than assumed.
			effectiveCeiling := ceiling
			if effectiveCeiling <= 0 {
				effectiveCeiling = maxSnippetBeats
			}
			was := min(max(suggest+2, minBeats), effectiveCeiling)
			if maxBeats < was {
				t.Errorf("target=%ds ceiling=%d: maxBeats narrowed from %d to %d", target, ceiling, was, maxBeats)
			}
			if minBeats > maxBeats {
				t.Errorf("target=%ds ceiling=%d: range is inverted (%d-%d)", target, ceiling, minBeats, maxBeats)
			}
			if minBeats < floorSnippetBeats {
				t.Errorf("target=%ds ceiling=%d: minBeats %d is below the floor", target, ceiling, minBeats)
			}
		}
	}
}

// Short clips are the case the whole budget-derived range was built for, and the
// documented table is the contract. A 10- or 20-second clip must still be two
// beats, or the arithmetic that made those runtimes expressible is undone.
func TestShortClipsKeepTheirDocumentedSuggestion(t *testing.T) {
	for _, c := range []struct{ target, wantSuggest int }{
		{10, 2}, {20, 2}, {45, 3},
	} {
		words, _, _ := wordBudget(c.target, 174)
		_, _, suggest, _ := beatBounds(words, 0)
		if suggest != c.wantSuggest {
			t.Errorf("%ds suggests %d beats, want %d", c.target, suggest, c.wantSuggest)
		}
	}
}

// A hand-built plan has no budget to size against and must not divide by zero.
func TestBeatBoundsWithNoBudget(t *testing.T) {
	minBeats, maxBeats, suggest, perBeat := beatBounds(0, 0)
	if minBeats != floorSnippetBeats || maxBeats != maxSnippetBeats {
		t.Errorf("no-budget bounds = %d-%d", minBeats, maxBeats)
	}
	if suggest <= 0 || perBeat <= 0 {
		t.Errorf("no-budget suggest=%d perBeat=%d", suggest, perBeat)
	}
}
