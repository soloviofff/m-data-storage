package broker

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// MockCryptoBroker implements CryptoBroker interface for testing
type MockCryptoBroker struct {
	*BaseBroker

	// Specific channels for crypto exchanges
	fundingRateChan chan interfaces.FundingRate
	markPriceChan   chan interfaces.MarkPrice
	liquidationChan chan interfaces.Liquidation

	// Data for simulation
	spotMarkets    []interfaces.SpotMarket
	futuresMarkets []interfaces.FuturesMarket
	contractInfos  map[string]*interfaces.ContractInfo

	// Data generator
	dataGenerator *CryptoDataGenerator

	// Simulation control
	simulationRunning bool
	simulationMu      sync.RWMutex
}

// NewMockCryptoBroker creates a new mock crypto broker
func NewMockCryptoBroker(config interfaces.BrokerConfig, logger *logrus.Logger) *MockCryptoBroker {
	baseBroker := NewBaseBroker(config, logger)

	mock := &MockCryptoBroker{
		BaseBroker:      baseBroker,
		fundingRateChan: make(chan interfaces.FundingRate, config.Defaults.BufferSize),
		markPriceChan:   make(chan interfaces.MarkPrice, config.Defaults.BufferSize),
		liquidationChan: make(chan interfaces.Liquidation, config.Defaults.BufferSize),
		contractInfos:   make(map[string]*interfaces.ContractInfo),
	}

	// Initialize test data
	mock.initializeTestData()

	// Create data generator
	mock.dataGenerator = NewCryptoDataGenerator(mock.spotMarkets, mock.futuresMarkets)

	return mock
}

// initializeTestData initializes test markets and contracts
func (m *MockCryptoBroker) initializeTestData() {
	// Spot markets
	m.spotMarkets = []interfaces.SpotMarket{
		{
			Symbol:      "BTCUSDT",
			BaseAsset:   "BTC",
			QuoteAsset:  "USDT",
			Status:      "TRADING",
			MinPrice:    0.01,
			MaxPrice:    1000000,
			TickSize:    0.01,
			MinQuantity: 0.00001,
			MaxQuantity: 9000,
			StepSize:    0.00001,
			MinNotional: 10,
		},
		{
			Symbol:      "ETHUSDT",
			BaseAsset:   "ETH",
			QuoteAsset:  "USDT",
			Status:      "TRADING",
			MinPrice:    0.01,
			MaxPrice:    100000,
			TickSize:    0.01,
			MinQuantity: 0.0001,
			MaxQuantity: 10000,
			StepSize:    0.0001,
			MinNotional: 10,
		},
		{
			Symbol:      "ADAUSDT",
			BaseAsset:   "ADA",
			QuoteAsset:  "USDT",
			Status:      "TRADING",
			MinPrice:    0.0001,
			MaxPrice:    1000,
			TickSize:    0.0001,
			MinQuantity: 0.1,
			MaxQuantity: 90000000,
			StepSize:    0.1,
			MinNotional: 10,
		},
	}

	// Futures markets
	m.futuresMarkets = []interfaces.FuturesMarket{
		{
			Symbol:                "BTCUSDT",
			BaseAsset:             "BTC",
			QuoteAsset:            "USDT",
			ContractType:          "PERPETUAL",
			Status:                "TRADING",
			MaintMarginPercent:    2.5,
			RequiredMarginPercent: 5.0,
			BaseAssetPrecision:    8,
			QuotePrecision:        8,
			UnderlyingType:        "COIN",
			SettlePlan:            0,
		},
		{
			Symbol:                "ETHUSDT",
			BaseAsset:             "ETH",
			QuoteAsset:            "USDT",
			ContractType:          "PERPETUAL",
			Status:                "TRADING",
			MaintMarginPercent:    2.5,
			RequiredMarginPercent: 5.0,
			BaseAssetPrecision:    8,
			QuotePrecision:        8,
			UnderlyingType:        "COIN",
			SettlePlan:            0,
		},
	}

	// Contract information
	for _, market := range m.futuresMarkets {
		m.contractInfos[market.Symbol] = &interfaces.ContractInfo{
			Symbol:                market.Symbol,
			Status:                market.Status,
			MaintMarginPercent:    market.MaintMarginPercent,
			RequiredMarginPercent: market.RequiredMarginPercent,
			BaseAssetPrecision:    market.BaseAssetPrecision,
			QuotePrecision:        market.QuotePrecision,
			TriggerProtect:        0.1,
			UnderlyingType:        market.UnderlyingType,
			SettlePlan:            market.SettlePlan,
		}
	}
}

// GetSpotMarkets returns list of spot markets
func (m *MockCryptoBroker) GetSpotMarkets() ([]interfaces.SpotMarket, error) {
	return m.spotMarkets, nil
}

// GetFuturesMarkets returns list of futures markets
func (m *MockCryptoBroker) GetFuturesMarkets() ([]interfaces.FuturesMarket, error) {
	return m.futuresMarkets, nil
}

// SubscribeSpot subscribes to spot markets
func (m *MockCryptoBroker) SubscribeSpot(ctx context.Context, symbols []string) error {
	subscriptions := make([]entities.InstrumentSubscription, 0, len(symbols))

	for _, symbol := range symbols {
		subscriptions = append(subscriptions, entities.InstrumentSubscription{
			Symbol: symbol,
			Type:   entities.InstrumentTypeSpot,
			Market: entities.MarketTypeSpot,
		})
	}

	return m.Subscribe(ctx, subscriptions)
}

// SubscribeFutures subscribes to futures markets
func (m *MockCryptoBroker) SubscribeFutures(ctx context.Context, symbols []string) error {
	subscriptions := make([]entities.InstrumentSubscription, 0, len(symbols))

	for _, symbol := range symbols {
		subscriptions = append(subscriptions, entities.InstrumentSubscription{
			Symbol: symbol,
			Type:   entities.InstrumentTypeFutures,
			Market: entities.MarketTypeFutures,
		})
	}

	return m.Subscribe(ctx, subscriptions)
}

// GetContractInfo returns contract information
func (m *MockCryptoBroker) GetContractInfo(symbol string) (*interfaces.ContractInfo, error) {
	if info, exists := m.contractInfos[symbol]; exists {
		return info, nil
	}
	return nil, fmt.Errorf("contract info not found for symbol: %s", symbol)
}

// GetFundingRateChannel returns funding rate channel
func (m *MockCryptoBroker) GetFundingRateChannel() <-chan interfaces.FundingRate {
	return m.fundingRateChan
}

// GetMarkPriceChannel returns mark price channel
func (m *MockCryptoBroker) GetMarkPriceChannel() <-chan interfaces.MarkPrice {
	return m.markPriceChan
}

// GetLiquidationChannel returns liquidation channel
func (m *MockCryptoBroker) GetLiquidationChannel() <-chan interfaces.Liquidation {
	return m.liquidationChan
}

// GetSupportedInstruments returns list of supported instruments
func (m *MockCryptoBroker) GetSupportedInstruments() []entities.InstrumentInfo {
	instruments := make([]entities.InstrumentInfo, 0, len(m.spotMarkets)+len(m.futuresMarkets))

	// Add spot instruments
	for _, market := range m.spotMarkets {
		instruments = append(instruments, entities.InstrumentInfo{
			Symbol:            market.Symbol,
			BaseAsset:         market.BaseAsset,
			QuoteAsset:        market.QuoteAsset,
			Type:              entities.InstrumentTypeSpot,
			Market:            entities.MarketTypeSpot,
			IsActive:          market.Status == "TRADING",
			MinPrice:          market.MinPrice,
			MaxPrice:          market.MaxPrice,
			MinQuantity:       market.MinQuantity,
			MaxQuantity:       market.MaxQuantity,
			PricePrecision:    2,
			QuantityPrecision: 8,
		})
	}

	// Add futures instruments
	for _, market := range m.futuresMarkets {
		instruments = append(instruments, entities.InstrumentInfo{
			Symbol:            market.Symbol,
			BaseAsset:         market.BaseAsset,
			QuoteAsset:        market.QuoteAsset,
			Type:              entities.InstrumentTypeFutures,
			Market:            entities.MarketTypeFutures,
			IsActive:          market.Status == "TRADING",
			PricePrecision:    market.QuotePrecision,
			QuantityPrecision: market.BaseAssetPrecision,
		})
	}

	return instruments
}

// Start starts data simulation
func (m *MockCryptoBroker) Start(ctx context.Context) error {
	if err := m.BaseBroker.Start(ctx); err != nil {
		return err
	}

	m.simulationMu.Lock()
	m.simulationRunning = true
	m.simulationMu.Unlock()

	// Start data generation
	m.wg.Add(1)
	go m.runDataSimulation()

	return nil
}

// Stop stops data simulation
func (m *MockCryptoBroker) Stop() error {
	m.simulationMu.Lock()
	m.simulationRunning = false
	m.simulationMu.Unlock()

	return m.BaseBroker.Stop()
}

// runDataSimulation starts market data simulation
func (m *MockCryptoBroker) runDataSimulation() {
	defer m.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond) // Generate data every 100ms
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.simulationMu.RLock()
			running := m.simulationRunning
			m.simulationMu.RUnlock()

			if !running {
				return
			}

			m.generateMarketData()
		}
	}
}

// generateMarketData generates random market data
func (m *MockCryptoBroker) generateMarketData() {
	now := time.Now()

	// Generate tickers for all active subscriptions
	m.subMu.RLock()
	for _, subscription := range m.subscriptions {
		// Generate ticker
		if ticker := m.dataGenerator.GenerateTicker(subscription.Symbol, subscription.Market, now); ticker != nil {
			ticker.BrokerID = m.config.ID
			m.SendTicker(*ticker)
		}

		// Sometimes generate candles
		if rand.Float32() < 0.1 { // 10% probability
			if candle := m.dataGenerator.GenerateCandle(subscription.Symbol, subscription.Market, now); candle != nil {
				candle.BrokerID = m.config.ID
				m.SendCandle(*candle)
			}
		}

		// Sometimes generate order books
		if rand.Float32() < 0.05 { // 5% probability
			if orderBook := m.dataGenerator.GenerateOrderBook(subscription.Symbol, subscription.Market, now); orderBook != nil {
				orderBook.BrokerID = m.config.ID
				m.SendOrderBook(*orderBook)
			}
		}

		// For futures generate additional data
		if subscription.Market == entities.MarketTypeFutures {
			// Funding rates
			if rand.Float32() < 0.02 { // 2% probability
				fundingRate := m.dataGenerator.GenerateFundingRate(subscription.Symbol, now)
				fundingRate.BrokerID = m.config.ID
				m.sendFundingRate(fundingRate)
			}

			// Mark prices
			if rand.Float32() < 0.05 { // 5% probability
				markPrice := m.dataGenerator.GenerateMarkPrice(subscription.Symbol, now)
				markPrice.BrokerID = m.config.ID
				m.sendMarkPrice(markPrice)
			}

			// Liquidations
			if rand.Float32() < 0.01 { // 1% probability
				liquidation := m.dataGenerator.GenerateLiquidation(subscription.Symbol, now)
				liquidation.BrokerID = m.config.ID
				m.sendLiquidation(liquidation)
			}
		}
	}
	m.subMu.RUnlock()
}

// sendFundingRate sends funding rate
func (m *MockCryptoBroker) sendFundingRate(rate interfaces.FundingRate) {
	select {
	case m.fundingRateChan <- rate:
	default:
		// Channel is full, skip
	}
}

// sendMarkPrice sends mark price
func (m *MockCryptoBroker) sendMarkPrice(price interfaces.MarkPrice) {
	select {
	case m.markPriceChan <- price:
	default:
		// Channel is full, skip
	}
}

// sendLiquidation sends liquidation
func (m *MockCryptoBroker) sendLiquidation(liquidation interfaces.Liquidation) {
	select {
	case m.liquidationChan <- liquidation:
	default:
		// Channel is full, skip
	}
}
