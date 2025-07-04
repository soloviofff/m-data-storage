package entities

import (
	"testing"
	"time"
)

func TestCandle_Validate(t *testing.T) {
	tests := []struct {
		name    string
		candle  Candle
		wantErr bool
	}{
		{
			name: "valid candle",
			candle: Candle{
				Symbol:    "BTC/USDT",
				Open:      49000.0,
				High:      51000.0,
				Low:       48000.0,
				Close:     50000.0,
				Volume:    1000.0,
				Timeframe: "1h",
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: false,
		},
		{
			name: "empty symbol",
			candle: Candle{
				Symbol:    "",
				Open:      49000.0,
				High:      51000.0,
				Low:       48000.0,
				Close:     50000.0,
				Volume:    1000.0,
				Timeframe: "1h",
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "high less than low",
			candle: Candle{
				Symbol:    "BTC/USDT",
				Open:      49000.0,
				High:      47000.0, // High < Low
				Low:       48000.0,
				Close:     50000.0,
				Volume:    1000.0,
				Timeframe: "1h",
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "open above high",
			candle: Candle{
				Symbol:    "BTC/USDT",
				Open:      52000.0, // Open > High
				High:      51000.0,
				Low:       48000.0,
				Close:     50000.0,
				Volume:    1000.0,
				Timeframe: "1h",
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "close below low",
			candle: Candle{
				Symbol:    "BTC/USDT",
				Open:      49000.0,
				High:      51000.0,
				Low:       48000.0,
				Close:     47000.0, // Close < Low
				Volume:    1000.0,
				Timeframe: "1h",
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
		{
			name: "negative volume",
			candle: Candle{
				Symbol:    "BTC/USDT",
				Open:      49000.0,
				High:      51000.0,
				Low:       48000.0,
				Close:     50000.0,
				Volume:    -1000.0, // Negative volume
				Timeframe: "1h",
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.candle.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Candle.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCandle_GetBodySize(t *testing.T) {
	tests := []struct {
		name   string
		candle Candle
		want   float64
	}{
		{
			name: "bullish candle",
			candle: Candle{
				Open:  49000.0,
				Close: 50000.0,
			},
			want: 1000.0,
		},
		{
			name: "bearish candle",
			candle: Candle{
				Open:  50000.0,
				Close: 49000.0,
			},
			want: 1000.0,
		},
		{
			name: "doji candle",
			candle: Candle{
				Open:  50000.0,
				Close: 50000.0,
			},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.candle.GetBodySize(); got != tt.want {
				t.Errorf("Candle.GetBodySize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCandle_GetUpperShadow(t *testing.T) {
	tests := []struct {
		name   string
		candle Candle
		want   float64
	}{
		{
			name: "bullish candle with upper shadow",
			candle: Candle{
				Open:  49000.0,
				High:  51000.0,
				Close: 50000.0,
			},
			want: 1000.0, // High - Close
		},
		{
			name: "bearish candle with upper shadow",
			candle: Candle{
				Open:  50000.0,
				High:  51000.0,
				Close: 49000.0,
			},
			want: 1000.0, // High - Open
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.candle.GetUpperShadow(); got != tt.want {
				t.Errorf("Candle.GetUpperShadow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCandle_GetLowerShadow(t *testing.T) {
	tests := []struct {
		name   string
		candle Candle
		want   float64
	}{
		{
			name: "bullish candle with lower shadow",
			candle: Candle{
				Open:  49000.0,
				Low:   48000.0,
				Close: 50000.0,
			},
			want: 1000.0, // Open - Low
		},
		{
			name: "bearish candle with lower shadow",
			candle: Candle{
				Open:  50000.0,
				Low:   48000.0,
				Close: 49000.0,
			},
			want: 1000.0, // Close - Low
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.candle.GetLowerShadow(); got != tt.want {
				t.Errorf("Candle.GetLowerShadow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCandle_IsBullish(t *testing.T) {
	tests := []struct {
		name   string
		candle Candle
		want   bool
	}{
		{
			name: "bullish candle",
			candle: Candle{
				Open:  49000.0,
				Close: 50000.0,
			},
			want: true,
		},
		{
			name: "bearish candle",
			candle: Candle{
				Open:  50000.0,
				Close: 49000.0,
			},
			want: false,
		},
		{
			name: "doji candle",
			candle: Candle{
				Open:  50000.0,
				Close: 50000.0,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.candle.IsBullish(); got != tt.want {
				t.Errorf("Candle.IsBullish() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCandle_IsBearish(t *testing.T) {
	tests := []struct {
		name   string
		candle Candle
		want   bool
	}{
		{
			name: "bearish candle",
			candle: Candle{
				Open:  50000.0,
				Close: 49000.0,
			},
			want: true,
		},
		{
			name: "bullish candle",
			candle: Candle{
				Open:  49000.0,
				Close: 50000.0,
			},
			want: false,
		},
		{
			name: "doji candle",
			candle: Candle{
				Open:  50000.0,
				Close: 50000.0,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.candle.IsBearish(); got != tt.want {
				t.Errorf("Candle.IsBearish() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCandle_IsDoji(t *testing.T) {
	tests := []struct {
		name   string
		candle Candle
		want   bool
	}{
		{
			name: "doji candle",
			candle: Candle{
				Open:  50000.0,
				Close: 50000.0,
			},
			want: true,
		},
		{
			name: "bullish candle",
			candle: Candle{
				Open:  49000.0,
				Close: 50000.0,
			},
			want: false,
		},
		{
			name: "bearish candle",
			candle: Candle{
				Open:  50000.0,
				Close: 49000.0,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.candle.IsDoji(); got != tt.want {
				t.Errorf("Candle.IsDoji() = %v, want %v", got, tt.want)
			}
		})
	}
}
