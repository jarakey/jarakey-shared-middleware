package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// Service call metrics
	serviceCallDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "service_call_duration_seconds",
			Help:    "Duration of external service calls in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "status"},
	)
	
	serviceCallTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_calls_total",
			Help: "Total number of external service calls",
		},
		[]string{"service", "method", "status"},
	)
	
	serviceCallErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_call_errors_total",
			Help: "Total number of external service call errors",
		},
		[]string{"service", "method", "error_type"},
	)
	
	// Circuit breaker metrics
	circuitBreakerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Current state of circuit breakers (0=closed, 1=half-open, 2=open)",
		},
		[]string{"service"},
	)
	
	circuitBreakerFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_failures_total",
			Help: "Total number of circuit breaker failures",
		},
		[]string{"service"},
	)
	
	circuitBreakerTransitions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_transitions_total",
			Help: "Total number of circuit breaker state transitions",
		},
		[]string{"service", "from_state", "to_state"},
	)
	
	// Retry metrics
	retryAttempts = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "retry_attempts_total",
			Help: "Total number of retry attempts",
		},
		[]string{"service", "method"},
	)
	
	retryFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "retry_failures_total",
			Help: "Total number of retry failures after all attempts",
		},
		[]string{"service", "method"},
	)
	
	// Health check metrics
	healthCheckStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "health_check_status",
			Help: "Health check status (0=unhealthy, 1=degraded, 2=healthy)",
		},
		[]string{"service", "dependency"},
	)
	
	healthCheckDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "health_check_duration_seconds",
			Help:    "Duration of health checks in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "dependency"},
	)
	
	// HTTP request metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)
	
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	
	httpRequestsInFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
		[]string{"method", "endpoint"},
	)
	
	// Database metrics
	databaseConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "database_connections",
			Help: "Current number of database connections",
		},
		[]string{"service", "database"},
	)
	
	databaseQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "database", "query_type"},
	)
	
	databaseErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"service", "database", "error_type"},
	)
	
	// Redis metrics
	redisConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "redis_connections",
			Help: "Current number of Redis connections",
		},
		[]string{"service"},
	)
	
	redisOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"service", "operation", "status"},
	)
	
	redisOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_operation_duration_seconds",
			Help:    "Duration of Redis operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "operation"},
	)
)

// MetricsRegistry holds all metrics for a service
type MetricsRegistry struct {
	serviceName string
	metrics     map[string]prometheus.Collector
}

// NewMetricsRegistry creates a new metrics registry
func NewMetricsRegistry(serviceName string) *MetricsRegistry {
	registry := &MetricsRegistry{
		serviceName: serviceName,
		metrics:     make(map[string]prometheus.Collector),
	}
	
	// Register all metrics
	registry.registerMetrics()
	
	return registry
}

// registerMetrics registers all Prometheus metrics
func (mr *MetricsRegistry) registerMetrics() {
	// Service call metrics
	registerIfNotExists(serviceCallDuration)
	registerIfNotExists(serviceCallTotal)
	registerIfNotExists(serviceCallErrors)
	
	// Circuit breaker metrics
	registerIfNotExists(circuitBreakerState)
	registerIfNotExists(circuitBreakerFailures)
	registerIfNotExists(circuitBreakerTransitions)
	
	// Retry metrics
	registerIfNotExists(retryAttempts)
	registerIfNotExists(retryFailures)
	
	// Health check metrics
	registerIfNotExists(healthCheckStatus)
	registerIfNotExists(healthCheckDuration)
	
	// HTTP request metrics
	registerIfNotExists(httpRequestsTotal)
	registerIfNotExists(httpRequestDuration)
	registerIfNotExists(httpRequestsInFlight)
	
	// Database metrics
	registerIfNotExists(databaseConnections)
	registerIfNotExists(databaseQueryDuration)
	registerIfNotExists(databaseErrors)
	
	// Redis metrics
	registerIfNotExists(redisConnections)
	registerIfNotExists(redisOperations)
	registerIfNotExists(redisOperationDuration)
}

// registerIfNotExists registers a metric only if it's not already registered
func registerIfNotExists(collector prometheus.Collector) {
	// Try to register, ignore if already registered
	prometheus.Register(collector)
}

// RecordServiceCall records metrics for a service call
func (mr *MetricsRegistry) RecordServiceCall(service, method, status string, duration time.Duration) {
	serviceCallDuration.WithLabelValues(service, method, status).Observe(duration.Seconds())
	serviceCallTotal.WithLabelValues(service, method, status).Inc()
	
	if status != "success" {
		serviceCallErrors.WithLabelValues(service, method, status).Inc()
	}
}

// RecordCircuitBreakerState records the current state of a circuit breaker
func (mr *MetricsRegistry) RecordCircuitBreakerState(service string, state CircuitBreakerState) {
	var stateValue float64
	switch state {
	case StateClosed:
		stateValue = 0
	case StateHalfOpen:
		stateValue = 1
	case StateOpen:
		stateValue = 2
	}
	
	circuitBreakerState.WithLabelValues(service).Set(stateValue)
}

// RecordCircuitBreakerFailure records a circuit breaker failure
func (mr *MetricsRegistry) RecordCircuitBreakerFailure(service string) {
	circuitBreakerFailures.WithLabelValues(service).Inc()
}

// RecordCircuitBreakerTransition records a circuit breaker state transition
func (mr *MetricsRegistry) RecordCircuitBreakerTransition(service string, fromState, toState CircuitBreakerState) {
	circuitBreakerTransitions.WithLabelValues(service, fromState.String(), toState.String()).Inc()
}

// RecordRetryAttempt records a retry attempt
func (mr *MetricsRegistry) RecordRetryAttempt(service, method string) {
	retryAttempts.WithLabelValues(service, method).Inc()
}

// RecordRetryFailure records a retry failure after all attempts
func (mr *MetricsRegistry) RecordRetryFailure(service, method string) {
	retryFailures.WithLabelValues(service, method).Inc()
}

// RecordHealthCheck records health check metrics
func (mr *MetricsRegistry) RecordHealthCheck(dependency string, status HealthStatus, duration time.Duration) {
	var statusValue float64
	switch status {
	case StatusUnhealthy:
		statusValue = 0
	case StatusDegraded:
		statusValue = 1
	case StatusHealthy:
		statusValue = 2
	}
	
	healthCheckStatus.WithLabelValues(mr.serviceName, dependency).Set(statusValue)
	healthCheckDuration.WithLabelValues(mr.serviceName, dependency).Observe(duration.Seconds())
}

// RecordHTTPRequest records HTTP request metrics
func (mr *MetricsRegistry) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration) {
	httpRequestsTotal.WithLabelValues(method, endpoint, strconv.Itoa(statusCode)).Inc()
	httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordHTTPRequestStart records the start of an HTTP request
func (mr *MetricsRegistry) RecordHTTPRequestStart(method, endpoint string) {
	httpRequestsInFlight.WithLabelValues(method, endpoint).Inc()
}

// RecordHTTPRequestEnd records the end of an HTTP request
func (mr *MetricsRegistry) RecordHTTPRequestEnd(method, endpoint string) {
	httpRequestsInFlight.WithLabelValues(method, endpoint).Dec()
}

// RecordDatabaseConnection records database connection metrics
func (mr *MetricsRegistry) RecordDatabaseConnection(database string, count int) {
	databaseConnections.WithLabelValues(mr.serviceName, database).Set(float64(count))
}

// RecordDatabaseQuery records database query metrics
func (mr *MetricsRegistry) RecordDatabaseQuery(database, queryType string, duration time.Duration) {
	databaseQueryDuration.WithLabelValues(mr.serviceName, database, queryType).Observe(duration.Seconds())
}

// RecordDatabaseError records database error metrics
func (mr *MetricsRegistry) RecordDatabaseError(database, errorType string) {
	databaseErrors.WithLabelValues(mr.serviceName, database, errorType).Inc()
}

// RecordRedisConnection records Redis connection metrics
func (mr *MetricsRegistry) RecordRedisConnection(count int) {
	redisConnections.WithLabelValues(mr.serviceName).Set(float64(count))
}

// RecordRedisOperation records Redis operation metrics
func (mr *MetricsRegistry) RecordRedisOperation(operation, status string, duration time.Duration) {
	redisOperations.WithLabelValues(mr.serviceName, operation, status).Inc()
	redisOperationDuration.WithLabelValues(mr.serviceName, operation).Observe(duration.Seconds())
}

// HTTPHandler returns an HTTP handler for the metrics endpoint
func (mr *MetricsRegistry) HTTPHandler() http.Handler {
	return promhttp.Handler()
}

// MetricsMiddleware creates middleware for recording HTTP request metrics
func (mr *MetricsRegistry) MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			
			// Record request start
			mr.RecordHTTPRequestStart(r.Method, r.URL.Path)
			
			// Create response writer wrapper to capture status code
			wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: 200}
			
			// Process request
			next.ServeHTTP(wrappedWriter, r)
			
			// Record request end
			mr.RecordHTTPRequestEnd(r.Method, r.URL.Path)
			
			// Record request metrics
			duration := time.Since(start)
			mr.RecordHTTPRequest(r.Method, r.URL.Path, wrappedWriter.statusCode, duration)
		})
	}
}

// GinMetricsMiddleware creates middleware for Gin framework
func (mr *MetricsRegistry) GinMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Record request start
		mr.RecordHTTPRequestStart(c.Request.Method, c.Request.URL.Path)
		
		// Process request
		c.Next()
		
		// Record request end
		mr.RecordHTTPRequestEnd(c.Request.Method, c.Request.URL.Path)
		
		// Record request metrics
		duration := time.Since(start)
		mr.RecordHTTPRequest(c.Request.Method, c.Request.URL.Path, c.Writer.Status(), duration)
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}

// GetMetricsSummary returns a summary of all metrics
func (mr *MetricsRegistry) GetMetricsSummary() map[string]interface{} {
	return map[string]interface{}{
		"service": mr.serviceName,
		"metrics": map[string]interface{}{
			"service_calls": map[string]interface{}{
				"duration": "histogram",
				"total":    "counter",
				"errors":   "counter",
			},
			"circuit_breaker": map[string]interface{}{
				"state":        "gauge",
				"failures":     "counter",
				"transitions":  "counter",
			},
			"retry": map[string]interface{}{
				"attempts": "counter",
				"failures": "counter",
			},
			"health_check": map[string]interface{}{
				"status":   "gauge",
				"duration": "histogram",
			},
			"http_requests": map[string]interface{}{
				"total":     "counter",
				"duration":  "histogram",
				"in_flight": "gauge",
			},
			"database": map[string]interface{}{
				"connections": "gauge",
				"query_duration": "histogram",
				"errors":     "counter",
			},
			"redis": map[string]interface{}{
				"connections": "gauge",
				"operations":  "counter",
				"duration":    "histogram",
			},
		},
	}
} 