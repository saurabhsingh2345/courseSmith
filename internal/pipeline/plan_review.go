package pipeline

// The quality gate on the short-form path.
//
// The lesson path has always had one: script → three-pass review (accuracy,
// pedagogy, tone) against a threshold, regenerating with the critique. The
// snippet and reel path had none, and said so in a comment — SnippetStageOrder
// listed "no review rounds" as a feature, on the reasoning that the plan stage
// does all the thinking in one call and everything after it is the ordinary
// video path.
//
// That reasoning holds for everything after the plan. It does not hold for the
// plan itself, which is exactly where the thinking happens and therefore exactly
// where a bad decision becomes a rendered video nobody looks at again. The
// surface being sold ran one unreviewed call straight to render, which is how a
// gauge captioned with a number that does not exist got voiced, timed,
// animated and published.
//
// So the gate goes on the plan. It reuses reviewGateWith — the round loop, the
// persisted audit trail in generated/reviews/, and the soft-fail that keeps the
// best draft — with the plan rubric rather than the prose one. What is new here is
// only the marshalling: a SnippetPlan is a struct, the gate speaks []byte, and a
// regeneration means re-planning through the template rather than re-prompting a
// writer.

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/project"
)

// gateSegmentPlan reviews one planned segment and returns the best draft.
//
// NEVER fails the run, and returns the original plan on any error. That is a
// deliberate asymmetry with the validators inside planning: a plan that breaks
// its template's rules cannot be rendered, so failing is the only option, but a
// plan the *reviewer* dislikes renders perfectly well. Losing a finished video to
// a critic that timed out would be the gate doing more harm than the flaw it was
// looking for. Every verdict is on disk either way.
func (e *Env) gateSegmentPlan(ctx context.Context, l *project.Lesson, cfg config.Config, spec SnippetSpec, plan *SnippetPlan) *SnippetPlan {
	if plan == nil || e.Router == nil {
		return plan
	}
	kind := "plan:" + spec.ID
	if spec.ID == "" {
		kind = "plan:" + spec.Template
	}

	artifact, err := json.Marshal(plan)
	if err != nil {
		return plan
	}

	best, _, err := e.reviewGateWith(ctx, l, cfg, planRubric, kind, artifact,
		func(ctx context.Context, critique string) ([]byte, error) {
			// Re-plan through the template with the critique in hand. The template's
			// own validators run again, so a regeneration cannot fix a fabrication
			// by breaking the beat budget.
			retry := spec
			retry.Critique = critique
			tpl, ok := SnippetTemplates[retry.Template]
			if !ok {
				return nil, fmt.Errorf("template %q disappeared between planning and review", retry.Template)
			}
			planner := tpl.Plan
			if planner == nil {
				planner = planSnippetDefault
			}
			next, err := planner(ctx, e, retry, cfg)
			if err != nil {
				return nil, err
			}
			next.Template = retry.Template
			return json.Marshal(next)
		})
	if err != nil {
		fmt.Fprintf(e.out(), "    ! review of %s could not run, keeping the plan as-is (%s)\n", kind, errSummary(err))
		return plan
	}

	var out SnippetPlan
	if err := json.Unmarshal(best, &out); err != nil {
		return plan
	}
	// The round trip through JSON drops the unexported fields the validators set
	// (targetWords), and Scenes runs off the exported ones — but a plan that came
	// back from the gate has already passed its template's validator inside the
	// planner, so nothing downstream re-checks them. Template is restored because
	// scene dispatch reads it.
	out.Template = spec.Template
	return &out
}
