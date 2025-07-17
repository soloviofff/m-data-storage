package broker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

func TestCryptoDataGenerator_GenerateTicker(t *testing.T) {
	spotMarkets := []interfaces.SpotMarket{
		{Symbol: "BTCUSDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	}
	futuresMarkets := []interfaces.FuturesMarket{
		{Symbol: "ETHUSDT", BaseAsset: "ETH", QuoteAsset: "USDT"},
	}

	generator := NewCryptoDataGenerator(spotMarkets, futuresMarkets)
	now := time.Now()

	// Test ticker generation for spot
	ticker := generator.GenerateTicker("BTCUSDT", entities.MarketTypeSpot, now)
	assert.NotNil(t, ticker)
	assert.Equal(t, "BTCUSDT", ticker.Symbol)
	assert.Equal(t, entities.MarketTypeSpot, ticker.Market)
	assert.Equal(t, entities.InstrumentTypeSpot, ticker.Type)
	assert.Greater(t, ticker.Price, 0.0)
	assert.Greater(t, ticker.Volume, 0.0)
	assert.NotZero(t, ticker.BidPrice)
	assert.NotZero(t, ticker.AskPrice)
	assert.Greater(t, ticker.AskPrice, ticker.BidPrice) // Ask should be higher than Bid

	// Test ticker generation for futures
	futuresTicker := generator.GenerateTicker("ETHUSDT", entities.MarketTypeFutures, now)
	assert.NotNil(t, futuresTicker)
	assert.Equal(t, "ETHUSDT", futuresTicker.Symbol)
	assert.Equal(t, entities.MarketTypeFutures, futuresTicker.Market)
	assert.Equal(t, entities.InstrumentTypeFutures, futuresTicker.Type)
	assert.Greater(t, futuresTicker.OpenInterest, 0.0)
}

func TestCryptoDataGenerator_GenerateCandle(t *testing.T) {
	spotMarkets := []interfaces.SpotMarket{
		{Symbol: "BTCUSDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	}
	futuresMarkets := []interfaces.FuturesMarket{}

	generator := NewCryptoDataGenerator(spotMarkets, futuresMarkets)
	now := time.Now()

	candle := generator.GenerateCandle("BTCUSDT", entities.MarketTypeSpot, now)
	assert.NotNil(t, candle)
	assert.Equal(t, "BTCUSDT", candle.Symbol)
	assert.Equal(t, entities.MarketTypeSpot, candle.Market)
	assert.Equal(t, entities.InstrumentTypeSpot, candle.Type)
	assert.Greater(t, candle.Open, 0.0)
	assert.Greater(t, candle.High, 0.0)
	assert.Greater(t, candle.Low, 0.0)
	assert.Greater(t, candle.Close, 0.0)
	assert.Greater(t, candle.Volume, 0.0)
	assert.Equal(t, "1m", candle.Timeframe)

	// Check OHLC logic
	assert.GreaterOrEqual(t, candle.High, candle.Open)
	assert.GreaterOrEqual(t, candle.High, candle.Close)
	assert.LessOrEqual(t, candle.Low, candle.Open)
	assert.LessOrEqual(t, candle.Low, candle.Close)
}

func TestCryptoDataGenerator_GenerateOrderBook(t *testing.T) {
	spotMarkets := []interfaces.SpotMarket{
		{Symbol: "BTCUSDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	}
	futuresMarkets := []interfaces.FuturesMarket{}

	generator := NewCryptoDataGenerator(spotMarkets, futuresMarkets)
	now := time.Now()

	orderBook := generator.GenerateOrderBook("BTCUSDT", entities.MarketTypeSpot, now)
	assert.NotNil(t, orderBook)
	assert.Equal(t, "BTCUSDT", orderBook.Symbol)
	assert.Equal(t, entities.MarketTypeSpot, orderBook.Market)
	assert.Equal(t, entities.InstrumentTypeSpot, orderBook.Type)
	assert.Len(t, orderBook.Bids, 20)
	assert.Len(t, orderBook.Asks, 20)

	// Check that bid prices are decreasing
	for i := 1; i < len(orderBook.Bids); i++ {
		assert.Greater(t, orderBook.Bids[i-1].Price, orderBook.Bids[i].Price)
	}

	// Check that ask prices are increasing
	for i := 1; i < len(orderBook.Asks); i++ {
		assert.Greater(t, orderBook.Asks[i].Price, orderBook.Asks[i-1].Price)
	}

	// Check that best ask is higher than best bid
	if len(orderBook.Bids) > 0 && len(orderBook.Asks) > 0 {
		assert.Greater(t, orderBook.Asks[0].Price, orderBook.Bids[0].Price)
	}
}

func TestCryptoDataGenerator_GenerateFundingRate(t *testing.T) {
	spotMarkets := []interfaces.SpotMarket{}
	futuresMarkets := []interfaces.FuturesMarket{
		{Symbol: "BTCUSDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	}

	generator := NewCryptoDataGenerator(spotMarkets, futuresMarkets)
	now := time.Now()

	fundingRate := generator.GenerateFundingRate("BTCUSDT", now)
	assert.Equal(t, "BTCUSDT", fundingRate.Symbol)
	assert.True(t, fundingRate.Rate >= -0.001 && fundingRate.Rate <= 0.001) // ±0.1%
	assert.True(t, fundingRate.NextTime.After(now))
}

func TestCryptoDataGenerator_GenerateMarkPrice(t *testing.T) {
	spotMarkets := []interfaces.SpotMarket{}
	futuresMarkets := []interfaces.FuturesMarket{
		{Symbol: "BTCUSDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	}

	generator := NewCryptoDataGenerator(spotMarkets, futuresMarkets)
	now := time.Now()

	markPrice := generator.GenerateMarkPrice("BTCUSDT", now)
	assert.Equal(t, "BTCUSDT", markPrice.Symbol)
	assert.Greater(t, markPrice.Price, 0.0)
	assert.Greater(t, markPrice.IndexPrice, 0.0)
}

func TestCryptoDataGenerator_GenerateLiquidation(t *testing.T) {
	spotMarkets := []interfaces.SpotMarket{}
	futuresMarkets := []interfaces.FuturesMarket{
		{Symbol: "BTCUSDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	}

	generator := NewCryptoDataGenerator(spotMarkets, futuresMarkets)
	now := time.Now()

	liquidation := generator.GenerateLiquidation("BTCUSDT", now)
	assert.Equal(t, "BTCUSDT", liquidation.Symbol)
	assert.True(t, liquidation.Side == "BUY" || liquidation.Side == "SELL")
	assert.Greater(t, liquidation.Price, 0.0)
	assert.Greater(t, liquidation.Quantity, 0.0)
}

func TestStockDataGenerator_GenerateTicker(t *testing.T) {
	stockMarkets := []interfaces.StockMarket{
		{Symbol: "AAPL", CompanyName: "Apple Inc.", Exchange: "NASDAQ"},
	}
	sectors := []interfaces.Sector{
		{Code: "TECH", Name: "Technology"},
	}

	generator := NewStockDataGenerator(stockMarkets, sectors)
	now := time.Now()

	ticker := generator.GenerateTicker("AAPL", entities.MarketTypeStock, now)
	assert.NotNil(t, ticker)
	assert.Equal(t, "AAPL", ticker.Symbol)
	assert.Equal(t, entities.MarketTypeStock, ticker.Market)
	assert.Equal(t, entities.InstrumentTypeStock, ticker.Type)
	assert.Greater(t, ticker.Price, 0.0)
	assert.Greater(t, ticker.Volume, 0.0)
	assert.Greater(t, ticker.BidPrice, 0.0)
	assert.Greater(t, ticker.AskPrice, 0.0)
	assert.Greater(t, ticker.AskPrice, ticker.BidPrice)

	// Check that spread is reasonable (less than 1%)
	spread := (ticker.AskPrice - ticker.BidPrice) / ticker.Price
	assert.Less(t, spread, 0.01)
}

func TestStockDataGenerator_GenerateCandle(t *testing.T) {
	stockMarkets := []interfaces.StockMarket{
		{Symbol: "AAPL", CompanyName: "Apple Inc.", Exchange: "NASDAQ"},
	}
	sectors := []interfaces.Sector{}

	generator := NewStockDataGenerator(stockMarkets, sectors)
	now := time.Now()

	candle := generator.GenerateCandle("AAPL", entities.MarketTypeStock, now)
	assert.NotNil(t, candle)
	assert.Equal(t, "AAPL", candle.Symbol)
	assert.Equal(t, entities.MarketTypeStock, candle.Market)
	assert.Equal(t, entities.InstrumentTypeStock, candle.Type)
	assert.Greater(t, candle.Open, 0.0)
	assert.Greater(t, candle.High, 0.0)
	assert.Greater(t, candle.Low, 0.0)
	assert.Greater(t, candle.Close, 0.0)
	assert.Greater(t, candle.Volume, 0.0)

	// Check OHLC logic
	assert.GreaterOrEqual(t, candle.High, candle.Open)
	assert.GreaterOrEqual(t, candle.High, candle.Close)
	assert.LessOrEqual(t, candle.Low, candle.Open)
	assert.LessOrEqual(t, candle.Low, candle.Close)
}

func TestStockDataGenerator_GenerateDividend(t *testing.T) {
	stockMarkets := []interfaces.StockMarket{
		{Symbol: "AAPL", CompanyName: "Apple Inc.", Exchange: "NASDAQ"},
	}
	sectors := []interfaces.Sector{}

	generator := NewStockDataGenerator(stockMarkets, sectors)
	now := time.Now()

	dividend := generator.GenerateDividend("AAPL", now)
	assert.Equal(t, "AAPL", dividend.Symbol)
	assert.Greater(t, dividend.Amount, 0.0)
	assert.Equal(t, "USD", dividend.Currency)
	assert.True(t, dividend.ExDate.After(now))
	assert.True(t, dividend.PayDate.After(dividend.ExDate))
	assert.NotEmpty(t, dividend.Frequency)
	assert.NotEmpty(t, dividend.Type)
}

func TestStockDataGenerator_GenerateCorporateAction(t *testing.T) {
	stockMarkets := []interfaces.StockMarket{
		{Symbol: "AAPL", CompanyName: "Apple Inc.", Exchange: "NASDAQ"},
	}
	sectors := []interfaces.Sector{}

	generator := NewStockDataGenerator(stockMarkets, sectors)
	now := time.Now()

	action := generator.GenerateCorporateAction("AAPL", now)
	assert.Equal(t, "AAPL", action.Symbol)
	assert.NotEmpty(t, action.Type)
	assert.NotEmpty(t, action.Ratio)
	assert.NotEmpty(t, action.Description)
	assert.True(t, action.ExDate.After(now))
}

func TestStockDataGenerator_GenerateEarnings(t *testing.T) {
	stockMarkets := []interfaces.StockMarket{
		{Symbol: "AAPL", CompanyName: "Apple Inc.", Exchange: "NASDAQ"},
	}
	sectors := []interfaces.Sector{}

	generator := NewStockDataGenerator(stockMarkets, sectors)
	now := time.Now()

	earnings := generator.GenerateEarnings("AAPL", now)
	assert.Equal(t, "AAPL", earnings.Symbol)
	assert.Greater(t, earnings.EPS, 0.0)
	assert.Greater(t, earnings.EPSEstimate, 0.0)
	assert.Greater(t, earnings.Revenue, 0.0)
	assert.Greater(t, earnings.RevenueEstimate, 0.0)
	assert.NotEmpty(t, earnings.Period)
	assert.True(t, earnings.Time == "BMO" || earnings.Time == "AMC")
	assert.True(t, earnings.ReportDate.After(now))
}

func TestDataGenerator_PriceConsistency(t *testing.T) {
	spotMarkets := []interfaces.SpotMarket{
		{Symbol: "BTCUSDT", BaseAsset: "BTC", QuoteAsset: "USDT"},
	}
	futuresMarkets := []interfaces.FuturesMarket{}

	generator := NewCryptoDataGenerator(spotMarkets, futuresMarkets)
	now := time.Now()

	// Generate several tickers in sequence
	ticker1 := generator.GenerateTicker("BTCUSDT", entities.MarketTypeSpot, now)
	ticker2 := generator.GenerateTicker("BTCUSDT", entities.MarketTypeSpot, now.Add(time.Second))
	ticker3 := generator.GenerateTicker("BTCUSDT", entities.MarketTypeSpot, now.Add(2*time.Second))

	// Check that prices change gradually (no more than 5% at a time)
	change1 := abs(ticker2.Price-ticker1.Price) / ticker1.Price
	change2 := abs(ticker3.Price-ticker2.Price) / ticker2.Price

	assert.Less(t, change1, 0.05, "Price change should be less than 5%")
	assert.Less(t, change2, 0.05, "Price change should be less than 5%")
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
