package services

import (
	"strings"
	"time"

	"github.com/pkg/errors"

	"m-data-storage/internal/domain/entities"
)

// DataValidatorService реализует интерфейс DataValidator
type DataValidatorService struct {
	// Настройки валидации
	maxSymbolLength   int
	maxBrokerIDLength int
	maxTimeframLength int
	maxPriceLevels    int
	maxFutureTime     time.Duration
	maxPastTime       time.Duration
	minPrice          float64
	maxPrice          float64
	minVolume         float64
	maxVolume         float64
}

// NewDataValidatorService создает новый сервис валидации данных
func NewDataValidatorService() *DataValidatorService {
	return &DataValidatorService{
		maxSymbolLength:   50,
		maxBrokerIDLength: 50,
		maxTimeframLength: 10,
		maxPriceLevels:    1000,
		maxFutureTime:     5 * time.Minute,      // Максимум 5 минут в будущем
		maxPastTime:       365 * 24 * time.Hour, // Максимум год в прошлом
		minPrice:          0.0000001,            // Минимальная цена
		maxPrice:          1000000000,           // Максимальная цена
		minVolume:         0,                    // Минимальный объем
		maxVolume:         1000000000000,        // Максимальный объем
	}
}

// ValidateTicker валидирует тикер
func (v *DataValidatorService) ValidateTicker(ticker entities.Ticker) error {
	// Проверяем базовые поля
	if err := v.validateSymbol(ticker.Symbol); err != nil {
		return errors.Wrap(err, "invalid symbol")
	}

	if err := v.validateBrokerID(ticker.BrokerID); err != nil {
		return errors.Wrap(err, "invalid broker_id")
	}

	if err := v.validateTimestamp(ticker.Timestamp); err != nil {
		return errors.Wrap(err, "invalid timestamp")
	}

	// Проверяем цену
	if err := v.validatePrice(ticker.Price); err != nil {
		return errors.Wrap(err, "invalid price")
	}

	// Проверяем объем
	if err := v.validateVolume(ticker.Volume); err != nil {
		return errors.Wrap(err, "invalid volume")
	}

	// Проверяем тип рынка
	if err := v.validateMarketType(ticker.Market); err != nil {
		return errors.Wrap(err, "invalid market type")
	}

	// Проверяем тип инструмента
	if err := v.validateInstrumentType(ticker.Type); err != nil {
		return errors.Wrap(err, "invalid instrument type")
	}

	// Проверяем дополнительные поля для акций
	if ticker.Type == entities.InstrumentTypeStock {
		if ticker.BidPrice > 0 && ticker.AskPrice > 0 {
			if ticker.BidPrice >= ticker.AskPrice {
				return errors.New("bid price must be less than ask price")
			}
		}

		if ticker.BidSize < 0 || ticker.AskSize < 0 {
			return errors.New("bid/ask sizes cannot be negative")
		}
	}

	// Проверяем дополнительные поля
	if ticker.High24h > 0 && ticker.Low24h > 0 {
		if ticker.High24h < ticker.Low24h {
			return errors.New("high24h cannot be less than low24h")
		}

		if ticker.Price > ticker.High24h || ticker.Price < ticker.Low24h {
			return errors.New("price must be within 24h range")
		}
	}

	if ticker.Volume24h < 0 {
		return errors.New("volume24h cannot be negative")
	}

	if ticker.OpenInterest < 0 {
		return errors.New("open interest cannot be negative")
	}

	return nil
}

// ValidateCandle валидирует свечу
func (v *DataValidatorService) ValidateCandle(candle entities.Candle) error {
	// Проверяем базовые поля
	if err := v.validateSymbol(candle.Symbol); err != nil {
		return errors.Wrap(err, "invalid symbol")
	}

	if err := v.validateBrokerID(candle.BrokerID); err != nil {
		return errors.Wrap(err, "invalid broker_id")
	}

	if err := v.validateTimestamp(candle.Timestamp); err != nil {
		return errors.Wrap(err, "invalid timestamp")
	}

	if err := v.validateTimeframe(candle.Timeframe); err != nil {
		return errors.Wrap(err, "invalid timeframe")
	}

	// Проверяем цены OHLC
	if err := v.validatePrice(candle.Open); err != nil {
		return errors.Wrap(err, "invalid open price")
	}

	if err := v.validatePrice(candle.High); err != nil {
		return errors.Wrap(err, "invalid high price")
	}

	if err := v.validatePrice(candle.Low); err != nil {
		return errors.Wrap(err, "invalid low price")
	}

	if err := v.validatePrice(candle.Close); err != nil {
		return errors.Wrap(err, "invalid close price")
	}

	// Проверяем объем
	if err := v.validateVolume(candle.Volume); err != nil {
		return errors.Wrap(err, "invalid volume")
	}

	// Проверяем тип рынка
	if err := v.validateMarketType(candle.Market); err != nil {
		return errors.Wrap(err, "invalid market type")
	}

	// Проверяем тип инструмента
	if err := v.validateInstrumentType(candle.Type); err != nil {
		return errors.Wrap(err, "invalid instrument type")
	}

	// Проверяем логику OHLC
	if candle.High < candle.Low {
		return errors.New("high price cannot be less than low price")
	}

	if candle.High < candle.Open || candle.High < candle.Close {
		return errors.New("high price must be >= open and close prices")
	}

	if candle.Low > candle.Open || candle.Low > candle.Close {
		return errors.New("low price must be <= open and close prices")
	}

	// Проверяем дополнительные поля
	if candle.Trades < 0 {
		return errors.New("trades count cannot be negative")
	}

	if candle.QuoteVolume < 0 {
		return errors.New("quote volume cannot be negative")
	}

	if candle.OpenInterest < 0 {
		return errors.New("open interest cannot be negative")
	}

	return nil
}

// ValidateOrderBook валидирует ордербук
func (v *DataValidatorService) ValidateOrderBook(orderBook entities.OrderBook) error {
	// Проверяем базовые поля
	if err := v.validateSymbol(orderBook.Symbol); err != nil {
		return errors.Wrap(err, "invalid symbol")
	}

	if err := v.validateBrokerID(orderBook.BrokerID); err != nil {
		return errors.Wrap(err, "invalid broker_id")
	}

	if err := v.validateTimestamp(orderBook.Timestamp); err != nil {
		return errors.Wrap(err, "invalid timestamp")
	}

	// Проверяем тип рынка
	if err := v.validateMarketType(orderBook.Market); err != nil {
		return errors.Wrap(err, "invalid market type")
	}

	// Проверяем тип инструмента
	if err := v.validateInstrumentType(orderBook.Type); err != nil {
		return errors.Wrap(err, "invalid instrument type")
	}

	// Проверяем количество уровней
	if len(orderBook.Bids) > v.maxPriceLevels {
		return errors.Errorf("too many bid levels: %d > %d", len(orderBook.Bids), v.maxPriceLevels)
	}

	if len(orderBook.Asks) > v.maxPriceLevels {
		return errors.Errorf("too many ask levels: %d > %d", len(orderBook.Asks), v.maxPriceLevels)
	}

	// Проверяем уровни bids
	for i, bid := range orderBook.Bids {
		if err := v.validatePrice(bid.Price); err != nil {
			return errors.Wrapf(err, "invalid bid price at level %d", i)
		}

		if err := v.validateVolume(bid.Quantity); err != nil {
			return errors.Wrapf(err, "invalid bid quantity at level %d", i)
		}
	}

	// Проверяем уровни asks
	for i, ask := range orderBook.Asks {
		if err := v.validatePrice(ask.Price); err != nil {
			return errors.Wrapf(err, "invalid ask price at level %d", i)
		}

		if err := v.validateVolume(ask.Quantity); err != nil {
			return errors.Wrapf(err, "invalid ask quantity at level %d", i)
		}
	}

	// Проверяем, что лучший bid меньше лучшего ask
	bestBid := orderBook.GetBestBid()
	bestAsk := orderBook.GetBestAsk()

	if bestBid != nil && bestAsk != nil {
		if bestBid.Price >= bestAsk.Price {
			return errors.New("best bid price must be less than best ask price")
		}
	}

	return nil
}

// ValidateInstrument валидирует информацию об инструменте
func (v *DataValidatorService) ValidateInstrument(instrument entities.InstrumentInfo) error {
	if err := v.validateSymbol(instrument.Symbol); err != nil {
		return errors.Wrap(err, "invalid symbol")
	}

	// Для не-акций проверяем базовый и котируемый активы
	if instrument.Type != entities.InstrumentTypeStock {
		if strings.TrimSpace(instrument.BaseAsset) == "" {
			return errors.New("base asset is required for non-stock instruments")
		}

		if strings.TrimSpace(instrument.QuoteAsset) == "" {
			return errors.New("quote asset is required for non-stock instruments")
		}
	}

	// Проверяем тип рынка
	if err := v.validateMarketType(instrument.Market); err != nil {
		return errors.Wrap(err, "invalid market type")
	}

	// Проверяем тип инструмента
	if err := v.validateInstrumentType(instrument.Type); err != nil {
		return errors.Wrap(err, "invalid instrument type")
	}

	// Проверяем ценовые ограничения
	if instrument.MinPrice < 0 {
		return errors.New("min price cannot be negative")
	}

	if instrument.MaxPrice > 0 && instrument.MaxPrice < instrument.MinPrice {
		return errors.New("max price cannot be less than min price")
	}

	// Проверяем ограничения по количеству
	if instrument.MinQuantity < 0 {
		return errors.New("min quantity cannot be negative")
	}

	if instrument.MaxQuantity > 0 && instrument.MaxQuantity < instrument.MinQuantity {
		return errors.New("max quantity cannot be less than min quantity")
	}

	// Проверяем точность
	if instrument.PricePrecision < 0 || instrument.PricePrecision > 18 {
		return errors.New("price precision must be between 0 and 18")
	}

	if instrument.QuantityPrecision < 0 || instrument.QuantityPrecision > 18 {
		return errors.New("quantity precision must be between 0 and 18")
	}

	return nil
}

// ValidateSubscription валидирует подписку на инструмент
func (v *DataValidatorService) ValidateSubscription(subscription entities.InstrumentSubscription) error {
	if err := v.validateSymbol(subscription.Symbol); err != nil {
		return errors.Wrap(err, "invalid symbol")
	}

	if err := v.validateBrokerID(subscription.BrokerID); err != nil {
		return errors.Wrap(err, "invalid broker_id")
	}

	// Проверяем тип рынка
	if err := v.validateMarketType(subscription.Market); err != nil {
		return errors.Wrap(err, "invalid market type")
	}

	// Проверяем тип инструмента
	if err := v.validateInstrumentType(subscription.Type); err != nil {
		return errors.Wrap(err, "invalid instrument type")
	}

	// Проверяем типы данных
	if len(subscription.DataTypes) == 0 {
		return errors.New("at least one data type is required")
	}

	for _, dataType := range subscription.DataTypes {
		if err := v.validateDataType(dataType); err != nil {
			return errors.Wrap(err, "invalid data type")
		}
	}

	// Проверяем дату начала
	if subscription.StartDate.IsZero() {
		return errors.New("start date is required")
	}

	if subscription.StartDate.After(time.Now().Add(v.maxFutureTime)) {
		return errors.New("start date cannot be too far in the future")
	}

	return nil
}

// Вспомогательные методы валидации

func (v *DataValidatorService) validateSymbol(symbol string) error {
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return errors.New("symbol cannot be empty")
	}

	if len(symbol) > v.maxSymbolLength {
		return errors.Errorf("symbol too long: %d > %d", len(symbol), v.maxSymbolLength)
	}

	return nil
}

func (v *DataValidatorService) validateBrokerID(brokerID string) error {
	brokerID = strings.TrimSpace(brokerID)
	if brokerID == "" {
		return errors.New("broker_id cannot be empty")
	}

	if len(brokerID) > v.maxBrokerIDLength {
		return errors.Errorf("broker_id too long: %d > %d", len(brokerID), v.maxBrokerIDLength)
	}

	return nil
}

func (v *DataValidatorService) validateTimestamp(timestamp time.Time) error {
	if timestamp.IsZero() {
		return errors.New("timestamp cannot be zero")
	}

	now := time.Now()
	if timestamp.After(now.Add(v.maxFutureTime)) {
		return errors.New("timestamp cannot be too far in the future")
	}

	if timestamp.Before(now.Add(-v.maxPastTime)) {
		return errors.New("timestamp cannot be too far in the past")
	}

	return nil
}

func (v *DataValidatorService) validatePrice(price float64) error {
	if price <= v.minPrice {
		return errors.Errorf("price too small: %f <= %f", price, v.minPrice)
	}

	if price > v.maxPrice {
		return errors.Errorf("price too large: %f > %f", price, v.maxPrice)
	}

	return nil
}

func (v *DataValidatorService) validateVolume(volume float64) error {
	if volume < v.minVolume {
		return errors.Errorf("volume too small: %f < %f", volume, v.minVolume)
	}

	if volume > v.maxVolume {
		return errors.Errorf("volume too large: %f > %f", volume, v.maxVolume)
	}

	return nil
}

func (v *DataValidatorService) validateTimeframe(timeframe string) error {
	timeframe = strings.TrimSpace(timeframe)
	if timeframe == "" {
		return errors.New("timeframe cannot be empty")
	}

	if len(timeframe) > v.maxTimeframLength {
		return errors.Errorf("timeframe too long: %d > %d", len(timeframe), v.maxTimeframLength)
	}

	// Проверяем допустимые значения
	validTimeframes := []string{"1s", "5s", "15s", "30s", "1m", "3m", "5m", "15m", "30m", "1h", "2h", "4h", "6h", "8h", "12h", "1d", "3d", "1w", "1M"}
	for _, valid := range validTimeframes {
		if timeframe == valid {
			return nil
		}
	}

	return errors.Errorf("invalid timeframe: %s", timeframe)
}

func (v *DataValidatorService) validateMarketType(market entities.MarketType) error {
	switch market {
	case entities.MarketTypeSpot, entities.MarketTypeFutures, entities.MarketTypeStock:
		return nil
	default:
		return errors.Errorf("invalid market type: %s", market)
	}
}

func (v *DataValidatorService) validateInstrumentType(instrumentType entities.InstrumentType) error {
	switch instrumentType {
	case entities.InstrumentTypeSpot, entities.InstrumentTypeFutures, entities.InstrumentTypeStock, entities.InstrumentTypeETF, entities.InstrumentTypeBond:
		return nil
	default:
		return errors.Errorf("invalid instrument type: %s", instrumentType)
	}
}

func (v *DataValidatorService) validateDataType(dataType entities.DataType) error {
	switch dataType {
	case entities.DataTypeTicker, entities.DataTypeCandle, entities.DataTypeOrderBook:
		return nil
	default:
		return errors.Errorf("invalid data type: %s", dataType)
	}
}
