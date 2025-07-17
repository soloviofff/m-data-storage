package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"m-data-storage/api/dto"
	internalerrors "m-data-storage/internal/errors"
	"m-data-storage/pkg/broker"
)

// Repository - interface for configuration storage operations
type Repository interface {
	GetSystemConfig(ctx context.Context) (dto.SystemConfig, error)
	UpdateSystemConfig(ctx context.Context, config dto.SystemConfig) error
	GetBrokerConfig(ctx context.Context, id string) (broker.BrokerConfig, error)
	SetBrokerConfig(ctx context.Context, config broker.BrokerConfig) error
	ListBrokerConfigs(ctx context.Context) ([]broker.BrokerConfig, error)
}

// repository - implementation of Repository interface
type repository struct {
	db *sql.DB
}

// NewRepository - creates a new repository instance
func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// GetSystemConfig - retrieves system configuration
func (r *repository) GetSystemConfig(ctx context.Context) (dto.SystemConfig, error) {
	query := `SELECT 
		storage_retention, vacuum_interval, max_storage_size,
		api_port, api_host, read_timeout, write_timeout, shutdown_timeout,
		jwt_secret, api_key_header, allowed_origins, metrics_enabled,
		metrics_port, tracing_enabled
	FROM system_config LIMIT 1`

	var (
		storageRetention int64
		vacuumInterval   int64
		maxStorageSize   int64
		readTimeout      int64
		writeTimeout     int64
		shutdownTimeout  int64
		allowedOrigins   string
		config           dto.SystemConfig
	)

	err := r.db.QueryRowContext(ctx, query).Scan(
		&storageRetention,
		&vacuumInterval,
		&maxStorageSize,
		&config.APIPort,
		&config.APIHost,
		&readTimeout,
		&writeTimeout,
		&shutdownTimeout,
		&config.JWTSecret,
		&config.APIKeyHeader,
		&allowedOrigins,
		&config.MetricsEnabled,
		&config.MetricsPort,
		&config.TracingEnabled,
	)

	if err == sql.ErrNoRows {
		return dto.SystemConfig{}, internalerrors.ErrNotFound
	}
	if err != nil {
		return dto.SystemConfig{}, fmt.Errorf("failed to get system config: %w", err)
	}

	// Convert seconds to durations
	config.StorageRetention = time.Duration(storageRetention) * time.Second
	config.VacuumInterval = time.Duration(vacuumInterval) * time.Second
	config.MaxStorageSize = maxStorageSize
	config.ReadTimeout = time.Duration(readTimeout) * time.Second
	config.WriteTimeout = time.Duration(writeTimeout) * time.Second
	config.ShutdownTimeout = time.Duration(shutdownTimeout) * time.Second

	// Parse allowed origins
	if err := json.Unmarshal([]byte(allowedOrigins), &config.AllowedOrigins); err != nil {
		return dto.SystemConfig{}, fmt.Errorf("failed to parse allowed origins: %w", err)
	}

	return config, nil
}

// UpdateSystemConfig - updates system configuration
func (r *repository) UpdateSystemConfig(ctx context.Context, cfg dto.SystemConfig) error {
	// Convert durations to seconds for storage
	storageRetention := int64(cfg.StorageRetention.Seconds())
	vacuumInterval := int64(cfg.VacuumInterval.Seconds())
	readTimeout := int64(cfg.ReadTimeout.Seconds())
	writeTimeout := int64(cfg.WriteTimeout.Seconds())
	shutdownTimeout := int64(cfg.ShutdownTimeout.Seconds())

	// Convert allowed origins to JSON
	allowedOrigins, err := json.Marshal(cfg.AllowedOrigins)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed origins: %w", err)
	}

	query := `INSERT OR REPLACE INTO system_config (
		storage_retention, vacuum_interval, max_storage_size,
		api_port, api_host, read_timeout, write_timeout, shutdown_timeout,
		jwt_secret, api_key_header, allowed_origins, metrics_enabled,
		metrics_port, tracing_enabled
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		storageRetention,
		vacuumInterval,
		cfg.MaxStorageSize,
		cfg.APIPort,
		cfg.APIHost,
		readTimeout,
		writeTimeout,
		shutdownTimeout,
		cfg.JWTSecret,
		cfg.APIKeyHeader,
		string(allowedOrigins),
		cfg.MetricsEnabled,
		cfg.MetricsPort,
		cfg.TracingEnabled,
	)

	if err != nil {
		return fmt.Errorf("failed to update system config: %w", err)
	}

	return nil
}

// GetBrokerConfig - retrieves broker configuration
func (r *repository) GetBrokerConfig(ctx context.Context, brokerID string) (broker.BrokerConfig, error) {
	query := `SELECT config_json FROM broker_config WHERE id = ?`

	var configJSON string
	err := r.db.QueryRowContext(ctx, query, brokerID).Scan(&configJSON)
	if err == sql.ErrNoRows {
		return broker.BrokerConfig{}, internalerrors.ErrNotFound
	}
	if err != nil {
		return broker.BrokerConfig{}, fmt.Errorf("failed to get broker config: %w", err)
	}

	var config broker.BrokerConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return broker.BrokerConfig{}, fmt.Errorf("failed to parse broker config: %w", err)
	}

	return config, nil
}

// SetBrokerConfig - saves broker configuration
func (r *repository) SetBrokerConfig(ctx context.Context, cfg broker.BrokerConfig) error {
	configJSON, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal broker config: %w", err)
	}

	query := `INSERT OR REPLACE INTO broker_config (id, name, type, enabled, config_json) VALUES (?, ?, ?, ?, ?)`
	_, err = r.db.ExecContext(ctx, query, cfg.ID, cfg.Name, string(cfg.Type), cfg.Enabled, string(configJSON))
	if err != nil {
		return fmt.Errorf("failed to set broker config: %w", err)
	}

	return nil
}

// ListBrokerConfigs - retrieves list of all broker configurations
func (r *repository) ListBrokerConfigs(ctx context.Context) ([]broker.BrokerConfig, error) {
	query := `SELECT config_json FROM broker_config`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list broker configs: %w", err)
	}
	defer rows.Close()

	var configs []broker.BrokerConfig
	for rows.Next() {
		var configJSON string
		if err := rows.Scan(&configJSON); err != nil {
			return nil, fmt.Errorf("failed to scan broker config: %w", err)
		}

		var config broker.BrokerConfig
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return nil, fmt.Errorf("failed to parse broker config: %w", err)
		}

		configs = append(configs, config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating broker configs: %w", err)
	}

	return configs, nil
}
