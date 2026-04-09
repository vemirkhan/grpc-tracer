package sampler_test

import (
	"testing"

	"github.com/user/grpc-tracer/internal/sampler"
)

func TestNew_Defaults(t *testing.T) {
	s := sampler.New(sampler.AlwaysSample, 0.5)
	if s.Strategy() != sampler.AlwaysSample {
		t.Errorf("expected AlwaysSample, got %s", s.Strategy())
	}
	if s.Rate() != 0.5 {
		t.Errorf("expected rate 0.5, got %f", s.Rate())
	}
}

func TestNew_ClampRate(t *testing.T) {
	s := sampler.New(sampler.ProbabilisticSample, 1.5)
	if s.Rate() != 1.0 {
		t.Errorf("rate should be clamped to 1.0, got %f", s.Rate())
	}
	s2 := sampler.New(sampler.ProbabilisticSample, -0.3)
	if s2.Rate() != 0.0 {
		t.Errorf("rate should be clamped to 0.0, got %f", s2.Rate())
	}
}

func TestShouldSample_Always(t *testing.T) {
	s := sampler.New(sampler.AlwaysSample, 1.0)
	for i := 0; i < 20; i++ {
		if !s.ShouldSample("trace-id") {
			t.Error("AlwaysSample should always return true")
		}
	}
}

func TestShouldSample_Never(t *testing.T) {
	s := sampler.New(sampler.NeverSample, 0.0)
	for i := 0; i < 20; i++ {
		if s.ShouldSample("trace-id") {
			t.Error("NeverSample should always return false")
		}
	}
}

func TestShouldSample_Probabilistic_FullRate(t *testing.T) {
	s := sampler.New(sampler.ProbabilisticSample, 1.0)
	for i := 0; i < 50; i++ {
		if !s.ShouldSample("trace-id") {
			t.Error("rate=1.0 should always sample")
		}
	}
}

func TestShouldSample_Probabilistic_ZeroRate(t *testing.T) {
	s := sampler.New(sampler.ProbabilisticSample, 0.0)
	for i := 0; i < 50; i++ {
		if s.ShouldSample("trace-id") {
			t.Error("rate=0.0 should never sample")
		}
	}
}

func TestShouldSample_Probabilistic_Approximate(t *testing.T) {
	s := sampler.New(sampler.ProbabilisticSample, 0.5)
	sampled := 0
	total := 1000
	for i := 0; i < total; i++ {
		if s.ShouldSample("trace-id") {
			sampled++
		}
	}
	// With seed=42 and rate=0.5, expect roughly 50% ± 10%
	if sampled < 300 || sampled > 700 {
		t.Errorf("expected ~50%% sampled, got %d/%d", sampled, total)
	}
}
