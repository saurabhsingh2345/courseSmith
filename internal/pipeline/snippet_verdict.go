package pipeline

// The verdict template: the ruling, and the conditions under which it is wrong.
//
// Every one of the reference clips ends this way, and the catalog had no frame
// for it. `compare` introduces two things and marks a winner, which is the
// *argument*; this is what comes after the argument — a recommendation stated
// plainly, the ground it holds on, and the asterisk that keeps it honest.
//
// It is deliberately not `showcase`. A showcase is an introduction: what a tool
// is, what it costs, what you are signing up for, and it runs the first time a
// course meets something. A verdict runs last. It assumes the viewer already
// knows the options and wants to be told what to do, which is a different
// frame, a different beat order, and a different failure mode.
//
// That failure mode is the rule this template exists for. The tempting shape is
// a recommendation with no conditions attached — clean, confident, and useless,
// because the viewer cannot tell whether they are the person it was written
// for. So `breaks` is required and validated: at least one honest statement of
// where the call stops being right. A ruling with no asterisk is not advice, it
// is an advert, and the model will not write the awkward half unless it must.
//
// The closing frame is the call alone, set at headline size on an otherwise
// empty stage. That is the frame people screenshot and the one they quote, so
// it gets the whole screen and nothing competes with it.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "verdict",
		Category:    CatDecisions,
		Since:       SinceV1,
		Title:       "The verdict",
		Description: "A recommendation, the ground it holds on, the asterisk that qualifies it — and the call, alone on the frame.",
		Example:     "After all that: should you actually self-host your database?",
		PromptFile:  snippetVerdictTemplateName,
		NeedsCode:   false,
		// Holds, breaks and the call is three beats minimum, and a verdict that
		// spends one beat on each is a verdict nobody believes. Five is the
		// realistic shape.
		MinTargetSec:     35,
		DefaultTargetSec: 55,
		MaxBeats:         9,
		Owns:             beatFields{Verdict: true},
		OwnsPlan:         planFields{Verdict: true},
		Normalize:        normalizeVerdictPlan,
		Validate:         validateVerdictPlan,
		Scenes:           verdictScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":           strings.Join(VerdictShows(), ", "),
				"MinHolds":        minVerdictHolds,
				"MaxHolds":        maxVerdictHolds,
				"MinBreaks":       minVerdictBreaks,
				"MaxBreaks":       maxVerdictBreaks,
				"MaxCallWords":    maxVerdictCallWords,
				"MaxSubjectWords": maxVerdictSubjectWords,
				"MaxPointWords":   maxVerdictPointWords,
			}
		},
	})
}

const snippetVerdictTemplateName = "snippet_verdict.tmpl"

const (
	minVerdictHolds = 2
	maxVerdictHolds = 4
	// One honest condition is the floor, and it is the whole point.
	minVerdictBreaks = 1
	maxVerdictBreaks = 3

	// The call is set at headline size alone on the frame. Ten words is what
	// fits across the stage at that size on two lines; past it the closing
	// frame becomes a paragraph, which nobody screenshots.
	maxVerdictCallWords    = 10
	maxVerdictSubjectWords = 5
	maxVerdictPointWords   = 9
)

// verdictShows is the closed vocabulary of what a beat does.
var verdictShows = map[string]bool{
	// Name what is being ruled on.
	"subject": true,
	// The ground the call holds on.
	"holds": true,
	// Where it stops being right. Required.
	"breaks": true,
	// The ruling, alone on the frame. The last beat.
	"call": true,
}

// VerdictShows returns the beat vocabulary sorted.
func VerdictShows() []string {
	out := make([]string, 0, len(verdictShows))
	for k := range verdictShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// VerdictSpec is the ruling. On the plan rather than per-beat because the call
// is the subject of the clip and the beats only build toward it.
type VerdictSpec struct {
	// Subject is what is being ruled on — "Self-hosting Postgres".
	Subject string `json:"subject"`
	// Call is the ruling itself, in one line, in the imperative if it can be.
	// It takes the whole closing frame.
	Call string `json:"call"`
	// Holds are the conditions under which the call is right.
	Holds []string `json:"holds"`
	// Breaks are the conditions under which it is not. Required, and the reason
	// this template is worth having.
	Breaks []string `json:"breaks"`
}

// VerdictBeat is one move.
type VerdictBeat struct {
	Show string `json:"show"`
	// At indexes Holds or Breaks, for a beat that lights one of them. A beat
	// with no index lights the whole column.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to the ground
// the call holds on — the bulk of the clip.
func (b VerdictBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if verdictShows[s] {
		return s
	}
	return "holds"
}

func normalizeVerdictPlan(p *SnippetPlan) {
	v := p.Verdict
	if v == nil {
		return
	}
	v.Subject = clampWords(collapseSpaces(v.Subject), maxVerdictSubjectWords)
	v.Call = clampWords(collapseSpaces(v.Call), maxVerdictCallWords)
	v.Holds = clampVerdictPoints(v.Holds, maxVerdictHolds)
	v.Breaks = clampVerdictPoints(v.Breaks, maxVerdictBreaks)

	for i := range p.Beats {
		b := p.Beats[i].Verdict
		if b == nil {
			continue
		}
		if !verdictShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			// The shape says what the ends are for: the clip opens by naming
			// the subject and closes on the call.
			switch {
			case i == 0:
				b.Show = "subject"
			case i == len(p.Beats)-1:
				b.Show = "call"
			default:
				b.Show = "holds"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.At < 0 {
			b.At = 0
		}
		switch b.Show {
		case "holds":
			if n := len(v.Holds); n > 0 && b.At >= n {
				b.At = n - 1
			}
		case "breaks":
			if n := len(v.Breaks); n > 0 && b.At >= n {
				b.At = n - 1
			}
		default:
			b.At = 0
		}
	}
}

// clampVerdictPoints trims a column: each entry to the point-word limit, the
// column to its maximum, and empties dropped.
func clampVerdictPoints(in []string, max int) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = clampWords(collapseSpaces(s), maxVerdictPointWords)
		if s != "" && len(out) < max {
			out = append(out, s)
		}
	}
	return out
}

func validateVerdictPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Verdict: true}); err != nil {
		return err
	}

	v := p.Verdict
	if v == nil {
		return fmt.Errorf("the plan has no verdict — this template is one ruling and the conditions on it")
	}
	if strings.TrimSpace(v.Subject) == "" {
		return fmt.Errorf("the verdict has no subject — say what is being ruled on")
	}
	if strings.TrimSpace(v.Call) == "" {
		return fmt.Errorf("there is no call. The whole clip builds to one line of advice; write the line")
	}
	if n := len(v.Holds); n < minVerdictHolds || n > maxVerdictHolds {
		return fmt.Errorf("the call holds on %d conditions, want %d-%d", n, minVerdictHolds, maxVerdictHolds)
	}
	// The rule this template exists for.
	if n := len(v.Breaks); n < minVerdictBreaks || n > maxVerdictBreaks {
		return fmt.Errorf("there are %d conditions where the call breaks down, want %d-%d. A recommendation with no asterisk is not advice, it is an advert — the viewer cannot tell whether they are the person %q was written for unless you say who it is wrong for",
			n, minVerdictBreaks, maxVerdictBreaks, v.Call)
	}

	lit := map[string]map[int]bool{"holds": {}, "breaks": {}}
	counts := map[string]int{}
	for i, b := range p.Beats {
		if b.Verdict == nil {
			return fmt.Errorf("beat %q has no verdict direction — every beat names the subject, walks a condition, or delivers the call", b.ID)
		}
		show := b.Verdict.ResolvedShow()
		counts[show]++
		if show == "call" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q delivers the call but the clip carries on afterwards. The call is the last frame — it is the one people screenshot, and nothing follows it", b.ID)
		}
		if show != "holds" && show != "breaks" {
			continue
		}
		list := v.Holds
		if show == "breaks" {
			list = v.Breaks
		}
		if b.Verdict.At < 0 || b.Verdict.At >= len(list) {
			return fmt.Errorf("beat %q lights %s %d, which does not exist", b.ID, show, b.Verdict.At)
		}
		if lit[show][b.Verdict.At] {
			return fmt.Errorf("beat %q lights %s %d again; each condition gets one beat", b.ID, show, b.Verdict.At)
		}
		lit[show][b.Verdict.At] = true
	}
	if counts["call"] != 1 {
		return fmt.Errorf("there are %d call beats; the ruling lands exactly once, at the end", counts["call"])
	}
	if p.Beats[len(p.Beats)-1].Verdict.ResolvedShow() != "call" {
		return fmt.Errorf("the clip does not end on the call. Close with {\"show\": \"call\"} — that frame is the whole point of the template")
	}
	// Written but never spoken is the same as not said, which is exactly how a
	// required asterisk gets quietly skipped.
	if len(lit["breaks"]) == 0 {
		return fmt.Errorf("the conditions where the call breaks down are written but no beat says them out loud. An asterisk nobody narrates is an asterisk nobody read — give it a beat")
	}
	if len(lit["holds"]) == 0 {
		return fmt.Errorf("no beat covers the ground the call holds on")
	}
	return nil
}

// verdictScenes lays the clip out as ONE scene: the two columns are the frame
// throughout, and the last beat pushes them back for the call.
func verdictScenes(in SnippetSceneInput) ([]Scene, error) {
	v := in.Plan.Verdict
	if v == nil {
		return nil, fmt.Errorf("the plan has no verdict")
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Verdict == nil {
			return nil, fmt.Errorf("beat %q has no verdict direction", beat.ID)
		}
		show := beat.Verdict.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "holds" || show == "breaks" {
			step["at"] = beat.Verdict.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneVerdict,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":   in.Plan.Title,
			"subject": v.Subject,
			"call":    v.Call,
			"holds":   v.Holds,
			"breaks":  v.Breaks,
			"steps":   steps,
		},
	}}, nil
}
