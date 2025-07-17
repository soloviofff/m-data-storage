package entities

import (
	"fmt"
	"time"
)

// Candle represents a candle with support for different market types
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

	// For futures
	OpenInterest float64 `json:"open_interest,omitempty"`
}

// IsValid checks candle validity
func (c *Candle) IsValid() bool {
	if c.Symbol == "" {
		return false
	}
	if c.Open <= 0 || c.High <= 0 || c.Low <= 0 || c.Close <= 0 {
		return false
	}
	if c.High < c.Low {
		return false
	}
	if c.High < c.Open || c.High < c.Close {
		return false
	}
	if c.Low > c.Open || c.Low > c.Close {
		return false
	}
	if c.Volume < 0 {
		return false
	}
	if c.BrokerID == "" {
		return false
	}
	if c.Timestamp.IsZero() {
		return false
	}
	if c.Timeframe == "" {
		return false
	}
	return true
}

// GetBodySize returns candle body size
func (c *Candle) GetBodySize() float64 {
	if c.Close > c.Open {
		return c.Close - c.Open
	}
	return c.Open - c.Close
}

// GetUpperShadow returns upper shadow size
func (c *Candle) GetUpperShadow() float64 {
	if c.Close > c.Open {
		return c.High - c.Close
	}
	return c.High - c.Open
}

// GetLowerShadow returns lower shadow size
func (c *Candle) GetLowerShadow() float64 {
	if c.Close > c.Open {
		return c.Open - c.Low
	}
	return c.Close - c.Low
}

// IsBullish checks if candle is bullish
func (c *Candle) IsBullish() bool {
	return c.Close > c.Open
}

// IsBearish checks if candle is bearish
func (c *Candle) IsBearish() bool {
	return c.Close < c.Open
}

// IsDoji checks if candle is a doji
func (c *Candle) IsDoji() bool {
	return c.Close == c.Open
}

// GetRange returns candle range
func (c *Candle) GetRange() float64 {
	return c.High - c.Low
}

// GetTypicalPrice returns typical price (HLC/3)
func (c *Candle) GetTypicalPrice() float64 {
	return (c.High + c.Low + c.Close) / 3
}

// GetWeightedPrice returns weighted price (HLCC/4)
func (c *Candle) GetWeightedPrice() float64 {
	return (c.High + c.Low + c.Close + c.Close) / 4
}

// Validate checks candle validity
func (c *Candle) Validate() error {
	if c.Symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}
	if c.Open <= 0 || c.High <= 0 || c.Low <= 0 || c.Close <= 0 {
		return fmt.Errorf("OHLC prices must be positive")
	}
	if c.High < c.Low {
		return fmt.Errorf("high price cannot be less than low price")
	}
	if c.High < c.Open || c.High < c.Close {
		return fmt.Errorf("high price must be greater than or equal to open and close prices")
	}
	if c.Low > c.Open || c.Low > c.Close {
		return fmt.Errorf("low price must be less than or equal to open and close prices")
	}
	if c.Volume < 0 {
		return fmt.Errorf("volume cannot be negative")
	}
	if c.BrokerID == "" {
		return fmt.Errorf("broker ID cannot be empty")
	}
	if c.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}
	if c.Timeframe == "" {
		return fmt.Errorf("timeframe cannot be empty")
	}
	if c.Trades < 0 {
		return fmt.Errorf("trades count cannot be negative")
	}
	return nil
}
