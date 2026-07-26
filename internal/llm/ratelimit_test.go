package llm

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

// fakeClock drives a Limiter deterministically: sleeps advance time instantly.
type fakeClock struct {
	now    time.Time
	slept  []time.Duration
	sleeps int
}

func newFakeClock() *fakeClock {
	return &fakeClock{now: time.Date(2026, 7, 17, 9, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) install(l *Limiter) {
	l.now = func() time.Time { return c.now }
	l.sleep = func(_ context.Context, d time.Duration) error {
		c.slept = append(c.slept, d)
		c.sleeps++
		c.now = c.now.Add(d)
		return nil
	}
}

func TestLimiterBurstThenThrottle(t *testing.T) {
	clock := newFakeClock()
	l := NewLimiter("groq", LimitConfig{PerMinute: 3, PerDay: 100}, "")
	clock.install(l)
	ctx := context.Background()

	// Full bucket: 3 requests pass without sleeping.
	for i := range 3 {
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("Wait #%d: %v", i+1, err)
		}
	}
	if clock.sleeps != 0 {
		t.Fatalf("burst within capacity slept %d times", clock.sleeps)
	}

	// Bucket empty: the 4th request must wait for one token to refill
	// (3/min → 20s per token).
	if err := l.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	if clock.sleeps == 0 {
		t.Fatal("4th request did not sleep")
	}
	total := time.Duration(0)
	for _, d := range clock.slept {
		total += d
	}
	if total < 19*time.Second || total > 21*time.Second {
		t.Errorf("slept %v total, want ~20s", total)
	}
}

func TestLimiterRefillCapsAtCapacity(t *testing.T) {
	clock := newFakeClock()
	l := NewLimiter("groq", LimitConfig{PerMinute: 2, PerDay: 100}, "")
	clock.install(l)
	ctx := context.Background()

	if err := l.Wait(ctx); err != nil { // establish LastRefill
		t.Fatal(err)
	}
	clock.now = clock.now.Add(time.Hour) // long idle must not over-fill

	for range 2 { // capacity is 2, one was restored plus one left
		if err := l.Wait(ctx); err != nil {
			t.Fatal(err)
		}
	}
	before := clock.sleeps
	if err := l.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	if clock.sleeps == before {
		t.Error("bucket exceeded capacity after long idle")
	}
}

func TestLimiterDailyQuota(t *testing.T) {
	clock := newFakeClock()
	l := NewLimiter("groq", LimitConfig{PerMinute: 100, PerDay: 2}, "")
	clock.install(l)
	ctx := context.Background()

	for i := range 2 {
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("Wait #%d: %v", i+1, err)
		}
	}

	err := l.Wait(ctx)
	var quota *QuotaError
	if !errors.As(err, &quota) {
		t.Fatalf("Wait() = %v, want *QuotaError", err)
	}
	if quota.Provider != "groq" || quota.Limit != 2 {
		t.Errorf("QuotaError = %+v", quota)
	}
	wantReset := clock.now.Add(24 * time.Hour)
	if !quota.ResetAt.Equal(wantReset) {
		t.Errorf("ResetAt = %v, want %v", quota.ResetAt, wantReset)
	}

	// After the 24h window rolls, requests flow again.
	clock.now = clock.now.Add(25 * time.Hour)
	if err := l.Wait(ctx); err != nil {
		t.Errorf("Wait() after window rollover: %v", err)
	}
}

func TestLimiterPersistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ratelimit", "groq.json")
	clock := newFakeClock()
	cfg := LimitConfig{PerMinute: 100, PerDay: 5}

	l1 := NewLimiter("groq", cfg, path)
	clock.install(l1)
	for range 3 {
		if err := l1.Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	// A new process (same clock) must remember 3 of 5 daily requests used.
	l2 := NewLimiter("groq", cfg, path)
	clock.install(l2)
	for i := range 2 {
		if err := l2.Wait(context.Background()); err != nil {
			t.Fatalf("Wait #%d after reload: %v", i+1, err)
		}
	}
	err := l2.Wait(context.Background())
	var quota *QuotaError
	if !errors.As(err, &quota) {
		t.Fatalf("daily budget not enforced across restart: %v", err)
	}
}

func TestLimiterContextCancelled(t *testing.T) {
	l := NewLimiter("groq", LimitConfig{PerMinute: 1, PerDay: 100}, "")
	ctx := context.Background()
	if err := l.Wait(ctx); err != nil { // drain the single token
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := l.Wait(cancelled); !errors.Is(err, context.Canceled) {
		t.Errorf("Wait() with cancelled ctx = %v, want context.Canceled", err)
	}
}
