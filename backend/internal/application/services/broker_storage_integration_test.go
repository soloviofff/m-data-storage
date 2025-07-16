package services

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
	"m-data-storage/internal/infrastructure/broker"
)

// MockBrokerManager для тестирования
type MockBrokerManager struct {
	mock.Mock
}

func (m *MockBrokerManager) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockBrokerManager) AddBroker(ctx context.Context, config interfaces.BrokerConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockBrokerManager) RemoveBroker(ctx context.Context, brokerID string) error {
	args := m.Called(ctx, brokerID)
	return args.Error(0)
}

func (m *MockBrokerManager) GetBroker(brokerID string) (interfaces.Broker, error) {
	args := m.Called(brokerID)
	return args.Get(0).(interfaces.Broker), args.Error(1)
}

func (m *MockBrokerManager) GetAllBrokers() map[string]interfaces.Broker {
	args := m.Called()
	return args.Get(0).(map[string]interfaces.Broker)
}

func (m *MockBrokerManager) ListBrokers() []interfaces.BrokerInfo {
	args := m.Called()
	return args.Get(0).([]interfaces.BrokerInfo)
}

func (m *MockBrokerManager) StartAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockBrokerManager) StopAll() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBrokerManager) Health() map[string]error {
	args := m.Called()
	return args.Get(0).(map[string]error)
}

func (m *MockBrokerManager) HealthCheck() map[string]error {
	args := m.Called()
	return args.Get(0).(map[string]error)
}

// MockStorageService для тестирования
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) SaveTicker(ctx context.Context, ticker entities.Ticker) error {
	args := m.Called(ctx, ticker)
	return args.Error(0)
}

func (m *MockStorageService) SaveCandle(ctx context.Context, candle entities.Candle) error {
	args := m.Called(ctx, candle)
	return args.Error(0)
}

func (m *MockStorageService) SaveOrderBook(ctx context.Context, orderBook entities.OrderBook) error {
	args := m.Called(ctx, orderBook)
	return args.Error(0)
}

func (m *MockStorageService) SaveTickers(ctx context.Context, tickers []entities.Ticker) error {
	args := m.Called(ctx, tickers)
	return args.Error(0)
}

func (m *MockStorageService) SaveCandles(ctx context.Context, candles []entities.Candle) error {
	args := m.Called(ctx, candles)
	return args.Error(0)
}

func (m *MockStorageService) SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error {
	args := m.Called(ctx, orderBooks)
	return args.Error(0)
}

func (m *MockStorageService) GetTickers(ctx context.Context, filter interfaces.TickerFilter) ([]entities.Ticker, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Ticker), args.Error(1)
}

func (m *MockStorageService) GetCandles(ctx context.Context, filter interfaces.CandleFilter) ([]entities.Candle, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Candle), args.Error(1)
}

func (m *MockStorageService) GetOrderBooks(ctx context.Context, filter interfaces.OrderBookFilter) ([]entities.OrderBook, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.OrderBook), args.Error(1)
}

func (m *MockStorageService) FlushAll(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorageService) GetStats() interfaces.StorageServiceStats {
	args := m.Called()
	return args.Get(0).(interfaces.StorageServiceStats)
}

func (m *MockStorageService) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewBrokerStorageIntegration(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageService := &MockStorageService{}
	logger := logrus.New()

	integration := NewBrokerStorageIntegrationService(mockBrokerManager, mockStorageService, logger)

	assert.NotNil(t, integration)
	assert.NotNil(t, integration.(*BrokerStorageIntegration).brokerManager)
	assert.NotNil(t, integration.(*BrokerStorageIntegration).storageService)
	assert.NotNil(t, integration.(*BrokerStorageIntegration).logger)
	assert.NotNil(t, integration.(*BrokerStorageIntegration).integrations)
}

func TestBrokerStorageIntegration_StartStop(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageService := &MockStorageService{}
	logger := logrus.New()

	// Настраиваем мок для возврата пустого списка брокеров
	mockBrokerManager.On("GetAllBrokers").Return(map[string]interfaces.Broker{})

	integration := NewBrokerStorageIntegration(mockBrokerManager, mockStorageService, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Тестируем запуск
	err := integration.Start(ctx)
	require.NoError(t, err)

	// Проверяем статистику
	stats := integration.GetStats()
	assert.Equal(t, 0, stats.ActiveBrokers)
	assert.False(t, stats.StartedAt.IsZero())

	// Тестируем остановку
	err = integration.Stop()
	require.NoError(t, err)

	mockBrokerManager.AssertExpectations(t)
}

func TestBrokerStorageIntegration_AddRemoveBroker(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageService := &MockStorageService{}
	logger := logrus.New()

	// Настраиваем мок для возврата пустого списка брокеров
	mockBrokerManager.On("GetAllBrokers").Return(map[string]interfaces.Broker{})

	integration := NewBrokerStorageIntegration(mockBrokerManager, mockStorageService, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Запускаем интеграцию
	err := integration.Start(ctx)
	require.NoError(t, err)

	// Создаем тестовый брокер
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	testBroker := broker.NewMockCryptoBroker(config, logger)

	// Тестируем добавление брокера
	err = integration.AddBroker("test-broker", testBroker)
	require.NoError(t, err)

	// Проверяем статистику
	stats := integration.GetStats()
	assert.Equal(t, 1, stats.ActiveBrokers)

	// Проверяем статистику по брокеру
	brokerStats, err := integration.GetBrokerStats("test-broker")
	require.NoError(t, err)
	assert.Equal(t, "test-broker", brokerStats.BrokerID)
	assert.False(t, brokerStats.StartedAt.IsZero())

	// Тестируем удаление брокера
	err = integration.RemoveBroker("test-broker")
	require.NoError(t, err)

	// Проверяем, что брокер удален
	stats = integration.GetStats()
	assert.Equal(t, 0, stats.ActiveBrokers)

	// Проверяем, что статистика по брокеру недоступна
	_, err = integration.GetBrokerStats("test-broker")
	assert.Error(t, err)

	// Останавливаем интеграцию
	err = integration.Stop()
	require.NoError(t, err)

	mockBrokerManager.AssertExpectations(t)
}

func TestBrokerStorageIntegration_DataProcessing(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageService := &MockStorageService{}
	logger := logrus.New()

	// Настраиваем мок для возврата пустого списка брокеров
	mockBrokerManager.On("GetAllBrokers").Return(map[string]interfaces.Broker{})

	// Настраиваем мок для сохранения данных
	mockStorageService.On("SaveTicker", mock.Anything, mock.AnythingOfType("entities.Ticker")).Return(nil)
	// Candle и OrderBook генерируются с определенной вероятностью, поэтому делаем их опциональными
	mockStorageService.On("SaveCandle", mock.Anything, mock.AnythingOfType("entities.Candle")).Return(nil).Maybe()
	mockStorageService.On("SaveOrderBook", mock.Anything, mock.AnythingOfType("entities.OrderBook")).Return(nil).Maybe()

	integration := NewBrokerStorageIntegration(mockBrokerManager, mockStorageService, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Запускаем интеграцию
	err := integration.Start(ctx)
	require.NoError(t, err)

	// Создаем тестовый брокер
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	testBroker := broker.NewMockCryptoBroker(config, logger)

	// Запускаем брокер для генерации данных
	err = testBroker.Start(ctx)
	require.NoError(t, err)

	// Добавляем брокер в интеграцию
	err = integration.AddBroker("test-broker", testBroker)
	require.NoError(t, err)

	// Подписываемся на данные
	subscriptions := []entities.InstrumentSubscription{
		{
			Symbol: "BTCUSDT",
			Type:   entities.InstrumentTypeSpot,
			Market: entities.MarketTypeSpot,
		},
	}
	err = testBroker.Subscribe(ctx, subscriptions)
	require.NoError(t, err)

	// Ждем обработки данных
	time.Sleep(2 * time.Second)

	// Проверяем статистику
	stats := integration.GetStats()
	assert.Equal(t, 1, stats.ActiveBrokers)

	brokerStats, err := integration.GetBrokerStats("test-broker")
	require.NoError(t, err)
	assert.Equal(t, "test-broker", brokerStats.BrokerID)

	// Останавливаем брокер
	err = testBroker.Stop()
	require.NoError(t, err)

	// Останавливаем интеграцию
	err = integration.Stop()
	require.NoError(t, err)

	mockBrokerManager.AssertExpectations(t)
	mockStorageService.AssertExpectations(t)
}

func TestBrokerStorageIntegration_Health(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageService := &MockStorageService{}
	logger := logrus.New()

	// Настраиваем мок для возврата пустого списка брокеров
	mockBrokerManager.On("GetAllBrokers").Return(map[string]interfaces.Broker{})

	integration := NewBrokerStorageIntegration(mockBrokerManager, mockStorageService, logger)

	// Запускаем интеграцию
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := integration.Start(ctx)
	require.NoError(t, err)
	defer integration.Stop()

	// Тестируем здоровье без активных брокеров
	err = integration.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active broker integrations")

	// Создаем и добавляем тестовый брокер
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	testBroker := broker.NewMockCryptoBroker(config, logger)
	err = integration.AddBroker("test-broker", testBroker)
	require.NoError(t, err)

	// Тестируем здоровье с активным брокером
	err = integration.Health()
	assert.NoError(t, err)

	mockBrokerManager.AssertExpectations(t)
}
