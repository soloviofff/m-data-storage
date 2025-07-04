package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"m-data-storage/api/dto"
)

// RateLimiter represents a simple rate limiter
type RateLimiter struct {
	mu              sync.RWMutex
	clients         map[string]*clientInfo
	limit           int           // requests per window
	window          time.Duration // time window
	cleanupInterval time.Duration
}

// clientInfo stores information about a client's requests
type clientInfo struct {
	requests []time.Time
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients:         make(map[string]*clientInfo),
		limit:           limit,
		window:          window,
		cleanupInterval: window * 2,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// RateLimit middleware function
func (rl *RateLimiter) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		if !rl.allow(clientIP) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)

			response := dto.Response{
				Success: false,
				Error: &dto.Error{
					Code:    "RATE_LIMITED",
					Message: "Too many requests",
					Details: "Rate limit exceeded. Please try again later.",
				},
			}

			json.NewEncoder(w).Encode(response)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// allow checks if a request from the given client should be allowed
func (rl *RateLimiter) allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientIP]

	if !exists {
		client = &clientInfo{
			requests: make([]time.Time, 0),
			lastSeen: now,
		}
		rl.clients[clientIP] = client
	}

	// Remove old requests outside the window
	cutoff := now.Add(-rl.window)
	validRequests := make([]time.Time, 0)

	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	client.requests = validRequests
	client.lastSeen = now

	// Check if limit is exceeded
	if len(client.requests) >= rl.limit {
		return false
	}

	// Add current request
	client.requests = append(client.requests, now)
	return true
}

// cleanup removes old client entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.cleanupInterval)

		for clientIP, client := range rl.clients {
			if client.lastSeen.Before(cutoff) {
				delete(rl.clients, clientIP)
			}
		}
		rl.mu.Unlock()
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
