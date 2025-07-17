package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// InstrumentManagerService implements the InstrumentManager interface
type InstrumentManagerService struct {
	metadataStorage  interfaces.MetadataStorage
	dataPipeline     interfaces.DataPipeline
	validatorService interfaces.DataValidator
	logger           *logrus.Logger

	// In-memory cache for subscriptions
	subscriptions map[string]*entities.InstrumentSubscription
	mu            sync.RWMutex

	// Service state
	isRunning bool
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewInstrumentManagerService creates a new InstrumentManagerService
func NewInstrumentManagerService(
	metadataStorage interfaces.MetadataStorage,
	dataPipeline interfaces.DataPipeline,
	validatorService interfaces.DataValidator,
	logger *logrus.Logger,
) *InstrumentManagerService {
	if logger == nil {
		logger = logrus.New()
	}

	return &InstrumentManagerService{
		metadataStorage:  metadataStorage,
		dataPipeline:     dataPipeline,
		validatorService: validatorService,
		logger:           logger,
		subscriptions:    make(map[string]*entities.InstrumentSubscription),
	}
}

// Start starts the instrument manager service
func (ims *InstrumentManagerService) Start(ctx context.Context) error {
	ims.mu.Lock()
	defer ims.mu.Unlock()

	if ims.isRunning {
		return fmt.Errorf("instrument manager is already running")
	}

	ims.ctx, ims.cancel = context.WithCancel(ctx)
	ims.isRunning = true

	// Load existing subscriptions from storage
	if err := ims.loadSubscriptionsFromStorage(); err != nil {
		ims.logger.WithError(err).Warn("Failed to load subscriptions from storage")
	}

	ims.logger.Info("Instrument manager service started")
	return nil
}

// Stop stops the instrument manager service
func (ims *InstrumentManagerService) Stop() error {
	ims.mu.Lock()
	defer ims.mu.Unlock()

	if !ims.isRunning {
		return nil
	}

	if ims.cancel != nil {
		ims.cancel()
	}

	ims.wg.Wait()
	ims.isRunning = false

	ims.logger.Info("Instrument manager service stopped")
	return nil
}

// Health checks the health of the instrument manager
func (ims *InstrumentManagerService) Health() error {
	ims.mu.RLock()
	defer ims.mu.RUnlock()

	if !ims.isRunning {
		return fmt.Errorf("instrument manager is not running")
	}

	// Check metadata storage health (if available)
	if ims.metadataStorage != nil {
		if err := ims.metadataStorage.Health(); err != nil {
			return fmt.Errorf("metadata storage health check failed: %w", err)
		}
	}

	// Check data pipeline health (if available)
	if ims.dataPipeline != nil {
		if err := ims.dataPipeline.Health(); err != nil {
			return fmt.Errorf("data pipeline health check failed: %w", err)
		}
	}

	return nil
}

// AddInstrument adds a new instrument
func (ims *InstrumentManagerService) AddInstrument(ctx context.Context, instrument entities.InstrumentInfo) error {
	// Validate instrument
	if ims.validatorService != nil {
		if err := ims.validatorService.ValidateInstrument(instrument); err != nil {
			return fmt.Errorf("instrument validation failed: %w", err)
		}
	}

	// Save to storage (if available)
	if ims.metadataStorage != nil {
		if err := ims.metadataStorage.SaveInstrument(ctx, instrument); err != nil {
			return fmt.Errorf("failed to save instrument: %w", err)
		}
	} else {
		ims.logger.Warn("MetadataStorage not available, instrument not persisted")
	}

	ims.logger.WithFields(logrus.Fields{
		"symbol": instrument.Symbol,
		"type":   instrument.Type,
		"market": instrument.Market,
	}).Info("Instrument added successfully")

	return nil
}

// GetInstrument retrieves an instrument by symbol
func (ims *InstrumentManagerService) GetInstrument(ctx context.Context, symbol string) (*entities.InstrumentInfo, error) {
	if ims.metadataStorage == nil {
		return nil, fmt.Errorf("metadata storage not available")
	}

	instrument, err := ims.metadataStorage.GetInstrument(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get instrument: %w", err)
	}

	return instrument, nil
}

// ListInstruments retrieves all instruments
func (ims *InstrumentManagerService) ListInstruments(ctx context.Context) ([]entities.InstrumentInfo, error) {
	if ims.metadataStorage == nil {
		return []entities.InstrumentInfo{}, nil // Return empty list if storage not available
	}

	instruments, err := ims.metadataStorage.ListInstruments(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list instruments: %w", err)
	}

	return instruments, nil
}

// AddSubscription adds a new subscription
func (ims *InstrumentManagerService) AddSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	// Validate subscription
	if ims.validatorService != nil {
		if err := ims.validatorService.ValidateSubscription(subscription); err != nil {
			return fmt.Errorf("subscription validation failed: %w", err)
		}
	}

	// Check if instrument exists (if storage available)
	if ims.metadataStorage != nil {
		_, err := ims.metadataStorage.GetInstrument(ctx, subscription.Symbol)
		if err != nil {
			return fmt.Errorf("instrument not found: %w", err)
		}

		// Save to storage
		if err := ims.metadataStorage.SaveSubscription(ctx, subscription); err != nil {
			return fmt.Errorf("failed to save subscription: %w", err)
		}
	} else {
		ims.logger.Warn("MetadataStorage not available, subscription not persisted")
	}

	// Add to in-memory cache
	ims.mu.Lock()
	ims.subscriptions[subscription.ID] = &subscription
	ims.mu.Unlock()

	ims.logger.WithFields(logrus.Fields{
		"subscription_id": subscription.ID,
		"symbol":          subscription.Symbol,
		"broker_id":       subscription.BrokerID,
	}).Info("Subscription added successfully")

	return nil
}

// RemoveSubscription removes a subscription
func (ims *InstrumentManagerService) RemoveSubscription(ctx context.Context, subscriptionID string) error {
	// Get subscription first
	subscription, err := ims.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Stop tracking if active
	if subscription.IsActive {
		if err := ims.StopTracking(ctx, subscriptionID); err != nil {
			ims.logger.WithError(err).Warn("Failed to stop tracking before removing subscription")
		}
	}

	// Remove from storage (if available)
	if ims.metadataStorage != nil {
		if err := ims.metadataStorage.DeleteSubscription(ctx, subscriptionID); err != nil {
			return fmt.Errorf("failed to delete subscription: %w", err)
		}
	}

	// Remove from in-memory cache
	ims.mu.Lock()
	delete(ims.subscriptions, subscriptionID)
	ims.mu.Unlock()

	ims.logger.WithFields(logrus.Fields{
		"subscription_id": subscriptionID,
	}).Info("Subscription removed successfully")

	return nil
}

// UpdateSubscription updates an existing subscription
func (ims *InstrumentManagerService) UpdateSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	// Validate subscription
	if ims.validatorService != nil {
		if err := ims.validatorService.ValidateSubscription(subscription); err != nil {
			return fmt.Errorf("subscription validation failed: %w", err)
		}
	}

	// Check if subscription exists
	_, err := ims.GetSubscription(ctx, subscription.ID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Update in storage (if available)
	if ims.metadataStorage != nil {
		if err := ims.metadataStorage.UpdateSubscription(ctx, subscription); err != nil {
			return fmt.Errorf("failed to update subscription: %w", err)
		}
	}

	// Update in-memory cache
	ims.mu.Lock()
	ims.subscriptions[subscription.ID] = &subscription
	ims.mu.Unlock()

	ims.logger.WithFields(logrus.Fields{
		"subscription_id": subscription.ID,
	}).Info("Subscription updated successfully")

	return nil
}

// GetSubscription retrieves a subscription by ID
func (ims *InstrumentManagerService) GetSubscription(ctx context.Context, subscriptionID string) (*entities.InstrumentSubscription, error) {
	// Try in-memory cache first
	ims.mu.RLock()
	if subscription, exists := ims.subscriptions[subscriptionID]; exists {
		ims.mu.RUnlock()
		return subscription, nil
	}
	ims.mu.RUnlock()

	// Fallback to storage (if available)
	if ims.metadataStorage != nil {
		subscription, err := ims.metadataStorage.GetSubscription(ctx, subscriptionID)
		if err != nil {
			return nil, fmt.Errorf("failed to get subscription: %w", err)
		}

		// Add to cache
		ims.mu.Lock()
		ims.subscriptions[subscriptionID] = subscription
		ims.mu.Unlock()

		return subscription, nil
	}

	return nil, fmt.Errorf("subscription not found: %s", subscriptionID)
}

// ListSubscriptions retrieves all subscriptions
func (ims *InstrumentManagerService) ListSubscriptions(ctx context.Context) ([]entities.InstrumentSubscription, error) {
	if ims.metadataStorage != nil {
		subscriptions, err := ims.metadataStorage.ListSubscriptions(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list subscriptions: %w", err)
		}

		// Update in-memory cache
		ims.mu.Lock()
		for i := range subscriptions {
			ims.subscriptions[subscriptions[i].ID] = &subscriptions[i]
		}
		ims.mu.Unlock()

		return subscriptions, nil
	}

	// Return from in-memory cache if storage not available
	ims.mu.RLock()
	defer ims.mu.RUnlock()

	subscriptions := make([]entities.InstrumentSubscription, 0, len(ims.subscriptions))
	for _, subscription := range ims.subscriptions {
		subscriptions = append(subscriptions, *subscription)
	}

	return subscriptions, nil
}

// SyncWithBrokers synchronizes instruments with brokers
func (ims *InstrumentManagerService) SyncWithBrokers(ctx context.Context) error {
	// This is a placeholder implementation
	// In a real implementation, this would sync with broker APIs to get available instruments
	ims.logger.Info("Syncing with brokers (placeholder implementation)")
	return nil
}

// StartTracking starts tracking a subscription
func (ims *InstrumentManagerService) StartTracking(ctx context.Context, subscriptionID string) error {
	subscription, err := ims.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Subscribe to data pipeline (if available)
	if ims.dataPipeline != nil {
		subscriptions := []entities.InstrumentSubscription{*subscription}
		if err := ims.dataPipeline.Subscribe(ctx, subscription.BrokerID, subscriptions); err != nil {
			return fmt.Errorf("failed to start tracking: %w", err)
		}
	} else {
		ims.logger.Warn("DataPipeline not available, tracking not started")
	}

	// Update subscription status
	subscription.IsActive = true
	subscription.UpdatedAt = time.Now()
	if err := ims.UpdateSubscription(ctx, *subscription); err != nil {
		ims.logger.WithError(err).Warn("Failed to update subscription status")
	}

	ims.logger.WithFields(logrus.Fields{
		"subscription_id": subscriptionID,
		"symbol":          subscription.Symbol,
		"broker_id":       subscription.BrokerID,
	}).Info("Tracking started successfully")

	return nil
}

// StopTracking stops tracking a subscription
func (ims *InstrumentManagerService) StopTracking(ctx context.Context, subscriptionID string) error {
	subscription, err := ims.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("subscription not found: %w", err)
	}

	// Unsubscribe from data pipeline (if available)
	if ims.dataPipeline != nil {
		subscriptions := []entities.InstrumentSubscription{*subscription}
		if err := ims.dataPipeline.Unsubscribe(ctx, subscription.BrokerID, subscriptions); err != nil {
			return fmt.Errorf("failed to stop tracking: %w", err)
		}
	} else {
		ims.logger.Warn("DataPipeline not available, tracking not stopped")
	}

	// Update subscription status
	subscription.IsActive = false
	subscription.UpdatedAt = time.Now()
	if err := ims.UpdateSubscription(ctx, *subscription); err != nil {
		ims.logger.WithError(err).Warn("Failed to update subscription status")
	}

	ims.logger.WithFields(logrus.Fields{
		"subscription_id": subscriptionID,
		"symbol":          subscription.Symbol,
		"broker_id":       subscription.BrokerID,
	}).Info("Tracking stopped successfully")

	return nil
}

// loadSubscriptionsFromStorage loads subscriptions from storage into memory
func (ims *InstrumentManagerService) loadSubscriptionsFromStorage() error {
	if ims.metadataStorage == nil {
		ims.logger.Info("MetadataStorage not available, skipping subscription loading")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	subscriptions, err := ims.metadataStorage.ListSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to load subscriptions: %w", err)
	}

	ims.mu.Lock()
	defer ims.mu.Unlock()

	for i := range subscriptions {
		ims.subscriptions[subscriptions[i].ID] = &subscriptions[i]
	}

	ims.logger.WithField("count", len(subscriptions)).Info("Loaded subscriptions from storage")
	return nil
}
