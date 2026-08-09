package pipeline

// The critic: reading the finished piece as a piece.
//
// Everything before this judges one segment at a time. The template validators
// check a plan against its own rules; the review gate (plan_review.go) scores a
// plan against a quality rubric. Both are looking at one clip with the others
// out of frame, and there is a whole class of defect they structurally cannot
// see — the one a viewer notices first.
//
// A segment can satisfy every rule it is subject to and still be wrong in the
// piece:
//
//   - it covers ground segment three already covered, in different words
//   - it contradicts something an earlier segment established
//   - it is fine, and it does not advance the argument the piece is making
//   - it opens as though the viewer has not been watching for four minutes
//
// Each of those is invisible from inside the segment. The plan is valid, the
// rubric scores it well, and it ships. This is what "some clips had bad content
// and did not belong" is describing, and it is why the fix could not be a better
// per-segment prompt: no amount of care writing segment seven tells it what
// segment three said.
//
// == What it does about it ==
//
// One call reads every segment's narration in order, with the piece's angle in
// hand, and returns only the segments that are wrong and why. Those are re-planned
// through their own template with the criticism attached — the same
// SnippetSpec.Critique path the review gate uses, so a repair still passes the
// template's validators and cannot fix a repetition by breaking the beat budget.
//
// == Why it never fails the run ==
//
// The same asymmetry gateSegmentPlan documents. A plan that breaks its
// template's rules cannot be rendered, so failing is the only option there. A
// plan the critic dislikes renders perfectly well — losing a finished video to a
// critic that timed out would be the gate doing more harm than the flaw it was
// looking for. Every verdict is printed either way, so a repair that did not
// happen is visible rather than silent.

import (
	"context"
	"fmt"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
)

const comboCriticTemplateName = "combo_critic.tmpl"

const (
	// criticRepairRounds is how many times the critic is told its own reply was
	// malformed. Low: the reply shape is a short list, not a budget to satisfy.
	criticRepairRounds = 2
	// maxCriticRepairs caps how many segments one critic pass re-plans.
	//
	// Every repair is a full planning call, so an uncapped critic on a
	// twelve-segment piece can double the cost of the run. Four is the point past
	// which the problem is not individual segments: a critic flagging five of
	// eight is describing an outline that did not work, and re-planning five
	// segments against the same outline produces five differently-wrong ones.
	//
	// Whatever is not repaired is REPORTED rather than dropped quietly — a cap
	// nobody is told about reads as "the critic found four problems", which is a
	// different and more comforting statement than the truth.
	maxCriticRepairs = 4
	// criticNarrationWords is how much of each segment the critic reads.
	//
	// Enough to judge what a segment says and not enough to judge how it says it.
	// That is the intended limit: this pass is about whether a segment belongs,
	// and a critic handed every word starts editing prose, which the review gate
	// already does better with the whole rubric in hand.
	criticNarrationWords = 70
)

// ComboVerdict is the critic's finding on one segment.
type ComboVerdict struct {
	// ID names the segment. Echoed from what the critic was given, so a reply
	// about a segment that does not exist is caught rather than silently dropped.
	ID string `json:"id"`
	// Problem is what is wrong with this segment IN THE PIECE — not in itself.
	Problem string `json:"problem"`
	// Fix is what the rewrite should do instead. Handed to the segment's writer
	// verbatim, so it has to be an instruction rather than a complaint.
	Fix string `json:"fix"`
}

// ComboCritique is the whole pass's finding.
type ComboCritique struct {
	// Verdicts holds only the segments that are wrong. An empty list is the
	// expected result and is not a failure of the critic.
	Verdicts []ComboVerdict `json:"verdicts"`
}

// criticSegment is one segment as the critic reads it.
type criticSegment struct {
	ID        string
	N         int
	Template  string
	Role      string
	Heading   string
	Title     string
	Narration string
}

// criticiseCombo reads the planned piece as a whole and repairs what does not
// belong. It returns the plan it is prepared to ship — the original when there
// is nothing to fix, or when fixing it did not work.
func (e *Env) criticiseCombo(ctx context.Context, spec *ComboSpec, plan *ComboPlan, cfg config.Config, segSpecs map[string]SnippetSpec) *ComboPlan {
	if e.Router == nil || plan == nil || len(plan.Segments) < 2 {
		return plan
	}

	read := make([]criticSegment, 0, len(plan.Segments))
	byID := map[string]int{}
	for i, seg := range plan.Segments {
		if seg.Plan == nil {
			continue
		}
		byID[seg.ID] = i
		var sb strings.Builder
		for _, b := range seg.Plan.Beats {
			sb.WriteString(b.Narration)
			sb.WriteString(" ")
		}
		read = append(read, criticSegment{
			ID:        seg.ID,
			N:         i + 1,
			Template:  seg.Template,
			Role:      segmentRole(spec, seg.ID),
			Heading:   segmentHeading(spec, seg.ID),
			Title:     seg.Plan.Title,
			Narration: clampWords(collapseSpaces(sb.String()), criticNarrationWords),
		})
	}
	if len(read) < 2 {
		return plan
	}

	system, user, err := e.renderPrompt(comboCriticTemplateName, map[string]any{
		"Title":    plan.Title,
		"Angle":    spec.Angle,
		"Brief":    spec.Brief,
		"Segments": read,
		"Audience": cfg.Style.Audience,
	})
	if err != nil {
		fmt.Fprintf(e.out(), "    ! the critic could not run: %s\n", errSummary(err))
		return plan
	}

	var critique ComboCritique
	err = e.completeJSONLenientRounds(ctx, cfg.Pipeline, llm.TaskContent, system, user, 0.4, thinkingBudget(6144), criticRepairRounds, effortInterlocking, &critique, func() error {
		return normalizeCritique(&critique, byID)
	})
	if err != nil {
		// A critic that cannot answer is not a reason to lose a finished piece.
		fmt.Fprintf(e.out(), "    ! the critic could not read the piece, shipping it unrepaired (%s)\n", errSummary(err))
		return plan
	}
	if len(critique.Verdicts) == 0 {
		fmt.Fprintf(e.out(), "    the critic found nothing out of place across %d segments\n", len(read))
		return plan
	}

	fmt.Fprintf(e.out(), "    the critic flagged %d of %d segments:\n", len(critique.Verdicts), len(read))
	for _, v := range critique.Verdicts {
		fmt.Fprintf(e.out(), "      %-16s %s\n", v.ID, truncateForLog(v.Problem, 88))
	}

	repair := critique.Verdicts
	if len(repair) > maxCriticRepairs {
		// Said out loud. A cap nobody is told about reads as a smaller problem
		// than the one that was actually found, and this particular overflow is
		// worth reading as a signal: past four, the defect is usually the outline.
		fmt.Fprintf(e.out(), "    ! %d flagged but only %d will be re-planned — past this the problem is the outline rather than the segments, so re-directing beats repairing\n",
			len(repair), maxCriticRepairs)
		repair = repair[:maxCriticRepairs]
	}

	repaired := 0
	for _, v := range repair {
		i := byID[v.ID]
		seg := plan.Segments[i]
		base, ok := segSpecs[v.ID]
		if !ok {
			continue
		}
		tpl, known := SnippetTemplates[seg.Template]
		if !known {
			continue
		}
		retry := base
		retry.Template = seg.Template
		// The problem AND the instruction. A writer told only what is wrong
		// rewrites around the complaint and usually reintroduces it from another
		// direction; told what to do instead, it has somewhere to go.
		retry.Critique = fmt.Sprintf("This segment was read in the context of the whole piece and does not belong as written.\n\nWhat is wrong: %s\n\nWhat to do instead: %s", v.Problem, v.Fix)
		planner := tpl.Plan
		if planner == nil {
			planner = planSnippetDefault
		}
		next, err := planner(ctx, e, retry, cfg)
		if err != nil {
			fmt.Fprintf(e.out(), "      ! %s could not be re-planned, keeping the original (%s)\n", v.ID, errSummary(err))
			continue
		}
		next.Template = seg.Template
		plan.Segments[i].Plan = next
		repaired++
	}
	fmt.Fprintf(e.out(), "    %d of %d flagged segments re-planned\n", repaired, len(repair))
	return plan
}

// normalizeCritique cleans the reply and rejects verdicts that do not point at
// a real segment.
//
// A verdict naming a segment that does not exist is not a harmless typo: it is
// the critic having lost track of which segment it was reading, which means the
// verdicts around it are about the wrong ones too. Sending it back is cheaper
// than repairing a good segment against a complaint aimed at another.
func normalizeCritique(c *ComboCritique, byID map[string]int) error {
	out := make([]ComboVerdict, 0, len(c.Verdicts))
	seen := map[string]bool{}
	for _, v := range c.Verdicts {
		v.ID = strings.TrimSpace(v.ID)
		v.Problem = collapseSpaces(v.Problem)
		v.Fix = collapseSpaces(v.Fix)
		if v.ID == "" || v.Problem == "" {
			continue
		}
		if _, ok := byID[v.ID]; !ok {
			ids := make([]string, 0, len(byID))
			for id := range byID {
				ids = append(ids, id)
			}
			return fmt.Errorf("verdict names segment %q, which is not in this piece. Use the ids exactly as given: %s", v.ID, strings.Join(ids, ", "))
		}
		if seen[v.ID] {
			continue // two complaints about one segment; the first is enough to re-plan it
		}
		if strings.TrimSpace(v.Fix) == "" {
			return fmt.Errorf("the verdict on %q says what is wrong but not what to do instead. The fix is handed to the writer verbatim, so it has to be an instruction", v.ID)
		}
		seen[v.ID] = true
		out = append(out, v)
	}
	c.Verdicts = out
	return nil
}

// segmentRole and segmentHeading pull a segment's argument-level facts off the
// spec, which is where the director recorded them. The plan does not carry
// either: a template's plan is about the clip, and the role and heading are
// about the clip's place in the piece.
func segmentRole(spec *ComboSpec, id string) string {
	for _, s := range spec.Segments {
		if s.ID == id {
			return s.Role
		}
	}
	return ""
}

func segmentHeading(spec *ComboSpec, id string) string {
	for _, s := range spec.Segments {
		if s.ID == id {
			return s.Heading
		}
	}
	return ""
}
