// Package tracepruner provides utilities for removing old or excess spans
// from a trace store based on configurable retention policies.
package tracepruner

import (
	"time"

	"github.com/example/grpc-tracer/internal/storage"
)

// Options configures the pruning behaviour.
type Options struct {
	// MaxAge is the maximum age of a span before it is pruned.
	// Spans older than this duration are removed. Zero means no age limit.
	MaxAge time.Duration

	// MaxSpansPerTrace caps the number of spans retained per trace.
	// When exceeded the oldest spans are dropped. Zero means no limit.
	MaxSpansPerTrace int
}

func defaultOptions() Options {
	return Options{
		MaxAge:           30 * time.Minute,
		MaxSpansPerTrace: 500,
	}
}

// Pruner removes stale or excess spans from a TraceStore.
type Pruner struct {
	store *storage.TraceStore
	opts  Options
}

// New creates a Pruner with the given store and options.
// Any zero-value option field falls back to its default.
func New(store *storage.TraceStore, opts Options) *Pruner {
	defaults := defaultOptions()
	if opts.MaxAge == 0 {
		opts.MaxAge = defaults.MaxAge
	}
	if opts.MaxSpansPerTrace == 0 {
		opts.MaxSpansPerTrace = defaults.MaxSpansPerTrace
	}
	return &Pruner{store: store, opts: opts}
}

// Prune iterates over all traces and removes spans that violate the
// configured retention policy. It returns the total number of spans removed.
func (p *Pruner) Prune() int {
	cutoff := time.Now().Add(-p.opts.MaxAge)
	traces := p.store.GetAllTraces()
	removed := 0

	for _, spans := range traces {
		if len(spans) == 0 {
			continue
		}
		traceID := spans[0].TraceID

		// Filter by age.
		var kept []storage.Span
		for _, s := range spans {
			if s.StartTime.After(cutoff) {
				kept = append(kept, s)
			} else {
				removed++
			}
		}

		// Cap per-trace span count (keep the most recent).
		if p.opts.MaxSpansPerTrace > 0 && len(kept) > p.opts.MaxSpansPerTrace {
			drop := len(kept) - p.opts.MaxSpansPerTrace
			removed += drop
			kept = kept[drop:]
		}

		p.store.ReplaceSpans(traceID, kept)
	}
	return removed
}
