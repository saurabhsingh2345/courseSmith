package pipeline

// The lookup template: one key, resolved in steps.
//
// A career-switcher meets the same shape a dozen times in a foundations course
// and is never told it is the same shape. A virtual address becomes a physical
// frame through a page table. A hostname becomes an IP address through a
// resolver, a root server, a TLD server and finally an authoritative one. A
// file path becomes an inode through a directory. A symbol becomes a value
// through a scope chain. Every one of those is a key that nobody holds the
// answer to, handed along a chain of parties until one of them does.
//
// The catalog could draw a flow of stages and a hand-off between two parties;
// it had nothing that draws the QUESTION travelling. That distinction is the
// whole template: the picture is not the tables, it is one card moving between
// them collecting stamps, because the thing a beginner gets wrong is thinking
// the answer was somewhere all along rather than assembled on the way.
//
// So the validator enforces a walk, not a diagram. The hops are visited in
// order and none is skipped, because a chain with a station missed is not a
// shorter chain, it is a wrong one — the whole claim of the picture is that
// each party only knows enough to point at the next. The answer comes back
// exactly once, and only after the last hop has been asked, because a hit that
// lands mid-chain says the remaining stations were decoration. And the cache
// shortcut, when it is drawn at all, is drawn last: it is the arc that exists
// BECAUSE the long way was walked first, and a shortcut shown before the road
// is a shortcut past nothing.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "lookup",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "Resolved in steps",
		Description: "One key carried along a chain of tables, collecting an answer at each stop until the last one resolves it. Reach for it when the subject is a name being turned into a thing — DNS, page tables, inodes, scope chains — and the point is that no single party knows the answer.",
		Example:     "How a hostname becomes an IP address",
		PromptFile:  snippetLookupTemplateName,
		NeedsCode:   false,
		// The key, two to five stops, the return and the shortcut: under
		// thirty-five seconds the card arrives at a station before the viewer
		// has read what the previous one said.
		MinTargetSec:     35,
		DefaultTargetSec: 50,
		// Opener + up to five hops + the hit + the cache. Past eight the chain
		// is longer than the picture can hold at a legible station width.
		MaxBeats: 8,
		// A beat here is a SHOT — the card moving one station — not a step in
		// an argument. Twenty-eight words is about nine seconds, which is how
		// long one stamp landing stays interesting.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Lookup: true},
		OwnsPlan:          planFields{Lookup: true},
		Normalize:         normalizeLookupPlan,
		Validate:          validateLookupPlan,
		Scenes:            lookupScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":          strings.Join(MetricRoles(), ", "),
				"Shows":          strings.Join(LookupShows(), ", "),
				"MinHops":        minLookupHops,
				"MaxHops":        maxLookupHops,
				"MaxKeyWords":    maxLookupKeyWords,
				"MaxWhereWords":  maxLookupWhereWords,
				"MaxGivesWords":  maxLookupGivesWords,
				"MaxMissWords":   maxLookupMissWords,
				"MaxAnswerWords": maxLookupAnswerWords,
			}
		},
	})
}

const snippetLookupTemplateName = "snippet_lookup.tmpl"

const (
	// One stop is not a chain, it is a function call — the point of the picture
	// is that the answer is assembled across parties, which needs two.
	minLookupHops = 2
	// Past five the stations are narrower than their own captions on a 1920
	// frame, and every hop wants its own beat, which no runtime funds.
	maxLookupHops = 5

	// The key rides on a card a few hundred pixels wide, set in mono.
	maxLookupKeyWords = 6
	// A station's name is a nameplate: "the root server", "page table".
	maxLookupWhereWords = 4
	// What a station answers is one line under its plate.
	maxLookupGivesWords = 8
	// The miss line is the small print beside a station and competes with the
	// answer for the same space, so it is held to the same length.
	maxLookupMissWords = 8
	// The answer is the payoff, set large. Six words is a value; more is a
	// sentence pretending to be one.
	maxLookupAnswerWords = 6
)

// lookupShows is the closed vocabulary of what a beat does.
var lookupShows = map[string]bool{
	// The key alone, before anybody has been asked. The opener.
	"ask": true,
	// The card travels to hop At and that station's answer stamps onto it.
	"hop": true,
	// The answer returns to the asker along the lit chain.
	"hit": true,
	// The shortcut arc is drawn over the top, for next time. Optional closer.
	"cache": true,
}

// LookupShows returns the beat vocabulary sorted.
func LookupShows() []string {
	out := make([]string, 0, len(lookupShows))
	for k := range lookupShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// LookupSpec is the question, the chain and the answer. On the plan because
// the same key is on screen for the whole clip.
type LookupSpec struct {
	// Key is what is being resolved — "www.example.com".
	Key string `json:"key"`
	// Hops are the stations asked, in the order they are asked.
	Hops []LookupHop `json:"hops"`
	// Answer is what the key resolves to — "93.184.216.34".
	Answer string `json:"answer"`
}

// LookupHop is one station on the chain.
type LookupHop struct {
	// Where names the table or party asked — "the root server".
	Where string `json:"where"`
	// Gives is what this station answers — "go ask the .com servers".
	Gives string `json:"gives"`
	// Miss is what happens when this station does not know. Optional.
	Miss string `json:"miss,omitempty"`
}

// LookupBeat is one shot: which state of the chain this beat shows.
type LookupBeat struct {
	// Show is a lookupShows name.
	Show string `json:"show"`
	// At indexes LookupSpec.Hops, for a "hop" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a hop —
// the workhorse state most beats of this template are in.
func (b LookupBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if lookupShows[s] {
		return s
	}
	return "hop"
}

func normalizeLookupPlan(p *SnippetPlan) {
	l := p.Lookup
	if l == nil {
		return
	}
	l.Key = clampWords(collapseSpaces(l.Key), maxLookupKeyWords)
	l.Answer = clampWords(collapseSpaces(l.Answer), maxLookupAnswerWords)

	hops := make([]LookupHop, 0, len(l.Hops))
	for _, h := range l.Hops {
		h.Where = clampWords(collapseSpaces(h.Where), maxLookupWhereWords)
		h.Gives = clampWords(collapseSpaces(h.Gives), maxLookupGivesWords)
		h.Miss = clampWords(collapseSpaces(h.Miss), maxLookupMissWords)
		if len(hops) < maxLookupHops {
			hops = append(hops, h)
		}
	}
	l.Hops = hops

	for i := range p.Beats {
		b := p.Beats[i].Lookup
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "hop" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(l.Hops); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateLookupPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Lookup: true}); err != nil {
		return err
	}

	l := p.Lookup
	if l == nil {
		return fmt.Errorf("the plan has no lookup — this template is one key carried along a chain of tables, so the key and the chain are the clip")
	}
	if strings.TrimSpace(l.Key) == "" {
		return fmt.Errorf("the plan gives no key. The card that travels the chain has to have something written on it — a hostname, a virtual address, a path")
	}
	if strings.TrimSpace(l.Answer) == "" {
		return fmt.Errorf("the plan gives no answer. The clip is a question being resolved, so the thing it resolves TO is the last frame the viewer keeps")
	}
	if n := len(l.Hops); n < minLookupHops || n > maxLookupHops {
		return fmt.Errorf("the chain has %d hop(s), want %d-%d. One stop is a function call rather than a chain — the point of the picture is that the answer is assembled across parties — and past %d the stations are narrower than their own captions",
			n, minLookupHops, maxLookupHops, maxLookupHops)
	}
	for i, h := range l.Hops {
		if strings.TrimSpace(h.Where) == "" {
			return fmt.Errorf("hop %d has no name — an unlabelled station is a box the card visits for no stated reason", i)
		}
		if strings.TrimSpace(h.Gives) == "" {
			return fmt.Errorf("hop %d (%q) answers nothing. The stamp IS the answer: a station the card visits and leaves unchanged is a stop the chain did not need", i, h.Where)
		}
	}

	// The shape. The key is seen alone before anybody is asked.
	if p.Beats[0].Lookup == nil || p.Beats[0].Lookup.ResolvedShow() != "ask" {
		return fmt.Errorf("beat %q does not open on the key alone. The stations mean nothing until there is a question travelling between them — the first beat is {\"show\": \"ask\"}",
			p.Beats[0].ID)
	}
	// THE ORDER, which is the claim of the picture: each party knows only
	// enough to point at the next, so the walk is checked hop by hop.
	next := 0
	lastHopBeat := -1
	hits, caches, hitAt := 0, 0, -1
	for i, b := range p.Beats {
		if b.Lookup == nil {
			return fmt.Errorf("beat %q has no lookup direction — every beat shows one state of the chain", b.ID)
		}
		switch b.Lookup.ResolvedShow() {
		case "ask":
			if i != 0 {
				return fmt.Errorf("beat %q goes back to the bare key part-way through. The ask is the opener; returning to it un-stamps a card the chain already answered", b.ID)
			}
		case "hop":
			at := b.Lookup.At
			if at < 0 || at >= len(l.Hops) {
				return fmt.Errorf("beat %q visits hop %d, which does not exist — the chain holds hops 0-%d", b.ID, at, len(l.Hops)-1)
			}
			if at != next {
				return fmt.Errorf("beat %q visits hop %d, but the card is standing at hop %d. The chain is walked in order with no skips: each station only knows enough to point at the next one, so jumping to %d means the viewer never learns who sent the card there",
					b.ID, at, next, at)
			}
			next++
			lastHopBeat = i
		case "hit":
			hits++
			if hits > 1 {
				return fmt.Errorf("beat %q returns the answer a second time. The answer comes back once — a rerun says the first return did not count", b.ID)
			}
			hitAt = i
		case "cache":
			caches++
			if caches > 1 {
				return fmt.Errorf("beat %q draws the shortcut a second time. There is one cache arc, and drawing it twice says the first one was not kept", b.ID)
			}
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q draws the cache shortcut before the end. The shortcut is the closer: it exists BECAUSE the long way was walked, so anything after it is the clip walking the road it just replaced", b.ID)
			}
		}
	}
	if next < len(l.Hops) {
		where := fmt.Sprintf("the card stops at hop %d", next-1)
		if next == 0 {
			where = "the card never leaves the asker"
		}
		return fmt.Errorf("%d of the %d hops are never visited — %s. A chain with a station missed is not a shorter chain, it is a wrong one: give every hop its own beat, or write fewer hops",
			len(l.Hops)-next, len(l.Hops), where)
	}
	if hits == 0 {
		return fmt.Errorf("no beat returns the answer. The clip is a question being resolved, and one that never comes back leaves the card sitting at the last table with the viewer still holding the question")
	}
	if hitAt < lastHopBeat {
		return fmt.Errorf("beat %q returns the answer before hop %d has been asked. A hit that lands mid-chain says the remaining stations were decoration — put the {\"show\": \"hit\"} beat after the last hop",
			p.Beats[hitAt].ID, len(l.Hops)-1)
	}
	return nil
}

// lookupScenes lays the clip out as ONE scene. The chain persists; the steps
// say where the card is standing, which stations have stamped it, and whether
// the answer has come home.
func lookupScenes(in SnippetSceneInput) ([]Scene, error) {
	l := in.Plan.Lookup
	if l == nil {
		return nil, fmt.Errorf("the plan has no lookup")
	}

	hops := make([]map[string]any, len(l.Hops))
	for i, h := range l.Hops {
		hops[i] = map[string]any{
			"where": h.Where,
			"gives": h.Gives,
			"miss":  h.Miss,
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	stamped := map[int]bool{}
	answered, cached := false, false
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Lookup == nil {
			return nil, fmt.Errorf("beat %q has no lookup direction", beat.ID)
		}
		show := beat.Lookup.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "hop":
			stamped[beat.Lookup.At] = true
			step["at"] = beat.Lookup.At
		case "hit":
			answered = true
		case "cache":
			cached = true
		}
		visited := make([]int, 0, len(stamped))
		for at := range stamped {
			visited = append(visited, at)
		}
		sort.Ints(visited)
		step["visited"] = visited
		step["answered"] = answered
		step["cached"] = cached
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneLookup,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"key":    l.Key,
			"answer": l.Answer,
			"hops":   hops,
			"steps":  steps,
		}),
	}}, nil
}
