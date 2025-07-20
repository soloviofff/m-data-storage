package interfaces

import (
	"context"
	"time"

	"m-data-storage/internal/domain/entities"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// User registration
	Register(ctx context.Context, username, email, password string) (*entities.User, error)

	// Authentication
	Login(ctx context.Context, req *entities.LoginRequest) (*entities.LoginResponse, error)
	Logout(ctx context.Context, sessionID string) error
	RefreshToken(ctx context.Context, req *entities.RefreshTokenRequest) (*entities.LoginResponse, error)
	ValidateSession(ctx context.Context, sessionID string) (*entities.User, error)

	// User management
	GetUserByID(ctx context.Context, userID string) (*entities.User, error)

	// Password management
	ChangePassword(ctx context.Context, userID string, req *entities.ChangePasswordRequest) error
	ResetPassword(ctx context.Context, email string) error
	ValidateResetToken(ctx context.Context, token string) (*entities.User, error)
	SetNewPassword(ctx context.Context, token, newPassword string) error

	// Session management
	GetUserSessions(ctx context.Context, userID string) ([]*entities.UserSession, error)
	RevokeSession(ctx context.Context, sessionID string) error
	RevokeAllUserSessions(ctx context.Context, userID string) error
}

// UserService defines the interface for user management operations
type UserService interface {
	// User CRUD operations
	CreateUser(ctx context.Context, req *entities.CreateUserRequest) (*entities.User, error)
	GetUserByID(ctx context.Context, id string) (*entities.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUsers(ctx context.Context, filter entities.UserFilter) ([]*entities.User, error)
	UpdateUser(ctx context.Context, id string, req *entities.UpdateUserRequest) (*entities.User, error)
	DeleteUser(ctx context.Context, id string) error

	// User status management
	ActivateUser(ctx context.Context, id string) error
	DeactivateUser(ctx context.Context, id string) error
	SuspendUser(ctx context.Context, id string) error

	// User validation
	ValidateUserCredentials(ctx context.Context, username, password string) (*entities.User, error)
	CheckUserExists(ctx context.Context, username, email string) (bool, error)
}

// RoleService defines the interface for role management operations
type RoleService interface {
	// Role CRUD operations
	CreateRole(ctx context.Context, req *entities.CreateRoleRequest) (*entities.Role, error)
	GetRoleByID(ctx context.Context, id string) (*entities.Role, error)
	GetRoleByName(ctx context.Context, name string) (*entities.Role, error)
	GetRoles(ctx context.Context, filter entities.RoleFilter) ([]*entities.Role, error)
	UpdateRole(ctx context.Context, id string, req *entities.UpdateRoleRequest) (*entities.Role, error)
	DeleteRole(ctx context.Context, id string) error

	// Role-Permission management
	AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]*entities.Permission, error)

	// Role validation
	ValidateRolePermissions(ctx context.Context, roleID string, requiredPermissions []string) (bool, error)
}

// PermissionService defines the interface for permission management operations
type PermissionService interface {
	// Permission CRUD operations
	CreatePermission(ctx context.Context, req *entities.CreatePermissionRequest) (*entities.Permission, error)
	GetPermissionByID(ctx context.Context, id string) (*entities.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*entities.Permission, error)
	GetPermissions(ctx context.Context, filter entities.PermissionFilter) ([]*entities.Permission, error)
	UpdatePermission(ctx context.Context, id string, req *entities.UpdatePermissionRequest) (*entities.Permission, error)
	DeletePermission(ctx context.Context, id string) error

	// Permission validation
	CheckPermission(ctx context.Context, userID, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
}

// APIKeyService defines the interface for API key management operations
type APIKeyService interface {
	// API Key CRUD operations
	CreateAPIKey(ctx context.Context, userID string, req *entities.CreateAPIKeyRequest) (*entities.APIKey, string, error) // Returns APIKey, raw key, error
	GetAPIKeyByID(ctx context.Context, id string) (*entities.APIKey, error)
	GetAPIKeysByUserID(ctx context.Context, userID string) ([]*entities.APIKey, error)
	GetAPIKeys(ctx context.Context, filter entities.APIKeyFilter) ([]*entities.APIKey, error)
	UpdateAPIKey(ctx context.Context, id string, req *entities.CreateAPIKeyRequest) (*entities.APIKey, error)
	DeleteAPIKey(ctx context.Context, id string) error

	// API Key validation
	ValidateAPIKey(ctx context.Context, key string) (*entities.APIKey, *entities.User, error)
	RevokeAPIKey(ctx context.Context, id string) error
	UpdateAPIKeyLastUsed(ctx context.Context, id string) error

	// API Key utilities
	GenerateAPIKey() (string, string, error) // Returns key, hash, error
	HashAPIKey(key string) (string, error)
	ValidateAPIKeyFormat(key string) bool
}

// TokenService defines the interface for JWT token operations
type TokenService interface {
	// JWT operations
	GenerateAccessToken(ctx context.Context, user *entities.User) (string, time.Time, error)
	GenerateRefreshToken(ctx context.Context, user *entities.User) (string, time.Time, error)
	ValidateAccessToken(ctx context.Context, token string) (*entities.User, error)
	ValidateRefreshToken(ctx context.Context, token string) (*entities.User, error)
	RevokeToken(ctx context.Context, token string) error

	// Token utilities
	ExtractUserFromToken(ctx context.Context, token string) (*entities.User, error)
	GetTokenExpiration(token string) (time.Time, error)
	IsTokenExpired(token string) bool
}

// PasswordService defines the interface for password operations
type PasswordService interface {
	// Password hashing
	HashPassword(password string) (string, error)
	ValidatePassword(password, hash string) bool

	// Password validation
	ValidatePasswordStrength(password string) error
	GenerateRandomPassword(length int) string

	// Password reset
	GenerateResetToken(userID string) (string, time.Time, error)
	ValidateResetToken(token string) (string, error) // Returns userID
}

// AuthorizationService defines the interface for authorization operations
type AuthorizationService interface {
	// Permission checking
	HasPermission(ctx context.Context, userID, permission string) (bool, error)
	HasAnyPermission(ctx context.Context, userID string, permissions []string) (bool, error)
	HasAllPermissions(ctx context.Context, userID string, permissions []string) (bool, error)

	// Role checking
	HasRole(ctx context.Context, userID, roleName string) (bool, error)
	HasAnyRole(ctx context.Context, userID string, roleNames []string) (bool, error)

	// Resource-based authorization
	CanAccessResource(ctx context.Context, userID, resource, action string) (bool, error)
	CanModifyUser(ctx context.Context, actorID, targetUserID string) (bool, error)
	CanManageRole(ctx context.Context, userID, roleID string) (bool, error)

	// API Key authorization
	CanAPIKeyAccess(ctx context.Context, apiKey *entities.APIKey, resource, action string) (bool, error)
}

// SecurityService defines the interface for security operations
type SecurityService interface {
	// Rate limiting
	CheckRateLimit(ctx context.Context, identifier string, action string) (bool, error)
	IncrementRateLimit(ctx context.Context, identifier string, action string) error

	// Security logging
	LogSecurityEvent(ctx context.Context, event SecurityEvent) error
	GetSecurityEvents(ctx context.Context, filter SecurityEventFilter) ([]*SecurityEvent, error)

	// Account security
	CheckAccountLockout(ctx context.Context, userID string) (bool, time.Time, error)
	LockAccount(ctx context.Context, userID string, duration time.Duration) error
	UnlockAccount(ctx context.Context, userID string) error

	// Session security
	ValidateSessionSecurity(ctx context.Context, session *entities.UserSession, ipAddress, userAgent string) (bool, error)
	DetectSuspiciousActivity(ctx context.Context, userID string) (bool, []string, error)
}

// SecurityEvent represents a security-related event
type SecurityEvent struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id,omitempty"`
	EventType   SecurityEventType      `json:"event_type"`
	Description string                 `json:"description"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Severity    SecuritySeverity       `json:"severity"`
}

// SecurityEventType represents the type of security event
type SecurityEventType string

const (
	SecurityEventLogin              SecurityEventType = "login"
	SecurityEventLoginFailed        SecurityEventType = "login_failed"
	SecurityEventLogout             SecurityEventType = "logout"
	SecurityEventPasswordChange     SecurityEventType = "password_change"
	SecurityEventAccountLocked      SecurityEventType = "account_locked"
	SecurityEventAccountUnlocked    SecurityEventType = "account_unlocked"
	SecurityEventAPIKeyCreated      SecurityEventType = "api_key_created"
	SecurityEventAPIKeyRevoked      SecurityEventType = "api_key_revoked"
	SecurityEventPermissionDenied   SecurityEventType = "permission_denied"
	SecurityEventSuspiciousActivity SecurityEventType = "suspicious_activity"
)

// SecuritySeverity represents the severity of a security event
type SecuritySeverity string

const (
	SecuritySeverityLow      SecuritySeverity = "low"
	SecuritySeverityMedium   SecuritySeverity = "medium"
	SecuritySeverityHigh     SecuritySeverity = "high"
	SecuritySeverityCritical SecuritySeverity = "critical"
)

// SecurityEventFilter represents filters for querying security events
type SecurityEventFilter struct {
	UserID    *string            `json:"user_id,omitempty"`
	EventType *SecurityEventType `json:"event_type,omitempty"`
	Severity  *SecuritySeverity  `json:"severity,omitempty"`
	StartTime *time.Time         `json:"start_time,omitempty"`
	EndTime   *time.Time         `json:"end_time,omitempty"`
	Limit     int                `json:"limit,omitempty"`
	Offset    int                `json:"offset,omitempty"`
}
