package tagselector_test

import (
	"testing"
	"time"

	"github.com/your-org/grpc-tracer/internal/storage"
	"github.com/your-org/grpc-tracer/internal/tagselector"
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

func TestMatchSpan_EmptyCriteria(t *testing.T) {
	s := makeSpan(map[string]string{"env": "prod"})
	if !tagselector.MatchSpan(s, tagselector.Criteria{}) {
		t.Fatal("empty criteria should match any span")
	}
}

func TestMatchSpan_AllTagsPresent(t *testing.T) {
	s := makeSpan(map[string]string{"env": "prod", "region": "us-east"})
	c := tagselector.Criteria{Tags: map[string]string{"env": "prod", "region": "us-east"}}
	if !tagselector.MatchSpan(s, c) {
		t.Fatal("expected match when all tags are present")
	}
}

func TestMatchSpan_MissingTag(t *testing.T) {
	s := makeSpan(map[string]string{"env": "prod"})
	c := tagselector.Criteria{Tags: map[string]string{"env": "prod", "region": "us-east"}}
	if tagselector.MatchSpan(s, c) {
		t.Fatal("expected no match when a required tag is missing")
	}
}

func TestMatchSpan_WrongValue(t *testing.T) {
	s := makeSpan(map[string]string{"env": "staging"})
	c := tagselector.Criteria{Tags: map[string]string{"env": "prod"}}
	if tagselector.MatchSpan(s, c) {
		t.Fatal("expected no match when tag value differs")
	}
}

func TestSelectSpans_FiltersCorrectly(t *testing.T) {
	spans := []storage.Span{
		makeSpan(map[string]string{"env": "prod"}),
		makeSpan(map[string]string{"env": "staging"}),
		makeSpan(map[string]string{"env": "prod", "region": "eu"}),
	}
	c := tagselector.Criteria{Tags: map[string]string{"env": "prod"}}
	got := tagselector.SelectSpans(spans, c)
	if len(got) != 2 {
		t.Fatalf("expected 2 spans, got %d", len(got))
	}
}

func TestSelectTraces_ExcludesNonMatchingTraces(t *testing.T) {
	traces := map[string][]storage.Span{
		"trace-a": {makeSpan(map[string]string{"env": "prod"})},
		"trace-b": {makeSpan(map[string]string{"env": "staging"})},
	}
	c := tagselector.Criteria{Tags: map[string]string{"env": "prod"}}
	got := tagselector.SelectTraces(traces, c)
	if len(got) != 1 {
		t.Fatalf("expected 1 trace, got %d", len(got))
	}
	if _, ok := got["trace-a"]; !ok {
		t.Fatal("expected trace-a to be selected")
	}
}

func TestSelectTraces_EmptyCriteriaReturnsAll(t *testing.T) {
	traces := map[string][]storage.Span{
		"t1": {makeSpan(nil)},
		"t2": {makeSpan(nil)},
	}
	got := tagselector.SelectTraces(traces, tagselector.Criteria{})
	if len(got) != 2 {
		t.Fatalf("expected 2 traces, got %d", len(got))
	}
}
