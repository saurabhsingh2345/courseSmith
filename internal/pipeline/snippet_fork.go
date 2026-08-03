package pipeline

// The fork template: two processes over one memory, and what happens on a write.
//
// `trace` draws two actors contending for a shared value — the interesting thing
// is a collision. This is a different shape the catalog could not draw: two
// actors over shared state where nothing collides, because the *first write*
// silently gives the writer its own copy and everybody else keeps the original.
//
// Copy-on-write is taught badly almost everywhere, and the reason is that the
// usual diagram shows the copy and stops. What that diagram misses is the part
// that makes the technique worth anything: the pages nobody wrote to are still
// one page, shared, costing nothing. A picture that copies everything has drawn
// two separate memories, which is the thing copy-on-write exists to avoid.
//
// So the frame holds both states at once — copied pages and still-shared pages —
// and the validators are about keeping it honest.
//
// The shared state is established before anything diverges. A page splitting on
// memory the viewer has not seen whole is a page splitting for no reason.
//
// **A page copies at most once.** After the first write the writer has its own
// page; a second write by the same side lands on the copy it already owns and
// nothing further happens. Drawing a second copy is the classic misunderstanding
// of the mechanism, and it is exactly the mistake a model writing from memory
// makes.
//
// **At least one page stays shared.** This is the rule the template exists for.
// If every page ends up copied, the clip has drawn a full duplicate and the word
// "copy-on-write" no longer describes anything on screen — the saving IS the
// pages nobody touched, and a frame with none of them left has argued against
// its own subject.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "fork",
		Category:    CatSystems,
		Since:       SinceV4,
		Family:      FamilyReplica,
		Title:       "Shared until someone writes",
		Description: "Two processes over one memory, where a write splits a single page and leaves the rest shared. Reach for it for copy-on-write, snapshots, forks, structural sharing in immutable data.",
		Example:     "How Redis takes a snapshot without pausing the database",
		PromptFile:  snippetForkTemplateName,
		NeedsCode:   false,
		// The fork, the shared memory, a write or two, and what stayed shared.
		MinTargetSec:     35,
		DefaultTargetSec: 55,
		MaxBeats:         8,
		Owns:             beatFields{Fork: true},
		OwnsPlan:         planFields{Fork: true},
		Normalize:        normalizeForkPlan,
		Validate:         validateForkPlan,
		Scenes:           forkScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(ForkShows(), ", "),
				"Sides":         strings.Join(ForkSides(), ", "),
				"MinPages":      minForkPages,
				"MaxPages":      maxForkPages,
				"MinWrites":     minForkWrites,
				"MaxWrites":     maxForkWrites,
				"MaxLabelWords": maxForkLabelWords,
				"MaxLabelChars": maxForkLabelChars,
				"MaxNoteWords":  maxForkNoteWords,
			}
		},
	})
}

const snippetForkTemplateName = "snippet_fork.tmpl"

const (
	// Below four pages there is not enough memory for "most of it stayed
	// shared" to be visible; past eight the cells are too small to carry a copy
	// appearing beneath one of them.
	minForkPages = 4
	maxForkPages = 8

	// One write shows the mechanism; three is as many divergences as a viewer
	// can hold while still tracking which side owns what.
	minForkWrites = 1
	maxForkWrites = 3

	maxForkLabelWords = 3
	maxForkLabelChars = 14
	maxForkNoteWords  = 16
)

// forkShows is the closed vocabulary of what a beat does.
var forkShows = map[string]bool{
	// The two processes over one memory, nothing diverged. The first beat.
	"shared": true,
	// One side writes, and that page copies.
	"write": true,
	// Hold the picture and say what stayed shared.
	"read": true,
}

// ForkShows returns the beat vocabulary sorted.
func ForkShows() []string {
	out := make([]string, 0, len(forkShows))
	for k := range forkShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// forkSides is which process a write comes from. A closed vocabulary because the
// renderer draws the copy on the writer's side, and "either" has nowhere to go.
var forkSides = map[string]bool{"parent": true, "child": true}

// ForkSides returns the side vocabulary sorted.
func ForkSides() []string {
	out := make([]string, 0, len(forkSides))
	for k := range forkSides {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ForkSpec is the two processes and the memory under them. On the plan because
// the memory is one object that persists for the whole clip.
type ForkSpec struct {
	// Origin is what forked — "redis-server". Optional; the frame opens on it.
	Origin string `json:"origin,omitempty"`
	// OriginNote is the line under it — "pid 4271". Optional.
	OriginNote string `json:"originNote,omitempty"`
	// Parent and Child name the two sides ("" takes "parent" and "child").
	Parent string `json:"parent,omitempty"`
	Child  string `json:"child,omitempty"`
	// Pages are the shared memory, drawn left to right.
	Pages []ForkPage `json:"pages"`
	// Writes are the divergences, in the order they happen.
	Writes []ForkWrite `json:"writes"`
}

// ForkPage is one page of the shared memory.
type ForkPage struct {
	// Label is what lives in this page — "user:42", "the dataset". Optional:
	// a memory of unlabelled cells is a perfectly honest picture of pages.
	Label string `json:"label,omitempty"`
}

// ForkWrite is one divergence.
type ForkWrite struct {
	// At indexes ForkSpec.Pages.
	At int `json:"at"`
	// By is which side wrote: a forkSides name.
	By string `json:"by"`
	// Note says what just happened, and ideally what did NOT.
	Note string `json:"note,omitempty"`
	// Role is what this write is doing: a metricRoles name.
	Role string `json:"role,omitempty"`
}

// ResolvedBy returns the writing side, defaulting to the parent — which is the
// side that keeps serving traffic in every real example of this.
func (w ForkWrite) ResolvedBy() string {
	s := strings.ToLower(strings.TrimSpace(w.By))
	if forkSides[s] {
		return s
	}
	return "parent"
}

// ResolvedRole returns the write's role, defaulting to neutral.
func (w ForkWrite) ResolvedRole() string {
	s := strings.ToLower(strings.TrimSpace(w.Role))
	if metricRoles[s] {
		return s
	}
	return "neutral"
}

// ResolvedParent and ResolvedChild name the two sides on screen.
func (s *ForkSpec) ResolvedParent() string {
	if l := strings.TrimSpace(s.Parent); l != "" {
		return l
	}
	return "parent"
}

func (s *ForkSpec) ResolvedChild() string {
	if l := strings.TrimSpace(s.Child); l != "" {
		return l
	}
	return "child"
}

// ForkBeat is one move: which write this beat performs.
type ForkBeat struct {
	// Show is a forkShows name.
	Show string `json:"show"`
	// At indexes ForkSpec.Writes, for a "write" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to a write.
func (b ForkBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if forkShows[s] {
		return s
	}
	return "write"
}

func normalizeForkPlan(p *SnippetPlan) {
	f := p.Fork
	if f == nil {
		return
	}
	f.Origin = clampChars(collapseSpaces(f.Origin), maxForkLabelChars)
	f.OriginNote = clampWords(collapseSpaces(f.OriginNote), maxForkLabelWords)
	f.Parent = clampWords(collapseSpaces(f.Parent), maxForkLabelWords)
	f.Child = clampWords(collapseSpaces(f.Child), maxForkLabelWords)

	pages := make([]ForkPage, 0, len(f.Pages))
	for _, pg := range f.Pages {
		pg.Label = clampChars(collapseSpaces(pg.Label), maxForkLabelChars)
		if len(pages) < maxForkPages {
			pages = append(pages, pg)
		}
	}
	f.Pages = pages

	writes := make([]ForkWrite, 0, len(f.Writes))
	seen := map[string]bool{}
	for _, w := range f.Writes {
		w.Note = clampWords(collapseSpaces(w.Note), maxForkNoteWords)
		w.By = w.ResolvedBy()
		w.Role = w.ResolvedRole()
		if w.At < 0 || w.At >= len(f.Pages) {
			continue
		}
		// A page copies once per side. A second write by the same side lands on
		// the copy it already owns, so drawing it again would be a second copy —
		// which is the classic misunderstanding of the mechanism. Dropping is
		// the repair; the model meant one divergence and described two.
		key := fmt.Sprintf("%d/%s", w.At, w.By)
		if seen[key] {
			continue
		}
		seen[key] = true
		if len(writes) < maxForkWrites {
			writes = append(writes, w)
		}
	}
	f.Writes = writes

	for i := range p.Beats {
		b := p.Beats[i].Fork
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "write" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(f.Writes); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateForkPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Fork: true}); err != nil {
		return err
	}

	f := p.Fork
	if f == nil {
		return fmt.Errorf("the plan has no memory — this template is two processes over one memory and what a write does to it")
	}
	if n := len(f.Pages); n < minForkPages || n > maxForkPages {
		return fmt.Errorf("the memory has %d pages, want %d-%d. Below four there is not enough of it for \"most of it stayed shared\" to be visible, which is the whole point; past eight the cells are too small to carry a copy appearing beneath one",
			n, minForkPages, maxForkPages)
	}
	if n := len(f.Writes); n < minForkWrites || n > maxForkWrites {
		return fmt.Errorf("there are %d writes, want %d-%d. One shows the mechanism; past three a viewer cannot track which side owns what",
			n, minForkWrites, maxForkWrites)
	}

	copied := map[int]bool{}
	for i, w := range f.Writes {
		if w.At < 0 || w.At >= len(f.Pages) {
			return fmt.Errorf("write %d touches page %d, which does not exist", i, w.At)
		}
		if s := strings.ToLower(strings.TrimSpace(w.By)); s != "" && !forkSides[s] {
			return fmt.Errorf("write %d comes from %q, which is not one of: %s. The copy is drawn on the writer's side, so a write from neither has nowhere to go",
				i, w.By, strings.Join(ForkSides(), ", "))
		}
		if r := strings.ToLower(strings.TrimSpace(w.Role)); r != "" && !metricRoles[r] {
			return fmt.Errorf("write %d has role %q, which is not one of: %s", i, w.Role, strings.Join(MetricRoles(), ", "))
		}
		// A page copies at most once. Two writes to the same page — from either
		// side — is the misunderstanding this template exists to correct: after
		// the first, the writer already has its own copy.
		if copied[w.At] {
			return fmt.Errorf("write %d copies page %d, which has already been copied. Copy-on-write copies a page ONCE — after the first write the writer owns its copy and writes to it directly, so a second copy of the same page is the mistake this picture is supposed to prevent",
				i, w.At)
		}
		copied[w.At] = true
	}

	// The rule the template exists for. The saving IS the pages nobody touched.
	if len(copied) >= len(f.Pages) {
		return fmt.Errorf("every one of the %d pages ends up copied, so the clip has drawn a full duplicate rather than copy-on-write. The saving is the pages nobody wrote to — leave at least one shared, or the word stops describing anything on screen",
			len(f.Pages))
	}

	// The memory is seen whole before anything splits it.
	if p.Beats[0].Fork == nil || p.Beats[0].Fork.ResolvedShow() != "shared" {
		return fmt.Errorf("beat %q does not open on the shared memory. A page splitting on memory the viewer has not seen whole is a page splitting for no reason",
			p.Beats[0].ID)
	}

	done := map[int]bool{}
	shareds := 0
	for i, b := range p.Beats {
		if b.Fork == nil {
			return fmt.Errorf("beat %q has no fork direction — every beat shows the shared memory, performs a write, or reads the result", b.ID)
		}
		switch b.Fork.ResolvedShow() {
		case "shared":
			shareds++
			if i != 0 {
				return fmt.Errorf("beat %q shows the memory shared again part-way through. Once a page has diverged it does not go back, and drawing it whole again would say it did", b.ID)
			}
		case "write":
			if b.Fork.At < 0 || b.Fork.At >= len(f.Writes) {
				return fmt.Errorf("beat %q performs write %d, which does not exist", b.ID, b.Fork.At)
			}
			if done[b.Fork.At] {
				return fmt.Errorf("beat %q performs write %d again; each divergence gets one beat", b.ID, b.Fork.At)
			}
			done[b.Fork.At] = true
		}
	}
	if shareds != 1 {
		return fmt.Errorf("there are %d beats opening on the shared memory, want exactly 1", shareds)
	}
	if len(done) != len(f.Writes) {
		return fmt.Errorf("%d of the %d writes never happen. A divergence the narrator skips is one nobody saw — give it a beat or cut it",
			len(f.Writes)-len(done), len(f.Writes))
	}
	return nil
}

// forkScenes lays the clip out as ONE scene. The memory persists and the beats
// only say which pages have diverged and to whom.
func forkScenes(in SnippetSceneInput) ([]Scene, error) {
	f := in.Plan.Fork
	if f == nil {
		return nil, fmt.Errorf("the plan has no memory")
	}

	pages := make([]map[string]any, len(f.Pages))
	for i, pg := range f.Pages {
		pages[i] = map[string]any{"label": pg.Label}
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	// Which pages have copied by each beat, and to which side. Accumulated in Go
	// so the renderer never replays the write list to find out.
	copies := map[int]string{}
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Fork == nil {
			return nil, fmt.Errorf("beat %q has no fork direction", beat.ID)
		}
		show := beat.Fork.ResolvedShow()
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
		}
		if show == "write" {
			w := f.Writes[beat.Fork.At]
			copies[w.At] = w.ResolvedBy()
			step["at"] = w.At
			step["by"] = w.ResolvedBy()
			step["note"] = w.Note
			step["role"] = w.ResolvedRole()
		}
		// A fresh map per step: the renderer reads one step and draws the whole
		// frame from it, so each has to carry the state as it stands.
		snapshot := make(map[string]any, len(copies))
		for at, by := range copies {
			snapshot[fmt.Sprintf("%d", at)] = by
		}
		step["copied"] = snapshot
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneFork,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":      in.Plan.Title,
			"origin":     f.Origin,
			"originNote": f.OriginNote,
			"parent":     f.ResolvedParent(),
			"child":      f.ResolvedChild(),
			"pages":      pages,
			"steps":      steps,
		}),
	}}, nil
}
