package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOrderBook_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		orderbook OrderBook
		want      bool
	}{
		{
			name: "valid orderbook",
			orderbook: OrderBook{
				Symbol: "BTC/USDT",
				Bids: []PriceLevel{
					{Price: 50000.0, Quantity: 1.0},
				},
				Asks: []PriceLevel{
					{Price: 50100.0, Quantity: 1.0},
				},
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			want: true,
		},
		{
			name: "empty symbol",
			orderbook: OrderBook{
				Symbol: "",
				Bids: []PriceLevel{
					{Price: 50000.0, Quantity: 1.0},
				},
				Asks: []PriceLevel{
					{Price: 50100.0, Quantity: 1.0},
				},
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			want: false,
		},
		{
			name: "empty broker ID",
			orderbook: OrderBook{
				Symbol: "BTC/USDT",
				Bids: []PriceLevel{
					{Price: 50000.0, Quantity: 1.0},
				},
				Asks: []PriceLevel{
					{Price: 50100.0, Quantity: 1.0},
				},
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "",
			},
			want: false,
		},
		{
			name: "zero timestamp",
			orderbook: OrderBook{
				Symbol: "BTC/USDT",
				Bids: []PriceLevel{
					{Price: 50000.0, Quantity: 1.0},
				},
				Asks: []PriceLevel{
					{Price: 50100.0, Quantity: 1.0},
				},
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Time{},
				BrokerID:  "binance",
			},
			want: false,
		},
		{
			name: "empty bids and asks",
			orderbook: OrderBook{
				Symbol:    "BTC/USDT",
				Bids:      []PriceLevel{},
				Asks:      []PriceLevel{},
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			want: false,
		},
		{
			name: "invalid bid price",
			orderbook: OrderBook{
				Symbol: "BTC/USDT",
				Bids: []PriceLevel{
					{Price: -50000.0, Quantity: 1.0},
				},
				Asks: []PriceLevel{
					{Price: 50100.0, Quantity: 1.0},
				},
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			want: false,
		},
		{
			name: "invalid bid quantity",
			orderbook: OrderBook{
				Symbol: "BTC/USDT",
				Bids: []PriceLevel{
					{Price: 50000.0, Quantity: -1.0},
				},
				Asks: []PriceLevel{
					{Price: 50100.0, Quantity: 1.0},
				},
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			want: false,
		},
		{
			name: "invalid ask price",
			orderbook: OrderBook{
				Symbol: "BTC/USDT",
				Bids: []PriceLevel{
					{Price: 50000.0, Quantity: 1.0},
				},
				Asks: []PriceLevel{
					{Price: 0.0, Quantity: 1.0},
				},
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			want: false,
		},
		{
			name: "invalid ask quantity",
			orderbook: OrderBook{
				Symbol: "BTC/USDT",
				Bids: []PriceLevel{
					{Price: 50000.0, Quantity: 1.0},
				},
				Asks: []PriceLevel{
					{Price: 50100.0, Quantity: 0.0},
				},
				Market:    MarketTypeSpot,
				Type:      InstrumentTypeSpot,
				Timestamp: time.Now(),
				BrokerID:  "binance",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.orderbook.IsValid())
		})
	}
}

func TestOrderBook_GetBestBid(t *testing.T) {
	tests := []struct {
		name      string
		orderbook OrderBook
		want      *PriceLevel
	}{
		{
			name: "multiple bids",
			orderbook: OrderBook{
				Bids: []PriceLevel{
					{Price: 49900.0, Quantity: 1.0},
					{Price: 50000.0, Quantity: 2.0},
					{Price: 49950.0, Quantity: 1.5},
				},
			},
			want: &PriceLevel{Price: 50000.0, Quantity: 2.0},
		},
		{
			name: "single bid",
			orderbook: OrderBook{
				Bids: []PriceLevel{
					{Price: 50000.0, Quantity: 1.0},
				},
			},
			want: &PriceLevel{Price: 50000.0, Quantity: 1.0},
		},
		{
			name: "no bids",
			orderbook: OrderBook{
				Bids: []PriceLevel{},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.orderbook.GetBestBid()
			if tt.want == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.want.Price, result.Price)
				assert.Equal(t, tt.want.Quantity, result.Quantity)
			}
		})
	}
}

func TestOrderBook_GetBestAsk(t *testing.T) {
	tests := []struct {
		name      string
		orderbook OrderBook
		want      *PriceLevel
	}{
		{
			name: "multiple asks",
			orderbook: OrderBook{
				Asks: []PriceLevel{
					{Price: 50200.0, Quantity: 1.0},
					{Price: 50100.0, Quantity: 2.0},
					{Price: 50150.0, Quantity: 1.5},
				},
			},
			want: &PriceLevel{Price: 50100.0, Quantity: 2.0},
		},
		{
			name: "single ask",
			orderbook: OrderBook{
				Asks: []PriceLevel{
					{Price: 50100.0, Quantity: 1.0},
				},
			},
			want: &PriceLevel{Price: 50100.0, Quantity: 1.0},
		},
		{
			name: "no asks",
			orderbook: OrderBook{
				Asks: []PriceLevel{},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.orderbook.GetBestAsk()
			if tt.want == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.want.Price, result.Price)
				assert.Equal(t, tt.want.Quantity, result.Quantity)
			}
		})
	}
}

func TestOrderBook_GetSpread(t *testing.T) {
	tests := []struct {
		name      string
		orderbook OrderBook
		want      float64
	}{
		{
			name: "normal spread",
			orderbook: OrderBook{
				Bids: []PriceLevel{
					{Price: 50000.0, Quantity: 1.0},
				},
				Asks: []PriceLevel{
					{Price: 50100.0, Quantity: 1.0},
				},
			},
			want: 100.0,
		},
		{
			name: "no bids",
			orderbook: OrderBook{
				Bids: []PriceLevel{},
				Asks: []PriceLevel{
					{Price: 50100.0, Quantity: 1.0},
				},
			},
			want: 0.0,
		},
		{
			name: "no asks",
			orderbook: OrderBook{
				Bids: []PriceLevel{
					{Price: 50000.0, Quantity: 1.0},
				},
				Asks: []PriceLevel{},
			},
			want: 0.0,
		},
		{
			name: "no bids and asks",
			orderbook: OrderBook{
				Bids: []PriceLevel{},
				Asks: []PriceLevel{},
			},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.orderbook.GetSpread())
		})
	}
}

func TestOrderBook_GetMidPrice(t *testing.T) {
	tests := []struct {
		name      string
		orderbook OrderBook
		want      float64
	}{
		{
			name: "normal mid price",
			orderbook: OrderBook{
				Bids: []PriceLevel{
					{Price: 50000.0, Quantity: 1.0},
				},
				Asks: []PriceLevel{
					{Price: 50100.0, Quantity: 1.0},
				},
			},
			want: 50050.0,
		},
		{
			name: "no bids",
			orderbook: OrderBook{
				Bids: []PriceLevel{},
				Asks: []PriceLevel{
					{Price: 50100.0, Quantity: 1.0},
				},
			},
			want: 0.0,
		},
		{
			name: "no asks",
			orderbook: OrderBook{
				Bids: []PriceLevel{
					{Price: 50000.0, Quantity: 1.0},
				},
				Asks: []PriceLevel{},
			},
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.orderbook.GetMidPrice())
		})
	}
}

func TestOrderBook_Sort(t *testing.T) {
	orderbook := OrderBook{
		Bids: []PriceLevel{
			{Price: 49900.0, Quantity: 1.0},
			{Price: 50000.0, Quantity: 2.0},
			{Price: 49950.0, Quantity: 1.5},
		},
		Asks: []PriceLevel{
			{Price: 50200.0, Quantity: 1.0},
			{Price: 50100.0, Quantity: 2.0},
			{Price: 50150.0, Quantity: 1.5},
		},
	}

	orderbook.Sort()

	// Check bids are sorted by descending price
	assert.Equal(t, 50000.0, orderbook.Bids[0].Price)
	assert.Equal(t, 49950.0, orderbook.Bids[1].Price)
	assert.Equal(t, 49900.0, orderbook.Bids[2].Price)

	// Check asks are sorted by ascending price
	assert.Equal(t, 50100.0, orderbook.Asks[0].Price)
	assert.Equal(t, 50150.0, orderbook.Asks[1].Price)
	assert.Equal(t, 50200.0, orderbook.Asks[2].Price)
}

func TestOrderBook_SortBids(t *testing.T) {
	orderbook := OrderBook{
		Bids: []PriceLevel{
			{Price: 49900.0, Quantity: 1.0},
			{Price: 50000.0, Quantity: 2.0},
			{Price: 49950.0, Quantity: 1.5},
		},
	}

	orderbook.SortBids()

	// Check bids are sorted by descending price
	assert.Equal(t, 50000.0, orderbook.Bids[0].Price)
	assert.Equal(t, 49950.0, orderbook.Bids[1].Price)
	assert.Equal(t, 49900.0, orderbook.Bids[2].Price)
}

func TestOrderBook_SortAsks(t *testing.T) {
	orderbook := OrderBook{
		Asks: []PriceLevel{
			{Price: 50200.0, Quantity: 1.0},
			{Price: 50100.0, Quantity: 2.0},
			{Price: 50150.0, Quantity: 1.5},
		},
	}

	orderbook.SortAsks()

	// Check asks are sorted by ascending price
	assert.Equal(t, 50100.0, orderbook.Asks[0].Price)
	assert.Equal(t, 50150.0, orderbook.Asks[1].Price)
	assert.Equal(t, 50200.0, orderbook.Asks[2].Price)
}

func TestOrderBook_GetTotalBidVolume(t *testing.T) {
	orderbook := OrderBook{
		Bids: []PriceLevel{
			{Price: 50000.0, Quantity: 1.0},
			{Price: 49950.0, Quantity: 2.5},
			{Price: 49900.0, Quantity: 1.5},
		},
	}

	total := orderbook.GetTotalBidVolume()
	assert.Equal(t, 5.0, total)
}

func TestOrderBook_GetTotalAskVolume(t *testing.T) {
	orderbook := OrderBook{
		Asks: []PriceLevel{
			{Price: 50100.0, Quantity: 1.0},
			{Price: 50150.0, Quantity: 2.5},
			{Price: 50200.0, Quantity: 1.5},
		},
	}

	total := orderbook.GetTotalAskVolume()
	assert.Equal(t, 5.0, total)
}

func TestOrderBook_GetDepth(t *testing.T) {
	orderbook := OrderBook{
		Bids: []PriceLevel{
			{Price: 50000.0, Quantity: 1.0},
			{Price: 49950.0, Quantity: 2.5},
		},
		Asks: []PriceLevel{
			{Price: 50100.0, Quantity: 1.0},
			{Price: 50150.0, Quantity: 2.5},
			{Price: 50200.0, Quantity: 1.5},
		},
	}

	bidDepth, askDepth := orderbook.GetDepth()
	assert.Equal(t, 2, bidDepth)
	assert.Equal(t, 3, askDepth)
}
