// Package tracededuplicator provides span-level deduplication for gRPC trace
// pipelines.
//
// In distributed systems the same span can be emitted more than once due to
// retries, at-least-once delivery semantics, or fan-out collectors. The
// Deduplicator sits in front of a storage.TraceStore and silently discards any
// span whose SpanID has already been recorded, ensuring that each logical unit
// of work appears exactly once in the trace store.
//
// Usage:
//
//	store := storage.NewTraceStore()
//	dd, err := tracededuplicator.New(store)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// accepted == true  → span was new and forwarded to store
//	// accepted == false → span was a duplicate and dropped
//	accepted := dd.AddSpan(span)
//
// The internal seen-set can be cleared at any time via Reset(), which is useful
// when rotating trace windows or during testing.
package tracededuplicator
