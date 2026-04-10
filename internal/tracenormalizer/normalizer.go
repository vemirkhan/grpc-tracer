// Package tracenormalizer provides utilities for normalizing span fields
// within a trace, ensuring consistent formatting of service names, methods,
// and tag values before storage or export.
package tracenormalizer

import (
	"strings"
	"time"

	"github.com/user/grpc-tracer/internal/collector"
)

// Options controls normalization behaviour.
type Options struct {
	// LowercaseService forces service names to lowercase.
	LowercaseService bool
	// TrimMethod removes leading slashes from method names.
	TrimMethod bool
	// DefaultDuration is applied when a span has a zero duration.
	DefaultDuration time.Duration
}

// defaultOptions returns sensible defaults.
func defaultOptions() Options {
	return Options{
		LowercaseService: true,
		TrimMethod:       true,
		DefaultDuration:  0,
	}
}

// Normalizer applies field normalization to spans.
type Normalizer struct {
	opts Options
}

// New creates a Normalizer with the given options.
// Zero-value option fields fall back to package defaults.
func New(opts Options) *Normalizer {
	def := defaultOptions()
	if !opts.LowercaseService {
		def.LowercaseService = false
	}
	if !opts.TrimMethod {
		def.TrimMethod = false
	}
	if opts.DefaultDuration > 0 {
		def.DefaultDuration = opts.DefaultDuration
	}
	return &Normalizer{opts: def}
}

// NormalizeSpan returns a copy of span with normalized fields applied.
func (n *Normalizer) NormalizeSpan(s collector.Span) collector.Span {
	out := s

	if n.opts.LowercaseService {
		out.ServiceName = strings.ToLower(strings.TrimSpace(out.ServiceName))
	}

	if n.opts.TrimMethod {
		out.Method = strings.TrimLeft(strings.TrimSpace(out.Method), "/")
	}

	if out.Duration == 0 && n.opts.DefaultDuration > 0 {
		out.Duration = n.opts.DefaultDuration
	}

	return out
}

// NormalizeTrace normalizes every span in the slice and returns a new slice.
func (n *Normalizer) NormalizeTrace(spans []collector.Span) []collector.Span {
	result := make([]collector.Span, len(spans))
	for i, s := range spans {
		result[i] = n.NormalizeSpan(s)
	}
	return result
}
