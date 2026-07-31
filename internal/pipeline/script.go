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

// ScriptFileName is the Stage 1 output in the lesson's generated dir.
const ScriptFileName = "script.json"

// Script is the narration script for one lesson.
type Script struct {
	Title    string    `json:"title"`
	Sections []Section `json:"sections"`
}

// Section is one narrated segment, mapped from a top-level outline heading.
type Section struct {
	ID string `json:"id"`
	// Title is the section's human heading, when whatever built the script knew
	// it. Empty for a hand-written lesson, where the heading lives in lesson.md
	// and is recovered by matching slugs (see sectionTitles).
	//
	// It exists for the assembled paths. A reel's section id is
	// "<segment>--<beat>", which matches no heading in the generated lesson.md,
	// so title recovery fell through to humanizing the id — and every chapter
	// read "Myth 1  Everyone Says": the template name, the cast ordinal, and a
	// double space where the "--" was. The heading was known at the moment the
	// section was built and thrown away one field short of here.
	Title          string `json:"title,omitempty"`
	Narration      string `json:"narration"`
	DurationEstSec int    `json:"duration_est_sec"`
	Cues           []Cue  `json:"cues"`
}

// Cue anchors a visual event to a word position in the narration.
type Cue struct {
	// Type is "diagram", "demo", or "pause".
	Type string `json:"type"`
	// Ref is the diagram id, the demo description, or "" for a pause.
	Ref string `json:"ref"`
	// AtWord is the 0-based word index in the section narration.
	AtWord int `json:"at_word"`
}

// Cue types.
const (
	CueDiagram = "diagram"
	CueDemo    = "demo"
	CuePause   = "pause"
)

// Validate checks structural soundness and that diagram cues only reference
// diagrams declared in the lesson front-matter.
func (s *Script) Validate(declaredDiagrams map[string]bool) error {
	if strings.TrimSpace(s.Title) == "" {
		return fmt.Errorf("script title is empty")
	}
	if len(s.Sections) == 0 {
		return fmt.Errorf("script has no sections")
	}
	seen := map[string]bool{}
	for i, sec := range s.Sections {
		if strings.TrimSpace(sec.ID) == "" {
			return fmt.Errorf("section %d has an empty id", i)
		}
		if seen[sec.ID] {
			return fmt.Errorf("duplicate section id %q", sec.ID)
		}
		seen[sec.ID] = true
		if strings.TrimSpace(sec.Narration) == "" {
			return fmt.Errorf("section %q has empty narration", sec.ID)
		}
		if sec.DurationEstSec <= 0 {
			return fmt.Errorf("section %q has non-positive duration_est_sec %d", sec.ID, sec.DurationEstSec)
		}
		for j, cue := range sec.Cues {
			switch cue.Type {
			case CueDiagram:
				if !declaredDiagrams[cue.Ref] {
					return fmt.Errorf("section %q cue %d references undeclared diagram %q", sec.ID, j, cue.Ref)
				}
			case CueDemo:
				if strings.TrimSpace(cue.Ref) == "" {
					return fmt.Errorf("section %q cue %d is a demo with an empty description", sec.ID, j)
				}
			case CuePause:
				// no ref required
			default:
				return fmt.Errorf("section %q cue %d has unknown type %q (want diagram, demo, or pause)", sec.ID, j, cue.Type)
			}
			// Models miscount words, so out-of-range indices are clamped by
			// consumers (the video stage) rather than rejected here.
			if cue.AtWord < 0 {
				return fmt.Errorf("section %q cue %d has negative at_word %d", sec.ID, j, cue.AtWord)
			}
		}
	}
	return nil
}

// TotalDurationSec sums the section duration estimates.
func (s *Script) TotalDurationSec() int {
	total := 0
	for _, sec := range s.Sections {
		total += sec.DurationEstSec
	}
	return total
}

// scriptPromptData feeds prompts/script.tmpl.
type scriptPromptData struct {
	Tone     string
	Audience string
	Language string
	PaceWPM  int
	Title    string
	Diagrams []project.DiagramSpec
	Outline  string
	// Critique is reviewer feedback injected on regeneration rounds ("" on
	// the first draft).
	Critique string
	// Notes are unresolved human reviewer notes from review-notes.yaml.
	Notes string
	// Captures describe what this lesson's real recordings actually caught.
	// Empty before anything has been recorded, which is the ordinary state of
	// a first run and the reason the prompt treats them as optional.
	Captures []string
}

// generateScript asks the content model for a narration script; the review
// stage reuses it with a critique to regenerate rejected drafts. Unresolved
// human review notes are always injected.
func generateScript(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, critique string) (*Script, error) {
	notes, err := LoadReviewNotes(courseDirOf(l.Dir))
	if err != nil {
		return nil, err
	}
	data := scriptPromptData{
		Tone:     cfg.Style.Tone,
		Audience: cfg.Style.Audience,
		Language: cfg.Style.Language,
		PaceWPM:  cfg.Style.PaceWPM,
		Title:    l.FrontMatter.Title,
		Diagrams: l.FrontMatter.Diagrams,
		Outline:  l.Body,
		Critique: critique,
		Captures: CaptureBriefing(l),
		Notes:    notes.UnresolvedText(l.ID),
	}
	system, user, err := e.renderPrompt(scriptTemplateName, data)
	if err != nil {
		return nil, err
	}
	declared := make(map[string]bool, len(l.FrontMatter.Diagrams))
	for _, d := range l.FrontMatter.Diagrams {
		declared[d.ID] = true
	}
	var script Script
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.4, 8192, &script,
		func() error { return script.Validate(declared) })
	if err != nil {
		return nil, fmt.Errorf("generating script: %w", err)
	}
	return &script, nil
}

// runScriptStage is Stage 1: lesson.md → generated/script.json. Unresolved
// human review notes are injected into the draft and then marked resolved,
// so the runner's post-run input hashes stabilize immediately.
func runScriptStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	notes, err := LoadReviewNotes(courseDirOf(l.Dir))
	if err != nil {
		return err
	}
	unresolved := notes.UnresolvedText(l.ID)
	if unresolved != "" {
		fmt.Fprintf(e.out(), "  → script    injecting human review notes from %s:\n", ReviewNotesFileName)
		for _, line := range strings.Split(unresolved, "\n") {
			fmt.Fprintf(e.out(), "      %s\n", line)
		}
	}

	fmt.Fprintf(e.out(), "  → script    writing narration draft (%s)...\n", cfg.Pipeline.LLMContent)
	script, err := generateScript(ctx, e, l, cfg, "")
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(l.GeneratedDir(), ScriptFileName), script); err != nil {
		return err
	}

	if unresolved != "" {
		if err := writeJSON(filepath.Join(l.GeneratedDir(), ReviewsDirName, NotesAppliedFileName),
			map[string]string{"notes": unresolved}); err != nil {
			return err
		}
		if n := notes.MarkResolved(l.ID); n > 0 {
			if err := notes.Save(); err != nil {
				return err
			}
			fmt.Fprintf(e.out(), "    %d review note(s) applied and marked resolved\n", n)
		}
	}
	fmt.Fprintf(e.out(), "    %d sections, ~%ds narration\n", len(script.Sections), script.TotalDurationSec())
	return nil
}

// writeJSON writes v as indented JSON, atomically (temp file + rename).
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding %s: %w", filepath.Base(path), err)
	}
	return writeFileAtomic(path, append(data, '\n'))
}

func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file for %s: %w", path, err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file for %s: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("replacing %s: %w", path, err)
	}
	return nil
}
