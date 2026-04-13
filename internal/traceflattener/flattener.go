// Package traceflattener converts a hierarchical trace tree into a flat,
// ordered list of spans sorted by start time, suitable for tabular display
// or sequential processing.
package traceflattener

import (
	"errors"
	"sort"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// Options controls flattener behaviour.
type Options struct {
	// IncludeErrorOnly restricts output to spans that have an error.
	IncludeErrorOnly bool
	// MaxSpans caps the number of spans returned (0 = unlimited).
	MaxSpans int
}

// FlatSpan is a single span with its depth in the original tree.
type FlatSpan struct {
	Span  storage.Span
	Depth int
}

// Flattener converts traces into ordered flat span lists.
type Flattener struct {
	store *storage.TraceStore
	opts  Options
}

// New creates a Flattener backed by store.
func New(store *storage.TraceStore, opts Options) (*Flattener, error) {
	if store == nil {
		return nil, errors.New("traceflattener: store must not be nil")
	}
	if opts.MaxSpans < 0 {
		opts.MaxSpans = 0
	}
	return &Flattener{store: store, opts: opts}, nil
}

// Flatten returns the spans of traceID as a flat, start-time-ordered slice.
func (f *Flattener) Flatten(traceID string) ([]FlatSpan, error) {
	spans, ok := f.store.GetTrace(traceID)
	if !ok {
		return nil, errors.New("traceflattener: trace not found: " + traceID)
	}

	// Build parent→children index.
	children := make(map[string][]storage.Span)
	var roots []storage.Span
	for _, s := range spans {
		if s.ParentID == "" {
			roots = append(roots, s)
		} else {
			children[s.ParentID] = append(children[s.ParentID], s)
		}
	}

	sortByStart := func(sl []storage.Span) {
		sort.Slice(sl, func(i, j int) bool {
			return sl[i].StartTime.Before(sl[j].StartTime)
		})
	}
	sortByStart(roots)

	var result []FlatSpan
	var walk func(s storage.Span, depth int)
	walk = func(s storage.Span, depth int) {
		if f.opts.MaxSpans > 0 && len(result) >= f.opts.MaxSpans {
			return
		}
		if !f.opts.IncludeErrorOnly || s.Error {
			result = append(result, FlatSpan{Span: s, Depth: depth})
		}
		kids := children[s.SpanID]
		sortByStart(kids)
		for _, kid := range kids {
			walk(kid, depth+1)
		}
	}
	for _, r := range roots {
		walk(r, 0)
	}
	return result, nil
}

// FlattenAll flattens every trace in the store and returns a combined slice
// ordered by each span's start time across all traces.
func (f *Flattener) FlattenAll() []FlatSpan {
	all := f.store.GetAllTraces()
	var combined []FlatSpan
	for traceID := range all {
		fs, err := f.Flatten(traceID)
		if err == nil {
			combined = append(combined, fs...)
		}
	}
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Span.StartTime.Before(combined[j].Span.StartTime)
	})
	if f.opts.MaxSpans > 0 && len(combined) > f.opts.MaxSpans {
		combined = combined[:f.opts.MaxSpans]
	}
	_ = time.Now // keep import if StartTime is time.Time
	return combined
}
