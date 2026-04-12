// Package tracerouter routes incoming spans to one or more destination stores
// based on configurable routing rules (e.g. service name, tag matching).
package tracerouter

import (
	"errors"
	"sync"

	"github.com/example/grpc-tracer/internal/storage"
)

// Rule decides whether a span should be forwarded to a destination store.
type Rule func(span storage.Span) bool

// Route pairs a matching rule with a destination store.
type Route struct {
	Rule Rule
	Dest *storage.TraceStore
}

// Router forwards spans to destination stores based on registered routes.
// Spans that match no route are silently dropped unless a fallback is set.
type Router struct {
	mu       sync.RWMutex
	routes   []Route
	fallback *storage.TraceStore
}

// New returns an initialised Router.
func New() *Router {
	return &Router{}
}

// AddRoute registers a new routing rule and its destination.
func (r *Router) AddRoute(rule Rule, dest *storage.TraceStore) error {
	if rule == nil {
		return errors.New("tracerouter: rule must not be nil")
	}
	if dest == nil {
		return errors.New("tracerouter: destination store must not be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = append(r.routes, Route{Rule: rule, Dest: dest})
	return nil
}

// SetFallback sets a catch-all destination for spans that match no route.
func (r *Router) SetFallback(dest *storage.TraceStore) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallback = dest
}

// Route evaluates all registered rules against the span and forwards it to
// every matching destination. Returns the number of destinations written to.
func (r *Router) Route(span storage.Span) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	matched := 0
	for _, route := range r.routes {
		if route.Rule(span) {
			route.Dest.AddSpan(span)
			matched++
		}
	}
	if matched == 0 && r.fallback != nil {
		r.fallback.AddSpan(span)
		matched++
	}
	return matched
}
