// Package tracethrottler provides span-rate throttling per trace or service,
// dropping spans that exceed a configured burst limit within a sliding window.
package tracethrottler

import (
	"sync"
	"time"
)

// Options configures the throttler.
type Options struct {
	// MaxSpansPerSecond is the maximum number of spans accepted per key per second.
	MaxSpansPerSecond int
	// BurstSize is the maximum instantaneous burst allowed.
	BurstSize int
}

func defaultOptions() Options {
	return Options{
		MaxSpansPerSecond: 100,
		BurstSize:         20,
	}
}

type bucket struct {
	tokens    float64
	lastRefil time.Time
}

// Throttler rate-limits spans keyed by an arbitrary string (e.g. service name or trace ID).
type Throttler struct {
	opts    Options
	mu      sync.Mutex
	buckets map[string]*bucket
	now     func() time.Time
}

// New creates a Throttler. Zero-value fields in opts are replaced with defaults.
func New(opts Options) *Throttler {
	d := defaultOptions()
	if opts.MaxSpansPerSecond <= 0 {
		opts.MaxSpansPerSecond = d.MaxSpansPerSecond
	}
	if opts.BurstSize <= 0 {
		opts.BurstSize = d.BurstSize
	}
	return &Throttler{
		opts:    opts,
		buckets: make(map[string]*bucket),
		now:     time.Now,
	}
}

// Allow returns true if the span for the given key should be accepted.
// It uses a token-bucket algorithm: tokens refill at MaxSpansPerSecond
// and are capped at BurstSize.
func (t *Throttler) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	b, ok := t.buckets[key]
	if !ok {
		b = &bucket{
			tokens:    float64(t.opts.BurstSize),
			lastRefil: now,
		}
		t.buckets[key] = b
	}

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(b.lastRefil).Seconds()
	b.tokens += elapsed * float64(t.opts.MaxSpansPerSecond)
	if b.tokens > float64(t.opts.BurstSize) {
		b.tokens = float64(t.opts.BurstSize)
	}
	b.lastRefil = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Reset clears all token buckets.
func (t *Throttler) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buckets = make(map[string]*bucket)
}
