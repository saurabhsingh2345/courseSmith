package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// CaptionsFileName is the captions stage output in the lesson's generated dir.
const CaptionsFileName = "captions.vtt"

// Transcriber turns an audio file into a timed transcription. Satisfied by
// *llm.GroqTranscriber; an interface so tests can substitute a fake. Used by
// the align stage as the fallback when whisperX is unavailable.
type Transcriber interface {
	Transcribe(ctx context.Context, audioPath, model string) (*llm.Transcription, error)
}

// Caption cue grouping bounds.
const (
	cueMaxWords = 8
	cueMaxChars = 42
	cueBreakGap = 800 // ms of silence that forces a new cue
)

// loadAlignment reads the lesson's generated alignment.json.
func loadAlignment(l *project.Lesson) (*Alignment, error) {
	path := filepath.Join(l.GeneratedDir(), AlignmentFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no %s yet — the align stage must run first", AlignmentFileName)
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var a Alignment
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("parsing %s (delete it and re-run the align stage): %w", path, err)
	}
	if len(a.Words) == 0 {
		return nil, fmt.Errorf("%s has no words — re-run the align stage", AlignmentFileName)
	}
	return &a, nil
}

// runCaptionsStage builds captions.vtt from the word-level alignment.
func runCaptionsStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	alignment, err := loadAlignment(l)
	if err != nil {
		return err
	}
	fmt.Fprintf(e.out(), "  → captions  building WebVTT from %d aligned words...\n", len(alignment.CaptionWords()))

	vtt, cueCount := vttFromWords(alignment.CaptionWords())
	path := filepath.Join(l.GeneratedDir(), CaptionsFileName)
	if err := writeFileAtomic(path, []byte(vtt)); err != nil {
		return err
	}
	fmt.Fprintf(e.out(), "    %s written (%d cues)\n", CaptionsFileName, cueCount)

	// Keyword emphasis (optional enhancement): mark the words the burned-in
	// captions hold in the accent colour. Degrades to plain captions when no
	// LLM is configured or the pass fails.
	if e.Router != nil {
		emphasis, err := e.generateCaptionEmphasis(ctx, cfg, alignment)
		if err != nil {
			fmt.Fprintf(e.out(), "    warning: caption emphasis skipped: %v\n", err)
			return nil
		}
		if err := writeJSON(filepath.Join(l.GeneratedDir(), CaptionEmphasisFileName), emphasis); err != nil {
			return err
		}
		fmt.Fprintf(e.out(), "    %s written (%d keywords)\n", CaptionEmphasisFileName, len(emphasis.Indices))
	}
	return nil
}

// vttFromWords groups aligned words into readable caption cues: a new cue
// starts at sentence ends, long silences, or when the line would run past
// the word/char limits.
func vttFromWords(words []AlignedWord) (string, int) {
	var b strings.Builder
	b.WriteString("WEBVTT\n\n")

	var cue []AlignedWord
	cueChars := 0
	count := 0
	flush := func() {
		if len(cue) == 0 {
			return
		}
		count++
		texts := make([]string, len(cue))
		for i, w := range cue {
			texts[i] = w.Word
		}
		fmt.Fprintf(&b, "%d\n%s --> %s\n%s\n\n",
			count,
			vttTimestamp(float64(cue[0].StartMs)/1000),
			vttTimestamp(float64(cue[len(cue)-1].EndMs)/1000),
			strings.Join(texts, " "),
		)
		cue = cue[:0]
		cueChars = 0
	}

	for _, w := range words {
		if strings.TrimSpace(w.Word) == "" {
			continue
		}
		if len(cue) > 0 {
			prev := cue[len(cue)-1]
			gap := w.StartMs - prev.EndMs
			endOfSentence := strings.ContainsAny(suffixPunct(prev.Word), ".!?")
			if endOfSentence || gap > cueBreakGap ||
				len(cue) >= cueMaxWords || cueChars+1+len(w.Word) > cueMaxChars {
				flush()
			}
		}
		cue = append(cue, w)
		if cueChars == 0 {
			cueChars = len(w.Word)
		} else {
			cueChars += 1 + len(w.Word)
		}
	}
	flush()
	return b.String(), count
}

// suffixPunct returns the trailing punctuation of a word ("" if none).
func suffixPunct(w string) string {
	trimmed := strings.TrimRight(w, ".,!?;:…")
	return w[len(trimmed):]
}

// vttTimestamp formats seconds as a WebVTT timestamp (HH:MM:SS.mmm).
func vttTimestamp(seconds float64) string {
	if seconds < 0 {
		seconds = 0
	}
	ms := int64(seconds*1000 + 0.5)
	h := ms / 3_600_000
	m := ms % 3_600_000 / 60_000
	s := ms % 60_000 / 1000
	ms %= 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}
