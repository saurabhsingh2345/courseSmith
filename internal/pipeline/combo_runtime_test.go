package pipeline

import (
	"strings"
	"testing"
)

func TestParseRequestedRuntime(t *testing.T) {
	cases := []struct {
		brief string
		want  int
		ok    bool
	}{
		// The brief that started this: an en-dash range, which is how people
		// actually type it, and the exact string that shipped as 9:43.
		{"Create a 25–30 minute beginner-friendly introduction video.", 27 * 60, true},
		{"Create a 25-30 minute intro", 27 * 60, true},
		{"a 2 minute explainer", 120, true},
		{"90 seconds on why indexes matter", 90, true},
		{"about 10 min", 600, true},
		{"1 hour of material", 3600, true},
		{"45 sec hook", 45, true},
		{"Two minutes", 0, false}, // spelled out; not worth guessing at
		// A bare number is far more often a count than a duration, and reading
		// "13 tools" as thirteen seconds would be worse than reading nothing.
		{"covers 13 tools including Webflow and Bubble", 0, false},
		{"an introduction to no-code tools", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		got, ok := ParseRequestedRuntime(c.brief)
		if ok != c.ok || got != c.want {
			t.Errorf("ParseRequestedRuntime(%q) = %d, %v; want %d, %v", c.brief, got, ok, c.want, c.ok)
		}
	}
}

// The whole point: the ask that produced 9:43 must now produce a plan that adds
// up to roughly what was asked for.
func TestTheTwentyFiveMinuteAskIsBudgeted(t *testing.T) {
	requested, ok := ParseRequestedRuntime("Create a 25–30 minute beginner-friendly introduction video.")
	if !ok {
		t.Fatal("the runtime was not read out of the brief")
	}
	b := BudgetRuntime(requested, 5)

	if b.Segments <= 5 {
		t.Errorf("budgeted %d segments for %ds; five segments cannot carry it", b.Segments, requested)
	}
	if b.Segments > maxComboSegments {
		t.Errorf("budgeted %d segments, over the %d cap", b.Segments, maxComboSegments)
	}
	if b.PerSegmentSec > maxSnippetTargetSec {
		t.Errorf("per-segment %ds exceeds the %ds template ceiling", b.PerSegmentSec, maxSnippetTargetSec)
	}
	// Within a couple of minutes of the ask, rather than a third of it.
	if got := b.Achievable(); got < requested*8/10 {
		t.Errorf("achievable %ds is less than 80%% of the %ds ask", got, requested)
	}
}

// An ask beyond what the segment cap can carry must SAY so. Silently delivering a
// fraction is the original bug, and a budget that quietly clamps is the same bug
// wearing a helper function.
func TestAnImpossibleAskReportsItsShortfall(t *testing.T) {
	// Two hours: far past maxComboSegments x maxSnippetTargetSec.
	b := BudgetRuntime(2*3600, 5)
	if b.Shortfall <= 0 {
		t.Fatal("a two-hour ask reported no shortfall")
	}
	if b.Segments != maxComboSegments {
		t.Errorf("an impossible ask used %d segments, not the %d available", b.Segments, maxComboSegments)
	}
	if b.PerSegmentSec != maxSnippetTargetSec {
		t.Errorf("an impossible ask used %ds segments, not the %ds ceiling", b.PerSegmentSec, maxSnippetTargetSec)
	}
	if desc := b.Describe(); !strings.Contains(desc, "short of the ask") {
		t.Errorf("the description does not admit the shortfall: %q", desc)
	}
}

// A satisfiable ask must not claim a shortfall, or the warning stops meaning
// anything.
func TestASatisfiableAskReportsNoShortfall(t *testing.T) {
	for _, sec := range []int{180, 300, 600, 900} {
		b := BudgetRuntime(sec, 5)
		if b.Shortfall != 0 {
			t.Errorf("%ds reported a %ds shortfall but %d segments of %ds is %ds",
				sec, b.Shortfall, b.Segments, b.PerSegmentSec, b.Achievable())
		}
		if strings.Contains(b.Describe(), "short of the ask") {
			t.Errorf("%ds warned about a shortfall it does not have", sec)
		}
	}
}

// No stated length means no budget, and the segments keep their template
// defaults — the behaviour every brief without a runtime relies on.
func TestNoRuntimeMeansNoBudget(t *testing.T) {
	b := BudgetRuntime(0, 5)
	if b.PerSegmentSec != 0 || b.Segments != 0 {
		t.Errorf("an unstated runtime produced a budget: %+v", b)
	}
	if b.Describe() != "" {
		t.Errorf("an unstated runtime described itself: %q", b.Describe())
	}
}

// A budget must never produce a spec that fails validation the moment it is
// written. MinTargetSec is arithmetic — a template needing eight beats cannot be
// planned in twenty seconds — and Validate rejects a TargetSec below it.
func TestBudgetedTargetsAlwaysValidate(t *testing.T) {
	for _, requested := range []int{60, 120, 300, 900, 1620, 7200} {
		b := BudgetRuntime(requested, 5)
		for _, name := range SnippetTemplateNames() {
			target := segmentTargetFor(name, b.PerSegmentSec)
			spec := ComboSpec{
				ID: "budgeted",
				Segments: []ComboSegment{
					{ID: "a", Template: name, Prompt: "the first part", TargetSec: target},
					{ID: "b", Template: "myth", Prompt: "the belief this corrects", TargetSec: segmentTargetFor("myth", b.PerSegmentSec)},
				},
			}
			if err := spec.Validate(); err != nil {
				t.Errorf("a %ds ask budgeted %s to %ds, which does not validate: %v", requested, name, target, err)
			}
		}
	}
}

// A very short ask must not fall below the snippet floor.
func TestAVeryShortAskClampsToTheFloor(t *testing.T) {
	b := BudgetRuntime(12, 5)
	if b.PerSegmentSec < minSnippetTargetSec {
		t.Errorf("per-segment %ds is below the %ds floor", b.PerSegmentSec, minSnippetTargetSec)
	}
	if b.Segments < minComboSegments {
		t.Errorf("budgeted %d segments, below the %d minimum", b.Segments, minComboSegments)
	}
}

func TestDurationWords(t *testing.T) {
	for _, c := range []struct {
		sec  int
		want string
	}{
		{45, "45s"}, {60, "1m"}, {90, "1m30s"}, {1620, "27m"},
	} {
		if got := durationWords(c.sec); got != c.want {
			t.Errorf("durationWords(%d) = %q, want %q", c.sec, got, c.want)
		}
	}
}
