// Package traceparent implements W3C Trace Context traceparent header
// encoding and decoding for propagating trace context across service boundaries.
package traceparent

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	// Header is the canonical W3C traceparent header name.
	Header = "traceparent"

	// version is the only supported version byte.
	version = "00"
)

var (
	ErrInvalidFormat  = errors.New("traceparent: invalid format")
	ErrUnsupportedVer = errors.New("traceparent: unsupported version")
	ErrZeroTraceID    = errors.New("traceparent: all-zero trace-id")
	ErrZeroParentID   = errors.New("traceparent: all-zero parent-id")

	reValid = regexp.MustCompile(`^[0-9a-f]{2}-[0-9a-f]{32}-[0-9a-f]{16}-[0-9a-f]{2}$`)
)

// TraceParent holds the parsed fields of a W3C traceparent header value.
type TraceParent struct {
	Version  string
	TraceID  string
	ParentID string
	Flags    string
}

// Sampled reports whether the sampled flag bit is set.
func (tp TraceParent) Sampled() bool {
	return tp.Flags == "01"
}

// String encodes the TraceParent back to its header value representation.
func (tp TraceParent) String() string {
	return fmt.Sprintf("%s-%s-%s-%s", tp.Version, tp.TraceID, tp.ParentID, tp.Flags)
}

// New creates a TraceParent with the given traceID and parentID.
// sampled controls whether the sampled flag is set.
func New(traceID, parentID string, sampled bool) TraceParent {
	flags := "00"
	if sampled {
		flags = "01"
	}
	return TraceParent{
		Version:  version,
		TraceID:  strings.ToLower(traceID),
		ParentID: strings.ToLower(parentID),
		Flags:    flags,
	}
}

// Parse decodes a traceparent header value into a TraceParent.
// It returns an error if the value is malformed or uses an unsupported version.
func Parse(value string) (TraceParent, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if !reValid.MatchString(value) {
		return TraceParent{}, ErrInvalidFormat
	}
	parts := strings.Split(value, "-")
	if parts[0] != version {
		return TraceParent{}, ErrUnsupportedVer
	}
	if parts[1] == strings.Repeat("0", 32) {
		return TraceParent{}, ErrZeroTraceID
	}
	if parts[2] == strings.Repeat("0", 16) {
		return TraceParent{}, ErrZeroParentID
	}
	return TraceParent{
		Version:  parts[0],
		TraceID:  parts[1],
		ParentID: parts[2],
		Flags:    parts[3],
	}, nil
}
