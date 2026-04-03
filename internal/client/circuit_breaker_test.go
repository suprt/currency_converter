package client

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_Success(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)

	err := cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if cb.GetState() != stateClosed {
		t.Errorf("expected 0 (CLOSED), got %d", cb.GetState())
	}
}

func TestCircuitBreaker_OpenAfterFailures(t *testing.T) {
	cb := NewCircuitBreaker(1, time.Second)
	err := cb.Execute(func() error {
		return errors.New("some error")
	})
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if cb.GetState() != stateOpen {
		t.Errorf("expected 1 (OPEN), got %d", cb.GetState())
	}
}

func TestCircuitBreaker_BlockWhenOpen(t *testing.T) {
	cb := NewCircuitBreaker(3, 5*time.Second)
	cb.state = stateOpen
	cb.lastFailureTime = time.Now()
	err := cb.Execute(func() error {
		return nil
	})
	if err != nil {
		if !errors.Is(err, errCircuitOpen) {
			t.Errorf("unexpected error: %s", err)
		}
	} else {
		t.Errorf("expected error, got nil")
	}
}

func TestCircuitBreaker_HalfOpenAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)
	cb.state = stateOpen
	cb.failureCount = 1
	cb.lastFailureTime = time.Now()
	time.Sleep(200 * time.Millisecond)
	err := cb.Execute(func() error {
		return errors.New("some error")
	})
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if cb.GetState() != stateHalfOpen {
		t.Errorf("expected 2 (Half-Open), got %d", cb.GetState())
	}
}

func TestCircuitBreaker_CloseAfterSuccess(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)
	cb.state = stateHalfOpen
	err := cb.Execute(func() error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if cb.GetState() != stateClosed {
		t.Errorf("expected 0 (CLOSED), got %d", cb.GetState())
	}
}

func TestCircuitBreaker_OpenOnFailureInHalfOpen(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)
	cb.state = stateHalfOpen
	cb.failureCount = 2
	err := cb.Execute(func() error {
		return errors.New("some error")
	})
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	if cb.GetState() != stateOpen {
		t.Errorf("expected 1 (OPEN), got %d", cb.GetState())
	}
}
