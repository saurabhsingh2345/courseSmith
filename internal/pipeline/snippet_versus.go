package pipeline

// The versus template: two things head to head, and a verdict at the bottom.
//
// A foundations course is one long sequence of forced choices — array or
// linked list, TCP or UDP, process or thread, SQL or document store — and the
// learner's problem is never that they cannot list the differences. It is that
// a list of differences does not tell them which to pick. So this template is
// deliberately shaped like the decision rather than like the data: two panels
// facing each other across a centre spine, one dimension at a time landing on
// both sides, the winning cell tinted, and a verdict strip that says WHEN to
// use which.
//
// It is the course's compare workhorse, which is exactly why the validator is
// strict about the ending. Two rules exist because of the two ways a
// comparison quietly fails to be advice.
//
// The first: a verdict that is merely the name of one side. "TCP" is a
// preference, not guidance — it tells a viewer nothing about the case where
// the other one is right, and a clip that ends there has spent forty seconds
// building a scoreboard and then read out the score. So a verdict that reduces
// to Left or Right is rejected outright.
//
// The second: the clean sweep. When every row goes the same way, the honest
// question in the viewer's head is "then why does the other one exist?", and
// that is the moment a one-word verdict does the most damage. So a sweep
// demands a longer verdict — enough words to name the case the loser wins —
// because a sweep is the shape that most needs explaining, not the least.
//
// Rows land exactly once and in order, because a comparison that revisits a
// dimension is arguing rather than comparing, and one that skips a row leaves
// a cell on screen the voice never accounted for.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "versus",
		Category:    CatDecisions,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "This or that",
		Description: "Two contenders across a centre spine, their dimensions landing one at a time with the winning cell tinted, and a verdict strip that says when to reach for which. Reach for it when the subject is a forced choice a learner keeps meeting — TCP or UDP, array or list, process or thread.",
		Example:     "TCP against UDP: reliability or speed",
		PromptFile:  snippetVersusTemplateName,
		NeedsCode:   false,
		// Two names, up to six rows and a verdict. Under thirty seconds the
		// rows land faster than the eye can read across the spine.
		MinTargetSec:     30,
		DefaultTargetSec: 50,
		// The opener, six rows and the verdict. That is the widest this shape
		// gets, and a seventh row is a spec sheet rather than a choice.
		MaxBeats:  9,
		Owns:      beatFields{Versus: true},
		OwnsPlan:  planFields{Versus: true},
		Normalize: normalizeVersusPlan,
		Validate:  validateVersusPlan,
		Scenes:    versusScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":           strings.Join(MetricRoles(), ", "),
				"Shows":           strings.Join(VersusShows(), ", "),
				"Edges":           strings.Join(VersusEdges(), ", "),
				"MinRows":         minVersusRows,
				"MaxRows":         maxVersusRows,
				"MaxSideWords":    maxVersusSideWords,
				"MaxDimWords":     maxVersusDimWords,
				"MaxValWords":     maxVersusValWords,
				"MaxVerdictWords": maxVersusVerdictWords,
				"MinSweepWords":   minVersusSweepVerdictWords,
			}
		},
	})
}

const snippetVersusTemplateName = "snippet_versus.tmpl"

const (
	// Below three rows this is not a comparison, it is an assertion with an
	// example. Past six the cells shrink below the size a six-word value needs
	// and the stage stops being two panels and becomes a table.
	minVersusRows = 3
	maxVersusRows = 6

	// A contender is a name set in display type — "TCP", "linked list".
	maxVersusSideWords = 3
	// A dimension is a label on the spine — "ordering", "setup cost".
	maxVersusDimWords = 3
	// A cell value is a phrase, not a sentence: "guaranteed, in order".
	maxVersusValWords = 6
	// The verdict is one line across the foot of the frame.
	maxVersusVerdictWords = 14
	// A clean sweep needs enough words to name the case the loser still wins.
	// Six is the shortest sentence that can hold a "use X when Y" clause.
	minVersusSweepVerdictWords = 6
)

// versusShows is the closed vocabulary of what a beat does.
var versusShows = map[string]bool{
	// The two contenders alone, big. The opener.
	"face": true,
	// Row At lands on both panels, its edge tinting the winning cell.
	"row": true,
	// The verdict strip lands full width. The closer.
	"verdict": true,
}

// VersusShows returns the beat vocabulary sorted.
func VersusShows() []string {
	out := make([]string, 0, len(versusShows))
	for k := range versusShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// versusEdges is the closed vocabulary of who wins a row.
var versusEdges = map[string]bool{
	// The left panel's cell is tinted.
	"left": true,
	// The right panel's cell is tinted.
	"right": true,
	// Neither: the row is a difference, not an advantage.
	"even": true,
}

// VersusEdges returns the edge vocabulary sorted.
func VersusEdges() []string {
	out := make([]string, 0, len(versusEdges))
	for k := range versusEdges {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// VersusSpec is the head-to-head. On the plan because both panels and every
// row persist for the whole clip — beats reveal them, they do not change them.
type VersusSpec struct {
	// Left is the contender on the left of the spine.
	Left string `json:"left"`
	// Right is the contender on the right.
	Right string `json:"right"`
	// Rows are the dimensions compared, in the order they land.
	Rows []VersusRow `json:"rows"`
	// Verdict is the when-to-use-which line across the foot of the frame.
	Verdict string `json:"verdict"`
}

// VersusRow is one dimension, valued on both sides.
type VersusRow struct {
	// Dim names the dimension — "ordering", "setup cost".
	Dim string `json:"dim"`
	// LeftVal is the left panel's cell.
	LeftVal string `json:"leftVal"`
	// RightVal is the right panel's cell.
	RightVal string `json:"rightVal"`
	// Edge is which side this row favours: left, right or even.
	Edge string `json:"edge,omitempty"`
}

// ResolvedEdge returns who wins the row, defaulting the unknown to even — the
// honest reading of a row whose winner was never stated.
func (r VersusRow) ResolvedEdge() string {
	s := strings.ToLower(strings.TrimSpace(r.Edge))
	if versusEdges[s] {
		return s
	}
	return "even"
}

// VersusBeat is one shot of the head-to-head.
type VersusBeat struct {
	// Show is a versusShows name.
	Show string `json:"show"`
	// At indexes VersusSpec.Rows, for the "row" beats.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults the unknown to a row landing, which is what most beats
// of this template are.
func (b VersusBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if versusShows[s] {
		return s
	}
	return "row"
}

// versusFold reduces a name to the letters and digits in it, so "TCP." and
// "tcp" are the same answer. Used only to catch a verdict that is nothing but
// the name of one side.
func versusFold(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizeVersusPlan(p *SnippetPlan) {
	v := p.Versus
	if v == nil {
		return
	}
	v.Left = clampWords(collapseSpaces(v.Left), maxVersusSideWords)
	v.Right = clampWords(collapseSpaces(v.Right), maxVersusSideWords)
	v.Verdict = clampWords(collapseSpaces(v.Verdict), maxVersusVerdictWords)

	rows := make([]VersusRow, 0, len(v.Rows))
	for _, r := range v.Rows {
		r.Dim = clampWords(collapseSpaces(r.Dim), maxVersusDimWords)
		r.LeftVal = clampWords(collapseSpaces(r.LeftVal), maxVersusValWords)
		r.RightVal = clampWords(collapseSpaces(r.RightVal), maxVersusValWords)
		r.Edge = r.ResolvedEdge()
		if len(rows) < maxVersusRows {
			rows = append(rows, r)
		}
	}
	v.Rows = rows

	for i := range p.Beats {
		b := p.Beats[i].Versus
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "row" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(v.Rows); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateVersusPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Versus: true}); err != nil {
		return err
	}

	v := p.Versus
	if v == nil {
		return fmt.Errorf("the plan has no comparison — this template is two contenders across a spine, so the two names and their rows are the clip")
	}
	if strings.TrimSpace(v.Left) == "" || strings.TrimSpace(v.Right) == "" {
		return fmt.Errorf("the plan names %q against %q, and a head-to-head needs both sides named. The two names are set in display type at the top of the frame and they are what the whole clip is about", v.Left, v.Right)
	}
	if versusFold(v.Left) == versusFold(v.Right) {
		return fmt.Errorf("both sides are called %q, so there is nothing to choose between. Name two different things, or use a template that explains one thing rather than weighing two", v.Left)
	}
	if n := len(v.Rows); n < minVersusRows || n > maxVersusRows {
		return fmt.Errorf("the comparison has %d rows, want %d-%d. Below three this is an assertion with an example rather than a comparison, and past %d the cells shrink until the stage reads as a spreadsheet instead of a choice",
			n, minVersusRows, maxVersusRows, maxVersusRows)
	}
	for i, r := range v.Rows {
		if strings.TrimSpace(r.Dim) == "" {
			return fmt.Errorf("row %d has no dimension. The label on the spine is what makes two cells comparable — without it the row is two unrelated phrases facing each other", i)
		}
		if strings.TrimSpace(r.LeftVal) == "" || strings.TrimSpace(r.RightVal) == "" {
			return fmt.Errorf("row %d (%q) leaves one side empty: %q against %q. Both panels get a cell on every row, or the row draws as half a comparison", i, r.Dim, r.LeftVal, r.RightVal)
		}
	}

	if strings.TrimSpace(v.Verdict) == "" {
		return fmt.Errorf("the plan has no verdict. The strip across the foot of the frame is the only part of this clip that tells anybody what to DO, and without it the viewer leaves with a scoreboard")
	}
	// A verdict that reduces to one contender's name is a preference, not
	// advice, and this is the single most common way this template fails.
	if f := versusFold(v.Verdict); f == versusFold(v.Left) || f == versusFold(v.Right) {
		return fmt.Errorf("the verdict is just %q, which is the name of one side rather than guidance. %q is not advice — say WHEN to reach for it and when the other one is the right call, in a sentence somebody could act on",
			v.Verdict, v.Verdict)
	}
	// THE SWEEP. Counted here rather than trusted, because a clean sweep is the
	// shape that most needs explaining and the one a model is most likely to
	// sign off with a single word.
	sweep := true
	first := v.Rows[0].ResolvedEdge()
	for _, r := range v.Rows {
		if r.ResolvedEdge() != first {
			sweep = false
			break
		}
	}
	if sweep {
		if n := len(strings.Fields(v.Verdict)); n < minVersusSweepVerdictWords {
			return fmt.Errorf("every one of the %d rows goes the same way (%q), and the verdict is %d words: %q. A clean sweep is exactly the case that needs explaining — the viewer is already asking why the other one exists — so give it at least %d words and name the case %q still wins",
				len(v.Rows), first, n, v.Verdict, minVersusSweepVerdictWords, sweepLoser(v, first))
		}
	}

	if p.Beats[0].Versus == nil || p.Beats[0].Versus.ResolvedShow() != "face" {
		return fmt.Errorf("beat %q does not open on the two contenders. A row landing before the viewer knows who is being compared is four words either side of a line — open with {\"show\": \"face\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Versus == nil || last.Versus.ResolvedShow() != "verdict" {
		return fmt.Errorf("beat %q does not close on the verdict. The strip is the payoff of the whole shape, and ending on a row leaves the comparison unjudged — end with {\"show\": \"verdict\"}", last.ID)
	}

	next := 0
	for _, b := range p.Beats {
		d := b.Versus
		if d == nil {
			return fmt.Errorf("beat %q has no versus direction — every beat shows the contenders, lands a row, or delivers the verdict", b.ID)
		}
		if d.ResolvedShow() != "row" {
			continue
		}
		if d.At < 0 || d.At >= len(v.Rows) {
			return fmt.Errorf("beat %q lands row %d, which does not exist — the comparison has rows 0-%d", b.ID, d.At, len(v.Rows)-1)
		}
		if d.At != next {
			return fmt.Errorf("beat %q lands row %d (%q) when row %d (%q) is the next one due. Rows land once each, in order — a comparison that revisits a dimension is arguing rather than comparing",
				b.ID, d.At, v.Rows[d.At].Dim, next, v.Rows[next].Dim)
		}
		next++
	}
	if next != len(v.Rows) {
		return fmt.Errorf("the clip lands %d of %d rows, so %q is on screen with nothing said about it. Every row needs its own beat, or drop it from the comparison",
			next, len(v.Rows), v.Rows[next].Dim)
	}
	return nil
}

// sweepLoser names the side that lost every row, for the sweep rejection. On
// an "even" sweep neither side lost, so the phrasing stays neutral.
func sweepLoser(v *VersusSpec, edge string) string {
	switch edge {
	case "left":
		return v.Right
	case "right":
		return v.Left
	default:
		return "either side"
	}
}

// versusScenes lays the clip out as ONE scene. The tally and the sweep flag are
// counted here so the component paints a result it was handed rather than
// re-deriving the argument in TypeScript.
func versusScenes(in SnippetSceneInput) ([]Scene, error) {
	v := in.Plan.Versus
	if v == nil {
		return nil, fmt.Errorf("the plan has no comparison")
	}
	if len(v.Rows) == 0 {
		return nil, fmt.Errorf("the comparison has no rows")
	}

	rows := make([]map[string]any, len(v.Rows))
	tally := map[string]int{"left": 0, "right": 0, "even": 0}
	for i, r := range v.Rows {
		edge := r.ResolvedEdge()
		tally[edge]++
		rows[i] = map[string]any{
			"dim":      r.Dim,
			"leftVal":  r.LeftVal,
			"rightVal": r.RightVal,
			"edge":     edge,
		}
	}
	sweep := tally[v.Rows[0].ResolvedEdge()] == len(v.Rows)

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	landed := make([]int, 0, len(v.Rows))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Versus == nil {
			return nil, fmt.Errorf("beat %q has no versus direction", beat.ID)
		}
		show := beat.Versus.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "row":
			at := beat.Versus.At
			if at < 0 || at >= len(v.Rows) {
				return nil, fmt.Errorf("beat %q lands row %d, which does not exist", beat.ID, at)
			}
			landed = append(landed, at)
			step["at"] = at
		case "verdict":
			// The strip lands over the finished table, so the closer shows
			// every row whatever the beats managed to cover.
			landed = landed[:0]
			for j := range v.Rows {
				landed = append(landed, j)
			}
		}
		up := make([]int, len(landed))
		copy(up, landed)
		sort.Ints(up)
		step["landed"] = up
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneVersus,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":     in.Plan.Title,
			"left":      v.Left,
			"right":     v.Right,
			"rows":      rows,
			"verdict":   v.Verdict,
			"leftWins":  tally["left"],
			"rightWins": tally["right"],
			"evens":     tally["even"],
			"sweep":     sweep,
			"steps":     steps,
		}),
	}}, nil
}
