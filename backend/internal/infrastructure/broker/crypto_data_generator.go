package broker

import (
	"math/rand"
	"time"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// CryptoDataGenerator generates realistic data for crypto exchanges
type CryptoDataGenerator struct {
	spotMarkets    []interfaces.SpotMarket
	futuresMarkets []interfaces.FuturesMarket

	// Current prices for simulation
	currentPrices map[string]float64

	// Base prices for different symbols
	basePrices map[string]float64
}

// NewCryptoDataGenerator creates a new data generator
func NewCryptoDataGenerator(spotMarkets []interfaces.SpotMarket, futuresMarkets []interfaces.FuturesMarket) *CryptoDataGenerator {
	generator := &CryptoDataGenerator{
		spotMarkets:    spotMarkets,
		futuresMarkets: futuresMarkets,
		currentPrices:  make(map[string]float64),
		basePrices:     make(map[string]float64),
	}

	// Initialize base prices
	generator.initializeBasePrices()

	return generator
}

// initializeBasePrices sets initial prices for symbols
func (g *CryptoDataGenerator) initializeBasePrices() {
	// Realistic base prices for popular cryptocurrencies
	g.basePrices["BTCUSDT"] = 45000.0
	g.basePrices["ETHUSDT"] = 3000.0
	g.basePrices["ADAUSDT"] = 0.5
	g.basePrices["BNBUSDT"] = 300.0
	g.basePrices["XRPUSDT"] = 0.6
	g.basePrices["SOLUSDT"] = 100.0
	g.basePrices["DOTUSDT"] = 7.0
	g.basePrices["LINKUSDT"] = 15.0

	// Copy base prices to current prices
	for symbol, price := range g.basePrices {
		g.currentPrices[symbol] = price
	}
}

// GenerateTicker generates realistic ticker
func (g *CryptoDataGenerator) GenerateTicker(symbol string, market entities.MarketType, timestamp time.Time) *entities.Ticker {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		// If symbol not found, use random price
		basePrice = rand.Float64() * 1000
		g.currentPrices[symbol] = basePrice
	}

	// Generate small price change (±2%)
	priceChange := (rand.Float64() - 0.5) * 0.04 * basePrice
	newPrice := basePrice + priceChange

	// Update current price
	g.currentPrices[symbol] = newPrice

	// Generate volume
	volume := rand.Float64() * 100

	ticker := &entities.Ticker{
		Symbol:    symbol,
		Price:     newPrice,
		Volume:    volume,
		Market:    market,
		Timestamp: timestamp,
	}

	// Set instrument type
	switch market {
	case entities.MarketTypeSpot:
		ticker.Type = entities.InstrumentTypeSpot
	case entities.MarketTypeFutures:
		ticker.Type = entities.InstrumentTypeFutures
	}

	// Add additional fields
	ticker.Change = priceChange
	ticker.ChangePercent = (priceChange / basePrice) * 100
	ticker.High24h = newPrice * (1 + rand.Float64()*0.05)
	ticker.Low24h = newPrice * (1 - rand.Float64()*0.05)
	ticker.Volume24h = volume * (20 + rand.Float64()*10)

	// For futures add open interest
	if market == entities.MarketTypeFutures {
		ticker.OpenInterest = rand.Float64() * 1000000
	}

	// For spot add bid/ask
	if market == entities.MarketTypeSpot {
		spread := newPrice * 0.001 // 0.1% spread
		ticker.BidPrice = newPrice - spread/2
		ticker.AskPrice = newPrice + spread/2
		ticker.BidSize = rand.Float64() * 10
		ticker.AskSize = rand.Float64() * 10
	}

	return ticker
}

// GenerateCandle generates realistic candle
func (g *CryptoDataGenerator) GenerateCandle(symbol string, market entities.MarketType, timestamp time.Time) *entities.Candle {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = rand.Float64() * 1000
		g.currentPrices[symbol] = basePrice
	}

	// Generate OHLC data
	open := basePrice
	volatility := 0.02 // 2% volatility

	high := open * (1 + rand.Float64()*volatility)
	low := open * (1 - rand.Float64()*volatility)
	close := low + rand.Float64()*(high-low)

	// Update current price
	g.currentPrices[symbol] = close

	volume := rand.Float64() * 1000

	candle := &entities.Candle{
		Symbol:    symbol,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
		Market:    market,
		Timestamp: timestamp,
		Timeframe: "1m",
	}

	// Set instrument type
	switch market {
	case entities.MarketTypeSpot:
		candle.Type = entities.InstrumentTypeSpot
	case entities.MarketTypeFutures:
		candle.Type = entities.InstrumentTypeFutures
	}

	// Add additional fields
	candle.Trades = int64(rand.Intn(1000) + 100)
	candle.QuoteVolume = volume * close

	// For futures add open interest
	if market == entities.MarketTypeFutures {
		candle.OpenInterest = rand.Float64() * 1000000
	}

	return candle
}

// GenerateOrderBook generates realistic order book
func (g *CryptoDataGenerator) GenerateOrderBook(symbol string, market entities.MarketType, timestamp time.Time) *entities.OrderBook {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = rand.Float64() * 1000
		g.currentPrices[symbol] = basePrice
	}

	// Generate bid and ask levels
	bids := make([]entities.PriceLevel, 0, 20)
	asks := make([]entities.PriceLevel, 0, 20)

	// Generate bid levels (below current price)
	for i := 0; i < 20; i++ {
		price := basePrice * (1 - float64(i+1)*0.001) // Each level 0.1% lower
		quantity := rand.Float64() * 10
		bids = append(bids, entities.PriceLevel{
			Price:    price,
			Quantity: quantity,
		})
	}

	// Generate ask levels (above current price)
	for i := 0; i < 20; i++ {
		price := basePrice * (1 + float64(i+1)*0.001) // Each level 0.1% higher
		quantity := rand.Float64() * 10
		asks = append(asks, entities.PriceLevel{
			Price:    price,
			Quantity: quantity,
		})
	}

	orderBook := &entities.OrderBook{
		Symbol:    symbol,
		Bids:      bids,
		Asks:      asks,
		Market:    market,
		Timestamp: timestamp,
	}

	// Set instrument type
	switch market {
	case entities.MarketTypeSpot:
		orderBook.Type = entities.InstrumentTypeSpot
	case entities.MarketTypeFutures:
		orderBook.Type = entities.InstrumentTypeFutures
	}

	// Add additional fields
	orderBook.LastUpdateID = rand.Int63n(1000000)
	orderBook.EventTime = timestamp

	return orderBook
}

// GenerateFundingRate generates funding rate
func (g *CryptoDataGenerator) GenerateFundingRate(symbol string, timestamp time.Time) interfaces.FundingRate {
	// Generate realistic funding rate (-0.1% to 0.1%)
	rate := (rand.Float64() - 0.5) * 0.002

	// Next funding time (every 8 hours)
	nextTime := timestamp.Truncate(8 * time.Hour).Add(8 * time.Hour)

	return interfaces.FundingRate{
		Symbol:    symbol,
		Rate:      rate,
		NextTime:  nextTime,
		Timestamp: timestamp,
	}
}

// GenerateMarkPrice generates mark price
func (g *CryptoDataGenerator) GenerateMarkPrice(symbol string, timestamp time.Time) interfaces.MarkPrice {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = rand.Float64() * 1000
	}

	// Mark price is usually close to spot price
	markPrice := basePrice * (1 + (rand.Float64()-0.5)*0.001)   // ±0.05%
	indexPrice := basePrice * (1 + (rand.Float64()-0.5)*0.0005) // ±0.025%

	return interfaces.MarkPrice{
		Symbol:     symbol,
		Price:      markPrice,
		IndexPrice: indexPrice,
		Timestamp:  timestamp,
	}
}

// GenerateLiquidation generates liquidation
func (g *CryptoDataGenerator) GenerateLiquidation(symbol string, timestamp time.Time) interfaces.Liquidation {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = rand.Float64() * 1000
	}

	// Liquidation price may differ from current price
	liquidationPrice := basePrice * (1 + (rand.Float64()-0.5)*0.02) // ±1%

	// Random side
	side := "BUY"
	if rand.Float32() < 0.5 {
		side = "SELL"
	}

	// Random quantity
	quantity := rand.Float64() * 100

	return interfaces.Liquidation{
		Symbol:    symbol,
		Side:      side,
		Price:     liquidationPrice,
		Quantity:  quantity,
		Timestamp: timestamp,
	}
}

// GetCurrentPrice returns current price of symbol
func (g *CryptoDataGenerator) GetCurrentPrice(symbol string) float64 {
	if price, exists := g.currentPrices[symbol]; exists {
		return price
	}
	return 0
}

// SetCurrentPrice sets current price of symbol
func (g *CryptoDataGenerator) SetCurrentPrice(symbol string, price float64) {
	g.currentPrices[symbol] = price
}

// GenerateRealisticPriceMovement generates realistic price movement
func (g *CryptoDataGenerator) GenerateRealisticPriceMovement(symbol string, volatility float64) float64 {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = g.basePrices[symbol]
		if basePrice == 0 {
			basePrice = rand.Float64() * 1000
		}
	}

	// Use random walk model with mean reversion
	meanReversion := 0.01 // Mean reversion strength
	baseValue := g.basePrices[symbol]

	// Random component
	randomComponent := (rand.Float64() - 0.5) * volatility * basePrice

	// Mean reversion component
	meanReversionComponent := (baseValue - basePrice) * meanReversion

	// New price
	newPrice := basePrice + randomComponent + meanReversionComponent

	// Ensure price is positive
	if newPrice <= 0 {
		newPrice = basePrice * 0.99
	}

	g.currentPrices[symbol] = newPrice
	return newPrice
}
