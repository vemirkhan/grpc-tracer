// Package traceparenthook provides lifecycle hooks that fire before and after
// a span is committed to a TraceStore.
//
// # Overview
//
// A Hook holds two ordered lists of HookFn callbacks — pre-store and
// post-store. Callers register functions via RegisterPre and RegisterPost,
// then either invoke RunPre/RunPost directly or use the convenience method
// Wrap which sequences the full pre → store → post lifecycle.
//
// # Usage
//
//	h := traceparenthook.New()
//	h.RegisterPre(func(s storage.Span) {
//		log.Printf("about to store span %s", s.SpanID)
//	})
//	h.RegisterPost(func(s storage.Span) {
//		metrics.Inc("spans_stored")
//	})
//
//	if err := h.Wrap(store, span); err != nil {
//		log.Printf("store error: %v", err)
//	}
package traceparenthook
