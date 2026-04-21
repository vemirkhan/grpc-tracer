// Package tracecloner provides utilities for deep-copying traces and spans
// between stores, enabling safe forking of trace data for parallel processing.
package tracecloner

import (
	"errors"
	"fmt"

	"github.com/user/grpc-tracer/internal/storage"
)

// ErrNilSource is returned when the source store is nil.
var ErrNilSource = errors.New("tracecloner: source store must not be nil")

// ErrNilDestination is returned when the destination store is nil.
var ErrNilDestination = errors.New("tracecloner: destination store must not be nil")

// Cloner copies traces from one TraceStore to another.
type Cloner struct {
	src  *storage.TraceStore
	dest *storage.TraceStore
}

// New creates a Cloner that reads from src and writes cloned spans to dest.
func New(src, dest *storage.TraceStore) (*Cloner, error) {
	if src == nil {
		return nil, ErrNilSource
	}
	if dest == nil {
		return nil, ErrNilDestination
	}
	return &Cloner{src: src, dest: dest}, nil
}

// CloneTrace copies all spans belonging to traceID from src into dest.
// Returns an error if the trace does not exist in the source store.
func (c *Cloner) CloneTrace(traceID string) error {
	spans, ok := c.src.GetTrace(traceID)
	if !ok {
		return fmt.Errorf("tracecloner: trace %q not found in source", traceID)
	}
	for _, s := range spans {
		cloned := cloneSpan(s)
		c.dest.AddSpan(cloned)
	}
	return nil
}

// CloneAll copies every trace from src into dest.
func (c *Cloner) CloneAll() int {
	all := c.src.GetAllTraces()
	count := 0
	for _, spans := range all {
		for _, s := range spans {
			c.dest.AddSpan(cloneSpan(s))
		}
		count++
	}
	return count
}

// cloneSpan returns a deep copy of a storage.Span.
func cloneSpan(s storage.Span) storage.Span {
	copy := s
	if s.Tags != nil {
		copy.Tags = make(map[string]string, len(s.Tags))
		for k, v := range s.Tags {
			copy.Tags[k] = v
		}
	}
	return copy
}
