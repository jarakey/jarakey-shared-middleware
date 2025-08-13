package middleware

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	
	if config.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts to be 3, got %d", config.MaxAttempts)
	}
	
	if config.InitialDelay != 100*time.Millisecond {
		t.Errorf("Expected InitialDelay to be 100ms, got %v", config.InitialDelay)
	}
	
	if config.MaxDelay != 30*time.Second {
		t.Errorf("Expected MaxDelay to be 30s, got %v", config.MaxDelay)
	}
	
	if config.BackoffFactor != 2.0 {
		t.Errorf("Expected BackoffFactor to be 2.0, got %f", config.BackoffFactor)
	}
	
	if !config.Jitter {
		t.Error("Expected Jitter to be true")
	}
	
	expectedRetryableErrors := []int{408, 429, 500, 502, 503, 504}
	if len(config.RetryableErrors) != len(expectedRetryableErrors) {
		t.Errorf("Expected %d retryable errors, got %d", len(expectedRetryableErrors), len(config.RetryableErrors))
	}
}

func TestRetryableError(t *testing.T) {
	err := &RetryableError{
		StatusCode: 500,
		Message:    "Internal Server Error",
	}
	
	expected := "retryable error (status: 500): Internal Server Error"
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestIsRetryableError(t *testing.T) {
	config := DefaultRetryConfig()
	
	// Test retryable error
	retryableErr := &RetryableError{StatusCode: 500, Message: "Server Error"}
	if !config.IsRetryableError(retryableErr) {
		t.Error("Expected 500 error to be retryable")
	}
	
	// Test non-retryable error
	nonRetryableErr := &RetryableError{StatusCode: 400, Message: "Bad Request"}
	if config.IsRetryableError(nonRetryableErr) {
		t.Error("Expected 400 error to not be retryable")
	}
	
	// Test regular error
	regularErr := errors.New("regular error")
	if config.IsRetryableError(regularErr) {
		t.Error("Expected regular error to not be retryable")
	}
}

func TestRetrySuccess(t *testing.T) {
	config := DefaultRetryConfig()
	attempts := 0
	
	err := config.Retry(context.Background(), func() error {
		attempts++
		if attempts == 1 {
			return &RetryableError{StatusCode: 500, Message: "Server Error"}
		}
		return nil
	})
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryMaxAttemptsExceeded(t *testing.T) {
	config := DefaultRetryConfig()
	attempts := 0
	
	err := config.Retry(context.Background(), func() error {
		attempts++
		return &RetryableError{StatusCode: 500, Message: "Server Error"}
	})
	
	if err == nil {
		t.Error("Expected error for max attempts exceeded")
	}
	
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
	
	expectedMsg := "max retry attempts (3) exceeded: retryable error (status: 500): Server Error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestRetryNonRetryableError(t *testing.T) {
	config := DefaultRetryConfig()
	attempts := 0
	
	err := config.Retry(context.Background(), func() error {
		attempts++
		return &RetryableError{StatusCode: 400, Message: "Bad Request"}
	})
	
	if err == nil {
		t.Error("Expected error for non-retryable status")
	}
	
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetryWithResult(t *testing.T) {
	config := DefaultRetryConfig()
	attempts := 0
	
	result, err := config.RetryWithResult(context.Background(), func() (interface{}, error) {
		attempts++
		if attempts == 1 {
			return nil, &RetryableError{StatusCode: 500, Message: "Server Error"}
		}
		return "success", nil
	})
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if result != "success" {
		t.Errorf("Expected result 'success', got %v", result)
	}
	
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryContextCancellation(t *testing.T) {
	config := DefaultRetryConfig()
	ctx, cancel := context.WithCancel(context.Background())
	
	// Cancel context immediately
	cancel()
	
	attempts := 0
	err := config.Retry(ctx, func() error {
		attempts++
		return &RetryableError{StatusCode: 500, Message: "Server Error"}
	})
	
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
	
	if attempts != 0 {
		t.Errorf("Expected 0 attempts, got %d", attempts)
	}
}

func TestRetryWithBackoff(t *testing.T) {
	config := DefaultRetryConfig()
	attempts := 0
	
	err := config.RetryWithBackoff(context.Background(), func() error {
		attempts++
		if attempts == 1 {
			return &RetryableError{StatusCode: 500, Message: "Server Error"}
		}
		return nil
	}, func(attempt int) time.Duration {
		return time.Duration(attempt) * 100 * time.Millisecond
	})
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestExponentialBackoff(t *testing.T) {
	baseDelay := 100 * time.Millisecond
	factor := 2.0
	
	testCases := []struct {
		attempt    int
		expected  time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 400 * time.Millisecond},
		{3, 800 * time.Millisecond},
	}
	
	for _, tc := range testCases {
		result := ExponentialBackoff(tc.attempt, baseDelay, factor)
		if result != tc.expected {
			t.Errorf("Expected backoff for attempt %d to be %v, got %v", tc.attempt, tc.expected, result)
		}
	}
}

func TestLinearBackoff(t *testing.T) {
	baseDelay := 100 * time.Millisecond
	
	testCases := []struct {
		attempt    int
		expected  time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 200 * time.Millisecond},
		{2, 300 * time.Millisecond},
		{3, 400 * time.Millisecond},
	}
	
	for _, tc := range testCases {
		result := LinearBackoff(tc.attempt, baseDelay)
		if result != tc.expected {
			t.Errorf("Expected backoff for attempt %d to be %v, got %v", tc.attempt, tc.expected, result)
		}
	}
}

func TestConstantBackoff(t *testing.T) {
	baseDelay := 100 * time.Millisecond
	
	for attempt := 0; attempt < 5; attempt++ {
		result := ConstantBackoff(attempt, baseDelay)
		if result != baseDelay {
			t.Errorf("Expected backoff for attempt %d to be %v, got %v", attempt, baseDelay, result)
		}
	}
}

func TestFibonacciBackoff(t *testing.T) {
	baseDelay := 100 * time.Millisecond
	
	testCases := []struct {
		attempt    int
		expected  time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 100 * time.Millisecond},
		{2, 200 * time.Millisecond},
		{3, 300 * time.Millisecond},
		{4, 500 * time.Millisecond},
		{5, 800 * time.Millisecond},
	}
	
	for _, tc := range testCases {
		result := FibonacciBackoff(tc.attempt, baseDelay)
		if result != tc.expected {
			t.Errorf("Expected backoff for attempt %d to be %v, got %v", tc.attempt, tc.expected, result)
		}
	}
}

func TestRetryWithCustomConfig(t *testing.T) {
	config := &RetryConfig{
		MaxAttempts:   2,
		InitialDelay:  50 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 1.5,
		RetryableErrors: []int{500},
		Jitter:        false,
	}
	
	attempts := 0
	err := config.Retry(context.Background(), func() error {
		attempts++
		return &RetryableError{StatusCode: 500, Message: "Server Error"}
	})
	
	if err == nil {
		t.Error("Expected error for max attempts exceeded")
	}
	
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryJitter(t *testing.T) {
	config := &RetryConfig{
		MaxAttempts:   2,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      1 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []int{500},
		Jitter:        true,
	}
	
	attempts := 0
	err := config.Retry(context.Background(), func() error {
		attempts++
		return &RetryableError{StatusCode: 500, Message: "Server Error"}
	})
	
	if err == nil {
		t.Error("Expected error for max attempts exceeded")
	}
	
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
} 