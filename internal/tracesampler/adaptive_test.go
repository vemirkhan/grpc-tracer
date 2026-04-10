package tracesampler

import (
	"testing"
	"time"
)

func TestNew_AppliesDefaults(t *testing.T) {
	s := New(Config{})
	if s.cfg.BaseRate != 0.1 {
		t.Errorf("expected BaseRate 0.1, got %v", s.cfg.BaseRate)
	}
	if s.cfg.ErrorThreshold != 0.2 {
		t.Errorf("expected ErrorThreshold 0.2, got %v", s.cfg.ErrorThreshold)
	}
	if s.cfg.Window != 30*time.Second {
		t.Errorf("expected Window 30s, got %v", s.cfg.Window)
	}
}

func TestNew_CustomConfig(t *testing.T) {
	s := New(Config{BaseRate: 0.5, ErrorThreshold: 0.3, Window: 10 * time.Second})
	if s.cfg.BaseRate != 0.5 {
		t.Errorf("expected BaseRate 0.5, got %v", s.cfg.BaseRate)
	}
}

func TestErrorRate_NoObservations(t *testing.T) {
	s := New(Config{})
	if r := s.ErrorRate(); r != 0.0 {
		t.Errorf("expected 0.0, got %v", r)
	}
}

func TestErrorRate_AllErrors(t *testing.T) {
	s := New(Config{Window: 5 * time.Second})
	s.Record(true)
	s.Record(true)
	s.Record(true)
	if r := s.ErrorRate(); r != 1.0 {
		t.Errorf("expected 1.0, got %v", r)
	}
}

func TestErrorRate_Mixed(t *testing.T) {
	s := New(Config{Window: 5 * time.Second})
	s.Record(false)
	s.Record(false)
	s.Record(true)
	s.Record(true)
	r := s.ErrorRate()
	if r != 0.5 {
		t.Errorf("expected 0.5, got %v", r)
	}
}

func TestShouldSample_BoostsOnHighErrorRate(t *testing.T) {
	s := New(Config{
		BaseRate:       0.01,
		ErrorThreshold: 0.5,
		Window:         5 * time.Second,
	})
	// Record enough errors to exceed threshold.
	for i := 0; i < 10; i++ {
		s.Record(true)
	}
	// All requests should be sampled when error rate is 100%.
	for i := 0; i < 20; i++ {
		if !s.ShouldSample() {
			t.Error("expected ShouldSample to return true under high error rate")
		}
	}
}

func TestEviction_RemovesStaleObservations(t *testing.T) {
	s := New(Config{Window: 50 * time.Millisecond})
	s.Record(true)
	s.Record(true)
	if r := s.ErrorRate(); r != 1.0 {
		t.Fatalf("expected 1.0 before eviction, got %v", r)
	}
	time.Sleep(60 * time.Millisecond)
	if r := s.ErrorRate(); r != 0.0 {
		t.Errorf("expected 0.0 after eviction, got %v", r)
	}
}
