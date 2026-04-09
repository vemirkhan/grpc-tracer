// Package propagator provides utilities for injecting and extracting
// distributed trace context (trace ID, span ID, parent span ID) into and
// from gRPC metadata.
//
// Typical usage on the client side:
//
//	ctx = propagator.Inject(ctx, propagator.TraceContext{
//		TraceID: "abc",
//		SpanID:  "def",
//	})
//
// Typical usage on the server side (inside an interceptor):
//
//	tc, ok := propagator.Extract(ctx)
//	if ok {
//		// continue existing trace
//	}
package propagator
