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
)

// MockDataPipeline for testing
type MockDataPipeline struct {
	mock.Mock
}

func (m *MockDataPipeline) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDataPipeline) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDataPipeline) AddBroker(ctx context.Context, config interfaces.BrokerConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockDataPipeline) RemoveBroker(ctx context.Context, brokerID string) error {
	args := m.Called(ctx, brokerID)
	return args.Error(0)
}

func (m *MockDataPipeline) Subscribe(ctx context.Context, brokerID string, subscriptions []entities.InstrumentSubscription) error {
	args := m.Called(ctx, brokerID, subscriptions)
	return args.Error(0)
}

func (m *MockDataPipeline) Unsubscribe(ctx context.Context, brokerID string, subscriptions []entities.InstrumentSubscription) error {
	args := m.Called(ctx, brokerID, subscriptions)
	return args.Error(0)
}

func (m *MockDataPipeline) GetStats() interfaces.DataPipelineStats {
	args := m.Called()
	return args.Get(0).(interfaces.DataPipelineStats)
}

func (m *MockDataPipeline) GetIntegrationStats() interfaces.BrokerStorageIntegrationStats {
	args := m.Called()
	return args.Get(0).(interfaces.BrokerStorageIntegrationStats)
}

func (m *MockDataPipeline) Health() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewDataPipelineService(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageIntegration := &MockBrokerStorageIntegration{}
	logger := logrus.New()
	config := DefaultDataPipelineConfig()

	pipeline := NewDataPipelineService(mockBrokerManager, mockStorageIntegration, logger, config)

	assert.NotNil(t, pipeline)
	assert.Equal(t, mockBrokerManager, pipeline.brokerManager)
	assert.Equal(t, mockStorageIntegration, pipeline.storageIntegration)
	assert.Equal(t, logger, pipeline.logger)
	assert.Equal(t, config, pipeline.config)
}

func TestDataPipelineService_StartStop(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageIntegration := &MockBrokerStorageIntegration{}
	logger := logrus.New()
	config := DefaultDataPipelineConfig()
	config.AutoConnectBrokers = false // Disable auto-connect for simplicity

	// Setup mocks
	mockStorageIntegration.On("Start", mock.Anything).Return(nil)
	mockStorageIntegration.On("Stop").Return(nil)

	pipeline := NewDataPipelineService(mockBrokerManager, mockStorageIntegration, logger, config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test start
	err := pipeline.Start(ctx)
	require.NoError(t, err)

	// Check statistics
	stats := pipeline.GetStats()
	assert.False(t, stats.StartedAt.IsZero())

	// Test stop
	err = pipeline.Stop()
	require.NoError(t, err)

	mockStorageIntegration.AssertExpectations(t)
}

func TestDataPipelineService_AddRemoveBroker(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageIntegration := &MockBrokerStorageIntegration{}
	logger := logrus.New()
	config := DefaultDataPipelineConfig()
	config.AutoConnectBrokers = false

	// Create mock broker
	mockBroker := &MockBroker{}
	mockBroker.On("Start", mock.Anything).Return(nil)
	mockBroker.On("Stop").Return(nil).Maybe()
	mockBroker.On("IsConnected").Return(true)
	mockBroker.On("Connect", mock.Anything).Return(nil)

	// Setup mocks
	mockStorageIntegration.On("Start", mock.Anything).Return(nil)
	mockStorageIntegration.On("Stop").Return(nil).Maybe()
	mockBrokerManager.On("AddBroker", mock.Anything, mock.Anything).Return(nil)
	mockBrokerManager.On("GetBroker", "test-broker").Return(mockBroker, nil)
	mockBrokerManager.On("RemoveBroker", mock.Anything, "test-broker").Return(nil)
	mockStorageIntegration.On("AddBroker", "test-broker", mockBroker).Return(nil)
	mockStorageIntegration.On("RemoveBroker", "test-broker").Return(nil)

	pipeline := NewDataPipelineService(mockBrokerManager, mockStorageIntegration, logger, config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start pipeline
	err := pipeline.Start(ctx)
	require.NoError(t, err)
	defer pipeline.Stop()

	// Create broker configuration
	brokerConfig := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	// Test adding broker
	err = pipeline.AddBroker(ctx, brokerConfig)
	require.NoError(t, err)

	// Test removing broker
	err = pipeline.RemoveBroker(ctx, "test-broker")
	require.NoError(t, err)

	mockBrokerManager.AssertExpectations(t)
	mockStorageIntegration.AssertExpectations(t)
}

func TestDataPipelineService_Health(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageIntegration := &MockBrokerStorageIntegration{}
	logger := logrus.New()
	config := DefaultDataPipelineConfig()

	// Setup mocks
	mockStorageIntegration.On("Health").Return(nil)
	mockBrokerManager.On("Health").Return(map[string]error{})

	pipeline := NewDataPipelineService(mockBrokerManager, mockStorageIntegration, logger, config)

	// Test health check
	err := pipeline.Health()
	assert.NoError(t, err)

	mockStorageIntegration.AssertExpectations(t)
	mockBrokerManager.AssertExpectations(t)
}

func TestDataPipelineService_HealthWithErrors(t *testing.T) {
	mockBrokerManager := &MockBrokerManager{}
	mockStorageIntegration := &MockBrokerStorageIntegration{}
	logger := logrus.New()
	config := DefaultDataPipelineConfig()

	// Setup mocks with errors
	mockStorageIntegration.On("Health").Return(assert.AnError)

	pipeline := NewDataPipelineService(mockBrokerManager, mockStorageIntegration, logger, config)

	// Test health check with error
	err := pipeline.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "storage integration health check failed")

	mockStorageIntegration.AssertExpectations(t)
}

func TestDefaultDataPipelineConfig(t *testing.T) {
	config := DefaultDataPipelineConfig()

	assert.True(t, config.AutoConnectBrokers)
	assert.Equal(t, 30*time.Second, config.HealthCheckInterval)
	assert.Equal(t, 10*time.Second, config.OperationTimeout)
	assert.True(t, config.AutoReconnect)
	assert.Equal(t, 5*time.Second, config.ReconnectInterval)
}

// MockBrokerStorageIntegration for testing
type MockBrokerStorageIntegration struct {
	mock.Mock
}

func (m *MockBrokerStorageIntegration) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockBrokerStorageIntegration) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBrokerStorageIntegration) AddBroker(brokerID string, broker interfaces.Broker) error {
	args := m.Called(brokerID, broker)
	return args.Error(0)
}

func (m *MockBrokerStorageIntegration) RemoveBroker(brokerID string) error {
	args := m.Called(brokerID)
	return args.Error(0)
}

func (m *MockBrokerStorageIntegration) GetStats() interfaces.BrokerStorageIntegrationStats {
	args := m.Called()
	return args.Get(0).(interfaces.BrokerStorageIntegrationStats)
}

func (m *MockBrokerStorageIntegration) Health() error {
	args := m.Called()
	return args.Error(0)
}

// MockBroker for testing
type MockBroker struct {
	mock.Mock
}

func (m *MockBroker) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockBroker) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBroker) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockBroker) GetInfo() interfaces.BrokerInfo {
	args := m.Called()
	return args.Get(0).(interfaces.BrokerInfo)
}

func (m *MockBroker) GetSupportedInstruments() []entities.InstrumentInfo {
	args := m.Called()
	return args.Get(0).([]entities.InstrumentInfo)
}

func (m *MockBroker) Subscribe(ctx context.Context, instruments []entities.InstrumentSubscription) error {
	args := m.Called(ctx, instruments)
	return args.Error(0)
}

func (m *MockBroker) Unsubscribe(ctx context.Context, instruments []entities.InstrumentSubscription) error {
	args := m.Called(ctx, instruments)
	return args.Error(0)
}

func (m *MockBroker) GetTickerChannel() <-chan entities.Ticker {
	args := m.Called()
	return args.Get(0).(<-chan entities.Ticker)
}

func (m *MockBroker) GetCandleChannel() <-chan entities.Candle {
	args := m.Called()
	return args.Get(0).(<-chan entities.Candle)
}

func (m *MockBroker) GetOrderBookChannel() <-chan entities.OrderBook {
	args := m.Called()
	return args.Get(0).(<-chan entities.OrderBook)
}

func (m *MockBroker) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockBroker) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBroker) Health() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockBrokerStorageIntegration) GetBrokerStats(brokerID string) (interfaces.BrokerIntegrationStats, error) {
	args := m.Called(brokerID)
	return args.Get(0).(interfaces.BrokerIntegrationStats), args.Error(1)
}
