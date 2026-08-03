package pipeline

import (
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

// The structural advice was in the prompt from the start and nothing checked it,
// so it was read as decoration: casts came back as a run of parts in whatever
// order the brief listed them — each individually sensible, collectively
// shapeless. A viewer given eight equally-weighted segments has a table of
// contents, not a piece.

func TestCastArcAccepted(t *testing.T) {
	if err := validateCastArc(castFixture()); err != nil {
		t.Fatalf("a well-shaped arc was rejected: %v", err)
	}
}

// Opening on a definition or a ruling gives nobody a reason to watch part two.
func TestCastMustOpenOnAHook(t *testing.T) {
	c := castFixture()
	c.Segments[0].Role = RoleDevelop
	err := validateCastArc(c)
	if err == nil {
		t.Fatal("a piece opening on a develop segment was accepted")
	}
	if !strings.Contains(err.Error(), "at stake") {
		t.Errorf("the error does not say what an opener is for: %v", err)
	}
}

// Ending mid-explanation leaves the viewer with notes rather than an answer.
func TestCastMustCloseOnAPayoff(t *testing.T) {
	c := castFixture()
	c.Segments[len(c.Segments)-1].Role = RoleDevelop
	err := validateCastArc(c)
	if err == nil {
		t.Fatal("a piece ending on a develop segment was accepted")
	}
	if !strings.Contains(err.Error(), "resolve") {
		t.Errorf("the error does not say what a closer is for: %v", err)
	}
}

// The failure shapelessness actually looks like: hook, develop, hook, develop
// reads as four unrelated clips, because every return to a hook restarts the
// piece before the first question was paid off.
func TestCastArcMustNotGoBackwards(t *testing.T) {
	c := castFixture()
	c.Segments[2].Role = RoleHook
	err := validateCastArc(c)
	if err == nil {
		t.Fatal("an arc that returned to a hook halfway through was accepted")
	}
	if !strings.Contains(err.Error(), "restarts the video") {
		t.Errorf("the error does not explain why going back is fatal: %v", err)
	}
}

// A hook and a payoff with no middle promises something and rules on it without
// ever explaining it.
func TestCastArcNeedsAMiddle(t *testing.T) {
	c := &CastResult{Title: "t", Segments: []CastPick{
		{Template: "myth", Role: RoleHook, Covers: "the belief", Material: "belief: x; truth: y"},
		{Template: "verdict", Role: RolePayoff, Covers: "what to do", Material: "holds: a; breaks: b"},
	}}
	err := validateCastArc(c)
	if err == nil {
		t.Fatal("a piece with no develop segment was accepted")
	}
	if !strings.Contains(err.Error(), "no middle") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCastRejectsAnInventedRole(t *testing.T) {
	c := castFixture()
	c.Segments[1].Role = "climax"
	err := validateCastArc(c)
	if err == nil {
		t.Fatal("an invented role was accepted")
	}
	if !strings.Contains(err.Error(), "hook, develop, payoff") {
		t.Errorf("the error does not list the vocabulary: %v", err)
	}
}

// Casing and whitespace are the model's mechanical mistakes, repaired before the
// arc is judged — the same treatment the template name gets.
func TestCastNormalizesTheRole(t *testing.T) {
	c := &CastResult{Segments: []CastPick{
		{Template: "myth", Role: "  HOOK ", Covers: "a", Material: "m"},
	}}
	normalizeCast(c)
	if c.Segments[0].Role != RoleHook {
		t.Errorf("role = %q, want %q", c.Segments[0].Role, RoleHook)
	}
}

// The arc must be checked as part of the ordinary validation, not only when
// called directly — otherwise the repair loop never sees it and a shapeless cast
// reaches the pipeline.
func TestValidateCastEnforcesTheArc(t *testing.T) {
	c := castFixture()
	c.Segments[0].Role = RolePayoff
	if err := validateCast(c); err == nil {
		t.Fatal("validateCast accepted a piece that opens on its payoff")
	}
}

// The caster is given the facts AND the gaps, and the gaps are the load-bearing
// half: a part whose numbers were looked for and not found must not be cast as a
// chart, because the segment would then have to invent them.
func TestCastPromptCarriesFactsAndGaps(t *testing.T) {
	cfg := config.Defaults()
	sub := substanceFixture()
	sub.Gaps = []string{"No current figure for professional no-code adoption"}
	sub.Misconceptions = []string{"You need a CS degree to build software"}

	data := map[string]any{
		"Brief": "a brief", "Title": "", "Catalog": SnippetCatalogForPrompt(),
		"WantSegments": 5, "MinSegments": minReelSegments, "MaxSegments": maxReelSegments,
		"MaxSame": maxSameTemplate, "Audience": cfg.Style.Audience, "Tone": cfg.Style.Tone,
		"PerSegmentSec": 0, "TotalSec": 0,
		"Facts":          substanceLines(sub),
		"Gaps":           substanceGaps(sub),
		"Misconceptions": substanceMisconceptions(sub),
		"Roles":          strings.Join([]string{RoleHook, RoleDevelop, RolePayoff}, ", "),
	}
	system, _, healed, err := renderPromptFileHealed(repoPromptsDir, reelCastTemplateName, data)
	if err != nil {
		t.Fatalf("rendering %s: %v", reelCastTemplateName, err)
	}
	if len(healed) > 0 {
		t.Errorf("the cast prompt references keys nothing supplies: %v", healed)
	}
	if !strings.Contains(system, "https://webflow.com/pricing") {
		t.Error("the caster was not given the facts, so it can only guess which templates are fillable")
	}
	if !strings.Contains(system, "professional no-code adoption") {
		t.Error("the caster was not given the gaps")
	}
	if !strings.Contains(system, "Do NOT cast a chart") {
		t.Error("the gaps are shown without the instruction that makes them useful")
	}
	if !strings.Contains(system, "CS degree") {
		t.Error("the audience's beliefs never reached the caster, which is where the strongest hooks are")
	}
	// The role vocabulary must be named, or the model cannot return a valid one.
	for _, r := range []string{RoleHook, RoleDevelop, RolePayoff} {
		if !strings.Contains(system, r) {
			t.Errorf("the prompt never mentions the %q role it will be validated on", r)
		}
	}
}

// An ungrounded cast renders without those blocks and behaves as it did before.
func TestCastPromptWithoutASheet(t *testing.T) {
	cfg := config.Defaults()
	data := map[string]any{
		"Brief": "a brief", "Title": "", "Catalog": SnippetCatalogForPrompt(),
		"WantSegments": 5, "MinSegments": minReelSegments, "MaxSegments": maxReelSegments,
		"MaxSame": maxSameTemplate, "Audience": cfg.Style.Audience, "Tone": cfg.Style.Tone,
		"PerSegmentSec": 0, "TotalSec": 0,
		"Facts":          substanceLines(nil),
		"Gaps":           substanceGaps(nil),
		"Misconceptions": substanceMisconceptions(nil),
		"Roles":          strings.Join([]string{RoleHook, RoleDevelop, RolePayoff}, ", "),
	}
	system, _, healed, err := renderPromptFileHealed(repoPromptsDir, reelCastTemplateName, data)
	if err != nil {
		t.Fatal(err)
	}
	if len(healed) > 0 {
		t.Errorf("unsupplied keys: %v", healed)
	}
	if strings.Contains(system, "THE FACTS THIS PIECE HAS") {
		t.Error("the facts block rendered with nothing in it")
	}
	if strings.Contains(system, "WHAT WAS LOOKED FOR AND NOT FOUND") {
		t.Error("the gaps block rendered with nothing in it")
	}
	// The arc is not conditional on grounding.
	if !strings.Contains(system, "hook → develop → payoff") {
		t.Error("the arc instruction is missing from an ungrounded cast")
	}
}
