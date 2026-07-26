package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// AlignmentFileName is the Stage "align" output in the lesson's generated dir.
const AlignmentFileName = "alignment.json"

// TTSAccuracyFileName is the WER QA report, under generated/reviews/.
const TTSAccuracyFileName = "tts_accuracy.json"

// Alignment sources.
const (
	AlignSourceWhisperX  = "whisperx"           // word-level, precise
	AlignSourceEstimated = "segments-estimated" // Groq segments, word times interpolated
)

// Silence handling: inter-word gaps longer than maxGapMs are compressed down
// to keepGapMs by cutting audio.
const (
	maxGapMs  = 1500
	keepGapMs = 500
)

// werFlagThreshold flags sections whose word error rate suggests the TTS
// misread the script.
const werFlagThreshold = 0.05

// AlignedWord is one spoken word with millisecond timestamps.
type AlignedWord struct {
	Word    string `json:"word"`
	StartMs int    `json:"startMs"`
	EndMs   int    `json:"endMs"`
}

// SectionSpan locates one script section inside the aligned word stream.
type SectionSpan struct {
	ID        string  `json:"id"`
	StartMs   int     `json:"startMs"`
	EndMs     int     `json:"endMs"`
	WordStart int     `json:"wordStart"` // index into Alignment.Words
	WordEnd   int     `json:"wordEnd"`   // exclusive
	WER       float64 `json:"wer"`
}

// Alignment is the word-level timing map for a lesson voiceover.
type Alignment struct {
	Source   string        `json:"source"`
	Words    []AlignedWord `json:"words"`
	Sections []SectionSpan `json:"sections"`
	// DisplayWords is the caption-facing track: the WRITTEN narration words
	// (original punctuation and casing) carrying timestamps mapped over from
	// the transcribed words. Captions built from raw ASR words show
	// recognition errors ("PATH box" → "Pathbox"); captions built from this
	// track always show what the author wrote.
	DisplayWords []AlignedWord `json:"displayWords,omitempty"`
	// DisplaySections locates each script section inside DisplayWords
	// (WordStart/WordEnd index into DisplayWords; WER is unused).
	DisplaySections []SectionSpan `json:"displaySections,omitempty"`
}

// CaptionWords returns the track captions should render: the written-text
// display track when present, else the raw transcribed words.
func (a *Alignment) CaptionWords() []AlignedWord {
	if len(a.DisplayWords) > 0 {
		return a.DisplayWords
	}
	return a.Words
}

// CaptionSections returns the section spans matching CaptionWords.
func (a *Alignment) CaptionSections() []SectionSpan {
	if len(a.DisplayWords) > 0 {
		return a.DisplaySections
	}
	return a.Sections
}

// Aligner produces word-level timestamps for an audio file. nil on Env means
// whisperX is unavailable and the stage falls back to segment estimates.
type Aligner interface {
	Align(ctx context.Context, audioPath string) ([]AlignedWord, error)
}

// SubprocessAligner shells out to tools/align/align.py (whisperX).
type SubprocessAligner struct {
	// Cmd is the interpreter + script, e.g.
	// ["tools/align/.venv/bin/python", "tools/align/align.py"].
	Cmd []string
}

func (a *SubprocessAligner) Align(ctx context.Context, audioPath string) ([]AlignedWord, error) {
	if len(a.Cmd) == 0 {
		return nil, fmt.Errorf("aligner command is empty")
	}
	args := append(append([]string{}, a.Cmd[1:]...), "--audio", audioPath)
	cmd := exec.CommandContext(ctx, a.Cmd[0], args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("whisperX aligner failed: %w\n%s", err, tailLines(stderr.String(), 8))
	}
	var out struct {
		Words []AlignedWord `json:"words"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return nil, fmt.Errorf("parsing aligner output: %w", err)
	}
	if len(out.Words) == 0 {
		return nil, fmt.Errorf("aligner returned no words")
	}
	return out.Words, nil
}

// wordsFromSegments estimates word timestamps from segment-level
// transcription: each segment's duration is distributed across its words
// proportional to word length. Used when whisperX is unavailable.
func wordsFromSegments(segments []llm.TranscriptSegment) []AlignedWord {
	var words []AlignedWord
	for _, seg := range segments {
		fields := strings.Fields(seg.Text)
		if len(fields) == 0 {
			continue
		}
		total := 0
		for _, f := range fields {
			total += len(f) + 1
		}
		startMs := int(seg.Start * 1000)
		durMs := int((seg.End - seg.Start) * 1000)
		if durMs <= 0 {
			durMs = len(fields) * 300
		}
		at := startMs
		for _, f := range fields {
			w := durMs * (len(f) + 1) / total
			words = append(words, AlignedWord{Word: f, StartMs: at, EndMs: at + w})
			at += w
		}
	}
	return words
}

// audioCut is a time region removed from the voiceover.
type audioCut struct {
	FromMs int `json:"fromMs"`
	ToMs   int `json:"toMs"`
}

// planGapCompression finds inter-word silences longer than maxGap and plans
// cuts that shrink each to keep. It returns the cuts (in the original
// timeline) and the word list with timestamps shifted onto the compressed
// timeline. Words are assumed sorted by StartMs.
func planGapCompression(words []AlignedWord, maxGap, keep int) ([]audioCut, []AlignedWord) {
	var cuts []audioCut
	adjusted := make([]AlignedWord, len(words))
	removed := 0
	for i, w := range words {
		if i > 0 {
			gap := w.StartMs - words[i-1].EndMs
			if gap > maxGap {
				cut := audioCut{
					FromMs: words[i-1].EndMs + keep/2,
					ToMs:   w.StartMs - keep/2,
				}
				cuts = append(cuts, cut)
				removed += cut.ToMs - cut.FromMs
			}
		}
		adjusted[i] = AlignedWord{Word: w.Word, StartMs: w.StartMs - removed, EndMs: w.EndMs - removed}
	}
	return cuts, adjusted
}

// compressAudio removes the cut regions from a WAV file in place using
// ffmpeg atrim+concat.
func compressAudio(ctx context.Context, e *Env, path string, cuts []audioCut) error {
	if len(cuts) == 0 {
		return nil
	}
	var filter strings.Builder
	var labels []string
	cursor := 0
	seg := 0
	writeSegment := func(fromMs int, toMs int) { // toMs < 0 means "to EOF"
		label := fmt.Sprintf("[s%d]", seg)
		if toMs < 0 {
			fmt.Fprintf(&filter, "[0:a]atrim=start=%.3f,asetpts=PTS-STARTPTS%s;", float64(fromMs)/1000, label)
		} else {
			fmt.Fprintf(&filter, "[0:a]atrim=start=%.3f:end=%.3f,asetpts=PTS-STARTPTS%s;", float64(fromMs)/1000, float64(toMs)/1000, label)
		}
		labels = append(labels, label)
		seg++
	}
	for _, cut := range cuts {
		writeSegment(cursor, cut.FromMs)
		cursor = cut.ToMs
	}
	writeSegment(cursor, -1)
	fmt.Fprintf(&filter, "%sconcat=n=%d:v=0:a=1[out]", strings.Join(labels, ""), len(labels))

	tmp := path + ".compressed.wav"
	if err := e.runFFmpeg(ctx, "-y", "-i", path, "-filter_complex", filter.String(), "-map", "[out]", tmp); err != nil {
		return fmt.Errorf("compressing silence gaps: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}

// normalizeToken lowercases a word and strips everything but letters and
// digits, so "Python," and "python" compare equal. Returns "" for tokens
// that are pure punctuation.
func normalizeToken(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normalizeWords(raw []string) []string {
	out := make([]string, 0, len(raw))
	for _, w := range raw {
		if n := normalizeToken(w); n != "" {
			out = append(out, n)
		}
	}
	return out
}

// spokenFillerTokens are spoken separators dropped from BOTH sides of the
// WER comparison: the narrator says "python dot org", whisper writes
// "python.org" — after expansion and filler-dropping the two agree.
var spokenFillerTokens = map[string]bool{"dot": true, "slash": true}

var tokenSplitRe = regexp.MustCompile(`[./\-_@:]+`)
var allDigitsRe = regexp.MustCompile(`^[0-9]+$`)

// expandTokens normalizes one raw token from either side of the WER
// comparison into comparable word tokens: URLs and identifiers split into
// their words ("python.org/downloads" → python org downloads), digit runs
// are spelled out ("1991" → the words whisper heard spoken), and spoken
// separators vanish. Both reference and hypothesis go through this, so
// whisper's inverse text normalization stops reading as TTS misreads.
func expandTokens(raw string) []string {
	var out []string
	for _, part := range tokenSplitRe.Split(strings.ToLower(raw), -1) {
		var b strings.Builder
		for _, r := range part {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				b.WriteRune(r)
			}
		}
		p := b.String()
		if p == "" || spokenFillerTokens[p] {
			continue
		}
		if allDigitsRe.MatchString(p) && len(p) <= 6 {
			words := strings.FieldsFunc(numberWords(p), func(r rune) bool { return r == ' ' || r == '-' })
			out = append(out, words...)
			continue
		}
		out = append(out, p)
	}
	return out
}

// expandWords is expandTokens over a whitespace-split text.
func expandWords(text string) []string {
	var out []string
	for _, f := range strings.Fields(text) {
		out = append(out, expandTokens(f)...)
	}
	return out
}

// mergeCompounds merges adjacent tokens whose concatenation appears in the
// other side's token set ("mac" + "os" → "macos" when the reference wrote
// "macOS"), keeping the first token's metadata.
func mergeCompounds(tokens []string, meta []int, other map[string]bool) ([]string, []int) {
	var outT []string
	var outM []int
	for i := 0; i < len(tokens); i++ {
		if i+1 < len(tokens) && other[tokens[i]+tokens[i+1]] {
			outT = append(outT, tokens[i]+tokens[i+1])
			outM = append(outM, meta[i])
			i++
			continue
		}
		outT = append(outT, tokens[i])
		outM = append(outM, meta[i])
	}
	return outT, outM
}

// editOp is one step of the reference→hypothesis alignment path.
type editOp struct {
	op     byte // 'm' match, 's' substitute, 'd' delete (ref only), 'i' insert (hyp only)
	refIdx int  // valid for m/s/d
	hypIdx int  // valid for m/s/i
}

// alignWordSequences computes the Levenshtein alignment path between
// reference and hypothesis word sequences.
func alignWordSequences(ref, hyp []string) []editOp {
	n, m := len(ref), len(hyp)
	// dp[i][j] = edit distance between ref[:i] and hyp[:j].
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
		dp[i][0] = i
	}
	for j := 0; j <= m; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			cost := 1
			if ref[i-1] == hyp[j-1] {
				cost = 0
			}
			best := dp[i-1][j-1] + cost
			if v := dp[i-1][j] + 1; v < best {
				best = v
			}
			if v := dp[i][j-1] + 1; v < best {
				best = v
			}
			dp[i][j] = best
		}
	}
	// Backtrace.
	var rev []editOp
	i, j := n, m
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && dp[i][j] == dp[i-1][j-1] && ref[i-1] == hyp[j-1]:
			rev = append(rev, editOp{op: 'm', refIdx: i - 1, hypIdx: j - 1})
			i, j = i-1, j-1
		case i > 0 && j > 0 && dp[i][j] == dp[i-1][j-1]+1:
			rev = append(rev, editOp{op: 's', refIdx: i - 1, hypIdx: j - 1})
			i, j = i-1, j-1
		case i > 0 && dp[i][j] == dp[i-1][j]+1:
			rev = append(rev, editOp{op: 'd', refIdx: i - 1, hypIdx: -1})
			i--
		default:
			rev = append(rev, editOp{op: 'i', refIdx: -1, hypIdx: j - 1})
			j--
		}
	}
	// Reverse in place.
	for a, b := 0, len(rev)-1; a < b; a, b = a+1, b-1 {
		rev[a], rev[b] = rev[b], rev[a]
	}
	return rev
}

// sectionRef is one section's WER reference text — the text that was
// actually sent to the TTS (tts_script.json), not the written narration.
type sectionRef struct {
	ID   string
	Text string
}

// Misread is one word the TTS did not speak as written: the reference token
// and what the transcription heard instead ("" when dropped entirely).
type Misread struct {
	Section string `json:"section"`
	Ref     string `json:"ref"`
	Heard   string `json:"heard"`
}

// matchSections maps each section's reference text onto the aligned word
// stream and computes per-section and overall word error rates, plus the
// exact misread words.
func matchSections(refs []sectionRef, words []AlignedWord) ([]SectionSpan, float64, []Misread) {
	// Reference words per section, with a section index for every ref word.
	// refOrig keeps the pre-normalization token for misread reporting.
	var ref []string
	var refOrig []string
	var refSection []int
	for si, sec := range refs {
		for _, raw := range strings.Fields(sec.Text) {
			for _, n := range expandTokens(raw) {
				ref = append(ref, n)
				refOrig = append(refOrig, raw)
				refSection = append(refSection, si)
			}
		}
	}
	// Hypothesis words, keeping the mapping back to original indices
	// (expansion can produce several comparable tokens per aligned word, and
	// drop pure-punctuation ones).
	var hyp []string
	var hypOrig []int
	for i, w := range words {
		for _, n := range expandTokens(w.Word) {
			hyp = append(hyp, n)
			hypOrig = append(hypOrig, i)
		}
	}

	// Compound merge: whisper splits some written compounds into words
	// ("macOS" comes back as "Mac OS") and joins some spoken sequences into
	// compounds. Adjacent tokens on either side whose concatenation exists
	// on the other side compare as one token.
	refSet := make(map[string]bool, len(ref))
	for _, t := range ref {
		refSet[t] = true
	}
	hypSet := make(map[string]bool, len(hyp))
	for _, t := range hyp {
		hypSet[t] = true
	}
	hyp, hypOrig = mergeCompounds(hyp, hypOrig, refSet)
	{
		var mT []string
		var mO []string
		var mS []int
		for i := 0; i < len(ref); i++ {
			if i+1 < len(ref) && refSection[i] == refSection[i+1] && hypSet[ref[i]+ref[i+1]] {
				mT = append(mT, ref[i]+ref[i+1])
				mO = append(mO, refOrig[i])
				mS = append(mS, refSection[i])
				i++
				continue
			}
			mT = append(mT, ref[i])
			mO = append(mO, refOrig[i])
			mS = append(mS, refSection[i])
		}
		ref, refOrig, refSection = mT, mO, mS
	}
	refCounts := make([]int, len(refs))
	for _, si := range refSection {
		refCounts[si]++
	}

	ops := alignWordSequences(ref, hyp)
	errCounts := make([]int, len(refs))
	firstHyp := make([]int, len(refs))
	lastHyp := make([]int, len(refs))
	for i := range firstHyp {
		firstHyp[i], lastHyp[i] = -1, -1
	}
	var misreads []Misread
	totalErrs := 0
	lastRefSection := 0
	for _, op := range ops {
		sec := lastRefSection
		if op.refIdx >= 0 {
			sec = refSection[op.refIdx]
			lastRefSection = sec
		}
		switch op.op {
		case 'm':
			if firstHyp[sec] == -1 {
				firstHyp[sec] = hypOrig[op.hypIdx]
			}
			lastHyp[sec] = hypOrig[op.hypIdx]
		case 's':
			errCounts[sec]++
			totalErrs++
			misreads = append(misreads, Misread{
				Section: refs[sec].ID,
				Ref:     refOrig[op.refIdx],
				Heard:   words[hypOrig[op.hypIdx]].Word,
			})
			if firstHyp[sec] == -1 {
				firstHyp[sec] = hypOrig[op.hypIdx]
			}
			lastHyp[sec] = hypOrig[op.hypIdx]
		case 'd':
			errCounts[sec]++
			totalErrs++
			misreads = append(misreads, Misread{Section: refs[sec].ID, Ref: refOrig[op.refIdx]})
		case 'i':
			errCounts[sec]++
			totalErrs++
		}
	}

	spans := make([]SectionSpan, len(refs))
	prevEndMs, prevWordEnd := 0, 0
	for si, sec := range refs {
		span := SectionSpan{ID: sec.ID}
		if refCounts[si] > 0 {
			span.WER = float64(errCounts[si]) / float64(refCounts[si])
		}
		if firstHyp[si] >= 0 {
			span.WordStart = firstHyp[si]
			span.WordEnd = lastHyp[si] + 1
			span.StartMs = words[firstHyp[si]].StartMs
			span.EndMs = words[lastHyp[si]].EndMs
		} else {
			// Nothing matched (all words dropped): zero-length span at the
			// previous boundary, so downstream timing stays monotonic.
			span.WordStart, span.WordEnd = prevWordEnd, prevWordEnd
			span.StartMs, span.EndMs = prevEndMs, prevEndMs
		}
		prevEndMs, prevWordEnd = span.EndMs, span.WordEnd
		spans[si] = span
	}
	overall := 0.0
	if len(ref) > 0 {
		overall = float64(totalErrs) / float64(len(ref))
	}
	return spans, overall, misreads
}

// buildDisplayWords maps the written narration onto the transcribed word
// stream: every whitespace-separated written token becomes one display word
// (original text preserved) whose timestamps come from the ASR words it
// aligned to; written words the transcription missed get times interpolated
// between their matched neighbours. Returns the display track plus one span
// per section locating it inside the track.
func buildDisplayWords(refs []sectionRef, words []AlignedWord) ([]AlignedWord, []SectionSpan) {
	// Written tokens (display form) + expanded comparable tokens.
	var orig []string
	var origSection []int
	var refExp []string
	var refWordIdx []int
	for si, sec := range refs {
		for _, raw := range strings.Fields(sec.Text) {
			wi := len(orig)
			orig = append(orig, raw)
			origSection = append(origSection, si)
			for _, n := range expandTokens(raw) {
				refExp = append(refExp, n)
				refWordIdx = append(refWordIdx, wi)
			}
		}
	}
	var hypExp []string
	var hypWordIdx []int
	for i, w := range words {
		for _, n := range expandTokens(w.Word) {
			hypExp = append(hypExp, n)
			hypWordIdx = append(hypWordIdx, i)
		}
	}
	// Compound merge both ways, mirroring matchSections, so "macOS" vs
	// "Mac OS" style splits still line up.
	refSet := make(map[string]bool, len(refExp))
	for _, t := range refExp {
		refSet[t] = true
	}
	hypSet := make(map[string]bool, len(hypExp))
	for _, t := range hypExp {
		hypSet[t] = true
	}
	hypExp, hypWordIdx = mergeCompounds(hypExp, hypWordIdx, refSet)
	refExp, refWordIdx = mergeCompounds(refExp, refWordIdx, hypSet)

	start := make([]int, len(orig))
	end := make([]int, len(orig))
	for i := range start {
		start[i], end[i] = -1, -1
	}
	for _, op := range alignWordSequences(refExp, hypExp) {
		if (op.op != 'm' && op.op != 's') || op.refIdx < 0 || op.hypIdx < 0 {
			continue
		}
		wi := refWordIdx[op.refIdx]
		w := words[hypWordIdx[op.hypIdx]]
		if start[wi] == -1 || w.StartMs < start[wi] {
			start[wi] = w.StartMs
		}
		if w.EndMs > end[wi] {
			end[wi] = w.EndMs
		}
	}

	// Interpolate unmatched runs between their known neighbours,
	// distributing the gap proportionally to word length.
	lastEnd := 0
	if len(words) > 0 {
		lastEnd = words[len(words)-1].EndMs
	}
	for i := 0; i < len(orig); i++ {
		if start[i] != -1 {
			continue
		}
		j := i
		for j < len(orig) && start[j] == -1 {
			j++
		}
		lo := 0
		if i > 0 {
			lo = end[i-1]
		}
		hi := lastEnd
		if j < len(orig) {
			hi = start[j]
		}
		if hi < lo {
			hi = lo
		}
		total := 0
		for k := i; k < j; k++ {
			total += len(orig[k]) + 1
		}
		at := lo
		for k := i; k < j; k++ {
			share := (hi - lo) * (len(orig[k]) + 1) / total
			start[k] = at
			at += share
			end[k] = at
		}
		if j > i {
			end[j-1] = hi
		}
		i = j - 1
	}
	// Enforce a monotonic timeline.
	prev := 0
	for i := range orig {
		if start[i] < prev {
			start[i] = prev
		}
		if end[i] < start[i] {
			end[i] = start[i]
		}
		prev = end[i]
	}

	display := make([]AlignedWord, len(orig))
	for i, w := range orig {
		display[i] = AlignedWord{Word: w, StartMs: start[i], EndMs: end[i]}
	}
	spans := make([]SectionSpan, len(refs))
	for si := range refs {
		spans[si] = SectionSpan{ID: refs[si].ID, WordStart: -1}
	}
	for wi, si := range origSection {
		sp := &spans[si]
		if sp.WordStart == -1 {
			sp.WordStart = wi
			sp.StartMs = display[wi].StartMs
		}
		sp.WordEnd = wi + 1
		sp.EndMs = display[wi].EndMs
	}
	for si := range spans {
		if spans[si].WordStart == -1 {
			spans[si].WordStart, spans[si].WordEnd = 0, 0
		}
	}
	return display, spans
}

// sectionAccuracy is one row of the TTS accuracy QA report.
type sectionAccuracy struct {
	ID      string  `json:"id"`
	Words   int     `json:"words"`
	Errors  int     `json:"errors"`
	WER     float64 `json:"wer"`
	Flagged bool    `json:"flagged"`
}

type ttsAccuracyReport struct {
	Source     string            `json:"source"`
	OverallWER float64           `json:"overall_wer"`
	Threshold  float64           `json:"threshold"`
	Sections   []sectionAccuracy `json:"sections"`
	Flagged    []string          `json:"flagged"`
	// Misreads lists the exact words the TTS got wrong in flagged sections.
	Misreads []Misread `json:"misreads,omitempty"`
	// AutoFixes are misread→spoken entries written to tts_fixes.json because
	// the pronunciation dictionary knows a fix; re-run the audio stage to
	// apply them.
	AutoFixes map[string]string `json:"auto_fixes,omitempty"`
	CheckedAt time.Time         `json:"checked_at"`
}

// PaceFileName is the pace verification report, under generated/reviews/.
const PaceFileName = "pace.json"

// paceTolerance is the acceptable deviation from style.pace_wpm.
const paceTolerance = 0.15

// sectionPace is one row of the pace verification report.
type sectionPace struct {
	ID         string  `json:"id"`
	Words      int     `json:"words"`
	DurationMs int     `json:"duration_ms"`
	WPM        float64 `json:"wpm"`
	Deviation  float64 `json:"deviation"` // fraction vs target, e.g. 0.22 = 22% off
	Flagged    bool    `json:"flagged"`
}

type paceReport struct {
	TargetWPM int           `json:"target_wpm"`
	Tolerance float64       `json:"tolerance"`
	Sections  []sectionPace `json:"sections"`
	Flagged   []string      `json:"flagged"`
	CheckedAt time.Time     `json:"checked_at"`
}

// buildPaceReport computes actual words-per-minute per section from the
// alignment spans and flags sections outside target ± tolerance.
func buildPaceReport(refs []sectionRef, spans []SectionSpan, targetWPM int) paceReport {
	report := paceReport{TargetWPM: targetWPM, Tolerance: paceTolerance, CheckedAt: time.Now().UTC()}
	for si, span := range spans {
		words := len(normalizeWords(strings.Fields(refs[si].Text)))
		row := sectionPace{ID: span.ID, Words: words, DurationMs: span.EndMs - span.StartMs}
		if row.DurationMs > 0 && words > 0 {
			row.WPM = float64(words) / (float64(row.DurationMs) / 60000)
		}
		if targetWPM > 0 && row.WPM > 0 {
			row.Deviation = (row.WPM - float64(targetWPM)) / float64(targetWPM)
			row.Flagged = row.Deviation > paceTolerance || row.Deviation < -paceTolerance
		}
		if row.Flagged {
			report.Flagged = append(report.Flagged, span.ID)
		}
		report.Sections = append(report.Sections, row)
	}
	return report
}

// planAutoFixes looks each misread word up in the pronunciation dictionary
// and returns the fixes to write to tts_fixes.json. A misread whose written
// form has a known spoken form gets fixed automatically; everything else
// stays in the report for a human.
//
// A fix is only planned when the written form actually survived into the
// spoken text (spokenBySection) — i.e. the prep pass missed it. When prep
// already applied the dictionary entry, re-synthesizing would produce the
// identical audio, so the misread is surfaced instead of "fixed".
func planAutoFixes(misreads []Misread, dict, existing map[string]string, spokenBySection map[string]string) map[string]string {
	fixes := map[string]string{}
	for _, m := range misreads {
		key := strings.Trim(m.Ref, ".,;:!?…\"'()[]{}")
		if key == "" {
			continue
		}
		if _, done := existing[key]; done {
			continue
		}
		if spoken, ok := spokenBySection[m.Section]; ok &&
			!strings.Contains(strings.ToLower(spoken), strings.ToLower(key)) {
			continue // prep already rewrote this word; nothing to fix by re-running
		}
		norm := normalizeToken(key)
		for from, to := range dict {
			if normalizeToken(from) == norm && to != key {
				fixes[key] = to
				break
			}
		}
	}
	return fixes
}

// runAlignStage is the "align" stage: voiceover.wav + script.json →
// alignment.json (word-level timestamps + section spans), with long silences
// compressed out of the audio and a TTS accuracy (WER) report.
func runAlignStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	voiceover := filepath.Join(l.GeneratedDir(), VoiceoverFileName)
	if _, err := os.Stat(voiceover); err != nil {
		return fmt.Errorf("no %s yet — the audio stage must run first", VoiceoverFileName)
	}
	script, err := loadScript(l)
	if err != nil {
		return err
	}

	var words []AlignedWord
	source := AlignSourceWhisperX
	switch {
	case e.Aligner != nil:
		fmt.Fprintf(e.out(), "  → align     word-aligning with whisperX...\n")
		words, err = e.Aligner.Align(ctx, voiceover)
		if err != nil {
			return err
		}
	case e.Transcriber != nil:
		source = AlignSourceEstimated
		fmt.Fprintf(e.out(), "  → align     whisperX not installed — estimating word timing from %s segments\n", cfg.Pipeline.CaptionsModel)
		fmt.Fprintf(e.out(), "  ⚠ align     scene sync will use estimates; install tools/align (uv sync) for word-level precision\n")
		tr, err := e.Transcriber.Transcribe(ctx, voiceover, cfg.Pipeline.CaptionsModel)
		if err != nil {
			return fmt.Errorf("fallback transcription: %w", err)
		}
		words = wordsFromSegments(tr.Segments)
		if len(words) == 0 {
			return fmt.Errorf("fallback transcription produced no words")
		}
	default:
		return fmt.Errorf(
			"no aligner available — install whisperX (cd tools/align && uv sync) or set %s for the Groq fallback",
			llm.EnvGroqKey,
		)
	}

	// Compress long silences, shifting word times onto the new timeline.
	cuts, adjusted := planGapCompression(words, maxGapMs, keepGapMs)
	if len(cuts) > 0 {
		saved := 0
		for _, c := range cuts {
			saved += c.ToMs - c.FromMs
		}
		fmt.Fprintf(e.out(), "    compressing %d silence gap(s), saving %.1fs\n", len(cuts), float64(saved)/1000)
		if err := compressAudio(ctx, e, voiceover, cuts); err != nil {
			return err
		}
		words = adjusted
	}

	// WER references. The TTS speaks the prepped text (tts_script.json), but
	// whisper inverse-normalizes speech back to written form ("python dot
	// org" comes back as "python.org", spelled-out numbers come back as
	// digits). Judging against a single reference produces false misreads,
	// so each section is scored against BOTH the written narration and the
	// spoken text, and the better match wins: the TTS only misread a word if
	// the transcript matches neither form.
	narrRefs := make([]sectionRef, len(script.Sections))
	for si, sec := range script.Sections {
		narrRefs[si] = sectionRef{ID: sec.ID, Text: sec.Narration}
	}
	refs := narrRefs
	spokenRefs := narrRefs
	haveSpoken := false
	if spoken, err := loadTTSScript(l); err != nil {
		return err
	} else if spoken != nil {
		bySection := map[string]string{}
		for _, s := range spoken.Sections {
			bySection[s.ID] = s.SpokenText()
		}
		spokenRefs = make([]sectionRef, len(narrRefs))
		copy(spokenRefs, narrRefs)
		for si := range spokenRefs {
			if text, ok := bySection[spokenRefs[si].ID]; ok {
				spokenRefs[si].Text = text
			}
		}
		// refs is mutated by the per-section merge below; it must not share
		// a backing array with spokenRefs.
		refs = make([]sectionRef, len(spokenRefs))
		copy(refs, spokenRefs)
		haveSpoken = true
	}

	// Timing spans come from the spoken-text alignment (it tracks the audio
	// most closely); WER and misreads take the per-section best of the two.
	sections, overall, misreads := matchSections(spokenRefs, words)
	if haveSpoken {
		narrSpans, _, narrMisreads := matchSections(narrRefs, words)
		misreadsBySection := map[string][]Misread{}
		for _, m := range misreads {
			misreadsBySection[m.Section] = append(misreadsBySection[m.Section], m)
		}
		narrBySection := map[string][]Misread{}
		for _, m := range narrMisreads {
			narrBySection[m.Section] = append(narrBySection[m.Section], m)
		}
		totalWords, totalErrs := 0, 0.0
		var merged []Misread
		for si := range sections {
			refWords := len(expandWords(spokenRefs[si].Text))
			if narrSpans[si].WER < sections[si].WER {
				sections[si].WER = narrSpans[si].WER
				refs[si] = narrRefs[si]
				refWords = len(expandWords(narrRefs[si].Text))
				merged = append(merged, narrBySection[sections[si].ID]...)
			} else {
				merged = append(merged, misreadsBySection[sections[si].ID]...)
			}
			totalWords += refWords
			totalErrs += sections[si].WER * float64(refWords)
		}
		misreads = merged
		if totalWords > 0 {
			overall = totalErrs / float64(totalWords)
		}
	}
	alignment := Alignment{Source: source, Words: words, Sections: sections}
	alignment.DisplayWords, alignment.DisplaySections = buildDisplayWords(narrRefs, words)
	if err := writeJSON(filepath.Join(l.GeneratedDir(), AlignmentFileName), alignment); err != nil {
		return err
	}

	// TTS accuracy QA report.
	report := ttsAccuracyReport{
		Source:     source,
		OverallWER: overall,
		Threshold:  werFlagThreshold,
		CheckedAt:  time.Now().UTC(),
	}
	flaggedSection := map[string]bool{}
	for si, span := range sections {
		refWords := len(expandWords(refs[si].Text))
		flagged := span.WER > werFlagThreshold
		report.Sections = append(report.Sections, sectionAccuracy{
			ID:      span.ID,
			Words:   refWords,
			Errors:  int(span.WER*float64(refWords) + 0.5),
			WER:     span.WER,
			Flagged: flagged,
		})
		if flagged {
			report.Flagged = append(report.Flagged, span.ID)
			flaggedSection[span.ID] = true
		}
	}
	// Misreads from flagged sections: auto-fix the ones the pronunciation
	// dictionary knows, surface the rest verbatim.
	for _, m := range misreads {
		if flaggedSection[m.Section] {
			report.Misreads = append(report.Misreads, m)
		}
	}
	if len(report.Misreads) > 0 {
		dict := SpeechDict(cfg.Style.Pronunciations)
		existing, err := loadTTSFixes(l)
		if err != nil {
			return err
		}
		spokenBySection := make(map[string]string, len(spokenRefs))
		for _, r := range spokenRefs {
			spokenBySection[r.ID] = r.Text
		}
		fixes := planAutoFixes(report.Misreads, dict, existing, spokenBySection)
		if len(fixes) > 0 {
			report.AutoFixes = fixes
			merged := make(map[string]string, len(existing)+len(fixes))
			for k, v := range existing {
				merged[k] = v
			}
			for k, v := range fixes {
				merged[k] = v
			}
			if err := writeJSON(filepath.Join(l.GeneratedDir(), TTSFixesFileName), merged); err != nil {
				return err
			}
		}
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), ReviewsDirName, TTSAccuracyFileName), report); err != nil {
		return err
	}

	// Pace verification: actual WPM per section vs style.pace_wpm.
	pace := buildPaceReport(refs, sections, cfg.Style.PaceWPM)
	if err := writeJSON(filepath.Join(l.GeneratedDir(), ReviewsDirName, PaceFileName), pace); err != nil {
		return err
	}

	// Auto-pace: when the lesson as a whole runs outside the band, write an
	// absolute speed correction; its appearance re-stales the audio stage.
	var totalWords int
	var totalMs int
	for _, row := range pace.Sections {
		totalWords += row.Words
		totalMs += row.DurationMs
	}
	var lessonWPM float64
	if totalMs > 0 {
		lessonWPM = float64(totalWords) / (float64(totalMs) / 60000)
	}
	oldFix := 1.0
	if prev, err := loadTTSSpeedFix(l); err != nil {
		return err
	} else if prev != nil {
		oldFix = prev.Speed
	}
	if fix := computeSpeedFix(lessonWPM, cfg.Style.PaceWPM, oldFix); fix != nil {
		if err := writeJSON(filepath.Join(l.GeneratedDir(), TTSSpeedFileName), fix); err != nil {
			return err
		}
		fmt.Fprintf(e.out(), "    auto-pace: lesson runs at %.0f wpm vs target %d — wrote %s (speed %.2f); re-run to apply\n",
			lessonWPM, cfg.Style.PaceWPM, TTSSpeedFileName, fix.Speed)
	}

	fmt.Fprintf(e.out(), "    %d words aligned (%s), overall WER %.1f%%\n", len(words), source, overall*100)
	if len(report.Flagged) > 0 {
		fmt.Fprintf(e.out(), "  ⚠ align     WER above %.0f%% in section(s) %s — see %s\n",
			werFlagThreshold*100, strings.Join(report.Flagged, ", "),
			filepath.Join(project.GeneratedDirName, ReviewsDirName, TTSAccuracyFileName))
		for _, m := range report.Misreads {
			heard := m.Heard
			if heard == "" {
				heard = "(dropped)"
			}
			fmt.Fprintf(e.out(), "      %s: wrote %q, heard %q\n", m.Section, m.Ref, heard)
		}
		if len(report.AutoFixes) > 0 {
			fmt.Fprintf(e.out(), "    auto-applied %d pronunciation fix(es) to %s — re-run to synthesize with them\n",
				len(report.AutoFixes), TTSFixesFileName)
		}
	}
	if len(pace.Flagged) > 0 {
		fmt.Fprintf(e.out(), "  ⚠ align     pace outside %d±%.0f%% wpm in section(s) %s — see %s\n",
			pace.TargetWPM, paceTolerance*100, strings.Join(pace.Flagged, ", "),
			filepath.Join(project.GeneratedDirName, ReviewsDirName, PaceFileName))
	}
	return nil
}

// loadTTSScript reads tts_script.json; a missing file returns nil.
func loadTTSScript(l *project.Lesson) (*TTSScript, error) {
	data, err := os.ReadFile(filepath.Join(l.GeneratedDir(), TTSScriptFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", TTSScriptFileName, err)
	}
	var out TTSScript
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parsing %s (delete it and re-run the audio stage): %w", TTSScriptFileName, err)
	}
	return &out, nil
}
