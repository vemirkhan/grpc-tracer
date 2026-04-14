// Package traceprojector provides a way to project (select a subset of fields
// from) spans in a trace, producing lightweight span summaries.
package traceprojector

import (
	"errors"
	"fmt"

	"github.com/example/grpc-tracer/internal/storage"
)

// Field identifies a span field that can be included in a projection.
type Field string

const (
	FieldTraceID  Field = "trace_id"
	FieldSpanID   Field = "span_id"
	FieldService  Field = "service"
	FieldMethod   Field = "method"
	FieldDuration Field = "duration"
	FieldError    Field = "error"
	FieldTags     Field = "tags"
)

// Projection holds the selected fields for a projected span.
type Projection map[Field]interface{}

// Projector selects a configured set of fields from spans.
type Projector struct {
	fields map[Field]struct{}
	store  *storage.TraceStore
}

// New creates a Projector that reads from store and projects the given fields.
// Returns an error if store is nil or no fields are specified.
func New(store *storage.TraceStore, fields ...Field) (*Projector, error) {
	if store == nil {
		return nil, errors.New("traceprojector: store must not be nil")
	}
	if len(fields) == 0 {
		return nil, errors.New("traceprojector: at least one field must be specified")
	}
	f := make(map[Field]struct{}, len(fields))
	for _, field := range fields {
		f[field] = struct{}{}
	}
	return &Projector{fields: f, store: store}, nil
}

// ProjectTrace returns a slice of Projection for every span in the given trace.
// Returns an error if the trace is not found.
func (p *Projector) ProjectTrace(traceID string) ([]Projection, error) {
	spans, ok := p.store.GetTrace(traceID)
	if !ok {
		return nil, fmt.Errorf("traceprojector: trace %q not found", traceID)
	}
	result := make([]Projection, 0, len(spans))
	for _, s := range spans {
		proj := make(Projection)
		if _, ok := p.fields[FieldTraceID]; ok {
			proj[FieldTraceID] = s.TraceID
		}
		if _, ok := p.fields[FieldSpanID]; ok {
			proj[FieldSpanID] = s.SpanID
		}
		if _, ok := p.fields[FieldService]; ok {
			proj[FieldService] = s.ServiceName
		}
		if _, ok := p.fields[FieldMethod]; ok {
			proj[FieldMethod] = s.Method
		}
		if _, ok := p.fields[FieldDuration]; ok {
			proj[FieldDuration] = s.Duration
		}
		if _, ok := p.fields[FieldError]; ok {
			proj[FieldError] = s.Error
		}
		if _, ok := p.fields[FieldTags]; ok {
			proj[FieldTags] = s.Tags
		}
		result = append(result, proj)
	}
	return result, nil
}
