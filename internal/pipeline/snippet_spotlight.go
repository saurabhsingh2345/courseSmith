package pipeline

// The spotlight template: one named thing, and the short list of what it is for.
//
// This is the asymmetric member of the showroom family. `cards` puts things in a
// row and `duel` puts two face to face; both are frames about a CHOICE. This one
// has nothing to choose between — it is a single product held up on the left of
// the frame with its claims stacked beside it, one landing at a time.
//
// The catalog's nearest neighbour is `showcase`, and the difference is runtime
// rather than subject. A showcase is a seventy-second introduction: four facts, a
// strengths column, an honest limitations column, and a designed hand-off to a
// screen recording. That is the right shape the first time a course meets a tool
// and much too much apparatus for the fourth. This is the fifteen-second version
// — the thing, its mark, and two to four claims — and having both means a course
// can introduce a tool properly once and then mention the next six without every
// mention costing a minute.
//
// The composition is the point, and it is the one thing this family has that the
// rest of the catalog does not: the frame is deliberately UNBALANCED. A card on
// the left, rows on the right, nothing in the lower band. Every centred layout in
// the catalog says "here is a diagram"; this one says "here is a thing, and here
// is what I claim about it", which is a different sentence.
//
// Two rules earn the shape.
//
// A point is a CLAIM, not a feature. "Ten gigabyte context window" is a spec and
// belongs on a showcase's fact row; "reads a whole repository at once" is what the
// spec is FOR, and it is what a viewer can act on. The word ceiling is the
// enforcement — a claim fits in nine words and a feature list does not.
//
// Every point gets its own beat, in order, exactly once. Same rule as the cards
// row, and here it is load-bearing in a second way: the rows land one at a time
// on their beats, so a point with no beat is a row that never appears at all.

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "spotlight",
		Category: CatDecisions,
		Since:    SinceV9,
		Family:   FamilyShowroom,
		Title:    "One thing, and what it is for",
		Description: "A single product held up on the left wearing its real logo, with two to four claims landing one at a time beside it. " +
			"Reach for it to introduce a named tool in fifteen seconds — showcase is the long version.",
		Example:    "Claude Code: what it is actually good at",
		PromptFile: snippetSpotlightTemplateName,
		NeedsCode:  false,
		Plan:       planSpotlightSnippet,
		// The card, two points, the finished frame: four beats. That is genuinely
		// short, and it is the reason this template exists next to showcase.
		MinTargetSec:     20,
		DefaultTargetSec: 35,
		MaxBeats:         7,
		// A beat is a shot. Twenty-two words is about seven seconds, which is what
		// one row landing wants — long enough to say the claim and short enough
		// that four of them do not add up to a minute.
		IdealWordsPerBeat: 22,
		Owns:              beatFields{Spotlight: true},
		OwnsPlan:          planFields{Spotlight: true},
		Normalize:         normalizeSpotlightPlan,
		Validate:          validateSpotlightPlan,
		Scenes:            spotlightScenes,
		PromptData: func(spec SnippetSpec, cfg config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(SpotlightShows(), ", "),
				"Icons":         strings.Join(PointIconNames(), ", "),
				"MinPoints":     minSpotlightPoints,
				"MaxPoints":     spotlightCeilingFor(spec, cfg),
				"MaxNameWords":  maxSpotlightNameWords,
				"MaxNoteWords":  maxSpotlightNoteWords,
				"MaxPointWords": maxSpotlightPointWords,
			}
		},
	})
}

const snippetSpotlightTemplateName = "snippet_spotlight.tmpl"

const (
	// One claim is a tagline with extra steps; past four the rows are taller than
	// the card beside them and the composition stops being a card with a list and
	// becomes a list with a logo on it.
	minSpotlightPoints = 2
	maxSpotlightPoints = 4

	maxSpotlightNameWords = 4
	// The line under the name, inside the card.
	maxSpotlightNoteWords = 12
	// A claim. Nine words is the ceiling and it is doing real work — see the file
	// header on why a feature list must not fit.
	maxSpotlightPointWords = 9
)

// spotlightShows is the closed vocabulary of what a beat does.
var spotlightShows = map[string]bool{
	// The card alone, no rows yet. The opener.
	"card": true,
	// Row At lands, the ones before it holding.
	"point": true,
	// Everything up and even. The closer.
	"all": true,
}

// SpotlightShows returns the beat vocabulary sorted.
func SpotlightShows() []string {
	out := make([]string, 0, len(spotlightShows))
	for k := range spotlightShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// SpotlightSpec is the subject and its claims.
type SpotlightSpec struct {
	// Title is the thing's NAME.
	Title string `json:"title"`
	// Note is one line of what it is, inside the card under the name.
	Note string `json:"note"`
	// Role tints the card: a metricRoles name.
	Role string `json:"role,omitempty"`
	// Points are the claims, top to bottom.
	Points []SpotlightPoint `json:"points"`

	// Where the mark comes from — the same fields a Card carries.
	Brand    string `json:"brand,omitempty"`
	Site     string `json:"site,omitempty"`
	ImageURL string `json:"imageUrl,omitempty"`
	Icon     string `json:"icon,omitempty"`

	// Resolved by the pipeline.
	Mark     string `json:"mark,omitempty"`
	Tint     string `json:"tint,omitempty"`
	Image    string `json:"image,omitempty"`
	MarkFrom string `json:"markFrom,omitempty"`
}

// SpotlightPoint is one claim.
type SpotlightPoint struct {
	// Text is the claim.
	Text string `json:"text"`
	// Icon is a PointIconNames name, drawn in a chip at the head of the row.
	Icon string `json:"icon,omitempty"`
}

// ResolvedIcon returns the row's glyph. Never empty: an empty chip at the head of
// a row is a hole where every other row has a mark.
func (p SpotlightPoint) ResolvedIcon() string {
	if n := normalizePointIconName(p.Icon); n != "" {
		return n
	}
	return "check"
}

// ResolvedRole returns the card's tint, defaulting to neutral.
func (s SpotlightSpec) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(s.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// ResolvedIcon returns the card's fallback glyph. Never empty.
func (s SpotlightSpec) ResolvedIcon() string {
	if n := normalizePointIconName(s.Icon); n != "" {
		return n
	}
	return "box"
}

// asCard views the subject as a Card, so the cards template's art resolver works
// on it unchanged.
func (s *SpotlightSpec) asCard() Card {
	return Card{
		Title: s.Title, Note: s.Note, Role: s.Role,
		Brand: s.Brand, Site: s.Site, ImageURL: s.ImageURL, Icon: s.Icon,
		Mark: s.Mark, Tint: s.Tint, Image: s.Image, MarkFrom: s.MarkFrom,
	}
}

// SpotlightBeat is one shot.
type SpotlightBeat struct {
	// Show is a spotlightShows name.
	Show string `json:"show"`
	// At indexes SpotlightSpec.Points, for a "point" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults the unknown to a point landing, which is what most beats
// of this template are.
func (b SpotlightBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if spotlightShows[s] {
		return s
	}
	return "point"
}

// spotlightPointBudget is how many claims the beat budget funds: the card opens
// and the finished frame closes, so everything between them is one point.
func spotlightPointBudget(targetWords int) int {
	_, maxBeats, _, _ := beatBounds(targetWords, templateBeatCeiling("spotlight"), templateIdealWords("spotlight"))
	return min(max(maxBeats-2, minSpotlightPoints), maxSpotlightPoints)
}

// spotlightCeilingFor is the same arithmetic for a request that has not been
// planned yet, so the prompt quotes the number the validator will score against.
func spotlightCeilingFor(spec SnippetSpec, cfg config.Config) int {
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	want, _, _ := wordBudget(spec.ResolvedTargetSec(), pace)
	return spotlightPointBudget(want)
}

// planSpotlightSnippet plans the clip and then goes and gets the mark.
func planSpotlightSnippet(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config) (*SnippetPlan, error) {
	plan, err := planSnippetDefault(ctx, e, spec, cfg)
	if err != nil {
		return nil, err
	}
	if s := plan.Spotlight; s != nil {
		row := &CardsSpec{Items: []Card{s.asCard()}}
		resolveCardArt(ctx, e, row)
		s.Mark, s.Tint = row.Items[0].Mark, row.Items[0].Tint
		s.Image, s.MarkFrom = row.Items[0].Image, row.Items[0].MarkFrom
	}
	return plan, nil
}

func normalizeSpotlightPlan(p *SnippetPlan) {
	s := p.Spotlight
	if s == nil {
		return
	}
	s.Title = clampWords(collapseSpaces(s.Title), maxSpotlightNameWords)
	s.Note = clampWords(collapseSpaces(s.Note), maxSpotlightNoteWords)
	s.Role = s.ResolvedRole()
	s.Icon = s.ResolvedIcon()
	s.Brand = cardsSlug(s.Brand)
	s.Site = cardsHost(s.Site)

	points := make([]SpotlightPoint, 0, len(s.Points))
	for _, pt := range s.Points {
		pt.Text = clampWords(collapseSpaces(pt.Text), maxSpotlightPointWords)
		pt.Icon = pt.ResolvedIcon()
		if pt.Text != "" && len(points) < maxSpotlightPoints {
			points = append(points, pt)
		}
	}
	s.Points = points

	for i := range p.Beats {
		b := p.Beats[i].Spotlight
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "point" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(s.Points); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateSpotlightPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Spotlight: true}); err != nil {
		return err
	}

	s := p.Spotlight
	if s == nil {
		return fmt.Errorf("the plan has no subject — this template is one named thing and what it is for, so the card is the clip")
	}
	if strings.TrimSpace(s.Title) == "" {
		return fmt.Errorf("the card has no name. The name is what the mark above it identifies")
	}
	if strings.TrimSpace(s.Note) == "" {
		return fmt.Errorf("the card has no line under the name. A logo and a name is a sticker: say in one clause what %q IS, so the claims beside it have something to attach to", s.Title)
	}
	budget := spotlightPointBudget(p.targetWords)
	if n := len(s.Points); n < minSpotlightPoints || n > maxSpotlightPoints {
		return fmt.Errorf("the card has %d claims beside it, want %d-%d. One claim is a tagline with extra steps; past %d the rows are taller than the card and the frame becomes a list with a logo on it",
			n, minSpotlightPoints, maxSpotlightPoints, maxSpotlightPoints)
	}
	if n := len(s.Points); n > budget {
		return fmt.Errorf("the card has %d claims but this runtime funds only %d: the first beat lands the card, the last evens the frame, and every beat between is one claim. Use %d claims, or ask for a longer clip",
			n, budget, budget)
	}
	seen := map[string]bool{}
	for i, pt := range s.Points {
		if strings.TrimSpace(pt.Text) == "" {
			return fmt.Errorf("claim %d is empty", i)
		}
		key := cardsFold(pt.Text)
		if seen[key] {
			return fmt.Errorf("two claims are both %q, so the frame makes the same point twice", pt.Text)
		}
		seen[key] = true
	}

	if p.Beats[0].Spotlight == nil || p.Beats[0].Spotlight.ResolvedShow() != "card" {
		return fmt.Errorf("beat %q does not open on the card alone. The thing has to be on screen and named before anything is claimed about it — open with {\"show\": \"card\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Spotlight == nil || last.Spotlight.ResolvedShow() != "all" {
		return fmt.Errorf("beat %q does not close on the finished frame. Ending with the last row still lit leaves the earlier claims dimmed on the final frame, which says only the last one counted — end with {\"show\": \"all\"}", last.ID)
	}

	next := 0
	for _, b := range p.Beats {
		d := b.Spotlight
		if d == nil {
			return fmt.Errorf("beat %q has no spotlight direction — every beat lands the card, lands one claim, or evens the frame", b.ID)
		}
		if d.ResolvedShow() != "point" {
			continue
		}
		if d.At < 0 || d.At >= len(s.Points) {
			return fmt.Errorf("beat %q lands claim %d, which does not exist — there are claims 0-%d", b.ID, d.At, len(s.Points)-1)
		}
		if next >= len(s.Points) {
			return fmt.Errorf("beat %q lands claim %d (%q) when all %d have already landed. Each claim lands once, top to bottom",
				b.ID, d.At, s.Points[d.At].Text, len(s.Points))
		}
		if d.At != next {
			return fmt.Errorf("beat %q lands claim %d (%q) when claim %d (%q) is next. The rows stack downward in order — landing them out of order makes the stack jump",
				b.ID, d.At, s.Points[d.At].Text, next, s.Points[next].Text)
		}
		next++
	}
	if next != len(s.Points) {
		return fmt.Errorf("the clip lands %d of %d claims, so %q never appears on screen at all — the rows are drawn on their beats, so a claim with no beat is not a dim row, it is a missing one",
			next, len(s.Points), s.Points[next].Text)
	}
	return nil
}

// spotlightScenes lays the clip out as ONE scene: the card holds, and the steps
// say how many rows have landed and which one is newest.
func spotlightScenes(in SnippetSceneInput) ([]Scene, error) {
	s := in.Plan.Spotlight
	if s == nil {
		return nil, fmt.Errorf("the plan has no subject")
	}
	if len(s.Points) == 0 {
		return nil, fmt.Errorf("the card has no claims beside it")
	}

	points := make([]map[string]any, len(s.Points))
	for i, pt := range s.Points {
		points[i] = map[string]any{"text": pt.Text, "icon": pt.ResolvedIcon()}
	}

	card := map[string]any{
		"title": s.Title,
		"note":  s.Note,
		"role":  s.ResolvedRole(),
		"icon":  s.ResolvedIcon(),
	}
	if s.Mark != "" {
		card["mark"] = s.Mark
	}
	if s.Tint != "" {
		card["tint"] = s.Tint
	}
	if s.Image != "" {
		card["image"] = s.Image
	}
	if s.MarkFrom != "" {
		card["markFrom"] = s.MarkFrom
	}

	// `shown` is how many rows are on screen and `at` which one just landed. Both
	// are carried on the step rather than derived in the component, so a frame is
	// a function of one object — the same discipline the cards row keeps.
	shown := 0
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Spotlight == nil {
			return nil, fmt.Errorf("beat %q has no spotlight direction", beat.ID)
		}
		show := beat.Spotlight.ResolvedShow()
		step := map[string]any{"startMs": startMs, "endMs": endMs, "show": show}
		switch show {
		case "point":
			at := beat.Spotlight.At
			if at < 0 || at >= len(s.Points) {
				return nil, fmt.Errorf("beat %q lands claim %d, which does not exist", beat.ID, at)
			}
			shown = at + 1
			step["at"] = at
		case "all":
			shown = len(s.Points)
			step["at"] = -1
		default:
			step["at"] = -1
		}
		step["shown"] = shown
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneSpotlight,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"card":   card,
			"points": points,
			"steps":  steps,
		}),
	}}, nil
}
