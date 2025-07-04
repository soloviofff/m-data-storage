package middleware

import (
	"encoding/json"
	"net/http"
	"runtime/debug"

	"m-data-storage/api/dto"
	"m-data-storage/internal/infrastructure/logger"
)

// RecoveryMiddleware handles panic recovery
type RecoveryMiddleware struct {
	logger *logger.Logger
}

// NewRecoveryMiddleware creates a new RecoveryMiddleware instance
func NewRecoveryMiddleware(logger *logger.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		logger: logger,
	}
}

// Recover recovers from panics and returns a proper error response
func (m *RecoveryMiddleware) Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				m.logger.WithFields(map[string]interface{}{
					"panic":      err,
					"stack":      string(debug.Stack()),
					"method":     r.Method,
					"path":       r.URL.Path,
					"remote_ip":  r.RemoteAddr,
					"user_agent": r.UserAgent(),
				}).Error("Panic recovered")

				// Return error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				response := dto.Response{
					Success: false,
					Error: &dto.Error{
						Code:    "INTERNAL_SERVER_ERROR",
						Message: "Internal server error occurred",
					},
				}

				json.NewEncoder(w).Encode(response)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
