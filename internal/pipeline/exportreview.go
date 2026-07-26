package pipeline

// export-review: one markdown document per lesson bundling everything a
// human subject-matter expert needs — script, diagrams (inline), quiz with
// answers, common mistakes, exercises, and the flags from every automated
// pass — plus instructions for leaving notes in review-notes.yaml.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/enfec/coursesmith/internal/project"
)

// ReviewExportDirName is created under the course directory.
const ReviewExportDirName = "review-export"

// ExportReview writes the per-lesson review documents and returns the
// output directory.
func ExportReview(e *Env, course *project.Course) (string, error) {
	lessons, err := course.Lessons()
	if err != nil {
		return "", err
	}
	outDir := filepath.Join(course.Dir, ReviewExportDirName)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", fmt.Errorf("creating %s: %w", outDir, err)
	}
	for _, l := range lessons {
		doc := lessonReviewDoc(l)
		path := filepath.Join(outDir, l.ID+".md")
		if err := writeFileAtomic(path, []byte(doc)); err != nil {
			return "", err
		}
		fmt.Fprintf(e.out(), "  ✓ %s\n", path)
	}
	return outDir, nil
}

// readJSONFile decodes a generated JSON artifact into out; false when the
// artifact does not exist or does not parse.
func readJSONFile(path string, out any) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return json.Unmarshal(data, out) == nil
}

func lessonReviewDoc(l *project.Lesson) string {
	gen := func(parts ...string) string {
		return filepath.Join(append([]string{l.GeneratedDir()}, parts...)...)
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s — %s\n\n", l.ID, l.FrontMatter.Title)
	b.WriteString("_Generated for human review. Leave feedback in `review-notes.yaml` (template at the bottom); the next pipeline run applies your notes and marks them resolved._\n")

	// Narration script with inline diagrams at their cue points.
	var script Script
	hasScript := readJSONFile(gen(ScriptFileName), &script)
	if !hasScript {
		b.WriteString("\n## Narration script\n\n_Not generated yet — run the pipeline first._\n")
	} else {
		b.WriteString("\n## Narration script\n")
		for _, sec := range script.Sections {
			fmt.Fprintf(&b, "\n### %s (~%ds)\n\n%s\n", sec.ID, sec.DurationEstSec, strings.TrimSpace(sec.Narration))
			for _, cue := range sec.Cues {
				switch cue.Type {
				case CueDiagram:
					svg, err := os.ReadFile(gen(DiagramsDirName, cue.Ref+".svg"))
					if err == nil {
						fmt.Fprintf(&b, "\n**Diagram `%s`** (shown at word %d):\n\n%s\n", cue.Ref, cue.AtWord, strings.TrimSpace(string(svg)))
					} else {
						fmt.Fprintf(&b, "\n**Diagram `%s`** (not rendered yet)\n", cue.Ref)
					}
				case CueDemo:
					fmt.Fprintf(&b, "\n**Demo** (at word %d): %s\n", cue.AtWord, cue.Ref)
				}
			}
		}
	}

	// Quiz with answers.
	var quiz Quiz
	if readJSONFile(gen(QuizFileName), &quiz) {
		fmt.Fprintf(&b, "\n## Quiz — %s\n", quiz.Title)
		for i, q := range quiz.Questions {
			tags := q.Type
			if q.Review {
				tags += ", spaced-repetition review"
			}
			fmt.Fprintf(&b, "\n**%d. [%s]** %s\n\n", i+1, tags, strings.TrimSpace(q.Prompt))
			for j, opt := range q.Options {
				mark := " "
				if j == q.AnswerIndex {
					mark = "x"
				}
				fmt.Fprintf(&b, "- [%s] %s\n", mark, opt)
			}
			fmt.Fprintf(&b, "\n_%s_\n", strings.TrimSpace(q.Explanation))
		}
	}

	// Common mistakes with real tracebacks.
	var mistakes MistakesDoc
	if readJSONFile(gen(MistakesFileName), &mistakes) {
		b.WriteString("\n## Common mistakes (tracebacks are real sandbox output)\n")
		for _, m := range mistakes.Mistakes {
			fmt.Fprintf(&b, "\n### %s\n\n%s\n\n```python\n%s\n```\n\n```\n%s\n```\n\n**Fix:** %s\n\n```python\n%s\n```\n",
				m.Title, m.Explanation, strings.TrimSpace(m.BrokenCode), m.Traceback, m.Fix, strings.TrimSpace(m.FixedCode))
		}
	}

	// Exercises.
	var exercises ExercisesDoc
	if readJSONFile(gen(ExercisesDirName, ExercisesManifestName), &exercises) {
		b.WriteString("\n## Practice exercises (solutions verified by execution)\n")
		for _, ex := range exercises.Exercises {
			fmt.Fprintf(&b, "\n### %s — %s\n\n%s\n\n**Starter:**\n\n```python\n%s\n```\n\n**Hidden tests:**\n\n```python\n%s\n```\n",
				ex.Slug, ex.Title, strings.TrimSpace(ex.Description), strings.TrimSpace(ex.StarterCode), strings.TrimSpace(ex.TestCode))
		}
	}

	// Flags from every automated pass.
	b.WriteString("\n## Automated quality flags\n")
	flags := collectQualityFlags(l)
	if len(flags) == 0 {
		b.WriteString("\n_No review artifacts found — run the pipeline first._\n")
	}
	for _, f := range flags {
		fmt.Fprintf(&b, "\n- %s", f)
	}
	b.WriteString("\n")

	// Notes template.
	fmt.Fprintf(&b, "\n## Leaving feedback\n\nAdd notes to `%s` in the course directory:\n\n```yaml\nlessons:\n  %s:\n    notes:\n      - note: \"Lesson-wide feedback here\"\n    sections:\n", ReviewNotesFileName, l.ID)
	if hasScript && len(script.Sections) > 0 {
		fmt.Fprintf(&b, "      %s:\n        - note: \"Section-specific feedback here\"\n", script.Sections[0].ID)
	} else {
		b.WriteString("      some-section-id:\n        - note: \"Section-specific feedback here\"\n")
	}
	b.WriteString("```\n")
	return b.String()
}

// collectQualityFlags summarizes every persisted QA report for one lesson.
func collectQualityFlags(l *project.Lesson) []string {
	reviews := func(name string) string {
		return filepath.Join(l.GeneratedDir(), ReviewsDirName, name)
	}
	var flags []string

	// Latest multipass review round.
	for round := maxReviewRounds; round >= 1; round-- {
		var rec MultiReviewRecord
		if readJSONFile(reviews(fmt.Sprintf("script-multipass-round-%d.json", round)), &rec) {
			flags = append(flags, fmt.Sprintf(
				"script review (round %d): accuracy %.1f, pedagogy %.1f, tone %.1f → weighted %.1f/%.1f (%s)",
				rec.Round, rec.Accuracy.Score, rec.Pedagogy.Score, rec.Tone.Score, rec.Weighted, rec.Threshold, passFailWord(rec.Passed)))
			for _, v := range rec.Accuracy.Verdicts {
				if v.Verdict != "correct" {
					flags = append(flags, fmt.Sprintf("  %s claim: %q — %s", v.Verdict, v.Claim, v.Citation))
				}
			}
			break
		}
	}

	var wer ttsAccuracyReport
	if readJSONFile(reviews(TTSAccuracyFileName), &wer) {
		flags = append(flags, fmt.Sprintf("TTS accuracy: overall WER %.1f%% (flagged sections: %s)",
			wer.OverallWER*100, orNone(wer.Flagged)))
		for _, m := range wer.Misreads {
			heard := m.Heard
			if heard == "" {
				heard = "(dropped)"
			}
			flags = append(flags, fmt.Sprintf("  misread in %s: wrote %q, heard %q", m.Section, m.Ref, heard))
		}
	}
	var pace paceReport
	if readJSONFile(reviews(PaceFileName), &pace) {
		flags = append(flags, fmt.Sprintf("pace: target %d wpm ±%.0f%% (out-of-band: %s)",
			pace.TargetWPM, pace.Tolerance*100, orNone(pace.Flagged)))
	}
	var loud LoudnessReport
	if readJSONFile(reviews(LoudnessFileName), &loud) {
		flags = append(flags, fmt.Sprintf("loudness: %.1f → %.1f LUFS (target %.0f), true peak %.1f dBTP",
			loud.Before.IntegratedLUFS, loud.After.IntegratedLUFS, loud.TargetLUFS, loud.After.TruePeakDB))
	}
	var distractors distractorReport
	if readJSONFile(reviews(QuizDistractorsFileName), &distractors) && len(distractors.Rounds) > 0 {
		last := distractors.Rounds[len(distractors.Rounds)-1]
		flags = append(flags, fmt.Sprintf("quiz distractors: %s weak after %d round(s)",
			orNone(last.Weak), len(distractors.Rounds)))
	}
	var difficulty difficultyReport
	if readJSONFile(reviews(QuizDifficultyFileName), &difficulty) {
		flags = append(flags, fmt.Sprintf("quiz difficulty (simulated %d students): %s",
			difficulty.Students, orNone(difficulty.Flagged)))
	}
	return flags
}

func orNone(items []string) string {
	if len(items) == 0 {
		return "none"
	}
	return strings.Join(items, "; ")
}
