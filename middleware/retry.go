package middleware

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryConfig holds the configuration for retry logic
type RetryConfig struct {
	MaxAttempts     int           `json:"max_attempts"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []int         `json:"retryable_errors"`
	Jitter          bool          `json:"jitter"`
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RetryableErrors: []int{
			408, // Request Timeout
			429, // Too Many Requests
			500, // Internal Server Error
			502, // Bad Gateway
			503, // Service Unavailable
			504, // Gateway Timeout
		},
		Jitter: true,
	}
}

// RetryableError represents an error that can be retried
type RetryableError struct {
	StatusCode int
	Message    string
}

func (e RetryableError) Error() string {
	return fmt.Sprintf("retryable error (status: %d): %s", e.StatusCode, e.Message)
}

// IsRetryableError checks if an error is retryable based on the configuration
func (rc *RetryConfig) IsRetryableError(err error) bool {
	if retryableErr, ok := err.(*RetryableError); ok {
		for _, statusCode := range rc.RetryableErrors {
			if retryableErr.StatusCode == statusCode {
				return true
			}
		}
	}
	return false
}

// Retry executes a function with retry logic and exponential backoff
func (rc *RetryConfig) Retry(ctx context.Context, fn func() error) error {
	var lastErr error
	delay := rc.InitialDelay

	for attempt := 0; attempt < rc.MaxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// If this was the last attempt, don't wait
		if attempt == rc.MaxAttempts-1 {
			break
		}

		// Check if the error is retryable
		if !rc.IsRetryableError(lastErr) {
			return lastErr
		}

		// Calculate next delay with exponential backoff
		nextDelay := time.Duration(float64(delay) * rc.BackoffFactor)
		if nextDelay > rc.MaxDelay {
			nextDelay = rc.MaxDelay
		}

		// Add jitter if enabled
		if rc.Jitter {
			jitter := time.Duration(rand.Float64() * float64(nextDelay) * 0.1) // 10% jitter
			nextDelay += jitter
		}

		// Wait for the delay
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		delay = nextDelay
	}

	return fmt.Errorf("max retry attempts (%d) exceeded: %w", rc.MaxAttempts, lastErr)
}

// RetryWithResult executes a function with retry logic and returns a result
func (rc *RetryConfig) RetryWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	var lastErr error
	delay := rc.InitialDelay

	for attempt := 0; attempt < rc.MaxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Execute the function
		if res, err := fn(); err == nil {
			return res, nil
		} else {
			lastErr = err
		}

		// If this was the last attempt, don't wait
		if attempt == rc.MaxAttempts-1 {
			break
		}

		// Check if the error is retryable
		if !rc.IsRetryableError(lastErr) {
			return nil, lastErr
		}

		// Calculate next delay with exponential backoff
		nextDelay := time.Duration(float64(delay) * rc.BackoffFactor)
		if nextDelay > rc.MaxDelay {
			nextDelay = rc.MaxDelay
		}

		// Add jitter if enabled
		if rc.Jitter {
			jitter := time.Duration(rand.Float64() * float64(nextDelay) * 0.1) // 10% jitter
			nextDelay += jitter
		}

		// Wait for the delay
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}

		delay = nextDelay
	}

	return nil, fmt.Errorf("max retry attempts (%d) exceeded: %w", rc.MaxAttempts, lastErr)
}

// RetryWithBackoff executes a function with custom backoff calculation
func (rc *RetryConfig) RetryWithBackoff(ctx context.Context, fn func() error, backoffFunc func(attempt int) time.Duration) error {
	var lastErr error

	for attempt := 0; attempt < rc.MaxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
		}

		// If this was the last attempt, don't wait
		if attempt == rc.MaxAttempts-1 {
			break
		}

		// Check if the error is retryable
		if !rc.IsRetryableError(lastErr) {
			return lastErr
		}

		// Calculate delay using custom backoff function
		delay := backoffFunc(attempt)
		if delay > rc.MaxDelay {
			delay = rc.MaxDelay
		}

		// Wait for the delay
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("max retry attempts (%d) exceeded: %w", rc.MaxAttempts, lastErr)
}

// ExponentialBackoff calculates exponential backoff delay
func ExponentialBackoff(attempt int, baseDelay time.Duration, factor float64) time.Duration {
	delay := float64(baseDelay) * math.Pow(factor, float64(attempt))
	return time.Duration(delay)
}

// LinearBackoff calculates linear backoff delay
func LinearBackoff(attempt int, baseDelay time.Duration) time.Duration {
	return baseDelay * time.Duration(attempt+1)
}

// ConstantBackoff returns constant delay
func ConstantBackoff(attempt int, baseDelay time.Duration) time.Duration {
	return baseDelay
}

// FibonacciBackoff calculates Fibonacci backoff delay
func FibonacciBackoff(attempt int, baseDelay time.Duration) time.Duration {
	if attempt <= 1 {
		return baseDelay
	}
	
	fib := 1
	prev := 1
	for i := 2; i <= attempt; i++ {
		fib, prev = fib+prev, fib
	}
	
	return baseDelay * time.Duration(fib)
} 