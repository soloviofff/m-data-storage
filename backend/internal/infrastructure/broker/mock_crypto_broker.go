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

// MockCryptoBroker реализует интерфейс CryptoBroker для тестирования
type MockCryptoBroker struct {
	*BaseBroker
	
	// Специфичные каналы для криптобирж
	fundingRateChan   chan interfaces.FundingRate
	markPriceChan     chan interfaces.MarkPrice
	liquidationChan   chan interfaces.Liquidation
	
	// Данные для симуляции
	spotMarkets    []interfaces.SpotMarket
	futuresMarkets []interfaces.FuturesMarket
	contractInfos  map[string]*interfaces.ContractInfo
	
	// Генератор данных
	dataGenerator *CryptoDataGenerator
	
	// Управление симуляцией
	simulationRunning bool
	simulationMu      sync.RWMutex
}

// NewMockCryptoBroker создает новый mock криптоброкер
func NewMockCryptoBroker(config interfaces.BrokerConfig, logger *logrus.Logger) *MockCryptoBroker {
	baseBroker := NewBaseBroker(config, logger)
	
	mock := &MockCryptoBroker{
		BaseBroker:      baseBroker,
		fundingRateChan: make(chan interfaces.FundingRate, config.Defaults.BufferSize),
		markPriceChan:   make(chan interfaces.MarkPrice, config.Defaults.BufferSize),
		liquidationChan: make(chan interfaces.Liquidation, config.Defaults.BufferSize),
		contractInfos:   make(map[string]*interfaces.ContractInfo),
	}
	
	// Инициализируем тестовые данные
	mock.initializeTestData()
	
	// Создаем генератор данных
	mock.dataGenerator = NewCryptoDataGenerator(mock.spotMarkets, mock.futuresMarkets)
	
	return mock
}

// initializeTestData инициализирует тестовые рынки и контракты
func (m *MockCryptoBroker) initializeTestData() {
	// Спот рынки
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
	
	// Фьючерс рынки
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
	
	// Информация о контрактах
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

// GetSpotMarkets возвращает список спот рынков
func (m *MockCryptoBroker) GetSpotMarkets() ([]interfaces.SpotMarket, error) {
	return m.spotMarkets, nil
}

// GetFuturesMarkets возвращает список фьючерс рынков
func (m *MockCryptoBroker) GetFuturesMarkets() ([]interfaces.FuturesMarket, error) {
	return m.futuresMarkets, nil
}

// SubscribeSpot подписывается на спот рынки
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

// SubscribeFutures подписывается на фьючерс рынки
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

// GetContractInfo возвращает информацию о контракте
func (m *MockCryptoBroker) GetContractInfo(symbol string) (*interfaces.ContractInfo, error) {
	if info, exists := m.contractInfos[symbol]; exists {
		return info, nil
	}
	return nil, fmt.Errorf("contract info not found for symbol: %s", symbol)
}

// GetFundingRateChannel возвращает канал ставок финансирования
func (m *MockCryptoBroker) GetFundingRateChannel() <-chan interfaces.FundingRate {
	return m.fundingRateChan
}

// GetMarkPriceChannel возвращает канал маркировочных цен
func (m *MockCryptoBroker) GetMarkPriceChannel() <-chan interfaces.MarkPrice {
	return m.markPriceChan
}

// GetLiquidationChannel возвращает канал ликвидаций
func (m *MockCryptoBroker) GetLiquidationChannel() <-chan interfaces.Liquidation {
	return m.liquidationChan
}

// GetSupportedInstruments возвращает список поддерживаемых инструментов
func (m *MockCryptoBroker) GetSupportedInstruments() []entities.InstrumentInfo {
	instruments := make([]entities.InstrumentInfo, 0, len(m.spotMarkets)+len(m.futuresMarkets))
	
	// Добавляем спот инструменты
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
	
	// Добавляем фьючерс инструменты
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

// Start запускает симуляцию данных
func (m *MockCryptoBroker) Start(ctx context.Context) error {
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
func (m *MockCryptoBroker) Stop() error {
	m.simulationMu.Lock()
	m.simulationRunning = false
	m.simulationMu.Unlock()
	
	return m.BaseBroker.Stop()
}

// runDataSimulation запускает симуляцию рыночных данных
func (m *MockCryptoBroker) runDataSimulation() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(100 * time.Millisecond) // Генерируем данные каждые 100мс
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
func (m *MockCryptoBroker) generateMarketData() {
	now := time.Now()
	
	// Генерируем тикеры для всех активных подписок
	m.subMu.RLock()
	for _, subscription := range m.subscriptions {
		// Генерируем тикер
		if ticker := m.dataGenerator.GenerateTicker(subscription.Symbol, subscription.Market, now); ticker != nil {
			ticker.BrokerID = m.config.ID
			m.SendTicker(*ticker)
		}
		
		// Иногда генерируем свечи
		if rand.Float32() < 0.1 { // 10% вероятность
			if candle := m.dataGenerator.GenerateCandle(subscription.Symbol, subscription.Market, now); candle != nil {
				candle.BrokerID = m.config.ID
				m.SendCandle(*candle)
			}
		}
		
		// Иногда генерируем стаканы
		if rand.Float32() < 0.05 { // 5% вероятность
			if orderBook := m.dataGenerator.GenerateOrderBook(subscription.Symbol, subscription.Market, now); orderBook != nil {
				orderBook.BrokerID = m.config.ID
				m.SendOrderBook(*orderBook)
			}
		}
		
		// Для фьючерсов генерируем дополнительные данные
		if subscription.Market == entities.MarketTypeFutures {
			// Ставки финансирования
			if rand.Float32() < 0.02 { // 2% вероятность
				fundingRate := m.dataGenerator.GenerateFundingRate(subscription.Symbol, now)
				fundingRate.BrokerID = m.config.ID
				m.sendFundingRate(fundingRate)
			}
			
			// Маркировочные цены
			if rand.Float32() < 0.05 { // 5% вероятность
				markPrice := m.dataGenerator.GenerateMarkPrice(subscription.Symbol, now)
				markPrice.BrokerID = m.config.ID
				m.sendMarkPrice(markPrice)
			}
			
			// Ликвидации
			if rand.Float32() < 0.01 { // 1% вероятность
				liquidation := m.dataGenerator.GenerateLiquidation(subscription.Symbol, now)
				liquidation.BrokerID = m.config.ID
				m.sendLiquidation(liquidation)
			}
		}
	}
	m.subMu.RUnlock()
}

// sendFundingRate отправляет ставку финансирования
func (m *MockCryptoBroker) sendFundingRate(rate interfaces.FundingRate) {
	select {
	case m.fundingRateChan <- rate:
	default:
		// Канал переполнен, пропускаем
	}
}

// sendMarkPrice отправляет маркировочную цену
func (m *MockCryptoBroker) sendMarkPrice(price interfaces.MarkPrice) {
	select {
	case m.markPriceChan <- price:
	default:
		// Канал переполнен, пропускаем
	}
}

// sendLiquidation отправляет ликвидацию
func (m *MockCryptoBroker) sendLiquidation(liquidation interfaces.Liquidation) {
	select {
	case m.liquidationChan <- liquidation:
	default:
		// Канал переполнен, пропускаем
	}
}
