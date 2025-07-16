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

// MockStockBroker реализует интерфейс StockBroker для тестирования
type MockStockBroker struct {
	*BaseBroker
	
	// Специфичные каналы для фондового рынка
	dividendChan         chan interfaces.Dividend
	corporateActionChan  chan interfaces.CorporateAction
	earningsChan         chan interfaces.Earnings
	
	// Данные для симуляции
	stockMarkets  []interfaces.StockMarket
	sectors       []interfaces.Sector
	companyInfos  map[string]*interfaces.CompanyInfo
	
	// Генератор данных
	dataGenerator *StockDataGenerator
	
	// Управление симуляцией
	simulationRunning bool
	simulationMu      sync.RWMutex
}

// NewMockStockBroker создает новый mock фондовый брокер
func NewMockStockBroker(config interfaces.BrokerConfig, logger *logrus.Logger) *MockStockBroker {
	baseBroker := NewBaseBroker(config, logger)
	
	mock := &MockStockBroker{
		BaseBroker:          baseBroker,
		dividendChan:        make(chan interfaces.Dividend, config.Defaults.BufferSize),
		corporateActionChan: make(chan interfaces.CorporateAction, config.Defaults.BufferSize),
		earningsChan:        make(chan interfaces.Earnings, config.Defaults.BufferSize),
		companyInfos:        make(map[string]*interfaces.CompanyInfo),
	}
	
	// Инициализируем тестовые данные
	mock.initializeTestData()
	
	// Создаем генератор данных
	mock.dataGenerator = NewStockDataGenerator(mock.stockMarkets, mock.sectors)
	
	return mock
}

// initializeTestData инициализирует тестовые рынки и компании
func (m *MockStockBroker) initializeTestData() {
	// Секторы
	m.sectors = []interfaces.Sector{
		{Code: "TECH", Name: "Technology", Description: "Technology companies"},
		{Code: "FINL", Name: "Financial", Description: "Financial services"},
		{Code: "HLTH", Name: "Healthcare", Description: "Healthcare and pharmaceuticals"},
		{Code: "ENRG", Name: "Energy", Description: "Energy and utilities"},
		{Code: "CONS", Name: "Consumer", Description: "Consumer goods and services"},
	}
	
	// Фондовые рынки
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
	
	// Информация о компаниях
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
			PERatio:           15.0 + rand.Float64()*20.0, // P/E от 15 до 35
			DividendYield:     rand.Float64() * 3.0,       // Дивидендная доходность 0-3%
			Beta:              0.5 + rand.Float64()*1.5,   // Бета от 0.5 до 2.0
			EPS:               rand.Float64() * 10.0,       // EPS от 0 до 10
			BookValue:         rand.Float64() * 50.0,       // Балансовая стоимость
			Description:       fmt.Sprintf("%s is a leading company in %s sector", market.CompanyName, market.Industry),
			Website:           fmt.Sprintf("https://www.%s.com", market.Symbol),
			CEO:               "John Doe", // Упрощенно
			Employees:         int(rand.Intn(500000) + 10000),
			Founded:           time.Date(1900+rand.Intn(100), time.Month(rand.Intn(12)+1), rand.Intn(28)+1, 0, 0, 0, 0, time.UTC),
			IPODate:           market.ListingDate,
		}
	}
}

// GetStockMarkets возвращает список фондовых рынков
func (m *MockStockBroker) GetStockMarkets() ([]interfaces.StockMarket, error) {
	return m.stockMarkets, nil
}

// GetSectors возвращает список секторов
func (m *MockStockBroker) GetSectors() ([]interfaces.Sector, error) {
	return m.sectors, nil
}

// SubscribeStocks подписывается на акции
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

// GetCompanyInfo возвращает информацию о компании
func (m *MockStockBroker) GetCompanyInfo(symbol string) (*interfaces.CompanyInfo, error) {
	if info, exists := m.companyInfos[symbol]; exists {
		return info, nil
	}
	return nil, fmt.Errorf("company info not found for symbol: %s", symbol)
}

// GetDividendChannel возвращает канал дивидендов
func (m *MockStockBroker) GetDividendChannel() <-chan interfaces.Dividend {
	return m.dividendChan
}

// GetCorporateActionChannel возвращает канал корпоративных действий
func (m *MockStockBroker) GetCorporateActionChannel() <-chan interfaces.CorporateAction {
	return m.corporateActionChan
}

// GetEarningsChannel возвращает канал отчетов о прибыли
func (m *MockStockBroker) GetEarningsChannel() <-chan interfaces.Earnings {
	return m.earningsChan
}

// GetSupportedInstruments возвращает список поддерживаемых инструментов
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

// Start запускает симуляцию данных
func (m *MockStockBroker) Start(ctx context.Context) error {
	if err := m.BaseBroker.Start(ctx); err != nil {
		return err
	}
	
	m.simulationMu.Lock()
	m.simulationRunning = true
	m.simulationMu.Unlock()
	
	// Запускаем генерацию данных
	m.wg.Add(1)
	go m.runDataSimulation()
	
	return nil
}

// Stop останавливает симуляцию данных
func (m *MockStockBroker) Stop() error {
	m.simulationMu.Lock()
	m.simulationRunning = false
	m.simulationMu.Unlock()
	
	return m.BaseBroker.Stop()
}

// runDataSimulation запускает симуляцию рыночных данных
func (m *MockStockBroker) runDataSimulation() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(200 * time.Millisecond) // Генерируем данные каждые 200мс (медленнее чем крипто)
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

// generateMarketData генерирует случайные рыночные данные
func (m *MockStockBroker) generateMarketData() {
	now := time.Now()
	
	// Генерируем данные только в торговые часы (упрощенно)
	if !m.isTradingHours(now) {
		return
	}
	
	// Генерируем тикеры для всех активных подписок
	m.subMu.RLock()
	for _, subscription := range m.subscriptions {
		// Генерируем тикер
		if ticker := m.dataGenerator.GenerateTicker(subscription.Symbol, subscription.Market, now); ticker != nil {
			ticker.BrokerID = m.config.ID
			m.SendTicker(*ticker)
		}
		
		// Иногда генерируем свечи
		if rand.Float32() < 0.05 { // 5% вероятность
			if candle := m.dataGenerator.GenerateCandle(subscription.Symbol, subscription.Market, now); candle != nil {
				candle.BrokerID = m.config.ID
				m.SendCandle(*candle)
			}
		}
		
		// Иногда генерируем стаканы
		if rand.Float32() < 0.02 { // 2% вероятность
			if orderBook := m.dataGenerator.GenerateOrderBook(subscription.Symbol, subscription.Market, now); orderBook != nil {
				orderBook.BrokerID = m.config.ID
				m.SendOrderBook(*orderBook)
			}
		}
		
		// Редко генерируем дивиденды
		if rand.Float32() < 0.001 { // 0.1% вероятность
			dividend := m.dataGenerator.GenerateDividend(subscription.Symbol, now)
			dividend.BrokerID = m.config.ID
			m.sendDividend(dividend)
		}
		
		// Редко генерируем корпоративные действия
		if rand.Float32() < 0.0005 { // 0.05% вероятность
			action := m.dataGenerator.GenerateCorporateAction(subscription.Symbol, now)
			action.BrokerID = m.config.ID
			m.sendCorporateAction(action)
		}
		
		// Редко генерируем отчеты о прибыли
		if rand.Float32() < 0.0002 { // 0.02% вероятность
			earnings := m.dataGenerator.GenerateEarnings(subscription.Symbol, now)
			earnings.BrokerID = m.config.ID
			m.sendEarnings(earnings)
		}
	}
	m.subMu.RUnlock()
}

// isTradingHours проверяет, торговые ли часы (упрощенная проверка)
func (m *MockStockBroker) isTradingHours(t time.Time) bool {
	// Упрощенно: торговля с 9:30 до 16:00 по будням
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

// sendDividend отправляет дивиденд
func (m *MockStockBroker) sendDividend(dividend interfaces.Dividend) {
	select {
	case m.dividendChan <- dividend:
	default:
		// Канал переполнен, пропускаем
	}
}

// sendCorporateAction отправляет корпоративное действие
func (m *MockStockBroker) sendCorporateAction(action interfaces.CorporateAction) {
	select {
	case m.corporateActionChan <- action:
	default:
		// Канал переполнен, пропускаем
	}
}

// sendEarnings отправляет отчет о прибыли
func (m *MockStockBroker) sendEarnings(earnings interfaces.Earnings) {
	select {
	case m.earningsChan <- earnings:
	default:
		// Канал переполнен, пропускаем
	}
}
