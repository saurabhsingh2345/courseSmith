package pipeline

// The history template: the commit graph, drawn as the graph it actually is.
//
// Version control is the first tool a career-switcher is handed and the last
// one they understand, and the reason is that every explanation of it is
// prose. "A branch is a pointer" and "a merge joins two histories" are true
// sentences that leave no picture behind, so the learner builds their own,
// and the one they build is a LIST — commits in a row, a branch as a detour,
// a merge as a thing that happens to files. Almost every git mistake a
// beginner makes is that list model failing.
//
// So this template draws the directed acyclic graph and nothing else: lanes as
// hairlines, commits as discs on them, an edge from every commit back to each
// parent, and HEAD as a chip that rides the newest commit. A branch is visible
// as two edges leaving one disc. A merge is visible as one disc with two edges
// arriving from two different lanes. Nothing has to be asserted, because it is
// all on screen.
//
// The validator is a graph checker, because a DAG drawn wrong teaches the list
// model harder than saying nothing. It insists on exactly ONE root, since a
// second parentless commit is a second repository and the extra one is named
// in the rejection rather than left to be hunted. It insists every parent
// index is strictly less than its child's, because history flows forward and
// an edge pointing the other way is a cycle a viewer will read as time
// travel. And it insists a two-parent commit's parents sit on DIFFERENT lanes:
// a merge whose parents share a lane joined nothing, which is a picture of the
// exact misconception the clip exists to kill.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "history",
		Category:    CatSystems,
		Since:       SinceV7,
		Family:      FamilyFoundations,
		Title:       "The commit graph",
		Description: "The git DAG drawn honestly: lanes, commits landing one at a time, a branch leaving as two edges, a merge arriving as two, and HEAD riding the newest commit. Reach for it when the subject is version control's shape — what a branch really is, what a merge really joins, why history is a graph and not a list.",
		Example:     "Branch, commit, merge: one feature's life in the git graph",
		PromptFile:  snippetHistoryTemplateName,
		NeedsCode:   false,
		// Eight commits landing one at a time, plus the opener and the log:
		// under thirty-five seconds each disc gets barely three seconds and the
		// edges have no time to draw.
		MinTargetSec:     35,
		DefaultTargetSec: 55,
		// The opener, eight commits, and the closing log. That is the whole
		// graph at its widest, and anything past it is the same picture again.
		MaxBeats: 10,
		// A beat here is a SHOT — one disc landing, one pair of edges lighting
		// — not a step in an argument.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{History: true},
		OwnsPlan:          planFields{History: true},
		Normalize:         normalizeHistoryPlan,
		Validate:          validateHistoryPlan,
		Scenes:            historyScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(HistoryShows(), ", "),
				"MinLanes":      minHistoryLanes,
				"MaxLanes":      maxHistoryLanes,
				"MinCommits":    minHistoryCommits,
				"MaxCommits":    maxHistoryCommits,
				"MaxLaneWords":  maxHistoryLaneWords,
				"MaxLabelWords": maxHistoryLabelWords,
			}
		},
	})
}

const snippetHistoryTemplateName = "snippet_history.tmpl"

const (
	// One lane is a straight history, which is a legitimate clip — "this is
	// what main looks like before anybody branches". Three is the most lanes
	// that fit at a disc size where the labels underneath stay readable.
	minHistoryLanes = 1
	maxHistoryLanes = 3

	// Below three commits there is no history, only a before and an after.
	// Past eight the discs crowd and the labels collide, and eight is also the
	// most a MaxBeats of ten can land one at a time.
	minHistoryCommits = 3
	maxHistoryCommits = 8

	// A lane is a ref name — "main", "feature/login".
	maxHistoryLaneWords = 3
	// A commit label is a commit message in the register people actually use:
	// "add login form", not a sentence about it.
	maxHistoryLabelWords = 4
	// Two parents is a merge. Three is an octopus merge, which is real, rare,
	// and undrawable at this size without the edges crossing every lane.
	maxHistoryParents = 2
)

// historyShows is the closed vocabulary of what a beat does.
var historyShows = map[string]bool{
	// The empty lanes and the story so far. The opener.
	"graph": true,
	// Commit At lands on its lane, its edges draw, and HEAD moves to it.
	"commit": true,
	// A dwell on a divergence: the two child edges leaving commit At light.
	"branch": true,
	// A two-parent commit At lands and both of its edges draw at once.
	"merge": true,
	// The closer: the whole graph lit, lanes labelled.
	"log": true,
}

// HistoryShows returns the beat vocabulary sorted.
func HistoryShows() []string {
	out := make([]string, 0, len(historyShows))
	for k := range historyShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// HistorySpec is the graph. On the plan because the DAG persists for the whole
// clip — beats reveal it, they do not change it.
type HistorySpec struct {
	// Lanes are the ref names, left-labelled: "main", "feature".
	Lanes []string `json:"lanes"`
	// Commits are the nodes, oldest first. Index order IS time order.
	Commits []HistoryCommit `json:"commits"`
}

// HistoryCommit is one node of the DAG.
type HistoryCommit struct {
	// Lane indexes HistorySpec.Lanes.
	Lane int `json:"lane"`
	// Label is the commit message — "add login form".
	Label string `json:"label"`
	// Parents are the indices of the commits this one descends from. Empty for
	// the root; one for an ordinary commit; two for a merge.
	Parents []int `json:"parents,omitempty"`
}

// HistoryBeat is one shot of the graph.
type HistoryBeat struct {
	// Show is a historyShows name.
	Show string `json:"show"`
	// At indexes HistorySpec.Commits — the commit that lands, or the commit a
	// branch beat dwells on.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults the unknown to a commit landing, which is what most
// beats of this template are.
func (b HistoryBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if historyShows[s] {
		return s
	}
	return "commit"
}

func normalizeHistoryPlan(p *SnippetPlan) {
	h := p.History
	if h == nil {
		return
	}
	lanes := make([]string, 0, len(h.Lanes))
	for _, l := range h.Lanes {
		if len(lanes) < maxHistoryLanes {
			lanes = append(lanes, clampWords(collapseSpaces(l), maxHistoryLaneWords))
		}
	}
	h.Lanes = lanes

	commits := make([]HistoryCommit, 0, len(h.Commits))
	for _, c := range h.Commits {
		c.Label = clampWords(collapseSpaces(c.Label), maxHistoryLabelWords)
		if c.Lane < 0 {
			c.Lane = 0
		}
		if n := len(h.Lanes); n > 0 && c.Lane >= n {
			c.Lane = n - 1
		}
		// Duplicate parents are a merge with itself, which draws one edge and
		// reads as a bug; drop them rather than argue.
		parents := make([]int, 0, len(c.Parents))
		seen := map[int]bool{}
		for _, pi := range c.Parents {
			if seen[pi] || len(parents) >= maxHistoryParents {
				continue
			}
			seen[pi] = true
			parents = append(parents, pi)
		}
		c.Parents = parents
		if len(commits) < maxHistoryCommits {
			commits = append(commits, c)
		}
	}
	h.Commits = commits

	for i := range p.Beats {
		b := p.Beats[i].History
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		switch b.Show {
		case "commit", "merge", "branch":
			if b.At < 0 {
				b.At = 0
			}
			if n := len(h.Commits); n > 0 && b.At >= n {
				b.At = n - 1
			}
		default:
			b.At = 0
		}
	}
}

func validateHistoryPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{History: true}); err != nil {
		return err
	}

	h := p.History
	if h == nil {
		return fmt.Errorf("the plan has no history — this template is a commit graph, so the lanes and commits are the clip")
	}
	if n := len(h.Lanes); n < minHistoryLanes || n > maxHistoryLanes {
		return fmt.Errorf("the graph has %d lanes, want %d-%d. A lane is a ref the viewer can name, and past %d the discs shrink until their labels collide",
			n, minHistoryLanes, maxHistoryLanes, maxHistoryLanes)
	}
	for i, l := range h.Lanes {
		if strings.TrimSpace(l) == "" {
			return fmt.Errorf("lane %d has no name. An unnamed lane is a hairline with commits on it, and the whole point of a lane is that it is a branch somebody can check out", i)
		}
	}
	if n := len(h.Commits); n < minHistoryCommits || n > maxHistoryCommits {
		return fmt.Errorf("the graph has %d commits, want %d-%d. Below three there is no history, only a before and an after — timeline draws that better; past %d the discs crowd and there are not enough beats to land them one at a time",
			n, minHistoryCommits, maxHistoryCommits, maxHistoryCommits)
	}

	// THE DAG, checked as a graph. Every rule below is a way the picture can be
	// drawn wrong while reading as plausible prose.
	for i, c := range h.Commits {
		if strings.TrimSpace(c.Label) == "" {
			return fmt.Errorf("commit %d has no label. A disc with nothing written under it is a dot, and the viewer cannot tell one dot from another", i)
		}
		if c.Lane < 0 || c.Lane >= len(h.Lanes) {
			return fmt.Errorf("commit %d (%q) is on lane %d, which does not exist — the graph has lanes 0-%d", i, c.Label, c.Lane, len(h.Lanes)-1)
		}
		if len(c.Parents) > maxHistoryParents {
			return fmt.Errorf("commit %d (%q) has %d parents. Two parents is a merge and draws as two edges arriving; more is an octopus merge, which is real but cannot be drawn at this size without the edges crossing every lane",
				i, c.Label, len(c.Parents))
		}
		for _, pi := range c.Parents {
			if pi < 0 || pi >= len(h.Commits) {
				return fmt.Errorf("commit %d (%q) claims parent %d, which does not exist — the graph has commits 0-%d", i, c.Label, pi, len(h.Commits)-1)
			}
			if pi >= i {
				return fmt.Errorf("commit %d (%q) claims commit %d (%q) as a parent, but a parent has to come earlier. History flows forward, so every parent index is strictly less than its child's — an edge pointing the other way is a cycle, and a viewer reads it as time travel",
					i, c.Label, pi, h.Commits[pi].Label)
			}
		}
	}
	roots := []int{}
	for i, c := range h.Commits {
		if len(c.Parents) == 0 {
			roots = append(roots, i)
		}
	}
	if len(roots) == 0 {
		return fmt.Errorf("every commit has a parent, so the graph has no root. Some commit was the first one — give it an empty \"parents\" list")
	}
	if len(roots) > 1 {
		extra := make([]string, 0, len(roots)-1)
		for _, i := range roots[1:] {
			extra = append(extra, fmt.Sprintf("commit %d (%q)", i, h.Commits[i].Label))
		}
		return fmt.Errorf("the graph has %d parentless commits, and a repository has exactly one root. %s float free with no edge back to anything — give each of them a parent",
			len(roots), strings.Join(extra, " and "))
	}
	for i, c := range h.Commits {
		if len(c.Parents) != maxHistoryParents {
			continue
		}
		a, b := h.Commits[c.Parents[0]], h.Commits[c.Parents[1]]
		if a.Lane == b.Lane {
			return fmt.Errorf("commit %d (%q) is a merge whose two parents, %q and %q, are both on lane %d (%q). A merge joins two lanes — if both parents already sit on the same one, nothing was joined and the picture shows two edges arriving from the same hairline",
				i, c.Label, a.Label, b.Label, a.Lane, h.Lanes[a.Lane])
		}
	}

	children := make([][]int, len(h.Commits))
	for i, c := range h.Commits {
		for _, pi := range c.Parents {
			children[pi] = append(children[pi], i)
		}
	}

	if p.Beats[0].History == nil || p.Beats[0].History.ResolvedShow() != "graph" {
		return fmt.Errorf("beat %q does not open on the graph. A commit landing on lanes the viewer has not seen is a disc appearing in the dark — open with {\"show\": \"graph\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.History == nil || last.History.ResolvedShow() != "log" {
		return fmt.Errorf("beat %q does not close on the log. The last frame is the whole graph lit with its lane names, which is the picture somebody keeps — end with {\"show\": \"log\"}", last.ID)
	}

	landed := 0
	seen := map[int]bool{}
	for _, b := range p.Beats {
		d := b.History
		if d == nil {
			return fmt.Errorf("beat %q has no history direction — every beat shows the graph, lands a commit, dwells on a branch, lands a merge, or lights the log", b.ID)
		}
		show := d.ResolvedShow()
		if show == "graph" || show == "log" {
			continue
		}
		if d.At < 0 || d.At >= len(h.Commits) {
			return fmt.Errorf("beat %q acts on commit %d, which does not exist — the graph has commits 0-%d", b.ID, d.At, len(h.Commits)-1)
		}
		c := h.Commits[d.At]
		switch show {
		case "commit", "merge":
			if d.At != landed {
				return fmt.Errorf("beat %q lands commit %d (%q) when commit %d (%q) is the next one due. Commits land in index order, oldest first — a commit arriving before its parent draws an edge into empty space",
					b.ID, d.At, c.Label, landed, h.Commits[landed].Label)
			}
			if show == "merge" && len(c.Parents) != maxHistoryParents {
				return fmt.Errorf("beat %q calls commit %d (%q) a merge, but it has %d parent(s). A merge is a commit with two parents, and the shot is both edges arriving at once — use {\"show\": \"commit\"} for this one",
					b.ID, d.At, c.Label, len(c.Parents))
			}
			if show == "commit" && len(c.Parents) == maxHistoryParents {
				return fmt.Errorf("beat %q lands commit %d (%q) as an ordinary commit, but it has two parents, which makes it a merge. Give it {\"show\": \"merge\"} so both edges draw together — that arrival IS what a merge looks like",
					b.ID, d.At, c.Label)
			}
			seen[d.At] = true
			landed++
		case "branch":
			if len(children[d.At]) < 2 {
				return fmt.Errorf("beat %q dwells on a branch at commit %d (%q), but %d commit(s) descend from it. A branch is visible as TWO edges leaving one disc, so point this beat at the commit the history actually forks from",
					b.ID, d.At, c.Label, len(children[d.At]))
			}
			for _, kid := range children[d.At] {
				if !seen[kid] {
					return fmt.Errorf("beat %q highlights the fork at commit %d (%q) before commit %d (%q) has landed. There is nothing to see until both children are on screen — move this beat after them",
						b.ID, d.At, c.Label, kid, h.Commits[kid].Label)
				}
			}
		}
	}
	if landed != len(h.Commits) {
		return fmt.Errorf("the clip lands %d of %d commits, so the closing log lights discs the viewer never saw arrive. Every commit needs its own beat, or the graph is smaller than you wrote it",
			landed, len(h.Commits))
	}
	return nil
}

// historyScenes lays the clip out as ONE scene. The edge list, the lane
// crossings and the children of every commit are derived here, so the
// component draws a graph it was handed rather than one it has to walk.
func historyScenes(in SnippetSceneInput) ([]Scene, error) {
	h := in.Plan.History
	if h == nil {
		return nil, fmt.Errorf("the plan has no history")
	}
	if len(h.Commits) == 0 || len(h.Lanes) == 0 {
		return nil, fmt.Errorf("the graph has no commits or no lanes")
	}

	children := make([][]int, len(h.Commits))
	for i, c := range h.Commits {
		for _, pi := range c.Parents {
			if pi >= 0 && pi < len(h.Commits) {
				children[pi] = append(children[pi], i)
			}
		}
	}

	commits := make([]map[string]any, len(h.Commits))
	edges := make([]map[string]any, 0, len(h.Commits))
	for i, c := range h.Commits {
		parents := make([]int, len(c.Parents))
		copy(parents, c.Parents)
		sort.Ints(parents)
		kids := make([]int, len(children[i]))
		copy(kids, children[i])
		sort.Ints(kids)
		commits[i] = map[string]any{
			// col is the horizontal slot: index order is time order, so the
			// graph reads left to right with no layout pass in TypeScript.
			"col":      i,
			"lane":     c.Lane,
			"label":    c.Label,
			"parents":  parents,
			"children": kids,
			"merge":    len(c.Parents) == maxHistoryParents,
		}
		for _, pi := range c.Parents {
			if pi < 0 || pi >= len(h.Commits) {
				continue
			}
			edges = append(edges, map[string]any{
				"from":     pi,
				"to":       i,
				"fromLane": h.Commits[pi].Lane,
				"toLane":   c.Lane,
				// An edge that changes lane is drawn as a curve; one that does
				// not is a straight run along the hairline.
				"curved": h.Commits[pi].Lane != c.Lane,
			})
		}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	landed := make([]int, 0, len(h.Commits))
	head := -1
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.History == nil {
			return nil, fmt.Errorf("beat %q has no history direction", beat.ID)
		}
		show := beat.History.ResolvedShow()
		at := beat.History.At
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		switch show {
		case "commit", "merge":
			if at < 0 || at >= len(h.Commits) {
				return nil, fmt.Errorf("beat %q lands commit %d, which does not exist", beat.ID, at)
			}
			landed = append(landed, at)
			head = at
			step["at"] = at
		case "branch":
			if at < 0 || at >= len(h.Commits) {
				return nil, fmt.Errorf("beat %q dwells on commit %d, which does not exist", beat.ID, at)
			}
			step["at"] = at
			kids := make([]int, len(children[at]))
			copy(kids, children[at])
			sort.Ints(kids)
			step["kids"] = kids
		case "log":
			// The closer lights everything, whatever the beats landed.
			landed = landed[:0]
			for j := range h.Commits {
				landed = append(landed, j)
			}
			head = len(h.Commits) - 1
		}
		up := make([]int, len(landed))
		copy(up, landed)
		sort.Ints(up)
		step["landed"] = up
		step["head"] = head
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneHistory,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":   in.Plan.Title,
			"lanes":   h.Lanes,
			"commits": commits,
			"edges":   edges,
			"steps":   steps,
		}),
	}}, nil
}
