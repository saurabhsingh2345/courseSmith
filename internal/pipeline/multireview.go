package pipeline

// Multi-dimensional script review: three sequential passes — accuracy
// (claims extracted and, where possible, executed in the sandbox),
// pedagogy, and tone — combined into a weighted gate. Below-threshold
// scripts are regenerated with every critique injected, at most twice.
// Every pass of every round is persisted to generated/reviews/.

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// Review weights per the pedagogy-brain spec: accuracy dominates.
const (
	weightAccuracy = 0.50
	weightPedagogy = 0.35
	weightTone     = 0.15
)

// Claim is one extracted factual claim about Python.
type Claim struct {
	Claim     string `json:"claim"`
	Section   string `json:"section"`
	Checkable bool   `json:"checkable"`
	Code      string `json:"code,omitempty"`
}

type claimExtraction struct {
	Claims []Claim `json:"claims"`
}

func (c *claimExtraction) Validate() error {
	for i, cl := range c.Claims {
		if strings.TrimSpace(cl.Claim) == "" {
			return fmt.Errorf("claims[%d] has empty claim text", i)
		}
		if cl.Checkable && strings.TrimSpace(cl.Code) == "" {
			return fmt.Errorf("claims[%d] is checkable but has no code", i)
		}
	}
	return nil
}

// ClaimResult is a claim plus its sandbox execution outcome.
type ClaimResult struct {
	Claim
	Executed bool   `json:"executed"`
	Held     *bool  `json:"held,omitempty"` // nil when not executed
	Output   string `json:"output,omitempty"`
}

// ClaimVerdict is the reviewer's ruling on one claim.
type ClaimVerdict struct {
	Claim    string `json:"claim"`
	Verdict  string `json:"verdict"` // correct | incorrect | unsupported
	Citation string `json:"citation"`
}

// AccuracyReview is the accuracy pass result.
type AccuracyReview struct {
	Score    float64        `json:"score"`
	Verdicts []ClaimVerdict `json:"verdicts"`
	Critique string         `json:"critique"`
}

func (r *AccuracyReview) Validate() error {
	if r.Score < 1 || r.Score > 10 {
		return fmt.Errorf("score %v is outside 1-10", r.Score)
	}
	for i, v := range r.Verdicts {
		switch v.Verdict {
		case "correct", "incorrect", "unsupported":
		default:
			return fmt.Errorf("verdicts[%d] has unknown verdict %q (want correct, incorrect, or unsupported)", i, v.Verdict)
		}
	}
	if strings.TrimSpace(r.Critique) == "" {
		return fmt.Errorf("critique is empty")
	}
	return nil
}

// pedagogyDimensions are the required rubric keys of the pedagogy pass.
var pedagogyDimensions = []string{"concept_ordering", "cognitive_load", "concrete_before_abstract", "worked_examples"}

// PedagogyReview is the pedagogy pass result.
type PedagogyReview struct {
	Scores   map[string]float64 `json:"scores"`
	Score    float64            `json:"score"`
	Critique string             `json:"critique"`
}

func (r *PedagogyReview) Validate() error {
	for _, key := range pedagogyDimensions {
		v, ok := r.Scores[key]
		if !ok {
			return fmt.Errorf("scores is missing %q (required: %s)", key, strings.Join(pedagogyDimensions, ", "))
		}
		if v < 1 || v > 10 {
			return fmt.Errorf("score %q = %v is outside 1-10", key, v)
		}
	}
	if r.Score < 1 || r.Score > 10 {
		return fmt.Errorf("score %v is outside 1-10", r.Score)
	}
	if strings.TrimSpace(r.Critique) == "" {
		return fmt.Errorf("critique is empty")
	}
	return nil
}

// ToneReview is the tone pass result.
type ToneReview struct {
	Score    float64 `json:"score"`
	Critique string  `json:"critique"`
}

func (r *ToneReview) Validate() error {
	if r.Score < 1 || r.Score > 10 {
		return fmt.Errorf("score %v is outside 1-10", r.Score)
	}
	if strings.TrimSpace(r.Critique) == "" {
		return fmt.Errorf("critique is empty")
	}
	return nil
}

// MultiReviewRecord is one round's full three-pass record, persisted to
// generated/reviews/script-multipass-round-N.json.
type MultiReviewRecord struct {
	Round        int                `json:"round"`
	Model        string             `json:"model"`
	ArtifactHash string             `json:"artifact_hash"`
	Claims       []ClaimResult      `json:"claims"`
	Accuracy     AccuracyReview     `json:"accuracy"`
	Pedagogy     PedagogyReview     `json:"pedagogy"`
	Tone         ToneReview         `json:"tone"`
	Weights      map[string]float64 `json:"weights"`
	Weighted     float64            `json:"weighted"`
	Threshold    float64            `json:"threshold"`
	Passed       bool               `json:"passed"`
	ReviewedAt   time.Time          `json:"reviewed_at"`
}

// Critiques joins the three passes' critiques for regeneration injection.
func (r *MultiReviewRecord) Critiques() string {
	var b strings.Builder
	fmt.Fprintf(&b, "ACCURACY (scored %.1f/10): %s\n", r.Accuracy.Score, r.Accuracy.Critique)
	for _, v := range r.Accuracy.Verdicts {
		if v.Verdict != "correct" {
			fmt.Fprintf(&b, "  - %s claim: %q (%s)\n", v.Verdict, v.Claim, v.Citation)
		}
	}
	fmt.Fprintf(&b, "PEDAGOGY (scored %.1f/10): %s\n", r.Pedagogy.Score, r.Pedagogy.Critique)
	fmt.Fprintf(&b, "TONE (scored %.1f/10): %s", r.Tone.Score, r.Tone.Critique)
	return b.String()
}

// accuracyPromptData feeds review_claims.tmpl and review_accuracy.tmpl.
type accuracyPromptData struct {
	Audience        string
	Outline         string
	Artifact        string
	ClaimResults    string
	VerifiedOutputs string
}

// pedagogyPromptData feeds review_pedagogy.tmpl.
type pedagogyPromptData struct {
	Audience string
	PaceWPM  int
	Outline  string
	Artifact string
}

// tonePromptData feeds review_tone.tmpl.
type tonePromptData struct {
	Tone     string
	Audience string
	Artifact string
}

// extractClaims asks the review model for the script's factual claims.
func extractClaims(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, artifact []byte) ([]Claim, error) {
	data := accuracyPromptData{Audience: cfg.Style.Audience, Outline: l.Body, Artifact: string(artifact)}
	system, user, err := e.renderPrompt(reviewClaimsTemplateName, data)
	if err != nil {
		return nil, err
	}
	var out claimExtraction
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskReview, system, user, 0, 4096, &out, out.Validate)
	if err != nil {
		return nil, fmt.Errorf("extracting claims: %w", err)
	}
	return out.Claims, nil
}

// executeClaims runs each checkable claim's verification program in the
// sandbox. With no CodeRunner available every claim stays un-executed and
// the scoring model judges them all.
func executeClaims(ctx context.Context, e *Env, claims []Claim) []ClaimResult {
	results := make([]ClaimResult, 0, len(claims))
	for _, c := range claims {
		result := ClaimResult{Claim: c}
		if c.Checkable && c.Code != "" && e.CodeRunner != nil {
			res, err := e.CodeRunner.RunPython(ctx, c.Code, codeTimeout)
			if err == nil {
				held := res.Ok()
				result.Executed = true
				result.Held = &held
				if !held {
					result.Output = tailLines(res.Stderr, 4)
				}
			}
		}
		results = append(results, result)
	}
	return results
}

// claimResultsSummary renders claim execution outcomes for the accuracy
// scoring prompt.
func claimResultsSummary(results []ClaimResult) string {
	if len(results) == 0 {
		return "(no claims extracted)"
	}
	var b strings.Builder
	for i, r := range results {
		fmt.Fprintf(&b, "%d. [%s] %s\n", i+1, r.Section, r.Claim.Claim)
		switch {
		case r.Executed && r.Held != nil && *r.Held:
			fmt.Fprintf(&b, "   EXECUTED: verification program passed — the claim held\n")
		case r.Executed && r.Held != nil:
			fmt.Fprintf(&b, "   EXECUTED: verification program FAILED — the claim is false\n   %s\n", r.Output)
		case r.Checkable:
			fmt.Fprintf(&b, "   not executed (no sandbox available) — judge it yourself\n")
		default:
			fmt.Fprintf(&b, "   not code-checkable — judge it against Python documentation\n")
		}
	}
	return strings.TrimSpace(b.String())
}

// runMultiReview performs the three passes on one script draft.
func runMultiReview(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, artifact []byte, round int) (*MultiReviewRecord, error) {
	record := &MultiReviewRecord{
		Round:        round,
		Model:        cfg.Pipeline.LLMReview,
		ArtifactHash: project.HashBytes(artifact),
		Threshold:    cfg.Pipeline.ReviewThreshold,
		Weights: map[string]float64{
			"accuracy": weightAccuracy, "pedagogy": weightPedagogy, "tone": weightTone,
		},
		ReviewedAt: time.Now().UTC(),
	}

	// Pass 1: accuracy — extract claims, execute the checkable ones, score.
	fmt.Fprintf(e.out(), "  → review    round %d accuracy: extracting claims (%s)...\n", round, cfg.Pipeline.LLMReview)
	claims, err := extractClaims(ctx, e, l, cfg, artifact)
	if err != nil {
		return nil, err
	}
	record.Claims = executeClaims(ctx, e, claims)
	executed := 0
	for _, c := range record.Claims {
		if c.Executed {
			executed++
		}
	}
	fmt.Fprintf(e.out(), "    %d claim(s), %d executed in the sandbox\n", len(record.Claims), executed)

	data := accuracyPromptData{
		Audience:        cfg.Style.Audience,
		Outline:         l.Body,
		Artifact:        string(artifact),
		ClaimResults:    claimResultsSummary(record.Claims),
		VerifiedOutputs: verifiedOutputsSummary(l),
	}
	system, user, err := e.renderPrompt(reviewAccuracyTemplateName, data)
	if err != nil {
		return nil, err
	}
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskReview, system, user, 0, 4096, &record.Accuracy, record.Accuracy.Validate)
	if err != nil {
		return nil, fmt.Errorf("accuracy pass: %w", err)
	}

	// Pass 2: pedagogy.
	fmt.Fprintf(e.out(), "  → review    round %d pedagogy pass...\n", round)
	pdata := pedagogyPromptData{Audience: cfg.Style.Audience, PaceWPM: cfg.Style.PaceWPM, Outline: l.Body, Artifact: string(artifact)}
	system, user, err = e.renderPrompt(reviewPedagogyTemplateName, pdata)
	if err != nil {
		return nil, err
	}
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskReview, system, user, 0, 2048, &record.Pedagogy, record.Pedagogy.Validate)
	if err != nil {
		return nil, fmt.Errorf("pedagogy pass: %w", err)
	}

	// Pass 3: tone.
	fmt.Fprintf(e.out(), "  → review    round %d tone pass...\n", round)
	tdata := tonePromptData{Tone: cfg.Style.Tone, Audience: cfg.Style.Audience, Artifact: string(artifact)}
	system, user, err = e.renderPrompt(reviewToneTemplateName, tdata)
	if err != nil {
		return nil, err
	}
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskReview, system, user, 0, 2048, &record.Tone, record.Tone.Validate)
	if err != nil {
		return nil, fmt.Errorf("tone pass: %w", err)
	}

	record.Weighted = weightAccuracy*record.Accuracy.Score +
		weightPedagogy*record.Pedagogy.Score +
		weightTone*record.Tone.Score
	record.Passed = record.Weighted >= record.Threshold
	return record, nil
}

// runReviewStage is the review stage: gate generated/script.json through
// the three-pass review; below-threshold drafts are regenerated with every
// critique injected (max 2 regenerations), and the best draft wins.
func runReviewStage(ctx context.Context, e *Env, _ *project.Course, l *project.Lesson, cfg config.Config) error {
	scriptPath := filepath.Join(l.GeneratedDir(), ScriptFileName)
	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no %s yet — the script stage must run first", ScriptFileName)
		}
		return fmt.Errorf("reading %s: %w", scriptPath, err)
	}
	var probe Script
	if err := json.Unmarshal(raw, &probe); err != nil {
		return fmt.Errorf("parsing %s (delete it and re-run the script stage): %w", scriptPath, err)
	}

	current := raw
	var (
		bestScore float64
		bestDraft = raw
		passed    bool
		rounds    int
	)
	for round := 1; round <= maxReviewRounds; round++ {
		rounds = round
		record, err := runMultiReview(ctx, e, l, cfg, current, round)
		if err != nil {
			return err
		}
		recordPath := filepath.Join(l.GeneratedDir(), ReviewsDirName, fmt.Sprintf("script-multipass-round-%d.json", round))
		if err := writeJSON(recordPath, record); err != nil {
			return err
		}
		fmt.Fprintf(e.out(), "    accuracy %.1f × %.2f + pedagogy %.1f × %.2f + tone %.1f × %.2f = %.1f (threshold %.1f) — %s\n",
			record.Accuracy.Score, weightAccuracy, record.Pedagogy.Score, weightPedagogy,
			record.Tone.Score, weightTone, record.Weighted, record.Threshold, passFailWord(record.Passed))

		if record.Weighted > bestScore {
			bestScore = record.Weighted
			bestDraft = current
		}
		if record.Passed {
			passed = true
			break
		}
		if round == maxReviewRounds {
			break
		}

		fmt.Fprintf(e.out(), "    regenerating script with all three critiques...\n")
		next, err := generateScript(ctx, e, l, cfg, record.Critiques())
		if err != nil {
			return err
		}
		encoded, err := json.MarshalIndent(next, "", "  ")
		if err != nil {
			return fmt.Errorf("encoding regenerated script: %w", err)
		}
		current = append(encoded, '\n')
	}

	if !passed {
		fmt.Fprintf(e.out(),
			"  ⚠ review    weighted score %.1f after %d rounds (threshold %.1f) — kept the best draft; critiques are in %s\n",
			bestScore, rounds, cfg.Pipeline.ReviewThreshold, filepath.Join(project.GeneratedDirName, ReviewsDirName),
		)
	}
	return writeFileAtomic(scriptPath, bestDraft)
}
