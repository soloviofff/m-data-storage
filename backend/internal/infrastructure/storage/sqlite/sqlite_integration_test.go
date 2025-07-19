package sqlite

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// TestSQLiteIntegration_RealDatabase tests actual SQLite operations
func TestSQLiteIntegration_RealDatabase(t *testing.T) {
	// Create temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Create SQLite repository
	sqliteRepo, err := NewMetadataRepository(dbPath)
	require.NoError(t, err, "Failed to create SQLite repository")
	defer func() {
		if db := sqliteRepo.GetDB(); db != nil {
			db.Close()
		}
	}()

	// Initialize schema (migrations)
	err = sqliteRepo.Migrate()
	require.NoError(t, err, "Failed to migrate SQLite schema")

	ctx := context.Background()

	// Run subtests
	t.Run("InstrumentOperations", func(t *testing.T) {
		testInstrumentOperations(t, sqliteRepo, ctx)
	})

	t.Run("SubscriptionOperations", func(t *testing.T) {
		testSubscriptionOperations(t, sqliteRepo, ctx)
	})

	t.Run("BrokerOperations", func(t *testing.T) {
		testBrokerOperations(t, sqliteRepo, ctx)
	})
}

func testInstrumentOperations(t *testing.T, repo *MetadataRepository, ctx context.Context) {
	// Test data
	testInstrument := entities.InstrumentInfo{
		Symbol:            "BTCUSDT_TEST",
		BaseAsset:         "BTC",
		QuoteAsset:        "USDT",
		Type:              entities.InstrumentTypeSpot,
		Market:            entities.MarketTypeSpot,
		IsActive:          true,
		MinPrice:          0.01,
		MaxPrice:          1000000.0,
		MinQuantity:       0.001,
		MaxQuantity:       10000.0,
		PricePrecision:    2,
		QuantityPrecision: 3,
	}

	// Test adding instrument
	err := repo.SaveInstrument(ctx, testInstrument)
	require.NoError(t, err)

	// Test getting instrument
	retrieved, err := repo.GetInstrument(ctx, testInstrument.Symbol)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, testInstrument.Symbol, retrieved.Symbol)
	assert.Equal(t, testInstrument.BaseAsset, retrieved.BaseAsset)
	assert.Equal(t, testInstrument.QuoteAsset, retrieved.QuoteAsset)
	assert.Equal(t, testInstrument.Type, retrieved.Type)
	assert.Equal(t, testInstrument.Market, retrieved.Market)
	assert.Equal(t, testInstrument.IsActive, retrieved.IsActive)

	// Test listing instruments
	instruments, err := repo.ListInstruments(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(instruments), 1)

	// Find our test instrument
	found := false
	for _, instr := range instruments {
		if instr.Symbol == testInstrument.Symbol {
			found = true
			break
		}
	}
	assert.True(t, found, "Test instrument should be in the list")

	// Test updating instrument (add again with same symbol should update)
	updatedInstrument := testInstrument
	updatedInstrument.IsActive = false

	err = repo.SaveInstrument(ctx, updatedInstrument)
	require.NoError(t, err)

	// Verify update
	retrieved, err = repo.GetInstrument(ctx, testInstrument.Symbol)
	require.NoError(t, err)
	assert.False(t, retrieved.IsActive)

	// Clean up
	cleanupInstrument(t, repo, testInstrument.Symbol)
}

func testSubscriptionOperations(t *testing.T, repo *MetadataRepository, ctx context.Context) {
	// First create a test instrument
	testInstrument := entities.InstrumentInfo{
		Symbol:            "ETHUSDT_TEST",
		BaseAsset:         "ETH",
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

	err := repo.SaveInstrument(ctx, testInstrument)
	require.NoError(t, err)

	// Test data
	testSubscription := entities.InstrumentSubscription{
		ID:        "sub_test_123",
		Symbol:    testInstrument.Symbol,
		Type:      entities.InstrumentTypeSpot,
		Market:    entities.MarketTypeSpot,
		DataTypes: []entities.DataType{entities.DataTypeTicker, entities.DataTypeCandle},
		StartDate: time.Now().UTC(),
		BrokerID:  "binance",
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Test adding subscription
	err = repo.SaveSubscription(ctx, testSubscription)
	require.NoError(t, err)

	// Test getting subscription
	retrieved, err := repo.GetSubscription(ctx, testSubscription.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, testSubscription.ID, retrieved.ID)
	assert.Equal(t, testSubscription.Symbol, retrieved.Symbol)
	assert.Equal(t, testSubscription.Type, retrieved.Type)
	assert.Equal(t, testSubscription.Market, retrieved.Market)
	assert.Equal(t, testSubscription.BrokerID, retrieved.BrokerID)
	assert.Equal(t, testSubscription.IsActive, retrieved.IsActive)

	// Test listing subscriptions
	subscriptions, err := repo.ListSubscriptions(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(subscriptions), 1)

	// Find our test subscription
	found := false
	for _, sub := range subscriptions {
		if sub.ID == testSubscription.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Test subscription should be in the list")

	// Test removing subscription
	err = repo.DeleteSubscription(ctx, testSubscription.ID)
	require.NoError(t, err)

	// Verify removal
	retrieved, err = repo.GetSubscription(ctx, testSubscription.ID)
	assert.Error(t, err) // Should return error for non-existent subscription
	assert.Nil(t, retrieved)

	// Clean up
	cleanupInstrument(t, repo, testInstrument.Symbol)
}

func testBrokerOperations(t *testing.T, repo *MetadataRepository, ctx context.Context) {
	// Test data - using BrokerConfig which is what the repository actually stores
	testBrokerConfig := interfaces.BrokerConfig{
		ID:      "test_broker",
		Name:    "Test Broker",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: true,
		Connection: interfaces.ConnectionConfig{
			WebSocketURL: "wss://ws.testbroker.com",
			RestAPIURL:   "https://api.testbroker.com",
			Timeout:      30 * time.Second,
		},
		Auth: interfaces.AuthConfig{
			APIKey:    "test_key",
			SecretKey: "test_secret",
		},
		Settings: map[string]interface{}{
			"test_setting": "test_value",
		},
	}

	// Test adding broker config
	err := repo.SaveBrokerConfig(ctx, testBrokerConfig)
	require.NoError(t, err)

	// Test getting broker config
	retrieved, err := repo.GetBrokerConfig(ctx, testBrokerConfig.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, testBrokerConfig.ID, retrieved.ID)
	assert.Equal(t, testBrokerConfig.Name, retrieved.Name)
	assert.Equal(t, testBrokerConfig.Type, retrieved.Type)
	assert.Equal(t, testBrokerConfig.Enabled, retrieved.Enabled)

	// Test listing broker configs
	brokerConfigs, err := repo.ListBrokerConfigs(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(brokerConfigs), 1)

	// Find our test broker config
	found := false
	for _, config := range brokerConfigs {
		if config.ID == testBrokerConfig.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "Test broker config should be in the list")

	// Test updating broker config
	updatedBrokerConfig := testBrokerConfig
	updatedBrokerConfig.Enabled = false

	err = repo.SaveBrokerConfig(ctx, updatedBrokerConfig)
	require.NoError(t, err)

	// Verify update
	retrieved, err = repo.GetBrokerConfig(ctx, testBrokerConfig.ID)
	require.NoError(t, err)
	assert.False(t, retrieved.Enabled)

	// Clean up
	cleanupBrokerConfig(t, repo, testBrokerConfig.ID)
}

// Helper functions for cleanup
func cleanupInstrument(t *testing.T, repo *MetadataRepository, symbol string) {
	db := repo.GetDB()
	if db != nil {
		_, err := db.Exec("DELETE FROM instruments WHERE symbol = ?", symbol)
		if err != nil {
			t.Logf("Cleanup warning for instrument %s: %v", symbol, err)
		}
	}
}

func cleanupBrokerConfig(t *testing.T, repo *MetadataRepository, brokerID string) {
	db := repo.GetDB()
	if db != nil {
		_, err := db.Exec("DELETE FROM broker_configs WHERE id = ?", brokerID)
		if err != nil {
			t.Logf("Cleanup warning for broker config %s: %v", brokerID, err)
		}
	}
}
