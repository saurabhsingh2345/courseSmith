package pipeline

import (
	"strings"
	"testing"

	"github.com/enfec/coursesmith/internal/config"
)

// The enrich prompt renders in two shapes from one file: a standalone snippet,
// where the creator's prompt is the only source there is, and a combo segment,
// where the piece's brief, the caster's material and the ground already covered
// are all in hand. Every one of those three blocks is conditional, so the
// standalone shape exercises none of them and would keep passing while the
// segment shape was broken.

func enrichData(spec SnippetSpec) map[string]any {
	tpl := SnippetTemplates[spec.Template]
	return map[string]any{
		"Prompt":       spec.Prompt,
		"Template":     tpl.Name,
		"TemplateName": tpl.Title,
		"Description":  tpl.Description,
		"Example":      tpl.Example,
		"Needs":        templateNeeds(tpl.Name),
		"TargetSec":    spec.ResolvedTargetSec(),
		"Audience":     config.Defaults().Style.Audience,
		"Brief":        strings.TrimSpace(spec.Brief),
		"Material":     strings.TrimSpace(spec.Material),
		"Priors":       spec.Priors,
		"Facts":        substanceLines(spec.Substance),
		"Gaps":         substanceGaps(spec.Substance),
	}
}

// Every key the enrich prompt references must be supplied, not healed. Healing
// renders the key empty and warns that the prompt and the binary have drifted,
// so a prompt relying on it looks like a build problem on every run while
// silently dropping the data it asked for. This is the guard that caught exactly
// that after .Facts and .Gaps were added to the prompt and nowhere else.
func TestEnrichPromptNeedsNoHealing(t *testing.T) {
	spec := SnippetSpec{Prompt: "why indexes matter", Template: "gauge"}
	_, _, healed, err := renderPromptFileHealed(repoPromptsDir, snippetEnrichTemplateName, enrichData(spec))
	if err != nil {
		t.Fatalf("rendering %s: %v", snippetEnrichTemplateName, err)
	}
	if len(healed) > 0 {
		t.Errorf("the enrich prompt references keys nothing supplies: %v", healed)
	}
}

func TestEnrichPromptRendersForAStandaloneSnippet(t *testing.T) {
	spec := SnippetSpec{
		Prompt:   "What a 70B model costs to run at home",
		Template: "costing",
	}
	system, user, err := renderPromptFile(repoPromptsDir, snippetEnrichTemplateName, enrichData(spec))
	if err != nil {
		t.Fatalf("rendering %s: %v", snippetEnrichTemplateName, err)
	}
	if !strings.Contains(user, spec.Prompt) {
		t.Error("the request did not reach the user message")
	}
	// None of the combo-only blocks may appear when there is no combo.
	for _, phrase := range []string{"THE WHOLE PIECE", "MATERIAL ALREADY CHOSEN", "ALREADY COVERED"} {
		if strings.Contains(system, phrase) {
			t.Errorf("standalone snippet got the combo-only block %q", phrase)
		}
	}
	if strings.Contains(system, "<no value>") || strings.Contains(user, "<no value>") {
		t.Error("rendered a <no value> placeholder")
	}
}

func TestEnrichPromptCarriesComboContext(t *testing.T) {
	spec := SnippetSpec{
		Prompt:   "how various tools help people create faster",
		Template: "showcase",
		Brief:    "An introduction to no-code. Cover how Webflow, Bubble, Zapier, Supabase and Cursor help people ship without writing code.",
		Material: "Webflow (visual site builder); Bubble (app builder); Zapier (automation); Supabase (database); Cursor (AI editor)",
		Priors: []string{
			"What everyone gets wrong about building software (myth): What everyone says; Not quite",
			"Understanding no-code (constellation): The core concept; Visual programming",
		},
	}
	system, user, err := renderPromptFile(repoPromptsDir, snippetEnrichTemplateName, enrichData(spec))
	if err != nil {
		t.Fatalf("rendering %s: %v", snippetEnrichTemplateName, err)
	}
	if !strings.Contains(user, spec.Prompt) {
		t.Error("the request did not reach the user message")
	}
	// The specifics this whole change exists to deliver. The one-line prompt
	// names no tool; the brief and the material name five between them, and if
	// they do not reach the model the segment is planned exactly as blindly as
	// it was before.
	for _, tool := range []string{"Webflow", "Bubble", "Zapier", "Supabase", "Cursor"} {
		if !strings.Contains(system, tool) {
			t.Errorf("%q reached neither the brief nor the material in the rendered prompt", tool)
		}
	}
	// Every prior must render, not just the first — {{range}} over one element
	// looks identical to a bug that drops the tail.
	for _, prior := range spec.Priors {
		if !strings.Contains(system, prior) {
			t.Errorf("prior %q did not render", prior)
		}
	}
	if !strings.Contains(system, "Do not cover these again") {
		t.Error("the priors block rendered without its instruction")
	}
	if strings.Contains(system, "<no value>") || strings.Contains(user, "<no value>") {
		t.Error("rendered a <no value> placeholder")
	}
}

// The fabrication rules are the load-bearing part of this prompt: they are what
// stands between a template that needs a number and a number that does not
// exist. Asserted by content rather than by rendering alone, because a prompt
// that renders and no longer forbids invention is the exact regression that put
// "Emily built MoodMatch, thousands of downloads" into a finished video.
func TestEnrichPromptForbidsInvention(t *testing.T) {
	spec := SnippetSpec{Prompt: "why indexes matter", Template: "gauge"}
	system, _, err := renderPromptFile(repoPromptsDir, snippetEnrichTemplateName, enrichData(spec))
	if err != nil {
		t.Fatalf("rendering %s: %v", snippetEnrichTemplateName, err)
	}
	if !strings.Contains(system, "DO NOT INVENT") {
		t.Error("the enrich prompt no longer forbids invention")
	}
	for _, subject := range []string{"case study", "statistic"} {
		if !strings.Contains(system, subject) {
			t.Errorf("the invention rules no longer mention %q", subject)
		}
	}
}
