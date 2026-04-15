package tracescorer_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/tracescorer"
)

func makeStore() storage.TraceStore {
	return storage.NewTraceStore()
}

func addSpan(store storage.TraceStore, traceID, spanID, service string, dur time.Duration, errMsg string) {
	store.AddSpan(storage.Span{
		TraceID:   traceID,
		SpanID:    spanID,
		Service:   service,
		StartTime: time.Now(),
		Duration:  dur,
		Error:     errMsg,
	})
}

func TestNew_NilStoreReturnsError(t *testing.T) {
	_, err := tracescorer.New(nil, tracescorer.Options{})
	if err == nil {
		t.Fatal("expected error for nil store")
	}
}

func TestNew_ValidStore(t *testing.T) {
	s, err := tracescorer.New(makeStore(), tracescorer.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil scorer")
	}
}

func TestScoreTrace_NotFound(t *testing.T) {
	s, _ := tracescorer.New(makeStore(), tracescorer.Options{})
	_, err := s.ScoreTrace("missing")
	if err == nil {
		t.Fatal("expected error for missing trace")
	}
}

func TestScoreTrace_HealthyTrace_ScoreIsOne(t *testing.T) {
	store := makeStore()
	addSpan(store, "t1", "s1", "svc", 100*time.Millisecond, "")
	addSpan(store, "t1", "s2", "svc", 50*time.Millisecond, "")

	scorer, _ := tracescorer.New(store, tracescorer.Options{})
	score, err := scorer.ScoreTrace("t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Value != 1.0 {
		t.Errorf("expected 1.0, got %f", score.Value)
	}
	if len(score.Penalties) != 0 {
		t.Errorf("expected no penalties, got %v", score.Penalties)
	}
}

func TestScoreTrace_ErrorSpan_ReducesScore(t *testing.T) {
	store := makeStore()
	addSpan(store, "t2", "s1", "svc", 10*time.Millisecond, "rpc error")

	scorer, _ := tracescorer.New(store, tracescorer.Options{ErrorWeight: 0.3})
	score, err := scorer.ScoreTrace("t2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if score.Value >= 1.0 {
		t.Errorf("expected reduced score, got %f", score.Value)
	}
	if !contains(score.Penalties, "errors_present") {
		t.Errorf("expected errors_present penalty, got %v", score.Penalties)
	}
}

func TestScoreTrace_SlowTrace_AppliesLatencyPenalty(t *testing.T) {
	store := makeStore()
	addSpan(store, "t3", "s1", "svc", 5*time.Second, "")

	scorer, _ := tracescorer.New(store, tracescorer.Options{MaxDuration: time.Second})
	score, _ := scorer.ScoreTrace("t3")
	if !contains(score.Penalties, "high_latency") {
		t.Errorf("expected high_latency penalty, got %v", score.Penalties)
	}
}

func TestScoreTrace_ScoreNeverNegative(t *testing.T) {
	store := makeStore()
	for i := 0; i < 10; i++ {
		addSpan(store, "t4", string(rune('a'+i)), "svc", 10*time.Second, "boom")
	}

	scorer, _ := tracescorer.New(store, tracescorer.Options{
		MaxDuration: time.Millisecond,
		MaxSpans:    1,
		ErrorWeight: 0.5,
	})
	score, _ := scorer.ScoreTrace("t4")
	if score.Value < 0 {
		t.Errorf("score must not be negative, got %f", score.Value)
	}
}

func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
