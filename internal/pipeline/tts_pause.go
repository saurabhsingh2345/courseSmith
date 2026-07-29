package pipeline

// Sentence pauses: giving the read room to breathe without touching the read.
//
// Kokoro synthesizes a paragraph as one continuous take, which is why the
// voice keeps its intonation across a full stop — ask it for one sentence at a
// time and every sentence restarts the prosody, which sounds clipped. So the
// pauses are not synthesized in. They are cut into the finished audio
// afterwards, at the word boundaries the aligner has already located.
//
// That ordering is what keeps the video aligned. This is the mirror of
// planGapCompression: that one removes regions and shifts every later
// timestamp back, this one adds regions and shifts every later timestamp
// forward. Both compute the shift arithmetically rather than re-transcribing,
// so the alignment stays exact to the millisecond and every downstream
// consumer — captions, chapters, the scene graph's cue timings — follows for
// free.
//
// The pause is a FLOOR, not an addition. A sentence end that already breathes
// for longer than the floor is left exactly as the narrator read it; only the
// ones that run on get widened. That is also why this needs no special case
// for paragraph and section joins, whose baked-in 350ms and 700ms already
// clear the floor.

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// silenceInsert is a widening of the natural gap that follows one word.
// AtMs is a position on the pre-insertion timeline.
type silenceInsert struct {
	AtMs  int `json:"atMs"`
	AddMs int `json:"addMs"`
}

// endsSentence reports whether a word token closes a sentence.
//
// Trailing quotes and brackets are stripped first so that a quoted sentence
// ("...like this.") still registers. A single trailing "." is deliberately
// NOT treated as an abbreviation problem: the narration has already been
// through PrepForSpeech, which expands the abbreviations that would otherwise
// produce a false boundary.
func endsSentence(word string) bool {
	trimmed := strings.TrimRight(word, `"'”’)]}`)
	if trimmed == "" {
		return false
	}
	switch trimmed[len(trimmed)-1] {
	case '.', '!', '?':
		return true
	}
	return false
}

// planSentencePauses finds sentence ends whose following gap is shorter than
// floorMs and plans the silence that brings each up to the floor.
//
// It returns the insertions (on the original timeline, in order) and the word
// list with timestamps moved onto the widened timeline. The final word is
// never a boundary — a pause after the last word is just a longer video.
//
// capMs bounds a single insertion so this cannot plan a gap the align stage's
// own long-gap compression would have cut back out on the next run.
func planSentencePauses(words []AlignedWord, floorMs, capMs int) ([]silenceInsert, []AlignedWord) {
	if floorMs <= 0 || len(words) < 2 {
		return nil, words
	}
	var inserts []silenceInsert
	shifted := make([]AlignedWord, len(words))
	added := 0
	for i, w := range words {
		shifted[i] = AlignedWord{Word: w.Word, StartMs: w.StartMs + added, EndMs: w.EndMs + added}
		if i == len(words)-1 || !endsSentence(w.Word) {
			continue
		}
		gap := max(words[i+1].StartMs-w.EndMs, 0)
		need := floorMs - gap
		if need <= 0 {
			continue // already breathes for long enough
		}
		if capMs > 0 && gap+need > capMs {
			need = capMs - gap
			if need <= 0 {
				continue
			}
		}
		// Split the existing gap: the quietest point is its middle, so a
		// join there cannot clip the tail of the word or the attack of the
		// next one.
		inserts = append(inserts, silenceInsert{AtMs: w.EndMs + gap/2, AddMs: need})
		added += need
	}
	return inserts, shifted
}

// shiftForInserts maps a timestamp from the pre-insertion timeline onto the
// widened one. inserts must be ordered by AtMs.
func shiftForInserts(ms int, inserts []silenceInsert) int {
	added := 0
	for _, in := range inserts {
		if in.AtMs > ms {
			break
		}
		added += in.AddMs
	}
	return ms + added
}

// shiftSpans moves section spans onto the widened timeline. Word indices are
// untouched: inserting silence adds no words.
func shiftSpans(spans []SectionSpan, inserts []silenceInsert) []SectionSpan {
	out := make([]SectionSpan, len(spans))
	for i, s := range spans {
		out[i] = s
		out[i].StartMs = shiftForInserts(s.StartMs, inserts)
		out[i].EndMs = shiftForInserts(s.EndMs, inserts)
	}
	return out
}

// applySentencePauses widens the alignment's sentence gaps in place and cuts
// the matching silence into the voiceover.
//
// Sentence ends are read off the caption track — the written narration with
// the author's own punctuation — rather than the raw transcript, because ASR
// punctuation is a guess and the author's is not. Both tracks share one
// timeline, so a single set of insertions shifts all four collections.
//
// Returns the insertions so the caller can keep reporting pace on speaking
// time rather than on time that is now silence.
func applySentencePauses(ctx context.Context, e *Env, a *Alignment, voiceoverPath string, floorMs int) ([]silenceInsert, error) {
	if floorMs <= 0 {
		return nil, nil
	}
	// keepGapMs is what long-gap compression leaves behind, so planning a
	// wider gap than that would be undone on the next align run.
	capMs := max(floorMs, keepGapMs)

	inserts, _ := planSentencePauses(a.CaptionWords(), floorMs, capMs)
	if len(inserts) == 0 {
		return nil, nil
	}
	if err := insertSilence(ctx, e, voiceoverPath, inserts); err != nil {
		return nil, err
	}

	for i, w := range a.Words {
		a.Words[i] = AlignedWord{
			Word:    w.Word,
			StartMs: shiftForInserts(w.StartMs, inserts),
			EndMs:   shiftForInserts(w.EndMs, inserts),
		}
	}
	for i, w := range a.DisplayWords {
		a.DisplayWords[i] = AlignedWord{
			Word:    w.Word,
			StartMs: shiftForInserts(w.StartMs, inserts),
			EndMs:   shiftForInserts(w.EndMs, inserts),
		}
	}
	a.Sections = shiftSpans(a.Sections, inserts)
	a.DisplaySections = shiftSpans(a.DisplaySections, inserts)
	return inserts, nil
}

// insertSilence cuts silence into a WAV at the given points, in place.
//
// The silence is generated inside the filter graph rather than as extra
// inputs, so a lesson with two hundred sentences still runs as one ffmpeg
// invocation with one input.
func insertSilence(ctx context.Context, e *Env, path string, inserts []silenceInsert) error {
	if len(inserts) == 0 {
		return nil
	}
	const fmtChain = "aformat=sample_fmts=s16:channel_layouts=mono"

	var filter strings.Builder
	var labels []string
	cursor := 0
	for i, in := range inserts {
		speech := fmt.Sprintf("[s%d]", i)
		fmt.Fprintf(&filter, "[0:a]atrim=start=%.3f:end=%.3f,asetpts=PTS-STARTPTS,%s%s;",
			float64(cursor)/1000, float64(in.AtMs)/1000, fmtChain, speech)
		labels = append(labels, speech)

		pause := fmt.Sprintf("[p%d]", i)
		fmt.Fprintf(&filter, "anullsrc=r=%d:cl=mono,atrim=end=%.3f,asetpts=PTS-STARTPTS,%s%s;",
			ttsSampleRate, float64(in.AddMs)/1000, fmtChain, pause)
		labels = append(labels, pause)

		cursor = in.AtMs
	}
	tail := fmt.Sprintf("[s%d]", len(inserts))
	fmt.Fprintf(&filter, "[0:a]atrim=start=%.3f,asetpts=PTS-STARTPTS,%s%s;", float64(cursor)/1000, fmtChain, tail)
	labels = append(labels, tail)

	fmt.Fprintf(&filter, "%sconcat=n=%d:v=0:a=1[out]", strings.Join(labels, ""), len(labels))

	tmp := path + ".paused.wav"
	if err := e.runFFmpeg(ctx, "-y", "-i", path, "-filter_complex", filter.String(), "-map", "[out]",
		"-ar", fmt.Sprint(ttsSampleRate), "-ac", "1", "-c:a", "pcm_s16le", tmp); err != nil {
		return fmt.Errorf("inserting %d sentence pause(s): %w", len(inserts), err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}
