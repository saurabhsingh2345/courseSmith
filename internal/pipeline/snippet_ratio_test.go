package pipeline

import (
	"strings"
	"testing"
)

const rtNarration = "Eight hundred gigabytes a second against two hundred and seventy is not a small gap."

func ratioPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "ratio",
		Title:    "The spec sheet leaves this out",
		Ratio: &RatioSpec{
			Unit:      "GB/s",
			Reference: RatioSide{Label: "Mac Studio M3 Ultra", Value: 800, Role: "rival"},
			Subject:   RatioSide{Label: "DGX Spark", Value: 270, Role: "limit"},
			Phrase:    "a third",
			Note:      "So every token comes out slower",
		},
		Beats: []SnippetBeat{
			{ID: "mac", Heading: "What the Mac does", Narration: rtNarration, Ratio: &RatioBeat{Show: "reference"}},
			{ID: "spark", Heading: "What the Spark does", Narration: rtNarration, Ratio: &RatioBeat{Show: "subject"}},
			{ID: "third", Heading: "A third", Narration: rtNarration, Ratio: &RatioBeat{Show: "fraction"}},
			{ID: "cost", Heading: "What that costs", Narration: rtNarration, Ratio: &RatioBeat{Show: "read"}},
		},
	}
	p.targetWords = 4 * 40
	return p
}

func TestRatioPlanAccepted(t *testing.T) {
	if err := validateRatioPlan(ratioPlan()); err != nil {
		t.Fatalf("a well-formed ratio plan was rejected: %v", err)
	}
}

// The rule the template exists for. The phrase is the one line the clip is built
// to be remembered by, so it cannot be the false one.
func TestRatioRejectsAPhraseTheNumbersDoNotSupport(t *testing.T) {
	p := ratioPlan()
	p.Ratio.Phrase = "half"
	err := validateRatioPlan(p)
	if err == nil {
		t.Fatal("270 out of 800 called \"half\" was accepted")
	}
	if !strings.Contains(err.Error(), "0.338") && !strings.Contains(err.Error(), "0.337") {
		t.Fatalf("the error does not state the real fraction: %v", err)
	}
}

// Rounding to a friendly fraction is expected and correct: 0.3375 is "a third"
// to every person who has ever said it out loud.
func TestRatioAcceptsAFriendlyRounding(t *testing.T) {
	p := ratioPlan()
	// 0.3375 against an exact third of 0.3333 — inside the tolerance.
	if err := validateRatioPlan(p); err != nil {
		t.Fatalf("a friendly rounding was rejected: %v", err)
	}
}

// A phrase the validator cannot read is one a viewer cannot convert either.
func TestRatioRejectsAPhraseNobodySaysOutLoud(t *testing.T) {
	p := ratioPlan()
	p.Ratio.Phrase = "0.3375 times"
	err := validateRatioPlan(p)
	if err == nil {
		t.Fatal("a phrase that is not a spoken fraction was accepted")
	}
	if !strings.Contains(err.Error(), "a third") {
		t.Fatalf("the error does not offer the vocabulary: %v", err)
	}
}

// Qualifiers do not change the fraction: "roughly a third" is a third.
func TestRatioReadsQualifiedPhrases(t *testing.T) {
	for _, phrase := range []string{"a third", "a third of", "roughly a third", "barely a third of"} {
		p := ratioPlan()
		p.Ratio.Phrase = phrase
		if err := validateRatioPlan(p); err != nil {
			t.Errorf("phrase %q was rejected: %v", phrase, err)
		}
	}
}

// A fraction is only worth stating when it is striking.
func TestRatioRejectsASpreadTooSmallToBeWorthAPhrase(t *testing.T) {
	p := ratioPlan()
	p.Ratio.Subject.Value = 700
	p.Ratio.Phrase = "three quarters"
	err := validateRatioPlan(p)
	if err == nil {
		t.Fatal("a spread of 1.14x was accepted")
	}
	if !strings.Contains(err.Error(), "data template") {
		t.Fatalf("the error does not name the alternative: %v", err)
	}
}

// Which way round the model filled the fields is not a claim about the subject —
// the values are — so it is swapped rather than rejected.
func TestRatioNormalizeSwapsAnInvertedPair(t *testing.T) {
	p := ratioPlan()
	p.Ratio.Reference = RatioSide{Label: "DGX Spark", Value: 270, Role: "limit"}
	p.Ratio.Subject = RatioSide{Label: "Mac Studio M3 Ultra", Value: 800, Role: "rival"}
	normalizeRatioPlan(p)
	if p.Ratio.Reference.Value != 800 {
		t.Fatalf("reference is %v after normalize, want the larger value 800", p.Ratio.Reference.Value)
	}
	if p.Ratio.Subject.Value != 270 {
		t.Fatalf("subject is %v after normalize, want the smaller value 270", p.Ratio.Subject.Value)
	}
	// And the swapped pair then validates against the same phrase.
	if err := validateRatioPlan(p); err != nil {
		t.Fatalf("the swapped pair was rejected: %v", err)
	}
}

func TestRatioRequiresOneSharedUnit(t *testing.T) {
	p := ratioPlan()
	p.Ratio.Unit = ""
	err := validateRatioPlan(p)
	if err == nil {
		t.Fatal("a pair with no unit was accepted")
	}
	if !strings.Contains(err.Error(), "rate") {
		t.Fatalf("the error does not explain why one unit matters: %v", err)
	}
}

// The fraction is OF something, so that something has to exist first.
func TestRatioRequiresTheReferenceFirst(t *testing.T) {
	p := ratioPlan()
	p.Beats[0].Ratio = &RatioBeat{Show: "subject"}
	p.Beats[1].Ratio = &RatioBeat{Show: "reference"}
	if err := validateRatioPlan(p); err == nil {
		t.Fatal("a clip that states the subject before the reference was accepted")
	}
}

// Naming the fraction early throws away the only surprise the clip has.
func TestRatioRejectsTheFractionBeforeTheSubject(t *testing.T) {
	p := ratioPlan()
	p.Beats[1].Ratio = &RatioBeat{Show: "fraction"}
	p.Beats[2].Ratio = &RatioBeat{Show: "subject"}
	if err := validateRatioPlan(p); err == nil {
		t.Fatal("a clip that names the fraction before the measurement was accepted")
	}
}

func TestRatioRequiresAPhrase(t *testing.T) {
	p := ratioPlan()
	p.Ratio.Phrase = ""
	err := validateRatioPlan(p)
	if err == nil {
		t.Fatal("a pair with no phrase was accepted")
	}
	if !strings.Contains(err.Error(), "chart") {
		t.Fatalf("the error does not say what it would be instead: %v", err)
	}
}

// The renderer never divides: the share it draws is the same division the
// validator checked.
func TestRatioScenesCarryTheCheckedShare(t *testing.T) {
	p := ratioPlan()
	scenes, err := ratioScenes(sceneInput(t, p, 16000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	sub, _ := scenes[0].Props["subject"].(map[string]any)
	if got := sub["frac"].(float64); got < 0.337 || got > 0.338 {
		t.Fatalf("the subject's share reached the scene as %v, want 0.3375", got)
	}
	if scenes[0].Props["phrase"] != "a third" {
		t.Fatalf("the phrase did not reach the scene: %v", scenes[0].Props["phrase"])
	}
}
