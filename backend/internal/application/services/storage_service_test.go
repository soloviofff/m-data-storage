package services

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// MockStorageManager для тестирования
type MockStorageManager struct {
	mock.Mock
}

func (m *MockStorageManager) SaveTickers(ctx context.Context, tickers []entities.Ticker) error {
	args := m.Called(ctx, tickers)
	return args.Error(0)
}

func (m *MockStorageManager) SaveCandles(ctx context.Context, candles []entities.Candle) error {
	args := m.Called(ctx, candles)
	return args.Error(0)
}

func (m *MockStorageManager) SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error {
	args := m.Called(ctx, orderBooks)
	return args.Error(0)
}

func (m *MockStorageManager) GetTickers(ctx context.Context, filter interfaces.TickerFilter) ([]entities.Ticker, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Ticker), args.Error(1)
}

func (m *MockStorageManager) GetCandles(ctx context.Context, filter interfaces.CandleFilter) ([]entities.Candle, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Candle), args.Error(1)
}

func (m *MockStorageManager) GetOrderBooks(ctx context.Context, filter interfaces.OrderBookFilter) ([]entities.OrderBook, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.OrderBook), args.Error(1)
}

func (m *MockStorageManager) GetMetadataStorage() interfaces.MetadataStorage {
	args := m.Called()
	return args.Get(0).(interfaces.MetadataStorage)
}

func (m *MockStorageManager) GetTimeSeriesStorage() interfaces.TimeSeriesStorage {
	args := m.Called()
	return args.Get(0).(interfaces.TimeSeriesStorage)
}

func (m *MockStorageManager) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorageManager) Shutdown() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStorageManager) Health() map[string]error {
	args := m.Called()
	return args.Get(0).(map[string]error)
}

// Методы миграций не входят в интерфейс StorageManager, поэтому убираем их

// MockDataValidator для тестирования
type MockDataValidator struct {
	mock.Mock
}

func (m *MockDataValidator) ValidateTicker(ticker entities.Ticker) error {
	args := m.Called(ticker)
	return args.Error(0)
}

func (m *MockDataValidator) ValidateCandle(candle entities.Candle) error {
	args := m.Called(candle)
	return args.Error(0)
}

func (m *MockDataValidator) ValidateOrderBook(orderBook entities.OrderBook) error {
	args := m.Called(orderBook)
	return args.Error(0)
}

func (m *MockDataValidator) ValidateTimeframe(timeframe string) error {
	args := m.Called(timeframe)
	return args.Error(0)
}

func (m *MockDataValidator) ValidateMarketType(marketType string) error {
	args := m.Called(marketType)
	return args.Error(0)
}

func (m *MockDataValidator) ValidateInstrumentType(instrumentType string) error {
	args := m.Called(instrumentType)
	return args.Error(0)
}

func (m *MockDataValidator) ValidateInstrument(instrument entities.InstrumentInfo) error {
	args := m.Called(instrument)
	return args.Error(0)
}

func (m *MockDataValidator) ValidateSubscription(subscription entities.InstrumentSubscription) error {
	args := m.Called(subscription)
	return args.Error(0)
}

func TestNewStorageService(t *testing.T) {
	mockStorage := &MockStorageManager{}
	mockValidator := &MockDataValidator{}
	logger := logrus.New()
	config := DefaultStorageServiceConfig()

	service := NewStorageService(mockStorage, mockValidator, logger, config)

	assert.NotNil(t, service)
	assert.Equal(t, mockStorage, service.storageManager)
	assert.Equal(t, mockValidator, service.validator)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, config.BatchSize, service.batchSize)
	assert.Equal(t, config.FlushInterval, service.flushInterval)

	// Закрываем сервис
	ctx := context.Background()
	service.Close(ctx)
}

func TestStorageService_SaveTicker(t *testing.T) {
	mockStorage := &MockStorageManager{}
	mockValidator := &MockDataValidator{}
	logger := logrus.New()
	config := StorageServiceConfig{
		BatchSize:     2,             // Маленький размер для тестирования
		FlushInterval: 1 * time.Hour, // Большой интервал, чтобы не мешал тестам
	}

	service := NewStorageService(mockStorage, mockValidator, logger, config)
	defer service.Close(context.Background())

	ctx := context.Background()
	ticker := entities.Ticker{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		Volume:    1.5,
		Timestamp: time.Now(),
		BrokerID:  "binance",
	}

	// Настраиваем моки
	mockValidator.On("ValidateTicker", ticker).Return(nil)

	// Первый тикер - должен добавиться в буфер
	err := service.SaveTicker(ctx, ticker)
	assert.NoError(t, err)

	// Второй тикер - должен вызвать флаш
	mockStorage.On("SaveTickers", ctx, mock.MatchedBy(func(tickers []entities.Ticker) bool {
		return len(tickers) == 2
	})).Return(nil)

	err = service.SaveTicker(ctx, ticker)
	assert.NoError(t, err)

	// Проверяем, что моки были вызваны
	mockValidator.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestStorageService_SaveTickers(t *testing.T) {
	mockStorage := &MockStorageManager{}
	mockValidator := &MockDataValidator{}
	logger := logrus.New()
	config := DefaultStorageServiceConfig()

	service := NewStorageService(mockStorage, mockValidator, logger, config)
	defer service.Close(context.Background())

	ctx := context.Background()
	tickers := []entities.Ticker{
		{
			Symbol:    "BTCUSDT",
			Price:     50000.0,
			Volume:    1.5,
			Timestamp: time.Now(),
			BrokerID:  "binance",
		},
		{
			Symbol:    "ETHUSDT",
			Price:     3000.0,
			Volume:    2.5,
			Timestamp: time.Now(),
			BrokerID:  "binance",
		},
	}

	// Настраиваем моки
	for _, ticker := range tickers {
		mockValidator.On("ValidateTicker", ticker).Return(nil)
	}
	mockStorage.On("SaveTickers", ctx, tickers).Return(nil)

	err := service.SaveTickers(ctx, tickers)
	assert.NoError(t, err)

	// Проверяем статистику
	stats := service.GetStats()
	assert.Equal(t, int64(2), stats.TickersSaved)

	// Проверяем, что моки были вызваны
	mockValidator.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestStorageService_FlushAll(t *testing.T) {
	mockStorage := &MockStorageManager{}
	mockValidator := &MockDataValidator{}
	logger := logrus.New()
	config := StorageServiceConfig{
		BatchSize:     10, // Большой размер, чтобы не было автофлаша
		FlushInterval: 1 * time.Hour,
	}

	service := NewStorageService(mockStorage, mockValidator, logger, config)
	defer service.Close(context.Background())

	ctx := context.Background()
	ticker := entities.Ticker{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		Volume:    1.5,
		Timestamp: time.Now(),
		BrokerID:  "binance",
	}

	// Настраиваем моки
	mockValidator.On("ValidateTicker", ticker).Return(nil)
	mockStorage.On("SaveTickers", ctx, mock.MatchedBy(func(tickers []entities.Ticker) bool {
		return len(tickers) == 1
	})).Return(nil)

	// Добавляем тикер в буфер
	err := service.SaveTicker(ctx, ticker)
	assert.NoError(t, err)

	// Принудительно сбрасываем буферы
	err = service.FlushAll(ctx)
	assert.NoError(t, err)

	// Проверяем, что моки были вызваны
	mockValidator.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

func TestStorageService_GetStats(t *testing.T) {
	mockStorage := &MockStorageManager{}
	mockValidator := &MockDataValidator{}
	logger := logrus.New()
	config := DefaultStorageServiceConfig()

	service := NewStorageService(mockStorage, mockValidator, logger, config)
	defer service.Close(context.Background())

	// Проверяем начальную статистику
	stats := service.GetStats()
	assert.Equal(t, int64(0), stats.TickersSaved)
	assert.Equal(t, int64(0), stats.CandlesSaved)
	assert.Equal(t, int64(0), stats.OrderBooksSaved)
	assert.Equal(t, int64(0), stats.ErrorsCount)
}
