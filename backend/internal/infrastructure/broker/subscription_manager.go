package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/entities"
)

// SubscriptionStatus represents subscription status
type SubscriptionStatus string

const (
	SubscriptionStatusPending  SubscriptionStatus = "pending"
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusInactive SubscriptionStatus = "inactive"
	SubscriptionStatusError    SubscriptionStatus = "error"
)

// SubscriptionInfo contains subscription information
type SubscriptionInfo struct {
	Subscription entities.InstrumentSubscription `json:"subscription"`
	Status       SubscriptionStatus              `json:"status"`
	CreatedAt    time.Time                       `json:"created_at"`
	UpdatedAt    time.Time                       `json:"updated_at"`
	LastError    string                          `json:"last_error,omitempty"`
	DataCount    int64                           `json:"data_count"`
	LastDataAt   time.Time                       `json:"last_data_at"`
}

// SubscriptionManager manages instrument subscriptions
type SubscriptionManager struct {
	subscriptions map[string]*SubscriptionInfo
	logger        *logrus.Logger
	mu            sync.RWMutex

	// Subscription callbacks
	onSubscribe   func(ctx context.Context, subscription entities.InstrumentSubscription) error
	onUnsubscribe func(ctx context.Context, subscription entities.InstrumentSubscription) error
}

// NewSubscriptionManager creates a new subscription manager
func NewSubscriptionManager(logger *logrus.Logger) *SubscriptionManager {
	if logger == nil {
		logger = logrus.New()
	}

	return &SubscriptionManager{
		subscriptions: make(map[string]*SubscriptionInfo),
		logger:        logger,
	}
}

// SetCallbacks sets callbacks for subscriptions
func (sm *SubscriptionManager) SetCallbacks(
	onSubscribe func(ctx context.Context, subscription entities.InstrumentSubscription) error,
	onUnsubscribe func(ctx context.Context, subscription entities.InstrumentSubscription) error,
) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.onSubscribe = onSubscribe
	sm.onUnsubscribe = onUnsubscribe
}

// Subscribe subscribes to instruments
func (sm *SubscriptionManager) Subscribe(ctx context.Context, subscriptions []entities.InstrumentSubscription) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var errors []error

	for _, subscription := range subscriptions {
		key := sm.getSubscriptionKey(subscription)

		// Check if such subscription already exists
		if existingInfo, exists := sm.subscriptions[key]; exists {
			if existingInfo.Status == SubscriptionStatusActive {
				sm.logger.WithFields(logrus.Fields{
					"symbol": subscription.Symbol,
					"type":   subscription.Type,
					"market": subscription.Market,
				}).Debug("Subscription already active")
				continue
			}
		}

		// Create subscription information
		info := &SubscriptionInfo{
			Subscription: subscription,
			Status:       SubscriptionStatusPending,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		sm.subscriptions[key] = info

		// Call subscription callback
		if sm.onSubscribe != nil {
			if err := sm.onSubscribe(ctx, subscription); err != nil {
				info.Status = SubscriptionStatusError
				info.LastError = err.Error()
				info.UpdatedAt = time.Now()

				errors = append(errors, fmt.Errorf("failed to subscribe to %s: %w", key, err))

				sm.logger.WithFields(logrus.Fields{
					"symbol": subscription.Symbol,
					"type":   subscription.Type,
					"market": subscription.Market,
				}).WithError(err).Error("Subscription failed")

				continue
			}
		}

		// Mark subscription as active
		info.Status = SubscriptionStatusActive
		info.UpdatedAt = time.Now()

		sm.logger.WithFields(logrus.Fields{
			"symbol": subscription.Symbol,
			"type":   subscription.Type,
			"market": subscription.Market,
		}).Info("Subscription created successfully")
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to create %d subscriptions: %v", len(errors), errors)
	}

	return nil
}

// Unsubscribe unsubscribes from instruments
func (sm *SubscriptionManager) Unsubscribe(ctx context.Context, subscriptions []entities.InstrumentSubscription) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var errors []error

	for _, subscription := range subscriptions {
		key := sm.getSubscriptionKey(subscription)

		_, exists := sm.subscriptions[key]
		if !exists {
			sm.logger.WithFields(logrus.Fields{
				"symbol": subscription.Symbol,
				"type":   subscription.Type,
				"market": subscription.Market,
			}).Debug("Subscription not found")
			continue
		}

		// Call unsubscription callback
		if sm.onUnsubscribe != nil {
			if err := sm.onUnsubscribe(ctx, subscription); err != nil {
				errors = append(errors, fmt.Errorf("failed to unsubscribe from %s: %w", key, err))

				sm.logger.WithFields(logrus.Fields{
					"symbol": subscription.Symbol,
					"type":   subscription.Type,
					"market": subscription.Market,
				}).WithError(err).Error("Unsubscription failed")

				continue
			}
		}

		// Remove subscription
		delete(sm.subscriptions, key)

		sm.logger.WithFields(logrus.Fields{
			"symbol": subscription.Symbol,
			"type":   subscription.Type,
			"market": subscription.Market,
		}).Info("Subscription removed successfully")
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to remove %d subscriptions: %v", len(errors), errors)
	}

	return nil
}

// GetSubscription returns subscription information
func (sm *SubscriptionManager) GetSubscription(subscription entities.InstrumentSubscription) (*SubscriptionInfo, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	key := sm.getSubscriptionKey(subscription)
	info, exists := sm.subscriptions[key]
	if !exists {
		return nil, false
	}

	// Return copy for safety
	infoCopy := *info
	return &infoCopy, true
}

// GetAllSubscriptions returns all subscriptions
func (sm *SubscriptionManager) GetAllSubscriptions() map[string]*SubscriptionInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]*SubscriptionInfo)
	for key, info := range sm.subscriptions {
		infoCopy := *info
		result[key] = &infoCopy
	}

	return result
}

// GetActiveSubscriptions returns active subscriptions
func (sm *SubscriptionManager) GetActiveSubscriptions() []entities.InstrumentSubscription {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var active []entities.InstrumentSubscription
	for _, info := range sm.subscriptions {
		if info.Status == SubscriptionStatusActive {
			active = append(active, info.Subscription)
		}
	}

	return active
}

// GetSubscriptionCount returns subscription count by status
func (sm *SubscriptionManager) GetSubscriptionCount() map[SubscriptionStatus]int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	counts := make(map[SubscriptionStatus]int)
	for _, info := range sm.subscriptions {
		counts[info.Status]++
	}

	return counts
}

// UpdateDataReceived updates data reception statistics for subscription
func (sm *SubscriptionManager) UpdateDataReceived(subscription entities.InstrumentSubscription) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := sm.getSubscriptionKey(subscription)
	if info, exists := sm.subscriptions[key]; exists {
		info.DataCount++
		info.LastDataAt = time.Now()
		info.UpdatedAt = time.Now()
	}
}

// MarkSubscriptionError marks subscription as erroneous
func (sm *SubscriptionManager) MarkSubscriptionError(subscription entities.InstrumentSubscription, err error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	key := sm.getSubscriptionKey(subscription)
	if info, exists := sm.subscriptions[key]; exists {
		info.Status = SubscriptionStatusError
		info.LastError = err.Error()
		info.UpdatedAt = time.Now()

		sm.logger.WithFields(logrus.Fields{
			"symbol": subscription.Symbol,
			"type":   subscription.Type,
			"market": subscription.Market,
		}).WithError(err).Error("Subscription marked as error")
	}
}

// CleanupInactiveSubscriptions removes inactive subscriptions
func (sm *SubscriptionManager) CleanupInactiveSubscriptions(maxAge time.Duration) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	var removed []string

	for key, info := range sm.subscriptions {
		if info.Status == SubscriptionStatusInactive && info.UpdatedAt.Before(cutoff) {
			removed = append(removed, key)
		}
	}

	for _, key := range removed {
		delete(sm.subscriptions, key)
	}

	if len(removed) > 0 {
		sm.logger.WithField("count", len(removed)).Info("Cleaned up inactive subscriptions")
	}

	return len(removed)
}

// getSubscriptionKey generates key for subscription
func (sm *SubscriptionManager) getSubscriptionKey(subscription entities.InstrumentSubscription) string {
	return fmt.Sprintf("%s_%s_%s", subscription.Symbol, subscription.Type, subscription.Market)
}

// Shutdown gracefully shuts down the subscription manager
func (sm *SubscriptionManager) Shutdown(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.logger.Info("Shutting down subscription manager")

	// Unsubscribe from all active subscriptions
	var activeSubscriptions []entities.InstrumentSubscription
	for _, info := range sm.subscriptions {
		if info.Status == SubscriptionStatusActive {
			activeSubscriptions = append(activeSubscriptions, info.Subscription)
		}
	}

	if len(activeSubscriptions) > 0 {
		sm.logger.WithField("count", len(activeSubscriptions)).Info("Unsubscribing from active subscriptions")

		// Temporarily unlock mutex for Unsubscribe call
		sm.mu.Unlock()
		err := sm.Unsubscribe(ctx, activeSubscriptions)
		sm.mu.Lock()

		if err != nil {
			sm.logger.WithError(err).Error("Error unsubscribing during shutdown")
			return err
		}
	}

	sm.logger.Info("Subscription manager shutdown completed")
	return nil
}
