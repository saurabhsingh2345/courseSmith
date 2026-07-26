package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

type fakeAligner struct {
	words []AlignedWord
	err   error
}

func (f *fakeAligner) Align(_ context.Context, _ string) ([]AlignedWord, error) {
	return f.words, f.err
}

// wordSeq builds evenly spaced AlignedWords: each word 200ms with a 50ms gap.
func wordSeq(startMs int, words ...string) []AlignedWord {
	out := make([]AlignedWord, len(words))
	at := startMs
	for i, w := range words {
		out[i] = AlignedWord{Word: w, StartMs: at, EndMs: at + 200}
		at += 250
	}
	return out
}

func TestPlanGapCompression(t *testing.T) {
	tests := []struct {
		name      string
		words     []AlignedWord
		wantCuts  []audioCut
		wantShift []int // expected StartMs per word after adjustment
	}{
		{
			name:      "no gaps",
			words:     wordSeq(0, "a", "b", "c"),
			wantCuts:  nil,
			wantShift: []int{0, 250, 500},
		},
		{
			name: "one long gap",
			words: []AlignedWord{
				{Word: "a", StartMs: 0, EndMs: 500},
				{Word: "b", StartMs: 3000, EndMs: 3400}, // 2500ms gap
			},
			// keep 250ms each side of the cut; remove 2000ms.
			wantCuts:  []audioCut{{FromMs: 750, ToMs: 2750}},
			wantShift: []int{0, 1000},
		},
		{
			name: "two gaps accumulate",
			words: []AlignedWord{
				{Word: "a", StartMs: 0, EndMs: 500},
				{Word: "b", StartMs: 2500, EndMs: 3000}, // 2000ms gap → remove 1500
				{Word: "c", StartMs: 6000, EndMs: 6500}, // 3000ms gap → remove 2500
			},
			wantCuts: []audioCut{
				{FromMs: 750, ToMs: 2250},
				{FromMs: 3250, ToMs: 5750},
			},
			wantShift: []int{0, 1000, 2000},
		},
		{
			name: "gap exactly at threshold is kept",
			words: []AlignedWord{
				{Word: "a", StartMs: 0, EndMs: 500},
				{Word: "b", StartMs: 2000, EndMs: 2400}, // 1500ms = maxGap, not >
			},
			wantCuts:  nil,
			wantShift: []int{0, 2000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cuts, adjusted := planGapCompression(tt.words, maxGapMs, keepGapMs)
			if fmt.Sprint(cuts) != fmt.Sprint(tt.wantCuts) {
				t.Errorf("cuts = %v, want %v", cuts, tt.wantCuts)
			}
			for i, want := range tt.wantShift {
				if adjusted[i].StartMs != want {
					t.Errorf("word %d StartMs = %d, want %d", i, adjusted[i].StartMs, want)
				}
				origDur := tt.words[i].EndMs - tt.words[i].StartMs
				if got := adjusted[i].EndMs - adjusted[i].StartMs; got != origDur {
					t.Errorf("word %d duration changed: %d → %d", i, origDur, got)
				}
			}
		})
	}
}

func TestWordsFromSegments(t *testing.T) {
	segs := []llm.TranscriptSegment{
		{Start: 0, End: 2, Text: " Hello there world."},
		{Start: 2.5, End: 4, Text: "Second segment"},
	}
	words := wordsFromSegments(segs)
	if len(words) != 5 {
		t.Fatalf("got %d words, want 5", len(words))
	}
	if words[0].StartMs != 0 || words[0].Word != "Hello" {
		t.Errorf("first word = %+v", words[0])
	}
	if words[3].StartMs != 2500 {
		t.Errorf("second segment starts at %d, want 2500", words[3].StartMs)
	}
	for i := 1; i < len(words); i++ {
		if words[i].StartMs < words[i-1].StartMs {
			t.Errorf("word times not monotonic at %d: %+v", i, words)
		}
	}
	// Word ends stay within their segment.
	if words[2].EndMs > 2000+1 {
		t.Errorf("segment 1 word overruns: %+v", words[2])
	}
}

func TestMatchSections(t *testing.T) {
	refs := []sectionRef{
		{ID: "one", Text: "Python is easy to learn."},
		{ID: "two", Text: "Now let us try it."},
	}

	t.Run("perfect read", func(t *testing.T) {
		words := wordSeq(0, "Python", "is", "easy", "to", "learn.", "Now", "let", "us", "try", "it.")
		spans, overall, misreads := matchSections(refs, words)
		if overall != 0 {
			t.Errorf("overall WER = %v, want 0", overall)
		}
		if len(misreads) != 0 {
			t.Errorf("misreads = %+v, want none", misreads)
		}
		if spans[0].WordStart != 0 || spans[0].WordEnd != 5 || spans[0].WER != 0 {
			t.Errorf("span one = %+v", spans[0])
		}
		if spans[1].WordStart != 5 || spans[1].WordEnd != 10 {
			t.Errorf("span two = %+v", spans[1])
		}
		if spans[0].StartMs != 0 || spans[1].StartMs != words[5].StartMs {
			t.Errorf("span times: %+v %+v", spans[0], spans[1])
		}
	})

	t.Run("misread words flag the right section", func(t *testing.T) {
		// Kokoro misreads two words of section two.
		words := wordSeq(0, "Python", "is", "easy", "to", "learn.", "Cow", "yet", "us", "try", "it.")
		spans, overall, misreads := matchSections(refs, words)
		if spans[0].WER != 0 {
			t.Errorf("section one WER = %v, want 0", spans[0].WER)
		}
		if spans[1].WER <= 0.3 {
			t.Errorf("section two WER = %v, want > 0.3", spans[1].WER)
		}
		if overall <= 0 || overall >= 0.5 {
			t.Errorf("overall = %v", overall)
		}
		if len(misreads) != 2 {
			t.Fatalf("misreads = %+v, want 2", misreads)
		}
		if misreads[0].Section != "two" || misreads[0].Ref != "Now" || misreads[0].Heard != "Cow" {
			t.Errorf("misread[0] = %+v", misreads[0])
		}
		if misreads[1].Ref != "let" || misreads[1].Heard != "yet" {
			t.Errorf("misread[1] = %+v", misreads[1])
		}
	})

	t.Run("punctuation and case are ignored", func(t *testing.T) {
		words := wordSeq(0, "python", "IS", "easy,", "to", "LEARN", "now", "let", "us", "try", "it")
		_, overall, _ := matchSections(refs, words)
		if overall != 0 {
			t.Errorf("overall WER = %v, want 0 (normalization)", overall)
		}
	})
}

func TestPlanAutoFixes(t *testing.T) {
	dict := map[string]string{"PyPI": "pie pee eye", "Groq": "grock"}
	misreads := []Misread{
		{Section: "s", Ref: "PyPI,", Heard: "peepee"},    // in dict (punctuation stripped)
		{Section: "s", Ref: "Anaconda", Heard: "banana"}, // not in dict
	}
	// The spoken text still contains "PyPI" verbatim — prep missed it.
	spoken := map[string]string{"s": "Install from PyPI, then check Anaconda."}
	fixes := planAutoFixes(misreads, dict, nil, spoken)
	if len(fixes) != 1 || fixes["PyPI"] != "pie pee eye" {
		t.Errorf("fixes = %+v", fixes)
	}
	// Already-applied fixes are not re-planned.
	fixes = planAutoFixes(misreads, dict, map[string]string{"PyPI": "pie pee eye"}, spoken)
	if len(fixes) != 0 {
		t.Errorf("fixes = %+v, want none when already applied", fixes)
	}
	// When prep already rewrote the word, re-running cannot help — no fix.
	fixes = planAutoFixes(misreads, dict, nil, map[string]string{"s": "Install from pie pee eye, then check Anaconda."})
	if len(fixes) != 0 {
		t.Errorf("fixes = %+v, want none when prep already applied", fixes)
	}
}

func TestExpandTokens(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"python.org/downloads", "python org downloads"},
		{"Hello,", "hello"},
		{"1991", "one thousand nine hundred ninety one"},
		{"3", "three"},
		{"dot", ""},   // spoken separator drops from both sides
		{"slash", ""}, // spoken separator drops from both sides
		{"ninety-one,", "ninety one"},
		{"macOS", "macos"},
	}
	for _, tt := range tests {
		if got := strings.Join(expandTokens(tt.in), " "); got != tt.want {
			t.Errorf("expandTokens(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestBuildPaceReport(t *testing.T) {
	refs := []sectionRef{
		{ID: "on-pace", Text: strings.Repeat("word ", 150)},  // 150 words
		{ID: "rushed", Text: strings.Repeat("word ", 150)},   // same words, half the time
	}
	spans := []SectionSpan{
		{ID: "on-pace", StartMs: 0, EndMs: 60000},   // 150 wpm
		{ID: "rushed", StartMs: 60000, EndMs: 90000}, // 300 wpm
	}
	report := buildPaceReport(refs, spans, 150)
	if report.Sections[0].Flagged {
		t.Errorf("on-pace section flagged: %+v", report.Sections[0])
	}
	if !report.Sections[1].Flagged || report.Sections[1].WPM < 290 {
		t.Errorf("rushed section not flagged: %+v", report.Sections[1])
	}
	if len(report.Flagged) != 1 || report.Flagged[0] != "rushed" {
		t.Errorf("flagged = %v", report.Flagged)
	}
}

func TestAlignStageWithAligner(t *testing.T) {
	requireFFmpeg(t)
	course, lesson := testCourse(t)
	seedScript(t, lesson) // narrations: "Python reads code line by line." / "Now let us try it live."

	// 6s of real audio; the aligner reports a 2.5s silence in the middle.
	if err := os.MkdirAll(lesson.GeneratedDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(lesson.GeneratedDir(), VoiceoverFileName), makeWAV(6), 0o644); err != nil {
		t.Fatal(err)
	}
	words := append(
		wordSeq(0, "Python", "reads", "code", "line", "by", "line."),
		wordSeq(4000, "Now", "let", "us", "try", "it", "live.")..., // gap 4000-1450=2550ms
	)
	env, out := runEnv(t, &fakeRouter{})
	env.Aligner = &fakeAligner{words: words}

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageAlign}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), AlignmentFileName))
	if err != nil {
		t.Fatal(err)
	}
	var alignment Alignment
	if err := json.Unmarshal(data, &alignment); err != nil {
		t.Fatal(err)
	}
	if alignment.Source != AlignSourceWhisperX {
		t.Errorf("source = %q", alignment.Source)
	}
	if len(alignment.Words) != 12 {
		t.Fatalf("words = %d, want 12", len(alignment.Words))
	}
	// The 2550ms gap was compressed to 500ms: second half shifted left 2050ms.
	if got := alignment.Words[6].StartMs; got != 4000-2050 {
		t.Errorf("first word of section two at %dms, want %d", got, 4000-2050)
	}
	if len(alignment.Sections) != 2 || alignment.Sections[1].ID != "second-idea" {
		t.Errorf("sections = %+v", alignment.Sections)
	}
	if alignment.Sections[0].WER != 0 || alignment.Sections[1].WER != 0 {
		t.Errorf("WER nonzero: %+v", alignment.Sections)
	}

	// The audio itself must have shrunk by ~2.05s.
	dur, err := wavDuration(filepath.Join(lesson.GeneratedDir(), VoiceoverFileName))
	if err != nil {
		t.Fatal(err)
	}
	want := 6*time.Second - 2050*time.Millisecond
	if diff := (dur - want).Abs(); diff > 100*time.Millisecond {
		t.Errorf("compressed voiceover = %v, want ~%v", dur, want)
	}

	if _, err := os.Stat(filepath.Join(lesson.GeneratedDir(), ReviewsDirName, TTSAccuracyFileName)); err != nil {
		t.Errorf("tts accuracy report missing: %v", err)
	}
	if !strings.Contains(out.String(), "compressing 1 silence gap") {
		t.Errorf("output missing compression notice:\n%s", out.String())
	}
}

func TestAlignStageFallbackToSegments(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson)
	seedVoiceover(t, lesson)

	env, out := runEnv(t, &fakeRouter{})
	env.Transcriber = &fakeTranscriber{result: &llm.Transcription{
		Text: "Python reads code line by line. Now let us try it live.",
		Segments: []llm.TranscriptSegment{
			{Start: 0, End: 0.5, Text: " Python reads code line by line."},
			{Start: 0.5, End: 1, Text: " Now let us try it live."},
		},
	}}

	if err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageAlign}); err != nil {
		t.Fatal(err)
	}
	var alignment Alignment
	data, err := os.ReadFile(filepath.Join(lesson.GeneratedDir(), AlignmentFileName))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &alignment); err != nil {
		t.Fatal(err)
	}
	if alignment.Source != AlignSourceEstimated {
		t.Errorf("source = %q, want fallback", alignment.Source)
	}
	if !strings.Contains(out.String(), "whisperX not installed") {
		t.Errorf("output missing fallback warning:\n%s", out.String())
	}
}

func TestAlignStageNeedsSomeAligner(t *testing.T) {
	course, lesson := testCourse(t)
	seedScript(t, lesson)
	seedVoiceover(t, lesson)

	env, _ := runEnv(t, &fakeRouter{}) // no Aligner, no Transcriber
	err := env.RunLesson(context.Background(), course, lesson, RunOptions{Stage: project.StageAlign})
	if err == nil || !strings.Contains(err.Error(), "uv sync") {
		t.Errorf("error = %v, want install instructions", err)
	}
}

func TestSubprocessAligner(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "fake-align")
	body := "#!/bin/sh\necho '{\"words\":[{\"word\":\"hi\",\"startMs\":0,\"endMs\":300}]}'\n"
	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	a := &SubprocessAligner{Cmd: []string{script}}
	words, err := a.Align(context.Background(), "unused.wav")
	if err != nil {
		t.Fatal(err)
	}
	if len(words) != 1 || words[0].Word != "hi" || words[0].EndMs != 300 {
		t.Errorf("words = %+v", words)
	}

	failing := filepath.Join(dir, "fake-fail")
	if err := os.WriteFile(failing, []byte("#!/bin/sh\necho 'model not found' >&2\nexit 3\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	a = &SubprocessAligner{Cmd: []string{failing}}
	if _, err := a.Align(context.Background(), "unused.wav"); err == nil || !strings.Contains(err.Error(), "model not found") {
		t.Errorf("error = %v, want stderr surfaced", err)
	}
}

func TestBuildDisplayWords(t *testing.T) {
	refs := []sectionRef{
		{ID: "one", Text: "Check the Add Python to PATH box during install."},
		{ID: "two", Text: "Visit python.org to download it."},
	}
	// The ASR heard "Pathbox" for "PATH box" and "python.org" for the
	// written "python.org" (inverse-normalized), and dropped "during".
	words := wordSeq(0,
		"Check", "the", "Add", "Python", "to", "Pathbox", "install.",
		"Visit", "python.org", "to", "download", "it.")

	display, spans := buildDisplayWords(refs, words)

	var got []string
	for _, w := range display {
		got = append(got, w.Word)
	}
	want := []string{
		"Check", "the", "Add", "Python", "to", "PATH", "box", "during", "install.",
		"Visit", "python.org", "to", "download", "it.",
	}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("display words = %q, want %q", got, want)
	}
	// Monotonic, non-negative timeline.
	prev := 0
	for i, w := range display {
		if w.StartMs < prev || w.EndMs < w.StartMs {
			t.Errorf("word %d %q times not monotonic: %+v (prev end %d)", i, w.Word, w, prev)
		}
		prev = w.EndMs
	}
	// Matched words carry the ASR timing.
	if display[0].StartMs != words[0].StartMs || display[0].EndMs != words[0].EndMs {
		t.Errorf("matched word timing = %+v, want %+v", display[0], words[0])
	}
	// Section spans cover the display track contiguously.
	if len(spans) != 2 {
		t.Fatalf("spans = %+v", spans)
	}
	if spans[0].WordStart != 0 || spans[0].WordEnd != 9 {
		t.Errorf("span one = %+v", spans[0])
	}
	if spans[1].WordStart != 9 || spans[1].WordEnd != len(display) {
		t.Errorf("span two = %+v", spans[1])
	}
	if spans[0].EndMs > spans[1].StartMs {
		t.Errorf("span times overlap: %+v %+v", spans[0], spans[1])
	}
}
