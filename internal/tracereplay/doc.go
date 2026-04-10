// Package tracereplay replays previously recorded gRPC traces into a
// destination TraceStore for offline debugging, regression testing, and
// performance benchmarking.
//
// Basic usage:
//
//	r, err := tracereplay.New(srcStore, dstStore, tracereplay.Options{
//		SpeedFactor: 2.0, // replay at 2× real-time speed
//		MaxSpans:    100,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	n, err := r.ReplayTrace("abc123")
//
// SpeedFactor of 0 disables inter-span delays entirely, which is useful
// in unit tests and CI pipelines.
package tracereplay
