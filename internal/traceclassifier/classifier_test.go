package traceclassifier_test

import (
	"testing"
	"time"

	"github.com/user/grpc-tracer/internal/storage"
	"github.com/user/grpc-tracer/internal/traceclassifier"
)

func makeSpan(service, errMsg string, dur time.Duration, tags map[string]string) storage.Span {
	if tags == nil {
		tags = map[string]string{}
	}
	return storage.Span{
		ServiceName: service,
		Error:       errMsg,
		Duration:    dur,
		Tags:        tags,
	}
}

func TestClassify_ErrorSpan(t *testing.T) {
	c := traceclassifier.New(traceclassifier.DefaultRules(500 * time.Millisecond)...)
	span := makeSpan("svc", "rpc error", 100*time.Millisecond, nil)
	labels := c.Classify(span)
	if !containsLabel(labels, traceclassifier.LabelError) {
		t.Errorf("expected LabelError, got %v", labels)
	}
}

func TestClassify_SlowSpan(t *testing.T) {
	c := traceclassifier.New(traceclassifier.DefaultRules(500 * time.Millisecond)...)
	span := makeSpan("svc", "", 600*time.Millisecond, nil)
	labels := c.Classify(span)
	if !containsLabel(labels, traceclassifier.LabelSlow) {
		t.Errorf("expected LabelSlow, got %v", labels)
	}
	if containsLabel(labels, traceclassifier.LabelFast) {
		t.Error("fast label should not be set for slow span")
	}
}

func TestClassify_FastSpan(t *testing.T) {
	c := traceclassifier.New(traceclassifier.DefaultRules(500 * time.Millisecond)...)
	span := makeSpan("svc", "", 50*time.Millisecond, nil)
	labels := c.Classify(span)
	if !containsLabel(labels, traceclassifier.LabelFast) {
		t.Errorf("expected LabelFast, got %v", labels)
	}
}

func TestClassify_InternalService(t *testing.T) {
	c := traceclassifier.New(traceclassifier.DefaultRules(500 * time.Millisecond)...)
	span := makeSpan("auth-internal", "", 10*time.Millisecond, nil)
	labels := c.Classify(span)
	if !containsLabel(labels, traceclassifier.LabelInternal) {
		t.Errorf("expected LabelInternal, got %v", labels)
	}
}

func TestClassify_ExternalSpan(t *testing.T) {
	c := traceclassifier.New(traceclassifier.DefaultRules(500 * time.Millisecond)...)
	span := makeSpan("gateway", "", 20*time.Millisecond, map[string]string{"span.kind": "client"})
	labels := c.Classify(span)
	if !containsLabel(labels, traceclassifier.LabelExternal) {
		t.Errorf("expected LabelExternal, got %v", labels)
	}
}

func TestClassify_CustomRule(t *testing.T) {
	rule := traceclassifier.Rule{
		Name:  "database",
		Label: "database",
		Match: func(s storage.Span) bool { return s.Tags["db.type"] != "" },
	}
	c := traceclassifier.New(rule)
	span := makeSpan("svc", "", 5*time.Millisecond, map[string]string{"db.type": "postgres"})
	labels := c.Classify(span)
	if !containsLabel(labels, "database") {
		t.Errorf("expected database label, got %v", labels)
	}
}

func TestClassify_NoDuplicateLabels(t *testing.T) {
	rule := traceclassifier.Rule{
		Name: "dup1", Label: traceclassifier.LabelError,
		Match: func(s storage.Span) bool { return true },
	}
	rule2 := traceclassifier.Rule{
		Name: "dup2", Label: traceclassifier.LabelError,
		Match: func(s storage.Span) bool { return true },
	}
	c := traceclassifier.New(rule, rule2)
	span := makeSpan("svc", "", 0, nil)
	labels := c.Classify(span)
	count := 0
	for _, l := range labels {
		if l == traceclassifier.LabelError {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 LabelError, got %d", count)
	}
}

func containsLabel(labels []traceclassifier.Label, target traceclassifier.Label) bool {
	for _, l := range labels {
		if l == target {
			return true
		}
	}
	return false
}
