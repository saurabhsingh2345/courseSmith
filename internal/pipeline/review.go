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

// rubric is which set of questions an artifact is judged against.
//
// Two exist. The original judges prose against accuracy/clarity/engagement/
// pacing, which is the right set for a script somebody will read. A snippet plan
// needs a different set entirely — its failures are inventing a figure and
// narrating the template instead of the subject, neither of which any of those
// four dimensions asks about. What must NOT differ is the machinery around the
// judgement: the round loop, the persisted audit trail, and the soft-fail that
// keeps the best draft. Those were the expensive part to get right, so the rubric
// is a parameter rather than a second copy of the gate.
type rubric struct {
	// Template is the prompt file that renders the judgement.
	Template string
	// Scores are the dimensions the reviewer must return, and exactly those. A
	// missing one is a rejected review, so this doubles as the contract between
	// the prompt and Validate.
	Scores []string
	// FatalBelow fails a dimension outright, whatever the overall says.
	//
	// Fabrication needs this. An average cannot express "one invented figure
	// ruins the clip": a plan that is excellent on three dimensions and made up a
	// number scores well and ships, which is precisely the outcome observed. A
	// floor on the dimension that matters is the only thing an averaging rubric
	// cannot route around.
	FatalBelow map[string]float64
}

var (
	// scriptRubric judges written prose. The original four dimensions.
	scriptRubric = rubric{Template: reviewTemplateName, Scores: requiredScores}

	// planRubric judges a snippet or reel plan before it is spoken aloud.
	//
	// The dimensions are the four failures actually observed in shipped output,
	// not a general notion of quality:
	//   fabrication      — a figure, product, person or case study that does not exist
	//   concreteness     — "streamline workflows" where a real name belonged
	//   teaching         — narration that describes the picture instead of the subject
	//   non_redundancy   — beats that restate rather than advance
	planRubric = rubric{
		Template: reviewPlanTemplateName,
		Scores:   []string{"fabrication", "concreteness", "teaching", "non_redundancy"},
		// Nothing else is fatal. A dull clip is worth shipping and fixing; a clip
		// that states something untrue is not, because the cost lands on whoever
		// believes it.
		FatalBelow: map[string]float64{"fabrication": 7},
	}
)

// validate checks a review against this rubric: every dimension present, no
// extras, everything in range.
func (rb rubric) validate(r *Review) error {
	for _, key := range rb.Scores {
		v, ok := r.Scores[key]
		if !ok {
			return fmt.Errorf("scores is missing %q (required: %s)", key, strings.Join(rb.Scores, ", "))
		}
		if v < 1 || v > 10 {
			return fmt.Errorf("score %q = %v is outside 1-10", key, v)
		}
	}
	if len(r.Scores) != len(rb.Scores) {
		return fmt.Errorf("scores has unexpected extra keys (want exactly: %s)", strings.Join(rb.Scores, ", "))
	}
	if r.Overall < 1 || r.Overall > 10 {
		return fmt.Errorf("overall = %v is outside 1-10", r.Overall)
	}
	if strings.TrimSpace(r.Critique) == "" {
		return fmt.Errorf("critique is empty")
	}
	return nil
}

// failedDimension returns the dimension that fell below its own floor, if any,
// so the caller can say which one rather than only that something did.
func (rb rubric) failedDimension(r *Review) (string, float64, bool) {
	for key, floor := range rb.FatalBelow {
		if v, ok := r.Scores[key]; ok && v < floor {
			return key, v, true
		}
	}
	return "", 0, false
}

// Review is the critic's verdict on one artifact.
type Review struct {
	Scores   map[string]float64 `json:"scores"`
	Overall  float64            `json:"overall"`
	Critique string             `json:"critique"`
}

// Validate checks the script rubric's shape: all four dimensions present and
// every score in 1-10. Delegates rather than repeating the checks, so the two
// rubrics cannot disagree about what a well-formed review is.
func (r *Review) Validate() error {
	return scriptRubric.validate(r)
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

// reviewPromptData feeds prompts/review_rubric.tmpl and review_plan.tmpl.
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
	// Facts and Gaps are the plan rubric's ground truth: what the piece
	// established, and what it looked for and could not find. A fabrication
	// score is only meaningful against them — without the sheet the judge is
	// being asked whether a claim *sounds* invented, which is the same guess the
	// writer already made.
	Facts []string
	Gaps  []string
	// Grounded says whether a search actually ran. A judge holding an ungrounded
	// sheet to a sourced standard would reject everything.
	Grounded bool
}

// reviewArtifact scores one artifact against a rubric with the review model.
func reviewArtifact(ctx context.Context, e *Env, l *project.Lesson, cfg config.Config, rb rubric, kind string, artifact []byte) (*Review, error) {
	sub, err := LoadSubstance(l)
	if err != nil {
		return nil, err
	}
	data := reviewPromptData{
		Kind:            kind,
		Audience:        cfg.Style.Audience,
		Tone:            cfg.Style.Tone,
		PaceWPM:         cfg.Style.PaceWPM,
		Outline:         l.Body,
		Artifact:        string(artifact),
		VerifiedOutputs: verifiedOutputsSummary(l),
		Facts:           substanceLines(sub),
		Gaps:            substanceGaps(sub),
		Grounded:        sub != nil && sub.Grounded,
	}
	system, user, err := e.renderPrompt(rb.Template, data)
	if err != nil {
		return nil, err
	}
	var review Review
	err = e.completeJSON(ctx, cfg.Pipeline, llm.TaskReview, system, user, 0, 2048, &review, func() error {
		return rb.validate(&review)
	})
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
	return e.reviewGateWith(ctx, l, cfg, scriptRubric, kind, artifact, regenerate)
}

// reviewGateWith is reviewGate against a chosen rubric. See rubric for why the
// questions vary and the machinery does not.
func (e *Env) reviewGateWith(ctx context.Context, l *project.Lesson, cfg config.Config, rb rubric, kind string, artifact []byte, regenerate func(ctx context.Context, critique string) ([]byte, error)) ([]byte, bool, error) {
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
		review, err := reviewArtifact(ctx, e, l, cfg, rb, kind, current)
		if err != nil {
			return nil, false, err
		}
		pass := review.Overall >= threshold
		// A fatal dimension overrides a passing average. Reported by name,
		// because "it failed" sends somebody reading four scores to work out
		// which.
		if dim, score, fatal := rb.failedDimension(review); fatal {
			pass = false
			fmt.Fprintf(e.out(), "    %s scored %.1f (floor %.1f) — fatal whatever the average says\n",
				dim, score, rb.FatalBelow[dim])
		}
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
