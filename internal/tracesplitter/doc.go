// Package tracesplitter splits a single distributed trace into multiple
// sub-traces according to a caller-supplied Predicate.
//
// # Overview
//
// A Predicate is a function that maps a Span to a string bucket key. Spans
// that map to the same key are collected into a dedicated TraceStore. Spans
// that return an empty string are silently dropped.
//
// # Built-in predicates
//
//   - ByService – buckets spans by their ServiceName field.
//   - ByTag(key) – buckets spans by the value of a specific tag key.
//
// # Usage
//
//	spl, err := tracesplitter.New(sourceStore, tracesplitter.ByService)
//	if err != nil { ... }
//
//	buckets, err := spl.Split(traceID)
//	if err != nil { ... }
//
//	for service, store := range buckets {
//		// process each sub-trace independently
//	}
package tracesplitter
