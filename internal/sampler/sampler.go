// Package sampler provides trace sampling strategies for grpc-tracer.
// It allows controlling which traces are recorded based on configurable policies.
package sampler

import (
	"math/rand"
	"sync"
)

// Strategy defines the sampling strategy type.
type Strategy string

const (
	// AlwaysSample records every trace.
	AlwaysSample Strategy = "always"
	// NeverSample discards every trace.
	NeverSample Strategy = "never"
	// ProbabilisticSample records traces based on a probability rate.
	ProbabilisticSample Strategy = "probabilistic"
)

// Sampler decides whether a given trace should be recorded.
type Sampler struct {
	strategy Strategy
	rate     float64
	mu       sync.Mutex
	rng      *rand.Rand
}

// New creates a new Sampler with the given strategy.
// For ProbabilisticSample, rate must be in [0.0, 1.0].
func New(strategy Strategy, rate float64) *Sampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &Sampler{
		strategy: strategy,
		rate:     rate,
		rng:      rand.New(rand.NewSource(42)),
	}
}

// ShouldSample returns true if the trace identified by traceID should be recorded.
func (s *Sampler) ShouldSample(traceID string) bool {
	switch s.strategy {
	case AlwaysSample:
		return true
	case NeverSample:
		return false
	case ProbabilisticSample:
		s.mu.Lock()
		v := s.rng.Float64()
		s.mu.Unlock()
		return v < s.rate
	}
	return true
}

// Strategy returns the current sampling strategy.
func (s *Sampler) Strategy() Strategy {
	return s.strategy
}

// Rate returns the configured sampling rate.
func (s *Sampler) Rate() float64 {
	return s.rate
}
