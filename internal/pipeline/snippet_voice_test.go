package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func planWithNarration(narrations ...string) *SnippetPlan {
	p := &SnippetPlan{Title: "A title"}
	for i, n := range narrations {
		p.Beats = append(p.Beats, SnippetBeat{
			ID:        string(rune('a' + i)),
			Heading:   "Heading",
			Narration: n,
		})
	}
	return p
}

func TestSpokenVoiceRejectsWrittenPunctuation(t *testing.T) {
	for _, tc := range []struct{ name, narration, want string }{
		{
			"semicolon",
			"ANN indexes trade perfect recall for speed; recall and latency vary.",
			"semicolon",
		},
		{
			"em dash",
			"A vector database stores embeddings — numerical vectors it can compare.",
			"dash",
		},
		{
			"en dash",
			"It finds the closest vectors – not the matching words.",
			"dash",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSpokenVoice(planWithNarration(tc.narration))
			if err == nil {
				t.Fatalf("accepted unspeakable narration: %q", tc.narration)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error does not name the problem (%q): %v", tc.want, err)
			}
			// The correction round needs to know WHICH beat, or the model rewrites
			// the wrong one — the plan is up to eight beats long.
			if !strings.Contains(err.Error(), `beat "a"`) {
				t.Errorf("error does not name the beat: %v", err)
			}
		})
	}
}

func TestSpokenVoiceAcceptsOrdinaryNarration(t *testing.T) {
	// Real sentences from shipped narration in this repo.
	ok := planWithNarration(
		"Paste the error. Paste the relevant part. Say which screen you are on.",
		"One change. Look at the result. Say the next thing.",
		"You are not the expert on the code. You are the expert on what is wrong.",
		"It finds the closest vectors, not the matching words.",
	)
	if err := validateSpokenVoice(ok); err != nil {
		t.Fatalf("rejected ordinary spoken narration: %v", err)
	}
}

// The rules were admitted on measured separation against shipped narration, and
// they stay admitted only while that holds. If the corpus drifts — a course
// written in a denser register, a template that legitimately needs a dash — this
// fails and the rule gets re-argued rather than quietly costing a correction
// round on every clip.
//
// The threshold is deliberately loose (5%): the measured rate was ~1% and the
// point is to catch a regime change, not to pin a number.
func TestVoiceRulesStayCalibratedAgainstShippedNarration(t *testing.T) {
	root := repoRootFromTest(t)
	scripts, err := filepath.Glob(filepath.Join(root, "courses", "*", "lessons", "*", "generated", "script.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) == 0 {
		t.Skip("no generated scripts in this checkout to calibrate against")
	}

	var total, flagged int
	for _, f := range scripts {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var script Script
		if err := json.Unmarshal(b, &script); err != nil {
			continue
		}
		for _, s := range script.Sections {
			for _, sent := range splitSentencesForTest(s.Narration) {
				total++
				if strings.ContainsAny(sent, ";—–") || hedges.MatchString(sent) {
					flagged++
				}
			}
		}
	}
	if total == 0 {
		t.Skip("no narration found")
	}
	rate := 100 * float64(flagged) / float64(total)
	t.Logf("voice rules would flag %d/%d = %.2f%% of shipped narration", flagged, total, rate)
	if rate > 5 {
		t.Errorf("the voice rules now fire on %.2f%% of shipped narration (%d/%d). They were admitted at ~1%%; at this rate they cost a correction round on ordinary work and should be re-argued, not left in", rate, flagged, total)
	}
}

// splitSentencesForTest breaks on sentence-ending punctuation. Go's regexp is
// RE2 and has no lookbehind, so the terminator is captured and re-attached
// rather than matched from behind.
func splitSentencesForTest(s string) []string {
	var out []string
	var cur strings.Builder
	for _, r := range s {
		cur.WriteRune(r)
		if r == '.' || r == '!' || r == '?' {
			if p := strings.TrimSpace(cur.String()); p != "" {
				out = append(out, p)
			}
			cur.Reset()
		}
	}
	if p := strings.TrimSpace(cur.String()); p != "" {
		out = append(out, p)
	}
	return out
}

// repoRootFromTest walks up from this test file to the module root.
func repoRootFromTest(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the test file")
	}
	dir := filepath.Dir(file)
	for range 6 {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	t.Fatal("no go.mod above the test file")
	return ""
}

func TestSpokenVoiceAdviceReachesEveryTemplate(t *testing.T) {
	advice := spokenVoiceAdvice()
	for _, want := range []string{"semicolon", "dash", "ELEVEN", "hedge"} {
		if !strings.Contains(advice, want) {
			t.Errorf("shared voice advice does not mention %q", want)
		}
	}
	// The advice must name the two ENFORCED rules, or a model reads the rejection
	// as arbitrary when it has never been told.
	if !strings.Contains(advice, "rejected") {
		t.Error("the advice does not say that the punctuation rules are enforced")
	}
}
