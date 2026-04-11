package tracethrottler

import (
	"testing"
	"time"
)

func TestNew_Defaults(t *testing.T) {
	th := New(Options{})
	if th.opts.MaxSpansPerSecond != 100 {
		t.Fatalf("expected 100, got %d", th.opts.MaxSpansPerSecond)
	}
	if th.opts.BurstSize != 20 {
		t.Fatalf("expected 20, got %d", th.opts.BurstSize)
	}
}

func TestNew_Custom(t *testing.T) {
	th := New(Options{MaxSpansPerSecond: 10, BurstSize: 5})
	if th.opts.MaxSpansPerSecond != 10 {
		t.Fatalf("expected 10, got %d", th.opts.MaxSpansPerSecond)
	}
}

func TestAllow_BurstConsumed(t *testing.T) {
	th := New(Options{MaxSpansPerSecond: 1, BurstSize: 3})
	// First 3 calls should succeed (burst).
	for i := 0; i < 3; i++ {
		if !th.Allow("svc-a") {
			t.Fatalf("call %d should be allowed", i)
		}
	}
	// 4th call exceeds burst.
	if th.Allow("svc-a") {
		t.Fatal("4th call should be throttled")
	}
}

func TestAllow_RefillOverTime(t *testing.T) {
	base := time.Now()
	th := New(Options{MaxSpansPerSecond: 10, BurstSize: 2})
	th.now = func() time.Time { return base }

	// Exhaust burst.
	th.Allow("svc")
	th.Allow("svc")
	if th.Allow("svc") {
		t.Fatal("should be throttled after burst")
	}

	// Advance time by 1 second — should refill 10 tokens (capped at 2).
	th.now = func() time.Time { return base.Add(time.Second) }
	if !th.Allow("svc") {
		t.Fatal("should be allowed after refill")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	th := New(Options{MaxSpansPerSecond: 1, BurstSize: 1})
	if !th.Allow("a") {
		t.Fatal("key a should be allowed")
	}
	if !th.Allow("b") {
		t.Fatal("key b should be allowed independently")
	}
	// Both now exhausted.
	if th.Allow("a") {
		t.Fatal("key a should be throttled")
	}
	if th.Allow("b") {
		t.Fatal("key b should be throttled")
	}
}

func TestReset_ClearsBuckets(t *testing.T) {
	th := New(Options{MaxSpansPerSecond: 1, BurstSize: 1})
	th.Allow("svc")
	if th.Allow("svc") {
		t.Fatal("should be throttled before reset")
	}
	th.Reset()
	if !th.Allow("svc") {
		t.Fatal("should be allowed after reset")
	}
}
