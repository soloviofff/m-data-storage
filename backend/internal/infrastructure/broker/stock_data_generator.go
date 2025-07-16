package broker

import (
	"fmt"
	"math/rand"
	"time"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// StockDataGenerator генерирует реалистичные данные для фондового рынка
type StockDataGenerator struct {
	stockMarkets []interfaces.StockMarket
	sectors      []interfaces.Sector
	
	// Текущие цены для симуляции
	currentPrices map[string]float64
	
	// Базовые цены для различных символов
	basePrices map[string]float64
}

// NewStockDataGenerator создает новый генератор данных
func NewStockDataGenerator(stockMarkets []interfaces.StockMarket, sectors []interfaces.Sector) *StockDataGenerator {
	generator := &StockDataGenerator{
		stockMarkets:  stockMarkets,
		sectors:       sectors,
		currentPrices: make(map[string]float64),
		basePrices:    make(map[string]float64),
	}
	
	// Инициализируем базовые цены
	generator.initializeBasePrices()
	
	return generator
}

// initializeBasePrices устанавливает начальные цены для символов
func (g *StockDataGenerator) initializeBasePrices() {
	// Реалистичные базовые цены для популярных акций
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
	
	// Копируем базовые цены в текущие
	for symbol, price := range g.basePrices {
		g.currentPrices[symbol] = price
	}
}

// GenerateTicker генерирует реалистичный тикер для акции
func (g *StockDataGenerator) GenerateTicker(symbol string, market entities.MarketType, timestamp time.Time) *entities.Ticker {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		// Если символ не найден, используем случайную цену
		basePrice = 50.0 + rand.Float64()*200.0 // От $50 до $250
		g.currentPrices[symbol] = basePrice
	}
	
	// Генерируем небольшое изменение цены (±1% для акций, меньше волатильности чем у крипто)
	priceChange := (rand.Float64() - 0.5) * 0.02 * basePrice
	newPrice := basePrice + priceChange
	
	// Убеждаемся, что цена положительная
	if newPrice <= 0 {
		newPrice = basePrice * 0.99
	}
	
	// Обновляем текущую цену
	g.currentPrices[symbol] = newPrice
	
	// Генерируем объем (акции торгуются меньшими объемами)
	volume := rand.Float64() * 1000000 // До 1M акций
	
	ticker := &entities.Ticker{
		Symbol:    symbol,
		Price:     newPrice,
		Volume:    volume,
		Market:    market,
		Type:      entities.InstrumentTypeStock,
		Timestamp: timestamp,
	}
	
	// Добавляем дополнительные поля
	ticker.Change = priceChange
	ticker.ChangePercent = (priceChange / basePrice) * 100
	ticker.High24h = newPrice * (1 + rand.Float64()*0.03) // Дневной максимум
	ticker.Low24h = newPrice * (1 - rand.Float64()*0.03)  // Дневной минимум
	ticker.Volume24h = volume * (5 + rand.Float64()*5)    // Дневной объем
	
	// Bid/Ask спред для акций
	spread := newPrice * (0.0001 + rand.Float64()*0.0009) // 0.01-0.1% спред
	ticker.BidPrice = newPrice - spread/2
	ticker.AskPrice = newPrice + spread/2
	ticker.BidSize = rand.Float64() * 1000  // Размер bid
	ticker.AskSize = rand.Float64() * 1000  // Размер ask
	
	return ticker
}

// GenerateCandle генерирует реалистичную свечу для акции
func (g *StockDataGenerator) GenerateCandle(symbol string, market entities.MarketType, timestamp time.Time) *entities.Candle {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = 50.0 + rand.Float64()*200.0
		g.currentPrices[symbol] = basePrice
	}
	
	// Генерируем OHLC данные (меньшая волатильность для акций)
	open := basePrice
	volatility := 0.01 // 1% волатильность
	
	high := open * (1 + rand.Float64()*volatility)
	low := open * (1 - rand.Float64()*volatility)
	close := low + rand.Float64()*(high-low)
	
	// Обновляем текущую цену
	g.currentPrices[symbol] = close
	
	volume := rand.Float64() * 10000000 // До 10M акций
	
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
	
	// Добавляем дополнительные поля
	candle.Trades = int64(rand.Intn(10000) + 1000)     // Количество сделок
	candle.QuoteVolume = volume * close                // Объем в долларах
	
	return candle
}

// GenerateOrderBook генерирует реалистичный стакан заявок для акции
func (g *StockDataGenerator) GenerateOrderBook(symbol string, market entities.MarketType, timestamp time.Time) *entities.OrderBook {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = 50.0 + rand.Float64()*200.0
	}
	
	// Генерируем уровни bid и ask (меньше уровней чем у крипто)
	bids := make([]entities.PriceLevel, 0, 10)
	asks := make([]entities.PriceLevel, 0, 10)
	
	// Генерируем bid уровни (ниже текущей цены)
	for i := 0; i < 10; i++ {
		price := basePrice * (1 - float64(i+1)*0.0005) // Каждый уровень на 0.05% ниже
		quantity := rand.Float64() * 1000               // До 1000 акций
		bids = append(bids, entities.PriceLevel{
			Price:    price,
			Quantity: quantity,
		})
	}
	
	// Генерируем ask уровни (выше текущей цены)
	for i := 0; i < 10; i++ {
		price := basePrice * (1 + float64(i+1)*0.0005) // Каждый уровень на 0.05% выше
		quantity := rand.Float64() * 1000               // До 1000 акций
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
	
	// Добавляем дополнительные поля
	orderBook.LastUpdateID = rand.Int63n(1000000)
	orderBook.EventTime = timestamp
	
	return orderBook
}

// GenerateDividend генерирует дивиденд
func (g *StockDataGenerator) GenerateDividend(symbol string, timestamp time.Time) interfaces.Dividend {
	// Генерируем реалистичную сумму дивиденда
	amount := rand.Float64() * 2.0 // До $2 за акцию
	
	// Даты дивиденда
	exDate := timestamp.AddDate(0, 0, rand.Intn(30)+1)      // Ex-date через 1-30 дней
	payDate := exDate.AddDate(0, 0, rand.Intn(30)+15)       // Pay date через 15-45 дней после ex-date
	recordDate := exDate.AddDate(0, 0, 2)                   // Record date через 2 дня после ex-date
	declareDate := timestamp.AddDate(0, 0, -rand.Intn(30))  // Declare date до 30 дней назад
	
	// Частота выплат
	frequencies := []string{"QUARTERLY", "MONTHLY", "ANNUALLY", "SEMI_ANNUALLY"}
	frequency := frequencies[rand.Intn(len(frequencies))]
	
	// Тип дивиденда
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

// GenerateCorporateAction генерирует корпоративное действие
func (g *StockDataGenerator) GenerateCorporateAction(symbol string, timestamp time.Time) interfaces.CorporateAction {
	// Типы корпоративных действий
	actionTypes := []string{"SPLIT", "MERGER", "SPINOFF", "DIVIDEND", "RIGHTS_OFFERING"}
	actionType := actionTypes[rand.Intn(len(actionTypes))]
	
	// Генерируем коэффициент в зависимости от типа
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
	
	// Даты
	exDate := timestamp.AddDate(0, 0, rand.Intn(60)+1)     // Ex-date через 1-60 дней
	payDate := exDate.AddDate(0, 0, rand.Intn(30)+1)       // Pay date через 1-30 дней после ex-date
	recordDate := exDate.AddDate(0, 0, 2)                  // Record date через 2 дня после ex-date
	
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

// GenerateEarnings генерирует отчет о прибыли
func (g *StockDataGenerator) GenerateEarnings(symbol string, timestamp time.Time) interfaces.Earnings {
	// Генерируем реалистичные данные о прибыли
	eps := rand.Float64() * 5.0                    // EPS от 0 до $5
	epsEstimate := eps * (0.9 + rand.Float64()*0.2) // Оценка ±10%
	
	revenue := (rand.Float64() * 50 + 10) * 1000000000 // Выручка от $10B до $60B
	revenueEstimate := revenue * (0.95 + rand.Float64()*0.1) // Оценка ±5%
	
	// Период отчета
	periods := []string{"Q1", "Q2", "Q3", "Q4", "FY"}
	period := periods[rand.Intn(len(periods))]
	
	// Дата отчета
	reportDate := timestamp.AddDate(0, 0, rand.Intn(30)+1)
	
	// Время публикации
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

// GetCurrentPrice возвращает текущую цену символа
func (g *StockDataGenerator) GetCurrentPrice(symbol string) float64 {
	if price, exists := g.currentPrices[symbol]; exists {
		return price
	}
	return 0
}

// SetCurrentPrice устанавливает текущую цену символа
func (g *StockDataGenerator) SetCurrentPrice(symbol string, price float64) {
	g.currentPrices[symbol] = price
}

// GenerateRealisticPriceMovement генерирует реалистичное движение цены для акции
func (g *StockDataGenerator) GenerateRealisticPriceMovement(symbol string, volatility float64) float64 {
	basePrice, exists := g.currentPrices[symbol]
	if !exists {
		basePrice = g.basePrices[symbol]
		if basePrice == 0 {
			basePrice = 50.0 + rand.Float64()*200.0
		}
	}
	
	// Используем модель случайного блуждания с возвратом к среднему
	// Акции имеют меньшую волатильность и более сильный возврат к среднему
	meanReversion := 0.02 // Более сильный возврат к среднему для акций
	baseValue := g.basePrices[symbol]
	
	// Случайный компонент (меньше для акций)
	randomComponent := (rand.Float64() - 0.5) * volatility * basePrice * 0.5
	
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
