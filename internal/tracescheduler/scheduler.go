// Package tracescheduler provides a periodic job scheduler that runs
// maintenance tasks against a trace store, such as evicting expired traces,
// compacting storage, or triggering snapshot captures on a fixed interval.
package tracescheduler

import (
	"context"
	"sync"
	"time"
)

// Job is a function that is executed on each tick of the scheduler.
type Job func(ctx context.Context)

// Scheduler runs one or more Jobs at a configurable interval.
type Scheduler struct {
	mu       sync.Mutex
	jobs     []namedJob
	interval time.Duration
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

type namedJob struct {
	name string
	fn   Job
}

// Options configures the Scheduler.
type Options struct {
	// Interval is how often registered jobs are executed.
	// Defaults to 30 seconds.
	Interval time.Duration
}

func defaultOptions() Options {
	return Options{
		Interval: 30 * time.Second,
	}
}

// New creates a new Scheduler. If opts is nil, sensible defaults are applied.
func New(opts *Options) *Scheduler {
	defaults := defaultOptions()
	if opts != nil {
		if opts.Interval > 0 {
			defaults.Interval = opts.Interval
		}
	}
	return &Scheduler{
		interval: defaults.Interval,
		stopCh:   make(chan struct{}),
	}
}

// Register adds a named job to the scheduler. Jobs are executed in
// registration order on every tick. Register is safe to call before Start.
func (s *Scheduler) Register(name string, fn Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, namedJob{name: name, fn: fn})
}

// Start begins the scheduling loop in a background goroutine.
// The loop runs until Stop is called or ctx is cancelled.
func (s *Scheduler) Start(ctx context.Context) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.runAll(ctx)
			case <-s.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop signals the scheduling loop to exit and waits for it to finish.
func (s *Scheduler) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

// RunNow executes all registered jobs immediately in the calling goroutine.
// Useful for triggering maintenance on demand or in tests.
func (s *Scheduler) RunNow(ctx context.Context) {
	s.runAll(ctx)
}

// JobCount returns the number of registered jobs.
func (s *Scheduler) JobCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.jobs)
}

func (s *Scheduler) runAll(ctx context.Context) {
	s.mu.Lock()
	jobs := make([]namedJob, len(s.jobs))
	copy(jobs, s.jobs)
	s.mu.Unlock()

	for _, j := range jobs {
		if ctx.Err() != nil {
			return
		}
		j.fn(ctx)
	}
}
