package circuitbreaker

import (
	"testing"
	"time"
)

func TestNew_Defaults(t *testing.T) {
	cb := New(0, 0)
	if cb.maxFailures != 5 {
		t.Errorf("expected default maxFailures=5, got %d", cb.maxFailures)
	}
	if cb.resetTimeout != 10*time.Second {
		t.Errorf("expected default resetTimeout=10s, got %v", cb.resetTimeout)
	}
}

func TestNew_Custom(t *testing.T) {
	cb := New(3, 5*time.Second)
	if cb.maxFailures != 3 {
		t.Errorf("expected maxFailures=3, got %d", cb.maxFailures)
	}
}

func TestAllow_Closed(t *testing.T) {
	cb := New(3, time.Second)
	if err := cb.Allow(); err != nil {
		t.Errorf("expected nil error in closed state, got %v", err)
	}
}

func TestRecordFailure_OpensCircuit(t *testing.T) {
	cb := New(3, 10*time.Second)
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.CurrentState() != StateOpen {
		t.Errorf("expected StateOpen after %d failures", 3)
	}
	if err := cb.Allow(); err != ErrCircuitOpen {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestRecordSuccess_ResetsClosed(t *testing.T) {
	cb := New(2, 10*time.Second)
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.CurrentState() != StateOpen {
		t.Fatal("expected circuit to be open")
	}
	// Manually move to half-open to simulate timeout
	cb.mu.Lock()
	cb.state = StateHalfOpen
	cb.mu.Unlock()

	cb.RecordSuccess()
	if cb.CurrentState() != StateClosed {
		t.Errorf("expected StateClosed after success, got %v", cb.CurrentState())
	}
	if cb.failures != 0 {
		t.Errorf("expected failures reset to 0, got %d", cb.failures)
	}
}

func TestAllow_TransitionsToHalfOpen(t *testing.T) {
	cb := New(1, 50*time.Millisecond)
	cb.RecordFailure()
	if cb.CurrentState() != StateOpen {
		t.Fatal("expected StateOpen")
	}
	time.Sleep(60 * time.Millisecond)
	if err := cb.Allow(); err != nil {
		t.Errorf("expected nil after reset timeout, got %v", err)
	}
	if cb.CurrentState() != StateHalfOpen {
		t.Errorf("expected StateHalfOpen, got %v", cb.CurrentState())
	}
}
