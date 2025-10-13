package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jarakey/jarakey-shared-middleware/types"
	"github.com/jarakey/jarakey-shared-middleware/utils"
	"github.com/stretchr/testify/assert"
)

func TestAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "No Authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authorization header required",
		},
		{
			name:           "Invalid Authorization format",
			authHeader:     "Invalid token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid authorization format",
		},
		{
			name:           "Empty Bearer token",
			authHeader:     "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Token required",
		},
		{
			name:           "Invalid JWT token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AuthRequired())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}
		})
	}
}

func TestAuthRequiredValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a valid JWT token using the same secret as the middleware
	user := &types.User{
		ID:    "test-user-id",
		Email: "test@example.com",
		Role:  types.RoleAdmin,
		OrgID: "test-org-id",
	}

	// Use the same secret that the middleware uses by default
	secretKey := "super-secret-jwt-key-to-change-in-production"
	jwtManager := utils.NewJWTManager(secretKey)
	token, err := jwtManager.GenerateToken(user)
	assert.NoError(t, err)

	router := gin.New()
	router.Use(AuthRequired())
	router.GET("/test", func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, "test-user-id", userID)

		email, exists := c.Get("user_email")
		assert.True(t, exists)
		assert.Equal(t, "test@example.com", email)

		role, exists := c.Get("user_role")
		assert.True(t, exists)
		assert.Equal(t, types.RoleAdmin, role)

		orgID, exists := c.Get("org_id")
		assert.True(t, exists)
		assert.Equal(t, "test-org-id", orgID)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireAdminPermission(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRole       types.UserRole
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Admin role",
			userRole:       types.RoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Member role",
			userRole:       types.RoleMember,
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions",
		},
		{
			name:           "Guard role",
			userRole:       types.RoleGuard,
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Mock authentication
				c.Set("user_role", tt.userRole)
				c.Next()
			})
			router.Use(RequireAdminPermission())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}
		})
	}
}

func TestRequireRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRole       types.UserRole
		requiredRole   types.UserRole
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Correct role",
			userRole:       types.RoleAdmin,
			requiredRole:   types.RoleAdmin,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Incorrect role",
			userRole:       types.RoleMember,
			requiredRole:   types.RoleAdmin,
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Mock authentication
				c.Set("user_role", tt.userRole)
				c.Next()
			})
			router.Use(RequireRole(tt.requiredRole))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}
		})
	}
}

func TestRequireAnyRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		userRole       types.UserRole
		requiredRoles  []types.UserRole
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Has one of the required roles",
			userRole:       types.RoleAdmin,
			requiredRoles:  []types.UserRole{types.RoleAdmin, types.RoleMember},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Does not have any required role",
			userRole:       types.RoleGuard,
			requiredRoles:  []types.UserRole{types.RoleAdmin, types.RoleMember},
			expectedStatus: http.StatusForbidden,
			expectedError:  "Insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(func(c *gin.Context) {
				// Mock authentication
				c.Set("user_role", tt.userRole)
				c.Next()
			})
			router.Use(RequireAnyRole(tt.requiredRoles...))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, response["error"])
			}
		})
	}
}

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(SecurityHeaders())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	router.OPTIONS("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "options"})
	})

	// Test GET request
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, Authorization, X-Requested-With, X-Correlation-ID", w.Header().Get("Access-Control-Allow-Headers"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

	// Test OPTIONS request
	req, _ = http.NewRequest("OPTIONS", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}
