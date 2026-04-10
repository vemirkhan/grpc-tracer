// Package traceexporter provides exporters that emit trace metrics
// to external monitoring systems such as Prometheus.
package traceexporter

import (
	"fmt"
	"sync"

	"github.com/grpc-tracer/internal/storage"
)

// Counter is a minimal interface for a monotonically increasing counter.
type Counter interface {
	Inc(labels ...string)
}

// Histogram is a minimal interface for recording duration observations.
type Histogram interface {
	Observe(value float64, labels ...string)
}

// Metrics holds the counters and histograms used by PrometheusExporter.
type Metrics struct {
	SpansTotal   Counter
	ErrorsTotal  Counter
	DurationMS   Histogram
}

// PrometheusExporter reads spans from a TraceStore and emits
// metrics through the provided Metrics sinks.
type PrometheusExporter struct {
	mu      sync.Mutex
	store   *storage.TraceStore
	metrics Metrics
	seen    map[string]struct{}
}

// NewPrometheusExporter creates a PrometheusExporter backed by store.
func NewPrometheusExporter(store *storage.TraceStore, m Metrics) *PrometheusExporter {
	return &PrometheusExporter{
		store:   store,
		metrics: m,
		seen:    make(map[string]struct{}),
	}
}

// Flush iterates all traces in the store and emits metrics for any
// spans that have not yet been exported. It is safe to call concurrently.
func (p *PrometheusExporter) Flush() error {
	traces := p.store.GetAllTraces()

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, spans := range traces {
		for _, span := range spans {
			key := fmt.Sprintf("%s:%s", span.TraceID, span.SpanID)
			if _, ok := p.seen[key]; ok {
				continue
			}
			p.seen[key] = struct{}{}

			p.metrics.SpansTotal.Inc(span.ServiceName)
			if span.Error != "" {
				p.metrics.ErrorsTotal.Inc(span.ServiceName)
			}
			duration := span.EndTime.Sub(span.StartTime).Seconds() * 1000
			p.metrics.DurationMS.Observe(duration, span.ServiceName)
		}
	}
	return nil
}

// Reset clears the set of already-exported span keys, causing the
// next Flush to re-emit all spans currently in the store.
func (p *PrometheusExporter) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.seen = make(map[string]struct{})
}
