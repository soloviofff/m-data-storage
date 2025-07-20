package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"m-data-storage/internal/domain/interfaces"
	"m-data-storage/internal/infrastructure/logger"
)

// AuthMiddleware handles JWT and API key authentication
type AuthMiddleware struct {
	tokenService      interfaces.TokenService
	apiKeyService     interfaces.APIKeyService
	authzService      interfaces.AuthorizationService
	permissionService interfaces.PermissionService
	securityService   interfaces.SecurityService
	logger            *logger.Logger
}

// NewAuthMiddleware creates a new AuthMiddleware instance
func NewAuthMiddleware(
	tokenService interfaces.TokenService,
	apiKeyService interfaces.APIKeyService,
	authzService interfaces.AuthorizationService,
	permissionService interfaces.PermissionService,
	securityService interfaces.SecurityService,
	logger *logger.Logger,
) *AuthMiddleware {
	return &AuthMiddleware{
		tokenService:      tokenService,
		apiKeyService:     apiKeyService,
		authzService:      authzService,
		permissionService: permissionService,
		securityService:   securityService,
		logger:            logger,
	}
}

// AuthContext represents authenticated user context
type AuthContext struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	AuthMethod  string   `json:"auth_method"` // "jwt" or "api_key"
	APIKeyID    string   `json:"api_key_id,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// JWTAuthenticate validates JWT token from Authorization header
func (m *AuthMiddleware) JWTAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract JWT token
		tokenStr := extractBearerToken(r)
		if tokenStr == "" {
			m.writeUnauthorizedError(w, "Missing or invalid authorization header")
			return
		}

		// Validate JWT token
		user, err := m.tokenService.ValidateAccessToken(r.Context(), tokenStr)
		if err != nil {
			m.logger.Warn("JWT validation failed", "error", err.Error(), "ip", getClientIP(r))
			m.writeUnauthorizedError(w, "Invalid or expired token")
			return
		}

		// Get user roles
		var roles []string
		if user.Role != nil {
			roles = []string{user.Role.Name}
		}

		// Create auth context
		authCtx := &AuthContext{
			UserID:     user.ID,
			Username:   user.Username,
			Email:      user.Email,
			Roles:      roles,
			AuthMethod: "jwt",
		}

		// Add to request context
		ctx := context.WithValue(r.Context(), "auth", authCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// APIKeyAuthenticate validates API key from X-API-Key header
func (m *AuthMiddleware) APIKeyAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			m.writeUnauthorizedError(w, "Missing API key")
			return
		}

		// Validate API key
		keyInfo, user, err := m.apiKeyService.ValidateAPIKey(r.Context(), apiKey)
		if err != nil {
			m.logger.Warn("API key validation failed", "error", err.Error(), "ip", getClientIP(r))
			m.writeUnauthorizedError(w, "Invalid API key")
			return
		}

		// Get user permissions
		permissions, err := m.permissionService.GetUserPermissions(r.Context(), user.ID)
		if err != nil {
			m.logger.Error("Failed to get user permissions", "error", err.Error(), "user_id", user.ID)
			m.writeInternalError(w, "Authentication error")
			return
		}

		// Create auth context
		authCtx := &AuthContext{
			UserID:      user.ID,
			Username:    user.Username,
			Email:       user.Email,
			AuthMethod:  "api_key",
			APIKeyID:    keyInfo.ID,
			Permissions: permissions,
		}

		// Add to request context
		ctx := context.WithValue(r.Context(), "auth", authCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthenticate tries both JWT and API key authentication (non-blocking)
func (m *AuthMiddleware) OptionalAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var authCtx *AuthContext

		// Try JWT first
		if tokenStr := extractBearerToken(r); tokenStr != "" {
			if user, err := m.tokenService.ValidateAccessToken(r.Context(), tokenStr); err == nil {
				var roles []string
				if user.Role != nil {
					roles = []string{user.Role.Name}
				}
				authCtx = &AuthContext{
					UserID:     user.ID,
					Username:   user.Username,
					Email:      user.Email,
					Roles:      roles,
					AuthMethod: "jwt",
				}
			}
		}

		// Try API key if JWT failed
		if authCtx == nil {
			if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
				if keyInfo, user, err := m.apiKeyService.ValidateAPIKey(r.Context(), apiKey); err == nil {
					if permissions, err := m.permissionService.GetUserPermissions(r.Context(), user.ID); err == nil {
						authCtx = &AuthContext{
							UserID:      user.ID,
							Username:    user.Username,
							Email:       user.Email,
							AuthMethod:  "api_key",
							APIKeyID:    keyInfo.ID,
							Permissions: permissions,
						}
					}
				}
			}
		}

		// Add auth context if available (can be nil for anonymous access)
		ctx := context.WithValue(r.Context(), "auth", authCtx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission middleware that checks if user has required permission
func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx := GetAuthContext(r)
			if authCtx == nil {
				m.writeUnauthorizedError(w, "Authentication required")
				return
			}

			// Check permission
			hasPermission, err := m.authzService.HasPermission(r.Context(), authCtx.UserID, permission)
			if err != nil {
				m.logger.Error("Permission check failed", "error", err.Error(), "user_id", authCtx.UserID, "permission", permission)
				m.writeInternalError(w, "Authorization error")
				return
			}

			if !hasPermission {
				m.logger.Warn("Permission denied", "user_id", authCtx.UserID, "permission", permission, "ip", getClientIP(r))
				m.writeForbiddenError(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole middleware that checks if user has required role
func (m *AuthMiddleware) RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx := GetAuthContext(r)
			if authCtx == nil {
				m.writeUnauthorizedError(w, "Authentication required")
				return
			}

			// Check role
			hasRole, err := m.authzService.HasRole(r.Context(), authCtx.UserID, roleName)
			if err != nil {
				m.logger.Error("Role check failed", "error", err.Error(), "user_id", authCtx.UserID, "role", roleName)
				m.writeInternalError(w, "Authorization error")
				return
			}

			if !hasRole {
				m.logger.Warn("Role denied", "user_id", authCtx.UserID, "role", roleName, "ip", getClientIP(r))
				m.writeForbiddenError(w, "Insufficient role")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetAuthContext extracts auth context from request
func GetAuthContext(r *http.Request) *AuthContext {
	if authCtx := r.Context().Value("auth"); authCtx != nil {
		if ctx, ok := authCtx.(*AuthContext); ok {
			return ctx
		}
	}
	return nil
}

// Helper functions
func extractBearerToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	if bearToken == "" {
		return ""
	}

	parts := strings.Split(bearToken, " ")
	if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
		return parts[1]
	}
	return ""
}

func (m *AuthMiddleware) writeUnauthorizedError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	response := map[string]interface{}{
		"error":   "unauthorized",
		"message": message,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *AuthMiddleware) writeForbiddenError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	response := map[string]interface{}{
		"error":   "forbidden",
		"message": message,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *AuthMiddleware) writeInternalError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	response := map[string]interface{}{
		"error":   "internal_error",
		"message": message,
	}
	json.NewEncoder(w).Encode(response)
}
