// Package circuitbreaker implements a thread-safe circuit breaker pattern
// for use with gRPC client interceptors.
//
// The circuit breaker transitions through three states:
//
//   - Closed: normal operation; all requests pass through.
//   - Open: the failure threshold has been exceeded; requests are rejected
//     immediately with ErrCircuitOpen.
//   - HalfOpen: the reset timeout has elapsed; one probe request is allowed
//     through to test whether the downstream service has recovered.
//
// Usage:
//
//	cb := circuitbreaker.New(5, 10*time.Second)
//	if err := cb.Allow(); err != nil {
//		return err // circuit is open
//	}
//	if err := invoke(); err != nil {
//		cb.RecordFailure()
//		return err
//	}
//	cb.RecordSuccess()
package circuitbreaker
