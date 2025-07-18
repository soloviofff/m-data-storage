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

// DataCollectorService implements the DataCollector interface
type DataCollectorService struct {
	dataPipeline      interfaces.DataPipeline
	instrumentManager interfaces.InstrumentManager
	dataProcessor     interfaces.DataProcessor
	logger            *logrus.Logger

	// Data channels
	tickerChan    chan entities.Ticker
	candleChan    chan entities.Candle
	orderBookChan chan entities.OrderBook

	// Collection state
	isCollecting bool
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	mu           sync.RWMutex

	// Statistics
	stats   InternalCollectionStats
	statsMu sync.RWMutex

	// Configuration
	config DataCollectorConfig
}

// DataCollectorConfig holds configuration for data collector
type DataCollectorConfig struct {
	ChannelBufferSize int           `yaml:"channel_buffer_size"`
	WorkerCount       int           `yaml:"worker_count"`
	ProcessTimeout    time.Duration `yaml:"process_timeout"`
	StatsInterval     time.Duration `yaml:"stats_interval"`
}

// InternalCollectionStats holds internal collection statistics
type InternalCollectionStats struct {
	StartedAt           time.Time `json:"started_at"`
	CollectedTickers    int64     `json:"collected_tickers"`
	CollectedCandles    int64     `json:"collected_candles"`
	CollectedOrderBooks int64     `json:"collected_orderbooks"`
	ProcessedTickers    int64     `json:"processed_tickers"`
	ProcessedCandles    int64     `json:"processed_candles"`
	ProcessedOrderBooks int64     `json:"processed_orderbooks"`
	Errors              int64     `json:"errors"`
	LastActivity        time.Time `json:"last_activity"`
	ActiveSubscriptions int       `json:"active_subscriptions"`
}

// NewDataCollectorService creates a new data collector service
func NewDataCollectorService(
	dataPipeline interfaces.DataPipeline,
	instrumentManager interfaces.InstrumentManager,
	dataProcessor interfaces.DataProcessor,
	logger *logrus.Logger,
) *DataCollectorService {
	if logger == nil {
		logger = logrus.New()
	}

	config := DataCollectorConfig{
		ChannelBufferSize: 1000,
		WorkerCount:       3,
		ProcessTimeout:    30 * time.Second,
		StatsInterval:     60 * time.Second,
	}

	return &DataCollectorService{
		dataPipeline:      dataPipeline,
		instrumentManager: instrumentManager,
		dataProcessor:     dataProcessor,
		logger:            logger,
		config:            config,
		tickerChan:        make(chan entities.Ticker, config.ChannelBufferSize),
		candleChan:        make(chan entities.Candle, config.ChannelBufferSize),
		orderBookChan:     make(chan entities.OrderBook, config.ChannelBufferSize),
		stats: InternalCollectionStats{
			StartedAt: time.Now(),
		},
	}
}

// StartCollection starts data collection
func (dcs *DataCollectorService) StartCollection(ctx context.Context) error {
	dcs.mu.Lock()
	defer dcs.mu.Unlock()

	if dcs.isCollecting {
		return fmt.Errorf("data collection is already running")
	}

	dcs.ctx, dcs.cancel = context.WithCancel(ctx)
	dcs.isCollecting = true

	// Start data processor
	if err := dcs.dataProcessor.Start(dcs.ctx); err != nil {
		dcs.isCollecting = false
		return fmt.Errorf("failed to start data processor: %w", err)
	}

	// Start data pipeline
	if err := dcs.dataPipeline.Start(dcs.ctx); err != nil {
		dcs.dataProcessor.Stop()
		dcs.isCollecting = false
		return fmt.Errorf("failed to start data pipeline: %w", err)
	}

	// Start workers
	dcs.wg.Add(dcs.config.WorkerCount + 1) // +1 for stats worker
	go dcs.tickerWorker()
	go dcs.candleWorker()
	go dcs.orderBookWorker()
	go dcs.statsWorker()

	dcs.stats.StartedAt = time.Now()
	dcs.logger.Info("Data collection started")

	return nil
}

// StopCollection stops data collection
func (dcs *DataCollectorService) StopCollection() error {
	dcs.mu.Lock()
	defer dcs.mu.Unlock()

	if !dcs.isCollecting {
		return fmt.Errorf("data collection is not running")
	}

	// Cancel context
	if dcs.cancel != nil {
		dcs.cancel()
	}

	// Close channels
	close(dcs.tickerChan)
	close(dcs.candleChan)
	close(dcs.orderBookChan)

	// Wait for workers to finish
	dcs.wg.Wait()

	// Stop services
	if err := dcs.dataProcessor.Stop(); err != nil {
		dcs.logger.WithError(err).Error("Failed to stop data processor")
	}

	if err := dcs.dataPipeline.Stop(); err != nil {
		dcs.logger.WithError(err).Error("Failed to stop data pipeline")
	}

	dcs.isCollecting = false
	dcs.logger.Info("Data collection stopped")

	return nil
}

// Subscribe subscribes to broker instruments
func (dcs *DataCollectorService) Subscribe(ctx context.Context, brokerID string, subscription entities.InstrumentSubscription) error {
	// Add subscription to instrument manager
	if err := dcs.instrumentManager.AddSubscription(ctx, subscription); err != nil {
		return fmt.Errorf("failed to add subscription to instrument manager: %w", err)
	}

	// Subscribe to data pipeline
	subscriptions := []entities.InstrumentSubscription{subscription}
	if err := dcs.dataPipeline.Subscribe(ctx, brokerID, subscriptions); err != nil {
		// Try to remove from instrument manager if pipeline subscription fails
		dcs.instrumentManager.RemoveSubscription(ctx, subscription.ID)
		return fmt.Errorf("failed to subscribe to data pipeline: %w", err)
	}

	dcs.logger.WithFields(logrus.Fields{
		"broker_id":       brokerID,
		"subscription_id": subscription.ID,
		"symbol":          subscription.Symbol,
	}).Info("Successfully subscribed to instrument")

	return nil
}

// Unsubscribe unsubscribes from broker instruments
func (dcs *DataCollectorService) Unsubscribe(ctx context.Context, brokerID string, subscriptionID string) error {
	// Get subscription details
	subscription, err := dcs.instrumentManager.GetSubscription(ctx, subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	// Unsubscribe from data pipeline
	subscriptions := []entities.InstrumentSubscription{*subscription}
	if err := dcs.dataPipeline.Unsubscribe(ctx, brokerID, subscriptions); err != nil {
		dcs.logger.WithError(err).Error("Failed to unsubscribe from data pipeline")
		// Continue with removal from instrument manager even if pipeline fails
	}

	// Remove from instrument manager
	if err := dcs.instrumentManager.RemoveSubscription(ctx, subscriptionID); err != nil {
		return fmt.Errorf("failed to remove subscription from instrument manager: %w", err)
	}

	dcs.logger.WithFields(logrus.Fields{
		"broker_id":       brokerID,
		"subscription_id": subscriptionID,
		"symbol":          subscription.Symbol,
	}).Info("Successfully unsubscribed from instrument")

	return nil
}

// GetTickerChannel returns ticker data channel
func (dcs *DataCollectorService) GetTickerChannel() <-chan entities.Ticker {
	return dcs.tickerChan
}

// GetCandleChannel returns candle data channel
func (dcs *DataCollectorService) GetCandleChannel() <-chan entities.Candle {
	return dcs.candleChan
}

// GetOrderBookChannel returns order book data channel
func (dcs *DataCollectorService) GetOrderBookChannel() <-chan entities.OrderBook {
	return dcs.orderBookChan
}

// GetCollectionStats returns collection statistics
func (dcs *DataCollectorService) GetCollectionStats() interfaces.CollectionStats {
	dcs.statsMu.RLock()
	defer dcs.statsMu.RUnlock()

	// Update active subscriptions count
	activeSubscriptions := 0
	if subscriptions, err := dcs.instrumentManager.ListSubscriptions(context.Background()); err == nil {
		for _, sub := range subscriptions {
			if sub.IsActive {
				activeSubscriptions++
			}
		}
	}

	// Calculate rates (per second)
	duration := time.Since(dcs.stats.StartedAt).Seconds()
	if duration == 0 {
		duration = 1 // Avoid division by zero
	}

	return interfaces.CollectionStats{
		TotalTickers:        dcs.stats.CollectedTickers,
		TotalCandles:        dcs.stats.CollectedCandles,
		TotalOrderBooks:     dcs.stats.CollectedOrderBooks,
		TickersPerSecond:    float64(dcs.stats.CollectedTickers) / duration,
		CandlesPerSecond:    float64(dcs.stats.CollectedCandles) / duration,
		OrderBooksPerSecond: float64(dcs.stats.CollectedOrderBooks) / duration,
		ActiveSubscriptions: activeSubscriptions,
		ConnectedBrokers:    0, // TODO: Get from broker manager
		LastUpdate:          dcs.stats.LastActivity,
		Errors:              dcs.stats.Errors,
	}
}

// Health checks collector health
func (dcs *DataCollectorService) Health() error {
	dcs.mu.RLock()
	defer dcs.mu.RUnlock()

	if !dcs.isCollecting {
		return fmt.Errorf("data collection is not running")
	}

	// Check data pipeline health
	if err := dcs.dataPipeline.Health(); err != nil {
		return fmt.Errorf("data pipeline health check failed: %w", err)
	}

	// Check data processor health
	if err := dcs.dataProcessor.Health(); err != nil {
		return fmt.Errorf("data processor health check failed: %w", err)
	}

	return nil
}

// tickerWorker processes ticker data
func (dcs *DataCollectorService) tickerWorker() {
	defer dcs.wg.Done()

	for {
		select {
		case ticker, ok := <-dcs.tickerChan:
			if !ok {
				return
			}

			dcs.updateStats(1, 0, 0, false)

			// Process ticker through data processor
			ctx, cancel := context.WithTimeout(context.Background(), dcs.config.ProcessTimeout)
			if err := dcs.dataProcessor.ProcessTicker(ctx, ticker); err != nil {
				dcs.logger.WithError(err).WithField("symbol", ticker.Symbol).Error("Failed to process ticker")
				dcs.updateStats(0, 0, 0, true)
			} else {
				dcs.updateProcessedStats(1, 0, 0)
			}
			cancel()

		case <-dcs.ctx.Done():
			return
		}
	}
}

// candleWorker processes candle data
func (dcs *DataCollectorService) candleWorker() {
	defer dcs.wg.Done()

	for {
		select {
		case candle, ok := <-dcs.candleChan:
			if !ok {
				return
			}

			dcs.updateStats(0, 1, 0, false)

			// Process candle through data processor
			ctx, cancel := context.WithTimeout(context.Background(), dcs.config.ProcessTimeout)
			if err := dcs.dataProcessor.ProcessCandle(ctx, candle); err != nil {
				dcs.logger.WithError(err).WithField("symbol", candle.Symbol).Error("Failed to process candle")
				dcs.updateStats(0, 0, 0, true)
			} else {
				dcs.updateProcessedStats(0, 1, 0)
			}
			cancel()

		case <-dcs.ctx.Done():
			return
		}
	}
}

// orderBookWorker processes order book data
func (dcs *DataCollectorService) orderBookWorker() {
	defer dcs.wg.Done()

	for {
		select {
		case orderBook, ok := <-dcs.orderBookChan:
			if !ok {
				return
			}

			dcs.updateStats(0, 0, 1, false)

			// Process order book through data processor
			ctx, cancel := context.WithTimeout(context.Background(), dcs.config.ProcessTimeout)
			if err := dcs.dataProcessor.ProcessOrderBook(ctx, orderBook); err != nil {
				dcs.logger.WithError(err).WithField("symbol", orderBook.Symbol).Error("Failed to process order book")
				dcs.updateStats(0, 0, 0, true)
			} else {
				dcs.updateProcessedStats(0, 0, 1)
			}
			cancel()

		case <-dcs.ctx.Done():
			return
		}
	}
}

// statsWorker periodically logs statistics
func (dcs *DataCollectorService) statsWorker() {
	defer dcs.wg.Done()

	ticker := time.NewTicker(dcs.config.StatsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := dcs.GetCollectionStats()
			dcs.logger.WithFields(logrus.Fields{
				"total_tickers":         stats.TotalTickers,
				"total_candles":         stats.TotalCandles,
				"total_orderbooks":      stats.TotalOrderBooks,
				"tickers_per_second":    stats.TickersPerSecond,
				"candles_per_second":    stats.CandlesPerSecond,
				"orderbooks_per_second": stats.OrderBooksPerSecond,
				"errors":                stats.Errors,
				"active_subscriptions":  stats.ActiveSubscriptions,
				"connected_brokers":     stats.ConnectedBrokers,
			}).Info("Data collection statistics")

		case <-dcs.ctx.Done():
			return
		}
	}
}

// updateStats updates collection statistics
func (dcs *DataCollectorService) updateStats(tickers, candles, orderBooks int64, isError bool) {
	dcs.statsMu.Lock()
	defer dcs.statsMu.Unlock()

	dcs.stats.CollectedTickers += tickers
	dcs.stats.CollectedCandles += candles
	dcs.stats.CollectedOrderBooks += orderBooks
	dcs.stats.LastActivity = time.Now()

	if isError {
		dcs.stats.Errors++
	}
}

// updateProcessedStats updates processed statistics
func (dcs *DataCollectorService) updateProcessedStats(tickers, candles, orderBooks int64) {
	dcs.statsMu.Lock()
	defer dcs.statsMu.Unlock()

	dcs.stats.ProcessedTickers += tickers
	dcs.stats.ProcessedCandles += candles
	dcs.stats.ProcessedOrderBooks += orderBooks
}

// IsCollecting returns whether collection is active
func (dcs *DataCollectorService) IsCollecting() bool {
	dcs.mu.RLock()
	defer dcs.mu.RUnlock()
	return dcs.isCollecting
}

// GetConfig returns collector configuration
func (dcs *DataCollectorService) GetConfig() DataCollectorConfig {
	return dcs.config
}

// SetConfig updates collector configuration
func (dcs *DataCollectorService) SetConfig(config DataCollectorConfig) {
	dcs.config = config
}
