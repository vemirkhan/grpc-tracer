package redactor_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/redactor"
	"github.com/user/grpc-tracer/internal/storage"
)

func makeSpan(meta map[string]string) storage.Span {
	return storage.Span{
		TraceID:   "t1",
		SpanID:    "s1",
		Service:   "svc",
		Method:    "/pkg.Svc/Method",
		StartTime: time.Now(),
		Duration:  10 * time.Millisecond,
		Metadata:  meta,
	}
}

func TestNew_DefaultKeys(t *testing.T) {
	r := redactor.New(nil, "")
	if r == nil {
		t.Fatal("expected non-nil Redactor")
	}
}

func TestRedactSpan_SensitiveKeyReplaced(t *testing.T) {
	r := redactor.New(nil, "[REDACTED]")
	span := makeSpan(map[string]string{
		"authorization": "Bearer secret-token",
		"x-request-id":  "abc123",
	})

	got := r.RedactSpan(span)

	if got.Metadata["authorization"] != "[REDACTED]" {
		t.Errorf("expected authorization to be redacted, got %q", got.Metadata["authorization"])
	}
	if got.Metadata["x-request-id"] != "abc123" {
		t.Errorf("expected x-request-id to be unchanged, got %q", got.Metadata["x-request-id"])
	}
}

func TestRedactSpan_CaseInsensitive(t *testing.T) {
	r := redactor.New(nil, "***")
	span := makeSpan(map[string]string{
		"Authorization": "Bearer token",
		"X-API-KEY":     "my-key",
	})

	got := r.RedactSpan(span)

	for _, key := range []string{"Authorization", "X-API-KEY"} {
		if got.Metadata[key] != "***" {
			t.Errorf("expected %q to be redacted, got %q", key, got.Metadata[key])
		}
	}
}

func TestRedactSpan_NoMetadata(t *testing.T) {
	r := redactor.New(nil, "")
	span := makeSpan(nil)
	got := r.RedactSpan(span)
	if got.Metadata != nil {
		t.Errorf("expected nil metadata, got %v", got.Metadata)
	}
}

func TestRedactTrace_AllSpansRedacted(t *testing.T) {
	r := redactor.New([]string{"secret"}, "[HIDDEN]")
	trace := storage.Trace{
		TraceID: "t1",
		Spans: []storage.Span{
			makeSpan(map[string]string{"secret": "val1", "ok": "keep"}),
			makeSpan(map[string]string{"secret": "val2", "also-ok": "keep"}),
		},
	}

	got := r.RedactTrace(trace)

	for i, s := range got.Spans {
		if s.Metadata["secret"] != "[HIDDEN]" {
			t.Errorf("span %d: expected secret to be hidden, got %q", i, s.Metadata["secret"])
		}
	}
	if got.Spans[0].Metadata["ok"] != "keep" {
		t.Errorf("expected ok key to be preserved")
	}
}

func TestRedactSpan_OriginalUnmodified(t *testing.T) {
	r := redactor.New(nil, "[REDACTED]")
	orig := makeSpan(map[string]string{"authorization": "secret"})
	_ = r.RedactSpan(orig)
	if orig.Metadata["authorization"] != "secret" {
		t.Error("original span metadata was modified")
	}
}
