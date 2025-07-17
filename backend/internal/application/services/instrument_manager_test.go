package services

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// MockMetadataStorage for testing
type MockMetadataStorage struct {
	mock.Mock
}

func (m *MockMetadataStorage) SaveInstrument(ctx context.Context, instrument entities.InstrumentInfo) error {
	args := m.Called(ctx, instrument)
	return args.Error(0)
}

func (m *MockMetadataStorage) GetInstrument(ctx context.Context, symbol string) (*entities.InstrumentInfo, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.InstrumentInfo), args.Error(1)
}

func (m *MockMetadataStorage) ListInstruments(ctx context.Context) ([]entities.InstrumentInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.InstrumentInfo), args.Error(1)
}

func (m *MockMetadataStorage) SaveSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockMetadataStorage) GetSubscription(ctx context.Context, subscriptionID string) (*entities.InstrumentSubscription, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.InstrumentSubscription), args.Error(1)
}

func (m *MockMetadataStorage) ListSubscriptions(ctx context.Context) ([]entities.InstrumentSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.InstrumentSubscription), args.Error(1)
}

func (m *MockMetadataStorage) DeleteSubscription(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

func (m *MockMetadataStorage) UpdateSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockMetadataStorage) Health() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMetadataStorage) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMetadataStorage) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMetadataStorage) Migrate() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMetadataStorage) GetDB() *sql.DB {
	args := m.Called()
	return args.Get(0).(*sql.DB)
}

func (m *MockMetadataStorage) SaveBrokerConfig(ctx context.Context, config interfaces.BrokerConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockMetadataStorage) GetBrokerConfig(ctx context.Context, brokerID string) (*interfaces.BrokerConfig, error) {
	args := m.Called(ctx, brokerID)
	return args.Get(0).(*interfaces.BrokerConfig), args.Error(1)
}

func (m *MockMetadataStorage) ListBrokerConfigs(ctx context.Context) ([]interfaces.BrokerConfig, error) {
	args := m.Called(ctx)
	return args.Get(0).([]interfaces.BrokerConfig), args.Error(1)
}

func (m *MockMetadataStorage) DeleteBrokerConfig(ctx context.Context, brokerID string) error {
	args := m.Called(ctx, brokerID)
	return args.Error(0)
}

func (m *MockMetadataStorage) DeleteInstrument(ctx context.Context, symbol string) error {
	args := m.Called(ctx, symbol)
	return args.Error(0)
}

func TestNewInstrumentManagerService(t *testing.T) {
	mockStorage := &MockMetadataStorage{}
	mockPipeline := &MockDataPipeline{}
	mockValidator := &MockDataValidator{}
	testLogger := logrus.New()

	service := NewInstrumentManagerService(mockStorage, mockPipeline, mockValidator, testLogger)

	assert.NotNil(t, service)
	assert.Equal(t, mockStorage, service.metadataStorage)
	assert.Equal(t, mockPipeline, service.dataPipeline)
	assert.Equal(t, mockValidator, service.validatorService)
	assert.Equal(t, testLogger, service.logger)
	assert.NotNil(t, service.subscriptions)
}

func TestInstrumentManagerService_AddInstrument(t *testing.T) {
	mockStorage := &MockMetadataStorage{}
	mockValidator := &MockDataValidator{}
	logger := logrus.New()

	service := NewInstrumentManagerService(mockStorage, nil, mockValidator, logger)

	instrument := entities.InstrumentInfo{
		Symbol:     "BTCUSDT",
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		Type:       entities.InstrumentTypeSpot,
		Market:     entities.MarketTypeSpot,
		IsActive:   true,
	}

	ctx := context.Background()

	// Test successful addition
	t.Run("successful addition", func(t *testing.T) {
		mockValidator.On("ValidateInstrument", instrument).Return(nil)
		mockStorage.On("SaveInstrument", ctx, instrument).Return(nil)

		err := service.AddInstrument(ctx, instrument)

		assert.NoError(t, err)
		mockValidator.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	// Test validation error
	t.Run("validation error", func(t *testing.T) {
		mockValidator.On("ValidateInstrument", instrument).Return(fmt.Errorf("validation failed"))

		err := service.AddInstrument(ctx, instrument)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "instrument validation failed")
		mockValidator.AssertExpectations(t)
	})

	// Test storage error
	t.Run("storage error", func(t *testing.T) {
		mockValidator.On("ValidateInstrument", instrument).Return(nil)
		mockStorage.On("SaveInstrument", ctx, instrument).Return(fmt.Errorf("storage error"))

		err := service.AddInstrument(ctx, instrument)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save instrument")
		mockValidator.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})
}

func TestInstrumentManagerService_AddInstrument_NilDependencies(t *testing.T) {
	logger := logrus.New()
	service := NewInstrumentManagerService(nil, nil, nil, logger)

	instrument := entities.InstrumentInfo{
		Symbol:     "BTCUSDT",
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		Type:       entities.InstrumentTypeSpot,
		Market:     entities.MarketTypeSpot,
		IsActive:   true,
	}

	ctx := context.Background()

	// Test with nil dependencies - should not fail
	err := service.AddInstrument(ctx, instrument)
	assert.NoError(t, err)
}

func TestInstrumentManagerService_GetInstrument(t *testing.T) {
	mockStorage := &MockMetadataStorage{}
	logger := logrus.New()

	service := NewInstrumentManagerService(mockStorage, nil, nil, logger)

	instrument := &entities.InstrumentInfo{
		Symbol:     "BTCUSDT",
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		Type:       entities.InstrumentTypeSpot,
		Market:     entities.MarketTypeSpot,
		IsActive:   true,
	}

	ctx := context.Background()

	// Test successful retrieval
	t.Run("successful retrieval", func(t *testing.T) {
		mockStorage.On("GetInstrument", ctx, "BTCUSDT").Return(instrument, nil)

		result, err := service.GetInstrument(ctx, "BTCUSDT")

		assert.NoError(t, err)
		assert.Equal(t, instrument, result)
		mockStorage.AssertExpectations(t)
	})

	// Test storage error
	t.Run("storage error", func(t *testing.T) {
		mockStorage.On("GetInstrument", ctx, "NOTFOUND").Return(nil, fmt.Errorf("not found"))

		result, err := service.GetInstrument(ctx, "NOTFOUND")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockStorage.AssertExpectations(t)
	})
}

func TestInstrumentManagerService_GetInstrument_NilStorage(t *testing.T) {
	logger := logrus.New()
	service := NewInstrumentManagerService(nil, nil, nil, logger)

	ctx := context.Background()

	// Test with nil storage - should return error
	result, err := service.GetInstrument(ctx, "BTCUSDT")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "MetadataStorage not available")
}

func TestInstrumentManagerService_AddSubscription(t *testing.T) {
	mockStorage := &MockMetadataStorage{}
	mockValidator := &MockDataValidator{}
	logger := logrus.New()

	service := NewInstrumentManagerService(mockStorage, nil, mockValidator, logger)

	subscription := entities.InstrumentSubscription{
		ID:        "sub-123",
		Symbol:    "BTCUSDT",
		Type:      entities.InstrumentTypeSpot,
		Market:    entities.MarketTypeSpot,
		DataTypes: []entities.DataType{entities.DataTypeTicker, entities.DataTypeCandle},
		StartDate: time.Now(),
		BrokerID:  "test-broker",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	instrument := &entities.InstrumentInfo{
		Symbol:     "BTCUSDT",
		BaseAsset:  "BTC",
		QuoteAsset: "USDT",
		Type:       entities.InstrumentTypeSpot,
		Market:     entities.MarketTypeSpot,
		IsActive:   true,
	}

	ctx := context.Background()

	// Test successful addition
	t.Run("successful addition", func(t *testing.T) {
		mockValidator.On("ValidateSubscription", subscription).Return(nil)
		mockStorage.On("GetInstrument", ctx, "BTCUSDT").Return(instrument, nil)
		mockStorage.On("SaveSubscription", ctx, subscription).Return(nil)

		err := service.AddSubscription(ctx, subscription)

		assert.NoError(t, err)
		mockValidator.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})

	// Test validation error
	t.Run("validation error", func(t *testing.T) {
		mockValidator.On("ValidateSubscription", subscription).Return(fmt.Errorf("validation failed"))

		err := service.AddSubscription(ctx, subscription)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "subscription validation failed")
		mockValidator.AssertExpectations(t)
	})
}

func TestInstrumentManagerService_Health(t *testing.T) {
	mockStorage := &MockMetadataStorage{}
	mockPipeline := &MockDataPipeline{}
	logger := logrus.New()

	service := NewInstrumentManagerService(mockStorage, mockPipeline, nil, logger)

	// Test healthy state
	t.Run("healthy state", func(t *testing.T) {
		mockStorage.On("Health").Return(nil)
		mockPipeline.On("Health").Return(nil)

		err := service.Health()

		assert.NoError(t, err)
		mockStorage.AssertExpectations(t)
		mockPipeline.AssertExpectations(t)
	})

	// Test with nil dependencies
	t.Run("nil dependencies", func(t *testing.T) {
		serviceWithNil := NewInstrumentManagerService(nil, nil, nil, logger)
		err := serviceWithNil.Health()
		assert.NoError(t, err)
	})
}
