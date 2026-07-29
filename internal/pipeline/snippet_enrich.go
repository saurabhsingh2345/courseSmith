package pipeline

// Prompt enrichment: turning what somebody typed into what the template needs.
//
// A template's planner is asked for a lot at once — narration inside a word
// budget, a beat structure, and a visual spec that satisfies that template's
// own rules. It can do all of that from a rich brief. From "what a 70B model
// costs at home" it has to invent every number in the clip before it can even
// start, and the usual failure is not a slightly worse clip: it is a reply that
// does not decode at all, and three correction rounds spent saying so.
//
// The fix is not more correction rounds. It is to stop asking the planner to do
// two jobs in one call. Enrichment does the first job on its own: take the
// prompt, look at what this template actually needs, and write the fuller brief
// a person would have written if they had known. The planner then gets the kind
// of input it is good at.
//
// This runs BEFORE planning rather than as a retry after it. A retry would only
// rescue clips that had already failed, and the thin prompts that fail loudly
// are a fraction of the thin prompts that quietly produce a mediocre clip.

import (
	"context"
	"fmt"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
)

const snippetEnrichTemplateName = "snippet_enrich.tmpl"

// enrichedPrompt is the model's rewrite.
type enrichedPrompt struct {
	// Prompt is the fuller brief, in the same voice as the original.
	Prompt string `json:"prompt"`
	// Notes is what it added and why. Logged, not used — but asking for it
	// visibly improves the rewrite, and it is what somebody reads when the
	// clip covers something they did not ask for.
	Notes string `json:"notes,omitempty"`
}

// EnrichSnippetPrompt rewrites a request into something the template can be
// planned from. It never fails the pipeline: a rewrite that does not come back
// leaves the original prompt in place, because a thin prompt is a worse clip
// and a dead enrichment call would otherwise be no clip at all.
func EnrichSnippetPrompt(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config) string {
	original := strings.TrimSpace(spec.Prompt)
	tpl, ok := SnippetTemplates[spec.Template]
	if !ok || e.Router == nil || original == "" {
		return original
	}

	system, user, err := e.renderPrompt(snippetEnrichTemplateName, map[string]any{
		"Prompt":       original,
		"Template":     tpl.Name,
		"TemplateName": tpl.Title,
		"Description":  tpl.Description,
		"Example":      tpl.Example,
		"Needs":        templateNeeds(tpl.Name),
		"TargetSec":    spec.ResolvedTargetSec(),
		"Audience":     cfg.Style.Audience,
	})
	if err != nil {
		return original
	}

	var out enrichedPrompt
	// One round, no correction loop. If the rewrite does not come back cleanly
	// the original is a perfectly good input — spending three more calls to
	// improve an input is worse than planning from what the user typed.
	err = e.completeJSONLenientRounds(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.5, 1024, 1, &out, func() error {
		if len(strings.Fields(out.Prompt)) < 8 {
			return fmt.Errorf("the rewrite is shorter than the request; expand it into the specifics the template needs")
		}
		return nil
	})
	if err != nil || strings.TrimSpace(out.Prompt) == "" {
		return original
	}
	return strings.TrimSpace(out.Prompt)
}

// templateNeeds is the material each template cannot be filled without.
//
// The same knowledge the caster is given, in the same words, because both are
// answering the same question — what does this look actually require — and two
// descriptions of that would drift. Templates not listed here are ones whose
// prompt carries no hard data requirement beyond a subject.
func templateNeeds(name string) string {
	switch name {
	case "metric":
		return "three to six real figures, each with a unit and what it counts, and at least one that is a ceiling or a limit rather than plain context"
	case "gauge":
		return "one ceiling with a number and a name, and two to five things measured against it, none more than four times the ceiling"
	case "costing":
		return "three to six line items with amounts that actually add up to the total, and at least one cost people do not budget for"
	case "decision":
		return "one question that separates the options, and two to four bands along it, each ending in an instruction"
	case "verdict":
		return "the ruling in one line, two to four conditions it holds under, and at least one where it is wrong"
	case "myth":
		return "the belief in the words somebody really uses, what is true instead as a statement rather than a denial, and evidence for it"
	case "analogy":
		return "a genuinely familiar thing, three or four of its parts mapped one-to-one onto the real subject, and where the picture stops working"
	case "trace":
		return "two or three actors, one shared value with a starting state, and three to six operations that change it in a specific order"
	case "rundown":
		return "a set of three to five things that genuinely belong together, and a line on why each earns its place"
	case "constellation":
		return "one idea and three to six properties, each joined to it by a relation word that completes a sentence"
	case "compare":
		return "two named options and the specific dimensions they differ on"
	case "showcase":
		return "one tool, what it costs, what happens if you leave, and something it is honestly bad at"
	case "timeline":
		return "dated or ordered milestones, and why each one mattered"
	case "quiz":
		return "one question with a real answer and wrong options that are genuinely tempting"
	default:
		return ""
	}
}
