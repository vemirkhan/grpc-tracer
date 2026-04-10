// Package tracereplay provides functionality to replay recorded traces
// through the processing pipeline for debugging and regression testing.
package tracereplay

import (
	"errors"
	"time"

	"github.com/your-org/grpc-tracer/internal/storage"
)

// Options configures replay behaviour.
type Options struct {
	// SpeedFactor scales the inter-span delay. 1.0 = real time, 0 = no delay.
	SpeedFactor float64
	// MaxSpans limits how many spans are replayed. 0 means unlimited.
	MaxSpans int
}

func defaultOptions() Options {
	return Options{SpeedFactor: 1.0}
}

// Replayer replays spans from a source store into a destination store.
type Replayer struct {
	src  *storage.TraceStore
	dst  *storage.TraceStore
	opts Options
}

// New creates a Replayer that reads from src and writes into dst.
func New(src, dst *storage.TraceStore, opts Options) (*Replayer, error) {
	if src == nil {
		return nil, errors.New("tracereplay: source store must not be nil")
	}
	if dst == nil {
		return nil, errors.New("tracereplay: destination store must not be nil")
	}
	if opts.SpeedFactor < 0 {
		opts.SpeedFactor = 0
	}
	return &Replayer{src: src, dst: dst, opts: opts}, nil
}

// ReplayTrace replays all spans of the given traceID into the destination store.
// Spans are emitted in start-time order with optional pacing.
func (r *Replayer) ReplayTrace(traceID string) (int, error) {
	spans, err := r.src.GetTrace(traceID)
	if err != nil {
		return 0, err
	}
	if len(spans) == 0 {
		return 0, nil
	}

	sortByStart(spans)

	limit := len(spans)
	if r.opts.MaxSpans > 0 && r.opts.MaxSpans < limit {
		limit = r.opts.MaxSpans
	}

	var prev time.Time
	for i := 0; i < limit; i++ {
		sp := spans[i]
		if r.opts.SpeedFactor > 0 && !prev.IsZero() {
			gap := sp.StartTime.Sub(prev)
			delay := time.Duration(float64(gap) / r.opts.SpeedFactor)
			if delay > 0 {
				time.Sleep(delay)
			}
		}
		prev = sp.StartTime
		r.dst.AddSpan(sp)
	}
	return limit, nil
}

// ReplayAll replays every trace found in the source store.
func (r *Replayer) ReplayAll() (int, error) {
	traces := r.src.GetAllTraces()
	total := 0
	for id := range traces {
		n, err := r.ReplayTrace(id)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}
