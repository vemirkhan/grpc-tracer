// Package tracepruner implements a retention-policy engine for grpc-tracer
// trace stores.
//
// A Pruner is constructed with a *storage.TraceStore and an Options struct
// that controls two independent policies:
//
//   - MaxAge: spans whose StartTime is older than this duration are
//     discarded on the next Prune call.
//
//   - MaxSpansPerTrace: when a single trace accumulates more spans than
//     this limit the oldest spans are dropped, keeping only the most
//     recent MaxSpansPerTrace entries.
//
// Both limits default to sensible values (30 minutes / 500 spans) when
// the corresponding Options field is left at its zero value.
//
// Typical usage:
//
//	store := storage.NewTraceStore()
//	pruner := tracepruner.New(store, tracepruner.Options{
//	    MaxAge:           15 * time.Minute,
//	    MaxSpansPerTrace: 200,
//	})
//	// Call periodically, e.g. from a ticker goroutine.
//	n := pruner.Prune()
//	fmt.Printf("pruned %d spans\n", n)
package tracepruner
