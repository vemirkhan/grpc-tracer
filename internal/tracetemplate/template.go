// Package tracetemplate provides span template creation and application
// for stamping common fields onto new spans before they are stored.
package tracetemplate

import (
	"errors"
	"strings"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// Template holds default field values applied to spans before storage.
type Template struct {
	Service  string
	Tags     map[string]string
	minDur   time.Duration
}

// Option configures a Template.
type Option func(*Template)

// WithService sets a default service name applied when the span's service is empty.
func WithService(svc string) Option {
	return func(t *Template) { t.Service = svc }
}

// WithTag adds a default tag applied when the key is absent from the span.
func WithTag(key, value string) Option {
	return func(t *Template) { t.Tags[key] = value }
}

// WithMinDuration sets a floor duration; spans shorter than this are bumped up.
func WithMinDuration(d time.Duration) Option {
	return func(t *Template) { t.minDur = d }
}

// New creates a Template with the supplied options.
func New(opts ...Option) (*Template, error) {
	t := &Template{Tags: make(map[string]string)}
	for _, o := range opts {
		o(t)
	}
	return t, nil
}

// Apply stamps template defaults onto span and returns the modified copy.
func (t *Template) Apply(span storage.Span) storage.Span {
	if span.Service == "" && t.Service != "" {
		span.Service = t.Service
	}
	span.Service = strings.TrimSpace(span.Service)

	if span.Tags == nil {
		span.Tags = make(map[string]string)
	}
	for k, v := range t.Tags {
		if _, exists := span.Tags[k]; !exists {
			span.Tags[k] = v
		}
	}

	if t.minDur > 0 && span.Duration < t.minDur {
		span.Duration = t.minDur
	}
	return span
}

// ApplyToTrace fetches all spans for traceID, applies the template, and
// re-stores each modified span. Returns an error if the trace is not found.
func (t *Template) ApplyToTrace(store *storage.TraceStore, traceID string) error {
	spans, ok := store.GetTrace(traceID)
	if !ok {
		return errors.New("tracetemplate: trace not found: " + traceID)
	}
	for _, s := range spans {
		modified := t.Apply(s)
		store.AddSpan(modified)
	}
	return nil
}
