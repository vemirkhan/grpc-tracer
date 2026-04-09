// Package circuitbreaker provides a simple circuit breaker for gRPC client calls.
// It transitions between Closed, Open, and HalfOpen states based on failure thresholds.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // blocking requests
	StateHalfOpen              // testing if service recovered
)

// ErrCircuitOpen is returned when the circuit is open and requests are blocked.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreaker tracks failures and opens the circuit when a threshold is exceeded.
type CircuitBreaker struct {
	mu            sync.Mutex
	state         State
	failures      int
	maxFailures   int
	resetTimeout  time.Duration
	lastFailure   time.Time
}

// New creates a CircuitBreaker with the given failure threshold and reset timeout.
// Defaults are applied if zero values are provided.
func New(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	if maxFailures <= 0 {
		maxFailures = 5
	}
	if resetTimeout <= 0 {
		resetTimeout = 10 * time.Second
	}
	return &CircuitBreaker{
		state:        StateClosed,
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
	}
}

// Allow reports whether a request should be allowed through.
// It transitions from Open to HalfOpen after the reset timeout.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return nil
	case StateOpen:
		if time.Since(cb.lastFailure) >= cb.resetTimeout {
			cb.state = StateHalfOpen
			return nil
		}
		return ErrCircuitOpen
	case StateHalfOpen:
		return nil
	}
	return nil
}

// RecordSuccess resets the circuit breaker to Closed on a successful call.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = StateClosed
}

// RecordFailure increments the failure count and opens the circuit if threshold is reached.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
	}
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) CurrentState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
