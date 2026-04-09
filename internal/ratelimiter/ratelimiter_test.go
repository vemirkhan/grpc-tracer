package ratelimiter

import (
	"testing"
	"time"
)

func TestNew_Defaults(t *testing.T) {
	rl := New(0)
	if rl.Rate() != 100 {
		t.Fatalf("expected default rate 100, got %v", rl.Rate())
	}
}

func TestNew_CustomRate(t *testing.T) {
	rl := New(50)
	if rl.Rate() != 50 {
		t.Fatalf("expected rate 50, got %v", rl.Rate())
	}
}

func TestAllow_BucketFull(t *testing.T) {
	rl := New(10)
	// Bucket starts full; first 10 calls should be allowed.
	for i := 0; i < 10; i++ {
		if !rl.Allow() {
			t.Fatalf("expected Allow()=true on call %d", i+1)
		}
	}
}

func TestAllow_BucketExhausted(t *testing.T) {
	rl := New(5)
	for i := 0; i < 5; i++ {
		rl.Allow()
	}
	if rl.Allow() {
		t.Fatal("expected Allow()=false when bucket is exhausted")
	}
}

func TestAllow_RefillOverTime(t *testing.T) {
	fixed := time.Now()
	rl := New(10)
	rl.now = func() time.Time { return fixed }

	// Drain the bucket.
	for i := 0; i < 10; i++ {
		rl.Allow()
	}
	if rl.Allow() {
		t.Fatal("expected bucket to be empty")
	}

	// Advance clock by 1 second — should refill 10 tokens.
	rl.now = func() time.Time { return fixed.Add(1 * time.Second) }

	for i := 0; i < 10; i++ {
		if !rl.Allow() {
			t.Fatalf("expected Allow()=true after refill, call %d", i+1)
		}
	}
}

func TestAllow_CapAtMax(t *testing.T) {
	fixed := time.Now()
	rl := New(5)
	rl.now = func() time.Time { return fixed }

	// Drain the bucket.
	for i := 0; i < 5; i++ {
		rl.Allow()
	}

	// Advance by 10 seconds — would add 50 tokens but cap is 5.
	rl.now = func() time.Time { return fixed.Add(10 * time.Second) }

	allowed := 0
	for i := 0; i < 10; i++ {
		if rl.Allow() {
			allowed++
		}
	}
	if allowed != 5 {
		t.Fatalf("expected 5 allowed after cap refill, got %d", allowed)
	}
}
