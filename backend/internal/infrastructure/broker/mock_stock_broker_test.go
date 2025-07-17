package broker

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

func TestMockStockBroker_Creation(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-stock",
		Name:    "mock",
		Type:    interfaces.BrokerTypeStock,
		Enabled: true,
		Connection: interfaces.ConnectionConfig{
			WebSocketURL: "ws://localhost:8080",
			RestAPIURL:   "http://localhost:8080",
		},
		Limits: interfaces.LimitsConfig{
			MaxSubscriptions:  100,
			RequestsPerSecond: 1000,
		},
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockStockBroker(config, logger)

	assert.NotNil(t, broker)

	info := broker.GetInfo()
	assert.Equal(t, config.ID, info.ID)
	assert.Equal(t, config.Name, info.Name)
	assert.Equal(t, config.Type, info.Type)
}

func TestMockStockBroker_GetMarkets(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-stock",
		Name:    "mock",
		Type:    interfaces.BrokerTypeStock,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockStockBroker(config, logger)

	// Test stock markets retrieval
	stockMarkets, err := broker.GetStockMarkets()
	require.NoError(t, err)
	assert.NotEmpty(t, stockMarkets)

	// Check that popular stocks exist
	symbols := make(map[string]bool)
	for _, market := range stockMarkets {
		symbols[market.Symbol] = true
	}
	assert.True(t, symbols["AAPL"])
	assert.True(t, symbols["MSFT"])
	assert.True(t, symbols["GOOGL"])

	// Test sectors retrieval
	sectors, err := broker.GetSectors()
	require.NoError(t, err)
	assert.NotEmpty(t, sectors)

	// Check that main sectors exist
	sectorCodes := make(map[string]bool)
	for _, sector := range sectors {
		sectorCodes[sector.Code] = true
	}
	assert.True(t, sectorCodes["TECH"])
	assert.True(t, sectorCodes["FINL"])
	assert.True(t, sectorCodes["HLTH"])
}

func TestMockStockBroker_GetCompanyInfo(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-stock",
		Name:    "mock",
		Type:    interfaces.BrokerTypeStock,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockStockBroker(config, logger)

	// Test company information retrieval
	companyInfo, err := broker.GetCompanyInfo("AAPL")
	require.NoError(t, err)
	assert.NotNil(t, companyInfo)
	assert.Equal(t, "AAPL", companyInfo.Symbol)
	assert.Equal(t, "Apple Inc.", companyInfo.Name)
	assert.Equal(t, "NASDAQ", companyInfo.Exchange)
	assert.Greater(t, companyInfo.MarketCap, 0.0)

	// Test non-existent symbol
	_, err = broker.GetCompanyInfo("NONEXISTENT")
	assert.Error(t, err)
}

func TestMockStockBroker_Subscriptions(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-stock",
		Name:    "mock",
		Type:    interfaces.BrokerTypeStock,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockStockBroker(config, logger)

	ctx := context.Background()

	// Test stocks subscription
	err := broker.SubscribeStocks(ctx, []string{"AAPL", "MSFT", "GOOGL"})
	assert.NoError(t, err)

	// Check that subscriptions are created
	subscriptions := broker.GetSubscriptions()
	assert.Len(t, subscriptions, 3)
}

func TestMockStockBroker_DataChannels(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-stock",
		Name:    "mock",
		Type:    interfaces.BrokerTypeStock,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockStockBroker(config, logger)

	// Check that channels are available
	tickerChan := broker.GetTickerChannel()
	assert.NotNil(t, tickerChan)

	candleChan := broker.GetCandleChannel()
	assert.NotNil(t, candleChan)

	orderBookChan := broker.GetOrderBookChannel()
	assert.NotNil(t, orderBookChan)

	dividendChan := broker.GetDividendChannel()
	assert.NotNil(t, dividendChan)

	corporateActionChan := broker.GetCorporateActionChannel()
	assert.NotNil(t, corporateActionChan)

	earningsChan := broker.GetEarningsChannel()
	assert.NotNil(t, earningsChan)
}

func TestMockStockBroker_DataGeneration(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-stock",
		Name:    "mock",
		Type:    interfaces.BrokerTypeStock,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockStockBroker(config, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Start broker
	err := broker.Start(ctx)
	require.NoError(t, err)

	// Subscribe to data
	err = broker.SubscribeStocks(ctx, []string{"AAPL", "MSFT"})
	require.NoError(t, err)

	// Wait for data generation
	time.Sleep(1 * time.Second)

	// Check that data is being generated
	tickerChan := broker.GetTickerChannel()
	candleChan := broker.GetCandleChannel()
	orderBookChan := broker.GetOrderBookChannel()
	dividendChan := broker.GetDividendChannel()
	corporateActionChan := broker.GetCorporateActionChannel()
	earningsChan := broker.GetEarningsChannel()

	// Should receive at least some data
	receivedTickers := 0
	receivedCandles := 0
	receivedOrderBooks := 0
	receivedDividends := 0
	receivedCorporateActions := 0
	receivedEarnings := 0

	timeout := time.After(1 * time.Second)
	for {
		select {
		case ticker := <-tickerChan:
			assert.NotEmpty(t, ticker.Symbol)
			assert.Greater(t, ticker.Price, 0.0)
			assert.Equal(t, config.ID, ticker.BrokerID)
			assert.Greater(t, ticker.BidPrice, 0.0)
			assert.Greater(t, ticker.AskPrice, 0.0)
			receivedTickers++
		case candle := <-candleChan:
			assert.NotEmpty(t, candle.Symbol)
			assert.Greater(t, candle.Close, 0.0)
			assert.Equal(t, config.ID, candle.BrokerID)
			receivedCandles++
		case orderBook := <-orderBookChan:
			assert.NotEmpty(t, orderBook.Symbol)
			assert.NotEmpty(t, orderBook.Bids)
			assert.NotEmpty(t, orderBook.Asks)
			assert.Equal(t, config.ID, orderBook.BrokerID)
			receivedOrderBooks++
		case dividend := <-dividendChan:
			assert.NotEmpty(t, dividend.Symbol)
			assert.Greater(t, dividend.Amount, 0.0)
			assert.Equal(t, config.ID, dividend.BrokerID)
			receivedDividends++
		case action := <-corporateActionChan:
			assert.NotEmpty(t, action.Symbol)
			assert.NotEmpty(t, action.Type)
			assert.Equal(t, config.ID, action.BrokerID)
			receivedCorporateActions++
		case earnings := <-earningsChan:
			assert.NotEmpty(t, earnings.Symbol)
			assert.Greater(t, earnings.EPS, 0.0)
			assert.Equal(t, config.ID, earnings.BrokerID)
			receivedEarnings++
		case <-timeout:
			goto checkResults
		}
	}

checkResults:
	// Check that we received data
	assert.Greater(t, receivedTickers, 0, "Should receive ticker data")

	// Stop broker
	err = broker.Stop()
	assert.NoError(t, err)
}

func TestMockStockBroker_TradingHours(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-stock",
		Name:    "mock",
		Type:    interfaces.BrokerTypeStock,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockStockBroker(config, logger)

	// Test trading hours check
	// Monday 10:00 - trading hours
	tradingTime := time.Date(2024, 1, 8, 10, 0, 0, 0, time.UTC) // Monday
	assert.True(t, broker.isTradingHours(tradingTime))

	// Saturday - not trading hours
	weekendTime := time.Date(2024, 1, 6, 10, 0, 0, 0, time.UTC) // Saturday
	assert.False(t, broker.isTradingHours(weekendTime))

	// Monday 8:00 - before market open
	earlyTime := time.Date(2024, 1, 8, 8, 0, 0, 0, time.UTC)
	assert.False(t, broker.isTradingHours(earlyTime))

	// Monday 17:00 - after market close
	lateTime := time.Date(2024, 1, 8, 17, 0, 0, 0, time.UTC)
	assert.False(t, broker.isTradingHours(lateTime))
}

func TestMockStockBroker_SupportedInstruments(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-stock",
		Name:    "mock",
		Type:    interfaces.BrokerTypeStock,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockStockBroker(config, logger)

	instruments := broker.GetSupportedInstruments()
	assert.NotEmpty(t, instruments)

	// Check that all instruments are stocks
	for _, instrument := range instruments {
		assert.Equal(t, entities.MarketTypeStock, instrument.Market)
		assert.Equal(t, entities.InstrumentTypeStock, instrument.Type)
		assert.True(t, instrument.IsActive)
	}
}
