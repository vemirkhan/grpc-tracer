package visualizer

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"grpc-tracer/internal/storage"
)

// Visualizer formats and displays trace data
type Visualizer struct {
	store *storage.TraceStore
}

// NewVisualizer creates a new visualizer instance
func NewVisualizer(store *storage.TraceStore) *Visualizer {
	return &Visualizer{store: store}
}

// FormatTrace returns a formatted string representation of a trace
func (v *Visualizer) FormatTrace(traceID string) (string, error) {
	spans, err := v.store.GetTrace(traceID)
	if err != nil {
		return "", err
	}

	if len(spans) == 0 {
		return "", fmt.Errorf("no spans found for trace %s", traceID)
	}

	// Sort spans by start time
	sort.Slice(spans, func(i, j int) bool {
		return spans[i].StartTime.Before(spans[j].StartTime)
	})

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\n=== Trace: %s ===\n", traceID))

	for i, span := range spans {
		duration := span.EndTime.Sub(span.StartTime)
		indent := strings.Repeat("  ", span.Depth)
		
		builder.WriteString(fmt.Sprintf("%d. %s[%s] %s\n", i+1, indent, span.SpanID[:8], span.Method))
		builder.WriteString(fmt.Sprintf("   %sService: %s\n", indent, span.ServiceName))
		builder.WriteString(fmt.Sprintf("   %sDuration: %v\n", indent, duration))
		
		if span.Error != "" {
			builder.WriteString(fmt.Sprintf("   %sError: %s\n", indent, span.Error))
		}
		
		if len(span.Metadata) > 0 {
			builder.WriteString(fmt.Sprintf("   %sMetadata: %v\n", indent, span.Metadata))
		}
		builder.WriteString("\n")
	}

	return builder.String(), nil
}

// FormatAllTraces returns a summary of all traces
func (v *Visualizer) FormatAllTraces() string {
	traces := v.store.GetAllTraces()
	
	var builder strings.Builder
	builder.WriteString("\n=== All Traces ===\n")
	builder.WriteString(fmt.Sprintf("Total traces: %d\n\n", len(traces)))

	for traceID, spans := range traces {
		if len(spans) == 0 {
			continue
		}
		
		minTime := spans[0].StartTime
		maxTime := spans[0].EndTime
		
		for _, span := range spans {
			if span.StartTime.Before(minTime) {
				minTime = span.StartTime
			}
			if span.EndTime.After(maxTime) {
				maxTime = span.EndTime
			}
		}
		
		duration := maxTime.Sub(minTime)
		builder.WriteString(fmt.Sprintf("- %s: %d spans, duration: %v\n", traceID[:16], len(spans), duration))
	}

	return builder.String()
}

// FormatTraceSince returns a formatted summary of all traces started after the given time
func (v *Visualizer) FormatTraceSince(since time.Time) string {
	traces := v.store.GetAllTraces()

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("\n=== Traces since %s ===\n", since.Format(time.RFC3339)))

	count := 0
	for traceID, spans := range traces {
		if len(spans) == 0 {
			continue
		}
		// Use the earliest span start time as the trace start time
		traceStart := spans[0].StartTime
		for _, span := range spans {
			if span.StartTime.Before(traceStart) {
				traceStart = span.StartTime
			}
		}
		if traceStart.Before(since) {
			continue
		}
		builder.WriteString(fmt.Sprintf("- %s: %d spans, started: %s\n", traceID[:16], len(spans), traceStart.Format(time.RFC3339)))
		count++
	}

	if count == 0 {
		builder.WriteString("No traces found.\n")
	} else {
		builder.WriteString(fmt.Sprintf("\nTotal: %d trace(s)\n", count))
	}

	return builder.String()
}
