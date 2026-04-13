// Package traceprofiler provides latency profiling for gRPC traces,
// computing percentile statistics (p50, p90, p99) per service and method.
package traceprofiler

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/romangurevitch/grpc-tracer/internal/storage"
)

// Profile holds latency percentile statistics for a single service+method pair.
type Profile struct {
	Service string
	Method  string
	Count   int
	P50     time.Duration
	P90     time.Duration
	P99     time.Duration
	Min     time.Duration
	Max     time.Duration
}

// Profiler computes latency profiles from a TraceStore.
type Profiler struct {
	mu    sync.Mutex
	store *storage.TraceStore
}

// New creates a new Profiler backed by the given TraceStore.
// It returns an error if store is nil.
func New(store *storage.TraceStore) (*Profiler, error) {
	if store == nil {
		return nil, fmt.Errorf("traceprofiler: store must not be nil")
	}
	return &Profiler{store: store}, nil
}

// ProfileAll computes latency profiles for every service+method pair
// found across all traces in the store.
func (p *Profiler) ProfileAll() []Profile {
	p.mu.Lock()
	defer p.mu.Unlock()

	// bucket durations by "service\x00method"
	buckets := make(map[string][]time.Duration)
	keys := make(map[string][2]string) // key -> {service, method}

	for _, trace := range p.store.GetAllTraces() {
		for _, span := range trace {
			if span.Duration <= 0 {
				continue
			}
			k := span.ServiceName + "\x00" + span.Method
			buckets[k] = append(buckets[k], span.Duration)
			if _, ok := keys[k]; !ok {
				keys[k] = [2]string{span.ServiceName, span.Method}
			}
		}
	}

	profiles := make([]Profile, 0, len(buckets))
	for k, durations := range buckets {
		profiles = append(profiles, buildProfile(keys[k][0], keys[k][1], durations))
	}

	// stable sort by service then method for deterministic output
	sort.Slice(profiles, func(i, j int) bool {
		if profiles[i].Service != profiles[j].Service {
			return profiles[i].Service < profiles[j].Service
		}
		return profiles[i].Method < profiles[j].Method
	})

	return profiles
}

// ProfileService returns latency profiles filtered to a single service name.
func (p *Profiler) ProfileService(service string) []Profile {
	all := p.ProfileAll()
	out := all[:0]
	for _, pr := range all {
		if pr.Service == service {
			out = append(out, pr)
		}
	}
	return out
}

// buildProfile computes percentile stats from a slice of durations.
func buildProfile(service, method string, durations []time.Duration) Profile {
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	n := len(sorted)
	return Profile{
		Service: service,
		Method:  method,
		Count:   n,
		Min:     sorted[0],
		Max:     sorted[n-1],
		P50:     percentile(sorted, 50),
		P90:     percentile(sorted, 90),
		P99:     percentile(sorted, 99),
	}
}

// percentile returns the p-th percentile value from a sorted slice.
func percentile(sorted []time.Duration, p int) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := (p * len(sorted)) / 100
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
