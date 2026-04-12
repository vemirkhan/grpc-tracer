// Package traceenricher attaches additional metadata to spans based on
// configurable rules, such as environment tags, host info, or custom labels.
package traceenricher

import (
	"strings"

	"github.com/example/grpc-tracer/internal/storage"
)

// Rule is a function that adds or modifies tags on a span.
type Rule func(span *storage.Span) *storage.Span

// Enricher applies a set of rules to spans before they are stored.
type Enricher struct {
	rules []Rule
}

// New creates an Enricher with the provided rules.
func New(rules ...Rule) *Enricher {
	return &Enricher{rules: rules}
}

// Enrich applies all rules to the given span and returns the modified span.
func (e *Enricher) Enrich(span storage.Span) storage.Span {
	s := &span
	for _, rule := range e.rules {
		s = rule(s)
	}
	return *s
}

// WithStaticTag returns a Rule that unconditionally sets a tag on every span.
func WithStaticTag(key, value string) Rule {
	return func(span *storage.Span) *storage.Span {
		if span.Tags == nil {
			span.Tags = make(map[string]string)
		}
		span.Tags[key] = value
		return span
	}
}

// NormalizeService returns a Rule that lowercases the service name.
func NormalizeService() Rule {
	return func(span *storage.Span) *storage.Span {
		span.ServiceName = strings.ToLower(strings.TrimSpace(span.ServiceName))
		return span
	}
}

// DefaultIfEmpty returns a Rule that sets tag key to defaultVal when the tag
// is absent or blank.
func DefaultIfEmpty(key, defaultVal string) Rule {
	return func(span *storage.Span) *storage.Span {
		if span.Tags == nil {
			span.Tags = make(map[string]string)
		}
		if strings.TrimSpace(span.Tags[key]) == "" {
			span.Tags[key] = defaultVal
		}
		return span
	}
}
