package services

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// DataProcessorService implements the DataProcessor interface
type DataProcessorService struct {
	storage   interfaces.StorageManager
	validator interfaces.DataValidator
	logger    *logrus.Logger

	// Channels for batch processing
	tickerChan    chan entities.Ticker
	candleChan    chan entities.Candle
	orderBookChan chan entities.OrderBook

	// Batch processing settings
	batchSize     int
	flushInterval time.Duration

	// Advanced buffering settings
	adaptiveBatching  bool
	priorityBuffering bool
	maxBatchSize      int
	minBatchSize      int
	adaptiveThreshold float64 // Channel utilization threshold for adaptive batching

	// Buffer monitoring
	bufferStats BufferStats
	bufferMutex sync.RWMutex

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Statistics
	stats ProcessorStats
	mutex sync.RWMutex
}

// ProcessorStats - data processor statistics
type ProcessorStats struct {
	ProcessedTickers    int64     `json:"processed_tickers"`
	ProcessedCandles    int64     `json:"processed_candles"`
	ProcessedOrderBooks int64     `json:"processed_orderbooks"`
	Errors              int64     `json:"errors"`
	LastProcessed       time.Time `json:"last_processed"`
}

// BufferStats - buffer monitoring statistics
type BufferStats struct {
	TickerChannelUtilization    float64   `json:"ticker_channel_utilization"`
	CandleChannelUtilization    float64   `json:"candle_channel_utilization"`
	OrderBookChannelUtilization float64   `json:"orderbook_channel_utilization"`
	AdaptiveBatchSizes          []int     `json:"adaptive_batch_sizes"`
	OverflowEvents              int64     `json:"overflow_events"`
	LastBufferCheck             time.Time `json:"last_buffer_check"`
}

// NewDataProcessorService creates a new data processing service
func NewDataProcessorService(
	storage interfaces.StorageManager,
	validator interfaces.DataValidator,
	logger *logrus.Logger,
) *DataProcessorService {
	return &DataProcessorService{
		storage:       storage,
		validator:     validator,
		logger:        logger,
		batchSize:     100,
		flushInterval: 5 * time.Second,
		tickerChan:    make(chan entities.Ticker, 1000),
		candleChan:    make(chan entities.Candle, 1000),
		orderBookChan: make(chan entities.OrderBook, 1000),

		// Advanced buffering settings
		adaptiveBatching:  true,
		priorityBuffering: true,
		maxBatchSize:      500,
		minBatchSize:      10,
		adaptiveThreshold: 0.7, // 70% channel utilization threshold

		// Initialize buffer stats
		bufferStats: BufferStats{
			AdaptiveBatchSizes: make([]int, 0, 100), // Keep last 100 batch sizes
		},
	}
}

// Start starts the data processing service
func (s *DataProcessorService) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	// Start workers for batch processing
	s.wg.Add(3)
	go s.tickerWorker()
	go s.candleWorker()
	go s.orderBookWorker()

	s.logger.Info("Data processor service started")
	return nil
}

// Stop stops the data processing service
func (s *DataProcessorService) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}

	// Close channels
	close(s.tickerChan)
	close(s.candleChan)
	close(s.orderBookChan)

	// Wait for workers to finish
	s.wg.Wait()

	s.logger.Info("Data processor service stopped")
	return nil
}

// ProcessTicker processes a single ticker
func (s *DataProcessorService) ProcessTicker(ctx context.Context, ticker entities.Ticker) error {
	if err := s.validator.ValidateTicker(ticker); err != nil {
		s.incrementErrors()
		return errors.Wrap(err, "ticker validation failed")
	}

	select {
	case s.tickerChan <- ticker:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.incrementErrors()
		return errors.New("ticker channel is full")
	}
}

// ProcessCandle processes a single candle
func (s *DataProcessorService) ProcessCandle(ctx context.Context, candle entities.Candle) error {
	if err := s.validator.ValidateCandle(candle); err != nil {
		s.incrementErrors()
		return errors.Wrap(err, "candle validation failed")
	}

	select {
	case s.candleChan <- candle:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.incrementErrors()
		return errors.New("candle channel is full")
	}
}

// ProcessOrderBook processes a single orderbook
func (s *DataProcessorService) ProcessOrderBook(ctx context.Context, orderBook entities.OrderBook) error {
	if err := s.validator.ValidateOrderBook(orderBook); err != nil {
		s.incrementErrors()
		return errors.Wrap(err, "orderbook validation failed")
	}

	select {
	case s.orderBookChan <- orderBook:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		s.incrementErrors()
		return errors.New("orderbook channel is full")
	}
}

// ProcessTickerBatch processes a batch of tickers
func (s *DataProcessorService) ProcessTickerBatch(ctx context.Context, tickers []entities.Ticker) error {
	// Validate all tickers
	for _, ticker := range tickers {
		if err := s.validator.ValidateTicker(ticker); err != nil {
			s.incrementErrors()
			return errors.Wrap(err, "ticker validation failed")
		}
	}

	// Save to storage
	if err := s.storage.SaveTickers(ctx, tickers); err != nil {
		s.incrementErrors()
		return errors.Wrap(err, "failed to save tickers")
	}

	s.updateStats(len(tickers), 0, 0)
	return nil
}

// ProcessCandleBatch processes a batch of candles
func (s *DataProcessorService) ProcessCandleBatch(ctx context.Context, candles []entities.Candle) error {
	// Validate all candles
	for _, candle := range candles {
		if err := s.validator.ValidateCandle(candle); err != nil {
			s.incrementErrors()
			return errors.Wrap(err, "candle validation failed")
		}
	}

	// Save to storage
	if err := s.storage.SaveCandles(ctx, candles); err != nil {
		s.incrementErrors()
		return errors.Wrap(err, "failed to save candles")
	}

	s.updateStats(0, len(candles), 0)
	return nil
}

// ProcessOrderBookBatch processes a batch of orderbooks
func (s *DataProcessorService) ProcessOrderBookBatch(ctx context.Context, orderBooks []entities.OrderBook) error {
	// Validate all orderbooks
	for _, orderBook := range orderBooks {
		if err := s.validator.ValidateOrderBook(orderBook); err != nil {
			s.incrementErrors()
			return errors.Wrap(err, "orderbook validation failed")
		}
	}

	// Save to storage
	if err := s.storage.SaveOrderBooks(ctx, orderBooks); err != nil {
		s.incrementErrors()
		return errors.Wrap(err, "failed to save orderbooks")
	}

	s.updateStats(0, 0, len(orderBooks))
	return nil
}

// Health checks the service health
func (s *DataProcessorService) Health() error {
	// Check channel states
	if len(s.tickerChan) > cap(s.tickerChan)*9/10 {
		return errors.New("ticker channel is nearly full")
	}
	if len(s.candleChan) > cap(s.candleChan)*9/10 {
		return errors.New("candle channel is nearly full")
	}
	if len(s.orderBookChan) > cap(s.orderBookChan)*9/10 {
		return errors.New("orderbook channel is nearly full")
	}

	return nil
}

// GetStats returns processor statistics
func (s *DataProcessorService) GetStats() ProcessorStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.stats
}

// tickerWorker processes tickers in batch mode
func (s *DataProcessorService) tickerWorker() {
	defer s.wg.Done()

	batch := make([]entities.Ticker, 0, s.maxBatchSize)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	// Buffer monitoring ticker
	bufferTicker := time.NewTicker(1 * time.Second)
	defer bufferTicker.Stop()

	for {
		select {
		case data, ok := <-s.tickerChan:
			if !ok {
				// Channel closed, process remaining data
				if len(batch) > 0 {
					s.processBatch(batch, "ticker")
				}
				return
			}

			batch = append(batch, data)

			// Calculate adaptive batch size
			adaptiveSize := s.calculateAdaptiveBatchSize(len(s.tickerChan), cap(s.tickerChan))

			if len(batch) >= adaptiveSize {
				s.processBatch(batch, "ticker")
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.processBatch(batch, "ticker")
				batch = batch[:0]
			}

		case <-bufferTicker.C:
			// Update buffer statistics
			s.updateBufferStats()

		case <-s.ctx.Done():
			return
		}
	}
}

// candleWorker processes candles in batch mode
func (s *DataProcessorService) candleWorker() {
	defer s.wg.Done()

	batch := make([]entities.Candle, 0, s.maxBatchSize)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case data, ok := <-s.candleChan:
			if !ok {
				if len(batch) > 0 {
					s.processCandleBatch(batch)
				}
				return
			}

			batch = append(batch, data)

			// Calculate adaptive batch size
			adaptiveSize := s.calculateAdaptiveBatchSize(len(s.candleChan), cap(s.candleChan))

			if len(batch) >= adaptiveSize {
				s.processCandleBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.processCandleBatch(batch)
				batch = batch[:0]
			}

		case <-s.ctx.Done():
			return
		}
	}
}

// orderBookWorker processes orderbooks in batch mode
func (s *DataProcessorService) orderBookWorker() {
	defer s.wg.Done()

	batch := make([]entities.OrderBook, 0, s.maxBatchSize)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case data, ok := <-s.orderBookChan:
			if !ok {
				if len(batch) > 0 {
					s.processOrderBookBatch(batch)
				}
				return
			}

			batch = append(batch, data)

			// Calculate adaptive batch size
			adaptiveSize := s.calculateAdaptiveBatchSize(len(s.orderBookChan), cap(s.orderBookChan))

			if len(batch) >= adaptiveSize {
				s.processOrderBookBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.processOrderBookBatch(batch)
				batch = batch[:0]
			}

		case <-s.ctx.Done():
			return
		}
	}
}

// processBatch processes a batch of tickers
func (s *DataProcessorService) processBatch(batch []entities.Ticker, dataType string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.storage.SaveTickers(ctx, batch); err != nil {
		s.logger.WithError(err).WithField("data_type", dataType).Error("Failed to save ticker batch")
		s.incrementErrors()
		return
	}

	s.updateStats(len(batch), 0, 0)
}

// processCandleBatch processes a batch of candles
func (s *DataProcessorService) processCandleBatch(batch []entities.Candle) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.storage.SaveCandles(ctx, batch); err != nil {
		s.logger.WithError(err).Error("Failed to save candle batch")
		s.incrementErrors()
		return
	}

	s.updateStats(0, len(batch), 0)
}

// processOrderBookBatch processes a batch of orderbooks
func (s *DataProcessorService) processOrderBookBatch(batch []entities.OrderBook) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.storage.SaveOrderBooks(ctx, batch); err != nil {
		s.logger.WithError(err).Error("Failed to save orderbook batch")
		s.incrementErrors()
		return
	}

	s.updateStats(0, 0, len(batch))
}

// updateStats updates statistics
func (s *DataProcessorService) updateStats(tickers, candles, orderBooks int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stats.ProcessedTickers += int64(tickers)
	s.stats.ProcessedCandles += int64(candles)
	s.stats.ProcessedOrderBooks += int64(orderBooks)
	s.stats.LastProcessed = time.Now()
}

// incrementErrors increments the error counter
func (s *DataProcessorService) incrementErrors() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats.Errors++
}

// calculateAdaptiveBatchSize calculates optimal batch size based on channel utilization
func (s *DataProcessorService) calculateAdaptiveBatchSize(channelLen, channelCap int) int {
	if !s.adaptiveBatching {
		return s.batchSize
	}

	utilization := float64(channelLen) / float64(channelCap)

	var adaptiveSize int
	if utilization > s.adaptiveThreshold {
		// High utilization - increase batch size for better throughput
		adaptiveSize = int(float64(s.batchSize) * (1.0 + utilization))
		if adaptiveSize > s.maxBatchSize {
			adaptiveSize = s.maxBatchSize
		}
	} else {
		// Low utilization - decrease batch size for better latency
		adaptiveSize = int(float64(s.batchSize) * utilization)
		if adaptiveSize < s.minBatchSize {
			adaptiveSize = s.minBatchSize
		}
	}

	// Track adaptive batch sizes
	s.bufferMutex.Lock()
	s.bufferStats.AdaptiveBatchSizes = append(s.bufferStats.AdaptiveBatchSizes, adaptiveSize)
	if len(s.bufferStats.AdaptiveBatchSizes) > 100 {
		s.bufferStats.AdaptiveBatchSizes = s.bufferStats.AdaptiveBatchSizes[1:]
	}
	s.bufferMutex.Unlock()

	return adaptiveSize
}

// updateBufferStats updates buffer monitoring statistics
func (s *DataProcessorService) updateBufferStats() {
	s.bufferMutex.Lock()
	defer s.bufferMutex.Unlock()

	tickerCap := cap(s.tickerChan)
	candleCap := cap(s.candleChan)
	orderBookCap := cap(s.orderBookChan)

	s.bufferStats.TickerChannelUtilization = float64(len(s.tickerChan)) / float64(tickerCap)
	s.bufferStats.CandleChannelUtilization = float64(len(s.candleChan)) / float64(candleCap)
	s.bufferStats.OrderBookChannelUtilization = float64(len(s.orderBookChan)) / float64(orderBookCap)
	s.bufferStats.LastBufferCheck = time.Now()

	// Check for overflow conditions
	if s.bufferStats.TickerChannelUtilization > 0.95 ||
		s.bufferStats.CandleChannelUtilization > 0.95 ||
		s.bufferStats.OrderBookChannelUtilization > 0.95 {
		s.bufferStats.OverflowEvents++
		s.logger.Warn("Buffer overflow risk detected",
			"ticker_util", s.bufferStats.TickerChannelUtilization,
			"candle_util", s.bufferStats.CandleChannelUtilization,
			"orderbook_util", s.bufferStats.OrderBookChannelUtilization)
	}
}

// GetBufferStats returns current buffer statistics
func (s *DataProcessorService) GetBufferStats() BufferStats {
	s.bufferMutex.RLock()
	stats := s.bufferStats
	s.bufferMutex.RUnlock()

	return stats
}

// SetAdaptiveBatching enables or disables adaptive batching
func (s *DataProcessorService) SetAdaptiveBatching(enabled bool) {
	s.adaptiveBatching = enabled
	s.logger.Info("Adaptive batching setting changed", "enabled", enabled)
}

// SetPriorityBuffering enables or disables priority buffering
func (s *DataProcessorService) SetPriorityBuffering(enabled bool) {
	s.priorityBuffering = enabled
	s.logger.Info("Priority buffering setting changed", "enabled", enabled)
}

// SetBatchSizeLimits sets the minimum and maximum batch sizes for adaptive batching
func (s *DataProcessorService) SetBatchSizeLimits(min, max int) {
	if min > 0 && max > min {
		s.minBatchSize = min
		s.maxBatchSize = max
		s.logger.Info("Batch size limits updated", "min", min, "max", max)
	}
}

// SetAdaptiveThreshold sets the channel utilization threshold for adaptive batching
func (s *DataProcessorService) SetAdaptiveThreshold(threshold float64) {
	if threshold > 0 && threshold < 1 {
		s.adaptiveThreshold = threshold
		s.logger.Info("Adaptive threshold updated", "threshold", threshold)
	}
}
