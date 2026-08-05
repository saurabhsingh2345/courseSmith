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
	// Role is this segment's job in the arc: hook, develop, or payoff.
	//
	// Casting used to emit segments that were each individually sensible and
	// collectively shapeless — a run of eight parts in the order the brief
	// happened to list them, with nothing at stake at the start and nothing
	// resolved at the end. The prompt has always carried structural advice ("do
	// not open on a verdict, do not close on a definition") and nothing checked
	// it, so it was read as decoration.
	//
	// Naming the role per segment is what makes the shape checkable, and asking
	// for it makes the caster decide the shape rather than discover it. The
	// vocabulary is three words on purpose: any finer and it becomes a taxonomy
	// nobody applies consistently.
	Role string `json:"role"`
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
	// A length in the brief decides the segment count, because the two are the
	// same decision: five segments cannot carry twenty-five minutes without each
	// running five, which is twice any template's ceiling. Previously the brief's
	// runtime was never read and every segment took its template default, so a
	// 25-minute ask shipped as 9:43 with nothing saying so.
	var budget RuntimeBudget
	if requested, ok := ParseRequestedRuntime(brief); ok {
		budget = BudgetRuntime(requested, want)
		want = budget.Segments
		fmt.Fprintf(e.out(), "    %s\n", budget.Describe())
	}
	want = min(max(want, minReelSegments), maxReelSegments)

	// Establish the facts before choosing the looks — the same inversion the
	// substance stage exists for, applied one step earlier.
	//
	// Casting picks the data-hungry templates, so it is the decision that most
	// needs to know which data exists: `gauge` on a subject with no ceiling in it
	// is the original failure, and the caster could only avoid it by guessing. Now
	// it can see the figures and, more usefully, the GAPS — a part whose numbers
	// were looked for and not found is a part that must not be cast as a chart.
	//
	// This does not double the cost. establishSubstance is a pure function of
	// (brief, cfg), the substance stage will call it again with identical inputs,
	// and the LLM cache is keyed on the whole request — so the stage's call is a
	// cache hit and the facts are paid for once. Best-effort for the same reason
	// the stage is: a failure here should produce a less well-informed cast, never
	// no cast at all.
	var sub *Substance
	if s, err := e.establishSubstance(ctx, brief, cfg); err == nil {
		sub = s
		fmt.Fprintf(e.out(), "    %d facts established, %d gaps\n", len(s.Renderable()), len(s.Gaps))
	} else {
		fmt.Fprintf(e.out(), "    ! casting without a fact sheet: %s\n", errSummary(err))
	}

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
		// Zero when the brief named no length. The prompt branches on it, so a
		// piece with no stated runtime reads exactly as it did before.
		"PerSegmentSec": budget.PerSegmentSec,
		"TotalSec":      budget.Achievable(),
		// The established facts, and the holes. Nil-safe: an ungrounded cast
		// renders without these blocks and behaves as it did before.
		"Facts":          substanceLines(sub),
		"Gaps":           substanceGaps(sub),
		"Misconceptions": substanceMisconceptions(sub),
		"Roles":          strings.Join([]string{RoleHook, RoleDevelop, RolePayoff}, ", "),
	})
	if err != nil {
		return nil, err
	}

	var cast CastResult
	err = e.completeJSONLenientRounds(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.6, thinkingBudget(4096), castRepairRounds, effortInterlocking, &cast, func() error {
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
		seg := ReelSegment{
			Template: p.Template,
			Prompt:   p.Covers,
			Role:     p.Role,
			// The material is why the template was chosen, and it is what the
			// segment's writer needs to fill it. It used to be validated here
			// and then dropped, so every writer started from the one-line
			// `covers` and invented the rest.
			Material: p.Material,
		}
		// Only when the brief asked for a length. Left at zero, a segment takes
		// its template's own default, which is the right answer for a piece whose
		// brief says nothing about how long it should be.
		if budget.PerSegmentSec > 0 {
			seg.TargetSec = segmentTargetFor(p.Template, budget.PerSegmentSec)
		}
		spec.Segments = append(spec.Segments, seg)
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
		p.Role = strings.ToLower(strings.TrimSpace(p.Role))
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
	return validateCastArc(c)
}

// validateCastArc enforces the shape the prompt has always described.
//
// The structural advice was there from the start and nothing checked it, so it
// was read as decoration: casts came back as a run of parts in whatever order the
// brief listed them, each individually sensible and collectively shapeless. A
// viewer given eight equally-weighted segments has been given a table of contents,
// not a piece — nothing is at stake at the start, so there is no reason to reach
// the end, and nothing resolves there if they do.
//
// The checks are the three that carry the shape and nothing more. Over-specifying
// an arc would reject good structures for not matching a template, and the point
// is to rule out shapelessness, not to prescribe one form.
func validateCastArc(c *CastResult) error {
	for i, p := range c.Segments {
		if _, ok := castRoleOrder[p.Role]; !ok {
			return fmt.Errorf("segment %d (%s) has role %q, which is not one of %s, %s, %s. Every segment has a job in the arc: what is at stake, how it works, or what to do about it",
				i+1, p.Template, p.Role, RoleHook, RoleDevelop, RolePayoff)
		}
	}

	if r := c.Segments[0].Role; r != RoleHook {
		return fmt.Errorf("the piece opens on a %s segment. The first part has to put something at stake — a belief to correct, a figure that does not fit, a question the viewer already has — or there is no reason to watch the second",
			r)
	}
	last := c.Segments[len(c.Segments)-1]
	if last.Role != RolePayoff {
		return fmt.Errorf("the piece ends on a %s segment (%s). It has to resolve: a ruling on what to do, or one picture that holds the whole thing together. Ending mid-explanation leaves the viewer with notes rather than an answer",
			last.Role, last.Template)
	}

	// Backwards is the failure that shapelessness actually looks like: hook,
	// develop, hook, develop reads as four unrelated clips rather than one
	// argument, because every return to a hook restarts the piece.
	prev := 0
	for i, p := range c.Segments {
		rank := castRoleOrder[p.Role]
		if rank < prev {
			return fmt.Errorf("segment %d (%s) is a %s after the piece had already moved on. The arc runs hook → develop → payoff and does not go back; a second hook halfway through restarts the video",
				i+1, p.Template, p.Role)
		}
		prev = rank
	}

	developed := 0
	for _, p := range c.Segments {
		if p.Role == RoleDevelop {
			developed++
		}
	}
	if developed == 0 {
		return fmt.Errorf("no segment develops anything — the piece is a hook and a payoff with no middle, which promises something and then rules on it without ever explaining it")
	}
	return nil
}
