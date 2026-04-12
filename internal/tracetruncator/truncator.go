// Package tracetruncator provides utilities for truncating span fields
// that exceed configurable size limits, preventing oversized traces from
// consuming excessive memory or storage.
package tracetruncator

import (
	"github.com/user/grpc-tracer/internal/storage"
)

// Options controls truncation behaviour.
type Options struct {
	// MaxServiceLen is the maximum allowed length for the service name.
	MaxServiceLen int
	// MaxMethodLen is the maximum allowed length for the method name.
	MaxMethodLen int
	// MaxTagValueLen is the maximum allowed length for any single tag value.
	MaxTagValueLen int
	// MaxTagCount is the maximum number of tags retained per span.
	MaxTagCount int
}

func defaultOptions() Options {
	return Options{
		MaxServiceLen:  64,
		MaxMethodLen:   128,
		MaxTagValueLen: 256,
		MaxTagCount:    32,
	}
}

// Truncator trims span fields to configured limits.
type Truncator struct {
	opts Options
}

// New returns a Truncator with the provided options.
// Zero-value fields in opts fall back to sensible defaults.
func New(opts Options) *Truncator {
	d := defaultOptions()
	if opts.MaxServiceLen <= 0 {
		opts.MaxServiceLen = d.MaxServiceLen
	}
	if opts.MaxMethodLen <= 0 {
		opts.MaxMethodLen = d.MaxMethodLen
	}
	if opts.MaxTagValueLen <= 0 {
		opts.MaxTagValueLen = d.MaxTagValueLen
	}
	if opts.MaxTagCount <= 0 {
		opts.MaxTagCount = d.MaxTagCount
	}
	return &Truncator{opts: opts}
}

// TruncateSpan returns a copy of s with all fields trimmed to the
// configured limits. The original span is never mutated.
func (t *Truncator) TruncateSpan(s storage.Span) storage.Span {
	s.ServiceName = trunc(s.ServiceName, t.opts.MaxServiceLen)
	s.Method = trunc(s.Method, t.opts.MaxMethodLen)

	if len(s.Tags) > 0 {
		truncated := make(map[string]string, len(s.Tags))
		count := 0
		for k, v := range s.Tags {
			if count >= t.opts.MaxTagCount {
				break
			}
			truncated[k] = trunc(v, t.opts.MaxTagValueLen)
			count++
		}
		s.Tags = truncated
	}
	return s
}

func trunc(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
