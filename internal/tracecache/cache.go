// Package tracecache provides a read-through cache layer for trace lookups,
// reducing repeated reads from the underlying TraceStore.
package tracecache

import (
	"sync"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
)

// Options configures the TraceCache.
type Options struct {
	// TTL is how long a cached trace is considered valid.
	TTL time.Duration
	// MaxEntries is the maximum number of traces to hold in the cache.
	MaxEntries int
}

func defaultOptions() Options {
	return Options{
		TTL:        30 * time.Second,
		MaxEntries: 256,
	}
}

type entry struct {
	spans    []storage.Span
	fetchedAt time.Time
}

// TraceCache wraps a TraceStore with an in-memory LRU-style cache.
type TraceCache struct {
	mu      sync.Mutex
	store   *storage.TraceStore
	opts    Options
	cache   map[string]entry
	order   []string // insertion order for eviction
}

// New creates a TraceCache backed by store.
func New(store *storage.TraceStore, opts ...Options) (*TraceCache, error) {
	if store == nil {
		return nil, errNilStore
	}
	o := defaultOptions()
	if len(opts) > 0 {
		if opts[0].TTL > 0 {
			o.TTL = opts[0].TTL
		}
		if opts[0].MaxEntries > 0 {
			o.MaxEntries = opts[0].MaxEntries
		}
	}
	return &TraceCache{
		store: store,
		opts:  o,
		cache: make(map[string]entry),
	}, nil
}

// GetTrace returns spans for traceID, using the cache when possible.
func (tc *TraceCache) GetTrace(traceID string) ([]storage.Span, bool) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if e, ok := tc.cache[traceID]; ok {
		if time.Since(e.fetchedAt) < tc.opts.TTL {
			return e.spans, true
		}
		tc.evict(traceID)
	}

	spans, ok := tc.store.GetTrace(traceID)
	if !ok {
		return nil, false
	}
	tc.insert(traceID, spans)
	return spans, true
}

// Invalidate removes a traceID from the cache.
func (tc *TraceCache) Invalidate(traceID string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.evict(traceID)
}

// Len returns the number of cached entries.
func (tc *TraceCache) Len() int {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return len(tc.cache)
}

func (tc *TraceCache) insert(traceID string, spans []storage.Span) {
	if len(tc.cache) >= tc.opts.MaxEntries && len(tc.order) > 0 {
		oldest := tc.order[0]
		tc.order = tc.order[1:]
		delete(tc.cache, oldest)
	}
	tc.cache[traceID] = entry{spans: spans, fetchedAt: time.Now()}
	tc.order = append(tc.order, traceID)
}

func (tc *TraceCache) evict(traceID string) {
	delete(tc.cache, traceID)
	for i, id := range tc.order {
		if id == traceID {
			tc.order = append(tc.order[:i], tc.order[i+1:]...)
			break
		}
	}
}
