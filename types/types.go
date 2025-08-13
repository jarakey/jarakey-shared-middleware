package types

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	Name         string    `json:"name" db:"name"`
	Avatar       string    `json:"avatar" db:"avatar"`
	Role         UserRole  `json:"role" db:"role"`
	OrgID        string    `json:"org_id" db:"org_id"`
	Provider     string    `json:"provider" db:"provider"` // google, facebook, apple
	ProviderID   string    `json:"provider_id" db:"provider_id"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole represents user roles
type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleMember UserRole = "member"
	RoleGuard  UserRole = "guard"
)

// Organization represents an organization
type Organization struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Domain      string    `json:"domain" db:"domain"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// AccessCode represents a generated access code
type AccessCode struct {
	ID          string    `json:"id" db:"id"`
	Code        string    `json:"code" db:"code"`
	UserID      string    `json:"user_id" db:"user_id"`
	OrgID       string    `json:"org_id" db:"org_id"`
	Purpose     string    `json:"purpose" db:"purpose"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
	IsUsed      bool      `json:"is_used" db:"is_used"`
	UsedAt      *time.Time `json:"used_at" db:"used_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ValidationLog represents a code validation attempt
type ValidationLog struct {
	ID            string    `json:"id" db:"id"`
	CodeID        string    `json:"code_id" db:"code_id"`
	ValidatorID   string    `json:"validator_id" db:"validator_id"`
	Status        string    `json:"status" db:"status"` // valid, invalid, expired
	IPAddress     string    `json:"ip_address" db:"ip_address"`
	UserAgent     string    `json:"user_agent" db:"user_agent"`
	Location      string    `json:"location" db:"location"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// Validator represents a guard/validator account
type Validator struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	OrgID     string    `json:"org_id" db:"org_id"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CodeGenerationRequest represents a request to generate a code
type CodeGenerationRequest struct {
	Purpose   string `json:"purpose" validate:"required"`
	Duration  string `json:"duration" validate:"required,oneof=10min 30min 1hour unlimited"`
}

// CodeValidationRequest represents a request to validate a code
type CodeValidationRequest struct {
	Code      string `json:"code" validate:"required,len=6"`
	Validator string `json:"validator,omitempty"`
}

// CodeValidationResponse represents the response from code validation
type CodeValidationResponse struct {
	Valid     bool      `json:"valid"`
	Code      string    `json:"code,omitempty"`
	Purpose   string    `json:"purpose,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Message   string    `json:"message"`
}

// QRCodeData represents the data encoded in QR codes for offline validation
type QRCodeData struct {
	Code      string    `json:"code"`
	Signature string    `json:"signature"`
	ExpiresAt time.Time `json:"expires_at"`
	Purpose   string    `json:"purpose"`
	OrgID     string    `json:"org_id"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Pagination represents pagination parameters
type Pagination struct {
	Page     int `json:"page" query:"page"`
	PageSize int `json:"page_size" query:"page_size"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// OAuthProvider represents OAuth provider configuration
type OAuthProvider struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURL  string `json:"redirect_url"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Role   UserRole `json:"role"`
	OrgID  string   `json:"org_id"`
	Exp    int64    `json:"exp"`
	Iat    int64    `json:"iat"`
}

// Duration constants
const (
	Duration10Min    = "10min"
	Duration30Min    = "30min"
	Duration1Hour    = "1hour"
	DurationUnlimited = "unlimited"
)

// Status constants
const (
	StatusValid   = "valid"
	StatusInvalid = "invalid"
	StatusExpired = "expired"
) 