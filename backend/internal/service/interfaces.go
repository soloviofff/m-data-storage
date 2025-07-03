package service

import (
	"context"
	"time"

	"m-data-storage/api/dto"
	"m-data-storage/pkg/broker"
)

// BrokerService handles broker management and market data operations
type BrokerService interface {
	// Broker management
	AddBroker(ctx context.Context, config broker.BrokerConfig) error
	RemoveBroker(ctx context.Context, brokerID string) error
	GetBroker(ctx context.Context, brokerID string) (broker.Broker, error)
	ListBrokers(ctx context.Context) ([]broker.BrokerInfo, error)

	// Market data operations
	Subscribe(ctx context.Context, brokerID string, subscriptions []broker.InstrumentSubscription) error
	Unsubscribe(ctx context.Context, brokerID string, subscriptions []broker.InstrumentSubscription) error
	GetMarketData(ctx context.Context, brokerID, symbol string) ([]broker.Ticker, error)

	// Health checks
	HealthCheck(ctx context.Context) map[string]error
}

// StorageService handles data persistence operations
type StorageService interface {
	// Market data storage
	SaveTicker(ctx context.Context, ticker broker.Ticker) error
	SaveCandle(ctx context.Context, candle broker.Candle) error
	SaveOrderBook(ctx context.Context, orderBook broker.OrderBook) error

	// Data retrieval
	GetTickers(ctx context.Context, symbol string, from, to *time.Time) ([]broker.Ticker, error)
	GetCandles(ctx context.Context, symbol, timeframe string, from, to *time.Time) ([]broker.Candle, error)
	GetOrderBook(ctx context.Context, symbol string, depth int) (*broker.OrderBook, error)

	// Maintenance
	Cleanup(ctx context.Context, before time.Time) error
	Vacuum(ctx context.Context) error
}

// ConfigService handles configuration management
type ConfigService interface {
	// Broker configuration
	GetBrokerConfig(ctx context.Context, brokerID string) (*broker.BrokerConfig, error)
	SetBrokerConfig(ctx context.Context, config broker.BrokerConfig) error
	ListBrokerConfigs(ctx context.Context) ([]broker.BrokerConfig, error)

	// System configuration
	GetSystemConfig(ctx context.Context) (*dto.SystemConfig, error)
	UpdateSystemConfig(ctx context.Context, config dto.SystemConfig) error
}
