package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/enfec/coursesmith/internal/config"
	"github.com/enfec/coursesmith/internal/llm"
)

// How hard a stage asks its model to think ON A RETRY. Everything else, and
// every FIRST attempt, inherits the course's reasoning_effort.
//
// The axis is NOT importance — it is whether the stage has to satisfy several
// independent rules at once. Reasoning tokens bill as output tokens, so effort
// spent where a single well-specified field was wanted is money for nothing;
// but a stage juggling a beat count against a per-beat word range against a
// total word budget cannot land them all in one pass, and measurably does not:
// at low effort a 45-second clip came back 101 words (34s) against a budget the
// prompt calls a hard requirement, and at medium the same prompt landed 136
// words (46s). The correction rounds hid it — they were being spent re-fixing
// arithmetic instead of the content.
const (
	// effortInherit takes whatever the course configured.
	effortInherit = ""
	// effortInterlocking is for artifacts with several interacting numeric
	// rules, where one pass cannot satisfy all of them at once. It applies only
	// after a reply has been rejected — see completeWithRepairRounds.
	effortInterlocking = "medium"
)

// thinkingBudget widens an output budget to cover reasoning.
//
// A reasoning model spends its thinking from the SAME max_completion_tokens
// pool as the answer, and spends it FIRST. So a budget sized for the artifact —
// 6144 for a snippet plan, which is about four times what the JSON needs — is
// not generous once effort goes up: the thinking consumes it and the reply is
// cut off mid-object.
//
// That failure is nasty because it does not look like a budget problem. It
// surfaces as `response was truncated`, after the retry stack has burned four
// attempts, and the stage reports that it could not run — so the visible symptom
// is a missing quality gate, not a number that needs raising. It was introduced
// here the moment planning moved to effortInterlocking, and caught only because
// the truncation error names ReasoningEffort.
//
// Tripled rather than tuned: the point is headroom, and an unused ceiling costs
// nothing (only tokens actually emitted are billed). Callers that raise effort
// must pass their budget through here — the coupling is easy to forget, which
// is exactly why it is a named function and not a bigger literal at each site.
func thinkingBudget(contentTokens int) int { return contentTokens * 3 }

// completeWithRepair sends a system+user prompt and passes the reply through
// accept, which validates it and returns the normalized artifact (e.g. JSON
// as-is, or an SVG extracted from surrounding noise). If accept rejects the
// reply, the model gets exactly one correction round with the error appended
// to the conversation.
func (e *Env) completeWithRepair(
	ctx context.Context,
	pcfg config.Pipeline,
	task llm.TaskType,
	system, user string,
	temperature float64,
	maxTokens int,
	jsonMode bool,
	accept func(content string) (string, error),
) (string, error) {
	return e.completeWithRepairImages(ctx, pcfg, task, system, user, nil, temperature, maxTokens, jsonMode, accept)
}

// completeWithRepairImages is completeWithRepair with base64 PNGs attached
// to the user message (vision models).
func (e *Env) completeWithRepairImages(
	ctx context.Context,
	pcfg config.Pipeline,
	task llm.TaskType,
	system, user string,
	images []string,
	temperature float64,
	maxTokens int,
	jsonMode bool,
	accept func(content string) (string, error),
) (string, error) {
	return e.completeWithRepairRounds(ctx, pcfg, task, system, user, images, temperature, maxTokens, jsonMode, 1, effortInherit, accept)
}

// completeWithRepairRounds is the general form: up to `rounds` correction
// attempts after the first reply.
//
// Every rejection is carried forward, not just the latest one. A single-round
// loop quoting only the most recent error made models oscillate on artifacts
// with several independent numeric rules — attempt one missed the item count,
// attempt two fixed the count and blew the word budget, and the call failed
// having never seen both constraints at once. Showing the whole history of
// misses converges instead.
func (e *Env) completeWithRepairRounds(
	ctx context.Context,
	pcfg config.Pipeline,
	task llm.TaskType,
	system, user string,
	images []string,
	temperature float64,
	maxTokens int,
	jsonMode bool,
	rounds int,
	effort string,
	accept func(content string) (string, error),
) (string, error) {
	if e.Router == nil {
		return "", fmt.Errorf("no LLM router configured")
	}
	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: system},
		{Role: llm.RoleUser, Content: user},
	}
	req := llm.Request{
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		JSONMode:    jsonMode,
		Images:      images,
		// The FIRST attempt always runs at the course's configured effort, even
		// when this call asked for more. See the escalation note below.
		ReasoningEffort: effortInherit,
	}

	resp, err := e.Router.Complete(ctx, pcfg, task, req)
	if err != nil {
		return "", err
	}
	artifact, acceptErr := accept(resp.Content)
	if acceptErr == nil {
		return artifact, nil
	}

	rejections := []string{acceptErr.Error()}
	for round := 0; round < rounds; round++ {
		var sb strings.Builder
		sb.WriteString("Your reply was rejected. Every rule below has been broken by one of your attempts so far — satisfy all of them at once:\n")
		for _, r := range rejections {
			sb.WriteString("- " + r + "\n")
		}
		sb.WriteString("Respond again with only the corrected output — no fences, no commentary.")

		// Escalate the thinking budget only now, on a reply that was actually
		// rejected — this is the whole cost model of the stage.
		//
		// Paying for effort on the first attempt was measured and it is bad
		// value: a snippet plan's JSON is one to two thousand tokens, and at
		// medium the calls averaged 3,398 completion tokens, so roughly two
		// thirds of the bill was thinking — on every plan, including the many
		// that would have been right at low effort. Reasoning tokens bill as
		// completion, so that is not a rounding error: it took a 30-second clip
		// from about two cents to twenty.
		//
		// A rejection is the signal that this particular plan is one of the hard
		// ones — the beat count fought the word budget, or the template's own
		// rules did. THAT is worth thinking about, and only that. Most plans pass
		// first time and never pay.
		req.ReasoningEffort = effort

		req.Messages = append(messages,
			llm.Message{Role: llm.RoleAssistant, Content: resp.Content},
			llm.Message{Role: llm.RoleUser, Content: sb.String()},
		)
		resp, err = e.Router.Complete(ctx, pcfg, task, req)
		if err != nil {
			return "", fmt.Errorf("retrying after rejected output (%s): %w", strings.Join(rejections, "; "), err)
		}
		artifact, acceptErr = accept(resp.Content)
		if acceptErr == nil {
			return artifact, nil
		}
		if msg := acceptErr.Error(); !slices.Contains(rejections, msg) {
			rejections = append(rejections, msg)
		}
	}
	return "", fmt.Errorf("%s response invalid after %d correction round(s): %s",
		task, rounds, strings.Join(rejections, "; "))
}

// completeJSON is completeWithRepair specialized for JSON artifacts: the
// reply is parsed strictly into out (a non-nil pointer) and validated.
func (e *Env) completeJSON(
	ctx context.Context,
	pcfg config.Pipeline,
	task llm.TaskType,
	system, user string,
	temperature float64,
	maxTokens int,
	out any,
	validate func() error,
) error {
	return e.completeJSONWithImages(ctx, pcfg, task, system, user, nil, temperature, maxTokens, out, validate)
}

// completeJSONWithImages is completeJSON with PNGs attached (vision review).
func (e *Env) completeJSONWithImages(
	ctx context.Context,
	pcfg config.Pipeline,
	task llm.TaskType,
	system, user string,
	images []string,
	temperature float64,
	maxTokens int,
	out any,
	validate func() error,
) error {
	return e.completeJSONRounds(ctx, pcfg, task, system, user, images, temperature, maxTokens, 1, effortInherit, out, validate)
}

// completeJSONRounds is completeJSON with an explicit correction-round budget,
// for artifacts whose validation enforces several independent rules at once.
func (e *Env) completeJSONRounds(
	ctx context.Context,
	pcfg config.Pipeline,
	task llm.TaskType,
	system, user string,
	images []string,
	temperature float64,
	maxTokens int,
	rounds int,
	effort string,
	out any,
	validate func() error,
) error {
	_, err := e.completeWithRepairRounds(ctx, pcfg, task, system, user, images, temperature, maxTokens, true, rounds, effort,
		func(content string) (string, error) {
			if err := parseJSONStrict(content, out, validate); err != nil {
				return "", err
			}
			return content, nil
		})
	return err
}

// completeJSONLenientRounds is completeJSONRounds for artifacts where a field
// the model invented should be dropped rather than argued about.
//
// Strict decoding is the right default: for most of the pipeline an unexpected
// key means the model misread the schema, and saying so is cheap. It is the
// wrong default for an artifact with a large optional surface — a snippet plan
// has a dozen optional payloads and models routinely garnish one with a
// "description" or a "duration_sec" nobody asked for. That reply is a good clip
// with a spare key on it, and spending a correction round (and the whole reply)
// on the key is the worst of both: the creator waits longer for a plan that was
// already sitting there.
func (e *Env) completeJSONLenientRounds(
	ctx context.Context,
	pcfg config.Pipeline,
	task llm.TaskType,
	system, user string,
	temperature float64,
	maxTokens int,
	rounds int,
	effort string,
	out any,
	validate func() error,
) error {
	_, err := e.completeWithRepairRounds(ctx, pcfg, task, system, user, nil, temperature, maxTokens, true, rounds, effort,
		func(content string) (string, error) {
			if err := parseJSONLenient(content, out, validate); err != nil {
				return "", err
			}
			return content, nil
		})
	return err
}

// parseJSONStrict decodes content into out, rejecting unknown fields and
// trailing data, then runs validate. out is zeroed first so a failed prior
// attempt cannot leak fields into this one.
func parseJSONStrict(content string, out any, validate func() error) error {
	reflect.ValueOf(out).Elem().SetZero()
	dec := json.NewDecoder(bytes.NewReader([]byte(stripFences(content))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if dec.More() {
		return fmt.Errorf("trailing content after the JSON object")
	}
	if validate != nil {
		return validate()
	}
	return nil
}

// parseJSONLenient decodes the first JSON object in content into out, ignoring
// fields the schema does not know and anything written after the object, then
// runs validate.
//
// The trailing-content tolerance matters as much as the unknown-field one: a
// model that answers with the object and then a sentence about it has still
// answered, and the sentence is not worth a round trip.
func parseJSONLenient(content string, out any, validate func() error) error {
	reflect.ValueOf(out).Elem().SetZero()
	dec := json.NewDecoder(bytes.NewReader([]byte(stripFences(content))))
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if validate != nil {
		return validate()
	}
	return nil
}

// stripFences tolerates models that wrap output in markdown code fences
// despite instructions.
func stripFences(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	if i := strings.Index(s, "\n"); i >= 0 {
		s = s[i+1:] // drop the ```json / ```svg line
	}
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
