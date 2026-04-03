package client

import (
	"errors"
	"sync"
	"time"
)

// Circuit breaker states
const (
	stateClosed = iota
	stateOpen
	stateHalfOpen
)

type CircuitBreaker struct {
	mu               sync.Mutex
	state            int
	failureCount     int
	failureThreshold int
	timeout          time.Duration
	lastFailureTime  time.Time
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            stateClosed,
		failureThreshold: threshold,
		timeout:          timeout,
	}
}

var errCircuitOpen = errors.New("circuit breaker is open")

func (cb *CircuitBreaker) Execute(fn func() error) error {

	cb.mu.Lock()

	if cb.state == stateOpen {
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = stateHalfOpen
		} else {
			cb.mu.Unlock()
			return errCircuitOpen
		}
	}
	cb.mu.Unlock()
	err := fn()
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		if cb.failureCount >= cb.failureThreshold {
			cb.state = stateOpen
			return errCircuitOpen
		}
		return err
	}

	cb.failureCount = 0
	cb.state = stateClosed
	cb.lastFailureTime = time.Time{}
	return nil
}

// GetState return the current state of the circuit breaker
//
// Returns:
//
//	0 - Closed: Normal operation, requests are allowed
//	1 - Open: Circuit is open, requests are blocked
//	2 - Half-Open: Testing if service recovered
func (cb *CircuitBreaker) GetState() int {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
