// Package traceindex provides a secondary index over trace spans,
// allowing fast lookups by service name, method, or tag key-value pairs.
package traceindex

import (
	"sync"

	"github.com/example/grpc-tracer/internal/storage"
)

// Index maintains secondary indexes over spans stored in a TraceStore.
type Index struct {
	mu      sync.RWMutex
	byService map[string][]string // service name -> []traceID
	byMethod  map[string][]string // method -> []traceID
	byTag     map[string][]string // "key=value" -> []traceID
}

// New creates an empty Index.
func New() *Index {
	return &Index{
		byService: make(map[string][]string),
		byMethod:  make(map[string][]string),
		byTag:     make(map[string][]string),
	}
}

// Index builds the secondary index from all traces in the given store.
func (idx *Index) Build(store *storage.TraceStore) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.byService = make(map[string][]string)
	idx.byMethod = make(map[string][]string)
	idx.byTag = make(map[string][]string)

	for _, trace := range store.GetAllTraces() {
		for _, span := range trace {
			traceID := span.TraceID
			addUnique(idx.byService, span.ServiceName, traceID)
			addUnique(idx.byMethod, span.Method, traceID)
			for k, v := range span.Tags {
				key := k + "=" + v
				addUnique(idx.byTag, key, traceID)
			}
		}
	}
}

// ByService returns trace IDs associated with the given service name.
func (idx *Index) ByService(name string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return copySlice(idx.byService[name])
}

// ByMethod returns trace IDs associated with the given method.
func (idx *Index) ByMethod(method string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return copySlice(idx.byMethod[method])
}

// ByTag returns trace IDs that contain a span with the given tag key=value.
func (idx *Index) ByTag(key, value string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return copySlice(idx.byTag[key+"="+value])
}

func addUnique(m map[string][]string, key, value string) {
	for _, v := range m[key] {
		if v == value {
			return
		}
	}
	m[key] = append(m[key], value)
}

func copySlice(s []string) []string {
	if len(s) == 0 {
		return nil
	}
	out := make([]string, len(s))
	copy(out, s)
	return out
}
