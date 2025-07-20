package handlers

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"m-data-storage/api/middleware"
	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
	"m-data-storage/internal/infrastructure/logger"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService          interfaces.AuthService
	userService          interfaces.UserService
	tokenService         interfaces.TokenService
	apiKeyService        interfaces.APIKeyService
	authorizationService interfaces.AuthorizationService
	permissionService    interfaces.PermissionService
	securityService      interfaces.SecurityService
	passwordService      interfaces.PasswordService
	logger               logger.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	authService interfaces.AuthService,
	userService interfaces.UserService,
	tokenService interfaces.TokenService,
	apiKeyService interfaces.APIKeyService,
	authorizationService interfaces.AuthorizationService,
	permissionService interfaces.PermissionService,
	securityService interfaces.SecurityService,
	passwordService interfaces.PasswordService,
	logger logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService:          authService,
		userService:          userService,
		tokenService:         tokenService,
		apiKeyService:        apiKeyService,
		authorizationService: authorizationService,
		permissionService:    permissionService,
		securityService:      securityService,
		passwordService:      passwordService,
		logger:               logger,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         UserInfo  `json:"user"`
}

// UserInfo represents user information in responses
type UserInfo struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	RoleID   string `json:"role_id,omitempty"`
}

// CreateAPIKeyRequest represents an API key creation request
type CreateAPIKeyRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=100"`
	Permissions []string `json:"permissions,omitempty"`
}

// APIKeyResponse represents an API key response
type APIKeyResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Key         string     `json:"key,omitempty"` // Only returned on creation
	Permissions []string   `json:"permissions"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	IsActive    bool       `json:"is_active"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	RoleID   string `json:"role_id,omitempty"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	ctx := r.Context()

	// Create login request for AuthService
	loginReq := &entities.LoginRequest{
		Username:  req.Username,
		Password:  req.Password,
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	}

	// Use AuthService to handle login
	loginResponse, err := h.authService.Login(ctx, loginReq)
	if err != nil {
		h.logger.Warning("Login failed", "username", req.Username, "error", err.Error())
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	// Return the login response directly
	h.writeJSONResponse(w, http.StatusOK, loginResponse)
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetAuthContext(r)
	if authCtx == nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated", nil)
		return
	}

	ctx := r.Context()

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "No authorization header", nil)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid authorization header format", nil)
		return
	}

	// Revoke token
	if err := h.tokenService.RevokeToken(ctx, token); err != nil {
		h.logger.Error("Failed to revoke token", "user_id", authCtx.UserID, "error", err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Logout failed", err)
		return
	}

	// Log security event
	event := interfaces.SecurityEvent{
		UserID:      authCtx.UserID,
		EventType:   interfaces.SecurityEventLogout,
		Description: "User logged out successfully",
		IPAddress:   getClientIP(r),
		UserAgent:   r.UserAgent(),
		Timestamp:   time.Now(),
		Severity:    interfaces.SecuritySeverityLow,
	}
	h.securityService.LogSecurityEvent(ctx, event)

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	ctx := r.Context()

	// Validate refresh token and get user
	user, err := h.tokenService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		h.logger.Warning("Token refresh failed", "error", err.Error())
		h.writeErrorResponse(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	// Generate new tokens
	accessToken, accessExpiry, err := h.tokenService.GenerateAccessToken(ctx, user)
	if err != nil {
		h.logger.Error("Failed to generate access token", "user_id", user.ID, "error", err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Token generation failed", err)
		return
	}

	refreshToken, _, err := h.tokenService.GenerateRefreshToken(ctx, user)
	if err != nil {
		h.logger.Error("Failed to generate refresh token", "user_id", user.ID, "error", err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Token generation failed", err)
		return
	}

	response := map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_at":    accessExpiry,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	ctx := r.Context()

	// Create user through auth service
	user, err := h.authService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		h.logger.Warning("User registration failed", "username", req.Username, "email", req.Email, "error", err.Error())
		h.writeErrorResponse(w, http.StatusBadRequest, "Registration failed", err)
		return
	}

	// Log security event
	event := interfaces.SecurityEvent{
		UserID:      user.ID,
		EventType:   interfaces.SecurityEventType("register"),
		Description: "User registered successfully",
		IPAddress:   getClientIP(r),
		UserAgent:   r.UserAgent(),
		Timestamp:   time.Now(),
		Severity:    interfaces.SecuritySeverityLow,
	}
	h.securityService.LogSecurityEvent(ctx, event)

	userInfo := UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    []string{}, // Will be populated by role service
	}

	h.writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"message": "User registered successfully",
		"user":    userInfo,
	})
}

// writeJSONResponse writes a JSON response
func (h *AuthHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err.Error())
	}
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetAuthContext(r)
	if authCtx == nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated", nil)
		return
	}

	ctx := r.Context()

	// Get user details
	user, err := h.authService.GetUserByID(ctx, authCtx.UserID)
	if err != nil {
		h.logger.Error("Failed to get user profile", "user_id", authCtx.UserID, "error", err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get profile", err)
		return
	}

	// Get user roles (if role is loaded)
	roles := []string{}
	if user.Role != nil {
		roles = []string{user.Role.Name}
	}

	userInfo := UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    roles,
	}

	h.writeJSONResponse(w, http.StatusOK, userInfo)
}

// UpdateProfile updates the current user's profile
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetAuthContext(r)
	if authCtx == nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated", nil)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	ctx := r.Context()

	// Create update request
	updateReq := &entities.UpdateUserRequest{
		Email:     &req.Email,
		FirstName: nil, // Not provided in API request
		LastName:  nil, // Not provided in API request
		RoleID:    &req.RoleID,
	}

	// Update user
	user, err := h.userService.UpdateUser(ctx, authCtx.UserID, updateReq)
	if err != nil {
		h.logger.Error("Failed to update user profile", "user_id", authCtx.UserID, "error", err.Error())
		h.writeErrorResponse(w, http.StatusBadRequest, "Failed to update profile", err)
		return
	}

	// Log security event
	event := interfaces.SecurityEvent{
		UserID:      authCtx.UserID,
		EventType:   interfaces.SecurityEventType("profile_update"),
		Description: "User profile updated",
		IPAddress:   getClientIP(r),
		UserAgent:   r.UserAgent(),
		Timestamp:   time.Now(),
		Severity:    interfaces.SecuritySeverityLow,
	}
	h.securityService.LogSecurityEvent(ctx, event)

	// Get updated roles (if role is loaded)
	roles := []string{}
	if user.Role != nil {
		roles = []string{user.Role.Name}
	}

	userInfo := UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Roles:    roles,
	}

	h.writeJSONResponse(w, http.StatusOK, userInfo)
}

// ChangePassword changes the current user's password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetAuthContext(r)
	if authCtx == nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated", nil)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	ctx := r.Context()

	// Create change password request
	changeReq := &entities.ChangePasswordRequest{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}

	// Change password
	err := h.authService.ChangePassword(ctx, authCtx.UserID, changeReq)
	if err != nil {
		h.logger.Warning("Password change failed", "user_id", authCtx.UserID, "error", err.Error())
		h.writeErrorResponse(w, http.StatusBadRequest, "Failed to change password", err)
		return
	}

	// Log security event
	event := interfaces.SecurityEvent{
		UserID:      authCtx.UserID,
		EventType:   interfaces.SecurityEventPasswordChange,
		Description: "User password changed",
		IPAddress:   getClientIP(r),
		UserAgent:   r.UserAgent(),
		Timestamp:   time.Now(),
		Severity:    interfaces.SecuritySeverityMedium,
	}
	h.securityService.LogSecurityEvent(ctx, event)

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}

// CreateAPIKey creates a new API key for the current user
func (h *AuthHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetAuthContext(r)
	if authCtx == nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated", nil)
		return
	}

	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	ctx := r.Context()

	// Create API key request
	createReq := &entities.CreateAPIKeyRequest{
		Name:        req.Name,
		Permissions: req.Permissions,
	}

	// Create API key
	apiKey, rawKey, err := h.apiKeyService.CreateAPIKey(ctx, authCtx.UserID, createReq)
	if err != nil {
		h.logger.Error("Failed to create API key", "user_id", authCtx.UserID, "error", err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create API key", err)
		return
	}

	// Log security event
	event := interfaces.SecurityEvent{
		UserID:      authCtx.UserID,
		EventType:   interfaces.SecurityEventAPIKeyCreated,
		Description: "API key created: " + req.Name,
		IPAddress:   getClientIP(r),
		UserAgent:   r.UserAgent(),
		Timestamp:   time.Now(),
		Severity:    interfaces.SecuritySeverityMedium,
	}
	h.securityService.LogSecurityEvent(ctx, event)

	response := APIKeyResponse{
		ID:          apiKey.ID,
		Name:        apiKey.Name,
		Key:         rawKey, // Only returned on creation
		Permissions: apiKey.Permissions,
		CreatedAt:   apiKey.CreatedAt,
		LastUsedAt:  apiKey.LastUsedAt,
		IsActive:    apiKey.IsActive,
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

// ListAPIKeys lists all API keys for the current user
func (h *AuthHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetAuthContext(r)
	if authCtx == nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated", nil)
		return
	}

	ctx := r.Context()

	// Get user's API keys
	apiKeys, err := h.apiKeyService.GetAPIKeysByUserID(ctx, authCtx.UserID)
	if err != nil {
		h.logger.Error("Failed to get API keys", "user_id", authCtx.UserID, "error", err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get API keys", err)
		return
	}

	// Convert to response format (without raw keys)
	response := make([]APIKeyResponse, len(apiKeys))
	for i, key := range apiKeys {
		response[i] = APIKeyResponse{
			ID:          key.ID,
			Name:        key.Name,
			Permissions: key.Permissions,
			CreatedAt:   key.CreatedAt,
			LastUsedAt:  key.LastUsedAt,
			IsActive:    key.IsActive,
		}
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// RevokeAPIKey revokes an API key
func (h *AuthHandler) RevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetAuthContext(r)
	if authCtx == nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Not authenticated", nil)
		return
	}

	vars := mux.Vars(r)
	keyID := vars["id"]
	if keyID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "API key ID is required", nil)
		return
	}

	ctx := r.Context()

	// Get API key to verify ownership
	apiKey, err := h.apiKeyService.GetAPIKeyByID(ctx, keyID)
	if err != nil {
		h.logger.Error("Failed to get API key", "key_id", keyID, "error", err.Error())
		h.writeErrorResponse(w, http.StatusNotFound, "API key not found", err)
		return
	}

	// Verify ownership
	if apiKey.UserID != authCtx.UserID {
		h.writeErrorResponse(w, http.StatusForbidden, "Access denied", nil)
		return
	}

	// Revoke API key
	if err := h.apiKeyService.RevokeAPIKey(ctx, keyID); err != nil {
		h.logger.Error("Failed to revoke API key", "key_id", keyID, "error", err.Error())
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to revoke API key", err)
		return
	}

	// Log security event
	event := interfaces.SecurityEvent{
		UserID:      authCtx.UserID,
		EventType:   interfaces.SecurityEventAPIKeyRevoked,
		Description: "API key revoked: " + apiKey.Name,
		IPAddress:   getClientIP(r),
		UserAgent:   r.UserAgent(),
		Timestamp:   time.Now(),
		Severity:    interfaces.SecuritySeverityMedium,
	}
	h.securityService.LogSecurityEvent(ctx, event)

	h.writeJSONResponse(w, http.StatusOK, map[string]string{"message": "API key revoked successfully"})
}

// writeErrorResponse writes an error response
func (h *AuthHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	response := map[string]interface{}{
		"error":   message,
		"status":  statusCode,
		"success": false,
	}

	if err != nil {
		h.logger.Error("API error", "message", message, "error", err.Error())
	}

	h.writeJSONResponse(w, statusCode, response)
}
