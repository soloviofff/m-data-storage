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

// BrokerStorageIntegration integrates brokers with storage system
type BrokerStorageIntegration struct {
	brokerManager  interfaces.BrokerManager
	storageService interfaces.StorageService
	logger         *logrus.Logger

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Active integrations by broker
	integrations map[string]*BrokerIntegration
	mu           sync.RWMutex

	// Statistics
	stats   IntegrationStats
	statsMu sync.RWMutex
}

// BrokerIntegration represents integration of one broker with storage
type BrokerIntegration struct {
	brokerID string
	broker   interfaces.Broker

	// Channels for stopping workers
	stopChan chan struct{}
	wg       sync.WaitGroup

	// Broker statistics
	stats BrokerIntegrationStats
	mu    sync.RWMutex
}

// IntegrationStats contains overall integration statistics
type IntegrationStats = interfaces.BrokerStorageIntegrationStats

// BrokerIntegrationStats contains broker integration statistics
type BrokerIntegrationStats = interfaces.BrokerIntegrationStats

// NewBrokerStorageIntegration creates a new integration service
func NewBrokerStorageIntegration(
	brokerManager interfaces.BrokerManager,
	storageService interfaces.StorageService,
	logger *logrus.Logger,
) *BrokerStorageIntegration {
	if logger == nil {
		logger = logrus.New()
	}

	return &BrokerStorageIntegration{
		brokerManager:  brokerManager,
		storageService: storageService,
		logger:         logger,
		integrations:   make(map[string]*BrokerIntegration),
		stats: IntegrationStats{
			StartedAt: time.Now(),
		},
	}
}

// NewBrokerStorageIntegrationService creates a new integration service (returns interface)
func NewBrokerStorageIntegrationService(
	brokerManager interfaces.BrokerManager,
	storageService interfaces.StorageService,
	logger *logrus.Logger,
) interfaces.BrokerStorageIntegration {
	return NewBrokerStorageIntegration(brokerManager, storageService, logger)
}

// Start starts integration of all brokers with storage
func (bis *BrokerStorageIntegration) Start(ctx context.Context) error {
	bis.ctx, bis.cancel = context.WithCancel(ctx)

	bis.logger.Info("Starting broker-storage integration")

	// Get all brokers
	brokers := bis.brokerManager.GetAllBrokers()

	bis.mu.Lock()
	defer bis.mu.Unlock()

	// Start integration for each broker
	for brokerID, broker := range brokers {
		if err := bis.startBrokerIntegrationUnsafe(brokerID, broker); err != nil {
			bis.logger.WithError(err).WithField("broker_id", brokerID).Error("Failed to start broker integration")
			continue
		}
	}

	bis.updateStatsUnsafe()
	bis.logger.WithField("active_brokers", len(bis.integrations)).Info("Broker-storage integration started")

	return nil
}

// Stop stops integration
func (bis *BrokerStorageIntegration) Stop() error {
	bis.logger.Info("Stopping broker-storage integration")

	if bis.cancel != nil {
		bis.cancel()
	}

	bis.mu.Lock()
	defer bis.mu.Unlock()

	// Stop all broker integrations
	for brokerID, integration := range bis.integrations {
		bis.stopBrokerIntegrationUnsafe(brokerID, integration)
	}

	// Wait for all workers to complete
	bis.wg.Wait()

	bis.logger.Info("Broker-storage integration stopped")
	return nil
}

// AddBroker adds a new broker to integration
func (bis *BrokerStorageIntegration) AddBroker(brokerID string, broker interfaces.Broker) error {
	bis.mu.Lock()
	defer bis.mu.Unlock()

	if _, exists := bis.integrations[brokerID]; exists {
		return fmt.Errorf("broker %s already integrated", brokerID)
	}

	if err := bis.startBrokerIntegrationUnsafe(brokerID, broker); err != nil {
		return fmt.Errorf("failed to start broker integration: %w", err)
	}

	bis.updateStatsUnsafe()
	bis.logger.WithField("broker_id", brokerID).Info("Broker added to integration")

	return nil
}

// RemoveBroker removes a broker from integration
func (bis *BrokerStorageIntegration) RemoveBroker(brokerID string) error {
	bis.mu.Lock()
	defer bis.mu.Unlock()

	integration, exists := bis.integrations[brokerID]
	if !exists {
		return fmt.Errorf("broker %s not found in integration", brokerID)
	}

	bis.stopBrokerIntegrationUnsafe(brokerID, integration)
	delete(bis.integrations, brokerID)

	bis.updateStatsUnsafe()
	bis.logger.WithField("broker_id", brokerID).Info("Broker removed from integration")

	return nil
}

// GetStats returns integration statistics
func (bis *BrokerStorageIntegration) GetStats() IntegrationStats {
	bis.statsMu.RLock()
	defer bis.statsMu.RUnlock()
	return bis.stats
}

// GetBrokerStats returns statistics for a specific broker
func (bis *BrokerStorageIntegration) GetBrokerStats(brokerID string) (BrokerIntegrationStats, error) {
	bis.mu.RLock()
	defer bis.mu.RUnlock()

	integration, exists := bis.integrations[brokerID]
	if !exists {
		return BrokerIntegrationStats{}, fmt.Errorf("broker %s not found", brokerID)
	}

	integration.mu.RLock()
	defer integration.mu.RUnlock()
	return integration.stats, nil
}

// Health checks integration health
func (bis *BrokerStorageIntegration) Health() error {
	bis.mu.RLock()
	defer bis.mu.RUnlock()

	if len(bis.integrations) == 0 {
		return fmt.Errorf("no active broker integrations")
	}

	// Check that data is being received
	bis.statsMu.RLock()
	lastData := bis.stats.LastDataReceived
	bis.statsMu.RUnlock()

	if !lastData.IsZero() && time.Since(lastData) > 5*time.Minute {
		return fmt.Errorf("no data received for %v", time.Since(lastData))
	}

	return nil
}

// startBrokerIntegrationUnsafe starts integration for broker (without locking)
func (bis *BrokerStorageIntegration) startBrokerIntegrationUnsafe(brokerID string, broker interfaces.Broker) error {
	integration := &BrokerIntegration{
		brokerID: brokerID,
		broker:   broker,
		stopChan: make(chan struct{}),
		stats: BrokerIntegrationStats{
			BrokerID:  brokerID,
			StartedAt: time.Now(),
		},
	}

	// Start workers for data processing
	integration.wg.Add(3)
	go bis.tickerWorker(integration)
	go bis.candleWorker(integration)
	go bis.orderBookWorker(integration)

	bis.integrations[brokerID] = integration

	bis.logger.WithField("broker_id", brokerID).Info("Broker integration started")
	return nil
}

// stopBrokerIntegrationUnsafe stops broker integration (without locking)
func (bis *BrokerStorageIntegration) stopBrokerIntegrationUnsafe(brokerID string, integration *BrokerIntegration) {
	close(integration.stopChan)
	integration.wg.Wait()
	bis.logger.WithField("broker_id", brokerID).Info("Broker integration stopped")
}

// updateStatsUnsafe updates overall statistics (without locking)
func (bis *BrokerStorageIntegration) updateStatsUnsafe() {
	bis.statsMu.Lock()
	defer bis.statsMu.Unlock()

	bis.stats.ActiveBrokers = len(bis.integrations)
}

// tickerWorker processes tickers from broker
func (bis *BrokerStorageIntegration) tickerWorker(integration *BrokerIntegration) {
	defer integration.wg.Done()

	tickerChan := integration.broker.GetTickerChannel()

	bis.logger.WithField("broker_id", integration.brokerID).Debug("Ticker worker started")

	for {
		select {
		case <-integration.stopChan:
			bis.logger.WithField("broker_id", integration.brokerID).Debug("Ticker worker stopped")
			return

		case <-bis.ctx.Done():
			bis.logger.WithField("broker_id", integration.brokerID).Debug("Ticker worker stopped by context")
			return

		case ticker, ok := <-tickerChan:
			if !ok {
				bis.logger.WithField("broker_id", integration.brokerID).Warn("Ticker channel closed")
				return
			}

			if err := bis.processTicker(integration, ticker); err != nil {
				bis.logger.WithError(err).WithFields(logrus.Fields{
					"broker_id": integration.brokerID,
					"symbol":    ticker.Symbol,
				}).Error("Failed to process ticker")

				integration.mu.Lock()
				integration.stats.Errors++
				integration.mu.Unlock()

				bis.statsMu.Lock()
				bis.stats.TotalErrors++
				bis.statsMu.Unlock()
			}
		}
	}
}

// candleWorker processes candles from broker
func (bis *BrokerStorageIntegration) candleWorker(integration *BrokerIntegration) {
	defer integration.wg.Done()

	candleChan := integration.broker.GetCandleChannel()

	bis.logger.WithField("broker_id", integration.brokerID).Debug("Candle worker started")

	for {
		select {
		case <-integration.stopChan:
			bis.logger.WithField("broker_id", integration.brokerID).Debug("Candle worker stopped")
			return

		case <-bis.ctx.Done():
			bis.logger.WithField("broker_id", integration.brokerID).Debug("Candle worker stopped by context")
			return

		case candle, ok := <-candleChan:
			if !ok {
				bis.logger.WithField("broker_id", integration.brokerID).Warn("Candle channel closed")
				return
			}

			if err := bis.processCandle(integration, candle); err != nil {
				bis.logger.WithError(err).WithFields(logrus.Fields{
					"broker_id": integration.brokerID,
					"symbol":    candle.Symbol,
				}).Error("Failed to process candle")

				integration.mu.Lock()
				integration.stats.Errors++
				integration.mu.Unlock()

				bis.statsMu.Lock()
				bis.stats.TotalErrors++
				bis.statsMu.Unlock()
			}
		}
	}
}

// orderBookWorker processes order books from broker
func (bis *BrokerStorageIntegration) orderBookWorker(integration *BrokerIntegration) {
	defer integration.wg.Done()

	orderBookChan := integration.broker.GetOrderBookChannel()

	bis.logger.WithField("broker_id", integration.brokerID).Debug("OrderBook worker started")

	for {
		select {
		case <-integration.stopChan:
			bis.logger.WithField("broker_id", integration.brokerID).Debug("OrderBook worker stopped")
			return

		case <-bis.ctx.Done():
			bis.logger.WithField("broker_id", integration.brokerID).Debug("OrderBook worker stopped by context")
			return

		case orderBook, ok := <-orderBookChan:
			if !ok {
				bis.logger.WithField("broker_id", integration.brokerID).Warn("OrderBook channel closed")
				return
			}

			if err := bis.processOrderBook(integration, orderBook); err != nil {
				bis.logger.WithError(err).WithFields(logrus.Fields{
					"broker_id": integration.brokerID,
					"symbol":    orderBook.Symbol,
				}).Error("Failed to process orderbook")

				integration.mu.Lock()
				integration.stats.Errors++
				integration.mu.Unlock()

				bis.statsMu.Lock()
				bis.stats.TotalErrors++
				bis.statsMu.Unlock()
			}
		}
	}
}

// processTicker processes ticker and saves it to storage
func (bis *BrokerStorageIntegration) processTicker(integration *BrokerIntegration, ticker entities.Ticker) error {
	// Add broker information to ticker
	ticker.BrokerID = integration.brokerID

	// Save ticker through StorageService
	ctx, cancel := context.WithTimeout(bis.ctx, 10*time.Second)
	defer cancel()

	if err := bis.storageService.SaveTicker(ctx, ticker); err != nil {
		return fmt.Errorf("failed to save ticker: %w", err)
	}

	// Update statistics
	now := time.Now()

	integration.mu.Lock()
	integration.stats.TickersProcessed++
	integration.stats.LastDataReceived = now
	integration.mu.Unlock()

	bis.statsMu.Lock()
	bis.stats.TotalTickers++
	bis.stats.LastDataReceived = now
	bis.statsMu.Unlock()

	bis.logger.WithFields(logrus.Fields{
		"broker_id": integration.brokerID,
		"symbol":    ticker.Symbol,
		"price":     ticker.Price,
	}).Debug("Ticker processed and saved")

	return nil
}

// processCandle processes candle and saves it to storage
func (bis *BrokerStorageIntegration) processCandle(integration *BrokerIntegration, candle entities.Candle) error {
	// Add broker information to candle
	candle.BrokerID = integration.brokerID

	// Save candle through StorageService
	ctx, cancel := context.WithTimeout(bis.ctx, 10*time.Second)
	defer cancel()

	if err := bis.storageService.SaveCandle(ctx, candle); err != nil {
		return fmt.Errorf("failed to save candle: %w", err)
	}

	// Update statistics
	now := time.Now()

	integration.mu.Lock()
	integration.stats.CandlesProcessed++
	integration.stats.LastDataReceived = now
	integration.mu.Unlock()

	bis.statsMu.Lock()
	bis.stats.TotalCandles++
	bis.stats.LastDataReceived = now
	bis.statsMu.Unlock()

	bis.logger.WithFields(logrus.Fields{
		"broker_id": integration.brokerID,
		"symbol":    candle.Symbol,
		"timeframe": candle.Timeframe,
		"open":      candle.Open,
		"close":     candle.Close,
	}).Debug("Candle processed and saved")

	return nil
}

// processOrderBook processes order book and saves it to storage
func (bis *BrokerStorageIntegration) processOrderBook(integration *BrokerIntegration, orderBook entities.OrderBook) error {
	// Add broker information to order book
	orderBook.BrokerID = integration.brokerID

	// Save order book through StorageService
	ctx, cancel := context.WithTimeout(bis.ctx, 10*time.Second)
	defer cancel()

	if err := bis.storageService.SaveOrderBook(ctx, orderBook); err != nil {
		return fmt.Errorf("failed to save orderbook: %w", err)
	}

	// Update statistics
	now := time.Now()

	integration.mu.Lock()
	integration.stats.OrderBooksProcessed++
	integration.stats.LastDataReceived = now
	integration.mu.Unlock()

	bis.statsMu.Lock()
	bis.stats.TotalOrderBooks++
	bis.stats.LastDataReceived = now
	bis.statsMu.Unlock()

	bis.logger.WithFields(logrus.Fields{
		"broker_id": integration.brokerID,
		"symbol":    orderBook.Symbol,
		"bids":      len(orderBook.Bids),
		"asks":      len(orderBook.Asks),
	}).Debug("OrderBook processed and saved")

	return nil
}
