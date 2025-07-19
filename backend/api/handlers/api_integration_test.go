package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"m-data-storage/api/dto"
	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/logger"
)

// MockInstrumentManagerAPI for integration tests
type MockInstrumentManagerAPI struct {
	mock.Mock
}

func (m *MockInstrumentManagerAPI) ListInstruments(ctx context.Context) ([]entities.InstrumentInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.InstrumentInfo), args.Error(1)
}

func (m *MockInstrumentManagerAPI) GetInstrument(ctx context.Context, symbol string) (*entities.InstrumentInfo, error) {
	args := m.Called(ctx, symbol)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.InstrumentInfo), args.Error(1)
}

func (m *MockInstrumentManagerAPI) AddInstrument(ctx context.Context, instrument entities.InstrumentInfo) error {
	args := m.Called(ctx, instrument)
	return args.Error(0)
}

func (m *MockInstrumentManagerAPI) UpdateInstrument(ctx context.Context, symbol string, instrument entities.InstrumentInfo) error {
	args := m.Called(ctx, symbol, instrument)
	return args.Error(0)
}

func (m *MockInstrumentManagerAPI) DeleteInstrument(ctx context.Context, symbol string) error {
	args := m.Called(ctx, symbol)
	return args.Error(0)
}

func (m *MockInstrumentManagerAPI) GetSubscriptions(ctx context.Context, filter interfaces.SubscriptionFilter) ([]entities.InstrumentSubscription, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.InstrumentSubscription), args.Error(1)
}

func (m *MockInstrumentManagerAPI) Subscribe(ctx context.Context, subscription entities.InstrumentSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockInstrumentManagerAPI) Unsubscribe(ctx context.Context, symbol, brokerID string) error {
	args := m.Called(ctx, symbol, brokerID)
	return args.Error(0)
}

func (m *MockInstrumentManagerAPI) Health() error {
	args := m.Called()
	return args.Error(0)
}

// Additional methods required by InstrumentManager interface
func (m *MockInstrumentManagerAPI) AddSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockInstrumentManagerAPI) RemoveSubscription(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

func (m *MockInstrumentManagerAPI) UpdateSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	args := m.Called(ctx, subscription)
	return args.Error(0)
}

func (m *MockInstrumentManagerAPI) GetSubscription(ctx context.Context, subscriptionID string) (*entities.InstrumentSubscription, error) {
	args := m.Called(ctx, subscriptionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.InstrumentSubscription), args.Error(1)
}

func (m *MockInstrumentManagerAPI) ListSubscriptions(ctx context.Context) ([]entities.InstrumentSubscription, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.InstrumentSubscription), args.Error(1)
}

func (m *MockInstrumentManagerAPI) SyncWithBrokers(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockInstrumentManagerAPI) StartTracking(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

func (m *MockInstrumentManagerAPI) StopTracking(ctx context.Context, subscriptionID string) error {
	args := m.Called(ctx, subscriptionID)
	return args.Error(0)
}

func (m *MockInstrumentManagerAPI) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockInstrumentManagerAPI) Stop() error {
	args := m.Called()
	return args.Error(0)
}

// MockDataQueryAPI for integration tests
type MockDataQueryAPI struct {
	mock.Mock
}

func (m *MockDataQueryAPI) GetTickers(ctx context.Context, filter interfaces.TickerFilter) ([]entities.Ticker, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Ticker), args.Error(1)
}

func (m *MockDataQueryAPI) GetCandles(ctx context.Context, filter interfaces.CandleFilter) ([]entities.Candle, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Candle), args.Error(1)
}

func (m *MockDataQueryAPI) GetOrderBooks(ctx context.Context, filter interfaces.OrderBookFilter) ([]entities.OrderBook, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.OrderBook), args.Error(1)
}

func (m *MockDataQueryAPI) Health() error {
	args := m.Called()
	return args.Error(0)
}

// Additional methods required by DataQuery interface
func (m *MockDataQueryAPI) GetTickerAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]interfaces.TickerAggregate, error) {
	args := m.Called(ctx, symbol, interval, startTime, endTime)
	return args.Get(0).([]interfaces.TickerAggregate), args.Error(1)
}

func (m *MockDataQueryAPI) GetCandleAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]interfaces.CandleAggregate, error) {
	args := m.Called(ctx, symbol, interval, startTime, endTime)
	return args.Get(0).([]interfaces.CandleAggregate), args.Error(1)
}

func (m *MockDataQueryAPI) GetDataStats(ctx context.Context) (interfaces.DataStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(interfaces.DataStats), args.Error(1)
}

// setupTestRouter creates a test router with all handlers
func setupTestRouter(instrumentManager interfaces.InstrumentManager, dataQuery interfaces.DataQuery) *mux.Router {
	router := mux.NewRouter()

	// Create logger
	testLogger, _ := logger.New(config.LoggingConfig{
		Level:  "debug",
		Format: "text",
	})

	// Create handlers
	instrumentHandler := NewInstrumentHandler(instrumentManager, testLogger)
	subscriptionHandler := NewSubscriptionHandler(instrumentManager, testLogger)
	dataHandler := NewDataHandler(dataQuery, testLogger)

	// Setup routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Instrument routes
	instruments := api.PathPrefix("/instruments").Subrouter()
	instruments.HandleFunc("", instrumentHandler.ListInstruments).Methods("GET")
	instruments.HandleFunc("", instrumentHandler.CreateInstrument).Methods("POST")
	instruments.HandleFunc("/{symbol}", instrumentHandler.GetInstrument).Methods("GET")
	instruments.HandleFunc("/{symbol}", instrumentHandler.UpdateInstrument).Methods("PUT")
	instruments.HandleFunc("/{symbol}", instrumentHandler.DeleteInstrument).Methods("DELETE")

	// Subscription routes
	subscriptions := api.PathPrefix("/subscriptions").Subrouter()
	subscriptions.HandleFunc("", subscriptionHandler.ListSubscriptions).Methods("GET")
	subscriptions.HandleFunc("", subscriptionHandler.CreateSubscription).Methods("POST")
	subscriptions.HandleFunc("/{id}", subscriptionHandler.DeleteSubscription).Methods("DELETE")

	// Data routes
	data := api.PathPrefix("/data").Subrouter()
	data.HandleFunc("/tickers", dataHandler.GetTickers).Methods("GET")
	data.HandleFunc("/candles", dataHandler.GetCandles).Methods("GET")
	data.HandleFunc("/orderbooks", dataHandler.GetOrderBooks).Methods("GET")

	return router
}

// TestAPIIntegration_InstrumentWorkflow tests the complete instrument management workflow
func TestAPIIntegration_InstrumentWorkflow(t *testing.T) {
	// Setup mocks
	mockInstrumentManager := &MockInstrumentManagerAPI{}
	mockDataQuery := &MockDataQueryAPI{}

	// Setup router
	router := setupTestRouter(mockInstrumentManager, mockDataQuery)

	// Test data
	testInstrument := entities.InstrumentInfo{
		Symbol:            "BTCUSDT",
		BaseAsset:         "BTC",
		QuoteAsset:        "USDT",
		Type:              entities.InstrumentTypeSpot,
		Market:            entities.MarketTypeSpot,
		MinPrice:          0.01,
		MaxPrice:          1000000.0,
		MinQuantity:       0.001,
		MaxQuantity:       10000.0,
		PricePrecision:    2,
		QuantityPrecision: 3,
	}

	// 1. Test creating instrument
	mockInstrumentManager.On("AddInstrument", mock.Anything, mock.AnythingOfType("entities.InstrumentInfo")).Return(nil)

	createReq := dto.CreateInstrumentRequest{
		Symbol:            testInstrument.Symbol,
		BaseAsset:         testInstrument.BaseAsset,
		QuoteAsset:        testInstrument.QuoteAsset,
		Type:              testInstrument.Type,
		Market:            testInstrument.Market,
		MinPrice:          testInstrument.MinPrice,
		MaxPrice:          testInstrument.MaxPrice,
		MinQuantity:       testInstrument.MinQuantity,
		MaxQuantity:       testInstrument.MaxQuantity,
		PricePrecision:    testInstrument.PricePrecision,
		QuantityPrecision: testInstrument.QuantityPrecision,
	}

	createBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/instruments", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// 2. Test getting instrument
	mockInstrumentManager.On("GetInstrument", mock.Anything, "BTCUSDT").Return(&testInstrument, nil)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/instruments/BTCUSDT", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	// 3. Test listing instruments
	instruments := []entities.InstrumentInfo{testInstrument}
	mockInstrumentManager.On("ListInstruments", mock.Anything).Return(instruments, nil)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/instruments", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var listResponse dto.Response
	err = json.Unmarshal(w.Body.Bytes(), &listResponse)
	require.NoError(t, err)
	assert.True(t, listResponse.Success)

	// 4. Test updating instrument
	mockInstrumentManager.On("GetInstrument", mock.Anything, "BTCUSDT").Return(&testInstrument, nil)
	mockInstrumentManager.On("AddInstrument", mock.Anything, mock.AnythingOfType("entities.InstrumentInfo")).Return(nil)

	minPrice := 0.02
	maxPrice := 2000000.0
	minQuantity := 0.002
	maxQuantity := 20000.0

	updateReq := dto.UpdateInstrumentRequest{
		MinPrice:    &minPrice,
		MaxPrice:    &maxPrice,
		MinQuantity: &minQuantity,
		MaxQuantity: &maxQuantity,
	}

	updateBody, _ := json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/instruments/BTCUSDT", bytes.NewBuffer(updateBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Test deleting instrument (currently not implemented)
	mockInstrumentManager.On("GetInstrument", mock.Anything, "BTCUSDT").Return(&testInstrument, nil)

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/instruments/BTCUSDT", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code) // Current implementation returns not implemented

	// Verify all expectations
	mockInstrumentManager.AssertExpectations(t)
}

// TestAPIIntegration_SubscriptionWorkflow tests the complete subscription management workflow
func TestAPIIntegration_SubscriptionWorkflow(t *testing.T) {
	// Setup mocks
	mockInstrumentManager := &MockInstrumentManagerAPI{}
	mockDataQuery := &MockDataQueryAPI{}

	// Setup router
	router := setupTestRouter(mockInstrumentManager, mockDataQuery)

	// Test data
	testSubscription := entities.InstrumentSubscription{
		ID:       "sub-1",
		Symbol:   "BTCUSDT",
		BrokerID: "binance",
		DataTypes: []entities.DataType{
			entities.DataTypeTicker,
			entities.DataTypeCandle,
		},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 1. Test creating subscription
	mockInstrumentManager.On("AddSubscription", mock.Anything, mock.AnythingOfType("entities.InstrumentSubscription")).Return(nil)

	createReq := dto.CreateSubscriptionRequest{
		Symbol:    testSubscription.Symbol,
		Type:      entities.InstrumentTypeSpot,
		Market:    entities.MarketTypeSpot,
		BrokerID:  testSubscription.BrokerID,
		DataTypes: testSubscription.DataTypes,
		StartDate: time.Now(),
	}

	createBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBuffer(createBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// 2. Test listing subscriptions
	subscriptions := []entities.InstrumentSubscription{testSubscription}
	mockInstrumentManager.On("ListSubscriptions", mock.Anything).Return(subscriptions, nil)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var listResponse dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &listResponse)
	require.NoError(t, err)
	assert.True(t, listResponse.Success)

	// 3. Test deleting subscription
	mockInstrumentManager.On("GetSubscription", mock.Anything, "sub-1").Return(&testSubscription, nil)
	mockInstrumentManager.On("RemoveSubscription", mock.Anything, "sub-1").Return(nil)

	req = httptest.NewRequest(http.MethodDelete, "/api/v1/subscriptions/sub-1", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify all expectations
	mockInstrumentManager.AssertExpectations(t)
}

// TestAPIIntegration_DataQueryWorkflow tests the data retrieval workflow
func TestAPIIntegration_DataQueryWorkflow(t *testing.T) {
	// Setup mocks
	mockInstrumentManager := &MockInstrumentManagerAPI{}
	mockDataQuery := &MockDataQueryAPI{}

	// Setup router
	router := setupTestRouter(mockInstrumentManager, mockDataQuery)

	// Test data
	testTicker := entities.Ticker{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		Volume:    1000.0,
		Timestamp: time.Now(),
		BrokerID:  "binance",
	}

	testCandle := entities.Candle{
		Symbol:    "BTCUSDT",
		Timeframe: "1m",
		Open:      49900.0,
		High:      50100.0,
		Low:       49800.0,
		Close:     50000.0,
		Volume:    500.0,
		Timestamp: time.Now(),
		BrokerID:  "binance",
	}

	testOrderBook := entities.OrderBook{
		Symbol:    "BTCUSDT",
		Bids:      []entities.PriceLevel{{Price: 49990.0, Quantity: 1.0}},
		Asks:      []entities.PriceLevel{{Price: 50010.0, Quantity: 1.0}},
		Timestamp: time.Now(),
		BrokerID:  "binance",
	}

	// 1. Test getting tickers
	tickers := []entities.Ticker{testTicker}
	mockDataQuery.On("GetTickers", mock.Anything, mock.AnythingOfType("interfaces.TickerFilter")).Return(tickers, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/data/tickers?symbol=BTCUSDT", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var tickerResponse dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &tickerResponse)
	require.NoError(t, err)
	assert.True(t, tickerResponse.Success)

	// 2. Test getting candles
	candles := []entities.Candle{testCandle}
	mockDataQuery.On("GetCandles", mock.Anything, mock.AnythingOfType("interfaces.CandleFilter")).Return(candles, nil)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/data/candles?symbol=BTCUSDT&timeframe=1m", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var candleResponse dto.Response
	err = json.Unmarshal(w.Body.Bytes(), &candleResponse)
	require.NoError(t, err)
	assert.True(t, candleResponse.Success)

	// 3. Test getting order books
	orderBooks := []entities.OrderBook{testOrderBook}
	mockDataQuery.On("GetOrderBooks", mock.Anything, mock.AnythingOfType("interfaces.OrderBookFilter")).Return(orderBooks, nil)

	req = httptest.NewRequest(http.MethodGet, "/api/v1/data/orderbooks?symbol=BTCUSDT", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var orderBookResponse dto.Response
	err = json.Unmarshal(w.Body.Bytes(), &orderBookResponse)
	require.NoError(t, err)
	assert.True(t, orderBookResponse.Success)

	// Verify all expectations
	mockDataQuery.AssertExpectations(t)
}

// TestAPIIntegration_ErrorHandling tests error handling scenarios
func TestAPIIntegration_ErrorHandling(t *testing.T) {
	// Setup mocks
	mockInstrumentManager := &MockInstrumentManagerAPI{}
	mockDataQuery := &MockDataQueryAPI{}

	// Setup router
	router := setupTestRouter(mockInstrumentManager, mockDataQuery)

	// 1. Test 404 for non-existent instrument
	mockInstrumentManager.On("GetInstrument", mock.Anything, "NONEXISTENT").Return(nil, fmt.Errorf("instrument not found"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/instruments/NONEXISTENT", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errorResponse dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.False(t, errorResponse.Success)
	assert.NotNil(t, errorResponse.Error)

	// 2. Test invalid JSON in request body
	req = httptest.NewRequest(http.MethodPost, "/api/v1/instruments", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.False(t, errorResponse.Success)
	assert.NotNil(t, errorResponse.Error)

	// 3. Test missing required fields
	invalidReq := dto.CreateInstrumentRequest{
		Symbol: "", // Missing required field
	}

	invalidBody, _ := json.Marshal(invalidReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/instruments", bytes.NewBuffer(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.False(t, errorResponse.Success)
	assert.NotNil(t, errorResponse.Error)

	// Verify all expectations
	mockInstrumentManager.AssertExpectations(t)
}
