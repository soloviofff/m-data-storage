package middleware

import (
	"net/http"
	"time"

	"m-data-storage/internal/infrastructure/logger"
)

// LoggingMiddleware handles request logging
type LoggingMiddleware struct {
	logger *logger.Logger
}

// NewLoggingMiddleware creates a new LoggingMiddleware instance
func NewLoggingMiddleware(logger *logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Log logs request information
func (m *LoggingMiddleware) Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create response wrapper to capture status code
		wrapped := wrapResponseWriter(w)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Get request ID from context
		requestID := GetRequestIDFromContext(r.Context())

		// Calculate duration
		duration := time.Since(start)

		// Log request details using our structured logger
		m.logger.LogAPIRequest(
			requestID,
			r.Method,
			r.URL.Path,
			r.UserAgent(),
			duration.String(),
			wrapped.status,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

// wrapResponseWriter creates new responseWriter
func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w}
}

// WriteHeader captures status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
