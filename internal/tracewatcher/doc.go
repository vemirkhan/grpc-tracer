// Package tracewatcher provides a publish-subscribe mechanism for trace events.
//
// A Watcher allows multiple handlers to be registered and notified in real time
// whenever a new span is recorded. This enables live dashboards, alerting, and
// streaming trace visualizations without polling the TraceStore.
//
// Basic usage:
//
//	w := tracewatcher.New()
//	unsub := w.Subscribe(func(e tracewatcher.Event) {
//		fmt.Printf("new span on trace %s\n", e.TraceID)
//	})
//	defer unsub()
//
// The UnaryServerInterceptor integrates the watcher into the gRPC middleware
// chain, automatically emitting events after each handled RPC.
package tracewatcher
