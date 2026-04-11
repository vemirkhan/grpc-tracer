// Package tracewatcher provides real-time watch functionality for trace events,
// notifying registered subscribers whenever a new span is added to a trace.
package tracewatcher

import (
	"sync"

	"github.com/example/grpc-tracer/internal/storage"
)

// EventKind describes the type of trace event.
type EventKind string

const (
	EventSpanAdded EventKind = "span_added"
)

// Event represents a single trace event emitted by the watcher.
type Event struct {
	Kind    EventKind
	TraceID string
	Span    storage.Span
}

// Handler is a callback invoked when an event occurs.
type Handler func(Event)

// Watcher watches a TraceStore and notifies subscribers of new spans.
type Watcher struct {
	mu          sync.RWMutex
	subscribers []Handler
}

// New creates a new Watcher.
func New() *Watcher {
	return &Watcher{}
}

// Subscribe registers a handler that will be called on every trace event.
// Returns an unsubscribe function.
func (w *Watcher) Subscribe(h Handler) func() {
	w.mu.Lock()
	idx := len(w.subscribers)
	w.subscribers = append(w.subscribers, h)
	w.mu.Unlock()

	return func() {
		w.mu.Lock()
		defer w.mu.Unlock()
		w.subscribers[idx] = nil
	}
}

// Notify dispatches an event to all active subscribers.
func (w *Watcher) Notify(e Event) {
	w.mu.RLock()
	handlers := make([]Handler, len(w.subscribers))
	copy(handlers, w.subscribers)
	w.mu.RUnlock()

	for _, h := range handlers {
		if h != nil {
			h(e)
		}
	}
}

// WatchStore wraps an AddSpan call, emitting an event after each successful add.
func (w *Watcher) WatchStore(store *storage.TraceStore, traceID string, span storage.Span) {
	store.AddSpan(traceID, span)
	w.Notify(Event{
		Kind:    EventSpanAdded,
		TraceID: traceID,
		Span:    span,
	})
}
