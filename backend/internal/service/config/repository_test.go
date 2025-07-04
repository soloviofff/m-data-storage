package config

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"m-data-storage/api/dto"
	"m-data-storage/pkg/broker"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// Create temporary database
	dbFile := "test.db"
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Initialize schema
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		db.Close()
		os.Remove(dbFile)
	}

	return db, cleanup
}

func TestRepository_SystemConfig(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	// Test data
	cfg := dto.SystemConfig{
		StorageRetention: 24 * time.Hour,
		VacuumInterval:   1 * time.Hour,
		MaxStorageSize:   1024 * 1024 * 1024, // 1GB
		APIPort:          8080,
		APIHost:          "localhost",
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		ShutdownTimeout:  10 * time.Second,
		JWTSecret:        "test-secret-key-must-be-at-least-32-chars",
		APIKeyHeader:     "X-API-Key",
		AllowedOrigins:   []string{"http://localhost:3000"},
		MetricsEnabled:   true,
		MetricsPort:      9090,
		TracingEnabled:   false,
	}

	// Test UpdateSystemConfig
	if err := repo.UpdateSystemConfig(ctx, cfg); err != nil {
		t.Fatalf("Failed to update system config: %v", err)
	}

	// Test GetSystemConfig
	retrieved, err := repo.GetSystemConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get system config: %v", err)
	}

	// Compare values
	if retrieved.StorageRetention != cfg.StorageRetention {
		t.Errorf("StorageRetention mismatch: got %v, want %v", retrieved.StorageRetention, cfg.StorageRetention)
	}
	if retrieved.APIPort != cfg.APIPort {
		t.Errorf("APIPort mismatch: got %v, want %v", retrieved.APIPort, cfg.APIPort)
	}
	if retrieved.JWTSecret != cfg.JWTSecret {
		t.Errorf("JWTSecret mismatch: got %v, want %v", retrieved.JWTSecret, cfg.JWTSecret)
	}
	if len(retrieved.AllowedOrigins) != len(cfg.AllowedOrigins) {
		t.Errorf("AllowedOrigins length mismatch: got %v, want %v", len(retrieved.AllowedOrigins), len(cfg.AllowedOrigins))
	}
}

func TestRepository_BrokerConfig(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	// Test data
	cfg := broker.BrokerConfig{
		ID:      "test-broker",
		Name:    "Test Broker",
		Type:    broker.BrokerTypeCrypto,
		Enabled: true,
		Connection: broker.ConnectionConfig{
			WebSocketURL:         "wss://test.broker/ws",
			RestAPIURL:           "https://test.broker/api",
			Timeout:              5 * time.Second,
			PingInterval:         30 * time.Second,
			ReconnectDelay:       5 * time.Second,
			MaxReconnectAttempts: 3,
		},
		Auth: broker.AuthConfig{
			APIKey:    "test-key",
			SecretKey: "test-secret",
			Sandbox:   true,
		},
	}

	// Test SetBrokerConfig
	if err := repo.SetBrokerConfig(ctx, cfg); err != nil {
		t.Fatalf("Failed to set broker config: %v", err)
	}

	// Test GetBrokerConfig
	retrieved, err := repo.GetBrokerConfig(ctx, cfg.ID)
	if err != nil {
		t.Fatalf("Failed to get broker config: %v", err)
	}

	// Compare values
	if retrieved.ID != cfg.ID {
		t.Errorf("ID mismatch: got %v, want %v", retrieved.ID, cfg.ID)
	}
	if retrieved.Name != cfg.Name {
		t.Errorf("Name mismatch: got %v, want %v", retrieved.Name, cfg.Name)
	}
	if retrieved.Type != cfg.Type {
		t.Errorf("Type mismatch: got %v, want %v", retrieved.Type, cfg.Type)
	}
	if retrieved.Connection.WebSocketURL != cfg.Connection.WebSocketURL {
		t.Errorf("WebSocketURL mismatch: got %v, want %v", retrieved.Connection.WebSocketURL, cfg.Connection.WebSocketURL)
	}

	// Test ListBrokerConfigs
	configs, err := repo.ListBrokerConfigs(ctx)
	if err != nil {
		t.Fatalf("Failed to list broker configs: %v", err)
	}
	if len(configs) != 1 {
		t.Errorf("Expected 1 broker config, got %d", len(configs))
	}
}

func TestRepository_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	ctx := context.Background()

	// Test GetSystemConfig when not exists
	_, err := repo.GetSystemConfig(ctx)
	if err == nil {
		t.Error("Expected error when getting non-existent system config")
	}

	// Test GetBrokerConfig when not exists
	_, err = repo.GetBrokerConfig(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent broker config")
	}

	// Test ListBrokerConfigs when empty
	configs, err := repo.ListBrokerConfigs(ctx)
	if err != nil {
		t.Fatalf("Failed to list broker configs: %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("Expected empty broker configs list, got %d items", len(configs))
	}
}
