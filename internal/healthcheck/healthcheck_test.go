package healthcheck

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("expected non-nil Checker")
	}
	if len(c.All()) != 0 {
		t.Fatal("expected empty checker")
	}
}

func TestRecord_And_Get(t *testing.T) {
	c := New()
	before := time.Now()
	c.Record("svc-a", StatusHealthy, "all good")

	h, ok := c.Get("svc-a")
	if !ok {
		t.Fatal("expected to find svc-a")
	}
	if h.Status != StatusHealthy {
		t.Errorf("expected healthy, got %s", h.Status)
	}
	if h.Message != "all good" {
		t.Errorf("unexpected message: %s", h.Message)
	}
	if h.LastChecked.Before(before) {
		t.Error("LastChecked should be after test start")
	}
}

func TestGet_Unknown(t *testing.T) {
	c := New()
	_, ok := c.Get("ghost")
	if ok {
		t.Fatal("expected false for unknown service")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	c := New()
	c.Record("svc-a", StatusHealthy, "")
	c.Record("svc-b", StatusDegraded, "high latency")

	all := c.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestIsHealthy_AllHealthy(t *testing.T) {
	c := New()
	c.Record("svc-a", StatusHealthy, "")
	c.Record("svc-b", StatusHealthy, "")
	if !c.IsHealthy() {
		t.Error("expected IsHealthy to return true")
	}
}

func TestIsHealthy_OneDegraded(t *testing.T) {
	c := New()
	c.Record("svc-a", StatusHealthy, "")
	c.Record("svc-b", StatusDegraded, "slow")
	if c.IsHealthy() {
		t.Error("expected IsHealthy to return false when one service is degraded")
	}
}

func TestRecord_Overwrite(t *testing.T) {
	c := New()
	c.Record("svc-a", StatusHealthy, "fine")
	c.Record("svc-a", StatusUnhealthy, "crashed")

	h, _ := c.Get("svc-a")
	if h.Status != StatusUnhealthy {
		t.Errorf("expected unhealthy after overwrite, got %s", h.Status)
	}
	if h.Message != "crashed" {
		t.Errorf("unexpected message: %s", h.Message)
	}
}
