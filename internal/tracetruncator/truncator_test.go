package tracetruncator

import (
	"strings"
	"testing"

	"github.com/user/grpc-tracer/internal/storage"
)

func makeSpan(service, method string, tags map[string]string) storage.Span {
	return storage.Span{
		TraceID:     "trace-1",
		SpanID:      "span-1",
		ServiceName: service,
		Method:      method,
		Tags:        tags,
	}
}

func TestNew_Defaults(t *testing.T) {
	tr := New(Options{})
	if tr.opts.MaxServiceLen != 64 {
		t.Errorf("expected MaxServiceLen 64, got %d", tr.opts.MaxServiceLen)
	}
	if tr.opts.MaxTagCount != 32 {
		t.Errorf("expected MaxTagCount 32, got %d", tr.opts.MaxTagCount)
	}
}

func TestTruncateSpan_ServiceName(t *testing.T) {
	tr := New(Options{MaxServiceLen: 5})
	s := makeSpan("verylongservicename", "SomeMethod", nil)
	out := tr.TruncateSpan(s)
	if out.ServiceName != "veryl" {
		t.Errorf("expected 'veryl', got %q", out.ServiceName)
	}
	// original must not be mutated
	if s.ServiceName != "verylongservicename" {
		t.Error("original span was mutated")
	}
}

func TestTruncateSpan_Method(t *testing.T) {
	tr := New(Options{MaxMethodLen: 4})
	s := makeSpan("svc", "LongMethodName", nil)
	out := tr.TruncateSpan(s)
	if out.Method != "Long" {
		t.Errorf("expected 'Long', got %q", out.Method)
	}
}

func TestTruncateSpan_TagValue(t *testing.T) {
	tr := New(Options{MaxTagValueLen: 3})
	s := makeSpan("svc", "m", map[string]string{"key": "toolongvalue"})
	out := tr.TruncateSpan(s)
	if out.Tags["key"] != "too" {
		t.Errorf("expected 'too', got %q", out.Tags["key"])
	}
}

func TestTruncateSpan_TagCount(t *testing.T) {
	tr := New(Options{MaxTagCount: 3})
	tags := map[string]string{}
	for i := 0; i < 10; i++ {
		tags[strings.Repeat("k", i+1)] = "v"
	}
	s := makeSpan("svc", "m", tags)
	out := tr.TruncateSpan(s)
	if len(out.Tags) > 3 {
		t.Errorf("expected at most 3 tags, got %d", len(out.Tags))
	}
}

func TestTruncateSpan_ShortFieldsUnchanged(t *testing.T) {
	tr := New(Options{})
	s := makeSpan("svc", "Method", map[string]string{"env": "prod"})
	out := tr.TruncateSpan(s)
	if out.ServiceName != "svc" || out.Method != "Method" || out.Tags["env"] != "prod" {
		t.Error("short fields should not be modified")
	}
}

func TestTruncateSpan_NilTags(t *testing.T) {
	tr := New(Options{})
	s := makeSpan("svc", "m", nil)
	out := tr.TruncateSpan(s)
	if out.Tags != nil {
		t.Error("nil tags should remain nil")
	}
}
