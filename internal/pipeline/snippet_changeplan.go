package pipeline

// The changeplan template: what an agent says it is about to do, file by file.
//
// The artifact this draws is now one of the most-read documents in software and
// the catalog had no frame for it: an agent's plan, before it touches anything.
// The tool it is borrowed from prints that plan as a wall of monospace — a
// numbered list of files, each with bullets under it, scrolling past — and on
// video that is close to unreadable. The words are 14pt, there are forty lines,
// and the viewer has no idea how much is left.
//
// So this is the same document with the scroll taken out of it. The files stand
// still in a RAIL down the left, each with its own count of changes, and one
// file's bullets are open beside it. Nothing scrolls; the light moves. The viewer
// can see at any moment how many files this touches and which one is being
// discussed, which is the thing the original cannot tell them.
//
// Three rules earn the shape.
//
// A file has to say WHAT CHANGES IN IT, not what it is. "src/routes/auth.ts — the
// auth routes" names the file twice; "src/routes/auth.ts — convert and write in
// the handler" is the plan. The rail's summary line is the plan in miniature and
// it is required.
//
// Every file gets its own beat, in order. Same rule as the cards row, and here it
// is stronger: a file in the rail that the voice never opens is a file the viewer
// counted and then never heard about, which is worse than not listing it.
//
// And a plan is allowed to say NOTHING CHANGES in a file. That is the field
// `verdict: "unchanged"`, and it exists because it is the most reassuring thing a
// plan can contain and the first thing a summary drops. A viewer watching an agent
// touch their codebase wants to know what it decided to leave alone.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "changeplan",
		Category: CatCode,
		Since:    SinceV9,
		Family:   FamilyShowroom,
		Title:    "Changes by file",
		Description: "An agent's plan before it edits: the files it will touch standing in a rail with their change counts, and one file's changes open beside it. " +
			"Reach for it to show what a tool proposes, or to walk a patch you are about to make.",
		Example:    "What Claude Code proposes before it touches your repo",
		PromptFile: snippetChangePlanTemplateName,
		NeedsCode:  false,
		// The rail up, a beat per file, and the summing-up. Three files is four
		// beats; six files is eight and needs a longer clip.
		MinTargetSec:     30,
		DefaultTargetSec: 50,
		MaxBeats:         9,
		// A beat is one file being opened and explained. Twenty-four words is
		// about eight seconds, which is what two or three bullets want.
		IdealWordsPerBeat: 24,
		Owns:              beatFields{ChangePlan: true},
		OwnsPlan:          planFields{ChangePlan: true},
		Normalize:         normalizeChangePlanPlan,
		Validate:          validateChangePlanPlan,
		Scenes:            changePlanScenes,
		PromptData: func(spec SnippetSpec, cfg config.Config) map[string]any {
			return map[string]any{
				"Shows":           strings.Join(ChangePlanShows(), ", "),
				"Verdicts":        strings.Join(ChangeVerdicts(), ", "),
				"MinFiles":        minChangeFiles,
				"MaxFiles":        changeFileCeilingFor(spec, cfg),
				"MaxPathChars":    maxChangePathChars,
				"MaxSummaryWords": maxChangeSummaryWords,
				"MaxEditWords":    maxChangeEditWords,
				"MinEdits":        minChangeEdits,
				"MaxEdits":        maxChangeEdits,
				"MaxCloserWords":  maxChangePlanCloserWords,
			}
		},
	})
}

const snippetChangePlanTemplateName = "snippet_changeplan.tmpl"

const (
	// One file is not a plan, it is an edit. Past six the rail's rows drop below
	// the size a path can be read at, and the clip cannot fund them anyway.
	minChangeFiles = 2
	maxChangeFiles = 6

	// A path, not a prose description. Truncated from the LEFT when it is too
	// long, because the filename is the part that identifies it.
	maxChangePathChars = 42
	// The rail's line: what changes in this file.
	maxChangeSummaryWords = 9
	// One bullet inside a file.
	maxChangeEditWords = 14
	minChangeEdits     = 1
	maxChangeEdits     = 4
	// The line under the finished rail.
	maxChangePlanCloserWords = 16
)

// changeVerdicts is what a file's row says about it at a glance.
var changeVerdicts = map[string]bool{
	// Lines are added to it.
	"add": true,
	// Lines are replaced in it.
	"edit": true,
	// The file goes away.
	"delete": true,
	// Deliberately left alone, and the plan says why. See the file header.
	"unchanged": true,
}

// ChangeVerdicts returns the vocabulary sorted.
func ChangeVerdicts() []string {
	out := make([]string, 0, len(changeVerdicts))
	for k := range changeVerdicts {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// changePlanShows is the closed vocabulary of what a beat does.
var changePlanShows = map[string]bool{
	// The whole rail up, no file open. The opener.
	"rail": true,
	// File At open beside the rail.
	"file": true,
	// Every row settled and the closing line. The closer.
	"all": true,
}

// ChangePlanShows returns the beat vocabulary sorted.
func ChangePlanShows() []string {
	out := make([]string, 0, len(changePlanShows))
	for k := range changePlanShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// ChangePlanSpec is the plan.
type ChangePlanSpec struct {
	// Files are the rail, top to bottom, in the order the plan works through
	// them.
	Files []ChangeFile `json:"files"`
	// Closer is the line under the finished rail — what the plan adds up to.
	Closer string `json:"closer,omitempty"`
}

// ChangeFile is one row of the rail.
type ChangeFile struct {
	// Path is the file, as it appears in the repo.
	Path string `json:"path"`
	// Summary is what changes in it — not what it is. See the file header.
	Summary string `json:"summary"`
	// Verdict is a changeVerdicts name.
	Verdict string `json:"verdict,omitempty"`
	// Edits are the bullets that open beside the rail. Empty is legal only for
	// an `unchanged` file, where the summary is the whole story.
	Edits []string `json:"edits,omitempty"`
}

// ResolvedVerdict defaults to an edit, which is what most rows of most plans are.
func (f ChangeFile) ResolvedVerdict() string {
	v := strings.ToLower(strings.TrimSpace(f.Verdict))
	if changeVerdicts[v] {
		return v
	}
	return "edit"
}

// ChangePlanBeat is one shot.
type ChangePlanBeat struct {
	// Show is a changePlanShows name.
	Show string `json:"show"`
	// At indexes ChangePlanSpec.Files, for a "file" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow defaults to a file opening, which is what most beats are.
func (b ChangePlanBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if changePlanShows[s] {
		return s
	}
	return "file"
}

// changeFileBudget is how many files the beat budget funds: the rail goes up, the
// closer lands, and everything between is one file.
func changeFileBudget(targetWords int) int {
	_, maxBeats, _, _ := beatBounds(targetWords, templateBeatCeiling("changeplan"), templateIdealWords("changeplan"))
	return min(max(maxBeats-2, minChangeFiles), maxChangeFiles)
}

func changeFileCeilingFor(spec SnippetSpec, cfg config.Config) int {
	pace := cfg.Style.PaceWPM
	if pace <= 0 {
		pace = 150
	}
	want, _, _ := wordBudget(spec.ResolvedTargetSec(), pace)
	return changeFileBudget(want)
}

// clampPathLeft trims a path from the LEFT, keeping the end.
//
// The opposite of clampChars, and the reason is that the identifying part of a
// path is its tail. "…/services/userService.ts" is the file; "src/main/java/com/…"
// is four directories and no answer.
func clampPathLeft(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n || n < 4 {
		return s
	}
	return "…" + s[len(s)-(n-1):]
}

func normalizeChangePlanPlan(p *SnippetPlan) {
	c := p.ChangePlan
	if c == nil {
		return
	}
	c.Closer = clampWords(collapseSpaces(c.Closer), maxChangePlanCloserWords)

	files := make([]ChangeFile, 0, len(c.Files))
	for _, f := range c.Files {
		f.Path = clampPathLeft(collapseSpaces(f.Path), maxChangePathChars)
		f.Summary = clampWords(collapseSpaces(f.Summary), maxChangeSummaryWords)
		f.Verdict = f.ResolvedVerdict()
		edits := make([]string, 0, len(f.Edits))
		for _, e := range f.Edits {
			if e = clampWords(collapseSpaces(e), maxChangeEditWords); e != "" && len(edits) < maxChangeEdits {
				edits = append(edits, e)
			}
		}
		f.Edits = edits
		if f.Path != "" && len(files) < maxChangeFiles {
			files = append(files, f)
		}
	}
	c.Files = files

	for i := range p.Beats {
		b := p.Beats[i].ChangePlan
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "file" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(c.Files); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateChangePlanPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{ChangePlan: true}); err != nil {
		return err
	}

	c := p.ChangePlan
	if c == nil {
		return fmt.Errorf("the plan has no files — this template is a change plan, so the rail of files is the clip")
	}
	budget := changeFileBudget(p.targetWords)
	if n := len(c.Files); n < minChangeFiles || n > maxChangeFiles {
		return fmt.Errorf("the rail has %d files, want %d-%d. One file is an edit rather than a plan; past %d the rows drop below the size a path can be read at",
			n, minChangeFiles, maxChangeFiles, maxChangeFiles)
	}
	if n := len(c.Files); n > budget {
		return fmt.Errorf("the rail has %d files but this runtime funds only %d: the first beat raises the rail, the last sums it up, and every beat between opens one file. Use %d files, or ask for a longer clip",
			n, budget, budget)
	}

	seen := map[string]bool{}
	for i, f := range c.Files {
		if strings.TrimSpace(f.Path) == "" {
			return fmt.Errorf("file %d has no path", i)
		}
		if seen[f.Path] {
			return fmt.Errorf("two rows are both %q, so the plan touches one file twice", f.Path)
		}
		seen[f.Path] = true
		if strings.TrimSpace(f.Summary) == "" {
			return fmt.Errorf("file %d (%q) has no summary. The rail's line is the plan in miniature — say what CHANGES in it, not what the file is", i, f.Path)
		}
		// The failure this template invites, and the one the header calls out.
		if changePlanNamesItself(f.Path, f.Summary) {
			return fmt.Errorf("file %d's summary (%q) is just the file's own name again. The path is already on the row; the line beside it has to say what happens to it — \"convert and write in the handler\", \"switch to memory storage\"", i, f.Summary)
		}
		if f.ResolvedVerdict() == "unchanged" {
			continue // the summary is the whole story; no bullets required
		}
		if len(f.Edits) < minChangeEdits {
			return fmt.Errorf("file %d (%q) is marked %q and has nothing under it. A file the plan says it will change has to say how — or mark it \"unchanged\", which is a real and useful answer",
				i, f.Path, f.ResolvedVerdict())
		}
	}

	if p.Beats[0].ChangePlan == nil || p.Beats[0].ChangePlan.ResolvedShow() != "rail" {
		return fmt.Errorf("beat %q does not open on the rail. The count of files is the first thing this frame gives a viewer — how much of their codebase is about to move — so open with {\"show\": \"rail\"}", p.Beats[0].ID)
	}
	if last := p.Beats[len(p.Beats)-1]; last.ChangePlan == nil || last.ChangePlan.ResolvedShow() != "all" {
		return fmt.Errorf("beat %q does not close on the finished rail. Ending inside one file leaves the plan half-read — end with {\"show\": \"all\"}", last.ID)
	}

	next := 0
	for _, b := range p.Beats {
		d := b.ChangePlan
		if d == nil {
			return fmt.Errorf("beat %q has no changeplan direction — every beat raises the rail, opens one file, or sums up", b.ID)
		}
		if d.ResolvedShow() != "file" {
			continue
		}
		if d.At < 0 || d.At >= len(c.Files) {
			return fmt.Errorf("beat %q opens file %d, which does not exist — the rail has files 0-%d", b.ID, d.At, len(c.Files)-1)
		}
		if next >= len(c.Files) {
			return fmt.Errorf("beat %q opens %q when every file has already been opened. Each row is opened once, top to bottom", b.ID, c.Files[d.At].Path)
		}
		if d.At != next {
			return fmt.Errorf("beat %q opens %q when %q is the next row down. The rail is read top to bottom, once each — jumping about is how a viewer loses their place in a list of files",
				b.ID, c.Files[d.At].Path, c.Files[next].Path)
		}
		next++
	}
	if next != len(c.Files) {
		return fmt.Errorf("the clip opens %d of %d files, so %q sits in the rail counted and never discussed — which is worse than leaving it out",
			next, len(c.Files), c.Files[next].Path)
	}
	return nil
}

// changePlanNamesItself reports whether a summary is just the path's own filename
// with the punctuation taken out.
func changePlanNamesItself(path, summary string) bool {
	base := path
	if i := strings.LastIndexByte(base, '/'); i >= 0 {
		base = base[i+1:]
	}
	if i := strings.IndexByte(base, '.'); i > 0 {
		base = base[:i]
	}
	return openerFold(base) != "" && openerFold(base) == openerFold(summary)
}

// changePlanScenes lays the clip out as ONE scene: the rail persists and the
// steps say which file is open.
func changePlanScenes(in SnippetSceneInput) ([]Scene, error) {
	c := in.Plan.ChangePlan
	if c == nil {
		return nil, fmt.Errorf("the plan has no files")
	}
	if len(c.Files) == 0 {
		return nil, fmt.Errorf("the rail has no files")
	}

	files := make([]map[string]any, len(c.Files))
	for i, f := range c.Files {
		files[i] = map[string]any{
			"path":    f.Path,
			"summary": f.Summary,
			"verdict": f.ResolvedVerdict(),
			"edits":   f.Edits,
		}
	}

	// `done` counts the rows already walked, so the rail can mark them without the
	// component replaying the timeline to work out which.
	done := 0
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.ChangePlan == nil {
			return nil, fmt.Errorf("beat %q has no changeplan direction", beat.ID)
		}
		show := beat.ChangePlan.ResolvedShow()
		step := map[string]any{"startMs": startMs, "endMs": endMs, "show": show, "at": -1}
		switch show {
		case "file":
			at := beat.ChangePlan.At
			if at < 0 || at >= len(c.Files) {
				return nil, fmt.Errorf("beat %q opens file %d, which does not exist", beat.ID, at)
			}
			step["at"] = at
			done = at + 1
		case "all":
			done = len(c.Files)
		}
		step["done"] = done
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneChangePlan,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":  in.Plan.Title,
			"files":  files,
			"closer": c.Closer,
			"steps":  steps,
		}),
	}}, nil
}
