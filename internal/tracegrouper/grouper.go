// Package tracegrouper provides functionality for grouping spans and traces
// by configurable dimensions such as service name, method, status, or custom tags.
// Grouped results can be used for aggregation, reporting, or batch processing.
package tracegrouper

import (
	"errors"
	"sync"

	"github.com/user/grpc-tracer/internal/storage"
)

// GroupKey identifies a grouping dimension.
type GroupKey string

const (
	GroupByService GroupKey = "service"
	GroupByMethod  GroupKey = "method"
	GroupByStatus  GroupKey = "status"
	GroupByTag     GroupKey = "tag"
)

// Group holds spans that share a common key value.
type Group struct {
	Key   string
	Spans []storage.Span
}

// Grouper partitions spans from a trace store into named groups.
type Grouper struct {
	mu    sync.RWMutex
	store *storage.TraceStore
}

// New creates a new Grouper backed by the provided store.
// Returns an error if store is nil.
func New(store *storage.TraceStore) (*Grouper, error) {
	if store == nil {
		return nil, errors.New("tracegrouper: store must not be nil")
	}
	return &Grouper{store: store}, nil
}

// GroupTrace groups all spans in the given trace by the specified key.
// For GroupByTag, tagName specifies which tag to group on; it is ignored
// for other group keys. Spans missing the tag are placed under the empty
// string key when grouping by tag.
func (g *Grouper) GroupTrace(traceID string, by GroupKey, tagName string) ([]Group, error) {
	g.mu.RLock()
	spans, ok := g.store.GetTrace(traceID)
	g.mu.RUnlock()

	if !ok {
		return nil, errors.New("tracegrouper: trace not found: " + traceID)
	}

	return groupSpans(spans, by, tagName), nil
}

// GroupAll groups spans across all traces in the store by the specified key.
func (g *Grouper) GroupAll(by GroupKey, tagName string) []Group {
	g.mu.RLock()
	allTraces := g.store.GetAllTraces()
	g.mu.RUnlock()

	var all []storage.Span
	for _, spans := range allTraces {
		all = append(all, spans...)
	}
	return groupSpans(all, by, tagName)
}

// groupSpans partitions spans into groups according to the given key dimension.
func groupSpans(spans []storage.Span, by GroupKey, tagName string) []Group {
	buckets := make(map[string][]storage.Span)

	for _, s := range spans {
		k := resolveKey(s, by, tagName)
		buckets[k] = append(buckets[k], s)
	}

	groups := make([]Group, 0, len(buckets))
	for k, ss := range buckets {
		groups = append(groups, Group{Key: k, Spans: ss})
	}
	return groups
}

// resolveKey extracts the grouping value from a span.
func resolveKey(s storage.Span, by GroupKey, tagName string) string {
	switch by {
	case GroupByService:
		return s.ServiceName
	case GroupByMethod:
		return s.Method
	case GroupByStatus:
		if s.Error != "" {
			return "error"
		}
		return "ok"
	case GroupByTag:
		if s.Tags != nil {
			if v, ok := s.Tags[tagName]; ok {
				return v
			}
		}
		return ""
	default:
		return ""
	}
}
