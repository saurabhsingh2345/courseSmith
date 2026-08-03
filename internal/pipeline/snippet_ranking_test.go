package pipeline

import (
	"strings"
	"testing"
)

const rkNarration = "One write lands, and the board is in order again before the call returns."

func rankingPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "ranking",
		Title:    "One write, and the board re-sorts",
		Ranking: &RankingSpec{
			Metric: "score",
			Rows: []RankingEntry{
				{Label: "shadowwolf", Value: 9842},
				{Label: "neon_blade", Value: 9610},
				{Label: "kira_07", Value: 9354},
				{Label: "mochi", Value: 9128},
				{Label: "vortex", Value: 8901},
			},
			Arrivals: []RankingEntry{
				{Label: "phoenix", Value: 9501, Note: "The write and the re-sort are one operation", Role: "quantity"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "board", Heading: "Five players", Narration: rkNarration, Ranking: &RankingBeat{Show: "board"}},
			{ID: "write", Heading: "A score arrives", Narration: rkNarration, Ranking: &RankingBeat{Show: "insert", At: 0}},
			{ID: "cost", Heading: "What that cost", Narration: rkNarration, Ranking: &RankingBeat{Show: "read"}},
		},
	}
	p.targetWords = 3 * 40
	return p
}

func TestRankingPlanAccepted(t *testing.T) {
	if err := validateRankingPlan(rankingPlan()); err != nil {
		t.Fatalf("a well-formed ranking plan was rejected: %v", err)
	}
}

// The rule the template exists for. An arrival that lands below the board does
// not move anything, so the narrator says "and this one comes in" over a picture
// that sits still — which reads as a broken render rather than as a small score.
func TestRankingRejectsAnArrivalThatDoesNotPlace(t *testing.T) {
	p := rankingPlan()
	p.Ranking.Arrivals[0].Value = 12
	err := validateRankingPlan(p)
	if err == nil {
		t.Fatal("an arrival that changes nothing on screen was accepted")
	}
	if !strings.Contains(err.Error(), "phoenix") {
		t.Fatalf("the error does not name the arrival: %v", err)
	}
}

// The same rule on the second arrival, which is the case a single-arrival test
// misses: the board has already moved once, so "does it place" has to be asked
// against the board as it stands, not as it started.
func TestRankingChecksEachArrivalAgainstTheBoardAsItStands(t *testing.T) {
	p := rankingPlan()
	p.Ranking.Arrivals = append(p.Ranking.Arrivals, RankingEntry{
		Label: "straggler", Value: 9000, Role: "neutral",
	})
	p.Beats = append(p.Beats, SnippetBeat{
		ID: "second", Heading: "And another", Narration: rkNarration,
		Ranking: &RankingBeat{Show: "insert", At: 1},
	})
	// 9000 clears the original bottom row (8901) but not the bottom of the board
	// once phoenix has pushed vortex off, so it places against the starting
	// board and does not place against the real one.
	if err := validateRankingPlan(p); err == nil {
		t.Fatal("an arrival judged against the starting board rather than the current one was accepted")
	}
}

func TestRankingRequiresTheBoardFirst(t *testing.T) {
	p := rankingPlan()
	p.Beats[0].Ranking = &RankingBeat{Show: "insert", At: 0}
	p.Beats[1].Ranking = &RankingBeat{Show: "board"}
	if err := validateRankingPlan(p); err == nil {
		t.Fatal("a clip that lands an arrival before establishing the board was accepted")
	}
}

func TestRankingRejectsADuplicateName(t *testing.T) {
	p := rankingPlan()
	p.Ranking.Arrivals[0].Label = "mochi"
	err := validateRankingPlan(p)
	if err == nil {
		t.Fatal("an arrival already on the board was accepted")
	}
	if !strings.Contains(err.Error(), "update") {
		t.Fatalf("the error does not explain why it is wrong: %v", err)
	}
}

func TestRankingRejectsAnUnreadableBoard(t *testing.T) {
	p := rankingPlan()
	p.Ranking.Rows = p.Ranking.Rows[:3]
	if err := validateRankingPlan(p); err == nil {
		t.Fatal("a three-row board was accepted")
	}
}

func TestRankingRequiresAMetric(t *testing.T) {
	p := rankingPlan()
	p.Ranking.Metric = ""
	if err := validateRankingPlan(p); err == nil {
		t.Fatal("a board with no stated axis was accepted")
	}
}

func TestRankingRequiresEveryArrivalToLand(t *testing.T) {
	p := rankingPlan()
	p.Ranking.Arrivals = append(p.Ranking.Arrivals, RankingEntry{Label: "late", Value: 9700})
	if err := validateRankingPlan(p); err == nil {
		t.Fatal("an arrival that no beat ever lands was accepted")
	}
}

// Which order the model listed the rows in is not a claim about the subject —
// the values are — so sorting is a mechanical repair rather than a round.
func TestRankingNormalizeSortsTheBoardDescending(t *testing.T) {
	p := rankingPlan()
	p.Ranking.Rows = []RankingEntry{
		{Label: "low", Value: 100},
		{Label: "high", Value: 900},
		{Label: "mid", Value: 500},
		{Label: "lowest", Value: 50},
	}
	normalizeRankingPlan(p)
	got := make([]string, len(p.Ranking.Rows))
	for i, r := range p.Ranking.Rows {
		got[i] = r.Label
	}
	want := []string{"high", "mid", "low", "lowest"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("board sorted to %v, want %v", got, want)
		}
	}
}

// rankingOrder is the single answer to "which rows are visible and in what
// order". The validator and the renderer both read it, so a tie has to break the
// same way for both.
func TestRankingOrderBreaksTiesTowardTheIncumbent(t *testing.T) {
	r := &RankingSpec{
		Rows: []RankingEntry{
			{Label: "a", Value: 900},
			{Label: "b", Value: 500},
			{Label: "c", Value: 400},
			{Label: "d", Value: 300},
		},
		Arrivals: []RankingEntry{{Label: "tie", Value: 500}},
	}
	order := rankingOrder(r, 1)
	// The incumbent at 500 is index 1; the arrival is index 4. Equal scores
	// leave the incumbent ahead, because a new row that matches has not beaten
	// anything.
	if order[1] != 1 {
		t.Fatalf("order is %v, want the incumbent (index 1) still in second", order)
	}
}

func TestRankingScenesCarryOneOrderPerBeat(t *testing.T) {
	p := rankingPlan()
	scenes, err := rankingScenes(sceneInput(t, p, 12000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if len(steps) != len(p.Beats) {
		t.Fatalf("want %d steps, got %d", len(p.Beats), len(steps))
	}
	for i, s := range steps {
		if _, ok := s["order"].([]int); !ok {
			t.Fatalf("step %d carries no order for the renderer to draw", i)
		}
	}
	// The insert beat names which flat index arrived, so the renderer can light
	// it without re-deriving anything.
	if steps[1]["entered"] != len(p.Ranking.Rows) {
		t.Fatalf("insert beat entered %v, want %d", steps[1]["entered"], len(p.Ranking.Rows))
	}
}
