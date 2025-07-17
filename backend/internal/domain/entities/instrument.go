package entities

import (
	"time"
)

// DataType - data type for subscription
type DataType string

const (
	DataTypeTicker    DataType = "ticker"
	DataTypeCandle    DataType = "candle"
	DataTypeOrderBook DataType = "orderbook"
)

// InstrumentInfo represents instrument information
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

// InstrumentSubscription represents instrument subscription
type InstrumentSubscription struct {
	ID        string                 `json:"id"`
	Symbol    string                 `json:"symbol"`
	Type      InstrumentType         `json:"type"`
	Market    MarketType             `json:"market"`
	DataTypes []DataType             `json:"data_types"`
	StartDate time.Time              `json:"start_date"`
	Settings  map[string]interface{} `json:"settings"`
	BrokerID  string                 `json:"broker_id"`
	IsActive  bool                   `json:"is_active"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// IsValid checks instrument information validity
func (ii *InstrumentInfo) IsValid() bool {
	if ii.Symbol == "" {
		return false
	}
	if ii.BaseAsset == "" && ii.Type != InstrumentTypeStock {
		return false
	}
	if ii.QuoteAsset == "" && ii.Type != InstrumentTypeStock {
		return false
	}
	if ii.MinPrice < 0 {
		return false
	}
	if ii.MaxPrice > 0 && ii.MaxPrice < ii.MinPrice {
		return false
	}
	if ii.MinQuantity < 0 {
		return false
	}
	if ii.MaxQuantity > 0 && ii.MaxQuantity < ii.MinQuantity {
		return false
	}
	if ii.PricePrecision < 0 {
		return false
	}
	if ii.QuantityPrecision < 0 {
		return false
	}
	return true
}

// IsValid checks instrument subscription validity
func (is *InstrumentSubscription) IsValid() bool {
	if is.Symbol == "" {
		return false
	}
	if is.BrokerID == "" {
		return false
	}
	if len(is.DataTypes) == 0 {
		return false
	}
	if is.StartDate.IsZero() {
		return false
	}

	// Check data types validity
	for _, dataType := range is.DataTypes {
		if dataType != DataTypeTicker && dataType != DataTypeCandle && dataType != DataTypeOrderBook {
			return false
		}
	}

	return true
}

// HasDataType checks if specified data type is included in subscription
func (is *InstrumentSubscription) HasDataType(dataType DataType) bool {
	for _, dt := range is.DataTypes {
		if dt == dataType {
			return true
		}
	}
	return false
}

// GetSetting returns setting value by key
func (is *InstrumentSubscription) GetSetting(key string) (interface{}, bool) {
	if is.Settings == nil {
		return nil, false
	}
	value, exists := is.Settings[key]
	return value, exists
}

// SetSetting sets setting value
func (is *InstrumentSubscription) SetSetting(key string, value interface{}) {
	if is.Settings == nil {
		is.Settings = make(map[string]interface{})
	}
	is.Settings[key] = value
}

// GetStringSetting returns string setting value
func (is *InstrumentSubscription) GetStringSetting(key string) (string, bool) {
	value, exists := is.GetSetting(key)
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

// GetIntSetting returns integer setting value
func (is *InstrumentSubscription) GetIntSetting(key string) (int, bool) {
	value, exists := is.GetSetting(key)
	if !exists {
		return 0, false
	}

	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// GetBoolSetting returns boolean setting value
func (is *InstrumentSubscription) GetBoolSetting(key string) (bool, bool) {
	value, exists := is.GetSetting(key)
	if !exists {
		return false, false
	}
	b, ok := value.(bool)
	return b, ok
}
