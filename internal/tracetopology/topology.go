// Package tracetopology builds a service dependency graph from collected traces.
package tracetopology

import (
	"errors"
	"sync"

	"github.com/example/grpc-tracer/internal/storage"
)

// Edge represents a directed call between two services.
type Edge struct {
	From  string
	To    string
	Calls int
}

// Graph holds the adjacency information for all observed services.
type Graph struct {
	mu    sync.RWMutex
	edges map[string]map[string]int // from -> to -> count
}

// Topology builds and queries a service dependency graph.
type Topology struct {
	store *storage.TraceStore
}

// New returns a Topology backed by the given store.
func New(store *storage.TraceStore) (*Topology, error) {
	if store == nil {
		return nil, errors.New("tracetopology: store must not be nil")
	}
	return &Topology{store: store}, nil
}

// Build walks all traces in the store and constructs a dependency Graph.
func (t *Topology) Build() *Graph {
	g := &Graph{edges: make(map[string]map[string]int)}

	traces := t.store.GetAllTraces()
	for _, spans := range traces {
		// index spans by spanID for parent lookup
		byID := make(map[string]storage.Span, len(spans))
		for _, s := range spans {
			byID[s.SpanID] = s
		}
		for _, s := range spans {
			if s.ParentSpanID == "" {
				continue
			}
			parent, ok := byID[s.ParentSpanID]
			if !ok || parent.ServiceName == s.ServiceName {
				continue
			}
			g.mu.Lock()
			if g.edges[parent.ServiceName] == nil {
				g.edges[parent.ServiceName] = make(map[string]int)
			}
			g.edges[parent.ServiceName][s.ServiceName]++
			g.mu.Unlock()
		}
	}
	return g
}

// Edges returns a flat slice of all directed edges in the graph.
func (g *Graph) Edges() []Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var result []Edge
	for from, tos := range g.edges {
		for to, count := range tos {
			result = append(result, Edge{From: from, To: to, Calls: count})
		}
	}
	return result
}

// Services returns the set of all service names observed in the graph.
func (g *Graph) Services() []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	seen := make(map[string]struct{})
	for from, tos := range g.edges {
		seen[from] = struct{}{}
		for to := range tos {
			seen[to] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	return out
}
