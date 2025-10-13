package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jarakey/jarakey-shared-middleware/types"
	"github.com/jarakey/jarakey-shared-middleware/utils"
)

// AuthRequired middleware checks if user is authenticated
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization header required",
				"message": "Please provide a valid authorization token",
			})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization format",
				"message": "Authorization header must be in format 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Token required",
				"message": "Please provide a valid token",
			})
			c.Abort()
			return
		}

		// Get JWT secret from environment
		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			secretKey = "super-secret-jwt-key-to-change-in-production"
		}

		// Validate token using shared JWT manager
		jwtManager := utils.NewJWTManager(secretKey)
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			log.Printf("JWT validation error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("org_id", claims.OrgID)
		c.Set("user_permissions", claims.Permissions)

		c.Next()
	}
}

// RequireAdminPermission middleware checks if user has admin permissions
func RequireAdminPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "User not authenticated",
				"message": "Please log in to access this resource",
			})
			c.Abort()
			return
		}

		// Check if user has admin role
		role, ok := userRole.(types.UserRole)
		if !ok || !isAdminRole(string(role)) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Insufficient permissions",
				"message": "Admin access required for this operation",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole middleware checks if user has specific role
func RequireRole(requiredRole types.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "User not authenticated",
				"message": "Please log in to access this resource",
			})
			c.Abort()
			return
		}

		// Check if user has required role
		role, ok := userRole.(types.UserRole)
		if !ok || role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Insufficient permissions",
				"message": "Required role: " + string(requiredRole),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole middleware checks if user has any of the specified roles
func RequireAnyRole(requiredRoles ...types.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if user is authenticated
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "User not authenticated",
				"message": "Please log in to access this resource",
			})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		role, ok := userRole.(types.UserRole)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Invalid user role",
				"message": "User role is not valid",
			})
			c.Abort()
			return
		}

		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			roleStrings := make([]string, len(requiredRoles))
			for i, r := range requiredRoles {
				roleStrings[i] = string(r)
			}
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Insufficient permissions",
				"message": "Required roles: " + strings.Join(roleStrings, ", "),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// isAdminRole checks if the role is an admin role
func isAdminRole(role string) bool {
	adminRoles := []string{"admin", "super_admin", "jarakey-manager"}
	for _, adminRole := range adminRoles {
		if role == adminRole {
			return true
		}
	}
	return false
}

// SecurityHeaders middleware adds security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}

// CORS middleware handles CORS
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Correlation-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
