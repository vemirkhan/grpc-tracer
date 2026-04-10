// Package traceannotator provides utilities for attaching structured
// annotations (key-value metadata) to spans after they have been created.
package traceannotator

import (
	"fmt"
	"sync"

	"github.com/user/grpc-tracer/internal/storage"
)

// Annotator holds a reference to the trace store and applies annotations
// to spans identified by their span ID.
type Annotator struct {
	mu    sync.Mutex
	store *storage.TraceStore
}

// New creates a new Annotator backed by the given TraceStore.
func New(store *storage.TraceStore) *Annotator {
	return &Annotator{store: store}
}

// Annotate adds or updates a key-value annotation on the span identified by
// traceID and spanID. Returns an error if the trace or span is not found.
func (a *Annotator) Annotate(traceID, spanID, key, value string) error {
	if key == "" {
		return fmt.Errorf("traceannotator: annotation key must not be empty")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	spans := a.store.GetTrace(traceID)
	if len(spans) == 0 {
		return fmt.Errorf("traceannotator: trace %q not found", traceID)
	}

	for i := range spans {
		if spans[i].SpanID == spanID {
			if spans[i].Tags == nil {
				spans[i].Tags = make(map[string]string)
			}
			spans[i].Tags[key] = value
			a.store.ReplaceSpan(traceID, spans[i])
			return nil
		}
	}

	return fmt.Errorf("traceannotator: span %q not found in trace %q", spanID, traceID)
}

// AnnotateAll adds a key-value annotation to every span within a trace.
func (a *Annotator) AnnotateAll(traceID, key, value string) error {
	if key == "" {
		return fmt.Errorf("traceannotator: annotation key must not be empty")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	spans := a.store.GetTrace(traceID)
	if len(spans) == 0 {
		return fmt.Errorf("traceannotator: trace %q not found", traceID)
	}

	for i := range spans {
		if spans[i].Tags == nil {
			spans[i].Tags = make(map[string]string)
		}
		spans[i].Tags[key] = value
		a.store.ReplaceSpan(traceID, spans[i])
	}
	return nil
}
