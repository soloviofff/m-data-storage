package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m-data-storage/internal/infrastructure/config"
	"m-data-storage/internal/infrastructure/logger"
)

func TestServerIntegration(t *testing.T) {
	// Create test configuration
	cfg := createTestConfig()

	// Create logger
	log, err := logger.New(cfg.Logging)
	require.NoError(t, err)

	// Test container initialization
	ctx := context.Background()
	container, err := initializeContainer(ctx, cfg, log)
	require.NoError(t, err)
	require.NotNil(t, container)

	// Test HTTP server initialization
	httpServer, err := initializeHTTPServer(cfg, container, log)
	require.NoError(t, err)
	require.NotNil(t, httpServer)

	// Check that server is configured correctly
	assert.NotNil(t, httpServer)

	// Shutdown container
	err = container.Shutdown()
	require.NoError(t, err)
}

func TestConfigLoading(t *testing.T) {
	// Test configuration loading
	cfg, err := config.Load("")
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check basic parameters
	assert.NotEmpty(t, cfg.App.Name)
	assert.NotEmpty(t, cfg.App.Version)
	assert.Greater(t, cfg.API.Port, 0)
	assert.Greater(t, cfg.API.ReadTimeout, time.Duration(0))
	assert.Greater(t, cfg.API.WriteTimeout, time.Duration(0))
}

// Helper function for creating test configuration
func createTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:        "m-data-storage-test",
			Version:     "1.0.0",
			Environment: "test",
		},
		API: config.APIConfig{
			Host:            "localhost",
			Port:            8080,
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    30 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			CORS: config.CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{"*"},
			},
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "json",
		},
		Database: config.DatabaseConfig{
			SQLite: config.SQLiteConfig{
				Path:            ":memory:",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: time.Hour,
				WALMode:         true,
				ForeignKeys:     true,
			},
			QuestDB: config.QuestDBConfig{
				Host:            "localhost",
				Port:            8812,
				Database:        "qdb",
				QueryTimeout:    30 * time.Second,
				MaxOpenConns:    20,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Hour,
			},
		},
		Storage: config.StorageConfig{
			BatchSize:       1000,
			FlushInterval:   5 * time.Second,
			RetentionPeriod: 720 * time.Hour,
			VacuumInterval:  24 * time.Hour,
			MaxStorageSize:  10737418240,
		},
		Brokers: config.BrokersConfig{
			ConfigPath:          "./configs/brokers",
			ReconnectDelay:      5 * time.Second,
			MaxReconnects:       10,
			HealthCheckInterval: 30 * time.Second,
		},
	}
}
