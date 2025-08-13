package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"jarakey/shared/types"
)

// CryptoManager handles cryptographic operations
type CryptoManager struct {
	secretKey string
}

// NewCryptoManager creates a new crypto manager
func NewCryptoManager(secretKey string) *CryptoManager {
	return &CryptoManager{
		secretKey: secretKey,
	}
}

// GenerateSecureCode generates a secure 6-digit code
func (c *CryptoManager) GenerateSecureCode() (string, error) {
	// Generate a random number between 100000 and 999999
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	
	// Add 100000 to ensure it's 6 digits
	code := n.Add(n, big.NewInt(100000))
	return code.String(), nil
}

// GenerateSignature creates a HMAC signature for QR code data
func (c *CryptoManager) GenerateSignature(data string) string {
	h := hmac.New(sha256.New, []byte(c.secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifySignature verifies a HMAC signature
func (c *CryptoManager) VerifySignature(data, signature string) bool {
	expectedSignature := c.GenerateSignature(data)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// CreateQRCodeData creates data for QR code generation
func (c *CryptoManager) CreateQRCodeData(code *types.AccessCode, orgID string) (*types.QRCodeData, error) {
	// Create data string for signing
	dataString := fmt.Sprintf("%s:%s:%s:%d", code.Code, code.Purpose, orgID, code.ExpiresAt.Unix())
	
	// Generate signature
	signature := c.GenerateSignature(dataString)
	
	return &types.QRCodeData{
		Code:      code.Code,
		Signature: signature,
		ExpiresAt: code.ExpiresAt,
		Purpose:   code.Purpose,
		OrgID:     orgID,
	}, nil
}

// ValidateQRCodeData validates QR code data offline
func (c *CryptoManager) ValidateQRCodeData(qrData *types.QRCodeData) bool {
	// Check if code is expired
	if time.Now().After(qrData.ExpiresAt) {
		return false
	}
	
	// Create data string for verification
	dataString := fmt.Sprintf("%s:%s:%s:%d", qrData.Code, qrData.Purpose, qrData.OrgID, qrData.ExpiresAt.Unix())
	
	// Verify signature
	return c.VerifySignature(dataString, qrData.Signature)
}

// GenerateRandomString generates a random string of specified length
func (c *CryptoManager) GenerateRandomString(length int) (string, error) {
	// Validate input length
	if length <= 0 {
		return "", fmt.Errorf("length must be positive, got %d", length)
	}
	
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		randomIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[randomIndex.Int64()]
	}
	return string(b), nil
}

// HashPassword creates a secure hash of a password
func (c *CryptoManager) HashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password + c.secretKey))
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyPasswordHash verifies a password against its hash
func (c *CryptoManager) VerifyPasswordHash(password, hash string) bool {
	expectedHash := c.HashPassword(password)
	return hash == expectedHash
} 