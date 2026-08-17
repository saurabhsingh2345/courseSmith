package pipeline

// The pitfall template: the mistake you are about to make, and how it will look.
//
// The catalog already has `myth`, and the difference between them is the reason
// this one exists. A myth is a wrong BELIEF — something people think is true,
// corrected by saying what is true instead. A pitfall is a wrong ACTION at a
// specific step, and correcting a belief does not help somebody who has already
// taken it. They do not need to be told they were wrong; they need to recognise
// the thing they are currently looking at.
//
// Three rules earn it its place.
//
// **There is a symptom, and it is observable.** This is the field that separates
// the template from `myth` and the one a model omits every time, because
// describing the mistake feels like the work. It is not: the viewer already made
// the mistake and does not know it. What they have is a screen with something
// wrong on it, and the only thing that connects their screen to this clip is a
// symptom they can match against — an error string, a number that is too high, a
// thing that works locally and not deployed.
//
// **The symptom is not a restatement of the mistake.** "The mistake is retrying
// forever; the symptom is that it retries forever" has drawn a circle. The
// symptom must be what you SEE, not what is happening — the sentence a validator
// cannot fully check, so the check is that the two strings differ substantially
// and the prompt carries the rest.
//
// **There is a fix, and it is one move.** A pitfall clip that ends on "it
// depends" has described a trap and left the viewer in it.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:             "pitfall",
		Category:         CatDecisions,
		Since:            SinceV5,
		Title:            "How this goes wrong",
		Description:      "The mistake people actually make at one step, the symptom that shows it happened, and the single move that fixes it. Reach for it right after teaching a step people get wrong.",
		Example:          "The mistake everyone makes the first time they add retries to an API client",
		PromptFile:       snippetPitfallTemplateName,
		MinTargetSec:     25,
		DefaultTargetSec: 45,
		MaxBeats:         8,
		Owns:             beatFields{Pitfall: true},
		OwnsPlan:         planFields{Pitfall: true},
		Normalize:        normalizePitfallPlan,
		Validate:         validatePitfallPlan,
		Scenes:           pitfallScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				// The emphasis roles every headline picks from.
				"Roles":            strings.Join(MetricRoles(), ", "),
				"Shows":            strings.Join(PitfallShows(), ", "),
				"MaxMistakeWords":  maxPitfallMistakeWords,
				"MaxSymptomWords":  maxPitfallSymptomWords,
				"MaxFixWords":      maxPitfallFixWords,
				"MaxWhyWords":      maxPitfallWhyWords,
				"MinSymptomShared": minPitfallSymptomDistinct,
			}
		},
	})
}

const snippetPitfallTemplateName = "snippet_pitfall.tmpl"

const (
	maxPitfallMistakeWords = 14
	maxPitfallSymptomWords = 16
	maxPitfallFixWords     = 14
	maxPitfallWhyWords     = 20

	// How many words the symptom must have that the mistake does not. Two is
	// low on purpose: the check is for a restatement, not for prose variety, and
	// a rule that fires on a genuinely tight pair costs a correction round.
	minPitfallSymptomDistinct = 2
)

// PitfallSpec is one trap and the way out of it.
type PitfallSpec struct {
	// Mistake is the wrong ACTION, at the step where people take it.
	Mistake string `json:"mistake"`
	// Symptom is what you SEE when it has happened.
	Symptom string `json:"symptom"`
	// Why is what makes the mistake tempting. Not blame — the mistake is
	// usually the locally sensible move, and saying so is what stops the clip
	// sounding like a telling-off.
	Why string `json:"why,omitempty"`
	// Fix is the single move that gets out of it.
	Fix string `json:"fix"`
}

// PitfallBeat says which part of the trap this beat is on.
type PitfallBeat struct {
	Show string `json:"show"`
}

var pitfallShows = map[string]bool{
	// The step, taken the way people take it.
	"mistake": true,
	// What shows up on screen when it has been taken.
	"symptom": true,
	// Why it was the sensible-looking move.
	"why": true,
	// The one move out.
	"fix": true,
}

// PitfallShows returns the beat vocabulary, sorted, for the prompt.
func PitfallShows() []string {
	out := make([]string, 0, len(pitfallShows))
	for s := range pitfallShows {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// ResolvedShow defaults to the symptom, which is the beat most clips of this
// template spend their time on.
func (b PitfallBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if pitfallShows[s] {
		return s
	}
	return "symptom"
}

func normalizePitfallPlan(p *SnippetPlan) {
	if p.Pitfall == nil {
		return
	}
	f := p.Pitfall
	f.Mistake = strings.TrimSpace(f.Mistake)
	f.Symptom = strings.TrimSpace(f.Symptom)
	f.Why = strings.TrimSpace(f.Why)
	f.Fix = strings.TrimSpace(f.Fix)
	for i := range p.Beats {
		if b := p.Beats[i].Pitfall; b != nil {
			b.Show = b.ResolvedShow()
		}
	}
}

// wordSet lowercases and splits, for the restatement check.
func wordSet(s string) map[string]bool {
	out := map[string]bool{}
	for _, w := range strings.Fields(strings.ToLower(s)) {
		out[strings.Trim(w, ".,:;!?\"'()")] = true
	}
	return out
}

func validatePitfallPlan(p *SnippetPlan) error {
	f := p.Pitfall
	if f == nil {
		return fmt.Errorf("the plan has no pitfall — this template shows one mistake, so the trap is the clip")
	}
	if strings.TrimSpace(f.Mistake) == "" {
		return fmt.Errorf("the pitfall names no mistake")
	}
	if n := len(strings.Fields(f.Mistake)); n > maxPitfallMistakeWords {
		return fmt.Errorf("the mistake is %d words; keep it to %d", n, maxPitfallMistakeWords)
	}
	if strings.TrimSpace(f.Symptom) == "" {
		return fmt.Errorf("the pitfall names no symptom, and the symptom is what this template is FOR. The viewer has already made the mistake and does not know it — what they have is a screen with something wrong on it. Name what they would SEE: the error string, the number that is too high, the thing that works locally and not deployed. Without it this is a `myth` clip, which corrects a belief rather than a step")
	}
	if n := len(strings.Fields(f.Symptom)); n > maxPitfallSymptomWords {
		return fmt.Errorf("the symptom is %d words; keep it to %d", n, maxPitfallSymptomWords)
	}
	if strings.TrimSpace(f.Fix) == "" {
		return fmt.Errorf("the pitfall names no fix. A clip that describes a trap and stops has left the viewer in it")
	}
	if n := len(strings.Fields(f.Fix)); n > maxPitfallFixWords {
		return fmt.Errorf("the fix is %d words; keep it to %d — it is one move, not a procedure", n, maxPitfallFixWords)
	}
	if n := len(strings.Fields(f.Why)); n > maxPitfallWhyWords {
		return fmt.Errorf("the reason is %d words; keep it to %d", n, maxPitfallWhyWords)
	}

	// The restatement check. A symptom built out of the mistake's own words is
	// the failure this template's whole value rests on avoiding.
	mistake := wordSet(f.Mistake)
	distinct := 0
	for w := range wordSet(f.Symptom) {
		if !mistake[w] && len(w) > 3 {
			distinct++
		}
	}
	if distinct < minPitfallSymptomDistinct {
		return fmt.Errorf("the symptom (%q) is the mistake (%q) said again. A symptom is what you SEE, not what is happening — %q is a description of the fault; the symptom is the log line, the latency number, the blank screen. Rewrite it as the thing the viewer would notice before they knew what was wrong",
			f.Symptom, f.Mistake, f.Mistake)
	}

	shown := map[string]bool{}
	for _, b := range p.Beats {
		if b.Pitfall == nil {
			return fmt.Errorf("beat %q has no pitfall direction — say whether it shows the mistake, the symptom, why it was tempting, or the fix", b.ID)
		}
		shown[b.Pitfall.ResolvedShow()] = true
	}
	for _, need := range []string{"mistake", "symptom", "fix"} {
		if !shown[need] {
			return fmt.Errorf("no beat shows the %s. All three of mistake, symptom and fix have to be spoken — a clip missing the symptom cannot be matched against the viewer's screen, and one missing the fix leaves them stuck", need)
		}
	}
	return nil
}

func pitfallScenes(in SnippetSceneInput) ([]Scene, error) {
	f := in.Plan.Pitfall
	if f == nil {
		return nil, fmt.Errorf("the plan has no pitfall")
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Pitfall == nil {
			return nil, fmt.Errorf("beat %q has no pitfall direction", beat.ID)
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Pitfall.ResolvedShow(),
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    ScenePitfall,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":   in.Plan.Title,
			"mistake": f.Mistake,
			"symptom": f.Symptom,
			"why":     f.Why,
			"fix":     f.Fix,
			"steps":   steps,
		}),
	}}, nil
}
