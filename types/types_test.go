package types

import (
	"testing"
	"time"
)

func TestUserRoleConstants(t *testing.T) {
	// Test that user role constants are defined correctly
	if RoleAdmin != "admin" {
		t.Errorf("Expected RoleAdmin to be 'admin', got %s", RoleAdmin)
	}
	
	if RoleMember != "member" {
		t.Errorf("Expected RoleMember to be 'member', got %s", RoleMember)
	}
	
	if RoleGuard != "guard" {
		t.Errorf("Expected RoleGuard to be 'guard', got %s", RoleGuard)
	}
}

func TestUserStruct(t *testing.T) {
	// Test User struct creation
	user := User{
		ID:        "test-id",
		Email:     "test@example.com",
		Name:      "Test User",
		Role:      RoleMember,
		OrgID:     "org-123",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if user.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", user.ID)
	}
	
	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", user.Email)
	}
	
	if user.Role != RoleMember {
		t.Errorf("Expected role %s, got %s", RoleMember, user.Role)
	}
}

func TestOrganizationStruct(t *testing.T) {
	// Test Organization struct creation
	org := Organization{
		ID:          "org-123",
		Name:        "Test Org",
		Description: "Test Description",
		Domain:      "test.com",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	if org.ID != "org-123" {
		t.Errorf("Expected ID 'org-123', got %s", org.ID)
	}
	
	if org.Name != "Test Org" {
		t.Errorf("Expected name 'Test Org', got %s", org.Name)
	}
	
	if org.Domain != "test.com" {
		t.Errorf("Expected domain 'test.com', got %s", org.Domain)
	}
}

func TestAccessCodeStruct(t *testing.T) {
	// Test AccessCode struct creation
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)
	
	code := AccessCode{
		ID:        "code-123",
		Code:      "123456",
		UserID:    "user-123",
		OrgID:     "org-123",
		Purpose:   "test",
		ExpiresAt: expiresAt,
		IsUsed:    false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	if code.ID != "code-123" {
		t.Errorf("Expected ID 'code-123', got %s", code.ID)
	}
	
	if code.Code != "123456" {
		t.Errorf("Expected code '123456', got %s", code.Code)
	}
	
	if code.Purpose != "test" {
		t.Errorf("Expected purpose 'test', got %s", code.Purpose)
	}
	
	if code.IsUsed {
		t.Error("Expected IsUsed to be false")
	}
}

func TestValidatorStruct(t *testing.T) {
	// Test Validator struct creation
	validator := Validator{
		ID:        "validator-123",
		UserID:    "user-123",
		OrgID:     "org-123",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if validator.ID != "validator-123" {
		t.Errorf("Expected ID 'validator-123', got %s", validator.ID)
	}
	
	if validator.UserID != "user-123" {
		t.Errorf("Expected UserID 'user-123', got %s", validator.UserID)
	}
	
	if !validator.IsActive {
		t.Error("Expected IsActive to be true")
	}
}

func TestCodeGenerationRequestStruct(t *testing.T) {
	// Test CodeGenerationRequest struct creation
	req := CodeGenerationRequest{
		Purpose:  "test",
		Duration: "1hour",
	}
	
	if req.Purpose != "test" {
		t.Errorf("Expected purpose 'test', got %s", req.Purpose)
	}
	
	if req.Duration != "1hour" {
		t.Errorf("Expected duration '1hour', got %s", req.Duration)
	}
}

func TestCodeValidationRequestStruct(t *testing.T) {
	// Test CodeValidationRequest struct creation
	req := CodeValidationRequest{
		Code:      "123456",
		Validator: "validator-123",
	}
	
	if req.Code != "123456" {
		t.Errorf("Expected code '123456', got %s", req.Code)
	}
	
	if req.Validator != "validator-123" {
		t.Errorf("Expected validator 'validator-123', got %s", req.Validator)
	}
}

func TestCodeValidationResponseStruct(t *testing.T) {
	// Test CodeValidationResponse struct creation
	now := time.Now()
	expiresAt := now.Add(1 * time.Hour)
	
	resp := CodeValidationResponse{
		Valid:     true,
		Code:      "123456",
		Purpose:   "test",
		ExpiresAt: expiresAt,
		Message:   "Code is valid",
	}
	
	if !resp.Valid {
		t.Error("Expected Valid to be true")
	}
	
	if resp.Code != "123456" {
		t.Errorf("Expected code '123456', got %s", resp.Code)
	}
	
	if resp.Message != "Code is valid" {
		t.Errorf("Expected message 'Code is valid', got %s", resp.Message)
	}
}
