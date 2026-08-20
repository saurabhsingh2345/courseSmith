package pipeline

// The doc template: a markdown file open in an editor, read section by section.
//
// The catalog has three ways to put a file on screen and none of them draw this.
// `vscode` types code into a buffer and RUNS it, so it carries a verify stage and
// an interpreter's output — the whole point of it is that the code is real.
// `workspace` moves a camera across several files like a screencast. `patch` shows
// what changed in a diff. All three are about code that executes.
//
// A markdown file executes nothing, and that is precisely why it needs its own
// treatment. What matters about a document is its STRUCTURE — which headings exist,
// what sits under each one, and which section the narration is on right now. So
// this template holds the whole file still, dims it, and lights one section at a
// time. The viewer always sees the shape of the document and never loses their
// place in it, which is impossible in a template that scrolls a camera around.
//
// Three details do the work.
//
// THE GUTTER IS REAL. Line numbers are computed from the flattened document, so
// section three starts at line 14 because the two sections above it are thirteen
// lines long. Numbers that do not add up are the fastest way to make an editor
// look drawn, and they are free to get right.
//
// THE TREE IS OPTIONAL AND IT ANSWERS A DIFFERENT QUESTION. A document on its own
// says what to write; a document beside a file tree says WHERE THE FILE LIVES, and
// for anything about configuration that second question is usually the lesson. One
// entry can be marked, so the frame can point at a file among its siblings without
// a callout.
//
// AND MARKDOWN IS STYLED, NOT PRINTED. Headings, bullets and inline code are
// picked out by their own syntax as the author wrote it, because a config file that
// renders as flat grey text teaches nothing about how to write one. The styling is
// derived from the line's first characters — no parser, no AST, and no way for an
// author to get a heading by asking for one in a field.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:     "doc",
		Category: CatCode,
		Since:    SinceV10,
		Family:   FamilyAtelier,
		Title:    "A file, section by section",
		Description: "A markdown file open in an editor with real line numbers, dimmed except for the section being read, and optionally a file tree beside it saying where the file lives. " +
			"Reach for it to teach what belongs IN a file — a config, a spec, a memory file; `vscode` is for code that runs.",
		Example:    "what belongs in a project's memory file",
		PromptFile: snippetDocTemplateName,
		NeedsCode:  false,
		// A document needs long enough to walk two or three sections; under about
		// twenty seconds it is a screenshot with a voice over it.
		MinTargetSec:     20,
		DefaultTargetSec: 50,
		MaxTargetSec:     180,
		// The file, up to five sections, the tree, the pull-back.
		MaxBeats: 8,
		// A section of a config file is worth a sentence or two, like a session
		// event rather than a caption.
		IdealWordsPerBeat: 28,
		Owns:              beatFields{Doc: true},
		OwnsPlan:          planFields{Doc: true},
		Normalize:         normalizeDocPlan,
		Validate:          validateDocPlan,
		Scenes:            docScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Shows":         strings.Join(DocShows(), ", "),
				"Kinds":         strings.Join(DocTreeKinds(), ", "),
				"MinBlocks":     minDocBlocks,
				"MaxBlocks":     maxDocBlocks,
				"MaxBlockLines": maxDocBlockLines,
				"MaxLineChars":  maxDocLineChars,
				"MaxTree":       maxDocTreeItems,
				"MaxNoteWords":  maxDocNoteWords,
			}
		},
	})
}

const snippetDocTemplateName = "snippet_doc.tmpl"

const (
	// Two sections is a document with a shape; past five the file is longer than
	// the frame and the dimming stops helping because everything is dim.
	minDocBlocks = 2
	maxDocBlocks = 5

	maxDocBlockLines = 7
	maxDocLineChars  = 74
	maxDocTreeItems  = 9
	maxDocNoteWords  = 12
	maxDocFileChars  = 34
	maxDocCrumbChars = 40
)

// docShows is the closed vocabulary of what a beat does.
var docShows = map[string]bool{
	// The editor arrives with the file open and the whole document visible.
	"open": true,
	// The file tree beside it, with the marked entry lit.
	"tree": true,
	// Section At lights and everything else dims.
	"block": true,
	// Every section back up at full strength: the finished file.
	"whole": true,
}

// docTreeKinds is what one row of the tree is.
var docTreeKinds = map[string]bool{
	"dir":  true,
	"file": true,
	// The one the frame is about. At most one row may be marked — two marks is
	// a frame pointing at two things, which points at neither.
	"mark": true,
}

// DocShows returns the beat vocabulary sorted.
func DocShows() []string {
	out := make([]string, 0, len(docShows))
	for k := range docShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// DocTreeKinds returns the tree-row vocabulary sorted.
func DocTreeKinds() []string {
	out := make([]string, 0, len(docTreeKinds))
	for k := range docTreeKinds {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// DocSpec is the file on screen.
type DocSpec struct {
	// File is the name in the tab — "CLAUDE.md".
	File string `json:"file"`
	// Crumb is the breadcrumb tail an editor shows beside the filename.
	Crumb string `json:"crumb,omitempty"`
	// Tree is the optional file list beside the document.
	Tree []DocTreeItem `json:"tree,omitempty"`
	// Blocks are the document's sections, top to bottom.
	Blocks []DocBlock `json:"blocks"`
}

// DocTreeItem is one row of the file tree.
type DocTreeItem struct {
	Name  string `json:"name"`
	Kind  string `json:"kind,omitempty"`
	Depth int    `json:"depth,omitempty"`
}

// ResolvedKind defaults to a plain file.
func (i DocTreeItem) ResolvedKind() string {
	k := strings.ToLower(strings.TrimSpace(i.Kind))
	if docTreeKinds[k] {
		return k
	}
	return "file"
}

// DocBlock is one section of the document.
type DocBlock struct {
	// Text is the section's first line, usually its heading — written in markdown
	// as it would appear in the file ("## Commands").
	Text string `json:"text"`
	// Lines are the rest of the section, again as written.
	Lines []string `json:"lines,omitempty"`
	// Note is the annotation in the margin while this section is lit.
	Note string `json:"note,omitempty"`
}

// DocBeat is one shot of the document.
type DocBeat struct {
	Show string `json:"show"`
	At   int    `json:"at,omitempty"`
}

// ResolvedShow defaults to a section lighting, which is what most beats are.
func (b DocBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if docShows[s] {
		return s
	}
	return "block"
}

func normalizeDocPlan(p *SnippetPlan) {
	d := p.Doc
	if d == nil {
		return
	}
	d.File = clampCodeLine(collapseSpaces(d.File), maxDocFileChars)
	d.Crumb = clampCodeLine(collapseSpaces(d.Crumb), maxDocCrumbChars)

	tree := make([]DocTreeItem, 0, maxDocTreeItems)
	marked := false
	for _, it := range d.Tree {
		it.Name = clampCodeLine(collapseSpaces(it.Name), 30)
		it.Kind = it.ResolvedKind()
		if it.Kind == "mark" {
			// Only the first mark survives. See docTreeKinds.
			if marked {
				it.Kind = "file"
			}
			marked = true
		}
		if it.Depth < 0 {
			it.Depth = 0
		}
		if it.Depth > 3 {
			it.Depth = 3
		}
		if it.Name != "" && len(tree) < maxDocTreeItems {
			tree = append(tree, it)
		}
	}
	d.Tree = tree

	blocks := make([]DocBlock, 0, maxDocBlocks)
	for _, b := range d.Blocks {
		// keepIndent, not collapseSpaces: a markdown list nests by indentation
		// and a config file's structure is the thing being taught.
		b.Text = keepIndent(b.Text, maxDocLineChars)
		b.Note = clampWords(collapseSpaces(b.Note), maxDocNoteWords)
		lines := make([]string, 0, maxDocBlockLines)
		for _, l := range b.Lines {
			if l = keepIndent(l, maxDocLineChars); l != "" && len(lines) < maxDocBlockLines {
				lines = append(lines, l)
			}
		}
		b.Lines = lines
		if b.Text != "" && len(blocks) < maxDocBlocks {
			blocks = append(blocks, b)
		}
	}
	d.Blocks = blocks

	for i := range p.Beats {
		if b := p.Beats[i].Doc; b != nil {
			b.Show = b.ResolvedShow()
			if b.Show != "block" {
				b.At = 0
			}
		}
	}
}

func validateDocPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Doc: true}); err != nil {
		return err
	}

	d := p.Doc
	if d == nil {
		return fmt.Errorf("the plan has no document — this template is one file held open, so the file is the clip")
	}
	if strings.TrimSpace(d.File) == "" {
		return fmt.Errorf("the document has no filename. The tab is what says which file this is, and a config lesson about an unnamed file teaches nothing")
	}
	if n := len(d.Blocks); n < minDocBlocks || n > maxDocBlocks {
		return fmt.Errorf("the document has %d section(s) and wants %d-%d: one section is a screenshot, and past %d the file is taller than the frame",
			n, minDocBlocks, maxDocBlocks, maxDocBlocks)
	}

	var (
		lit    = map[int]bool{}
		opens  int
		wholes int
		trees  int
	)
	for i, b := range p.Beats {
		if b.Doc == nil {
			return fmt.Errorf("beat %q has no doc direction", b.ID)
		}
		show := b.Doc.ResolvedShow()
		switch show {
		case "open":
			opens++
			if i != 0 {
				return fmt.Errorf("beat %q opens the file, but it is beat %d — the file opens once, first", b.ID, i+1)
			}
		case "block":
			at := b.Doc.At
			if at < 0 || at >= len(d.Blocks) {
				return fmt.Errorf("beat %q lights section %d of %d", b.ID, at, len(d.Blocks))
			}
			if lit[at] {
				return fmt.Errorf("beat %q lights section %d again. Lighting a section twice means the voice went back, and the dimming has already told the viewer that section was finished with", b.ID, at)
			}
			lit[at] = true
		case "tree":
			trees++
			if len(d.Tree) == 0 {
				return fmt.Errorf("beat %q shows the file tree but the document has none", b.ID)
			}
		case "whole":
			wholes++
			if i != len(p.Beats)-1 {
				return fmt.Errorf("beat %q brings the whole file back up, but %d beat(s) follow it — that is the closing frame",
					b.ID, len(p.Beats)-1-i)
			}
		}
	}
	if opens != 1 {
		return fmt.Errorf("there are %d \"open\" beats and there must be exactly one", opens)
	}
	if wholes != 1 {
		return fmt.Errorf("there are %d \"whole\" beats and there must be exactly one, last", wholes)
	}
	if trees > 1 {
		return fmt.Errorf("the tree is shown %d times; show it once", trees)
	}
	if len(lit) != len(d.Blocks) {
		return fmt.Errorf("%d of the %d sections are never lit. A section on screen that the voice never reaches is one the viewer read on their own while being told about something else",
			len(d.Blocks)-len(lit), len(d.Blocks))
	}
	return nil
}

func docScenes(in SnippetSceneInput) ([]Scene, error) {
	d := in.Plan.Doc
	if d == nil {
		return nil, fmt.Errorf("the plan has no document")
	}

	blocks := make([]map[string]any, 0, len(d.Blocks))
	for _, b := range d.Blocks {
		blocks = append(blocks, map[string]any{
			"text":  b.Text,
			"lines": b.Lines,
			"note":  b.Note,
		})
	}
	tree := make([]map[string]any, 0, len(d.Tree))
	for _, it := range d.Tree {
		tree = append(tree, map[string]any{
			"name":  it.Name,
			"kind":  it.ResolvedKind(),
			"depth": it.Depth,
		})
	}

	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Doc == nil {
			return nil, fmt.Errorf("beat %q has no doc direction", beat.ID)
		}
		show := beat.Doc.ResolvedShow()
		// The lit section is NOT latched, unlike a session's transcript: exactly
		// one section is lit at a time and that is the whole mechanic. -1 means
		// the file is up at full strength with nothing singled out.
		at := -1
		if show == "block" {
			at = beat.Doc.At
		}
		steps = append(steps, map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			"at":      at,
		})
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneDoc,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: map[string]any{
			"file":   d.File,
			"crumb":  d.Crumb,
			"tree":   tree,
			"blocks": blocks,
			"steps":  steps,
		},
	}}, nil
}
