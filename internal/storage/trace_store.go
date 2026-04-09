package storage

import (
	"sync"
	"time"
)

// TraceSpan represents a single span in a distributed trace
type TraceSpan struct {
	TraceID    string                 `json:"trace_id"`
	SpanID     string                 `json:"span_id"`
	ParentID   string                 `json:"parent_id,omitempty"`
	Service    string                 `json:"service"`
	Method     string                 `json:"method"`
	StartTime  time.Time              `json:"start_time"`
	Duration   time.Duration          `json:"duration"`
	StatusCode string                 `json:"status_code"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// TraceStore manages in-memory storage of trace spans
type TraceStore struct {
	mu     sync.RWMutex
	spans  map[string][]*TraceSpan // traceID -> spans
	maxAge time.Duration
}

// NewTraceStore creates a new trace store with specified max age for traces
func NewTraceStore(maxAge time.Duration) *TraceStore {
	store := &TraceStore{
		spans:  make(map[string][]*TraceSpan),
		maxAge: maxAge,
	}
	go store.cleanup()
	return store
}

// AddSpan adds a span to the store
func (ts *TraceStore) AddSpan(span *TraceSpan) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.spans[span.TraceID] = append(ts.spans[span.TraceID], span)
}

// GetTrace retrieves all spans for a given trace ID
func (ts *TraceStore) GetTrace(traceID string) []*TraceSpan {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	spans := ts.spans[traceID]
	result := make([]*TraceSpan, len(spans))
	copy(result, spans)
	return result
}

// GetAllTraces returns all active traces
func (ts *TraceStore) GetAllTraces() map[string][]*TraceSpan {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	result := make(map[string][]*TraceSpan, len(ts.spans))
	for traceID, spans := range ts.spans {
		copied := make([]*TraceSpan, len(spans))
		copy(copied, spans)
		result[traceID] = copied
	}
	return result
}

// cleanup periodically removes old traces
func (ts *TraceStore) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ts.mu.Lock()
		now := time.Now()
		for traceID, spans := range ts.spans {
			if len(spans) > 0 && now.Sub(spans[0].StartTime) > ts.maxAge {
				delete(ts.spans, traceID)
			}
		}
		ts.mu.Unlock()
	}
}
