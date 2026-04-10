package spanlinker_test

import (
	"testing"
	"time"

	"github.com/grpc-tracer/internal/spanlinker"
	"github.com/grpc-tracer/internal/storage"
)

func makeLink(fromTrace, fromSpan, toTrace, toSpan string, kind spanlinker.LinkKind) spanlinker.SpanLink {
	return spanlinker.SpanLink{
		FromTraceID: fromTrace,
		FromSpanID:  fromSpan,
		ToTraceID:   toTrace,
		ToSpanID:    toSpan,
		Kind:        kind,
	}
}

func TestAdd_ValidLink(t *testing.T) {
	l := spanlinker.New()
	err := l.Add(makeLink("t1", "s1", "t2", "s2", spanlinker.LinkKindChildOf))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := len(l.All()); got != 1 {
		t.Fatalf("expected 1 link, got %d", got)
	}
}

func TestAdd_MissingFromID_ReturnsError(t *testing.T) {
	l := spanlinker.New()
	err := l.Add(makeLink("", "", "t2", "s2", spanlinker.LinkKindChildOf))
	if err == nil {
		t.Fatal("expected error for empty from IDs")
	}
}

func TestAdd_MissingToID_ReturnsError(t *testing.T) {
	l := spanlinker.New()
	err := l.Add(makeLink("t1", "s1", "", "", spanlinker.LinkKindFollowsFrom))
	if err == nil {
		t.Fatal("expected error for empty to IDs")
	}
}

func TestAdd_DefaultsKindToChildOf(t *testing.T) {
	l := spanlinker.New()
	link := spanlinker.SpanLink{FromTraceID: "t1", FromSpanID: "s1", ToTraceID: "t2", ToSpanID: "s2"}
	_ = l.Add(link)
	if got := l.All()[0].Kind; got != spanlinker.LinkKindChildOf {
		t.Fatalf("expected kind child_of, got %q", got)
	}
}

func TestLinksFrom_ReturnsMatchingLinks(t *testing.T) {
	l := spanlinker.New()
	_ = l.Add(makeLink("t1", "s1", "t2", "s2", spanlinker.LinkKindChildOf))
	_ = l.Add(makeLink("t1", "s1", "t3", "s3", spanlinker.LinkKindFollowsFrom))
	_ = l.Add(makeLink("t9", "s9", "t2", "s2", spanlinker.LinkKindChildOf))

	links := l.LinksFrom("t1", "s1")
	if len(links) != 2 {
		t.Fatalf("expected 2 links from t1/s1, got %d", len(links))
	}
}

func TestLinksTo_ReturnsMatchingLinks(t *testing.T) {
	l := spanlinker.New()
	_ = l.Add(makeLink("t1", "s1", "t2", "s2", spanlinker.LinkKindChildOf))
	_ = l.Add(makeLink("t3", "s3", "t2", "s2", spanlinker.LinkKindFollowsFrom))

	links := l.LinksTo("t2", "s2")
	if len(links) != 2 {
		t.Fatalf("expected 2 links to t2/s2, got %d", len(links))
	}
}

func TestResolveLinkedSpans_ReturnsSpansFromStore(t *testing.T) {
	store := storage.NewTraceStore()
	store.AddSpan(storage.Span{
		TraceID: "t2", SpanID: "s2", Service: "svc-b",
		StartTime: time.Now(), EndTime: time.Now(),
	})

	l := spanlinker.New()
	_ = l.Add(makeLink("t1", "s1", "t2", "s2", spanlinker.LinkKindChildOf))

	spans := l.ResolveLinkedSpans(store, "t1", "s1")
	if len(spans) != 1 {
		t.Fatalf("expected 1 resolved span, got %d", len(spans))
	}
	if spans[0].SpanID != "s2" {
		t.Errorf("expected span s2, got %q", spans[0].SpanID)
	}
}

func TestResolveLinkedSpans_MissingTrace_Skipped(t *testing.T) {
	store := storage.NewTraceStore()
	l := spanlinker.New()
	_ = l.Add(makeLink("t1", "s1", "t_missing", "s_missing", spanlinker.LinkKindChildOf))

	spans := l.ResolveLinkedSpans(store, "t1", "s1")
	if len(spans) != 0 {
		t.Fatalf("expected 0 spans for missing trace, got %d", len(spans))
	}
}
