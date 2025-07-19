package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthMiddleware(t *testing.T) {
	secret := "test-secret"
	middleware := NewAuthMiddleware(secret)

	assert.NotNil(t, middleware)
	assert.Equal(t, []byte(secret), middleware.jwtSecret)
}

func TestAuthMiddleware_Authenticate(t *testing.T) {
	secret := "test-secret"
	middleware := NewAuthMiddleware(secret)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectContext  bool
	}{
		{
			name:           "valid token",
			token:          createValidToken(secret),
			expectedStatus: http.StatusOK,
			expectContext:  true,
		},
		{
			name:           "invalid token",
			token:          "invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:           "expired token",
			token:          createExpiredToken(secret),
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:           "wrong secret",
			token:          createValidToken("wrong-secret"),
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
		{
			name:           "no token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectContext:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that checks for user context
			var userFromContext interface{}
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userFromContext = r.Context().Value("user")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// Wrap with auth middleware
			authHandler := middleware.Authenticate(handler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Execute request
			authHandler.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check context
			if tt.expectContext {
				assert.NotNil(t, userFromContext)
				claims, ok := userFromContext.(jwt.MapClaims)
				assert.True(t, ok)
				assert.Equal(t, "test-user", claims["sub"])
			} else {
				assert.Nil(t, userFromContext)
			}
		})
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name          string
		authorization string
		expectedToken string
	}{
		{
			name:          "valid bearer token",
			authorization: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		},
		{
			name:          "no bearer prefix",
			authorization: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedToken: "",
		},
		{
			name:          "empty authorization",
			authorization: "",
			expectedToken: "",
		},
		{
			name:          "only bearer",
			authorization: "Bearer",
			expectedToken: "",
		},
		{
			name:          "multiple spaces - invalid format",
			authorization: "Bearer  token",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authorization != "" {
				req.Header.Set("Authorization", tt.authorization)
			}

			token := extractToken(req)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

func TestAuthMiddleware_InvalidSigningMethod(t *testing.T) {
	secret := "test-secret"
	middleware := NewAuthMiddleware(secret)

	// Use a token with wrong signing method (RS256 instead of HS256)
	// This will fail because we don't have RSA keys, but that's expected
	tokenString := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0LXVzZXIiLCJleHAiOjE2MjM5MzQ0MDB9.invalid"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	authHandler := middleware.Authenticate(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	authHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// Helper functions for creating test tokens
func createValidToken(secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func createExpiredToken(secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	})

	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}
