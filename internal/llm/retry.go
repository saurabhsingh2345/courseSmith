package llm

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// retryingProvider wraps a provider with a rate limiter and retry-with-backoff.
// Each attempt (including retries) acquires a rate-limit token first, so
// retries can never exceed the provider's request budget.
type retryingProvider struct {
	inner       Provider
	limiter     *Limiter // may be nil (no throttling)
	maxAttempts int
	baseBackoff time.Duration

	sleep func(ctx context.Context, d time.Duration) error
}

const (
	defaultMaxAttempts = 4
	defaultBaseBackoff = 2 * time.Second
	// maxRetryAfter caps how long a server-sent Retry-After hint is
	// honored: a bad hint must never silently stall a pipeline run.
	maxRetryAfter = 2 * time.Minute
)

// withRetry wraps inner with throttling and retries.
func withRetry(inner Provider, limiter *Limiter) Provider {
	return &retryingProvider{
		inner:       inner,
		limiter:     limiter,
		maxAttempts: defaultMaxAttempts,
		baseBackoff: defaultBaseBackoff,
		sleep:       sleepCtx,
	}
}

func (r *retryingProvider) Name() string { return r.inner.Name() }

func (r *retryingProvider) Complete(ctx context.Context, req Request) (*Response, error) {
	var lastErr error
	for attempt := 0; attempt < r.maxAttempts; attempt++ {
		if attempt > 0 {
			delay := r.baseBackoff << (attempt - 1)
			if apiErr, ok := errors.AsType[*APIError](lastErr); ok && apiErr.RetryAfter > delay {
				delay = min(apiErr.RetryAfter, maxRetryAfter)
			}
			if err := r.sleep(ctx, delay); err != nil {
				return nil, err
			}
		}
		if r.limiter != nil {
			if err := r.limiter.Wait(ctx); err != nil {
				return nil, err // quota exhausted or ctx done — retrying won't help
			}
		}
		resp, err := r.inner.Complete(ctx, req)
		if err == nil {
			return resp, nil
		}
		if !isRetryable(ctx, err) {
			return nil, err
		}
		lastErr = err
	}
	return nil, fmt.Errorf("%s: giving up after %d attempts: %w", r.inner.Name(), r.maxAttempts, lastErr)
}

// isRetryable classifies an error from a provider attempt. API errors follow
// their status code; context ends are terminal; anything else (transport
// errors, timeouts) is worth retrying.
func isRetryable(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	}
	if apiErr, ok := errors.AsType[*APIError](err); ok {
		return apiErr.Retryable()
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	return true
}
