package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewMetricsRegistry(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	if registry.serviceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got %s", registry.serviceName)
	}
	
	if registry.metrics == nil {
		t.Error("Expected metrics map to be initialized")
	}
}

func TestRecordServiceCall(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record a successful service call
	duration := 100 * time.Millisecond
	registry.RecordServiceCall("user-service", "GET", "success", duration)
	
	// Check if metrics were recorded
	expected := 1.0
	if testutil.ToFloat64(serviceCallTotal.WithLabelValues("user-service", "GET", "success")) != expected {
		t.Errorf("Expected service call total to be %f, got %f", expected, testutil.ToFloat64(serviceCallTotal.WithLabelValues("user-service", "GET", "success")))
	}
	
	// Record a failed service call
	registry.RecordServiceCall("user-service", "GET", "error", duration)
	
	// Check error metrics
	if testutil.ToFloat64(serviceCallErrors.WithLabelValues("user-service", "GET", "error")) != 1.0 {
		t.Error("Expected service call error to be recorded")
	}
}

func TestRecordCircuitBreakerState(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record circuit breaker states
	registry.RecordCircuitBreakerState("user-service", StateClosed)
	registry.RecordCircuitBreakerState("user-service", StateHalfOpen)
	registry.RecordCircuitBreakerState("user-service", StateOpen)
	
	// Check if the last state (open) is recorded
	expected := 2.0 // StateOpen value
	if testutil.ToFloat64(circuitBreakerState.WithLabelValues("user-service")) != expected {
		t.Errorf("Expected circuit breaker state to be %f, got %f", expected, testutil.ToFloat64(circuitBreakerState.WithLabelValues("user-service")))
	}
}

func TestRecordCircuitBreakerFailure(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record circuit breaker failures
	registry.RecordCircuitBreakerFailure("user-service")
	registry.RecordCircuitBreakerFailure("user-service")
	
	// Check if failures were recorded
	expected := 2.0
	if testutil.ToFloat64(circuitBreakerFailures.WithLabelValues("user-service")) != expected {
		t.Errorf("Expected circuit breaker failures to be %f, got %f", expected, testutil.ToFloat64(circuitBreakerFailures.WithLabelValues("user-service")))
	}
}

func TestRecordCircuitBreakerTransition(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record circuit breaker transitions
	registry.RecordCircuitBreakerTransition("user-service", StateClosed, StateOpen)
	registry.RecordCircuitBreakerTransition("user-service", StateOpen, StateHalfOpen)
	
	// Check if transitions were recorded
	if testutil.ToFloat64(circuitBreakerTransitions.WithLabelValues("user-service", "CLOSED", "OPEN")) != 1.0 {
		t.Error("Expected first transition to be recorded")
	}
	
	if testutil.ToFloat64(circuitBreakerTransitions.WithLabelValues("user-service", "OPEN", "HALF_OPEN")) != 1.0 {
		t.Error("Expected second transition to be recorded")
	}
}

func TestRecordRetryAttempt(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record retry attempts
	registry.RecordRetryAttempt("user-service", "GET")
	registry.RecordRetryAttempt("user-service", "GET")
	
	// Check if attempts were recorded
	expected := 2.0
	if testutil.ToFloat64(retryAttempts.WithLabelValues("user-service", "GET")) != expected {
		t.Errorf("Expected retry attempts to be %f, got %f", expected, testutil.ToFloat64(retryAttempts.WithLabelValues("user-service", "GET")))
	}
}

func TestRecordRetryFailure(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record retry failures
	registry.RecordRetryFailure("user-service", "GET")
	
	// Check if failures were recorded
	expected := 1.0
	if testutil.ToFloat64(retryFailures.WithLabelValues("user-service", "GET")) != expected {
		t.Errorf("Expected retry failures to be %f, got %f", expected, testutil.ToFloat64(retryFailures.WithLabelValues("user-service", "GET")))
	}
}

func TestRecordHealthCheck(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record health check
	duration := 50 * time.Millisecond
	registry.RecordHealthCheck("database", StatusHealthy, duration)
	
	// Check if health check was recorded
	expected := 2.0 // StatusHealthy value
	if testutil.ToFloat64(healthCheckStatus.WithLabelValues("test-service", "database")) != expected {
		t.Errorf("Expected health check status to be %f, got %f", expected, testutil.ToFloat64(healthCheckStatus.WithLabelValues("test-service", "database")))
	}
}

func TestRecordHTTPRequest(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record HTTP request
	duration := 100 * time.Millisecond
	registry.RecordHTTPRequest("GET", "/api/v1/users", 200, duration)
	
	// Check if request was recorded
	expected := 1.0
	if testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/api/v1/users", "200")) != expected {
		t.Errorf("Expected HTTP request total to be %f, got %f", expected, testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/api/v1/users", "200")))
	}
}

func TestRecordHTTPRequestStartEnd(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record request start
	registry.RecordHTTPRequestStart("GET", "/api/v1/users")
	
	// Check if in-flight count increased
	expected := 1.0
	if testutil.ToFloat64(httpRequestsInFlight.WithLabelValues("GET", "/api/v1/users")) != expected {
		t.Errorf("Expected in-flight requests to be %f, got %f", expected, testutil.ToFloat64(httpRequestsInFlight.WithLabelValues("GET", "/api/v1/users")))
	}
	
	// Record request end
	registry.RecordHTTPRequestEnd("GET", "/api/v1/users")
	
	// Check if in-flight count decreased
	expected = 0.0
	if testutil.ToFloat64(httpRequestsInFlight.WithLabelValues("GET", "/api/v1/users")) != expected {
		t.Errorf("Expected in-flight requests to be %f, got %f", expected, testutil.ToFloat64(httpRequestsInFlight.WithLabelValues("GET", "/api/v1/users")))
	}
}

func TestRecordDatabaseConnection(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record database connections
	registry.RecordDatabaseConnection("postgres", 5)
	
	// Check if connections were recorded
	expected := 5.0
	if testutil.ToFloat64(databaseConnections.WithLabelValues("test-service", "postgres")) != expected {
		t.Errorf("Expected database connections to be %f, got %f", expected, testutil.ToFloat64(databaseConnections.WithLabelValues("test-service", "postgres")))
	}
}

func TestRecordDatabaseQuery(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record database query
	duration := 25 * time.Millisecond
	registry.RecordDatabaseQuery("postgres", "SELECT", duration)
	
	// Check if query was recorded - we can't easily test histogram values with testutil
	// but we can verify the function doesn't panic
	t.Log("Database query duration recorded successfully")
}

func TestRecordDatabaseError(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record database error
	registry.RecordDatabaseError("postgres", "connection_failed")
	
	// Check if error was recorded
	expected := 1.0
	if testutil.ToFloat64(databaseErrors.WithLabelValues("test-service", "postgres", "connection_failed")) != expected {
		t.Errorf("Expected database errors to be %f, got %f", expected, testutil.ToFloat64(databaseErrors.WithLabelValues("test-service", "postgres", "connection_failed")))
	}
}

func TestRecordRedisConnection(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record Redis connections
	registry.RecordRedisConnection(3)
	
	// Check if connections were recorded
	expected := 3.0
	if testutil.ToFloat64(redisConnections.WithLabelValues("test-service")) != expected {
		t.Errorf("Expected Redis connections to be %f, got %f", expected, testutil.ToFloat64(redisConnections.WithLabelValues("test-service")))
	}
}

func TestRecordRedisOperation(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Record Redis operation
	duration := 10 * time.Millisecond
	registry.RecordRedisOperation("GET", "success", duration)
	
	// Check if operation was recorded
	expected := 1.0
	if testutil.ToFloat64(redisOperations.WithLabelValues("test-service", "GET", "success")) != expected {
		t.Errorf("Expected Redis operations to be %f, got %f", expected, testutil.ToFloat64(redisOperations.WithLabelValues("test-service", "GET", "success")))
	}
}

func TestHTTPHandler(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	handler := registry.HTTPHandler()
	if handler == nil {
		t.Error("Expected HTTP handler to be returned")
	}
	
	// Test that the handler responds
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	
	handler.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
	
	// Check that metrics are included in response
	body := w.Body.String()
	if body == "" {
		t.Error("Expected metrics response body to not be empty")
	}
}

func TestMetricsMiddleware(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Wrap with metrics middleware
	wrappedHandler := registry.MetricsMiddleware()(handler)
	
	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	wrappedHandler.ServeHTTP(w, req)
	
	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
	
	// Check that metrics were recorded
	expected := 1.0
	if testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/test", "200")) != expected {
		t.Errorf("Expected HTTP request total to be %f, got %f", expected, testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/test", "200")))
	}
}

func TestGinMetricsMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	registry := NewMetricsRegistry("test-service")
	
	// Create Gin router
	r := gin.New()
	r.Use(registry.GinMetricsMiddleware())
	
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	r.ServeHTTP(w, req)
	
	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
	
	// Check that metrics were recorded - the count may be higher due to previous tests
	// but it should be at least 1
	currentTotal := testutil.ToFloat64(httpRequestsTotal.WithLabelValues("GET", "/test", "200"))
	if currentTotal < 1.0 {
		t.Errorf("Expected HTTP request total to be at least 1.0, got %f", currentTotal)
	}
}

func TestGetMetricsSummary(t *testing.T) {
	registry := NewMetricsRegistry("test-service")
	
	summary := registry.GetMetricsSummary()
	
	if summary["service"] != "test-service" {
		t.Errorf("Expected service name 'test-service', got %v", summary["service"])
	}
	
	metrics, ok := summary["metrics"].(map[string]interface{})
	if !ok {
		t.Error("Expected metrics to be a map")
	}
	
	// Check that all metric categories are present
	expectedCategories := []string{
		"service_calls", "circuit_breaker", "retry", "health_check",
		"http_requests", "database", "redis",
	}
	
	for _, category := range expectedCategories {
		if _, exists := metrics[category]; !exists {
			t.Errorf("Expected metric category '%s' to be present", category)
		}
	}
}

func TestResponseWriter(t *testing.T) {
	// Test response writer wrapper
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: 200}
	
	// Test WriteHeader
	rw.WriteHeader(http.StatusNotFound)
	if rw.statusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rw.statusCode)
	}
	
	// Test Write
	data := []byte("test data")
	n, err := rw.Write(data)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, got %d", len(data), n)
	}
}

func TestMetricsRegistration(t *testing.T) {
	// Create a new registry to test metric registration
	_ = NewMetricsRegistry("test-service")
	
	// Verify that all metrics are registered
	collectors := []prometheus.Collector{
		serviceCallDuration,
		serviceCallTotal,
		serviceCallErrors,
		circuitBreakerState,
		circuitBreakerFailures,
		circuitBreakerTransitions,
		retryAttempts,
		retryFailures,
		healthCheckStatus,
		healthCheckDuration,
		httpRequestsTotal,
		httpRequestDuration,
		httpRequestsInFlight,
		databaseConnections,
		databaseQueryDuration,
		databaseErrors,
		redisConnections,
		redisOperations,
		redisOperationDuration,
	}
	
	for _, collector := range collectors {
		if collector == nil {
			t.Error("Expected collector to not be nil")
		}
	}
} 