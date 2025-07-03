package broker

import "time"

// BrokerType represents the type of broker
type BrokerType string

const (
	BrokerTypeCrypto BrokerType = "crypto"
	BrokerTypeStock  BrokerType = "stock"
)

// MarketType represents the type of market
type MarketType string

const (
	MarketTypeSpot    MarketType = "spot"
	MarketTypeFutures MarketType = "futures"
	MarketTypeStock   MarketType = "stock"
)

// InstrumentType represents the type of trading instrument
type InstrumentType string

const (
	InstrumentTypeSpot    InstrumentType = "spot"
	InstrumentTypeFutures InstrumentType = "futures"
	InstrumentTypeStock   InstrumentType = "stock"
	InstrumentTypeETF     InstrumentType = "etf"
	InstrumentTypeBond    InstrumentType = "bond"
)

// BrokerInfo contains metadata about a broker
type BrokerInfo struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Type        BrokerType `json:"type"`
	Description string     `json:"description"`
	Website     string     `json:"website"`
	Status      string     `json:"status"`
	Features    []string   `json:"features"`
}

// InstrumentInfo contains metadata about a trading instrument
type InstrumentInfo struct {
	Symbol            string         `json:"symbol"`
	BaseAsset         string         `json:"base_asset"`
	QuoteAsset        string         `json:"quote_asset"`
	Type              InstrumentType `json:"type"`
	Market            MarketType     `json:"market"`
	IsActive          bool           `json:"is_active"`
	MinPrice          float64        `json:"min_price"`
	MaxPrice          float64        `json:"max_price"`
	MinQuantity       float64        `json:"min_quantity"`
	MaxQuantity       float64        `json:"max_quantity"`
	PricePrecision    int            `json:"price_precision"`
	QuantityPrecision int            `json:"quantity_precision"`
}

// InstrumentSubscription represents a subscription to instrument data
type InstrumentSubscription struct {
	Symbol    string                 `json:"symbol"`
	Type      InstrumentType         `json:"type"`
	Market    MarketType             `json:"market"`
	DataTypes []string               `json:"data_types"`
	StartDate time.Time              `json:"start_date"`
	Settings  map[string]interface{} `json:"settings"`
}
