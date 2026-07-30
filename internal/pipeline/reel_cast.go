package pipeline

// Casting: a brief in, an ordered run of templates out.
//
// This is the stage that turns "explain what decides whether a model runs
// locally" into nine segments, each with a template and what it covers. It is
// the only genuinely new thinking in the reel path — planning, assembly and
// rendering were all reuse.
//
// The difficulty is not picking looks. It is that every template has a
// validator that will reject material it cannot express: cast `gauge` on a
// segment with no threshold, or `costing` where nothing adds up, and the plan
// stage burns correction rounds discovering it — after the caster has gone.
//
// So the caster is made to do the awkward part up front, the same move every
// template's own validator makes. Each segment must name the MATERIAL it will
// use — the concrete facts the template needs — and a segment that cannot
// supply any is a segment whose template was wrong. That moves failure from
// expensive and late (mid-pipeline, after nine planning calls) to cheap and
// early (one call, before anything is spent).
//
// The other thing the caster owns is rhythm. Nine `metric` segments in a row is
// unwatchable however good each one is, so repetition is capped and a template
// may not follow itself. Those are enforced rather than requested: a prompt can
// ask for variety and a model will still produce a run of five identical looks
// when the subject leans that way.

import (
	"context"
	"fmt"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
)

const reelCastTemplateName = "reel_cast.tmpl"

const (
	// castRepairRounds is how many times the caster is told what it got wrong.
	// Fewer than planning needs: the rules here are about shape and variety
	// rather than a stack of interacting numeric budgets.
	castRepairRounds = 3

	// maxSameTemplate is how often one template may appear in a reel. Three is
	// generous — a piece that genuinely wants four gauges wants a different
	// structure — and the cap exists because the failure it prevents is a video
	// nobody finishes rather than a video that is wrong.
	maxSameTemplate = 3

	maxCastMaterialWords = 24
)

// CastResult is what the caster returns, before it becomes a ReelSpec.
type CastResult struct {
	Title    string     `json:"title"`
	Segments []CastPick `json:"segments"`
}

// CastPick is one chosen segment.
type CastPick struct {
	// Template is the look this part is cut in.
	Template string `json:"template"`
	// Covers is what this segment should say — it becomes the segment's prompt
	// and is planned through the template's own prompt.
	Covers string `json:"covers"`
	// Material is the concrete facts this template needs and the caster claims
	// to have. Required, and the rule the stage exists for: a segment that
	// cannot name its material is one whose template cannot be filled.
	Material string `json:"material"`
	// Why is the caster's reason for this look. Not used by the pipeline; kept
	// because it is the first thing a person reads when deciding whether to
	// override the choice, and asking for it also visibly improves the picks.
	Why string `json:"why,omitempty"`
}

// SnippetCatalogForPrompt renders the template catalog as the caster sees it:
// grouped, with the copy the studio gallery shows.
//
// Built from the live registry rather than written into the prompt file, so a
// template added to the catalog is castable the moment it registers. A prompt
// with a hand-maintained list would silently stop offering the newest looks,
// which is exactly the failure nobody notices.
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

// CastReel turns a brief into an ordered run of segments.
func CastReel(ctx context.Context, e *Env, brief, title string, want int, cfg config.Config) (*ReelSpec, error) {
	if strings.TrimSpace(brief) == "" {
		return nil, fmt.Errorf("casting needs a brief — say what the whole piece is about")
	}
	if e.Router == nil {
		return nil, fmt.Errorf("casting a reel needs an LLM — set GROQ_API_KEY (or an OpenAI-compatible provider) and retry")
	}
	if want <= 0 {
		want = 5
	}
	want = min(max(want, minReelSegments), maxReelSegments)

	system, user, err := e.renderPrompt(reelCastTemplateName, map[string]any{
		"Brief":        brief,
		"Title":        title,
		"Catalog":      SnippetCatalogForPrompt(),
		"WantSegments": want,
		"MinSegments":  minReelSegments,
		"MaxSegments":  maxReelSegments,
		"MaxSame":      maxSameTemplate,
		"Audience":     cfg.Style.Audience,
		"Tone":         cfg.Style.Tone,
	})
	if err != nil {
		return nil, err
	}

	var cast CastResult
	err = e.completeJSONLenientRounds(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.6, 4096, castRepairRounds, &cast, func() error {
		normalizeCast(&cast)
		return validateCast(&cast)
	})
	if err != nil {
		return nil, err
	}

	spec := &ReelSpec{Title: cast.Title, Brief: brief}
	if strings.TrimSpace(title) != "" {
		spec.Title = strings.TrimSpace(title)
	}
	for _, p := range cast.Segments {
		spec.Segments = append(spec.Segments, ReelSegment{
			Template: p.Template,
			Prompt:   p.Covers,
			// The material is why the template was chosen, and it is what the
			// segment's writer needs to fill it. It used to be validated here
			// and then dropped, so every writer started from the one-line
			// `covers` and invented the rest.
			Material: p.Material,
		})
	}
	spec.EnsureSegmentIDs()
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	return spec, nil
}

// normalizeCast repairs the caster's mechanical mistakes before it is judged:
// stray casing on a template name, whitespace, an over-long material line.
func normalizeCast(c *CastResult) {
	c.Title = collapseSpaces(c.Title)
	out := make([]CastPick, 0, len(c.Segments))
	for _, p := range c.Segments {
		p.Template = strings.ToLower(strings.TrimSpace(p.Template))
		p.Covers = collapseSpaces(p.Covers)
		p.Material = clampWords(collapseSpaces(p.Material), maxCastMaterialWords)
		p.Why = collapseSpaces(p.Why)
		if p.Template != "" && p.Covers != "" {
			out = append(out, p)
		}
	}
	c.Segments = out
}

// validateCast enforces what the caster cannot be trusted to do unprompted.
func validateCast(c *CastResult) error {
	if n := len(c.Segments); n < minReelSegments || n > maxReelSegments {
		return fmt.Errorf("cast %d segments, want %d-%d", n, minReelSegments, maxReelSegments)
	}
	counts := map[string]int{}
	for i, p := range c.Segments {
		if _, ok := SnippetTemplates[p.Template]; !ok {
			return fmt.Errorf("segment %d names template %q, which does not exist. Choose only from the catalog you were given", i+1, p.Template)
		}
		// The rule this stage exists for.
		if strings.TrimSpace(p.Material) == "" {
			return fmt.Errorf("segment %d (%s) names no material. Every template needs particular facts to be filled — a gauge needs a ceiling and things measured against it, a costing needs line items that add up. If you cannot name what this segment will actually contain, the template is the wrong one: pick another",
				i+1, p.Template)
		}
		if i > 0 && c.Segments[i-1].Template == p.Template {
			return fmt.Errorf("segments %d and %d are both %s. Two of the same look running together reads as one long segment; put something else between them or merge them",
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
