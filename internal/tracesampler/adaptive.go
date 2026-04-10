// Package tracesampler provides adaptive sampling that adjusts the sampling
// rate based on observed error rates and request throughput.
package tracesampler

import (
	"sync"
	"time"
)

// Config holds configuration for the adaptive sampler.
type Config struct {
	// BaseRate is the default sampling rate (0.0–1.0).
	BaseRate float64
	// ErrorRate above which the sampler boosts to 1.0.
	ErrorThreshold float64
	// Window is the duration over which error rate is computed.
	Window time.Duration
}

func (c *Config) applyDefaults() {
	if c.BaseRate <= 0 || c.BaseRate > 1 {
		c.BaseRate = 0.1
	}
	if c.ErrorThreshold <= 0 || c.ErrorThreshold > 1 {
		c.ErrorThreshold = 0.2
	}
	if c.Window <= 0 {
		c.Window = 30 * time.Second
	}
}

// AdaptiveSampler adjusts its sampling rate based on recent error rate.
type AdaptiveSampler struct {
	cfg     Config
	mu      sync.Mutex
	total   int64
	errors  int64
	window  []observation
}

type observation struct {
	at    time.Time
	isErr bool
}

// New creates a new AdaptiveSampler with the given config.
func New(cfg Config) *AdaptiveSampler {
	cfg.applyDefaults()
	return &AdaptiveSampler{cfg: cfg}
}

// Record records a completed request. isErr should be true if the request
// resulted in an error.
func (a *AdaptiveSampler) Record(isErr bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	a.window = append(a.window, observation{at: now, isErr: isErr})
	a.evict(now)
}

// ShouldSample returns true if the current request should be sampled.
func (a *AdaptiveSampler) ShouldSample() bool {
	a.mu.Lock()
	rate := a.effectiveRate(time.Now())
	a.mu.Unlock()
	if rate >= 1.0 {
		return true
	}
	if rate <= 0.0 {
		return false
	}
	// Simple deterministic bucket: use current nanosecond mod 1000.
	return float64(time.Now().UnixNano()%1000)/1000.0 < rate
}

// ErrorRate returns the observed error rate over the configured window.
func (a *AdaptiveSampler) ErrorRate() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.errorRate(time.Now())
}

// Stats returns a snapshot of the current sampler state, including the number
// of observations in the window, the current error rate, and the effective
// sampling rate.
type Stats struct {
	WindowSize  int
	ErrorRate   float64
	EffectiveRate float64
}

// Snapshot returns a Stats snapshot of the sampler's current state.
func (a *AdaptiveSampler) Snapshot() Stats {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	a.evict(now)
	return Stats{
		WindowSize:    len(a.window),
		ErrorRate:     a.errorRate(now),
		EffectiveRate: a.effectiveRate(now),
	}
}

func (a *AdaptiveSampler) effectiveRate(now time.Time) float64 {
	a.evict(now)
	if a.errorRate(now) >= a.cfg.ErrorThreshold {
		return 1.0
	}
	return a.cfg.BaseRate
}

func (a *AdaptiveSampler) errorRate(now time.Time) float64 {
	a.evict(now)
	if len(a.window) == 0 {
		return 0.0
	}
	var errs int
	for _, o := range a.window {
		if o.isErr {
			errs++
		}
	}
	return float64(errs) / float64(len(a.window))
}

func (a *AdaptiveSampler) evict(now time.Time) {
	cutoff := now.Add(-a.cfg.Window)
	i := 0
	for i < len(a.window) && a.window[i].at.Before(cutoff) {
		i++
	}
	a.window = a.window[i:]
}
