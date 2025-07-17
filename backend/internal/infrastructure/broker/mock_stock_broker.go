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

// MockStockBroker implements StockBroker interface for testing
type MockStockBroker struct {
	*BaseBroker

	// Specific channels for stock market
	dividendChan        chan interfaces.Dividend
	corporateActionChan chan interfaces.CorporateAction
	earningsChan        chan interfaces.Earnings

	// Data for simulation
	stockMarkets []interfaces.StockMarket
	sectors      []interfaces.Sector
	companyInfos map[string]*interfaces.CompanyInfo

	// Data generator
	dataGenerator *StockDataGenerator

	// Simulation control
	simulationRunning bool
	simulationMu      sync.RWMutex
}

// NewMockStockBroker creates a new mock stock broker
func NewMockStockBroker(config interfaces.BrokerConfig, logger *logrus.Logger) *MockStockBroker {
	baseBroker := NewBaseBroker(config, logger)

	mock := &MockStockBroker{
		BaseBroker:          baseBroker,
		dividendChan:        make(chan interfaces.Dividend, config.Defaults.BufferSize),
		corporateActionChan: make(chan interfaces.CorporateAction, config.Defaults.BufferSize),
		earningsChan:        make(chan interfaces.Earnings, config.Defaults.BufferSize),
		companyInfos:        make(map[string]*interfaces.CompanyInfo),
	}

	// Initialize test data
	mock.initializeTestData()

	// Create data generator
	mock.dataGenerator = NewStockDataGenerator(mock.stockMarkets, mock.sectors)

	return mock
}

// initializeTestData initializes test markets and companies
func (m *MockStockBroker) initializeTestData() {
	// Sectors
	m.sectors = []interfaces.Sector{
		{Code: "TECH", Name: "Technology", Description: "Technology companies"},
		{Code: "FINL", Name: "Financial", Description: "Financial services"},
		{Code: "HLTH", Name: "Healthcare", Description: "Healthcare and pharmaceuticals"},
		{Code: "ENRG", Name: "Energy", Description: "Energy and utilities"},
		{Code: "CONS", Name: "Consumer", Description: "Consumer goods and services"},
	}

	// Stock markets
	m.stockMarkets = []interfaces.StockMarket{
		{
			Symbol:            "AAPL",
			CompanyName:       "Apple Inc.",
			Exchange:          "NASDAQ",
			Currency:          "USD",
			Country:           "US",
			Sector:            "TECH",
			Industry:          "Consumer Electronics",
			MarketCap:         3000000000000, // 3T
			SharesOutstanding: 16000000000,   // 16B
			Status:            "ACTIVE",
			ListingDate:       time.Date(1980, 12, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			Symbol:            "MSFT",
			CompanyName:       "Microsoft Corporation",
			Exchange:          "NASDAQ",
			Currency:          "USD",
			Country:           "US",
			Sector:            "TECH",
			Industry:          "Software",
			MarketCap:         2800000000000, // 2.8T
			SharesOutstanding: 7400000000,    // 7.4B
			Status:            "ACTIVE",
			ListingDate:       time.Date(1986, 3, 13, 0, 0, 0, 0, time.UTC),
		},
		{
			Symbol:            "GOOGL",
			CompanyName:       "Alphabet Inc.",
			Exchange:          "NASDAQ",
			Currency:          "USD",
			Country:           "US",
			Sector:            "TECH",
			Industry:          "Internet Services",
			MarketCap:         1700000000000, // 1.7T
			SharesOutstanding: 13000000000,   // 13B
			Status:            "ACTIVE",
			ListingDate:       time.Date(2004, 8, 19, 0, 0, 0, 0, time.UTC),
		},
		{
			Symbol:            "TSLA",
			CompanyName:       "Tesla, Inc.",
			Exchange:          "NASDAQ",
			Currency:          "USD",
			Country:           "US",
			Sector:            "CONS",
			Industry:          "Electric Vehicles",
			MarketCap:         800000000000, // 800B
			SharesOutstanding: 3200000000,   // 3.2B
			Status:            "ACTIVE",
			ListingDate:       time.Date(2010, 6, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			Symbol:            "JPM",
			CompanyName:       "JPMorgan Chase & Co.",
			Exchange:          "NYSE",
			Currency:          "USD",
			Country:           "US",
			Sector:            "FINL",
			Industry:          "Banking",
			MarketCap:         450000000000, // 450B
			SharesOutstanding: 3000000000,   // 3B
			Status:            "ACTIVE",
			ListingDate:       time.Date(1969, 3, 5, 0, 0, 0, 0, time.UTC),
		},
	}

	// Company information
	for _, market := range m.stockMarkets {
		m.companyInfos[market.Symbol] = &interfaces.CompanyInfo{
			Symbol:            market.Symbol,
			Name:              market.CompanyName,
			Exchange:          market.Exchange,
			Currency:          market.Currency,
			Country:           market.Country,
			Sector:            market.Sector,
			Industry:          market.Industry,
			MarketCap:         market.MarketCap,
			SharesOutstanding: market.SharesOutstanding,
			PERatio:           15.0 + rand.Float64()*20.0, // P/E from 15 to 35
			DividendYield:     rand.Float64() * 3.0,       // Dividend yield 0-3%
			Beta:              0.5 + rand.Float64()*1.5,   // Beta from 0.5 to 2.0
			EPS:               rand.Float64() * 10.0,      // EPS from 0 to 10
			BookValue:         rand.Float64() * 50.0,      // Book value
			Description:       fmt.Sprintf("%s is a leading company in %s sector", market.CompanyName, market.Industry),
			Website:           fmt.Sprintf("https://www.%s.com", market.Symbol),
			CEO:               "John Doe", // Simplified
			Employees:         int(rand.Intn(500000) + 10000),
			Founded:           time.Date(1900+rand.Intn(100), time.Month(rand.Intn(12)+1), rand.Intn(28)+1, 0, 0, 0, 0, time.UTC),
			IPODate:           market.ListingDate,
		}
	}
}

// GetStockMarkets returns list of stock markets
func (m *MockStockBroker) GetStockMarkets() ([]interfaces.StockMarket, error) {
	return m.stockMarkets, nil
}

// GetSectors returns list of sectors
func (m *MockStockBroker) GetSectors() ([]interfaces.Sector, error) {
	return m.sectors, nil
}

// SubscribeStocks subscribes to stocks
func (m *MockStockBroker) SubscribeStocks(ctx context.Context, symbols []string) error {
	subscriptions := make([]entities.InstrumentSubscription, 0, len(symbols))

	for _, symbol := range symbols {
		subscriptions = append(subscriptions, entities.InstrumentSubscription{
			Symbol: symbol,
			Type:   entities.InstrumentTypeStock,
			Market: entities.MarketTypeStock,
		})
	}

	return m.Subscribe(ctx, subscriptions)
}

// GetCompanyInfo returns company information
func (m *MockStockBroker) GetCompanyInfo(symbol string) (*interfaces.CompanyInfo, error) {
	if info, exists := m.companyInfos[symbol]; exists {
		return info, nil
	}
	return nil, fmt.Errorf("company info not found for symbol: %s", symbol)
}

// GetDividendChannel returns dividend channel
func (m *MockStockBroker) GetDividendChannel() <-chan interfaces.Dividend {
	return m.dividendChan
}

// GetCorporateActionChannel returns corporate action channel
func (m *MockStockBroker) GetCorporateActionChannel() <-chan interfaces.CorporateAction {
	return m.corporateActionChan
}

// GetEarningsChannel returns earnings channel
func (m *MockStockBroker) GetEarningsChannel() <-chan interfaces.Earnings {
	return m.earningsChan
}

// GetSupportedInstruments returns list of supported instruments
func (m *MockStockBroker) GetSupportedInstruments() []entities.InstrumentInfo {
	instruments := make([]entities.InstrumentInfo, 0, len(m.stockMarkets))

	for _, market := range m.stockMarkets {
		instruments = append(instruments, entities.InstrumentInfo{
			Symbol:            market.Symbol,
			BaseAsset:         market.Symbol,
			QuoteAsset:        market.Currency,
			Type:              entities.InstrumentTypeStock,
			Market:            entities.MarketTypeStock,
			IsActive:          market.Status == "ACTIVE",
			PricePrecision:    2,
			QuantityPrecision: 0,
		})
	}

	return instruments
}

// Start starts data simulation
func (m *MockStockBroker) Start(ctx context.Context) error {
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
func (m *MockStockBroker) Stop() error {
	m.simulationMu.Lock()
	m.simulationRunning = false
	m.simulationMu.Unlock()

	return m.BaseBroker.Stop()
}

// runDataSimulation starts market data simulation
func (m *MockStockBroker) runDataSimulation() {
	defer m.wg.Done()

	ticker := time.NewTicker(200 * time.Millisecond) // Generate data every 200ms (slower than crypto)
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
func (m *MockStockBroker) generateMarketData() {
	now := time.Now()

	// Generate data only during trading hours (simplified)
	if !m.isTradingHours(now) {
		return
	}

	// Generate tickers for all active subscriptions
	m.subMu.RLock()
	for _, subscription := range m.subscriptions {
		// Generate ticker
		if ticker := m.dataGenerator.GenerateTicker(subscription.Symbol, subscription.Market, now); ticker != nil {
			ticker.BrokerID = m.config.ID
			m.SendTicker(*ticker)
		}

		// Sometimes generate candles
		if rand.Float32() < 0.05 { // 5% probability
			if candle := m.dataGenerator.GenerateCandle(subscription.Symbol, subscription.Market, now); candle != nil {
				candle.BrokerID = m.config.ID
				m.SendCandle(*candle)
			}
		}

		// Sometimes generate order books
		if rand.Float32() < 0.02 { // 2% probability
			if orderBook := m.dataGenerator.GenerateOrderBook(subscription.Symbol, subscription.Market, now); orderBook != nil {
				orderBook.BrokerID = m.config.ID
				m.SendOrderBook(*orderBook)
			}
		}

		// Rarely generate dividends
		if rand.Float32() < 0.001 { // 0.1% probability
			dividend := m.dataGenerator.GenerateDividend(subscription.Symbol, now)
			dividend.BrokerID = m.config.ID
			m.sendDividend(dividend)
		}

		// Rarely generate corporate actions
		if rand.Float32() < 0.0005 { // 0.05% probability
			action := m.dataGenerator.GenerateCorporateAction(subscription.Symbol, now)
			action.BrokerID = m.config.ID
			m.sendCorporateAction(action)
		}

		// Rarely generate earnings reports
		if rand.Float32() < 0.0002 { // 0.02% probability
			earnings := m.dataGenerator.GenerateEarnings(subscription.Symbol, now)
			earnings.BrokerID = m.config.ID
			m.sendEarnings(earnings)
		}
	}
	m.subMu.RUnlock()
}

// isTradingHours checks if it's trading hours (simplified check)
func (m *MockStockBroker) isTradingHours(t time.Time) bool {
	// Simplified: trading from 9:30 to 16:00 on weekdays
	weekday := t.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}

	hour := t.Hour()
	minute := t.Minute()

	// 9:30 - 16:00
	if hour < 9 || hour > 16 {
		return false
	}
	if hour == 9 && minute < 30 {
		return false
	}

	return true
}

// sendDividend sends dividend
func (m *MockStockBroker) sendDividend(dividend interfaces.Dividend) {
	select {
	case m.dividendChan <- dividend:
	default:
		// Channel is full, skip
	}
}

// sendCorporateAction sends corporate action
func (m *MockStockBroker) sendCorporateAction(action interfaces.CorporateAction) {
	select {
	case m.corporateActionChan <- action:
	default:
		// Channel is full, skip
	}
}

// sendEarnings sends earnings report
func (m *MockStockBroker) sendEarnings(earnings interfaces.Earnings) {
	select {
	case m.earningsChan <- earnings:
	default:
		// Channel is full, skip
	}
}
