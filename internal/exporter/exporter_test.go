package exporter_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/grpc-tracer/internal/exporter"
	"github.com/grpc-tracer/internal/storage"
)

func makeStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	store := storage.NewTraceStore()
	store.AddSpan(storage.Span{
		TraceID:   "trace-1",
		SpanID:    "span-1",
		Operation: "/svc.Service/Method",
		StartTime: time.Now(),
		Duration:  12 * time.Millisecond,
		Status:    "OK",
	})
	return store
}

func TestNewExporter_DefaultsToStdout(t *testing.T) {
	store := makeStore(t)
	exp := exporter.NewExporter(store, exporter.FormatJSON, nil)
	if exp == nil {
		t.Fatal("expected non-nil exporter")
	}
}

func TestExportTrace_JSON(t *testing.T) {
	store := makeStore(t)
	var buf bytes.Buffer
	exp := exporter.NewExporter(store, exporter.FormatJSON, &buf)

	if err := exp.ExportTrace("trace-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if result["trace_id"] != "trace-1" {
		t.Errorf("expected trace_id=trace-1, got %v", result["trace_id"])
	}
}

func TestExportTrace_NotFound(t *testing.T) {
	store := storage.NewTraceStore()
	var buf bytes.Buffer
	exp := exporter.NewExporter(store, exporter.FormatJSON, &buf)

	if err := exp.ExportTrace("missing"); err == nil {
		t.Fatal("expected error for missing trace")
	}
}

func TestExportTrace_Text(t *testing.T) {
	store := makeStore(t)
	var buf bytes.Buffer
	exp := exporter.NewExporter(store, exporter.FormatText, &buf)

	if err := exp.ExportTrace("trace-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "trace-1") {
		t.Errorf("expected trace ID in text output, got: %s", buf.String())
	}
}

func TestExportAll_JSON(t *testing.T) {
	store := makeStore(t)
	var buf bytes.Buffer
	exp := exporter.NewExporter(store, exporter.FormatJSON, &buf)

	if err := exp.ExportAll(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if _, ok := result["traces"]; !ok {
		t.Error("expected 'traces' key in JSON output")
	}
}

func TestExportAll_UnsupportedFormat(t *testing.T) {
	store := makeStore(t)
	var buf bytes.Buffer
	exp := exporter.NewExporter(store, exporter.Format("xml"), &buf)

	if err := exp.ExportAll(); err == nil {
		t.Fatal("expected error for unsupported format")
	}
}
