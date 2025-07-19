package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/logger"
)

func TestNewLoggingMiddleware(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	middleware := NewLoggingMiddleware(realLogger)

	assert.NotNil(t, middleware)
	assert.Equal(t, realLogger, middleware.logger)
}

func TestLoggingMiddleware_Log(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		userAgent      string
		requestID      string
		handlerStatus  int
		expectedStatus int
	}{
		{
			name:           "GET request with request ID",
			method:         "GET",
			path:           "/api/v1/test",
			userAgent:      "test-agent/1.0",
			requestID:      "test-request-id-123",
			handlerStatus:  http.StatusOK,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST request without request ID",
			method:         "POST",
			path:           "/api/v1/data",
			userAgent:      "curl/7.68.0",
			requestID:      "",
			handlerStatus:  http.StatusCreated,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "error response",
			method:         "GET",
			path:           "/api/v1/notfound",
			userAgent:      "browser/1.0",
			requestID:      "error-request-id",
			handlerStatus:  http.StatusNotFound,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LoggingConfig{
				Level:  "info",
				Format: "json",
				Output: "stdout",
			}
			realLogger, err := logger.New(cfg)
			assert.NoError(t, err)

			middleware := NewLoggingMiddleware(realLogger)

			// Create a test handler that sets the status code
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.handlerStatus)
				w.Write([]byte("test response"))
			})

			// Wrap with logging middleware
			loggingHandler := middleware.Log(handler)

			// Create request with context containing request ID
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("User-Agent", tt.userAgent)

			// Add request ID to context if provided
			if tt.requestID != "" {
				ctx := context.WithValue(req.Context(), "request_id", tt.requestID)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			loggingHandler.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check that response body is preserved
			assert.Equal(t, "test response", rr.Body.String())
		})
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	// Test that responseWriter correctly captures status code
	rr := httptest.NewRecorder()
	wrapped := wrapResponseWriter(rr)

	// Initially status should be 0 (not set)
	assert.Equal(t, 0, wrapped.status)

	// Set status code
	wrapped.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, wrapped.status)

	// Check that the underlying ResponseWriter also got the status
	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestResponseWriter_Write(t *testing.T) {
	// Test that responseWriter correctly handles writes
	rr := httptest.NewRecorder()
	wrapped := wrapResponseWriter(rr)

	// Write data
	data := []byte("test data")
	n, err := wrapped.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, "test data", rr.Body.String())
}

func TestResponseWriter_DefaultStatus(t *testing.T) {
	// Test that if WriteHeader is not called explicitly,
	// the status is captured when Write is called
	rr := httptest.NewRecorder()
	wrapped := wrapResponseWriter(rr)

	// Write without explicitly setting header
	wrapped.Write([]byte("test"))

	// Status should be 200 (default)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestLoggingMiddleware_WithoutRequestID(t *testing.T) {
	cfg := config.LoggingConfig{
		Level:  "info",
		Format: "json",
		Output: "stdout",
	}
	realLogger, err := logger.New(cfg)
	assert.NoError(t, err)

	middleware := NewLoggingMiddleware(realLogger)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggingHandler := middleware.Log(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// Execute request - should not panic even without request ID
	loggingHandler.ServeHTTP(rr, req)

	// Check that request completed successfully
	assert.Equal(t, http.StatusOK, rr.Code)
}
