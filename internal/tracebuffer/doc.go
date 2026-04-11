// Package tracebuffer implements a thread-safe, fixed-capacity ring buffer
// designed for high-throughput span collection in the grpc-tracer pipeline.
//
// # Overview
//
// A Buffer holds up to a configurable number of [storage.Span] values in
// FIFO order. When the buffer reaches capacity, the oldest span is silently
// evicted to make room for the incoming span, bounding memory consumption
// regardless of ingestion rate.
//
// # Usage
//
//	buf := tracebuffer.New(1024)
//
//	// producer
//	buf.Push(span)
//
//	// consumer — drain periodically
//	for _, s := range buf.Flush() {
//		store.AddSpan(s.TraceID, s)
//	}
//
// Buffer is safe for concurrent use by multiple goroutines.
package tracebuffer
