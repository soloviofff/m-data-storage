package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"m-data-storage/api/dto"
	"m-data-storage/pkg/broker"
)

// Service implements configuration management
type Service struct {
	repo *Repository
	db   *sql.DB
}

// NewService creates a new configuration service
func NewService(dbPath string) (*Service, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config database: %w", err)
	}

	// Initialize schema
	if err := initSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &Service{
		repo: NewRepository(db),
		db:   db,
	}, nil
}

// Close closes the database connection
func (s *Service) Close() error {
	return s.db.Close()
}

// GetBrokerConfig implements service.ConfigService
func (s *Service) GetBrokerConfig(ctx context.Context, brokerID string) (*broker.BrokerConfig, error) {
	return s.repo.GetBrokerConfig(ctx, brokerID)
}

// SetBrokerConfig implements service.ConfigService
func (s *Service) SetBrokerConfig(ctx context.Context, config broker.BrokerConfig) error {
	return s.repo.SetBrokerConfig(ctx, config)
}

// ListBrokerConfigs implements service.ConfigService
func (s *Service) ListBrokerConfigs(ctx context.Context) ([]broker.BrokerConfig, error) {
	return s.repo.ListBrokerConfigs(ctx)
}

// GetSystemConfig implements service.ConfigService
func (s *Service) GetSystemConfig(ctx context.Context) (*dto.SystemConfig, error) {
	return s.repo.GetSystemConfig(ctx)
}

// UpdateSystemConfig implements service.ConfigService
func (s *Service) UpdateSystemConfig(ctx context.Context, config dto.SystemConfig) error {
	return s.repo.UpdateSystemConfig(ctx, config)
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
