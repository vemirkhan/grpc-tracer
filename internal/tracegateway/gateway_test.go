package tracegateway_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/grpc-tracer/internal/collector"
	"github.com/grpc-tracer/internal/storage"
	"github.com/grpc-tracer/internal/tracegateway"
	"github.com/grpc-tracer/internal/visualizer"
)

func makeGateway(t *testing.T) (*tracegateway.Gateway, *storage.TraceStore) {
	t.Helper()
	store := storage.NewTraceStore()
	col := collector.NewCollector(store)
	vis := visualizer.NewVisualizer(store)
	_ = col
	gw := tracegateway.New(store, vis)
	return gw, store
}

func addSpan(t *testing.T, store *storage.TraceStore, traceID, spanID, service string) {
	t.Helper()
	store.AddSpan(storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   service,
		Method:    "/svc/Method",
		StartTime: time.Now(),
		Duration:  10 * time.Millisecond,
	})
}

func TestHandleHealth(t *testing.T) {
	gw, _ := makeGateway(t)
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %q", body["status"])
	}
}

func TestHandleListAll_Empty(t *testing.T) {
	gw, _ := makeGateway(t)
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/traces", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&body)
	if int(body["count"].(float64)) != 0 {
		t.Errorf("expected 0 traces")
	}
}

func TestHandleListAll_WithTraces(t *testing.T) {
	gw, store := makeGateway(t)
	addSpan(t, store, "trace-1", "span-1", "svc-a")
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/traces", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&body)
	if int(body["count"].(float64)) != 1 {
		t.Errorf("expected 1 trace")
	}
}

func TestHandleGetTrace_Found(t *testing.T) {
	gw, store := makeGateway(t)
	addSpan(t, store, "trace-abc", "span-1", "svc-a")
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/traces/trace-abc", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&body)
	if body["trace_id"] != "trace-abc" {
		t.Errorf("unexpected trace_id: %v", body["trace_id"])
	}
}

func TestHandleGetTrace_NotFound(t *testing.T) {
	gw, _ := makeGateway(t)
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/traces/missing", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestHandleListAll_MethodNotAllowed(t *testing.T) {
	gw, _ := makeGateway(t)
	rec := httptest.NewRecorder()
	gw.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/traces", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
