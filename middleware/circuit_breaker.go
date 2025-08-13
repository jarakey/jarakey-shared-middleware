package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitBreakerState represents the current state of the circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// String returns the string representation of the circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig holds the configuration for a circuit breaker
type CircuitBreakerConfig struct {
	MaxFailures    int           `json:"max_failures"`
	Timeout        time.Duration `json:"timeout"`
	ResetTimeout   time.Duration `json:"reset_timeout"`
	MonitorTimeout time.Duration `json:"monitor_timeout"`
}

// DefaultCircuitBreakerConfig returns a default configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		MaxFailures:    5,
		Timeout:        30 * time.Second,
		ResetTimeout:   60 * time.Second,
		MonitorTimeout: 10 * time.Second,
	}
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config     *CircuitBreakerConfig
	state      CircuitBreakerState
	failures   int
	lastError  error
	lastFailure time.Time
	mutex      sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}
	
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Execute runs the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	if !cb.Ready() {
		return fmt.Errorf("circuit breaker is %s", cb.state.String())
	}

	err := fn()
	cb.recordResult(err)
	return err
}

// ExecuteWithResult runs the given function with circuit breaker protection and returns a result
func (cb *CircuitBreaker) ExecuteWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	if !cb.Ready() {
		return nil, fmt.Errorf("circuit breaker is %s", cb.state.String())
	}

	result, err := fn()
	cb.recordResult(err)
	return result, err
}

// Ready checks if the circuit breaker is ready to execute requests
func (cb *CircuitBreaker) Ready() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Check if we need to transition from Open to HalfOpen
	if cb.state == StateOpen && time.Since(cb.lastFailure) >= cb.config.ResetTimeout {
		cb.state = StateHalfOpen
	}

	return cb.state != StateOpen
}

// recordResult records the result of an operation and updates the circuit breaker state
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.failures++
		cb.lastError = err
		cb.lastFailure = time.Now()

		if cb.failures >= cb.config.MaxFailures {
			cb.state = StateOpen
		}
	} else {
		// Success - reset failures and close circuit
		cb.failures = 0
		cb.lastError = nil
		if cb.state == StateHalfOpen {
			cb.state = StateClosed
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// GetFailures returns the current failure count
func (cb *CircuitBreaker) GetFailures() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.failures
}

// GetLastError returns the last error that occurred
func (cb *CircuitBreaker) GetLastError() error {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.lastError
}

// GetLastFailure returns the timestamp of the last failure
func (cb *CircuitBreaker) GetLastFailure() time.Time {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.lastFailure
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.state = StateOpen
	cb.lastFailure = time.Now()
}

// ForceClose forces the circuit breaker to closed state
func (cb *CircuitBreaker) ForceClose() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.lastError = nil
}

// Reset resets the circuit breaker to its initial state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.state = StateClosed
	cb.failures = 0
	cb.lastError = nil
	cb.lastFailure = time.Time{}
}

// GetStats returns statistics about the circuit breaker
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	// Calculate ready status without calling Ready() method
	ready := cb.state != StateOpen
	if cb.state == StateOpen && time.Since(cb.lastFailure) >= cb.config.ResetTimeout {
		ready = true
	}

	return map[string]interface{}{
		"state":         cb.state.String(),
		"failures":      cb.failures,
		"max_failures":  cb.config.MaxFailures,
		"last_error":    cb.lastError,
		"last_failure":  cb.lastFailure,
		"ready":         ready,
		"timeout":       cb.config.Timeout,
		"reset_timeout": cb.config.ResetTimeout,
	}
} 