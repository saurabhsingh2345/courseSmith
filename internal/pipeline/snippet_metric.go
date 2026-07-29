package pipeline

// The metric template: one number at a time, large enough to be the whole frame.
//
// The catalog could already draw numbers — `data` puts them on a chart or a map
// — and that is a different job. A chart is for a number whose meaning is its
// *shape*: a trend, a distribution, a comparison across categories. This
// template is for a number whose meaning is its *size*, where the picture adds
// nothing and the only honest visual is the figure itself, set very large, with
// the unit small beside it and one line saying what it counts.
//
// The reference clips lean on this constantly and it is the hardest thing to do
// badly-but-passably: a number read out over a stock diagram is forgettable,
// and the same number alone on a dark frame at three hundred points is not.
//
// Two rules earn this template its place, and both are validators.
//
// The first is that every figure carries a unit and a label. "2.8 trillion" is
// a quantity; "2.8 trillion parameters, and each one has to be in memory before
// a token comes out" is an argument. A bare number is trivia, and a course full
// of trivia teaches nobody anything they can use.
//
// The second is `role`. Each figure is coloured by what it is *doing* in the
// argument — the quantity being measured, the ceiling it runs into, the
// alternative it is being weighed against — and those three colours are fixed
// across the whole catalog rather than derived from branding. A viewer who has
// watched two of these clips knows what red means before the narrator says it.
// See videoskin.go for why the semantic accents are not brand colours.

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "metric",
		Category:    CatNumbers,
		Since:       SinceV1,
		Title:       "Big numbers",
		Description: "One figure at a time, counting up on an almost empty frame. Reach for it when the point is how big something is, and a chart would add nothing.",
		Example:     "What it really costs to run a 70-billion-parameter model at home",
		PromptFile:  snippetMetricTemplateName,
		NeedsCode:   false,
		// A number needs saying, and a number needs a beat. Four figures plus a
		// recap is five beats before anything optional, which the shared 45s
		// budget funds; six figures is the ceiling and needs the longer runtime.
		MinTargetSec:     40,
		DefaultTargetSec: 55,
		// The beat count here is a property of the content — a subject with five
		// numbers worth stating has five beats whether or not attention lasts
		// seven pictures — so the ceiling goes above the shared one, same as
		// `stack` and `breakdown`.
		MaxBeats:  9,
		Owns:      beatFields{Metric: true},
		OwnsPlan:  planFields{Metric: true},
		Normalize: normalizeMetricPlan,
		Validate:  validateMetricPlan,
		Scenes:    metricScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(MetricShows(), ", "),
				"MinMetrics":    minMetrics,
				"MaxMetrics":    maxMetrics,
				"MaxValueChars": maxMetricValueChars,
				"MaxUnitWords":  maxMetricUnitWords,
				"MaxLabelWords": maxMetricLabelWords,
				"MaxNoteWords":  maxMetricNoteWords,
			}
		},
	})
}

const snippetMetricTemplateName = "snippet_metric.tmpl"

// Capacity. Two figures is a comparison and belongs to `compare`; seven is a
// table, and a table set at this size is a list nobody finishes.
const (
	minMetrics = 3
	maxMetrics = 6

	// A figure is set at roughly 260px. Eight characters is what fits across
	// the stage at that size with its unit beside it — "313K–577K" is the
	// realistic worst case and it is exactly nine, so the cap is generous by
	// one rather than tight.
	maxMetricValueChars = 9
	maxMetricUnitWords  = 3
	maxMetricLabelWords = 7
	maxMetricNoteWords  = 14
)

// metricRoles is the closed vocabulary of what a figure is doing in the
// argument. It maps onto the semantic accents, which is why it is a vocabulary
// rather than a colour: the template says what the number means and the design
// system decides what colour that is.
var metricRoles = map[string]bool{
	// The thing being measured — the subject's own number.
	"quantity": true,
	// A ceiling, a floor, a budget: the number the quantity runs into.
	"limit": true,
	// The alternative being weighed against the subject.
	"rival": true,
	// A number that is merely context and should not compete for meaning.
	"neutral": true,
}

// MetricRoles returns the role vocabulary sorted, for prompts and docs.
func MetricRoles() []string {
	out := make([]string, 0, len(metricRoles))
	for k := range metricRoles {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// metricShows is the closed vocabulary of what a beat does.
var metricShows = map[string]bool{
	// State one figure: it counts up and takes the frame.
	"state": true,
	// Bring every figure stated so far back as a row. The last beat only.
	"recap": true,
}

// MetricShows returns the beat vocabulary sorted.
func MetricShows() []string {
	out := make([]string, 0, len(metricShows))
	for k := range metricShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// MetricSpec is the set of figures the clip states. On the plan rather than
// per-beat because the recap needs all of them at once.
type MetricSpec struct {
	Figures []Metric `json:"figures"`
}

// Metric is one figure.
type Metric struct {
	// Value is the number as it should read on screen — "2.8", "512", "313K",
	// "<1". Not a sentence, and not carrying its own unit.
	Value string `json:"value"`
	// Unit is what the figure is counted in — "TB", "GB/s", "params", "ms".
	// Set small and tight against the value.
	Unit string `json:"unit,omitempty"`
	// Label says what is being counted. It sits above the figure in the
	// eyebrow slot.
	Label string `json:"label"`
	// Note is the one line under the figure that turns a quantity into an
	// argument. Optional, and the template is much weaker without it.
	Note string `json:"note,omitempty"`
	// Role is what this figure is doing: a metricRoles name. It picks the
	// semantic accent, so it is a claim about meaning rather than about colour.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the figure's role, defaulting to the neutral one — a
// number whose job the model did not state should not be shouting in red.
func (m Metric) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(m.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// countUpTarget is the numeric part of a value, for the count-up animation, and
// whether there was one at all.
//
// Not every figure counts: "<1" and "313K–577K" are ranges and thresholds that
// have no single number to run to, and animating them to a wrong intermediate
// reads as the clip lying for a few frames. Those are revealed instead.
var metricNumberRe = regexp.MustCompile(`^-?\d+(\.\d+)?$`)

func (m Metric) countsUp() bool {
	return metricNumberRe.MatchString(strings.TrimSpace(m.Value))
}

// MetricBeat is one move: which figure this beat states.
type MetricBeat struct {
	// Show is a metricShows name.
	Show string `json:"show"`
	// At indexes MetricSpec.Figures, for a "state" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to stating a
// figure — which is what all but one beat of this template does.
func (b MetricBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if metricShows[s] {
		return s
	}
	return "state"
}

func normalizeMetricPlan(p *SnippetPlan) {
	m := p.Metric
	if m == nil {
		return
	}

	figures := make([]Metric, 0, len(m.Figures))
	for _, f := range m.Figures {
		f.Value = clampChars(collapseSpaces(f.Value), maxMetricValueChars)
		f.Unit = clampWords(collapseSpaces(f.Unit), maxMetricUnitWords)
		f.Label = clampWords(collapseSpaces(f.Label), maxMetricLabelWords)
		f.Note = clampWords(collapseSpaces(f.Note), maxMetricNoteWords)
		f.Role = f.ResolvedRole()
		// A figure with no number is not a figure. Dropping it is the repair;
		// inventing one would be a claim about the subject.
		if f.Value != "" && f.Label != "" && len(figures) < maxMetrics {
			figures = append(figures, f)
		}
	}
	m.Figures = figures

	for i := range p.Beats {
		b := p.Beats[i].Metric
		if b == nil {
			continue
		}
		if !metricShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			// The shape says what the ends are for: a clip like this closes on
			// the recap if it has one, and everything else states a figure.
			b.Show = "state"
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "state" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(m.Figures); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

// clampChars trims a string to at most n characters, on a rune boundary.
func clampChars(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return strings.TrimSpace(string(r[:n]))
}

func validateMetricPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Metric: true}); err != nil {
		return err
	}

	m := p.Metric
	if m == nil {
		return fmt.Errorf("the plan has no figures — this template is a handful of numbers, each one taking the whole frame in turn")
	}
	if n := len(m.Figures); n < minMetrics || n > maxMetrics {
		return fmt.Errorf("there are %d figures, want %d-%d. Two numbers set against each other is the compare template; seven is a table, and a table set this large is a list nobody finishes",
			n, minMetrics, maxMetrics)
	}

	seenValue := map[string]bool{}
	for i, f := range m.Figures {
		if strings.TrimSpace(f.Value) == "" {
			return fmt.Errorf("figure %d has no value", i)
		}
		// The rule the template exists for. A number with no unit is not a
		// measurement, and a number with no label is trivia.
		if strings.TrimSpace(f.Label) == "" {
			return fmt.Errorf("figure %d (%q) has no label — say what is being counted. A number nobody labels is trivia, and a course full of trivia teaches nothing anyone can use",
				i, f.Value)
		}
		if strings.TrimSpace(f.Unit) == "" && f.countsUp() {
			return fmt.Errorf("figure %d (%q) has no unit. %q of what? A bare count is a quantity; a count with a unit is a measurement",
				i, f.Value, f.Value)
		}
		key := strings.ToLower(strings.TrimSpace(f.Value)) + "\x00" + strings.ToLower(strings.TrimSpace(f.Unit))
		if seenValue[key] {
			return fmt.Errorf("two figures are both %q %q — each beat states a different number", f.Value, f.Unit)
		}
		seenValue[key] = true
		if r := strings.ToLower(strings.TrimSpace(f.Role)); r != "" && !metricRoles[r] {
			return fmt.Errorf("figure %d has role %q, which is not one of: %s", i, f.Role, strings.Join(MetricRoles(), ", "))
		}
	}

	stated := map[int]bool{}
	counts := map[string]int{}
	for i, b := range p.Beats {
		if b.Metric == nil {
			return fmt.Errorf("beat %q has no metric direction — every beat states a figure or recaps them", b.ID)
		}
		show := b.Metric.ResolvedShow()
		counts[show]++
		if show == "recap" && i != len(p.Beats)-1 {
			return fmt.Errorf("beat %q recaps the figures but the clip carries on afterwards. The recap is the closing frame — it is the one that gets screenshotted", b.ID)
		}
		if show != "state" {
			continue
		}
		if b.Metric.At < 0 || b.Metric.At >= len(m.Figures) {
			return fmt.Errorf("beat %q states figure %d, which does not exist", b.ID, b.Metric.At)
		}
		if stated[b.Metric.At] {
			return fmt.Errorf("beat %q states figure %d again; each number gets one beat", b.ID, b.Metric.At)
		}
		stated[b.Metric.At] = true
	}
	if len(stated) != len(m.Figures) {
		return fmt.Errorf("%d of the %d figures are never said out loud. A number on screen that the narrator skips is a number nobody read — give it a beat or cut it",
			len(m.Figures)-len(stated), len(m.Figures))
	}
	if counts["recap"] > 1 {
		return fmt.Errorf("there are %d recap beats; the figures come back together once, at the end", counts["recap"])
	}
	// At least one figure has to be doing something in the argument. A clip
	// where every number is "neutral" is a clip with no point of view, which is
	// a list of facts rather than an explanation.
	roles := map[string]int{}
	for _, f := range m.Figures {
		roles[f.ResolvedRole()]++
	}
	if roles["neutral"] == len(m.Figures) {
		return fmt.Errorf("every figure is role %q, so nothing in the clip is being argued. Mark the number the subject is measured by as %q, and the ceiling or budget it runs into as %q — the colours are how a viewer reads the argument before the narrator finishes the sentence",
			"neutral", "quantity", "limit")
	}
	return nil
}

// metricScenes lays the clip out as ONE scene. The figures share a frame that
// only ever holds one of them at a time, so the count-up, the unit and the note
// are one component's business and the beats just say which figure is up.
func metricScenes(in SnippetSceneInput) ([]Scene, error) {
	m := in.Plan.Metric
	if m == nil {
		return nil, fmt.Errorf("the plan has no figures")
	}

	figures := make([]map[string]any, len(m.Figures))
	for i, f := range m.Figures {
		figures[i] = map[string]any{
			"value":    f.Value,
			"unit":     f.Unit,
			"label":    f.Label,
			"note":     f.Note,
			"role":     f.ResolvedRole(),
			"countsUp": f.countsUp(),
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Metric == nil {
			return nil, fmt.Errorf("beat %q has no metric direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Metric.ResolvedShow(),
		}
		if beat.Metric.ResolvedShow() == "state" {
			step["at"] = beat.Metric.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneMetric,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"title":   in.Plan.Title,
			"figures": figures,
			"steps":   steps,
		},
	}}, nil
}
