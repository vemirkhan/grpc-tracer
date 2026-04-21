// Package tracepartitioner provides utilities for splitting a trace
// store into named sub-stores (partitions) based on a user-supplied
// key function.
//
// # Overview
//
// A Partitioner reads all spans from a source TraceStore and routes
// each span into a dedicated sub-store whose name is derived by
// applying a KeyFunc to the span. Built-in key functions are provided
// for the most common partitioning strategies:
//
//	// Partition by service name
//	p, _ := tracepartitioner.New(src, tracepartitioner.ByService)
//	parts := p.Partition()
//
//	// Partition by RPC method
//	p, _ := tracepartitioner.New(src, tracepartitioner.ByMethod)
//	parts := p.Partition()
//
// # gRPC Integration
//
// UnaryServerInterceptor wraps a gRPC handler and routes the span
// stored in the request context into the matching partition store
// after the handler returns.
package tracepartitioner
