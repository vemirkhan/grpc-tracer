// Package tracethrottler implements per-key span-rate throttling for the
// grpc-tracer pipeline.
//
// A Throttler uses a token-bucket algorithm keyed by an arbitrary string
// (typically a service name or trace ID). Tokens refill at a configurable
// rate (MaxSpansPerSecond) and allow short bursts up to BurstSize.
//
// Typical usage:
//
//	th := tracethrottler.New(tracethrottler.Options{
//		MaxSpansPerSecond: 500,
//		BurstSize:         50,
//	})
//
//	if th.Allow(span.ServiceName) {
//		store.AddSpan(traceID, span)
//	}
//
// When MaxSpansPerSecond is set to 0, all spans are allowed through without
// any rate limiting. When BurstSize is set to 0, it defaults to 1, ensuring
// at least one span can always be admitted per refill interval.
//
// The Throttler is safe for concurrent use.
package tracethrottler
