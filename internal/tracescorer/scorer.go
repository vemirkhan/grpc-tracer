// Package tracescorer assigns a numeric quality score to a trace based on
// configurable criteria such as error rate, span count, and total duration.
package tracescorer

import (
	"errors"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// Score holds the computed result for a single trace.
type Score struct {
	TraceID   string
	Value     float64 // 0.0 (worst) – 1.0 (best)
	Penalties []string
}

// Options tunes the scoring algorithm.
type Options struct {
	// MaxDuration is the threshold above which a latency penalty is applied.
	MaxDuration time.Duration
	// MaxSpans is the upper span-count limit before a complexity penalty fires.
	MaxSpans int
	// ErrorWeight is subtracted from the score for each error span (0–1).
	ErrorWeight float64
}

func defaultOptions() Options {
	return Options{
		MaxDuration: 2 * time.Second,
		MaxSpans:    50,
		ErrorWeight: 0.3,
	}
}

// Scorer computes quality scores for traces.
type Scorer struct {
	store storage.TraceStore
	opts  Options
}

// New creates a Scorer backed by store. Any zero-value option field falls back
// to its default.
func New(store storage.TraceStore, opts Options) (*Scorer, error) {
	if store == nil {
		return nil, errors.New("tracescorer: store must not be nil")
	}
	d := defaultOptions()
	if opts.MaxDuration == 0 {
		opts.MaxDuration = d.MaxDuration
	}
	if opts.MaxSpans == 0 {
		opts.MaxSpans = d.MaxSpans
	}
	if opts.ErrorWeight == 0 {
		opts.ErrorWeight = d.ErrorWeight
	}
	return &Scorer{store: store, opts: opts}, nil
}

// ScoreTrace computes a Score for the given traceID.
func (s *Scorer) ScoreTrace(traceID string) (Score, error) {
	spans, ok := s.store.GetTrace(traceID)
	if !ok || len(spans) == 0 {
		return Score{}, errors.New("tracescorer: trace not found: " + traceID)
	}

	result := Score{TraceID: traceID, Value: 1.0}

	// Latency penalty: find total wall time (max end – min start).
	var minStart, maxEnd time.Time
	errorCount := 0
	for i, sp := range spans {
		if i == 0 || sp.StartTime.Before(minStart) {
			minStart = sp.StartTime
		}
		end := sp.StartTime.Add(sp.Duration)
		if end.After(maxEnd) {
			maxEnd = end
		}
		if sp.Error != "" {
			errorCount++
		}
	}

	total := maxEnd.Sub(minStart)
	if total > s.opts.MaxDuration {
		result.Value -= 0.2
		result.Penalties = append(result.Penalties, "high_latency")
	}

	// Complexity penalty.
	if len(spans) > s.opts.MaxSpans {
		result.Value -= 0.1
		result.Penalties = append(result.Penalties, "high_span_count")
	}

	// Error penalty.
	if errorCount > 0 {
		penalty := s.opts.ErrorWeight * float64(errorCount) / float64(len(spans))
		result.Value -= penalty
		result.Penalties = append(result.Penalties, "errors_present")
	}

	if result.Value < 0 {
		result.Value = 0
	}
	return result, nil
}
