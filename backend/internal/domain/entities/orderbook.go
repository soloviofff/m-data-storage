package entities

import (
	"sort"
	"time"
)

// PriceLevel представляет уровень цены в ордербуке
type PriceLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

// OrderBook представляет ордербук с поддержкой разных типов рынков
type OrderBook struct {
	Symbol    string         `json:"symbol"`
	Bids      []PriceLevel   `json:"bids"`
	Asks      []PriceLevel   `json:"asks"`
	Market    MarketType     `json:"market"`
	Type      InstrumentType `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	BrokerID  string         `json:"broker_id"`

	// Дополнительные поля
	LastUpdateID int64     `json:"last_update_id,omitempty"`
	EventTime    time.Time `json:"event_time,omitempty"`
}

// IsValid проверяет валидность ордербука
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

	// Проверяем валидность уровней цен
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

// GetBestBid возвращает лучшую цену покупки
func (ob *OrderBook) GetBestBid() *PriceLevel {
	if len(ob.Bids) == 0 {
		return nil
	}

	// Находим максимальную цену среди bids
	best := ob.Bids[0]
	for _, bid := range ob.Bids[1:] {
		if bid.Price > best.Price {
			best = bid
		}
	}
	return &best
}

// GetBestAsk возвращает лучшую цену продажи
func (ob *OrderBook) GetBestAsk() *PriceLevel {
	if len(ob.Asks) == 0 {
		return nil
	}

	// Находим минимальную цену среди asks
	best := ob.Asks[0]
	for _, ask := range ob.Asks[1:] {
		if ask.Price < best.Price {
			best = ask
		}
	}
	return &best
}

// GetSpread возвращает спред между лучшими ценами
func (ob *OrderBook) GetSpread() float64 {
	bestBid := ob.GetBestBid()
	bestAsk := ob.GetBestAsk()

	if bestBid == nil || bestAsk == nil {
		return 0
	}

	return bestAsk.Price - bestBid.Price
}

// GetMidPrice возвращает среднюю цену
func (ob *OrderBook) GetMidPrice() float64 {
	bestBid := ob.GetBestBid()
	bestAsk := ob.GetBestAsk()

	if bestBid == nil || bestAsk == nil {
		return 0
	}

	return (bestBid.Price + bestAsk.Price) / 2
}

// SortBids сортирует bids по убыванию цены
func (ob *OrderBook) SortBids() {
	sort.Slice(ob.Bids, func(i, j int) bool {
		return ob.Bids[i].Price > ob.Bids[j].Price
	})
}

// SortAsks сортирует asks по возрастанию цены
func (ob *OrderBook) SortAsks() {
	sort.Slice(ob.Asks, func(i, j int) bool {
		return ob.Asks[i].Price < ob.Asks[j].Price
	})
}

// Sort сортирует и bids, и asks
func (ob *OrderBook) Sort() {
	ob.SortBids()
	ob.SortAsks()
}

// GetTotalBidVolume возвращает общий объем заявок на покупку
func (ob *OrderBook) GetTotalBidVolume() float64 {
	total := 0.0
	for _, bid := range ob.Bids {
		total += bid.Quantity
	}
	return total
}

// GetTotalAskVolume возвращает общий объем заявок на продажу
func (ob *OrderBook) GetTotalAskVolume() float64 {
	total := 0.0
	for _, ask := range ob.Asks {
		total += ask.Quantity
	}
	return total
}

// GetDepth возвращает глубину ордербука (количество уровней)
func (ob *OrderBook) GetDepth() (bidDepth, askDepth int) {
	return len(ob.Bids), len(ob.Asks)
}
