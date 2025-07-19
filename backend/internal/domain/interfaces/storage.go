package interfaces

import (
	"context"
	"database/sql"
	"time"

	"m-data-storage/internal/domain/entities"
)

// TickerFilter - filter for ticker retrieval
type TickerFilter struct {
	Symbols   []string                  `json:"symbols,omitempty"`
	BrokerIDs []string                  `json:"broker_ids,omitempty"`
	Markets   []entities.MarketType     `json:"markets,omitempty"`
	Types     []entities.InstrumentType `json:"types,omitempty"`
	StartTime *time.Time                `json:"start_time,omitempty"`
	EndTime   *time.Time                `json:"end_time,omitempty"`
	Limit     int                       `json:"limit,omitempty"`
	Offset    int                       `json:"offset,omitempty"`
}

// CandleFilter - filter for candle retrieval
type CandleFilter struct {
	Symbols    []string                  `json:"symbols,omitempty"`
	BrokerIDs  []string                  `json:"broker_ids,omitempty"`
	Markets    []entities.MarketType     `json:"markets,omitempty"`
	Types      []entities.InstrumentType `json:"types,omitempty"`
	Timeframes []string                  `json:"timeframes,omitempty"`
	StartTime  *time.Time                `json:"start_time,omitempty"`
	EndTime    *time.Time                `json:"end_time,omitempty"`
	Limit      int                       `json:"limit,omitempty"`
	Offset     int                       `json:"offset,omitempty"`
}

// OrderBookFilter - filter for order book retrieval
type OrderBookFilter struct {
	Symbols   []string                  `json:"symbols,omitempty"`
	BrokerIDs []string                  `json:"broker_ids,omitempty"`
	Markets   []entities.MarketType     `json:"markets,omitempty"`
	Types     []entities.InstrumentType `json:"types,omitempty"`
	StartTime *time.Time                `json:"start_time,omitempty"`
	EndTime   *time.Time                `json:"end_time,omitempty"`
	Limit     int                       `json:"limit,omitempty"`
	Offset    int                       `json:"offset,omitempty"`
}

// SubscriptionFilter - filter for subscription retrieval
type SubscriptionFilter struct {
	Symbols   []string `json:"symbols,omitempty"`
	BrokerIDs []string `json:"broker_ids,omitempty"`
	Active    *bool    `json:"active,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
}

// Storage - main interface for data storage
type Storage interface {
	// Data saving
	SaveTickers(ctx context.Context, tickers []entities.Ticker) error
	SaveCandles(ctx context.Context, candles []entities.Candle) error
	SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error

	// Data retrieval
	GetTickers(ctx context.Context, filter TickerFilter) ([]entities.Ticker, error)
	GetCandles(ctx context.Context, filter CandleFilter) ([]entities.Candle, error)
	GetOrderBooks(ctx context.Context, filter OrderBookFilter) ([]entities.OrderBook, error)

	// Lifecycle management
	Connect(ctx context.Context) error
	Disconnect() error
	Health() error

	// Statistics
	GetStats(ctx context.Context) (StorageStats, error)
}

// MetadataStorage - interface for metadata storage (SQLite)
type MetadataStorage interface {
	// Instrument management
	SaveInstrument(ctx context.Context, instrument entities.InstrumentInfo) error
	GetInstrument(ctx context.Context, symbol string) (*entities.InstrumentInfo, error)
	ListInstruments(ctx context.Context) ([]entities.InstrumentInfo, error)
	DeleteInstrument(ctx context.Context, symbol string) error

	// Subscription management
	SaveSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error
	GetSubscription(ctx context.Context, id string) (*entities.InstrumentSubscription, error)
	GetSubscriptions(ctx context.Context, filter SubscriptionFilter) ([]entities.InstrumentSubscription, error)
	ListSubscriptions(ctx context.Context) ([]entities.InstrumentSubscription, error)
	UpdateSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error
	DeleteSubscription(ctx context.Context, id string) error

	// Broker configuration management
	SaveBrokerConfig(ctx context.Context, config BrokerConfig) error
	GetBrokerConfig(ctx context.Context, brokerID string) (*BrokerConfig, error)
	ListBrokerConfigs(ctx context.Context) ([]BrokerConfig, error)
	DeleteBrokerConfig(ctx context.Context, brokerID string) error

	// Lifecycle management
	Connect(ctx context.Context) error
	Disconnect() error
	Health() error
	Migrate() error

	// Database access for migrations
	GetDB() *sql.DB
}

// TimeSeriesStorage - interface for time series storage (QuestDB)
type TimeSeriesStorage interface {
	// Time series saving
	SaveTickers(ctx context.Context, tickers []entities.Ticker) error
	SaveCandles(ctx context.Context, candles []entities.Candle) error
	SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error

	// Time series retrieval
	GetTickers(ctx context.Context, filter TickerFilter) ([]entities.Ticker, error)
	GetCandles(ctx context.Context, filter CandleFilter) ([]entities.Candle, error)
	GetOrderBooks(ctx context.Context, filter OrderBookFilter) ([]entities.OrderBook, error)

	// Data aggregation
	GetTickerAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]TickerAggregate, error)
	GetCandleAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]CandleAggregate, error)

	// Lifecycle management
	Connect(ctx context.Context) error
	Disconnect() error
	Health() error

	// Database access for migrations
	GetDB() *sql.DB

	// Old data cleanup
	CleanupOldData(ctx context.Context, retentionPeriod time.Duration) error
}

// StorageManager - manager for storage management
type StorageManager interface {
	// Basic operations
	SaveTickers(ctx context.Context, tickers []entities.Ticker) error
	SaveCandles(ctx context.Context, candles []entities.Candle) error
	SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error

	GetTickers(ctx context.Context, filter TickerFilter) ([]entities.Ticker, error)
	GetCandles(ctx context.Context, filter CandleFilter) ([]entities.Candle, error)
	GetOrderBooks(ctx context.Context, filter OrderBookFilter) ([]entities.OrderBook, error)

	// Metadata management
	GetMetadataStorage() MetadataStorage
	GetTimeSeriesStorage() TimeSeriesStorage

	// Lifecycle management
	Initialize(ctx context.Context) error
	Shutdown() error
	Health() map[string]error
}

// StorageStats - storage statistics
type StorageStats struct {
	TotalTickers    int64     `json:"total_tickers"`
	TotalCandles    int64     `json:"total_candles"`
	TotalOrderBooks int64     `json:"total_orderbooks"`
	OldestRecord    time.Time `json:"oldest_record"`
	NewestRecord    time.Time `json:"newest_record"`
	StorageSize     int64     `json:"storage_size_bytes"`
}

// TickerAggregate - aggregated ticker data
type TickerAggregate struct {
	Symbol    string    `json:"symbol"`
	Timestamp time.Time `json:"timestamp"`
	AvgPrice  float64   `json:"avg_price"`
	MinPrice  float64   `json:"min_price"`
	MaxPrice  float64   `json:"max_price"`
	Volume    float64   `json:"volume"`
	Count     int64     `json:"count"`
}

// CandleAggregate - aggregated candle data
type CandleAggregate struct {
	Symbol    string    `json:"symbol"`
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	Count     int64     `json:"count"`
}

// DateFilter - interface for date-based filtering operations
type DateFilter interface {
	FilterTickersBySubscriptionDate(ctx context.Context, filter TickerFilter) ([]entities.Ticker, error)
	FilterCandlesBySubscriptionDate(ctx context.Context, filter CandleFilter) ([]entities.Candle, error)
	FilterOrderBooksBySubscriptionDate(ctx context.Context, filter OrderBookFilter) ([]entities.OrderBook, error)
}

// UserStorage defines the interface for user storage operations
type UserStorage interface {
	// User operations
	CreateUser(ctx context.Context, user *entities.User) error
	GetUserByID(ctx context.Context, id string) (*entities.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entities.User, error)
	GetUserByEmail(ctx context.Context, email string) (*entities.User, error)
	GetUsers(ctx context.Context, filter entities.UserFilter) ([]*entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
	DeleteUser(ctx context.Context, id string) error
	UpdateLastLogin(ctx context.Context, userID string) error

	// Session operations
	CreateSession(ctx context.Context, session *entities.UserSession) error
	GetSessionByID(ctx context.Context, id string) (*entities.UserSession, error)
	GetSessionsByUserID(ctx context.Context, userID string) ([]*entities.UserSession, error)
	UpdateSession(ctx context.Context, session *entities.UserSession) error
	RevokeSession(ctx context.Context, id string) error
	RevokeUserSessions(ctx context.Context, userID string) error
	CleanupExpiredSessions(ctx context.Context) error

	// API Key operations
	CreateAPIKey(ctx context.Context, apiKey *entities.APIKey) error
	GetAPIKeyByID(ctx context.Context, id string) (*entities.APIKey, error)
	GetAPIKeysByUserID(ctx context.Context, userID string) ([]*entities.APIKey, error)
	GetAPIKeys(ctx context.Context, filter entities.APIKeyFilter) ([]*entities.APIKey, error)
	UpdateAPIKey(ctx context.Context, apiKey *entities.APIKey) error
	DeleteAPIKey(ctx context.Context, id string) error
	UpdateAPIKeyLastUsed(ctx context.Context, id string) error

	// Health check
	Health(ctx context.Context) error
}

// RoleStorage defines the interface for role and permission storage operations
type RoleStorage interface {
	// Role operations
	CreateRole(ctx context.Context, role *entities.Role) error
	GetRoleByID(ctx context.Context, id string) (*entities.Role, error)
	GetRoleByName(ctx context.Context, name string) (*entities.Role, error)
	GetRoles(ctx context.Context, filter entities.RoleFilter) ([]*entities.Role, error)
	UpdateRole(ctx context.Context, role *entities.Role) error
	DeleteRole(ctx context.Context, id string) error

	// Permission operations
	CreatePermission(ctx context.Context, permission *entities.Permission) error
	GetPermissionByID(ctx context.Context, id string) (*entities.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*entities.Permission, error)
	GetPermissions(ctx context.Context, filter entities.PermissionFilter) ([]*entities.Permission, error)
	UpdatePermission(ctx context.Context, permission *entities.Permission) error
	DeletePermission(ctx context.Context, id string) error

	// Role-Permission operations
	AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]*entities.Permission, error)
	GetPermissionRoles(ctx context.Context, permissionID string) ([]*entities.Role, error)

	// Health check
	Health(ctx context.Context) error
}
