package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"

	_ "github.com/mattn/go-sqlite3"
)

// MetadataRepository implements MetadataStorage interface for SQLite
type MetadataRepository struct {
	db *sql.DB
}

// NewMetadataRepository creates a new SQLite metadata repository
func NewMetadataRepository(dbPath string) (*MetadataRepository, error) {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	repo := &MetadataRepository{db: db}

	// Note: Schema initialization is now handled by migrations
	// No longer calling initSchema() here

	return repo, nil
}

// GetDB returns the underlying database connection
func (r *MetadataRepository) GetDB() *sql.DB {
	return r.db
}

// initSchema initializes the database schema
func (r *MetadataRepository) initSchema() error {
	// Try multiple possible paths for schema file
	possiblePaths := []string{
		"internal/infrastructure/storage/sqlite/schema.sql",
		"../../../storage/sqlite/schema.sql",
		"schema.sql",
	}

	var schema []byte
	var err error

	for _, schemaPath := range possiblePaths {
		schema, err = os.ReadFile(schemaPath)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("failed to read schema file from any location: %w", err)
	}

	if _, err := r.db.Exec(string(schema)); err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// Connect establishes database connection
func (r *MetadataRepository) Connect(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

// Disconnect closes database connection
func (r *MetadataRepository) Disconnect() error {
	return r.db.Close()
}

// Health checks database health
func (r *MetadataRepository) Health() error {
	return r.db.Ping()
}

// Migrate runs database migrations
func (r *MetadataRepository) Migrate() error {
	return r.initSchema()
}

// SaveInstrument saves instrument information
func (r *MetadataRepository) SaveInstrument(ctx context.Context, instrument entities.InstrumentInfo) error {
	query := `INSERT OR REPLACE INTO instruments
		(symbol, type, market, base_asset, quote_asset,
		 min_price, max_price, min_quantity, max_quantity,
		 price_precision, quantity_precision, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		instrument.Symbol,
		string(instrument.Type),
		string(instrument.Market),
		instrument.BaseAsset,
		instrument.QuoteAsset,
		instrument.MinPrice,
		instrument.MaxPrice,
		instrument.MinQuantity,
		instrument.MaxQuantity,
		instrument.PricePrecision,
		instrument.QuantityPrecision,
		instrument.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to save instrument: %w", err)
	}

	return nil
}

// GetInstrument retrieves instrument by symbol
func (r *MetadataRepository) GetInstrument(ctx context.Context, symbol string) (*entities.InstrumentInfo, error) {
	query := `SELECT symbol, type, market, base_asset, quote_asset,
		min_price, max_price, min_quantity, max_quantity,
		price_precision, quantity_precision, is_active,
		created_at, updated_at
		FROM instruments WHERE symbol = ?`

	var instrument entities.InstrumentInfo
	var instrumentType, market string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, symbol).Scan(
		&instrument.Symbol,
		&instrumentType,
		&market,
		&instrument.BaseAsset,
		&instrument.QuoteAsset,
		&instrument.MinPrice,
		&instrument.MaxPrice,
		&instrument.MinQuantity,
		&instrument.MaxQuantity,
		&instrument.PricePrecision,
		&instrument.QuantityPrecision,
		&instrument.IsActive,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("instrument not found: %s", symbol)
		}
		return nil, fmt.Errorf("failed to get instrument: %w", err)
	}

	instrument.Type = entities.InstrumentType(instrumentType)
	instrument.Market = entities.MarketType(market)

	return &instrument, nil
}

// ListInstruments retrieves all instruments
func (r *MetadataRepository) ListInstruments(ctx context.Context) ([]entities.InstrumentInfo, error) {
	query := `SELECT symbol, type, market, base_asset, quote_asset,
		min_price, max_price, min_quantity, max_quantity,
		price_precision, quantity_precision, is_active,
		created_at, updated_at
		FROM instruments ORDER BY symbol`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list instruments: %w", err)
	}
	defer rows.Close()

	var instruments []entities.InstrumentInfo
	for rows.Next() {
		var instrument entities.InstrumentInfo
		var instrumentType, market string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&instrument.Symbol,
			&instrumentType,
			&market,
			&instrument.BaseAsset,
			&instrument.QuoteAsset,
			&instrument.MinPrice,
			&instrument.MaxPrice,
			&instrument.MinQuantity,
			&instrument.MaxQuantity,
			&instrument.PricePrecision,
			&instrument.QuantityPrecision,
			&instrument.IsActive,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan instrument: %w", err)
		}

		instrument.Type = entities.InstrumentType(instrumentType)
		instrument.Market = entities.MarketType(market)

		instruments = append(instruments, instrument)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate instruments: %w", err)
	}

	return instruments, nil
}

// DeleteInstrument removes instrument by symbol
func (r *MetadataRepository) DeleteInstrument(ctx context.Context, symbol string) error {
	query := `DELETE FROM instruments WHERE symbol = ?`

	result, err := r.db.ExecContext(ctx, query, symbol)
	if err != nil {
		return fmt.Errorf("failed to delete instrument: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("instrument not found: %s", symbol)
	}

	return nil
}

// SaveSubscription saves subscription information
func (r *MetadataRepository) SaveSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	dataTypesJSON, err := json.Marshal(subscription.DataTypes)
	if err != nil {
		return fmt.Errorf("failed to marshal data types: %w", err)
	}

	settingsJSON, err := json.Marshal(subscription.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `INSERT OR REPLACE INTO subscriptions
		(id, symbol, type, market, data_types, start_date, end_date, settings, broker_id, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		subscription.ID,
		subscription.Symbol,
		string(subscription.Type),
		string(subscription.Market),
		string(dataTypesJSON),
		subscription.StartDate,
		nil, // end_date - not in entity
		string(settingsJSON),
		subscription.BrokerID,
		subscription.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to save subscription: %w", err)
	}

	return nil
}

// GetSubscription retrieves subscription by ID
func (r *MetadataRepository) GetSubscription(ctx context.Context, id string) (*entities.InstrumentSubscription, error) {
	query := `SELECT id, symbol, type, market, data_types, start_date, end_date, settings, broker_id, is_active,
		created_at, updated_at
		FROM subscriptions WHERE id = ?`

	var subscription entities.InstrumentSubscription
	var dataTypesJSON, settingsJSON, subscriptionType, market string
	var endDate sql.NullTime
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&subscription.ID,
		&subscription.Symbol,
		&subscriptionType,
		&market,
		&dataTypesJSON,
		&subscription.StartDate,
		&endDate,
		&settingsJSON,
		&subscription.BrokerID,
		&subscription.IsActive,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subscription not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	// Set type and market
	subscription.Type = entities.InstrumentType(subscriptionType)
	subscription.Market = entities.MarketType(market)

	// Unmarshal JSON fields
	if err := json.Unmarshal([]byte(dataTypesJSON), &subscription.DataTypes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data types: %w", err)
	}

	if err := json.Unmarshal([]byte(settingsJSON), &subscription.Settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	subscription.CreatedAt = createdAt
	subscription.UpdatedAt = updatedAt

	return &subscription, nil
}

// ListSubscriptions retrieves all subscriptions
func (r *MetadataRepository) ListSubscriptions(ctx context.Context) ([]entities.InstrumentSubscription, error) {
	query := `SELECT id, symbol, type, market, data_types, start_date, end_date, settings, broker_id, is_active,
		created_at, updated_at
		FROM subscriptions ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subscriptions []entities.InstrumentSubscription
	for rows.Next() {
		var subscription entities.InstrumentSubscription
		var dataTypesJSON, settingsJSON, subscriptionType, market string
		var endDate sql.NullTime
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&subscription.ID,
			&subscription.Symbol,
			&subscriptionType,
			&market,
			&dataTypesJSON,
			&subscription.StartDate,
			&endDate,
			&settingsJSON,
			&subscription.BrokerID,
			&subscription.IsActive,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}

		// Set type and market
		subscription.Type = entities.InstrumentType(subscriptionType)
		subscription.Market = entities.MarketType(market)

		// Unmarshal JSON fields
		if err := json.Unmarshal([]byte(dataTypesJSON), &subscription.DataTypes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal data types: %w", err)
		}

		if err := json.Unmarshal([]byte(settingsJSON), &subscription.Settings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal settings: %w", err)
		}

		subscription.CreatedAt = createdAt
		subscription.UpdatedAt = updatedAt

		subscriptions = append(subscriptions, subscription)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate subscriptions: %w", err)
	}

	return subscriptions, nil
}

// UpdateSubscription updates subscription information
func (r *MetadataRepository) UpdateSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	return r.SaveSubscription(ctx, subscription) // Use INSERT OR REPLACE
}

// DeleteSubscription removes subscription by ID
func (r *MetadataRepository) DeleteSubscription(ctx context.Context, id string) error {
	query := `DELETE FROM subscriptions WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found: %s", id)
	}

	return nil
}

// SaveBrokerConfig saves broker configuration
func (r *MetadataRepository) SaveBrokerConfig(ctx context.Context, config interfaces.BrokerConfig) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal broker config: %w", err)
	}

	query := `INSERT OR REPLACE INTO broker_configs
		(id, name, type, enabled, config_json)
		VALUES (?, ?, ?, ?, ?)`

	_, err = r.db.ExecContext(ctx, query,
		config.ID,
		config.Name,
		string(config.Type),
		config.Enabled,
		string(configJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to save broker config: %w", err)
	}

	return nil
}

// GetBrokerConfig retrieves broker configuration by ID
func (r *MetadataRepository) GetBrokerConfig(ctx context.Context, brokerID string) (*interfaces.BrokerConfig, error) {
	query := `SELECT id, name, type, enabled, config_json, created_at, updated_at
		FROM broker_configs WHERE id = ?`

	var config interfaces.BrokerConfig
	var configJSON string
	var brokerType string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, brokerID).Scan(
		&config.ID,
		&config.Name,
		&brokerType,
		&config.Enabled,
		&configJSON,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("broker config not found: %s", brokerID)
		}
		return nil, fmt.Errorf("failed to get broker config: %w", err)
	}

	// Unmarshal the full config
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal broker config: %w", err)
	}

	config.Type = interfaces.BrokerType(brokerType)

	return &config, nil
}

// ListBrokerConfigs retrieves all broker configurations
func (r *MetadataRepository) ListBrokerConfigs(ctx context.Context) ([]interfaces.BrokerConfig, error) {
	query := `SELECT id, name, type, enabled, config_json, created_at, updated_at
		FROM broker_configs ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list broker configs: %w", err)
	}
	defer rows.Close()

	var configs []interfaces.BrokerConfig
	for rows.Next() {
		var config interfaces.BrokerConfig
		var configJSON string
		var brokerType string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&config.ID,
			&config.Name,
			&brokerType,
			&config.Enabled,
			&configJSON,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan broker config: %w", err)
		}

		// Unmarshal the full config
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal broker config: %w", err)
		}

		config.Type = interfaces.BrokerType(brokerType)
		configs = append(configs, config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate broker configs: %w", err)
	}

	return configs, nil
}

// DeleteBrokerConfig removes broker configuration by ID
func (r *MetadataRepository) DeleteBrokerConfig(ctx context.Context, brokerID string) error {
	query := `DELETE FROM broker_configs WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, brokerID)
	if err != nil {
		return fmt.Errorf("failed to delete broker config: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("broker config not found: %s", brokerID)
	}

	return nil
}
