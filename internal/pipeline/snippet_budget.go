package pipeline

// The budget template: a fixed pot, the claims against it, and what is left.
//
// This is `costing` run backwards, and the direction is the whole difference.
// `costing` starts at nothing and adds line items until a total lands — the
// argument is "it adds up to more than you thought". Here the total is the FIRST
// thing on screen and every beat takes a bite out of it. The argument is "there
// is less room than you thought", and the number that teaches is the remainder,
// which is the one figure nobody in the source material ever computes.
//
// It is also not `gauge`, and the line between them is a count. A gauge measures
// candidates against a ceiling one at a time: each bar is independent and the
// question is asked separately of each. A budget stacks claims that all come out
// of the same pot, so the third claim's fate depends on the first two. That
// dependency is why the picture has to accumulate rather than reset.
//
// Three rules earn it its place, and all three are validators.
//
// The pot is established before anything claims it. You cannot spend something
// the viewer has not been handed.
//
// **There are at least two claims.** This is the rule that keeps the template
// from being a worse `gauge`. One claim against a pot is a single quantity
// measured against a limit, which `gauge` draws better and with a threshold. The
// only reason to accumulate is that claims interact, and two is the minimum
// number that can.
//
// **No single claim may exceed the pot on its own.** If the first bite is bigger
// than the whole budget, everything after it is noise — the remainder was already
// decided and the clip is spending beats on figures that cannot matter. That is
// a gauge with one bar, and the error says so.
//
// The remainder is allowed to go negative, and that is the punchline case rather
// than an error: a budget that busts is exactly what the reference clips keep
// landing on.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "budget",
		Category:    CatNumbers,
		Since:       SinceV4,
		Family:      FamilyReplica,
		Title:       "What is left after everything",
		Description: "A fixed pot with claims taken out of it one at a time, closing on the remainder. Reach for it when the point is how little room is left once the obvious costs are paid.",
		Example:     "What is actually left of 24GB of VRAM once the model is loaded",
		PromptFile:  snippetBudgetTemplateName,
		NeedsCode:   false,
		// The pot, two or three claims, and the remainder. The remainder needs a
		// beat of its own: it is the number the whole clip is for.
		MinTargetSec:     30,
		DefaultTargetSec: 50,
		MaxBeats:         8,
		Owns:             beatFields{Budget: true},
		OwnsPlan:         planFields{Budget: true},
		Normalize:        normalizeBudgetPlan,
		Validate:         validateBudgetPlan,
		Scenes:           budgetScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(BudgetShows(), ", "),
				"MinClaims":     minBudgetClaims,
				"MaxClaims":     maxBudgetClaims,
				"MaxLabelWords": maxBudgetLabelWords,
				"MaxNoteWords":  maxBudgetNoteWords,
				"MaxUnitWords":  maxBudgetUnitWords,
			}
		},
	})
}

const snippetBudgetTemplateName = "snippet_budget.tmpl"

const (
	// Two is the floor for a reason — see the file header. Past four the bar is
	// a stack of slivers and the remainder is a sliver among them.
	minBudgetClaims = 2
	maxBudgetClaims = 4

	maxBudgetLabelWords = 5
	maxBudgetNoteWords  = 16
	maxBudgetUnitWords  = 3
)

// budgetShows is the closed vocabulary of what a beat does.
var budgetShows = map[string]bool{
	// The pot, whole and unclaimed. The first beat, always.
	"pot": true,
	// One claim comes out of it.
	"claim": true,
	// Land on the remainder. The last beat.
	"remainder": true,
}

// BudgetShows returns the beat vocabulary sorted.
func BudgetShows() []string {
	out := make([]string, 0, len(budgetShows))
	for k := range budgetShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// BudgetSpec is the pot and what comes out of it. On the plan because the pot is
// one object every beat is measured against.
type BudgetSpec struct {
	// Pot is how much there is to start with.
	Pot float64 `json:"pot"`
	// Unit is what everything here is counted in — "GB", "$/mo", "ms".
	Unit string `json:"unit"`
	// PotLabel names the pot — "what the card holds", "the monthly budget".
	PotLabel string `json:"potLabel"`
	// RemainderLabel names what is left — "left for everything else". Optional.
	RemainderLabel string `json:"remainderLabel,omitempty"`
	// Claims are the bites, in the order they are taken.
	Claims []BudgetClaim `json:"claims"`
}

// BudgetClaim is one bite out of the pot.
type BudgetClaim struct {
	// Amount is how much this claim takes, in the pot's unit.
	Amount float64 `json:"amount"`
	// Label is what is taking it.
	Label string `json:"label"`
	// Note is what this claim means. One sentence.
	Note string `json:"note,omitempty"`
	// Role is what this claim is doing: a metricRoles name.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the claim's role, defaulting to neutral.
func (c BudgetClaim) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(c.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// Claimed is everything the claims take together.
func (s *BudgetSpec) Claimed() float64 {
	t := 0.0
	for _, c := range s.Claims {
		t += c.Amount
	}
	return t
}

// Remainder is what is left, which may be negative — a budget that busts is the
// punchline rather than a mistake.
func (s *BudgetSpec) Remainder() float64 { return s.Pot - s.Claimed() }

// ResolvedRemainderLabel names the leftover on screen.
func (s *BudgetSpec) ResolvedRemainderLabel() string {
	if l := strings.TrimSpace(s.RemainderLabel); l != "" {
		return l
	}
	if s.Remainder() < 0 {
		return "over budget"
	}
	return "left"
}

// BudgetBeat is one move: which claim this beat takes.
type BudgetBeat struct {
	// Show is a budgetShows name.
	Show string `json:"show"`
	// At indexes BudgetSpec.Claims, for a "claim" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to a claim.
func (b BudgetBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if budgetShows[s] {
		return s
	}
	return "claim"
}

func normalizeBudgetPlan(p *SnippetPlan) {
	b := p.Budget
	if b == nil {
		return
	}
	b.Unit = clampWords(collapseSpaces(b.Unit), maxBudgetUnitWords)
	b.PotLabel = clampWords(collapseSpaces(b.PotLabel), maxBudgetLabelWords)
	b.RemainderLabel = clampWords(collapseSpaces(b.RemainderLabel), maxBudgetLabelWords)

	claims := make([]BudgetClaim, 0, len(b.Claims))
	for _, c := range b.Claims {
		c.Label = clampWords(collapseSpaces(c.Label), maxBudgetLabelWords)
		c.Note = clampWords(collapseSpaces(c.Note), maxBudgetNoteWords)
		c.Role = c.ResolvedRole()
		// A claim of nothing takes no bite, so it is not a claim. Dropping is the
		// repair; inventing an amount would be a claim about the subject.
		if c.Amount > 0 && c.Label != "" && len(claims) < maxBudgetClaims {
			claims = append(claims, c)
		}
	}
	b.Claims = claims

	for i := range p.Beats {
		bb := p.Beats[i].Budget
		if bb == nil {
			continue
		}
		bb.Show = bb.ResolvedShow()
		if bb.Show != "claim" {
			bb.At = 0
			continue
		}
		if bb.At < 0 {
			bb.At = 0
		}
		if n := len(b.Claims); n > 0 && bb.At >= n {
			bb.At = n - 1
		}
	}
}

func validateBudgetPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Budget: true}); err != nil {
		return err
	}

	b := p.Budget
	if b == nil {
		return fmt.Errorf("the plan has no pot — this template is one fixed amount and the claims that come out of it")
	}
	if b.Pot <= 0 {
		return fmt.Errorf("the pot is %v; there has to be something there to spend", b.Pot)
	}
	if strings.TrimSpace(b.Unit) == "" {
		return fmt.Errorf("the pot has no unit — %v of what? Every figure in this clip is in the same unit, and without it the remainder is a bare number", b.Pot)
	}
	if strings.TrimSpace(b.PotLabel) == "" {
		return fmt.Errorf("the pot has no label — say what the %v %s IS. A budget nobody has named is an arbitrary number to subtract from", b.Pot, b.Unit)
	}

	// The rule that keeps this from being a worse gauge.
	if n := len(b.Claims); n < minBudgetClaims || n > maxBudgetClaims {
		return fmt.Errorf("there are %d claims, want %d-%d. One claim against a pot is a single quantity measured against a limit, which the gauge template draws better and with a threshold — the only reason to accumulate is that claims interact, and two is the fewest that can. Past four the bar is a stack of slivers",
			n, minBudgetClaims, maxBudgetClaims)
	}

	seen := map[string]bool{}
	for i, c := range b.Claims {
		if c.Amount <= 0 {
			return fmt.Errorf("claim %d (%q) takes %v — a claim of nothing is not a claim", i, c.Label, c.Amount)
		}
		if strings.TrimSpace(c.Label) == "" {
			return fmt.Errorf("claim %d has no label — say what is taking those %v %s", i, c.Amount, b.Unit)
		}
		key := strings.ToLower(strings.TrimSpace(c.Label))
		if seen[key] {
			return fmt.Errorf("two claims are both %q — each bite out of the pot is a different one", c.Label)
		}
		seen[key] = true
		// If one bite is bigger than the whole pot, everything after it is noise:
		// the remainder was already decided.
		if c.Amount > b.Pot {
			return fmt.Errorf("claim %d (%q) takes %v %s out of a pot of %v, so it busts the budget on its own and every claim after it is spending a beat on a figure that cannot matter. That is one bar against a ceiling — use the gauge template, or pick a pot the claims are actually near",
				i, c.Label, c.Amount, b.Unit, b.Pot)
		}
		if r := strings.ToLower(strings.TrimSpace(c.Role)); r != "" && !metricRoles[r] {
			return fmt.Errorf("claim %d has role %q, which is not one of: %s", i, c.Role, strings.Join(MetricRoles(), ", "))
		}
	}

	// The pot is on screen before anything comes out of it.
	if p.Beats[0].Budget == nil || p.Beats[0].Budget.ResolvedShow() != "pot" {
		return fmt.Errorf("beat %q does not establish the pot. You cannot spend something the viewer has not been handed, and a bar that starts part-eaten has no full state to be read against",
			p.Beats[0].ID)
	}

	taken := map[int]bool{}
	pots := 0
	remainders := 0
	for i, beat := range p.Beats {
		if beat.Budget == nil {
			return fmt.Errorf("beat %q has no budget direction — every beat shows the pot, takes a claim, or lands on the remainder", beat.ID)
		}
		switch beat.Budget.ResolvedShow() {
		case "pot":
			pots++
			if i != 0 {
				return fmt.Errorf("beat %q shows the pot whole again part-way through. It is established once, at the start — a claim does not get given back", beat.ID)
			}
		case "claim":
			at := beat.Budget.At
			if at < 0 || at >= len(b.Claims) {
				return fmt.Errorf("beat %q takes claim %d, which does not exist", beat.ID, at)
			}
			if taken[at] {
				return fmt.Errorf("beat %q takes claim %d again; each bite gets one beat", beat.ID, at)
			}
			taken[at] = true
		case "remainder":
			remainders++
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q lands on the remainder but the clip carries on afterwards. The remainder is the closing frame — it is the number the whole clip is for, and it is the one that gets screenshotted", beat.ID)
			}
		}
	}
	if pots != 1 {
		return fmt.Errorf("there are %d beats establishing the pot, want exactly 1", pots)
	}
	if len(taken) != len(b.Claims) {
		return fmt.Errorf("%d of the %d claims are never taken. A bite the narrator skips is one nobody saw, and the remainder on screen will not match the arithmetic — give it a beat or cut it",
			len(b.Claims)-len(taken), len(b.Claims))
	}
	if remainders != 1 {
		return fmt.Errorf("there are %d beats landing on the remainder, want exactly 1, last. What is left is the number this template exists to state", remainders)
	}
	return nil
}

// budgetScenes lays the clip out as ONE scene. The pot persists and the beats
// only say how much of it has been claimed.
func budgetScenes(in SnippetSceneInput) ([]Scene, error) {
	b := in.Plan.Budget
	if b == nil {
		return nil, fmt.Errorf("the plan has no pot")
	}

	// Each claim's share of the pot, so the renderer never divides. Widths are
	// fractions of the pot rather than of the claimed total: a bar that
	// re-normalised as claims landed would show the first claim shrinking, which
	// is the opposite of what is happening.
	claims := make([]map[string]any, len(b.Claims))
	for i, c := range b.Claims {
		claims[i] = map[string]any{
			"amount": c.Amount,
			"label":  c.Label,
			"note":   c.Note,
			"role":   c.ResolvedRole(),
			"frac":   roundTo(c.Amount/b.Pot, 4),
		}
	}

	spent := 0.0
	landed := map[int]bool{}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Budget == nil {
			return nil, fmt.Errorf("beat %q has no budget direction", beat.ID)
		}
		show := beat.Budget.ResolvedShow()
		if show == "claim" {
			spent += b.Claims[beat.Budget.At].Amount
			landed[beat.Budget.At] = true
		}
		// The claims taken by this beat, as a set rather than a count.
		//
		// A count only works if the beats take the claims in declaration order,
		// and nothing enforces that — the validator requires one beat per claim,
		// not a particular sequence. A clip that took claim 2 before claim 0
		// would have reported three claims landed with one on screen.
		taken := make([]int, 0, len(landed))
		for at := range landed {
			taken = append(taken, at)
		}
		sort.Ints(taken)

		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"taken":   taken,
			"left":    roundTo(b.Pot-spent, 4),
		}
		if show == "claim" {
			step["at"] = beat.Budget.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneBudget,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":          in.Plan.Title,
			"pot":            b.Pot,
			"unit":           b.Unit,
			"potLabel":       b.PotLabel,
			"remainderLabel": b.ResolvedRemainderLabel(),
			"remainder":      roundTo(b.Remainder(), 4),
			"remainderFrac":  roundTo(b.Remainder()/b.Pot, 4),
			"claims":         claims,
			"steps":          steps,
		}),
	}}, nil
}
