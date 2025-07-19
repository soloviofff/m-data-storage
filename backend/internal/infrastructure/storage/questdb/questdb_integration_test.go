package questdb

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// TestQuestDBIntegration_RealDatabase tests actual QuestDB operations
func TestQuestDBIntegration_RealDatabase(t *testing.T) {
	// Skip if running in CI or if QuestDB is not available
	if testing.Short() {
		t.Skip("Skipping QuestDB integration test in short mode")
	}

	// Setup test database configuration
	cfg := Config{
		Host:     "localhost",
		Port:     8812, // QuestDB PostgreSQL wire protocol port
		Database: "qdb",
		Username: "admin",
		Password: "quest",
		SSLMode:  "disable",
	}

	// Create QuestDB repository
	questRepo := NewTimeSeriesRepository(cfg)

	// Connect to database
	ctx := context.Background()
	err := questRepo.Connect(ctx)
	if err != nil {
		t.Skip("Skipping QuestDB integration test - database not available:", err)
	}
	defer questRepo.Disconnect()

	// Test connection
	err = questRepo.Health()
	require.NoError(t, err, "Failed to connect to QuestDB")

	// Run subtests
	t.Run("TickerOperations", func(t *testing.T) {
		testTickerOperations(t, questRepo, ctx)
	})

	t.Run("CandleOperations", func(t *testing.T) {
		testCandleOperations(t, questRepo, ctx)
	})

	t.Run("OrderBookOperations", func(t *testing.T) {
		testOrderBookOperations(t, questRepo, ctx)
	})

	t.Run("QueryFiltering", func(t *testing.T) {
		testQueryFiltering(t, questRepo, ctx)
	})
}

func testTickerOperations(t *testing.T, repo *TimeSeriesRepository, ctx context.Context) {
	// Test data
	testTicker := entities.Ticker{
		Symbol:    "BTCUSDT_TEST",
		Price:     50000.0,
		Volume:    1000.0,
		Timestamp: time.Now().UTC(),
		BrokerID:  "binance",
		Market:    entities.MarketTypeSpot,
		Type:      entities.InstrumentTypeSpot,
	}

	// Clean up any existing test data
	cleanupTestData(t, repo, "tickers", "symbol = 'BTCUSDT_TEST'")

	// Test storing ticker
	err := repo.SaveTickers(ctx, []entities.Ticker{testTicker})
	require.NoError(t, err)

	// Test retrieving ticker
	filter := interfaces.TickerFilter{
		Symbols:   []string{testTicker.Symbol},
		BrokerIDs: []string{testTicker.BrokerID},
		Limit:     10,
	}

	tickers, err := repo.GetTickers(ctx, filter)
	require.NoError(t, err)
	require.Len(t, tickers, 1)

	retrieved := tickers[0]
	assert.Equal(t, testTicker.Symbol, retrieved.Symbol)
	assert.Equal(t, testTicker.Price, retrieved.Price)
	assert.Equal(t, testTicker.Volume, retrieved.Volume)
	assert.Equal(t, testTicker.BrokerID, retrieved.BrokerID)
	assert.Equal(t, testTicker.Market, retrieved.Market)
	assert.Equal(t, testTicker.Type, retrieved.Type)

	// Clean up
	cleanupTestData(t, repo, "tickers", "symbol = 'BTCUSDT_TEST'")
}

func testCandleOperations(t *testing.T, repo *TimeSeriesRepository, ctx context.Context) {
	// Test data
	testCandle := entities.Candle{
		Symbol:    "ETHUSDT_TEST",
		Timeframe: "1m",
		Open:      3000.0,
		High:      3100.0,
		Low:       2950.0,
		Close:     3050.0,
		Volume:    500.0,
		Timestamp: time.Now().UTC(),
		BrokerID:  "binance",
		Market:    entities.MarketTypeSpot,
		Type:      entities.InstrumentTypeSpot,
	}

	// Clean up any existing test data
	cleanupTestData(t, repo, "candles", "symbol = 'ETHUSDT_TEST'")

	// Test storing candle
	err := repo.SaveCandles(ctx, []entities.Candle{testCandle})
	require.NoError(t, err)

	// Test retrieving candle
	filter := interfaces.CandleFilter{
		Symbols:    []string{testCandle.Symbol},
		Timeframes: []string{testCandle.Timeframe},
		BrokerIDs:  []string{testCandle.BrokerID},
		Limit:      10,
	}

	candles, err := repo.GetCandles(ctx, filter)
	require.NoError(t, err)
	require.Len(t, candles, 1)

	retrieved := candles[0]
	assert.Equal(t, testCandle.Symbol, retrieved.Symbol)
	assert.Equal(t, testCandle.Timeframe, retrieved.Timeframe)
	assert.Equal(t, testCandle.Open, retrieved.Open)
	assert.Equal(t, testCandle.High, retrieved.High)
	assert.Equal(t, testCandle.Low, retrieved.Low)
	assert.Equal(t, testCandle.Close, retrieved.Close)
	assert.Equal(t, testCandle.Volume, retrieved.Volume)
	assert.Equal(t, testCandle.BrokerID, retrieved.BrokerID)

	// Clean up
	cleanupTestData(t, repo, "candles", "symbol = 'ETHUSDT_TEST'")
}

func testOrderBookOperations(t *testing.T, repo *TimeSeriesRepository, ctx context.Context) {
	// Test data
	testOrderBook := entities.OrderBook{
		Symbol: "ADAUSDT_TEST",
		Bids: []entities.PriceLevel{
			{Price: 1.0, Quantity: 100.0},
			{Price: 0.99, Quantity: 200.0},
		},
		Asks: []entities.PriceLevel{
			{Price: 1.01, Quantity: 150.0},
			{Price: 1.02, Quantity: 250.0},
		},
		Timestamp: time.Now().UTC(),
		BrokerID:  "binance",
		Market:    entities.MarketTypeSpot,
		Type:      entities.InstrumentTypeSpot,
	}

	// Clean up any existing test data
	cleanupTestData(t, repo, "orderbooks", "symbol = 'ADAUSDT_TEST'")

	// Test storing order book
	err := repo.SaveOrderBooks(ctx, []entities.OrderBook{testOrderBook})
	require.NoError(t, err)

	// Test retrieving order book
	filter := interfaces.OrderBookFilter{
		Symbols:   []string{testOrderBook.Symbol},
		BrokerIDs: []string{testOrderBook.BrokerID},
		Limit:     10,
	}

	orderBooks, err := repo.GetOrderBooks(ctx, filter)
	require.NoError(t, err)
	require.Len(t, orderBooks, 1)

	retrieved := orderBooks[0]
	assert.Equal(t, testOrderBook.Symbol, retrieved.Symbol)
	assert.Equal(t, testOrderBook.BrokerID, retrieved.BrokerID)
	assert.Len(t, retrieved.Bids, 2)
	assert.Len(t, retrieved.Asks, 2)

	// Verify bid/ask data
	assert.Equal(t, testOrderBook.Bids[0].Price, retrieved.Bids[0].Price)
	assert.Equal(t, testOrderBook.Bids[0].Quantity, retrieved.Bids[0].Quantity)
	assert.Equal(t, testOrderBook.Asks[0].Price, retrieved.Asks[0].Price)
	assert.Equal(t, testOrderBook.Asks[0].Quantity, retrieved.Asks[0].Quantity)

	// Clean up
	cleanupTestData(t, repo, "orderbooks", "symbol = 'ADAUSDT_TEST'")
}

func testQueryFiltering(t *testing.T, repo *TimeSeriesRepository, ctx context.Context) {
	// Test time-based filtering with multiple records
	baseTime := time.Now().UTC().Truncate(time.Minute)

	testTickers := []entities.Ticker{
		{
			Symbol:    "FILTER_TEST",
			Price:     100.0,
			Volume:    10.0,
			Timestamp: baseTime,
			BrokerID:  "binance",
			Market:    entities.MarketTypeSpot,
			Type:      entities.InstrumentTypeSpot,
		},
		{
			Symbol:    "FILTER_TEST",
			Price:     101.0,
			Volume:    11.0,
			Timestamp: baseTime.Add(1 * time.Minute),
			BrokerID:  "binance",
			Market:    entities.MarketTypeSpot,
			Type:      entities.InstrumentTypeSpot,
		},
		{
			Symbol:    "FILTER_TEST",
			Price:     102.0,
			Volume:    12.0,
			Timestamp: baseTime.Add(2 * time.Minute),
			BrokerID:  "binance",
			Market:    entities.MarketTypeSpot,
			Type:      entities.InstrumentTypeSpot,
		},
	}

	// Clean up any existing test data
	cleanupTestData(t, repo, "tickers", "symbol = 'FILTER_TEST'")

	// Store test data
	err := repo.SaveTickers(ctx, testTickers)
	require.NoError(t, err)

	// Test time range filtering
	startTime := baseTime.Add(30 * time.Second)
	endTime := baseTime.Add(90 * time.Second)

	filter := interfaces.TickerFilter{
		Symbols:   []string{testTickers[0].Symbol},
		StartTime: &startTime,
		EndTime:   &endTime,
		Limit:     10,
	}

	filteredTickers, err := repo.GetTickers(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, filteredTickers, 1) // Should only get the middle ticker
	assert.Equal(t, 101.0, filteredTickers[0].Price)

	// Test limit filtering
	limitFilter := interfaces.TickerFilter{
		Symbols: []string{testTickers[0].Symbol},
		Limit:   2,
	}

	limitedTickers, err := repo.GetTickers(ctx, limitFilter)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(limitedTickers), 2)

	// Clean up
	cleanupTestData(t, repo, "tickers", "symbol = 'FILTER_TEST'")
}

// Helper function to clean up test data
func cleanupTestData(t *testing.T, repo *TimeSeriesRepository, table, condition string) {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", table, condition)
	_, err := repo.db.Exec(query)
	if err != nil {
		// Ignore errors for cleanup - table might not exist yet
		t.Logf("Cleanup warning: %v", err)
	}
}
