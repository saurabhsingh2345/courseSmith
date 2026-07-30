package pipeline

import (
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

// The substance stage collected misconceptions and nothing read them: a populated
// field with no consumer, which is worse than an absent one because it looks done.
// The myth prompt meanwhile demanded a claim "in the words somebody actually uses"
// and warned that a strawman corrects nobody — while giving the model no source
// for what anybody believes, so the warning could not bite.
func TestMisconceptionsReachTheMythPrompt(t *testing.T) {
	cfg := config.Defaults()
	cfg.Style.PaceWPM = 175
	held := []string{
		"You need to know how to code before you can build software",
		"No-code tools can build anything",
	}
	spec := SnippetSpec{
		Prompt:    "the belief that building software needs years of experience",
		Template:  "myth",
		Substance: &Substance{Subject: "no-code", Misconceptions: held},
	}

	tpl := SnippetTemplates[spec.Template]
	data := sharedPromptData(spec, cfg)
	for k, v := range tpl.PromptData(spec, cfg) {
		data[k] = v
	}
	system, _, healed, err := renderPromptFileHealed(repoPromptsDir, tpl.PromptFile, data)
	if err != nil {
		t.Fatalf("rendering %s: %v", tpl.PromptFile, err)
	}
	if len(healed) > 0 {
		t.Errorf("the myth prompt references keys nothing supplies: %v", healed)
	}
	for _, m := range held {
		if !strings.Contains(system, m) {
			t.Errorf("misconception %q never reached the prompt", m)
		}
	}
	// And the instruction that makes them load-bearing rather than decorative.
	if !strings.Contains(system, "Take your claim from this list") {
		t.Error("the beliefs are shown but the model is not told to use them")
	}
}

// A piece with no established misconceptions must render exactly as it did before
// the field existed — no empty heading, no instruction pointing at nothing.
func TestMythPromptWithoutMisconceptions(t *testing.T) {
	cfg := config.Defaults()
	spec := SnippetSpec{Prompt: "the belief that indexes always help", Template: "myth"}

	tpl := SnippetTemplates[spec.Template]
	data := sharedPromptData(spec, cfg)
	for k, v := range tpl.PromptData(spec, cfg) {
		data[k] = v
	}
	system, _, healed, err := renderPromptFileHealed(repoPromptsDir, tpl.PromptFile, data)
	if err != nil {
		t.Fatal(err)
	}
	if len(healed) > 0 {
		t.Errorf("unsupplied keys: %v", healed)
	}
	if strings.Contains(system, "BELIEFS THIS AUDIENCE ACTUALLY HOLDS") {
		t.Error("the misconceptions block rendered with nothing in it")
	}
	// The prompt's own strawman warning has to survive — it is the fallback when
	// nothing was established.
	if !strings.Contains(system, "strawman") {
		t.Error("the strawman warning was lost")
	}
}

// Shared as well as myth-owned, so any template can open on a belief the viewer
// recognises, and so a prompt referencing the key does not fall through the
// healing path and render empty with a spurious drift warning.
func TestMisconceptionsAreSharedPromptData(t *testing.T) {
	cfg := config.Defaults()
	spec := SnippetSpec{
		Template:  "verdict",
		Substance: &Substance{Misconceptions: []string{"No-code cannot scale"}},
	}
	got, _ := sharedPromptData(spec, cfg)["Misconceptions"].([]string)
	if len(got) != 1 || got[0] != "No-code cannot scale" {
		t.Errorf("Misconceptions = %v, want the established belief", got)
	}
}

func TestSubstanceMisconceptionsToleratesNil(t *testing.T) {
	if got := substanceMisconceptions(nil); got != nil {
		t.Errorf("substanceMisconceptions(nil) = %v", got)
	}
}

// The field is only worth wiring if the stage actually fills it. This pins the
// contract the prompt asks for: beliefs in the audience's own voice, kept apart
// from the subject facts.
func TestSubstanceKeepsMisconceptionsApartFromFacts(t *testing.T) {
	s := substanceFixture()
	s.Misconceptions = []string{"You need a CS degree to build software"}
	if err := s.Validate(); err != nil {
		t.Fatalf("a sheet with misconceptions was rejected: %v", err)
	}
	// A misconception must not leak into the renderable facts — it is a fact about
	// the audience, and rendering it as a claim about the subject would put a
	// falsehood on screen as though it were true.
	for _, f := range s.Renderable() {
		if strings.Contains(f.Claim, "CS degree") {
			t.Error("a misconception appeared among the renderable facts")
		}
	}
	for _, line := range substanceLines(s) {
		if strings.Contains(line, "CS degree") {
			t.Error("a misconception was handed to a writer as an established fact")
		}
	}
}
