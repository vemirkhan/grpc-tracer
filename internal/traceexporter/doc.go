// Package traceexporter provides adapters that push trace and span
// data collected by grpc-tracer into external observability backends.
//
// # PrometheusExporter
//
// PrometheusExporter reads all spans stored in a [storage.TraceStore] and
// emits three metrics via pluggable Counter / Histogram sinks:
//
//   - SpansTotal   – incremented once per span, labelled by service name.
//   - ErrorsTotal  – incremented when a span carries a non-empty Error field.
//   - DurationMS   – records the span duration in milliseconds.
//
// The exporter is idempotent: each span is exported at most once per
// lifetime of the exporter (keyed by traceID+spanID). Call Reset to
// clear the deduplication set and allow re-emission of all current spans.
//
// # Usage
//
//	exp := traceexporter.NewPrometheusExporter(store, traceexporter.Metrics{
//	    SpansTotal:  myCounter,
//	    ErrorsTotal: myErrCounter,
//	    DurationMS:  myHistogram,
//	})
//
//	// call periodically, e.g. from a ticker goroutine
//	exp.Flush()
package traceexporter
