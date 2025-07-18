package services

import (
	"context"

	"github.com/stretchr/testify/mock"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// MockStorageManagerForDataQuery is a mock implementation of StorageManager interface for data query tests
type MockStorageManagerForDataQuery struct {
	mock.Mock
}

func (m *MockStorageManagerForDataQuery) SaveTickers(ctx context.Context, tickers []entities.Ticker) error {
	args := m.Called(ctx, tickers)
	return args.Error(0)
}

func (m *MockStorageManagerForDataQuery) SaveCandles(ctx context.Context, candles []entities.Candle) error {
	args := m.Called(ctx, candles)
	return args.Error(0)
}

func (m *MockStorageManagerForDataQuery) SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error {
	args := m.Called(ctx, orderBooks)
	return args.Error(0)
}

func (m *MockStorageManagerForDataQuery) GetTickers(ctx context.Context, filter interfaces.TickerFilter) ([]entities.Ticker, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Ticker), args.Error(1)
}

func (m *MockStorageManagerForDataQuery) GetCandles(ctx context.Context, filter interfaces.CandleFilter) ([]entities.Candle, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.Candle), args.Error(1)
}

func (m *MockStorageManagerForDataQuery) GetOrderBooks(ctx context.Context, filter interfaces.OrderBookFilter) ([]entities.OrderBook, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]entities.OrderBook), args.Error(1)
}

func (m *MockStorageManagerForDataQuery) GetMetadataStorage() interfaces.MetadataStorage {
	args := m.Called()
	return args.Get(0).(interfaces.MetadataStorage)
}

func (m *MockStorageManagerForDataQuery) GetTimeSeriesStorage() interfaces.TimeSeriesStorage {
	args := m.Called()
	return args.Get(0).(interfaces.TimeSeriesStorage)
}

func (m *MockStorageManagerForDataQuery) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorageManagerForDataQuery) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStorageManagerForDataQuery) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStorageManagerForDataQuery) Shutdown() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStorageManagerForDataQuery) Health() map[string]error {
	args := m.Called()
	return args.Get(0).(map[string]error)
}
