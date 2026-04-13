// Package tracesummarizer provides utilities for producing concise summaries
// of trace data collected across microservices.
package tracesummarizer

import (
	"fmt"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// Summary holds aggregated statistics for a single trace.
type Summary struct {
	TraceID      string
	SpanCount    int
	Services     []string
	TotalDuration time.Duration
	HasErrors    bool
	RootMethod   string
}

// Summarizer produces summaries from a TraceStore.
type Summarizer struct {
	store *storage.TraceStore
}

// New creates a new Summarizer backed by the given store.
func New(store *storage.TraceStore) (*Summarizer, error) {
	if store == nil {
		return nil, fmt.Errorf("tracesummarizer: store must not be nil")
	}
	return &Summarizer{store: store}, nil
}

// Summarize returns a Summary for the trace identified by traceID.
// Returns an error if the trace is not found.
func (s *Summarizer) Summarize(traceID string) (Summary, error) {
	spans := s.store.GetTrace(traceID)
	if len(spans) == 0 {
		return Summary{}, fmt.Errorf("tracesummarizer: trace %q not found", traceID)
	}

	svcSet := make(map[string]struct{})
	var totalDur time.Duration
	hasErrors := false
	rootMethod := ""

	for _, sp := range spans {
		if sp.ServiceName != "" {
			svcSet[sp.ServiceName] = struct{}{}
		}
		totalDur += sp.Duration
		if sp.Error != "" {
			hasErrors = true
		}
		if sp.ParentSpanID == "" && sp.Method != "" {
			rootMethod = sp.Method
		}
	}

	services := make([]string, 0, len(svcSet))
	for svc := range svcSet {
		services = append(services, svc)
	}

	return Summary{
		TraceID:       traceID,
		SpanCount:     len(spans),
		Services:      services,
		TotalDuration: totalDur,
		HasErrors:     hasErrors,
		RootMethod:    rootMethod,
	}, nil
}

// SummarizeAll returns summaries for every trace currently in the store.
func (s *Summarizer) SummarizeAll() []Summary {
	allTraces := s.store.GetAllTraces()
	summaries := make([]Summary, 0, len(allTraces))
	for traceID := range allTraces {
		if sum, err := s.Summarize(traceID); err == nil {
			summaries = append(summaries, sum)
		}
	}
	return summaries
}
