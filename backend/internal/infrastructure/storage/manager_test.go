package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

func TestNewManager(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test_manager_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := Config{
		SQLite: SQLiteConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		QuestDB: QuestDBConfig{
			Host:     "localhost",
			Port:     8812,
			Database: "qdb",
			Username: "admin",
			Password: "quest",
			SSLMode:  "disable",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager == nil {
		t.Fatal("Manager is nil")
	}

	if manager.metadata == nil {
		t.Error("Metadata storage is nil")
	}

	if manager.timeSeries == nil {
		t.Error("Time series storage is nil")
	}

	if manager.migrations != nil {
		t.Error("Migration manager should be nil before initialization")
	}

	if manager.logger == nil {
		t.Error("Logger is nil")
	}
}

func TestNewManagerWithNilLogger(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test_manager_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := Config{
		SQLite: SQLiteConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		QuestDB: QuestDBConfig{
			Host:     "localhost",
			Port:     8812,
			Database: "qdb",
			Username: "admin",
			Password: "quest",
			SSLMode:  "disable",
		},
	}

	manager, err := NewManager(config, nil)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if manager.logger == nil {
		t.Error("Logger should be created when nil is passed")
	}
}

func TestManagerMigrationMethods(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test_manager_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := Config{
		SQLite: SQLiteConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		QuestDB: QuestDBConfig{
			Host:     "localhost",
			Port:     8812,
			Database: "qdb",
			Username: "admin",
			Password: "quest",
			SSLMode:  "disable",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Test methods before initialization - should fail
	status, err := manager.GetMigrationStatus(ctx)
	if err == nil {
		t.Error("GetMigrationStatus should fail before initialization")
	}
	if status != nil {
		t.Error("Migration status should be nil before initialization")
	}

	err = manager.RunMigrations(ctx)
	if err == nil {
		t.Error("RunMigrations should fail before initialization")
	}

	err = manager.RollbackMigrations(ctx, 0)
	if err == nil {
		t.Error("RollbackMigrations should fail before initialization")
	}

	t.Log("All migration methods correctly fail before initialization")
}

func TestManagerGetters(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test_manager_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := Config{
		SQLite: SQLiteConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		QuestDB: QuestDBConfig{
			Host:     "localhost",
			Port:     8812,
			Database: "qdb",
			Username: "admin",
			Password: "quest",
			SSLMode:  "disable",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test getters
	metadataStorage := manager.GetMetadataStorage()
	if metadataStorage == nil {
		t.Error("GetMetadataStorage returned nil")
	}

	timeSeriesStorage := manager.GetTimeSeriesStorage()
	if timeSeriesStorage == nil {
		t.Error("GetTimeSeriesStorage returned nil")
	}
}

func TestManagerHealth(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test_manager_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := Config{
		SQLite: SQLiteConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		QuestDB: QuestDBConfig{
			Host:     "localhost",
			Port:     8812,
			Database: "qdb",
			Username: "admin",
			Password: "quest",
			SSLMode:  "disable",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test health check
	health := manager.Health()
	if health == nil {
		t.Error("Health check returned nil")
	}

	if _, exists := health["metadata"]; !exists {
		t.Error("Health check missing metadata status")
	}

	if _, exists := health["timeseries"]; !exists {
		t.Error("Health check missing timeseries status")
	}
}

func TestManagerShutdown(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test_manager_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := Config{
		SQLite: SQLiteConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		QuestDB: QuestDBConfig{
			Host:     "localhost",
			Port:     8812,
			Database: "qdb",
			Username: "admin",
			Password: "quest",
			SSLMode:  "disable",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Test shutdown - should not error even if not connected
	err = manager.Shutdown()
	if err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestManagerProxyMethods(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test_manager_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := Config{
		SQLite: SQLiteConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		QuestDB: QuestDBConfig{
			Host:     "localhost",
			Port:     8812,
			Database: "qdb",
			Username: "admin",
			Password: "quest",
			SSLMode:  "disable",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Test time series proxy methods with empty data (should not error)
	err = manager.SaveTickers(ctx, []entities.Ticker{})
	assert.NoError(t, err)

	err = manager.SaveCandles(ctx, []entities.Candle{})
	assert.NoError(t, err)

	err = manager.SaveOrderBooks(ctx, []entities.OrderBook{})
	assert.NoError(t, err)

	// Test time series proxy methods with nil DB (should panic)
	ticker := entities.Ticker{
		Symbol:    "BTCUSD",
		Price:     50000.0,
		Volume:    1.5,
		Market:    entities.MarketTypeSpot,
		Type:      entities.InstrumentTypeSpot,
		Timestamp: time.Now(),
		BrokerID:  "test-broker",
	}

	assert.Panics(t, func() {
		manager.SaveTickers(ctx, []entities.Ticker{ticker})
	})

	assert.Panics(t, func() {
		manager.GetTickers(ctx, interfaces.TickerFilter{})
	})

	assert.Panics(t, func() {
		manager.GetCandles(ctx, interfaces.CandleFilter{})
	})

	assert.Panics(t, func() {
		manager.GetOrderBooks(ctx, interfaces.OrderBookFilter{})
	})

	// Test metadata proxy methods (should fail without initialization)
	instrument := entities.InstrumentInfo{
		Symbol:     "BTCUSD",
		BaseAsset:  "BTC",
		QuoteAsset: "USD",
		Type:       entities.InstrumentTypeSpot,
		Market:     entities.MarketTypeSpot,
		IsActive:   true,
	}

	// These should fail because database is not initialized
	err = manager.SaveInstrument(ctx, instrument)
	assert.Error(t, err)

	retrievedInstrument, err := manager.GetInstrument(ctx, "BTCUSD")
	assert.Error(t, err)
	assert.Nil(t, retrievedInstrument)

	instruments, err := manager.ListInstruments(ctx)
	assert.Error(t, err)
	assert.Nil(t, instruments)

	err = manager.DeleteInstrument(ctx, "BTCUSD")
	assert.Error(t, err)

	// Test subscription proxy methods (should also fail)
	subscription := entities.InstrumentSubscription{
		ID:        "test-sub-1",
		Symbol:    "BTCUSD",
		Type:      entities.InstrumentTypeSpot,
		Market:    entities.MarketTypeSpot,
		DataTypes: []entities.DataType{entities.DataTypeTicker},
		StartDate: time.Now(),
		BrokerID:  "test-broker",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = manager.SaveSubscription(ctx, subscription)
	assert.Error(t, err)

	retrievedSub, err := manager.GetSubscription(ctx, "test-sub-1")
	assert.Error(t, err)
	assert.Nil(t, retrievedSub)

	subscriptions, err := manager.ListSubscriptions(ctx)
	assert.Error(t, err)
	assert.Nil(t, subscriptions)

	err = manager.UpdateSubscription(ctx, subscription)
	assert.Error(t, err)

	err = manager.DeleteSubscription(ctx, "test-sub-1")
	assert.Error(t, err)
}

func TestManagerUtilityMethods(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "test_manager_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := Config{
		SQLite: SQLiteConfig{
			DatabasePath: filepath.Join(tempDir, "test.db"),
		},
		QuestDB: QuestDBConfig{
			Host:     "localhost",
			Port:     8812,
			Database: "qdb",
			Username: "admin",
			Password: "quest",
			SSLMode:  "disable",
		},
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	manager, err := NewManager(config, logger)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	ctx := context.Background()

	// Test CleanupOldData (currently returns nil)
	err = manager.CleanupOldData(ctx, time.Hour*24)
	assert.NoError(t, err)

	// Test GetStats (currently returns empty stats)
	stats, err := manager.GetStats(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}
