package snapshot

import "github.com/user/grpc-tracer/internal/storage"

// Diff describes the changes between two snapshots.
type Diff struct {
	Added   []storage.Trace `json:"added"`
	Removed []storage.Trace `json:"removed"`
}

// Compare returns a Diff describing traces present in next but not in base
// (Added) and traces present in base but not in next (Removed).
func Compare(base, next *Snapshot) Diff {
	baseIndex := indexTraces(base.Traces)
	nextIndex := indexTraces(next.Traces)

	var added, removed []storage.Trace

	for id, tr := range nextIndex {
		if _, exists := baseIndex[id]; !exists {
			added = append(added, tr)
		}
	}

	for id, tr := range baseIndex {
		if _, exists := nextIndex[id]; !exists {
			removed = append(removed, tr)
		}
	}

	return Diff{Added: added, Removed: removed}
}

// indexTraces builds a map from TraceID to Trace for fast lookup.
func indexTraces(traces []storage.Trace) map[string]storage.Trace {
	m := make(map[string]storage.Trace, len(traces))
	for _, tr := range traces {
		m[tr.TraceID] = tr
	}
	return m
}
