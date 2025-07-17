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

// MockBrokerManager for testing
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

// MockStorageService for testing
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

	// Setup mock to return empty broker list
	mockBrokerManager.On("GetAllBrokers").Return(map[string]interfaces.Broker{})

	integration := NewBrokerStorageIntegration(mockBrokerManager, mockStorageService, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test start
	err := integration.Start(ctx)
	require.NoError(t, err)

	// Check statistics
	stats := integration.GetStats()
	assert.Equal(t, 0, stats.ActiveBrokers)
	assert.False(t, stats.StartedAt.IsZero())

	// Test stop
	err = integration.Stop()
	require.NoError(t, err)

	mockBrokerManager.AssertExpectations(t)
}

func TestBrokerStorageIntegration_AddRemoveBroker(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageService := &MockStorageService{}
	logger := logrus.New()

	// Setup mock to return empty broker list
	mockBrokerManager.On("GetAllBrokers").Return(map[string]interfaces.Broker{})

	integration := NewBrokerStorageIntegration(mockBrokerManager, mockStorageService, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start integration
	err := integration.Start(ctx)
	require.NoError(t, err)

	// Create test broker
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	testBroker := broker.NewMockCryptoBroker(config, logger)

	// Test adding broker
	err = integration.AddBroker("test-broker", testBroker)
	require.NoError(t, err)

	// Check statistics
	stats := integration.GetStats()
	assert.Equal(t, 1, stats.ActiveBrokers)

	// Check broker statistics
	brokerStats, err := integration.GetBrokerStats("test-broker")
	require.NoError(t, err)
	assert.Equal(t, "test-broker", brokerStats.BrokerID)
	assert.False(t, brokerStats.StartedAt.IsZero())

	// Test removing broker
	err = integration.RemoveBroker("test-broker")
	require.NoError(t, err)

	// Check that broker is removed
	stats = integration.GetStats()
	assert.Equal(t, 0, stats.ActiveBrokers)

	// Check that broker statistics are unavailable
	_, err = integration.GetBrokerStats("test-broker")
	assert.Error(t, err)

	// Stop integration
	err = integration.Stop()
	require.NoError(t, err)

	mockBrokerManager.AssertExpectations(t)
}

func TestBrokerStorageIntegration_DataProcessing(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageService := &MockStorageService{}
	logger := logrus.New()

	// Setup mock to return empty broker list
	mockBrokerManager.On("GetAllBrokers").Return(map[string]interfaces.Broker{})

	// Setup mock for data saving
	mockStorageService.On("SaveTicker", mock.Anything, mock.AnythingOfType("entities.Ticker")).Return(nil)
	// Candle and OrderBook are generated with certain probability, so make them optional
	mockStorageService.On("SaveCandle", mock.Anything, mock.AnythingOfType("entities.Candle")).Return(nil).Maybe()
	mockStorageService.On("SaveOrderBook", mock.Anything, mock.AnythingOfType("entities.OrderBook")).Return(nil).Maybe()

	integration := NewBrokerStorageIntegration(mockBrokerManager, mockStorageService, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start integration
	err := integration.Start(ctx)
	require.NoError(t, err)

	// Create test broker
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	testBroker := broker.NewMockCryptoBroker(config, logger)

	// Start broker for data generation
	err = testBroker.Start(ctx)
	require.NoError(t, err)

	// Add broker to integration
	err = integration.AddBroker("test-broker", testBroker)
	require.NoError(t, err)

	// Subscribe to data
	subscriptions := []entities.InstrumentSubscription{
		{
			Symbol: "BTCUSDT",
			Type:   entities.InstrumentTypeSpot,
			Market: entities.MarketTypeSpot,
		},
	}
	err = testBroker.Subscribe(ctx, subscriptions)
	require.NoError(t, err)

	// Wait for data processing
	time.Sleep(2 * time.Second)

	// Check statistics
	stats := integration.GetStats()
	assert.Equal(t, 1, stats.ActiveBrokers)

	brokerStats, err := integration.GetBrokerStats("test-broker")
	require.NoError(t, err)
	assert.Equal(t, "test-broker", brokerStats.BrokerID)

	// Stop broker
	err = testBroker.Stop()
	require.NoError(t, err)

	// Stop integration
	err = integration.Stop()
	require.NoError(t, err)

	mockBrokerManager.AssertExpectations(t)
	mockStorageService.AssertExpectations(t)
}

func TestBrokerStorageIntegration_Health(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageService := &MockStorageService{}
	logger := logrus.New()

	// Setup mock to return empty broker list
	mockBrokerManager.On("GetAllBrokers").Return(map[string]interfaces.Broker{})

	integration := NewBrokerStorageIntegration(mockBrokerManager, mockStorageService, logger)

	// Start integration
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := integration.Start(ctx)
	require.NoError(t, err)
	defer integration.Stop()

	// Test health without active brokers
	err = integration.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no active broker integrations")

	// Create and add test broker
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

	// Test health with active broker
	err = integration.Health()
	assert.NoError(t, err)

	mockBrokerManager.AssertExpectations(t)
}
