package pipeline

// The analogy template: the familiar thing, and what each part of it really is.
//
// The strongest of the reference clips hangs its whole argument on one image —
// a librarian in a library — and keeps returning to it: the room is the memory,
// the walk is the bandwidth, the reading is the compute. By the end the viewer
// is reasoning about hardware using a picture they already had, which is what
// an explanation is *for*.
//
// The catalog could not do this. `illustration` can put a metaphor in a
// headline, and a metaphor stated once is a decoration; what teaches is the
// *mapping*, held on screen, part by part, so the viewer can check each
// correspondence as it is claimed.
//
// Two rules earn it its place, and both are validated.
//
// Every part of the metaphor maps to something real. A picture with a piece
// that corresponds to nothing is a picture that will generate a wrong
// prediction the moment the viewer reasons with it — and they will, because
// that is what you have just trained them to do.
//
// And the analogy has to say where it breaks. Every analogy breaks somewhere,
// and the one that never admits it is the one that quietly becomes the
// learner's actual mental model, wrong parts included. This is the same shape
// as the showcase's required limitation and it is required for the same reason:
// the model will not write the awkward half unless it must.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "analogy",
		Category:    CatConcepts,
		Since:       SinceV1,
		Title:       "The mental model",
		Description: "A familiar picture mapped onto the real thing, part by part — and an honest note on where it stops working.",
		Example:     "A database index is a library card catalogue — until it isn't",
		PromptFile:  snippetAnalogyTemplateName,
		NeedsCode:   false,
		// Setting the picture up, mapping three pairs and admitting the break is
		// five beats, and the setup beat cannot be rushed — an analogy the
		// viewer has not pictured yet maps onto nothing.
		MinTargetSec:     45,
		DefaultTargetSec: 65,
		MaxBeats:         8,
		Owns:             beatFields{Analogy: true},
		OwnsPlan:         planFields{Analogy: true},
		Normalize:        normalizeAnalogyPlan,
		Validate:         validateAnalogyPlan,
		Scenes:           analogyScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":           strings.Join(AnalogyShows(), ", "),
				"Icons":           strings.Join(PointIconNames(), ", "),
				"MinPairs":        minAnalogyPairs,
				"MaxPairs":        maxAnalogyPairs,
				"MaxSubjectWords": maxAnalogySubjectWords,
				"MaxSideWords":    maxAnalogySideWords,
				"MaxNoteWords":    maxAnalogyNoteWords,
				"MaxBreaksWords":  maxAnalogyBreaksWords,
			}
		},
	})
}

const snippetAnalogyTemplateName = "snippet_analogy.tmpl"

const (
	// Two correspondences is a simile, not a model. Five rows down the stage
	// leaves each one too short to carry a note.
	minAnalogyPairs = 3
	maxAnalogyPairs = 4

	maxAnalogySubjectWords = 5
	maxAnalogySideWords    = 5
	maxAnalogyNoteWords    = 14
	maxAnalogyBreaksWords  = 18
)

// analogyShows is the closed vocabulary of what a beat does.
var analogyShows = map[string]bool{
	// Set the picture up: the familiar thing, alone, before anything is mapped.
	"picture": true,
	// Light one correspondence.
	"pair": true,
	// Where the analogy stops working. Required.
	"breaks": true,
}

// AnalogyShows returns the beat vocabulary sorted.
func AnalogyShows() []string {
	out := make([]string, 0, len(analogyShows))
	for k := range analogyShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// AnalogySpec is the picture and what it maps onto.
type AnalogySpec struct {
	// Familiar is the thing the viewer already understands — "A library".
	Familiar string `json:"familiar"`
	// FamiliarIcon is a PointIconNames name drawn over the left column.
	FamiliarIcon string `json:"familiarIcon,omitempty"`
	// Real is what is actually being explained — "Running a model locally".
	Real string `json:"real"`
	// RealIcon is a PointIconNames name drawn over the right column.
	RealIcon string `json:"realIcon,omitempty"`
	// Pairs are the correspondences, in the order they are walked.
	Pairs []AnalogyPair `json:"pairs"`
	// Breaks is where the analogy stops working. Required, and the reason this
	// template is worth having: an analogy that never admits its limits becomes
	// the learner's actual mental model, wrong parts included.
	Breaks string `json:"breaks"`
}

// ResolvedFamiliarIcon returns the left column's glyph.
func (a AnalogySpec) ResolvedFamiliarIcon() string {
	if icon := normalizePointIconName(a.FamiliarIcon); icon != "" {
		return icon
	}
	return "book"
}

// ResolvedRealIcon returns the right column's glyph.
func (a AnalogySpec) ResolvedRealIcon() string {
	if icon := normalizePointIconName(a.RealIcon); icon != "" {
		return icon
	}
	return "server"
}

// AnalogyPair is one correspondence: a piece of the picture, and the thing it
// actually is.
type AnalogyPair struct {
	// From is the part of the familiar thing — "The walk to the shelf".
	From string `json:"from"`
	// To is what it really is — "Memory bandwidth".
	To string `json:"to"`
	// Note is the line that makes the correspondence land.
	Note string `json:"note,omitempty"`
}

// AnalogyBeat is one move.
type AnalogyBeat struct {
	Show string `json:"show"`
	At   int    `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to a pair —
// the bulk of the clip.
func (b AnalogyBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if analogyShows[s] {
		return s
	}
	return "pair"
}

func normalizeAnalogyPlan(p *SnippetPlan) {
	a := p.Analogy
	if a == nil {
		return
	}
	a.Familiar = clampWords(collapseSpaces(a.Familiar), maxAnalogySubjectWords)
	a.Real = clampWords(collapseSpaces(a.Real), maxAnalogySubjectWords)
	a.Breaks = clampWords(collapseSpaces(a.Breaks), maxAnalogyBreaksWords)
	a.FamiliarIcon = a.ResolvedFamiliarIcon()
	a.RealIcon = a.ResolvedRealIcon()

	pairs := make([]AnalogyPair, 0, len(a.Pairs))
	for _, pr := range a.Pairs {
		pr.From = clampWords(collapseSpaces(pr.From), maxAnalogySideWords)
		pr.To = clampWords(collapseSpaces(pr.To), maxAnalogySideWords)
		pr.Note = clampWords(collapseSpaces(pr.Note), maxAnalogyNoteWords)
		// Half a correspondence is a piece of the picture pointing at nothing,
		// which is exactly the failure the validator exists to catch. Dropping
		// it is the repair; inventing the other half would be a claim about the
		// subject.
		if pr.From != "" && pr.To != "" && len(pairs) < maxAnalogyPairs {
			pairs = append(pairs, pr)
		}
	}
	a.Pairs = pairs

	for i := range p.Beats {
		b := p.Beats[i].Analogy
		if b == nil {
			continue
		}
		if !analogyShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			switch {
			case i == 0:
				b.Show = "picture"
			case i == len(p.Beats)-1:
				b.Show = "breaks"
			default:
				b.Show = "pair"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "pair" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(a.Pairs); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateAnalogyPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Analogy: true}); err != nil {
		return err
	}

	a := p.Analogy
	if a == nil {
		return fmt.Errorf("the plan has no analogy — this template is one familiar picture mapped onto the real thing")
	}
	if strings.TrimSpace(a.Familiar) == "" {
		return fmt.Errorf("there is no familiar thing. The whole template rests on a picture the viewer already has; name it")
	}
	if strings.TrimSpace(a.Real) == "" {
		return fmt.Errorf("there is nothing being explained — say what the picture is standing in for")
	}
	if n := len(a.Pairs); n < minAnalogyPairs || n > maxAnalogyPairs {
		return fmt.Errorf("there are %d correspondences, want %d-%d. Two is a simile rather than a model, and five rows down the stage leaves each too short to carry its note",
			n, minAnalogyPairs, maxAnalogyPairs)
	}

	// The first rule: nothing in the picture may point at nothing.
	seenFrom, seenTo := map[string]bool{}, map[string]bool{}
	for i, pr := range a.Pairs {
		if strings.TrimSpace(pr.From) == "" || strings.TrimSpace(pr.To) == "" {
			return fmt.Errorf("correspondence %d is half-written. A piece of the picture that maps to nothing will generate a wrong prediction the moment the viewer reasons with it — and reasoning with it is exactly what you have just taught them to do", i)
		}
		kf, kt := strings.ToLower(pr.From), strings.ToLower(pr.To)
		if seenFrom[kf] {
			return fmt.Errorf("two correspondences both start from %q — each part of the picture maps to one thing", pr.From)
		}
		if seenTo[kt] {
			return fmt.Errorf("two correspondences both land on %q — mapping two parts of the picture onto one real thing is where an analogy starts lying", pr.To)
		}
		seenFrom[kf], seenTo[kt] = true, true
	}

	// The second rule, and the one this template exists for.
	if strings.TrimSpace(a.Breaks) == "" {
		return fmt.Errorf("the analogy never says where it stops working. Every analogy breaks somewhere, and the one that does not admit it becomes the learner's actual mental model with the wrong parts still in it — say what %q does NOT do that %q does",
			a.Familiar, a.Real)
	}

	walked := map[int]bool{}
	counts := map[string]int{}
	for i, b := range p.Beats {
		if b.Analogy == nil {
			return fmt.Errorf("beat %q has no analogy direction — every beat sets the picture up, walks one correspondence, or admits the break", b.ID)
		}
		show := b.Analogy.ResolvedShow()
		counts[show]++
		if i == 0 && show != "picture" {
			return fmt.Errorf("the clip opens on %q. Set the picture up first — a correspondence drawn before the viewer has pictured the familiar thing maps onto nothing", show)
		}
		if show != "pair" {
			continue
		}
		if b.Analogy.At < 0 || b.Analogy.At >= len(a.Pairs) {
			return fmt.Errorf("beat %q walks correspondence %d, which does not exist", b.ID, b.Analogy.At)
		}
		if walked[b.Analogy.At] {
			return fmt.Errorf("beat %q walks correspondence %d again; each one gets a beat", b.ID, b.Analogy.At)
		}
		walked[b.Analogy.At] = true
	}
	if counts["picture"] != 1 {
		return fmt.Errorf("there are %d picture beats; the image is set up once", counts["picture"])
	}
	if len(walked) != len(a.Pairs) {
		return fmt.Errorf("%d of the %d correspondences are never spoken. A mapping drawn but not narrated is one the viewer has to work out alone, which is what the analogy was supposed to save them",
			len(a.Pairs)-len(walked), len(a.Pairs))
	}
	// Written but never said is the same as not said — the same way the
	// showcase's limitation can be quietly skipped.
	if counts["breaks"] == 0 {
		return fmt.Errorf("the analogy says where it breaks but no beat speaks it. That admission is what stops the picture becoming the learner's real model — give it a beat, and make it the last one")
	}
	return nil
}

// analogyScenes lays the clip out as ONE scene: both columns are on screen
// throughout and the beats only move which row is lit.
func analogyScenes(in SnippetSceneInput) ([]Scene, error) {
	a := in.Plan.Analogy
	if a == nil {
		return nil, fmt.Errorf("the plan has no analogy")
	}

	pairs := make([]map[string]any, len(a.Pairs))
	for i, pr := range a.Pairs {
		pairs[i] = map[string]any{"from": pr.From, "to": pr.To, "note": pr.Note}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Analogy == nil {
			return nil, fmt.Errorf("beat %q has no analogy direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Analogy.ResolvedShow(),
		}
		if beat.Analogy.ResolvedShow() == "pair" {
			step["at"] = beat.Analogy.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneAnalogy,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":        in.Plan.Title,
			"familiar":     a.Familiar,
			"familiarIcon": a.ResolvedFamiliarIcon(),
			"real":         a.Real,
			"realIcon":     a.ResolvedRealIcon(),
			"pairs":        pairs,
			"breaks":       a.Breaks,
			"steps":        steps,
		},
	}}, nil
}
