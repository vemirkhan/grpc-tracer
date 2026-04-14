package traceprojector_test

import (
	"testing"
	"time"

	"github.com/example/grpc-tracer/internal/storage"
	"github.com/example/grpc-tracer/internal/traceprojector"
)

func makeStore() *storage.TraceStore {
	return storage.NewTraceStore()
}

func addSpan(s *storage.TraceStore, traceID, spanID, service, method string, dur time.Duration, hasErr bool) {
	span := storage.Span{
		TraceID:     traceID,
		SpanID:      spanID,
		ServiceName: service,
		Method:      method,
		Duration:    dur,
		Error:       hasErr,
		Tags:        map[string]string{"env": "test"},
	}
	s.AddSpan(span)
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := traceprojector.New(nil, traceprojector.FieldService)
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_NoFieldsReturnsError(t *testing.T) {
	st := makeStore()
	_, err := traceprojector.New(st)
	if err == nil {
		t.Fatal("expected error when no fields specified")
	}
}

func TestNew_ValidCreation(t *testing.T) {
	st := makeStore()
	p, err := traceprojector.New(st, traceprojector.FieldService, traceprojector.FieldMethod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil projector")
	}
}

func TestProjectTrace_NonExistent(t *testing.T) {
	st := makeStore()
	p, _ := traceprojector.New(st, traceprojector.FieldService)
	_, err := p.ProjectTrace("missing-trace")
	if err == nil {
		t.Fatal("expected error for non-existent trace")
	}
}

func TestProjectTrace_SelectedFields(t *testing.T) {
	st := makeStore()
	addSpan(st, "t1", "s1", "auth", "/Login", 10*time.Millisecond, false)
	addSpan(st, "t1", "s2", "auth", "/Logout", 5*time.Millisecond, true)

	p, _ := traceprojector.New(st, traceprojector.FieldService, traceprojector.FieldError)
	projs, err := p.ProjectTrace("t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projs) != 2 {
		t.Fatalf("expected 2 projections, got %d", len(projs))
	}
	for _, proj := range projs {
		if _, ok := proj[traceprojector.FieldService]; !ok {
			t.Error("expected FieldService to be present")
		}
		if _, ok := proj[traceprojector.FieldError]; !ok {
			t.Error("expected FieldError to be present")
		}
		if _, ok := proj[traceprojector.FieldMethod]; ok {
			t.Error("expected FieldMethod to be absent")
		}
	}
}

func TestProjectTrace_AllFields(t *testing.T) {
	st := makeStore()
	addSpan(st, "t2", "s1", "order", "/Place", 20*time.Millisecond, false)

	p, _ := traceprojector.New(st,
		traceprojector.FieldTraceID, traceprojector.FieldSpanID,
		traceprojector.FieldService, traceprojector.FieldMethod,
		traceprojector.FieldDuration, traceprojector.FieldError,
		traceprojector.FieldTags,
	)
	projs, err := p.ProjectTrace("t2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projs) != 1 {
		t.Fatalf("expected 1 projection, got %d", len(projs))
	}
	proj := projs[0]
	for _, f := range []traceprojector.Field{
		traceprojector.FieldTraceID, traceprojector.FieldSpanID,
		traceprojector.FieldService, traceprojector.FieldMethod,
		traceprojector.FieldDuration, traceprojector.FieldError,
		traceprojector.FieldTags,
	} {
		if _, ok := proj[f]; !ok {
			t.Errorf("expected field %q to be present", f)
		}
	}
}
