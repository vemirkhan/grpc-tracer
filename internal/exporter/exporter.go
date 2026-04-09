package exporter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/grpc-tracer/internal/storage"
)

// Format represents the output format for exported traces.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Exporter handles exporting trace data to various outputs.
type Exporter struct {
	store  *storage.TraceStore
	format Format
	writer io.Writer
}

// NewExporter creates a new Exporter with the given store, format, and writer.
func NewExporter(store *storage.TraceStore, format Format, writer io.Writer) *Exporter {
	if writer == nil {
		writer = os.Stdout
	}
	return &Exporter{
		store:  store,
		format: format,
		writer: writer,
	}
}

// ExportTrace exports a single trace by ID.
func (e *Exporter) ExportTrace(traceID string) error {
	spans, ok := e.store.GetTrace(traceID)
	if !ok {
		return fmt.Errorf("trace %q not found", traceID)
	}
	switch e.format {
	case FormatJSON:
		return e.writeJSON(map[string]interface{}{"trace_id": traceID, "spans": spans})
	case FormatText:
		return e.writeText(traceID, spans)
	default:
		return fmt.Errorf("unsupported format: %s", e.format)
	}
}

// ExportAll exports all traces.
func (e *Exporter) ExportAll() error {
	traces := e.store.GetAllTraces()
	switch e.format {
	case FormatJSON:
		return e.writeJSON(map[string]interface{}{"exported_at": time.Now().UTC(), "traces": traces})
	case FormatText:
		for id, spans := range traces {
			if err := e.writeText(id, spans); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("unsupported format: %s", e.format)
	}
}

func (e *Exporter) writeJSON(v interface{}) error {
	enc := json.NewEncoder(e.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func (e *Exporter) writeText(traceID string, spans interface{}) error {
	_, err := fmt.Fprintf(e.writer, "=== Trace: %s ===\n%+v\n", traceID, spans)
	return err
}
