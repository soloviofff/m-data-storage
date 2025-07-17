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

	batch := make([]entities.Ticker, 0, s.batchSize)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

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
			if len(batch) >= s.batchSize {
				s.processBatch(batch, "ticker")
				batch = batch[:0]
			}

		case <-ticker.C:
			if len(batch) > 0 {
				s.processBatch(batch, "ticker")
				batch = batch[:0]
			}

		case <-s.ctx.Done():
			return
		}
	}
}

// candleWorker processes candles in batch mode
func (s *DataProcessorService) candleWorker() {
	defer s.wg.Done()

	batch := make([]entities.Candle, 0, s.batchSize)
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
			if len(batch) >= s.batchSize {
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

	batch := make([]entities.OrderBook, 0, s.batchSize)
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
			if len(batch) >= s.batchSize {
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
		s.logger.WithError(err).Error("Failed to save ticker batch")
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
