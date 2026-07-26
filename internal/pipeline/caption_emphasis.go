package pipeline

// Caption keyword emphasis: an LLM pass over the aligned transcript marks
// the handful of words that carry each phrase (max ~1 per caption page) so
// the burned-in captions can hold them in the accent colour. No open-source
// tool does this (2026 survey) — it's one cached prompt per section.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// CaptionEmphasisFileName is the emphasis artifact in the generated dir.
const CaptionEmphasisFileName = "captions_emphasis.json"

// maxEmphasisPerWords caps density: at most one keyword per this many words
// (the renderer shows ~5 words per caption page; 8 keeps pages mostly calm).
const maxEmphasisPerWords = 8

// CaptionEmphasis is the persisted emphasis artifact: global indices into
// the alignment's word list.
type CaptionEmphasis struct {
	Indices []int `json:"indices"`
}

type emphasisReply struct {
	Indices []int `json:"indices"`
}

// generateCaptionEmphasis asks the LLM, section by section, which words to
// keep highlighted. Failures degrade to no emphasis (captions still render).
func (e *Env) generateCaptionEmphasis(ctx context.Context, cfg config.Config, a *Alignment) (*CaptionEmphasis, error) {
	out := &CaptionEmphasis{Indices: []int{}}
	capWords := a.CaptionWords()
	for si, span := range a.CaptionSections() {
		if span.WordEnd <= span.WordStart || span.WordEnd > len(capWords) {
			continue
		}
		words := capWords[span.WordStart:span.WordEnd]
		var numbered strings.Builder
		for i, w := range words {
			fmt.Fprintf(&numbered, "%d:%s ", span.WordStart+i, w.Word)
		}
		budget := max(1, len(words)/maxEmphasisPerWords)

		system, user, err := e.renderPrompt(captionEmphasisTemplateName, map[string]any{
			"Words":  strings.TrimSpace(numbered.String()),
			"Budget": budget,
		})
		if err != nil {
			return nil, err
		}
		var reply emphasisReply
		if err := e.completeJSON(ctx, cfg.Pipeline, llm.TaskReview, system, user, 0, 1024, &reply, nil); err != nil {
			return nil, fmt.Errorf("caption emphasis for section %d: %w", si, err)
		}
		// Validate: in-range, deduped, capped to the budget.
		seen := map[int]bool{}
		kept := 0
		for _, idx := range reply.Indices {
			if idx < span.WordStart || idx >= span.WordEnd || seen[idx] || kept >= budget {
				continue
			}
			seen[idx] = true
			kept++
			out.Indices = append(out.Indices, idx)
		}
	}
	sort.Ints(out.Indices)
	return out, nil
}

// loadCaptionEmphasis reads captions_emphasis.json; a missing file is not an
// error (emphasis is an optional enhancement).
func loadCaptionEmphasis(l *project.Lesson) (*CaptionEmphasis, error) {
	path := filepath.Join(l.GeneratedDir(), CaptionEmphasisFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	var ce CaptionEmphasis
	if err := json.Unmarshal(data, &ce); err != nil {
		return nil, fmt.Errorf("parsing %s (delete it and re-run the captions stage): %w", path, err)
	}
	return &ce, nil
}
