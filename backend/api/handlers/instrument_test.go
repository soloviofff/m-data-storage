package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"m-data-storage/api/dto"
	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/logger"
)

// MockInstrumentManager for testing
type MockInstrumentManager struct {
	mock.Mock
}

func (m *MockInstrumentManager) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockInstrumentManager) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockInstrumentManager) Health() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockInstrumentManager) AddInstrument(ctx context.Context, instrument entities.InstrumentInfo) error {
	args := m.Called(ctx, instrument)
	return args.Error(0)
}

func (m *MockInstrumentManager) GetInstrument(ctx context.Context, symbol string) (*entities.InstrumentInfo, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.InstrumentInfo), args.Error(1)
}

func (m *MockInstrumentManager) ListInstruments(ctx context.Context) ([]entities.InstrumentInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.InstrumentInfo), args.Error(1)
}

func (m *MockInstrumentManager) AddSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockInstrumentManager) RemoveSubscription(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

func (m *MockInstrumentManager) UpdateSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockInstrumentManager) GetSubscription(ctx context.Context, subscriptionID string) (*entities.InstrumentSubscription, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.InstrumentSubscription), args.Error(1)
}

func (m *MockInstrumentManager) ListSubscriptions(ctx context.Context) ([]entities.InstrumentSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.InstrumentSubscription), args.Error(1)
}

func (m *MockInstrumentManager) SyncWithBrokers(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockInstrumentManager) StartTracking(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

func (m *MockInstrumentManager) StopTracking(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

func TestNewInstrumentHandler(t *testing.T) {
	mockManager := &MockInstrumentManager{}
	cfg := config.LoggingConfig{Level: "debug", Format: "text"}
	testLogger, _ := logger.New(cfg)

	handler := NewInstrumentHandler(mockManager, testLogger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockManager, handler.instrumentManager)
	assert.Equal(t, testLogger, handler.logger)
}

func TestInstrumentHandler_CreateInstrument(t *testing.T) {
	// Test successful creation
	t.Run("successful creation", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewInstrumentHandler(mockManager, testLogger)

		req := dto.CreateInstrumentRequest{
			Symbol:            "BTCUSDT",
			BaseAsset:         "BTC",
			QuoteAsset:        "USDT",
			Type:              entities.InstrumentTypeSpot,
			Market:            entities.MarketTypeSpot,
			IsActive:          true,
			MinPrice:          0.01,
			MaxPrice:          100000.0,
			MinQuantity:       0.001,
			MaxQuantity:       1000.0,
			PricePrecision:    2,
			QuantityPrecision: 3,
		}

		body, _ := json.Marshal(req)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/instruments", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		mockManager.On("AddInstrument", mock.Anything, mock.AnythingOfType("entities.InstrumentInfo")).Return(nil)

		handler.CreateInstrument(recorder, request)

		assert.Equal(t, http.StatusCreated, recorder.Code)
		mockManager.AssertExpectations(t)
	})

	// Test invalid JSON
	t.Run("invalid JSON", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewInstrumentHandler(mockManager, testLogger)

		request := httptest.NewRequest(http.MethodPost, "/api/v1/instruments", bytes.NewBufferString("invalid json"))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.CreateInstrument(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	// Test service error
	t.Run("service error", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewInstrumentHandler(mockManager, testLogger)

		req := dto.CreateInstrumentRequest{
			Symbol:     "BTCUSDT",
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
			Type:       entities.InstrumentTypeSpot,
			Market:     entities.MarketTypeSpot,
		}

		body, _ := json.Marshal(req)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/instruments", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		mockManager.On("AddInstrument", mock.Anything, mock.AnythingOfType("entities.InstrumentInfo")).Return(fmt.Errorf("service error"))

		handler.CreateInstrument(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		mockManager.AssertExpectations(t)
	})
}

func TestInstrumentHandler_GetInstrument(t *testing.T) {
	// Test successful retrieval
	t.Run("successful retrieval", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewInstrumentHandler(mockManager, testLogger)

		instrument := &entities.InstrumentInfo{
			Symbol:     "BTCUSDT",
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
			Type:       entities.InstrumentTypeSpot,
			Market:     entities.MarketTypeSpot,
			IsActive:   true,
		}

		request := httptest.NewRequest(http.MethodGet, "/api/v1/instruments/BTCUSDT", nil)
		recorder := httptest.NewRecorder()

		// Set up router to extract path variables
		router := mux.NewRouter()
		router.HandleFunc("/api/v1/instruments/{symbol}", handler.GetInstrument)

		mockManager.On("GetInstrument", mock.Anything, "BTCUSDT").Return(instrument, nil)

		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockManager.AssertExpectations(t)
	})

	// Test instrument not found
	t.Run("instrument not found", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewInstrumentHandler(mockManager, testLogger)

		request := httptest.NewRequest(http.MethodGet, "/api/v1/instruments/NOTFOUND", nil)
		recorder := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/api/v1/instruments/{symbol}", handler.GetInstrument)

		mockManager.On("GetInstrument", mock.Anything, "NOTFOUND").Return(nil, fmt.Errorf("not found"))

		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		mockManager.AssertExpectations(t)
	})
}

func TestInstrumentHandler_ListInstruments(t *testing.T) {
	// Test successful listing
	t.Run("successful listing", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewInstrumentHandler(mockManager, testLogger)

		instruments := []entities.InstrumentInfo{
			{
				Symbol:     "BTCUSDT",
				BaseAsset:  "BTC",
				QuoteAsset: "USDT",
				Type:       entities.InstrumentTypeSpot,
				Market:     entities.MarketTypeSpot,
				IsActive:   true,
			},
			{
				Symbol:     "ETHUSDT",
				BaseAsset:  "ETH",
				QuoteAsset: "USDT",
				Type:       entities.InstrumentTypeSpot,
				Market:     entities.MarketTypeSpot,
				IsActive:   true,
			},
		}

		request := httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil)
		recorder := httptest.NewRecorder()

		mockManager.On("ListInstruments", mock.Anything).Return(instruments, nil)

		handler.ListInstruments(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockManager.AssertExpectations(t)
	})

	// Test service error
	t.Run("service error", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewInstrumentHandler(mockManager, testLogger)

		request := httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil)
		recorder := httptest.NewRecorder()

		mockManager.On("ListInstruments", mock.Anything).Return([]entities.InstrumentInfo{}, fmt.Errorf("service error"))

		handler.ListInstruments(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		mockManager.AssertExpectations(t)
	})
}
