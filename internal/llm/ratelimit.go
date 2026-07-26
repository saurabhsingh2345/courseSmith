package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LimitConfig bounds request volume for one provider.
type LimitConfig struct {
	// PerMinute is the sustained request rate; it is also the burst
	// capacity of the token bucket.
	PerMinute int
	// PerDay caps total requests per rolling 24h window. Exceeding it makes
	// Wait return a *QuotaError instead of blocking (the wait would be hours).
	PerDay int
}

// DefaultLimits stays safely under each provider's free-tier ceilings
// (Groq free tier: 30 req/min, ~1,000 req/day for llama-3.3-70b).
var DefaultLimits = map[string]LimitConfig{
	"groq":   {PerMinute: 28, PerDay: 950},
	"openai": {PerMinute: 60, PerDay: 4000},
}

// limiterState is the persisted part of a Limiter, so restarts cannot blow
// the daily budget or the burst bucket.
type limiterState struct {
	Tokens     float64   `json:"tokens"`
	LastRefill time.Time `json:"last_refill"`
	DayStart   time.Time `json:"day_start"`
	DayUsed    int       `json:"day_used"`
}

// Limiter is a token-bucket rate limiter with a daily budget.
// Wait blocks until a request may be sent (or the context is done).
type Limiter struct {
	mu       sync.Mutex
	provider string
	cfg      LimitConfig
	state    limiterState
	// path persists state across process restarts; empty means memory-only.
	path string

	// Injectable for tests.
	now   func() time.Time
	sleep func(ctx context.Context, d time.Duration) error
}

// NewLimiter creates a limiter for one provider. statePath may be "" to
// disable persistence; otherwise prior state is loaded from it if present.
func NewLimiter(provider string, cfg LimitConfig, statePath string) *Limiter {
	l := &Limiter{
		provider: provider,
		cfg:      cfg,
		path:     statePath,
		now:      time.Now,
		sleep:    sleepCtx,
	}
	l.state.Tokens = float64(cfg.PerMinute) // full bucket on first ever run
	if statePath != "" {
		if data, err := os.ReadFile(statePath); err == nil {
			var s limiterState
			if err := json.Unmarshal(data, &s); err == nil && !s.LastRefill.IsZero() {
				l.state = s
			}
		}
	}
	return l
}

// Wait blocks until one request token is available and consumes it.
// It returns *QuotaError immediately when the daily budget is exhausted,
// and ctx.Err() if the context ends while waiting.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		now := l.now()
		l.refill(now)

		if l.cfg.PerDay > 0 && l.state.DayUsed >= l.cfg.PerDay {
			err := &QuotaError{
				Provider: l.provider,
				Limit:    l.cfg.PerDay,
				ResetAt:  l.state.DayStart.Add(24 * time.Hour),
			}
			l.mu.Unlock()
			return err
		}

		if l.state.Tokens >= 1 {
			l.state.Tokens--
			l.state.DayUsed++
			if l.state.DayStart.IsZero() {
				l.state.DayStart = now
			}
			l.persistLocked()
			l.mu.Unlock()
			return nil
		}

		ratePerSec := float64(l.cfg.PerMinute) / 60.0
		wait := time.Duration((1 - l.state.Tokens) / ratePerSec * float64(time.Second))
		l.mu.Unlock()

		if err := l.sleep(ctx, wait); err != nil {
			return err
		}
	}
}

// refill tops up the bucket for elapsed time and rolls the daily window.
// Callers must hold l.mu.
func (l *Limiter) refill(now time.Time) {
	if !l.state.LastRefill.IsZero() {
		elapsed := now.Sub(l.state.LastRefill).Seconds()
		if elapsed > 0 {
			l.state.Tokens += elapsed * float64(l.cfg.PerMinute) / 60.0
			if capacity := float64(l.cfg.PerMinute); l.state.Tokens > capacity {
				l.state.Tokens = capacity
			}
		}
	}
	l.state.LastRefill = now

	if !l.state.DayStart.IsZero() && now.Sub(l.state.DayStart) >= 24*time.Hour {
		l.state.DayStart = now
		l.state.DayUsed = 0
	}
}

// persistLocked best-effort saves state; a failed write must not fail the
// request, so the error is reported to stderr and dropped.
// Callers must hold l.mu.
func (l *Limiter) persistLocked() {
	if l.path == "" {
		return
	}
	data, err := json.Marshal(l.state)
	if err == nil {
		if mkErr := os.MkdirAll(filepath.Dir(l.path), 0o755); mkErr != nil {
			err = mkErr
		} else {
			err = os.WriteFile(l.path, data, 0o644)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not persist %s rate-limit state: %v\n", l.provider, err)
	}
}

// sleepCtx sleeps for d or until ctx is done, whichever comes first.
func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
