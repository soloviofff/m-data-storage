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

func TestMockCryptoBroker_Creation(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-crypto",
		Name:    "mock",
		Type:    interfaces.BrokerTypeCrypto,
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
	broker := NewMockCryptoBroker(config, logger)

	assert.NotNil(t, broker)

	info := broker.GetInfo()
	assert.Equal(t, config.ID, info.ID)
	assert.Equal(t, config.Name, info.Name)
	assert.Equal(t, config.Type, info.Type)
}

func TestMockCryptoBroker_GetMarkets(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-crypto",
		Name:    "mock",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockCryptoBroker(config, logger)

	// Test spot markets retrieval
	spotMarkets, err := broker.GetSpotMarkets()
	require.NoError(t, err)
	assert.NotEmpty(t, spotMarkets)

	// Check that popular pairs exist
	symbols := make(map[string]bool)
	for _, market := range spotMarkets {
		symbols[market.Symbol] = true
	}
	assert.True(t, symbols["BTCUSDT"])
	assert.True(t, symbols["ETHUSDT"])

	// Test futures markets retrieval
	futuresMarkets, err := broker.GetFuturesMarkets()
	require.NoError(t, err)
	assert.NotEmpty(t, futuresMarkets)

	// Check that popular futures exist
	futuresSymbols := make(map[string]bool)
	for _, market := range futuresMarkets {
		futuresSymbols[market.Symbol] = true
	}
	assert.True(t, futuresSymbols["BTCUSDT"])
	assert.True(t, futuresSymbols["ETHUSDT"])
}

func TestMockCryptoBroker_GetContractInfo(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-crypto",
		Name:    "mock",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockCryptoBroker(config, logger)

	// Test contract information retrieval
	contractInfo, err := broker.GetContractInfo("BTCUSDT")
	require.NoError(t, err)
	assert.NotNil(t, contractInfo)
	assert.Equal(t, "BTCUSDT", contractInfo.Symbol)
	assert.Greater(t, contractInfo.MaintMarginPercent, 0.0)
	assert.Greater(t, contractInfo.RequiredMarginPercent, 0.0)

	// Test non-existent symbol
	_, err = broker.GetContractInfo("NONEXISTENT")
	assert.Error(t, err)
}

func TestMockCryptoBroker_Subscriptions(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-crypto",
		Name:    "mock",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockCryptoBroker(config, logger)

	ctx := context.Background()

	// Test spot markets subscription
	err := broker.SubscribeSpot(ctx, []string{"BTCUSDT", "ETHUSDT"})
	assert.NoError(t, err)

	// Test futures markets subscription
	err = broker.SubscribeFutures(ctx, []string{"BTCUSDT", "ETHUSDT"})
	assert.NoError(t, err)

	// Check that subscriptions are created
	subscriptions := broker.GetSubscriptions()
	assert.Len(t, subscriptions, 4) // 2 spot + 2 futures
}

func TestMockCryptoBroker_DataChannels(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-crypto",
		Name:    "mock",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockCryptoBroker(config, logger)

	// Check that channels are available
	tickerChan := broker.GetTickerChannel()
	assert.NotNil(t, tickerChan)

	candleChan := broker.GetCandleChannel()
	assert.NotNil(t, candleChan)

	orderBookChan := broker.GetOrderBookChannel()
	assert.NotNil(t, orderBookChan)

	fundingRateChan := broker.GetFundingRateChannel()
	assert.NotNil(t, fundingRateChan)

	markPriceChan := broker.GetMarkPriceChannel()
	assert.NotNil(t, markPriceChan)

	liquidationChan := broker.GetLiquidationChannel()
	assert.NotNil(t, liquidationChan)
}

func TestMockCryptoBroker_DataGeneration(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-crypto",
		Name:    "mock",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockCryptoBroker(config, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start broker
	err := broker.Start(ctx)
	require.NoError(t, err)

	// Subscribe to data
	err = broker.SubscribeSpot(ctx, []string{"BTCUSDT"})
	require.NoError(t, err)

	err = broker.SubscribeFutures(ctx, []string{"ETHUSDT"})
	require.NoError(t, err)

	// Wait for data generation
	time.Sleep(500 * time.Millisecond)

	// Check that data is being generated
	tickerChan := broker.GetTickerChannel()
	candleChan := broker.GetCandleChannel()
	orderBookChan := broker.GetOrderBookChannel()
	fundingRateChan := broker.GetFundingRateChannel()
	markPriceChan := broker.GetMarkPriceChannel()
	liquidationChan := broker.GetLiquidationChannel()

	// Should receive at least some data
	receivedTickers := 0
	receivedCandles := 0
	receivedOrderBooks := 0
	receivedFundingRates := 0
	receivedMarkPrices := 0
	receivedLiquidations := 0

	timeout := time.After(1 * time.Second)
	for {
		select {
		case ticker := <-tickerChan:
			assert.NotEmpty(t, ticker.Symbol)
			assert.Greater(t, ticker.Price, 0.0)
			assert.Equal(t, config.ID, ticker.BrokerID)
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
		case fundingRate := <-fundingRateChan:
			assert.NotEmpty(t, fundingRate.Symbol)
			assert.Equal(t, config.ID, fundingRate.BrokerID)
			receivedFundingRates++
		case markPrice := <-markPriceChan:
			assert.NotEmpty(t, markPrice.Symbol)
			assert.Greater(t, markPrice.Price, 0.0)
			assert.Equal(t, config.ID, markPrice.BrokerID)
			receivedMarkPrices++
		case liquidation := <-liquidationChan:
			assert.NotEmpty(t, liquidation.Symbol)
			assert.Greater(t, liquidation.Price, 0.0)
			assert.Equal(t, config.ID, liquidation.BrokerID)
			receivedLiquidations++
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

func TestMockCryptoBroker_SupportedInstruments(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:      "test-crypto",
		Name:    "mock",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: true,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
	}

	logger := logrus.New()
	broker := NewMockCryptoBroker(config, logger)

	instruments := broker.GetSupportedInstruments()
	assert.NotEmpty(t, instruments)

	// Check that both spot and futures instruments exist
	hasSpot := false
	hasFutures := false

	for _, instrument := range instruments {
		if instrument.Market == entities.MarketTypeSpot {
			hasSpot = true
		}
		if instrument.Market == entities.MarketTypeFutures {
			hasFutures = true
		}
	}

	assert.True(t, hasSpot, "Should have spot instruments")
	assert.True(t, hasFutures, "Should have futures instruments")
}
