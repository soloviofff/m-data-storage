package broker

import (
	"math/rand"
	"time"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// CryptoDataGenerator генерирует реалистичные данные для криптобирж
type CryptoDataGenerator struct {
	spotMarkets    []interfaces.SpotMarket
	futuresMarkets []interfaces.FuturesMarket

	// Текущие цены для симуляции
	currentPrices map[string]float64

	// Базовые цены для различных символов
	basePrices map[string]float64
}

// NewCryptoDataGenerator создает новый генератор данных
func NewCryptoDataGenerator(spotMarkets []interfaces.SpotMarket, futuresMarkets []interfaces.FuturesMarket) *CryptoDataGenerator {
	generator := &CryptoDataGenerator{
		spotMarkets:    spotMarkets,
		futuresMarkets: futuresMarkets,
		currentPrices:  make(map[string]float64),
		basePrices:     make(map[string]float64),
	}

	// Инициализируем базовые цены
	generator.initializeBasePrices()

	return generator
}

// initializeBasePrices устанавливает начальные цены для символов
func (g *CryptoDataGenerator) initializeBasePrices() {
	// Реалистичные базовые цены для популярных криптовалют
	g.basePrices["BTCUSDT"] = 45000.0
	g.basePrices["ETHUSDT"] = 3000.0
	g.basePrices["ADAUSDT"] = 0.5
	g.basePrices["BNBUSDT"] = 300.0
	g.basePrices["XRPUSDT"] = 0.6
	g.basePrices["SOLUSDT"] = 100.0
	g.basePrices["DOTUSDT"] = 7.0
	g.basePrices["LINKUSDT"] = 15.0

	// Копируем базовые цены в текущие
	for symbol, price := range g.basePrices {
		g.currentPrices[symbol] = price
	}
}

// GenerateTicker генерирует реалистичный тикер
func (g *CryptoDataGenerator) GenerateTicker(symbol string, market entities.MarketType, timestamp time.Time) *entities.Ticker {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		// Если символ не найден, используем случайную цену
		basePrice = rand.Float64() * 1000
		g.currentPrices[symbol] = basePrice
	}

	// Генерируем небольшое изменение цены (±2%)
	priceChange := (rand.Float64() - 0.5) * 0.04 * basePrice
	newPrice := basePrice + priceChange

	// Обновляем текущую цену
	g.currentPrices[symbol] = newPrice

	// Генерируем объем
	volume := rand.Float64() * 100

	ticker := &entities.Ticker{
		Symbol:    symbol,
		Price:     newPrice,
		Volume:    volume,
		Market:    market,
		Timestamp: timestamp,
	}

	// Устанавливаем тип инструмента
	switch market {
	case entities.MarketTypeSpot:
		ticker.Type = entities.InstrumentTypeSpot
	case entities.MarketTypeFutures:
		ticker.Type = entities.InstrumentTypeFutures
	}

	// Добавляем дополнительные поля
	ticker.Change = priceChange
	ticker.ChangePercent = (priceChange / basePrice) * 100
	ticker.High24h = newPrice * (1 + rand.Float64()*0.05)
	ticker.Low24h = newPrice * (1 - rand.Float64()*0.05)
	ticker.Volume24h = volume * (20 + rand.Float64()*10)

	// Для фьючерсов добавляем открытый интерес
	if market == entities.MarketTypeFutures {
		ticker.OpenInterest = rand.Float64() * 1000000
	}

	// Для спота добавляем bid/ask
	if market == entities.MarketTypeSpot {
		spread := newPrice * 0.001 // 0.1% спред
		ticker.BidPrice = newPrice - spread/2
		ticker.AskPrice = newPrice + spread/2
		ticker.BidSize = rand.Float64() * 10
		ticker.AskSize = rand.Float64() * 10
	}

	return ticker
}

// GenerateCandle генерирует реалистичную свечу
func (g *CryptoDataGenerator) GenerateCandle(symbol string, market entities.MarketType, timestamp time.Time) *entities.Candle {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = rand.Float64() * 1000
		g.currentPrices[symbol] = basePrice
	}

	// Генерируем OHLC данные
	open := basePrice
	volatility := 0.02 // 2% волатильность

	high := open * (1 + rand.Float64()*volatility)
	low := open * (1 - rand.Float64()*volatility)
	close := low + rand.Float64()*(high-low)

	// Обновляем текущую цену
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

	// Устанавливаем тип инструмента
	switch market {
	case entities.MarketTypeSpot:
		candle.Type = entities.InstrumentTypeSpot
	case entities.MarketTypeFutures:
		candle.Type = entities.InstrumentTypeFutures
	}

	// Добавляем дополнительные поля
	candle.Trades = int64(rand.Intn(1000) + 100)
	candle.QuoteVolume = volume * close

	// Для фьючерсов добавляем открытый интерес
	if market == entities.MarketTypeFutures {
		candle.OpenInterest = rand.Float64() * 1000000
	}

	return candle
}

// GenerateOrderBook генерирует реалистичный стакан заявок
func (g *CryptoDataGenerator) GenerateOrderBook(symbol string, market entities.MarketType, timestamp time.Time) *entities.OrderBook {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = rand.Float64() * 1000
		g.currentPrices[symbol] = basePrice
	}

	// Генерируем уровни bid и ask
	bids := make([]entities.PriceLevel, 0, 20)
	asks := make([]entities.PriceLevel, 0, 20)

	// Генерируем bid уровни (ниже текущей цены)
	for i := 0; i < 20; i++ {
		price := basePrice * (1 - float64(i+1)*0.001) // Каждый уровень на 0.1% ниже
		quantity := rand.Float64() * 10
		bids = append(bids, entities.PriceLevel{
			Price:    price,
			Quantity: quantity,
		})
	}

	// Генерируем ask уровни (выше текущей цены)
	for i := 0; i < 20; i++ {
		price := basePrice * (1 + float64(i+1)*0.001) // Каждый уровень на 0.1% выше
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

	// Устанавливаем тип инструмента
	switch market {
	case entities.MarketTypeSpot:
		orderBook.Type = entities.InstrumentTypeSpot
	case entities.MarketTypeFutures:
		orderBook.Type = entities.InstrumentTypeFutures
	}

	// Добавляем дополнительные поля
	orderBook.LastUpdateID = rand.Int63n(1000000)
	orderBook.EventTime = timestamp

	return orderBook
}

// GenerateFundingRate генерирует ставку финансирования
func (g *CryptoDataGenerator) GenerateFundingRate(symbol string, timestamp time.Time) interfaces.FundingRate {
	// Генерируем реалистичную ставку финансирования (-0.1% до 0.1%)
	rate := (rand.Float64() - 0.5) * 0.002

	// Следующее время финансирования (каждые 8 часов)
	nextTime := timestamp.Truncate(8 * time.Hour).Add(8 * time.Hour)

	return interfaces.FundingRate{
		Symbol:    symbol,
		Rate:      rate,
		NextTime:  nextTime,
		Timestamp: timestamp,
	}
}

// GenerateMarkPrice генерирует маркировочную цену
func (g *CryptoDataGenerator) GenerateMarkPrice(symbol string, timestamp time.Time) interfaces.MarkPrice {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = rand.Float64() * 1000
	}

	// Маркировочная цена обычно близка к спот цене
	markPrice := basePrice * (1 + (rand.Float64()-0.5)*0.001)   // ±0.05%
	indexPrice := basePrice * (1 + (rand.Float64()-0.5)*0.0005) // ±0.025%

	return interfaces.MarkPrice{
		Symbol:     symbol,
		Price:      markPrice,
		IndexPrice: indexPrice,
		Timestamp:  timestamp,
	}
}

// GenerateLiquidation генерирует ликвидацию
func (g *CryptoDataGenerator) GenerateLiquidation(symbol string, timestamp time.Time) interfaces.Liquidation {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = rand.Float64() * 1000
	}

	// Цена ликвидации может отличаться от текущей цены
	liquidationPrice := basePrice * (1 + (rand.Float64()-0.5)*0.02) // ±1%

	// Случайная сторона
	side := "BUY"
	if rand.Float32() < 0.5 {
		side = "SELL"
	}

	// Случайное количество
	quantity := rand.Float64() * 100

	return interfaces.Liquidation{
		Symbol:    symbol,
		Side:      side,
		Price:     liquidationPrice,
		Quantity:  quantity,
		Timestamp: timestamp,
	}
}

// GetCurrentPrice возвращает текущую цену символа
func (g *CryptoDataGenerator) GetCurrentPrice(symbol string) float64 {
	if price, exists := g.currentPrices[symbol]; exists {
		return price
	}
	return 0
}

// SetCurrentPrice устанавливает текущую цену символа
func (g *CryptoDataGenerator) SetCurrentPrice(symbol string, price float64) {
	g.currentPrices[symbol] = price
}

// GenerateRealisticPriceMovement генерирует реалистичное движение цены
func (g *CryptoDataGenerator) GenerateRealisticPriceMovement(symbol string, volatility float64) float64 {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = g.basePrices[symbol]
		if basePrice == 0 {
			basePrice = rand.Float64() * 1000
		}
	}

	// Используем модель случайного блуждания с возвратом к среднему
	meanReversion := 0.01 // Сила возврата к среднему
	baseValue := g.basePrices[symbol]

	// Случайный компонент
	randomComponent := (rand.Float64() - 0.5) * volatility * basePrice

	// Компонент возврата к среднему
	meanReversionComponent := (baseValue - basePrice) * meanReversion

	// Новая цена
	newPrice := basePrice + randomComponent + meanReversionComponent

	// Убеждаемся, что цена положительная
	if newPrice <= 0 {
		newPrice = basePrice * 0.99
	}

	g.currentPrices[symbol] = newPrice
	return newPrice
}
