package entities

import (
	"fmt"
	"time"
)

// MarketType - тип рынка
type MarketType string

const (
	MarketTypeSpot    MarketType = "spot"
	MarketTypeFutures MarketType = "futures"
	MarketTypeStock   MarketType = "stock"
)

// InstrumentType - тип инструмента
type InstrumentType string

const (
	InstrumentTypeSpot    InstrumentType = "spot"
	InstrumentTypeFutures InstrumentType = "futures"
	InstrumentTypeStock   InstrumentType = "stock"
	InstrumentTypeETF     InstrumentType = "etf"
	InstrumentTypeBond    InstrumentType = "bond"
)

// Ticker представляет данные тикера с поддержкой разных типов рынков
type Ticker struct {
	Symbol    string         `json:"symbol"`
	Price     float64        `json:"price"`
	Volume    float64        `json:"volume"`
	Market    MarketType     `json:"market"`
	Type      InstrumentType `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	BrokerID  string         `json:"broker_id"`

	// Дополнительные поля для разных типов рынков
	Change        float64 `json:"change,omitempty"`
	ChangePercent float64 `json:"change_percent,omitempty"`
	High24h       float64 `json:"high_24h,omitempty"`
	Low24h        float64 `json:"low_24h,omitempty"`
	Volume24h     float64 `json:"volume_24h,omitempty"`
	PrevClose24h  float64 `json:"prev_close_24h,omitempty"`

	// Для фьючерсов
	OpenInterest float64 `json:"open_interest,omitempty"`

	// Для акций
	BidPrice float64 `json:"bid_price,omitempty"`
	AskPrice float64 `json:"ask_price,omitempty"`
	BidSize  float64 `json:"bid_size,omitempty"`
	AskSize  float64 `json:"ask_size,omitempty"`
}

// IsValid проверяет валидность тикера
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

// GetDisplayPrice возвращает цену для отображения с учетом типа инструмента
func (t *Ticker) GetDisplayPrice() float64 {
	return t.Price
}

// GetSpread возвращает спред между bid и ask для акций
func (t *Ticker) GetSpread() float64 {
	if t.Type == InstrumentTypeStock && t.BidPrice > 0 && t.AskPrice > 0 {
		return t.AskPrice - t.BidPrice
	}
	return 0
}

// GetMidPrice возвращает среднюю цену между bid и ask
func (t *Ticker) GetMidPrice() float64 {
	if t.Type == InstrumentTypeStock && t.BidPrice > 0 && t.AskPrice > 0 {
		return (t.BidPrice + t.AskPrice) / 2
	}
	return t.Price
}

// Validate проверяет валидность тикера
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

// GetPriceChange24h возвращает изменение цены за 24 часа
func (t *Ticker) GetPriceChange24h() float64 {
	if t.PrevClose24h > 0 {
		return t.Price - t.PrevClose24h
	}
	return 0
}

// GetPriceChangePercent24h возвращает процентное изменение цены за 24 часа
func (t *Ticker) GetPriceChangePercent24h() float64 {
	if t.PrevClose24h > 0 {
		return (t.Price - t.PrevClose24h) / t.PrevClose24h * 100
	}
	return 0
}

// IsValidForMarket проверяет соответствие типа инструмента и рынка
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
