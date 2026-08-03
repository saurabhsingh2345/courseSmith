package pipeline

import (
	"strings"
	"testing"
)

func castFixture() *CastResult {
	return &CastResult{
		Title: "What decides whether a model runs",
		Segments: []CastPick{
			{Template: "myth", Role: RoleHook, Covers: "the belief that a bigger card is enough", Material: "belief: buy the biggest GPU; truth: bandwidth sets speed"},
			{Template: "rundown", Role: RoleDevelop, Covers: "the three numbers that decide it", Material: "capacity GB; bandwidth TB/s; compute TFLOPs"},
			{Template: "gauge", Role: RoleDevelop, Covers: "which models fit in 24GB", Material: "24GB ceiling; 7B 14GB, 13B 26GB, 13B-4bit 8GB"},
			{Template: "verdict", Role: RolePayoff, Covers: "what to actually buy", Material: "holds: under 24GB; breaks: compliance, past a few TB"},
		},
	}
}

func TestCastAccepted(t *testing.T) {
	if err := validateCast(castFixture()); err != nil {
		t.Fatalf("a well-formed cast was rejected: %v", err)
	}
}

// The rule this stage exists for: a template chosen with nothing to put in it
// fails later, expensively, after the caster has gone.
func TestCastRequiresMaterial(t *testing.T) {
	c := castFixture()
	c.Segments[2].Material = ""
	err := validateCast(c)
	if err == nil {
		t.Fatal("a segment with no material was accepted")
	}
	if !strings.Contains(err.Error(), "the wrong one") {
		t.Errorf("the error does not tell the model what to do: %v", err)
	}
}

// Rhythm is enforced rather than requested: a prompt can ask for variety and a
// model will still produce a run of identical looks when the subject leans that
// way.
func TestCastRejectsBackToBackRepeats(t *testing.T) {
	c := castFixture()
	c.Segments[2].Template = "rundown"
	err := validateCast(c)
	if err == nil {
		t.Fatal("two identical templates in a row were accepted")
	}
	if !strings.Contains(err.Error(), "one long segment") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCastCapsRepetition(t *testing.T) {
	c := castFixture()
	// Four gauges, alternated so nothing is back to back.
	c.Segments = []CastPick{
		{Template: "gauge", Covers: "a", Material: "m"},
		{Template: "metric", Covers: "b", Material: "m"},
		{Template: "gauge", Covers: "c", Material: "m"},
		{Template: "metric", Covers: "d", Material: "m"},
		{Template: "gauge", Covers: "e", Material: "m"},
		{Template: "metric", Covers: "f", Material: "m"},
		{Template: "gauge", Covers: "g", Material: "m"},
	}
	err := validateCast(c)
	if err == nil {
		t.Fatal("a template used four times was accepted")
	}
	if !strings.Contains(err.Error(), "nobody finishes") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCastRejectsInventedTemplate(t *testing.T) {
	c := castFixture()
	c.Segments[1].Template = "explainer"
	err := validateCast(c)
	if err == nil {
		t.Fatal("an invented template was accepted")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCastNormalizeRepairs(t *testing.T) {
	c := castFixture()
	c.Segments[0].Template = "  MYTH  "
	c.Segments[1].Material = strings.Repeat("word ", 60)
	c.Segments = append(c.Segments, CastPick{Template: "metric", Covers: ""})
	normalizeCast(c)

	if c.Segments[0].Template != "myth" {
		t.Errorf("a mis-cased template stayed %q", c.Segments[0].Template)
	}
	if n := len(strings.Fields(c.Segments[1].Material)); n > maxCastMaterialWords {
		t.Errorf("material is %d words after normalize, want <= %d", n, maxCastMaterialWords)
	}
	for _, p := range c.Segments {
		if p.Covers == "" {
			t.Error("a segment with nothing to cover survived normalize")
		}
	}
}

// The catalog handed to the caster is built from the live registry, so a
// template added today is castable today. A hand-maintained list in the prompt
// would silently stop offering the newest looks.
func TestCatalogForPromptCoversTheWholeRegistry(t *testing.T) {
	cat := SnippetCatalogForPrompt()
	for _, name := range SnippetTemplateNames() {
		if !strings.Contains(cat, name) {
			t.Errorf("template %q is missing from the caster's catalog", name)
		}
	}
	// And the group headings are there, since the caster picks by job.
	for _, g := range SnippetTemplatesByCategory() {
		if !strings.Contains(cat, g.Title) {
			t.Errorf("category %q is missing from the caster's catalog", g.Title)
		}
	}
}

// A cast has to survive becoming a reel: the same rules the spec enforces.
func TestCastProducesAValidSpec(t *testing.T) {
	c := castFixture()
	spec := &ReelSpec{Title: c.Title, Brief: "a brief"}
	for _, p := range c.Segments {
		spec.Segments = append(spec.Segments, ReelSegment{Template: p.Template, Prompt: p.Covers})
	}
	spec.EnsureSegmentIDs()
	if err := spec.Validate(); err != nil {
		t.Fatalf("a cast reel does not validate as a spec: %v", err)
	}
	seen := map[string]bool{}
	for _, s := range spec.Segments {
		if s.ID == "" || seen[s.ID] {
			t.Errorf("segment id %q is empty or duplicated", s.ID)
		}
		seen[s.ID] = true
	}
}
