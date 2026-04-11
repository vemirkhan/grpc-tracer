// Package tracemerger provides utilities for combining spans from multiple
// TraceStore instances into a single destination store.
//
// # Overview
//
// In a distributed tracing setup it is common for different services or
// middleware layers to accumulate spans in their own local stores before
// flushing them to a central location. The Merger type handles this
// consolidation step:
//
//	merger, err := tracemerger.New(centralStore)
//	if err != nil { ... }
//	n, err := merger.Merge(storeA, storeB, storeC)
//
// Duplicate spans (identified by SpanID) are silently dropped so that
// replaying or retrying a merge is safe.
//
// # gRPC Integration
//
// UnaryServerInterceptor wraps a gRPC handler and automatically merges the
// span stored in the request context into the destination store after each
// call completes.
package tracemerger
