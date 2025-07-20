package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"m-data-storage/internal/domain/interfaces"
	"m-data-storage/internal/infrastructure/logger"
)

// SecurityMiddleware handles security-related middleware
type SecurityMiddleware struct {
	securityService interfaces.SecurityService
	logger          *logger.Logger
}

// NewSecurityMiddleware creates a new SecurityMiddleware instance
func NewSecurityMiddleware(
	securityService interfaces.SecurityService,
	logger *logger.Logger,
) *SecurityMiddleware {
	return &SecurityMiddleware{
		securityService: securityService,
		logger:          logger,
	}
}

// RateLimit middleware that applies rate limiting based on user or IP
func (m *SecurityMiddleware) RateLimit(action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get identifier (user ID if authenticated, otherwise IP)
			identifier := getClientIP(r)
			if authCtx := GetAuthContext(r); authCtx != nil {
				identifier = authCtx.UserID
			}

			// Check rate limit
			allowed, err := m.securityService.CheckRateLimit(r.Context(), identifier, action)
			if err != nil {
				m.logger.Error("Rate limit check failed", "error", err.Error(), "identifier", identifier, "action", action)
				m.writeInternalError(w, "Rate limit check failed")
				return
			}

			if !allowed {
				m.logger.Warn("Rate limit exceeded", "identifier", identifier, "action", action, "ip", getClientIP(r))
				m.writeRateLimitError(w, "Rate limit exceeded")
				return
			}

			// Increment rate limit counter
			if err := m.securityService.IncrementRateLimit(r.Context(), identifier, action); err != nil {
				m.logger.Error("Failed to increment rate limit", "error", err.Error(), "identifier", identifier, "action", action)
				// Continue anyway - don't fail the request for this
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders middleware that adds security headers
func (m *SecurityMiddleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Remove server information
		w.Header().Del("Server")
		w.Header().Del("X-Powered-By")

		next.ServeHTTP(w, r)
	})
}

// SecurityLogging middleware that logs security events
func (m *SecurityMiddleware) SecurityLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &securityResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Process request
		next.ServeHTTP(wrapper, r)

		// Log security-relevant events
		duration := time.Since(start)
		authCtx := GetAuthContext(r)

		logData := map[string]interface{}{
			"method":     r.Method,
			"path":       r.URL.Path,
			"status":     wrapper.statusCode,
			"duration":   duration.Milliseconds(),
			"ip":         getClientIP(r),
			"user_agent": r.UserAgent(),
		}

		if authCtx != nil {
			logData["user_id"] = authCtx.UserID
			logData["auth_method"] = authCtx.AuthMethod
			if authCtx.APIKeyID != "" {
				logData["api_key_id"] = authCtx.APIKeyID
			}
		}

		// Log based on status code
		switch {
		case wrapper.statusCode >= 500:
			m.logger.Error("HTTP request failed", logData)
		case wrapper.statusCode >= 400:
			m.logger.Warn("HTTP request error", logData)
		default:
			m.logger.Info("HTTP request", logData)
		}

		// Log security events
		if wrapper.statusCode == http.StatusUnauthorized {
			m.logSecurityEvent(r.Context(), "authentication_failed", logData)
		} else if wrapper.statusCode == http.StatusForbidden {
			m.logSecurityEvent(r.Context(), "authorization_failed", logData)
		}
	})
}

// securityResponseWriter wrapper to capture status code
type securityResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *securityResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Helper methods
func (m *SecurityMiddleware) logSecurityEvent(ctx context.Context, eventType string, data map[string]interface{}) {
	// For now, just log using the logger since SecurityService.LogSecurityEvent expects SecurityEvent
	// In a full implementation, we would create a SecurityEvent struct
	m.logger.Warn("Security event", "event_type", eventType, "data", data)
}

func (m *SecurityMiddleware) writeRateLimitError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "60") // Suggest retry after 60 seconds
	w.WriteHeader(http.StatusTooManyRequests)
	response := map[string]interface{}{
		"error":   "rate_limit_exceeded",
		"message": message,
	}
	json.NewEncoder(w).Encode(response)
}

func (m *SecurityMiddleware) writeInternalError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	response := map[string]interface{}{
		"error":   "internal_error",
		"message": message,
	}
	json.NewEncoder(w).Encode(response)
}
