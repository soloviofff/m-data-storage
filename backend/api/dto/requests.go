package dto

import (
	"m-data-storage/internal/domain/entities"
	"m-data-storage/pkg/broker"
	"time"
)

// AddBrokerRequest represents request to add a new broker
type AddBrokerRequest struct {
	Config broker.BrokerConfig `json:"config" validate:"required"`
}

// SubscriptionRequest represents request to subscribe to market data
type SubscriptionRequest struct {
	BrokerID      string                          `json:"broker_id" validate:"required"`
	Subscriptions []broker.InstrumentSubscription `json:"subscriptions" validate:"required,min=1"`
}

// CreateInstrumentRequest represents request to create a new instrument
type CreateInstrumentRequest struct {
	Symbol            string                  `json:"symbol" validate:"required"`
	BaseAsset         string                  `json:"base_asset"`
	QuoteAsset        string                  `json:"quote_asset"`
	Type              entities.InstrumentType `json:"type" validate:"required"`
	Market            entities.MarketType     `json:"market" validate:"required"`
	IsActive          bool                    `json:"is_active"`
	MinPrice          float64                 `json:"min_price" validate:"min=0"`
	MaxPrice          float64                 `json:"max_price" validate:"min=0"`
	MinQuantity       float64                 `json:"min_quantity" validate:"min=0"`
	MaxQuantity       float64                 `json:"max_quantity" validate:"min=0"`
	PricePrecision    int                     `json:"price_precision" validate:"min=0"`
	QuantityPrecision int                     `json:"quantity_precision" validate:"min=0"`
}

// UpdateInstrumentRequest represents request to update an existing instrument
type UpdateInstrumentRequest struct {
	BaseAsset         *string                  `json:"base_asset,omitempty"`
	QuoteAsset        *string                  `json:"quote_asset,omitempty"`
	Type              *entities.InstrumentType `json:"type,omitempty"`
	Market            *entities.MarketType     `json:"market,omitempty"`
	IsActive          *bool                    `json:"is_active,omitempty"`
	MinPrice          *float64                 `json:"min_price,omitempty" validate:"omitempty,min=0"`
	MaxPrice          *float64                 `json:"max_price,omitempty" validate:"omitempty,min=0"`
	MinQuantity       *float64                 `json:"min_quantity,omitempty" validate:"omitempty,min=0"`
	MaxQuantity       *float64                 `json:"max_quantity,omitempty" validate:"omitempty,min=0"`
	PricePrecision    *int                     `json:"price_precision,omitempty" validate:"omitempty,min=0"`
	QuantityPrecision *int                     `json:"quantity_precision,omitempty" validate:"omitempty,min=0"`
}

// CreateSubscriptionRequest represents request to create a new subscription
type CreateSubscriptionRequest struct {
	Symbol    string                  `json:"symbol" validate:"required"`
	Type      entities.InstrumentType `json:"type" validate:"required"`
	Market    entities.MarketType     `json:"market" validate:"required"`
	DataTypes []entities.DataType     `json:"data_types" validate:"required,min=1"`
	StartDate time.Time               `json:"start_date" validate:"required"`
	Settings  map[string]interface{}  `json:"settings,omitempty"`
	BrokerID  string                  `json:"broker_id" validate:"required"`
}

// UpdateSubscriptionRequest represents request to update an existing subscription
type UpdateSubscriptionRequest struct {
	DataTypes *[]entities.DataType    `json:"data_types,omitempty" validate:"omitempty,min=1"`
	Settings  *map[string]interface{} `json:"settings,omitempty"`
	IsActive  *bool                   `json:"is_active,omitempty"`
}

// MarketDataRequest represents request to get market data
type MarketDataRequest struct {
	BrokerID string     `json:"broker_id" validate:"required"`
	Symbol   string     `json:"symbol" validate:"required"`
	From     *time.Time `json:"from,omitempty"`
	To       *time.Time `json:"to,omitempty"`
	Limit    int        `json:"limit,omitempty" validate:"omitempty,min=1,max=1000"`
}

// TickerDataRequest represents request to get ticker data
type TickerDataRequest struct {
	Symbol string     `json:"symbol" validate:"required"`
	From   *time.Time `json:"from,omitempty"`
	To     *time.Time `json:"to,omitempty"`
	Limit  int        `json:"limit,omitempty" validate:"omitempty,min=1,max=1000"`
}

// CandleDataRequest represents request to get candle data
type CandleDataRequest struct {
	Symbol    string     `json:"symbol" validate:"required"`
	Timeframe string     `json:"timeframe" validate:"required"`
	From      *time.Time `json:"from,omitempty"`
	To        *time.Time `json:"to,omitempty"`
	Limit     int        `json:"limit,omitempty" validate:"omitempty,min=1,max=1000"`
}

// OrderBookDataRequest represents request to get order book data
type OrderBookDataRequest struct {
	Symbol string `json:"symbol" validate:"required"`
	Depth  int    `json:"depth,omitempty" validate:"omitempty,min=1,max=100"`
}

// UpdateConfigRequest represents request to update system configuration
type UpdateConfigRequest struct {
	Config SystemConfig `json:"config" validate:"required"`
}

// LoginRequest represents authentication request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
