package tracecompressor_test

import (
	"compress/gzip"
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracecompressor"
)

func makeSpans() []storage.Span {
	now := time.Now()
	return []storage.Span{
		{TraceID: "trace-1", SpanID: "span-1", Service: "svc-a", Method: "/pkg.Svc/Call", StartTime: now, Duration: 5 * time.Millisecond},
		{TraceID: "trace-1", SpanID: "span-2", Service: "svc-b", Method: "/pkg.Svc/Other", StartTime: now, Duration: 2 * time.Millisecond, Error: "timeout"},
	}
}

func TestCompress_Gzip_RoundTrip(t *testing.T) {
	c := tracecompressor.New(tracecompressor.FormatGzip, gzip.DefaultCompression)
	spans := makeSpans()

	compressed, err := c.Compress(spans)
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}
	if len(compressed) == 0 {
		t.Fatal("expected non-empty compressed output")
	}

	got, err := c.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress: %v", err)
	}
	if len(got) != len(spans) {
		t.Fatalf("span count: want %d, got %d", len(spans), len(got))
	}
	for i, s := range got {
		if s.TraceID != spans[i].TraceID || s.SpanID != spans[i].SpanID {
			t.Errorf("span[%d] mismatch: want %+v, got %+v", i, spans[i], s)
		}
	}
}

func TestCompress_None_RoundTrip(t *testing.T) {
	c := tracecompressor.New(tracecompressor.FormatNone, 0)
	spans := makeSpans()

	raw, err := c.Compress(spans)
	if err != nil {
		t.Fatalf("Compress: %v", err)
	}

	got, err := c.Decompress(raw)
	if err != nil {
		t.Fatalf("Decompress: %v", err)
	}
	if len(got) != len(spans) {
		t.Fatalf("span count: want %d, got %d", len(spans), len(got))
	}
}

func TestCompress_EmptySlice(t *testing.T) {
	c := tracecompressor.New(tracecompressor.FormatGzip, gzip.BestSpeed)

	compressed, err := c.Compress([]storage.Span{})
	if err != nil {
		t.Fatalf("Compress empty: %v", err)
	}

	got, err := c.Decompress(compressed)
	if err != nil {
		t.Fatalf("Decompress empty: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 spans, got %d", len(got))
	}
}

func TestDecompress_InvalidData_ReturnsError(t *testing.T) {
	c := tracecompressor.New(tracecompressor.FormatGzip, gzip.DefaultCompression)
	_, err := c.Decompress([]byte("not-gzip-data"))
	if err == nil {
		t.Fatal("expected error for invalid gzip data, got nil")
	}
}

func TestNew_ClampsInvalidLevel(t *testing.T) {
	// Should not panic with an out-of-range level.
	c := tracecompressor.New(tracecompressor.FormatGzip, 999)
	spans := makeSpans()
	compressed, err := c.Compress(spans)
	if err != nil {
		t.Fatalf("Compress with clamped level: %v", err)
	}
	if _, err := c.Decompress(compressed); err != nil {
		t.Fatalf("Decompress with clamped level: %v", err)
	}
}
