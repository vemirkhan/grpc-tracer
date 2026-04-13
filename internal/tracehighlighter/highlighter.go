// Package tracehighlighter marks spans that match user-defined highlight
// criteria, adding a "highlight" tag so downstream visualizers can render
// them prominently.
package tracehighlighter

import (
	"errors"
	"strings"

	"github.com/user/grpc-tracer/internal/storage"
)

// Criteria holds the conditions that must ALL be satisfied for a span to be
// highlighted. Zero-value fields are ignored.
type Criteria struct {
	ServiceName string // substring match, case-insensitive
	Method      string // substring match, case-insensitive
	OnlyErrors  bool   // when true, only spans with Error==true are highlighted
}

// Highlighter applies highlight tags to matching spans inside a TraceStore.
type Highlighter struct {
	store    *storage.TraceStore
	criteria Criteria
}

// New creates a Highlighter. store must not be nil.
func New(store *storage.TraceStore, c Criteria) (*Highlighter, error) {
	if store == nil {
		return nil, errors.New("tracehighlighter: store must not be nil")
	}
	return &Highlighter{store: store, criteria: c}, nil
}

// HighlightTrace iterates over every span in the named trace and sets the
// "highlight" tag to "true" on any span that matches the criteria.
// It returns the number of spans that were highlighted.
func (h *Highlighter) HighlightTrace(traceID string) (int, error) {
	spans, err := h.store.GetTrace(traceID)
	if err != nil {
		return 0, err
	}

	count := 0
	for i := range spans {
		if h.matches(spans[i]) {
			if spans[i].Tags == nil {
				spans[i].Tags = map[string]string{}
			}
			spans[i].Tags["highlight"] = "true"
			count++
		}
	}
	return count, nil
}

// HighlightAll applies highlighting across every trace in the store.
func (h *Highlighter) HighlightAll() (int, error) {
	traces := h.store.GetAllTraces()
	total := 0
	for traceID := range traces {
		n, err := h.HighlightTrace(traceID)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

func (h *Highlighter) matches(span storage.Span) bool {
	c := h.criteria
	if c.ServiceName != "" && !strings.Contains(
		strings.ToLower(span.ServiceName), strings.ToLower(c.ServiceName)) {
		return false
	}
	if c.Method != "" && !strings.Contains(
		strings.ToLower(span.Method), strings.ToLower(c.Method)) {
		return false
	}
	if c.OnlyErrors && !span.Error {
		return false
	}
	return true
}
