package dto

import (
	"time"
)

// Response represents a generic API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents an API error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// BrokerResponse represents broker information
type BrokerResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Connected bool      `json:"connected"`
	LastSeen  time.Time `json:"last_seen"`
}

// MarketDataResponse represents market data response
type MarketDataResponse struct {
	Symbol    string      `json:"symbol"`
	BrokerID  string      `json:"broker_id"`
	DataType  string      `json:"data_type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// TickerResponse represents ticker data response
type TickerResponse struct {
	Symbol    string    `json:"symbol"`
	Price     float64   `json:"price"`
	BidPrice  float64   `json:"bid_price"`
	AskPrice  float64   `json:"ask_price"`
	Volume    float64   `json:"volume"`
	Change    float64   `json:"change"`
	ChangeP   float64   `json:"change_percent"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Open      float64   `json:"open"`
	Close     float64   `json:"close"`
	Timestamp time.Time `json:"timestamp"`
}

// CandleResponse represents candle data response
type CandleResponse struct {
	Symbol    string    `json:"symbol"`
	Timeframe string    `json:"timeframe"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

// OrderBookResponse represents order book data response
type OrderBookResponse struct {
	Symbol    string           `json:"symbol"`
	Bids      []OrderBookLevel `json:"bids"`
	Asks      []OrderBookLevel `json:"asks"`
	Timestamp time.Time        `json:"timestamp"`
}

// OrderBookLevel represents a single level in order book
type OrderBookLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// TickerListResponse represents list of tickers response
type TickerListResponse struct {
	Tickers []TickerResponse `json:"tickers"`
	Total   int              `json:"total"`
	From    *time.Time       `json:"from,omitempty"`
	To      *time.Time       `json:"to,omitempty"`
}

// CandleListResponse represents list of candles response
type CandleListResponse struct {
	Candles []CandleResponse `json:"candles"`
	Total   int              `json:"total"`
	From    *time.Time       `json:"from,omitempty"`
	To      *time.Time       `json:"to,omitempty"`
}

// LoginResponse represents successful authentication
type LoginResponse struct {
	Token        string   `json:"token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
}
