package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestCorrelationContext(t *testing.T) {
	corrCtx := &CorrelationContext{
		CorrelationID: "test-correlation-id",
		RequestID:     "test-request-id",
		TraceID:       "test-trace-id",
		SpanID:        "test-span-id",
		UserID:        "test-user-id",
		SessionID:     "test-session-id",
	}
	
	if corrCtx.CorrelationID != "test-correlation-id" {
		t.Errorf("Expected correlation ID 'test-correlation-id', got %s", corrCtx.CorrelationID)
	}
	
	if corrCtx.RequestID != "test-request-id" {
		t.Errorf("Expected request ID 'test-request-id', got %s", corrCtx.RequestID)
	}
	
	if corrCtx.TraceID != "test-trace-id" {
		t.Errorf("Expected trace ID 'test-trace-id', got %s", corrCtx.TraceID)
	}
	
	if corrCtx.SpanID != "test-span-id" {
		t.Errorf("Expected span ID 'test-span-id', got %s", corrCtx.SpanID)
	}
	
	if corrCtx.UserID != "test-user-id" {
		t.Errorf("Expected user ID 'test-user-id', got %s", corrCtx.UserID)
	}
	
	if corrCtx.SessionID != "test-session-id" {
		t.Errorf("Expected session ID 'test-session-id', got %s", corrCtx.SessionID)
	}
}

func TestCorrelationContextString(t *testing.T) {
	corrCtx := &CorrelationContext{
		CorrelationID: "test-correlation-id",
		RequestID:     "test-request-id",
		TraceID:       "test-trace-id",
		SpanID:        "test-span-id",
		UserID:        "test-user-id",
		SessionID:     "test-session-id",
	}
	
	expected := "correlation_id=test-correlation-id, request_id=test-request-id, trace_id=test-trace-id, span_id=test-span-id, user_id=test-user-id, session_id=test-session-id"
	if corrCtx.String() != expected {
		t.Errorf("Expected string representation '%s', got '%s'", expected, corrCtx.String())
	}
}

func TestCorrelationContextIsEmpty(t *testing.T) {
	// Test empty context
	emptyCtx := &CorrelationContext{}
	if !emptyCtx.IsEmpty() {
		t.Error("Expected empty context to return true")
	}
	
	// Test non-empty context
	nonEmptyCtx := &CorrelationContext{
		CorrelationID: "test-id",
	}
	if nonEmptyCtx.IsEmpty() {
		t.Error("Expected non-empty context to return false")
	}
}

func TestCorrelationMiddleware(t *testing.T) {
	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrCtx := GetCorrelationContext(r.Context())
		if corrCtx == nil {
			t.Error("Expected correlation context to be set")
			return
		}
		
		if corrCtx.CorrelationID == "" {
			t.Error("Expected correlation ID to be set")
		}
		
		if corrCtx.RequestID == "" {
			t.Error("Expected request ID to be set")
		}
		
		// Check response headers
		if w.Header().Get(CorrelationIDHeader) == "" {
			t.Error("Expected correlation ID header to be set in response")
		}
		
		if w.Header().Get(RequestIDHeader) == "" {
			t.Error("Expected request ID header to be set in response")
		}
	})
	
	// Create middleware
	middleware := CorrelationMiddleware()
	wrappedHandler := middleware(handler)
	
	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	wrappedHandler.ServeHTTP(w, req)
	
	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestCorrelationMiddlewareWithExistingHeaders(t *testing.T) {
	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrCtx := GetCorrelationContext(r.Context())
		if corrCtx == nil {
			t.Error("Expected correlation context to be set")
			return
		}
		
		if corrCtx.CorrelationID != "existing-correlation-id" {
			t.Errorf("Expected correlation ID 'existing-correlation-id', got %s", corrCtx.CorrelationID)
		}
		
		if corrCtx.RequestID != "existing-request-id" {
			t.Errorf("Expected request ID 'existing-request-id', got %s", corrCtx.RequestID)
		}
		
		if corrCtx.TraceID != "existing-trace-id" {
			t.Errorf("Expected trace ID 'existing-trace-id', got %s", corrCtx.TraceID)
		}
		
		if corrCtx.SpanID != "existing-span-id" {
			t.Errorf("Expected span ID 'existing-span-id', got %s", corrCtx.SpanID)
		}
	})
	
	// Create middleware
	middleware := CorrelationMiddleware()
	wrappedHandler := middleware(handler)
	
	// Create test request with existing headers
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(CorrelationIDHeader, "existing-correlation-id")
	req.Header.Set(RequestIDHeader, "existing-request-id")
	req.Header.Set(TraceIDHeader, "existing-trace-id")
	req.Header.Set(SpanIDHeader, "existing-span-id")
	
	w := httptest.NewRecorder()
	
	// Call handler
	wrappedHandler.ServeHTTP(w, req)
	
	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestGinCorrelationMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create Gin router
	r := gin.New()
	r.Use(GinCorrelationMiddleware())
	
	r.GET("/test", func(c *gin.Context) {
		corrCtx := GetCorrelationContext(c.Request.Context())
		if corrCtx == nil {
			t.Error("Expected correlation context to be set")
			return
		}
		
		if corrCtx.CorrelationID == "" {
			t.Error("Expected correlation ID to be set")
		}
		
		if corrCtx.RequestID == "" {
			t.Error("Expected request ID to be set")
		}
		
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// Call handler
	r.ServeHTTP(w, req)
	
	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
	
	// Check response headers
	if w.Header().Get(CorrelationIDHeader) == "" {
		t.Error("Expected correlation ID header to be set in response")
	}
	
	if w.Header().Get(RequestIDHeader) == "" {
		t.Error("Expected request ID header to be set in response")
	}
}

func TestExtractCorrelationID(t *testing.T) {
	// Test with correlation ID header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(CorrelationIDHeader, "test-correlation-id")
	
	correlationID := extractCorrelationID(req)
	if correlationID != "test-correlation-id" {
		t.Errorf("Expected correlation ID 'test-correlation-id', got %s", correlationID)
	}
	
	// Test with request ID header (fallback)
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(RequestIDHeader, "test-request-id")
	
	correlationID = extractCorrelationID(req)
	if correlationID != "test-request-id" {
		t.Errorf("Expected correlation ID 'test-request-id', got %s", correlationID)
	}
	
	// Test with trace ID header (fallback)
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(TraceIDHeader, "test-trace-id")
	
	correlationID = extractCorrelationID(req)
	if correlationID != "test-trace-id" {
		t.Errorf("Expected correlation ID 'test-trace-id', got %s", correlationID)
	}
	
	// Test with no headers (should generate new UUID)
	req = httptest.NewRequest("GET", "/test", nil)
	
	correlationID = extractCorrelationID(req)
	if correlationID == "" {
		t.Error("Expected correlation ID to be generated")
	}
	
	// Validate that it's a valid UUID
	if _, err := uuid.Parse(correlationID); err != nil {
		t.Errorf("Expected generated correlation ID to be valid UUID, got error: %v", err)
	}
}

func TestGetCorrelationContext(t *testing.T) {
	// Test with no correlation context
	ctx := context.Background()
	corrCtx := GetCorrelationContext(ctx)
	if corrCtx != nil {
		t.Error("Expected nil correlation context for empty context")
	}
	
	// Test with correlation context
	testCtx := &CorrelationContext{
		CorrelationID: "test-id",
	}
	ctx = context.WithValue(context.Background(), "correlation_context", testCtx)
	
	corrCtx = GetCorrelationContext(ctx)
	if corrCtx == nil {
		t.Error("Expected correlation context to be retrieved")
	}
	
	if corrCtx.CorrelationID != "test-id" {
		t.Errorf("Expected correlation ID 'test-id', got %s", corrCtx.CorrelationID)
	}
}

func TestGetCorrelationID(t *testing.T) {
	// Test with no correlation context
	ctx := context.Background()
	correlationID := GetCorrelationID(ctx)
	if correlationID != "" {
		t.Errorf("Expected empty correlation ID, got %s", correlationID)
	}
	
	// Test with correlation context
	testCtx := &CorrelationContext{
		CorrelationID: "test-id",
	}
	ctx = context.WithValue(context.Background(), "correlation_context", testCtx)
	
	correlationID = GetCorrelationID(ctx)
	if correlationID != "test-id" {
		t.Errorf("Expected correlation ID 'test-id', got %s", correlationID)
	}
}

func TestSetUserID(t *testing.T) {
	// Test with no correlation context
	ctx := context.Background()
	newCtx := SetUserID(ctx, "test-user-id")
	
	// Should return original context unchanged
	if newCtx != ctx {
		t.Error("Expected original context to be returned unchanged")
	}
	
	// Test with correlation context
	testCtx := &CorrelationContext{
		CorrelationID: "test-id",
	}
	ctx = context.WithValue(context.Background(), "correlation_context", testCtx)
	
	newCtx = SetUserID(ctx, "test-user-id")
	
	// Should return new context with user ID set
	corrCtx := GetCorrelationContext(newCtx)
	if corrCtx == nil {
		t.Error("Expected correlation context to be retrieved")
	}
	
	if corrCtx.UserID != "test-user-id" {
		t.Errorf("Expected user ID 'test-user-id', got %s", corrCtx.UserID)
	}
}

func TestSetSessionID(t *testing.T) {
	// Test with no correlation context
	ctx := context.Background()
	newCtx := SetSessionID(ctx, "test-session-id")
	
	// Should return original context unchanged
	if newCtx != ctx {
		t.Error("Expected original context to be returned unchanged")
	}
	
	// Test with correlation context
	testCtx := &CorrelationContext{
		CorrelationID: "test-id",
	}
	ctx = context.WithValue(context.Background(), "correlation_context", testCtx)
	
	newCtx = SetSessionID(ctx, "test-session-id")
	
	// Should return new context with session ID set
	corrCtx := GetCorrelationContext(newCtx)
	if corrCtx == nil {
		t.Error("Expected correlation context to be retrieved")
	}
	
	if corrCtx.SessionID != "test-session-id" {
		t.Errorf("Expected session ID 'test-session-id', got %s", corrCtx.SessionID)
	}
}

func TestWithCorrelationContext(t *testing.T) {
	ctx := context.Background()
	
	newCtx := WithCorrelationContext(ctx, "test-correlation-id", "test-request-id", "test-trace-id", "test-span-id")
	
	corrCtx := GetCorrelationContext(newCtx)
	if corrCtx == nil {
		t.Error("Expected correlation context to be created")
	}
	
	if corrCtx.CorrelationID != "test-correlation-id" {
		t.Errorf("Expected correlation ID 'test-correlation-id', got %s", corrCtx.CorrelationID)
	}
	
	if corrCtx.RequestID != "test-request-id" {
		t.Errorf("Expected request ID 'test-request-id', got %s", corrCtx.RequestID)
	}
	
	if corrCtx.TraceID != "test-trace-id" {
		t.Errorf("Expected trace ID 'test-trace-id', got %s", corrCtx.TraceID)
	}
	
	if corrCtx.SpanID != "test-span-id" {
		t.Errorf("Expected span ID 'test-span-id', got %s", corrCtx.SpanID)
	}
}

func TestPropagateCorrelationHeaders(t *testing.T) {
	// Create correlation context
	testCtx := &CorrelationContext{
		CorrelationID: "test-correlation-id",
		RequestID:     "test-request-id",
		TraceID:       "test-trace-id",
		SpanID:        "test-span-id",
	}
	ctx := context.WithValue(context.Background(), "correlation_context", testCtx)
	
	// Create HTTP request
	req := httptest.NewRequest("GET", "/test", nil)
	
	// Propagate headers
	PropagateCorrelationHeaders(req, ctx)
	
	// Check headers
	if req.Header.Get(CorrelationIDHeader) != "test-correlation-id" {
		t.Errorf("Expected correlation ID header 'test-correlation-id', got %s", req.Header.Get(CorrelationIDHeader))
	}
	
	if req.Header.Get(RequestIDHeader) != "test-request-id" {
		t.Errorf("Expected request ID header 'test-request-id', got %s", req.Header.Get(RequestIDHeader))
	}
	
	if req.Header.Get(TraceIDHeader) != "test-trace-id" {
		t.Errorf("Expected trace ID header 'test-trace-id', got %s", req.Header.Get(TraceIDHeader))
	}
	
	if req.Header.Get(SpanIDHeader) != "test-span-id" {
		t.Errorf("Expected span ID header 'test-span-id', got %s", req.Header.Get(SpanIDHeader))
	}
}

func TestLogCorrelationContext(t *testing.T) {
	// Test with no correlation context
	ctx := context.Background()
	fields := LogCorrelationContext(ctx)
	if fields != nil {
		t.Error("Expected nil fields for empty context")
	}
	
	// Test with correlation context
	testCtx := &CorrelationContext{
		CorrelationID: "test-correlation-id",
		RequestID:     "test-request-id",
		TraceID:       "test-trace-id",
		SpanID:        "test-span-id",
		UserID:        "test-user-id",
		SessionID:     "test-session-id",
	}
	ctx = context.WithValue(context.Background(), "correlation_context", testCtx)
	
	fields = LogCorrelationContext(ctx)
	if fields == nil {
		t.Error("Expected fields to be returned")
	}
	
	if fields["correlation_id"] != "test-correlation-id" {
		t.Errorf("Expected correlation_id 'test-correlation-id', got %v", fields["correlation_id"])
	}
	
	if fields["request_id"] != "test-request-id" {
		t.Errorf("Expected request_id 'test-request-id', got %v", fields["request_id"])
	}
	
	if fields["trace_id"] != "test-trace-id" {
		t.Errorf("Expected trace_id 'test-trace-id', got %v", fields["trace_id"])
	}
	
	if fields["span_id"] != "test-span-id" {
		t.Errorf("Expected span_id 'test-span-id', got %v", fields["span_id"])
	}
	
	if fields["user_id"] != "test-user-id" {
		t.Errorf("Expected user_id 'test-user-id', got %v", fields["user_id"])
	}
	
	if fields["session_id"] != "test-session-id" {
		t.Errorf("Expected session_id 'test-session-id', got %v", fields["session_id"])
	}
}

func TestValidateCorrelationID(t *testing.T) {
	// Test valid UUID
	validUUID := uuid.New().String()
	if !ValidateCorrelationID(validUUID) {
		t.Errorf("Expected UUID '%s' to be valid", validUUID)
	}
	
	// Test valid trace ID (32 hex characters)
	validTraceID := "1234567890abcdef1234567890abcdef"
	if !ValidateCorrelationID(validTraceID) {
		t.Errorf("Expected trace ID '%s' to be valid", validTraceID)
	}
	
	// Test valid span ID (16 hex characters)
	validSpanID := "1234567890abcdef"
	if !ValidateCorrelationID(validSpanID) {
		t.Errorf("Expected span ID '%s' to be valid", validSpanID)
	}
	
	// Test invalid correlation IDs
	invalidIDs := []string{
		"",
		"invalid-uuid",
		"1234567890abcdef1234567890abcdeg", // Invalid hex character
		"1234567890abcde", // Wrong length
	}
	
	for _, invalidID := range invalidIDs {
		if ValidateCorrelationID(invalidID) {
			t.Errorf("Expected correlation ID '%s' to be invalid", invalidID)
		}
	}
}

func TestSanitizeCorrelationID(t *testing.T) {
	// Test empty string
	if SanitizeCorrelationID("") != "" {
		t.Error("Expected empty string to be returned unchanged")
	}
	
	// Test valid correlation ID
	validID := "test-correlation-id-123"
	sanitized := SanitizeCorrelationID(validID)
	if sanitized != validID {
		t.Errorf("Expected valid ID to be unchanged, got '%s'", sanitized)
	}
	
	// Test ID with invalid characters
	invalidID := "test@correlation#id$123"
	sanitized = SanitizeCorrelationID(invalidID)
	expected := "testcorrelationid123"
	if sanitized != expected {
		t.Errorf("Expected sanitized ID '%s', got '%s'", expected, sanitized)
	}
	
	// Test ID that's too long
	longID := strings.Repeat("a", 70)
	sanitized = SanitizeCorrelationID(longID)
	if len(sanitized) <= 64 {
		t.Error("Expected long ID to be truncated")
	}
	
	if !strings.HasSuffix(sanitized, "...") {
		t.Error("Expected truncated ID to end with '...'")
	}
} 