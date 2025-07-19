package errors

import (
	"errors"
	"fmt"
)

// Authentication errors
var (
	// User errors
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrUserInactive         = errors.New("user account is inactive")
	ErrUserSuspended        = errors.New("user account is suspended")
	ErrUserDeleted          = errors.New("user account is deleted")
	ErrInvalidCredentials   = errors.New("invalid username or password")
	ErrAccountLocked        = errors.New("account is locked due to too many failed login attempts")

	// Password errors
	ErrInvalidPassword      = errors.New("invalid password")
	ErrPasswordTooWeak      = errors.New("password does not meet security requirements")
	ErrPasswordMismatch     = errors.New("current password is incorrect")
	ErrSamePassword         = errors.New("new password must be different from current password")

	// Session errors
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionExpired       = errors.New("session has expired")
	ErrSessionRevoked       = errors.New("session has been revoked")
	ErrInvalidSession       = errors.New("invalid session")

	// Token errors
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token has expired")
	ErrTokenRevoked         = errors.New("token has been revoked")
	ErrInvalidTokenFormat   = errors.New("invalid token format")
	ErrTokenGenerationFailed = errors.New("failed to generate token")

	// API Key errors
	ErrAPIKeyNotFound       = errors.New("API key not found")
	ErrAPIKeyInactive       = errors.New("API key is inactive")
	ErrAPIKeyExpired        = errors.New("API key has expired")
	ErrInvalidAPIKey        = errors.New("invalid API key")
	ErrAPIKeyRevoked        = errors.New("API key has been revoked")
	ErrAPIKeyLimitExceeded  = errors.New("API key limit exceeded")

	// Role and Permission errors
	ErrRoleNotFound         = errors.New("role not found")
	ErrRoleAlreadyExists    = errors.New("role already exists")
	ErrPermissionNotFound   = errors.New("permission not found")
	ErrPermissionAlreadyExists = errors.New("permission already exists")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrCannotDeleteSystemRole = errors.New("cannot delete system role")
	ErrCannotDeleteSystemPermission = errors.New("cannot delete system permission")

	// Authorization errors
	ErrUnauthorized         = errors.New("unauthorized access")
	ErrForbidden            = errors.New("access forbidden")
	ErrPermissionDenied     = errors.New("permission denied")
	ErrInvalidScope         = errors.New("invalid scope")

	// Rate limiting errors
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")
	ErrTooManyRequests      = errors.New("too many requests")

	// Validation errors
	ErrInvalidInput         = errors.New("invalid input")
	ErrMissingRequiredField = errors.New("missing required field")
	ErrInvalidEmailFormat   = errors.New("invalid email format")
	ErrInvalidUsernameFormat = errors.New("invalid username format")
	ErrUsernameTooShort     = errors.New("username is too short")
	ErrUsernameTooLong      = errors.New("username is too long")
	ErrPasswordTooShort     = errors.New("password is too short")
	ErrPasswordTooLong      = errors.New("password is too long")

	// Security errors
	ErrSuspiciousActivity   = errors.New("suspicious activity detected")
	ErrSecurityViolation    = errors.New("security violation")
	ErrIPBlocked            = errors.New("IP address is blocked")
	ErrDeviceNotTrusted     = errors.New("device is not trusted")
)

// AuthError represents an authentication-related error with additional context
type AuthError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Cause   error                  `json:"-"`
}

// Error implements the error interface
func (e *AuthError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AuthError) Unwrap() error {
	return e.Cause
}

// NewAuthError creates a new authentication error
func NewAuthError(code, message string, cause error) *AuthError {
	return &AuthError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Details: make(map[string]interface{}),
	}
}

// WithDetail adds a detail to the error
func (e *AuthError) WithDetail(key string, value interface{}) *AuthError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Error codes for structured error handling
const (
	// User error codes
	CodeUserNotFound      = "USER_NOT_FOUND"
	CodeUserExists        = "USER_ALREADY_EXISTS"
	CodeUserInactive      = "USER_INACTIVE"
	CodeUserSuspended     = "USER_SUSPENDED"
	CodeUserDeleted       = "USER_DELETED"
	CodeAccountLocked     = "ACCOUNT_LOCKED"

	// Authentication error codes
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
	CodeInvalidPassword   = "INVALID_PASSWORD"
	CodePasswordTooWeak   = "PASSWORD_TOO_WEAK"
	CodePasswordMismatch  = "PASSWORD_MISMATCH"

	// Session error codes
	CodeSessionNotFound   = "SESSION_NOT_FOUND"
	CodeSessionExpired    = "SESSION_EXPIRED"
	CodeSessionRevoked    = "SESSION_REVOKED"
	CodeInvalidSession    = "INVALID_SESSION"

	// Token error codes
	CodeInvalidToken      = "INVALID_TOKEN"
	CodeTokenExpired      = "TOKEN_EXPIRED"
	CodeTokenRevoked      = "TOKEN_REVOKED"

	// API Key error codes
	CodeAPIKeyNotFound    = "API_KEY_NOT_FOUND"
	CodeAPIKeyInactive    = "API_KEY_INACTIVE"
	CodeAPIKeyExpired     = "API_KEY_EXPIRED"
	CodeInvalidAPIKey     = "INVALID_API_KEY"

	// Authorization error codes
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodePermissionDenied  = "PERMISSION_DENIED"
	CodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"

	// Role/Permission error codes
	CodeRoleNotFound      = "ROLE_NOT_FOUND"
	CodeRoleExists        = "ROLE_ALREADY_EXISTS"
	CodePermissionNotFound = "PERMISSION_NOT_FOUND"
	CodePermissionExists  = "PERMISSION_ALREADY_EXISTS"

	// Rate limiting error codes
	CodeRateLimitExceeded = "RATE_LIMIT_EXCEEDED"
	CodeTooManyRequests   = "TOO_MANY_REQUESTS"

	// Validation error codes
	CodeInvalidInput      = "INVALID_INPUT"
	CodeValidationFailed  = "VALIDATION_FAILED"

	// Security error codes
	CodeSuspiciousActivity = "SUSPICIOUS_ACTIVITY"
	CodeSecurityViolation = "SECURITY_VIOLATION"
	CodeIPBlocked         = "IP_BLOCKED"
)

// Predefined authentication errors with codes
func NewUserNotFoundError(userID string) *AuthError {
	return NewAuthError(CodeUserNotFound, "User not found", ErrUserNotFound).
		WithDetail("user_id", userID)
}

func NewUserExistsError(username, email string) *AuthError {
	return NewAuthError(CodeUserExists, "User already exists", ErrUserAlreadyExists).
		WithDetail("username", username).
		WithDetail("email", email)
}

func NewInvalidCredentialsError() *AuthError {
	return NewAuthError(CodeInvalidCredentials, "Invalid username or password", ErrInvalidCredentials)
}

func NewAccountLockedError(unlockTime string) *AuthError {
	return NewAuthError(CodeAccountLocked, "Account is locked", ErrAccountLocked).
		WithDetail("unlock_time", unlockTime)
}

func NewSessionExpiredError() *AuthError {
	return NewAuthError(CodeSessionExpired, "Session has expired", ErrSessionExpired)
}

func NewTokenExpiredError() *AuthError {
	return NewAuthError(CodeTokenExpired, "Token has expired", ErrTokenExpired)
}

func NewPermissionDeniedError(permission string) *AuthError {
	return NewAuthError(CodePermissionDenied, "Permission denied", ErrPermissionDenied).
		WithDetail("required_permission", permission)
}

func NewInsufficientPermissionsError(required []string) *AuthError {
	return NewAuthError(CodeInsufficientPermissions, "Insufficient permissions", ErrInsufficientPermissions).
		WithDetail("required_permissions", required)
}

func NewAPIKeyNotFoundError(keyID string) *AuthError {
	return NewAuthError(CodeAPIKeyNotFound, "API key not found", ErrAPIKeyNotFound).
		WithDetail("key_id", keyID)
}

func NewAPIKeyExpiredError() *AuthError {
	return NewAuthError(CodeAPIKeyExpired, "API key has expired", ErrAPIKeyExpired)
}

func NewRateLimitExceededError(limit int, window string) *AuthError {
	return NewAuthError(CodeRateLimitExceeded, "Rate limit exceeded", ErrRateLimitExceeded).
		WithDetail("limit", limit).
		WithDetail("window", window)
}

func NewValidationError(field, message string) *AuthError {
	return NewAuthError(CodeValidationFailed, fmt.Sprintf("Validation failed for field '%s': %s", field, message), ErrInvalidInput).
		WithDetail("field", field).
		WithDetail("validation_message", message)
}

// IsAuthError checks if an error is an authentication error
func IsAuthError(err error) bool {
	var authErr *AuthError
	return errors.As(err, &authErr)
}

// GetAuthErrorCode extracts the error code from an authentication error
func GetAuthErrorCode(err error) string {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.Code
	}
	return ""
}

// IsUserNotFound checks if the error is a user not found error
func IsUserNotFound(err error) bool {
	return errors.Is(err, ErrUserNotFound) || GetAuthErrorCode(err) == CodeUserNotFound
}

// IsUnauthorized checks if the error is an unauthorized error
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized) || GetAuthErrorCode(err) == CodeUnauthorized
}

// IsForbidden checks if the error is a forbidden error
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden) || GetAuthErrorCode(err) == CodeForbidden
}

// IsTokenExpired checks if the error is a token expired error
func IsTokenExpired(err error) bool {
	return errors.Is(err, ErrTokenExpired) || GetAuthErrorCode(err) == CodeTokenExpired
}

// IsSessionExpired checks if the error is a session expired error
func IsSessionExpired(err error) bool {
	return errors.Is(err, ErrSessionExpired) || GetAuthErrorCode(err) == CodeSessionExpired
}
