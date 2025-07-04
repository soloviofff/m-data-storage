package entities

import (
	"fmt"
	"time"
)

// Candle представляет свечу с поддержкой разных типов рынков
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

	// Дополнительные поля
	Trades      int64   `json:"trades,omitempty"`
	QuoteVolume float64 `json:"quote_volume,omitempty"`

	// Для фьючерсов
	OpenInterest float64 `json:"open_interest,omitempty"`
}

// IsValid проверяет валидность свечи
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

// GetBodySize возвращает размер тела свечи
func (c *Candle) GetBodySize() float64 {
	if c.Close > c.Open {
		return c.Close - c.Open
	}
	return c.Open - c.Close
}

// GetUpperShadow возвращает размер верхней тени
func (c *Candle) GetUpperShadow() float64 {
	if c.Close > c.Open {
		return c.High - c.Close
	}
	return c.High - c.Open
}

// GetLowerShadow возвращает размер нижней тени
func (c *Candle) GetLowerShadow() float64 {
	if c.Close > c.Open {
		return c.Open - c.Low
	}
	return c.Close - c.Low
}

// IsBullish проверяет, является ли свеча бычьей
func (c *Candle) IsBullish() bool {
	return c.Close > c.Open
}

// IsBearish проверяет, является ли свеча медвежьей
func (c *Candle) IsBearish() bool {
	return c.Close < c.Open
}

// IsDoji проверяет, является ли свеча доджи
func (c *Candle) IsDoji() bool {
	return c.Close == c.Open
}

// GetRange возвращает диапазон свечи
func (c *Candle) GetRange() float64 {
	return c.High - c.Low
}

// GetTypicalPrice возвращает типичную цену (HLC/3)
func (c *Candle) GetTypicalPrice() float64 {
	return (c.High + c.Low + c.Close) / 3
}

// GetWeightedPrice возвращает взвешенную цену (HLCC/4)
func (c *Candle) GetWeightedPrice() float64 {
	return (c.High + c.Low + c.Close + c.Close) / 4
}

// Validate проверяет валидность свечи
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
