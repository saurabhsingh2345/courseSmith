package pipeline

// The growth template: how it scales.
//
// Big-O is taught as notation, and notation is the last thing about it that
// matters. The fact a career-switcher needs is physical: two pieces of code
// that are indistinguishable on ten items are not remotely the same program on
// a million, and the difference is not a constant factor you can buy your way
// out of with a faster machine. That fact is a SHAPE — one line that stays flat
// and another that leaves the top of the frame — and a shape is a picture, not
// a symbol.
//
// So the clip is a chart with an n axis, an ops axis and two to four curves
// drawn on in order from the slowest-growing to the fastest, followed by a
// vertical probe line dropped at a real input size that reads the damage off
// each curve. The ordering is not tidiness. A legend that runs slow to fast is
// itself the lesson — it is the hierarchy of complexity classes, in order, and
// a viewer who reads it top to bottom has learned something even if they miss
// every word of the narration.
//
// Which is why the class order is enforced and the two offending entries are
// named. The rest of the arithmetic is done in Go: every curve's polyline is
// sampled and normalised here, and so is its reading at the probe, formatted
// with separators or in scientific notation when the number stops fitting in a
// human sentence. The component receives geometry and strings. Nothing about
// this chart is computed in TypeScript, because a complexity chart whose curves
// are drawn approximately is a chart that teaches the approximation.

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
		Name:        "growth",
		Category:    CatNumbers,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "How it scales",
		Description: "Complexity curves drawn onto one n-against-ops chart in growth order, then a probe line dropped at a real input size to read what each one costs there. Reach for it when the subject is why an approach dies at scale — nested loops, a linear scan, the value of a sort.",
		Example:     "O(n) against O(n²): why the nested loop dies at a million items",
		PromptFile:  snippetGrowthTemplateName,
		NeedsCode:   false,
		// Empty axes, two or three curves drawing on, the probe and the moral:
		// under thirty-five seconds a curve does not finish its draw-on before
		// the next one starts, and the divergence is the whole picture.
		MinTargetSec: 35,
		// Three curves plus the axes, the probe and the moral is six beats,
		// which is what fifty seconds funds.
		DefaultTargetSec: 50,
		// Four curves plus the three framing shots is seven, with one spare.
		// A fifth curve is a chart nobody reads.
		MaxBeats: 8,
		// A beat here is a SHOT — one curve arriving, or the probe dropping —
		// so twenty-eight words is about as long as one of them holds.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Growth: true},
		OwnsPlan:          planFields{Growth: true},
		Normalize:         normalizeGrowthPlan,
		Validate:          validateGrowthPlan,
		Scenes:            growthScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(GrowthShows(), ", "),
				"Classes":       strings.Join(GrowthClasses(), ", "),
				"ClassOrder":    strings.Join(growthClassOrder, " < "),
				"MinCurves":     minGrowthCurves,
				"MaxCurves":     maxGrowthCurves,
				"MaxLabelWords": maxGrowthLabelWords,
				"MaxProbe":      maxGrowthProbe,
			}
		},
	})
}

const snippetGrowthTemplateName = "snippet_growth.tmpl"

const (
	// One curve is a line, and a line has no story. The entire content of this
	// picture is a comparison, so two is the floor.
	minGrowthCurves = 2
	// Four curves is the most the frame labels at their ends without the
	// labels colliding where the fast ones bunch together near the top.
	maxGrowthCurves = 4
	// "binary search", "the nested loop" — a name for the code, not a
	// description of it.
	maxGrowthLabelWords = 4

	// A billion items. Past that the reading is theatre: nobody has intuition
	// about the difference between 10^18 and 10^21 operations, so the probe
	// stops teaching and starts impressing.
	maxGrowthProbe = 1000000000

	// Twenty-four samples draws a smooth curve at this width without shipping
	// a polyline nobody can read in a scene graph.
	growthSamples = 24
	// The right-hand edge of the chart, in items. Small on purpose: the point
	// is that the curves separate almost immediately, and a wide x domain
	// flattens everything but the fastest into the same line along the bottom.
	growthVisN = 24
	// The ops value at the top of the frame. Set so O(n) climbs to about
	// three-fifths of the height at the right edge — which leaves the flat
	// classes visibly flat below it, and sends n log n, n² and 2ⁿ off the top
	// of the chart at three visibly different points. That exit is the shot.
	growthVisCeiling = 40.0

	// Above a thousand million million, digits with separators stop being a
	// quantity and become a texture, so the reading switches to scientific.
	growthPlainCeiling = 1e15
)

// growthClassOrder is the complexity hierarchy, slowest-growing first. It is
// the order the legend must be written in, because the legend read top to
// bottom is itself the lesson.
var growthClassOrder = []string{"1", "logn", "n", "nlogn", "n2", "2n"}

// growthClasses maps each class to its rank in the hierarchy.
var growthClasses = func() map[string]int {
	m := make(map[string]int, len(growthClassOrder))
	for i, c := range growthClassOrder {
		m[c] = i
	}
	return m
}()

// GrowthClasses returns the class vocabulary sorted.
func GrowthClasses() []string {
	out := make([]string, 0, len(growthClasses))
	for k := range growthClasses {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// growthNotation is how each class is set on screen, in mono, at the end of
// its curve.
var growthNotation = map[string]string{
	"1":     "O(1)",
	"logn":  "O(log n)",
	"n":     "O(n)",
	"nlogn": "O(n log n)",
	"n2":    "O(n²)",
	"2n":    "O(2ⁿ)",
}

// growthShows is the closed vocabulary of what a beat does.
var growthShows = map[string]bool{
	// The empty chart: two axes and nothing on them. The opener.
	"axes": true,
	// Curve At draws in from the origin with its label.
	"curve": true,
	// The vertical line drops at the probe and each curve's reading lands.
	"probe": true,
	// The one-line takeaway. The closer.
	"moral": true,
}

// GrowthShows returns the beat vocabulary sorted.
func GrowthShows() []string {
	out := make([]string, 0, len(growthShows))
	for k := range growthShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// GrowthCurve is one complexity class on the chart.
type GrowthCurve struct {
	// Class is a growthClasses key — "n", "n2", "logn".
	Class string `json:"class"`
	// Label is the code this curve stands for — "binary search".
	Label string `json:"label,omitempty"`
}

// ResolvedClass returns the curve's class, defaulting the unknown to linear.
// O(n) is the honest default: it is the middle of the hierarchy, so a class
// that did not survive lands somewhere that makes no dramatic claim in either
// direction.
func (c GrowthCurve) ResolvedClass() string {
	s := strings.ToLower(strings.TrimSpace(c.Class))
	s = strings.NewReplacer(" ", "", "^", "", "*", "").Replace(s)
	// The spellings a model reaches for when it is writing notation rather
	// than a vocabulary key. Repaired rather than argued about.
	switch s {
	case "o(1)", "const", "constant":
		s = "1"
	case "o(logn)", "log", "logn":
		s = "logn"
	case "o(n)", "linear":
		s = "n"
	case "o(nlogn)", "nlogn", "linearithmic":
		s = "nlogn"
	case "o(n2)", "n²", "o(n²)", "n2", "quadratic":
		s = "n2"
	case "o(2n)", "2ⁿ", "o(2ⁿ)", "2n", "exponential":
		s = "2n"
	}
	if _, ok := growthClasses[s]; ok {
		return s
	}
	return "n"
}

// GrowthSpec is the chart: which curves are on it and where the probe drops.
// On the plan because the axes stand for the whole clip.
type GrowthSpec struct {
	// Curves are the complexity classes, in growth order, slowest first.
	Curves []GrowthCurve `json:"curves"`
	// Probe is the input size the drop-line reads at, or 0 for no probe.
	Probe int `json:"probe,omitempty"`
}

// GrowthBeat is one shot: which state of the chart this beat shows.
type GrowthBeat struct {
	// Show is a growthShows name.
	Show string `json:"show"`
	// At is the curve this beat draws, for "curve".
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's shot, defaulting the unknown to a curve —
// the workhorse state most beats of this template are in.
func (b GrowthBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if growthShows[s] {
		return s
	}
	return "curve"
}

func normalizeGrowthPlan(p *SnippetPlan) {
	g := p.Growth
	if g == nil {
		return
	}
	if len(g.Curves) > maxGrowthCurves {
		g.Curves = g.Curves[:maxGrowthCurves]
	}
	for i := range g.Curves {
		g.Curves[i].Class = g.Curves[i].ResolvedClass()
		g.Curves[i].Label = clampWords(collapseSpaces(g.Curves[i].Label), maxGrowthLabelWords)
	}
	if g.Probe < 0 {
		g.Probe = 0
	}
	if g.Probe > maxGrowthProbe {
		g.Probe = maxGrowthProbe
	}

	last := len(g.Curves) - 1
	for i := range p.Beats {
		b := p.Beats[i].Growth
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.At < 0 {
			b.At = 0
		}
		if last >= 0 && b.At > last {
			b.At = last
		}
	}
}

func validateGrowthPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Growth: true}); err != nil {
		return err
	}

	g := p.Growth
	if g == nil {
		return fmt.Errorf("the plan has no curves — the entire content of this picture is a comparison between growth rates, so an empty chart is an empty clip")
	}
	if n := len(g.Curves); n < minGrowthCurves || n > maxGrowthCurves {
		return fmt.Errorf("the chart has %d curves, want %d-%d. One curve is a line and a line has no story, and past %d the labels collide where the fast curves bunch together near the top",
			n, minGrowthCurves, maxGrowthCurves, maxGrowthCurves)
	}
	seen := map[string]int{}
	for i, c := range g.Curves {
		class := c.ResolvedClass()
		if j, dup := seen[class]; dup {
			return fmt.Errorf("curves %d and %d are both %s. Two curves of the same class are one line drawn twice — if the point is that two different pieces of code share a complexity, say that in the narration and give the chart something to compare against",
				j, i, growthNotation[class])
		}
		seen[class] = i
		if strings.TrimSpace(c.Label) == "" {
			return fmt.Errorf("curve %d (%s) has no label. The notation goes on the curve automatically; the label is what the viewer came for — the actual code this line stands for, like \"the nested loop\"",
				i, growthNotation[class])
		}
	}
	// The hierarchy, in order, because the legend read top to bottom is itself
	// the lesson. The two offending entries are named so the fix is obvious.
	for i := 1; i < len(g.Curves); i++ {
		prev, cur := g.Curves[i-1].ResolvedClass(), g.Curves[i].ResolvedClass()
		if growthClasses[cur] < growthClasses[prev] {
			return fmt.Errorf("curve %d is %s and curve %d is %s, so the chart lists the faster-growing one first. The curves are drawn in the order you give them and the legend has to read slow to fast — %s — because that ordering is the hierarchy the viewer is here to learn. Swap those two",
				i-1, growthNotation[prev], i, growthNotation[cur], strings.Join(growthNotationOrder(), " < "))
		}
	}

	if p.Beats[0].Growth == nil || p.Beats[0].Growth.ResolvedShow() != "axes" {
		return fmt.Errorf("beat %q does not open on the empty chart. A curve arriving on axes the viewer has not read is a shape with no units — open with {\"show\": \"axes\"}",
			p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Growth == nil || last.Growth.ResolvedShow() != "moral" {
		return fmt.Errorf("the clip does not close on the takeaway. The chart is evidence and the moral is the verdict it supports — end with {\"show\": \"moral\"}")
	}

	want := 0
	probeAt := -1
	for i, b := range p.Beats {
		gb := b.Growth
		if gb == nil {
			return fmt.Errorf("beat %q has no growth direction — every beat shows one state of the chart", b.ID)
		}
		switch gb.ResolvedShow() {
		case "axes":
			if i != 0 {
				return fmt.Errorf("beat %q clears back to bare axes part-way through. Curves stay on the chart once drawn — the accumulation is the comparison", b.ID)
			}
		case "curve":
			if gb.At > want {
				return fmt.Errorf("beat %q draws curve %d (%s) while curve %d (%s) is not on the chart yet. The curves arrive in the order you listed them, slowest first, so the viewer watches the hierarchy build",
					b.ID, gb.At, growthCurveName(g, gb.At), want, growthCurveName(g, want))
			}
			if gb.At < want {
				return fmt.Errorf("beat %q draws curve %d (%s) a second time. Each curve arrives once — a beat that wants to talk about it again can, it just does not need to redraw it",
					b.ID, gb.At, growthCurveName(g, gb.At))
			}
			want++
		case "probe":
			if probeAt >= 0 {
				return fmt.Errorf("beat %q drops the probe again, after beat %q already did. One drop-line: a second reading at the same n is the same frame", b.ID, p.Beats[probeAt].ID)
			}
			if want < len(g.Curves) {
				return fmt.Errorf("beat %q drops the probe while only %d of the %d curves are drawn, so %s gets no reading. The probe is the moment the chart is cashed in for numbers, and it has to read every line on it",
					b.ID, want, len(g.Curves), growthCurveName(g, want))
			}
			if g.Probe <= 0 {
				return fmt.Errorf("beat %q drops a probe but growth.probe is %d. The drop-line has to land at a real input size — a million items, ten thousand rows — because the number it reads off each curve is the whole payoff of the beat",
					b.ID, g.Probe)
			}
			probeAt = i
		case "moral":
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q states the moral before the end. The verdict comes after the evidence — \"moral\" is the closer", b.ID)
			}
		}
	}
	if want != len(g.Curves) {
		return fmt.Errorf("the clip draws %d of the %d curves, so %s is in the plan and never appears on the chart. Either give it a beat or take it off the plan",
			want, len(g.Curves), growthCurveName(g, want))
	}
	return nil
}

// growthNotationOrder is the hierarchy spelled the way it is set on screen,
// for the ordering rejection.
func growthNotationOrder() []string {
	out := make([]string, 0, len(growthClassOrder))
	for _, c := range growthClassOrder {
		out = append(out, growthNotation[c])
	}
	return out
}

// growthCurveName quotes a curve for an error. Index-safe because the
// validator can be handed a plan that never went through normalize.
func growthCurveName(g *GrowthSpec, i int) string {
	if i < 0 || i >= len(g.Curves) {
		return "no such curve"
	}
	c := g.Curves[i]
	return fmt.Sprintf("%s, %q", growthNotation[c.ResolvedClass()], c.Label)
}

// growthOps is the class's cost at n. The one definition of what each curve
// means, used for both the polyline and the probe reading.
func growthOps(class string, n float64) float64 {
	switch class {
	case "1":
		return 1
	case "logn":
		return math.Log2(n)
	case "n":
		return n
	case "nlogn":
		return n * math.Log2(n)
	case "n2":
		return n * n
	case "2n":
		return math.Pow(2, n)
	}
	return n
}

// growthLog10 is the base-10 logarithm of growthOps, computed directly so a
// reading that overflows a float64 — 2^1000000 is not a number Go can hold —
// still has an exponent to print.
func growthLog10(class string, n float64) float64 {
	switch class {
	case "1":
		return 0
	case "logn":
		return math.Log10(math.Log2(n))
	case "n":
		return math.Log10(n)
	case "nlogn":
		return math.Log10(n) + math.Log10(math.Log2(n))
	case "n2":
		return 2 * math.Log10(n)
	case "2n":
		return n * math.Log10(2)
	}
	return math.Log10(n)
}

// growthCommas sets an integer with thousand separators. Local because there
// is no shared one and adding a package-level Commas() collides with three
// other templates' private helpers.
func growthCommas(v int64) string {
	s := strconv.FormatInt(v, 10)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	if neg {
		return "-" + b.String()
	}
	return b.String()
}

// growthReading is what a curve costs at n, as it is set on screen: separated
// digits while that is still a quantity a person can feel, and scientific
// notation once it is only a texture.
func growthReading(class string, n int) string {
	v := growthOps(class, float64(n))
	if !math.IsInf(v, 0) && !math.IsNaN(v) && v < growthPlainCeiling {
		return growthCommas(int64(math.Round(v)))
	}
	l := growthLog10(class, float64(n))
	exp := math.Floor(l)
	mant := math.Pow(10, l-exp)
	return fmt.Sprintf("%.1fe+%d", mant, int64(exp))
}

// growthScenes lays the clip out as ONE scene, with every curve already
// sampled, normalised and clamped, and every probe reading already formatted.
// The component draws polylines and prints strings; it does no maths.
func growthScenes(in SnippetSceneInput) ([]Scene, error) {
	g := in.Plan.Growth
	if g == nil {
		return nil, fmt.Errorf("the plan has no curves")
	}

	curves := make([]map[string]any, 0, len(g.Curves))
	for _, c := range g.Curves {
		class := c.ResolvedClass()
		points := make([]float64, growthSamples)
		for i := 0; i < growthSamples; i++ {
			t := float64(i) / float64(growthSamples-1)
			n := 1 + t*float64(growthVisN-1)
			y := growthOps(class, n) / growthVisCeiling
			// Clamped at the frame top on purpose: a curve that leaves the
			// chart is the shot, and where it leaves is the comparison.
			if y > 1 {
				y = 1
			}
			if y < 0 {
				y = 0
			}
			points[i] = math.Round(y*10000) / 10000
		}
		entry := map[string]any{
			"class":    class,
			"label":    c.Label,
			"notation": growthNotation[class],
			"points":   points,
		}
		if g.Probe > 0 {
			entry["reading"] = growthReading(class, g.Probe)
		}
		curves = append(curves, entry)
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	drawn := make([]int, 0, len(g.Curves))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Growth == nil {
			return nil, fmt.Errorf("beat %q has no growth direction", beat.ID)
		}
		show := beat.Growth.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "curve" {
			step["at"] = beat.Growth.At
			drawn = append(drawn, beat.Growth.At)
		}
		onChart := append([]int(nil), drawn...)
		sort.Ints(onChart)
		step["drawn"] = onChart
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	props := map[string]any{
		"title":  in.Plan.Title,
		"curves": curves,
		"probe":  g.Probe,
		"steps":  steps,
	}
	if g.Probe > 0 {
		props["probeLabel"] = growthCommas(int64(g.Probe))
		// The fastest-growing curve is the last one, because the ordering is
		// enforced. Shipped explicitly rather than left as "the last" so the
		// component paints the worst reading by what it IS, not by position.
		props["worst"] = len(g.Curves) - 1
	}
	return []Scene{{
		Type:    SceneGrowth,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props:   headlineProps(in.Plan, props),
	}}, nil
}
