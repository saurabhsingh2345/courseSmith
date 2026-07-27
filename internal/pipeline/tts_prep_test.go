package pipeline

import (
	"strings"
	"testing"
)

func TestPrepForSpeech(t *testing.T) {
	dict := SpeechDict(nil)
	tests := []struct {
		name string
		in   string
		want string
	}{
		// --- code-token dictionary ---
		{"f-string", "An f-string makes formatting easy.", "An eff string makes formatting easy."},
		{"f-strings plural", "We love f-strings here.", "We love eff strings here."},
		{"dunder init", "Python calls __init__ automatically.", "Python calls dunder init automatically."},
		{"dunder name", "Check __name__ before running.", "Check dunder name before running."},
		{"dunder main", "The __main__ module starts things.", "The dunder main module starts things."},
		{"pypi", "Install packages from PyPI today.", "Install packages from pie pee eye today."},
		{"pypi lowercase", "search pypi for it", "search pie pee eye for it"},
		{"python.org", "Download it from python.org now.", "Download it from python dot org now."},
		{"pip not pipeline", "Use pip inside your pipeline.", "Use pip inside your pipeline."},
		{"repl", "Open the REPL and type.", "Open the repple and type."},
		{"file extension", "Save it as a .py file.", "Save it as a dot pie file."},
		{"pep8", "Read PEP 8 for style.", "Read pep eight for style."},
		{"utf8", "Text is stored as utf-8 bytes.", "Text is stored as you tee eff eight bytes."},
		{"cpython", "CPython is the reference implementation.", "see python is the reference implementation."},
		{"user extension wins", "", ""}, // placeholder; handled below
		{"error names", "A TypeError means mixed types.", "A type error means mixed types."},
		{"contraction", "So you wanna code?", "So you want to code?"},
		{"python3", "Run python3 now.", "Run python three now."},
		{"no match inside words", "The integer count grows.", "The integer count grows."},

		// --- symbols and operators ---
		{"not equal", "Here x != y is true.", "Here x not equal to y is true."},
		{"double equals", "Use == to compare values.", "Use equals equals to compare values."},
		{"less or equal", "Check a <= b first.", "Check a less than or equal to b first."},
		{"empty list", "An empty [] means no items.", "An empty square brackets means no items."},
		{"empty dict", "Write {} for an empty dictionary.", "Write curly braces for an empty dictionary."},
		{"power", "Compute 2**8 in your head.", "Compute two to the power of eight in your head."},
		{"power spaced", "So 2 ** 10 is big.", "So two to the power of ten is big."},
		{"percent", "About 50% of learners quit.", "About fifty percent of learners quit."},

		// --- numbers ---
		{"version", "Python 3.11 is faster.", "Python three point eleven is faster."},
		{"version leading zero", "Version 3.05 was odd.", "Version three point zero five was odd."},
		{"small int", "You have 7 items.", "You have seven items."},
		{"teens", "There are 14 lines.", "There are fourteen lines."},
		{"tens", "Wait 90 seconds.", "Wait ninety seconds."},
		{"compound", "It ran 42 times.", "It ran forty-two times."},
		{"hundreds", "About 250 packages exist.", "About two hundred fifty packages exist."},
		{"thousands", "Over 3000 downloads happened.", "Over three thousand downloads happened."},
		{"millions", "Nearly 2000000 users agree.", "Nearly two million users agree."},
		{"zero", "It returns 0 on success.", "It returns zero on success."},

		// --- emphasis ---
		{"emphasis mid", "This is *really* important.", "This is, really, important."},
		{"emphasis start", "*Never* ignore errors.", "Never, ignore errors."},
		{"bold emphasis", "That is **the** key idea.", "That is, the, key idea."},

		// --- combinations ---
		{"dict then number", "Python 3.12 ships pip 24.", "Python three point twelve ships pip twenty-four."},
		{"operator with numbers", "Since 10 != 9, it prints.", "Since ten not equal to nine, it prints."},
	}
	for _, tt := range tests {
		if tt.in == "" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := PrepForSpeech(tt.in, dict); got != tt.want {
				t.Errorf("PrepForSpeech(%q)\n got: %q\nwant: %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestPrepForSpeechUserPronunciations(t *testing.T) {
	dict := SpeechDict(map[string]string{
		"Groq":    "grock",
		"CPython": "c python override",
	})
	if got := PrepForSpeech("Groq is fast.", dict); got != "grock is fast." {
		t.Errorf("user pronunciation not applied: %q", got)
	}
	if got := PrepForSpeech("CPython rules.", dict); got != "c python override rules." {
		t.Errorf("user override did not beat the default: %q", got)
	}
}

func TestGuardSentenceLength(t *testing.T) {
	// 38 words with a comma near the middle: must be split into two sentences.
	long := "Python is a language that reads almost like English, and because it reads almost like English it is the language most schools and universities now choose when they introduce complete beginners to the craft of programming."
	got := guardSentenceLength(long, 30)
	sentences := splitSentences(got)
	if len(sentences) != 2 {
		t.Fatalf("split into %d sentences, want 2: %q", len(sentences), got)
	}
	for _, s := range sentences {
		if n := len(strings.Fields(s)); n > 30 {
			t.Errorf("sentence still has %d words: %q", n, s)
		}
	}
	if !strings.HasPrefix(sentences[1], "And") {
		t.Errorf("second half not capitalized: %q", sentences[1])
	}

	short := "This one is fine."
	if got := guardSentenceLength(short, 30); got != short {
		t.Errorf("short sentence modified: %q", got)
	}

	// No clause boundary at all: left alone rather than mangled.
	noComma := strings.Repeat("word ", 35)
	noComma = strings.TrimSpace(noComma) + "."
	if got := guardSentenceLength(noComma, 30); got != noComma {
		t.Errorf("boundary-less sentence modified: %q", got)
	}
}

func TestSplitParagraphs(t *testing.T) {
	got := SplitParagraphs("First paragraph.\n\nSecond paragraph.\n\n\nThird.")
	if len(got) != 3 || got[0] != "First paragraph." || got[2] != "Third." {
		t.Errorf("SplitParagraphs = %q", got)
	}
	if got := SplitParagraphs("only one"); len(got) != 1 || got[0] != "only one" {
		t.Errorf("single paragraph = %q", got)
	}
}

func TestStripEmphasis(t *testing.T) {
	if got := StripEmphasis("This is *really* important."); got != "This is really important." {
		t.Errorf("StripEmphasis = %q", got)
	}
}
