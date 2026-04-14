// Package traceparenthook provides a mechanism for registering and invoking
// lifecycle hooks around span processing — before and after a span is stored.
package traceparenthook

import (
	"sync"

	"github.com/example/grpc-tracer/internal/storage"
)

// HookFn is a function invoked with a span at a lifecycle point.
type HookFn func(span storage.Span)

// Hook holds pre- and post-store callbacks for spans.
type Hook struct {
	mu      sync.RWMutex
	preFns  []HookFn
	postFns []HookFn
}

// New returns an initialised Hook with no registered functions.
func New() *Hook {
	return &Hook{}
}

// RegisterPre adds a function that is called before a span is stored.
func (h *Hook) RegisterPre(fn HookFn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.preFns = append(h.preFns, fn)
}

// RegisterPost adds a function that is called after a span is stored.
func (h *Hook) RegisterPost(fn HookFn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.postFns = append(h.postFns, fn)
}

// RunPre invokes all registered pre-store hooks for the given span.
func (h *Hook) RunPre(span storage.Span) {
	h.mu.RLock()
	fns := make([]HookFn, len(h.preFns))
	copy(fns, h.preFns)
	h.mu.RUnlock()
	for _, fn := range fns {
		fn(span)
	}
}

// RunPost invokes all registered post-store hooks for the given span.
func (h *Hook) RunPost(span storage.Span) {
	h.mu.RLock()
	fns := make([]HookFn, len(h.postFns))
	copy(fns, h.postFns)
	h.mu.RUnlock()
	for _, fn := range fns {
		fn(span)
	}
}

// Wrap calls RunPre, delegates to store.AddSpan, then calls RunPost.
// It returns any error from the underlying store.
func (h *Hook) Wrap(store *storage.TraceStore, span storage.Span) error {
	h.RunPre(span)
	err := store.AddSpan(span)
	if err == nil {
		h.RunPost(span)
	}
	return err
}
