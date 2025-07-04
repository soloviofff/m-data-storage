package entities

import (
	"time"
)

// DataType - тип данных для подписки
type DataType string

const (
	DataTypeTicker    DataType = "ticker"
	DataTypeCandle    DataType = "candle"
	DataTypeOrderBook DataType = "orderbook"
)

// InstrumentInfo представляет информацию об инструменте
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

// InstrumentSubscription представляет подписку на инструмент
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

// IsValid проверяет валидность информации об инструменте
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

// IsValid проверяет валидность подписки на инструмент
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

	// Проверяем валидность типов данных
	for _, dataType := range is.DataTypes {
		if dataType != DataTypeTicker && dataType != DataTypeCandle && dataType != DataTypeOrderBook {
			return false
		}
	}

	return true
}

// HasDataType проверяет, включен ли указанный тип данных в подписку
func (is *InstrumentSubscription) HasDataType(dataType DataType) bool {
	for _, dt := range is.DataTypes {
		if dt == dataType {
			return true
		}
	}
	return false
}

// GetSetting возвращает значение настройки по ключу
func (is *InstrumentSubscription) GetSetting(key string) (interface{}, bool) {
	if is.Settings == nil {
		return nil, false
	}
	value, exists := is.Settings[key]
	return value, exists
}

// SetSetting устанавливает значение настройки
func (is *InstrumentSubscription) SetSetting(key string, value interface{}) {
	if is.Settings == nil {
		is.Settings = make(map[string]interface{})
	}
	is.Settings[key] = value
}

// GetStringSetting возвращает строковое значение настройки
func (is *InstrumentSubscription) GetStringSetting(key string) (string, bool) {
	value, exists := is.GetSetting(key)
	if !exists {
		return "", false
	}
	str, ok := value.(string)
	return str, ok
}

// GetIntSetting возвращает целочисленное значение настройки
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

// GetBoolSetting возвращает булево значение настройки
func (is *InstrumentSubscription) GetBoolSetting(key string) (bool, bool) {
	value, exists := is.GetSetting(key)
	if !exists {
		return false, false
	}
	b, ok := value.(bool)
	return b, ok
}
