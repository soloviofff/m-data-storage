package sqlite

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// createTestRepository creates a repository with embedded schema for testing
func createTestRepository(dbPath string) (*MetadataRepository, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, err
	}

	// Create schema directly
	schema := `
	CREATE TABLE IF NOT EXISTS instruments (
		symbol TEXT PRIMARY KEY,
		base_asset TEXT NOT NULL,
		quote_asset TEXT NOT NULL,
		type TEXT NOT NULL,
		market TEXT NOT NULL,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		min_price REAL NOT NULL DEFAULT 0,
		max_price REAL NOT NULL DEFAULT 0,
		min_quantity REAL NOT NULL DEFAULT 0,
		max_quantity REAL NOT NULL DEFAULT 0,
		price_precision INTEGER NOT NULL DEFAULT 8,
		quantity_precision INTEGER NOT NULL DEFAULT 8,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS subscriptions (
		id TEXT PRIMARY KEY,
		symbol TEXT NOT NULL,
		type TEXT NOT NULL,
		market TEXT NOT NULL,
		data_types TEXT NOT NULL,
		start_date DATETIME NOT NULL,
		end_date DATETIME,
		settings TEXT,
		broker_id TEXT NOT NULL,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS broker_configs (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL,
		enabled BOOLEAN NOT NULL DEFAULT 1,
		config_json TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}

	return &MetadataRepository{db: db}, nil
}

func TestMetadataRepository_InstrumentOperations(t *testing.T) {
	// Create temporary database
	dbPath := "/tmp/test_metadata.db"
	defer os.Remove(dbPath)

	repo, err := createTestRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Disconnect()

	ctx := context.Background()

	// Test instrument
	instrument := entities.InstrumentInfo{
		Symbol:            "BTCUSDT",
		BaseAsset:         "BTC",
		QuoteAsset:        "USDT",
		Type:              entities.InstrumentTypeSpot,
		Market:            entities.MarketTypeSpot,
		IsActive:          true,
		MinPrice:          0.01,
		MaxPrice:          1000000,
		MinQuantity:       0.001,
		MaxQuantity:       1000,
		PricePrecision:    2,
		QuantityPrecision: 3,
	}

	// Test SaveInstrument
	err = repo.SaveInstrument(ctx, instrument)
	if err != nil {
		t.Fatalf("Failed to save instrument: %v", err)
	}

	// Test GetInstrument
	retrieved, err := repo.GetInstrument(ctx, "BTCUSDT")
	if err != nil {
		t.Fatalf("Failed to get instrument: %v", err)
	}

	if retrieved.Symbol != instrument.Symbol {
		t.Errorf("Expected symbol %s, got %s", instrument.Symbol, retrieved.Symbol)
	}

	if retrieved.Type != instrument.Type {
		t.Errorf("Expected type %s, got %s", instrument.Type, retrieved.Type)
	}

	// Test ListInstruments
	instruments, err := repo.ListInstruments(ctx)
	if err != nil {
		t.Fatalf("Failed to list instruments: %v", err)
	}

	if len(instruments) != 1 {
		t.Errorf("Expected 1 instrument, got %d", len(instruments))
	}

	// Test DeleteInstrument
	err = repo.DeleteInstrument(ctx, "BTCUSDT")
	if err != nil {
		t.Fatalf("Failed to delete instrument: %v", err)
	}

	// Verify deletion
	instruments, err = repo.ListInstruments(ctx)
	if err != nil {
		t.Fatalf("Failed to list instruments after deletion: %v", err)
	}

	if len(instruments) != 0 {
		t.Errorf("Expected 0 instruments after deletion, got %d", len(instruments))
	}
}

func TestMetadataRepository_SubscriptionOperations(t *testing.T) {
	// Create temporary database
	dbPath := "/tmp/test_metadata_sub.db"
	defer os.Remove(dbPath)

	repo, err := createTestRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Disconnect()

	ctx := context.Background()

	// Test subscription
	subscription := entities.InstrumentSubscription{
		ID:        "sub-1",
		Symbol:    "BTCUSDT",
		Type:      entities.InstrumentTypeSpot,
		Market:    entities.MarketTypeSpot,
		DataTypes: []entities.DataType{entities.DataTypeTicker, entities.DataTypeCandle},
		StartDate: time.Now(),
		BrokerID:  "binance",
		IsActive:  true,
		Settings: map[string]interface{}{
			"timeframes": []string{"1m", "5m", "1h"},
		},
	}

	// Test SaveSubscription
	err = repo.SaveSubscription(ctx, subscription)
	if err != nil {
		t.Fatalf("Failed to save subscription: %v", err)
	}

	// Test GetSubscription
	retrieved, err := repo.GetSubscription(ctx, "sub-1")
	if err != nil {
		t.Fatalf("Failed to get subscription: %v", err)
	}

	if retrieved.ID != subscription.ID {
		t.Errorf("Expected ID %s, got %s", subscription.ID, retrieved.ID)
	}

	if len(retrieved.DataTypes) != len(subscription.DataTypes) {
		t.Errorf("Expected %d data types, got %d", len(subscription.DataTypes), len(retrieved.DataTypes))
	}

	// Test ListSubscriptions
	subscriptions, err := repo.ListSubscriptions(ctx)
	if err != nil {
		t.Fatalf("Failed to list subscriptions: %v", err)
	}

	if len(subscriptions) != 1 {
		t.Errorf("Expected 1 subscription, got %d", len(subscriptions))
	}

	// Test DeleteSubscription
	err = repo.DeleteSubscription(ctx, "sub-1")
	if err != nil {
		t.Fatalf("Failed to delete subscription: %v", err)
	}

	// Verify deletion
	subscriptions, err = repo.ListSubscriptions(ctx)
	if err != nil {
		t.Fatalf("Failed to list subscriptions after deletion: %v", err)
	}

	if len(subscriptions) != 0 {
		t.Errorf("Expected 0 subscriptions after deletion, got %d", len(subscriptions))
	}
}

func TestMetadataRepository_BrokerConfigOperations(t *testing.T) {
	// Create temporary database
	dbPath := "/tmp/test_metadata_broker.db"
	defer os.Remove(dbPath)

	repo, err := createTestRepository(dbPath)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Disconnect()

	ctx := context.Background()

	// Test broker config
	config := interfaces.BrokerConfig{
		ID:      "binance",
		Name:    "Binance",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: true,
		Connection: interfaces.ConnectionConfig{
			WebSocketURL: "wss://stream.binance.com:9443/ws",
			RestAPIURL:   "https://api.binance.com",
		},
		Auth: interfaces.AuthConfig{
			APIKey:    "test-key",
			SecretKey: "test-secret",
		},
		Settings: map[string]interface{}{
			"test": "value",
		},
	}

	// Test SaveBrokerConfig
	err = repo.SaveBrokerConfig(ctx, config)
	if err != nil {
		t.Fatalf("Failed to save broker config: %v", err)
	}

	// Test GetBrokerConfig
	retrieved, err := repo.GetBrokerConfig(ctx, "binance")
	if err != nil {
		t.Fatalf("Failed to get broker config: %v", err)
	}

	if retrieved.ID != config.ID {
		t.Errorf("Expected ID %s, got %s", config.ID, retrieved.ID)
	}

	if retrieved.Type != config.Type {
		t.Errorf("Expected type %s, got %s", config.Type, retrieved.Type)
	}

	// Test ListBrokerConfigs
	configs, err := repo.ListBrokerConfigs(ctx)
	if err != nil {
		t.Fatalf("Failed to list broker configs: %v", err)
	}

	if len(configs) != 1 {
		t.Errorf("Expected 1 broker config, got %d", len(configs))
	}

	// Test DeleteBrokerConfig
	err = repo.DeleteBrokerConfig(ctx, "binance")
	if err != nil {
		t.Fatalf("Failed to delete broker config: %v", err)
	}

	// Verify deletion
	configs, err = repo.ListBrokerConfigs(ctx)
	if err != nil {
		t.Fatalf("Failed to list broker configs after deletion: %v", err)
	}

	if len(configs) != 0 {
		t.Errorf("Expected 0 broker configs after deletion, got %d", len(configs))
	}
}
