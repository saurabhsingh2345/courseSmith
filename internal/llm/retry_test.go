package llm

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// newTestRetry wraps inner with instant (recorded) sleeps.
func newTestRetry(inner Provider, limiter *Limiter) (*retryingProvider, *[]time.Duration) {
	slept := &[]time.Duration{}
	r := withRetry(inner, limiter).(*retryingProvider)
	r.sleep = func(_ context.Context, d time.Duration) error {
		*slept = append(*slept, d)
		return nil
	}
	return r, slept
}

func TestRetryRecoversFromRetryableErrors(t *testing.T) {
	inner := &fakeProvider{
		name: "groq",
		fn: func(call int, req Request) (*Response, error) {
			if call <= 2 {
				return nil, &APIError{Provider: "groq", StatusCode: 429, Message: "rate limited", RetryAfter: 7 * time.Second}
			}
			return okResponse("finally"), nil
		},
	}
	r, slept := newTestRetry(inner, nil)

	resp, err := r.Complete(context.Background(), testRequest())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "finally" || inner.calls != 3 {
		t.Errorf("content = %q, calls = %d", resp.Content, inner.calls)
	}
	// Retry-After (7s) exceeds base backoff (2s, 4s) and must win both times.
	if len(*slept) != 2 || (*slept)[0] != 7*time.Second || (*slept)[1] != 7*time.Second {
		t.Errorf("slept = %v, want [7s 7s]", *slept)
	}
}

func TestRetryUsesExponentialBackoffWithoutRetryAfter(t *testing.T) {
	inner := &fakeProvider{
		name: "groq",
		fn: func(call int, req Request) (*Response, error) {
			if call <= 2 {
				return nil, &APIError{Provider: "groq", StatusCode: 503, Message: "overloaded"}
			}
			return okResponse("ok"), nil
		},
	}
	r, slept := newTestRetry(inner, nil)

	if _, err := r.Complete(context.Background(), testRequest()); err != nil {
		t.Fatal(err)
	}
	if len(*slept) != 2 || (*slept)[0] != 2*time.Second || (*slept)[1] != 4*time.Second {
		t.Errorf("slept = %v, want [2s 4s]", *slept)
	}
}

func TestRetryDoesNotRetryClientErrors(t *testing.T) {
	inner := &fakeProvider{
		name: "openai",
		fn: func(call int, req Request) (*Response, error) {
			return nil, &APIError{Provider: "openai", StatusCode: 400, Message: "bad request"}
		},
	}
	r, _ := newTestRetry(inner, nil)

	_, err := r.Complete(context.Background(), testRequest())
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != 400 {
		t.Fatalf("error = %v, want the 400 passed through", err)
	}
	if inner.calls != 1 {
		t.Errorf("client error retried: %d calls", inner.calls)
	}
}

func TestRetryGivesUpAfterMaxAttempts(t *testing.T) {
	inner := &fakeProvider{
		name: "groq",
		fn: func(call int, req Request) (*Response, error) {
			return nil, &APIError{Provider: "groq", StatusCode: 500, Message: "persistent failure"}
		},
	}
	r, _ := newTestRetry(inner, nil)

	_, err := r.Complete(context.Background(), testRequest())
	if err == nil || !strings.Contains(err.Error(), "giving up after 4 attempts") {
		t.Fatalf("error = %v, want giving-up message", err)
	}
	if !strings.Contains(err.Error(), "persistent failure") {
		t.Errorf("error %q does not preserve the underlying cause", err)
	}
	if inner.calls != defaultMaxAttempts {
		t.Errorf("calls = %d, want %d", inner.calls, defaultMaxAttempts)
	}
}

func TestRetryRetriesTransportErrors(t *testing.T) {
	inner := &fakeProvider{
		name: "groq",
		fn: func(call int, req Request) (*Response, error) {
			if call == 1 {
				return nil, errors.New("connection reset by peer")
			}
			return okResponse("ok"), nil
		},
	}
	r, _ := newTestRetry(inner, nil)

	if _, err := r.Complete(context.Background(), testRequest()); err != nil {
		t.Fatalf("transport error not retried: %v", err)
	}
	if inner.calls != 2 {
		t.Errorf("calls = %d, want 2", inner.calls)
	}
}

func TestRetryStopsOnQuotaError(t *testing.T) {
	clock := newFakeClock()
	limiter := NewLimiter("groq", LimitConfig{PerMinute: 100, PerDay: 1}, "")
	clock.install(limiter)

	inner := &fakeProvider{
		name: "groq",
		fn: func(call int, req Request) (*Response, error) {
			return nil, &APIError{Provider: "groq", StatusCode: 500, Message: "flaky"}
		},
	}
	r, _ := newTestRetry(inner, limiter)

	_, err := r.Complete(context.Background(), testRequest())
	var quota *QuotaError
	if !errors.As(err, &quota) {
		t.Fatalf("error = %v, want *QuotaError once the daily budget is gone", err)
	}
	if inner.calls != 1 {
		t.Errorf("calls = %d, want 1 (no retries past the daily budget)", inner.calls)
	}
}
