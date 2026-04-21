// Package traceratelimiter provides per-trace-ID rate limiting to prevent
// a single trace from overwhelming the storage or processing pipeline.
package traceratelimiter

import (
	"errors"
	"sync"
	"time"
)

// ErrRateLimited is returned when a trace exceeds its allowed rate.
var ErrRateLimited = errors.New("traceratelimiter: rate limit exceeded")

// Options configures the per-trace rate limiter.
type Options struct {
	// MaxSpansPerSecond is the maximum number of spans allowed per trace per second.
	MaxSpansPerSecond int
	// BurstSize is the maximum burst of spans allowed before throttling.
	BurstSize int
}

func defaultOptions() Options {
	return Options{
		MaxSpansPerSecond: 100,
		BurstSize:         20,
	}
}

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// RateLimiter enforces per-trace-ID span ingestion limits.
type RateLimiter struct {
	mu      sync.Mutex
	opts    Options
	buckets map[string]*bucket
}

// New creates a new RateLimiter with the given options.
// Zero-value fields fall back to sensible defaults.
func New(opts Options) *RateLimiter {
	d := defaultOptions()
	if opts.MaxSpansPerSecond <= 0 {
		opts.MaxSpansPerSecond = d.MaxSpansPerSecond
	}
	if opts.BurstSize <= 0 {
		opts.BurstSize = d.BurstSize
	}
	return &RateLimiter{
		opts:    opts,
		buckets: make(map[string]*bucket),
	}
}

// Allow reports whether a span for the given traceID should be accepted.
// It returns ErrRateLimited when the trace has exceeded its quota.
func (r *RateLimiter) Allow(traceID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	b, ok := r.buckets[traceID]
	if !ok {
		b = &bucket{
			tokens:   float64(r.opts.BurstSize),
			lastSeen: now,
		}
		r.buckets[traceID] = b
	}

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens += elapsed * float64(r.opts.MaxSpansPerSecond)
	if b.tokens > float64(r.opts.BurstSize) {
		b.tokens = float64(r.opts.BurstSize)
	}
	b.lastSeen = now

	if b.tokens < 1 {
		return ErrRateLimited
	}
	b.tokens--
	return nil
}
