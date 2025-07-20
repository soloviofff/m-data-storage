package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/errors"
	"m-data-storage/internal/domain/interfaces"
)

// AuthService implements authentication operations
type AuthService struct {
	userStorage     interfaces.UserStorage
	tokenService    interfaces.TokenService
	passwordService interfaces.PasswordService
	securityService interfaces.SecurityService
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userStorage interfaces.UserStorage,
	tokenService interfaces.TokenService,
	passwordService interfaces.PasswordService,
	securityService interfaces.SecurityService,
) interfaces.AuthService {
	return &AuthService{
		userStorage:     userStorage,
		tokenService:    tokenService,
		passwordService: passwordService,
		securityService: securityService,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, username, email, password string) (*entities.User, error) {
	if username == "" || email == "" || password == "" {
		return nil, errors.NewAuthError(errors.CodeInvalidInput, "username, email, and password are required", nil)
	}

	// Validate password strength
	if err := s.passwordService.ValidatePasswordStrength(password); err != nil {
		return nil, err
	}

	// Check if user already exists
	existingUser, err := s.userStorage.GetUserByUsername(ctx, username)
	if err != nil && !errors.IsUserNotFound(err) {
		return nil, errors.NewAuthError("INTERNAL_ERROR", "failed to check username", err)
	}
	if existingUser != nil {
		return nil, errors.NewAuthError(errors.CodeUserExists, "username already exists", nil)
	}

	existingUser, err = s.userStorage.GetUserByEmail(ctx, email)
	if err != nil && !errors.IsUserNotFound(err) {
		return nil, errors.NewAuthError("INTERNAL_ERROR", "failed to check email", err)
	}
	if existingUser != nil {
		return nil, errors.NewAuthError(errors.CodeUserExists, "email already exists", nil)
	}

	// Hash password
	passwordHash, err := s.passwordService.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user entity
	user := &entities.User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Status:       entities.UserStatusActive,
		RoleID:       "user-role-id", // Default user role
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save user to storage
	err = s.userStorage.CreateUser(ctx, user)
	if err != nil {
		return nil, errors.NewAuthError("INTERNAL_ERROR", "failed to create user", err)
	}

	// Log security event
	event := interfaces.SecurityEvent{
		UserID:      user.ID,
		EventType:   "user_registered",
		Description: "New user registered",
		Severity:    interfaces.SecuritySeverityLow,
	}
	s.securityService.LogSecurityEvent(ctx, event)

	return user, nil
}

// Login authenticates a user and creates a session
func (s *AuthService) Login(ctx context.Context, req *entities.LoginRequest) (*entities.LoginResponse, error) {
	if req == nil {
		return nil, errors.NewAuthError(errors.CodeInvalidInput, "login request cannot be nil", nil)
	}

	if req.Username == "" || req.Password == "" {
		return nil, errors.NewAuthError(errors.CodeInvalidCredentials, "username and password are required", nil)
	}

	// Check rate limiting
	allowed, err := s.securityService.CheckRateLimit(ctx, req.Username, "login")
	if err != nil {
		return nil, err
	}
	if !allowed {
		// Log failed login attempt
		event := interfaces.SecurityEvent{
			EventType:   interfaces.SecurityEventLoginFailed,
			Description: "Rate limit exceeded for login attempts",
			IPAddress:   req.IPAddress,
			UserAgent:   req.UserAgent,
			Severity:    interfaces.SecuritySeverityHigh,
		}
		s.securityService.LogSecurityEvent(ctx, event)

		return nil, errors.NewAuthError(errors.CodeRateLimitExceeded, "too many login attempts", nil)
	}

	// Get user by username
	user, err := s.userStorage.GetUserByUsername(ctx, req.Username)
	if err != nil {
		// Log failed login attempt
		event := interfaces.SecurityEvent{
			EventType:   interfaces.SecurityEventLoginFailed,
			Description: "User not found",
			IPAddress:   req.IPAddress,
			UserAgent:   req.UserAgent,
			Severity:    interfaces.SecuritySeverityMedium,
		}
		s.securityService.LogSecurityEvent(ctx, event)

		return nil, errors.NewAuthError(errors.CodeInvalidCredentials, "invalid username or password", nil)
	}

	// Check if account is locked
	isLocked, lockoutExpiration, err := s.securityService.CheckAccountLockout(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if isLocked {
		return nil, errors.NewAuthError(errors.CodeAccountLocked,
			"account is locked until "+lockoutExpiration.Format(time.RFC3339), nil)
	}

	// Validate password
	if !s.passwordService.ValidatePassword(req.Password, user.PasswordHash) {
		// Log failed login attempt
		event := interfaces.SecurityEvent{
			UserID:      user.ID,
			EventType:   interfaces.SecurityEventLoginFailed,
			Description: "Invalid password",
			IPAddress:   req.IPAddress,
			UserAgent:   req.UserAgent,
			Severity:    interfaces.SecuritySeverityMedium,
		}
		s.securityService.LogSecurityEvent(ctx, event)

		return nil, errors.NewAuthError(errors.CodeInvalidCredentials, "invalid username or password", nil)
	}

	// Check user status
	if user.Status != entities.UserStatusActive {
		return nil, errors.NewAuthError(errors.CodeUserInactive, "user account is not active", nil)
	}

	// Generate tokens
	accessToken, accessExpiry, err := s.tokenService.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.tokenService.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, err
	}

	// Create session
	session := &entities.UserSession{
		ID:           uuid.New().String(),
		UserID:       user.ID,
		TokenHash:    "", // TODO: Hash the access token
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		CreatedAt:    time.Now(),
		LastUsedAt:   time.Now(),
	}

	err = s.userStorage.CreateSession(ctx, session)
	if err != nil {
		return nil, errors.NewAuthError("INTERNAL_ERROR", "failed to create session", err)
	}

	// Log successful login
	event := interfaces.SecurityEvent{
		UserID:      user.ID,
		EventType:   interfaces.SecurityEventLogin,
		Description: "User logged in successfully",
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
		Severity:    interfaces.SecuritySeverityLow,
	}
	s.securityService.LogSecurityEvent(ctx, event)

	// Update user last login
	now := time.Now()
	user.LastLoginAt = &now
	s.userStorage.UpdateUser(ctx, user)

	return &entities.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
		User:         user,
	}, nil
}

// Logout logs out a user by invalidating their session
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "session ID cannot be empty", nil)
	}

	// Get session
	session, err := s.userStorage.GetSessionByID(ctx, sessionID)
	if err != nil {
		return errors.NewAuthError(errors.CodeSessionNotFound, "session not found", err)
	}

	// Revoke session
	if err := s.userStorage.RevokeSession(ctx, sessionID); err != nil {
		return errors.NewAuthError("INTERNAL_ERROR", "failed to revoke session", err)
	}

	// Log logout
	event := interfaces.SecurityEvent{
		UserID:      session.UserID,
		EventType:   interfaces.SecurityEventLogout,
		Description: "User logged out",
		Severity:    interfaces.SecuritySeverityLow,
	}
	s.securityService.LogSecurityEvent(ctx, event)

	return nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, req *entities.RefreshTokenRequest) (*entities.LoginResponse, error) {
	if req == nil {
		return nil, errors.NewAuthError(errors.CodeInvalidInput, "refresh token request cannot be nil", nil)
	}

	if req.RefreshToken == "" {
		return nil, errors.NewAuthError(errors.CodeInvalidInput, "refresh token cannot be empty", nil)
	}

	// Validate refresh token
	user, err := s.tokenService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Get user with full details
	fullUser, err := s.userStorage.GetUserByID(ctx, user.ID)
	if err != nil {
		return nil, errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}

	// Check user status
	if fullUser.Status != entities.UserStatusActive {
		return nil, errors.NewAuthError(errors.CodeUserInactive, "user account is not active", nil)
	}

	// Generate new tokens
	accessToken, accessExpiry, err := s.tokenService.GenerateAccessToken(ctx, fullUser)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := s.tokenService.GenerateRefreshToken(ctx, fullUser)
	if err != nil {
		return nil, err
	}

	// Get user sessions and find the one with matching refresh token
	sessions, err := s.userStorage.GetSessionsByUserID(ctx, fullUser.ID)
	if err != nil {
		return nil, errors.NewAuthError("INTERNAL_ERROR", "failed to get user sessions", err)
	}

	var targetSession *entities.UserSession
	for _, session := range sessions {
		if session.RefreshToken == req.RefreshToken && !session.IsRevoked {
			targetSession = session
			break
		}
	}

	if targetSession == nil {
		return nil, errors.NewAuthError(errors.CodeSessionNotFound, "session not found", nil)
	}

	// Update session with new tokens
	targetSession.RefreshToken = refreshToken
	targetSession.ExpiresAt = accessExpiry
	targetSession.LastUsedAt = time.Now()

	err = s.userStorage.UpdateSession(ctx, targetSession)
	if err != nil {
		return nil, errors.NewAuthError("INTERNAL_ERROR", "failed to update session", err)
	}

	return &entities.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    accessExpiry,
		User:         fullUser,
	}, nil
}

// ValidateSession validates a session and returns the user
func (s *AuthService) ValidateSession(ctx context.Context, sessionID string) (*entities.User, error) {
	if sessionID == "" {
		return nil, errors.NewAuthError(errors.CodeInvalidInput, "session ID cannot be empty", nil)
	}

	// Get session
	session, err := s.userStorage.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, errors.NewAuthError(errors.CodeSessionNotFound, "session not found", err)
	}

	// Check if session is valid
	if !session.IsValid() {
		return nil, errors.NewAuthError(errors.CodeSessionExpired, "session is expired or revoked", nil)
	}

	// Get user
	user, err := s.userStorage.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}

	// Check user status
	if user.Status != entities.UserStatusActive {
		return nil, errors.NewAuthError(errors.CodeUserInactive, "user account is not active", nil)
	}

	return user, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID string, req *entities.ChangePasswordRequest) error {
	if userID == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "user ID cannot be empty", nil)
	}

	if req == nil {
		return errors.NewAuthError(errors.CodeInvalidInput, "change password request cannot be nil", nil)
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "current and new passwords are required", nil)
	}

	// Get user
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}

	// Validate current password
	if !s.passwordService.ValidatePassword(req.CurrentPassword, user.PasswordHash) {
		return errors.NewAuthError(errors.CodeInvalidCredentials, "current password is incorrect", nil)
	}

	// Validate new password strength
	if err := s.passwordService.ValidatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	// Hash new password
	newPasswordHash, err := s.passwordService.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update user password
	user.PasswordHash = newPasswordHash
	err = s.userStorage.UpdateUser(ctx, user)
	if err != nil {
		return errors.NewAuthError("INTERNAL_ERROR", "failed to update password", err)
	}

	// Log password change
	event := interfaces.SecurityEvent{
		UserID:      userID,
		EventType:   interfaces.SecurityEventPasswordChange,
		Description: "Password changed successfully",
		Severity:    interfaces.SecuritySeverityMedium,
	}
	s.securityService.LogSecurityEvent(ctx, event)

	// Revoke all user sessions except current one
	if req.KeepCurrentSession && req.CurrentSessionID != "" {
		sessions, err := s.userStorage.GetSessionsByUserID(ctx, userID)
		if err == nil {
			for _, session := range sessions {
				if session.ID != req.CurrentSessionID {
					s.userStorage.RevokeSession(ctx, session.ID)
				}
			}
		}
	} else {
		s.userStorage.RevokeUserSessions(ctx, userID)
	}

	return nil
}

// ResetPassword initiates a password reset process
func (s *AuthService) ResetPassword(ctx context.Context, email string) error {
	if email == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "email cannot be empty", nil)
	}

	// Check rate limiting
	allowed, err := s.securityService.CheckRateLimit(ctx, email, "password_reset")
	if err != nil {
		return err
	}
	if !allowed {
		return errors.NewAuthError(errors.CodeRateLimitExceeded, "too many password reset attempts", nil)
	}

	// Get user by email
	user, err := s.userStorage.GetUserByEmail(ctx, email)
	if err != nil {
		// Don't reveal if email exists or not for security
		return nil
	}

	// Check user status
	if user.Status != entities.UserStatusActive {
		return nil // Don't reveal user status
	}

	// Generate reset token
	_, expiresAt, err := s.passwordService.GenerateResetToken(user.ID)
	if err != nil {
		return err
	}

	// TODO: Store reset token in database
	// TODO: Send reset email to user

	// Log password reset request
	event := interfaces.SecurityEvent{
		UserID:      user.ID,
		EventType:   interfaces.SecurityEventPasswordChange,
		Description: "Password reset requested",
		Severity:    interfaces.SecuritySeverityMedium,
		Metadata: map[string]interface{}{
			"reset_token_expires": expiresAt,
		},
	}
	s.securityService.LogSecurityEvent(ctx, event)

	return nil
}

// ValidateResetToken validates a password reset token
func (s *AuthService) ValidateResetToken(ctx context.Context, token string) (*entities.User, error) {
	if token == "" {
		return nil, errors.NewAuthError(errors.CodeInvalidInput, "reset token cannot be empty", nil)
	}

	// Validate token format and get user ID
	userID, err := s.passwordService.ValidateResetToken(token)
	if err != nil {
		return nil, err
	}

	// Get user
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}

	// TODO: Check if token exists in database and is not expired

	return user, nil
}

// SetNewPassword sets a new password using a reset token
func (s *AuthService) SetNewPassword(ctx context.Context, token, newPassword string) error {
	if token == "" || newPassword == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "token and new password cannot be empty", nil)
	}

	// Validate reset token
	user, err := s.ValidateResetToken(ctx, token)
	if err != nil {
		return err
	}

	// Validate new password strength
	if err := s.passwordService.ValidatePasswordStrength(newPassword); err != nil {
		return err
	}

	// Hash new password
	newPasswordHash, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update user password
	user.PasswordHash = newPasswordHash
	err = s.userStorage.UpdateUser(ctx, user)
	if err != nil {
		return errors.NewAuthError("INTERNAL_ERROR", "failed to update password", err)
	}

	// TODO: Invalidate reset token in database

	// Log password reset completion
	event := interfaces.SecurityEvent{
		UserID:      user.ID,
		EventType:   interfaces.SecurityEventPasswordChange,
		Description: "Password reset completed",
		Severity:    interfaces.SecuritySeverityMedium,
	}
	s.securityService.LogSecurityEvent(ctx, event)

	// Revoke all user sessions
	s.userStorage.RevokeUserSessions(ctx, user.ID)

	return nil
}

// GetUserSessions returns all sessions for a user
func (s *AuthService) GetUserSessions(ctx context.Context, userID string) ([]*entities.UserSession, error) {
	if userID == "" {
		return nil, errors.NewAuthError(errors.CodeInvalidInput, "user ID cannot be empty", nil)
	}

	sessions, err := s.userStorage.GetSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, errors.NewAuthError("INTERNAL_ERROR", "failed to get user sessions", err)
	}

	return sessions, nil
}

// RevokeSession revokes a specific session
func (s *AuthService) RevokeSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "session ID cannot be empty", nil)
	}

	// Get session to log the user
	session, err := s.userStorage.GetSessionByID(ctx, sessionID)
	if err != nil {
		return errors.NewAuthError(errors.CodeSessionNotFound, "session not found", err)
	}

	// Revoke session
	if err := s.userStorage.RevokeSession(ctx, sessionID); err != nil {
		return errors.NewAuthError("INTERNAL_ERROR", "failed to revoke session", err)
	}

	// Log session revocation
	event := interfaces.SecurityEvent{
		UserID:      session.UserID,
		EventType:   interfaces.SecurityEventLogout,
		Description: "Session revoked",
		Severity:    interfaces.SecuritySeverityLow,
	}
	s.securityService.LogSecurityEvent(ctx, event)

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (s *AuthService) RevokeAllUserSessions(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "user ID cannot be empty", nil)
	}

	// Revoke all sessions
	if err := s.userStorage.RevokeUserSessions(ctx, userID); err != nil {
		return errors.NewAuthError("INTERNAL_ERROR", "failed to revoke all user sessions", err)
	}

	// Log session revocation
	event := interfaces.SecurityEvent{
		UserID:      userID,
		EventType:   interfaces.SecurityEventLogout,
		Description: "All sessions revoked",
		Severity:    interfaces.SecuritySeverityMedium,
	}
	s.securityService.LogSecurityEvent(ctx, event)

	return nil
}

// GetUserByID retrieves a user by their ID
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*entities.User, error) {
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}
	return user, nil
}
