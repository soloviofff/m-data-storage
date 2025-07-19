package questdb

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// Mock QuestDB repository for testing
type MockTimeSeriesRepository struct {
	tickers    []entities.Ticker
	candles    []entities.Candle
	orderBooks []entities.OrderBook
}

func NewMockTimeSeriesRepository() *MockTimeSeriesRepository {
	return &MockTimeSeriesRepository{
		tickers:    make([]entities.Ticker, 0),
		candles:    make([]entities.Candle, 0),
		orderBooks: make([]entities.OrderBook, 0),
	}
}

func (m *MockTimeSeriesRepository) Connect(ctx context.Context) error {
	return nil
}

func (m *MockTimeSeriesRepository) Disconnect() error {
	return nil
}

func (m *MockTimeSeriesRepository) Health() error {
	return nil
}

func (m *MockTimeSeriesRepository) SaveTicker(ctx context.Context, ticker entities.Ticker) error {
	m.tickers = append(m.tickers, ticker)
	return nil
}

func (m *MockTimeSeriesRepository) GetTickers(ctx context.Context, symbol string, from, to time.Time) ([]entities.Ticker, error) {
	var result []entities.Ticker
	for _, ticker := range m.tickers {
		if ticker.Symbol == symbol && ticker.Timestamp.After(from) && ticker.Timestamp.Before(to) {
			result = append(result, ticker)
		}
	}
	return result, nil
}

func (m *MockTimeSeriesRepository) SaveCandle(ctx context.Context, candle entities.Candle) error {
	m.candles = append(m.candles, candle)
	return nil
}

func (m *MockTimeSeriesRepository) GetCandles(ctx context.Context, symbol string, timeframe string, from, to time.Time) ([]entities.Candle, error) {
	var result []entities.Candle
	for _, candle := range m.candles {
		if candle.Symbol == symbol && candle.Timeframe == timeframe && candle.Timestamp.After(from) && candle.Timestamp.Before(to) {
			result = append(result, candle)
		}
	}
	return result, nil
}

func (m *MockTimeSeriesRepository) SaveOrderBook(ctx context.Context, orderBook entities.OrderBook) error {
	m.orderBooks = append(m.orderBooks, orderBook)
	return nil
}

func (m *MockTimeSeriesRepository) GetOrderBooks(ctx context.Context, symbol string, from, to time.Time) ([]entities.OrderBook, error) {
	var result []entities.OrderBook
	for _, orderBook := range m.orderBooks {
		if orderBook.Symbol == symbol && orderBook.Timestamp.After(from) && orderBook.Timestamp.Before(to) {
			result = append(result, orderBook)
		}
	}
	return result, nil
}

func (m *MockTimeSeriesRepository) CleanupOldData(ctx context.Context, before time.Time) error {
	return nil
}

func TestMockTimeSeriesRepository_TickerOperations(t *testing.T) {
	repo := NewMockTimeSeriesRepository()
	ctx := context.Background()

	// Test ticker
	ticker := entities.Ticker{
		Symbol:    "BTCUSDT",
		Timestamp: time.Now(),
		BidPrice:  50000.0,
		AskPrice:  50001.0,
		BidSize:   1.0,
		AskSize:   1.0,
	}

	// Save ticker
	err := repo.SaveTicker(ctx, ticker)
	if err != nil {
		t.Fatalf("Failed to save ticker: %v", err)
	}

	// Get tickers
	from := time.Now().Add(-time.Hour)
	to := time.Now().Add(time.Hour)
	tickers, err := repo.GetTickers(ctx, "BTCUSDT", from, to)
	if err != nil {
		t.Fatalf("Failed to get tickers: %v", err)
	}

	if len(tickers) != 1 {
		t.Fatalf("Expected 1 ticker, got %d", len(tickers))
	}

	if tickers[0].Symbol != ticker.Symbol {
		t.Errorf("Expected symbol %s, got %s", ticker.Symbol, tickers[0].Symbol)
	}
}

func TestMockTimeSeriesRepository_CandleOperations(t *testing.T) {
	repo := NewMockTimeSeriesRepository()
	ctx := context.Background()

	// Test candle
	candle := entities.Candle{
		Symbol:    "BTCUSDT",
		Timeframe: "1m",
		Timestamp: time.Now(),
		Open:      50000.0,
		High:      50100.0,
		Low:       49900.0,
		Close:     50050.0,
		Volume:    100.0,
		BrokerID:  "test-broker",
	}

	// Save candle
	err := repo.SaveCandle(ctx, candle)
	if err != nil {
		t.Fatalf("Failed to save candle: %v", err)
	}

	// Get candles
	from := time.Now().Add(-time.Hour)
	to := time.Now().Add(time.Hour)
	candles, err := repo.GetCandles(ctx, "BTCUSDT", "1m", from, to)
	if err != nil {
		t.Fatalf("Failed to get candles: %v", err)
	}

	if len(candles) != 1 {
		t.Fatalf("Expected 1 candle, got %d", len(candles))
	}

	if candles[0].Symbol != candle.Symbol {
		t.Errorf("Expected symbol %s, got %s", candle.Symbol, candles[0].Symbol)
	}
}

func TestMockTimeSeriesRepository_OrderBookOperations(t *testing.T) {
	repo := NewMockTimeSeriesRepository()
	ctx := context.Background()

	// Test order book
	orderBook := entities.OrderBook{
		Symbol:    "BTCUSDT",
		Timestamp: time.Now(),
		BrokerID:  "test-broker",
		Bids: []entities.PriceLevel{
			{Price: 50000.0, Quantity: 1.0},
		},
		Asks: []entities.PriceLevel{
			{Price: 50001.0, Quantity: 1.0},
		},
	}

	// Save order book
	err := repo.SaveOrderBook(ctx, orderBook)
	if err != nil {
		t.Fatalf("Failed to save order book: %v", err)
	}

	// Get order books
	from := time.Now().Add(-time.Hour)
	to := time.Now().Add(time.Hour)
	orderBooks, err := repo.GetOrderBooks(ctx, "BTCUSDT", from, to)
	if err != nil {
		t.Fatalf("Failed to get order books: %v", err)
	}

	if len(orderBooks) != 1 {
		t.Fatalf("Expected 1 order book, got %d", len(orderBooks))
	}

	if orderBooks[0].Symbol != orderBook.Symbol {
		t.Errorf("Expected symbol %s, got %s", orderBook.Symbol, orderBooks[0].Symbol)
	}
}

// Tests for real TimeSeriesRepository implementation

func TestNewTimeSeriesRepository(t *testing.T) {
	config := Config{
		Host:     "localhost",
		Port:     8812,
		Database: "qdb",
		Username: "admin",
		Password: "quest",
		SSLMode:  "disable",
	}

	repo := NewTimeSeriesRepository(config)

	assert.NotNil(t, repo)
	assert.Equal(t, config, repo.config)
	assert.Nil(t, repo.db) // DB should be nil until Connect is called
}

func TestTimeSeriesRepository_GetDB(t *testing.T) {
	repo := &TimeSeriesRepository{}

	db := repo.GetDB()
	assert.Nil(t, db)
}

func TestTimeSeriesRepository_Health_NilDB(t *testing.T) {
	repo := &TimeSeriesRepository{}

	err := repo.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}

func TestTimeSeriesRepository_Disconnect_NilDB(t *testing.T) {
	repo := &TimeSeriesRepository{}

	err := repo.Disconnect()
	assert.NoError(t, err) // Should not error when db is nil
}

func TestTimeSeriesRepository_SaveTickers_EmptySlice(t *testing.T) {
	repo := &TimeSeriesRepository{}
	ctx := context.Background()

	err := repo.SaveTickers(ctx, []entities.Ticker{})
	assert.NoError(t, err) // Should not error for empty slice
}

func TestTimeSeriesRepository_SaveCandles_EmptySlice(t *testing.T) {
	repo := &TimeSeriesRepository{}
	ctx := context.Background()

	err := repo.SaveCandles(ctx, []entities.Candle{})
	assert.NoError(t, err) // Should not error for empty slice
}

func TestTimeSeriesRepository_SaveOrderBooks_EmptySlice(t *testing.T) {
	repo := &TimeSeriesRepository{}
	ctx := context.Background()

	err := repo.SaveOrderBooks(ctx, []entities.OrderBook{})
	assert.NoError(t, err) // Should not error for empty slice
}

func TestTimeSeriesRepository_PriceLevelsToJSON(t *testing.T) {
	repo := &TimeSeriesRepository{}

	t.Run("empty slice", func(t *testing.T) {
		result := repo.priceLevelsToJSON([]entities.PriceLevel{})
		assert.Equal(t, "[]", result)
	})

	t.Run("single level", func(t *testing.T) {
		levels := []entities.PriceLevel{
			{Price: 50000.0, Quantity: 1.5},
		}
		result := repo.priceLevelsToJSON(levels)
		expected := `[{"price":50000.000000,"quantity":1.500000}]`
		assert.Equal(t, expected, result)
	})

	t.Run("multiple levels", func(t *testing.T) {
		levels := []entities.PriceLevel{
			{Price: 50000.0, Quantity: 1.5},
			{Price: 50001.0, Quantity: 2.0},
		}
		result := repo.priceLevelsToJSON(levels)
		expected := `[{"price":50000.000000,"quantity":1.500000},{"price":50001.000000,"quantity":2.000000}]`
		assert.Equal(t, expected, result)
	})
}

func TestTimeSeriesRepository_ParsePriceLevelsFromJSON(t *testing.T) {
	repo := &TimeSeriesRepository{}

	// This is a simplified implementation that always returns empty slice
	result := repo.parsePriceLevelsFromJSON(`[{"price":50000.0,"quantity":1.5}]`)
	assert.Empty(t, result)
}

func TestTimeSeriesRepository_BuildPlaceholders(t *testing.T) {
	repo := &TimeSeriesRepository{}

	t.Run("single placeholder", func(t *testing.T) {
		argIndex := 1
		result := repo.buildPlaceholders(1, &argIndex)
		assert.Equal(t, "$1", result)
		assert.Equal(t, 2, argIndex) // Should increment
	})

	t.Run("multiple placeholders", func(t *testing.T) {
		argIndex := 5
		result := repo.buildPlaceholders(3, &argIndex)
		// Note: Current implementation only returns first placeholder
		// This is a known limitation mentioned in the code
		assert.Equal(t, "$5", result)
		assert.Equal(t, 8, argIndex) // Should increment by count
	})
}

func TestTimeSeriesRepository_GetTickers_NilDB(t *testing.T) {
	repo := &TimeSeriesRepository{}
	ctx := context.Background()
	filter := interfaces.TickerFilter{}

	// Should panic with nil pointer dereference
	assert.Panics(t, func() {
		repo.GetTickers(ctx, filter)
	})
}

func TestTimeSeriesRepository_GetCandles_NilDB(t *testing.T) {
	repo := &TimeSeriesRepository{}
	ctx := context.Background()
	filter := interfaces.CandleFilter{}

	// Should panic with nil pointer dereference
	assert.Panics(t, func() {
		repo.GetCandles(ctx, filter)
	})
}

func TestTimeSeriesRepository_GetOrderBooks_NilDB(t *testing.T) {
	repo := &TimeSeriesRepository{}
	ctx := context.Background()
	filter := interfaces.OrderBookFilter{}

	// Should panic with nil pointer dereference
	assert.Panics(t, func() {
		repo.GetOrderBooks(ctx, filter)
	})
}

func TestTimeSeriesRepository_GetTickerAggregates_NilDB(t *testing.T) {
	repo := &TimeSeriesRepository{}
	ctx := context.Background()
	startTime := time.Now().Add(-time.Hour)
	endTime := time.Now()

	// Should panic with nil pointer dereference
	assert.Panics(t, func() {
		repo.GetTickerAggregates(ctx, "BTCUSD", "1h", startTime, endTime)
	})
}

func TestTimeSeriesRepository_GetCandleAggregates_NilDB(t *testing.T) {
	repo := &TimeSeriesRepository{}
	ctx := context.Background()
	startTime := time.Now().Add(-time.Hour)
	endTime := time.Now()

	// Should panic with nil pointer dereference
	assert.Panics(t, func() {
		repo.GetCandleAggregates(ctx, "BTCUSD", "1h", startTime, endTime)
	})
}

func TestTimeSeriesRepository_CleanupOldData_NilDB(t *testing.T) {
	repo := &TimeSeriesRepository{}
	ctx := context.Background()
	retentionPeriod := 24 * time.Hour

	// Should panic with nil pointer dereference
	assert.Panics(t, func() {
		repo.CleanupOldData(ctx, retentionPeriod)
	})
}

func TestTimeSeriesRepository_SaveTickers_NilDB(t *testing.T) {
	repo := &TimeSeriesRepository{}
	ctx := context.Background()

	ticker := entities.Ticker{
		Symbol:    "BTCUSD",
		Price:     50000.0,
		Volume:    1.5,
		Market:    entities.MarketTypeSpot,
		Type:      entities.InstrumentTypeSpot,
		Timestamp: time.Now(),
		BrokerID:  "test-broker",
	}

	// Should panic with nil pointer dereference
	assert.Panics(t, func() {
		repo.SaveTickers(ctx, []entities.Ticker{ticker})
	})
}

func TestTimeSeriesRepository_SaveCandles_NilDB(t *testing.T) {
	repo := &TimeSeriesRepository{}
	ctx := context.Background()

	candle := entities.Candle{
		Symbol:    "BTCUSD",
		Open:      49000.0,
		High:      51000.0,
		Low:       48000.0,
		Close:     50000.0,
		Volume:    100.0,
		Market:    entities.MarketTypeSpot,
		Type:      entities.InstrumentTypeSpot,
		Timestamp: time.Now(),
		Timeframe: "1h",
		BrokerID:  "test-broker",
	}

	// Should panic with nil pointer dereference
	assert.Panics(t, func() {
		repo.SaveCandles(ctx, []entities.Candle{candle})
	})
}

func TestTimeSeriesRepository_SaveOrderBooks_NilDB(t *testing.T) {
	repo := &TimeSeriesRepository{}
	ctx := context.Background()

	orderBook := entities.OrderBook{
		Symbol:    "BTCUSD",
		Market:    entities.MarketTypeSpot,
		Type:      entities.InstrumentTypeSpot,
		Timestamp: time.Now(),
		BrokerID:  "test-broker",
		Bids: []entities.PriceLevel{
			{Price: 49999.0, Quantity: 1.0},
		},
		Asks: []entities.PriceLevel{
			{Price: 50001.0, Quantity: 1.0},
		},
	}

	// Should panic with nil pointer dereference
	assert.Panics(t, func() {
		repo.SaveOrderBooks(ctx, []entities.OrderBook{orderBook})
	})
}
