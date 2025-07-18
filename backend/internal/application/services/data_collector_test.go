package services

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// MockDataCollectorPipeline for testing DataCollector
type MockDataCollectorPipeline struct {
	mock.Mock
}

func (m *MockDataCollectorPipeline) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDataCollectorPipeline) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDataCollectorPipeline) AddBroker(ctx context.Context, config interfaces.BrokerConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockDataCollectorPipeline) RemoveBroker(ctx context.Context, brokerID string) error {
	args := m.Called(ctx, brokerID)
	return args.Error(0)
}

func (m *MockDataCollectorPipeline) Subscribe(ctx context.Context, brokerID string, subscriptions []entities.InstrumentSubscription) error {
	args := m.Called(ctx, brokerID, subscriptions)
	return args.Error(0)
}

func (m *MockDataCollectorPipeline) Unsubscribe(ctx context.Context, brokerID string, subscriptions []entities.InstrumentSubscription) error {
	args := m.Called(ctx, brokerID, subscriptions)
	return args.Error(0)
}

func (m *MockDataCollectorPipeline) GetStats() interfaces.DataPipelineStats {
	args := m.Called()
	return args.Get(0).(interfaces.DataPipelineStats)
}

func (m *MockDataCollectorPipeline) GetIntegrationStats() interfaces.BrokerStorageIntegrationStats {
	args := m.Called()
	return args.Get(0).(interfaces.BrokerStorageIntegrationStats)
}

func (m *MockDataCollectorPipeline) Health() error {
	args := m.Called()
	return args.Error(0)
}

// MockDataCollectorInstrumentManager for testing DataCollector
type MockDataCollectorInstrumentManager struct {
	mock.Mock
}

func (m *MockDataCollectorInstrumentManager) AddSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockDataCollectorInstrumentManager) RemoveSubscription(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

func (m *MockDataCollectorInstrumentManager) GetSubscription(ctx context.Context, subscriptionID string) (*entities.InstrumentSubscription, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.InstrumentSubscription), args.Error(1)
}

func (m *MockDataCollectorInstrumentManager) ListSubscriptions(ctx context.Context) ([]entities.InstrumentSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.InstrumentSubscription), args.Error(1)
}

func (m *MockDataCollectorInstrumentManager) UpdateSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockDataCollectorInstrumentManager) AddInstrument(ctx context.Context, instrument entities.InstrumentInfo) error {
	args := m.Called(ctx, instrument)
	return args.Error(0)
}

func (m *MockDataCollectorInstrumentManager) GetInstrument(ctx context.Context, symbol string) (*entities.InstrumentInfo, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.InstrumentInfo), args.Error(1)
}

func (m *MockDataCollectorInstrumentManager) ListInstruments(ctx context.Context) ([]entities.InstrumentInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.InstrumentInfo), args.Error(1)
}

func (m *MockDataCollectorInstrumentManager) SyncWithBrokers(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDataCollectorInstrumentManager) StartTracking(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

func (m *MockDataCollectorInstrumentManager) StopTracking(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

func (m *MockDataCollectorInstrumentManager) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDataCollectorInstrumentManager) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDataCollectorInstrumentManager) Health() error {
	args := m.Called()
	return args.Error(0)
}

// MockDataCollectorProcessor for testing DataCollector
type MockDataCollectorProcessor struct {
	mock.Mock
}

func (m *MockDataCollectorProcessor) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDataCollectorProcessor) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDataCollectorProcessor) ProcessTicker(ctx context.Context, ticker entities.Ticker) error {
	args := m.Called(ctx, ticker)
	return args.Error(0)
}

func (m *MockDataCollectorProcessor) ProcessCandle(ctx context.Context, candle entities.Candle) error {
	args := m.Called(ctx, candle)
	return args.Error(0)
}

func (m *MockDataCollectorProcessor) ProcessOrderBook(ctx context.Context, orderBook entities.OrderBook) error {
	args := m.Called(ctx, orderBook)
	return args.Error(0)
}

func (m *MockDataCollectorProcessor) ProcessTickerBatch(ctx context.Context, tickers []entities.Ticker) error {
	args := m.Called(ctx, tickers)
	return args.Error(0)
}

func (m *MockDataCollectorProcessor) ProcessCandleBatch(ctx context.Context, candles []entities.Candle) error {
	args := m.Called(ctx, candles)
	return args.Error(0)
}

func (m *MockDataCollectorProcessor) ProcessOrderBookBatch(ctx context.Context, orderBooks []entities.OrderBook) error {
	args := m.Called(ctx, orderBooks)
	return args.Error(0)
}

func (m *MockDataCollectorProcessor) Health() error {
	args := m.Called()
	return args.Error(0)
}

func TestNewDataCollectorService(t *testing.T) {
	mockPipeline := &MockDataCollectorPipeline{}
	mockInstrumentManager := &MockDataCollectorInstrumentManager{}
	mockProcessor := &MockDataCollectorProcessor{}
	logger := logrus.New()

	service := NewDataCollectorService(mockPipeline, mockInstrumentManager, mockProcessor, logger)

	assert.NotNil(t, service)
	assert.Equal(t, mockPipeline, service.dataPipeline)
	assert.Equal(t, mockInstrumentManager, service.instrumentManager)
	assert.Equal(t, mockProcessor, service.dataProcessor)
	assert.Equal(t, logger, service.logger)
	assert.False(t, service.isCollecting)
	assert.Equal(t, 1000, service.config.ChannelBufferSize)
	assert.Equal(t, 3, service.config.WorkerCount)
}

func TestDataCollectorService_StartCollection(t *testing.T) {
	mockPipeline := &MockDataCollectorPipeline{}
	mockInstrumentManager := &MockDataCollectorInstrumentManager{}
	mockProcessor := &MockDataCollectorProcessor{}
	logger := logrus.New()

	service := NewDataCollectorService(mockPipeline, mockInstrumentManager, mockProcessor, logger)

	ctx := context.Background()

	// Setup expectations
	mockProcessor.On("Start", mock.AnythingOfType("*context.cancelCtx")).Return(nil)
	mockPipeline.On("Start", mock.AnythingOfType("*context.cancelCtx")).Return(nil)
	mockProcessor.On("Stop").Return(nil)
	mockPipeline.On("Stop").Return(nil)

	err := service.StartCollection(ctx)

	assert.NoError(t, err)
	assert.True(t, service.IsCollecting())

	// Cleanup
	err = service.StopCollection()
	assert.NoError(t, err)
	assert.False(t, service.IsCollecting())

	mockProcessor.AssertExpectations(t)
	mockPipeline.AssertExpectations(t)
}

func TestDataCollectorService_Subscribe(t *testing.T) {
	mockPipeline := &MockDataCollectorPipeline{}
	mockInstrumentManager := &MockDataCollectorInstrumentManager{}
	mockProcessor := &MockDataCollectorProcessor{}
	logger := logrus.New()

	service := NewDataCollectorService(mockPipeline, mockInstrumentManager, mockProcessor, logger)

	ctx := context.Background()
	brokerID := "test-broker"
	subscription := entities.InstrumentSubscription{
		ID:       "test-sub",
		Symbol:   "BTCUSD",
		BrokerID: brokerID,
	}

	// Setup expectations
	mockInstrumentManager.On("AddSubscription", ctx, subscription).Return(nil)
	mockPipeline.On("Subscribe", ctx, brokerID, []entities.InstrumentSubscription{subscription}).Return(nil)

	err := service.Subscribe(ctx, brokerID, subscription)

	assert.NoError(t, err)

	mockInstrumentManager.AssertExpectations(t)
	mockPipeline.AssertExpectations(t)
}

func TestDataCollectorService_GetCollectionStats(t *testing.T) {
	mockPipeline := &MockDataCollectorPipeline{}
	mockInstrumentManager := &MockDataCollectorInstrumentManager{}
	mockProcessor := &MockDataCollectorProcessor{}
	logger := logrus.New()

	service := NewDataCollectorService(mockPipeline, mockInstrumentManager, mockProcessor, logger)

	// Setup expectations
	mockInstrumentManager.On("ListSubscriptions", mock.AnythingOfType("context.backgroundCtx")).Return(
		[]entities.InstrumentSubscription{
			{ID: "sub1", IsActive: true},
			{ID: "sub2", IsActive: false},
			{ID: "sub3", IsActive: true},
		}, nil)

	stats := service.GetCollectionStats()

	assert.Equal(t, int64(0), stats.TotalTickers)
	assert.Equal(t, int64(0), stats.TotalCandles)
	assert.Equal(t, int64(0), stats.TotalOrderBooks)
	assert.Equal(t, 2, stats.ActiveSubscriptions) // Only active subscriptions
	assert.Equal(t, int64(0), stats.Errors)

	mockInstrumentManager.AssertExpectations(t)
}
