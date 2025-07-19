package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCORSMiddleware(t *testing.T) {
	origins := []string{"http://localhost:3000", "https://example.com"}
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	headers := []string{"Content-Type", "Authorization"}

	middleware := NewCORSMiddleware(origins, methods, headers)

	assert.NotNil(t, middleware)
	assert.Equal(t, origins, middleware.allowedOrigins)
	assert.Equal(t, methods, middleware.allowedMethods)
	assert.Equal(t, headers, middleware.allowedHeaders)
}

func TestCORSMiddleware_CORS(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		expectedOrigin string
		method         string
		expectedStatus int
	}{
		{
			name:           "allowed origin",
			allowedOrigins: []string{"http://localhost:3000", "https://example.com"},
			requestOrigin:  "http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wildcard origin",
			allowedOrigins: []string{"*"},
			requestOrigin:  "http://any-origin.com",
			expectedOrigin: "http://any-origin.com",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "disallowed origin",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://malicious.com",
			expectedOrigin: "",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "preflight request",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "http://localhost:3000",
			expectedOrigin: "http://localhost:3000",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no origin header",
			allowedOrigins: []string{"http://localhost:3000"},
			requestOrigin:  "",
			expectedOrigin: "",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			methods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
			headers := []string{"Content-Type", "Authorization"}
			middleware := NewCORSMiddleware(tt.allowedOrigins, methods, headers)

			// Create a test handler
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			// Wrap with CORS middleware
			corsHandler := middleware.CORS(handler)

			// Create request
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.requestOrigin != "" {
				req.Header.Set("Origin", tt.requestOrigin)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			corsHandler.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check CORS headers
			if tt.expectedOrigin != "" {
				assert.Equal(t, tt.expectedOrigin, rr.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
			}

			assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
			assert.Equal(t, "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
			assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
			assert.Equal(t, "86400", rr.Header().Get("Access-Control-Max-Age"))

			// For OPTIONS requests, body should be empty
			if tt.method == "OPTIONS" {
				assert.Empty(t, rr.Body.String())
			} else {
				assert.Equal(t, "OK", rr.Body.String())
			}
		})
	}
}

func TestCORSMiddleware_isOriginAllowed(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		origin         string
		expected       bool
	}{
		{
			name:           "exact match",
			allowedOrigins: []string{"http://localhost:3000", "https://example.com"},
			origin:         "http://localhost:3000",
			expected:       true,
		},
		{
			name:           "wildcard",
			allowedOrigins: []string{"*"},
			origin:         "http://any-origin.com",
			expected:       true,
		},
		{
			name:           "not allowed",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "http://malicious.com",
			expected:       false,
		},
		{
			name:           "empty origin",
			allowedOrigins: []string{"http://localhost:3000"},
			origin:         "",
			expected:       false,
		},
		{
			name:           "wildcard in list",
			allowedOrigins: []string{"http://localhost:3000", "*"},
			origin:         "http://any-origin.com",
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewCORSMiddleware(tt.allowedOrigins, []string{}, []string{})
			result := middleware.isOriginAllowed(tt.origin)
			assert.Equal(t, tt.expected, result)
		})
	}
}
