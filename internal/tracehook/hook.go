// Package tracehook provides lifecycle hooks that fire before and after
// a span is written to a TraceStore, enabling side-effects such as
// metrics, alerting, or audit logging without coupling them to core storage.
package tracehook

import (
	"errors"
	"sync"

	"github.com/example/grpc-tracer/internal/storage"
)

// HookFn is a function invoked with a span at a specific lifecycle point.
type HookFn func(span storage.Span)

// Hook manages pre- and post-store lifecycle callbacks.
type Hook struct {
	mu   sync.RWMutex
	pre  []HookFn
	post []HookFn
}

// New returns an initialised Hook.
func New() *Hook {
	return &Hook{}
}

// RegisterPre adds a function that is called before a span is stored.
func (h *Hook) RegisterPre(fn HookFn) error {
	if fn == nil {
		return errors.New("tracehook: pre-hook function must not be nil")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pre = append(h.pre, fn)
	return nil
}

// RegisterPost adds a function that is called after a span is stored.
func (h *Hook) RegisterPost(fn HookFn) error {
	if fn == nil {
		return errors.New("tracehook: post-hook function must not be nil")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.post = append(h.post, fn)
	return nil
}

// AddSpan fires pre-hooks, stores the span in the supplied store, then fires
// post-hooks. Any error from the store is returned; hooks always run.
func (h *Hook) AddSpan(store *storage.TraceStore, span storage.Span) error {
	h.mu.RLock()
	pre := make([]HookFn, len(h.pre))
	copy(pre, h.pre)
	post := make([]HookFn, len(h.post))
	copy(post, h.post)
	h.mu.RUnlock()

	for _, fn := range pre {
		fn(span)
	}

	store.AddSpan(span)

	for _, fn := range post {
		fn(span)
	}

	return nil
}
