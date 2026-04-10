package traceannotator_test

import (
	"testing"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/traceannotator"
)

func makeStore(t *testing.T) *storage.TraceStore {
	t.Helper()
	return storage.NewTraceStore()
}

func addSpan(store *storage.TraceStore, traceID, spanID, service string) {
	store.AddSpan(storage.Span{
		TraceID: traceID,
		SpanID:  spanID,
		Service: service,
		Tags:    map[string]string{},
	})
}

func TestAnnotate_Success(t *testing.T) {
	store := makeStore(t)
	addSpan(store, "t1", "s1", "svc")

	an := traceannotator.New(store)
	if err := an.Annotate("t1", "s1", "env", "prod"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	spans := store.GetTrace("t1")
	if spans[0].Tags["env"] != "prod" {
		t.Errorf("expected tag env=prod, got %q", spans[0].Tags["env"])
	}
}

func TestAnnotate_TraceNotFound(t *testing.T) {
	store := makeStore(t)
	an := traceannotator.New(store)

	err := an.Annotate("missing", "s1", "key", "val")
	if err == nil {
		t.Fatal("expected error for missing trace, got nil")
	}
}

func TestAnnotate_SpanNotFound(t *testing.T) {
	store := makeStore(t)
	addSpan(store, "t1", "s1", "svc")
	an := traceannotator.New(store)

	err := an.Annotate("t1", "no-such-span", "key", "val")
	if err == nil {
		t.Fatal("expected error for missing span, got nil")
	}
}

func TestAnnotate_EmptyKey(t *testing.T) {
	store := makeStore(t)
	addSpan(store, "t1", "s1", "svc")
	an := traceannotator.New(store)

	if err := an.Annotate("t1", "s1", "", "val"); err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestAnnotateAll_Success(t *testing.T) {
	store := makeStore(t)
	addSpan(store, "t1", "s1", "svcA")
	addSpan(store, "t1", "s2", "svcB")
	an := traceannotator.New(store)

	if err := an.AnnotateAll("t1", "region", "eu-west"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, sp := range store.GetTrace("t1") {
		if sp.Tags["region"] != "eu-west" {
			t.Errorf("span %s missing annotation, got %q", sp.SpanID, sp.Tags["region"])
		}
	}
}

func TestAnnotateAll_TraceNotFound(t *testing.T) {
	store := makeStore(t)
	an := traceannotator.New(store)

	if err := an.AnnotateAll("ghost", "k", "v"); err == nil {
		t.Fatal("expected error for missing trace")
	}
}
