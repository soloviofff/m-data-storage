package entities

import (
	"testing"
	"time"
)

func TestTicker_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ticker  Ticker
		wantErr bool
	}{
		{
			name: "valid ticker",
			ticker: Ticker{
				Symbol:    "BTC/USDT",
				Price:     50000.0,
				Volume:    1000.0,
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: false,
		},
		{
			name: "empty symbol",
			ticker: Ticker{
				Symbol:    "",
				Price:     50000.0,
				Volume:    1000.0,
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "negative price",
			ticker: Ticker{
				Symbol:    "BTC/USDT",
				Price:     -50000.0,
				Volume:    1000.0,
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "negative volume",
			ticker: Ticker{
				Symbol:    "BTC/USDT",
				Price:     50000.0,
				Volume:    -1000.0,
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "zero timestamp",
			ticker: Ticker{
				Symbol:    "BTC/USDT",
				Price:     50000.0,
				Volume:    1000.0,
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Time{},
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "empty broker ID",
			ticker: Ticker{
				Symbol:    "BTC/USDT",
				Price:     50000.0,
				Volume:    1000.0,
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ticker.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Ticker.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTicker_GetSpread(t *testing.T) {
	ticker := Ticker{
		Type:     InstrumentTypeStock,
		BidPrice: 49900.0,
		AskPrice: 50100.0,
	}

	spread := ticker.GetSpread()
	expected := 200.0

	if spread != expected {
		t.Errorf("Ticker.GetSpread() = %v, want %v", spread, expected)
	}
}

func TestTicker_GetMidPrice(t *testing.T) {
	ticker := Ticker{
		Type:     InstrumentTypeStock,
		BidPrice: 49900.0,
		AskPrice: 50100.0,
	}

	midPrice := ticker.GetMidPrice()
	expected := 50000.0

	if midPrice != expected {
		t.Errorf("Ticker.GetMidPrice() = %v, want %v", midPrice, expected)
	}
}

func TestTicker_GetPriceChange24h(t *testing.T) {
	ticker := Ticker{
		Price:        50000.0,
		PrevClose24h: 48000.0,
	}

	change := ticker.GetPriceChange24h()
	expected := 2000.0

	if change != expected {
		t.Errorf("Ticker.GetPriceChange24h() = %v, want %v", change, expected)
	}
}

func TestTicker_GetPriceChangePercent24h(t *testing.T) {
	ticker := Ticker{
		Price:        50000.0,
		PrevClose24h: 48000.0,
	}

	changePercent := ticker.GetPriceChangePercent24h()
	expected := 4.166666666666666 // (50000-48000)/48000 * 100

	if changePercent != expected {
		t.Errorf("Ticker.GetPriceChangePercent24h() = %v, want %v", changePercent, expected)
	}
}

func TestTicker_IsValidForMarket(t *testing.T) {
	tests := []struct {
		name   string
		ticker Ticker
		want   bool
	}{
		{
			name: "valid spot ticker",
			ticker: Ticker{
				Market: MarketTypeSpot,
				Type:   InstrumentTypeSpot,
			},
			want: true,
		},
		{
			name: "valid futures ticker",
			ticker: Ticker{
				Market: MarketTypeFutures,
				Type:   InstrumentTypeFutures,
			},
			want: true,
		},
		{
			name: "valid stock ticker",
			ticker: Ticker{
				Market: MarketTypeStock,
				Type:   InstrumentTypeStock,
			},
			want: true,
		},
		{
			name: "invalid market/type combination",
			ticker: Ticker{
				Market: MarketTypeSpot,
				Type:   InstrumentTypeFutures,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ticker.IsValidForMarket(); got != tt.want {
				t.Errorf("Ticker.IsValidForMarket() = %v, want %v", got, tt.want)
			}
		})
	}
}
