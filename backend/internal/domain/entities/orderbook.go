package entities

import (
	"sort"
	"time"
)

// PriceLevel represents a price level in the order book
type PriceLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// OrderBook represents order book with support for different market types
type OrderBook struct {
	Symbol    string         `json:"symbol"`
	Bids      []PriceLevel   `json:"bids"`
	Asks      []PriceLevel   `json:"asks"`
	Market    MarketType     `json:"market"`
	Type      InstrumentType `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	BrokerID  string         `json:"broker_id"`

	// Additional fields
	LastUpdateID int64     `json:"last_update_id,omitempty"`
	EventTime    time.Time `json:"event_time,omitempty"`
}

// IsValid checks order book validity
func (ob *OrderBook) IsValid() bool {
	if ob.Symbol == "" {
		return false
	}
	if ob.BrokerID == "" {
		return false
	}
	if ob.Timestamp.IsZero() {
		return false
	}
	if len(ob.Bids) == 0 && len(ob.Asks) == 0 {
		return false
	}

	// Check price levels validity
	for _, bid := range ob.Bids {
		if bid.Price <= 0 || bid.Quantity <= 0 {
			return false
		}
	}
	for _, ask := range ob.Asks {
		if ask.Price <= 0 || ask.Quantity <= 0 {
			return false
		}
	}

	return true
}

// GetBestBid returns best bid price
func (ob *OrderBook) GetBestBid() *PriceLevel {
	if len(ob.Bids) == 0 {
		return nil
	}

	// Find maximum price among bids
	best := ob.Bids[0]
	for _, bid := range ob.Bids[1:] {
		if bid.Price > best.Price {
			best = bid
		}
	}
	return &best
}

// GetBestAsk returns best ask price
func (ob *OrderBook) GetBestAsk() *PriceLevel {
	if len(ob.Asks) == 0 {
		return nil
	}

	// Find minimum price among asks
	best := ob.Asks[0]
	for _, ask := range ob.Asks[1:] {
		if ask.Price < best.Price {
			best = ask
		}
	}
	return &best
}

// GetSpread returns spread between best prices
func (ob *OrderBook) GetSpread() float64 {
	bestBid := ob.GetBestBid()
	bestAsk := ob.GetBestAsk()

	if bestBid == nil || bestAsk == nil {
		return 0
	}

	return bestAsk.Price - bestBid.Price
}

// GetMidPrice returns mid price
func (ob *OrderBook) GetMidPrice() float64 {
	bestBid := ob.GetBestBid()
	bestAsk := ob.GetBestAsk()

	if bestBid == nil || bestAsk == nil {
		return 0
	}

	return (bestBid.Price + bestAsk.Price) / 2
}

// SortBids sorts bids by descending price
func (ob *OrderBook) SortBids() {
	sort.Slice(ob.Bids, func(i, j int) bool {
		return ob.Bids[i].Price > ob.Bids[j].Price
	})
}

// SortAsks sorts asks by ascending price
func (ob *OrderBook) SortAsks() {
	sort.Slice(ob.Asks, func(i, j int) bool {
		return ob.Asks[i].Price < ob.Asks[j].Price
	})
}

// Sort sorts both bids and asks
func (ob *OrderBook) Sort() {
	ob.SortBids()
	ob.SortAsks()
}

// GetTotalBidVolume returns total bid volume
func (ob *OrderBook) GetTotalBidVolume() float64 {
	total := 0.0
	for _, bid := range ob.Bids {
		total += bid.Quantity
	}
	return total
}

// GetTotalAskVolume returns total ask volume
func (ob *OrderBook) GetTotalAskVolume() float64 {
	total := 0.0
	for _, ask := range ob.Asks {
		total += ask.Quantity
	}
	return total
}

// GetDepth returns order book depth (number of levels)
func (ob *OrderBook) GetDepth() (bidDepth, askDepth int) {
	return len(ob.Bids), len(ob.Asks)
}
