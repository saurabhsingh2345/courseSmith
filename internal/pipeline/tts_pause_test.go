package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestEndsSentence(t *testing.T) {
	cases := map[string]bool{
		"end.":      true,
		"really?":   true,
		"stop!":     true,
		`"quoted."`: true,
		"end.)":     true,
		"middle":    false,
		"comma,":    false,
		"dash—":     false,
		"":          false,
		`"`:         false,
	}
	for word, want := range cases {
		if got := endsSentence(word); got != want {
			t.Errorf("endsSentence(%q) = %v, want %v", word, got, want)
		}
	}
}

// A sentence end that already breathes past the floor must be left alone:
// the floor exists to stop run-on delivery, not to make every gap identical.
func TestPlanSentencePausesOnlyWidensShortGaps(t *testing.T) {
	words := []AlignedWord{
		{Word: "one", StartMs: 0, EndMs: 300},
		{Word: "two.", StartMs: 300, EndMs: 600}, // gap to next is 100 → widen
		{Word: "three", StartMs: 700, EndMs: 1000},
		{Word: "four.", StartMs: 1000, EndMs: 1300}, // gap is 700 → leave alone
		{Word: "five", StartMs: 2000, EndMs: 2300},
	}
	inserts, shifted := planSentencePauses(words, 400, 500)

	if len(inserts) != 1 {
		t.Fatalf("got %d insert(s), want 1: %+v", len(inserts), inserts)
	}
	if inserts[0].AddMs != 300 { // 400 floor - 100 existing
		t.Errorf("AddMs = %d, want 300", inserts[0].AddMs)
	}
	if inserts[0].AtMs != 650 { // midpoint of the 600..700 gap
		t.Errorf("AtMs = %d, want 650", inserts[0].AtMs)
	}

	// Words before the insertion keep their times; words after all move by
	// exactly the inserted amount.
	if shifted[1].EndMs != 600 {
		t.Errorf("word before insert moved: EndMs = %d, want 600", shifted[1].EndMs)
	}
	for i := 2; i < len(words); i++ {
		if want := words[i].StartMs + 300; shifted[i].StartMs != want {
			t.Errorf("word %d StartMs = %d, want %d", i, shifted[i].StartMs, want)
		}
	}
}

// Two insertions must accumulate: the third sentence shifts by both.
func TestPlanSentencePausesAccumulate(t *testing.T) {
	words := []AlignedWord{
		{Word: "a.", StartMs: 0, EndMs: 100},
		{Word: "b.", StartMs: 200, EndMs: 300},
		{Word: "c", StartMs: 400, EndMs: 500},
	}
	inserts, shifted := planSentencePauses(words, 400, 500)
	if len(inserts) != 2 {
		t.Fatalf("got %d insert(s), want 2", len(inserts))
	}
	// Each gap is 100, so each needs 300.
	if shifted[2].StartMs != 400+600 {
		t.Errorf("final word StartMs = %d, want 1000", shifted[2].StartMs)
	}
	// The plan's positions stay on the ORIGINAL timeline — ffmpeg trims the
	// unmodified file, so a shifted position would cut in the wrong place.
	if inserts[1].AtMs != 350 {
		t.Errorf("second insert AtMs = %d, want 350 (original timeline)", inserts[1].AtMs)
	}
}

// The last word is never a boundary: a pause after it is just dead video.
func TestPlanSentencePausesSkipsFinalWord(t *testing.T) {
	words := []AlignedWord{
		{Word: "only.", StartMs: 0, EndMs: 100},
	}
	inserts, _ := planSentencePauses(words, 400, 500)
	if len(inserts) != 0 {
		t.Errorf("got %d insert(s) for a single-word track, want 0", len(inserts))
	}
}

// The cap keeps a planned gap under what long-gap compression would cut back
// out, so the pause survives the next align run instead of oscillating.
func TestPlanSentencePausesRespectsCap(t *testing.T) {
	words := []AlignedWord{
		{Word: "a.", StartMs: 0, EndMs: 100},
		{Word: "b", StartMs: 150, EndMs: 250},
	}
	inserts, _ := planSentencePauses(words, 900, 500)
	if len(inserts) != 1 {
		t.Fatalf("got %d insert(s), want 1", len(inserts))
	}
	if got := 50 + inserts[0].AddMs; got != 500 {
		t.Errorf("resulting gap = %d, want capped at 500", got)
	}
}

func TestPlanSentencePausesDisabled(t *testing.T) {
	words := []AlignedWord{
		{Word: "a.", StartMs: 0, EndMs: 100},
		{Word: "b", StartMs: 150, EndMs: 250},
	}
	inserts, shifted := planSentencePauses(words, 0, 500)
	if len(inserts) != 0 {
		t.Errorf("got %d insert(s) with the floor off, want 0", len(inserts))
	}
	if shifted[1].StartMs != 150 {
		t.Errorf("timeline moved with the floor off: %d", shifted[1].StartMs)
	}
}

func TestShiftForInserts(t *testing.T) {
	inserts := []silenceInsert{{AtMs: 100, AddMs: 50}, {AtMs: 500, AddMs: 200}}
	cases := []struct{ in, want int }{
		{0, 0},
		{99, 99},
		{100, 150},   // at the boundary the insert has already landed
		{300, 350},   // between the two
		{1000, 1250}, // past both
	}
	for _, c := range cases {
		if got := shiftForInserts(c.in, inserts); got != c.want {
			t.Errorf("shiftForInserts(%d) = %d, want %d", c.in, got, c.want)
		}
	}
}

// Section spans must ride the same shift as the words they bracket, or the
// scene graph would cue visuals against a timeline the audio no longer has.
func TestShiftSpansKeepsWordIndices(t *testing.T) {
	spans := []SectionSpan{{ID: "s1", StartMs: 0, EndMs: 400, WordStart: 0, WordEnd: 5, WER: 0.1}}
	inserts := []silenceInsert{{AtMs: 200, AddMs: 300}}
	out := shiftSpans(spans, inserts)

	if out[0].EndMs != 700 {
		t.Errorf("EndMs = %d, want 700", out[0].EndMs)
	}
	if out[0].StartMs != 0 {
		t.Errorf("StartMs = %d, want 0", out[0].StartMs)
	}
	if out[0].WordStart != 0 || out[0].WordEnd != 5 || out[0].WER != 0.1 {
		t.Errorf("inserting silence changed word indices or WER: %+v", out[0])
	}
	if spans[0].EndMs != 400 {
		t.Error("shiftSpans mutated its input")
	}
}

// The filter graph is the part unit tests cannot reason about: anullsrc as an
// in-graph source, and concat refusing to join streams whose formats differ.
// This runs it against real ffmpeg and checks the audio actually grew by the
// planned amount.
func TestInsertSilenceGrowsAudioByPlannedAmount(t *testing.T) {
	requireFFmpeg(t)
	env, _ := runEnv(t, &fakeRouter{})

	path := filepath.Join(t.TempDir(), "voice.wav")
	if err := os.WriteFile(path, makeWAV(2), 0o644); err != nil {
		t.Fatal(err)
	}
	before, err := wavDuration(path)
	if err != nil {
		t.Fatal(err)
	}

	inserts := []silenceInsert{{AtMs: 500, AddMs: 400}, {AtMs: 1500, AddMs: 300}}
	if err := insertSilence(context.Background(), env, path, inserts); err != nil {
		t.Fatalf("insertSilence: %v", err)
	}

	after, err := wavDuration(path)
	if err != nil {
		t.Fatalf("result is not readable WAV: %v", err)
	}
	want := before + 700*time.Millisecond
	if diff := (after - want).Abs(); diff > 60*time.Millisecond {
		t.Errorf("duration = %v, want ~%v (was %v)", after, want, before)
	}
}

// Alignment and audio must move together: whatever the plan adds to the file,
// the same amount has to land in every timestamp the renderer reads.
func TestApplySentencePausesKeepsAudioAndTimestampsInStep(t *testing.T) {
	requireFFmpeg(t)
	env, _ := runEnv(t, &fakeRouter{})

	path := filepath.Join(t.TempDir(), "voice.wav")
	if err := os.WriteFile(path, makeWAV(3), 0o644); err != nil {
		t.Fatal(err)
	}
	before, err := wavDuration(path)
	if err != nil {
		t.Fatal(err)
	}

	a := &Alignment{
		Words: []AlignedWord{
			{Word: "one", StartMs: 0, EndMs: 400},
			{Word: "two", StartMs: 450, EndMs: 900},
			{Word: "three", StartMs: 1000, EndMs: 1400},
		},
		Sections: []SectionSpan{{ID: "s1", StartMs: 0, EndMs: 1400, WordStart: 0, WordEnd: 3}},
		// The caption track carries the author's punctuation, so the boundary
		// is found here even though the raw transcript above has none.
		DisplayWords: []AlignedWord{
			{Word: "One", StartMs: 0, EndMs: 400},
			{Word: "two.", StartMs: 450, EndMs: 900},
			{Word: "Three", StartMs: 1000, EndMs: 1400},
		},
		DisplaySections: []SectionSpan{{ID: "s1", StartMs: 0, EndMs: 1400, WordStart: 0, WordEnd: 3}},
	}

	inserts, err := applySentencePauses(context.Background(), env, a, path, 400)
	if err != nil {
		t.Fatalf("applySentencePauses: %v", err)
	}
	if len(inserts) != 1 {
		t.Fatalf("got %d insert(s), want 1", len(inserts))
	}
	added := inserts[0].AddMs
	if added != 300 { // 400 floor - 100 natural gap
		t.Fatalf("AddMs = %d, want 300", added)
	}

	after, err := wavDuration(path)
	if err != nil {
		t.Fatal(err)
	}
	if diff := (after - before - time.Duration(added)*time.Millisecond).Abs(); diff > 60*time.Millisecond {
		t.Errorf("audio grew by %v, want %dms", after-before, added)
	}

	// Both tracks and both span lists ride the same shift.
	if a.Words[2].StartMs != 1300 {
		t.Errorf("transcript word not shifted: %d, want 1300", a.Words[2].StartMs)
	}
	if a.DisplayWords[2].StartMs != 1300 {
		t.Errorf("caption word not shifted: %d, want 1300", a.DisplayWords[2].StartMs)
	}
	if a.Words[1].EndMs != 900 {
		t.Errorf("word before the pause moved: %d, want 900", a.Words[1].EndMs)
	}
	if a.Sections[0].EndMs != 1700 || a.DisplaySections[0].EndMs != 1700 {
		t.Errorf("section spans = %d/%d, want 1700", a.Sections[0].EndMs, a.DisplaySections[0].EndMs)
	}
}
