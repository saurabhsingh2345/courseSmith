package pipeline

// The patch template: one change at a time, at a size you can read.
//
// The catalog can type code out (`code`), walk a file in an editor (`vscode`) and
// show a whole workspace (`workspace`), and not one of those can show a CHANGE.
// That is the most-looked-at artifact in software — a diff — and the reason it was
// missing is that the obvious way to draw one does not work on video: a unified
// diff is forty lines of 14pt monospace, and a viewer watching it go past reads
// none of it. The tool this borrows from prints exactly that, and the clip is a
// screen recording you squint at.
//
// So the diff is taken apart. One hunk is on screen at a time, at nearly twice the
// size, with two or three lines of context around it and nothing else. The removed
// line is struck through in place and the added line arrives underneath it, which
// is the one animation this template needs and the only one it has: a viewer sees
// a change HAPPEN rather than seeing a document that contains a change.
//
// The tally is the second half of the idea. A standing `N files · +A −D` sits under
// the window and grows as hunks land, so the frame always answers "how big is this
// change" — the question a raw diff makes you scroll to the end to answer, and the
// one that decides whether somebody reads it at all.
//
// Three rules earn the shape.
//
// A hunk carries a NOTE saying why. A diff is what changed; the note is the only
// place what it is FOR can live, and without it this template is a slower way to
// read a patch file. It is required.
//
// A hunk is at most a handful of lines. Not a stylistic cap — the whole premise is
// that one change is readable at this size, and a twelve-line hunk at this size
// does not fit the window, so it would have to shrink back to the size that made
// the original unreadable. Big rewrites are several hunks or a different template.
//
// And the lines are CLAMPED, hard, at a width that fits. A diff line is as long as
// whoever wrote the code felt like making it; the frame is 1500 pixels of mono.
// Truncating in the pipeline with an ellipsis is honest and legible; letting it
// wrap turns a two-line change into five lines of ragged soup and destroys the
// alignment that makes a diff readable at all.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "patch",
		Category: CatCode,
		Since:    SinceV9,
		Family:   FamilyShowroom,
		Title:    "One change at a time",
		Description: "A diff taken apart: one hunk on screen at readable size, the old line struck through and the new one arriving under it, with a running count of what the change adds up to. " +
			"Reach for it to show what actually changed in a file.",
		Example:    "The three-line change that made the upload handler async",
		PromptFile: snippetPatchTemplateName,
		NeedsCode:  true,
		// The file up, a beat per hunk, the tally. Two hunks is four beats.
		MinTargetSec:     30,
		DefaultTargetSec: 50,
		// Six: the file, at most four hunks, the tally. 2 + maxPatchHunks, and the
		// arithmetic has to hold — see duel's MaxBeats for what happens when it
		// does not.
		// 144s: 6 beats x 60 words a beat, at 2.5 words a second. Past this the
		// shape cannot hold the narration — see MaxTargetSec.
		MaxTargetSec: 144,
		MaxBeats:     6,
		// A beat is one hunk landing and being explained, and it wants to be slow:
		// the viewer has to read two lines of code before the sentence about them
		// makes sense.
		IdealWordsPerBeat: 26,
		Owns:              beatFields{Patch: true},
		OwnsPlan:          planFields{Patch: true},
		Normalize:         normalizePatchPlan,
		Validate:          validatePatchPlan,
		Scenes:            patchScenes,
		PromptData: func(spec SnippetSpec, cfg config.Config) map[string]any {
			return map[string]any{
				"Shows":          strings.Join(PatchShows(), ", "),
				"MinHunks":       minPatchHunks,
				"MaxHunks":       patchHunkCeilingFor(spec, cfg),
				"MaxLineChars":   maxPatchLineChars,
				"MaxContext":     maxPatchContext,
				"MaxChanged":     maxPatchChangedLines,
				"MaxNoteWords":   maxPatchNoteWords,
				"MaxPathChars":   maxChangePathChars,
				"MaxCloserWords": maxPatchCloserWords,
			}
		},
	})
}

const snippetPatchTemplateName = "snippet_patch.tmpl"

const (
	// One hunk is a code scene with a strikethrough. Past four the clip is a patch
	// review rather than a lesson about one change.
	minPatchHunks = 1
	maxPatchHunks = 4

	// The line width the window actually fits, at the size this template sets code
	// at. See the file header on why this is clamped rather than wrapped.
	maxPatchLineChars = 62
	// Context lines on each side of the change.
	maxPatchContext = 3
	// Removed plus added lines in one hunk.
	maxPatchChangedLines = 5
	// Why this hunk exists.
	maxPatchNoteWords = 16
	// The line under the finished tally.
	maxPatchCloserWords = 16
)

// patchShows is the closed vocabulary of what a beat does.
var patchShows = map[string]bool{
	// The file open at rest, no hunk applied. The opener.
	"file": true,
	// Hunk At lands: the old line struck, the new one arriving.
	"hunk": true,
	// The tally, and what the change adds up to. The closer.
	"tally": true,
}

// PatchShows returns the beat vocabulary sorted.
func PatchShows() []string {
	out := make([]string, 0, len(patchShows))
	for k := range patchShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// PatchSpec is the change.
type PatchSpec struct {
	// Path is the file being changed, as it appears in the repo.
	Path string `json:"path"`
	// Lang is a syntax hint for the renderer's highlighter, e.g. "ts", "go".
	Lang string `json:"lang,omitempty"`
	// Hunks are the changes, in the order the clip walks them.
	Hunks []PatchHunk `json:"hunks"`
	// Closer is the line under the finished tally.
	Closer string `json:"closer,omitempty"`
}

// PatchHunk is one change in one place.
type PatchHunk struct {
	// At is the line number the hunk starts at, for the gutter. Cosmetic — it
	// numbers the rows and nothing reads it back.
	At int `json:"at,omitempty"`
	// Before are the lines being removed. Empty is legal: a hunk that only adds.
	Before []string `json:"before,omitempty"`
	// After are the lines being added. Empty is legal: a hunk that only removes.
	After []string `json:"after,omitempty"`
	// Context are the unchanged lines shown around the change, in order, with the
	// change sitting after them.
	Context []string `json:"context,omitempty"`
	// Note is why this change was made. Required — see the file header.
	Note string `json:"note"`
}

// PatchBeat is one shot.
type PatchBeat struct {
	// Show is a patchShows name.
	Show string `json:"show"`
	// At indexes PatchSpec.Hunks, for a "hunk" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults to a hunk landing.
func (b PatchBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if patchShows[s] {
		return s
	}
	return "hunk"
}

// Insertions and Deletions are the tally, summed over the hunks walked so far.
func (s *PatchSpec) Insertions(upto int) int { return patchCount(s.Hunks, upto, true) }
func (s *PatchSpec) Deletions(upto int) int  { return patchCount(s.Hunks, upto, false) }

func patchCount(hunks []PatchHunk, upto int, added bool) int {
	n := 0
	for i, h := range hunks {
		if i >= upto {
			break
		}
		if added {
			n += len(h.After)
		} else {
			n += len(h.Before)
		}
	}
	return n
}

// patchHunkBudget is how many hunks the beat budget funds.
func patchHunkBudget(targetWords int) int {
	_, maxBeats, _, _ := beatBounds(targetWords, templateBeatCeiling("patch"), templateIdealWords("patch"))
	return min(max(maxBeats-2, minPatchHunks), maxPatchHunks)
}

func patchHunkCeilingFor(spec SnippetSpec, cfg config.Config) int {
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	want, _, _ := wordBudget(spec.ResolvedTargetSec(), pace)
	return patchHunkBudget(want)
}

// clampCodeLine trims a line of code to what the window fits, from the right,
// marking the cut. See the file header on why this happens here rather than in CSS.
func clampCodeLine(s string, n int) string {
	s = strings.TrimRight(s, " \t")
	if len([]rune(s)) <= n {
		return s
	}
	r := []rune(s)
	return string(r[:n-1]) + "…"
}

func normalizePatchPlan(p *SnippetPlan) {
	pa := p.Patch
	if pa == nil {
		return
	}
	pa.Path = clampPathLeft(collapseSpaces(pa.Path), maxChangePathChars)
	pa.Lang = strings.ToLower(strings.TrimSpace(pa.Lang))
	pa.Closer = clampWords(collapseSpaces(pa.Closer), maxPatchCloserWords)

	lines := func(in []string, max int) []string {
		out := make([]string, 0, len(in))
		for _, l := range in {
			if l = clampCodeLine(l, maxPatchLineChars); strings.TrimSpace(l) != "" && len(out) < max {
				out = append(out, l)
			}
		}
		return out
	}

	hunks := make([]PatchHunk, 0, len(pa.Hunks))
	for _, h := range pa.Hunks {
		h.Context = lines(h.Context, maxPatchContext)
		h.Before = lines(h.Before, maxPatchChangedLines)
		h.After = lines(h.After, maxPatchChangedLines)
		// The cap is on the CHANGE, not on each side of it: three removed and
		// three added is six lines of moving type, which is past what one beat can
		// carry however it is split.
		for len(h.Before)+len(h.After) > maxPatchChangedLines && len(h.After) > 0 {
			h.After = h.After[:len(h.After)-1]
		}
		for len(h.Before)+len(h.After) > maxPatchChangedLines && len(h.Before) > 0 {
			h.Before = h.Before[:len(h.Before)-1]
		}
		h.Note = clampWords(collapseSpaces(h.Note), maxPatchNoteWords)
		if h.At < 0 {
			h.At = 0
		}
		if (len(h.Before) > 0 || len(h.After) > 0) && len(hunks) < maxPatchHunks {
			hunks = append(hunks, h)
		}
	}
	pa.Hunks = hunks

	for i := range p.Beats {
		b := p.Beats[i].Patch
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "hunk" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(pa.Hunks); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validatePatchPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Patch: true}); err != nil {
		return err
	}

	pa := p.Patch
	if pa == nil {
		return fmt.Errorf("the plan has no patch — this template draws one file's change, so the hunks are the clip")
	}
	if strings.TrimSpace(pa.Path) == "" {
		return fmt.Errorf("the patch has no path. The window's title is the file being changed, and a diff with no filename is a diff nobody can place")
	}
	budget := patchHunkBudget(p.targetWords)
	if n := len(pa.Hunks); n < minPatchHunks || n > maxPatchHunks {
		return fmt.Errorf("the patch has %d hunks, want %d-%d", n, minPatchHunks, maxPatchHunks)
	}
	if n := len(pa.Hunks); n > budget {
		return fmt.Errorf("the patch has %d hunks but this runtime funds only %d: the first beat opens the file, the last totals it, and every beat between lands one hunk. Use %d hunks, or ask for a longer clip",
			n, budget, budget)
	}

	for i, h := range pa.Hunks {
		if len(h.Before) == 0 && len(h.After) == 0 {
			return fmt.Errorf("hunk %d changes nothing — it has no removed lines and no added ones", i)
		}
		if n := len(h.Before) + len(h.After); n > maxPatchChangedLines {
			return fmt.Errorf("hunk %d moves %d lines, and %d is the ceiling. The whole premise here is that ONE change is readable at nearly twice the size a diff is normally set at — past that the window has to shrink the type back to the size that made the original unreadable. Split it into separate hunks, or use the code template to show the finished file",
				i, n, maxPatchChangedLines)
		}
		if strings.TrimSpace(h.Note) == "" {
			return fmt.Errorf("hunk %d has no note. The diff says WHAT changed and the note is the only place WHY can live — without it this template is a slower way to read a patch file", i)
		}
	}

	if p.Beats[0].Patch == nil || p.Beats[0].Patch.ResolvedShow() != "file" {
		return fmt.Errorf("beat %q does not open on the file at rest. A change has to be shown against the code it changes, so the viewer sees the before — open with {\"show\": \"file\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.Patch == nil || last.Patch.ResolvedShow() != "tally" {
		return fmt.Errorf("beat %q does not close on the tally. \"How big was that change\" is the question a diff makes you scroll to answer, and this template answers it on the last frame — end with {\"show\": \"tally\"}", last.ID)
	}

	next := 0
	for _, b := range p.Beats {
		d := b.Patch
		if d == nil {
			return fmt.Errorf("beat %q has no patch direction — every beat opens the file, lands one hunk, or totals the change", b.ID)
		}
		if d.ResolvedShow() != "hunk" {
			continue
		}
		if d.At < 0 || d.At >= len(pa.Hunks) {
			return fmt.Errorf("beat %q lands hunk %d, which does not exist — the patch has hunks 0-%d", b.ID, d.At, len(pa.Hunks)-1)
		}
		if next >= len(pa.Hunks) {
			return fmt.Errorf("beat %q lands a hunk when every one has already landed. Each hunk lands once, in file order", b.ID)
		}
		if d.At != next {
			return fmt.Errorf("beat %q lands hunk %d when hunk %d is next. A patch is walked in file order — jumping about is how a reviewer loses the thread of a change", b.ID, d.At, next)
		}
		next++
	}
	if next != len(pa.Hunks) {
		return fmt.Errorf("the clip lands %d of %d hunks, so part of the change is counted in the tally and never shown", next, len(pa.Hunks))
	}
	return nil
}

// patchScenes lays the clip out as ONE scene: the window persists, and each step
// says which hunk is showing and what the tally is up to.
func patchScenes(in SnippetSceneInput) ([]Scene, error) {
	pa := in.Plan.Patch
	if pa == nil {
		return nil, fmt.Errorf("the plan has no patch")
	}
	if len(pa.Hunks) == 0 {
		return nil, fmt.Errorf("the patch has no hunks")
	}

	hunks := make([]map[string]any, len(pa.Hunks))
	for i, h := range pa.Hunks {
		hunks[i] = map[string]any{
			"at":      h.At,
			"before":  h.Before,
			"after":   h.After,
			"context": h.Context,
			"note":    h.Note,
		}
	}

	// The tally is carried per step rather than recomputed in the renderer, so the
	// numbers on the frame and the numbers the validator scored are the same
	// arithmetic done once.
	landed := 0
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Patch == nil {
			return nil, fmt.Errorf("beat %q has no patch direction", beat.ID)
		}
		show := beat.Patch.ResolvedShow()
		step := map[string]any{"startMs": startMs, "endMs": endMs, "show": show, "at": -1}
		switch show {
		case "hunk":
			at := beat.Patch.At
			if at < 0 || at >= len(pa.Hunks) {
				return nil, fmt.Errorf("beat %q lands hunk %d, which does not exist", beat.ID, at)
			}
			step["at"] = at
			landed = at + 1
		case "tally":
			landed = len(pa.Hunks)
		}
		step["landed"] = landed
		step["added"] = pa.Insertions(landed)
		step["removed"] = pa.Deletions(landed)
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    ScenePatch,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"path":   pa.Path,
			"lang":   pa.Lang,
			"hunks":  hunks,
			"closer": pa.Closer,
			"steps":  steps,
		}),
	}}, nil
}
