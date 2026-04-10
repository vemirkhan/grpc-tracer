// Package tracegateway provides an HTTP gateway for querying and streaming
// trace data collected by grpc-tracer.
package tracegateway

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/grpc-tracer/internal/storage"
	"github.com/grpc-tracer/internal/visualizer"
)

// Gateway exposes trace data over HTTP.
type Gateway struct {
	store *storage.TraceStore
	vis   *visualizer.Visualizer
	mux   *http.ServeMux
}

// New creates a Gateway backed by the given store.
func New(store *storage.TraceStore, vis *visualizer.Visualizer) *Gateway {
	g := &Gateway{
		store: store,
		vis:   vis,
		mux:   http.NewServeMux(),
	}
	g.mux.HandleFunc("/traces", g.handleListAll)
	g.mux.HandleFunc("/traces/", g.handleGetTrace)
	g.mux.HandleFunc("/health", g.handleHealth)
	return g
}

// ServeHTTP implements http.Handler.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

func (g *Gateway) handleListAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	traces := g.store.GetAllTraces()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count":  len(traces),
		"traces": traces,
	})
}

func (g *Gateway) handleGetTrace(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/traces/")
	if id == "" {
		http.Error(w, "trace id required", http.StatusBadRequest)
		return
	}
	spans, ok := g.store.GetTrace(id)
	if !ok {
		http.Error(w, "trace not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"trace_id": id,
		"spans":    spans,
	})
}

func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
