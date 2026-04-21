// Package tracehook provides a lightweight lifecycle-hook mechanism for
// gRPC trace spans.
//
// # Overview
//
// A Hook holds two ordered lists of callbacks:
//
//   - Pre-hooks  – invoked with the span BEFORE it is written to storage.
//   - Post-hooks – invoked with the span AFTER it has been written to storage.
//
// # Usage
//
//	h := tracehook.New()
//
//	// Log every span before it lands in the store.
//	_ = h.RegisterPre(func(s storage.Span) {
//	    log.Printf("storing span %s", s.SpanID)
//	})
//
//	// Emit a metric once the span is safely persisted.
//	_ = h.RegisterPost(func(s storage.Span) {
//	    metrics.Inc("spans_stored")
//	})
//
//	// Store a span through the hook.
//	_ = h.AddSpan(store, span)
//
// The package also exposes UnaryServerInterceptor for drop-in use with the
// standard gRPC middleware chain.
package tracehook
