// Package snapshot provides point-in-time capture of trace state,
// allowing callers to persist or compare trace data at a given moment.
package snapshot

import (
	"encoding/json"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// Snapshot holds a captured view of all traces at a specific moment.
type Snapshot struct {
	CapturedAt time.Time              `json:"captured_at"`
	Traces     []storage.Trace        `json:"traces"`
	Meta       map[string]interface{} `json:"meta,omitempty"`
}

// Capture creates a Snapshot from all traces currently held in the store.
func Capture(store *storage.TraceStore) *Snapshot {
	traces := store.GetAllTraces()
	return &Snapshot{
		CapturedAt: time.Now().UTC(),
		Traces:     traces,
		Meta: map[string]interface{}{
			"trace_count": len(traces),
		},
	}
}

// MarshalJSON serialises the snapshot to JSON bytes.
func (s *Snapshot) MarshalJSON() ([]byte, error) {
	type alias Snapshot
	return json.Marshal((*alias)(s))
}

// Summary returns a brief human-readable description of the snapshot.
func (s *Snapshot) Summary() string {
	count := 0
	if c, ok := s.Meta["trace_count"]; ok {
		if n, ok := c.(int); ok {
			count = n
		}
	}
	return time.Now().UTC().Format(time.RFC3339) +
		" — snapshot contains " +
		itoa(count) + " trace(s) captured at " +
		s.CapturedAt.Format(time.RFC3339)
}

// itoa converts an int to a decimal string without importing strconv
// at the top level (keeps imports minimal).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
