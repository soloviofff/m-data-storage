package interfaces

import (
	"context"
	"time"

	"m-data-storage/internal/domain/entities"
)

// DataProcessor - interface for data processing
type DataProcessor interface {
	// Incoming data processing
	ProcessTicker(ctx context.Context, ticker entities.Ticker) error
	ProcessCandle(ctx context.Context, candle entities.Candle) error
	ProcessOrderBook(ctx context.Context, orderBook entities.OrderBook) error

	// Batch processing
	ProcessTickerBatch(ctx context.Context, tickers []entities.Ticker) error
	ProcessCandleBatch(ctx context.Context, candles []entities.Candle) error
	ProcessOrderBookBatch(ctx context.Context, orderBooks []entities.OrderBook) error

	// Lifecycle management
	Start(ctx context.Context) error
	Stop() error
	Health() error
}

// InstrumentManager - interface for instrument management
type InstrumentManager interface {
	// Subscription management
	AddSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error
	RemoveSubscription(ctx context.Context, subscriptionID string) error
	UpdateSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error
	GetSubscription(ctx context.Context, subscriptionID string) (*entities.InstrumentSubscription, error)
	ListSubscriptions(ctx context.Context) ([]entities.InstrumentSubscription, error)

	// Instrument management
	AddInstrument(ctx context.Context, instrument entities.InstrumentInfo) error
	GetInstrument(ctx context.Context, symbol string) (*entities.InstrumentInfo, error)
	ListInstruments(ctx context.Context) ([]entities.InstrumentInfo, error)

	// Broker synchronization
	SyncWithBrokers(ctx context.Context) error
	StartTracking(ctx context.Context, subscriptionID string) error
	StopTracking(ctx context.Context, subscriptionID string) error

	// Lifecycle management
	Start(ctx context.Context) error
	Stop() error
	Health() error
}

// DataValidator - interface for data validation
type DataValidator interface {
	ValidateTicker(ticker entities.Ticker) error
	ValidateCandle(candle entities.Candle) error
	ValidateOrderBook(orderBook entities.OrderBook) error
	ValidateTimeframe(timeframe string) error
	ValidateMarketType(marketType string) error
	ValidateInstrumentType(instrumentType string) error
	ValidateInstrument(instrument entities.InstrumentInfo) error
	ValidateSubscription(subscription entities.InstrumentSubscription) error
}

// StorageService - interface for high-level storage operations
type StorageService interface {
	// Individual record saving with batch processing
	SaveTicker(ctx context.Context, ticker entities.Ticker) error
	SaveCandle(ctx context.Context, candle entities.Candle) error
	SaveOrderBook(ctx context.Context, orderBook entities.OrderBook) error

	// Multiple record saving
	SaveTickers(ctx context.Context, tickers []entities.Ticker) error
	SaveCandles(ctx context.Context, candles []entities.Candle) error
	SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error

	// Data retrieval
	GetTickers(ctx context.Context, filter TickerFilter) ([]entities.Ticker, error)
	GetCandles(ctx context.Context, filter CandleFilter) ([]entities.Candle, error)
	GetOrderBooks(ctx context.Context, filter OrderBookFilter) ([]entities.OrderBook, error)

	// Buffer management
	FlushAll(ctx context.Context) error
	GetStats() StorageServiceStats

	// Lifecycle management
	Close(ctx context.Context) error
}

// StorageServiceStats - storage service statistics
type StorageServiceStats struct {
	TickersSaved    int64     `json:"tickers_saved"`
	CandlesSaved    int64     `json:"candles_saved"`
	OrderBooksSaved int64     `json:"order_books_saved"`
	BatchesFlashed  int64     `json:"batches_flashed"`
	ErrorsCount     int64     `json:"errors_count"`
	LastFlushTime   time.Time `json:"last_flush_time"`
}

// DataCollector - interface for data collection
type DataCollector interface {
	// Data collection startup
	StartCollection(ctx context.Context) error
	StopCollection() error

	// Subscription management
	Subscribe(ctx context.Context, brokerID string, subscription entities.InstrumentSubscription) error
	Unsubscribe(ctx context.Context, brokerID string, subscriptionID string) error

	// Data channel access
	GetTickerChannel() <-chan entities.Ticker
	GetCandleChannel() <-chan entities.Candle
	GetOrderBookChannel() <-chan entities.OrderBook

	// Statistics
	GetCollectionStats() CollectionStats
	Health() error
}

// BrokerStorageIntegration - interface for broker-storage integration
type BrokerStorageIntegration interface {
	// Lifecycle management
	Start(ctx context.Context) error
	Stop() error

	// Broker management
	AddBroker(brokerID string, broker Broker) error
	RemoveBroker(brokerID string) error

	// Statistics and monitoring
	GetStats() BrokerStorageIntegrationStats
	GetBrokerStats(brokerID string) (BrokerIntegrationStats, error)
	Health() error
}

// BrokerStorageIntegrationStats - broker-storage integration statistics
type BrokerStorageIntegrationStats struct {
	ActiveBrokers    int       `json:"active_brokers"`
	TotalTickers     int64     `json:"total_tickers"`
	TotalCandles     int64     `json:"total_candles"`
	TotalOrderBooks  int64     `json:"total_orderbooks"`
	TotalErrors      int64     `json:"total_errors"`
	LastDataReceived time.Time `json:"last_data_received"`
	StartedAt        time.Time `json:"started_at"`
}

// BrokerIntegrationStats - broker integration statistics
type BrokerIntegrationStats struct {
	BrokerID            string    `json:"broker_id"`
	TickersProcessed    int64     `json:"tickers_processed"`
	CandlesProcessed    int64     `json:"candles_processed"`
	OrderBooksProcessed int64     `json:"orderbooks_processed"`
	Errors              int64     `json:"errors"`
	LastDataReceived    time.Time `json:"last_data_received"`
	StartedAt           time.Time `json:"started_at"`
}

// DataPipeline - interface for data flow management
type DataPipeline interface {
	// Lifecycle management
	Start(ctx context.Context) error
	Stop() error

	// Broker management
	AddBroker(ctx context.Context, config BrokerConfig) error
	RemoveBroker(ctx context.Context, brokerID string) error

	// Subscription management
	Subscribe(ctx context.Context, brokerID string, subscriptions []entities.InstrumentSubscription) error
	Unsubscribe(ctx context.Context, brokerID string, subscriptions []entities.InstrumentSubscription) error

	// Statistics and monitoring
	GetStats() DataPipelineStats
	GetIntegrationStats() BrokerStorageIntegrationStats
	Health() error
}

// DataPipelineStats - data pipeline statistics
type DataPipelineStats struct {
	StartedAt          time.Time `json:"started_at"`
	ConnectedBrokers   int       `json:"connected_brokers"`
	TotalDataProcessed int64     `json:"total_data_processed"`
	TotalErrors        int64     `json:"total_errors"`
	LastHealthCheck    time.Time `json:"last_health_check"`
	HealthChecksPassed int64     `json:"health_checks_passed"`
	HealthChecksFailed int64     `json:"health_checks_failed"`
}

// DataQuery - interface for data queries
type DataQuery interface {
	// Data retrieval
	GetTickers(ctx context.Context, filter TickerFilter) ([]entities.Ticker, error)
	GetCandles(ctx context.Context, filter CandleFilter) ([]entities.Candle, error)
	GetOrderBooks(ctx context.Context, filter OrderBookFilter) ([]entities.OrderBook, error)

	// Aggregated data
	GetTickerAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]TickerAggregate, error)
	GetCandleAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]CandleAggregate, error)

	// Statistics
	GetDataStats(ctx context.Context) (DataStats, error)
}

// ConfigService - interface for configuration management
type ConfigService interface {
	// Broker configuration
	GetBrokerConfig(ctx context.Context, brokerID string) (*BrokerConfig, error)
	SetBrokerConfig(ctx context.Context, config BrokerConfig) error
	ListBrokerConfigs(ctx context.Context) ([]BrokerConfig, error)
	DeleteBrokerConfig(ctx context.Context, brokerID string) error

	// System configuration
	GetSystemConfig(ctx context.Context) (*SystemConfig, error)
	UpdateSystemConfig(ctx context.Context, config SystemConfig) error

	// Configuration validation
	ValidateBrokerConfig(config BrokerConfig) error
	ValidateSystemConfig(config SystemConfig) error
}

// SystemConfig - system configuration
type SystemConfig struct {
	// Storage settings
	StorageRetention time.Duration `json:"storage_retention"`
	VacuumInterval   time.Duration `json:"vacuum_interval"`
	MaxStorageSize   int64         `json:"max_storage_size"`

	// API settings
	APIPort         int           `json:"api_port"`
	APIHost         string        `json:"api_host"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

	// Security settings
	JWTSecret      string   `json:"jwt_secret"`
	APIKeyHeader   string   `json:"api_key_header"`
	AllowedOrigins []string `json:"allowed_origins"`

	// Monitoring settings
	MetricsEnabled bool `json:"metrics_enabled"`
	MetricsPort    int  `json:"metrics_port"`
	TracingEnabled bool `json:"tracing_enabled"`
}

// CollectionStats - data collection statistics
type CollectionStats struct {
	TotalTickers        int64     `json:"total_tickers"`
	TotalCandles        int64     `json:"total_candles"`
	TotalOrderBooks     int64     `json:"total_orderbooks"`
	TickersPerSecond    float64   `json:"tickers_per_second"`
	CandlesPerSecond    float64   `json:"candles_per_second"`
	OrderBooksPerSecond float64   `json:"orderbooks_per_second"`
	ActiveSubscriptions int       `json:"active_subscriptions"`
	ConnectedBrokers    int       `json:"connected_brokers"`
	LastUpdate          time.Time `json:"last_update"`
	Errors              int64     `json:"errors"`
}

// DataStats - data statistics
type DataStats struct {
	TotalRecords    int64            `json:"total_records"`
	RecordsByType   map[string]int64 `json:"records_by_type"`
	RecordsByBroker map[string]int64 `json:"records_by_broker"`
	RecordsBySymbol map[string]int64 `json:"records_by_symbol"`
	OldestRecord    time.Time        `json:"oldest_record"`
	NewestRecord    time.Time        `json:"newest_record"`
	StorageSize     int64            `json:"storage_size_bytes"`
}
