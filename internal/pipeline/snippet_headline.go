package pipeline

// The headline's one coloured phrase.
//
// Every headline in the reference look has exactly one span in an accent —
// "A $4000 BOX BUILT FOR AI", "ONE NUMBER MOST SPEC SHEETS BURY", "64
// ACCELERATORS MINIMUM" — and it is the single cheapest thing that makes set
// type read as designed type. No layout, no motion, one colour on a slice of a
// string.
//
// Two decisions carry over from templates that already had to solve this.
//
// The phrase is **quoted, not described**. The model writes the characters it
// means and Go finds them, exactly as an anatomy part is a literal substring of
// its subject. A model asked to name "the third word" gets it wrong often
// enough that the failure — a colour on the wrong word — is worse than no
// colour at all, and it fails silently.
//
// The colour is chosen by **role, not by taste**. The phrase says what it is
// doing in the argument and the design system decides what that looks like, the
// same trade the metric template makes for its figures. A headline that emphasised
// in the brand colour would say nothing; one that emphasised in red because red
// looked good would be lying on a frame where red already means "does not fit".
// See videoskin.go for why the semantic accents are not brand colours.

import (
	"fmt"
	"sort"
	"strings"
)

// emphasisRoles is the closed vocabulary of what an emphasised phrase is doing.
// It is deliberately the metric roles minus "neutral": a phrase singled out of a
// headline is by definition not neutral, and offering the value would invite
// plans that colour a word for decoration.
var emphasisRoles = map[string]bool{
	// The number or thing being measured — the subject's own quantity.
	"quantity": true,
	// The ceiling, the budget, the cost: what the subject runs into.
	"limit": true,
	// The alternative being weighed against the subject.
	"rival": true,
}

// EmphasisRoles returns the role vocabulary sorted, for prompts and docs.
func EmphasisRoles() []string {
	out := make([]string, 0, len(emphasisRoles))
	for k := range emphasisRoles {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// maxEmphasisWords caps the coloured span. Past four words the emphasis is the
// headline and the effect inverts: nothing is louder than anything else, and the
// frame loses the one mark the eye was supposed to land on first.
const maxEmphasisWords = 4

// normalizePlanEmphasis repairs what can be repaired and drops the rest.
//
// Dropping is the right repair here and not a loss: a headline with no coloured
// phrase is the look every existing template already has, so the failure mode is
// "less striking", not "broken". Correcting a phrase that does not occur in the
// title would mean guessing which words the model meant, which is the mistake
// this design exists to rule out.
func normalizePlanEmphasis(p *SnippetPlan) {
	p.Emphasis = clampWords(collapseSpaces(p.Emphasis), maxEmphasisWords)
	p.EmphasisRole = strings.ToLower(strings.TrimSpace(p.EmphasisRole))

	if p.Emphasis == "" || !containsPhrase(p.Title, p.Emphasis) {
		p.Emphasis = ""
		p.EmphasisRole = ""
		return
	}
	// A phrase that is the whole headline is not an emphasis. Clearing it keeps
	// the frame honest rather than painting every word.
	if phraseKey(p.Emphasis) == phraseKey(p.Title) {
		p.Emphasis = ""
		p.EmphasisRole = ""
		return
	}
	if !emphasisRoles[p.EmphasisRole] {
		// The role is a claim about meaning, and an unroled phrase still reads
		// correctly — the renderer paints it in the brand accent, which asserts
		// nothing. Clearing the role is safer than picking one on the model's
		// behalf.
		p.EmphasisRole = ""
	}
}

// validatePlanEmphasis rejects a role the vocabulary does not have.
//
// Only the role is validated, and deliberately so. normalizePlanEmphasis has
// already dropped a phrase that does not occur in the title, so by this point
// the phrase is either a real substring or absent — neither is worth a
// correction round. A misspelled *role*, though, is a model that thought it was
// making a claim and did not, and telling it so is how the next draft is right.
func validatePlanEmphasis(p *SnippetPlan) error {
	role := strings.ToLower(strings.TrimSpace(p.EmphasisRole))
	if role == "" {
		return nil
	}
	if !emphasisRoles[role] {
		return fmt.Errorf("the headline emphasis has role %q, which is not one of: %s. The colour says what the phrase is doing in the argument, so a role outside the vocabulary is a claim the design system cannot draw",
			p.EmphasisRole, strings.Join(EmphasisRoles(), ", "))
	}
	if strings.TrimSpace(p.Emphasis) == "" {
		return fmt.Errorf("the headline has emphasis role %q but no phrase to paint. Quote the words from the title %q that the role describes — the phrase is matched literally, not by position",
			role, p.Title)
	}
	return nil
}

// headlineProps adds the header's emphasis to a scene's props.
//
// A helper rather than two lines copied into each template's scenes function,
// because the prop names have to agree with SceneHeader's and there is no
// compiler between here and there. Omitting the keys entirely when there is no
// emphasis keeps scene graphs recorded before this existed byte-identical.
func headlineProps(p *SnippetPlan, props map[string]any) map[string]any {
	if strings.TrimSpace(p.Emphasis) == "" {
		return props
	}
	props["emphasis"] = p.Emphasis
	if role := strings.ToLower(strings.TrimSpace(p.EmphasisRole)); emphasisRoles[role] {
		props["emphasisRole"] = role
	}
	return props
}
