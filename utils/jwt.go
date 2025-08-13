package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"jarakey/shared/types"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
	}
}

// GenerateToken generates a new JWT token for a user
func (j *JWTManager) GenerateToken(user *types.User) (string, error) {
	claims := types.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		OrgID:  user.OrgID,
		Exp:    time.Now().Add(24 * time.Hour).Unix(),
		Iat:    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*types.JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &types.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*types.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken generates a new token with extended expiration
func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Create new claims with extended expiration
	newClaims := types.JWTClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
		OrgID:  claims.OrgID,
		Exp:    time.Now().Add(24 * time.Hour).Unix(),
		Iat:    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return token.SignedString([]byte(j.secretKey))
} 