// Package traceindex implements a secondary in-memory index over trace spans
// collected by a TraceStore. It enables O(1) lookups by service name, gRPC
// method, or arbitrary tag key-value pairs without scanning every trace.
//
// # Usage
//
//	idx := traceindex.New()
//	idx.Build(store) // populate from an existing TraceStore
//
//	// Query traces that involved the "auth" service:
//	traceIDs := idx.ByService("auth")
//
//	// Query traces that hit a specific method:
//	traceIDs = idx.ByMethod("/grpc.health.v1.Health/Check")
//
//	// Query traces tagged with env=prod:
//	traceIDs = idx.ByTag("env", "prod")
//
// The index can be kept up-to-date by calling Build after new spans are added,
// or by wiring the provided UnaryServerInterceptor into the gRPC server chain.
package traceindex
