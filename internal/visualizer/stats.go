package visualizer

import (
	"fmt"
	"sort"
	"time"

	"grpc-tracer/internal/storage"
)

// TraceStats contains statistics for a trace
type TraceStats struct {
	TraceID       string
	TotalSpans    int
	TotalDuration time.Duration
	Services      map[string]int
	Methods       map[string]int
	Errors        int
	AvgDuration   time.Duration
}

// CalculateStats computes statistics for a trace
func (v *Visualizer) CalculateStats(traceID string) (*TraceStats, error) {
	spans, err := v.store.GetTrace(traceID)
	if err != nil {
		return nil, err
	}

	if len(spans) == 0 {
		return nil, fmt.Errorf("no spans found for trace %s", traceID)
	}

	stats := &TraceStats{
		TraceID:  traceID,
		Services: make(map[string]int),
		Methods:  make(map[string]int),
	}

	var minTime, maxTime time.Time
	var totalSpanDuration time.Duration

	for i, span := range spans {
		if i == 0 {
			minTime = span.StartTime
			maxTime = span.EndTime
		} else {
			if span.StartTime.Before(minTime) {
				minTime = span.StartTime
			}
			if span.EndTime.After(maxTime) {
				maxTime = span.EndTime
			}
		}

		spanDuration := span.EndTime.Sub(span.StartTime)
		totalSpanDuration += spanDuration

		stats.Services[span.ServiceName]++
		stats.Methods[span.Method]++

		if span.Error != "" {
			stats.Errors++
		}
	}

	stats.TotalSpans = len(spans)
	stats.TotalDuration = maxTime.Sub(minTime)
	stats.AvgDuration = totalSpanDuration / time.Duration(len(spans))

	return stats, nil
}

// FormatStats returns a formatted string of trace statistics
func (v *Visualizer) FormatStats(traceID string) (string, error) {
	stats, err := v.CalculateStats(traceID)
	if err != nil {
		return "", err
	}

	var result string
	result += fmt.Sprintf("\n=== Trace Statistics: %s ===\n\n", traceID)
	result += fmt.Sprintf("Total Spans: %d\n", stats.TotalSpans)
	result += fmt.Sprintf("Total Duration: %v\n", stats.TotalDuration)
	result += fmt.Sprintf("Average Span Duration: %v\n", stats.AvgDuration)
	result += fmt.Sprintf("Errors: %d\n\n", stats.Errors)

	result += "Services:\n"
	services := make([]string, 0, len(stats.Services))
	for service := range stats.Services {
		services = append(services, service)
	}
	sort.Strings(services)
	for _, service := range services {
		result += fmt.Sprintf("  - %s: %d spans\n", service, stats.Services[service])
	}

	result += "\nMethods:\n"
	methods := make([]string, 0, len(stats.Methods))
	for method := range stats.Methods {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	for _, method := range methods {
		result += fmt.Sprintf("  - %s: %d calls\n", method, stats.Methods[method])
	}

	return result, nil
}
