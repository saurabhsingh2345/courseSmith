package pipeline

// The ratio template: two quantities, and the fraction between them.
//
// The boundary with `compare` is numbers. `compare` puts two subjects in columns,
// fills them with code or a figure, and lands a verdict — it is qualitative, and
// it has nothing to divide. This template's entire content is one piece of
// arithmetic: two measurements in the same unit, and the proportion the viewer
// would otherwise have to work out. "Two hundred and seventy against eight
// hundred" is two numbers; "a third" is an argument, and it is the form the
// reference clips reach for when they want a comparison to be memorable rather
// than merely accurate.
//
// It is not `latency` either, and the difference is what the picture claims.
// Latency says "these are in different categories" over a log axis with several
// rows. This says "this one is exactly this fraction of that one", which is a
// precise proportional claim about exactly two things.
//
// Two rules earn it its place, and both are validators.
//
// **The stated fraction must be the real one.** Same rule as `multiply` and for
// the same reason: the arithmetic IS the content, so a clip that calls 270 out of
// 800 "half" has spent its one memorable line on something false. Go divides and
// checks. Rounding to a friendly fraction is expected — "a third" for 0.3375 is
// what a person says — so the check is a tolerance.
//
// **The two quantities must be at least twice apart.** A fraction is only worth
// stating when it is striking. "Nine tenths of" is not a headline, it is a
// rounding error with a preposition, and a clip built on it should be a chart —
// `data` draws that, or `compare` if the point is qualitative.

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:             "ratio",
		Category:         CatNumbers,
		Since:            SinceV4,
		Family:           FamilyReplica,
		Title:            "A fraction of the other",
		Description:      "Two measurements in one unit and the proportion between them, stated as the line the viewer remembers. Reach for it when \"a third of\" lands harder than the two numbers do.",
		Example:          "How the DGX Spark's memory bandwidth compares to a Mac Studio's",
		PromptFile:       snippetRatioTemplateName,
		NeedsCode:        false,
		MinTargetSec:     20,
		DefaultTargetSec: 40,
		MaxBeats:         7,
		Owns:             beatFields{Ratio: true},
		OwnsPlan:         planFields{Ratio: true},
		Normalize:        normalizeRatioPlan,
		Validate:         validateRatioPlan,
		Scenes:           ratioScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":          strings.Join(MetricRoles(), ", "),
				"Shows":          strings.Join(RatioShows(), ", "),
				"MinSpread":      minRatioSpread,
				"MaxLabelWords":  maxRatioLabelWords,
				"MaxNoteWords":   maxRatioNoteWords,
				"MaxUnitWords":   maxRatioUnitWords,
				"MaxPhraseWords": maxRatioPhraseWords,
			}
		},
	})
}

const snippetRatioTemplateName = "snippet_ratio.tmpl"

const (
	// The smaller quantity must be at most half the larger, or the fraction is
	// not worth a clip. See the file header.
	minRatioSpread = 2.0

	maxRatioLabelWords  = 5
	maxRatioNoteWords   = 16
	maxRatioUnitWords   = 3
	maxRatioPhraseWords = 4

	// The arithmetic tolerance. Fifteen percent, which is much looser than
	// `multiply`'s half a percent, and deliberately: a friendly fraction is the
	// whole point of the template. "A third" for 0.3375 is what a person says and
	// what a viewer remembers, and rejecting it would force the clip to say
	// "thirty-three point seven five percent of", which nobody does.
	ratioTolerance = 0.15
)

// ratioShows is the closed vocabulary of what a beat does.
var ratioShows = map[string]bool{
	// The larger quantity, alone. The first beat, always: the fraction is OF it.
	"reference": true,
	// The smaller quantity arrives beside it.
	"subject": true,
	// The fraction between them lands.
	"fraction": true,
	// What that proportion means in practice.
	"read": true,
}

// RatioShows returns the beat vocabulary sorted.
func RatioShows() []string {
	out := make([]string, 0, len(ratioShows))
	for k := range ratioShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// RatioSpec is the pair and the proportion. On the plan because the clip builds
// one statement rather than a per-beat visual.
type RatioSpec struct {
	// Unit is what BOTH quantities are measured in. One unit: a proportion
	// across two different units is a rate, not a ratio, and it would not mean
	// what the clip says it means.
	Unit string `json:"unit"`
	// Reference is the thing being measured against — the larger one.
	Reference RatioSide `json:"reference"`
	// Subject is the thing whose size is the point — the smaller one.
	Subject RatioSide `json:"subject"`
	// Phrase is the proportion in words, and it is the line the whole clip is
	// for — "a third of", "half", "a tenth of what you expected".
	Phrase string `json:"phrase"`
	// Note is what the proportion means in practice. One sentence.
	Note string `json:"note,omitempty"`
}

// RatioSide is one of the two quantities.
type RatioSide struct {
	// Label names it.
	Label string `json:"label"`
	// Value is the measurement, in the spec's unit.
	Value float64 `json:"value"`
	// Role is what this side is doing: a metricRoles name.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the side's role, defaulting to neutral.
func (s RatioSide) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(s.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// Fraction is the subject as a proportion of the reference.
func (s *RatioSpec) Fraction() float64 {
	if s.Reference.Value == 0 {
		return 0
	}
	return s.Subject.Value / s.Reference.Value
}

// Spread is how many times larger the reference is than the subject.
func (s *RatioSpec) Spread() float64 {
	if s.Subject.Value == 0 {
		return math.Inf(1)
	}
	return s.Reference.Value / s.Subject.Value
}

// ratioPhraseValue reads a spoken fraction back as a number, so the arithmetic
// can be checked against what the clip actually says.
//
// Only the fractions people say out loud are recognised. That is not a
// limitation to work around — a phrase this cannot read is a phrase a viewer
// cannot convert either, and the validator says so rather than guessing.
var ratioPhrases = map[string]float64{
	"half":           1.0 / 2,
	"a half":         1.0 / 2,
	"a third":        1.0 / 3,
	"a quarter":      1.0 / 4,
	"a fifth":        1.0 / 5,
	"a sixth":        1.0 / 6,
	"an eighth":      1.0 / 8,
	"a tenth":        1.0 / 10,
	"a twentieth":    1.0 / 20,
	"a fiftieth":     1.0 / 50,
	"a hundredth":    1.0 / 100,
	"two thirds":     2.0 / 3,
	"three quarters": 3.0 / 4,
}

// ratioPhraseFraction extracts the fraction a phrase claims, and whether it
// claimed one at all. "a third of" and "roughly a third" both read as 1/3.
func ratioPhraseFraction(phrase string) (float64, bool) {
	key := phraseKey(phrase)
	// Strip the words that qualify a fraction without changing it.
	for _, filler := range []string{"roughly ", "about ", "around ", "nearly ", "barely ", "just ", "over ", "under "} {
		key = strings.TrimPrefix(key, phraseKey(filler)+" ")
	}
	key = strings.TrimSuffix(key, " of")
	key = strings.TrimSuffix(key, " of the")
	if f, ok := ratioPhrases[key]; ok {
		return f, true
	}
	return 0, false
}

// RatioBeat is one move.
type RatioBeat struct {
	// Show is a ratioShows name.
	Show string `json:"show"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to the
// reference — which is where every clip of this shape starts.
func (b RatioBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if ratioShows[s] {
		return s
	}
	return "reference"
}

func normalizeRatioPlan(p *SnippetPlan) {
	r := p.Ratio
	if r == nil {
		return
	}
	r.Unit = clampWords(collapseSpaces(r.Unit), maxRatioUnitWords)
	r.Phrase = clampWords(collapseSpaces(r.Phrase), maxRatioPhraseWords)
	r.Note = clampWords(collapseSpaces(r.Note), maxRatioNoteWords)
	r.Reference.Label = clampWords(collapseSpaces(r.Reference.Label), maxRatioLabelWords)
	r.Subject.Label = clampWords(collapseSpaces(r.Subject.Label), maxRatioLabelWords)
	r.Reference.Role = r.Reference.ResolvedRole()
	r.Subject.Role = r.Subject.ResolvedRole()

	// The reference is the larger one by definition, and which way round the
	// model happened to fill the fields is not a claim about the subject — the
	// values are. Swapping is a mechanical repair, and it is the difference
	// between a clip that says "a third of" and one that says "three times".
	if r.Subject.Value > r.Reference.Value {
		r.Reference, r.Subject = r.Subject, r.Reference
	}

	for i := range p.Beats {
		b := p.Beats[i].Ratio
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
	}
}

func validateRatioPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Ratio: true}); err != nil {
		return err
	}

	r := p.Ratio
	if r == nil {
		return fmt.Errorf("the plan has no pair — this template is two measurements and the fraction between them")
	}
	if strings.TrimSpace(r.Unit) == "" {
		return fmt.Errorf("the pair has no unit. BOTH quantities are measured in the same one: a proportion across two different units is a rate rather than a ratio, and it would not mean what the clip says it means")
	}
	for name, side := range map[string]RatioSide{"reference": r.Reference, "subject": r.Subject} {
		if strings.TrimSpace(side.Label) == "" {
			return fmt.Errorf("the %s has no label — a number with nothing attached to it cannot be a side of a comparison", name)
		}
		if side.Value <= 0 {
			return fmt.Errorf("the %s (%q) measures %v; both sides need a positive quantity to have a proportion between them", name, side.Label, side.Value)
		}
		if role := strings.ToLower(strings.TrimSpace(side.Role)); role != "" && !metricRoles[role] {
			return fmt.Errorf("the %s has role %q, which is not one of: %s", name, side.Role, strings.Join(MetricRoles(), ", "))
		}
	}

	// A fraction is only worth stating when it is striking.
	if spread := r.Spread(); spread < minRatioSpread {
		return fmt.Errorf("%v against %v is a spread of %.2gx, and a fraction that close is a rounding error with a preposition rather than a line anybody repeats. This template wants at least %gx — use the data template for a chart, or compare if the real point is qualitative",
			r.Subject.Value, r.Reference.Value, spread, minRatioSpread)
	}

	// The rule the template exists for: the spoken fraction has to be the real
	// one, because it is the line the whole clip is built to be remembered by.
	if strings.TrimSpace(r.Phrase) == "" {
		return fmt.Errorf("there is no phrase. The proportion in words — \"a third of\", \"half\" — is the line this whole template exists to land, and two numbers without it is a chart")
	}
	claimed, readable := ratioPhraseFraction(r.Phrase)
	if !readable {
		return fmt.Errorf("the phrase %q is not a fraction anybody says out loud, so it cannot be checked against the numbers — and a viewer cannot convert it either. Use one of: half, a third, a quarter, a fifth, a sixth, an eighth, a tenth, a twentieth, a fiftieth, a hundredth, two thirds, three quarters",
			r.Phrase)
	}
	actual := r.Fraction()
	if math.Abs(claimed-actual)/claimed > ratioTolerance {
		return fmt.Errorf("the clip says %q but %v out of %v is %.3g, not %.3g. The arithmetic is the content here, so the one memorable line cannot be the false one — pick the phrase that fits, or check the measurements",
			r.Phrase, r.Subject.Value, r.Reference.Value, actual, claimed)
	}

	// The reference comes first: the fraction is OF it, so it has to exist before
	// anything can be a fraction of it.
	if p.Beats[0].Ratio == nil || p.Beats[0].Ratio.ResolvedShow() != "reference" {
		return fmt.Errorf("beat %q does not establish the larger quantity. The fraction is OF something, so that something has to be on screen before the comparison can mean anything",
			p.Beats[0].ID)
	}

	counts := map[string]int{}
	order := make([]string, 0, len(p.Beats))
	for _, b := range p.Beats {
		if b.Ratio == nil {
			return fmt.Errorf("beat %q has no ratio direction — every beat states the reference, the subject, the fraction, or what it means", b.ID)
		}
		show := b.Ratio.ResolvedShow()
		counts[show]++
		order = append(order, show)
	}
	for _, once := range []string{"reference", "subject", "fraction"} {
		if counts[once] != 1 {
			return fmt.Errorf("there are %d beats showing %q, want exactly 1", counts[once], once)
		}
	}
	iRef, iSub, iFrac := indexOfShow(order, "reference"), indexOfShow(order, "subject"), indexOfShow(order, "fraction")
	if !(iRef < iSub && iSub < iFrac) {
		return fmt.Errorf("the beats run %v. The order is the argument — what it is measured against, then the measurement, then the proportion — and naming the fraction early throws away the only surprise the clip has",
			order)
	}
	return nil
}

// ratioScenes lays the clip out as ONE scene: two bars against a shared scale
// and the phrase between them.
func ratioScenes(in SnippetSceneInput) ([]Scene, error) {
	r := in.Plan.Ratio
	if r == nil {
		return nil, fmt.Errorf("the plan has no pair")
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Ratio == nil {
			return nil, fmt.Errorf("beat %q has no ratio direction", beat.ID)
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Ratio.ResolvedShow(),
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneRatio,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title": in.Plan.Title,
			"unit":  r.Unit,
			"reference": map[string]any{
				"label": r.Reference.Label,
				"value": r.Reference.Value,
				"role":  r.Reference.ResolvedRole(),
			},
			"subject": map[string]any{
				"label": r.Subject.Label,
				"value": r.Subject.Value,
				"role":  r.Subject.ResolvedRole(),
				// The subject's share of the reference, so the renderer never
				// divides — the same division the validator checked.
				"frac": roundTo(r.Fraction(), 4),
			},
			"phrase": r.Phrase,
			"note":   r.Note,
			"steps":  steps,
		}),
	}}, nil
}
