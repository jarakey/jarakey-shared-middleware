package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// CorrelationIDHeader is the standard header name for correlation IDs
	CorrelationIDHeader = "X-Correlation-ID"
	
	// RequestIDHeader is an alternative header name for request IDs
	RequestIDHeader = "X-Request-ID"
	
	// TraceIDHeader is the OpenTelemetry trace ID header
	TraceIDHeader = "X-Trace-ID"
	
	// SpanIDHeader is the OpenTelemetry span ID header
	SpanIDHeader = "X-Span-ID"
)

// CorrelationContext holds correlation information for a request
type CorrelationContext struct {
	CorrelationID string
	RequestID     string
	TraceID       string
	SpanID        string
	UserID        string
	SessionID     string
}

// String returns a string representation of the correlation context
func (cc *CorrelationContext) String() string {
	return fmt.Sprintf("correlation_id=%s, request_id=%s, trace_id=%s, span_id=%s, user_id=%s, session_id=%s",
		cc.CorrelationID, cc.RequestID, cc.TraceID, cc.SpanID, cc.UserID, cc.SessionID)
}

// IsEmpty checks if the correlation context has any meaningful data
func (cc *CorrelationContext) IsEmpty() bool {
	return cc.CorrelationID == "" && cc.RequestID == "" && cc.TraceID == "" && 
		   cc.SpanID == "" && cc.UserID == "" && cc.SessionID == ""
}

// CorrelationMiddleware creates middleware for handling correlation IDs
func CorrelationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract or generate correlation ID
			correlationID := extractCorrelationID(r)
			
			// Extract other correlation headers
			requestID := r.Header.Get(RequestIDHeader)
			traceID := r.Header.Get(TraceIDHeader)
			spanID := r.Header.Get(SpanIDHeader)
			
			// Generate request ID if not provided
			if requestID == "" {
				requestID = uuid.New().String()
			}
			
			// Create correlation context
			corrCtx := &CorrelationContext{
				CorrelationID: correlationID,
				RequestID:     requestID,
				TraceID:       traceID,
				SpanID:        spanID,
			}
			
			// Add correlation headers to response
			w.Header().Set(CorrelationIDHeader, correlationID)
			w.Header().Set(RequestIDHeader, requestID)
			if traceID != "" {
				w.Header().Set(TraceIDHeader, traceID)
			}
			if spanID != "" {
				w.Header().Set(SpanIDHeader, spanID)
			}
			
			// Add correlation context to request context
			ctx := context.WithValue(r.Context(), "correlation_context", corrCtx)
			r = r.WithContext(ctx)
			
			next.ServeHTTP(w, r)
		})
	}
}

// GinCorrelationMiddleware creates middleware for Gin framework
func GinCorrelationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract or generate correlation ID
		correlationID := extractCorrelationID(c.Request)
		
		// Extract other correlation headers
		requestID := c.GetHeader(RequestIDHeader)
		traceID := c.GetHeader(TraceIDHeader)
		spanID := c.GetHeader(SpanIDHeader)
		
		// Generate request ID if not provided
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Create correlation context
		corrCtx := &CorrelationContext{
			CorrelationID: correlationID,
			RequestID:     requestID,
			TraceID:       traceID,
			SpanID:        spanID,
		}
		
		// Add correlation context to Gin context
		c.Set("correlation_context", corrCtx)
		
		// Add correlation context to request context
		ctx := context.WithValue(c.Request.Context(), "correlation_context", corrCtx)
		c.Request = c.Request.WithContext(ctx)
		
		// Add correlation headers to response
		c.Header(CorrelationIDHeader, correlationID)
		c.Header(RequestIDHeader, requestID)
		if traceID != "" {
			c.Header(TraceIDHeader, traceID)
		}
		if spanID != "" {
			c.Header(SpanIDHeader, spanID)
		}
		
		c.Next()
	}
}

// extractCorrelationID extracts correlation ID from request headers
func extractCorrelationID(r *http.Request) string {
	// Check for correlation ID header first
	if correlationID := r.Header.Get(CorrelationIDHeader); correlationID != "" {
		return correlationID
	}
	
	// Check for request ID header as fallback
	if requestID := r.Header.Get(RequestIDHeader); requestID != "" {
		return requestID
	}
	
	// Check for trace ID header as fallback
	if traceID := r.Header.Get(TraceIDHeader); traceID != "" {
		return traceID
	}
	
	// Generate new correlation ID if none found
	return uuid.New().String()
}

// GetCorrelationContext extracts correlation context from context
func GetCorrelationContext(ctx context.Context) *CorrelationContext {
	if corrCtx, ok := ctx.Value("correlation_context").(*CorrelationContext); ok {
		return corrCtx
	}
	return nil
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if corrCtx := GetCorrelationContext(ctx); corrCtx != nil {
		return corrCtx.CorrelationID
	}
	return ""
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if corrCtx := GetCorrelationContext(ctx); corrCtx != nil {
		return corrCtx.RequestID
	}
	return ""
}

// GetTraceID extracts trace ID from context
func GetTraceID(ctx context.Context) string {
	if corrCtx := GetCorrelationContext(ctx); corrCtx != nil {
		return corrCtx.TraceID
	}
	return ""
}

// GetSpanID extracts span ID from context
func GetSpanID(ctx context.Context) string {
	if corrCtx := GetCorrelationContext(ctx); corrCtx != nil {
		return corrCtx.SpanID
	}
	return ""
}

// SetUserID sets user ID in correlation context
func SetUserID(ctx context.Context, userID string) context.Context {
	if corrCtx := GetCorrelationContext(ctx); corrCtx != nil {
		corrCtx.UserID = userID
		return context.WithValue(ctx, "correlation_context", corrCtx)
	}
	return ctx
}

// SetSessionID sets session ID in correlation context
func SetSessionID(ctx context.Context, sessionID string) context.Context {
	if corrCtx := GetCorrelationContext(ctx); corrCtx != nil {
		corrCtx.SessionID = sessionID
		return context.WithValue(ctx, "correlation_context", corrCtx)
	}
	return ctx
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if corrCtx := GetCorrelationContext(ctx); corrCtx != nil {
		return corrCtx.UserID
	}
	return ""
}

// GetSessionID extracts session ID from context
func GetSessionID(ctx context.Context) string {
	if corrCtx := GetCorrelationContext(ctx); corrCtx != nil {
		return corrCtx.SessionID
	}
	return ""
}

// WithCorrelationContext creates a new context with correlation information
func WithCorrelationContext(ctx context.Context, correlationID, requestID, traceID, spanID string) context.Context {
	corrCtx := &CorrelationContext{
		CorrelationID: correlationID,
		RequestID:     requestID,
		TraceID:       traceID,
		SpanID:        spanID,
	}
	return context.WithValue(ctx, "correlation_context", corrCtx)
}

// PropagateCorrelationHeaders adds correlation headers to HTTP request
func PropagateCorrelationHeaders(req *http.Request, ctx context.Context) {
	if corrCtx := GetCorrelationContext(ctx); corrCtx != nil {
		if corrCtx.CorrelationID != "" {
			req.Header.Set(CorrelationIDHeader, corrCtx.CorrelationID)
		}
		if corrCtx.RequestID != "" {
			req.Header.Set(RequestIDHeader, corrCtx.RequestID)
		}
		if corrCtx.TraceID != "" {
			req.Header.Set(TraceIDHeader, corrCtx.TraceID)
		}
		if corrCtx.SpanID != "" {
			req.Header.Set(SpanIDHeader, corrCtx.SpanID)
		}
	}
}

// LogCorrelationContext returns a map of correlation fields for logging
func LogCorrelationContext(ctx context.Context) map[string]interface{} {
	if corrCtx := GetCorrelationContext(ctx); corrCtx != nil {
		fields := make(map[string]interface{})
		if corrCtx.CorrelationID != "" {
			fields["correlation_id"] = corrCtx.CorrelationID
		}
		if corrCtx.RequestID != "" {
			fields["request_id"] = corrCtx.RequestID
		}
		if corrCtx.TraceID != "" {
			fields["trace_id"] = corrCtx.TraceID
		}
		if corrCtx.SpanID != "" {
			fields["span_id"] = corrCtx.SpanID
		}
		if corrCtx.UserID != "" {
			fields["user_id"] = corrCtx.UserID
		}
		if corrCtx.SessionID != "" {
			fields["session_id"] = corrCtx.SessionID
		}
		return fields
	}
	return nil
}

// ValidateCorrelationID validates if a correlation ID is properly formatted
func ValidateCorrelationID(correlationID string) bool {
	if correlationID == "" {
		return false
	}
	
	// Check if it's a valid UUID
	if _, err := uuid.Parse(correlationID); err == nil {
		return true
	}
	
	// Check if it's a valid trace ID (32 hex characters)
	if len(correlationID) == 32 {
		for _, char := range correlationID {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
		return true
	}
	
	// Check if it's a valid span ID (16 hex characters)
	if len(correlationID) == 16 {
		for _, char := range correlationID {
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
				return false
			}
		}
		return true
	}
	
	return false
}

// SanitizeCorrelationID sanitizes a correlation ID for safe logging
func SanitizeCorrelationID(correlationID string) string {
	if correlationID == "" {
		return ""
	}
	
	// Truncate if too long
	if len(correlationID) > 64 {
		return correlationID[:64] + "..."
	}
	
	// Remove any potentially dangerous characters
	sanitized := strings.Map(func(r rune) rune {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '-' || r == '_' {
			return r
		}
		return -1
	}, correlationID)
	
	return sanitized
} 