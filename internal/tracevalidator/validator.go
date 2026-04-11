// Package tracevalidator provides validation logic for spans and traces
// before they are stored or forwarded in the grpc-tracer pipeline.
package tracevalidator

import (
	"errors"
	"strings"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// ErrInvalidSpan is returned when a span fails validation.
var (
	ErrMissingTraceID  = errors.New("tracevalidator: missing trace_id")
	ErrMissingSpanID   = errors.New("tracevalidator: missing span_id")
	ErrMissingService  = errors.New("tracevalidator: missing service_name")
	ErrNegativeDuration = errors.New("tracevalidator: negative duration")
	ErrFutureStart     = errors.New("tracevalidator: start_time is in the future")
)

// Options configures the validator behaviour.
type Options struct {
	// AllowFutureStart disables the future-timestamp check.
	AllowFutureStart bool
	// MaxDuration, if non-zero, rejects spans longer than this value.
	MaxDuration time.Duration
}

// Validator checks spans against a set of rules.
type Validator struct {
	opts Options
}

// New creates a Validator with the supplied options.
func New(opts Options) *Validator {
	return &Validator{opts: opts}
}

// ValidateSpan returns the first validation error found in span, or nil.
func (v *Validator) ValidateSpan(span storage.Span) error {
	if strings.TrimSpace(span.TraceID) == "" {
		return ErrMissingTraceID
	}
	if strings.TrimSpace(span.SpanID) == "" {
		return ErrMissingSpanID
	}
	if strings.TrimSpace(span.ServiceName) == "" {
		return ErrMissingService
	}
	if span.Duration < 0 {
		return ErrNegativeDuration
	}
	if !v.opts.AllowFutureStart && span.StartTime.After(time.Now().Add(5*time.Second)) {
		return ErrFutureStart
	}
	if v.opts.MaxDuration > 0 && span.Duration > v.opts.MaxDuration {
		return errors.New("tracevalidator: duration exceeds maximum allowed")
	}
	return nil
}

// ValidateAll validates every span in the slice and returns all errors keyed
// by span index.
func (v *Validator) ValidateAll(spans []storage.Span) map[int]error {
	errs := make(map[int]error)
	for i, s := range spans {
		if err := v.ValidateSpan(s); err != nil {
			errs[i] = err
		}
	}
	return errs
}
