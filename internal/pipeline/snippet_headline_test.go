package pipeline

import (
	"strings"
	"testing"
)

func headlinePlan(title, emphasis, role string) *SnippetPlan {
	return &SnippetPlan{
		Template:     "metric",
		Title:        title,
		Emphasis:     emphasis,
		EmphasisRole: role,
	}
}

func TestEmphasisSurvivesWhenItOccursInTheTitle(t *testing.T) {
	p := headlinePlan("64 accelerators minimum", "64", "limit")
	normalizePlanEmphasis(p)
	if p.Emphasis != "64" {
		t.Fatalf("emphasis was dropped, want %q, got %q", "64", p.Emphasis)
	}
	if p.EmphasisRole != "limit" {
		t.Fatalf("role was dropped, want %q, got %q", "limit", p.EmphasisRole)
	}
}

// The rule the design exists for. The model quotes the characters it means, and
// a phrase that is not in the title is a colour landing on the wrong words —
// which fails silently, so it is dropped rather than guessed at.
func TestEmphasisDroppedWhenAbsentFromTitle(t *testing.T) {
	p := headlinePlan("64 accelerators minimum", "a whole rack", "limit")
	normalizePlanEmphasis(p)
	if p.Emphasis != "" || p.EmphasisRole != "" {
		t.Fatalf("an emphasis absent from the title survived: %q / %q", p.Emphasis, p.EmphasisRole)
	}
}

// Matching ignores case and punctuation so the model naming a phrase in lower
// case does not have to reproduce the title's typography to be understood.
func TestEmphasisMatchesLooselyButKeepsItsOwnText(t *testing.T) {
	p := headlinePlan("A $4000 box built for AI", "$4000", "limit")
	normalizePlanEmphasis(p)
	if p.Emphasis != "$4000" {
		t.Fatalf("emphasis with punctuation was dropped: %q", p.Emphasis)
	}
}

// A phrase that is the whole headline is not an emphasis: painting every word
// leaves nothing for the eye to land on first, which inverts the effect.
func TestEmphasisSpanningWholeTitleIsDropped(t *testing.T) {
	p := headlinePlan("Speed, no space", "speed no space", "rival")
	normalizePlanEmphasis(p)
	if p.Emphasis != "" {
		t.Fatalf("an emphasis covering the whole title survived: %q", p.Emphasis)
	}
}

// An unroled phrase still reads correctly — the renderer paints it in the brand
// accent, which asserts nothing — so a bad role clears rather than guesses.
func TestUnknownRoleIsClearedNotGuessed(t *testing.T) {
	p := headlinePlan("64 accelerators minimum", "64", "important")
	normalizePlanEmphasis(p)
	if p.Emphasis != "64" {
		t.Fatalf("the phrase was dropped along with its bad role: %q", p.Emphasis)
	}
	if p.EmphasisRole != "" {
		t.Fatalf("an unknown role survived: %q", p.EmphasisRole)
	}
}

// A misspelled role is a model that thought it was making a claim and did not,
// so validation says so rather than letting the draft ship uncoloured.
func TestValidateRejectsUnknownRole(t *testing.T) {
	err := validatePlanEmphasis(headlinePlan("64 accelerators minimum", "64", "important"))
	if err == nil {
		t.Fatal("a role outside the vocabulary was accepted")
	}
	if !strings.Contains(err.Error(), "quantity") {
		t.Fatalf("the error does not offer the vocabulary: %v", err)
	}
}

func TestValidateRejectsRoleWithNoPhrase(t *testing.T) {
	if err := validatePlanEmphasis(headlinePlan("64 accelerators minimum", "", "limit")); err == nil {
		t.Fatal("a role with nothing to paint was accepted")
	}
}

func TestValidateAcceptsNoEmphasisAtAll(t *testing.T) {
	if err := validatePlanEmphasis(headlinePlan("64 accelerators minimum", "", "")); err != nil {
		t.Fatalf("a plan with no emphasis was rejected: %v", err)
	}
}

// The emphasis is checked against the title the frame will carry, which for a
// plan that never set one is the first beat's heading.
func TestEmphasisIsCheckedAfterTheTitleFallback(t *testing.T) {
	p := &SnippetPlan{
		Template:     "metric",
		Emphasis:     "$4000",
		EmphasisRole: "limit",
		Beats: []SnippetBeat{
			{ID: "one", Heading: "A $4000 box built for AI", Narration: "A four thousand dollar box."},
		},
	}
	normalizeSnippetPlan(p)
	if p.Emphasis != "$4000" {
		t.Fatalf("emphasis was dropped against a fallback title %q: %q", p.Title, p.Emphasis)
	}
}

// headlineProps omits the keys entirely when there is nothing to paint, so a
// scene graph recorded before this existed stays byte-identical.
func TestHeadlinePropsAreOmittedWhenThereIsNoEmphasis(t *testing.T) {
	props := headlineProps(headlinePlan("64 accelerators minimum", "", ""), map[string]any{"title": "x"})
	if _, ok := props["emphasis"]; ok {
		t.Fatal("an empty emphasis wrote a prop")
	}
	if _, ok := props["emphasisRole"]; ok {
		t.Fatal("an empty role wrote a prop")
	}
}

func TestHeadlinePropsCarryPhraseAndRole(t *testing.T) {
	props := headlineProps(headlinePlan("64 accelerators minimum", "64", "limit"), map[string]any{})
	if props["emphasis"] != "64" {
		t.Fatalf("the phrase did not reach the scene: %v", props["emphasis"])
	}
	if props["emphasisRole"] != "limit" {
		t.Fatalf("the role did not reach the scene: %v", props["emphasisRole"])
	}
}
