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

// StorageService предоставляет высокоуровневые операции для работы с хранилищем данных
type StorageService struct {
	storageManager interfaces.StorageManager
	validator      interfaces.DataValidator
	logger         *logrus.Logger

	// Настройки пакетной обработки
	batchSize     int
	flushInterval time.Duration

	// Буферы для пакетной обработки
	tickerBuffer    []entities.Ticker
	candleBuffer    []entities.Candle
	orderBookBuffer []entities.OrderBook

	// Мьютексы для защиты буферов
	tickerMutex    sync.Mutex
	candleMutex    sync.Mutex
	orderBookMutex sync.Mutex

	// Каналы для уведомлений о флаше
	flushTicker    *time.Ticker
	stopChan       chan struct{}
	flushWaitGroup sync.WaitGroup

	// Статистика
	stats StorageStats
	mutex sync.RWMutex
}

// StorageStats содержит статистику работы сервиса хранения
type StorageStats = interfaces.StorageServiceStats

// StorageServiceConfig содержит конфигурацию для StorageService
type StorageServiceConfig struct {
	BatchSize     int           `yaml:"batch_size" json:"batch_size"`
	FlushInterval time.Duration `yaml:"flush_interval" json:"flush_interval"`
}

// DefaultStorageServiceConfig возвращает конфигурацию по умолчанию
func DefaultStorageServiceConfig() StorageServiceConfig {
	return StorageServiceConfig{
		BatchSize:     1000,
		FlushInterval: 5 * time.Second,
	}
}

// NewStorageService создает новый экземпляр StorageService
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

	// Запускаем периодический флаш
	service.startPeriodicFlush()

	return service
}

// SaveTicker сохраняет тикер с пакетной обработкой
func (s *StorageService) SaveTicker(ctx context.Context, ticker entities.Ticker) error {
	// Валидируем тикер
	if err := s.validator.ValidateTicker(ticker); err != nil {
		s.incrementErrors()
		return fmt.Errorf("ticker validation failed: %w", err)
	}

	s.tickerMutex.Lock()
	defer s.tickerMutex.Unlock()

	s.tickerBuffer = append(s.tickerBuffer, ticker)

	// Если буфер заполнен, сбрасываем его
	if len(s.tickerBuffer) >= s.batchSize {
		return s.flushTickersUnsafe(ctx)
	}

	return nil
}

// SaveCandle сохраняет свечу с пакетной обработкой
func (s *StorageService) SaveCandle(ctx context.Context, candle entities.Candle) error {
	// Валидируем свечу
	if err := s.validator.ValidateCandle(candle); err != nil {
		s.incrementErrors()
		return fmt.Errorf("candle validation failed: %w", err)
	}

	s.candleMutex.Lock()
	defer s.candleMutex.Unlock()

	s.candleBuffer = append(s.candleBuffer, candle)

	// Если буфер заполнен, сбрасываем его
	if len(s.candleBuffer) >= s.batchSize {
		return s.flushCandlesUnsafe(ctx)
	}

	return nil
}

// SaveOrderBook сохраняет ордербук с пакетной обработкой
func (s *StorageService) SaveOrderBook(ctx context.Context, orderBook entities.OrderBook) error {
	// Валидируем ордербук
	if err := s.validator.ValidateOrderBook(orderBook); err != nil {
		s.incrementErrors()
		return fmt.Errorf("orderbook validation failed: %w", err)
	}

	s.orderBookMutex.Lock()
	defer s.orderBookMutex.Unlock()

	s.orderBookBuffer = append(s.orderBookBuffer, orderBook)

	// Если буфер заполнен, сбрасываем его
	if len(s.orderBookBuffer) >= s.batchSize {
		return s.flushOrderBooksUnsafe(ctx)
	}

	return nil
}

// SaveTickers сохраняет множественные тикеры
func (s *StorageService) SaveTickers(ctx context.Context, tickers []entities.Ticker) error {
	// Валидируем все тикеры
	for i, ticker := range tickers {
		if err := s.validator.ValidateTicker(ticker); err != nil {
			s.incrementErrors()
			return fmt.Errorf("ticker validation failed at index %d: %w", i, err)
		}
	}

	// Сохраняем напрямую в storage manager
	if err := s.storageManager.SaveTickers(ctx, tickers); err != nil {
		s.incrementErrors()
		return fmt.Errorf("failed to save tickers: %w", err)
	}

	s.updateStats(int64(len(tickers)), 0, 0)
	return nil
}

// SaveCandles сохраняет множественные свечи
func (s *StorageService) SaveCandles(ctx context.Context, candles []entities.Candle) error {
	// Валидируем все свечи
	for i, candle := range candles {
		if err := s.validator.ValidateCandle(candle); err != nil {
			s.incrementErrors()
			return fmt.Errorf("candle validation failed at index %d: %w", i, err)
		}
	}

	// Сохраняем напрямую в storage manager
	if err := s.storageManager.SaveCandles(ctx, candles); err != nil {
		s.incrementErrors()
		return fmt.Errorf("failed to save candles: %w", err)
	}

	s.updateStats(0, int64(len(candles)), 0)
	return nil
}

// SaveOrderBooks сохраняет множественные ордербуки
func (s *StorageService) SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error {
	// Валидируем все ордербуки
	for i, orderBook := range orderBooks {
		if err := s.validator.ValidateOrderBook(orderBook); err != nil {
			s.incrementErrors()
			return fmt.Errorf("orderbook validation failed at index %d: %w", i, err)
		}
	}

	// Сохраняем напрямую в storage manager
	if err := s.storageManager.SaveOrderBooks(ctx, orderBooks); err != nil {
		s.incrementErrors()
		return fmt.Errorf("failed to save orderbooks: %w", err)
	}

	s.updateStats(0, 0, int64(len(orderBooks)))
	return nil
}

// GetTickers получает тикеры с фильтрацией
func (s *StorageService) GetTickers(ctx context.Context, filter interfaces.TickerFilter) ([]entities.Ticker, error) {
	return s.storageManager.GetTickers(ctx, filter)
}

// GetCandles получает свечи с фильтрацией
func (s *StorageService) GetCandles(ctx context.Context, filter interfaces.CandleFilter) ([]entities.Candle, error) {
	return s.storageManager.GetCandles(ctx, filter)
}

// GetOrderBooks получает ордербуки с фильтрацией
func (s *StorageService) GetOrderBooks(ctx context.Context, filter interfaces.OrderBookFilter) ([]entities.OrderBook, error) {
	return s.storageManager.GetOrderBooks(ctx, filter)
}

// FlushAll принудительно сбрасывает все буферы
func (s *StorageService) FlushAll(ctx context.Context) error {
	var errors []error

	// Сбрасываем тикеры
	s.tickerMutex.Lock()
	if len(s.tickerBuffer) > 0 {
		if err := s.flushTickersUnsafe(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to flush tickers: %w", err))
		}
	}
	s.tickerMutex.Unlock()

	// Сбрасываем свечи
	s.candleMutex.Lock()
	if len(s.candleBuffer) > 0 {
		if err := s.flushCandlesUnsafe(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to flush candles: %w", err))
		}
	}
	s.candleMutex.Unlock()

	// Сбрасываем ордербуки
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

// GetStats возвращает статистику сервиса
func (s *StorageService) GetStats() StorageStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.stats
}

// Close закрывает сервис и сбрасывает все буферы
func (s *StorageService) Close(ctx context.Context) error {
	// Останавливаем периодический флаш
	close(s.stopChan)
	if s.flushTicker != nil {
		s.flushTicker.Stop()
	}

	// Ждем завершения всех операций флаша
	s.flushWaitGroup.Wait()

	// Сбрасываем все оставшиеся данные
	return s.FlushAll(ctx)
}

// startPeriodicFlush запускает периодический сброс буферов
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

// flushTickersUnsafe сбрасывает буфер тикеров (должен вызываться под мьютексом)
func (s *StorageService) flushTickersUnsafe(ctx context.Context) error {
	if len(s.tickerBuffer) == 0 {
		return nil
	}

	if err := s.storageManager.SaveTickers(ctx, s.tickerBuffer); err != nil {
		s.incrementErrors()
		return fmt.Errorf("failed to save tickers batch: %w", err)
	}

	count := int64(len(s.tickerBuffer))
	s.tickerBuffer = s.tickerBuffer[:0] // Очищаем буфер
	s.updateStats(count, 0, 0)
	s.incrementBatchesFlashed()

	s.logger.WithField("count", count).Debug("Flushed tickers batch")
	return nil
}

// flushCandlesUnsafe сбрасывает буфер свечей (должен вызываться под мьютексом)
func (s *StorageService) flushCandlesUnsafe(ctx context.Context) error {
	if len(s.candleBuffer) == 0 {
		return nil
	}

	if err := s.storageManager.SaveCandles(ctx, s.candleBuffer); err != nil {
		s.incrementErrors()
		return fmt.Errorf("failed to save candles batch: %w", err)
	}

	count := int64(len(s.candleBuffer))
	s.candleBuffer = s.candleBuffer[:0] // Очищаем буфер
	s.updateStats(0, count, 0)
	s.incrementBatchesFlashed()

	s.logger.WithField("count", count).Debug("Flushed candles batch")
	return nil
}

// flushOrderBooksUnsafe сбрасывает буфер ордербуков (должен вызываться под мьютексом)
func (s *StorageService) flushOrderBooksUnsafe(ctx context.Context) error {
	if len(s.orderBookBuffer) == 0 {
		return nil
	}

	if err := s.storageManager.SaveOrderBooks(ctx, s.orderBookBuffer); err != nil {
		s.incrementErrors()
		return fmt.Errorf("failed to save orderbooks batch: %w", err)
	}

	count := int64(len(s.orderBookBuffer))
	s.orderBookBuffer = s.orderBookBuffer[:0] // Очищаем буфер
	s.updateStats(0, 0, count)
	s.incrementBatchesFlashed()

	s.logger.WithField("count", count).Debug("Flushed orderbooks batch")
	return nil
}

// updateStats обновляет статистику
func (s *StorageService) updateStats(tickers, candles, orderBooks int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stats.TickersSaved += tickers
	s.stats.CandlesSaved += candles
	s.stats.OrderBooksSaved += orderBooks
	s.stats.LastFlushTime = time.Now()
}

// incrementErrors увеличивает счетчик ошибок
func (s *StorageService) incrementErrors() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats.ErrorsCount++
}

// incrementBatchesFlashed увеличивает счетчик сброшенных пакетов
func (s *StorageService) incrementBatchesFlashed() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats.BatchesFlashed++
}
