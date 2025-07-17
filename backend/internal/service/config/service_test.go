package config

import (
	"context"
	"testing"
	"time"

	"m-data-storage/api/dto"
	internalerrors "m-data-storage/internal/errors"
	"m-data-storage/pkg/broker"
)

// mockRepository - mock repository for testing
type mockRepository struct {
	systemConfig  *dto.SystemConfig
	brokerConfigs map[string]broker.BrokerConfig
	err           error
}

func (m *mockRepository) GetSystemConfig(ctx context.Context) (dto.SystemConfig, error) {
	if m.err != nil {
		return dto.SystemConfig{}, m.err
	}
	if m.systemConfig == nil {
		return dto.SystemConfig{}, internalerrors.ErrNotFound
	}
	return *m.systemConfig, nil
}

func (m *mockRepository) UpdateSystemConfig(ctx context.Context, config dto.SystemConfig) error {
	if m.err != nil {
		return m.err
	}
	m.systemConfig = &config
	return nil
}

func (m *mockRepository) GetBrokerConfig(ctx context.Context, id string) (broker.BrokerConfig, error) {
	if m.err != nil {
		return broker.BrokerConfig{}, m.err
	}
	cfg, exists := m.brokerConfigs[id]
	if !exists {
		return broker.BrokerConfig{}, internalerrors.ErrNotFound
	}
	return cfg, nil
}

func (m *mockRepository) SetBrokerConfig(ctx context.Context, config broker.BrokerConfig) error {
	if m.err != nil {
		return m.err
	}
	if m.brokerConfigs == nil {
		m.brokerConfigs = make(map[string]broker.BrokerConfig)
	}
	m.brokerConfigs[config.ID] = config
	return nil
}

func (m *mockRepository) ListBrokerConfigs(ctx context.Context) ([]broker.BrokerConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	var configs []broker.BrokerConfig
	for _, cfg := range m.brokerConfigs {
		configs = append(configs, cfg)
	}
	return configs, nil
}

func TestService_SystemConfig(t *testing.T) {
	repo := &mockRepository{}
	service, err := NewService(repo)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	ctx := context.Background()

	// Test default config
	defaultCfg, err := service.GetSystemConfig(ctx)
	if err != nil {
		t.Fatalf("Failed to get default system config: %v", err)
	}

	// Verify default values
	expectedStorageRetention := 24 * time.Hour
	expectedAPIPort := 8080

	if defaultCfg.StorageRetention != expectedStorageRetention {
		t.Errorf("Default StorageRetention mismatch: got %v, want %v", defaultCfg.StorageRetention, expectedStorageRetention)
	}
	if defaultCfg.APIPort != expectedAPIPort {
		t.Errorf("Default APIPort mismatch: got %v, want %v", defaultCfg.APIPort, expectedAPIPort)
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

func TestService_BrokerConfig(t *testing.T) {
	repo := &mockRepository{}
	service, err := NewService(repo)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
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
	if err := service.SetBrokerConfig(ctx, cfg); err != nil {
		t.Fatalf("Failed to set broker config: %v", err)
	}

	// Test GetBrokerConfig
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

	// Test ListBrokerConfigs
	configs, err := service.ListBrokerConfigs(ctx)
	if err != nil {
		t.Fatalf("Failed to list broker configs: %v", err)
	}
	if len(configs) != 1 {
		t.Errorf("Expected 1 broker config, got %d", len(configs))
	}
}

func TestService_Validation(t *testing.T) {
	repo := &mockRepository{}
	service, err := NewService(repo)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
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
}
