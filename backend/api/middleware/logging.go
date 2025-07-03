package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware handles request logging
type LoggingMiddleware struct {
	logger *log.Logger
}

// NewLoggingMiddleware creates a new LoggingMiddleware instance
func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
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

		// Log request details
		m.logger.Printf(
			"Method: %s, Path: %s, Status: %d, Duration: %v, IP: %s, UserAgent: %s",
			r.Method,
			r.URL.Path,
			wrapped.status,
			time.Since(start),
			r.RemoteAddr,
			r.UserAgent(),
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
