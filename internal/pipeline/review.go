package pipeline

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
	"github.com/enfec/coursesmith/internal/project"
)

// ReviewsDirName is where per-round review records are persisted, under the
// lesson's generated dir.
const ReviewsDirName = "reviews"

// maxReviewRounds is the initial review plus up to two regenerations.
const maxReviewRounds = 3

// requiredScores are the rubric dimensions the reviewer must score.
var requiredScores = []string{"technical_accuracy", "clarity", "engagement", "pacing"}

// Review is the critic's verdict on one artifact.
type Review struct {
	Scores   map[string]float64 `json:"scores"`
	Overall  float64            `json:"overall"`
	Critique string             `json:"critique"`
}

// Validate checks the rubric shape: all four dimensions present and every
// score in 1-10.
func (r *Review) Validate() error {
	for _, key := range requiredScores {
		v, ok := r.Scores[key]
		if !ok {
			return fmt.Errorf("scores is missing %q (required: %s)", key, strings.Join(requiredScores, ", "))
		}
		if v < 1 || v > 10 {
			return fmt.Errorf("score %q = %v is outside 1-10", key, v)
		}
	}
	if len(r.Scores) != len(requiredScores) {
		return fmt.Errorf("scores has unexpected extra keys (want exactly: %s)", strings.Join(requiredScores, ", "))
	}
	if r.Overall < 1 || r.Overall > 10 {
		return fmt.Errorf("overall = %v is outside 1-10", r.Overall)
	}
	if strings.TrimSpace(r.Critique) == "" {
		return fmt.Errorf("critique is empty")
	}
	return nil
}

// reviewRecord is the persisted audit trail of one review round.
type reviewRecord struct {
	Kind         string    `json:"kind"`
	Round        int       `json:"round"`
	Model        string    `json:"model"`
	Threshold    float64   `json:"threshold"`
	Passed       bool      `json:"passed"`
	ArtifactHash string    `json:"artifact_hash"`
	Review       Review    `json:"review"`
	ReviewedAt   time.Time `json:"reviewed_at"`
}

// reviewPromptData feeds prompts/review_rubric.tmpl.
type reviewPromptData struct {
	Kind     string
	Audience string
	Tone     string
	PaceWPM  int
	Outline  string
	Artifact string
	// VerifiedOutputs is execution ground truth from the verify stage; the
	// critic flags artifact claims that contradict it.
	VerifiedOutputs string
}

// reviewArtifact scores one artifact against the rubric with the review model.
func reviewArtifact(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, kind string, artifact []byte) (*Review, error) {
	data := reviewPromptData{
		Kind:            kind,
		Audience:        cfg.Style.Audience,
		Tone:            cfg.Style.Tone,
		PaceWPM:         cfg.Style.PaceWPM,
		Outline:         l.Body,
		Artifact:        string(artifact),
		VerifiedOutputs: verifiedOutputsSummary(l),
	}
	system, user, err := e.renderPrompt(reviewTemplateName, data)
	if err != nil {
		return nil, err
	}
	var review Review
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskReview, system, user, 0, 2048, &review, review.Validate)
	if err != nil {
		return nil, fmt.Errorf("reviewing %s: %w", kind, err)
	}
	return &review, nil
}

var kindSlugRe = regexp.MustCompile(`[^a-z0-9]+`)

// kindSlug turns an artifact kind ("diagram:memory-model") into a safe file
// name fragment ("diagram-memory-model").
func kindSlug(kind string) string {
	return strings.Trim(kindSlugRe.ReplaceAllString(strings.ToLower(kind), "-"), "-")
}

// reviewGate is the quality gate every generated artifact goes through:
// review against the rubric; while below pipeline.review_threshold,
// regenerate with the critique injected (max 2 regenerations) and re-review.
// It returns the best-scoring version of the artifact and whether any round
// passed. Every round is persisted to generated/reviews/ for auditability.
//
// The gate is soft: when every round fails, the best draft is returned with
// a warning rather than an error, so long unattended runs complete and the
// critiques remain on disk to act on.
func (e *Env) reviewGate(ctx context.Context, l *project.Lesson, cfg config.Config, kind string, artifact []byte, regenerate func(ctx context.Context, critique string) ([]byte, error)) ([]byte, bool, error) {
	threshold := cfg.Pipeline.ReviewThreshold
	current := artifact
	var (
		bestScore float64
		bestDraft = artifact
		passed    bool
		rounds    int
	)
	for round := 1; round <= maxReviewRounds; round++ {
		rounds = round
		fmt.Fprintf(e.out(), "  → review    %s round %d: scoring (%s)...\n", kind, round, cfg.Pipeline.LLMReview)
		review, err := reviewArtifact(ctx, e, l, cfg, kind, current)
		if err != nil {
			return nil, false, err
		}
		pass := review.Overall >= threshold
		record := reviewRecord{
			Kind:         kind,
			Round:        round,
			Model:        cfg.Pipeline.LLMReview,
			Threshold:    threshold,
			Passed:       pass,
			ArtifactHash: project.HashBytes(current),
			Review:       *review,
			ReviewedAt:   time.Now().UTC(),
		}
		recordPath := filepath.Join(l.GeneratedDir(), ReviewsDirName, fmt.Sprintf("%s-round-%d.json", kindSlug(kind), round))
		if err := writeJSON(recordPath, record); err != nil {
			return nil, false, err
		}
		fmt.Fprintf(e.out(), "    overall %.1f (threshold %.1f) — %s\n", review.Overall, threshold, passFailWord(pass))

		if review.Overall > bestScore {
			bestScore = review.Overall
			bestDraft = current
		}
		if pass {
			passed = true
			break
		}
		if round == maxReviewRounds || regenerate == nil {
			break
		}

		fmt.Fprintf(e.out(), "    regenerating %s with critique...\n", kind)
		next, err := regenerate(ctx, review.Critique)
		if err != nil {
			return nil, false, err
		}
		current = next
	}

	if !passed {
		fmt.Fprintf(e.out(),
			"  ⚠ review    %s scored %.1f after %d rounds (threshold %.1f) — kept the best draft; critiques are in %s\n",
			kind, bestScore, rounds, threshold, filepath.Join(project.GeneratedDirName, ReviewsDirName),
		)
	}
	return bestDraft, passed, nil
}

// The review stage itself lives in multireview.go (three-pass accuracy /
// pedagogy / tone gate). reviewGate below remains the quality gate used by
// artifact-producing stages (visuals, quiz) at generation time.

func passFailWord(pass bool) string {
	if pass {
		return "passed"
	}
	return "below threshold"
}
