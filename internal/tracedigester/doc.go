// Package tracedigester computes stable structural fingerprints (digests) for
// gRPC traces stored in a TraceStore.
//
// A digest is a SHA-256 hash of the ordered sequence of "service:method" pairs
// across all spans in a trace, sorted by start time. Two traces that exercise
// the same call chain in the same order will always produce the same digest,
// regardless of their trace IDs, span IDs, or timing data.
//
// # Usage
//
//	dig, err := tracedigester.New(store)
//	if err != nil { ... }
//
//	// Fingerprint a single trace.
//	fingerprint, err := dig.Compute(traceID)
//
//	// Group all traces in the store by structural shape.
//	groups, err := dig.GroupByDigest()
//	for digest, ids := range groups {
//		fmt.Printf("%s -> %v\n", digest, ids)
//	}
package tracedigester
