package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HealthStatus represents the overall health status of a service
type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

// String returns the string representation of the health status
func (s HealthStatus) String() string {
	return string(s)
}

// DependencyHealth represents the health of a single dependency
type DependencyHealth struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) *DependencyHealth

// HealthChecker manages health checks for a service
type HealthChecker struct {
	serviceName string
	checks      map[string]HealthCheck
	mutex       sync.RWMutex
	timeout     time.Duration
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(serviceName string) *HealthChecker {
	return &HealthChecker{
		serviceName: serviceName,
		checks:      make(map[string]HealthCheck),
		timeout:     30 * time.Second,
	}
}

// AddCheck adds a health check for a dependency
func (hc *HealthChecker) AddCheck(name string, check HealthCheck) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.checks[name] = check
}

// RemoveCheck removes a health check
func (hc *HealthChecker) RemoveCheck(name string) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	delete(hc.checks, name)
}

// SetTimeout sets the timeout for health checks
func (hc *HealthChecker) SetTimeout(timeout time.Duration) {
	hc.mutex.Lock()
	defer hc.mutex.Unlock()
	hc.timeout = timeout
}

// CheckHealth performs all health checks and returns the overall status
func (hc *HealthChecker) CheckHealth(ctx context.Context) map[string]interface{} {
	hc.mutex.RLock()
	checks := make(map[string]HealthCheck)
	for name, check := range hc.checks {
		checks[name] = check
	}
	timeout := hc.timeout
	hc.mutex.RUnlock()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Run all health checks concurrently
	results := make(chan *DependencyHealth, len(checks))
	var wg sync.WaitGroup

	for name, check := range checks {
		wg.Add(1)
		go func(name string, check HealthCheck) {
			defer wg.Done()
			
			// Create a channel to receive the result
			resultChan := make(chan *DependencyHealth, 1)
			
			// Run the health check in a goroutine
			go func() {
				result := check(ctx)
				resultChan <- result
			}()
			
			// Wait for result or timeout
			select {
			case result := <-resultChan:
				if result != nil {
					result.Name = name
					if result.Timestamp.IsZero() {
						result.Timestamp = time.Now()
					}
				} else {
					result = &DependencyHealth{
						Name:      name,
						Status:    StatusUnhealthy,
						Message:   "Health check returned nil",
						Timestamp: time.Now(),
					}
				}
				results <- result
			case <-ctx.Done():
				// Timeout occurred
				results <- &DependencyHealth{
					Name:      name,
					Status:    StatusUnhealthy,
					Message:   "Health check timed out",
					Timestamp: time.Now(),
				}
			}
		}(name, check)
	}

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	dependencies := make(map[string]*DependencyHealth)
	overallStatus := StatusHealthy

	for result := range results {
		dependencies[result.Name] = result
		
		// Update overall status
		switch result.Status {
		case StatusUnhealthy:
			overallStatus = StatusUnhealthy
		case StatusDegraded:
			if overallStatus != StatusUnhealthy {
				overallStatus = StatusDegraded
			}
		}
	}

	return map[string]interface{}{
		"service":       hc.serviceName,
		"status":        overallStatus.String(),
		"timestamp":     time.Now().UTC(),
		"dependencies":  dependencies,
		"total_checks":  len(checks),
		"healthy":       countStatus(dependencies, StatusHealthy),
		"degraded":      countStatus(dependencies, StatusDegraded),
		"unhealthy":     countStatus(dependencies, StatusUnhealthy),
	}
}

// countStatus counts dependencies with a specific status
func countStatus(dependencies map[string]*DependencyHealth, status HealthStatus) int {
	count := 0
	for _, dep := range dependencies {
		if dep.Status == status {
			count++
		}
	}
	return count
}

// HTTPHandler returns an HTTP handler for health check endpoints
func (hc *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		health := hc.CheckHealth(ctx)
		
		// Set appropriate HTTP status code
		status := health["status"].(string)
		var httpStatus int
		switch status {
		case "healthy":
			httpStatus = http.StatusOK
		case "degraded":
			httpStatus = http.StatusOK // Service is running but some dependencies are down
		case "unhealthy":
			httpStatus = http.StatusServiceUnavailable
		default:
			httpStatus = http.StatusInternalServerError
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(httpStatus)
		
		json.NewEncoder(w).Encode(health)
	}
}

// Predefined health checks

// DatabaseHealthCheck creates a health check for database connectivity
func DatabaseHealthCheck(db interface{}) HealthCheck {
	return func(ctx context.Context) *DependencyHealth {
		// This is a generic interface - specific implementations should be provided by services
		return &DependencyHealth{
			Status:    StatusHealthy,
			Message:   "Database health check not implemented - override with specific implementation",
			Timestamp: time.Now(),
		}
	}
}

// HTTPHealthCheck creates a health check for HTTP service connectivity
func HTTPHealthCheck(name, url string, timeout time.Duration) HealthCheck {
	return func(ctx context.Context) *DependencyHealth {
		client := &http.Client{
			Timeout: timeout,
		}
		
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return &DependencyHealth{
				Status:    StatusUnhealthy,
				Message:   fmt.Sprintf("Failed to create request: %v", err),
				Timestamp: time.Now(),
			}
		}
		
		resp, err := client.Do(req)
		if err != nil {
			return &DependencyHealth{
				Status:    StatusUnhealthy,
				Message:   fmt.Sprintf("Request failed: %v", err),
				Timestamp: time.Now(),
			}
		}
		defer resp.Body.Close()
		
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return &DependencyHealth{
				Status:    StatusHealthy,
				Message:   fmt.Sprintf("Service responded with status %d", resp.StatusCode),
				Timestamp: time.Now(),
				Details: map[string]interface{}{
					"status_code": resp.StatusCode,
					"url":         url,
				},
			}
		}
		
		return &DependencyHealth{
			Status:    StatusDegraded,
			Message:   fmt.Sprintf("Service responded with status %d", resp.StatusCode),
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"status_code": resp.StatusCode,
				"url":         url,
			},
		}
	}
}

// RedisHealthCheck creates a health check for Redis connectivity
func RedisHealthCheck(redisClient interface{}) HealthCheck {
	return func(ctx context.Context) *DependencyHealth {
		// This is a generic interface - specific implementations should be provided by services
		return &DependencyHealth{
			Status:    StatusHealthy,
			Message:   "Redis health check not implemented - override with specific implementation",
			Timestamp: time.Now(),
		}
	}
}

// CustomHealthCheck creates a custom health check function
func CustomHealthCheck(check func(ctx context.Context) error) HealthCheck {
	return func(ctx context.Context) *DependencyHealth {
		if err := check(ctx); err != nil {
			return &DependencyHealth{
				Status:    StatusUnhealthy,
				Message:   err.Error(),
				Timestamp: time.Now(),
			}
		}
		
		return &DependencyHealth{
			Status:    StatusHealthy,
			Message:   "Custom health check passed",
			Timestamp: time.Now(),
		}
	}
} 