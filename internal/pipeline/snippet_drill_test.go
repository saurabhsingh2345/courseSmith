package pipeline

import (
	"strings"
	"testing"
)

const drNarration = "Two of these are mistakes people genuinely make, and only one of them is quietly right."

func drillPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "drill",
		Title:    "Why floats miss by a hair",
		Drill: &DrillSpec{
			Question: "Why does 0.1 plus 0.2 not equal 0.3?",
			Options: []string{
				"Computers round randomly",
				"The maths library has a bug",
				"One tenth has no exact binary form",
				"Floats only keep seven digits",
			},
			Answer: 2,
			Why:    "Base two cannot write one tenth exactly, so the stored value is already slightly off",
		},
		Beats: []SnippetBeat{
			{ID: "ask", Heading: "The classic surprise", Narration: drNarration, Drill: &DrillBeat{Show: "ask"}},
			{ID: "not-random", Heading: "Nothing is random", Narration: drNarration, Drill: &DrillBeat{Show: "eliminate", At: 0}},
			{ID: "not-a-bug", Heading: "Not a bug", Narration: drNarration, Drill: &DrillBeat{Show: "eliminate", At: 1}},
			{ID: "not-digits", Heading: "Not about digits", Narration: drNarration, Drill: &DrillBeat{Show: "eliminate", At: 3}},
			{ID: "reveal", Heading: "What survives", Narration: drNarration, Drill: &DrillBeat{Show: "reveal"}},
			{ID: "why", Heading: "Why it holds", Narration: drNarration, Drill: &DrillBeat{Show: "why"}},
		},
	}
	p.targetWords = 6 * 40
	return p
}

func TestDrillPlanAccepted(t *testing.T) {
	if err := validateDrillPlan(drillPlan()); err != nil {
		t.Fatalf("a well-formed drill plan was rejected: %v", err)
	}
}

func TestDrillRejectsAMissingQuestion(t *testing.T) {
	p := drillPlan()
	p.Drill.Question = "  "
	if err := validateDrillPlan(p); err == nil {
		t.Fatal("a board with no question was accepted, and the plates are then answers to nothing")
	}
}

func TestDrillRejectsTooFewOptions(t *testing.T) {
	p := drillPlan()
	p.Drill.Options = p.Drill.Options[:2]
	err := validateDrillPlan(p)
	if err == nil {
		t.Fatal("a two-option board was accepted, and two options is a coin flip")
	}
	if !strings.Contains(err.Error(), "coin flip") {
		t.Fatalf("the error does not explain why: %v", err)
	}
}

func TestDrillRejectsTooManyOptions(t *testing.T) {
	p := drillPlan()
	p.Drill.Options = append(p.Drill.Options, "Nobody would ever pick this")
	if err := validateDrillPlan(p); err == nil {
		t.Fatal("a five-option board was accepted, and striking a distractor nobody would pick teaches nothing")
	}
}

func TestDrillRejectsAnAnswerOffTheBoard(t *testing.T) {
	p := drillPlan()
	p.Drill.Answer = 9
	err := validateDrillPlan(p)
	if err == nil {
		t.Fatal("an answer index past the last plate was accepted")
	}
	if !strings.Contains(err.Error(), "options 0-3") {
		t.Fatalf("the error does not quote the range that exists: %v", err)
	}
}

func TestDrillRequiresOpeningOnTheAsk(t *testing.T) {
	p := drillPlan()
	p.Beats[0].Drill = &DrillBeat{Show: "eliminate", At: 0}
	if err := validateDrillPlan(p); err == nil {
		t.Fatal("a clip that strikes a plate before the viewer has read it was accepted")
	}
}

func TestDrillRejectsAskingTwice(t *testing.T) {
	p := drillPlan()
	p.Beats[2].Drill = &DrillBeat{Show: "ask"}
	err := validateDrillPlan(p)
	if err == nil {
		t.Fatal("a clip that re-asks the question was accepted")
	}
	if !strings.Contains(err.Error(), "un-strikes") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// The index arithmetic a model gets wrong: options get shuffled between drafts
// and the answer index goes stale, so the clip strikes its own answer.
func TestDrillRejectsStrikingTheAnswer(t *testing.T) {
	p := drillPlan()
	p.Beats[1].Drill = &DrillBeat{Show: "eliminate", At: 2}
	err := validateDrillPlan(p)
	if err == nil {
		t.Fatal("a strike through the answer plate was accepted")
	}
	if !strings.Contains(err.Error(), "ANSWER") {
		t.Fatalf("the error does not say the plate is the answer: %v", err)
	}
	if !strings.Contains(err.Error(), "One tenth has no exact binary form") {
		t.Fatalf("the error does not quote the plate it struck: %v", err)
	}
}

func TestDrillRejectsStrikingAnOptionTwice(t *testing.T) {
	p := drillPlan()
	p.Beats[2].Drill = &DrillBeat{Show: "eliminate", At: 0}
	err := validateDrillPlan(p)
	if err == nil {
		t.Fatal("a second strike on a struck plate was accepted")
	}
	if !strings.Contains(err.Error(), "corpse") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDrillRejectsAnEliminateOffTheBoard(t *testing.T) {
	p := drillPlan()
	p.Beats[1].Drill = &DrillBeat{Show: "eliminate", At: 9}
	if err := validateDrillPlan(p); err == nil {
		t.Fatal("a strike on a plate that does not exist was accepted")
	}
}

func TestDrillRejectsTwoReveals(t *testing.T) {
	p := drillPlan()
	p.Beats[3].Drill = &DrillBeat{Show: "reveal"}
	err := validateDrillPlan(p)
	if err == nil {
		t.Fatal("two reveal beats were accepted, and two reveals is two right answers")
	}
	if !strings.Contains(err.Error(), "want exactly 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDrillRejectsNoRevealAtAll(t *testing.T) {
	p := drillPlan()
	p.Beats = p.Beats[:4]
	p.targetWords = 4 * 40
	err := validateDrillPlan(p)
	if err == nil {
		t.Fatal("a check-question the clip never answers was accepted")
	}
	if !strings.Contains(err.Error(), "want exactly 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDrillRejectsAWhyBeforeTheReveal(t *testing.T) {
	p := drillPlan()
	p.Beats[4].Drill = &DrillBeat{Show: "why"}
	p.Beats[5].Drill = &DrillBeat{Show: "reveal"}
	err := validateDrillPlan(p)
	if err == nil {
		t.Fatal("a reason landing before the answer was accepted")
	}
	if !strings.Contains(err.Error(), "has not seen made") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDrillRejectsTwoWhys(t *testing.T) {
	p := drillPlan()
	p.Beats = append(p.Beats, SnippetBeat{
		ID: "why-again", Heading: "Saying it again", Narration: drNarration,
		Drill: &DrillBeat{Show: "why"},
	})
	err := validateDrillPlan(p)
	if err == nil {
		t.Fatal("two why beats were accepted, and the reason is one line that lands once")
	}
	if !strings.Contains(err.Error(), "2 why beats") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDrillRejectsAWhyBeatWithNoWhyLine(t *testing.T) {
	p := drillPlan()
	p.Drill.Why = "   "
	err := validateDrillPlan(p)
	if err == nil {
		t.Fatal("a why beat with nothing to say was accepted")
	}
	if !strings.Contains(err.Error(), "drop the beat") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Moving the answer index would silently change which plate the clip calls
// correct, so normalize leaves it alone and lets the validator argue.
func TestDrillNormalizeLeavesTheAnswerAlone(t *testing.T) {
	p := drillPlan()
	p.Drill.Answer = 9
	normalizeDrillPlan(p)
	if p.Drill.Answer != 9 {
		t.Fatalf("normalize moved the answer to %d, which would change which plate the clip calls correct", p.Drill.Answer)
	}
}

func TestDrillNormalizeClampsAndCapsTheBoard(t *testing.T) {
	p := drillPlan()
	p.Drill.Question = "why on earth does adding one tenth to two tenths not give you exactly three tenths in a program"
	p.Drill.Options[0] = "the computer rounds the numbers off in a completely random way"
	p.Drill.Options = append(p.Drill.Options, "  ", "a fifth spare plate", "a sixth spare plate")
	p.Drill.Why = "because base two cannot write one tenth exactly and so the value that gets stored is already a little bit wrong"
	normalizeDrillPlan(p)
	if n := len(strings.Fields(p.Drill.Question)); n != maxDrillQuestionWords {
		t.Fatalf("the question survived at %d words", n)
	}
	if n := len(strings.Fields(p.Drill.Options[0])); n != maxDrillOptionWords {
		t.Fatalf("an option survived at %d words", n)
	}
	if n := len(strings.Fields(p.Drill.Why)); n != maxDrillWhyWords {
		t.Fatalf("the why line survived at %d words", n)
	}
	if n := len(p.Drill.Options); n != maxDrillOptions {
		t.Fatalf("want %d options after normalize, got %d", maxDrillOptions, n)
	}
}

func TestDrillNormalizeClampsAnEliminateOffTheBoard(t *testing.T) {
	p := drillPlan()
	p.Beats[1].Drill.At = 99
	p.Beats[4].Drill.At = 3
	normalizeDrillPlan(p)
	if at := p.Beats[1].Drill.At; at != len(p.Drill.Options)-1 {
		t.Fatalf("want the strike clamped to the last plate, got %d", at)
	}
	// A reveal does not index anything, so an index on it is noise.
	if at := p.Beats[4].Drill.At; at != 0 {
		t.Fatalf("a reveal beat kept its index %d", at)
	}
}

func TestDrillShowDefaultsToEliminate(t *testing.T) {
	b := DrillBeat{Show: "wobble"}
	if got := b.ResolvedShow(); got != "eliminate" {
		t.Fatalf("an unknown show resolved to %q, want eliminate", got)
	}
	b = DrillBeat{Show: " ASK "}
	if got := b.ResolvedShow(); got != "ask" {
		t.Fatalf("a shouted ask resolved to %q", got)
	}
}

// The board accumulates: by the closer every wrong plate is struck and the
// answer is lit, which is the frame a viewer screenshots.
func TestDrillScenesAccumulateTheStrikes(t *testing.T) {
	p := drillPlan()
	scenes, err := drillScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if len(steps) != len(p.Beats) {
		t.Fatalf("want %d steps, got %d", len(p.Beats), len(steps))
	}

	first := steps[0]
	struck, _ := first["struck"].([]int)
	if first["show"] != "ask" || len(struck) != 0 {
		t.Fatalf("the opening board is already struck: %v", first)
	}
	if first["revealed"] != false || first["whyOn"] != false {
		t.Fatalf("the ask beat has already resolved the board: %v", first)
	}

	last := steps[len(steps)-1]
	strikes, _ := last["struck"].([]int)
	want := []int{0, 1, 3}
	if last["show"] != "why" || len(strikes) != len(want) {
		t.Fatalf("the closer does not carry every strike: %v", last)
	}
	for i, at := range strikes {
		if at != want[i] {
			t.Fatalf("the strike set is not the wrong plates in order: %v", strikes)
		}
	}
	if last["revealed"] != true || last["whyOn"] != true {
		t.Fatalf("the closer has not revealed the answer and its reason: %v", last)
	}
}
