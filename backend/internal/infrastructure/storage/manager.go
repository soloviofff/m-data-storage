package storage

import (
	"context"
	"fmt"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
	"m-data-storage/internal/infrastructure/storage/questdb"
	"m-data-storage/internal/infrastructure/storage/sqlite"
)

// Manager implements StorageManager interface
type Manager struct {
	metadata   interfaces.MetadataStorage
	timeSeries interfaces.TimeSeriesStorage
}

// Config holds configuration for storage manager
type Config struct {
	SQLite  SQLiteConfig  `yaml:"sqlite"`
	QuestDB QuestDBConfig `yaml:"questdb"`
}

// SQLiteConfig holds SQLite configuration
type SQLiteConfig struct {
	DatabasePath string `yaml:"database_path"`
}

// QuestDBConfig holds QuestDB configuration
type QuestDBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	SSLMode  string `yaml:"ssl_mode"`
}

// NewManager creates a new storage manager
func NewManager(config Config) (*Manager, error) {
	// Create SQLite repository for metadata
	metadataRepo, err := sqlite.NewMetadataRepository(config.SQLite.DatabasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata repository: %w", err)
	}

	// Create QuestDB repository for time series
	questdbConfig := questdb.Config{
		Host:     config.QuestDB.Host,
		Port:     config.QuestDB.Port,
		Database: config.QuestDB.Database,
		Username: config.QuestDB.Username,
		Password: config.QuestDB.Password,
		SSLMode:  config.QuestDB.SSLMode,
	}
	timeSeriesRepo := questdb.NewTimeSeriesRepository(questdbConfig)

	return &Manager{
		metadata:   metadataRepo,
		timeSeries: timeSeriesRepo,
	}, nil
}

// Initialize initializes both storage systems
func (m *Manager) Initialize(ctx context.Context) error {
	// Initialize metadata storage (SQLite)
	if err := m.metadata.Connect(ctx); err != nil {
		return fmt.Errorf("failed to initialize metadata storage: %w", err)
	}

	// Initialize time series storage (QuestDB)
	if err := m.timeSeries.Connect(ctx); err != nil {
		return fmt.Errorf("failed to initialize time series storage: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down both storage systems
func (m *Manager) Shutdown() error {
	var errors []error

	// Shutdown time series storage
	if err := m.timeSeries.Disconnect(); err != nil {
		errors = append(errors, fmt.Errorf("failed to shutdown time series storage: %w", err))
	}

	// Shutdown metadata storage
	if err := m.metadata.Disconnect(); err != nil {
		errors = append(errors, fmt.Errorf("failed to shutdown metadata storage: %w", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}

// Health checks the health of both storage systems
func (m *Manager) Health() map[string]error {
	health := make(map[string]error)

	health["metadata"] = m.metadata.Health()
	health["timeseries"] = m.timeSeries.Health()

	return health
}

// GetMetadataStorage returns the metadata storage interface
func (m *Manager) GetMetadataStorage() interfaces.MetadataStorage {
	return m.metadata
}

// GetTimeSeriesStorage returns the time series storage interface
func (m *Manager) GetTimeSeriesStorage() interfaces.TimeSeriesStorage {
	return m.timeSeries
}

// SaveTickers saves ticker data to time series storage
func (m *Manager) SaveTickers(ctx context.Context, tickers []entities.Ticker) error {
	return m.timeSeries.SaveTickers(ctx, tickers)
}

// SaveCandles saves candle data to time series storage
func (m *Manager) SaveCandles(ctx context.Context, candles []entities.Candle) error {
	return m.timeSeries.SaveCandles(ctx, candles)
}

// SaveOrderBooks saves order book data to time series storage
func (m *Manager) SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error {
	return m.timeSeries.SaveOrderBooks(ctx, orderBooks)
}

// GetTickers retrieves ticker data from time series storage
func (m *Manager) GetTickers(ctx context.Context, filter interfaces.TickerFilter) ([]entities.Ticker, error) {
	return m.timeSeries.GetTickers(ctx, filter)
}

// GetCandles retrieves candle data from time series storage
func (m *Manager) GetCandles(ctx context.Context, filter interfaces.CandleFilter) ([]entities.Candle, error) {
	return m.timeSeries.GetCandles(ctx, filter)
}

// GetOrderBooks retrieves order book data from time series storage
func (m *Manager) GetOrderBooks(ctx context.Context, filter interfaces.OrderBookFilter) ([]entities.OrderBook, error) {
	return m.timeSeries.GetOrderBooks(ctx, filter)
}

// SaveInstrument saves instrument metadata
func (m *Manager) SaveInstrument(ctx context.Context, instrument entities.InstrumentInfo) error {
	return m.metadata.SaveInstrument(ctx, instrument)
}

// GetInstrument retrieves instrument metadata
func (m *Manager) GetInstrument(ctx context.Context, symbol string) (*entities.InstrumentInfo, error) {
	return m.metadata.GetInstrument(ctx, symbol)
}

// ListInstruments retrieves all instruments metadata
func (m *Manager) ListInstruments(ctx context.Context) ([]entities.InstrumentInfo, error) {
	return m.metadata.ListInstruments(ctx)
}

// DeleteInstrument removes instrument metadata
func (m *Manager) DeleteInstrument(ctx context.Context, symbol string) error {
	return m.metadata.DeleteInstrument(ctx, symbol)
}

// SaveSubscription saves subscription metadata
func (m *Manager) SaveSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	return m.metadata.SaveSubscription(ctx, subscription)
}

// GetSubscription retrieves subscription metadata
func (m *Manager) GetSubscription(ctx context.Context, id string) (*entities.InstrumentSubscription, error) {
	return m.metadata.GetSubscription(ctx, id)
}

// ListSubscriptions retrieves all subscriptions metadata
func (m *Manager) ListSubscriptions(ctx context.Context) ([]entities.InstrumentSubscription, error) {
	return m.metadata.ListSubscriptions(ctx)
}

// UpdateSubscription updates subscription metadata
func (m *Manager) UpdateSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error {
	return m.metadata.UpdateSubscription(ctx, subscription)
}

// DeleteSubscription removes subscription metadata
func (m *Manager) DeleteSubscription(ctx context.Context, id string) error {
	return m.metadata.DeleteSubscription(ctx, id)
}

// SaveBrokerConfig saves broker configuration
func (m *Manager) SaveBrokerConfig(ctx context.Context, config interfaces.BrokerConfig) error {
	return m.metadata.SaveBrokerConfig(ctx, config)
}

// GetBrokerConfig retrieves broker configuration
func (m *Manager) GetBrokerConfig(ctx context.Context, brokerID string) (*interfaces.BrokerConfig, error) {
	return m.metadata.GetBrokerConfig(ctx, brokerID)
}

// ListBrokerConfigs retrieves all broker configurations
func (m *Manager) ListBrokerConfigs(ctx context.Context) ([]interfaces.BrokerConfig, error) {
	return m.metadata.ListBrokerConfigs(ctx)
}

// DeleteBrokerConfig removes broker configuration
func (m *Manager) DeleteBrokerConfig(ctx context.Context, brokerID string) error {
	return m.metadata.DeleteBrokerConfig(ctx, brokerID)
}

// CleanupOldData removes old time series data
func (m *Manager) CleanupOldData(ctx context.Context, retentionPeriod interface{}) error {
	// This would need proper implementation based on retention policy
	return nil
}

// GetStats returns storage statistics
func (m *Manager) GetStats(ctx context.Context) (interfaces.StorageStats, error) {
	// This would need proper implementation to gather stats from both storages
	return interfaces.StorageStats{}, nil
}
