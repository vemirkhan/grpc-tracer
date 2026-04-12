// Package tracecorrelator provides utilities for discovering and recording
// relationships between distinct traces that share common attributes.
//
// # Overview
//
// In a microservices system a single user action may spawn multiple root-level
// traces that are not connected by a parent-child span relationship. The
// correlator bridges those traces by scanning the trace store for spans that
// share the same value for a given tag key (e.g. "user-id", "request-id",
// "tenant").
//
// # Usage
//
//	c, err := tracecorrelator.New(store)
//	if err != nil { ... }
//
//	// Find all traces that share the same "user" tag value.
//	correlations := c.CorrelateByTag("user")
//
//	// Query correlations for a specific trace.
//	related := c.ForTrace("trace-abc123")
//
// # Interceptor
//
// UnaryServerInterceptor wraps a gRPC handler and automatically triggers
// tag-based correlation after each call completes, keeping the correlation
// index up to date in real time.
package tracecorrelator
