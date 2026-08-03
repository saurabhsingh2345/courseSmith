package pipeline

// The ranking template: an ordered board, and what happens when something new
// arrives on it.
//
// The catalog can draw a list (`rundown`), a set of magnitudes (`data`), and a
// number against a threshold (`gauge`). None of them can draw the thing that
// makes a leaderboard interesting, which is not the order — it is the order
// *changing*. A sorted set taking a write, a cache eviction, a queue reordering
// by priority, a benchmark table after the new entrant lands: in all of them the
// teaching moment is a row moving past another row, and a static list has no way
// to express it.
//
// So the board is one object that persists for the whole clip, and arrivals move
// through it. Rows slide; they are never redrawn in a new place. That is the
// whole design, and it is why this cannot be `data` with a bar chart: a chart
// re-sorted between beats is two charts, and the viewer has to diff them.
//
// Three rules earn it its place, and all three are validators.
//
// The board is established before anything arrives. A row moving on a board the
// viewer has not read yet is a row moving for no reason.
//
// **Every arrival must change the visible order.** This is the rule the template
// exists for. An entry that lands below the last visible row changes nothing on
// screen: the narrator says "and now this one comes in" and the picture does not
// move, which reads as a broken render rather than as a small number. If the
// honest value does not place, the clip should either show more rows or make a
// different point — and Go can tell, so it does.
//
// The board is short enough to read. Past eight rows the type is small and the
// movement is lost among rows that did not move; below four there is not enough
// board for a position to mean anything.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "ranking",
		Category:    CatSystems,
		Since:       SinceV4,
		Family:      FamilyReplica,
		Title:       "The board, and what moves it",
		Description: "An ordered board that re-sorts as new entries land. Reach for it when the point is not the ranking but the change to it — a write into a sorted set, an eviction, a new entrant.",
		Example:     "How a Redis sorted set keeps a live leaderboard in order",
		PromptFile:  snippetRankingTemplateName,
		NeedsCode:   false,
		// The board, then one beat per arrival, then what it means. Two arrivals
		// is already four beats, and an arrival not paused on is one nobody saw.
		MinTargetSec:     30,
		DefaultTargetSec: 50,
		MaxBeats:         8,
		Owns:             beatFields{Ranking: true},
		OwnsPlan:         planFields{Ranking: true},
		Normalize:        normalizeRankingPlan,
		Validate:         validateRankingPlan,
		Scenes:           rankingScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(RankingShows(), ", "),
				"MinRows":       minRankingRows,
				"MaxRows":       maxRankingRows,
				"MinArrivals":   minRankingArrivals,
				"MaxArrivals":   maxRankingArrivals,
				"MaxLabelWords": maxRankingLabelWords,
				"MaxNoteWords":  maxRankingNoteWords,
				"MaxUnitWords":  maxRankingUnitWords,
			}
		},
	})
}

const snippetRankingTemplateName = "snippet_ranking.tmpl"

const (
	// Below four rows a position means nothing — there is no board to move
	// through. Past eight the type is small and a single row sliding is lost
	// among seven that did not.
	minRankingRows = 4
	maxRankingRows = 8

	// One arrival is a demonstration; four is a montage nobody follows at this
	// pace, because each one needs its own beat to be read.
	minRankingArrivals = 1
	maxRankingArrivals = 3

	maxRankingLabelWords = 4
	maxRankingNoteWords  = 16
	maxRankingUnitWords  = 3
)

// rankingShows is the closed vocabulary of what a beat does.
var rankingShows = map[string]bool{
	// Draw the board as it stands. The first beat, always.
	"board": true,
	// One arrival lands and the board re-sorts around it.
	"insert": true,
	// Hold the settled board and say what the movement meant.
	"read": true,
}

// RankingShows returns the beat vocabulary sorted.
func RankingShows() []string {
	out := make([]string, 0, len(rankingShows))
	for k := range rankingShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// RankingSpec is the board and what lands on it. On the plan rather than
// per-beat because the board is one object the whole clip watches.
type RankingSpec struct {
	// Metric is what the board is ordered by — "score", "requests per second".
	Metric string `json:"metric"`
	// Unit is what a value is counted in ("" when the metric says it).
	Unit string `json:"unit,omitempty"`
	// Rows is the board as it stands before anything arrives.
	Rows []RankingEntry `json:"rows"`
	// Arrivals are the entries that land, in the order they land.
	Arrivals []RankingEntry `json:"arrivals"`
}

// RankingEntry is one line on the board, whether it started there or arrived.
type RankingEntry struct {
	// Label is the entry's name.
	Label string `json:"label"`
	// Value is what it is ranked by. Higher is better: the board is descending,
	// because every real leaderboard is and a viewer reads top as winning
	// without being told.
	Value float64 `json:"value"`
	// Note is the one line that says what this arrival means. Arrivals only —
	// a note on a starting row is furniture on a board nobody is reading yet.
	Note string `json:"note,omitempty"`
	// Role is what this entry is doing: a metricRoles name. Arrivals only.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the entry's role, defaulting to neutral.
func (e RankingEntry) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(e.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// RankingBeat is one move: which arrival lands.
type RankingBeat struct {
	// Show is a rankingShows name.
	Show string `json:"show"`
	// At indexes RankingSpec.Arrivals, for an "insert" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to an insert.
func (b RankingBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if rankingShows[s] {
		return s
	}
	return "insert"
}

func normalizeRankingPlan(p *SnippetPlan) {
	r := p.Ranking
	if r == nil {
		return
	}
	r.Metric = clampWords(collapseSpaces(r.Metric), maxRankingLabelWords)
	r.Unit = clampWords(collapseSpaces(r.Unit), maxRankingUnitWords)

	clean := func(in []RankingEntry, max int, keepNote bool) []RankingEntry {
		out := make([]RankingEntry, 0, len(in))
		for _, e := range in {
			e.Label = clampWords(collapseSpaces(e.Label), maxRankingLabelWords)
			if keepNote {
				e.Note = clampWords(collapseSpaces(e.Note), maxRankingNoteWords)
			} else {
				e.Note = ""
			}
			e.Role = e.ResolvedRole()
			// An entry with no name cannot be pointed at, and one with no value
			// has no place on an ordered board. Dropping is the repair.
			if e.Label != "" && e.Value > 0 && len(out) < max {
				out = append(out, e)
			}
		}
		return out
	}
	r.Rows = clean(r.Rows, maxRankingRows, false)
	r.Arrivals = clean(r.Arrivals, maxRankingArrivals, true)

	// The board is descending by construction rather than by instruction. Which
	// order the model happened to list the rows in is not a claim about the
	// subject — the values are — so sorting is a mechanical repair and not
	// something worth a correction round.
	sort.SliceStable(r.Rows, func(i, j int) bool { return r.Rows[i].Value > r.Rows[j].Value })

	for i := range p.Beats {
		b := p.Beats[i].Ranking
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "insert" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(r.Arrivals); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

// rankingOrder is the board after the first `arrivals` arrivals have landed:
// every entry revealed so far, sorted descending, truncated to the board height.
//
// In Go rather than in the renderer because "which rows are visible and in what
// order" is the template's entire logic, it wants unit tests, and two
// implementations of it — one deciding the validator's answer and one deciding
// what is drawn — would eventually disagree about a tie.
func rankingOrder(r *RankingSpec, arrivals int) []int {
	type slot struct {
		index int
		value float64
	}
	slots := make([]slot, 0, len(r.Rows)+arrivals)
	for i, row := range r.Rows {
		slots = append(slots, slot{index: i, value: row.Value})
	}
	for i := 0; i < arrivals && i < len(r.Arrivals); i++ {
		slots = append(slots, slot{index: len(r.Rows) + i, value: r.Arrivals[i].Value})
	}
	// Stable, and ties break toward the entry that was already on the board: a
	// new row equal to an existing one has not beaten it.
	sort.SliceStable(slots, func(i, j int) bool { return slots[i].value > slots[j].value })

	height := len(r.Rows)
	if height > maxRankingRows {
		height = maxRankingRows
	}
	if len(slots) < height {
		height = len(slots)
	}
	out := make([]int, height)
	for i := range out {
		out[i] = slots[i].index
	}
	return out
}

// sameOrder reports whether two boards are identical, position for position.
func sameOrder(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func validateRankingPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Ranking: true}); err != nil {
		return err
	}

	r := p.Ranking
	if r == nil {
		return fmt.Errorf("the plan has no board — this template is one ordered board and the things that land on it")
	}
	if strings.TrimSpace(r.Metric) == "" {
		return fmt.Errorf("the board has no metric — say what it is ordered by. A ranked list nobody has told you the axis of is a list of names")
	}
	if n := len(r.Rows); n < minRankingRows || n > maxRankingRows {
		return fmt.Errorf("the board has %d rows, want %d-%d. Below four a position means nothing because there is no board to move through; past eight a single row sliding is lost among the ones that did not",
			n, minRankingRows, maxRankingRows)
	}
	if n := len(r.Arrivals); n < minRankingArrivals || n > maxRankingArrivals {
		return fmt.Errorf("there are %d arrivals, want %d-%d. Each one needs its own beat to be read, and four is a montage nobody follows",
			n, minRankingArrivals, maxRankingArrivals)
	}

	seen := map[string]bool{}
	for _, group := range [][]RankingEntry{r.Rows, r.Arrivals} {
		for i, e := range group {
			if strings.TrimSpace(e.Label) == "" {
				return fmt.Errorf("entry %d has no label — a row nobody can name cannot be pointed at", i)
			}
			if e.Value <= 0 {
				return fmt.Errorf("entry %q has value %v; the board is ordered by a positive quantity", e.Label, e.Value)
			}
			key := strings.ToLower(strings.TrimSpace(e.Label))
			if seen[key] {
				return fmt.Errorf("two entries are both %q — the board cannot hold the same name twice, and an arrival that is already on it is an update rather than an insert", e.Label)
			}
			seen[key] = true
		}
	}
	for i, a := range r.Arrivals {
		if role := strings.ToLower(strings.TrimSpace(a.Role)); role != "" && !metricRoles[role] {
			return fmt.Errorf("arrival %d has role %q, which is not one of: %s", i, a.Role, strings.Join(MetricRoles(), ", "))
		}
	}

	// The board is read before anything moves it.
	if p.Beats[0].Ranking == nil || p.Beats[0].Ranking.ResolvedShow() != "board" {
		return fmt.Errorf("beat %q does not establish the board. A row moving on a board the viewer has not read yet is a row moving for no reason",
			p.Beats[0].ID)
	}

	inserted := map[int]bool{}
	boards := 0
	for i, b := range p.Beats {
		if b.Ranking == nil {
			return fmt.Errorf("beat %q has no ranking direction — every beat draws the board, lands an arrival, or reads the result", b.ID)
		}
		switch b.Ranking.ResolvedShow() {
		case "board":
			boards++
			if i != 0 {
				return fmt.Errorf("beat %q draws the board again part-way through. The board is established once, at the start", b.ID)
			}
		case "insert":
			if b.Ranking.At < 0 || b.Ranking.At >= len(r.Arrivals) {
				return fmt.Errorf("beat %q lands arrival %d, which does not exist", b.ID, b.Ranking.At)
			}
			if inserted[b.Ranking.At] {
				return fmt.Errorf("beat %q lands arrival %d again; each entry arrives once", b.ID, b.Ranking.At)
			}
			inserted[b.Ranking.At] = true
		}
	}
	if boards != 1 {
		return fmt.Errorf("there are %d beats establishing the board, want exactly 1", boards)
	}
	if len(inserted) != len(r.Arrivals) {
		return fmt.Errorf("%d of the %d arrivals never land. An entry the narrator skips is one nobody saw — give it a beat or cut it",
			len(r.Arrivals)-len(inserted), len(r.Arrivals))
	}

	// The rule the template exists for. An arrival that does not place lands
	// below the last visible row, so the picture does not move — which reads as
	// a broken render rather than as a small number.
	for i := range r.Arrivals {
		before := rankingOrder(r, i)
		after := rankingOrder(r, i+1)
		if sameOrder(before, after) {
			return fmt.Errorf("arrival %q scores %v, which does not place on a board whose lowest visible row is higher. The whole picture is a row moving, so an arrival that changes nothing on screen reads as a broken render — give it a value that places, or make the point with a different template",
				r.Arrivals[i].Label, r.Arrivals[i].Value)
		}
	}
	return nil
}

// rankingScenes lays the clip out as ONE scene. The board persists and the
// beats only say which orderings it passes through, so Go hands the renderer the
// orderings and the renderer animates between them.
func rankingScenes(in SnippetSceneInput) ([]Scene, error) {
	r := in.Plan.Ranking
	if r == nil {
		return nil, fmt.Errorf("the plan has no board")
	}

	// One flat entry list: starting rows first, then arrivals, which is the
	// indexing rankingOrder returns.
	entries := make([]map[string]any, 0, len(r.Rows)+len(r.Arrivals))
	for _, e := range r.Rows {
		entries = append(entries, map[string]any{
			"label": e.Label, "value": e.Value, "role": "neutral", "arrival": false,
		})
	}
	for _, e := range r.Arrivals {
		entries = append(entries, map[string]any{
			"label": e.Label, "value": e.Value, "note": e.Note,
			"role": e.ResolvedRole(), "arrival": true,
		})
	}

	landed := 0
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Ranking == nil {
			return nil, fmt.Errorf("beat %q has no ranking direction", beat.ID)
		}
		show := beat.Ranking.ResolvedShow()
		if show == "insert" {
			landed++
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			// The board as it stands during this beat. Precomputed so the
			// renderer never sorts: a second sort would be a second answer.
			"order": rankingOrder(r, landed),
		}
		if show == "insert" {
			step["at"] = beat.Ranking.At
			// Which flat index just landed, so the renderer can light it.
			step["entered"] = len(r.Rows) + beat.Ranking.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneRanking,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":   in.Plan.Title,
			"metric":  r.Metric,
			"unit":    r.Unit,
			"entries": entries,
			"steps":   steps,
		}),
	}}, nil
}
