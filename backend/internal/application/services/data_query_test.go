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

func TestNewDataQueryService(t *testing.T) {
	mockStorage := &MockStorageManagerForDataQuery{}
	testLogger := logrus.New()

	service := NewDataQueryService(mockStorage, testLogger)

	assert.NotNil(t, service)
	assert.Equal(t, mockStorage, service.storageManager)
	assert.Equal(t, testLogger, service.logger)
}

func TestNewDataQueryService_NilLogger(t *testing.T) {
	mockStorage := &MockStorageManagerForDataQuery{}

	service := NewDataQueryService(mockStorage, nil)

	assert.NotNil(t, service)
	assert.NotNil(t, service.logger)
}

func TestDataQueryService_GetTickers(t *testing.T) {
	mockStorage := &MockStorageManagerForDataQuery{}
	testLogger := logrus.New()
	service := NewDataQueryService(mockStorage, testLogger)

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
	service := NewDataQueryService(nil, testLogger)

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
	service := NewDataQueryService(mockStorage, testLogger)

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
	service := NewDataQueryService(nil, testLogger)

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
	service := NewDataQueryService(mockStorage, testLogger)

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
	service := NewDataQueryService(nil, testLogger)

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
	service := NewDataQueryService(nil, testLogger)

	// Execute
	result, err := service.GetTickerAggregates(context.Background(), "BTC/USDT", "1h", time.Now().Add(-time.Hour), time.Now())

	// Verify not implemented yet
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDataQueryService_GetCandleAggregates(t *testing.T) {
	testLogger := logrus.New()
	service := NewDataQueryService(nil, testLogger)

	// Execute
	result, err := service.GetCandleAggregates(context.Background(), "BTC/USDT", "1h", time.Now().Add(-time.Hour), time.Now())

	// Verify not implemented yet
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestDataQueryService_GetDataStats(t *testing.T) {
	testLogger := logrus.New()
	service := NewDataQueryService(nil, testLogger)

	// Execute
	result, err := service.GetDataStats(context.Background())

	// Verify not implemented yet
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(0), result.TotalRecords)
}
