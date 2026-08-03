package pipeline

// The multiplex template: many sources, one worker, and readiness.
//
// `trace` draws a system caught in the act — actors issuing work into a queue
// that drains against one shared value — and that is contention: the interesting
// thing is two things colliding. This is the opposite shape and the catalog had
// no way to draw it. Here nothing collides. There are many identical sources,
// exactly one worker, and the whole claim is that the worker is enough.
//
// It is the picture underneath every event loop, every `epoll`/`kqueue`
// explanation, every "why is nginx not thread-per-request", every actor mailbox.
// And it is the one that is most often drawn wrongly: a diagram showing one
// source being handled, then the next, then the next, has drawn POLLING. The
// difference — and the entire reason multiplexing is fast — is that several
// sources become ready together and the single worker takes them in one pass.
//
// Three rules earn it its place, and all three are validators.
//
// The pool is established before anything in it is ready. A source lighting up
// on a pool the viewer has not counted is a source lighting up for no reason.
//
// **At least one round has more than one source ready at once.** This is the
// rule the template exists for. A clip whose every round wakes exactly one
// source has drawn a loop calling accept() in turn, which is the thing
// multiplexing replaces. If the honest picture really is one-at-a-time, this is
// the wrong template and `flow` or `trace` is the right one.
//
// **No two consecutive rounds have the same ready set.** A round that wakes the
// same sources as the one before it is a beat where the picture does not move,
// and the narrator is left describing a still frame.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "multiplex",
		Category:    CatSystems,
		Since:       SinceV4,
		Family:      FamilyReplica,
		Title:       "Many waiting, one working",
		Description: "A pool of identical sources where several go ready at once and a single worker takes them in one pass. Reach for it for event loops, epoll, actor mailboxes — anywhere one thread is enough.",
		Example:     "How Redis serves a hundred thousand clients on one thread",
		PromptFile:  snippetMultiplexTemplateName,
		NeedsCode:   false,
		// The pool, two or three rounds, and what it means. A round that is not
		// paused on is a round nobody read.
		MinTargetSec:     30,
		DefaultTargetSec: 50,
		MaxBeats:         8,
		Owns:             beatFields{Multiplex: true},
		OwnsPlan:         planFields{Multiplex: true},
		Normalize:        normalizeMultiplexPlan,
		Validate:         validateMultiplexPlan,
		Scenes:           multiplexScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(MultiplexShows(), ", "),
				"MinSources":    minMultiplexSources,
				"MaxSources":    maxMultiplexSources,
				"MinRounds":     minMultiplexRounds,
				"MaxRounds":     maxMultiplexRounds,
				"MaxLabelWords": maxMultiplexLabelWords,
				"MaxLabelChars": maxMultiplexLabelChars,
				"MaxNoteWords":  maxMultiplexNoteWords,
			}
		},
	})
}

const snippetMultiplexTemplateName = "snippet_multiplex.tmpl"

const (
	// Below six the pool does not read as "many"; past fourteen the column runs
	// off the stage at a size where the labels are legible.
	minMultiplexSources = 6
	maxMultiplexSources = 14

	// One round shows a state rather than a mechanism. Past four the clip is a
	// loop the viewer is watching rather than a point being made.
	minMultiplexRounds = 2
	maxMultiplexRounds = 4

	// Source labels are handles — "#00428", "conn 41", "fd 7" — so they are cut
	// by characters. A handle clamped to a word count is not a handle.
	maxMultiplexLabelChars = 14
	maxMultiplexLabelWords = 4
	maxMultiplexNoteWords  = 16
)

// multiplexShows is the closed vocabulary of what a beat does.
var multiplexShows = map[string]bool{
	// The pool and the worker, everything idle. The first beat, always.
	"pool": true,
	// One round: some sources go ready and the worker takes them.
	"round": true,
	// Hold the picture and say what it means.
	"read": true,
}

// MultiplexShows returns the beat vocabulary sorted.
func MultiplexShows() []string {
	out := make([]string, 0, len(multiplexShows))
	for k := range multiplexShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// MultiplexSpec is the pool, the worker, and the rounds. On the plan because the
// pool is one object that persists for the whole clip.
type MultiplexSpec struct {
	// Sources are the things waiting, in the order they are drawn.
	Sources []MultiplexSource `json:"sources"`
	// SourceKind names what one of them is — "socket", "connection", "mailbox".
	SourceKind string `json:"sourceKind"`
	// Worker is the single thing serving them — "epoll", "the event loop".
	Worker string `json:"worker"`
	// WorkerNote is the line under the worker that carries the claim —
	// "1 thread", "one core". Optional but this template is much weaker without
	// it: the whole argument is how little is doing the work.
	WorkerNote string `json:"workerNote,omitempty"`
	// Rounds are the passes the worker makes, in order.
	Rounds []MultiplexRound `json:"rounds"`
}

// MultiplexSource is one thing waiting to be served.
type MultiplexSource struct {
	// Label is the source's handle as it reads on screen.
	Label string `json:"label"`
}

// MultiplexRound is one pass: which sources woke up, and what that shows.
type MultiplexRound struct {
	// Ready indexes MultiplexSpec.Sources — the ones ready in this pass.
	Ready []int `json:"ready"`
	// Note is what this pass means. One sentence.
	Note string `json:"note,omitempty"`
	// Role is what this round is doing: a metricRoles name.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the round's role, defaulting to neutral.
func (r MultiplexRound) ResolvedRole() string {
	s := strings.ToLower(strings.TrimSpace(r.Role))
	if metricRoles[s] {
		return s
	}
	return "neutral"
}

// MultiplexBeat is one move: which round this beat runs.
type MultiplexBeat struct {
	// Show is a multiplexShows name.
	Show string `json:"show"`
	// At indexes MultiplexSpec.Rounds, for a "round" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to a round.
func (b MultiplexBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if multiplexShows[s] {
		return s
	}
	return "round"
}

// sameReady reports whether two rounds wake the same set, order-insensitively.
func sameReady(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	x := append([]int(nil), a...)
	y := append([]int(nil), b...)
	sort.Ints(x)
	sort.Ints(y)
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

func normalizeMultiplexPlan(p *SnippetPlan) {
	m := p.Multiplex
	if m == nil {
		return
	}
	m.SourceKind = clampWords(collapseSpaces(m.SourceKind), maxMultiplexLabelWords)
	m.Worker = clampWords(collapseSpaces(m.Worker), maxMultiplexLabelWords)
	m.WorkerNote = clampWords(collapseSpaces(m.WorkerNote), maxMultiplexLabelWords)

	sources := make([]MultiplexSource, 0, len(m.Sources))
	for _, s := range m.Sources {
		s.Label = clampChars(collapseSpaces(s.Label), maxMultiplexLabelChars)
		if s.Label != "" && len(sources) < maxMultiplexSources {
			sources = append(sources, s)
		}
	}
	m.Sources = sources

	rounds := make([]MultiplexRound, 0, len(m.Rounds))
	for _, r := range m.Rounds {
		r.Note = clampWords(collapseSpaces(r.Note), maxMultiplexNoteWords)
		r.Role = r.ResolvedRole()
		// Drop indexes that point at nothing, and de-duplicate: a source listed
		// twice in one round is not twice as ready.
		seen := map[int]bool{}
		ready := make([]int, 0, len(r.Ready))
		for _, i := range r.Ready {
			if i >= 0 && i < len(m.Sources) && !seen[i] {
				seen[i] = true
				ready = append(ready, i)
			}
		}
		sort.Ints(ready)
		r.Ready = ready
		// A round where nothing is ready is a round where nothing happens.
		if len(r.Ready) > 0 && len(rounds) < maxMultiplexRounds {
			rounds = append(rounds, r)
		}
	}
	m.Rounds = rounds

	for i := range p.Beats {
		b := p.Beats[i].Multiplex
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "round" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(m.Rounds); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateMultiplexPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Multiplex: true}); err != nil {
		return err
	}

	m := p.Multiplex
	if m == nil {
		return fmt.Errorf("the plan has no pool — this template is many identical sources and the one worker that serves them")
	}
	if strings.TrimSpace(m.SourceKind) == "" {
		return fmt.Errorf("the pool has no source kind — say what one of them is. A column of unnamed handles is a list of strings")
	}
	if strings.TrimSpace(m.Worker) == "" {
		return fmt.Errorf("nothing is named as the worker. The claim this template makes is that ONE thing serves the whole pool, and a picture with no worker in it has not made it")
	}
	if n := len(m.Sources); n < minMultiplexSources || n > maxMultiplexSources {
		return fmt.Errorf("the pool has %d sources, want %d-%d. Below six it does not read as many; past fourteen the column runs off the stage at a size where the labels are legible",
			n, minMultiplexSources, maxMultiplexSources)
	}
	if n := len(m.Rounds); n < minMultiplexRounds || n > maxMultiplexRounds {
		return fmt.Errorf("there are %d rounds, want %d-%d. One round shows a state rather than a mechanism; past four the clip is a loop being watched rather than a point being made",
			n, minMultiplexRounds, maxMultiplexRounds)
	}

	seen := map[string]bool{}
	for i, s := range m.Sources {
		if strings.TrimSpace(s.Label) == "" {
			return fmt.Errorf("source %d has no label", i)
		}
		key := strings.ToLower(strings.TrimSpace(s.Label))
		if seen[key] {
			return fmt.Errorf("two sources are both %q — every handle in the pool is a different one", s.Label)
		}
		seen[key] = true
	}

	// The rule the template exists for. Every round waking exactly one source is
	// a loop calling accept() in turn, which is the thing multiplexing replaces.
	widest := 0
	for _, r := range m.Rounds {
		if len(r.Ready) > widest {
			widest = len(r.Ready)
		}
	}
	if widest < 2 {
		return fmt.Errorf("every round wakes exactly one source, so the clip has drawn polling rather than multiplexing. The whole reason one worker is enough is that several sources go ready TOGETHER and it takes them in one pass — give at least one round two or more ready sources, or use the flow template if the honest picture really is one at a time")
	}

	for i, r := range m.Rounds {
		if len(r.Ready) == 0 {
			return fmt.Errorf("round %d wakes nothing", i)
		}
		for _, at := range r.Ready {
			if at < 0 || at >= len(m.Sources) {
				return fmt.Errorf("round %d wakes source %d, which does not exist", i, at)
			}
		}
		if role := strings.ToLower(strings.TrimSpace(r.Role)); role != "" && !metricRoles[role] {
			return fmt.Errorf("round %d has role %q, which is not one of: %s", i, r.Role, strings.Join(MetricRoles(), ", "))
		}
		// A round that wakes the same sources as the one before is a beat where
		// the picture does not move.
		if i > 0 && sameReady(r.Ready, m.Rounds[i-1].Ready) {
			return fmt.Errorf("round %d wakes the same sources as round %d, so nothing on screen changes between them. Each pass has to find a different set ready — that is what makes it a pass rather than a photograph",
				i, i-1)
		}
	}

	// The pool is counted before anything in it lights up.
	if p.Beats[0].Multiplex == nil || p.Beats[0].Multiplex.ResolvedShow() != "pool" {
		return fmt.Errorf("beat %q does not establish the pool. A source lighting up on a pool the viewer has not counted is a source lighting up for no reason",
			p.Beats[0].ID)
	}

	ran := map[int]bool{}
	pools := 0
	for i, b := range p.Beats {
		if b.Multiplex == nil {
			return fmt.Errorf("beat %q has no multiplex direction — every beat draws the pool, runs a round, or reads the result", b.ID)
		}
		switch b.Multiplex.ResolvedShow() {
		case "pool":
			pools++
			if i != 0 {
				return fmt.Errorf("beat %q draws the pool again part-way through. It is established once, at the start", b.ID)
			}
		case "round":
			if b.Multiplex.At < 0 || b.Multiplex.At >= len(m.Rounds) {
				return fmt.Errorf("beat %q runs round %d, which does not exist", b.ID, b.Multiplex.At)
			}
			if ran[b.Multiplex.At] {
				return fmt.Errorf("beat %q runs round %d again; each pass gets one beat", b.ID, b.Multiplex.At)
			}
			ran[b.Multiplex.At] = true
		}
	}
	if pools != 1 {
		return fmt.Errorf("there are %d beats establishing the pool, want exactly 1", pools)
	}
	if len(ran) != len(m.Rounds) {
		return fmt.Errorf("%d of the %d rounds never run. A pass the narrator skips is one nobody saw — give it a beat or cut it",
			len(m.Rounds)-len(ran), len(m.Rounds))
	}
	return nil
}

// multiplexScenes lays the clip out as ONE scene. The pool and the worker
// persist; the beats only say which sources are ready.
func multiplexScenes(in SnippetSceneInput) ([]Scene, error) {
	m := in.Plan.Multiplex
	if m == nil {
		return nil, fmt.Errorf("the plan has no pool")
	}

	sources := make([]map[string]any, len(m.Sources))
	for i, s := range m.Sources {
		sources[i] = map[string]any{"label": s.Label}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Multiplex == nil {
			return nil, fmt.Errorf("beat %q has no multiplex direction", beat.ID)
		}
		show := beat.Multiplex.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "round" {
			r := m.Rounds[beat.Multiplex.At]
			step["at"] = beat.Multiplex.At
			step["ready"] = append([]int(nil), r.Ready...)
			step["note"] = r.Note
			step["role"] = r.ResolvedRole()
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneMultiplex,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":      in.Plan.Title,
			"sourceKind": m.SourceKind,
			"worker":     m.Worker,
			"workerNote": m.WorkerNote,
			"sources":    sources,
			"steps":      steps,
		}),
	}}, nil
}
