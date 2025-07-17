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

// DataPipelineService manages data flows from brokers to storage
type DataPipelineService struct {
	brokerManager      interfaces.BrokerManager
	storageIntegration interfaces.BrokerStorageIntegration
	logger             *logrus.Logger

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Configuration
	config DataPipelineConfig

	// Statistics
	stats PipelineStats
	mu    sync.RWMutex
}

// DataPipelineConfig contains data pipeline configuration
type DataPipelineConfig struct {
	// Automatic broker connection
	AutoConnectBrokers bool `yaml:"auto_connect_brokers" json:"auto_connect_brokers"`

	// Health check interval
	HealthCheckInterval time.Duration `yaml:"health_check_interval" json:"health_check_interval"`

	// Operation timeout
	OperationTimeout time.Duration `yaml:"operation_timeout" json:"operation_timeout"`

	// Automatic reconnection
	AutoReconnect bool `yaml:"auto_reconnect" json:"auto_reconnect"`

	// Reconnection interval
	ReconnectInterval time.Duration `yaml:"reconnect_interval" json:"reconnect_interval"`
}

// PipelineStats contains pipeline statistics
type PipelineStats = interfaces.DataPipelineStats

// DefaultDataPipelineConfig returns default configuration
func DefaultDataPipelineConfig() DataPipelineConfig {
	return DataPipelineConfig{
		AutoConnectBrokers:  true,
		HealthCheckInterval: 30 * time.Second,
		OperationTimeout:    10 * time.Second,
		AutoReconnect:       true,
		ReconnectInterval:   5 * time.Second,
	}
}

// NewDataPipelineService creates a new data pipeline service
func NewDataPipelineService(
	brokerManager interfaces.BrokerManager,
	storageIntegration interfaces.BrokerStorageIntegration,
	logger *logrus.Logger,
	config DataPipelineConfig,
) *DataPipelineService {
	if logger == nil {
		logger = logrus.New()
	}

	return &DataPipelineService{
		brokerManager:      brokerManager,
		storageIntegration: storageIntegration,
		logger:             logger,
		config:             config,
		stats: PipelineStats{
			StartedAt: time.Now(),
		},
	}
}

// Start starts the data pipeline
func (dps *DataPipelineService) Start(ctx context.Context) error {
	dps.ctx, dps.cancel = context.WithCancel(ctx)

	dps.logger.Info("Starting data pipeline service")

	// Start storage integration
	if err := dps.storageIntegration.Start(dps.ctx); err != nil {
		return fmt.Errorf("failed to start storage integration: %w", err)
	}

	// Connect existing brokers if auto-connect is enabled
	if dps.config.AutoConnectBrokers {
		if err := dps.connectAllBrokers(); err != nil {
			dps.logger.WithError(err).Warn("Failed to connect some brokers during startup")
		}
	}

	// Start background tasks
	dps.wg.Add(1)
	go dps.healthCheckWorker()

	if dps.config.AutoReconnect {
		dps.wg.Add(1)
		go dps.reconnectWorker()
	}

	dps.updateStats(func(stats *PipelineStats) {
		stats.StartedAt = time.Now()
	})

	dps.logger.Info("Data pipeline service started successfully")
	return nil
}

// Stop stops the data pipeline
func (dps *DataPipelineService) Stop() error {
	dps.logger.Info("Stopping data pipeline service")

	if dps.cancel != nil {
		dps.cancel()
	}

	// Stop storage integration
	if err := dps.storageIntegration.Stop(); err != nil {
		dps.logger.WithError(err).Error("Failed to stop storage integration")
	}

	// Wait for background tasks to finish
	dps.wg.Wait()

	dps.logger.Info("Data pipeline service stopped")
	return nil
}

// AddBroker adds a broker to the pipeline
func (dps *DataPipelineService) AddBroker(ctx context.Context, config interfaces.BrokerConfig) error {
	// Add broker to manager
	if err := dps.brokerManager.AddBroker(ctx, config); err != nil {
		return fmt.Errorf("failed to add broker to manager: %w", err)
	}

	// Get broker from manager
	broker, err := dps.brokerManager.GetBroker(config.ID)
	if err != nil {
		return fmt.Errorf("failed to get broker from manager: %w", err)
	}

	// Connect broker
	if err := dps.connectBroker(ctx, config.ID, broker); err != nil {
		dps.logger.WithError(err).WithField("broker_id", config.ID).Error("Failed to connect broker")
		return fmt.Errorf("failed to connect broker: %w", err)
	}

	dps.logger.WithField("broker_id", config.ID).Info("Broker added to pipeline")
	return nil
}

// RemoveBroker removes a broker from the pipeline
func (dps *DataPipelineService) RemoveBroker(ctx context.Context, brokerID string) error {
	// Remove broker from storage integration
	if err := dps.storageIntegration.RemoveBroker(brokerID); err != nil {
		dps.logger.WithError(err).WithField("broker_id", brokerID).Warn("Failed to remove broker from storage integration")
	}

	// Remove broker from manager
	if err := dps.brokerManager.RemoveBroker(ctx, brokerID); err != nil {
		return fmt.Errorf("failed to remove broker from manager: %w", err)
	}

	// Update statistics
	dps.updateStats(func(stats *PipelineStats) {
		if stats.ConnectedBrokers > 0 {
			stats.ConnectedBrokers--
		}
	})

	dps.logger.WithField("broker_id", brokerID).Info("Broker removed from pipeline")
	return nil
}

// Subscribe subscribes to broker instruments
func (dps *DataPipelineService) Subscribe(ctx context.Context, brokerID string, subscriptions []entities.InstrumentSubscription) error {
	broker, err := dps.brokerManager.GetBroker(brokerID)
	if err != nil {
		return fmt.Errorf("failed to get broker: %w", err)
	}

	if err := broker.Subscribe(ctx, subscriptions); err != nil {
		return fmt.Errorf("failed to subscribe to instruments: %w", err)
	}

	dps.logger.WithFields(logrus.Fields{
		"broker_id":     brokerID,
		"subscriptions": len(subscriptions),
	}).Info("Subscribed to instruments")

	return nil
}

// Unsubscribe unsubscribes from broker instruments
func (dps *DataPipelineService) Unsubscribe(ctx context.Context, brokerID string, subscriptions []entities.InstrumentSubscription) error {
	broker, err := dps.brokerManager.GetBroker(brokerID)
	if err != nil {
		return fmt.Errorf("failed to get broker: %w", err)
	}

	if err := broker.Unsubscribe(ctx, subscriptions); err != nil {
		return fmt.Errorf("failed to unsubscribe from instruments: %w", err)
	}

	dps.logger.WithFields(logrus.Fields{
		"broker_id":     brokerID,
		"subscriptions": len(subscriptions),
	}).Info("Unsubscribed from instruments")

	return nil
}

// GetStats returns pipeline statistics
func (dps *DataPipelineService) GetStats() PipelineStats {
	dps.mu.RLock()
	defer dps.mu.RUnlock()
	return dps.stats
}

// GetIntegrationStats returns storage integration statistics
func (dps *DataPipelineService) GetIntegrationStats() interfaces.BrokerStorageIntegrationStats {
	return dps.storageIntegration.GetStats()
}

// Health checks pipeline health
func (dps *DataPipelineService) Health() error {
	// Check storage integration
	if err := dps.storageIntegration.Health(); err != nil {
		return fmt.Errorf("storage integration health check failed: %w", err)
	}

	// Check brokers
	brokerHealth := dps.brokerManager.Health()
	for brokerID, err := range brokerHealth {
		if err != nil {
			return fmt.Errorf("broker %s health check failed: %w", brokerID, err)
		}
	}

	return nil
}

// connectAllBrokers connects all brokers from the manager
func (dps *DataPipelineService) connectAllBrokers() error {
	brokers := dps.brokerManager.GetAllBrokers()

	var errors []error
	for brokerID, broker := range brokers {
		ctx, cancel := context.WithTimeout(dps.ctx, dps.config.OperationTimeout)
		if err := dps.connectBroker(ctx, brokerID, broker); err != nil {
			errors = append(errors, fmt.Errorf("broker %s: %w", brokerID, err))
		}
		cancel()
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to connect %d brokers: %v", len(errors), errors)
	}

	return nil
}

// connectBroker connects an individual broker
func (dps *DataPipelineService) connectBroker(ctx context.Context, brokerID string, broker interfaces.Broker) error {
	// Connect broker
	if err := broker.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect broker: %w", err)
	}

	// Start broker
	if err := broker.Start(ctx); err != nil {
		return fmt.Errorf("failed to start broker: %w", err)
	}

	// Add broker to storage integration
	if err := dps.storageIntegration.AddBroker(brokerID, broker); err != nil {
		return fmt.Errorf("failed to add broker to storage integration: %w", err)
	}

	dps.updateStats(func(stats *PipelineStats) {
		stats.ConnectedBrokers++
	})

	dps.logger.WithField("broker_id", brokerID).Info("Broker connected successfully")
	return nil
}

// healthCheckWorker performs periodic health checks
func (dps *DataPipelineService) healthCheckWorker() {
	defer dps.wg.Done()

	ticker := time.NewTicker(dps.config.HealthCheckInterval)
	defer ticker.Stop()

	dps.logger.Debug("Health check worker started")

	for {
		select {
		case <-dps.ctx.Done():
			dps.logger.Debug("Health check worker stopped")
			return

		case <-ticker.C:
			dps.performHealthCheck()
		}
	}
}

// reconnectWorker performs reconnection of disconnected brokers
func (dps *DataPipelineService) reconnectWorker() {
	defer dps.wg.Done()

	ticker := time.NewTicker(dps.config.ReconnectInterval)
	defer ticker.Stop()

	dps.logger.Debug("Reconnect worker started")

	for {
		select {
		case <-dps.ctx.Done():
			dps.logger.Debug("Reconnect worker stopped")
			return

		case <-ticker.C:
			dps.performReconnect()
		}
	}
}

// performHealthCheck performs a health check
func (dps *DataPipelineService) performHealthCheck() {
	dps.logger.Debug("Performing health check")

	err := dps.Health()

	dps.updateStats(func(stats *PipelineStats) {
		stats.LastHealthCheck = time.Now()
		if err != nil {
			stats.HealthChecksFailed++
		} else {
			stats.HealthChecksPassed++
		}
	})

	if err != nil {
		dps.logger.WithError(err).Warn("Health check failed")
	} else {
		dps.logger.Debug("Health check passed")
	}
}

// performReconnect attempts to reconnect disconnected brokers
func (dps *DataPipelineService) performReconnect() {
	brokers := dps.brokerManager.GetAllBrokers()

	for brokerID, broker := range brokers {
		if !broker.IsConnected() {
			dps.logger.WithField("broker_id", brokerID).Info("Attempting to reconnect broker")

			ctx, cancel := context.WithTimeout(dps.ctx, dps.config.OperationTimeout)
			if err := dps.connectBroker(ctx, brokerID, broker); err != nil {
				dps.logger.WithError(err).WithField("broker_id", brokerID).Error("Failed to reconnect broker")
				dps.updateStats(func(stats *PipelineStats) {
					stats.TotalErrors++
				})
			}
			cancel()
		}
	}
}

// updateStats updates pipeline statistics
func (dps *DataPipelineService) updateStats(updateFunc func(*PipelineStats)) {
	dps.mu.Lock()
	defer dps.mu.Unlock()
	updateFunc(&dps.stats)
}
