package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInstrumentInfo_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		instrument InstrumentInfo
		want       bool
	}{
		{
			name: "valid spot instrument",
			instrument: InstrumentInfo{
				Symbol:            "BTC/USDT",
				BaseAsset:         "BTC",
				QuoteAsset:        "USDT",
				Type:              InstrumentTypeSpot,
				Market:            MarketTypeSpot,
				IsActive:          true,
				MinPrice:          0.01,
				MaxPrice:          1000000.0,
				MinQuantity:       0.001,
				MaxQuantity:       1000.0,
				PricePrecision:    2,
				QuantityPrecision: 3,
			},
			want: true,
		},
		{
			name: "valid stock instrument",
			instrument: InstrumentInfo{
				Symbol:            "AAPL",
				Type:              InstrumentTypeStock,
				Market:            MarketTypeStock,
				IsActive:          true,
				MinPrice:          0.01,
				MaxPrice:          1000.0,
				MinQuantity:       1.0,
				MaxQuantity:       10000.0,
				PricePrecision:    2,
				QuantityPrecision: 0,
			},
			want: true,
		},
		{
			name: "empty symbol",
			instrument: InstrumentInfo{
				Symbol:     "",
				BaseAsset:  "BTC",
				QuoteAsset: "USDT",
				Type:       InstrumentTypeSpot,
				Market:     MarketTypeSpot,
			},
			want: false,
		},
		{
			name: "empty base asset for non-stock",
			instrument: InstrumentInfo{
				Symbol:     "BTC/USDT",
				BaseAsset:  "",
				QuoteAsset: "USDT",
				Type:       InstrumentTypeSpot,
				Market:     MarketTypeSpot,
			},
			want: false,
		},
		{
			name: "empty quote asset for non-stock",
			instrument: InstrumentInfo{
				Symbol:     "BTC/USDT",
				BaseAsset:  "BTC",
				QuoteAsset: "",
				Type:       InstrumentTypeSpot,
				Market:     MarketTypeSpot,
			},
			want: false,
		},
		{
			name: "negative min price",
			instrument: InstrumentInfo{
				Symbol:     "BTC/USDT",
				BaseAsset:  "BTC",
				QuoteAsset: "USDT",
				Type:       InstrumentTypeSpot,
				Market:     MarketTypeSpot,
				MinPrice:   -0.01,
			},
			want: false,
		},
		{
			name: "max price less than min price",
			instrument: InstrumentInfo{
				Symbol:     "BTC/USDT",
				BaseAsset:  "BTC",
				QuoteAsset: "USDT",
				Type:       InstrumentTypeSpot,
				Market:     MarketTypeSpot,
				MinPrice:   100.0,
				MaxPrice:   50.0,
			},
			want: false,
		},
		{
			name: "negative min quantity",
			instrument: InstrumentInfo{
				Symbol:      "BTC/USDT",
				BaseAsset:   "BTC",
				QuoteAsset:  "USDT",
				Type:        InstrumentTypeSpot,
				Market:      MarketTypeSpot,
				MinQuantity: -0.001,
			},
			want: false,
		},
		{
			name: "max quantity less than min quantity",
			instrument: InstrumentInfo{
				Symbol:      "BTC/USDT",
				BaseAsset:   "BTC",
				QuoteAsset:  "USDT",
				Type:        InstrumentTypeSpot,
				Market:      MarketTypeSpot,
				MinQuantity: 10.0,
				MaxQuantity: 5.0,
			},
			want: false,
		},
		{
			name: "negative price precision",
			instrument: InstrumentInfo{
				Symbol:         "BTC/USDT",
				BaseAsset:      "BTC",
				QuoteAsset:     "USDT",
				Type:           InstrumentTypeSpot,
				Market:         MarketTypeSpot,
				PricePrecision: -1,
			},
			want: false,
		},
		{
			name: "negative quantity precision",
			instrument: InstrumentInfo{
				Symbol:            "BTC/USDT",
				BaseAsset:         "BTC",
				QuoteAsset:        "USDT",
				Type:              InstrumentTypeSpot,
				Market:            MarketTypeSpot,
				QuantityPrecision: -1,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.instrument.IsValid())
		})
	}
}

func TestInstrumentSubscription_IsValid(t *testing.T) {
	tests := []struct {
		name         string
		subscription InstrumentSubscription
		want         bool
	}{
		{
			name: "valid subscription",
			subscription: InstrumentSubscription{
				ID:        "sub-1",
				Symbol:    "BTC/USDT",
				Type:      InstrumentTypeSpot,
				Market:    MarketTypeSpot,
				DataTypes: []DataType{DataTypeTicker, DataTypeCandle},
				StartDate: time.Now(),
				BrokerID:  "binance",
				IsActive:  true,
			},
			want: true,
		},
		{
			name: "empty symbol",
			subscription: InstrumentSubscription{
				ID:        "sub-1",
				Symbol:    "",
				Type:      InstrumentTypeSpot,
				Market:    MarketTypeSpot,
				DataTypes: []DataType{DataTypeTicker},
				StartDate: time.Now(),
				BrokerID:  "binance",
			},
			want: false,
		},
		{
			name: "empty broker ID",
			subscription: InstrumentSubscription{
				ID:        "sub-1",
				Symbol:    "BTC/USDT",
				Type:      InstrumentTypeSpot,
				Market:    MarketTypeSpot,
				DataTypes: []DataType{DataTypeTicker},
				StartDate: time.Now(),
				BrokerID:  "",
			},
			want: false,
		},
		{
			name: "empty data types",
			subscription: InstrumentSubscription{
				ID:        "sub-1",
				Symbol:    "BTC/USDT",
				Type:      InstrumentTypeSpot,
				Market:    MarketTypeSpot,
				DataTypes: []DataType{},
				StartDate: time.Now(),
				BrokerID:  "binance",
			},
			want: false,
		},
		{
			name: "zero start date",
			subscription: InstrumentSubscription{
				ID:        "sub-1",
				Symbol:    "BTC/USDT",
				Type:      InstrumentTypeSpot,
				Market:    MarketTypeSpot,
				DataTypes: []DataType{DataTypeTicker},
				StartDate: time.Time{},
				BrokerID:  "binance",
			},
			want: false,
		},
		{
			name: "invalid data type",
			subscription: InstrumentSubscription{
				ID:        "sub-1",
				Symbol:    "BTC/USDT",
				Type:      InstrumentTypeSpot,
				Market:    MarketTypeSpot,
				DataTypes: []DataType{"invalid"},
				StartDate: time.Now(),
				BrokerID:  "binance",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.subscription.IsValid())
		})
	}
}

func TestInstrumentSubscription_HasDataType(t *testing.T) {
	subscription := InstrumentSubscription{
		DataTypes: []DataType{DataTypeTicker, DataTypeCandle},
	}

	assert.True(t, subscription.HasDataType(DataTypeTicker))
	assert.True(t, subscription.HasDataType(DataTypeCandle))
	assert.False(t, subscription.HasDataType(DataTypeOrderBook))
}

func TestInstrumentSubscription_Settings(t *testing.T) {
	subscription := InstrumentSubscription{}

	// Test GetSetting with nil settings
	value, exists := subscription.GetSetting("key")
	assert.False(t, exists)
	assert.Nil(t, value)

	// Test SetSetting
	subscription.SetSetting("string_key", "value")
	subscription.SetSetting("int_key", 42)
	subscription.SetSetting("bool_key", true)
	subscription.SetSetting("float_key", 3.14)

	// Test GetSetting
	value, exists = subscription.GetSetting("string_key")
	assert.True(t, exists)
	assert.Equal(t, "value", value)

	// Test GetStringSetting
	strValue, ok := subscription.GetStringSetting("string_key")
	assert.True(t, ok)
	assert.Equal(t, "value", strValue)

	strValue, ok = subscription.GetStringSetting("int_key")
	assert.False(t, ok)
	assert.Equal(t, "", strValue)

	strValue, ok = subscription.GetStringSetting("nonexistent")
	assert.False(t, ok)
	assert.Equal(t, "", strValue)

	// Test GetIntSetting
	intValue, ok := subscription.GetIntSetting("int_key")
	assert.True(t, ok)
	assert.Equal(t, 42, intValue)

	intValue, ok = subscription.GetIntSetting("float_key")
	assert.True(t, ok)
	assert.Equal(t, 3, intValue)

	intValue, ok = subscription.GetIntSetting("string_key")
	assert.False(t, ok)
	assert.Equal(t, 0, intValue)

	intValue, ok = subscription.GetIntSetting("nonexistent")
	assert.False(t, ok)
	assert.Equal(t, 0, intValue)

	// Test GetBoolSetting
	boolValue, ok := subscription.GetBoolSetting("bool_key")
	assert.True(t, ok)
	assert.True(t, boolValue)

	boolValue, ok = subscription.GetBoolSetting("string_key")
	assert.False(t, ok)
	assert.False(t, boolValue)

	boolValue, ok = subscription.GetBoolSetting("nonexistent")
	assert.False(t, ok)
	assert.False(t, boolValue)
}
