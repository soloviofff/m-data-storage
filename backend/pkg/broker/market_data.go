package broker

import "time"

// Ticker represents real-time price and volume information
type Ticker struct {
	Symbol    string         `json:"symbol"`
	Price     float64        `json:"price"`
	Volume    float64        `json:"volume"`
	Market    MarketType     `json:"market"`
	Type      InstrumentType `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	BrokerID  string         `json:"broker_id"`

	// Additional fields for different market types
	Change        float64 `json:"change,omitempty"`
	ChangePercent float64 `json:"change_percent,omitempty"`
	High24h       float64 `json:"high_24h,omitempty"`
	Low24h        float64 `json:"low_24h,omitempty"`
	Volume24h     float64 `json:"volume_24h,omitempty"`

	// Futures specific
	OpenInterest float64 `json:"open_interest,omitempty"`

	// Stock specific
	BidPrice float64 `json:"bid_price,omitempty"`
	AskPrice float64 `json:"ask_price,omitempty"`
	BidSize  float64 `json:"bid_size,omitempty"`
	AskSize  float64 `json:"ask_size,omitempty"`
}

// Candle represents OHLCV data for a specific timeframe
type Candle struct {
	Symbol    string         `json:"symbol"`
	Open      float64        `json:"open"`
	High      float64        `json:"high"`
	Low       float64        `json:"low"`
	Close     float64        `json:"close"`
	Volume    float64        `json:"volume"`
	Market    MarketType     `json:"market"`
	Type      InstrumentType `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	Timeframe string         `json:"timeframe"`
	BrokerID  string         `json:"broker_id"`

	// Additional fields
	Trades      int64   `json:"trades,omitempty"`
	QuoteVolume float64 `json:"quote_volume,omitempty"`

	// Futures specific
	OpenInterest float64 `json:"open_interest,omitempty"`
}

// PriceLevel represents a single level in the order book
type PriceLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// OrderBook represents the current state of the order book
type OrderBook struct {
	Symbol    string         `json:"symbol"`
	Bids      []PriceLevel   `json:"bids"`
	Asks      []PriceLevel   `json:"asks"`
	Market    MarketType     `json:"market"`
	Type      InstrumentType `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	BrokerID  string         `json:"broker_id"`

	// Additional fields
	LastUpdateID int64     `json:"last_update_id,omitempty"`
	EventTime    time.Time `json:"event_time,omitempty"`
}

// FundingRate represents funding rate information for perpetual contracts
type FundingRate struct {
	Symbol    string    `json:"symbol"`
	Rate      float64   `json:"rate"`
	NextTime  time.Time `json:"next_time"`
	Timestamp time.Time `json:"timestamp"`
	BrokerID  string    `json:"broker_id"`
}

// MarkPrice represents mark price information for futures
type MarkPrice struct {
	Symbol     string    `json:"symbol"`
	Price      float64   `json:"price"`
	IndexPrice float64   `json:"index_price"`
	Timestamp  time.Time `json:"timestamp"`
	BrokerID   string    `json:"broker_id"`
}

// Liquidation represents liquidation information
type Liquidation struct {
	Symbol    string    `json:"symbol"`
	Side      string    `json:"side"`
	Price     float64   `json:"price"`
	Quantity  float64   `json:"quantity"`
	Timestamp time.Time `json:"timestamp"`
	BrokerID  string    `json:"broker_id"`
}
