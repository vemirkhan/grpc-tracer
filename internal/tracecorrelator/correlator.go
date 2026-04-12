// Package tracecorrelator links related traces together by shared tags or
// parent-child relationships, enabling cross-service correlation queries.
package tracecorrelator

import (
	"errors"
	"sync"

	"github.com/example/grpc-tracer/internal/storage"
)

// Correlation represents a directional link between two trace IDs.
type Correlation struct {
	FromTraceID string
	ToTraceID   string
	Reason      string // e.g. "shared-tag", "parent-child"
}

// Correlator finds and stores relationships between traces.
type Correlator struct {
	mu           sync.RWMutex
	correlations []Correlation
	store        *storage.TraceStore
}

// New returns a new Correlator backed by the given store.
func New(store *storage.TraceStore) (*Correlator, error) {
	if store == nil {
		return nil, errors.New("tracecorrelator: store must not be nil")
	}
	return &Correlator{store: store}, nil
}

// CorrelateByTag scans all traces and links those that share the same value
// for the given tag key.
func (c *Correlator) CorrelateByTag(tagKey string) []Correlation {
	traces := c.store.GetAllTraces()

	// Build an index: tagValue -> []traceID
	index := make(map[string][]string)
	for _, spans := range traces {
		if len(spans) == 0 {
			continue
		}
		traceID := spans[0].TraceID
		for _, sp := range spans {
			if v, ok := sp.Tags[tagKey]; ok && v != "" {
				index[v] = appendUnique(index[v], traceID)
			}
		}
	}

	var results []Correlation
	for _, ids := range index {
		for i := 0; i < len(ids); i++ {
			for j := i + 1; j < len(ids); j++ {
				results = append(results, Correlation{
					FromTraceID: ids[i],
					ToTraceID:   ids[j],
					Reason:      "shared-tag:" + tagKey,
				})
			}
		}
	}

	c.mu.Lock()
	c.correlations = append(c.correlations, results...)
	c.mu.Unlock()
	return results
}

// All returns a snapshot of all recorded correlations.
func (c *Correlator) All() []Correlation {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Correlation, len(c.correlations))
	copy(out, c.correlations)
	return out
}

// ForTrace returns correlations that involve the given trace ID.
func (c *Correlator) ForTrace(traceID string) []Correlation {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var out []Correlation
	for _, corr := range c.correlations {
		if corr.FromTraceID == traceID || corr.ToTraceID == traceID {
			out = append(out, corr)
		}
	}
	return out
}

func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}
