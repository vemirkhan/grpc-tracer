// Package tracedigester computes a stable fingerprint (digest) for a trace
// based on its structural shape: the ordered sequence of service/method pairs
// across all spans. Identical call chains produce the same digest regardless
// of timing or IDs, making it easy to group or deduplicate traces by shape.
package tracedigester

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/mfreeman451/grpc-tracer/internal/storage"
)

// Digest is a hex-encoded SHA-256 fingerprint of a trace's structural shape.
type Digest string

// Digester computes structural fingerprints for traces.
type Digester struct {
	store *storage.TraceStore
}

// New returns a Digester backed by store.
// Returns an error if store is nil.
func New(store *storage.TraceStore) (*Digester, error) {
	if store == nil {
		return nil, fmt.Errorf("tracedigester: store must not be nil")
	}
	return &Digester{store: store}, nil
}

// Compute returns the structural Digest for the trace identified by traceID.
// Spans are sorted by start time before hashing so the digest is stable.
func (d *Digester) Compute(traceID string) (Digest, error) {
	spans, err := d.store.GetTrace(traceID)
	if err != nil {
		return "", fmt.Errorf("tracedigester: get trace: %w", err)
	}
	if len(spans) == 0 {
		return "", fmt.Errorf("tracedigester: trace %q has no spans", traceID)
	}

	// Stable sort by start time, then span ID for ties.
	sorted := make([]storage.Span, len(spans))
	copy(sorted, spans)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].StartTime.Equal(sorted[j].StartTime) {
			return sorted[i].SpanID < sorted[j].SpanID
		}
		return sorted[i].StartTime.Before(sorted[j].StartTime)
	})

	var sb strings.Builder
	for _, s := range sorted {
		fmt.Fprintf(&sb, "%s:%s;", s.ServiceName, s.Method)
	}

	sum := sha256.Sum256([]byte(sb.String()))
	return Digest(fmt.Sprintf("%x", sum)), nil
}

// GroupByDigest partitions all traces in the store by their structural digest.
// Traces that cannot be digested (e.g. empty) are silently skipped.
func (d *Digester) GroupByDigest() (map[Digest][]string, error) {
	allTraces := d.store.GetAllTraces()
	groups := make(map[Digest][]string)
	for traceID := range allTraces {
		dig, err := d.Compute(traceID)
		if err != nil {
			continue
		}
		groups[dig] = append(groups[dig], traceID)
	}
	return groups, nil
}
