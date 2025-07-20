package entities

import (
	"time"
)

// UserStatus represents the status of a user account
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusDeleted   UserStatus = "deleted"
)

// User represents a user in the system
type User struct {
	ID           string     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"` // Never expose password hash in JSON
	FirstName    string     `json:"first_name" db:"first_name"`
	LastName     string     `json:"last_name" db:"last_name"`
	Status       UserStatus `json:"status" db:"status"`
	RoleID       string     `json:"role_id" db:"role_id"`
	Role         *Role      `json:"role,omitempty" db:"-"` // Loaded separately
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at" db:"last_login_at"`
}

// IsActive returns true if the user is active
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// CanLogin returns true if the user can log in
func (u *User) CanLogin() bool {
	return u.Status == UserStatusActive
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	if u.FirstName == "" && u.LastName == "" {
		return u.Username
	}
	return u.FirstName + " " + u.LastName
}

// HasPermission checks if the user has a specific permission
func (u *User) HasPermission(permission string) bool {
	if u.Role == nil {
		return false
	}
	return u.Role.HasPermission(permission)
}

// HasRole checks if the user has a specific role
func (u *User) HasRole(roleName string) bool {
	if u.Role == nil {
		return false
	}
	return u.Role.Name == roleName
}

// UserSession represents an active user session
type UserSession struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	TokenHash    string    `json:"-" db:"token_hash"` // Never expose token hash
	RefreshToken string    `json:"-" db:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	LastUsedAt   time.Time `json:"last_used_at" db:"last_used_at"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	IsRevoked    bool      `json:"is_revoked" db:"is_revoked"`
}

// IsValid returns true if the session is valid and not expired
func (s *UserSession) IsValid() bool {
	return !s.IsRevoked && time.Now().Before(s.ExpiresAt)
}

// IsExpired returns true if the session has expired
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// APIKey represents an API key for programmatic access
type APIKey struct {
	ID          string     `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	Name        string     `json:"name" db:"name"`
	KeyHash     string     `json:"-" db:"key_hash"`              // Never expose key hash
	Prefix      string     `json:"prefix" db:"prefix"`           // First 8 chars for identification
	Permissions []string   `json:"permissions" db:"permissions"` // JSON array in DB
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at" db:"last_used_at"`
	IsActive    bool       `json:"is_active" db:"is_active"`
}

// IsValid returns true if the API key is valid and not expired
func (k *APIKey) IsValid() bool {
	if !k.IsActive {
		return false
	}
	if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
		return false
	}
	return true
}

// HasPermission checks if the API key has a specific permission
func (k *APIKey) HasPermission(permission string) bool {
	for _, p := range k.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"max=100"`
	LastName  string `json:"last_name" validate:"max=100"`
	RoleID    string `json:"role_id" validate:"required"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Email        *string     `json:"email,omitempty" validate:"omitempty,email"`
	FirstName    *string     `json:"first_name,omitempty" validate:"omitempty,max=100"`
	LastName     *string     `json:"last_name,omitempty" validate:"omitempty,max=100"`
	Status       *UserStatus `json:"status,omitempty"`
	RoleID       *string     `json:"role_id,omitempty"`
	PasswordHash *string     `json:"-"` // Never expose password hash in JSON
	LastLoginAt  *time.Time  `json:"last_login_at,omitempty"`
}

// ChangePasswordRequest represents a request to change user password
type ChangePasswordRequest struct {
	CurrentPassword    string `json:"current_password" validate:"required"`
	NewPassword        string `json:"new_password" validate:"required,min=8"`
	KeepCurrentSession bool   `json:"keep_current_session,omitempty"`
	CurrentSessionID   string `json:"current_session_id,omitempty"`
}

// CreateAPIKeyRequest represents a request to create a new API key
type CreateAPIKeyRequest struct {
	Name        string     `json:"name" validate:"required,min=1,max=100"`
	Permissions []string   `json:"permissions" validate:"required,min=1"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username  string `json:"username" validate:"required"`
	Password  string `json:"password" validate:"required"`
	IPAddress string `json:"ip_address,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *User     `json:"user"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UserFilter represents filters for querying users
type UserFilter struct {
	Username *string     `json:"username,omitempty"`
	Email    *string     `json:"email,omitempty"`
	Status   *UserStatus `json:"status,omitempty"`
	RoleID   *string     `json:"role_id,omitempty"`
	Limit    int         `json:"limit,omitempty"`
	Offset   int         `json:"offset,omitempty"`
}

// APIKeyFilter represents filters for querying API keys
type APIKeyFilter struct {
	UserID   *string `json:"user_id,omitempty"`
	Name     *string `json:"name,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
	Limit    int     `json:"limit,omitempty"`
	Offset   int     `json:"offset,omitempty"`
}

// SessionFilter represents filters for querying sessions
type SessionFilter struct {
	UserID    *string `json:"user_id,omitempty"`
	IsRevoked *bool   `json:"is_revoked,omitempty"`
	Limit     int     `json:"limit,omitempty"`
	Offset    int     `json:"offset,omitempty"`
}
