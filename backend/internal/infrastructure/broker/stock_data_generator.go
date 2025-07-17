package broker

import (
	"fmt"
	"math/rand"
	"time"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// StockDataGenerator generates realistic data for stock market
type StockDataGenerator struct {
	stockMarkets []interfaces.StockMarket
	sectors      []interfaces.Sector

	// Current prices for simulation
	currentPrices map[string]float64

	// Base prices for different symbols
	basePrices map[string]float64
}

// NewStockDataGenerator creates a new data generator
func NewStockDataGenerator(stockMarkets []interfaces.StockMarket, sectors []interfaces.Sector) *StockDataGenerator {
	generator := &StockDataGenerator{
		stockMarkets:  stockMarkets,
		sectors:       sectors,
		currentPrices: make(map[string]float64),
		basePrices:    make(map[string]float64),
	}

	// Initialize base prices
	generator.initializeBasePrices()

	return generator
}

// initializeBasePrices sets initial prices for symbols
func (g *StockDataGenerator) initializeBasePrices() {
	// Realistic base prices for popular stocks
	g.basePrices["AAPL"] = 180.0
	g.basePrices["MSFT"] = 380.0
	g.basePrices["GOOGL"] = 130.0
	g.basePrices["TSLA"] = 250.0
	g.basePrices["JPM"] = 150.0
	g.basePrices["AMZN"] = 140.0
	g.basePrices["NVDA"] = 450.0
	g.basePrices["META"] = 320.0
	g.basePrices["BRK.A"] = 540000.0 // Berkshire Hathaway Class A
	g.basePrices["JNJ"] = 160.0

	// Copy base prices to current prices
	for symbol, price := range g.basePrices {
		g.currentPrices[symbol] = price
	}
}

// GenerateTicker generates realistic ticker for stock
func (g *StockDataGenerator) GenerateTicker(symbol string, market entities.MarketType, timestamp time.Time) *entities.Ticker {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		// If symbol not found, use random price
		basePrice = 50.0 + rand.Float64()*200.0 // From $50 to $250
		g.currentPrices[symbol] = basePrice
	}

	// Generate small price change (±1% for stocks, less volatility than crypto)
	priceChange := (rand.Float64() - 0.5) * 0.02 * basePrice
	newPrice := basePrice + priceChange

	// Ensure price is positive
	if newPrice <= 0 {
		newPrice = basePrice * 0.99
	}

	// Update current price
	g.currentPrices[symbol] = newPrice

	// Generate volume (stocks trade in smaller volumes)
	volume := rand.Float64() * 1000000 // Up to 1M shares

	ticker := &entities.Ticker{
		Symbol:    symbol,
		Price:     newPrice,
		Volume:    volume,
		Market:    market,
		Type:      entities.InstrumentTypeStock,
		Timestamp: timestamp,
	}

	// Add additional fields
	ticker.Change = priceChange
	ticker.ChangePercent = (priceChange / basePrice) * 100
	ticker.High24h = newPrice * (1 + rand.Float64()*0.03) // Daily high
	ticker.Low24h = newPrice * (1 - rand.Float64()*0.03)  // Daily low
	ticker.Volume24h = volume * (5 + rand.Float64()*5)    // Daily volume

	// Bid/Ask spread for stocks
	spread := newPrice * (0.0001 + rand.Float64()*0.0009) // 0.01-0.1% spread
	ticker.BidPrice = newPrice - spread/2
	ticker.AskPrice = newPrice + spread/2
	ticker.BidSize = rand.Float64() * 1000 // Bid size
	ticker.AskSize = rand.Float64() * 1000 // Ask size

	return ticker
}

// GenerateCandle generates realistic candle for stock
func (g *StockDataGenerator) GenerateCandle(symbol string, market entities.MarketType, timestamp time.Time) *entities.Candle {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = 50.0 + rand.Float64()*200.0
		g.currentPrices[symbol] = basePrice
	}

	// Generate OHLC data (lower volatility for stocks)
	open := basePrice
	volatility := 0.01 // 1% volatility

	high := open * (1 + rand.Float64()*volatility)
	low := open * (1 - rand.Float64()*volatility)
	close := low + rand.Float64()*(high-low)

	// Update current price
	g.currentPrices[symbol] = close

	volume := rand.Float64() * 10000000 // Up to 10M shares

	candle := &entities.Candle{
		Symbol:    symbol,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
		Market:    market,
		Type:      entities.InstrumentTypeStock,
		Timestamp: timestamp,
		Timeframe: "1m",
	}

	// Add additional fields
	candle.Trades = int64(rand.Intn(10000) + 1000) // Number of trades
	candle.QuoteVolume = volume * close            // Volume in dollars

	return candle
}

// GenerateOrderBook generates realistic order book for stock
func (g *StockDataGenerator) GenerateOrderBook(symbol string, market entities.MarketType, timestamp time.Time) *entities.OrderBook {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = 50.0 + rand.Float64()*200.0
	}

	// Generate bid and ask levels (fewer levels than crypto)
	bids := make([]entities.PriceLevel, 0, 10)
	asks := make([]entities.PriceLevel, 0, 10)

	// Generate bid levels (below current price)
	for i := 0; i < 10; i++ {
		price := basePrice * (1 - float64(i+1)*0.0005) // Each level 0.05% lower
		quantity := rand.Float64() * 1000              // Up to 1000 shares
		bids = append(bids, entities.PriceLevel{
			Price:    price,
			Quantity: quantity,
		})
	}

	// Generate ask levels (above current price)
	for i := 0; i < 10; i++ {
		price := basePrice * (1 + float64(i+1)*0.0005) // Each level 0.05% higher
		quantity := rand.Float64() * 1000              // Up to 1000 shares
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
		Type:      entities.InstrumentTypeStock,
		Timestamp: timestamp,
	}

	// Add additional fields
	orderBook.LastUpdateID = rand.Int63n(1000000)
	orderBook.EventTime = timestamp

	return orderBook
}

// GenerateDividend generates dividend
func (g *StockDataGenerator) GenerateDividend(symbol string, timestamp time.Time) interfaces.Dividend {
	// Generate realistic dividend amount
	amount := rand.Float64() * 2.0 // Up to $2 per share

	// Dividend dates
	exDate := timestamp.AddDate(0, 0, rand.Intn(30)+1)     // Ex-date in 1-30 days
	payDate := exDate.AddDate(0, 0, rand.Intn(30)+15)      // Pay date 15-45 days after ex-date
	recordDate := exDate.AddDate(0, 0, 2)                  // Record date 2 days after ex-date
	declareDate := timestamp.AddDate(0, 0, -rand.Intn(30)) // Declare date up to 30 days ago

	// Payment frequency
	frequencies := []string{"QUARTERLY", "MONTHLY", "ANNUALLY", "SEMI_ANNUALLY"}
	frequency := frequencies[rand.Intn(len(frequencies))]

	// Dividend type
	types := []string{"CASH", "STOCK", "SPECIAL"}
	divType := types[rand.Intn(len(types))]

	return interfaces.Dividend{
		Symbol:      symbol,
		Amount:      amount,
		Currency:    "USD",
		ExDate:      exDate,
		PayDate:     payDate,
		RecordDate:  recordDate,
		DeclareDate: declareDate,
		Frequency:   frequency,
		Type:        divType,
	}
}

// GenerateCorporateAction generates corporate action
func (g *StockDataGenerator) GenerateCorporateAction(symbol string, timestamp time.Time) interfaces.CorporateAction {
	// Corporate action types
	actionTypes := []string{"SPLIT", "MERGER", "SPINOFF", "DIVIDEND", "RIGHTS_OFFERING"}
	actionType := actionTypes[rand.Intn(len(actionTypes))]

	// Generate ratio depending on type
	var ratio string
	var description string

	switch actionType {
	case "SPLIT":
		ratios := []string{"2:1", "3:1", "3:2", "4:1"}
		ratio = ratios[rand.Intn(len(ratios))]
		description = fmt.Sprintf("Stock split %s", ratio)
	case "MERGER":
		ratio = "1:1"
		description = "Merger with another company"
	case "SPINOFF":
		ratio = "1:1"
		description = "Spinoff of subsidiary"
	case "DIVIDEND":
		ratio = "1:1"
		description = "Special dividend distribution"
	case "RIGHTS_OFFERING":
		ratio = "1:5"
		description = "Rights offering to existing shareholders"
	}

	// Dates
	exDate := timestamp.AddDate(0, 0, rand.Intn(60)+1) // Ex-date in 1-60 days
	payDate := exDate.AddDate(0, 0, rand.Intn(30)+1)   // Pay date 1-30 days after ex-date
	recordDate := exDate.AddDate(0, 0, 2)              // Record date 2 days after ex-date

	return interfaces.CorporateAction{
		Symbol:      symbol,
		Type:        actionType,
		Ratio:       ratio,
		ExDate:      exDate,
		PayDate:     payDate,
		RecordDate:  recordDate,
		Description: description,
	}
}

// GenerateEarnings generates earnings report
func (g *StockDataGenerator) GenerateEarnings(symbol string, timestamp time.Time) interfaces.Earnings {
	// Generate realistic earnings data
	eps := rand.Float64() * 5.0                     // EPS from 0 to $5
	epsEstimate := eps * (0.9 + rand.Float64()*0.2) // Estimate ±10%

	revenue := (rand.Float64()*50 + 10) * 1000000000         // Revenue from $10B to $60B
	revenueEstimate := revenue * (0.95 + rand.Float64()*0.1) // Estimate ±5%

	// Report period
	periods := []string{"Q1", "Q2", "Q3", "Q4", "FY"}
	period := periods[rand.Intn(len(periods))]

	// Report date
	reportDate := timestamp.AddDate(0, 0, rand.Intn(30)+1)

	// Publication time
	times := []string{"BMO", "AMC"} // Before Market Open, After Market Close
	reportTime := times[rand.Intn(len(times))]

	return interfaces.Earnings{
		Symbol:          symbol,
		Period:          period,
		EPS:             eps,
		EPSEstimate:     epsEstimate,
		Revenue:         revenue,
		RevenueEstimate: revenueEstimate,
		ReportDate:      reportDate,
		Time:            reportTime,
	}
}

// GetCurrentPrice returns current price of symbol
func (g *StockDataGenerator) GetCurrentPrice(symbol string) float64 {
	if price, exists := g.currentPrices[symbol]; exists {
		return price
	}
	return 0
}

// SetCurrentPrice sets current price of symbol
func (g *StockDataGenerator) SetCurrentPrice(symbol string, price float64) {
	g.currentPrices[symbol] = price
}

// GenerateRealisticPriceMovement generates realistic price movement for stock
func (g *StockDataGenerator) GenerateRealisticPriceMovement(symbol string, volatility float64) float64 {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = g.basePrices[symbol]
		if basePrice == 0 {
			basePrice = 50.0 + rand.Float64()*200.0
		}
	}

	// Use random walk model with mean reversion
	// Stocks have lower volatility and stronger mean reversion
	meanReversion := 0.02 // Stronger mean reversion for stocks
	baseValue := g.basePrices[symbol]

	// Random component (smaller for stocks)
	randomComponent := (rand.Float64() - 0.5) * volatility * basePrice * 0.5

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
