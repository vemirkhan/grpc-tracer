// Package tracepipeline provides a composable pipeline for processing spans
// as they flow through the tracing system. A Pipeline chains together
// multiple processing stages — validation, normalization, redaction, enrichment,
// and storage — in a single call.
package tracepipeline

import (
	"errors"
	"fmt"

	"github.com/your-org/grpc-tracer/internal/storage"
)

// Stage is a single processing step in the pipeline.
// It receives a span, may mutate or inspect it, and returns the (possibly
// modified) span or an error. Returning a non-nil error halts the pipeline
// and the span is not stored.
type Stage func(span storage.Span) (storage.Span, error)

// Pipeline is an ordered sequence of Stages followed by a sink that persists
// the final span.
type Pipeline struct {
	stages []Stage
	store  *storage.TraceStore
}

// Option configures a Pipeline.
type Option func(*Pipeline)

// WithStage appends a Stage to the pipeline.
func WithStage(s Stage) Option {
	return func(p *Pipeline) {
		p.stages = append(p.stages, s)
	}
}

// New constructs a Pipeline that writes accepted spans to store.
// At least one Stage should be provided via WithStage; an empty pipeline
// simply stores every span as-is.
func New(store *storage.TraceStore, opts ...Option) (*Pipeline, error) {
	if store == nil {
		return nil, errors.New("tracepipeline: store must not be nil")
	}
	p := &Pipeline{store: store}
	for _, o := range opts {
		o(p)
	}
	return p, nil
}

// Process runs span through every Stage in order.
// If a Stage returns an error the span is dropped and the error is returned
// to the caller (wrapped with the stage index for easy debugging).
// On success the final span is persisted to the underlying TraceStore.
func (p *Pipeline) Process(span storage.Span) error {
	current := span
	for i, stage := range p.stages {
		var err error
		current, err = stage(current)
		if err != nil {
			return fmt.Errorf("tracepipeline: stage %d: %w", i, err)
		}
	}
	p.store.AddSpan(current)
	return nil
}

// ProcessAll runs Process for every span in the slice, collecting all errors.
// Processing continues even if individual spans fail so that a single bad span
// cannot block the rest of the batch.
func (p *Pipeline) ProcessAll(spans []storage.Span) []error {
	var errs []error
	for _, s := range spans {
		if err := p.Process(s); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
