package interfaces

import (
	"context"
	"database/sql"
	"time"

	"m-data-storage/internal/domain/entities"
)

// TickerFilter - фильтр для получения тикеров
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

// CandleFilter - фильтр для получения свечей
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

// OrderBookFilter - фильтр для получения ордербуков
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

// Storage - основной интерфейс для хранилища данных
type Storage interface {
	// Сохранение данных
	SaveTickers(ctx context.Context, tickers []entities.Ticker) error
	SaveCandles(ctx context.Context, candles []entities.Candle) error
	SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error

	// Получение данных
	GetTickers(ctx context.Context, filter TickerFilter) ([]entities.Ticker, error)
	GetCandles(ctx context.Context, filter CandleFilter) ([]entities.Candle, error)
	GetOrderBooks(ctx context.Context, filter OrderBookFilter) ([]entities.OrderBook, error)

	// Управление жизненным циклом
	Connect(ctx context.Context) error
	Disconnect() error
	Health() error

	// Статистика
	GetStats(ctx context.Context) (StorageStats, error)
}

// MetadataStorage - интерфейс для хранения метаданных (SQLite)
type MetadataStorage interface {
	// Управление инструментами
	SaveInstrument(ctx context.Context, instrument entities.InstrumentInfo) error
	GetInstrument(ctx context.Context, symbol string) (*entities.InstrumentInfo, error)
	ListInstruments(ctx context.Context) ([]entities.InstrumentInfo, error)
	DeleteInstrument(ctx context.Context, symbol string) error

	// Управление подписками
	SaveSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error
	GetSubscription(ctx context.Context, id string) (*entities.InstrumentSubscription, error)
	ListSubscriptions(ctx context.Context) ([]entities.InstrumentSubscription, error)
	UpdateSubscription(ctx context.Context, subscription entities.InstrumentSubscription) error
	DeleteSubscription(ctx context.Context, id string) error

	// Управление конфигурацией брокеров
	SaveBrokerConfig(ctx context.Context, config BrokerConfig) error
	GetBrokerConfig(ctx context.Context, brokerID string) (*BrokerConfig, error)
	ListBrokerConfigs(ctx context.Context) ([]BrokerConfig, error)
	DeleteBrokerConfig(ctx context.Context, brokerID string) error

	// Управление жизненным циклом
	Connect(ctx context.Context) error
	Disconnect() error
	Health() error
	Migrate() error

	// Доступ к базе данных для миграций
	GetDB() *sql.DB
}

// TimeSeriesStorage - интерфейс для хранения временных рядов (QuestDB)
type TimeSeriesStorage interface {
	// Сохранение временных рядов
	SaveTickers(ctx context.Context, tickers []entities.Ticker) error
	SaveCandles(ctx context.Context, candles []entities.Candle) error
	SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error

	// Получение временных рядов
	GetTickers(ctx context.Context, filter TickerFilter) ([]entities.Ticker, error)
	GetCandles(ctx context.Context, filter CandleFilter) ([]entities.Candle, error)
	GetOrderBooks(ctx context.Context, filter OrderBookFilter) ([]entities.OrderBook, error)

	// Агрегация данных
	GetTickerAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]TickerAggregate, error)
	GetCandleAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]CandleAggregate, error)

	// Управление жизненным циклом
	Connect(ctx context.Context) error
	Disconnect() error
	Health() error

	// Доступ к базе данных для миграций
	GetDB() *sql.DB

	// Очистка старых данных
	CleanupOldData(ctx context.Context, retentionPeriod time.Duration) error
}

// StorageManager - менеджер для управления хранилищами
type StorageManager interface {
	// Основные операции
	SaveTickers(ctx context.Context, tickers []entities.Ticker) error
	SaveCandles(ctx context.Context, candles []entities.Candle) error
	SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error

	GetTickers(ctx context.Context, filter TickerFilter) ([]entities.Ticker, error)
	GetCandles(ctx context.Context, filter CandleFilter) ([]entities.Candle, error)
	GetOrderBooks(ctx context.Context, filter OrderBookFilter) ([]entities.OrderBook, error)

	// Управление метаданными
	GetMetadataStorage() MetadataStorage
	GetTimeSeriesStorage() TimeSeriesStorage

	// Управление жизненным циклом
	Initialize(ctx context.Context) error
	Shutdown() error
	Health() map[string]error
}

// StorageStats - статистика хранилища
type StorageStats struct {
	TotalTickers    int64     `json:"total_tickers"`
	TotalCandles    int64     `json:"total_candles"`
	TotalOrderBooks int64     `json:"total_orderbooks"`
	OldestRecord    time.Time `json:"oldest_record"`
	NewestRecord    time.Time `json:"newest_record"`
	StorageSize     int64     `json:"storage_size_bytes"`
}

// TickerAggregate - агрегированные данные тикеров
type TickerAggregate struct {
	Symbol    string    `json:"symbol"`
	Timestamp time.Time `json:"timestamp"`
	AvgPrice  float64   `json:"avg_price"`
	MinPrice  float64   `json:"min_price"`
	MaxPrice  float64   `json:"max_price"`
	Volume    float64   `json:"volume"`
	Count     int64     `json:"count"`
}

// CandleAggregate - агрегированные данные свечей
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
