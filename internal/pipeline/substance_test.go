package pipeline

import (
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

func substanceFixture() *Substance {
	return &Substance{
		Subject: "No-code tools for beginners",
		Facts: []Fact{
			{Claim: "Webflow's Basic site plan is $14/month billed annually", Provenance: ProvSourced, Source: "https://webflow.com/pricing"},
			{Claim: "The brief covers no-code tools for complete beginners", Provenance: ProvGiven},
			{Claim: "A 70B model at 16-bit needs about 140GB", Provenance: ProvDerived, Working: "70e9 x 2 bytes = 140GB"},
			{Claim: "Most users ship their first project within a week", Provenance: ProvUnverified},
		},
		Sources:  []llm.Citation{{URL: "https://webflow.com/pricing", Title: "Pricing"}},
		Grounded: true,
	}
}

func TestSubstanceAccepted(t *testing.T) {
	if err := substanceFixture().Validate(); err != nil {
		t.Fatalf("a well-formed fact sheet was rejected: %v", err)
	}
}

// Only three of the four provenances may reach the screen, and `unverified`
// being excluded is the entire mechanism.
func TestOnlyBackedFactsAreRenderable(t *testing.T) {
	got := substanceFixture().Renderable()
	if len(got) != 3 {
		t.Fatalf("got %d renderable facts, want 3: %+v", len(got), got)
	}
	for _, f := range got {
		if f.Provenance == ProvUnverified {
			t.Errorf("an unverified fact is renderable: %q", f.Claim)
		}
	}
}

// A label without its backing is the failure this stage exists to prevent, so it
// is rejected rather than quietly downgraded — downgrading teaches the model that
// claiming a source is free.
func TestSourcedFactNeedsAURL(t *testing.T) {
	s := substanceFixture()
	s.Facts[0].Source = ""
	err := s.Validate()
	if err == nil {
		t.Fatal("a sourced fact with no URL was accepted")
	}
	if !strings.Contains(err.Error(), "mark it unverified") {
		t.Errorf("the error does not tell the model its way out: %v", err)
	}
}

func TestDerivedFactNeedsWorking(t *testing.T) {
	s := substanceFixture()
	s.Facts[2].Working = ""
	if err := s.Validate(); err == nil {
		t.Fatal("a derived fact with no working was accepted")
	}
}

func TestSubstanceRejectsUnknownProvenance(t *testing.T) {
	s := substanceFixture()
	s.Facts[1].Provenance = "probably"
	err := s.Validate()
	if err == nil {
		t.Fatal("an invented provenance label was accepted")
	}
	if !strings.Contains(err.Error(), "given, sourced, derived, captured, unverified") {
		t.Errorf("the error does not list the valid labels: %v", err)
	}
}

// A sheet where nothing may be rendered cannot produce a video, and failing here
// is much cheaper than failing per-segment five planning calls later.
func TestSubstanceRejectsAnEntirelyUnverifiedSheet(t *testing.T) {
	s := &Substance{Subject: "x", Facts: []Fact{
		{Claim: "a", Provenance: ProvUnverified},
		{Claim: "b", Provenance: ProvUnverified},
	}}
	if err := s.Validate(); err == nil {
		t.Fatal("a sheet with nothing renderable was accepted")
	}
}

// Normalizing must fail toward "cannot be rendered". A label the model invented
// becoming `given` would be the worst possible direction.
func TestNormalizeFailsTowardUnrenderable(t *testing.T) {
	s := &Substance{
		Subject: "  spaced   out  ",
		Facts: []Fact{
			{Claim: "a", Provenance: "SOURCED", Source: " https://example.com "},
			{Claim: "b", Provenance: "vibes"},
			// A "source" that is prose rather than a URL is worse than an empty
			// one: it looks like a citation and survives a glance.
			{Claim: "c", Provenance: ProvSourced, Source: "according to the docs"},
		},
	}
	normalizeSubstance(s)

	if s.Subject != "spaced out" {
		t.Errorf("subject = %q", s.Subject)
	}
	if s.Facts[0].Provenance != ProvSourced || s.Facts[0].Source != "https://example.com" {
		t.Errorf("casing and whitespace were not repaired: %+v", s.Facts[0])
	}
	if s.Facts[1].Provenance != "vibes" {
		t.Errorf("an unknown label should reach Validate to be reported, got %q", s.Facts[1].Provenance)
	}
	if s.Facts[2].Source != "" {
		t.Errorf("prose was kept as a source: %q", s.Facts[2].Source)
	}
	// And that fact must now fail, rather than passing with a plausible citation.
	if err := s.Validate(); err == nil {
		t.Error("a fact citing prose instead of a URL survived validation")
	}
}

// The check the whole provenance scheme rests on. A model that wants a fact
// rendered can write any plausible URL; the provider's citation list is the one
// part the model did not author, so it is the only thing worth checking against.
func TestUncitedSourcesAreDowngraded(t *testing.T) {
	s := substanceFixture()
	s.Facts = append(s.Facts, Fact{
		Claim:      "Bubble has 3 million users",
		Provenance: ProvSourced,
		Source:     "https://bubble.io/made-up-page",
	})
	dropUncitedSources(s, s.Sources, true)

	// The genuinely cited one survives.
	if s.Facts[0].Provenance != ProvSourced {
		t.Errorf("a genuinely cited fact was downgraded: %+v", s.Facts[0])
	}
	// The invented URL does not.
	last := s.Facts[len(s.Facts)-1]
	if last.Provenance != ProvUnverified {
		t.Errorf("a fact citing a URL the provider never returned stayed sourced: %+v", last)
	}
	if last.Source != "" {
		t.Errorf("the invented URL was kept: %q", last.Source)
	}
}

// With grounding off nothing was cited, so every `sourced` label is wrong by
// construction and none may survive.
func TestUngroundedRunHasNoSourcedFacts(t *testing.T) {
	s := substanceFixture()
	dropUncitedSources(s, nil, false)
	for _, f := range s.Facts {
		if f.Provenance == ProvSourced {
			t.Errorf("an ungrounded sheet kept a sourced fact: %+v", f)
		}
	}
	// And it still has something to say, from the brief.
	if len(s.Renderable()) == 0 {
		t.Error("downgrading left nothing renderable; given and derived facts should be untouched")
	}
}

// Only backed facts are shown to a writer. The label means something to this
// code and nothing to a model reading a list, so an unverified fact put in front
// of one is a fact that gets used.
func TestSubstanceLinesShowOnlyBackedFactsWithTheirBacking(t *testing.T) {
	lines := substanceLines(substanceFixture())
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3: %v", len(lines), lines)
	}
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "within a week") {
		t.Error("an unverified fact was shown to the writer")
	}
	if !strings.Contains(joined, "https://webflow.com/pricing") {
		t.Error("a sourced fact lost its URL, so nothing downstream can check it")
	}
	if !strings.Contains(joined, "70e9 x 2 bytes") {
		t.Error("a derived fact lost its working")
	}
}

func TestSubstanceHelpersTolerateNil(t *testing.T) {
	if got := substanceLines(nil); got != nil {
		t.Errorf("substanceLines(nil) = %v", got)
	}
	if got := substanceGaps(nil); got != nil {
		t.Errorf("substanceGaps(nil) = %v", got)
	}
}

func TestSearchEnabledRespectsTheOffSentinel(t *testing.T) {
	cfg := config.Defaults().Pipeline
	if !llm.SearchEnabled(cfg) {
		t.Error("grounding is off by default; the substance stage would never search")
	}
	cfg.LLMSearch = llm.SearchDisabled
	if llm.SearchEnabled(cfg) {
		t.Error(`llm_search: "off" did not disable grounding`)
	}
	// The sentinel exists because config layers merge non-empty-wins, so "" can
	// never override a default back off. Guard that reasoning.
	merged := config.Resolve(config.Defaults(), config.Config{}, config.Config{})
	if merged.Pipeline.LLMSearch == "" {
		t.Error("an empty override cleared the search model, so the sentinel is unnecessary — simplify")
	}
}

func TestSubstanceStagePrecedesPlan(t *testing.T) {
	order := project.SnippetStageOrder
	sub, plan := -1, -1
	for i, s := range order {
		switch s {
		case "substance":
			sub = i
		case "plan":
			plan = i
		}
	}
	if sub < 0 {
		t.Fatal("the substance stage is not in the snippet pipeline")
	}
	if sub > plan {
		t.Errorf("substance runs at %d, after plan at %d — the facts must exist before a template is chosen", sub, plan)
	}
}
