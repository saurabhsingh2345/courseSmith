package pipeline

// Casting: an outline in, a look for each part out.
//
// This stage used to do two jobs — divide the topic AND choose the templates —
// and doing both in one call is why finished pieces contained segments that did
// not belong. The division is now decided first and separately
// (combo_outline.go), and what is left here is the job this stage is actually
// good at: given a part of an argument that already exists, which of the
// available looks can hold it.
//
// The caster can no longer add a part, drop one, or change what one covers. It
// receives the parts and returns one template each. That is a real loss of
// latitude and it is the point — the previous version's latitude was spent
// inventing material to justify a template it liked the sound of, which is the
// decision running backwards.
//
// == What the caster is now told, that it was not ==
//
// Every template's bio: what it must be FILLED with, what it is wrongly reached
// for, and which arc roles it can carry (snippet_bio.go). Before this, the
// prompt described eighty-one looks and named the requirements of four of them
// by hand, so the caster's model of the other seventy-seven was "anything goes".
//
// And the pool: only templates whose house style agrees with the chosen theme
// are offered at all (combo_pool.go). A look from another family is not a bad
// look, it is a good look from a different production, and a piece that mixes
// them changes production partway through.
//
// == Why the checks are mechanical where they can be ==
//
// A prompt can ask for anything and a model will comply most of the time. The
// three rules here are the ones whose violation is invisible until the video is
// rendered, so they are checked rather than requested: a template outside the
// pool, a template that cannot carry the role its part was given, and a
// data-hungry template cast over material containing no data. The last is the
// commonest miscast in the catalog's history and it is catchable by looking for
// a digit, which is not clever and does not need to be.

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
)

const comboCastTemplateName = "combo_cast.tmpl"

const (
	// castRepairRounds is how many times the caster is told what it got wrong.
	// Fewer than planning needs: the rules here are about fit and variety rather
	// than a stack of interacting numeric budgets.
	castRepairRounds = 3

	// maxSameTemplate is how often one template may appear in a combo. Three is
	// generous — a piece that genuinely wants four gauges wants a different
	// structure — and the cap exists because the failure it prevents is a video
	// nobody finishes rather than a video that is wrong.
	maxSameTemplate = 3
)

// LookPick is the caster's choice for one part.
type LookPick struct {
	// Heading echoes the outline part this is for.
	//
	// Echoed rather than positional, because a reply that drops or reorders an
	// entry then silently pairs every later look with the wrong part — a failure
	// that produces a perfectly valid combo.yaml describing a completely
	// different video. Cheap to ask for, and it makes the misalignment an error
	// instead of a mystery.
	Heading string `json:"heading"`
	// Template is the look this part is cut in.
	Template string `json:"template"`
	// Why is the caster's reason. Not used by the pipeline; kept because it is
	// the first thing a person reads when deciding whether to override a choice,
	// and asking for it visibly improves the picks.
	Why string `json:"why,omitempty"`
}

// LookCast is what the caster returns.
type LookCast struct {
	Picks []LookPick `json:"picks"`
}

// The arc a piece is cast into.
const (
	// RoleHook puts something at stake — a belief to correct, a number that does
	// not fit, a question the viewer already has.
	RoleHook = "hook"
	// RoleDevelop is the body: how it works, what it costs, what the options are.
	RoleDevelop = "develop"
	// RolePayoff resolves — a ruling, or one picture to remember it by.
	RolePayoff = "payoff"
)

// castRoleOrder ranks the roles so a sequence can be checked for going backwards.
var castRoleOrder = map[string]int{RoleHook: 0, RoleDevelop: 1, RolePayoff: 2}

// digitRe is the test for whether a part's material contains any actual figure.
var digitRe = regexp.MustCompile(`\d`)

// CastLooks chooses a template for every part of an outline.
func CastLooks(ctx context.Context, e *Env, outline *ComboOutline, skin string, cfg config.Config) ([]LookPick, error) {
	if outline == nil || len(outline.Parts) == 0 {
		return nil, fmt.Errorf("casting needs an outline — the parts are decided before the looks are")
	}
	if e.Router == nil {
		return nil, fmt.Errorf("casting a combo needs an LLM — set GROQ_API_KEY (or an OpenAI-compatible provider) and retry")
	}
	skin = normalizeSkin(skin)
	pool := ComboPool(skin)
	if len(pool) == 0 {
		return nil, fmt.Errorf("the %s theme offers no templates to cast from", skin)
	}

	parts := make([]map[string]any, 0, len(outline.Parts))
	for i, p := range outline.Parts {
		parts = append(parts, map[string]any{
			"N":           i + 1,
			"Heading":     p.Heading,
			"Establishes": p.Establishes,
			"Material":    p.Material,
			"Role":        p.Role,
			// Whether this part has anything a data-hungry look could draw. Stated
			// per part rather than left for the model to notice, because "does this
			// text contain a number" is precisely the judgement it gets wrong — a
			// part about memory *sounds* numeric whether or not any figure is in it.
			"HasFigures": digitRe.MatchString(p.Material),
		})
	}

	system, user, err := e.renderPrompt(comboCastTemplateName, map[string]any{
		"Title":     outline.Title,
		"Angle":     outline.Angle,
		"Parts":     parts,
		"Catalog":   ComboCatalogForPrompt(skin),
		"Skin":      skin,
		"PoolNote":  ComboPoolDescribe(skin),
		"MaxSame":   maxSameTemplate,
		"Audience":  cfg.Style.Audience,
		"Tone":      cfg.Style.Tone,
		"Figures":   strings.Join(ComboFigureTemplates(skin), ", "),
		"HookLooks": strings.Join(ComboRoleTemplates(skin, RoleHook), ", "),
		"PayLooks":  strings.Join(ComboRoleTemplates(skin, RolePayoff), ", "),
	})
	if err != nil {
		return nil, err
	}

	var cast LookCast
	err = e.completeJSONLenientRounds(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.5, thinkingBudget(4096), castRepairRounds, effortInterlocking, &cast, func() error {
		normalizeLookCast(&cast)
		return validateLookCast(&cast, outline, skin)
	})
	if err != nil {
		return nil, fmt.Errorf("casting the looks: %w", err)
	}
	return cast.Picks, nil
}

// normalizeLookCast repairs the caster's mechanical mistakes before it is
// judged: stray casing on a template name, whitespace.
func normalizeLookCast(c *LookCast) {
	out := make([]LookPick, 0, len(c.Picks))
	for _, p := range c.Picks {
		p.Template = strings.ToLower(strings.TrimSpace(p.Template))
		p.Heading = collapseSpaces(p.Heading)
		p.Why = collapseSpaces(p.Why)
		if p.Template != "" {
			out = append(out, p)
		}
	}
	c.Picks = out
}

// validateLookCast enforces the three rules whose violation is invisible until
// the video renders, plus the rhythm rules.
func validateLookCast(c *LookCast, outline *ComboOutline, skin string) error {
	if len(c.Picks) != len(outline.Parts) {
		return fmt.Errorf("you returned %d looks for %d parts. Every part gets exactly one look, in order — do not merge, drop or add parts, that decision is already made",
			len(c.Picks), len(outline.Parts))
	}

	counts := map[string]int{}
	for i, p := range c.Picks {
		part := outline.Parts[i]

		// Misalignment first: every check below is against the wrong part if the
		// reply slipped, and the resulting errors would send the caster chasing
		// problems that are not there.
		if normalizeClaim(p.Heading) != normalizeClaim(part.Heading) {
			return fmt.Errorf("look %d says it is for %q but part %d is %q. Return one look per part, in the same order you were given them",
				i+1, p.Heading, i+1, part.Heading)
		}

		tpl, known := SnippetTemplates[p.Template]
		if !known {
			return fmt.Errorf("part %d (%s) names template %q, which does not exist. Choose only from the catalog you were given",
				i+1, part.Heading, p.Template)
		}
		if !InComboPool(skin, p.Template) {
			return fmt.Errorf("part %d (%s) names %q, which is not offered in the %s theme. The whole piece is cut in one house style, and a look from another one reads as a different video however good it is — pick from the catalog you were given",
				i+1, part.Heading, p.Template, skin)
		}
		if !tpl.Bio.CanCarry(part.Role) {
			return fmt.Errorf("part %d (%s) is a %s, and %q cannot carry that job — it can be %s. A %s look that opens the piece puts nothing at stake; pick one of: %s",
				i+1, part.Heading, part.Role, p.Template,
				strings.Join(tpl.Bio.Roles, " or "), p.Template,
				strings.Join(ComboRoleTemplates(skin, part.Role), ", "))
		}
		// The commonest miscast in this catalog's history, caught by looking for a
		// digit. Not clever, and it does not need to be: `gauge` over "how AI lets
		// people build faster" fails here in one cheap check instead of surviving
		// the cast, burning a planning call and three correction rounds, and
		// shipping as a salvaged clip with an empty chart in it.
		if tpl.Bio.Figures && !digitRe.MatchString(part.Material) {
			return fmt.Errorf("part %d (%s) is cast as %q, which cannot be planned without real figures — and this part's material contains none: %q. %s. Pick a look that carries a subject rather than data",
				i+1, part.Heading, p.Template, part.Material, tpl.Bio.Avoid)
		}

		if i > 0 && c.Picks[i-1].Template == p.Template {
			return fmt.Errorf("parts %d and %d are both %s. Two of the same look running together reads as one long segment; the cut between them is the whole reason this is a combo rather than a snippet",
				i, i+1, p.Template)
		}
		counts[p.Template]++
		if counts[p.Template] > maxSameTemplate {
			return fmt.Errorf("%s is used %d times, at most %d. However good each one is, a piece that keeps returning to the same picture is one nobody finishes",
				p.Template, counts[p.Template], maxSameTemplate)
		}
	}
	return nil
}

// SnippetCatalogForPrompt renders the whole catalog as the enrichment and
// no-code paths see it: grouped, with the copy the studio gallery shows.
//
// Kept alongside ComboCatalogForPrompt rather than replaced by it, because the
// two answer different questions. This one is the whole registry with gallery
// copy, for callers that are not casting a themed piece. The director's version
// is one theme's pool with bios, for the caller that is.
func SnippetCatalogForPrompt() string {
	var sb strings.Builder
	for _, g := range SnippetTemplatesByCategory() {
		fmt.Fprintf(&sb, "\n%s — %s\n", g.Title, g.Blurb)
		for _, t := range g.Templates {
			fmt.Fprintf(&sb, "  %-14s %s. %s\n", t.Name, t.Title, t.Description)
			fmt.Fprintf(&sb, "  %-14s e.g. %q\n", "", t.Example)
		}
	}
	return sb.String()
}
