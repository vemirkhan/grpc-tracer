// Package spanlinker provides utilities for linking related spans across
// trace boundaries, enabling parent-child and follow-from relationships
// to be established between spans in different traces.
package spanlinker

import (
	"fmt"

	"github.com/grpc-tracer/internal/storage"
)

// LinkKind describes the relationship between two linked spans.
type LinkKind string

const (
	// LinkKindChildOf indicates the linked span is a child of the source span.
	LinkKindChildOf LinkKind = "child_of"
	// LinkKindFollowsFrom indicates the linked span follows causally from the source.
	LinkKindFollowsFrom LinkKind = "follows_from"
)

// SpanLink represents a directional relationship between two spans.
type SpanLink struct {
	FromTraceID string   `json:"from_trace_id"`
	FromSpanID  string   `json:"from_span_id"`
	ToTraceID   string   `json:"to_trace_id"`
	ToSpanID    string   `json:"to_span_id"`
	Kind        LinkKind `json:"kind"`
}

// Linker maintains a registry of span links.
type Linker struct {
	links []SpanLink
}

// New creates a new Linker.
func New() *Linker {
	return &Linker{}
}

// Add records a new link between two spans.
func (l *Linker) Add(link SpanLink) error {
	if link.FromTraceID == "" || link.FromSpanID == "" {
		return fmt.Errorf("spanlinker: from trace/span IDs must not be empty")
	}
	if link.ToTraceID == "" || link.ToSpanID == "" {
		return fmt.Errorf("spanlinker: to trace/span IDs must not be empty")
	}
	if link.Kind == "" {
		link.Kind = LinkKindChildOf
	}
	l.links = append(l.links, link)
	return nil
}

// LinksFrom returns all links originating from the given span.
func (l *Linker) LinksFrom(traceID, spanID string) []SpanLink {
	var result []SpanLink
	for _, lnk := range l.links {
		if lnk.FromTraceID == traceID && lnk.FromSpanID == spanID {
			result = append(result, lnk)
		}
	}
	return result
}

// LinksTo returns all links pointing to the given span.
func (l *Linker) LinksTo(traceID, spanID string) []SpanLink {
	var result []SpanLink
	for _, lnk := range l.links {
		if lnk.ToTraceID == traceID && lnk.ToSpanID == spanID {
			result = append(result, lnk)
		}
	}
	return result
}

// All returns a copy of all registered links.
func (l *Linker) All() []SpanLink {
	out := make([]SpanLink, len(l.links))
	copy(out, l.links)
	return out
}

// ResolveLinkedSpans looks up the destination spans from the store for all
// links originating from the given span.
func (l *Linker) ResolveLinkedSpans(store *storage.TraceStore, traceID, spanID string) []storage.Span {
	links := l.LinksFrom(traceID, spanID)
	var spans []storage.Span
	for _, lnk := range links {
		trace, ok := store.GetTrace(lnk.ToTraceID)
		if !ok {
			continue
		}
		for _, s := range trace {
			if s.SpanID == lnk.ToSpanID {
				spans = append(spans, s)
				break
			}
		}
	}
	return spans
}
