package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"m-data-storage/api/dto"
	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/logger"
)

func TestNewSubscriptionHandler(t *testing.T) {
	mockManager := &MockInstrumentManager{}
	cfg := config.LoggingConfig{Level: "debug", Format: "text"}
	testLogger, _ := logger.New(cfg)

	handler := NewSubscriptionHandler(mockManager, testLogger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockManager, handler.instrumentManager)
	assert.Equal(t, testLogger, handler.logger)
}

func TestSubscriptionHandler_CreateSubscription(t *testing.T) {
	// Test successful creation
	t.Run("successful creation", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		req := dto.CreateSubscriptionRequest{
			Symbol:    "BTCUSDT",
			Type:      entities.InstrumentTypeSpot,
			Market:    entities.MarketTypeSpot,
			DataTypes: []entities.DataType{entities.DataTypeTicker, entities.DataTypeCandle},
			StartDate: time.Now(),
			BrokerID:  "test-broker",
		}

		body, _ := json.Marshal(req)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		mockManager.On("AddSubscription", mock.Anything, mock.AnythingOfType("entities.InstrumentSubscription")).Return(nil)

		handler.CreateSubscription(recorder, request)

		assert.Equal(t, http.StatusCreated, recorder.Code)
		mockManager.AssertExpectations(t)
	})

	// Test invalid JSON
	t.Run("invalid JSON", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		request := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBufferString("invalid json"))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		handler.CreateSubscription(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	// Test service error
	t.Run("service error", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		req := dto.CreateSubscriptionRequest{
			Symbol:    "BTCUSDT",
			Type:      entities.InstrumentTypeSpot,
			Market:    entities.MarketTypeSpot,
			DataTypes: []entities.DataType{entities.DataTypeTicker},
			StartDate: time.Now(),
			BrokerID:  "test-broker",
		}

		body, _ := json.Marshal(req)
		request := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBuffer(body))
		request.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()

		mockManager.On("AddSubscription", mock.Anything, mock.AnythingOfType("entities.InstrumentSubscription")).Return(fmt.Errorf("service error"))

		handler.CreateSubscription(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		mockManager.AssertExpectations(t)
	})
}

func TestSubscriptionHandler_GetSubscription(t *testing.T) {
	// Test successful retrieval
	t.Run("successful retrieval", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		subscription := &entities.InstrumentSubscription{
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

		request := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/sub-123", nil)
		recorder := httptest.NewRecorder()

		// Set up router to extract path variables
		router := mux.NewRouter()
		router.HandleFunc("/api/v1/subscriptions/{id}", handler.GetSubscription)

		mockManager.On("GetSubscription", mock.Anything, "sub-123").Return(subscription, nil)

		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockManager.AssertExpectations(t)
	})

	// Test subscription not found
	t.Run("subscription not found", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		request := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/notfound", nil)
		recorder := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/api/v1/subscriptions/{id}", handler.GetSubscription)

		mockManager.On("GetSubscription", mock.Anything, "notfound").Return(nil, fmt.Errorf("not found"))

		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
		mockManager.AssertExpectations(t)
	})
}

func TestSubscriptionHandler_ListSubscriptions(t *testing.T) {
	// Test successful listing
	t.Run("successful listing", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		subscriptions := []entities.InstrumentSubscription{
			{
				ID:        "sub-123",
				Symbol:    "BTCUSDT",
				Type:      entities.InstrumentTypeSpot,
				Market:    entities.MarketTypeSpot,
				DataTypes: []entities.DataType{entities.DataTypeTicker},
				StartDate: time.Now(),
				BrokerID:  "test-broker",
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        "sub-456",
				Symbol:    "ETHUSDT",
				Type:      entities.InstrumentTypeSpot,
				Market:    entities.MarketTypeSpot,
				DataTypes: []entities.DataType{entities.DataTypeCandle},
				StartDate: time.Now(),
				BrokerID:  "test-broker",
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		request := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions", nil)
		recorder := httptest.NewRecorder()

		mockManager.On("ListSubscriptions", mock.Anything).Return(subscriptions, nil)

		handler.ListSubscriptions(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockManager.AssertExpectations(t)
	})

	// Test service error
	t.Run("service error", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		request := httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions", nil)
		recorder := httptest.NewRecorder()

		mockManager.On("ListSubscriptions", mock.Anything).Return([]entities.InstrumentSubscription{}, fmt.Errorf("service error"))

		handler.ListSubscriptions(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		mockManager.AssertExpectations(t)
	})
}

func TestSubscriptionHandler_StartTracking(t *testing.T) {
	// Test successful start tracking
	t.Run("successful start tracking", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		subscription := &entities.InstrumentSubscription{
			ID:        "sub-123",
			Symbol:    "BTCUSDT",
			Type:      entities.InstrumentTypeSpot,
			Market:    entities.MarketTypeSpot,
			DataTypes: []entities.DataType{entities.DataTypeTicker},
			StartDate: time.Now(),
			BrokerID:  "test-broker",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		request := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/sub-123/start", nil)
		recorder := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/api/v1/subscriptions/{id}/start", handler.StartTracking)

		// StartTracking first calls GetSubscription, then StartTracking
		mockManager.On("GetSubscription", mock.Anything, "sub-123").Return(subscription, nil)
		mockManager.On("StartTracking", mock.Anything, "sub-123").Return(nil)

		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockManager.AssertExpectations(t)
	})

	// Test service error
	t.Run("service error", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		subscription := &entities.InstrumentSubscription{
			ID:        "sub-123",
			Symbol:    "BTCUSDT",
			Type:      entities.InstrumentTypeSpot,
			Market:    entities.MarketTypeSpot,
			DataTypes: []entities.DataType{entities.DataTypeTicker},
			StartDate: time.Now(),
			BrokerID:  "test-broker",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		request := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/sub-123/start", nil)
		recorder := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/api/v1/subscriptions/{id}/start", handler.StartTracking)

		// StartTracking first calls GetSubscription, then StartTracking (which fails)
		mockManager.On("GetSubscription", mock.Anything, "sub-123").Return(subscription, nil)
		mockManager.On("StartTracking", mock.Anything, "sub-123").Return(fmt.Errorf("service error"))

		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		mockManager.AssertExpectations(t)
	})
}

func TestSubscriptionHandler_StopTracking(t *testing.T) {
	// Test successful stop tracking
	t.Run("successful stop tracking", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		subscription := &entities.InstrumentSubscription{
			ID:        "sub-123",
			Symbol:    "BTCUSDT",
			Type:      entities.InstrumentTypeSpot,
			Market:    entities.MarketTypeSpot,
			DataTypes: []entities.DataType{entities.DataTypeTicker},
			StartDate: time.Now(),
			BrokerID:  "test-broker",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		request := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/sub-123/stop", nil)
		recorder := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/api/v1/subscriptions/{id}/stop", handler.StopTracking)

		// StopTracking first calls GetSubscription, then StopTracking
		mockManager.On("GetSubscription", mock.Anything, "sub-123").Return(subscription, nil)
		mockManager.On("StopTracking", mock.Anything, "sub-123").Return(nil)

		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockManager.AssertExpectations(t)
	})

	// Test service error
	t.Run("service error", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		subscription := &entities.InstrumentSubscription{
			ID:        "sub-123",
			Symbol:    "BTCUSDT",
			Type:      entities.InstrumentTypeSpot,
			Market:    entities.MarketTypeSpot,
			DataTypes: []entities.DataType{entities.DataTypeTicker},
			StartDate: time.Now(),
			BrokerID:  "test-broker",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		request := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions/sub-123/stop", nil)
		recorder := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/api/v1/subscriptions/{id}/stop", handler.StopTracking)

		// StopTracking first calls GetSubscription, then StopTracking (which fails)
		mockManager.On("GetSubscription", mock.Anything, "sub-123").Return(subscription, nil)
		mockManager.On("StopTracking", mock.Anything, "sub-123").Return(fmt.Errorf("service error"))

		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		mockManager.AssertExpectations(t)
	})
}

func TestSubscriptionHandler_DeleteSubscription(t *testing.T) {
	// Test successful deletion
	t.Run("successful deletion", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		subscription := &entities.InstrumentSubscription{
			ID:        "sub-123",
			Symbol:    "BTCUSDT",
			Type:      entities.InstrumentTypeSpot,
			Market:    entities.MarketTypeSpot,
			DataTypes: []entities.DataType{entities.DataTypeTicker},
			StartDate: time.Now(),
			BrokerID:  "test-broker",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		request := httptest.NewRequest(http.MethodDelete, "/api/v1/subscriptions/sub-123", nil)
		recorder := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/api/v1/subscriptions/{id}", handler.DeleteSubscription)

		// DeleteSubscription first calls GetSubscription, then RemoveSubscription
		mockManager.On("GetSubscription", mock.Anything, "sub-123").Return(subscription, nil)
		mockManager.On("RemoveSubscription", mock.Anything, "sub-123").Return(nil)

		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		mockManager.AssertExpectations(t)
	})

	// Test service error
	t.Run("service error", func(t *testing.T) {
		mockManager := &MockInstrumentManager{}
		cfg := config.LoggingConfig{Level: "debug", Format: "text"}
		testLogger, _ := logger.New(cfg)
		handler := NewSubscriptionHandler(mockManager, testLogger)

		subscription := &entities.InstrumentSubscription{
			ID:        "sub-123",
			Symbol:    "BTCUSDT",
			Type:      entities.InstrumentTypeSpot,
			Market:    entities.MarketTypeSpot,
			DataTypes: []entities.DataType{entities.DataTypeTicker},
			StartDate: time.Now(),
			BrokerID:  "test-broker",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		request := httptest.NewRequest(http.MethodDelete, "/api/v1/subscriptions/sub-123", nil)
		recorder := httptest.NewRecorder()

		router := mux.NewRouter()
		router.HandleFunc("/api/v1/subscriptions/{id}", handler.DeleteSubscription)

		// DeleteSubscription first calls GetSubscription, then RemoveSubscription (which fails)
		mockManager.On("GetSubscription", mock.Anything, "sub-123").Return(subscription, nil)
		mockManager.On("RemoveSubscription", mock.Anything, "sub-123").Return(fmt.Errorf("service error"))

		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
		mockManager.AssertExpectations(t)
	})
}
