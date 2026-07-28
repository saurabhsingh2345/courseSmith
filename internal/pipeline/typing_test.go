package pipeline

import (
	"strings"
	"testing"
)

// gapsFor returns the millisecond gap before each character, which is what the
// rhythm actually is — the absolute offsets are just its running total.
func gapsFor(t *testing.T, code string, budgetMs int) []int {
	t.Helper()
	sched := KeystrokeSchedule(code, budgetMs)
	if len(sched) != len([]rune(code)) {
		t.Fatalf("schedule has %d entries for %d runes", len(sched), len([]rune(code)))
	}
	gaps := make([]int, len(sched))
	prev := 0
	for i, ms := range sched {
		gaps[i] = ms - prev
		prev = ms
	}
	return gaps
}

// indexOfRune finds the nth occurrence of r in code, as a rune index.
func indexOfRune(code string, r rune, n int) int {
	seen := 0
	for i, c := range []rune(code) {
		if c == r {
			seen++
			if seen == n {
				return i
			}
		}
	}
	return -1
}

func TestKeystrokeScheduleFillsItsBudget(t *testing.T) {
	// A file that finishes early leaves a caret blinking on a static screen for
	// the rest of the beat, which is the single most obvious way this can look
	// wrong. The last keystroke is the end of the window, always.
	for _, code := range []string{
		"print('hi')",
		"def f(x):\n    return x * 2\n\nprint(f(21))",
		strings.Repeat("x = 1\n", 40),
	} {
		sched := KeystrokeSchedule(code, 4000)
		if got := sched[len(sched)-1]; got != 4000 {
			t.Errorf("last keystroke = %d, want the full 4000ms budget (code %q)", got, code[:min(12, len(code))])
		}
	}
}

func TestKeystrokeScheduleIsMonotonic(t *testing.T) {
	code := "def greet(name):\n    msg = f'hello {name}'\n    print(msg)\n"
	sched := KeystrokeSchedule(code, 6000)
	for i := 1; i < len(sched); i++ {
		if sched[i] < sched[i-1] {
			t.Fatalf("schedule went backwards at %d: %d then %d", i, sched[i-1], sched[i])
		}
	}
}

func TestIndentationIsNearlyFree(t *testing.T) {
	// Editors insert indentation; a clip that types four spaces one at a time
	// has never watched anyone code.
	code := "if x:\n    y = 1\n"
	gaps := gapsFor(t, code, 5000)
	runes := []rune(code)

	// The four spaces after the first newline.
	nl := indexOfRune(code, '\n', 1)
	var indentGap int
	for i := nl + 1; i < len(runes) && runes[i] == ' '; i++ {
		indentGap += gaps[i]
	}
	// Compare against a single ordinary character: `y`.
	y := indexOfRune(code, 'y', 1)
	if indentGap >= gaps[y] {
		t.Errorf("four spaces of indent took %dms, an ordinary character %dms — indent should be nearly free",
			indentGap, gaps[y])
	}
}

func TestBlockOpenerPausesLongest(t *testing.T) {
	// The pause before writing the body of something is the longest one in a
	// file, because it is where the next thought starts.
	code := "x = 1\ndef f():\n    return 2\n"
	gaps := gapsFor(t, code, 8000)

	afterAssign := indexOfRune(code, '\n', 1) // ends "x = 1"
	afterColon := indexOfRune(code, '\n', 2)  // ends "def f():"

	if gaps[afterColon] <= gaps[afterAssign] {
		t.Errorf("newline after a block opener = %dms, after a plain statement = %dms — the opener should pause longer",
			gaps[afterColon], gaps[afterAssign])
	}
}

func TestNewlinePausesLongerThanACharacter(t *testing.T) {
	code := "ab\ncd"
	gaps := gapsFor(t, code, 3000)
	nl := indexOfRune(code, '\n', 1)
	c := indexOfRune(code, 'c', 1)
	if gaps[nl] <= gaps[c] {
		t.Errorf("newline gap %dms is not longer than a character's %dms", gaps[nl], gaps[c])
	}
}

func TestAutoClosedBracketsAreNearlyFree(t *testing.T) {
	// `()` is one keystroke in a real editor: the closer arrives with the
	// opener. A clip that types it as two is a clip that is not imitating an
	// editor, which is the entire premise of these templates.
	code := "f()\nf(x)"
	gaps := gapsFor(t, code, 4000)

	autoClose := indexOfRune(code, ')', 1)  // the ')' in "f()", auto-inserted
	typedClose := indexOfRune(code, ')', 2) // the ')' in "f(x)", genuinely typed

	if gaps[autoClose] >= gaps[typedClose] {
		t.Errorf("auto-closed ')' took %dms and a typed ')' took %dms — the auto-closed one should be nearly free",
			gaps[autoClose], gaps[typedClose])
	}
}

func TestScheduleIsDeterministic(t *testing.T) {
	// The renderer animates from this and a separate process synthesises a
	// click track from it. If the two disagree by even one entry the clicks
	// drift off the characters, so this is not a nicety.
	code := "for i in range(10):\n    print(i)\n"
	a := KeystrokeSchedule(code, 5000)
	b := KeystrokeSchedule(code, 5000)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("schedule differs between calls at %d: %d vs %d", i, a[i], b[i])
		}
	}
}

func TestScheduleEdgeCases(t *testing.T) {
	if got := KeystrokeSchedule("", 1000); got != nil {
		t.Errorf("empty code should have no schedule, got %v", got)
	}
	// A window too short to type in: everything is already there. Better than a
	// schedule that keeps typing after the beat has ended.
	for _, budget := range []int{0, -500} {
		sched := KeystrokeSchedule("abc", budget)
		for i, ms := range sched {
			if ms != 0 {
				t.Errorf("budget %d: keystroke %d at %dms, want 0", budget, i, ms)
			}
		}
	}
}

func TestKeystrokeTimesMsShifts(t *testing.T) {
	got := KeystrokeTimesMs([]int{0, 100, 250}, 12000)
	want := []int{12000, 12100, 12250}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("KeystrokeTimesMs[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}
