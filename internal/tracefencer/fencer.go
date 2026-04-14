// Package tracefencer provides span-level access control by allowing or
// denying spans based on configurable service and method allow/deny lists.
package tracefencer

import (
	"errors"
	"strings"

	"github.com/example/grpc-tracer/internal/storage"
)

// ErrDenied is returned when a span is blocked by the fencer.
var ErrDenied = errors.New("tracefencer: span denied by policy")

// Policy defines the access control policy.
type Policy struct {
	// AllowServices is a whitelist of service names (case-insensitive).
	// If non-empty, only spans from listed services are allowed.
	AllowServices []string

	// DenyServices is a blacklist of service names (case-insensitive).
	DenyServices []string

	// DenyMethods is a blacklist of method names (case-insensitive substring).
	DenyMethods []string
}

// Fencer enforces a Policy against incoming spans.
type Fencer struct {
	policy Policy
}

// New creates a Fencer with the given Policy.
func New(p Policy) *Fencer {
	return &Fencer{policy: p}
}

// Allow returns nil if the span is permitted, or ErrDenied otherwise.
func (f *Fencer) Allow(span storage.Span) error {
	svc := strings.ToLower(span.ServiceName)
	method := strings.ToLower(span.Method)

	// Check deny-service list first.
	for _, d := range f.policy.DenyServices {
		if strings.ToLower(d) == svc {
			return ErrDenied
		}
	}

	// Check deny-method list.
	for _, dm := range f.policy.DenyMethods {
		if strings.Contains(method, strings.ToLower(dm)) {
			return ErrDenied
		}
	}

	// If allow list is non-empty, service must be present.
	if len(f.policy.AllowServices) > 0 {
		for _, a := range f.policy.AllowServices {
			if strings.ToLower(a) == svc {
				return nil
			}
		}
		return ErrDenied
	}

	return nil
}

// FilterSpans returns only the spans permitted by the policy.
func (f *Fencer) FilterSpans(spans []storage.Span) []storage.Span {
	out := make([]storage.Span, 0, len(spans))
	for _, s := range spans {
		if f.Allow(s) == nil {
			out = append(out, s)
		}
	}
	return out
}
