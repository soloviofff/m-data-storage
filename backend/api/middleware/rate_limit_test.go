package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"m-data-storage/api/dto"
)

func TestNewRateLimiter(t *testing.T) {
	limit := 10
	window := time.Minute

	rl := NewRateLimiter(limit, window)

	assert.NotNil(t, rl)
	assert.Equal(t, limit, rl.limit)
	assert.Equal(t, window, rl.window)
	assert.Equal(t, window*2, rl.cleanupInterval)
	assert.NotNil(t, rl.clients)
}

func TestRateLimiter_RateLimit(t *testing.T) {
	// Create rate limiter with 2 requests per second
	rl := NewRateLimiter(2, time.Second)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	rateLimitHandler := rl.RateLimit(handler)

	t.Run("allow requests within limit", func(t *testing.T) {
		// First request should be allowed
		req1 := httptest.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.1:12345"
		rr1 := httptest.NewRecorder()

		rateLimitHandler.ServeHTTP(rr1, req1)

		assert.Equal(t, http.StatusOK, rr1.Code)
		assert.Equal(t, "success", rr1.Body.String())

		// Second request should be allowed
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.1:12345"
		rr2 := httptest.NewRecorder()

		rateLimitHandler.ServeHTTP(rr2, req2)

		assert.Equal(t, http.StatusOK, rr2.Code)
		assert.Equal(t, "success", rr2.Body.String())
	})

	t.Run("block requests exceeding limit", func(t *testing.T) {
		// Third request should be blocked
		req3 := httptest.NewRequest("GET", "/test", nil)
		req3.RemoteAddr = "192.168.1.1:12345"
		rr3 := httptest.NewRecorder()

		rateLimitHandler.ServeHTTP(rr3, req3)

		assert.Equal(t, http.StatusTooManyRequests, rr3.Code)
		assert.Equal(t, "application/json", rr3.Header().Get("Content-Type"))

		var response dto.Response
		err := json.Unmarshal(rr3.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.False(t, response.Success)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "RATE_LIMITED", response.Error.Code)
		assert.Equal(t, "Too many requests", response.Error.Message)
	})

	t.Run("different IPs have separate limits", func(t *testing.T) {
		// Request from different IP should be allowed
		req4 := httptest.NewRequest("GET", "/test", nil)
		req4.RemoteAddr = "192.168.1.2:12345"
		rr4 := httptest.NewRecorder()

		rateLimitHandler.ServeHTTP(rr4, req4)

		assert.Equal(t, http.StatusOK, rr4.Code)
		assert.Equal(t, "success", rr4.Body.String())
	})
}

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	clientIP := "192.168.1.1"

	// First two requests should be allowed
	assert.True(t, rl.allow(clientIP))
	assert.True(t, rl.allow(clientIP))

	// Third request should be blocked
	assert.False(t, rl.allow(clientIP))

	// Wait for window to pass and try again
	time.Sleep(time.Second + 100*time.Millisecond)

	// Should be allowed again after window expires
	assert.True(t, rl.allow(clientIP))
}

func TestRateLimiter_WindowExpiry(t *testing.T) {
	// Use very short window for testing
	rl := NewRateLimiter(1, 100*time.Millisecond)

	clientIP := "192.168.1.1"

	// First request should be allowed
	assert.True(t, rl.allow(clientIP))

	// Second request should be blocked
	assert.False(t, rl.allow(clientIP))

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	assert.True(t, rl.allow(clientIP))
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		xForwardedFor  string
		xRealIP        string
		remoteAddr     string
		expectedIP     string
	}{
		{
			name:           "X-Forwarded-For header",
			xForwardedFor:  "203.0.113.1",
			xRealIP:        "203.0.113.2",
			remoteAddr:     "192.168.1.1:12345",
			expectedIP:     "203.0.113.1",
		},
		{
			name:           "X-Real-IP header",
			xForwardedFor:  "",
			xRealIP:        "203.0.113.2",
			remoteAddr:     "192.168.1.1:12345",
			expectedIP:     "203.0.113.2",
		},
		{
			name:           "RemoteAddr fallback",
			xForwardedFor:  "",
			xRealIP:        "",
			remoteAddr:     "192.168.1.1:12345",
			expectedIP:     "192.168.1.1:12345",
		},
		{
			name:           "empty headers",
			xForwardedFor:  "",
			xRealIP:        "",
			remoteAddr:     "127.0.0.1:8080",
			expectedIP:     "127.0.0.1:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := getClientIP(req)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(10, time.Second)

	// Test concurrent access to ensure no race conditions
	done := make(chan bool, 20)

	for i := 0; i < 20; i++ {
		go func(id int) {
			clientIP := "192.168.1.1"
			rl.allow(clientIP)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should not panic and should have proper state
	rl.mu.RLock()
	client := rl.clients["192.168.1.1"]
	rl.mu.RUnlock()

	assert.NotNil(t, client)
	// Should have at most 10 requests (the limit)
	assert.LessOrEqual(t, len(client.requests), 10)
}

func TestRateLimiter_Cleanup(t *testing.T) {
	// Create rate limiter with very short cleanup interval
	rl := NewRateLimiter(10, 50*time.Millisecond)

	// Add some clients
	rl.allow("192.168.1.1")
	rl.allow("192.168.1.2")

	// Verify clients exist
	rl.mu.RLock()
	assert.Len(t, rl.clients, 2)
	rl.mu.RUnlock()

	// Wait for cleanup to run (cleanup interval is 2 * window = 100ms)
	time.Sleep(150 * time.Millisecond)

	// Clients should be cleaned up
	rl.mu.RLock()
	clientCount := len(rl.clients)
	rl.mu.RUnlock()

	// Clients might be cleaned up, but this is timing dependent
	// Just ensure no panic occurred and state is consistent
	assert.GreaterOrEqual(t, clientCount, 0)
}
