package middleware

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(nil)
	if cb == nil {
		t.Fatal("NewCircuitBreaker should not return nil")
	}
	
	if cb.GetState() != StateClosed {
		t.Errorf("Expected initial state to be CLOSED, got %s", cb.GetState().String())
	}
	
	if cb.GetFailures() != 0 {
		t.Errorf("Expected initial failures to be 0, got %d", cb.GetFailures())
	}
}

func TestCircuitBreakerWithCustomConfig(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures:    3,
		Timeout:        10 * time.Second,
		ResetTimeout:   20 * time.Second,
		MonitorTimeout: 5 * time.Second,
	}
	
	cb := NewCircuitBreaker(config)
	if cb == nil {
		t.Fatal("NewCircuitBreaker should not return nil")
	}
	
	stats := cb.GetStats()
	if stats["max_failures"] != 3 {
		t.Errorf("Expected max_failures to be 3, got %v", stats["max_failures"])
	}
}

func TestCircuitBreakerExecuteSuccess(t *testing.T) {
	cb := NewCircuitBreaker(nil)
	
	err := cb.Execute(context.Background(), func() error {
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to remain CLOSED, got %s", cb.GetState().String())
	}
	
	if cb.GetFailures() != 0 {
		t.Errorf("Expected failures to remain 0, got %d", cb.GetFailures())
	}
}

func TestCircuitBreakerExecuteFailure(t *testing.T) {
	cb := NewCircuitBreaker(nil)
	testError := errors.New("test error")
	
	err := cb.Execute(context.Background(), func() error {
		return testError
	})
	
	if err != testError {
		t.Errorf("Expected test error, got %v", err)
	}
	
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to remain CLOSED after single failure, got %s", cb.GetState().String())
	}
	
	if cb.GetFailures() != 1 {
		t.Errorf("Expected failures to be 1, got %d", cb.GetFailures())
	}
}

func TestCircuitBreakerStateTransition(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures:  2,
		ResetTimeout: 100 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)
	
	// First failure
	cb.Execute(context.Background(), func() error {
		return errors.New("error 1")
	})
	
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be CLOSED after 1 failure, got %s", cb.GetState().String())
	}
	
	// Second failure - should open circuit
	cb.Execute(context.Background(), func() error {
		return errors.New("error 2")
	})
	
	if cb.GetState() != StateOpen {
		t.Errorf("Expected state to be OPEN after 2 failures, got %s", cb.GetState().String())
	}
	
	// Circuit should not be ready when open
	if cb.Ready() {
		t.Error("Circuit should not be ready when open")
	}
}

func TestCircuitBreakerHalfOpenState(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures:  1,
		ResetTimeout: 10 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)
	
	// Trigger circuit to open
	cb.Execute(context.Background(), func() error {
		return errors.New("error")
	})
	
	if cb.GetState() != StateOpen {
		t.Errorf("Expected state to be OPEN, got %s", cb.GetState().String())
	}
	
	// Wait for reset timeout
	time.Sleep(20 * time.Millisecond)
	
	// Call Ready() to trigger state transition to half-open
	if !cb.Ready() {
		t.Error("Circuit should be ready after reset timeout")
	}
	
	// Circuit should now be in half-open state
	if cb.GetState() != StateHalfOpen {
		t.Errorf("Expected state to be HALF_OPEN, got %s", cb.GetState().String())
	}
	
	// Circuit should be ready in half-open state
	if !cb.Ready() {
		t.Error("Circuit should be ready in half-open state")
	}
}

func TestCircuitBreakerRecovery(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures:  1,
		ResetTimeout: 10 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)
	
	// Trigger circuit to open
	cb.Execute(context.Background(), func() error {
		return errors.New("error")
	})
	
	// Wait for reset timeout
	time.Sleep(20 * time.Millisecond)
	
	// Success in half-open state should close circuit
	err := cb.Execute(context.Background(), func() error {
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be CLOSED after success, got %s", cb.GetState().String())
	}
	
	if cb.GetFailures() != 0 {
		t.Errorf("Expected failures to be reset to 0, got %d", cb.GetFailures())
	}
}

func TestCircuitBreakerExecuteWithResult(t *testing.T) {
	cb := NewCircuitBreaker(nil)
	
	result, err := cb.ExecuteWithResult(context.Background(), func() (interface{}, error) {
		return "success", nil
	})
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if result != "success" {
		t.Errorf("Expected result 'success', got %v", result)
	}
}

func TestCircuitBreakerForceOpen(t *testing.T) {
	cb := NewCircuitBreaker(nil)
	
	cb.ForceOpen()
	
	if cb.GetState() != StateOpen {
		t.Errorf("Expected state to be OPEN after ForceOpen, got %s", cb.GetState().String())
	}
	
	if cb.Ready() {
		t.Error("Circuit should not be ready after ForceOpen")
	}
}

func TestCircuitBreakerForceClose(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures: 1,
	}
	cb := NewCircuitBreaker(config)
	
	// Trigger circuit to open
	cb.Execute(context.Background(), func() error {
		return errors.New("error")
	})
	
	cb.ForceClose()
	
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be CLOSED after ForceClose, got %s", cb.GetState().String())
	}
	
	if cb.GetFailures() != 0 {
		t.Errorf("Expected failures to be 0 after ForceClose, got %d", cb.GetFailures())
	}
	
	if !cb.Ready() {
		t.Error("Circuit should be ready after ForceClose")
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	config := &CircuitBreakerConfig{
		MaxFailures: 1,
	}
	cb := NewCircuitBreaker(config)
	
	// Trigger circuit to open
	cb.Execute(context.Background(), func() error {
		return errors.New("error")
	})
	
	cb.Reset()
	
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state to be CLOSED after Reset, got %s", cb.GetState().String())
	}
	
	if cb.GetFailures() != 0 {
		t.Errorf("Expected failures to be 0 after Reset, got %d", cb.GetFailures())
	}
	
	if cb.GetLastError() != nil {
		t.Error("Expected last error to be nil after Reset")
	}
	
	if !cb.Ready() {
		t.Error("Circuit should be ready after Reset")
	}
}

func TestCircuitBreakerGetStats(t *testing.T) {
	cb := NewCircuitBreaker(nil)
	
	stats := cb.GetStats()
	
	requiredKeys := []string{
		"state", "failures", "max_failures", "last_error",
		"last_failure", "ready", "timeout", "reset_timeout",
	}
	
	for _, key := range requiredKeys {
		if _, exists := stats[key]; !exists {
			t.Errorf("Stats missing required key: %s", key)
		}
	}
	
	if stats["state"] != "CLOSED" {
		t.Errorf("Expected state to be 'CLOSED', got %v", stats["state"])
	}
	
	if stats["failures"] != 0 {
		t.Errorf("Expected failures to be 0, got %v", stats["failures"])
	}
	
	if stats["ready"] != true {
		t.Errorf("Expected ready to be true, got %v", stats["ready"])
	}
}

func TestCircuitBreakerStateString(t *testing.T) {
	testCases := []struct {
		state    CircuitBreakerState
		expected string
	}{
		{StateClosed, "CLOSED"},
		{StateOpen, "OPEN"},
		{StateHalfOpen, "HALF_OPEN"},
	}
	
	for _, tc := range testCases {
		if tc.state.String() != tc.expected {
			t.Errorf("Expected state %d to stringify to '%s', got '%s'", tc.state, tc.expected, tc.state.String())
		}
	}
}

func TestDefaultCircuitBreakerConfig(t *testing.T) {
	config := DefaultCircuitBreakerConfig()
	
	if config.MaxFailures != 5 {
		t.Errorf("Expected MaxFailures to be 5, got %d", config.MaxFailures)
	}
	
	if config.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout to be 30s, got %v", config.Timeout)
	}
	
	if config.ResetTimeout != 60*time.Second {
		t.Errorf("Expected ResetTimeout to be 60s, got %v", config.ResetTimeout)
	}
	
	if config.MonitorTimeout != 10*time.Second {
		t.Errorf("Expected MonitorTimeout to be 10s, got %v", config.MonitorTimeout)
	}
} 