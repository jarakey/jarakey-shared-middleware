package utils

import (
	"testing"
	"time"

	"github.com/jarakey/jarakey-shared-middleware/types"
	"github.com/stretchr/testify/assert"
)

func TestNewJWTManager(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	jwtManager := NewJWTManager(secretKey)
	
	assert.NotNil(t, jwtManager)
	assert.Equal(t, secretKey, jwtManager.secretKey)
}

func TestGenerateToken(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	jwtManager := NewJWTManager(secretKey)
	
	user := &types.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		Role:      types.RoleMember,
		OrgID:     "org-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	token, err := jwtManager.GenerateToken(user)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	
	// Token should be a valid JWT format (3 parts separated by dots)
	// This is a basic check - in production you'd want more thorough validation
	assert.Contains(t, token, ".")
}

func TestGenerateTokenWithDifferentUsers(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	jwtManager := NewJWTManager(secretKey)
	
	users := []*types.User{
		{
			ID:        "user-1",
			Email:     "user1@example.com",
			Name:      "User One",
			Role:      types.RoleMember,
			OrgID:     "org-1",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "user-2",
			Email:     "user2@example.com",
			Name:      "User Two",
			Role:      types.RoleAdmin,
			OrgID:     "org-2",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	tokens := make(map[string]bool)
	
	for _, user := range users {
		token, err := jwtManager.GenerateToken(user)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
		
		// Each token should be unique
		assert.False(t, tokens[token], "Token should be unique for user: %s", user.ID)
		tokens[token] = true
	}
}

func TestValidateToken(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	jwtManager := NewJWTManager(secretKey)
	
	user := &types.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		Role:      types.RoleMember,
		OrgID:     "org-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	token, err := jwtManager.GenerateToken(user)
	assert.NoError(t, err)
	
	// Valid token should validate
	claims, err := jwtManager.ValidateToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
	assert.Equal(t, user.OrgID, claims.OrgID)
	
	// Expiration should be in the future
	assert.True(t, claims.Exp > time.Now().Unix())
	
	// Issued at should be in the past
	assert.True(t, claims.Iat <= time.Now().Unix())
}

func TestValidateTokenInvalid(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	jwtManager := NewJWTManager(secretKey)
	
	// Invalid token format
	_, err := jwtManager.ValidateToken("invalid-token")
	assert.Error(t, err)
	
	// Empty token
	_, err = jwtManager.ValidateToken("")
	assert.Error(t, err)
	
	// Token with wrong signature
	user := &types.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		Role:      types.RoleMember,
		OrgID:     "org-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	token, err := jwtManager.GenerateToken(user)
	assert.NoError(t, err)
	
	// Create a different JWT manager with different secret
	differentJWTManager := NewJWTManager("different-secret-key-32-chars")
	_, err = differentJWTManager.ValidateToken(token)
	assert.Error(t, err)
}

func TestRefreshToken(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	jwtManager := NewJWTManager(secretKey)
	
	user := &types.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		Role:      types.RoleMember,
		OrgID:     "org-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	originalToken, err := jwtManager.GenerateToken(user)
	assert.NoError(t, err)
	
	// Refresh the token
	refreshedToken, err := jwtManager.RefreshToken(originalToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, refreshedToken)
	
	// Refreshed token should have different expiration time
	// Note: Tokens might be identical if generated at the same second
	// but we ensure the expiration time is extended
	
	// Both tokens should be valid
	originalClaims, err := jwtManager.ValidateToken(originalToken)
	assert.NoError(t, err)
	assert.NotNil(t, originalClaims)
	
	refreshedClaims, err := jwtManager.ValidateToken(refreshedToken)
	assert.NoError(t, err)
	assert.NotNil(t, refreshedClaims)
	
	// Claims should be the same (except for timing)
	assert.Equal(t, originalClaims.UserID, refreshedClaims.UserID)
	assert.Equal(t, originalClaims.Email, refreshedClaims.Email)
	assert.Equal(t, originalClaims.Role, refreshedClaims.Role)
	assert.Equal(t, originalClaims.OrgID, refreshedClaims.OrgID)
	
	// Refreshed token should have later expiration
	assert.True(t, refreshedClaims.Exp > originalClaims.Exp)
}

func TestRefreshTokenInvalid(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	jwtManager := NewJWTManager(secretKey)
	
	// Invalid token should not refresh
	_, err := jwtManager.RefreshToken("invalid-token")
	assert.Error(t, err)
	
	// Empty token should not refresh
	_, err = jwtManager.RefreshToken("")
	assert.Error(t, err)
}

func TestTokenExpiration(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	jwtManager := NewJWTManager(secretKey)
	
	user := &types.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		Role:      types.RoleMember,
		OrgID:     "org-456",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	token, err := jwtManager.GenerateToken(user)
	assert.NoError(t, err)
	
	claims, err := jwtManager.ValidateToken(token)
	assert.NoError(t, err)
	
	// Token should not be expired
	assert.True(t, claims.Exp > time.Now().Unix())
	
	// Token should have been issued recently
	assert.True(t, claims.Iat <= time.Now().Unix())
	assert.True(t, claims.Iat > time.Now().Add(-1*time.Minute).Unix())
} 