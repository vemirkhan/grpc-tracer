// Package tracescorer computes a numeric quality score (0.0–1.0) for a trace
// stored in a TraceStore.
//
// Scoring criteria
//
// The scorer applies three independent penalty rules:
//
//   - High latency  – total wall-clock duration of the trace exceeds MaxDuration.
//   - High span count – total number of spans exceeds MaxSpans.
//   - Errors present – one or more spans carry a non-empty Error field; the
//     penalty magnitude scales with the fraction of erroneous spans and the
//     configurable ErrorWeight.
//
// All penalties are subtracted from an initial score of 1.0. The result is
// clamped to [0.0, 1.0].
//
// Usage
//
//	scorer, err := tracescorer.New(store, tracescorer.Options{
//		MaxDuration: 2 * time.Second,
//		MaxSpans:    50,
//		ErrorWeight: 0.3,
//	})
//	if err != nil { ... }
//
//	score, err := scorer.ScoreTrace(traceID)
package tracescorer
