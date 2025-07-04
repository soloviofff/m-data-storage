package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// RequestIDMiddleware adds a unique request ID to each request
type RequestIDMiddleware struct{}

// NewRequestIDMiddleware creates a new RequestIDMiddleware instance
func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

// RequestID adds a unique request ID to the request context and response headers
func (m *RequestIDMiddleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID already exists in headers
		requestID := r.Header.Get("X-Request-ID")

		// Generate new request ID if not provided
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to request context
		ctx := context.WithValue(r.Context(), "request_id", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRequestID generates a random request ID
func generateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a simple counter-based ID if random generation fails
		return "req-" + hex.EncodeToString([]byte("fallback"))
	}
	return hex.EncodeToString(bytes)
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}
