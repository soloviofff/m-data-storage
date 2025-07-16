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

	// Тестируем получение спот рынков
	spotMarkets, err := broker.GetSpotMarkets()
	require.NoError(t, err)
	assert.NotEmpty(t, spotMarkets)

	// Проверяем, что есть популярные пары
	symbols := make(map[string]bool)
	for _, market := range spotMarkets {
		symbols[market.Symbol] = true
	}
	assert.True(t, symbols["BTCUSDT"])
	assert.True(t, symbols["ETHUSDT"])

	// Тестируем получение фьючерс рынков
	futuresMarkets, err := broker.GetFuturesMarkets()
	require.NoError(t, err)
	assert.NotEmpty(t, futuresMarkets)

	// Проверяем, что есть популярные фьючерсы
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

	// Тестируем получение информации о контракте
	contractInfo, err := broker.GetContractInfo("BTCUSDT")
	require.NoError(t, err)
	assert.NotNil(t, contractInfo)
	assert.Equal(t, "BTCUSDT", contractInfo.Symbol)
	assert.Greater(t, contractInfo.MaintMarginPercent, 0.0)
	assert.Greater(t, contractInfo.RequiredMarginPercent, 0.0)

	// Тестируем несуществующий символ
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

	// Тестируем подписку на спот рынки
	err := broker.SubscribeSpot(ctx, []string{"BTCUSDT", "ETHUSDT"})
	assert.NoError(t, err)

	// Тестируем подписку на фьючерс рынки
	err = broker.SubscribeFutures(ctx, []string{"BTCUSDT", "ETHUSDT"})
	assert.NoError(t, err)

	// Проверяем, что подписки созданы
	subscriptions := broker.GetSubscriptions()
	assert.Len(t, subscriptions, 4) // 2 спот + 2 фьючерс
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

	// Проверяем, что каналы доступны
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

	// Запускаем брокер
	err := broker.Start(ctx)
	require.NoError(t, err)

	// Подписываемся на данные
	err = broker.SubscribeSpot(ctx, []string{"BTCUSDT"})
	require.NoError(t, err)

	err = broker.SubscribeFutures(ctx, []string{"ETHUSDT"})
	require.NoError(t, err)

	// Ждем генерации данных
	time.Sleep(500 * time.Millisecond)

	// Проверяем, что данные генерируются
	tickerChan := broker.GetTickerChannel()
	candleChan := broker.GetCandleChannel()
	orderBookChan := broker.GetOrderBookChannel()
	fundingRateChan := broker.GetFundingRateChannel()
	markPriceChan := broker.GetMarkPriceChannel()
	liquidationChan := broker.GetLiquidationChannel()

	// Должны получить хотя бы некоторые данные
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
	// Проверяем, что получили данные
	assert.Greater(t, receivedTickers, 0, "Should receive ticker data")

	// Останавливаем брокер
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

	// Проверяем, что есть и спот и фьючерс инструменты
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
