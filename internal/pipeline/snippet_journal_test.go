package pipeline

import (
	"strings"
	"testing"
)

const jnNarration = "Every write goes on the end of the file, and nothing already written ever moves."

func journalPlan() *SnippetPlan {
	p := &SnippetPlan{
		Template: "journal",
		Title:    "The file that rebuilds your database",
		Journal: &JournalSpec{
			File: "appendonly.aof",
			Entries: []JournalEntry{
				{Text: "SET user:42 alice", Note: "The first write anyone made"},
				{Text: "SET cart:42 [items]"},
				{Text: "INCR visits"},
				{Text: "DEL cart:42", Note: "The delete is a record too", Role: "limit"},
				{Text: "SET score:42 1200"},
			},
		},
		Beats: []SnippetBeat{
			{ID: "empty", Heading: "An empty file", Narration: jnNarration, Journal: &JournalBeat{Show: "file"}},
			{ID: "first", Heading: "The first write", Narration: jnNarration, Journal: &JournalBeat{Show: "append", At: 0}},
			{ID: "grows", Heading: "It only grows", Narration: jnNarration, Journal: &JournalBeat{Show: "append", At: 1}},
			{ID: "third", Heading: "And again", Narration: jnNarration, Journal: &JournalBeat{Show: "append", At: 2}},
			{ID: "delete", Heading: "Even deletes", Narration: jnNarration, Journal: &JournalBeat{Show: "append", At: 3}},
			{ID: "replay-top", Heading: "Read it back", Narration: jnNarration, Journal: &JournalBeat{Show: "replay", At: 0}},
			{ID: "replay-del", Heading: "The delete replays", Narration: jnNarration, Journal: &JournalBeat{Show: "replay", At: 3}},
			{ID: "same", Heading: "Same state", Narration: jnNarration, Journal: &JournalBeat{Show: "read"}},
		},
	}
	p.targetWords = 8 * 40
	return p
}

func TestJournalPlanAccepted(t *testing.T) {
	if err := validateJournalPlan(journalPlan()); err != nil {
		t.Fatalf("a well-formed journal plan was rejected: %v", err)
	}
}

// The rule the template exists for. An append-only log has exactly one read
// order and it is the write order — a replay that jumps around is random
// access, and drawing it as a replay teaches the opposite of the truth.
func TestJournalRejectsAReplayThatGoesBackwards(t *testing.T) {
	p := journalPlan()
	p.Beats[5].Journal = &JournalBeat{Show: "replay", At: 3}
	p.Beats[6].Journal = &JournalBeat{Show: "replay", At: 1}
	err := validateJournalPlan(p)
	if err == nil {
		t.Fatal("a replay that walks backwards was accepted")
	}
	if !strings.Contains(err.Error(), "top to bottom") {
		t.Fatalf("the error does not say what a replay is: %v", err)
	}
}

// A log grows at the end. Writing above what is already there is the one thing
// it cannot do.
func TestJournalRejectsAnAppendThatGoesBackwards(t *testing.T) {
	p := journalPlan()
	// 0, 3, then back to 2.
	p.Beats[2].Journal = &JournalBeat{Show: "append", At: 3}
	p.Beats[3].Journal = &JournalBeat{Show: "append", At: 2}
	err := validateJournalPlan(p)
	if err == nil {
		t.Fatal("an append above the end of the file was accepted")
	}
	if !strings.Contains(err.Error(), "grows at the bottom") {
		t.Fatalf("the error does not explain the rule: %v", err)
	}
}

// Skipping forward is not a gap: it says several lines landed at once, which is
// how a log actually fills and how the reference clips draw it.
func TestJournalAllowsAnAppendThatSkipsForward(t *testing.T) {
	p := journalPlan()
	p.Beats[2].Journal = &JournalBeat{Show: "append", At: 3}
	p.Beats[3].Journal = &JournalBeat{Show: "read"}
	p.Beats[4].Journal = &JournalBeat{Show: "read"}
	if err := validateJournalPlan(p); err != nil {
		t.Fatalf("an append that skipped forward was rejected: %v", err)
	}
}

func TestJournalRejectsReplayingAnUnwrittenLine(t *testing.T) {
	p := journalPlan()
	// Line 4 is never appended by the fixture, so replaying it is replaying
	// something the clip did not record.
	p.Beats[6].Journal = &JournalBeat{Show: "replay", At: 4}
	if err := validateJournalPlan(p); err == nil {
		t.Fatal("a replay of a line that was never appended was accepted")
	}
}

// Both halves have to happen or this is a different template: appends with no
// replay is a file filling up, which workspace draws.
func TestJournalRequiresAReplay(t *testing.T) {
	p := journalPlan()
	p.Beats[5].Journal = &JournalBeat{Show: "read"}
	p.Beats[6].Journal = &JournalBeat{Show: "read"}
	err := validateJournalPlan(p)
	if err == nil {
		t.Fatal("a clip that only writes was accepted")
	}
	if !strings.Contains(err.Error(), "half the picture") {
		t.Fatalf("the error does not name what is missing: %v", err)
	}
}

func TestJournalRequiresTheEmptyFileFirst(t *testing.T) {
	p := journalPlan()
	p.Beats[0].Journal = &JournalBeat{Show: "append", At: 0}
	p.Beats[1].Journal = &JournalBeat{Show: "file"}
	if err := validateJournalPlan(p); err == nil {
		t.Fatal("a clip that appends before opening the file was accepted")
	}
}

func TestJournalRejectsAppendingTheSameLineTwice(t *testing.T) {
	p := journalPlan()
	p.Beats[4].Journal = &JournalBeat{Show: "append", At: 2}
	if err := validateJournalPlan(p); err == nil {
		t.Fatal("a line appended twice was accepted")
	}
}

func TestJournalRejectsAnUndrawableFile(t *testing.T) {
	p := journalPlan()
	p.Journal.Entries = p.Journal.Entries[:3]
	// Re-point the beats rather than deleting them, so the shared beat floor is
	// not what does the rejecting.
	p.Beats[4].Journal = &JournalBeat{Show: "read"}
	p.Beats[6].Journal = &JournalBeat{Show: "read"}
	if err := validateJournalPlan(p); err == nil {
		t.Fatal("a three-line log was accepted")
	}
}

func TestJournalRequiresAFileName(t *testing.T) {
	p := journalPlan()
	p.Journal.File = ""
	if err := validateJournalPlan(p); err == nil {
		t.Fatal("a log with no file name was accepted")
	}
}

// A log line is a command, so it is cut by characters rather than by words —
// clamping to a word count would produce something that is not a command.
func TestJournalNormalizeClampsLinesByCharacters(t *testing.T) {
	p := journalPlan()
	p.Journal.Entries[0].Text = strings.Repeat("x", maxJournalTextChars+20)
	normalizeJournalPlan(p)
	if got := len(p.Journal.Entries[0].Text); got > maxJournalTextChars {
		t.Fatalf("line is %d chars, want at most %d", got, maxJournalTextChars)
	}
}

// The scene carries how many lines exist at each beat, so the renderer never
// counts backwards through the steps to find out.
func TestJournalScenesCarryTheWrittenCount(t *testing.T) {
	p := journalPlan()
	scenes, err := journalScenes(sceneInput(t, p, 24000))
	if err != nil {
		t.Fatalf("scenes: %v", err)
	}
	steps, _ := scenes[0].Props["steps"].([]map[string]any)
	if steps[0]["written"] != 0 {
		t.Fatalf("the opening beat has %v lines written, want 0", steps[0]["written"])
	}
	if steps[4]["written"] != 4 {
		t.Fatalf("after four appends %v lines are written, want 4", steps[4]["written"])
	}
	// The replay beats do not add lines; the file stops growing once it is read.
	if steps[6]["written"] != 4 {
		t.Fatalf("a replay beat changed the written count to %v, want 4", steps[6]["written"])
	}
}

func TestJournalPhaseLabelsHaveDefaults(t *testing.T) {
	j := &JournalSpec{}
	if j.ResolvedWriteLabel() == "" || j.ResolvedReplayLabel() == "" {
		t.Fatal("a log that names neither phase renders with blank labels")
	}
	j.ReplayLabel = "rebuilding"
	if j.ResolvedReplayLabel() != "rebuilding" {
		t.Fatalf("a stated label was ignored: %q", j.ResolvedReplayLabel())
	}
}
