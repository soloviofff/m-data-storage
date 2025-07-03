package config

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"m-data-storage/api/dto"
	"m-data-storage/pkg/broker"
)

// Repository handles configuration storage operations
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new configuration repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetSystemConfig retrieves system configuration
func (r *Repository) GetSystemConfig(ctx context.Context) (*dto.SystemConfig, error) {
	query := `SELECT 
		storage_retention, vacuum_interval, max_storage_size,
		api_port, api_host, read_timeout, write_timeout, shutdown_timeout,
		jwt_secret, api_key_header, allowed_origins,
		metrics_enabled, metrics_port, tracing_enabled
		FROM system_config WHERE id = 1`

	var allowedOriginsJSON string
	var cfg dto.SystemConfig

	err := r.db.QueryRowContext(ctx, query).Scan(
		&cfg.StorageRetention,
		&cfg.VacuumInterval,
		&cfg.MaxStorageSize,
		&cfg.APIPort,
		&cfg.APIHost,
		&cfg.ReadTimeout,
		&cfg.WriteTimeout,
		&cfg.ShutdownTimeout,
		&cfg.JWTSecret,
		&cfg.APIKeyHeader,
		&allowedOriginsJSON,
		&cfg.MetricsEnabled,
		&cfg.MetricsPort,
		&cfg.TracingEnabled,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("system configuration not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get system config: %w", err)
	}

	// Parse allowed origins JSON
	if err := json.Unmarshal([]byte(allowedOriginsJSON), &cfg.AllowedOrigins); err != nil {
		return nil, fmt.Errorf("failed to parse allowed origins: %w", err)
	}

	return &cfg, nil
}

// UpdateSystemConfig updates system configuration
func (r *Repository) UpdateSystemConfig(ctx context.Context, cfg dto.SystemConfig) error {
	// Convert durations to seconds for storage
	storageRetention := int64(cfg.StorageRetention.Seconds())
	vacuumInterval := int64(cfg.VacuumInterval.Seconds())
	readTimeout := int64(cfg.ReadTimeout.Seconds())
	writeTimeout := int64(cfg.WriteTimeout.Seconds())
	shutdownTimeout := int64(cfg.ShutdownTimeout.Seconds())

	// Convert allowed origins to JSON
	allowedOriginsJSON, err := json.Marshal(cfg.AllowedOrigins)
	if err != nil {
		return fmt.Errorf("failed to marshal allowed origins: %w", err)
	}

	query := `INSERT INTO system_config (
		id, storage_retention, vacuum_interval, max_storage_size,
		api_port, api_host, read_timeout, write_timeout, shutdown_timeout,
		jwt_secret, api_key_header, allowed_origins,
		metrics_enabled, metrics_port, tracing_enabled
	) VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		storage_retention = excluded.storage_retention,
		vacuum_interval = excluded.vacuum_interval,
		max_storage_size = excluded.max_storage_size,
		api_port = excluded.api_port,
		api_host = excluded.api_host,
		read_timeout = excluded.read_timeout,
		write_timeout = excluded.write_timeout,
		shutdown_timeout = excluded.shutdown_timeout,
		jwt_secret = excluded.jwt_secret,
		api_key_header = excluded.api_key_header,
		allowed_origins = excluded.allowed_origins,
		metrics_enabled = excluded.metrics_enabled,
		metrics_port = excluded.metrics_port,
		tracing_enabled = excluded.tracing_enabled`

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
		allowedOriginsJSON,
		cfg.MetricsEnabled,
		cfg.MetricsPort,
		cfg.TracingEnabled,
	)

	if err != nil {
		return fmt.Errorf("failed to update system config: %w", err)
	}

	return nil
}

// GetBrokerConfig retrieves broker configuration
func (r *Repository) GetBrokerConfig(ctx context.Context, brokerID string) (*broker.BrokerConfig, error) {
	query := `SELECT config_json FROM broker_config WHERE id = ?`

	var configJSON string
	err := r.db.QueryRowContext(ctx, query, brokerID).Scan(&configJSON)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("broker configuration not found: %s", brokerID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get broker config: %w", err)
	}

	var cfg broker.BrokerConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse broker config: %w", err)
	}

	return &cfg, nil
}

// SetBrokerConfig saves broker configuration
func (r *Repository) SetBrokerConfig(ctx context.Context, cfg broker.BrokerConfig) error {
	configJSON, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal broker config: %w", err)
	}

	query := `INSERT INTO broker_config (id, name, type, enabled, config_json)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			type = excluded.type,
			enabled = excluded.enabled,
			config_json = excluded.config_json`

	_, err = r.db.ExecContext(ctx, query,
		cfg.ID,
		cfg.Name,
		cfg.Type,
		cfg.Enabled,
		configJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to set broker config: %w", err)
	}

	return nil
}

// ListBrokerConfigs retrieves all broker configurations
func (r *Repository) ListBrokerConfigs(ctx context.Context) ([]broker.BrokerConfig, error) {
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

		var cfg broker.BrokerConfig
		if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse broker config: %w", err)
		}

		configs = append(configs, cfg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating broker configs: %w", err)
	}

	return configs, nil
}
