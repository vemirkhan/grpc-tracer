// Package traceannotator provides post-hoc annotation of spans stored in a
// TraceStore.
//
// An Annotator wraps a *storage.TraceStore and exposes two methods:
//
//   - Annotate(traceID, spanID, key, value) – add or update a single tag on
//     a specific span.
//   - AnnotateAll(traceID, key, value) – add or update a tag on every span
//     belonging to a trace.
//
// A gRPC interceptor (UnaryServerInterceptor) is also provided so that a
// fixed set of static annotations (e.g. deployment environment, region, or
// version) can be stamped onto every incoming request's span automatically.
//
// Example usage:
//
//	an := traceannotator.New(store)
//	_ = an.Annotate("trace-abc", "span-001", "env", "production")
//	_ = an.AnnotateAll("trace-abc", "region", "us-east-1")
package traceannotator
