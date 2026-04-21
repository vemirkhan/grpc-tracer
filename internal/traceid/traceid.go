// Package traceid provides utilities for generating, validating, and
// parsing trace and span identifiers used throughout grpc-tracer.
package traceid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	// TraceIDLength is the number of random bytes used for a trace ID.
	TraceIDLength = 16
	// SpanIDLength is the number of random bytes used for a span ID.
	SpanIDLength = 8
)

// NewTraceID generates a new random 128-bit trace identifier encoded as a
// lowercase hex string.
func NewTraceID() (string, error) {
	return randomHex(TraceIDLength)
}

// NewSpanID generates a new random 64-bit span identifier encoded as a
// lowercase hex string.
func NewSpanID() (string, error) {
	return randomHex(SpanIDLength)
}

// MustTraceID is like NewTraceID but panics on error.
func MustTraceID() string {
	id, err := NewTraceID()
	if err != nil {
		panic(fmt.Sprintf("traceid: failed to generate trace ID: %v", err))
	}
	return id
}

// MustSpanID is like NewSpanID but panics on error.
func MustSpanID() string {
	id, err := NewSpanID()
	if err != nil {
		panic(fmt.Sprintf("traceid: failed to generate span ID: %v", err))
	}
	return id
}

// Validate returns true when s looks like a valid hex-encoded trace or span ID
// (non-empty, even length, all hex characters).
func Validate(s string) bool {
	if len(s) == 0 || len(s)%2 != 0 {
		return false
	}
	_, err := hex.DecodeString(strings.ToLower(s))
	return err == nil
}

// Normalize returns the lowercase form of a valid hex-encoded ID.
// It returns an error if the ID fails validation.
func Normalize(s string) (string, error) {
	if !Validate(s) {
		return "", fmt.Errorf("traceid: invalid ID %q: must be a non-empty, even-length hex string", s)
	}
	return strings.ToLower(s), nil
}

// randomHex returns n random bytes encoded as a lowercase hex string.
func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("traceid: rand.Read: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
