package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/errors"
	"m-data-storage/internal/domain/interfaces"
)

// SecurityService implements security operations
type SecurityService struct {
	userStorage interfaces.UserStorage

	// Rate limiting
	rateLimiters map[string]*RateLimiter
	rateMutex    sync.RWMutex

	// Security events storage (in-memory for now)
	securityEvents []*interfaces.SecurityEvent
	eventsMutex    sync.RWMutex
}

// RateLimiter represents a simple rate limiter
type RateLimiter struct {
	requests    []time.Time
	maxRequests int
	window      time.Duration
	mutex       sync.Mutex
}

// NewSecurityService creates a new security service
func NewSecurityService(userStorage interfaces.UserStorage) interfaces.SecurityService {
	return &SecurityService{
		userStorage:    userStorage,
		rateLimiters:   make(map[string]*RateLimiter),
		securityEvents: make([]*interfaces.SecurityEvent, 0),
	}
}

// CheckRateLimit checks if a request is within rate limits
func (s *SecurityService) CheckRateLimit(ctx context.Context, identifier string, action string) (bool, error) {
	if identifier == "" || action == "" {
		return false, errors.NewAuthError(errors.CodeInvalidInput, "identifier and action cannot be empty", nil)
	}

	key := fmt.Sprintf("%s:%s", identifier, action)

	s.rateMutex.RLock()
	limiter, exists := s.rateLimiters[key]
	s.rateMutex.RUnlock()

	if !exists {
		// Create new rate limiter based on action
		maxRequests, window := s.getRateLimitConfig(action)
		limiter = &RateLimiter{
			requests:    make([]time.Time, 0),
			maxRequests: maxRequests,
			window:      window,
		}

		s.rateMutex.Lock()
		s.rateLimiters[key] = limiter
		s.rateMutex.Unlock()
	}

	return limiter.Allow(), nil
}

// IncrementRateLimit increments the rate limit counter
func (s *SecurityService) IncrementRateLimit(ctx context.Context, identifier string, action string) error {
	// Rate limiting is handled in CheckRateLimit for this implementation
	return nil
}

// LogSecurityEvent logs a security event
func (s *SecurityService) LogSecurityEvent(ctx context.Context, event interfaces.SecurityEvent) error {
	// Validate event
	if event.EventType == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "event type cannot be empty", nil)
	}

	// Set default values
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Severity == "" {
		event.Severity = s.getDefaultSeverity(event.EventType)
	}

	// Store event (in-memory for now)
	s.eventsMutex.Lock()
	s.securityEvents = append(s.securityEvents, &event)

	// Keep only last 10000 events to prevent memory issues
	if len(s.securityEvents) > 10000 {
		s.securityEvents = s.securityEvents[len(s.securityEvents)-10000:]
	}
	s.eventsMutex.Unlock()

	// TODO: In a real implementation, you would:
	// 1. Store events in a database
	// 2. Send alerts for high-severity events
	// 3. Integrate with external security monitoring systems

	return nil
}

// GetSecurityEvents retrieves security events with filtering
func (s *SecurityService) GetSecurityEvents(ctx context.Context, filter interfaces.SecurityEventFilter) ([]*interfaces.SecurityEvent, error) {
	s.eventsMutex.RLock()
	defer s.eventsMutex.RUnlock()

	var filteredEvents []*interfaces.SecurityEvent

	for _, event := range s.securityEvents {
		if s.matchesFilter(event, filter) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	// Apply limit and offset
	start := filter.Offset
	if start > len(filteredEvents) {
		start = len(filteredEvents)
	}

	end := start + filter.Limit
	if filter.Limit == 0 || end > len(filteredEvents) {
		end = len(filteredEvents)
	}

	return filteredEvents[start:end], nil
}

// CheckAccountLockout checks if an account is locked
func (s *SecurityService) CheckAccountLockout(ctx context.Context, userID string) (bool, time.Time, error) {
	if userID == "" {
		return false, time.Time{}, errors.NewAuthError(errors.CodeInvalidInput, "user ID cannot be empty", nil)
	}

	// Get user
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return false, time.Time{}, errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}

	// Check if user is suspended (which we treat as locked)
	if user.Status == entities.UserStatusSuspended {
		// TODO: Get actual lockout expiration from database
		// For now, return a default lockout time
		lockoutExpiration := time.Now().Add(24 * time.Hour)
		return true, lockoutExpiration, nil
	}

	return false, time.Time{}, nil
}

// LockAccount locks a user account for a specified duration
func (s *SecurityService) LockAccount(ctx context.Context, userID string, duration time.Duration) error {
	if userID == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "user ID cannot be empty", nil)
	}

	// Get user
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}

	// Update user status to suspended
	suspendedStatus := entities.UserStatusSuspended
	user.Status = suspendedStatus

	err = s.userStorage.UpdateUser(ctx, user)
	if err != nil {
		return errors.NewAuthError("INTERNAL_ERROR", "failed to lock account", err)
	}

	// Log security event
	event := interfaces.SecurityEvent{
		UserID:      userID,
		EventType:   interfaces.SecurityEventAccountLocked,
		Description: fmt.Sprintf("Account locked for %v", duration),
		Severity:    interfaces.SecuritySeverityHigh,
		Metadata: map[string]interface{}{
			"duration_minutes": duration.Minutes(),
		},
	}
	s.LogSecurityEvent(ctx, event)

	// TODO: Store lockout expiration in database
	// TODO: Set up automatic unlock after duration

	return nil
}

// UnlockAccount unlocks a user account
func (s *SecurityService) UnlockAccount(ctx context.Context, userID string) error {
	if userID == "" {
		return errors.NewAuthError(errors.CodeInvalidInput, "user ID cannot be empty", nil)
	}

	// Get user
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}

	// Update user status to active
	activeStatus := entities.UserStatusActive
	user.Status = activeStatus

	err = s.userStorage.UpdateUser(ctx, user)
	if err != nil {
		return errors.NewAuthError("INTERNAL_ERROR", "failed to unlock account", err)
	}

	// Log security event
	event := interfaces.SecurityEvent{
		UserID:      userID,
		EventType:   interfaces.SecurityEventAccountUnlocked,
		Description: "Account unlocked",
		Severity:    interfaces.SecuritySeverityMedium,
	}
	s.LogSecurityEvent(ctx, event)

	return nil
}

// ValidateSessionSecurity validates session security parameters
func (s *SecurityService) ValidateSessionSecurity(ctx context.Context, session *entities.UserSession, ipAddress, userAgent string) (bool, error) {
	if session == nil {
		return false, errors.NewAuthError(errors.CodeInvalidInput, "session cannot be nil", nil)
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now()) {
		return false, errors.NewAuthError(errors.CodeTokenExpired, "session is expired", nil)
	}

	// Check if session is active
	if session.IsRevoked {
		return false, errors.NewAuthError("SESSION_INVALID", "session is revoked", nil)
	}

	// Validate IP address (if IP binding is enabled)
	if session.IPAddress != "" && session.IPAddress != ipAddress {
		// Log suspicious activity
		event := interfaces.SecurityEvent{
			UserID:      session.UserID,
			EventType:   interfaces.SecurityEventSuspiciousActivity,
			Description: fmt.Sprintf("Session accessed from different IP: expected %s, got %s", session.IPAddress, ipAddress),
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Severity:    interfaces.SecuritySeverityHigh,
		}
		s.LogSecurityEvent(ctx, event)

		return false, errors.NewAuthError("SUSPICIOUS_ACTIVITY", "session accessed from different IP address", nil)
	}

	// Validate user agent (basic check)
	if session.UserAgent != "" && !s.isUserAgentSimilar(session.UserAgent, userAgent) {
		// Log suspicious activity but don't fail (user agents can change)
		event := interfaces.SecurityEvent{
			UserID:      session.UserID,
			EventType:   interfaces.SecurityEventSuspiciousActivity,
			Description: fmt.Sprintf("Session accessed with different user agent"),
			IPAddress:   ipAddress,
			UserAgent:   userAgent,
			Severity:    interfaces.SecuritySeverityMedium,
		}
		s.LogSecurityEvent(ctx, event)
	}

	return true, nil
}

// DetectSuspiciousActivity detects suspicious activity for a user
func (s *SecurityService) DetectSuspiciousActivity(ctx context.Context, userID string) (bool, []string, error) {
	if userID == "" {
		return false, nil, errors.NewAuthError(errors.CodeInvalidInput, "user ID cannot be empty", nil)
	}

	var suspiciousActivities []string

	// Get recent security events for the user
	startTime := time.Now().Add(-24 * time.Hour)
	filter := interfaces.SecurityEventFilter{
		UserID:    &userID,
		StartTime: &startTime, // Last 24 hours
		Limit:     100,
	}

	events, err := s.GetSecurityEvents(ctx, filter)
	if err != nil {
		return false, nil, err
	}

	// Analyze events for suspicious patterns
	failedLogins := 0
	differentIPs := make(map[string]bool)

	for _, event := range events {
		switch event.EventType {
		case interfaces.SecurityEventLoginFailed:
			failedLogins++
		case interfaces.SecurityEventLogin:
			if event.IPAddress != "" {
				differentIPs[event.IPAddress] = true
			}
		}
	}

	// Check for multiple failed login attempts
	if failedLogins >= 5 {
		suspiciousActivities = append(suspiciousActivities, fmt.Sprintf("Multiple failed login attempts: %d", failedLogins))
	}

	// Check for logins from multiple IP addresses
	if len(differentIPs) >= 3 {
		suspiciousActivities = append(suspiciousActivities, fmt.Sprintf("Logins from multiple IP addresses: %d", len(differentIPs)))
	}

	return len(suspiciousActivities) > 0, suspiciousActivities, nil
}

// Helper methods

// Allow checks if a request is allowed by the rate limiter
func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Remove old requests outside the window
	cutoff := now.Add(-rl.window)
	validRequests := make([]time.Time, 0)
	for _, req := range rl.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	rl.requests = validRequests

	// Check if we can add a new request
	if len(rl.requests) >= rl.maxRequests {
		return false
	}

	// Add the current request
	rl.requests = append(rl.requests, now)
	return true
}

// getRateLimitConfig returns rate limit configuration for different actions
func (s *SecurityService) getRateLimitConfig(action string) (int, time.Duration) {
	switch action {
	case "login":
		return 5, 15 * time.Minute // 5 attempts per 15 minutes
	case "password_reset":
		return 3, 1 * time.Hour // 3 attempts per hour
	case "api_request":
		return 1000, 1 * time.Hour // 1000 requests per hour
	default:
		return 100, 1 * time.Hour // Default: 100 requests per hour
	}
}

// getDefaultSeverity returns default severity for event types
func (s *SecurityService) getDefaultSeverity(eventType interfaces.SecurityEventType) interfaces.SecuritySeverity {
	switch eventType {
	case interfaces.SecurityEventLoginFailed, interfaces.SecurityEventPermissionDenied:
		return interfaces.SecuritySeverityMedium
	case interfaces.SecurityEventAccountLocked, interfaces.SecurityEventSuspiciousActivity:
		return interfaces.SecuritySeverityHigh
	case interfaces.SecurityEventAPIKeyRevoked:
		return interfaces.SecuritySeverityCritical
	default:
		return interfaces.SecuritySeverityLow
	}
}

// matchesFilter checks if an event matches the filter criteria
func (s *SecurityService) matchesFilter(event *interfaces.SecurityEvent, filter interfaces.SecurityEventFilter) bool {
	if filter.UserID != nil && event.UserID != *filter.UserID {
		return false
	}
	if filter.EventType != nil && event.EventType != *filter.EventType {
		return false
	}
	if filter.Severity != nil && event.Severity != *filter.Severity {
		return false
	}
	if filter.StartTime != nil && event.Timestamp.Before(*filter.StartTime) {
		return false
	}
	if filter.EndTime != nil && event.Timestamp.After(*filter.EndTime) {
		return false
	}
	return true
}

// isUserAgentSimilar checks if two user agents are similar (basic implementation)
func (s *SecurityService) isUserAgentSimilar(ua1, ua2 string) bool {
	if ua1 == ua2 {
		return true
	}

	// Extract browser and OS information (very basic)
	ua1Lower := strings.ToLower(ua1)
	ua2Lower := strings.ToLower(ua2)

	// Check if they contain the same browser
	browsers := []string{"chrome", "firefox", "safari", "edge", "opera"}
	for _, browser := range browsers {
		if strings.Contains(ua1Lower, browser) && strings.Contains(ua2Lower, browser) {
			return true
		}
	}

	return false
}
