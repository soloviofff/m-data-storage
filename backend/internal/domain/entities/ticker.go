package entities

import (
	"fmt"
	"time"
)

// MarketType - market type
type MarketType string

const (
	MarketTypeSpot    MarketType = "spot"
	MarketTypeFutures MarketType = "futures"
	MarketTypeStock   MarketType = "stock"
)

// InstrumentType - instrument type
type InstrumentType string

const (
	InstrumentTypeSpot    InstrumentType = "spot"
	InstrumentTypeFutures InstrumentType = "futures"
	InstrumentTypeStock   InstrumentType = "stock"
	InstrumentTypeETF     InstrumentType = "etf"
	InstrumentTypeBond    InstrumentType = "bond"
)

// Ticker represents ticker data with support for different market types
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
	PrevClose24h  float64 `json:"prev_close_24h,omitempty"`

	// For futures
	OpenInterest float64 `json:"open_interest,omitempty"`

	// For stocks
	BidPrice float64 `json:"bid_price,omitempty"`
	AskPrice float64 `json:"ask_price,omitempty"`
	BidSize  float64 `json:"bid_size,omitempty"`
	AskSize  float64 `json:"ask_size,omitempty"`
}

// IsValid checks ticker validity
func (t *Ticker) IsValid() bool {
	if t.Symbol == "" {
		return false
	}
	if t.Price <= 0 {
		return false
	}
	if t.Volume < 0 {
		return false
	}
	if t.BrokerID == "" {
		return false
	}
	if t.Timestamp.IsZero() {
		return false
	}
	return true
}

// GetDisplayPrice returns price for display considering instrument type
func (t *Ticker) GetDisplayPrice() float64 {
	return t.Price
}

// GetSpread returns spread between bid and ask for stocks
func (t *Ticker) GetSpread() float64 {
	if t.Type == InstrumentTypeStock && t.BidPrice > 0 && t.AskPrice > 0 {
		return t.AskPrice - t.BidPrice
	}
	return 0
}

// GetMidPrice returns mid price between bid and ask
func (t *Ticker) GetMidPrice() float64 {
	if t.Type == InstrumentTypeStock && t.BidPrice > 0 && t.AskPrice > 0 {
		return (t.BidPrice + t.AskPrice) / 2
	}
	return t.Price
}

// Validate checks ticker validity
func (t *Ticker) Validate() error {
	if t.Symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	if t.Price <= 0 {
		return fmt.Errorf("price must be positive")
	}
	if t.Volume < 0 {
		return fmt.Errorf("volume cannot be negative")
	}
	if t.BrokerID == "" {
		return fmt.Errorf("broker ID cannot be empty")
	}
	if t.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}
	return nil
}

// GetPriceChange24h returns price change over 24 hours
func (t *Ticker) GetPriceChange24h() float64 {
	if t.PrevClose24h > 0 {
		return t.Price - t.PrevClose24h
	}
	return 0
}

// GetPriceChangePercent24h returns percentage price change over 24 hours
func (t *Ticker) GetPriceChangePercent24h() float64 {
	if t.PrevClose24h > 0 {
		return (t.Price - t.PrevClose24h) / t.PrevClose24h * 100
	}
	return 0
}

// IsValidForMarket checks instrument type and market compatibility
func (t *Ticker) IsValidForMarket() bool {
	switch t.Market {
	case MarketTypeSpot:
		return t.Type == InstrumentTypeSpot
	case MarketTypeFutures:
		return t.Type == InstrumentTypeFutures
	case MarketTypeStock:
		return t.Type == InstrumentTypeStock || t.Type == InstrumentTypeETF || t.Type == InstrumentTypeBond
	default:
		return false
	}
}
