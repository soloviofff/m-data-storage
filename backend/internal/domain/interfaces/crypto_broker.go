package interfaces

import (
	"context"
	"time"
)

// SpotMarket - spot market
type SpotMarket struct {
	Symbol      string  `json:"symbol"`
	BaseAsset   string  `json:"base_asset"`
	QuoteAsset  string  `json:"quote_asset"`
	Status      string  `json:"status"`
	MinPrice    float64 `json:"min_price"`
	MaxPrice    float64 `json:"max_price"`
	TickSize    float64 `json:"tick_size"`
	MinQuantity float64 `json:"min_quantity"`
	MaxQuantity float64 `json:"max_quantity"`
	StepSize    float64 `json:"step_size"`
	MinNotional float64 `json:"min_notional"`
}

// FuturesMarket - futures market
type FuturesMarket struct {
	Symbol                string    `json:"symbol"`
	BaseAsset             string    `json:"base_asset"`
	QuoteAsset            string    `json:"quote_asset"`
	ContractType          string    `json:"contract_type"` // PERPETUAL, CURRENT_QUARTER, etc.
	DeliveryDate          time.Time `json:"delivery_date"`
	OnboardDate           time.Time `json:"onboard_date"`
	Status                string    `json:"status"`
	MaintMarginPercent    float64   `json:"maint_margin_percent"`
	RequiredMarginPercent float64   `json:"required_margin_percent"`
	BaseAssetPrecision    int       `json:"base_asset_precision"`
	QuotePrecision        int       `json:"quote_precision"`
	UnderlyingType        string    `json:"underlying_type"`
	UnderlyingSubType     []string  `json:"underlying_sub_type"`
	SettlePlan            int       `json:"settle_plan"`
}

// ContractInfo - contract information
type ContractInfo struct {
	Symbol                string   `json:"symbol"`
	Status                string   `json:"status"`
	MaintMarginPercent    float64  `json:"maint_margin_percent"`
	RequiredMarginPercent float64  `json:"required_margin_percent"`
	BaseAssetPrecision    int      `json:"base_asset_precision"`
	QuotePrecision        int      `json:"quote_precision"`
	TriggerProtect        float64  `json:"trigger_protect"`
	UnderlyingType        string   `json:"underlying_type"`
	UnderlyingSubType     []string `json:"underlying_sub_type"`
	SettlePlan            int      `json:"settle_plan"`
}

// FundingRate - funding rate
type FundingRate struct {
	Symbol    string    `json:"symbol"`
	Rate      float64   `json:"rate"`
	NextTime  time.Time `json:"next_time"`
	Timestamp time.Time `json:"timestamp"`
	BrokerID  string    `json:"broker_id"`
}

// MarkPrice - mark price
type MarkPrice struct {
	Symbol     string    `json:"symbol"`
	Price      float64   `json:"price"`
	IndexPrice float64   `json:"index_price"`
	Timestamp  time.Time `json:"timestamp"`
	BrokerID   string    `json:"broker_id"`
}

// Liquidation - liquidation
type Liquidation struct {
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
	BrokerID  string    `json:"broker_id"`
}

// CryptoBroker - interface for crypto exchanges
type CryptoBroker interface {
	Broker

	// Specific methods for crypto exchanges
	GetSpotMarkets() ([]SpotMarket, error)
	GetFuturesMarkets() ([]FuturesMarket, error)

	// Subscription to different market types
	SubscribeSpot(ctx context.Context, symbols []string) error
	SubscribeFutures(ctx context.Context, symbols []string) error

	// Contract information retrieval
	GetContractInfo(symbol string) (*ContractInfo, error)

	// Specific channels for crypto exchanges
	GetFundingRateChannel() <-chan FundingRate
	GetMarkPriceChannel() <-chan MarkPrice
	GetLiquidationChannel() <-chan Liquidation
}

// CryptoBrokerFactory - factory for crypto exchanges
type CryptoBrokerFactory interface {
	Create(config BrokerConfig) (CryptoBroker, error)
	GetSupportedExchanges() []string
}
