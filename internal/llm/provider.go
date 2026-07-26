// Package llm provides rate-limited, disk-cached access to chat-completion
// providers (Groq, OpenAI). The stack, outermost first:
//
//	cache → retry → rate limit → HTTP
//
// so cache hits cost zero API calls and zero rate-limit tokens, and every
// real request is throttled and retried with backoff.
package llm

import (
	"context"
	"fmt"
	"time"
)

// Message roles, matching the OpenAI chat API.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

// Message is one turn of a chat-completion conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request is a provider-agnostic completion request. Model is the bare model
// name (no provider prefix); the router fills it in from config.
type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	// JSONMode asks the provider to emit a single valid JSON object
	// (response_format: json_object).
	JSONMode bool `json:"json_mode,omitempty"`
	// Images are base64-encoded PNGs attached to the final user message
	// (vision models only). They participate in the cache key like any
	// other request field.
	Images []string `json:"images,omitempty"`
}

// Validate checks a request before it is sent (or used as a cache key).
func (r Request) Validate() error {
	if r.Model == "" {
		return fmt.Errorf("llm request has no model")
	}
	if len(r.Messages) == 0 {
		return fmt.Errorf("llm request has no messages")
	}
	for i, m := range r.Messages {
		if m.Role == "" || m.Content == "" {
			return fmt.Errorf("llm request message %d has empty role or content", i)
		}
	}
	return nil
}

// Usage is the token accounting reported by the provider.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Response is a completed request.
type Response struct {
	Content string `json:"content"`
	Model   string `json:"model"` // model the provider actually served
	Usage   Usage  `json:"usage"`
	// FromCache is true when the response was served from the disk cache
	// without touching the provider. Never persisted as true.
	FromCache bool `json:"-"`
}

// Provider completes chat requests.
type Provider interface {
	// Name identifies the provider ("groq", "openai") for cache keys,
	// rate-limit buckets, and error messages.
	Name() string
	Complete(ctx context.Context, req Request) (*Response, error)
}

// APIError is a non-2xx response from a provider's API.
type APIError struct {
	Provider   string
	StatusCode int
	Message    string
	// RetryAfter is the provider-requested backoff (from the Retry-After
	// header), or 0 if none was given.
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s API error (HTTP %d): %s", e.Provider, e.StatusCode, e.Message)
}

// Retryable reports whether the request may succeed if retried.
func (e *APIError) Retryable() bool {
	return e.StatusCode == 429 || e.StatusCode >= 500
}

// QuotaError means a provider's daily request budget is exhausted. Callers
// should stop cleanly: completed work is cached and state.json records
// finished stages, so re-running after ResetAt resumes where it stopped.
type QuotaError struct {
	Provider string
	Limit    int
	ResetAt  time.Time
}

func (e *QuotaError) Error() string {
	return fmt.Sprintf(
		"%s daily request budget (%d) exhausted; resets at %s — re-run then, completed work is cached and will be skipped",
		e.Provider, e.Limit, e.ResetAt.Local().Format(time.RFC1123),
	)
}
