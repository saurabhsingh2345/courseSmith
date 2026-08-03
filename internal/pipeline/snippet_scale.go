package pipeline

// The scale template: worlds inside worlds, the camera pulling back.
//
// This is the template `gauge` refuses to be. A gauge caps at four times its
// ceiling and says so in its error message, because past that the line is a
// hairline against the left edge and the picture states nothing. But the
// interesting quantities in teaching are almost never within four times of each
// other — a kilobyte and a terabyte, a millisecond and a day, one user and a
// million — and drawing those as bars gives you one bar and a row of invisible
// stubs. Every honest chart of a thousand-fold gap is a chart nobody can read.
//
// So the frame does not compare lengths at all. It NESTS. Each level is a box
// containing the one before it, and the camera pulls back a level at a time, so
// the thing that filled the screen a moment ago is now a dot inside something
// bigger. The multiplier is written on the move. That is the only visual device
// that survives four orders of magnitude, and it is why the picture makes its
// point at ratios where a bar chart has already given up.
//
// Two rules earn it, and both are about refusing to draw a lie.
//
// The levels must strictly ascend. A zoom that goes back in is not a zoom, and
// the whole grammar here — the previous level shrinking to a dot — depends on
// there being a direction.
//
// And each step must be at least four times the last. Below that the nesting is
// invisible: a box 1.5x another box, drawn inside it, reads as a border. That
// number is deliberately the same 4x `gauge` gives up at, so the two templates
// meet exactly and neither has to draw the range the other is better at. Under
// four times, the honest picture is a bar against a ceiling, and the error says
// so by name.

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "scale",
		Category:    CatNumbers,
		Since:       SinceV2,
		Title:       "Ten times bigger",
		Description: "Each thing nested inside the next, the camera pulling back an order of magnitude at a time. Reach for it when the gap between two numbers is too wide for a chart to draw honestly — bytes to terabytes, one user to a million.",
		Example:     "How much data is a kilobyte, a photo, a film, and everything Netflix streams in a day",
		PromptFile:  snippetScaleTemplateName,
		NeedsCode:   false,
		// Three levels and a look at the whole ladder is four beats.
		MinTargetSec:     30,
		DefaultTargetSec: 50,
		Owns:             beatFields{Scale: true},
		OwnsPlan:         planFields{Scale: true},
		Normalize:        normalizeScalePlan,
		Validate:         validateScalePlan,
		Scenes:           scaleScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":         strings.Join(ScaleShows(), ", "),
				"Icons":         strings.Join(PointIconNames(), ", "),
				"MinLevels":     minScaleLevels,
				"MaxLevels":     maxScaleLevels,
				"MinStep":       int(minScaleStep),
				"MaxUnitWords":  maxScaleUnitWords,
				"MaxLabelWords": maxScaleLabelWords,
				"MaxNoteWords":  maxScaleNoteWords,
			}
		},
	})
}

const snippetScaleTemplateName = "snippet_scale.tmpl"

const (
	// Two levels is a comparison and `gauge` or `compare` draws it better. Five
	// nested boxes is already a 1000x span at the minimum step; a sixth is a
	// level the viewer cannot hold in their head alongside the other five.
	minScaleLevels = 3
	maxScaleLevels = 5

	// The least a step may be and still read as containment. A box 1.5x another
	// drawn inside it is a border, not a world. Deliberately the same figure
	// `gauge` gives up at, so the two templates meet exactly.
	minScaleStep = 4.0

	maxScaleUnitWords  = 2
	maxScaleLabelWords = 6
	maxScaleNoteWords  = 16
)

// scaleShows is the closed vocabulary of what a beat does.
var scaleShows = map[string]bool{
	// Pull back to one level: it fills the frame, the one before it is a box
	// inside it, and the multiplier lands on the move.
	"level": true,
	// The whole ladder at once, every level nested. The last beat, and optional.
	"whole": true,
}

// ScaleShows returns the beat vocabulary sorted.
func ScaleShows() []string {
	out := make([]string, 0, len(scaleShows))
	for k := range scaleShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ScaleSpec is the ladder.
type ScaleSpec struct {
	// Unit is what every level is measured in — "GB", "users", "ms". One unit
	// per clip: a ladder that changes units half way up is two ladders.
	Unit string `json:"unit"`
	// Levels are the rungs, smallest first.
	Levels []ScaleLevel `json:"levels"`
}

// ScaleLevel is one rung.
type ScaleLevel struct {
	// Label is the thing at this size — "One page of text", "Every film on
	// Netflix". A number the viewer cannot picture is a number they cannot
	// remember, so the label is an object, not a category.
	Label string `json:"label"`
	// Value is the quantity, in the ladder's unit. It is what the multipliers
	// are computed from, so it is a bare number rather than a written one.
	Value float64 `json:"value"`
	// Display is how the quantity is written on screen — "7.5 quintillion",
	// "2 KB". Filled in from Value when the model does not say.
	Display string `json:"display,omitempty"`
	Icon    string `json:"icon,omitempty"`
	// Note is the line that arrives when the camera settles on this level.
	Note string `json:"note,omitempty"`
}

// ScaleBeat is one move.
type ScaleBeat struct {
	Show string `json:"show"`
	// At indexes the level a `level` beat pulls back to.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to landing on
// a level — which is what most beats of this template do.
func (b ScaleBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if scaleShows[s] {
		return s
	}
	return "level"
}

func normalizeScalePlan(p *SnippetPlan) {
	s := p.Scale
	if s == nil {
		return
	}
	s.Unit = clampWords(collapseSpaces(s.Unit), maxScaleUnitWords)

	levels := make([]ScaleLevel, 0, len(s.Levels))
	for _, l := range s.Levels {
		l.Label = clampWords(collapseSpaces(l.Label), maxScaleLabelWords)
		l.Note = clampWords(collapseSpaces(l.Note), maxScaleNoteWords)
		l.Display = collapseSpaces(l.Display)
		l.Icon = normalizePointIconName(l.Icon)
		// A negative size is a sign error, not a quantity.
		l.Value = math.Abs(l.Value)
		if l.Display == "" {
			l.Display = scaleDisplay(l.Value)
		}
		if l.Label != "" && l.Value > 0 && len(levels) < maxScaleLevels {
			levels = append(levels, l)
		}
	}
	s.Levels = levels

	for i := range p.Beats {
		b := p.Beats[i].Scale
		if b == nil {
			continue
		}
		if !scaleShows[strings.ToLower(strings.TrimSpace(b.Show))] {
			if i == len(p.Beats)-1 && i > 0 {
				b.Show = "whole"
			} else {
				b.Show = "level"
			}
		} else {
			b.Show = b.ResolvedShow()
		}
		if b.Show != "level" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(s.Levels); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateScalePlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Scale: true}); err != nil {
		return err
	}

	s := p.Scale
	if s == nil {
		return fmt.Errorf("the plan has no scale — this template is a run of things, each one many times bigger than the last")
	}
	if strings.TrimSpace(s.Unit) == "" {
		return fmt.Errorf("the ladder has no unit — every rung is measured in the same thing, and the viewer needs to know what")
	}
	if n := len(s.Levels); n < minScaleLevels || n > maxScaleLevels {
		return fmt.Errorf("there are %d levels, want %d-%d. Two things is a comparison and gauge or compare draws it better; past %d the viewer cannot hold the ladder in their head",
			n, minScaleLevels, maxScaleLevels, maxScaleLevels)
	}

	seen := map[string]bool{}
	for i, l := range s.Levels {
		if strings.TrimSpace(l.Label) == "" {
			return fmt.Errorf("level %d has no label. A bare quantity is a number nobody can picture — say what is that big", i)
		}
		if l.Value <= 0 {
			return fmt.Errorf("level %q measures %v — every rung is a real quantity", l.Label, l.Value)
		}
		key := strings.ToLower(strings.TrimSpace(l.Label))
		if seen[key] {
			return fmt.Errorf("two levels are both called %q — each rung is a different thing at a different size", l.Label)
		}
		seen[key] = true

		if i == 0 {
			continue
		}
		prev := s.Levels[i-1]
		// The rules this template exists for.
		if l.Value <= prev.Value {
			return fmt.Errorf("level %q (%s) is not bigger than %q (%s). The camera only pulls back: a ladder that goes down a rung is not a ladder, and the picture — the last thing shrinking to a dot inside this one — has nothing to show",
				l.Label, l.Display, prev.Label, prev.Display)
		}
		if r := l.Value / prev.Value; r < minScaleStep {
			return fmt.Errorf("level %q is only %.1fx level %q, under the %gx this can draw. A box %.1f times another, drawn inside it, reads as a border rather than as a world — at that ratio the honest picture is a bar against a ceiling, which is the gauge template. Either drop a rung so the steps are wider, or use gauge",
				l.Label, r, prev.Label, minScaleStep, r)
		}
	}

	counts := map[string]int{}
	next := 0
	for i, b := range p.Beats {
		if b.Scale == nil {
			return fmt.Errorf("beat %q has no scale direction — every beat pulls back to a level, or shows the whole ladder", b.ID)
		}
		show := b.Scale.ResolvedShow()
		counts[show]++

		if show == "whole" {
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q shows the whole ladder but the clip carries on afterwards. Once every level is in frame there is nowhere further back to go", b.ID)
			}
			continue
		}
		if b.Scale.At != next {
			var at string
			if b.Scale.At >= 0 && b.Scale.At < len(s.Levels) {
				at = fmt.Sprintf(" (%q)", s.Levels[b.Scale.At].Label)
			}
			return fmt.Errorf("beat %q pulls back to level %d%s when the camera is at level %d. The ladder is climbed one rung at a time from the smallest — jumping a rung throws away the multiplier, which is the only thing making the gap legible",
				b.ID, b.Scale.At, at, next)
		}
		next++
	}

	if next != len(s.Levels) {
		return fmt.Errorf("%d of the %d levels are never reached. A rung the camera never stops at is a box with a label the viewer never gets to read",
			len(s.Levels)-next, len(s.Levels))
	}
	if counts["whole"] > 1 {
		return fmt.Errorf("there are %d beats showing the whole ladder; it lands once, at the end", counts["whole"])
	}
	return nil
}

// scaleScenes lays the clip out as ONE scene, for the reason `workspace` does:
// a scene per beat remounts the composition, and a remount is a cut. This whole
// template is one continuous camera move, so a cut in the middle of it would
// destroy the only thing it is doing.
func scaleScenes(in SnippetSceneInput) ([]Scene, error) {
	s := in.Plan.Scale
	if s == nil {
		return nil, fmt.Errorf("the plan has no scale")
	}
	if len(s.Levels) == 0 {
		return nil, fmt.Errorf("the ladder has no levels")
	}

	levels := make([]map[string]any, len(s.Levels))
	for i, l := range s.Levels {
		level := map[string]any{
			"label":   l.Label,
			"value":   l.Value,
			"display": l.Display,
			"icon":    l.Icon,
			"note":    l.Note,
		}
		// The multiplier is computed here, once, rather than in the renderer:
		// it is the number the validator enforced, and the figure on screen has
		// to be that same figure or the picture is arguing with its own rules.
		if i > 0 {
			level["times"] = scaleTimes(l.Value / s.Levels[i-1].Value)
		}
		levels[i] = level
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Scale == nil {
			return nil, fmt.Errorf("beat %q has no scale direction", beat.ID)
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    beat.Scale.ResolvedShow(),
		}
		if beat.Scale.ResolvedShow() == "level" {
			step["at"] = beat.Scale.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	props := map[string]any{
		"title":  in.Plan.Title,
		"unit":   s.Unit,
		"levels": levels,
		"steps":  steps,
	}
	// The whole span, for the closing frame. It is the number the clip is
	// actually about — the reason none of this could be a bar chart — and it is
	// the one figure no single level carries.
	if n := len(s.Levels); n > 1 {
		props["span"] = spanWords(s.Levels[n-1].Value / s.Levels[0].Value)
	}
	return []Scene{{
		Type:    SceneScale,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props:   props,
	}}, nil
}

// spanWords writes a ratio the way a narrator would say it. "×40000000000" is a
// figure nobody reads; "×40 billion" is the point of the clip.
func spanWords(r float64) string {
	for _, u := range []struct {
		at   float64
		word string
	}{
		{1e12, " trillion"},
		{1e9, " billion"},
		{1e6, " million"},
		{1e3, " thousand"},
	} {
		if r >= u.at {
			return scaleTimes(r/u.at) + u.word
		}
	}
	return scaleTimes(r)
}

// scaleTimes writes a multiplier the way somebody would say it out loud: "10",
// not "10.0"; "2.5", not "2.4999999".
func scaleTimes(r float64) string {
	if r >= 100 || r == math.Trunc(r) {
		return strconv.FormatFloat(math.Round(r), 'f', -1, 64)
	}
	return strconv.FormatFloat(math.Round(r*10)/10, 'f', -1, 64)
}

// scaleDisplay writes a quantity compactly, for the levels where the model gave
// a number and no wording for it.
//
// It is a fallback and not the main path: "7.5 quintillion grains" is better
// than anything derivable from 7.5e18, and the prompt asks for it. What this
// guarantees is that a level which forgot still renders a figure rather than
// Go's default float spelling, which puts "1.048576e+06" on the screen.
func scaleDisplay(v float64) string {
	abs := math.Abs(v)
	units := []struct {
		at   float64
		suff string
	}{
		{1e12, "T"},
		{1e9, "B"},
		{1e6, "M"},
		{1e3, "K"},
	}
	for _, u := range units {
		if abs >= u.at {
			return trimZeros(strconv.FormatFloat(v/u.at, 'f', 1, 64)) + u.suff
		}
	}
	if abs >= 1 {
		return trimZeros(strconv.FormatFloat(v, 'f', 1, 64))
	}
	return trimZeros(strconv.FormatFloat(v, 'f', 3, 64))
}

// trimZeros drops a trailing ".0" / ".500" so a whole number reads as one.
func trimZeros(s string) string {
	if !strings.Contains(s, ".") {
		return s
	}
	s = strings.TrimRight(s, "0")
	return strings.TrimSuffix(s, ".")
}
