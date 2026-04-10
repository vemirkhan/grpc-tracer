// Package spanprocessor provides a pipeline for processing spans before storage.
// Processors can transform, enrich, or drop spans as they flow through the chain.
package spanprocessor

import (
	"github.com/user/grpc-tracer/internal/collector"
)

// Processor is a function that receives a span and returns a (possibly modified)
// span and whether to keep it. Returning false drops the span from the pipeline.
type Processor func(span collector.Span) (collector.Span, bool)

// Pipeline is an ordered list of Processors applied sequentially to each span.
type Pipeline struct {
	steps []Processor
}

// New creates a new Pipeline with the given processors applied in order.
func New(processors ...Processor) *Pipeline {
	return &Pipeline{steps: processors}
}

// Process runs the span through all processors in order.
// If any processor drops the span (returns false), processing stops and
// the zero span is returned with ok=false.
func (p *Pipeline) Process(span collector.Span) (collector.Span, bool) {
	for _, step := range p.steps {
		var keep bool
		span, keep = step(span)
		if !keep {
			return collector.Span{}, false
		}
	}
	return span, true
}

// Enrich returns a Processor that adds the given key/value pair to the span's
// Tags map, creating the map if it does not already exist.
func Enrich(key, value string) Processor {
	return func(span collector.Span) (collector.Span, bool) {
		if span.Tags == nil {
			span.Tags = make(map[string]string)
		}
		span.Tags[key] = value
		return span, true
	}
}

// DropOnError returns a Processor that drops any span whose Error field is set.
func DropOnError() Processor {
	return func(span collector.Span) (collector.Span, bool) {
		if span.Error != "" {
			return collector.Span{}, false
		}
		return span, true
	}
}

// RequireService returns a Processor that drops spans not belonging to any of
// the specified service names.
func RequireService(names ...string) Processor {
	allowed := make(map[string]struct{}, len(names))
	for _, n := range names {
		allowed[n] = struct{}{}
	}
	return func(span collector.Span) (collector.Span, bool) {
		_, ok := allowed[span.Service]
		return span, ok
	}
}
