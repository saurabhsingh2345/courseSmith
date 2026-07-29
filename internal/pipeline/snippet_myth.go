package pipeline

// The myth template: what everyone believes, struck out, and what is true.
//
// Half of teaching is not adding knowledge but replacing it. A learner who
// thinks Redis is "just a cache" or that WebAssembly is "replacing JavaScript"
// does not have a gap to fill — they have a wrong model that will keep
// producing wrong predictions until something visibly removes it. Adding a
// correct explanation beside a wrong belief usually loses; the wrong one was
// there first.
//
// So the frame does the removing. The claim is set large, in the viewer's own
// words, and then struck through and replaced. That gesture is the teaching,
// and it is why this cannot be `illustration` with a clever headline: a
// strikethrough is a *retraction*, and no other template in the catalog can
// draw one.
//
// Two rules earn it its place.
//
// The claim must be stated in the words somebody actually uses. A strawman —
// "some people think Redis is bad" — corrects nothing, because no viewer
// recognises themselves in it and so nobody's model gets replaced. The
// validator cannot check for honesty, but the prompt can insist, and the
// length cap keeps it to something quotable rather than a hedge.
//
// The correction must not merely negate the claim. "Redis is not just a cache"
// leaves the viewer with a hole where their model was, which is worse than the
// wrong model — at least the wrong one made predictions. This one IS checked:
// a truth that is the claim with a negation bolted on is rejected.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "myth",
		Category:    CatConcepts,
		Since:       SinceV1,
		Title:       "What everyone gets wrong",
		Description: "The belief, in the words people use — struck through, replaced, and then shown why it was tempting.",
		Example:     "Redis is just a cache — and why that costs teams real money",
		PromptFile:  snippetMythTemplateName,
		NeedsCode:   false,
		// Claim, strike, truth and at least one piece of evidence is four beats.
		MinTargetSec:     35,
		DefaultTargetSec: 50,
		MaxBeats:         8,
		Owns:             beatFields{Myth: true},
		OwnsPlan:         planFields{Myth: true},
		Normalize:        normalizeMythPlan,
		Validate:         validateMythPlan,
		Scenes:           mythScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":            strings.Join(MythShows(), ", "),
				"MinEvidence":      minMythEvidence,
				"MaxEvidence":      maxMythEvidence,
				"MaxClaimWords":    maxMythClaimWords,
				"MaxTruthWords":    maxMythTruthWords,
				"MaxWhyWords":      maxMythWhyWords,
				"MaxEvidenceWords": maxMythEvidenceWords,
			}
		},
	})
}

const snippetMythTemplateName = "snippet_myth.tmpl"

const (
	minMythEvidence = 1
	maxMythEvidence = 4

	// The claim and the truth are both set at headline size, one above the
	// other, so they have to be roughly quotable lengths.
	maxMythClaimWords = 9
	maxMythTruthWords = 12
	// Why the belief was tempting — the beat that keeps the clip from being
	// smug.
	maxMythWhyWords      = 18
	maxMythEvidenceWords = 10
)

// mythShows is the closed vocabulary of what a beat does.
var mythShows = map[string]bool{
	// The claim alone, stated straight, before anything contradicts it.
	"claim": true,
	// The strike: the claim is crossed out and the truth lands under it.
	"strike": true,
	// One piece of what is actually the case.
	"evidence": true,
	// Why the belief was reasonable. Not optional in spirit; see the prompt.
	"why": true,
}

// MythShows returns the beat vocabulary sorted.
func MythShows() []string {
	out := make([]string, 0, len(mythShows))
	for k := range mythShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// MythSpec is the belief being replaced.
type MythSpec struct {
	// Claim is the wrong belief, in the words somebody actually uses. Set large
	// and then struck through.
	Claim string `json:"claim"`
	// Truth is what is actually the case. It replaces the claim on screen, so
	// it must stand on its own rather than negate.
	Truth string `json:"truth"`
	// Why is why the belief was reasonable to hold. The beat that keeps the
	// clip from being smug, and the one that makes it persuasive.
	Why string `json:"why,omitempty"`
	// Evidence is what makes the truth stick — the concrete things.
	Evidence []string `json:"evidence"`
}

// MythBeat is one move.
type MythBeat struct {
	Show string `json:"show"`
	At   int    `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to evidence —
// the bulk of the clip after the strike.
func (b MythBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if mythShows[s] {
		return s
	}
	return "evidence"
}

// negationPrefixes are the shapes a correction takes when it is only a denial.
var negationPrefixes = []string{
	"not ", "no, ", "it is not ", "it's not ", "that is not ", "that's not ",
	"actually, no", "wrong", "false", "nope",
}

// isMereNegation reports whether truth says nothing except that claim is false.
//
// The test is deliberately shallow — it catches the shape, not the semantics —
// because the failure it guards against has a very consistent shape: the model
// echoes the claim with a negation attached and calls it a correction. A truth
// that survives this may still be weak, but it will at least be a *statement*,
// which is the thing a viewer can put where their old model was.
func isMereNegation(claim, truth string) bool {
	c := normalizeForNegation(claim)
	t := normalizeForNegation(truth)
	if t == "" {
		return false
	}
	for _, p := range negationPrefixes {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	// "Redis is just a cache" -> "Redis is not just a cache". Turn the negated
	// verb back into its positive form; if what is left is the claim again,
	// nothing has been asserted.
	positive := t
	for _, sub := range negationSubstitutions {
		positive = strings.ReplaceAll(positive, sub[0], sub[1])
	}
	return positive != t && positive == c
}

// negationSubstitutions turn a negated verb back into its positive form. The
// surrounding spaces are part of both sides so the words stay separated — a
// substitution that collapses " is not " to "is" welds the neighbouring words
// together and the comparison silently never matches.
var negationSubstitutions = [][2]string{
	{" is not ", " is "}, {" isn't ", " is "},
	{" are not ", " are "}, {" aren't ", " are "},
	{" was not ", " was "}, {" wasn't ", " was "},
	{" does not ", " does "}, {" doesn't ", " does "},
	{" do not ", " do "}, {" don't ", " do "},
	{" cannot ", " can "}, {" can't ", " can "},
	{" will not ", " will "}, {" won't ", " will "},
}

// normalizeForNegation lowercases, collapses whitespace and drops the trailing
// punctuation that would otherwise make two identical sentences compare unequal.
func normalizeForNegation(s string) string {
	out := collapseSpaces(strings.ToLower(strings.TrimSpace(s)))
	return strings.TrimRight(out, ".!?")
}

func normalizeMythPlan(p *SnippetPlan) {
	m := p.Myth
	if m == nil {
		return
	}
	m.Claim = clampWords(collapseSpaces(m.Claim), maxMythClaimWords)
	m.Truth = clampWords(collapseSpaces(m.Truth), maxMythTruthWords)
	m.Why = clampWords(collapseSpaces(m.Why), maxMythWhyWords)

	ev := make([]string, 0, len(m.Evidence))
	for _, e := range m.Evidence {
		e = clampWords(collapseSpaces(e), maxMythEvidenceWords)
		if e != "" && len(ev) < maxMythEvidence {
			ev = append(ev, e)
		}
	}
	m.Evidence = ev

	for i := range p.Beats {
		b := p.Beats[i].Myth
		if b == nil {
			continue
		}
		if !mythShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			switch i {
			case 0:
				b.Show = "claim"
			case 1:
				b.Show = "strike"
			default:
				b.Show = "evidence"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "evidence" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(m.Evidence); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateMythPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Myth: true}); err != nil {
		return err
	}

	m := p.Myth
	if m == nil {
		return fmt.Errorf("the plan has no myth — this template is one wrong belief, removed, and what goes in its place")
	}
	if strings.TrimSpace(m.Claim) == "" {
		return fmt.Errorf("there is no claim. Write the belief in the words somebody actually uses — a strawman corrects nothing, because no viewer recognises themselves in it")
	}
	if strings.TrimSpace(m.Truth) == "" {
		return fmt.Errorf("there is nothing to put where the belief was")
	}
	// The rule this template exists for.
	if isMereNegation(m.Claim, m.Truth) {
		return fmt.Errorf("the correction %q only denies the claim. A viewer left with a hole where their model used to be is worse off than one with a wrong model, because at least the wrong one made predictions — say what IS the case, as a statement that stands on its own without the claim next to it",
			m.Truth)
	}
	if n := len(m.Evidence); n < minMythEvidence || n > maxMythEvidence {
		return fmt.Errorf("there are %d pieces of evidence, want %d-%d. A correction nobody backs up is just a different assertion, and the viewer has no reason to prefer yours",
			n, minMythEvidence, maxMythEvidence)
	}
	for i, e := range m.Evidence {
		if strings.TrimSpace(e) == "" {
			return fmt.Errorf("evidence %d is blank", i)
		}
	}

	shown := map[int]bool{}
	counts := map[string]int{}
	strikeAt := -1
	for i, b := range p.Beats {
		if b.Myth == nil {
			return fmt.Errorf("beat %q has no myth direction — every beat states the claim, strikes it, backs up the truth, or says why the belief was tempting", b.ID)
		}
		show := b.Myth.ResolvedShow()
		counts[show]++
		if i == 0 && show != "claim" {
			return fmt.Errorf("the clip opens on %q. State the belief first and state it straight — a correction that arrives before the viewer has recognised their own thought corrects nobody", show)
		}
		if show == "strike" {
			strikeAt = i
		}
		if show == "evidence" {
			if strikeAt < 0 {
				return fmt.Errorf("beat %q backs up the truth before the belief has been struck. Nothing can be evidence for a claim the clip has not made yet", b.ID)
			}
			if b.Myth.At < 0 || b.Myth.At >= len(m.Evidence) {
				return fmt.Errorf("beat %q shows evidence %d, which does not exist", b.ID, b.Myth.At)
			}
			if shown[b.Myth.At] {
				return fmt.Errorf("beat %q shows evidence %d again; each one gets a beat", b.ID, b.Myth.At)
			}
			shown[b.Myth.At] = true
		}
	}
	if counts["claim"] != 1 {
		return fmt.Errorf("there are %d claim beats; the belief is stated once", counts["claim"])
	}
	if counts["strike"] != 1 {
		return fmt.Errorf("there are %d strike beats; the belief is replaced exactly once, and that gesture is the whole template", counts["strike"])
	}
	if len(shown) != len(m.Evidence) {
		return fmt.Errorf("%d of the %d pieces of evidence are never spoken — give them a beat or cut them", len(m.Evidence)-len(shown), len(m.Evidence))
	}
	// Why the belief was tempting keeps the clip from being smug, and a smug
	// correction is one the viewer argues with rather than accepts.
	if strings.TrimSpace(m.Why) != "" && counts["why"] == 0 {
		return fmt.Errorf("the clip explains why the belief was reasonable but no beat says it. That line is what stops the correction reading as smug — give it a beat")
	}
	return nil
}

// mythScenes lays the clip out as ONE scene: the claim is on screen from the
// first frame and the strike happens in place.
func mythScenes(in SnippetSceneInput) ([]Scene, error) {
	m := in.Plan.Myth
	if m == nil {
		return nil, fmt.Errorf("the plan has no myth")
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Myth == nil {
			return nil, fmt.Errorf("beat %q has no myth direction", beat.ID)
		}
		show := beat.Myth.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "evidence" {
			step["at"] = beat.Myth.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneMyth,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":    in.Plan.Title,
			"claim":    m.Claim,
			"truth":    m.Truth,
			"why":      m.Why,
			"evidence": m.Evidence,
			"steps":    steps,
		},
	}}, nil
}
