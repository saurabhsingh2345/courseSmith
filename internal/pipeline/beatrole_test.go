package pipeline

import (
	"strings"
	"testing"
)

// Every template in the catalog has the same shape — name the subject, develop
// it, close on the whole — so the role comes from position rather than from a
// field the model declares. constellation is centre → spokes → whole, myth is
// claim → evidence → why, rundown is promise → items → all.
func TestRoleOf(t *testing.T) {
	cases := []struct {
		n    int
		want []beatRole
	}{
		// One beat is its own opener and nothing else; there is no "last" to be.
		{1, []beatRole{roleOpen}},
		// Two beats is one cut, which is what makes a ten-second clip a hook
		// rather than an explanation: open then land, no middle.
		{2, []beatRole{roleOpen, roleLand}},
		{3, []beatRole{roleOpen, roleDevelop, roleLand}},
		{6, []beatRole{roleOpen, roleDevelop, roleDevelop, roleDevelop, roleDevelop, roleLand}},
	}
	for _, c := range cases {
		for i, want := range c.want {
			if got := roleOf(i, c.n); got != want {
				t.Errorf("roleOf(%d, %d) = %v, want %v", i, c.n, got, want)
			}
		}
	}
}

// The flat ten-word floor was the last thing failing every run of the
// list-shaped templates, and it failed them on the beats those templates are
// designed to keep short.
func TestMinWordsForRelaxesTheEnds(t *testing.T) {
	if got := minWordsFor(roleDevelop); got != minWordsPerBeat {
		t.Errorf("develop floor = %d, want %d — the teaching beats must keep the full floor", got, minWordsPerBeat)
	}
	for _, r := range []beatRole{roleOpen, roleLand} {
		got := minWordsFor(r)
		if got != minWordsOpenLand {
			t.Errorf("role %v floor = %d, want %d", r, got, minWordsOpenLand)
		}
		if got >= minWordsPerBeat {
			t.Errorf("role %v floor (%d) is not relaxed below the develop floor (%d)", r, got, minWordsPerBeat)
		}
	}
	// Still a sentence, not a label with a voice track.
	if minWordsOpenLand < 5 {
		t.Errorf("minWordsOpenLand = %d; below about five words a beat is a caption", minWordsOpenLand)
	}
}

// The exact plans that were failing before this change: an opener and a closer
// one or two words shy of ten, with the middle beats fine. Both were the format
// being right and the rule being wrong.
func TestObservedOpenerAndCloserNowPass(t *testing.T) {
	// constellation: "No-code is visual programming, allowing anyone to build" is
	// eight words; "Together these components make technology accessible" is six.
	plan := &SnippetPlan{
		targetWords: 159,
		Beats: []SnippetBeat{
			{ID: "nocode", Narration: "No-code is visual programming, allowing anyone to build"},
			{ID: "builders", Narration: strings.Repeat("word ", 26)},
			{ID: "integrations", Narration: strings.Repeat("word ", 28)},
			{ID: "automation", Narration: strings.Repeat("word ", 27)},
			{ID: "iterations", Narration: strings.Repeat("word ", 25)},
			{ID: "whole", Narration: "Together these components make technology accessible"},
		},
	}
	if err := checkBeatShape(plan); err != nil {
		t.Errorf("a plan whose only short beats are its opener and closer was rejected: %v", err)
	}
}

// And a short beat in the MIDDLE is still rejected — that is a stated fact where
// an explanation belonged, which is the failure the floor exists to catch.
func TestShortMiddleBeatStillFails(t *testing.T) {
	plan := &SnippetPlan{
		targetWords: 159,
		Beats: []SnippetBeat{
			{ID: "open", Narration: "No-code is visual programming for everyone"},
			{ID: "middle", Narration: "It is quick"},
			{ID: "land", Narration: "Together these components make technology accessible"},
		},
	}
	err := checkBeatShape(plan)
	if err == nil {
		t.Fatal("a four-word middle beat was accepted")
	}
	if !strings.Contains(err.Error(), "middle") {
		t.Errorf("the error does not explain that the middle is held to a higher floor: %v", err)
	}
	if !strings.Contains(err.Error(), "floor 10") {
		t.Errorf("the middle beat was judged against the relaxed floor: %v", err)
	}
}

// The prompt has to ASK for variation, not merely permit it. Uniform beats were
// the commonest thing wrong with these plans and the arithmetic block is the only
// place that says so to every template at once.
func TestPlannerPromptAsksForVariedBeats(t *testing.T) {
	// The block is appended by planSnippetDefault rather than living in a prompt
	// file, so it is asserted here against the same construction.
	for _, phrase := range []string{
		"MUST NOT ALL BE THE SAME LENGTH",
		"FIRST beat",
		"MIDDLE beats",
		"LAST beat",
	} {
		if !strings.Contains(beatVariationAdvice(159, 3, 7), phrase) {
			t.Errorf("the beat-variation advice no longer says %q", phrase)
		}
	}
	// And it must quote the two different floors, or a model has no way to know
	// the ends are allowed to be shorter.
	adv := beatVariationAdvice(159, 3, 7)
	if !strings.Contains(adv, "6 words") {
		t.Error("the advice does not tell the model the ends may be short")
	}
}
