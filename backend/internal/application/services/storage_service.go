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

// StorageService provides high-level operations for working with data storage
type StorageService struct {
	storageManager interfaces.StorageManager
	validator      interfaces.DataValidator
	logger         *logrus.Logger

	// Batch processing settings
	batchSize     int
	flushInterval time.Duration

	// Buffers for batch processing
	tickerBuffer    []entities.Ticker
	candleBuffer    []entities.Candle
	orderBookBuffer []entities.OrderBook

	// Mutexes for buffer protection
	tickerMutex    sync.Mutex
	candleMutex    sync.Mutex
	orderBookMutex sync.Mutex

	// Channels for flush notifications
	flushTicker    *time.Ticker
	stopChan       chan struct{}
	flushWaitGroup sync.WaitGroup

	// Statistics
	stats StorageStats
	mutex sync.RWMutex
}

// StorageStats contains storage service statistics
type StorageStats = interfaces.StorageServiceStats

// StorageServiceConfig contains configuration for StorageService
type StorageServiceConfig struct {
	BatchSize     int           `yaml:"batch_size" json:"batch_size"`
	FlushInterval time.Duration `yaml:"flush_interval" json:"flush_interval"`
}

// DefaultStorageServiceConfig returns default configuration
func DefaultStorageServiceConfig() StorageServiceConfig {
	return StorageServiceConfig{
		BatchSize:     1000,
		FlushInterval: 5 * time.Second,
	}
}

// NewStorageService creates a new StorageService instance
func NewStorageService(
	storageManager interfaces.StorageManager,
	validator interfaces.DataValidator,
	logger *logrus.Logger,
	config StorageServiceConfig,
) *StorageService {
	if logger == nil {
		logger = logrus.New()
	}

	service := &StorageService{
		storageManager:  storageManager,
		validator:       validator,
		logger:          logger,
		batchSize:       config.BatchSize,
		flushInterval:   config.FlushInterval,
		tickerBuffer:    make([]entities.Ticker, 0, config.BatchSize),
		candleBuffer:    make([]entities.Candle, 0, config.BatchSize),
		orderBookBuffer: make([]entities.OrderBook, 0, config.BatchSize),
		stopChan:        make(chan struct{}),
	}

	// Start periodic flush
	service.startPeriodicFlush()

	return service
}

// SaveTicker saves a ticker with batch processing
func (s *StorageService) SaveTicker(ctx context.Context, ticker entities.Ticker) error {
	// Validate ticker
	if err := s.validator.ValidateTicker(ticker); err != nil {
		s.incrementErrors()
		return fmt.Errorf("ticker validation failed: %w", err)
	}

	s.tickerMutex.Lock()
	defer s.tickerMutex.Unlock()

	s.tickerBuffer = append(s.tickerBuffer, ticker)

	// If buffer is full, flush it
	if len(s.tickerBuffer) >= s.batchSize {
		return s.flushTickersUnsafe(ctx)
	}

	return nil
}

// SaveCandle saves a candle with batch processing
func (s *StorageService) SaveCandle(ctx context.Context, candle entities.Candle) error {
	// Validate candle
	if err := s.validator.ValidateCandle(candle); err != nil {
		s.incrementErrors()
		return fmt.Errorf("candle validation failed: %w", err)
	}

	s.candleMutex.Lock()
	defer s.candleMutex.Unlock()

	s.candleBuffer = append(s.candleBuffer, candle)

	// If buffer is full, flush it
	if len(s.candleBuffer) >= s.batchSize {
		return s.flushCandlesUnsafe(ctx)
	}

	return nil
}

// SaveOrderBook saves an order book with batch processing
func (s *StorageService) SaveOrderBook(ctx context.Context, orderBook entities.OrderBook) error {
	// Validate order book
	if err := s.validator.ValidateOrderBook(orderBook); err != nil {
		s.incrementErrors()
		return fmt.Errorf("orderbook validation failed: %w", err)
	}

	s.orderBookMutex.Lock()
	defer s.orderBookMutex.Unlock()

	s.orderBookBuffer = append(s.orderBookBuffer, orderBook)

	// If buffer is full, flush it
	if len(s.orderBookBuffer) >= s.batchSize {
		return s.flushOrderBooksUnsafe(ctx)
	}

	return nil
}

// SaveTickers saves multiple tickers
func (s *StorageService) SaveTickers(ctx context.Context, tickers []entities.Ticker) error {
	// Validate all tickers
	for i, ticker := range tickers {
		if err := s.validator.ValidateTicker(ticker); err != nil {
			s.incrementErrors()
			return fmt.Errorf("ticker validation failed at index %d: %w", i, err)
		}
	}

	// Save directly to storage manager
	if err := s.storageManager.SaveTickers(ctx, tickers); err != nil {
		s.incrementErrors()
		return fmt.Errorf("failed to save tickers: %w", err)
	}

	s.updateStats(int64(len(tickers)), 0, 0)
	return nil
}

// SaveCandles saves multiple candles
func (s *StorageService) SaveCandles(ctx context.Context, candles []entities.Candle) error {
	// Validate all candles
	for i, candle := range candles {
		if err := s.validator.ValidateCandle(candle); err != nil {
			s.incrementErrors()
			return fmt.Errorf("candle validation failed at index %d: %w", i, err)
		}
	}

	// Save directly to storage manager
	if err := s.storageManager.SaveCandles(ctx, candles); err != nil {
		s.incrementErrors()
		return fmt.Errorf("failed to save candles: %w", err)
	}

	s.updateStats(0, int64(len(candles)), 0)
	return nil
}

// SaveOrderBooks saves multiple order books
func (s *StorageService) SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error {
	// Validate all order books
	for i, orderBook := range orderBooks {
		if err := s.validator.ValidateOrderBook(orderBook); err != nil {
			s.incrementErrors()
			return fmt.Errorf("orderbook validation failed at index %d: %w", i, err)
		}
	}

	// Save directly to storage manager
	if err := s.storageManager.SaveOrderBooks(ctx, orderBooks); err != nil {
		s.incrementErrors()
		return fmt.Errorf("failed to save orderbooks: %w", err)
	}

	s.updateStats(0, 0, int64(len(orderBooks)))
	return nil
}

// GetTickers retrieves tickers with filtering
func (s *StorageService) GetTickers(ctx context.Context, filter interfaces.TickerFilter) ([]entities.Ticker, error) {
	return s.storageManager.GetTickers(ctx, filter)
}

// GetCandles retrieves candles with filtering
func (s *StorageService) GetCandles(ctx context.Context, filter interfaces.CandleFilter) ([]entities.Candle, error) {
	return s.storageManager.GetCandles(ctx, filter)
}

// GetOrderBooks retrieves order books with filtering
func (s *StorageService) GetOrderBooks(ctx context.Context, filter interfaces.OrderBookFilter) ([]entities.OrderBook, error) {
	return s.storageManager.GetOrderBooks(ctx, filter)
}

// FlushAll forcibly flushes all buffers
func (s *StorageService) FlushAll(ctx context.Context) error {
	var errors []error

	// Flush tickers
	s.tickerMutex.Lock()
	if len(s.tickerBuffer) > 0 {
		if err := s.flushTickersUnsafe(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to flush tickers: %w", err))
		}
	}
	s.tickerMutex.Unlock()

	// Flush candles
	s.candleMutex.Lock()
	if len(s.candleBuffer) > 0 {
		if err := s.flushCandlesUnsafe(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to flush candles: %w", err))
		}
	}
	s.candleMutex.Unlock()

	// Flush order books
	s.orderBookMutex.Lock()
	if len(s.orderBookBuffer) > 0 {
		if err := s.flushOrderBooksUnsafe(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to flush orderbooks: %w", err))
		}
	}
	s.orderBookMutex.Unlock()

	if len(errors) > 0 {
		return fmt.Errorf("flush errors: %v", errors)
	}

	return nil
}

// GetStats returns service statistics
func (s *StorageService) GetStats() StorageStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.stats
}

// Close closes the service and flushes all buffers
func (s *StorageService) Close(ctx context.Context) error {
	// Stop periodic flush
	close(s.stopChan)
	if s.flushTicker != nil {
		s.flushTicker.Stop()
	}

	// Wait for all flush operations to complete
	s.flushWaitGroup.Wait()

	// Flush all remaining data
	return s.FlushAll(ctx)
}

// startPeriodicFlush starts periodic buffer flushing
func (s *StorageService) startPeriodicFlush() {
	s.flushTicker = time.NewTicker(s.flushInterval)

	s.flushWaitGroup.Add(1)
	go func() {
		defer s.flushWaitGroup.Done()

		for {
			select {
			case <-s.flushTicker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				if err := s.FlushAll(ctx); err != nil {
					s.logger.WithError(err).Error("Failed to flush buffers periodically")
				}
				cancel()

			case <-s.stopChan:
				return
			}
		}
	}()
}

// flushTickersUnsafe flushes ticker buffer (must be called under mutex)
func (s *StorageService) flushTickersUnsafe(ctx context.Context) error {
	if len(s.tickerBuffer) == 0 {
		return nil
	}

	if err := s.storageManager.SaveTickers(ctx, s.tickerBuffer); err != nil {
		s.incrementErrors()
		return fmt.Errorf("failed to save tickers batch: %w", err)
	}

	count := int64(len(s.tickerBuffer))
	s.tickerBuffer = s.tickerBuffer[:0] // Clear buffer
	s.updateStats(count, 0, 0)
	s.incrementBatchesFlashed()

	s.logger.WithField("count", count).Debug("Flushed tickers batch")
	return nil
}

// flushCandlesUnsafe flushes candle buffer (must be called under mutex)
func (s *StorageService) flushCandlesUnsafe(ctx context.Context) error {
	if len(s.candleBuffer) == 0 {
		return nil
	}

	if err := s.storageManager.SaveCandles(ctx, s.candleBuffer); err != nil {
		s.incrementErrors()
		return fmt.Errorf("failed to save candles batch: %w", err)
	}

	count := int64(len(s.candleBuffer))
	s.candleBuffer = s.candleBuffer[:0] // Clear buffer
	s.updateStats(0, count, 0)
	s.incrementBatchesFlashed()

	s.logger.WithField("count", count).Debug("Flushed candles batch")
	return nil
}

// flushOrderBooksUnsafe flushes order book buffer (must be called under mutex)
func (s *StorageService) flushOrderBooksUnsafe(ctx context.Context) error {
	if len(s.orderBookBuffer) == 0 {
		return nil
	}

	if err := s.storageManager.SaveOrderBooks(ctx, s.orderBookBuffer); err != nil {
		s.incrementErrors()
		return fmt.Errorf("failed to save orderbooks batch: %w", err)
	}

	count := int64(len(s.orderBookBuffer))
	s.orderBookBuffer = s.orderBookBuffer[:0] // Clear buffer
	s.updateStats(0, 0, count)
	s.incrementBatchesFlashed()

	s.logger.WithField("count", count).Debug("Flushed orderbooks batch")
	return nil
}

// updateStats updates statistics
func (s *StorageService) updateStats(tickers, candles, orderBooks int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stats.TickersSaved += tickers
	s.stats.CandlesSaved += candles
	s.stats.OrderBooksSaved += orderBooks
	s.stats.LastFlushTime = time.Now()
}

// incrementErrors increments error counter
func (s *StorageService) incrementErrors() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats.ErrorsCount++
}

// incrementBatchesFlashed increments flushed batches counter
func (s *StorageService) incrementBatchesFlashed() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats.BatchesFlashed++
}
