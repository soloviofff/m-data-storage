package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"m-data-storage/api/dto"
	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
	"m-data-storage/internal/infrastructure/logger"
)

// MockDataQuery is a mock implementation of DataQuery interface
type MockDataQuery struct {
	mock.Mock
}

func (m *MockDataQuery) GetTickers(ctx context.Context, filter interfaces.TickerFilter) ([]entities.Ticker, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Ticker), args.Error(1)
}

func (m *MockDataQuery) GetCandles(ctx context.Context, filter interfaces.CandleFilter) ([]entities.Candle, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Candle), args.Error(1)
}

func (m *MockDataQuery) GetOrderBooks(ctx context.Context, filter interfaces.OrderBookFilter) ([]entities.OrderBook, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.OrderBook), args.Error(1)
}

func (m *MockDataQuery) GetTickerAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]interfaces.TickerAggregate, error) {
	args := m.Called(ctx, symbol, interval, startTime, endTime)
	return args.Get(0).([]interfaces.TickerAggregate), args.Error(1)
}

func (m *MockDataQuery) GetCandleAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]interfaces.CandleAggregate, error) {
	args := m.Called(ctx, symbol, interval, startTime, endTime)
	return args.Get(0).([]interfaces.CandleAggregate), args.Error(1)
}

func (m *MockDataQuery) GetDataStats(ctx context.Context) (interfaces.DataStats, error) {
	args := m.Called(ctx)
	return args.Get(0).(interfaces.DataStats), args.Error(1)
}

func TestNewDataHandler(t *testing.T) {
	mockDataQuery := &MockDataQuery{}
	testLogger := &logger.Logger{}

	handler := NewDataHandler(mockDataQuery, testLogger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockDataQuery, handler.dataQuery)
	assert.Equal(t, testLogger, handler.logger)
}

func TestDataHandler_GetTickers(t *testing.T) {
	mockDataQuery := &MockDataQuery{}
	testLogger := &logger.Logger{}
	handler := NewDataHandler(mockDataQuery, testLogger)

	// Test data
	testTickers := []entities.Ticker{
		{
			Symbol:        "BTC/USDT",
			Price:         50000.0,
			Volume:        1000.0,
			Change:        500.0,
			ChangePercent: 1.0,
			High24h:       51000.0,
			Low24h:        49000.0,
			PrevClose24h:  49500.0,
			Timestamp:     time.Now(),
		},
	}

	// Setup mock expectations
	mockDataQuery.On("GetTickers", mock.Anything, mock.Anything).Return(testTickers, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/data/tickers?symbol=BTC/USDT", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetTickers(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse as generic Response first
	var genericResponse dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &genericResponse)
	assert.NoError(t, err)
	assert.True(t, genericResponse.Success)

	// Extract the data as TickerListResponse
	dataBytes, err := json.Marshal(genericResponse.Data)
	assert.NoError(t, err)

	var tickerListResponse dto.TickerListResponse
	err = json.Unmarshal(dataBytes, &tickerListResponse)
	assert.NoError(t, err)

	assert.Len(t, tickerListResponse.Tickers, 1)
	assert.Equal(t, "BTC/USDT", tickerListResponse.Tickers[0].Symbol)
	assert.Equal(t, 50000.0, tickerListResponse.Tickers[0].Price)

	mockDataQuery.AssertExpectations(t)
}

func TestDataHandler_GetCandles(t *testing.T) {
	mockDataQuery := &MockDataQuery{}
	testLogger := &logger.Logger{}
	handler := NewDataHandler(mockDataQuery, testLogger)

	// Test data
	testCandles := []entities.Candle{
		{
			Symbol:    "BTC/USDT",
			Open:      49000.0,
			High:      51000.0,
			Low:       48000.0,
			Close:     50000.0,
			Volume:    1000.0,
			Timeframe: "1h",
			Timestamp: time.Now(),
		},
	}

	// Setup mock expectations
	mockDataQuery.On("GetCandles", mock.Anything, mock.Anything).Return(testCandles, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/data/candles?symbol=BTC/USDT&timeframe=1h", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetCandles(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse as generic Response first
	var genericResponse dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &genericResponse)
	assert.NoError(t, err)
	assert.True(t, genericResponse.Success)

	// Extract the data as CandleListResponse
	dataBytes, err := json.Marshal(genericResponse.Data)
	assert.NoError(t, err)

	var candleListResponse dto.CandleListResponse
	err = json.Unmarshal(dataBytes, &candleListResponse)
	assert.NoError(t, err)

	assert.Len(t, candleListResponse.Candles, 1)
	assert.Equal(t, "BTC/USDT", candleListResponse.Candles[0].Symbol)
	assert.Equal(t, "1h", candleListResponse.Candles[0].Timeframe)

	mockDataQuery.AssertExpectations(t)
}

func TestDataHandler_GetOrderBooks(t *testing.T) {
	mockDataQuery := &MockDataQuery{}
	testLogger := &logger.Logger{}
	handler := NewDataHandler(mockDataQuery, testLogger)

	// Test data
	testOrderBooks := []entities.OrderBook{
		{
			Symbol: "BTC/USDT",
			Bids: []entities.PriceLevel{
				{Price: 49900.0, Quantity: 1.0},
			},
			Asks: []entities.PriceLevel{
				{Price: 50100.0, Quantity: 1.0},
			},
			Timestamp: time.Now(),
		},
	}

	// Setup mock expectations
	mockDataQuery.On("GetOrderBooks", mock.Anything, mock.Anything).Return(testOrderBooks, nil)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/data/orderbooks?symbol=BTC/USDT", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetOrderBooks(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse as generic Response first
	var genericResponse dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &genericResponse)
	assert.NoError(t, err)
	assert.True(t, genericResponse.Success)

	// Extract the data as OrderBookResponse
	dataBytes, err := json.Marshal(genericResponse.Data)
	assert.NoError(t, err)

	var orderBookResponse dto.OrderBookResponse
	err = json.Unmarshal(dataBytes, &orderBookResponse)
	assert.NoError(t, err)

	assert.Equal(t, "BTC/USDT", orderBookResponse.Symbol)
	assert.Len(t, orderBookResponse.Bids, 1)
	assert.Len(t, orderBookResponse.Asks, 1)

	mockDataQuery.AssertExpectations(t)
}

func TestDataHandler_GetTickers_MissingSymbol(t *testing.T) {
	mockDataQuery := &MockDataQuery{}
	testLogger := &logger.Logger{}
	handler := NewDataHandler(mockDataQuery, testLogger)

	// Create request without symbol parameter
	req := httptest.NewRequest("GET", "/api/v1/data/tickers", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetTickers(w, req)

	// Verify error response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "missing_parameter", response.Error.Code)
}

func TestDataHandler_GetCandles_MissingTimeframe(t *testing.T) {
	mockDataQuery := &MockDataQuery{}
	testLogger := &logger.Logger{}
	handler := NewDataHandler(mockDataQuery, testLogger)

	// Test data - since GetCandles uses default timeframe "1m", it will still call the service
	testCandles := []entities.Candle{
		{
			Symbol:    "BTC/USDT",
			Open:      49000.0,
			High:      51000.0,
			Low:       48000.0,
			Close:     50000.0,
			Volume:    1000.0,
			Timeframe: "1m", // Default timeframe
			Timestamp: time.Now(),
		},
	}

	// Setup mock expectations for default timeframe
	mockDataQuery.On("GetCandles", mock.Anything, mock.Anything).Return(testCandles, nil)

	// Create request without timeframe parameter
	req := httptest.NewRequest("GET", "/api/v1/data/candles?symbol=BTC/USDT", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.GetCandles(w, req)

	// Verify successful response (since default timeframe is used)
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse as generic Response first
	var genericResponse dto.Response
	err := json.Unmarshal(w.Body.Bytes(), &genericResponse)
	assert.NoError(t, err)
	assert.True(t, genericResponse.Success)

	mockDataQuery.AssertExpectations(t)
}
