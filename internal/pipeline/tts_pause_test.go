package pipeline

import "testing"

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
