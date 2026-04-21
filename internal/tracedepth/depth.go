// Package tracedepth provides utilities for measuring and enforcing
// the maximum nesting depth of spans within a trace.
package tracedepth

import (
	"errors"

	"github.com/example/grpc-tracer/internal/storage"
)

// ErrDepthExceeded is returned when a trace exceeds the configured max depth.
var ErrDepthExceeded = errors.New("tracedepth: max span depth exceeded")

// Options configures the depth analyser.
type Options struct {
	// MaxDepth is the maximum allowed nesting depth. Zero means unlimited.
	MaxDepth int
}

func defaultOptions() Options {
	return Options{MaxDepth: 0}
}

// Analyser computes and optionally enforces span depth within a trace.
type Analyser struct {
	opts Options
}

// New returns a new Analyser with the given options.
func New(opts ...func(*Options)) *Analyser {
	o := defaultOptions()
	for _, apply := range opts {
		apply(&o)
	}
	return &Analyser{opts: o}
}

// WithMaxDepth sets the maximum allowed depth.
func WithMaxDepth(d int) func(*Options) {
	return func(o *Options) { o.MaxDepth = d }
}

// Depth returns the maximum nesting depth found in the given trace.
// Root spans (no parent) are at depth 1.
func (a *Analyser) Depth(spans []storage.Span) int {
	if len(spans) == 0 {
		return 0
	}
	children := make(map[string][]string)
	hasParent := make(map[string]bool)
	for _, s := range spans {
		if s.ParentID != "" {
			children[s.ParentID] = append(children[s.ParentID], s.SpanID)
			hasParent[s.SpanID] = true
		}
	}
	maxD := 0
	var dfs func(id string, d int)
	dfs = func(id string, d int) {
		if d > maxD {
			maxD = d
		}
		for _, child := range children[id] {
			dfs(child, d+1)
		}
	}
	for _, s := range spans {
		if !hasParent[s.SpanID] {
			dfs(s.SpanID, 1)
		}
	}
	return maxD
}

// Validate returns ErrDepthExceeded if MaxDepth > 0 and the trace depth
// exceeds it; otherwise it returns nil.
func (a *Analyser) Validate(spans []storage.Span) error {
	if a.opts.MaxDepth <= 0 {
		return nil
	}
	if d := a.Depth(spans); d > a.opts.MaxDepth {
		return ErrDepthExceeded
	}
	return nil
}
