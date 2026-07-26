package pipeline

// Mistakes stage: auto-generates the lesson's "common mistakes" section —
// the top beginner errors for the topic, each with the ACTUAL error message
// produced by running the broken code in the sandbox. Real tracebacks only:
// broken code that does not fail, or fixed code that does, rejects the
// draft and triggers a repair round.

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// MistakesFileName is the mistakes stage output in the lesson's generated dir.
const MistakesFileName = "mistakes.json"

// mistakesCount is how many common mistakes each lesson documents.
const mistakesCount = 3

// Mistake is one documented beginner error.
type Mistake struct {
	Title       string `json:"title"`
	Explanation string `json:"explanation"` // why beginners make it
	BrokenCode  string `json:"broken_code"`
	// Traceback is the real stderr Python produced in the sandbox.
	Traceback string `json:"traceback"`
	Fix       string `json:"fix"`
	FixedCode string `json:"fixed_code"`
	// ErrorType is the exception class the model claims broken_code raises.
	// It is matched against the real traceback so a snippet cannot fail for
	// an unrelated reason (e.g. a stray NameError) and still be accepted.
	ErrorType string `json:"error_type"`
}

// MistakesDoc is the persisted mistakes.json.
type MistakesDoc struct {
	Mistakes    []Mistake `json:"mistakes"`
	Runner      string    `json:"runner"`
	GeneratedAt time.Time `json:"generated_at"`
}

// mistakesPromptData feeds prompts/mistakes.tmpl.
type mistakesPromptData struct {
	Audience  string
	Title     string
	Outline   string
	Narration string
	Count     int
}

// validateMistakes checks shape and EXECUTES every code sample: broken code
// must fail, fixed code must succeed. Tracebacks are captured in place.
func validateMistakes(ctx context.Context, e *Env, doc *MistakesDoc) error {
	if len(doc.Mistakes) != mistakesCount {
		return fmt.Errorf("got %d mistakes, want exactly %d", len(doc.Mistakes), mistakesCount)
	}
	for i := range doc.Mistakes {
		m := &doc.Mistakes[i]
		if strings.TrimSpace(m.Title) == "" || strings.TrimSpace(m.Explanation) == "" || strings.TrimSpace(m.Fix) == "" {
			return fmt.Errorf("mistake %d is missing title, explanation, or fix", i+1)
		}
		if strings.TrimSpace(m.BrokenCode) == "" || strings.TrimSpace(m.FixedCode) == "" {
			return fmt.Errorf("mistake %d is missing broken_code or fixed_code", i+1)
		}
		broken, err := e.CodeRunner.RunPython(ctx, m.BrokenCode, codeTimeout)
		if err != nil {
			return err
		}
		if broken.Ok() {
			return fmt.Errorf("mistake %d (%q): broken_code runs WITHOUT an error — it must actually fail so the real traceback can be shown", i+1, m.Title)
		}
		m.Traceback = strings.TrimSpace(tailLines(broken.Stderr, 8))
		if m.Traceback == "" {
			return fmt.Errorf("mistake %d (%q): broken_code produced no error output", i+1, m.Title)
		}
		// The snippet must fail for the reason this mistake documents — a
		// stray NameError under a title about `=` vs `==` would show the
		// learner a traceback that teaches the wrong lesson. Only checked
		// when the model declared an error_type, so a missing field never
		// costs a run its single repair round.
		if want := strings.TrimSpace(m.ErrorType); want != "" {
			if got := exceptionType(m.Traceback); got != "" && got != want {
				return fmt.Errorf("mistake %d (%q): broken_code raises %s, but the mistake is documented as %s — broken_code must fail with the error this mistake actually causes", i+1, m.Title, got, want)
			}
		}
		fixed, err := e.CodeRunner.RunPython(ctx, m.FixedCode, codeTimeout)
		if err != nil {
			return err
		}
		if !fixed.Ok() {
			return fmt.Errorf("mistake %d (%q): fixed_code itself fails: %s", i+1, m.Title, tailLines(fixed.Stderr, 3))
		}
	}
	return nil
}

// exceptionType pulls the exception class out of a Python traceback — the
// last non-empty line is "SomeError: detail" for both ordinary tracebacks and
// SyntaxError's caret form. Dotted classes (json.decoder.JSONDecodeError)
// reduce to their final segment so they compare against a bare class name.
// Returns "" when the tail does not look like an exception line.
func exceptionType(traceback string) string {
	lines := strings.Split(traceback, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		head, _, _ := strings.Cut(line, ":")
		head = strings.TrimSpace(head)
		if idx := strings.LastIndex(head, "."); idx >= 0 {
			head = head[idx+1:]
		}
		// Exception classes are single identifiers ending in Error/Exception
		// (or a builtin like StopIteration); anything with spaces is prose.
		if head == "" || strings.ContainsAny(head, " \t") {
			return ""
		}
		return head
	}
	return ""
}

// runMistakesStage generates mistakes.json with sandbox-verified tracebacks.
func runMistakesStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	if e.CodeRunner == nil {
		return fmt.Errorf("no code runner available — the mistakes stage needs real tracebacks; install docker and run `%s`", sandboxBuildHelp)
	}
	script, err := loadScript(l)
	if err != nil {
		return err
	}
	narrations := make([]string, 0, len(script.Sections))
	for _, sec := range script.Sections {
		narrations = append(narrations, sec.Narration)
	}

	data := mistakesPromptData{
		Audience:  cfg.Style.Audience,
		Title:     l.FrontMatter.Title,
		Outline:   l.Body,
		Narration: strings.Join(narrations, "\n\n"),
		Count:     mistakesCount,
	}
	system, user, err := e.renderPrompt(mistakesTemplateName, data)
	if err != nil {
		return err
	}

	fmt.Fprintf(e.out(), "  → mistakes  drafting %d common mistakes (%s), verifying tracebacks via %s...\n",
		mistakesCount, cfg.Pipeline.LLMContent, e.CodeRunner.Name())
	var doc MistakesDoc
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.4, 4096, &doc,
		func() error { return validateMistakes(ctx, e, &doc) })
	if err != nil {
		return fmt.Errorf("generating mistakes: %w", err)
	}
	doc.Runner = e.CodeRunner.Name()
	doc.GeneratedAt = time.Now().UTC()

	if err := writeJSON(filepath.Join(l.GeneratedDir(), MistakesFileName), &doc); err != nil {
		return err
	}
	fmt.Fprintf(e.out(), "    %d mistake(s) with real tracebacks written to %s\n", len(doc.Mistakes), MistakesFileName)
	return nil
}
