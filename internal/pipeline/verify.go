package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// VerificationFileName is the verify stage output in the lesson's generated
// dir: block hash → real execution results.
const VerificationFileName = "verification.json"

// codeTimeout bounds each code block's execution.
const codeTimeout = 5 * time.Second

// CodeBlock is one fenced Python block found in lesson content.
type CodeBlock struct {
	// Source locates the block: "lesson.md" or "script:<section-id>".
	Source string `json:"source"`
	// Index is the block's ordinal within its source (0-based).
	Index int    `json:"index"`
	Code  string `json:"code"`
	// ClaimedOutput is the content of an ```output fence immediately
	// following the code block, if any.
	ClaimedOutput string `json:"claimed_output,omitempty"`
}

// VerifiedBlock is one executed block with its real results.
type VerifiedBlock struct {
	CodeBlock
	Hash          string    `json:"hash"`
	Stdout        string    `json:"stdout"`
	Stderr        string    `json:"stderr,omitempty"`
	ExitCode      int       `json:"exit_code"`
	TimedOut      bool      `json:"timed_out,omitempty"`
	OutputMatches *bool     `json:"output_matches,omitempty"` // nil when nothing was claimed
	VerifiedAt    time.Time `json:"verified_at"`
}

// VerificationReport is the persisted verification.json.
type VerificationReport struct {
	Runner string          `json:"runner"`
	Blocks []VerifiedBlock `json:"blocks"`
}

// extractCodeBlocks finds ```python fenced blocks in markdown, pairing each
// with an immediately following ```output fence as the claimed output.
func extractCodeBlocks(markdown, source string) []CodeBlock {
	lines := strings.Split(strings.ReplaceAll(markdown, "\r\n", "\n"), "\n")
	var blocks []CodeBlock

	// readFence returns the body of the fence starting at line i (which must
	// be an opening fence) and the index of the line after the closing fence.
	readFence := func(i int) (body string, next int) {
		var collected []string
		for j := i + 1; j < len(lines); j++ {
			if strings.HasPrefix(strings.TrimSpace(lines[j]), "```") {
				return strings.Join(collected, "\n"), j + 1
			}
			collected = append(collected, lines[j])
		}
		return strings.Join(collected, "\n"), len(lines)
	}

	i := 0
	for i < len(lines) {
		info := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[i]), "```")))
		if !strings.HasPrefix(strings.TrimSpace(lines[i]), "```") || (info != "python" && info != "py") {
			i++
			continue
		}
		code, next := readFence(i)
		block := CodeBlock{Source: source, Index: len(blocks), Code: code}

		// An ```output fence directly after (blank lines allowed) is a claim.
		j := next
		for j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			j++
		}
		if j < len(lines) {
			jinfo := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[j]), "```")))
			if strings.HasPrefix(strings.TrimSpace(lines[j]), "```") && jinfo == "output" {
				block.ClaimedOutput, next = readFence(j)
			}
		}
		blocks = append(blocks, block)
		i = next
	}
	return blocks
}

// normalizeOutput makes claimed and actual output comparable: CRLF→LF,
// trailing whitespace per line stripped, trailing newlines trimmed.
func normalizeOutput(s string) string {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// loadVerification reads verification.json; a missing file returns nil.
func loadVerification(l *project.Lesson) (*VerificationReport, error) {
	data, err := os.ReadFile(filepath.Join(l.GeneratedDir(), VerificationFileName))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %s: %w", VerificationFileName, err)
	}
	var report VerificationReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parsing %s (delete it to re-verify): %w", VerificationFileName, err)
	}
	return &report, nil
}

// runVerifyStage executes every Python code block from lesson.md and the
// generated script in the sandbox. Real stdout/stderr is recorded in
// verification.json (the ground truth downstream stages render); any block
// that errors fails the pipeline — a lesson with broken code cannot be
// published.
func runVerifyStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, _ config.Config) error {
	if e.CodeRunner == nil {
		return fmt.Errorf("no code runner available — install docker and run `%s`", sandboxBuildHelp)
	}
	script, err := loadScript(l)
	if err != nil {
		return err
	}

	blocks := extractCodeBlocks(l.Body, "lesson.md")
	for _, sec := range script.Sections {
		blocks = append(blocks, extractCodeBlocks(sec.Narration, "script:"+sec.ID)...)
	}
	if len(blocks) == 0 {
		fmt.Fprintf(e.out(), "  → verify    no Python code blocks found — nothing to execute\n")
		return writeJSON(filepath.Join(l.GeneratedDir(), VerificationFileName),
			VerificationReport{Runner: e.CodeRunner.Name()})
	}

	fmt.Fprintf(e.out(), "  → verify    executing %d code block(s) via %s...\n", len(blocks), e.CodeRunner.Name())
	if strings.Contains(e.CodeRunner.Name(), "UNSANDBOXED") {
		fmt.Fprintf(e.out(), "  ⚠ verify    %s\n", "running on the host without isolation — build the sandbox: "+sandboxBuildHelp)
	}

	// Unchanged blocks that already verified clean skip re-execution.
	cache := map[string]VerifiedBlock{}
	if prev, err := loadVerification(l); err == nil && prev != nil {
		for _, b := range prev.Blocks {
			if b.ExitCode == 0 && !b.TimedOut {
				cache[b.Hash] = b
			}
		}
	}

	report := VerificationReport{Runner: e.CodeRunner.Name()}
	var failures []string
	mismatches := 0
	for _, block := range blocks {
		if err := ctx.Err(); err != nil {
			return err
		}
		hash := project.HashBytes([]byte(block.Code))
		verified, ok := cache[hash]
		if ok {
			verified.CodeBlock = block // source/claims may move between runs
		} else {
			res, err := e.CodeRunner.RunPython(ctx, block.Code, codeTimeout)
			if err != nil {
				return fmt.Errorf("verify %s block %d: %w", block.Source, block.Index+1, err)
			}
			verified = VerifiedBlock{
				CodeBlock:  block,
				Hash:       hash,
				Stdout:     res.Stdout,
				Stderr:     res.Stderr,
				ExitCode:   res.ExitCode,
				TimedOut:   res.TimedOut,
				VerifiedAt: time.Now().UTC(),
			}
		}
		if verified.ExitCode != 0 || verified.TimedOut {
			reason := tailLines(verified.Stderr, 4)
			if verified.TimedOut {
				reason = fmt.Sprintf("timed out after %s", codeTimeout)
			}
			failures = append(failures, fmt.Sprintf(
				"%s block %d (%q...): %s",
				block.Source, block.Index+1, firstLine(block.Code), reason,
			))
		}
		if block.ClaimedOutput != "" {
			match := normalizeOutput(block.ClaimedOutput) == normalizeOutput(verified.Stdout)
			verified.OutputMatches = &match
			if !match {
				mismatches++
				fmt.Fprintf(e.out(), "  ⚠ verify    %s block %d: claimed output differs from what Python actually prints — the real output will be used\n",
					block.Source, block.Index+1)
			}
		}
		report.Blocks = append(report.Blocks, verified)
	}

	if err := writeJSON(filepath.Join(l.GeneratedDir(), VerificationFileName), report); err != nil {
		return err
	}
	if len(failures) > 0 {
		return fmt.Errorf(
			"%d code block(s) failed to execute — fix them before this lesson can be published:\n  %s",
			len(failures), strings.Join(failures, "\n  "),
		)
	}
	fmt.Fprintf(e.out(), "    %d block(s) verified", len(report.Blocks))
	if mismatches > 0 {
		fmt.Fprintf(e.out(), ", %d claimed output(s) corrected", mismatches)
	}
	fmt.Fprintln(e.out())
	return nil
}

// verifiedOutputsSummary renders verification results for injection into the
// review prompt, so the critic can catch narration that contradicts reality.
func verifiedOutputsSummary(l *project.Lesson) string {
	report, err := loadVerification(l)
	if err != nil || report == nil || len(report.Blocks) == 0 {
		return ""
	}
	var b strings.Builder
	for _, block := range report.Blocks {
		out := normalizeOutput(block.Stdout)
		if out == "" {
			continue
		}
		fmt.Fprintf(&b, "Code (%s):\n%s\nActual output:\n%s\n\n", block.Source, block.Code, truncate(out, 500))
	}
	return strings.TrimSpace(b.String())
}

// verifyQuizCode executes Python blocks inside quiz question prompts.
// Broken code fails validation (triggering regeneration) — EXCEPT in
// debugging-type questions, whose whole point is showing code that fails.
// When code runs cleanly and prints exactly one of the options,
// answer_index must point at it.
func verifyQuizCode(ctx context.Context, e *Env, q *Quiz) error {
	if e.CodeRunner == nil {
		return nil
	}
	for _, question := range q.Questions {
		for _, block := range extractCodeBlocks(question.Prompt, "quiz:"+question.ID) {
			res, err := e.CodeRunner.RunPython(ctx, block.Code, codeTimeout)
			if err != nil {
				return err
			}
			if !res.Ok() {
				if question.Type == QDebugging {
					continue // failing code is the exercise itself
				}
				return fmt.Errorf("question %q (type %s) contains code that fails when executed: %s",
					question.ID, question.Type, tailLines(res.Stderr, 3))
			}
			out := normalizeOutput(res.Stdout)
			if out == "" {
				continue
			}
			matching := -1
			multiple := false
			for i, opt := range question.Options {
				if normalizeOutput(opt) == out {
					if matching >= 0 {
						multiple = true
					}
					matching = i
				}
			}
			if matching >= 0 && !multiple && question.AnswerIndex != matching {
				return fmt.Errorf(
					"question %q: the code actually prints %q (option %d) but answer_index is %d",
					question.ID, out, matching, question.AnswerIndex,
				)
			}
		}
	}
	return nil
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	return truncate(line, 60)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
