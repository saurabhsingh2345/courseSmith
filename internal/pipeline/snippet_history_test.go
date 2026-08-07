package pipeline

import (
	"strings"
	"testing"
)

const gitNarration = "The disc lands on its lane and an edge draws back to the commit it grew out of."

func historyPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "history",
		Title:    "History is a graph not a line",
		History: &HistorySpec{
			Lanes: []string{"main", "feature"},
			Commits: []HistoryCommit{
				{Lane: 0, Label: "initial commit"},
				{Lane: 0, Label: "add readme", Parents: []int{0}},
				{Lane: 1, Label: "start the feature", Parents: []int{1}},
				{Lane: 0, Label: "fix a typo", Parents: []int{1}},
				{Lane: 0, Label: "merge the feature", Parents: []int{3, 2}},
			},
		},
		Beats: []SnippetBeat{
			{ID: "lanes", Heading: "Two lanes", Narration: gitNarration, History: &HistoryBeat{Show: "graph"}},
			{ID: "root", Heading: "The first commit", Narration: gitNarration, History: &HistoryBeat{Show: "commit", At: 0}},
			{ID: "readme", Heading: "Another commit", Narration: gitNarration, History: &HistoryBeat{Show: "commit", At: 1}},
			{ID: "feature", Heading: "A new lane", Narration: gitNarration, History: &HistoryBeat{Show: "commit", At: 2}},
			{ID: "typo", Heading: "Main moves on", Narration: gitNarration, History: &HistoryBeat{Show: "commit", At: 3}},
			{ID: "the-fork", Heading: "Where it split", Narration: gitNarration, History: &HistoryBeat{Show: "branch", At: 1}},
			{ID: "join", Heading: "Bringing it back", Narration: gitNarration, History: &HistoryBeat{Show: "merge", At: 4}},
			{ID: "the-log", Heading: "The whole graph", Narration: gitNarration, History: &HistoryBeat{Show: "log"}},
		},
	}
	// The template's ideal is 28 words per beat, so the fixture budget is sized
	// against that rather than the shared 40.
	p.targetWords = 8 * 28
	return p
}

func TestHistoryPlanAccepted(t *testing.T) {
	if err := validateHistoryPlan(historyPlan()); err != nil {
		t.Fatalf("a well-formed commit graph was rejected: %v", err)
	}
}

// The family's signature rule, applied to a graph: the DAG is walked in Go, and
// a second root is rejected with the extra one named rather than left to hunt.
func TestHistoryRejectsMoreThanOneRoot(t *testing.T) {
	p := historyPlan()
	p.History.Commits[2].Parents = nil
	err := validateHistoryPlan(p)
	if err == nil {
		t.Fatal("a graph with two parentless commits was accepted, which is two repositories")
	}
	if !strings.Contains(err.Error(), "start the feature") {
		t.Fatalf("the error does not name the extra root: %v", err)
	}
	if strings.Contains(err.Error(), "initial commit") {
		t.Fatalf("the error blames the real root as well as the extra one: %v", err)
	}
}

func TestHistoryRejectsAParentThatComesLater(t *testing.T) {
	p := historyPlan()
	p.History.Commits[1].Parents = []int{3}
	err := validateHistoryPlan(p)
	if err == nil {
		t.Fatal("an edge pointing backwards in time was accepted")
	}
	if !strings.Contains(err.Error(), "add readme") || !strings.Contains(err.Error(), "fix a typo") {
		t.Fatalf("the error does not quote both ends of the bad edge: %v", err)
	}
}

func TestHistoryRejectsAMergeWhoseParentsShareALane(t *testing.T) {
	p := historyPlan()
	p.History.Commits[2].Lane = 0
	err := validateHistoryPlan(p)
	if err == nil {
		t.Fatal("a merge whose two parents sit on the same lane was accepted, and it joins nothing")
	}
	if !strings.Contains(err.Error(), "merge the feature") || !strings.Contains(err.Error(), "main") {
		t.Fatalf("the error does not quote the merge and the shared lane: %v", err)
	}
}

func TestHistoryRejectsALaneIndexOutOfRange(t *testing.T) {
	p := historyPlan()
	p.History.Commits[2].Lane = 5
	if err := validateHistoryPlan(p); err == nil {
		t.Fatal("a commit on a lane that does not exist was accepted")
	}
}

func TestHistoryRequiresOpeningOnTheGraph(t *testing.T) {
	p := historyPlan()
	p.Beats[0].History = &HistoryBeat{Show: "commit", At: 0}
	if err := validateHistoryPlan(p); err == nil {
		t.Fatal("a commit landing before the lanes were shown was accepted")
	}
}

func TestHistoryRequiresClosingOnTheLog(t *testing.T) {
	p := historyPlan()
	p.Beats[7].History = &HistoryBeat{Show: "branch", At: 1}
	err := validateHistoryPlan(p)
	if err == nil {
		t.Fatal("a clip that never lights the whole graph was accepted")
	}
	if !strings.Contains(err.Error(), "log") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHistoryRejectsCommitsLandingOutOfOrder(t *testing.T) {
	p := historyPlan()
	p.Beats[2].History = &HistoryBeat{Show: "commit", At: 2}
	err := validateHistoryPlan(p)
	if err == nil {
		t.Fatal("a commit landing before its parent was accepted")
	}
	if !strings.Contains(err.Error(), "start the feature") || !strings.Contains(err.Error(), "add readme") {
		t.Fatalf("the error does not quote the commit landed and the one that was due: %v", err)
	}
}

func TestHistoryRejectsAMergeBeatOnASingleParentCommit(t *testing.T) {
	p := historyPlan()
	p.Beats[3].History = &HistoryBeat{Show: "merge", At: 2}
	err := validateHistoryPlan(p)
	if err == nil {
		t.Fatal("a merge beat on a commit with one parent was accepted")
	}
	if !strings.Contains(err.Error(), "start the feature") {
		t.Fatalf("the error does not name the commit: %v", err)
	}
}

func TestHistoryRejectsACommitBeatOnAMerge(t *testing.T) {
	p := historyPlan()
	p.Beats[6].History = &HistoryBeat{Show: "commit", At: 4}
	err := validateHistoryPlan(p)
	if err == nil {
		t.Fatal("a two-parent commit landed as an ordinary commit was accepted")
	}
	if !strings.Contains(err.Error(), "merge the feature") {
		t.Fatalf("the error does not name the merge: %v", err)
	}
}

func TestHistoryRejectsABranchBeatWhereNothingForks(t *testing.T) {
	p := historyPlan()
	p.Beats[5].History = &HistoryBeat{Show: "branch", At: 0}
	err := validateHistoryPlan(p)
	if err == nil {
		t.Fatal("a branch beat on a commit with one child was accepted, and one edge is not a fork")
	}
	if !strings.Contains(err.Error(), "initial commit") {
		t.Fatalf("the error does not name the commit: %v", err)
	}
}

func TestHistoryRejectsLeavingACommitUnlanded(t *testing.T) {
	p := historyPlan()
	p.History.Commits = append(p.History.Commits, HistoryCommit{Lane: 0, Label: "final touch", Parents: []int{4}})
	err := validateHistoryPlan(p)
	if err == nil {
		t.Fatal("a graph whose closing log lights a commit no beat landed was accepted")
	}
	if !strings.Contains(err.Error(), "5 of 6") {
		t.Fatalf("the error does not count the commits that landed: %v", err)
	}
}

// Over-long labels, a stray lane and a repeated parent are bookkeeping, not
// wrong answers, so they are repaired rather than argued with.
func TestHistoryNormalizeRepairsLabelsAndIndices(t *testing.T) {
	p := historyPlan()
	p.History.Lanes[1] = "the   long running feature branch"
	p.History.Commits[0].Label = "the very first commit of all"
	p.History.Commits[3].Lane = 9
	p.History.Commits[4].Parents = []int{3, 3, 2}
	p.Beats[1].History.At = 42
	normalizeHistoryPlan(p)

	if n := len(strings.Fields(p.History.Lanes[1])); n != maxHistoryLaneWords {
		t.Fatalf("the lane name normalized to %d words, want %d", n, maxHistoryLaneWords)
	}
	if n := len(strings.Fields(p.History.Commits[0].Label)); n != maxHistoryLabelWords {
		t.Fatalf("the commit label normalized to %d words, want %d", n, maxHistoryLabelWords)
	}
	if p.History.Commits[3].Lane != 1 {
		t.Fatalf("an out-of-range lane normalized to %d, want the last lane", p.History.Commits[3].Lane)
	}
	if got := p.History.Commits[4].Parents; len(got) != 2 || got[0] != 3 || got[1] != 2 {
		t.Fatalf("the repeated parent survived normalize: %v", got)
	}
	if p.Beats[1].History.At != 4 {
		t.Fatalf("an out-of-range commit index normalized to %d, want the last commit", p.Beats[1].History.At)
	}
}

func TestHistoryShowDefaultsToCommit(t *testing.T) {
	b := HistoryBeat{Show: "rebase"}
	if got := b.ResolvedShow(); got != "commit" {
		t.Fatalf("an unknown show resolved to %q, want commit", got)
	}
}

// The component draws a graph it was handed: edges, lane crossings and the
// landed set all arrive precomputed.
func TestHistoryScenesAccumulateTheGraph(t *testing.T) {
	p := historyPlan()
	scenes, err := historyScenes(sceneInput(t, p, 4000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	props := scenes[0].Props

	commits, _ := props["commits"].([]map[string]any)
	if len(commits) != 5 {
		t.Fatalf("want 5 commits, got %d", len(commits))
	}
	if commits[4]["merge"] != true || commits[0]["merge"] != false {
		t.Fatalf("the merge flag is on the wrong disc: %v", commits)
	}
	if kids, _ := commits[1]["children"].([]int); len(kids) != 2 || kids[0] != 2 || kids[1] != 3 {
		t.Fatalf("commit 1 forks into %v, want commits 2 and 3", commits[1]["children"])
	}

	edges, _ := props["edges"].([]map[string]any)
	if len(edges) != 5 {
		t.Fatalf("want 5 edges, got %d", len(edges))
	}
	curved := 0
	for _, e := range edges {
		if e["curved"] == true {
			curved++
		}
	}
	if curved != 2 {
		t.Fatalf("want 2 lane-changing edges, got %d", curved)
	}

	steps, _ := props["steps"].([]map[string]any)
	first := steps[0]
	if first["show"] != "graph" {
		t.Fatalf("the first step shows %v, want graph", first["show"])
	}
	if up, _ := first["landed"].([]int); len(up) != 0 {
		t.Fatalf("nothing has landed on the opener, but the step reports %v", up)
	}
	if first["head"] != -1 {
		t.Fatalf("HEAD points at %v before any commit landed", first["head"])
	}

	last := steps[len(steps)-1]
	if last["show"] != "log" {
		t.Fatalf("the last step shows %v, want log", last["show"])
	}
	up, _ := last["landed"].([]int)
	if len(up) != 5 || up[0] != 0 || up[4] != 4 {
		t.Fatalf("the log lights %v, want every commit sorted", up)
	}
	if last["head"] != 4 {
		t.Fatalf("HEAD ends on %v, want the newest commit", last["head"])
	}
}
