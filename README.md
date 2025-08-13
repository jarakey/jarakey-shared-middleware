# Jarakey Shared Middleware Package

A comprehensive Go middleware package providing essential service communication patterns for microservices architecture. This package implements industry-standard resilience, observability, and monitoring patterns that can be easily integrated into any Go service.

## 🚀 Features

### 1. Circuit Breaker Pattern
- **Location**: `middleware/circuit_breaker.go`
- **Purpose**: Prevents cascading failures in distributed systems
- **Features**:
  - Three states: CLOSED, OPEN, HALF_OPEN with automatic transitions
  - Configurable failure thresholds, timeouts, and reset intervals
  - Thread-safe implementation with proper locking strategy
  - Deadlock prevention through simplified locking
  - Statistics and monitoring capabilities

### 2. Retry Logic with Exponential Backoff
- **Location**: `middleware/retry.go`
- **Purpose**: Automatically retry failed operations with intelligent backoff
- **Features**:
  - Multiple backoff strategies (exponential, linear, constant, fibonacci)
  - Configurable retry attempts and delays
  - Context-aware cancellation support
  - Retryable error detection and handling
  - Jitter support for distributed systems

### 3. Enhanced Health Checks
- **Location**: `middleware/health_check.go`
- **Purpose**: Comprehensive service and dependency health monitoring
- **Features**:
  - Three health statuses: healthy, degraded, unhealthy
  - Concurrent health checks with timeout support
  - HTTP handler for health check endpoints
  - Predefined checks for common dependencies (HTTP, Database, Redis)
  - Custom health check support

### 4. Request Correlation IDs
- **Location**: `middleware/correlation.go`
- **Purpose**: Distributed tracing and request correlation across services
- **Features**:
  - Multiple header support (X-Correlation-ID, X-Request-ID, X-Trace-ID, X-Span-ID)
  - Automatic generation and propagation of correlation IDs
  - HTTP and Gin middleware support
  - User and session tracking capabilities
  - Logging context integration

### 5. Prometheus Metrics
- **Location**: `middleware/metrics.go`
- **Purpose**: Comprehensive application and infrastructure monitoring
- **Features**:
  - Service call metrics (duration, success/failure rates)
  - Circuit breaker state and failure metrics
  - Retry attempt and failure metrics
  - Health check status metrics
  - HTTP request metrics (duration, status codes)
  - Database and Redis operation metrics
  - HTTP and Gin middleware integration
  - Prometheus endpoint for metric scraping

## 📦 Installation

### Option 1: Direct Go Module Reference
```bash
go get github.com/jarakey/jarakey-shared-middleware@latest
```

### Option 2: Local Development
```bash
# Clone the repository
git clone https://github.com/jarakey/jarakey-shared-middleware.git
cd jarakey-shared-middleware

# Install dependencies
go mod download
go mod tidy

# Run tests
go test -v ./...
```

## 🔧 Usage Examples

### Circuit Breaker
```go
import "github.com/jarakey/jarakey-shared-middleware/middleware"

// Create circuit breaker with custom configuration
config := &middleware.CircuitBreakerConfig{
    MaxFailures:  5,
    Timeout:      30 * time.Second,
    ResetTimeout: 60 * time.Second,
}
cb := middleware.NewCircuitBreaker(config)

// Execute operation with circuit breaker protection
err := cb.Execute(context.Background(), func() error {
    // Your service call here
    return callExternalService()
})

if err != nil {
    log.Printf("Operation failed: %v", err)
}
```

### Retry Logic
```go
import "github.com/jarakey/jarakey-shared-middleware/middleware"

// Use default retry configuration
retryConfig := middleware.DefaultRetryConfig()

// Retry with exponential backoff
err := retryConfig.Retry(context.Background(), func() error {
    return callExternalService()
})

// Custom retry configuration
customConfig := &middleware.RetryConfig{
    MaxAttempts:  3,
    InitialDelay: 100 * time.Millisecond,
    MaxDelay:     5 * time.Second,
    Backoff:      middleware.ExponentialBackoff,
}
```

### Health Checks
```go
import "github.com/jarakey/jarakey-shared-middleware/middleware"

// Create health checker
checker := middleware.NewHealthChecker()
checker.SetTimeout(10 * time.Second)

// Add health checks
checker.AddCheck("database", middleware.DatabaseHealthCheck(db))
checker.AddCheck("redis", middleware.RedisHealthCheck(redisClient))
checker.AddCheck("external-api", middleware.HTTPHealthCheck("https://api.example.com/health"))

// Check health
health := checker.CheckHealth(context.Background())

// Use as HTTP handler
http.Handle("/health", checker.HTTPHandler())
```

### Correlation IDs
```go
import "github.com/jarakey/jarakey-shared-middleware/middleware"

// For standard HTTP
http.HandleFunc("/api", middleware.CorrelationMiddleware(func(w http.ResponseWriter, r *http.Request) {
    // Extract correlation ID
    corrID := middleware.GetCorrelationID(r.Context())
    log.Printf("Processing request with correlation ID: %s", corrID)
}))

// For Gin framework
router := gin.New()
router.Use(middleware.GinCorrelationMiddleware())

router.GET("/api", func(c *gin.Context) {
    corrID := middleware.GetCorrelationID(c.Request.Context())
    c.JSON(200, gin.H{"correlation_id": corrID})
})
```

### Prometheus Metrics
```go
import "github.com/jarakey/jarakey-shared-middleware/middleware"

// Create metrics registry
registry := middleware.NewMetricsRegistry("my-service")

// Record metrics
registry.RecordServiceCall("external-api", "GET", 150*time.Millisecond, nil)
registry.RecordHTTPRequest("GET", "/api/users", 200, 25*time.Millisecond)

// Use as HTTP handler for Prometheus scraping
http.Handle("/metrics", registry.HTTPHandler())

// Use as Gin middleware
router.Use(middleware.GinMetricsMiddleware())
```

## 🏗️ Architecture

### Package Structure
```
shared/
├── go.mod
├── go.sum
├── README.md
└── middleware/
    ├── circuit_breaker.go
    ├── circuit_breaker_test.go
    ├── retry.go
    ├── retry_test.go
    ├── health_check.go
    ├── health_check_test.go
    ├── correlation.go
    ├── correlation_test.go
    ├── metrics.go
    └── metrics_test.go
```

### Integration Points
- **HTTP Middleware**: Standard `net/http` middleware for each pattern
- **Gin Framework**: Dedicated middleware for Gin web framework
- **Standalone Usage**: Direct function calls for custom implementations
- **Configuration**: Environment-based configuration support
- **Monitoring**: Prometheus metrics integration

## 🧪 Testing

### Test Coverage
All middleware packages include comprehensive test coverage:

- **Circuit Breaker**: 14/14 tests passing
- **Retry Logic**: 32/32 tests passing  
- **Health Checks**: 47/47 tests passing
- **Correlation IDs**: 61/61 tests passing
- **Prometheus Metrics**: 75/75 tests passing

### Running Tests
```bash
# Run all tests
go test -v ./...

# Run specific package tests
go test -v ./middleware

# Run with coverage
go test -v -cover ./...
```

## 📊 Metrics and Monitoring

### Available Metrics
- **Service Calls**: Duration, success/failure rates
- **Circuit Breakers**: State changes, failure counts
- **Retry Attempts**: Attempt counts, failure rates
- **Health Checks**: Status changes, response times
- **HTTP Requests**: Duration, status codes, method distribution
- **Database Operations**: Query duration, connection status
- **Redis Operations**: Operation duration, connection status

### Prometheus Endpoint
Expose metrics at `/metrics` endpoint for Prometheus scraping:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'jarakey-services'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics
```

## 🔒 Security and Compliance

### Law.md Compliance
All implementations follow the security and coding standards outlined in `law.md`:
- Input validation and sanitization
- Secure error handling
- Proper logging practices
- Performance considerations
- Memory safety

### Best Practices
- Thread-safe implementations
- Proper error handling and propagation
- Context-aware operations
- Resource cleanup and management
- Comprehensive logging and monitoring

## 🚀 Deployment

### Publishing to GitHub
1. **Create Repository**:
   ```bash
   # Create new GitHub repository
   # Clone and push code
   git clone https://github.com/jarakey/jarakey-shared-middleware.git
   cd jarakey-shared-middleware
   git add .
   git commit -m "Initial commit: Shared middleware package"
   git push origin main
   ```

2. **Version Tagging**:
   ```bash
   # Create semantic version tag
   git tag v1.0.0
   git push origin v1.0.0
   ```

3. **Update Services**:
   ```bash
   # In each service, update go.mod
   go get github.com/jarakey/jarakey-shared-middleware@v1.0.0
   go mod tidy
   ```

### Environment Configuration
```bash
# Required environment variables
export SERVICE_NAME="my-service"
export METRICS_PORT="9090"
export HEALTH_CHECK_TIMEOUT="30s"
export CIRCUIT_BREAKER_MAX_FAILURES="5"
export CIRCUIT_BREAKER_RESET_TIMEOUT="60s"
```

## 🔄 Updates and Maintenance

### Version Management
- Follow semantic versioning (MAJOR.MINOR.PATCH)
- Breaking changes require major version bump
- Backward compatibility maintained within major versions
- Automated testing on each commit

### Contributing
1. Fork the repository
2. Create feature branch
3. Implement changes with tests
4. Submit pull request
5. Code review and approval
6. Merge and release

## 📚 References

### Documentation
- [Go Modules Documentation](https://golang.org/doc/modules)
- [Prometheus Go Client](https://pkg.go.dev/github.com/prometheus/client_golang)
- [Gin Web Framework](https://gin-gonic.com/)
- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)

### Standards
- [OpenTelemetry](https://opentelemetry.io/) for distributed tracing
- [Prometheus](https://prometheus.io/) for metrics collection
- [Health Check Standards](https://datatracker.ietf.org/doc/html/rfc7231#section-4.3.7)

## 🤝 Support

### Issues and Questions
- GitHub Issues: [Repository Issues](https://github.com/jarakey/jarakey-shared-middleware/issues)
- Documentation: This README and inline code comments
- Examples: See usage examples above

### Roadmap
- [ ] Additional backoff strategies
- [ ] More health check types
- [ ] OpenTelemetry integration
- [ ] GraphQL support
- [ ] gRPC middleware support

---

**Version**: 1.0.0  
**Go Version**: 1.21+  
**License**: MIT  
**Maintainer**: Jarakey Team 