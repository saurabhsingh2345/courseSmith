package pipeline

// The cards template: a row of named things, each wearing its own mark.
//
// Every other template in this catalog draws a picture the pipeline can make
// out of type and geometry — a bar, a ring, a boundary, a table. This one draws
// a picture that only works if the thing on screen is RECOGNISABLE, and that is
// a different problem. "Gemini against Claude" is not two words in a box; it is
// two products a viewer already has opinions about, and a frame that sets their
// names in the house font has thrown away the fastest identification the eye can
// make. So this is the one template whose art comes from outside the process:
// the mark is fetched, and snippet_cards_art.go is where that happens.
//
// The relation between the cards is declared rather than inferred, because the
// same row of three cards means three different things depending on what sits in
// the gaps. `versus` is a contest — these are alternatives and you pick one.
// `then` is a sequence — these happen in order and the arrow is the claim.
// `none` is a set — these are the players, and the row is orientation rather
// than argument. Drawing all three the same way would be drawing none of them.
//
// Three rules earn the shape.
//
// Every card carries a NOTE. A card that is a logo and a name is a sticker: the
// viewer recognises it and learns nothing, and a row of five of them is a
// sponsor slide. What makes this teach is the line under the name, so it is
// required rather than optional.
//
// Every card gets its own beat, in order, exactly once. The row is on screen
// from the first frame — that is the format's whole advantage, the viewer can
// see how many things there are before a word is read — which means a card the
// voice never accounts for is a card sitting lit on the stage saying nothing.
//
// A `versus` row must CLOSE on something. This is borrowed from the versus
// template's hardest-won rule and it is the same failure: a comparison that ends
// having named the contenders has spent forty seconds building a scoreboard and
// then read out the score. The closer says what to do about it, and a closer
// that reduces to one card's name is a preference rather than advice.

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "cards",
		Category: CatDecisions,
		Since:    SinceV8,
		// Moved off `replica` when the showroom skin arrived, and it was a
		// correction rather than a reshuffle: this template was always cut for a
		// light stage and was in the broadcast batch because that batch was the
		// only non-default one that existed. A brand mark on near-black has to be
		// recoloured to be seen, and recolouring it throws away the recognition
		// that is the entire reason to fetch it.
		Family: FamilyShowroom,
		Title:  "Things side by side",
		Description: "Two to five cards in a row, each with the real thing's logo, its name and a line saying what it is — with vs, an arrow, or nothing at all in the gaps between them. " +
			"Reach for it when the subject is named products, tools or services a viewer should be able to tell apart at a glance.",
		Example:    "Claude against Gemini: which one to reach for",
		PromptFile: snippetCardsTemplateName,
		NeedsCode:  false,
		// The custom planner is the fetch: the model names the brand, the
		// pipeline goes and gets its mark. See snippet_cards_art.go.
		Plan: planCardsSnippet,
		// The opener, a beat per card, and the closer. Two cards is the floor and
		// four beats funds it; five cards needs seven, which needs a minute.
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		// Seven: the row, at most five cards, the closer. 2 + maxCards.
		//
		// It was unset, which means the shared ceiling — and the shared ceiling is
		// higher than this template's shape, so a long enough cards clip was told
		// to write more beats than the row could ever fund. Every template whose
		// beat count follows from its content has to say so here.
		// 168s: 7 beats x 60 words a beat, at 2.5 words a second. Past this the
		// shape cannot hold the narration — see MaxTargetSec.
		MaxTargetSec: 168,
		MaxBeats:     7,
		// A beat here is a SHOT — one card lit with its note under it — not a step
		// in an argument. Forty words is thirteen seconds on one static card, which
		// is how a row of five becomes a slideshow. Twenty-six cuts the same
		// narration into more beats, which is what this template needs to reach
		// its own card ceiling at a sane runtime.
		IdealWordsPerBeat: 26,
		Owns:              beatFields{Cards: true},
		OwnsPlan:          planFields{Cards: true},
		Normalize:         normalizeCardsPlan,
		Validate:          validateCardsPlan,
		Scenes:            cardsScenes,
		PromptData: func(spec SnippetSpec, cfg config.Config) map[string]any {
			return map[string]any{
				"Roles":          strings.Join(MetricRoles(), ", "),
				"Relations":      strings.Join(CardRelations(), ", "),
				"Shows":          strings.Join(CardShows(), ", "),
				"Icons":          strings.Join(PointIconNames(), ", "),
				"MinCards":       minCards,
				"MaxCards":       cardCeilingFor(spec, cfg),
				"MaxTitleWords":  maxCardTitleWords,
				"MaxNoteWords":   maxCardNoteWords,
				"MaxCloserWords": maxCardsCloserWords,
				"MaxAskWords":    maxCardsAskWords,
			}
		},
	})
}

const snippetCardsTemplateName = "snippet_cards.tmpl"

const (
	// One card is a title slide and this template is about the gaps between
	// them. Past five the marks drop below the size that makes a logo do its
	// job, which is being recognised rather than read.
	minCards = 2
	maxCards = 5

	// A card's title is a NAME set in display type — "Claude", "Postgres",
	// "GitHub Actions".
	maxCardTitleWords = 3
	// The line under the name. One sentence, and it is the whole reason the card
	// is not a sticker.
	maxCardNoteWords = 18
	// The closer runs across the foot of the frame under the finished row.
	maxCardsCloserWords = 14
	// The row's shared question, set small-caps on every card. Three words is
	// already a long label at that size; the answer is the note, not this.
	maxCardsAskWords = 3
)

// cardRelations is the closed vocabulary of what sits in the gaps.
var cardRelations = map[string]bool{
	// A contest: "vs" between the cards, and the clip has to say which to pick.
	"versus": true,
	// A sequence: an arrow between the cards, left to right.
	"then": true,
	// A set: nothing in the gaps. The row is orientation, not argument.
	"none": true,
}

// CardRelations returns the relation vocabulary sorted.
func CardRelations() []string {
	out := make([]string, 0, len(cardRelations))
	for k := range cardRelations {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// cardShows is the closed vocabulary of what a beat does.
var cardShows = map[string]bool{
	// The whole row up, every card even. The opener.
	"row": true,
	// Card At lit, the rest receding, its note under the row.
	"card": true,
	// Every card lit and the closer landing. The closer.
	"all": true,
}

// CardShows returns the beat vocabulary sorted.
func CardShows() []string {
	out := make([]string, 0, len(cardShows))
	for k := range cardShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// CardsSpec is the row. On the plan rather than on a beat because every card is
// on screen for the whole clip — the beats only move the light around them,
// which is the format's whole advantage over a list revealed one item at a time.
type CardsSpec struct {
	// Relation is what sits in the gaps: a cardRelations name.
	Relation string `json:"relation,omitempty"`
	// Items are the cards, left to right.
	Items []Card `json:"items"`
	// Closer is the line under the finished row. Required for a versus row —
	// see validateCardsPlan.
	Closer string `json:"closer,omitempty"`

	// Ask turns the row into a question the clip answers one card at a time.
	//
	// It is a short small-caps label — "strongest at", "costs", "best for" — set
	// on EVERY card, above a slot where that card's note will land. Until the
	// card has its beat the slot reads `? ? ?`, and the note arrives when the
	// voice gets there.
	//
	// One field on the spec rather than a per-card label, because the row only
	// works if all the cards are answering the SAME question. Three cards each
	// labelled with a different heading is three unrelated cards that happen to
	// be adjacent, which is the failure the relation vocabulary already guards
	// against in the gaps and this guards against inside the cards.
	//
	// Optional. Empty means the notes simply sit under the names from the first
	// frame, which is the right shape for a row that is orientation rather than a
	// question.
	Ask string `json:"ask,omitempty"`
}

// ResolvedRelation returns what goes in the gaps, defaulting the unknown to
// nothing at all. A row with an invented connector drawn between its cards makes
// a claim the narration never made; a plain row makes none.
func (s *CardsSpec) ResolvedRelation() string {
	r := strings.ToLower(strings.TrimSpace(s.Relation))
	if cardRelations[r] {
		return r
	}
	return "none"
}

// Card is one thing in the row.
type Card struct {
	// Title is the thing's NAME — "Claude", "Gemini", "Postgres".
	Title string `json:"title"`
	// Note is what it is, in one line. Required: a card with a logo and a name
	// is a sticker.
	Note string `json:"note"`
	// Role tints the card: a metricRoles name.
	Role string `json:"role,omitempty"`

	// == Where the mark comes from. The model fills these in; the pipeline
	// resolves them into Mark/Image before the plan is written. ==

	// Brand is a Simple Icons slug — "claude", "googlegemini", "postgresql".
	// The first thing tried, and the one that looks right: a real brand mark,
	// monochrome, painted in the card's own colour.
	Brand string `json:"brand,omitempty"`
	// Site is the thing's domain — "anthropic.com". Tried when there is no
	// brand mark to be had, which is most of the time for anything that is not
	// a household name.
	Site string `json:"site,omitempty"`
	// ImageURL is an exact picture to use instead of a logo — a screenshot, a
	// photo, a diagram already published somewhere. Takes precedence over
	// everything else, because a model that named an exact image meant it.
	ImageURL string `json:"imageUrl,omitempty"`
	// Icon is a PointIconNames name, and it is the floor rather than a choice:
	// it is what the card wears when the network is down, the slug was wrong, or
	// the subject never had a logo in the first place. Always set.
	Icon string `json:"icon,omitempty"`

	// == Resolved by the pipeline, not written by the model. ==

	// Mark is SVG path data on a 0 0 24 24 viewBox. This is the good case: a
	// vector mark that stays sharp at any size.
	Mark string `json:"mark,omitempty"`
	// Tint is the brand's own hex, taken off the fetched mark. The colour the
	// mark is painted in, and the tile behind it tinted from — because the whole
	// reason to fetch a logo is that the viewer knows it on sight, and its colour
	// is half of how. Empty when the source served no plain hex fill, in which
	// case the card falls back to its role colour.
	Tint string `json:"tint,omitempty"`
	// Image is a data: URI for a mark that has to keep its own colours — a
	// favicon, a photo. Drawn as-is in the card's tile.
	Image string `json:"image,omitempty"`
	// MarkFrom records where the art came from ("simpleicons:claude"), so a
	// wrong logo on a finished frame can be traced to the thing that served it
	// rather than guessed at.
	MarkFrom string `json:"markFrom,omitempty"`
}

// ResolvedRole returns the card's tint, defaulting to neutral.
func (c Card) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(c.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// ResolvedImageURL returns the exact picture this card asked for, and "" for
// anything that is not a plain https URL.
//
// The scheme check is the point rather than tidiness. Everything else the
// pipeline fetches is a host it chose; this is the one field where a document a
// model wrote decides what gets requested, so it is held to the one shape a
// published image comes in.
func (c Card) ResolvedImageURL() string {
	raw := strings.TrimSpace(c.ImageURL)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return ""
	}
	return u.String()
}

// ResolvedIcon returns the card's fallback glyph. Never empty: this is what the
// card wears when no mark could be fetched, and a card with no art at all is a
// hole in the row rather than a plain card.
func (c Card) ResolvedIcon() string {
	if n := normalizePointIconName(c.Icon); n != "" {
		return n
	}
	return "box"
}

// CardsBeat is one shot of the row.
type CardsBeat struct {
	// Show is a cardShows name.
	Show string `json:"show"`
	// At indexes CardsSpec.Items, for a "card" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults the unknown to a card landing, which is what most beats
// of this template are.
func (b CardsBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if cardShows[s] {
		return s
	}
	return "card"
}

// cardsFold reduces a name to its letters and digits, so "Claude." and "claude"
// are the same answer. Used to catch two cards that are the same thing, and a
// closer that is nothing but one card's name.
func cardsFold(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// cardBeatBudget is how many cards the beat budget can actually fund: the row
// needs an opening beat and a closing one, and everything between them is a
// card.
//
// It is arithmetic rather than a constant because the answer changes with the
// runtime, and a template that tells the model "up to five" at a length that
// funds three is the contradiction documented at length in beatBounds — the
// model obeys the concrete number, fails the beat count, and burns every
// correction round discovering that the instructions could not both be met.
func cardBeatBudget(targetWords int) int {
	_, maxBeats, _, _ := beatBounds(targetWords, templateBeatCeiling("cards"), templateIdealWords("cards"))
	return min(max(maxBeats-2, minCards), maxCards)
}

// cardCeilingFor is cardBeatBudget for a request that has not been planned yet,
// so the prompt quotes the same ceiling the validator will score against.
func cardCeilingFor(spec SnippetSpec, cfg config.Config) int {
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	want, _, _ := wordBudget(spec.ResolvedTargetSec(), pace)
	return cardBeatBudget(want)
}

func normalizeCardsPlan(p *SnippetPlan) {
	c := p.Cards
	if c == nil {
		return
	}
	c.Relation = c.ResolvedRelation()
	c.Closer = clampWords(collapseSpaces(c.Closer), maxCardsCloserWords)
	c.Ask = clampWords(collapseSpaces(c.Ask), maxCardsAskWords)

	items := make([]Card, 0, len(c.Items))
	for _, it := range c.Items {
		it.Title = clampWords(collapseSpaces(it.Title), maxCardTitleWords)
		it.Note = clampWords(collapseSpaces(it.Note), maxCardNoteWords)
		it.Role = it.ResolvedRole()
		it.Icon = it.ResolvedIcon()
		it.Brand = cardsSlug(it.Brand)
		it.Site = cardsHost(it.Site)
		if it.Title != "" && len(items) < maxCards {
			items = append(items, it)
		}
	}
	c.Items = items

	for i := range p.Beats {
		b := p.Beats[i].Cards
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "card" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(c.Items); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateCardsPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Cards: true}); err != nil {
		return err
	}

	c := p.Cards
	if c == nil {
		return fmt.Errorf("the plan has no row — this template is two to five named things side by side, so the cards are the clip")
	}
	budget := cardBeatBudget(p.targetWords)
	if n := len(c.Items); n < minCards || n > maxCards {
		return fmt.Errorf("the row has %d cards, want %d-%d. One card is a title slide and this template is about the gaps between them; past %d the marks drop below the size that lets a logo be recognised rather than read",
			n, minCards, maxCards, maxCards)
	}
	// The arithmetic, quoted back. A row of five at a runtime that funds three
	// beats of cards cannot be covered, and saying "too many cards" without the
	// sum leaves the model to guess between cutting cards and cutting beats.
	if n := len(c.Items); n > budget {
		return fmt.Errorf("the row has %d cards but this runtime funds only %d: the first beat raises the row, the last lands all of them, and every beat between is one card. Use %d cards, or ask for a longer clip",
			n, budget, budget)
	}

	seen := map[string]bool{}
	for i, it := range c.Items {
		if strings.TrimSpace(it.Title) == "" {
			return fmt.Errorf("card %d has no name. The name is what the mark above it identifies — a card with art and no title is a logo nobody named", i)
		}
		key := cardsFold(it.Title)
		if seen[key] {
			return fmt.Errorf("two cards are both %q, so the row shows one thing twice. Name different things, or drop the duplicate", it.Title)
		}
		seen[key] = true
		if strings.TrimSpace(it.Note) == "" {
			return fmt.Errorf("card %d (%q) has no line under it. A card that is a logo and a name is a sticker: the viewer recognises it and learns nothing, and a row of those is a sponsor slide. Say what it IS", i, it.Title)
		}
		if r := strings.ToLower(strings.TrimSpace(it.Role)); r != "" && !metricRoles[r] {
			return fmt.Errorf("card %d has role %q, which is not one of: %s", i, it.Role, strings.Join(MetricRoles(), ", "))
		}
	}

	// A contest has to end on advice. Same rule as the versus template's
	// verdict, and it is here for the same reason: naming two contenders and
	// stopping is a scoreboard rather than a recommendation.
	if c.ResolvedRelation() == "versus" {
		closer := strings.TrimSpace(c.Closer)
		if closer == "" {
			return fmt.Errorf("the cards are set against each other with `vs` in the gaps and the clip closes on nothing. A contest that ends having named the contenders has built a scoreboard and not read out the score — write a closer saying WHEN to reach for which")
		}
		for _, it := range c.Items {
			if cardsFold(closer) == cardsFold(it.Title) {
				return fmt.Errorf("the closer is just %q, which is the name of one card rather than guidance. Say when to reach for it and when one of the others is the right call, in a line somebody could act on", c.Closer)
			}
		}
	}

	if p.Beats[0].Cards == nil || p.Beats[0].Cards.ResolvedShow() != "row" {
		return fmt.Errorf("beat %q does not open on the whole row. Every card being on screen from the first frame is what this format has over a list — the viewer sees how many things there are before a word is read — so open with {\"show\": \"row\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Cards == nil || last.Cards.ResolvedShow() != "all" {
		return fmt.Errorf("beat %q does not close on the finished row. Ending lit on one card leaves the others dimmed on the last frame, which says the clip was about that one — end with {\"show\": \"all\"}", last.ID)
	}

	next := 0
	for _, b := range p.Beats {
		d := b.Cards
		if d == nil {
			return fmt.Errorf("beat %q has no cards direction — every beat raises the row, lights one card, or lands them all", b.ID)
		}
		if d.ResolvedShow() != "card" {
			continue
		}
		if d.At < 0 || d.At >= len(c.Items) {
			return fmt.Errorf("beat %q lights card %d, which does not exist — the row has cards 0-%d", b.ID, d.At, len(c.Items)-1)
		}
		// Every card has already had its beat and here is another one. Checked
		// before the ordering rule below rather than folded into it, because
		// there is no "next card due" to name at this point — `next` has walked
		// off the end of the row, and the ordering message would read it.
		if next >= len(c.Items) {
			return fmt.Errorf("beat %q lights card %d (%q) when all %d cards have already had their beat. Each card is lit once, left to right — a second pass over the row says the first one did not land. Drop this beat, or give the row another card for it",
				b.ID, d.At, c.Items[d.At].Title, len(c.Items))
		}
		if d.At != next {
			return fmt.Errorf("beat %q lights card %d (%q) when card %d (%q) is the next one due. The cards are read left to right, once each — a row that jumps back is a row the eye has lost its place in",
				b.ID, d.At, c.Items[d.At].Title, next, c.Items[next].Title)
		}
		next++
	}
	if next != len(c.Items) {
		return fmt.Errorf("the clip lights %d of %d cards, so %q sits on the stage with nothing said about it. Every card gets its own beat, or drop it from the row",
			next, len(c.Items), c.Items[next].Title)
	}
	return nil
}

// cardsScenes lays the clip out as ONE scene. The row persists and the steps
// only say which card is lit, so the component paints a whole frame from one
// step rather than replaying the clip to work out what should be bright.
func cardsScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.Cards
	if c == nil {
		return nil, fmt.Errorf("the plan has no row")
	}
	if len(c.Items) == 0 {
		return nil, fmt.Errorf("the row has no cards")
	}

	all := make([]int, len(c.Items))
	items := make([]map[string]any, len(c.Items))
	for i, it := range c.Items {
		all[i] = i
		items[i] = map[string]any{
			"title": it.Title,
			"note":  it.Note,
			"role":  it.ResolvedRole(),
			"icon":  it.ResolvedIcon(),
		}
		// Omitted when empty for the same reason the art keys are: the component
		// reads a present tint as "this mark has a real brand colour, paint it in
		// that", and an empty string would paint the mark in nothing.
		if it.Tint != "" {
			items[i]["tint"] = it.Tint
		}
		// Omitted rather than sent empty: the component treats a present key as
		// art it must draw, and an empty string would draw a blank tile where
		// the fallback glyph belongs.
		if it.Mark != "" {
			items[i]["mark"] = it.Mark
		}
		if it.Image != "" {
			items[i]["image"] = it.Image
		}
		if it.MarkFrom != "" {
			items[i]["markFrom"] = it.MarkFrom
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Cards == nil {
			return nil, fmt.Errorf("beat %q has no cards direction", beat.ID)
		}
		show := beat.Cards.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "card":
			at := beat.Cards.At
			if at < 0 || at >= len(c.Items) {
				return nil, fmt.Errorf("beat %q lights card %d, which does not exist", beat.ID, at)
			}
			step["at"] = at
			step["lit"] = []int{at}
		default:
			// Both the opener and the closer show the row evenly. They differ in
			// what else is on the frame — the closer carries the closing line —
			// which is the component's business, not the timeline's.
			step["lit"] = all
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneCards,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":    in.Plan.Title,
			"relation": c.ResolvedRelation(),
			"items":    items,
			"closer":   c.Closer,
			"ask":      c.Ask,
			"steps":    steps,
		}),
	}}, nil
}
