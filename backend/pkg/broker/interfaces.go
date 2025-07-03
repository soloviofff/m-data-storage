package broker

import (
	"context"
)

// Broker is the base interface for all broker types
type Broker interface {
	// Connection methods
	Connect(ctx context.Context) error
	Disconnect() error
	IsConnected() bool

	// Broker information
	GetInfo() BrokerInfo
	GetSupportedInstruments() []InstrumentInfo

	// Subscription management
	Subscribe(ctx context.Context, instruments []InstrumentSubscription) error
	Unsubscribe(ctx context.Context, instruments []InstrumentSubscription) error

	// Data channels
	GetTickerChannel() <-chan Ticker
	GetCandleChannel() <-chan Candle
	GetOrderBookChannel() <-chan OrderBook

	// Lifecycle management
	Start(ctx context.Context) error
	Stop() error

	// Health check
	Health() error
}

// CryptoBroker extends Broker interface with crypto-specific methods
type CryptoBroker interface {
	Broker

	// Market information
	GetSpotMarkets() ([]SpotMarket, error)
	GetFuturesMarkets() ([]FuturesMarket, error)

	// Market-specific subscriptions
	SubscribeSpot(ctx context.Context, symbols []string) error
	SubscribeFutures(ctx context.Context, symbols []string) error

	// Contract information
	GetContractInfo(symbol string) (*ContractInfo, error)

	// Crypto-specific data channels
	GetFundingRateChannel() <-chan FundingRate
	GetMarkPriceChannel() <-chan MarkPrice
	GetLiquidationChannel() <-chan Liquidation
}

// StockBroker extends Broker interface with stock-specific methods
type StockBroker interface {
	Broker

	// Market information
	GetStockMarkets() ([]StockMarket, error)
	GetSectors() ([]Sector, error)

	// Stock-specific subscriptions
	SubscribeStocks(ctx context.Context, symbols []string) error

	// Company information
	GetCompanyInfo(symbol string) (*CompanyInfo, error)

	// Stock-specific data channels
	GetDividendChannel() <-chan Dividend
	GetCorporateActionChannel() <-chan CorporateAction
	GetEarningsChannel() <-chan Earnings
}

// BrokerFactory creates broker instances
type BrokerFactory interface {
	CreateBroker(config BrokerConfig) (Broker, error)
	GetSupportedTypes() []BrokerType
}
