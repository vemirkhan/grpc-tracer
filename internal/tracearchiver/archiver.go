// Package tracearchiver provides functionality for archiving completed or
// expired traces to a secondary store, freeing up space in the primary store
// while preserving trace data for later analysis.
package tracearchiver

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// ErrNilSource is returned when a nil source store is provided.
var ErrNilSource = errors.New("tracearchiver: source store must not be nil")

// ErrNilArchive is returned when a nil archive store is provided.
var ErrNilArchive = errors.New("tracearchiver: archive store must not be nil")

// Options configures archival behaviour.
type Options struct {
	// MaxAge is the minimum age of a trace's last span before it is eligible
	// for archival. Defaults to 5 minutes.
	MaxAge time.Duration

	// DeleteAfterArchive removes the trace from the source store once it has
	// been successfully copied to the archive. Defaults to true.
	DeleteAfterArchive bool
}

func defaultOptions() Options {
	return Options{
		MaxAge:             5 * time.Minute,
		DeleteAfterArchive: true,
	}
}

// Archiver moves eligible traces from a source store to an archive store.
type Archiver struct {
	mu      sync.Mutex
	source  *storage.TraceStore
	archive *storage.TraceStore
	opts    Options
}

// New creates a new Archiver.
// source is the live trace store; archive is the destination for old traces.
func New(source, archive *storage.TraceStore, opts *Options) (*Archiver, error) {
	if source == nil {
		return nil, ErrNilSource
	}
	if archive == nil {
		return nil, ErrNilArchive
	}

	o := defaultOptions()
	if opts != nil {
		if opts.MaxAge > 0 {
			o.MaxAge = opts.MaxAge
		}
		o.DeleteAfterArchive = opts.DeleteAfterArchive
	}

	return &Archiver{
		source:  source,
		archive: archive,
		opts:    o,
	}, nil
}

// ArchiveResult summarises a single archival run.
type ArchiveResult struct {
	Archived int
	Deleted  int
	Errors   []string
}

// Run scans the source store and archives all traces whose newest span is
// older than opts.MaxAge. It returns a summary of what was done.
func (a *Archiver) Run() ArchiveResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	result := ArchiveResult{}
	cutoff := time.Now().Add(-a.opts.MaxAge)

	traces := a.source.GetAllTraces()
	for traceID, spans := range traces {
		if len(spans) == 0 {
			continue
		}

		// Find the most recent span end time.
		var newest time.Time
		for _, sp := range spans {
			end := sp.StartTime.Add(sp.Duration)
			if end.After(newest) {
				newest = end
			}
		}

		if newest.After(cutoff) {
			// Trace is still fresh; skip it.
			continue
		}

		// Copy all spans to the archive store.
		var copyErr error
		for _, sp := range spans {
			if err := a.archive.AddSpan(sp); err != nil {
				copyErr = err
				result.Errors = append(result.Errors,
					fmt.Sprintf("archive span %s/%s: %v", traceID, sp.SpanID, err))
				break
			}
		}
		if copyErr != nil {
			continue
		}
		result.Archived++

		if a.opts.DeleteAfterArchive {
			a.source.DeleteTrace(traceID)
			result.Deleted++
		}
	}

	return result
}
