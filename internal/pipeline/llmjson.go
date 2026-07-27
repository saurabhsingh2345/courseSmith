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
	return e.completeWithRepairRounds(ctx, pcfg, task, system, user, images, temperature, maxTokens, jsonMode, 1, accept)
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
	return e.completeJSONRounds(ctx, pcfg, task, system, user, images, temperature, maxTokens, 1, out, validate)
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
	out any,
	validate func() error,
) error {
	_, err := e.completeWithRepairRounds(ctx, pcfg, task, system, user, images, temperature, maxTokens, true, rounds,
		func(content string) (string, error) {
			if err := parseJSONStrict(content, out, validate); err != nil {
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
