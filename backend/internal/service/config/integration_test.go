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

func setupIntegrationTest(t *testing.T) (*Service, func()) {
	// Create temporary database
	dbFile := "integration_test.db"
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

	// Create repository and service
	repo := NewRepository(db)
	service, err := NewService(repo)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		db.Close()
		os.Remove(dbFile)
	}

	return service, cleanup
}

func TestIntegration_SystemConfig(t *testing.T) {
	service, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	// Test default config
	defaultCfg, err := service.GetSystemConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get default system config: %v", err)
	}

	// Verify default values
	if defaultCfg.StorageRetention != 24*time.Hour {
		t.Errorf("Default StorageRetention mismatch: got %v, want %v", defaultCfg.StorageRetention, 24*time.Hour)
	}
	if defaultCfg.APIPort != 8080 {
		t.Errorf("Default APIPort mismatch: got %v, want %v", defaultCfg.APIPort, 8080)
	}

	// Test update config
	newCfg := dto.SystemConfig{
		StorageRetention: 48 * time.Hour,
		VacuumInterval:   2 * time.Hour,
		MaxStorageSize:   2048 * 1024 * 1024,
		APIPort:          9090,
		APIHost:          "0.0.0.0",
		ReadTimeout:      60 * time.Second,
		WriteTimeout:     60 * time.Second,
		ShutdownTimeout:  20 * time.Second,
		JWTSecret:        "new-test-secret-key-must-be-at-least-32-chars",
		APIKeyHeader:     "X-Custom-API-Key",
		AllowedOrigins:   []string{"https://app.example.com"},
		MetricsEnabled:   true,
		MetricsPort:      9091,
		TracingEnabled:   true,
	}

	if err := service.UpdateSystemConfig(ctx, newCfg); err != nil {
		t.Fatalf("Failed to update system config: %v", err)
	}

	// Verify updated config
	updatedCfg, err := service.GetSystemConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get updated system config: %v", err)
	}

	if updatedCfg.StorageRetention != newCfg.StorageRetention {
		t.Errorf("Updated StorageRetention mismatch: got %v, want %v", updatedCfg.StorageRetention, newCfg.StorageRetention)
	}
	if updatedCfg.APIPort != newCfg.APIPort {
		t.Errorf("Updated APIPort mismatch: got %v, want %v", updatedCfg.APIPort, newCfg.APIPort)
	}
}

func TestIntegration_BrokerConfig(t *testing.T) {
	service, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	// Test initial state
	configs, err := service.ListBrokerConfigs(ctx)
	if err != nil {
		t.Fatalf("Failed to list broker configs: %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("Expected empty broker configs list, got %d items", len(configs))
	}

	// Test adding broker config
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

	if err := service.SetBrokerConfig(ctx, cfg); err != nil {
		t.Fatalf("Failed to set broker config: %v", err)
	}

	// Test retrieving broker config
	retrieved, err := service.GetBrokerConfig(ctx, cfg.ID)
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

	// Test listing broker configs
	configs, err = service.ListBrokerConfigs(ctx)
	if err != nil {
		t.Fatalf("Failed to list broker configs: %v", err)
	}
	if len(configs) != 1 {
		t.Errorf("Expected 1 broker config, got %d", len(configs))
	}

	// Test updating broker config
	cfg.Name = "Updated Test Broker"
	if err := service.SetBrokerConfig(ctx, cfg); err != nil {
		t.Fatalf("Failed to update broker config: %v", err)
	}

	// Verify update
	updated, err := service.GetBrokerConfig(ctx, cfg.ID)
	if err != nil {
		t.Fatalf("Failed to get updated broker config: %v", err)
	}
	if updated.Name != cfg.Name {
		t.Errorf("Updated name mismatch: got %v, want %v", updated.Name, cfg.Name)
	}
}

func TestIntegration_Validation(t *testing.T) {
	service, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	// Test invalid system config
	invalidSystemCfg := dto.SystemConfig{
		StorageRetention: -1 * time.Hour, // Negative duration
		APIPort:          0,              // Invalid port
		JWTSecret:        "short",        // Too short secret
	}

	if err := service.UpdateSystemConfig(ctx, invalidSystemCfg); err == nil {
		t.Error("Expected validation error for invalid system config")
	}

	// Test invalid broker config
	invalidBrokerCfg := broker.BrokerConfig{
		ID:   "",                    // Empty ID
		Name: "",                    // Empty name
		Type: "invalid-broker-type", // Invalid type
	}

	if err := service.SetBrokerConfig(ctx, invalidBrokerCfg); err == nil {
		t.Error("Expected validation error for invalid broker config")
	}

	// Test non-existent broker
	_, err := service.GetBrokerConfig(ctx, "non-existent")
	if err == nil {
		t.Error("Expected error when getting non-existent broker config")
	}
}
