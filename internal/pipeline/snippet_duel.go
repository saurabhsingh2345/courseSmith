package pipeline

// The duel template: two named products, one measured axis, and a call.
//
// The catalog could already argue between two things three ways, so the first
// question is what this adds. `compare` introduces two subjects and judges them
// in prose. `versus` runs a table: several dimensions, a row landing at a time,
// a verdict at the end. Both are arguments made of WORDS about things the frame
// never shows. This one shows the things — the real marks, on two cards — and
// reduces the argument to a single number each, drawn as a bar.
//
// That reduction is the template, and it is a strong claim about the subject:
// that there is one axis the choice actually turns on. Most comparisons do not
// qualify, which is why `versus` still exists. But when one does — free against
// paid, one model's capability against another's, a managed service against
// running it yourself — the two-bar picture is worth more than a five-row table,
// because a viewer can see the size of the gap instead of reading that there is
// one.
//
// Four rules earn the shape.
//
// The scores must DIFFER. Two bars of the same length is a picture that says
// nothing, and a clip whose picture says nothing has spent its runtime drawing
// furniture. If the honest answer is that they are level on this axis, the axis
// is not the one the choice turns on, and this is the wrong template.
//
// Each contender carries a TAG — the pill under its name. Not decoration: the
// tag is what makes the two cards commensurable at a glance ("Free" against
// "$20/mo", "managed" against "self-hosted"), and it is the field that most often
// carries the reason the lower bar wins.
//
// The clip must PICK one. Same rule as the cards template's closer and the versus
// template's verdict, and it is the same failure without it: a frame with two bars
// on it and no call is a scoreboard the viewer has to interpret alone.
//
// And the pick is allowed to be the SHORTER bar. That is deliberately not
// validated, because it is the most useful thing this template can say — the free
// tier is worse and is still the right answer for most people — and forcing the
// pick to follow the measurement would make the template incapable of saying it.
// What the validator does instead is insist the verdict is more than a name, so
// there is somewhere for the reason to live.

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "duel",
		Category: CatDecisions,
		Since:    SinceV9,
		Family:   FamilyShowroom,
		Title:    "Two products, one axis",
		Description: "Two named things face to face, each with its real logo, a status pill and a bar measuring the one thing the choice turns on — then a call. " +
			"Reach for it when a decision between two products really does come down to a single axis.",
		Example:    "ChatGPT free or Gemini paid: what the money actually buys",
		PromptFile: snippetDuelTemplateName,
		NeedsCode:  false,
		// The same fetch the cards template runs, for the same reason.
		Plan: planDuelSnippet,
		// Both cards up, one beat each, the bars, the call: five beats is the floor
		// and there is no shorter honest version of this picture.
		MinTargetSec:     35,
		DefaultTargetSec: 50,
		MaxBeats:         7,
		// A beat is a shot here as it is in cards, and the shots are longer: the
		// bar-filling beat has an animation to cover and the verdict beat has to
		// land a recommendation.
		IdealWordsPerBeat: 26,
		Owns:              beatFields{Duel: true},
		OwnsPlan:          planFields{Duel: true},
		Normalize:         normalizeDuelPlan,
		Validate:          validateDuelPlan,
		Scenes:            duelScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":           strings.Join(MetricRoles(), ", "),
				"Shows":           strings.Join(DuelShows(), ", "),
				"Icons":           strings.Join(PointIconNames(), ", "),
				"MaxTitleWords":   maxDuelTitleWords,
				"MaxTagWords":     maxDuelTagWords,
				"MaxNoteWords":    maxDuelNoteWords,
				"MaxAxisWords":    maxDuelAxisWords,
				"MaxVerdictWords": maxDuelVerdictWords,
				"MinScoreGap":     minDuelScoreGap,
			}
		},
	})
}

const snippetDuelTemplateName = "snippet_duel.tmpl"

const (
	// A NAME, set in display type.
	maxDuelTitleWords = 3
	// The pill under the name: "Free", "$20 a month", "self-hosted".
	maxDuelTagWords = 3
	// The line under the pill. One clause.
	maxDuelNoteWords = 14
	// What the bars measure, set small-caps above them.
	maxDuelAxisWords = 3
	// The call, across the foot of the frame.
	maxDuelVerdictWords = 16

	// How far apart the two scores have to be to be a picture.
	//
	// Twelve points of a hundred is about four percent of the bar's width at the
	// size these are drawn, which is roughly the smallest length difference the
	// eye reads as deliberate rather than as two bars that were meant to match.
	// Below it the frame is claiming a gap the viewer cannot see.
	minDuelScoreGap = 12
)

// duelShows is the closed vocabulary of what a beat does.
var duelShows = map[string]bool{
	// Both cards up, neither favoured, no bars yet. The opener.
	"pair": true,
	// One contender lit, the other receded.
	"card": true,
	// Both bars fill to their scores. The measurement.
	"bars": true,
	// The pick rimmed and the call landing. The closer.
	"call": true,
}

// DuelShows returns the beat vocabulary sorted.
func DuelShows() []string {
	out := make([]string, 0, len(duelShows))
	for k := range duelShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// DuelSpec is the face-off. On the plan rather than on a beat because both cards
// are on screen for the whole clip; the beats move light and fill bars.
type DuelSpec struct {
	// Sides are the two contenders, left and right. A slice rather than two named
	// fields so the beat vocabulary can index it the way every other template in
	// the catalog indexes its items — {"show": "card", "at": 1} — instead of
	// inventing a left/right enum only this template understands.
	Sides []Contender `json:"sides"`
	// Axis is what the bars measure, one label for both — "capability", "speed",
	// "monthly cost". Required: two bars with no axis named is a chart with no
	// units, and the viewer cannot tell whether longer is better.
	Axis string `json:"axis"`
	// Pick indexes Sides: which one the clip recommends. May be the lower score.
	Pick int `json:"pick"`
	// Verdict is the call, under the finished frame. Required.
	Verdict string `json:"verdict"`
}

// Contender is one side of the duel.
type Contender struct {
	// Title is the thing's NAME.
	Title string `json:"title"`
	// Tag is the pill under the name — the terms, the tier, the deployment.
	Tag string `json:"tag"`
	// Note is one line of what it is.
	Note string `json:"note"`
	// Score is where its bar reaches, 0-100. Higher is more of Axis, whatever
	// Axis is — a cost duel draws the expensive one longer, and the verdict is
	// where "longer is worse" gets said.
	Score int `json:"score"`
	// Role tints the card: a metricRoles name.
	Role string `json:"role,omitempty"`

	// Where the mark comes from, resolved exactly as a Card's is. Shared with the
	// cards template through resolveContenderArt below rather than duplicated:
	// there is one fallback chain in this pipeline and one place it lives.
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

// asCard views a contender as a Card, so the art resolver written for the cards
// template works on it unchanged.
func (c *Contender) asCard() Card {
	return Card{
		Title: c.Title, Note: c.Note, Role: c.Role,
		Brand: c.Brand, Site: c.Site, ImageURL: c.ImageURL, Icon: c.Icon,
		Mark: c.Mark, Tint: c.Tint, Image: c.Image, MarkFrom: c.MarkFrom,
	}
}

// takeArt copies the resolved fields back off a Card.
func (c *Contender) takeArt(from Card) {
	c.Mark, c.Tint, c.Image, c.MarkFrom = from.Mark, from.Tint, from.Image, from.MarkFrom
}

// ResolvedRole returns the contender's tint, defaulting to neutral.
func (c Contender) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(c.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// ResolvedIcon returns the fallback glyph. Never empty.
func (c Contender) ResolvedIcon() string {
	if n := normalizePointIconName(c.Icon); n != "" {
		return n
	}
	return "box"
}

// DuelBeat is one shot of the face-off.
type DuelBeat struct {
	// Show is a duelShows name.
	Show string `json:"show"`
	// At indexes DuelSpec.Sides, for a "card" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults the unknown to a card landing.
func (b DuelBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if duelShows[s] {
		return s
	}
	return "card"
}

// planDuelSnippet plans the clip and then goes and gets the two marks.
func planDuelSnippet(ctx context.Context, e *Env, spec SnippetSpec, cfg config.Config) (*SnippetPlan, error) {
	plan, err := planSnippetDefault(ctx, e, spec, cfg)
	if err != nil {
		return nil, err
	}
	if d := plan.Duel; d != nil {
		// Borrowed wholesale: one fallback chain, one cache, one set of hard-won
		// facts about which CDN needs which header.
		row := &CardsSpec{Items: make([]Card, len(d.Sides))}
		for i := range d.Sides {
			row.Items[i] = d.Sides[i].asCard()
		}
		resolveCardArt(ctx, e, row)
		for i := range d.Sides {
			d.Sides[i].takeArt(row.Items[i])
		}
	}
	return plan, nil
}

func normalizeDuelPlan(p *SnippetPlan) {
	d := p.Duel
	if d == nil {
		return
	}
	d.Axis = clampWords(collapseSpaces(d.Axis), maxDuelAxisWords)
	d.Verdict = clampWords(collapseSpaces(d.Verdict), maxDuelVerdictWords)

	sides := make([]Contender, 0, len(d.Sides))
	for _, s := range d.Sides {
		s.Title = clampWords(collapseSpaces(s.Title), maxDuelTitleWords)
		s.Tag = clampWords(collapseSpaces(s.Tag), maxDuelTagWords)
		s.Note = clampWords(collapseSpaces(s.Note), maxDuelNoteWords)
		s.Role = s.ResolvedRole()
		s.Icon = s.ResolvedIcon()
		s.Brand = cardsSlug(s.Brand)
		s.Site = cardsHost(s.Site)
		s.Score = min(max(s.Score, 0), 100)
		if s.Title != "" && len(sides) < duelSides {
			sides = append(sides, s)
		}
	}
	d.Sides = sides

	if d.Pick < 0 || d.Pick >= len(d.Sides) {
		d.Pick = 0
	}

	for i := range p.Beats {
		b := p.Beats[i].Duel
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
		if n := len(d.Sides); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

// duelSides is not a range. Two is the template.
const duelSides = 2

func validateDuelPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Duel: true}); err != nil {
		return err
	}

	d := p.Duel
	if d == nil {
		return fmt.Errorf("the plan has no face-off — this template is two named things measured against each other, so the pair is the clip")
	}
	if len(d.Sides) != duelSides {
		return fmt.Errorf("the face-off has %d sides, want exactly %d. One thing measured against nothing is a statistic; three is a row, which is what the cards template draws", len(d.Sides), duelSides)
	}
	if strings.TrimSpace(d.Axis) == "" {
		return fmt.Errorf("the bars have no axis named. Two bars with no label is a chart with no units — the viewer cannot even tell whether longer is better. Name what is being measured: \"capability\", \"speed\", \"monthly cost\"")
	}
	for i, s := range d.Sides {
		if strings.TrimSpace(s.Title) == "" {
			return fmt.Errorf("side %d has no name. The name is what the mark above it identifies", i)
		}
		if strings.TrimSpace(s.Tag) == "" {
			return fmt.Errorf("side %d (%q) has no tag. The pill under the name is what makes the two cards comparable at a glance — the tier, the price, the way it is run — and it is usually where the reason the shorter bar wins actually lives", i, s.Title)
		}
		if strings.TrimSpace(s.Note) == "" {
			return fmt.Errorf("side %d (%q) has no line saying what it is", i, s.Title)
		}
	}
	if cardsFold(d.Sides[0].Title) == cardsFold(d.Sides[1].Title) {
		return fmt.Errorf("both sides are %q, so the frame sets one thing against itself", d.Sides[0].Title)
	}

	// The rule the template rests on.
	gap := d.Sides[0].Score - d.Sides[1].Score
	if gap < 0 {
		gap = -gap
	}
	if gap < minDuelScoreGap {
		return fmt.Errorf("the two bars are %d and %d, %d apart. Two bars the same length is a picture that says nothing — and if %q and %q really are level on %q, then that is not the axis the choice turns on and this is the wrong template for it. Pick the axis where they genuinely differ, or use versus to compare them across several",
			d.Sides[0].Score, d.Sides[1].Score, gap, d.Sides[0].Title, d.Sides[1].Title, d.Axis)
	}

	verdict := strings.TrimSpace(d.Verdict)
	if verdict == "" {
		return fmt.Errorf("the clip draws two bars and makes no call. That is a scoreboard the viewer has to interpret alone — say which one to reach for and when")
	}
	for _, s := range d.Sides {
		if cardsFold(verdict) == cardsFold(s.Title) {
			return fmt.Errorf("the verdict is just %q, which is a name rather than a call. Say when to reach for it — and, if it is the shorter bar, why that is still the right answer", d.Verdict)
		}
	}

	if p.Beats[0].Duel == nil || p.Beats[0].Duel.ResolvedShow() != "pair" {
		return fmt.Errorf("beat %q does not open on the pair. Both cards are on screen from the first frame — the viewer should know who is fighting before either one is described — so open with {\"show\": \"pair\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Duel == nil || last.Duel.ResolvedShow() != "call" {
		return fmt.Errorf("beat %q does not close on the call. Ending on a bar leaves the recommendation unsaid — end with {\"show\": \"call\"}", last.ID)
	}

	// Both contenders introduced, in order, before anything is measured. The
	// ordering matters more here than in a row of five: a bar drawn for something
	// the viewer has not met yet is a number about a stranger.
	next := 0
	bars := -1
	for i, b := range p.Beats {
		if b.Duel == nil {
			return fmt.Errorf("beat %q has no duel direction — every beat raises the pair, lights one side, fills the bars, or makes the call", b.ID)
		}
		switch b.Duel.ResolvedShow() {
		case "card":
			if next >= duelSides {
				return fmt.Errorf("beat %q lights a side when both have already had their beat. Each side is introduced once, left then right — a second pass says the first introduction did not land", b.ID)
			}
			if b.Duel.At != next {
				return fmt.Errorf("beat %q lights side %d (%q) when side %d (%q) is next. Left then right, once each",
					b.ID, b.Duel.At, d.Sides[b.Duel.At].Title, next, d.Sides[next].Title)
			}
			next++
		case "bars":
			if bars >= 0 {
				return fmt.Errorf("beat %q fills the bars a second time. They fill once and stay filled — refilling them says the first measurement was not believed", b.ID)
			}
			if next < duelSides {
				return fmt.Errorf("beat %q measures %q before both sides have been introduced. A bar drawn for something the viewer has not met is a number about a stranger — introduce %q first", b.ID, d.Axis, d.Sides[next].Title)
			}
			bars = i
		}
	}
	if next < duelSides {
		return fmt.Errorf("the clip introduces %d of the 2 sides, so %q is on screen with nothing said about it", next, d.Sides[next].Title)
	}
	if bars < 0 {
		return fmt.Errorf("no beat fills the bars, so %q is never measured and the frame is two cards with an empty chart under them. Add a {\"show\": \"bars\"} beat after both sides are introduced", d.Axis)
	}
	return nil
}

// duelScenes lays the clip out as ONE scene, the way cards does: the pair
// persists and the steps say what is lit and whether the bars have filled.
func duelScenes(in SnippetSceneInput) ([]Scene, error) {
	d := in.Plan.Duel
	if d == nil {
		return nil, fmt.Errorf("the plan has no face-off")
	}
	if len(d.Sides) != duelSides {
		return nil, fmt.Errorf("the face-off has %d sides, want %d", len(d.Sides), duelSides)
	}

	sides := make([]map[string]any, len(d.Sides))
	for i, s := range d.Sides {
		sides[i] = map[string]any{
			"title": s.Title,
			"tag":   s.Tag,
			"note":  s.Note,
			"score": s.Score,
			"role":  s.ResolvedRole(),
			"icon":  s.ResolvedIcon(),
		}
		if s.Mark != "" {
			sides[i]["mark"] = s.Mark
		}
		if s.Tint != "" {
			sides[i]["tint"] = s.Tint
		}
		if s.Image != "" {
			sides[i]["image"] = s.Image
		}
		if s.MarkFrom != "" {
			sides[i]["markFrom"] = s.MarkFrom
		}
	}

	// Once the bars have filled they stay filled for the rest of the clip, so the
	// step carries `bars` as a state rather than the component replaying the
	// timeline to work out whether the measurement has happened yet.
	filled := false
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Duel == nil {
			return nil, fmt.Errorf("beat %q has no duel direction", beat.ID)
		}
		show := beat.Duel.ResolvedShow()
		if show == "bars" {
			filled = true
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"bars":    filled,
		}
		switch show {
		case "card":
			at := beat.Duel.At
			if at < 0 || at >= len(d.Sides) {
				return nil, fmt.Errorf("beat %q lights side %d, which does not exist", beat.ID, at)
			}
			step["at"] = at
			step["lit"] = []int{at}
		case "call":
			step["lit"] = []int{d.Pick}
		default:
			step["lit"] = []int{0, 1}
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneDuel,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":   in.Plan.Title,
			"sides":   sides,
			"axis":    d.Axis,
			"pick":    d.Pick,
			"verdict": d.Verdict,
			"steps":   steps,
		}),
	}}, nil
}
