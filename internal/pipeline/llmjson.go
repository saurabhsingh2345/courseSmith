package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
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
	artifact, firstErr := accept(resp.Content)
	if firstErr == nil {
		return artifact, nil
	}

	// One correction round: show the model its own reply and the exact error.
	req.Messages = append(messages,
		llm.Message{Role: llm.RoleAssistant, Content: resp.Content},
		llm.Message{Role: llm.RoleUser, Content: fmt.Sprintf(
			"Your previous reply was rejected: %v. Respond again with only the corrected output — no fences, no commentary.",
			firstErr,
		)},
	)
	resp, err = e.Router.Complete(ctx, pcfg, task, req)
	if err != nil {
		return "", fmt.Errorf("retrying after rejected output (%v): %w", firstErr, err)
	}
	artifact, retryErr := accept(resp.Content)
	if retryErr != nil {
		return "", fmt.Errorf("%s response invalid after retry: %w (first attempt: %v)", task, retryErr, firstErr)
	}
	return artifact, nil
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
	_, err := e.completeWithRepairImages(ctx, pcfg, task, system, user, images, temperature, maxTokens, true,
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
