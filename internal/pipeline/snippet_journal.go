package pipeline

// The journal template: a log written line by line, and then replayed.
//
// The catalog can draw a sequence of milestones (`timeline`), a system caught
// mid-operation (`trace`), and a file being edited (`workspace`). None of them
// can draw the thing an append-only log actually is: a file that only ever grows
// at the bottom, and whose entire value is that replaying it from the top
// reconstructs the state that produced it.
//
// The picture has two halves and needs both. First the lines arrive, one at a
// time, at the end — never in the middle, which is the property being taught.
// Then a cursor walks back down from the top and each line happens again. A
// viewer who has watched that does not need to be told why an AOF is durable or
// why event sourcing works; they have seen the mechanism.
//
// Three rules earn it its place, and all three are validators.
//
// The empty file is established before anything is written to it. The lines
// arriving is the whole first half, and lines that were already there when the
// frame opened did not arrive.
//
// **A line cannot be replayed before it was appended.** Obvious, and worth
// enforcing precisely because it is obvious: a model writing a replay sequence
// from memory will happily walk lines in the order that makes the narration
// flow, and a log that replays something it never recorded is not a log.
//
// **Replay runs in append order.** This is the rule the template exists for. An
// append-only log has exactly one read order, and it is the write order — that
// is the difference between a journal and a table. A replay that jumps around is
// random access, and drawing it as a replay teaches the opposite of the truth.

import (
	"fmt"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
)

func init() {
	registerSnippetTemplate(&SnippetTemplate{
		Name:        "journal",
		Category:    CatSystems,
		Since:       SinceV4,
		Family:      FamilyReplica,
		Title:       "The log, and the replay",
		Description: "An append-only file that grows at the bottom, then replays from the top to rebuild what it recorded. Reach for it for write-ahead logs, event sourcing, migrations, audit trails.",
		Example:     "How Redis rebuilds your whole dataset from an append-only file",
		PromptFile:  snippetJournalTemplateName,
		NeedsCode:   false,
		// The empty file, a few appends, and a replay that lands somewhere is
		// six beats before anything optional. This template is longer than most
		// by construction and its default says so.
		MinTargetSec:     40,
		DefaultTargetSec: 60,
		MaxBeats:         10,
		Owns:             beatFields{Journal: true},
		OwnsPlan:         planFields{Journal: true},
		Normalize:        normalizeJournalPlan,
		Validate:         validateJournalPlan,
		Scenes:           journalScenes,
		PromptData: func(_ SnippetSpec, _ config.Config) map[string]any {
			return map[string]any{
				"Roles":         strings.Join(MetricRoles(), ", "),
				"Shows":         strings.Join(JournalShows(), ", "),
				"MinEntries":    minJournalEntries,
				"MaxEntries":    maxJournalEntries,
				"MaxTextChars":  maxJournalTextChars,
				"MaxNoteWords":  maxJournalNoteWords,
				"MaxFileChars":  maxJournalFileChars,
				"MaxLabelWords": maxJournalLabelWords,
			}
		},
	})
}

const snippetJournalTemplateName = "snippet_journal.tmpl"

const (
	// Three lines is not a log, it is a list. Past ten the gutter runs off the
	// stage at a size anybody can read, and the replay walk becomes a scroll —
	// which is a different picture and a worse one.
	minJournalEntries = 4
	maxJournalEntries = 10

	// One line of monospace across the panel. Past this it wraps, and a wrapped
	// log line breaks the gutter alignment the whole picture depends on.
	maxJournalTextChars = 46
	maxJournalNoteWords = 16
	maxJournalFileChars = 28
	// Labels for the two phases, so a clip about migrations is not forced to
	// call them "append" and "replay".
	maxJournalLabelWords = 3
)

// journalShows is the closed vocabulary of what a beat does.
var journalShows = map[string]bool{
	// The empty file, with its gutter and nothing in it. The first beat, always.
	"file": true,
	// One line is written at the end.
	"append": true,
	// The cursor lands on one line during the replay.
	"replay": true,
	// Hold the finished file and say what it means.
	"read": true,
}

// JournalShows returns the beat vocabulary sorted.
func JournalShows() []string {
	out := make([]string, 0, len(journalShows))
	for k := range journalShows {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// JournalSpec is the file and its lines. On the plan because the file is one
// object that persists for the whole clip — the beats only add to it or walk it.
type JournalSpec struct {
	// File is the log's name as it should read on screen — "appendonly.aof",
	// "events.jsonl", "0042_add_index.sql".
	File string `json:"file"`
	// WriteLabel and ReplayLabel name the two phases ("" takes the defaults).
	WriteLabel  string `json:"writeLabel,omitempty"`
	ReplayLabel string `json:"replayLabel,omitempty"`
	// Entries are the lines, in the order they are written. That order is also
	// the only order they can be replayed in — see the file header.
	Entries []JournalEntry `json:"entries"`
}

// JournalEntry is one line of the log.
type JournalEntry struct {
	// Text is the line exactly as it should appear — a command, an event, a
	// statement. Monospace, one line.
	Text string `json:"text"`
	// Note is what this line means when it is replayed. Optional.
	Note string `json:"note,omitempty"`
	// Role is what this line is doing: a metricRoles name. It picks the semantic
	// accent, so it is a claim about meaning rather than about colour.
	Role string `json:"role,omitempty"`
}

// ResolvedRole returns the line's role, defaulting to neutral — most lines of a
// log are ordinary, and a log where every line is shouting has no emphasis.
func (e JournalEntry) ResolvedRole() string {
	r := strings.ToLower(strings.TrimSpace(e.Role))
	if metricRoles[r] {
		return r
	}
	return "neutral"
}

// ResolvedWriteLabel and ResolvedReplayLabel name the phases on screen.
func (s *JournalSpec) ResolvedWriteLabel() string {
	if l := strings.TrimSpace(s.WriteLabel); l != "" {
		return l
	}
	return "appending"
}

func (s *JournalSpec) ResolvedReplayLabel() string {
	if l := strings.TrimSpace(s.ReplayLabel); l != "" {
		return l
	}
	return "replaying — top to bottom"
}

// JournalBeat is one move: which line this beat writes or walks to.
type JournalBeat struct {
	// Show is a journalShows name.
	Show string `json:"show"`
	// At indexes JournalSpec.Entries, for an "append" or "replay" beat.
	At int `json:"at,omitempty"`
}

// ResolvedShow returns the beat's action, defaulting the unknown to an append.
func (b JournalBeat) ResolvedShow() string {
	s := strings.ToLower(strings.TrimSpace(b.Show))
	if journalShows[s] {
		return s
	}
	return "append"
}

func normalizeJournalPlan(p *SnippetPlan) {
	j := p.Journal
	if j == nil {
		return
	}
	j.File = clampChars(collapseSpaces(j.File), maxJournalFileChars)
	j.WriteLabel = clampWords(collapseSpaces(j.WriteLabel), maxJournalLabelWords)
	j.ReplayLabel = clampWords(collapseSpaces(j.ReplayLabel), maxJournalLabelWords)

	entries := make([]JournalEntry, 0, len(j.Entries))
	for _, e := range j.Entries {
		// Collapsed but not word-clamped: a log line is a command, and cutting
		// it to a word count would produce something that is not one.
		e.Text = clampChars(collapseSpaces(e.Text), maxJournalTextChars)
		e.Note = clampWords(collapseSpaces(e.Note), maxJournalNoteWords)
		e.Role = e.ResolvedRole()
		if e.Text != "" && len(entries) < maxJournalEntries {
			entries = append(entries, e)
		}
	}
	j.Entries = entries

	for i := range p.Beats {
		b := p.Beats[i].Journal
		if b == nil {
			continue
		}
		b.Show = b.ResolvedShow()
		if b.Show != "append" && b.Show != "replay" {
			b.At = 0
			continue
		}
		if b.At < 0 {
			b.At = 0
		}
		if n := len(j.Entries); n > 0 && b.At >= n {
			b.At = n - 1
		}
	}
}

func validateJournalPlan(p *SnippetPlan) error {
	if err := checkBeatShape(p); err != nil {
		return err
	}
	if err := rejectForeignBeatFields(p, beatFields{Journal: true}); err != nil {
		return err
	}

	j := p.Journal
	if j == nil {
		return fmt.Errorf("the plan has no log — this template is one append-only file, written and then replayed")
	}
	if strings.TrimSpace(j.File) == "" {
		return fmt.Errorf("the log has no file name — an append-only file the viewer cannot name is a panel of text")
	}
	if n := len(j.Entries); n < minJournalEntries || n > maxJournalEntries {
		return fmt.Errorf("the log has %d lines, want %d-%d. Three lines is a list rather than a log; past ten the gutter runs off the stage at a readable size and the replay becomes a scroll",
			n, minJournalEntries, maxJournalEntries)
	}
	for i, e := range j.Entries {
		if strings.TrimSpace(e.Text) == "" {
			return fmt.Errorf("line %d is empty", i)
		}
		if r := strings.ToLower(strings.TrimSpace(e.Role)); r != "" && !metricRoles[r] {
			return fmt.Errorf("line %d has role %q, which is not one of: %s", i, e.Role, strings.Join(MetricRoles(), ", "))
		}
	}

	// The empty file is established before anything is written to it.
	if p.Beats[0].Journal == nil || p.Beats[0].Journal.ResolvedShow() != "file" {
		return fmt.Errorf("beat %q does not open on the empty file. The lines arriving is the whole first half of this clip, and lines that were already there when the frame opened did not arrive",
			p.Beats[0].ID)
	}

	appended := map[int]bool{}
	files := 0
	lastAppend := -1
	lastReplay := -1
	replays := 0
	for i, b := range p.Beats {
		if b.Journal == nil {
			return fmt.Errorf("beat %q has no journal direction — every beat opens the file, appends a line, replays one, or reads the result", b.ID)
		}
		show := b.Journal.ResolvedShow()
		at := b.Journal.At
		switch show {
		case "file":
			files++
			if i != 0 {
				return fmt.Errorf("beat %q opens the file again part-way through. The log is established once, at the start", b.ID)
			}
		case "append":
			if at < 0 || at >= len(j.Entries) {
				return fmt.Errorf("beat %q appends line %d, which does not exist", b.ID, at)
			}
			if appended[at] {
				return fmt.Errorf("beat %q appends line %d again; each line is written once, which is what append-only means", b.ID, at)
			}
			// A log grows at the end. Appending line 2 after line 4 is not an
			// append-only file, it is a file with a seek in it.
			//
			// Jumping FORWARD is allowed and is not a gap: a beat that appends
			// line 5 when line 2 was the last one says "and three more landed",
			// which is how a log actually fills and how the reference clips draw
			// it. `written` is at+1, so the skipped lines are on screen. What
			// cannot happen is writing above what is already written.
			if at <= lastAppend {
				return fmt.Errorf("beat %q appends line %d when the file already reaches line %d. An append-only log only ever grows at the bottom — writing above what is already there is the one thing it cannot do",
					b.ID, at, lastAppend)
			}
			lastAppend = at
			appended[at] = true
		case "replay":
			replays++
			if at < 0 || at >= len(j.Entries) {
				return fmt.Errorf("beat %q replays line %d, which does not exist", b.ID, at)
			}
			// A line cannot be replayed before it was written.
			if !appended[at] {
				return fmt.Errorf("beat %q replays line %d, which the clip never appended. A log that replays something it did not record is not a log",
					b.ID, at)
			}
			// The rule the template exists for. An append-only log has exactly
			// one read order and it is the write order.
			if at <= lastReplay {
				return fmt.Errorf("beat %q replays line %d after line %d. Replay runs top to bottom, in the order the lines were written — that is the difference between a journal and a table, and a replay that jumps around teaches the opposite of the truth",
					b.ID, at, lastReplay)
			}
			lastReplay = at
		}
	}
	if files != 1 {
		return fmt.Errorf("there are %d beats opening the file, want exactly 1", files)
	}
	if len(appended) == 0 {
		return fmt.Errorf("no line is ever written. The first half of this clip is the file growing, and a log nobody writes to has nothing to replay")
	}
	// Both halves have to happen, or this is a different template: appends with
	// no replay is a file filling up, which `workspace` draws; a replay with no
	// appends is a list being read.
	if replays == 0 {
		return fmt.Errorf("nothing is ever replayed. The point of an append-only log is that reading it back rebuilds what produced it, and a clip that only writes has drawn half the picture — if the replay is not the point, this is a different template")
	}
	return nil
}

// journalScenes lays the clip out as ONE scene. The file persists for the whole
// clip and the beats only say how much of it exists and where the cursor is.
func journalScenes(in SnippetSceneInput) ([]Scene, error) {
	j := in.Plan.Journal
	if j == nil {
		return nil, fmt.Errorf("the plan has no log")
	}

	entries := make([]map[string]any, len(j.Entries))
	for i, e := range j.Entries {
		entries[i] = map[string]any{
			"text": e.Text,
			"note": e.Note,
			"role": e.ResolvedRole(),
		}
	}

	written := 0
	steps := make([]map[string]any, 0, len(in.Plan.Beats))
	for i := range in.Plan.Beats {
		beat, startMs, endMs := in.Beat(i)
		if beat.Journal == nil {
			return nil, fmt.Errorf("beat %q has no journal direction", beat.ID)
		}
		show := beat.Journal.ResolvedShow()
		if show == "append" {
			written = beat.Journal.At + 1
		}
		step := map[string]any{
			"startMs": startMs,
			"endMs":   endMs,
			"show":    show,
			// How many lines exist by this beat, so the renderer never has to
			// count backwards through the steps to find out.
			"written": written,
		}
		if show == "append" || show == "replay" {
			step["at"] = beat.Journal.At
		}
		steps = append(steps, step)
	}

	_, clipStart, _ := in.Beat(0)
	_, _, clipEnd := in.Beat(len(in.Plan.Beats) - 1)
	return []Scene{{
		Type:    SceneJournal,
		StartMs: clipStart,
		EndMs:   clipEnd,
		Props: headlineProps(in.Plan, map[string]any{
			"title":       in.Plan.Title,
			"file":        j.File,
			"writeLabel":  j.ResolvedWriteLabel(),
			"replayLabel": j.ResolvedReplayLabel(),
			"entries":     entries,
			"steps":       steps,
		}),
	}}, nil
}
