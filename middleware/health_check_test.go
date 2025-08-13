package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHealthChecker(t *testing.T) {
	hc := NewHealthChecker("test-service")
	
	if hc.serviceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got %s", hc.serviceName)
	}
	
	if hc.timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", hc.timeout)
	}
	
	if len(hc.checks) != 0 {
		t.Errorf("Expected 0 checks initially, got %d", len(hc.checks))
	}
}

func TestHealthCheckerAddRemoveCheck(t *testing.T) {
	hc := NewHealthChecker("test-service")
	
	// Add a check
	check := func(ctx context.Context) *DependencyHealth {
		return &DependencyHealth{
			Status:    StatusHealthy,
			Message:   "Test check",
			Timestamp: time.Now(),
		}
	}
	
	hc.AddCheck("test-dependency", check)
	
	if len(hc.checks) != 1 {
		t.Errorf("Expected 1 check after adding, got %d", len(hc.checks))
	}
	
	// Remove the check
	hc.RemoveCheck("test-dependency")
	
	if len(hc.checks) != 0 {
		t.Errorf("Expected 0 checks after removing, got %d", len(hc.checks))
	}
}

func TestHealthCheckerSetTimeout(t *testing.T) {
	hc := NewHealthChecker("test-service")
	
	newTimeout := 60 * time.Second
	hc.SetTimeout(newTimeout)
	
	if hc.timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, hc.timeout)
	}
}

func TestHealthCheckerCheckHealth(t *testing.T) {
	hc := NewHealthChecker("test-service")
	
	// Add healthy check
	hc.AddCheck("healthy-dep", func(ctx context.Context) *DependencyHealth {
		return &DependencyHealth{
			Status:    StatusHealthy,
			Message:   "Healthy dependency",
			Timestamp: time.Now(),
		}
	})
	
	// Add degraded check
	hc.AddCheck("degraded-dep", func(ctx context.Context) *DependencyHealth {
		return &DependencyHealth{
			Status:    StatusDegraded,
			Message:   "Degraded dependency",
			Timestamp: time.Now(),
		}
	})
	
	// Add unhealthy check
	hc.AddCheck("unhealthy-dep", func(ctx context.Context) *DependencyHealth {
		return &DependencyHealth{
			Status:    StatusUnhealthy,
			Message:   "Unhealthy dependency",
			Timestamp: time.Now(),
		}
	})
	
	health := hc.CheckHealth(context.Background())
	
	if health["service"] != "test-service" {
		t.Errorf("Expected service name 'test-service', got %v", health["service"])
	}
	
	if health["status"] != "unhealthy" {
		t.Errorf("Expected overall status 'unhealthy', got %v", health["status"])
	}
	
	if health["total_checks"] != 3 {
		t.Errorf("Expected 3 total checks, got %v", health["total_checks"])
	}
	
	if health["healthy"] != 1 {
		t.Errorf("Expected 1 healthy dependency, got %v", health["healthy"])
	}
	
	if health["degraded"] != 1 {
		t.Errorf("Expected 1 degraded dependency, got %v", health["degraded"])
	}
	
	if health["unhealthy"] != 1 {
		t.Errorf("Expected 1 unhealthy dependency, got %v", health["unhealthy"])
	}
}

func TestHealthCheckerCheckHealthAllHealthy(t *testing.T) {
	hc := NewHealthChecker("test-service")
	
	// Add only healthy checks
	hc.AddCheck("healthy-dep1", func(ctx context.Context) *DependencyHealth {
		return &DependencyHealth{
			Status:    StatusHealthy,
			Message:   "Healthy dependency 1",
			Timestamp: time.Now(),
		}
	})
	
	hc.AddCheck("healthy-dep2", func(ctx context.Context) *DependencyHealth {
		return &DependencyHealth{
			Status:    StatusHealthy,
			Message:   "Healthy dependency 2",
			Timestamp: time.Now(),
		}
	})
	
	health := hc.CheckHealth(context.Background())
	
	if health["status"] != "healthy" {
		t.Errorf("Expected overall status 'healthy', got %v", health["status"])
	}
	
	if health["healthy"] != 2 {
		t.Errorf("Expected 2 healthy dependencies, got %v", health["healthy"])
	}
}

func TestHealthCheckerCheckHealthDegraded(t *testing.T) {
	hc := NewHealthChecker("test-service")
	
	// Add healthy and degraded checks (no unhealthy)
	hc.AddCheck("healthy-dep", func(ctx context.Context) *DependencyHealth {
		return &DependencyHealth{
			Status:    StatusHealthy,
			Message:   "Healthy dependency",
			Timestamp: time.Now(),
		}
	})
	
	hc.AddCheck("degraded-dep", func(ctx context.Context) *DependencyHealth {
		return &DependencyHealth{
			Status:    StatusDegraded,
			Message:   "Degraded dependency",
			Timestamp: time.Now(),
		}
	})
	
	health := hc.CheckHealth(context.Background())
	
	if health["status"] != "degraded" {
		t.Errorf("Expected overall status 'degraded', got %v", health["status"])
	}
}

func TestHealthCheckerCheckHealthTimeout(t *testing.T) {
	hc := NewHealthChecker("test-service")
	hc.SetTimeout(100 * time.Millisecond)
	
	// Add a check that takes longer than timeout
	hc.AddCheck("slow-dep", func(ctx context.Context) *DependencyHealth {
		time.Sleep(200 * time.Millisecond)
		return &DependencyHealth{
			Status:    StatusHealthy,
			Message:   "Slow dependency",
			Timestamp: time.Now(),
		}
	})
	
	health := hc.CheckHealth(context.Background())
	
	// Should get unhealthy status due to timeout
	if health["status"] != "unhealthy" {
		t.Errorf("Expected overall status 'unhealthy' due to timeout, got %v", health["status"])
	}
}

func TestHealthCheckerCheckHealthNilCheck(t *testing.T) {
	hc := NewHealthChecker("test-service")
	
	// Add a check that returns nil
	hc.AddCheck("nil-dep", func(ctx context.Context) *DependencyHealth {
		return nil
	})
	
	health := hc.CheckHealth(context.Background())
	
	if health["status"] != "unhealthy" {
		t.Errorf("Expected overall status 'unhealthy' due to nil check, got %v", health["status"])
	}
}

func TestHealthCheckerHTTPHandler(t *testing.T) {
	hc := NewHealthChecker("test-service")
	
	// Add a healthy check
	hc.AddCheck("healthy-dep", func(ctx context.Context) *DependencyHealth {
		return &DependencyHealth{
			Status:    StatusHealthy,
			Message:   "Healthy dependency",
			Timestamp: time.Now(),
		}
	})
	
	// Create request and response recorder
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	// Call the handler
	hc.HTTPHandler()(w, req)
	
	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
	
	// Parse response body
	var health map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if health["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}
}

func TestHealthCheckerHTTPHandlerUnhealthy(t *testing.T) {
	hc := NewHealthChecker("test-service")
	
	// Add an unhealthy check
	hc.AddCheck("unhealthy-dep", func(ctx context.Context) *DependencyHealth {
		return &DependencyHealth{
			Status:    StatusUnhealthy,
			Message:   "Unhealthy dependency",
			Timestamp: time.Now(),
		}
	})
	
	// Create request and response recorder
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	// Call the handler
	hc.HTTPHandler()(w, req)
	
	// Check response
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status code %d, got %d", http.StatusServiceUnavailable, w.Code)
	}
}

func TestHealthCheckerHTTPHandlerDegraded(t *testing.T) {
	hc := NewHealthChecker("test-service")
	
	// Add a degraded check
	hc.AddCheck("degraded-dep", func(ctx context.Context) *DependencyHealth {
		return &DependencyHealth{
			Status:    StatusDegraded,
			Message:   "Degraded dependency",
			Timestamp: time.Now(),
		}
	})
	
	// Create request and response recorder
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	// Call the handler
	hc.HTTPHandler()(w, req)
	
	// Check response - degraded should still return 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d for degraded, got %d", http.StatusOK, w.Code)
	}
}

func TestHTTPHealthCheck(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()
	
	// Create health check
	check := HTTPHealthCheck("test-service", server.URL, 5*time.Second)
	
	// Test healthy response
	health := check(context.Background())
	
	if health.Status != StatusHealthy {
		t.Errorf("Expected status 'healthy', got %s", health.Status)
	}
	
	if health.Name != "" { // Name should be set by the health checker
		t.Errorf("Expected empty name, got %s", health.Name)
	}
}

func TestHTTPHealthCheckUnhealthy(t *testing.T) {
	// Create a test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	}))
	defer server.Close()
	
	// Create health check
	check := HTTPHealthCheck("test-service", server.URL, 5*time.Second)
	
	// Test degraded response (non-2xx status)
	health := check(context.Background())
	
	if health.Status != StatusDegraded {
		t.Errorf("Expected status 'degraded', got %s", health.Status)
	}
}

func TestHTTPHealthCheckTimeout(t *testing.T) {
	// Create a test server that hangs
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	
	// Create health check with short timeout
	check := HTTPHealthCheck("test-service", server.URL, 100*time.Millisecond)
	
	// Test timeout
	health := check(context.Background())
	
	if health.Status != StatusUnhealthy {
		t.Errorf("Expected status 'unhealthy' due to timeout, got %s", health.Status)
	}
}

func TestCustomHealthCheck(t *testing.T) {
	// Test successful custom check
	successCheck := CustomHealthCheck(func(ctx context.Context) error {
		return nil
	})
	
	health := successCheck(context.Background())
	
	if health.Status != StatusHealthy {
		t.Errorf("Expected status 'healthy', got %s", health.Status)
	}
	
	// Test failed custom check
	failedCheck := CustomHealthCheck(func(ctx context.Context) error {
		return errors.New("custom error")
	})
	
	health = failedCheck(context.Background())
	
	if health.Status != StatusUnhealthy {
		t.Errorf("Expected status 'unhealthy', got %s", health.Status)
	}
	
	if health.Message != "custom error" {
		t.Errorf("Expected message 'custom error', got %s", health.Message)
	}
}

func TestHealthStatusString(t *testing.T) {
	testCases := []struct {
		status   HealthStatus
		expected string
	}{
		{StatusHealthy, "healthy"},
		{StatusDegraded, "degraded"},
		{StatusUnhealthy, "unhealthy"},
	}
	
	for _, tc := range testCases {
		if tc.status.String() != tc.expected {
			t.Errorf("Expected status %v to stringify to '%s', got '%s'", tc.status, tc.expected, tc.status.String())
		}
	}
}

func TestDependencyHealth(t *testing.T) {
	now := time.Now()
	health := &DependencyHealth{
		Name:      "test-dependency",
		Status:    StatusHealthy,
		Message:   "Test message",
		Timestamp: now,
		Details: map[string]interface{}{
			"key": "value",
		},
	}
	
	if health.Name != "test-dependency" {
		t.Errorf("Expected name 'test-dependency', got %s", health.Name)
	}
	
	if health.Status != StatusHealthy {
		t.Errorf("Expected status 'healthy', got %s", health.Status)
	}
	
	if health.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got %s", health.Message)
	}
	
	if health.Timestamp != now {
		t.Errorf("Expected timestamp %v, got %v", now, health.Timestamp)
	}
	
	if health.Details["key"] != "value" {
		t.Errorf("Expected detail value 'value', got %v", health.Details["key"])
	}
} 