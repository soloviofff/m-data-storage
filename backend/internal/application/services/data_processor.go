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

// DataProcessorService реализует интерфейс DataProcessor
type DataProcessorService struct {
	storage   interfaces.StorageManager
	validator interfaces.DataValidator
	logger    *logrus.Logger

	// Каналы для пакетной обработки
	tickerChan    chan entities.Ticker
	candleChan    chan entities.Candle
	orderBookChan chan entities.OrderBook

	// Настройки пакетной обработки
	batchSize     int
	flushInterval time.Duration

	// Управление жизненным циклом
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Статистика
	stats ProcessorStats
	mutex sync.RWMutex
}

// ProcessorStats - статистика процессора данных
type ProcessorStats struct {
	ProcessedTickers    int64     `json:"processed_tickers"`
	ProcessedCandles    int64     `json:"processed_candles"`
	ProcessedOrderBooks int64     `json:"processed_orderbooks"`
	Errors              int64     `json:"errors"`
	LastProcessed       time.Time `json:"last_processed"`
}

// NewDataProcessorService создает новый сервис обработки данных
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

// Start запускает сервис обработки данных
func (s *DataProcessorService) Start(ctx context.Context) error {
	s.ctx, s.cancel = context.WithCancel(ctx)

	// Запускаем воркеры для пакетной обработки
	s.wg.Add(3)
	go s.tickerWorker()
	go s.candleWorker()
	go s.orderBookWorker()

	s.logger.Info("Data processor service started")
	return nil
}

// Stop останавливает сервис обработки данных
func (s *DataProcessorService) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}

	// Закрываем каналы
	close(s.tickerChan)
	close(s.candleChan)
	close(s.orderBookChan)

	// Ждем завершения воркеров
	s.wg.Wait()

	s.logger.Info("Data processor service stopped")
	return nil
}

// ProcessTicker обрабатывает один тикер
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

// ProcessCandle обрабатывает одну свечу
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

// ProcessOrderBook обрабатывает один ордербук
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

// ProcessTickerBatch обрабатывает пакет тикеров
func (s *DataProcessorService) ProcessTickerBatch(ctx context.Context, tickers []entities.Ticker) error {
	// Валидируем все тикеры
	for _, ticker := range tickers {
		if err := s.validator.ValidateTicker(ticker); err != nil {
			s.incrementErrors()
			return errors.Wrap(err, "ticker validation failed")
		}
	}

	// Сохраняем в хранилище
	if err := s.storage.SaveTickers(ctx, tickers); err != nil {
		s.incrementErrors()
		return errors.Wrap(err, "failed to save tickers")
	}

	s.updateStats(len(tickers), 0, 0)
	return nil
}

// ProcessCandleBatch обрабатывает пакет свечей
func (s *DataProcessorService) ProcessCandleBatch(ctx context.Context, candles []entities.Candle) error {
	// Валидируем все свечи
	for _, candle := range candles {
		if err := s.validator.ValidateCandle(candle); err != nil {
			s.incrementErrors()
			return errors.Wrap(err, "candle validation failed")
		}
	}

	// Сохраняем в хранилище
	if err := s.storage.SaveCandles(ctx, candles); err != nil {
		s.incrementErrors()
		return errors.Wrap(err, "failed to save candles")
	}

	s.updateStats(0, len(candles), 0)
	return nil
}

// ProcessOrderBookBatch обрабатывает пакет ордербуков
func (s *DataProcessorService) ProcessOrderBookBatch(ctx context.Context, orderBooks []entities.OrderBook) error {
	// Валидируем все ордербуки
	for _, orderBook := range orderBooks {
		if err := s.validator.ValidateOrderBook(orderBook); err != nil {
			s.incrementErrors()
			return errors.Wrap(err, "orderbook validation failed")
		}
	}

	// Сохраняем в хранилище
	if err := s.storage.SaveOrderBooks(ctx, orderBooks); err != nil {
		s.incrementErrors()
		return errors.Wrap(err, "failed to save orderbooks")
	}

	s.updateStats(0, 0, len(orderBooks))
	return nil
}

// Health проверяет здоровье сервиса
func (s *DataProcessorService) Health() error {
	// Проверяем состояние каналов
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

// GetStats возвращает статистику процессора
func (s *DataProcessorService) GetStats() ProcessorStats {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.stats
}

// tickerWorker обрабатывает тикеры в пакетном режиме
func (s *DataProcessorService) tickerWorker() {
	defer s.wg.Done()

	batch := make([]entities.Ticker, 0, s.batchSize)
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case data, ok := <-s.tickerChan:
			if !ok {
				// Канал закрыт, обрабатываем оставшиеся данные
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

// candleWorker обрабатывает свечи в пакетном режиме
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

// orderBookWorker обрабатывает ордербуки в пакетном режиме
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

// processBatch обрабатывает пакет тикеров
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

// processCandleBatch обрабатывает пакет свечей
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

// processOrderBookBatch обрабатывает пакет ордербуков
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

// updateStats обновляет статистику
func (s *DataProcessorService) updateStats(tickers, candles, orderBooks int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.stats.ProcessedTickers += int64(tickers)
	s.stats.ProcessedCandles += int64(candles)
	s.stats.ProcessedOrderBooks += int64(orderBooks)
	s.stats.LastProcessed = time.Now()
}

// incrementErrors увеличивает счетчик ошибок
func (s *DataProcessorService) incrementErrors() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stats.Errors++
}
