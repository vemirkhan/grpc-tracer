// Package tracecompressor provides utilities for compressing and decompressing
// trace data to reduce memory and storage overhead when exporting or archiving spans.
package tracecompressor

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"

	"github.com/user/grpc-tracer/internal/storage"
)

// Format represents the compression format to use.
type Format int

const (
	// FormatGzip uses gzip compression.
	FormatGzip Format = iota
	// FormatNone skips compression (passthrough).
	FormatNone
)

// Compressor compresses and decompresses trace data.
type Compressor struct {
	format Format
	level  int
}

// New creates a new Compressor with the given format and compression level.
// Use gzip.DefaultCompression (-1) for the default level.
func New(format Format, level int) *Compressor {
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		level = gzip.DefaultCompression
	}
	return &Compressor{format: format, level: level}
}

// Compress serialises spans to JSON and compresses them.
func (c *Compressor) Compress(spans []storage.Span) ([]byte, error) {
	raw, err := json.Marshal(spans)
	if err != nil {
		return nil, fmt.Errorf("tracecompressor: marshal: %w", err)
	}
	if c.format == FormatNone {
		return raw, nil
	}
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, c.level)
	if err != nil {
		return nil, fmt.Errorf("tracecompressor: gzip writer: %w", err)
	}
	if _, err := w.Write(raw); err != nil {
		return nil, fmt.Errorf("tracecompressor: write: %w", err)
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("tracecompressor: close: %w", err)
	}
	return buf.Bytes(), nil
}

// Decompress decompresses bytes and deserialises them back into spans.
func (c *Compressor) Decompress(data []byte) ([]storage.Span, error) {
	var raw []byte
	if c.format == FormatNone {
		raw = data
	} else {
		r, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("tracecompressor: gzip reader: %w", err)
		}
		defer r.Close()
		decompressed, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("tracecompressor: read: %w", err)
		}
		raw = decompressed
	}
	var spans []storage.Span
	if err := json.Unmarshal(raw, &spans); err != nil {
		return nil, fmt.Errorf("tracecompressor: unmarshal: %w", err)
	}
	return spans, nil
}
