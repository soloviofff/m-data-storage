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

// MockDateFilterService is a mock implementation of DateFilterService
type MockDateFilterService struct {
	mock.Mock
}

func (m *MockDateFilterService) FilterTickersBySubscriptionDate(ctx context.Context, filter interfaces.TickerFilter) ([]entities.Ticker, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Ticker), args.Error(1)
}

func (m *MockDateFilterService) FilterCandlesBySubscriptionDate(ctx context.Context, filter interfaces.CandleFilter) ([]entities.Candle, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Candle), args.Error(1)
}

func (m *MockDateFilterService) FilterOrderBooksBySubscriptionDate(ctx context.Context, filter interfaces.OrderBookFilter) ([]entities.OrderBook, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.OrderBook), args.Error(1)
}

func TestNewDataQueryService(t *testing.T) {
	mockStorage := &MockStorageManagerForDataQuery{}
	testLogger := logrus.New()

	service := NewDataQueryService(mockStorage, nil, testLogger)

	assert.NotNil(t, service)
	assert.Equal(t, mockStorage, service.storageManager)
	assert.Nil(t, service.dateFilter)
	assert.Equal(t, testLogger, service.logger)
}

func TestNewDataQueryService_NilLogger(t *testing.T) {
	mockStorage := &MockStorageManagerForDataQuery{}

	service := NewDataQueryService(mockStorage, nil, nil)

	assert.NotNil(t, service)
	assert.NotNil(t, service.logger)
}

func TestDataQueryService_GetTickers(t *testing.T) {
	mockStorage := &MockStorageManagerForDataQuery{}
	testLogger := logrus.New()
	service := NewDataQueryService(mockStorage, nil, testLogger)

	// Test data
	testTickers := []entities.Ticker{
		{
			Symbol:    "BTC/USDT",
			Price:     50000.0,
			Volume:    1000.0,
			Timestamp: time.Now(),
			BrokerID:  "test-broker",
		},
	}

	filter := interfaces.TickerFilter{
		Symbols: []string{"BTC/USDT"},
		Limit:   10,
	}

	// Setup mock expectations
	mockStorage.On("GetTickers", mock.Anything, filter).Return(testTickers, nil)

	// Execute
	result, err := service.GetTickers(context.Background(), filter)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, testTickers, result)
	mockStorage.AssertExpectations(t)
}

func TestDataQueryService_GetTickers_NilStorage(t *testing.T) {
	testLogger := logrus.New()
	service := NewDataQueryService(nil, nil, testLogger)

	filter := interfaces.TickerFilter{
		Symbols: []string{"BTC/USDT"},
		Limit:   10,
	}

	// Execute
	result, err := service.GetTickers(context.Background(), filter)

	// Verify graceful degradation
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDataQueryService_GetCandles(t *testing.T) {
	mockStorage := &MockStorageManagerForDataQuery{}
	testLogger := logrus.New()
	service := NewDataQueryService(mockStorage, nil, testLogger)

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
			BrokerID:  "test-broker",
		},
	}

	filter := interfaces.CandleFilter{
		Symbols:    []string{"BTC/USDT"},
		Timeframes: []string{"1h"},
		Limit:      10,
	}

	// Setup mock expectations
	mockStorage.On("GetCandles", mock.Anything, filter).Return(testCandles, nil)

	// Execute
	result, err := service.GetCandles(context.Background(), filter)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, testCandles, result)
	mockStorage.AssertExpectations(t)
}

func TestDataQueryService_GetCandles_NilStorage(t *testing.T) {
	testLogger := logrus.New()
	service := NewDataQueryService(nil, nil, testLogger)

	filter := interfaces.CandleFilter{
		Symbols:    []string{"BTC/USDT"},
		Timeframes: []string{"1h"},
		Limit:      10,
	}

	// Execute
	result, err := service.GetCandles(context.Background(), filter)

	// Verify graceful degradation
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDataQueryService_GetOrderBooks(t *testing.T) {
	mockStorage := &MockStorageManagerForDataQuery{}
	testLogger := logrus.New()
	service := NewDataQueryService(mockStorage, nil, testLogger)

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
			BrokerID:  "test-broker",
		},
	}

	filter := interfaces.OrderBookFilter{
		Symbols: []string{"BTC/USDT"},
		Limit:   10,
	}

	// Setup mock expectations
	mockStorage.On("GetOrderBooks", mock.Anything, filter).Return(testOrderBooks, nil)

	// Execute
	result, err := service.GetOrderBooks(context.Background(), filter)

	// Verify
	assert.NoError(t, err)
	assert.Equal(t, testOrderBooks, result)
	mockStorage.AssertExpectations(t)
}

func TestDataQueryService_GetOrderBooks_NilStorage(t *testing.T) {
	testLogger := logrus.New()
	service := NewDataQueryService(nil, nil, testLogger)

	filter := interfaces.OrderBookFilter{
		Symbols: []string{"BTC/USDT"},
		Limit:   10,
	}

	// Execute
	result, err := service.GetOrderBooks(context.Background(), filter)

	// Verify graceful degradation
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDataQueryService_GetTickerAggregates(t *testing.T) {
	testLogger := logrus.New()
	service := NewDataQueryService(nil, nil, testLogger)

	// Execute
	result, err := service.GetTickerAggregates(context.Background(), "BTC/USDT", "1h", time.Now().Add(-time.Hour), time.Now())

	// Verify not implemented yet
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDataQueryService_GetCandleAggregates(t *testing.T) {
	testLogger := logrus.New()
	service := NewDataQueryService(nil, nil, testLogger)

	// Execute
	result, err := service.GetCandleAggregates(context.Background(), "BTC/USDT", "1h", time.Now().Add(-time.Hour), time.Now())

	// Verify not implemented yet
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDataQueryService_GetDataStats(t *testing.T) {
	testLogger := logrus.New()
	service := NewDataQueryService(nil, nil, testLogger)

	// Execute
	result, err := service.GetDataStats(context.Background())

	// Verify not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.TotalRecords)
}

func TestDataQueryService_GetTickers_WithDateFilter(t *testing.T) {
	mockStorage := &MockStorageManagerForDataQuery{}
	mockDateFilter := &MockDateFilterService{}
	testLogger := logrus.New()
	service := NewDataQueryService(mockStorage, mockDateFilter, testLogger)

	// Test data
	filter := interfaces.TickerFilter{
		Symbols: []string{"BTC/USDT"},
	}
	expectedTickers := []entities.Ticker{
		{
			Symbol:    "BTC/USDT",
			Price:     50000.0,
			Timestamp: time.Now(),
		},
	}

	// Setup mock
	mockDateFilter.On("FilterTickersBySubscriptionDate", mock.Anything, filter).Return(expectedTickers, nil)

	// Call the method
	result, err := service.GetTickers(context.Background(), filter)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedTickers, result)
	mockDateFilter.AssertExpectations(t)
	// Storage manager should not be called when date filter is used
	mockStorage.AssertNotCalled(t, "GetTickers")
}

func TestDataQueryService_GetCandles_WithDateFilter(t *testing.T) {
	mockStorage := &MockStorageManagerForDataQuery{}
	mockDateFilter := &MockDateFilterService{}
	testLogger := logrus.New()
	service := NewDataQueryService(mockStorage, mockDateFilter, testLogger)

	// Test data
	filter := interfaces.CandleFilter{
		Symbols: []string{"BTC/USDT"},
	}
	expectedCandles := []entities.Candle{
		{
			Symbol:    "BTC/USDT",
			Open:      50000.0,
			High:      51000.0,
			Low:       49000.0,
			Close:     50500.0,
			Timestamp: time.Now(),
		},
	}

	// Setup mock
	mockDateFilter.On("FilterCandlesBySubscriptionDate", mock.Anything, filter).Return(expectedCandles, nil)

	// Call the method
	result, err := service.GetCandles(context.Background(), filter)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedCandles, result)
	mockDateFilter.AssertExpectations(t)
	// Storage manager should not be called when date filter is used
	mockStorage.AssertNotCalled(t, "GetCandles")
}

func TestDataQueryService_GetOrderBooks_WithDateFilter(t *testing.T) {
	mockStorage := &MockStorageManagerForDataQuery{}
	mockDateFilter := &MockDateFilterService{}
	testLogger := logrus.New()
	service := NewDataQueryService(mockStorage, mockDateFilter, testLogger)

	// Test data
	filter := interfaces.OrderBookFilter{
		Symbols: []string{"BTC/USDT"},
	}
	expectedOrderBooks := []entities.OrderBook{
		{
			Symbol:    "BTC/USDT",
			Asks:      []entities.PriceLevel{{Price: 50100.0, Quantity: 1.0}},
			Bids:      []entities.PriceLevel{{Price: 49900.0, Quantity: 1.0}},
			Timestamp: time.Now(),
		},
	}

	// Setup mock
	mockDateFilter.On("FilterOrderBooksBySubscriptionDate", mock.Anything, filter).Return(expectedOrderBooks, nil)

	// Call the method
	result, err := service.GetOrderBooks(context.Background(), filter)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedOrderBooks, result)
	mockDateFilter.AssertExpectations(t)
	// Storage manager should not be called when date filter is used
	mockStorage.AssertNotCalled(t, "GetOrderBooks")
}
