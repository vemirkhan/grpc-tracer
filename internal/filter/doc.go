// Package filter provides span and trace filtering capabilities for grpc-tracer.
//
// It allows callers to narrow down collected traces by service name,
// duration range, error presence, or a specific trace ID. Filtering
// can be applied to a raw slice of spans or directly against a TraceStore.
//
// Example usage:
//
//	matches := filter.FilterTraces(store, filter.Criteria{
//		ServiceName: "auth",
//		OnlyErrors:  true,
//	})
package filter
