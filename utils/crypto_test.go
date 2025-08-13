package utils

import (
	"testing"
	"time"

	"github.com/jarakey/jarakey-shared-middleware/types"
	"github.com/stretchr/testify/assert"
)

func TestNewCryptoManager(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	crypto := NewCryptoManager(secretKey)
	
	assert.NotNil(t, crypto)
	assert.Equal(t, secretKey, crypto.secretKey)
}

func TestGenerateSecureCode(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	crypto := NewCryptoManager(secretKey)
	
	code, err := crypto.GenerateSecureCode()
	
	assert.NoError(t, err)
	assert.NotEmpty(t, code)
	assert.Len(t, code, 6) // Should be 6 characters
}

func TestGenerateSecureCodeMultiple(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	crypto := NewCryptoManager(secretKey)
	
	codes := make(map[string]bool)
	
	// Generate multiple codes to ensure uniqueness
	for i := 0; i < 100; i++ {
		code, err := crypto.GenerateSecureCode()
		assert.NoError(t, err)
		assert.Len(t, code, 6)
		
		// Check for uniqueness
		assert.False(t, codes[code], "Code should be unique: %s", code)
		codes[code] = true
	}
}

func TestGenerateSignature(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	crypto := NewCryptoManager(secretKey)
	
	data := "test-data-to-sign"
	signature := crypto.GenerateSignature(data)
	
	assert.NotEmpty(t, signature)
	assert.NotEqual(t, data, signature) // Signature should be different from data
}

func TestVerifySignature(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	crypto := NewCryptoManager(secretKey)
	
	data := "test-data-to-verify"
	signature := crypto.GenerateSignature(data)
	
	// Valid signature should verify
	assert.True(t, crypto.VerifySignature(data, signature))
	
	// Invalid signature should not verify
	assert.False(t, crypto.VerifySignature(data, "invalid-signature"))
	
	// Different data should not verify with same signature
	assert.False(t, crypto.VerifySignature("different-data", signature))
}

func TestCreateQRCodeData(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	crypto := NewCryptoManager(secretKey)
	
	code := &types.AccessCode{
		ID:        "code-123",
		Code:      "ABC123",
		UserID:    "user-123",
		OrgID:     "org-456",
		Purpose:   "entry",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		IsUsed:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	orgID := "org-456"
	qrData, err := crypto.CreateQRCodeData(code, orgID)
	
	assert.NoError(t, err)
	assert.NotNil(t, qrData)
	assert.Equal(t, code.Code, qrData.Code)
	assert.Equal(t, code.Purpose, qrData.Purpose)
	assert.Equal(t, orgID, qrData.OrgID)
	assert.NotEmpty(t, qrData.Signature)
	assert.Equal(t, code.ExpiresAt.Unix(), qrData.ExpiresAt.Unix())
}

func TestValidateQRCodeData(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	crypto := NewCryptoManager(secretKey)
	
	code := &types.AccessCode{
		ID:        "code-123",
		Code:      "ABC123",
		UserID:    "user-123",
		OrgID:     "org-456",
		Purpose:   "entry",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		IsUsed:    false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	orgID := "org-456"
	qrData, err := crypto.CreateQRCodeData(code, orgID)
	assert.NoError(t, err)
	
	// Valid QR data should validate
	assert.True(t, crypto.ValidateQRCodeData(qrData))
	
	// Invalid signature should not validate
	qrData.Signature = "invalid-signature"
	assert.False(t, crypto.ValidateQRCodeData(qrData))
	
	// Expired QR data should not validate
	qrData.ExpiresAt = time.Now().Add(-1 * time.Hour)
	assert.False(t, crypto.ValidateQRCodeData(qrData))
}

func TestGenerateRandomString(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	crypto := NewCryptoManager(secretKey)
	
	// Test different lengths
	lengths := []int{8, 16, 32, 64}
	
	for _, length := range lengths {
		randomStr, err := crypto.GenerateRandomString(length)
		assert.NoError(t, err)
		assert.Len(t, randomStr, length)
		
		// Generate another string of same length to ensure randomness
		randomStr2, err := crypto.GenerateRandomString(length)
		assert.NoError(t, err)
		assert.Len(t, randomStr2, length)
		assert.NotEqual(t, randomStr, randomStr2, "Random strings should be different")
	}
}

func TestGenerateRandomStringInvalidLength(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	crypto := NewCryptoManager(secretKey)
	
	// Test with invalid lengths
	invalidLengths := []int{0, -1, -10}
	
	for _, length := range invalidLengths {
		_, err := crypto.GenerateRandomString(length)
		assert.Error(t, err)
	}
}

func TestHashPassword(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	crypto := NewCryptoManager(secretKey)
	
	password := "my-secure-password"
	hash := crypto.HashPassword(password)
	
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash) // Hash should be different from password
	assert.Len(t, hash, 64) // SHA-256 hash is always 64 characters
}

func TestVerifyPasswordHash(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	crypto := NewCryptoManager(secretKey)
	
	password := "my-secure-password"
	hash := crypto.HashPassword(password)
	
	// Correct password should verify
	assert.True(t, crypto.VerifyPasswordHash(password, hash))
	
	// Wrong password should not verify
	assert.False(t, crypto.VerifyPasswordHash("wrong-password", hash))
	
	// Empty password should not verify
	assert.False(t, crypto.VerifyPasswordHash("", hash))
	
	// Empty hash should not verify
	assert.False(t, crypto.VerifyPasswordHash(password, ""))
}

func TestHashPasswordConsistency(t *testing.T) {
	secretKey := "test-secret-key-32-chars-long"
	crypto := NewCryptoManager(secretKey)
	
	password := "same-password"
	hash1 := crypto.HashPassword(password)
	hash2 := crypto.HashPassword(password)
	
	// Same password should produce the same hash (SHA-256 is deterministic)
	assert.Equal(t, hash1, hash2)
	
	// Both should verify correctly
	assert.True(t, crypto.VerifyPasswordHash(password, hash1))
	assert.True(t, crypto.VerifyPasswordHash(password, hash2))
} 