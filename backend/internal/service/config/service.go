package config

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"m-data-storage/api/dto"
	internalerrors "m-data-storage/internal/errors"
	"m-data-storage/pkg/broker"
)

// ErrNotFound - error when configuration is not found
var ErrNotFound = errors.New("configuration not found")

// Service - service for working with configuration
type Service struct {
	repo Repository
	db   *sql.DB
}

// NewService - creates a new service instance
func NewService(repo Repository) (*Service, error) {
	if repo == nil {
		return nil, errors.New("repository is required")
	}

	return &Service{
		repo: repo,
	}, nil
}

// Close closes the database connection
func (s *Service) Close() error {
	return s.db.Close()
}

// GetBrokerConfig implements service.ConfigService
func (s *Service) GetBrokerConfig(ctx context.Context, brokerID string) (broker.BrokerConfig, error) {
	return s.repo.GetBrokerConfig(ctx, brokerID)
}

// SetBrokerConfig implements service.ConfigService
func (s *Service) SetBrokerConfig(ctx context.Context, config broker.BrokerConfig) error {
	// Validate configuration
	if err := validateBrokerConfig(config); err != nil {
		return fmt.Errorf("invalid broker config: %w", err)
	}

	return s.repo.SetBrokerConfig(ctx, config)
}

// ListBrokerConfigs implements service.ConfigService
func (s *Service) ListBrokerConfigs(ctx context.Context) ([]broker.BrokerConfig, error) {
	return s.repo.ListBrokerConfigs(ctx)
}

// GetSystemConfig implements service.ConfigService
func (s *Service) GetSystemConfig(ctx context.Context) (dto.SystemConfig, error) {
	cfg, err := s.repo.GetSystemConfig(ctx)
	if err != nil {
		if errors.Is(err, internalerrors.ErrNotFound) {
			// Return default configuration
			return dto.SystemConfig{
				StorageRetention: 24 * time.Hour,
				VacuumInterval:   1 * time.Hour,
				MaxStorageSize:   1024 * 1024 * 1024, // 1GB
				APIPort:          8080,
				APIHost:          "localhost",
				ReadTimeout:      30 * time.Second,
				WriteTimeout:     30 * time.Second,
				ShutdownTimeout:  10 * time.Second,
				JWTSecret:        "default-secret-key-must-be-changed-in-prod",
				APIKeyHeader:     "X-API-Key",
				AllowedOrigins:   []string{"http://localhost:3000"},
				MetricsEnabled:   true,
				MetricsPort:      9090,
				TracingEnabled:   false,
			}, nil
		}
		return dto.SystemConfig{}, fmt.Errorf("failed to get system config: %w", err)
	}
	return cfg, nil
}

// UpdateSystemConfig implements service.ConfigService
func (s *Service) UpdateSystemConfig(ctx context.Context, cfg dto.SystemConfig) error {
	// Validate configuration
	if err := validateSystemConfig(cfg); err != nil {
		return fmt.Errorf("invalid system config: %w", err)
	}

	return s.repo.UpdateSystemConfig(ctx, cfg)
}

// initSchema initializes database schema
func initSchema(db *sql.DB) error {
	// Read schema file
	schemaPath := filepath.Join("internal", "service", "config", "schema.sql")
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute schema
	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// validateSystemConfig - validates system configuration correctness
func validateSystemConfig(cfg dto.SystemConfig) error {
	if cfg.StorageRetention <= 0 {
		return errors.New("storage retention must be positive")
	}
	if cfg.VacuumInterval <= 0 {
		return errors.New("vacuum interval must be positive")
	}
	if cfg.MaxStorageSize <= 0 {
		return errors.New("max storage size must be positive")
	}
	if cfg.APIPort <= 0 || cfg.APIPort > 65535 {
		return errors.New("invalid API port")
	}
	if cfg.APIHost == "" {
		return errors.New("API host is required")
	}
	if cfg.ReadTimeout <= 0 {
		return errors.New("read timeout must be positive")
	}
	if cfg.WriteTimeout <= 0 {
		return errors.New("write timeout must be positive")
	}
	if cfg.ShutdownTimeout <= 0 {
		return errors.New("shutdown timeout must be positive")
	}
	if len(cfg.JWTSecret) < 32 {
		return errors.New("JWT secret must be at least 32 characters")
	}
	if cfg.APIKeyHeader == "" {
		return errors.New("API key header is required")
	}
	if len(cfg.AllowedOrigins) == 0 {
		return errors.New("at least one allowed origin is required")
	}
	if cfg.MetricsEnabled && (cfg.MetricsPort <= 0 || cfg.MetricsPort > 65535) {
		return errors.New("invalid metrics port")
	}

	return nil
}

// validateBrokerConfig - validates broker configuration correctness
func validateBrokerConfig(cfg broker.BrokerConfig) error {
	if cfg.ID == "" {
		return errors.New("broker ID is required")
	}
	if cfg.Name == "" {
		return errors.New("broker name is required")
	}
	if cfg.Type != broker.BrokerTypeCrypto && cfg.Type != broker.BrokerTypeStock {
		return errors.New("invalid broker type")
	}
	if cfg.Connection.WebSocketURL == "" {
		return errors.New("WebSocket URL is required")
	}
	if cfg.Connection.RestAPIURL == "" {
		return errors.New("REST API URL is required")
	}
	if cfg.Connection.Timeout <= 0 {
		return errors.New("connection timeout must be positive")
	}
	if cfg.Connection.PingInterval <= 0 {
		return errors.New("ping interval must be positive")
	}
	if cfg.Connection.ReconnectDelay <= 0 {
		return errors.New("reconnect delay must be positive")
	}
	if cfg.Connection.MaxReconnectAttempts <= 0 {
		return errors.New("max reconnect attempts must be positive")
	}
	if cfg.Auth.APIKey == "" {
		return errors.New("API key is required")
	}
	if cfg.Auth.SecretKey == "" {
		return errors.New("secret key is required")
	}

	return nil
}
