package pipeline

// Outlining: deciding what the piece ARGUES, before anything decides what it
// looks like.
//
// This stage did not exist. Casting did both jobs at once — divide the topic and
// pick the looks — and that is why the finished pieces had segments that did not
// belong. A model asked for both in one breath does not weigh them equally: the
// catalog is concrete and eighty-one items long, the argument is abstract and
// has to be invented, so the looks win. The division of the topic came out as a
// by-product of which templates sounded appealing, which produces exactly the
// symptom reported — every segment individually plausible, the run of them
// shapeless, and one or two that are there because a template wanted filling
// rather than because the piece needed the part.
//
// Splitting it fixes that by removing the choice. The outline call cannot see
// the catalog at all. It has the subject, the runtime, the established facts and
// nothing else, so the only thing it can optimise is whether the parts add up to
// an argument. The caster then gets a fixed set of parts and may only choose how
// each is shown — it cannot add a part, drop one, or change what one covers.
//
// That constraint is the whole design. Two calls that can each override the
// other are one call with extra steps.
//
// == What a part is ==
//
// Not a topic. A part is a piece of the argument, and it declares what the
// viewer knows after it that they did not know before — which is the test for
// whether it earns its runtime. "Types of no-code tools" is a topic and could be
// two minutes or twenty. "After this they can tell a builder from an automator
// and know which one their problem needs" is a part, and it is obvious when it
// has been covered.

import (
	"context"
	"fmt"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
)

const comboOutlineTemplateName = "combo_outline.tmpl"

// outlineRepairRounds is how many times the outliner is told what it got wrong.
// The rules here are about shape and coverage rather than interlocking numeric
// budgets, so it needs fewer rounds than planning does.
const outlineRepairRounds = 3

// Bounds on what a part may claim. Both exist because the failure they catch is
// the same one: a part written as a topic rather than as an increment.
const (
	// maxPartWords caps `establishes`. A part that needs forty words to say what
	// it establishes has not decided what it establishes.
	maxPartWords = 30
	// maxOutlineMaterialWords caps the facts named per part. The caster reads
	// these to choose a template and the writer reads them to fill it; past this
	// it is a draft of the segment rather than the material for one.
	maxOutlineMaterialWords = 40
)

// ComboOutline is the argument the piece makes, before any look is chosen.
type ComboOutline struct {
	// Title is the finished piece's title.
	Title string `json:"title"`
	// Angle is the single claim the whole piece advances, in one sentence.
	//
	// Its job is to be the thing every part is checked against. A piece with no
	// angle is a list of true statements about a subject, which is the shape a
	// topic naturally falls into and the shape nobody watches to the end. It is
	// also what the critic scores each finished segment against: "does this
	// advance the angle" is answerable, where "is this good" is not.
	Angle string `json:"angle"`
	// Parts are the pieces of the argument, in the order they are made.
	Parts []ComboPart `json:"parts"`
}

// ComboPart is one piece of the argument.
type ComboPart struct {
	// Heading names the part in a few words.
	Heading string `json:"heading"`
	// Establishes is what the viewer knows after this part that they did not
	// know before — written as the increment, not as the topic.
	Establishes string `json:"establishes"`
	// Material is the concrete facts this part is built from.
	//
	// Named here rather than at cast time, and that is a change of who answers
	// for it. The caster used to invent the material to justify the template it
	// had already chosen, which is backwards: the facts are a property of the
	// subject and the template is a choice about how to show them. Now the
	// material exists first and the template has to be one that can hold it.
	Material string `json:"material"`
	// Role is this part's job in the arc: hook, develop or payoff.
	Role string `json:"role"`
}

// OutlineCombo turns a subject and a runtime into the argument the piece makes.
func OutlineCombo(ctx context.Context, e *Env, subject, title string, budget RuntimeBudget, sub *Substance, cfg config.Config) (*ComboOutline, error) {
	if strings.TrimSpace(subject) == "" {
		return nil, fmt.Errorf("outlining needs a subject — say what the piece is about")
	}
	if e.Router == nil {
		return nil, fmt.Errorf("directing a combo needs an LLM — set GROQ_API_KEY (or an OpenAI-compatible provider) and retry")
	}
	want := budget.Segments
	if want <= 0 {
		want = defaultComboSegments
	}
	want = min(max(want, minComboSegments), maxComboSegments)

	system, user, err := e.renderPrompt(comboOutlineTemplateName, map[string]any{
		"Subject":        subject,
		"Title":          title,
		"WantParts":      want,
		"MinParts":       minComboSegments,
		"MaxParts":       maxComboSegments,
		"PerPartSec":     budget.PerSegmentSec,
		"TotalSec":       budget.Achievable(),
		"Audience":       cfg.Style.Audience,
		"Tone":           cfg.Style.Tone,
		"MaxPartWords":   maxPartWords,
		"MaxMaterial":    maxOutlineMaterialWords,
		"Roles":          strings.Join([]string{RoleHook, RoleDevelop, RolePayoff}, ", "),
		"Facts":          substanceLines(sub),
		"Gaps":           substanceGaps(sub),
		"Misconceptions": substanceMisconceptions(sub),
	})
	if err != nil {
		return nil, err
	}

	var outline ComboOutline
	err = e.completeJSONLenientRounds(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.6, thinkingBudget(6144), outlineRepairRounds, effortInterlocking, &outline, func() error {
		normalizeOutline(&outline)
		return validateOutline(&outline, want)
	})
	if err != nil {
		return nil, fmt.Errorf("outlining the piece: %w", err)
	}
	if t := strings.TrimSpace(title); t != "" {
		outline.Title = t
	}
	return &outline, nil
}

// normalizeOutline repairs the outliner's mechanical mistakes before it is
// judged: whitespace, casing on a role, an over-long claim.
func normalizeOutline(o *ComboOutline) {
	o.Title = collapseSpaces(o.Title)
	o.Angle = collapseSpaces(o.Angle)
	out := make([]ComboPart, 0, len(o.Parts))
	for _, p := range o.Parts {
		p.Heading = collapseSpaces(p.Heading)
		p.Establishes = clampWords(collapseSpaces(p.Establishes), maxPartWords)
		p.Material = clampWords(collapseSpaces(p.Material), maxOutlineMaterialWords)
		p.Role = strings.ToLower(strings.TrimSpace(p.Role))
		if p.Heading != "" && p.Establishes != "" {
			out = append(out, p)
		}
	}
	o.Parts = out
}

// validateOutline enforces what the outliner cannot be trusted to do unprompted.
//
// The arc checks are the same three the cast has always applied, moved one stage
// earlier — which is where they belonged. Checking the arc after the looks are
// chosen means a shapeless piece is repaired by relabelling segments, because by
// then the roles are the only thing left that can move. Checking it here means a
// shapeless piece is repaired by rewriting the argument, which is the actual
// defect.
func validateOutline(o *ComboOutline, want int) error {
	if strings.TrimSpace(o.Angle) == "" {
		return fmt.Errorf("the outline names no angle. Say in one sentence what this piece ARGUES — not what it is about. Without it the parts are a list of true statements and nothing decides which of them belong")
	}
	if n := len(o.Parts); n < minComboSegments || n > maxComboSegments {
		return fmt.Errorf("the outline has %d parts, want %d-%d (aim for %d)", n, minComboSegments, maxComboSegments, want)
	}

	seen := map[string]int{}
	for i, p := range o.Parts {
		if _, ok := castRoleOrder[p.Role]; !ok {
			return fmt.Errorf("part %d (%s) has role %q, which is not one of %s, %s, %s",
				i+1, p.Heading, p.Role, RoleHook, RoleDevelop, RolePayoff)
		}
		if strings.TrimSpace(p.Material) == "" {
			return fmt.Errorf("part %d (%s) names no material. Every part is built from something concrete — figures, names, a belief people hold, the steps of a process. If you cannot name what this part contains, it is not a part of the argument and should be merged into one that is",
				i+1, p.Heading)
		}
		// Two parts that establish the same thing is the defect the whole stage
		// exists to catch, and it is the one a reader recognises instantly in a
		// finished video as "that bit was already covered". Compared on the claim
		// rather than the heading: two parts can be headed differently and land on
		// the same increment, which is how a piece repeats itself while looking
		// varied on paper.
		key := normalizeClaim(p.Establishes)
		if prior, dup := seen[key]; dup {
			return fmt.Errorf("parts %d (%s) and %d (%s) establish the same thing: %q. Merge them, or give one of them ground the other does not cover",
				prior+1, o.Parts[prior].Heading, i+1, p.Heading, p.Establishes)
		}
		seen[key] = i
	}
	return validateOutlineArc(o)
}

// validateOutlineArc enforces the shape: open on something at stake, close on
// something resolved, and never go backwards.
func validateOutlineArc(o *ComboOutline) error {
	if r := o.Parts[0].Role; r != RoleHook {
		return fmt.Errorf("the piece opens on a %s part. The first one has to put something at stake — a belief to correct, a figure that does not fit, a question the viewer already has — or there is no reason to watch the second", r)
	}
	if last := o.Parts[len(o.Parts)-1]; last.Role != RolePayoff {
		return fmt.Errorf("the piece ends on a %s part (%s). It has to resolve: a ruling on what to do, or one picture that holds the whole thing together", last.Role, last.Heading)
	}
	prev := 0
	for i, p := range o.Parts {
		rank := castRoleOrder[p.Role]
		if rank < prev {
			return fmt.Errorf("part %d (%s) is a %s after the piece had already moved on. The arc runs hook → develop → payoff and does not go back; a second hook halfway through restarts the video",
				i+1, p.Heading, p.Role)
		}
		prev = rank
	}
	developed := 0
	for _, p := range o.Parts {
		if p.Role == RoleDevelop {
			developed++
		}
	}
	if developed == 0 {
		return fmt.Errorf("no part develops anything — the piece is a hook and a payoff with no middle, which promises something and then rules on it without ever explaining it")
	}
	return nil
}

// normalizeClaim reduces a claim to something two phrasings of the same idea
// compare equal on: lowercased, stripped of punctuation and of the filler words
// that carry no content.
//
// Deliberately crude. It is a duplicate detector, not a semantic one, and its
// job is to catch the case that actually happens — the model restating a part in
// slightly different words — rather than to be right about paraphrase in
// general. A missed duplicate costs one repetitive segment that the critic gets
// a second shot at; a false positive would reject a good outline, so the bar is
// set to require near-identical wording.
func normalizeClaim(s string) string {
	filler := map[string]bool{
		"the": true, "a": true, "an": true, "of": true, "to": true, "and": true,
		"that": true, "is": true, "are": true, "it": true, "they": true, "you": true,
		"can": true, "will": true, "be": true, "how": true, "what": true, "why": true,
		"this": true, "their": true, "your": true, "in": true, "for": true, "on": true,
	}
	var keep []string
	for _, w := range strings.Fields(strings.ToLower(s)) {
		w = strings.Trim(w, ".,;:!?\"'()—–-")
		if w == "" || filler[w] {
			continue
		}
		keep = append(keep, w)
	}
	return strings.Join(keep, " ")
}

// Describe renders the outline for a progress log or a CLI, which is the point
// at which a person can still cheaply disagree with it.
func (o *ComboOutline) Describe() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s\n", o.Title)
	fmt.Fprintf(&sb, "  angle: %s\n", o.Angle)
	for i, p := range o.Parts {
		fmt.Fprintf(&sb, "  %d. [%s] %s — %s\n", i+1, p.Role, p.Heading, p.Establishes)
	}
	return sb.String()
}
