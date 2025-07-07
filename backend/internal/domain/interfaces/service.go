package interfaces

import (
	"context"
	"time"

	"m-data-storage/internal/domain/entities"
)

// DataProcessor - интерфейс для обработки данных
type DataProcessor interface {
	// Обработка входящих данных
	ProcessTicker(ctx context.Context, ticker entities.Ticker) error
	ProcessCandle(ctx context.Context, candle entities.Candle) error
	ProcessOrderBook(ctx context.Context, orderBook entities.OrderBook) error

	// Пакетная обработка
	ProcessTickerBatch(ctx context.Context, tickers []entities.Ticker) error
	ProcessCandleBatch(ctx context.Context, candles []entities.Candle) error
	ProcessOrderBookBatch(ctx context.Context, orderBooks []entities.OrderBook) error

	// Управление жизненным циклом
	Start(ctx context.Context) error
	Stop() error
	Health() error
}

// InstrumentManager - интерфейс для управления инструментами
type InstrumentManager interface {
	// Управление подписками
	AddSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error
	RemoveSubscription(ctx context.Context, subscriptionID string) error
	UpdateSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error
	GetSubscription(ctx context.Context, subscriptionID string) (*entities.InstrumentSubscription, error)
	ListSubscriptions(ctx context.Context) ([]entities.InstrumentSubscription, error)

	// Управление инструментами
	AddInstrument(ctx context.Context, instrument entities.InstrumentInfo) error
	GetInstrument(ctx context.Context, symbol string) (*entities.InstrumentInfo, error)
	ListInstruments(ctx context.Context) ([]entities.InstrumentInfo, error)

	// Синхронизация с брокерами
	SyncWithBrokers(ctx context.Context) error
	StartTracking(ctx context.Context, subscriptionID string) error
	StopTracking(ctx context.Context, subscriptionID string) error

	// Управление жизненным циклом
	Start(ctx context.Context) error
	Stop() error
	Health() error
}

// DataValidator - интерфейс для валидации данных
type DataValidator interface {
	ValidateTicker(ticker entities.Ticker) error
	ValidateCandle(candle entities.Candle) error
	ValidateOrderBook(orderBook entities.OrderBook) error
	ValidateTimeframe(timeframe string) error
	ValidateMarketType(marketType string) error
	ValidateInstrumentType(instrumentType string) error
}

// StorageService - интерфейс для высокоуровневых операций с хранилищем
type StorageService interface {
	// Сохранение отдельных записей с пакетной обработкой
	SaveTicker(ctx context.Context, ticker entities.Ticker) error
	SaveCandle(ctx context.Context, candle entities.Candle) error
	SaveOrderBook(ctx context.Context, orderBook entities.OrderBook) error

	// Сохранение множественных записей
	SaveTickers(ctx context.Context, tickers []entities.Ticker) error
	SaveCandles(ctx context.Context, candles []entities.Candle) error
	SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error

	// Получение данных
	GetTickers(ctx context.Context, filter TickerFilter) ([]entities.Ticker, error)
	GetCandles(ctx context.Context, filter CandleFilter) ([]entities.Candle, error)
	GetOrderBooks(ctx context.Context, filter OrderBookFilter) ([]entities.OrderBook, error)

	// Управление буферами
	FlushAll(ctx context.Context) error
	GetStats() StorageServiceStats

	// Управление жизненным циклом
	Close(ctx context.Context) error
}

// StorageServiceStats - статистика работы сервиса хранения
type StorageServiceStats struct {
	TickersSaved    int64     `json:"tickers_saved"`
	CandlesSaved    int64     `json:"candles_saved"`
	OrderBooksSaved int64     `json:"order_books_saved"`
	BatchesFlashed  int64     `json:"batches_flashed"`
	ErrorsCount     int64     `json:"errors_count"`
	LastFlushTime   time.Time `json:"last_flush_time"`
}

// DataCollector - интерфейс для сбора данных
type DataCollector interface {
	// Запуск сбора данных
	StartCollection(ctx context.Context) error
	StopCollection() error

	// Управление подписками
	Subscribe(ctx context.Context, brokerID string, subscription entities.InstrumentSubscription) error
	Unsubscribe(ctx context.Context, brokerID string, subscriptionID string) error

	// Получение каналов данных
	GetTickerChannel() <-chan entities.Ticker
	GetCandleChannel() <-chan entities.Candle
	GetOrderBookChannel() <-chan entities.OrderBook

	// Статистика
	GetCollectionStats() CollectionStats
	Health() error
}

// DataQuery - интерфейс для запросов данных
type DataQuery interface {
	// Получение данных
	GetTickers(ctx context.Context, filter TickerFilter) ([]entities.Ticker, error)
	GetCandles(ctx context.Context, filter CandleFilter) ([]entities.Candle, error)
	GetOrderBooks(ctx context.Context, filter OrderBookFilter) ([]entities.OrderBook, error)

	// Агрегированные данные
	GetTickerAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]TickerAggregate, error)
	GetCandleAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]CandleAggregate, error)

	// Статистика
	GetDataStats(ctx context.Context) (DataStats, error)
}

// ConfigService - интерфейс для управления конфигурацией
type ConfigService interface {
	// Конфигурация брокеров
	GetBrokerConfig(ctx context.Context, brokerID string) (*BrokerConfig, error)
	SetBrokerConfig(ctx context.Context, config BrokerConfig) error
	ListBrokerConfigs(ctx context.Context) ([]BrokerConfig, error)
	DeleteBrokerConfig(ctx context.Context, brokerID string) error

	// Системная конфигурация
	GetSystemConfig(ctx context.Context) (*SystemConfig, error)
	UpdateSystemConfig(ctx context.Context, config SystemConfig) error

	// Валидация конфигурации
	ValidateBrokerConfig(config BrokerConfig) error
	ValidateSystemConfig(config SystemConfig) error
}

// SystemConfig - системная конфигурация
type SystemConfig struct {
	// Настройки хранения
	StorageRetention time.Duration `json:"storage_retention"`
	VacuumInterval   time.Duration `json:"vacuum_interval"`
	MaxStorageSize   int64         `json:"max_storage_size"`

	// Настройки API
	APIPort         int           `json:"api_port"`
	APIHost         string        `json:"api_host"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

	// Настройки безопасности
	JWTSecret      string   `json:"jwt_secret"`
	APIKeyHeader   string   `json:"api_key_header"`
	AllowedOrigins []string `json:"allowed_origins"`

	// Настройки мониторинга
	MetricsEnabled bool `json:"metrics_enabled"`
	MetricsPort    int  `json:"metrics_port"`
	TracingEnabled bool `json:"tracing_enabled"`
}

// CollectionStats - статистика сбора данных
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

// DataStats - статистика данных
type DataStats struct {
	TotalRecords    int64            `json:"total_records"`
	RecordsByType   map[string]int64 `json:"records_by_type"`
	RecordsByBroker map[string]int64 `json:"records_by_broker"`
	RecordsBySymbol map[string]int64 `json:"records_by_symbol"`
	OldestRecord    time.Time        `json:"oldest_record"`
	NewestRecord    time.Time        `json:"newest_record"`
	StorageSize     int64            `json:"storage_size_bytes"`
}
