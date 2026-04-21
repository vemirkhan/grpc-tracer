package traceratelimiter_test

import (
	"testing"
	"time"

	"github.com/bevzzz/grpc-tracer/internal/traceratelimiter"
)

func TestNew_Defaults(t *testing.T) {
	rl := traceratelimiter.New(traceratelimiter.Options{})
	if rl == nil {
		t.Fatal("expected non-nil RateLimiter")
	}
}

func TestNew_CustomOptions(t *testing.T) {
	rl := traceratelimiter.New(traceratelimiter.Options{
		MaxSpansPerSecond: 10,
		BurstSize:         5,
	})
	if rl == nil {
		t.Fatal("expected non-nil RateLimiter")
	}
}

func TestAllow_BurstConsumed(t *testing.T) {
	rl := traceratelimiter.New(traceratelimiter.Options{
		MaxSpansPerSecond: 1,
		BurstSize:         3,
	})

	const traceID = "trace-abc"
	for i := 0; i < 3; i++ {
		if err := rl.Allow(traceID); err != nil {
			t.Fatalf("expected Allow to succeed on call %d, got: %v", i+1, err)
		}
	}

	// 4th call should be rate-limited.
	if err := rl.Allow(traceID); err == nil {
		t.Fatal("expected ErrRateLimited after burst exhausted")
	}
}

func TestAllow_IndependentTraces(t *testing.T) {
	rl := traceratelimiter.New(traceratelimiter.Options{
		MaxSpansPerSecond: 1,
		BurstSize:         1,
	})

	if err := rl.Allow("trace-1"); err != nil {
		t.Fatalf("trace-1: unexpected error: %v", err)
	}
	// trace-2 has its own bucket and should not be affected.
	if err := rl.Allow("trace-2"); err != nil {
		t.Fatalf("trace-2: unexpected error: %v", err)
	}
}

func TestAllow_RefillOverTime(t *testing.T) {
	rl := traceratelimiter.New(traceratelimiter.Options{
		MaxSpansPerSecond: 100,
		BurstSize:         1,
	})

	const traceID = "trace-refill"
	// Consume the single burst token.
	if err := rl.Allow(traceID); err != nil {
		t.Fatalf("first Allow failed: %v", err)
	}
	// Should be exhausted now.
	if err := rl.Allow(traceID); err == nil {
		t.Fatal("expected rate limit after burst consumed")
	}
	// Wait for refill (100 spans/s → 1 token in ~10ms).
	time.Sleep(20 * time.Millisecond)
	if err := rl.Allow(traceID); err != nil {
		t.Fatalf("expected Allow after refill, got: %v", err)
	}
}

func TestAllow_NegativeOptions_UsesDefaults(t *testing.T) {
	rl := traceratelimiter.New(traceratelimiter.Options{
		MaxSpansPerSecond: -5,
		BurstSize:         -1,
	})
	// With defaults (burst=20) the first call must succeed.
	if err := rl.Allow("trace-x"); err != nil {
		t.Fatalf("unexpected error with default options: %v", err)
	}
}
