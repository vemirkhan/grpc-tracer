// Package ratelimiter provides a token-bucket rate limiter for controlling
// the volume of spans accepted by the tracer per second.
package ratelimiter

import (
	"sync"
	"time"
)

// RateLimiter controls how many spans can be recorded per second using a
// token-bucket algorithm.
type RateLimiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTick time.Time
	now      func() time.Time
}

// New creates a RateLimiter that allows up to maxPerSecond spans per second.
// If maxPerSecond is <= 0 it defaults to 100.
func New(maxPerSecond float64) *RateLimiter {
	if maxPerSecond <= 0 {
		maxPerSecond = 100
	}
	return &RateLimiter{
		tokens:   maxPerSecond,
		max:      maxPerSecond,
		rate:     maxPerSecond,
		lastTick: time.Now(),
		now:      time.Now,
	}
}

// Allow reports whether a single span should be accepted. It refills tokens
// based on elapsed time since the last call and consumes one token if
// available.
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.now()
	elapsed := now.Sub(r.lastTick).Seconds()
	r.lastTick = now

	r.tokens += elapsed * r.rate
	if r.tokens > r.max {
		r.tokens = r.max
	}

	if r.tokens < 1 {
		return false
	}
	r.tokens--
	return true
}

// Rate returns the configured token replenishment rate (spans per second).
func (r *RateLimiter) Rate() float64 {
	return r.rate
}
