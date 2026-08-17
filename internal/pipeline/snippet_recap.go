package pipeline

// The recap template: what you already know, and where you got it.
//
// This is the template the course-continuity problem has been waiting for. The
// pipeline has scored the seam between lessons since bridge.tmpl was written —
// 1 to 10, with a suggested rewrite of the opening line — and scoring a seam
// after the fact cannot create one. A recap is the seam, made into a clip that
// the arc can actually contain.
//
// Two rules earn it its place, and they are the same rule pointed two ways.
//
// **Every claim carries the lesson it came from.** A recap whose claims float
// free is indistinguishable from a lesson, and the viewer cannot tell what they
// are supposed to already have. The `from` field is the whole point: it is what
// lets somebody who is lost go to exactly the right place rather than back to
// the beginning.
//
// **A recap introduces nothing.** This is the rule the template exists to
// enforce, and it is the one a model breaks every time, because "remind them
// what a retry is" and "explain what a retry is" produce nearly the same
// sentences. The tell is the `new` flag: a claim the model itself marks as new
// material is rejected outright, with the two templates that DO introduce
// things named in the error. A recap that teaches is a lesson with a
// misleading title, and the viewer who skipped it is now genuinely behind.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:             "recap",
		Category:         CatPresenting,
		Since:            SinceV5,
		Title:            "Where we got to",
		Description:      "What the lessons before this one established, each tagged with where it came from, and the one thread they were all building. Reach for it opening a lesson that depends on several earlier ones.",
		Example:          "Pick up a course after four lessons on queues, before the one about dropping messages",
		PromptFile:       snippetRecapTemplateName,
		MinTargetSec:     20,
		DefaultTargetSec: 40,
		MaxBeats:         7,
		Owns:             beatFields{Recap: true},
		OwnsPlan:         planFields{Recap: true},
		Normalize:        normalizeRecapPlan,
		Validate:         validateRecapPlan,
		Scenes:           recapScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				// The emphasis roles every headline picks from.
				"Roles":          strings.Join(MetricRoles(), ", "),
				"MinClaims":      minRecapClaims,
				"MaxClaims":      maxRecapClaims,
				"MaxClaimWords":  maxRecapClaimWords,
				"MaxFromWords":   maxRecapFromWords,
				"MaxThreadWords": maxRecapThreadWords,
			}
		},
	})
}

const snippetRecapTemplateName = "snippet_recap.tmpl"

const (
	minRecapClaims = 2
	maxRecapClaims = 5

	maxRecapClaimWords  = 12
	maxRecapFromWords   = 10
	maxRecapThreadWords = 16
)

// RecapSpec is what the course has established so far.
type RecapSpec struct {
	// Thread is the one line that says what all of it was building toward. It
	// is what turns a list of claims into a course.
	Thread string `json:"thread"`
	// Claims are what earlier lessons established, in the order they were.
	Claims []RecapClaim `json:"claims"`
}

// RecapClaim is one thing the course already established.
type RecapClaim struct {
	// Claim is the thing itself, in one line.
	Claim string `json:"claim"`
	// From names the lesson it came from, so somebody lost can go there.
	From string `json:"from"`
	// New marks a claim the model knows is not a recap. Always a rejection —
	// the field exists so the model can be honest and the validator can be
	// specific, rather than the model hiding it and the viewer finding out.
	New bool `json:"new,omitempty"`
}

// RecapBeat says which claim this beat is bringing back.
type RecapBeat struct {
	At   int    `json:"at,omitempty"`
	Show string `json:"show"`
}

var recapShows = map[string]bool{
	// One established claim, brought back.
	"claim": true,
	// The thread they were all building.
	"thread": true,
	// Everything at once, and where it leaves the viewer standing.
	"standing": true,
}

// RecapShows returns the beat vocabulary, sorted, for the prompt.
func RecapShows() []string {
	out := make([]string, 0, len(recapShows))
	for s := range recapShows {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// ResolvedShow defaults to bringing back a claim.
func (b RecapBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if recapShows[s] {
		return s
	}
	return "claim"
}

func normalizeRecapPlan(p *SnippetPlan) {
	if p.Recap == nil {
		return
	}
	r := p.Recap
	r.Thread = strings.TrimSpace(r.Thread)
	for i := range r.Claims {
		c := &r.Claims[i]
		c.Claim = strings.TrimSpace(c.Claim)
		c.From = strings.TrimSpace(c.From)
		// "In lesson 3 we saw that…" is a phrasing habit, and the lesson is
		// already carried in From. Stripping it keeps the claim a claim.
		for _, lead := range []string{
			"we saw that ", "we learned that ", "we established that ",
			"you saw that ", "you learned that ", "we covered ",
		} {
			if len(c.Claim) > len(lead) && strings.EqualFold(c.Claim[:len(lead)], lead) {
				c.Claim = strings.TrimSpace(c.Claim[len(lead):])
				break
			}
		}
	}
	for i := range p.Beats {
		if b := p.Beats[i].Recap; b != nil {
			b.Show = b.ResolvedShow()
			if b.At < 0 || b.At >= len(r.Claims) {
				b.At = 0
			}
		}
	}
}

func validateRecapPlan(p *SnippetPlan) error {
	r := p.Recap
	if r == nil {
		return fmt.Errorf("the plan has no recap — this template brings back what earlier lessons established, so the claims are the clip")
	}
	if strings.TrimSpace(r.Thread) == "" {
		return fmt.Errorf("the recap has no thread. A list of things the course said is a list; the thread is the line that says what they were all building toward, and it is the difference between a recap and an index")
	}
	if n := len(strings.Fields(r.Thread)); n > maxRecapThreadWords {
		return fmt.Errorf("the thread is %d words; keep it to %d", n, maxRecapThreadWords)
	}
	if n := len(r.Claims); n < minRecapClaims || n > maxRecapClaims {
		return fmt.Errorf("the recap brings back %d claim(s); this template takes %d to %d. One is a callback rather than a recap, and past %d nobody is remembering, they are taking notes",
			n, minRecapClaims, maxRecapClaims, maxRecapClaims)
	}

	seen := map[string]bool{}
	for i, c := range r.Claims {
		if strings.TrimSpace(c.Claim) == "" {
			return fmt.Errorf("claim %d is empty", i+1)
		}
		if c.New {
			return fmt.Errorf("claim %d (%q) is marked as new material. A recap introduces NOTHING — that is the entire contract, and a viewer who skipped this clip has to be no worse off for it. If the idea genuinely needs teaching, this is the wrong template: `chapter` opens new ground and `analogy` explains it. Drop the claim or replace it with what an earlier lesson actually established",
				i+1, c.Claim)
		}
		if n := len(strings.Fields(c.Claim)); n > maxRecapClaimWords {
			return fmt.Errorf("claim %d is %d words; keep it to %d. It is a reminder, and a reminder that needs a paragraph was never established", i+1, n, maxRecapClaimWords)
		}
		if strings.TrimSpace(c.From) == "" {
			return fmt.Errorf("claim %d (%q) does not say which lesson it came from. That field is the whole point of the template: it is what lets somebody who is lost go to exactly the right place instead of back to the beginning",
				i+1, c.Claim)
		}
		if n := len(strings.Fields(c.From)); n > maxRecapFromWords {
			return fmt.Errorf("the source for claim %d is %d words; keep it to %d — it is a lesson name, not a summary", i+1, n, maxRecapFromWords)
		}
		key := strings.ToLower(c.Claim)
		if seen[key] {
			return fmt.Errorf("claim %d repeats an earlier one (%q)", i+1, c.Claim)
		}
		seen[key] = true
	}

	brought := map[int]bool{}
	threaded := false
	for _, b := range p.Beats {
		if b.Recap == nil {
			return fmt.Errorf("beat %q has no recap direction — say whether it brings back a claim, names the thread, or shows where the viewer is standing", b.ID)
		}
		switch b.Recap.ResolvedShow() {
		case "claim":
			brought[b.Recap.At] = true
		case "thread", "standing":
			threaded = true
		}
	}
	if !threaded {
		return fmt.Errorf("no beat says the thread out loud. Without it the clip is a list read aloud, which is the failure this template exists to avoid")
	}
	for i, c := range r.Claims {
		if !brought[i] {
			return fmt.Errorf("claim %d (%q) is never brought back by a beat", i+1, c.Claim)
		}
	}
	return nil
}

func recapScenes(in SnippetSceneInput) ([]Scene, error) {
	r := in.Plan.Recap
	if r == nil {
		return nil, fmt.Errorf("the plan has no recap")
	}

	claims := make([]map[string]any, len(r.Claims))
	for i, c := range r.Claims {
		claims[i] = map[string]any{"claim": c.Claim, "from": c.From}
	}

	shown := map[int]bool{}
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Recap == nil {
			return nil, fmt.Errorf("beat %q has no recap direction", beat.ID)
		}
		show := beat.Recap.ResolvedShow()
		if show == "claim" {
			shown[beat.Recap.At] = true
		}
		if show == "standing" {
			for j := range r.Claims {
				shown[j] = true
			}
		}
		lit := make([]int, 0, len(shown))
		for at := range shown {
			lit = append(lit, at)
		}
		sort.Ints(lit)

		step := map[string]any{"startMs": startMs, "endMs": endMs, "show": show, "lit": lit}
		if show == "claim" {
			step["at"] = beat.Recap.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneRecap,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"thread": r.Thread,
			"claims": claims,
			"steps":  steps,
		}),
	}}, nil
}
