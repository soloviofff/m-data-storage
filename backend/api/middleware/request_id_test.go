package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRequestIDMiddleware(t *testing.T) {
	middleware := NewRequestIDMiddleware()
	assert.NotNil(t, middleware)
}

func TestRequestIDMiddleware_RequestID(t *testing.T) {
	tests := []struct {
		name              string
		existingRequestID string
		expectGenerated   bool
	}{
		{
			name:              "no existing request ID",
			existingRequestID: "",
			expectGenerated:   true,
		},
		{
			name:              "existing request ID",
			existingRequestID: "existing-request-id-123",
			expectGenerated:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewRequestIDMiddleware()

			// Create a test handler that captures the request context
			var capturedRequestID string
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedRequestID = GetRequestIDFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Wrap with RequestID middleware
			requestIDHandler := middleware.RequestID(handler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingRequestID != "" {
				req.Header.Set("X-Request-ID", tt.existingRequestID)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			requestIDHandler.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, http.StatusOK, rr.Code)

			// Check response header
			responseRequestID := rr.Header().Get("X-Request-ID")
			assert.NotEmpty(t, responseRequestID)

			// Check captured request ID from context
			assert.NotEmpty(t, capturedRequestID)
			assert.Equal(t, responseRequestID, capturedRequestID)

			if tt.expectGenerated {
				// Should be a generated ID (32 hex characters)
				assert.Len(t, responseRequestID, 32)
				assert.Regexp(t, "^[a-f0-9]{32}$", responseRequestID)
			} else {
				// Should be the existing request ID
				assert.Equal(t, tt.existingRequestID, responseRequestID)
			}
		})
	}
}

func TestGenerateRequestID(t *testing.T) {
	// Test that generateRequestID produces valid IDs
	for i := 0; i < 10; i++ {
		id := generateRequestID()
		assert.NotEmpty(t, id)
		
		// Should be 32 hex characters (16 bytes * 2)
		if id != "req-66616c6c6261636b" { // fallback case
			assert.Len(t, id, 32)
			assert.Regexp(t, "^[a-f0-9]{32}$", id)
		}
	}

	// Test that multiple calls produce different IDs
	id1 := generateRequestID()
	id2 := generateRequestID()
	if id1 != "req-66616c6c6261636b" && id2 != "req-66616c6c6261636b" {
		assert.NotEqual(t, id1, id2)
	}
}

func TestGetRequestIDFromContext(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		expected  string
	}{
		{
			name:     "context with request ID",
			ctx:      context.WithValue(context.Background(), "request_id", "test-request-id"),
			expected: "test-request-id",
		},
		{
			name:     "context without request ID",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "context with wrong type",
			ctx:      context.WithValue(context.Background(), "request_id", 123),
			expected: "",
		},
		{
			name:     "context with nil value",
			ctx:      context.WithValue(context.Background(), "request_id", nil),
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRequestIDFromContext(tt.ctx)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRequestIDMiddleware_Integration(t *testing.T) {
	middleware := NewRequestIDMiddleware()

	// Create a chain of handlers to test context propagation
	var requestIDs []string
	
	handler1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestIDs = append(requestIDs, GetRequestIDFromContext(r.Context()))
	})

	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestIDs = append(requestIDs, GetRequestIDFromContext(r.Context()))
		handler1.ServeHTTP(w, r)
	})

	// Wrap with RequestID middleware
	requestIDHandler := middleware.RequestID(handler2)

	// Create request
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// Execute request
	requestIDHandler.ServeHTTP(rr, req)

	// Check that the same request ID was available in all handlers
	assert.Len(t, requestIDs, 2)
	assert.Equal(t, requestIDs[0], requestIDs[1])
	assert.NotEmpty(t, requestIDs[0])
}
