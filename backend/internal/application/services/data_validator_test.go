package services

import (
	"testing"
	"time"

	"m-data-storage/internal/domain/entities"
)

func TestDataValidatorService_ValidateTicker(t *testing.T) {
	validator := NewDataValidatorService()

	tests := []struct {
		name    string
		ticker  entities.Ticker
		wantErr bool
	}{
		{
			name: "valid ticker",
			ticker: entities.Ticker{
				Symbol:    "BTC/USDT",
				Price:     50000.0,
				Volume:    1000.0,
				Market:    entities.MarketTypeSpot,
				Type:      entities.InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: false,
		},
		{
			name: "empty symbol",
			ticker: entities.Ticker{
				Symbol:    "",
				Price:     50000.0,
				Volume:    1000.0,
				Market:    entities.MarketTypeSpot,
				Type:      entities.InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "invalid price",
			ticker: entities.Ticker{
				Symbol:    "BTC/USDT",
				Price:     0.0,
				Volume:    1000.0,
				Market:    entities.MarketTypeSpot,
				Type:      entities.InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "negative volume",
			ticker: entities.Ticker{
				Symbol:    "BTC/USDT",
				Price:     50000.0,
				Volume:    -1000.0,
				Market:    entities.MarketTypeSpot,
				Type:      entities.InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "future timestamp",
			ticker: entities.Ticker{
				Symbol:    "BTC/USDT",
				Price:     50000.0,
				Volume:    1000.0,
				Market:    entities.MarketTypeSpot,
				Type:      entities.InstrumentTypeSpot,
				Timestamp: time.Now().Add(10 * time.Minute),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "invalid bid/ask prices for stock",
			ticker: entities.Ticker{
				Symbol:    "AAPL",
				Price:     150.0,
				Volume:    1000.0,
				Market:    entities.MarketTypeStock,
				Type:      entities.InstrumentTypeStock,
				Timestamp: time.Now(),
				BrokerID:  "alpaca",
				BidPrice:  151.0, // Bid > Ask
				AskPrice:  150.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTicker(tt.ticker)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataValidatorService.ValidateTicker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataValidatorService_ValidateCandle(t *testing.T) {
	validator := NewDataValidatorService()

	tests := []struct {
		name    string
		candle  entities.Candle
		wantErr bool
	}{
		{
			name: "valid candle",
			candle: entities.Candle{
				Symbol:    "BTC/USDT",
				Open:      49000.0,
				High:      51000.0,
				Low:       48000.0,
				Close:     50000.0,
				Volume:    1000.0,
				Timeframe: "1h",
				Market:    entities.MarketTypeSpot,
				Type:      entities.InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: false,
		},
		{
			name: "invalid OHLC logic",
			candle: entities.Candle{
				Symbol:    "BTC/USDT",
				Open:      49000.0,
				High:      47000.0, // High < Low
				Low:       48000.0,
				Close:     50000.0,
				Volume:    1000.0,
				Timeframe: "1h",
				Market:    entities.MarketTypeSpot,
				Type:      entities.InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "invalid timeframe",
			candle: entities.Candle{
				Symbol:    "BTC/USDT",
				Open:      49000.0,
				High:      51000.0,
				Low:       48000.0,
				Close:     50000.0,
				Volume:    1000.0,
				Timeframe: "invalid",
				Market:    entities.MarketTypeSpot,
				Type:      entities.InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "negative trades count",
			candle: entities.Candle{
				Symbol:    "BTC/USDT",
				Open:      49000.0,
				High:      51000.0,
				Low:       48000.0,
				Close:     50000.0,
				Volume:    1000.0,
				Timeframe: "1h",
				Market:    entities.MarketTypeSpot,
				Type:      entities.InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
				Trades:    -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateCandle(tt.candle)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataValidatorService.ValidateCandle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataValidatorService_ValidateOrderBook(t *testing.T) {
	validator := NewDataValidatorService()

	tests := []struct {
		name      string
		orderBook entities.OrderBook
		wantErr   bool
	}{
		{
			name: "valid order book",
			orderBook: entities.OrderBook{
				Symbol:    "BTC/USDT",
				Market:    entities.MarketTypeSpot,
				Type:      entities.InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
				Bids: []entities.PriceLevel{
					{Price: 49900.0, Quantity: 1.0},
					{Price: 49800.0, Quantity: 2.0},
				},
				Asks: []entities.PriceLevel{
					{Price: 50100.0, Quantity: 1.5},
					{Price: 50200.0, Quantity: 2.5},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid bid/ask spread",
			orderBook: entities.OrderBook{
				Symbol:    "BTC/USDT",
				Market:    entities.MarketTypeSpot,
				Type:      entities.InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
				Bids: []entities.PriceLevel{
					{Price: 50200.0, Quantity: 1.0}, // Bid > Ask
				},
				Asks: []entities.PriceLevel{
					{Price: 50100.0, Quantity: 1.5},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid price level",
			orderBook: entities.OrderBook{
				Symbol:    "BTC/USDT",
				Market:    entities.MarketTypeSpot,
				Type:      entities.InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
				Bids: []entities.PriceLevel{
					{Price: -49900.0, Quantity: 1.0}, // Negative price
				},
				Asks: []entities.PriceLevel{
					{Price: 50100.0, Quantity: 1.5},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateOrderBook(tt.orderBook)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataValidatorService.ValidateOrderBook() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataValidatorService_ValidateTimeframe(t *testing.T) {
	validator := NewDataValidatorService()

	tests := []struct {
		name      string
		timeframe string
		wantErr   bool
	}{
		{"valid 1m", "1m", false},
		{"valid 1h", "1h", false},
		{"valid 1d", "1d", false},
		{"valid 1w", "1w", false},
		{"valid 1M", "1M", false},
		{"invalid timeframe", "invalid", true},
		{"empty timeframe", "", true},
		{"too long timeframe", "verylongtimeframe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateTimeframe(tt.timeframe)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataValidatorService.validateTimeframe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataValidatorService_ValidateMarketType(t *testing.T) {
	validator := NewDataValidatorService()

	tests := []struct {
		name       string
		marketType entities.MarketType
		wantErr    bool
	}{
		{"valid spot", entities.MarketTypeSpot, false},
		{"valid futures", entities.MarketTypeFutures, false},
		{"valid stock", entities.MarketTypeStock, false},
		{"invalid market type", entities.MarketType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateMarketType(tt.marketType)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataValidatorService.validateMarketType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDataValidatorService_ValidateInstrumentType(t *testing.T) {
	validator := NewDataValidatorService()

	tests := []struct {
		name           string
		instrumentType entities.InstrumentType
		wantErr        bool
	}{
		{"valid spot", entities.InstrumentTypeSpot, false},
		{"valid futures", entities.InstrumentTypeFutures, false},
		{"valid stock", entities.InstrumentTypeStock, false},
		{"valid ETF", entities.InstrumentTypeETF, false},
		{"valid bond", entities.InstrumentTypeBond, false},
		{"invalid instrument type", entities.InstrumentType("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateInstrumentType(tt.instrumentType)
			if (err != nil) != tt.wantErr {
				t.Errorf("DataValidatorService.validateInstrumentType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
