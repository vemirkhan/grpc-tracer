package tracefencer_test

import (
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/tracefencer"
)

func makeSpan(svc, method string) storage.Span {
	return storage.Span{
		TraceID:     "trace-1",
		SpanID:      "span-1",
		ServiceName: svc,
		Method:      method,
		StartTime:   time.Now(),
		Duration:    10 * time.Millisecond,
	}
}

func TestAllow_NoPolicyPermitsAll(t *testing.T) {
	f := tracefencer.New(tracefencer.Policy{})
	if err := f.Allow(makeSpan("orders", "/Order/Create")); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_DenyServiceBlocks(t *testing.T) {
	f := tracefencer.New(tracefencer.Policy{
		DenyServices: []string{"payments"},
	})
	if err := f.Allow(makeSpan("payments", "/Pay/Submit")); err != tracefencer.ErrDenied {
		t.Fatalf("expected ErrDenied, got %v", err)
	}
}

func TestAllow_DenyService_CaseInsensitive(t *testing.T) {
	f := tracefencer.New(tracefencer.Policy{
		DenyServices: []string{"Payments"},
	})
	if err := f.Allow(makeSpan("PAYMENTS", "/Pay/Submit")); err != tracefencer.ErrDenied {
		t.Fatalf("expected ErrDenied, got %v", err)
	}
}

func TestAllow_AllowListPermitsMatchingService(t *testing.T) {
	f := tracefencer.New(tracefencer.Policy{
		AllowServices: []string{"orders", "inventory"},
	})
	if err := f.Allow(makeSpan("orders", "/Order/Get")); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_AllowListBlocksUnlistedService(t *testing.T) {
	f := tracefencer.New(tracefencer.Policy{
		AllowServices: []string{"orders"},
	})
	if err := f.Allow(makeSpan("payments", "/Pay/Submit")); err != tracefencer.ErrDenied {
		t.Fatalf("expected ErrDenied, got %v", err)
	}
}

func TestAllow_DenyMethodSubstring(t *testing.T) {
	f := tracefencer.New(tracefencer.Policy{
		DenyMethods: []string{"internal"},
	})
	if err := f.Allow(makeSpan("admin", "/Admin/InternalDebug")); err != tracefencer.ErrDenied {
		t.Fatalf("expected ErrDenied, got %v", err)
	}
}

func TestFilterSpans_RemovesDenied(t *testing.T) {
	f := tracefencer.New(tracefencer.Policy{
		DenyServices: []string{"payments"},
	})
	spans := []storage.Span{
		makeSpan("orders", "/Order/Get"),
		makeSpan("payments", "/Pay/Submit"),
		makeSpan("inventory", "/Inv/List"),
	}
	result := f.FilterSpans(spans)
	if len(result) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(result))
	}
	for _, s := range result {
		if s.ServiceName == "payments" {
			t.Fatal("payments span should have been filtered out")
		}
	}
}
