package pipeline

// The eras template: how we got here, in one band.
//
// A foundations course keeps needing the same short move — "before we look at
// how this works, here is where it came from" — and the catalog has had no
// good picture for it. Told as prose it becomes trivia, a list of decades and
// inventions with nothing holding them together, which is exactly the version
// of computing history that convinces a career-switcher the past is optional.
//
// The picture that fixes that is a band, not a list. Each era is a segment
// with its decade set above it in mono and one defining artifact inside it, so
// the whole span is present from the first frame and every era is read against
// its neighbours rather than in isolation. Then the arcs. Each era hands
// something to the next — a constraint solved, a cost that collapsed, an idea
// that outlived its hardware — and when those hand-offs draw as a continuous
// dashed thread across the band, the argument stops being "these things
// happened" and becomes "each of these caused the next". That thread is the
// whole reason this template exists, and it is why history in a CS course is
// worth forty seconds.
//
// Two rules are enforced because they are the two ways the picture stops
// arguing. Eras must be focused in order, once each, since history is walked
// forward and a clip that jumps back to the 1950s after the 1970s has given up
// on causation for the sake of a nice sentence. And a thread beat needs at
// least two eras carrying something forward, because one hand-off is not a
// thread — it is an anecdote, drawn as a single arc, and it would make the
// continuity claim look thinner than the truth.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "eras",
		Category:    CatConcepts,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "How we got here",
		Description: "A band of generations with a decade above each and one defining artifact inside it, then the hand-offs drawn as arcs so the history reads as causation rather than trivia. Reach for it when a subject only makes sense against what came before — why files look the way they do, how the cloud happened, where the terminal came from.",
		Example:     "From vacuum tubes to the cloud in five generations",
		PromptFile:  snippetErasTemplateName,
		NeedsCode:   false,
		// Six segments, a thread and a closer. Under thirty seconds the band is
		// a slideshow of decades nobody can read.
		MinTargetSec:     30,
		DefaultTargetSec: 45,
		// The band, six eras and the closer. Six eras leave no beat for the
		// thread, which is the honest trade: a wider history or a stated one.
		MaxBeats:  8,
		Owns:      beatFields{Eras: true},
		OwnsPlan:  planFields{Eras: true},
		Normalize: normalizeErasPlan,
		Validate:  validateErasPlan,
		Scenes:    erasScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(ErasShows(), ", "),
				"MinEras":       minErasEras,
				"MaxEras":       maxErasEras,
				"MaxLabelWords": maxErasLabelWords,
				"MaxWhenWords":  maxErasWhenWords,
				"MaxMarkWords":  maxErasMarkWords,
				"MaxCarryWords": maxErasCarryWords,
				"MinCarries":    minErasThreadCarries,
			}
		},
	})
}

const snippetErasTemplateName = "snippet_eras.tmpl"

const (
	// Below three eras there is no arc of history, only a before and an after,
	// and the band degenerates into a comparison. Past six the segments are
	// narrower than the artifact line inside them needs.
	minErasEras = 3
	maxErasEras = 6

	// An era is a name — "vacuum tubes", "the microprocessor".
	maxErasLabelWords = 3
	// A when is a decade or a year set in mono above the segment — "1940s".
	maxErasWhenWords = 2
	// The mark is the one defining thing, a line inside the segment.
	maxErasMarkWords = 10
	// A carry is the hand-off written along an arc, so it has to be short
	// enough to sit on a curve without wrapping.
	maxErasCarryWords = 8

	// One hand-off is an anecdote. Two is a thread, which is the claim the
	// beat exists to make.
	minErasThreadCarries = 2
)

// erasShows is the closed vocabulary of what a beat does.
var erasShows = map[string]bool{
	// The whole band, every segment dim. The opener.
	"band": true,
	// Era At lights: its card lifts with its when and its mark.
	"era": true,
	// The carry hand-offs draw as arcs across the band — the continuity claim.
	"thread": true,
	// The closer: the last era bright, with the line running on to today.
	"now": true,
}

// ErasShows returns the beat vocabulary sorted.
func ErasShows() []string {
	out := make([]string, 0, len(erasShows))
	for k := range erasShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ErasSpec is the band. On the plan because every era is on screen from the
// first frame — beats light them, they do not introduce them.
type ErasSpec struct {
	// Eras are the generations, oldest first.
	Eras []ErasEra `json:"eras"`
}

// ErasEra is one generation.
type ErasEra struct {
	// Label names the era — "vacuum tubes".
	Label string `json:"label"`
	// When is the decade above the segment — "1940s".
	When string `json:"when"`
	// Mark is the defining thing inside the segment.
	Mark string `json:"mark"`
	// Carry is what this era handed the next, written along the arc to it.
	// Optional: an era that changed nothing downstream should say so by
	// leaving this empty rather than by inventing a hand-off.
	Carry string `json:"carry,omitempty"`
}

// ErasBeat is one shot of the band.
type ErasBeat struct {
	// Show is an erasShows name.
	Show string `json:"show"`
	// At indexes ErasSpec.Eras, for the "era" beats.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults the unknown to one era in focus, which is what most
// beats of this template are.
func (b ErasBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if erasShows[s] {
		return s
	}
	return "era"
}

func normalizeErasPlan(p *SnippetPlan) {
	e := p.Eras
	if e == nil {
		return
	}
	eras := make([]ErasEra, 0, len(e.Eras))
	for _, era := range e.Eras {
		era.Label = clampWords(collapseSpaces(era.Label), maxErasLabelWords)
		era.When = clampWords(collapseSpaces(era.When), maxErasWhenWords)
		era.Mark = clampWords(collapseSpaces(era.Mark), maxErasMarkWords)
		era.Carry = clampWords(collapseSpaces(era.Carry), maxErasCarryWords)
		if len(eras) < maxErasEras {
			eras = append(eras, era)
		}
	}
	e.Eras = eras

	for i := range p.Beats {
		b := p.Beats[i].Eras
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "era" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(e.Eras); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateErasPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Eras: true}); err != nil {
		return err
	}

	e := p.Eras
	if e == nil {
		return fmt.Errorf("the plan has no eras — this template is a band of generations, so the eras are the clip")
	}
	if n := len(e.Eras); n < minErasEras || n > maxErasEras {
		return fmt.Errorf("the band has %d eras, want %d-%d. Below three there is no arc of history, only a before and an after — versus draws that better; past %d the segments are narrower than the line of text inside them needs",
			n, minErasEras, maxErasEras, maxErasEras)
	}
	carries := 0
	for i, era := range e.Eras {
		if strings.TrimSpace(era.Label) == "" {
			return fmt.Errorf("era %d has no label. A segment nobody can name is a coloured band, and the viewer cannot tell one generation from the next", i)
		}
		if strings.TrimSpace(era.When) == "" {
			return fmt.Errorf("era %d (%q) has no date. The decade above the segment is what makes the band a history rather than a list of categories", i, era.Label)
		}
		if strings.TrimSpace(era.Mark) == "" {
			return fmt.Errorf("era %d (%q) has no defining thing. An era with no artifact inside it is a label with a date on it, which is the trivia version of this clip", i, era.Label)
		}
		if strings.TrimSpace(era.Carry) != "" {
			carries++
		}
	}

	if p.Beats[0].Eras == nil || p.Beats[0].Eras.ResolvedShow() != "band" {
		return fmt.Errorf("beat %q does not open on the whole band. An era lighting up on a span the viewer has not seen is a card with a date on it — open with {\"show\": \"band\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Eras == nil || last.Eras.ResolvedShow() != "now" {
		return fmt.Errorf("beat %q does not close on now. The whole point of walking the band is arriving at the present, and ending mid-history leaves the viewer in a decade they do not live in — end with {\"show\": \"now\"}", last.ID)
	}

	next := 0
	for _, b := range p.Beats {
		d := b.Eras
		if d == nil {
			return fmt.Errorf("beat %q has no eras direction — every beat shows the band, lights an era, draws the thread, or arrives at now", b.ID)
		}
		switch d.ResolvedShow() {
		case "era":
			if d.At < 0 || d.At >= len(e.Eras) {
				return fmt.Errorf("beat %q lights era %d, which does not exist — the band has eras 0-%d", b.ID, d.At, len(e.Eras)-1)
			}
			if d.At != next {
				return fmt.Errorf("beat %q lights era %d (%q, %s) when era %d (%q, %s) is the next one due. History is walked forward, once through — jumping back for a nice sentence gives up the causation the band exists to show",
					b.ID, d.At, e.Eras[d.At].Label, e.Eras[d.At].When, next, e.Eras[next].Label, e.Eras[next].When)
			}
			next++
		case "thread":
			// Counted in Go, because "each one led to the next" is the claim
			// this beat makes and a single arc cannot make it.
			if carries < minErasThreadCarries {
				return fmt.Errorf("beat %q draws the thread, but only %d era carries anything forward. One hand-off is an anecdote, not a thread — give at least %d eras a \"carry\" saying what they handed the next, or drop this beat",
					b.ID, carries, minErasThreadCarries)
			}
		}
	}
	if next != len(e.Eras) {
		return fmt.Errorf("the clip lights %d of %d eras, so %q (%s) sits on the band with nothing said about it. Every era needs its own beat, or drop it from the history",
			next, len(e.Eras), e.Eras[next].Label, e.Eras[next].When)
	}
	return nil
}

// erasScenes lays the clip out as ONE scene. The hand-off arcs are resolved
// here — which era hands to which, and what the arc is labelled — so the
// component draws a thread it was given rather than pairing eras itself.
func erasScenes(in SnippetSceneInput) ([]Scene, error) {
	e := in.Plan.Eras
	if e == nil {
		return nil, fmt.Errorf("the plan has no eras")
	}
	if len(e.Eras) == 0 {
		return nil, fmt.Errorf("the band has no eras")
	}

	eras := make([]map[string]any, len(e.Eras))
	threads := make([]map[string]any, 0, len(e.Eras))
	for i, era := range e.Eras {
		eras[i] = map[string]any{
			"label": era.Label,
			"when":  era.When,
			"mark":  era.Mark,
			"carry": era.Carry,
		}
		// An arc needs somewhere to land, so the last era's carry is the line
		// that runs on to today rather than a hop between segments.
		if strings.TrimSpace(era.Carry) != "" && i < len(e.Eras)-1 {
			threads = append(threads, map[string]any{
				"from":  i,
				"to":    i + 1,
				"carry": era.Carry,
			})
		}
	}
	carryNow := strings.TrimSpace(e.Eras[len(e.Eras)-1].Carry)

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	lit := make([]int, 0, len(e.Eras))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Eras == nil {
			return nil, fmt.Errorf("beat %q has no eras direction", beat.ID)
		}
		show := beat.Eras.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "era":
			at := beat.Eras.At
			if at < 0 || at >= len(e.Eras) {
				return nil, fmt.Errorf("beat %q lights era %d, which does not exist", beat.ID, at)
			}
			lit = append(lit, at)
			step["at"] = at
		case "now":
			// The closer arrives at the present with the whole walk behind it.
			lit = lit[:0]
			for j := range e.Eras {
				lit = append(lit, j)
			}
			step["at"] = len(e.Eras) - 1
		}
		up := make([]int, len(lit))
		copy(up, lit)
		sort.Ints(up)
		step["lit"] = up
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneEras,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":    in.Plan.Title,
			"eras":     eras,
			"threads":  threads,
			"carryNow": carryNow,
			"steps":    steps,
		}),
	}}, nil
}
