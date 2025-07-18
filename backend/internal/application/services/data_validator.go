package services

import (
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"m-data-storage/internal/domain/entities"
)

// DataValidatorService implements the DataValidator interface
type DataValidatorService struct {
	// Validation settings
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

	// Advanced validation settings
	enableDuplicateDetection bool
	enableAnomalyDetection   bool
	maxPriceDeviation        float64 // Maximum allowed price deviation (percentage)
	maxVolumeSpike           float64 // Maximum allowed volume spike (multiplier)

	// Data consistency tracking
	lastTickerPrices map[string]float64         // symbol -> last price
	lastCandleData   map[string]entities.Candle // symbol -> last candle
	mutex            sync.RWMutex
}

// NewDataValidatorService creates a new data validation service
func NewDataValidatorService() *DataValidatorService {
	return &DataValidatorService{
		maxSymbolLength:   50,
		maxBrokerIDLength: 50,
		maxTimeframLength: 10,
		maxPriceLevels:    1000,
		maxFutureTime:     5 * time.Minute,      // Maximum 5 minutes in the future
		maxPastTime:       365 * 24 * time.Hour, // Maximum one year in the past
		minPrice:          0.0000001,            // Minimum price
		maxPrice:          1000000000,           // Maximum price
		minVolume:         0,                    // Minimum volume
		maxVolume:         1000000000000,        // Maximum volume

		// Advanced validation settings
		enableDuplicateDetection: true,
		enableAnomalyDetection:   true,
		maxPriceDeviation:        50.0, // 50% maximum price deviation
		maxVolumeSpike:           10.0, // 10x maximum volume spike

		// Initialize tracking maps
		lastTickerPrices: make(map[string]float64),
		lastCandleData:   make(map[string]entities.Candle),
	}
}

// ValidateTicker validates a ticker
func (v *DataValidatorService) ValidateTicker(ticker entities.Ticker) error {
	// Check basic fields
	if err := v.validateSymbol(ticker.Symbol); err != nil {
		return errors.Wrap(err, "invalid symbol")
	}

	if err := v.validateBrokerID(ticker.BrokerID); err != nil {
		return errors.Wrap(err, "invalid broker_id")
	}

	if err := v.validateTimestamp(ticker.Timestamp); err != nil {
		return errors.Wrap(err, "invalid timestamp")
	}

	// Check price
	if err := v.validatePrice(ticker.Price); err != nil {
		return errors.Wrap(err, "invalid price")
	}

	// Check volume
	if err := v.validateVolume(ticker.Volume); err != nil {
		return errors.Wrap(err, "invalid volume")
	}

	// Check market type
	if err := v.validateMarketType(ticker.Market); err != nil {
		return errors.Wrap(err, "invalid market type")
	}

	// Check instrument type
	if err := v.validateInstrumentType(ticker.Type); err != nil {
		return errors.Wrap(err, "invalid instrument type")
	}

	// Check additional fields for stocks
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

	// Check additional fields
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

	// Advanced validation
	if err := v.validateTickerAdvanced(ticker); err != nil {
		return errors.Wrap(err, "advanced validation failed")
	}

	return nil
}

// ValidateCandle validates a candle
func (v *DataValidatorService) ValidateCandle(candle entities.Candle) error {
	// Check basic fields
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

	// Check OHLC prices
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

	// Check volume
	if err := v.validateVolume(candle.Volume); err != nil {
		return errors.Wrap(err, "invalid volume")
	}

	// Check market type
	if err := v.validateMarketType(candle.Market); err != nil {
		return errors.Wrap(err, "invalid market type")
	}

	// Check instrument type
	if err := v.validateInstrumentType(candle.Type); err != nil {
		return errors.Wrap(err, "invalid instrument type")
	}

	// Check OHLC logic
	if candle.High < candle.Low {
		return errors.New("high price cannot be less than low price")
	}

	if candle.High < candle.Open || candle.High < candle.Close {
		return errors.New("high price must be >= open and close prices")
	}

	if candle.Low > candle.Open || candle.Low > candle.Close {
		return errors.New("low price must be <= open and close prices")
	}

	// Check additional fields
	if candle.Trades < 0 {
		return errors.New("trades count cannot be negative")
	}

	if candle.QuoteVolume < 0 {
		return errors.New("quote volume cannot be negative")
	}

	if candle.OpenInterest < 0 {
		return errors.New("open interest cannot be negative")
	}

	// Advanced validation
	if err := v.validateCandleAdvanced(candle); err != nil {
		return errors.Wrap(err, "advanced validation failed")
	}

	return nil
}

// ValidateOrderBook validates an order book
func (v *DataValidatorService) ValidateOrderBook(orderBook entities.OrderBook) error {
	// Check basic fields
	if err := v.validateSymbol(orderBook.Symbol); err != nil {
		return errors.Wrap(err, "invalid symbol")
	}

	if err := v.validateBrokerID(orderBook.BrokerID); err != nil {
		return errors.Wrap(err, "invalid broker_id")
	}

	if err := v.validateTimestamp(orderBook.Timestamp); err != nil {
		return errors.Wrap(err, "invalid timestamp")
	}

	// Check market type
	if err := v.validateMarketType(orderBook.Market); err != nil {
		return errors.Wrap(err, "invalid market type")
	}

	// Check instrument type
	if err := v.validateInstrumentType(orderBook.Type); err != nil {
		return errors.Wrap(err, "invalid instrument type")
	}

	// Check number of levels
	if len(orderBook.Bids) > v.maxPriceLevels {
		return errors.Errorf("too many bid levels: %d > %d", len(orderBook.Bids), v.maxPriceLevels)
	}

	if len(orderBook.Asks) > v.maxPriceLevels {
		return errors.Errorf("too many ask levels: %d > %d", len(orderBook.Asks), v.maxPriceLevels)
	}

	// Check bid levels
	for i, bid := range orderBook.Bids {
		if err := v.validatePrice(bid.Price); err != nil {
			return errors.Wrapf(err, "invalid bid price at level %d", i)
		}

		if err := v.validateVolume(bid.Quantity); err != nil {
			return errors.Wrapf(err, "invalid bid quantity at level %d", i)
		}
	}

	// Check ask levels
	for i, ask := range orderBook.Asks {
		if err := v.validatePrice(ask.Price); err != nil {
			return errors.Wrapf(err, "invalid ask price at level %d", i)
		}

		if err := v.validateVolume(ask.Quantity); err != nil {
			return errors.Wrapf(err, "invalid ask quantity at level %d", i)
		}
	}

	// Check that best bid is less than best ask
	bestBid := orderBook.GetBestBid()
	bestAsk := orderBook.GetBestAsk()

	if bestBid != nil && bestAsk != nil {
		if bestBid.Price >= bestAsk.Price {
			return errors.New("best bid price must be less than best ask price")
		}
	}

	return nil
}

// ValidateInstrument validates instrument information
func (v *DataValidatorService) ValidateInstrument(instrument entities.InstrumentInfo) error {
	if err := v.validateSymbol(instrument.Symbol); err != nil {
		return errors.Wrap(err, "invalid symbol")
	}

	// For non-stocks, check base and quote assets
	if instrument.Type != entities.InstrumentTypeStock {
		if strings.TrimSpace(instrument.BaseAsset) == "" {
			return errors.New("base asset is required for non-stock instruments")
		}

		if strings.TrimSpace(instrument.QuoteAsset) == "" {
			return errors.New("quote asset is required for non-stock instruments")
		}
	}

	// Check market type
	if err := v.validateMarketType(instrument.Market); err != nil {
		return errors.Wrap(err, "invalid market type")
	}

	// Check instrument type
	if err := v.validateInstrumentType(instrument.Type); err != nil {
		return errors.Wrap(err, "invalid instrument type")
	}

	// Check price constraints
	if instrument.MinPrice < 0 {
		return errors.New("min price cannot be negative")
	}

	if instrument.MaxPrice > 0 && instrument.MaxPrice < instrument.MinPrice {
		return errors.New("max price cannot be less than min price")
	}

	// Check quantity constraints
	if instrument.MinQuantity < 0 {
		return errors.New("min quantity cannot be negative")
	}

	if instrument.MaxQuantity > 0 && instrument.MaxQuantity < instrument.MinQuantity {
		return errors.New("max quantity cannot be less than min quantity")
	}

	// Check precision
	if instrument.PricePrecision < 0 || instrument.PricePrecision > 18 {
		return errors.New("price precision must be between 0 and 18")
	}

	if instrument.QuantityPrecision < 0 || instrument.QuantityPrecision > 18 {
		return errors.New("quantity precision must be between 0 and 18")
	}

	return nil
}

// ValidateSubscription validates an instrument subscription
func (v *DataValidatorService) ValidateSubscription(subscription entities.InstrumentSubscription) error {
	if err := v.validateSymbol(subscription.Symbol); err != nil {
		return errors.Wrap(err, "invalid symbol")
	}

	if err := v.validateBrokerID(subscription.BrokerID); err != nil {
		return errors.Wrap(err, "invalid broker_id")
	}

	// Check market type
	if err := v.validateMarketType(subscription.Market); err != nil {
		return errors.Wrap(err, "invalid market type")
	}

	// Check instrument type
	if err := v.validateInstrumentType(subscription.Type); err != nil {
		return errors.Wrap(err, "invalid instrument type")
	}

	// Check data types
	if len(subscription.DataTypes) == 0 {
		return errors.New("at least one data type is required")
	}

	for _, dataType := range subscription.DataTypes {
		if err := v.validateDataType(dataType); err != nil {
			return errors.Wrap(err, "invalid data type")
		}
	}

	// Check start date
	if subscription.StartDate.IsZero() {
		return errors.New("start date is required")
	}

	if subscription.StartDate.After(time.Now().Add(v.maxFutureTime)) {
		return errors.New("start date cannot be too far in the future")
	}

	return nil
}

// Helper validation methods

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

	// Check valid values
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

// ValidateMarketType validates market type string (public interface method)
func (v *DataValidatorService) ValidateMarketType(marketType string) error {
	return v.validateMarketType(entities.MarketType(marketType))
}

// ValidateInstrumentType validates instrument type string (public interface method)
func (v *DataValidatorService) ValidateInstrumentType(instrumentType string) error {
	return v.validateInstrumentType(entities.InstrumentType(instrumentType))
}

// ValidateTimeframe validates timeframe string (public interface method)
func (v *DataValidatorService) ValidateTimeframe(timeframe string) error {
	return v.validateTimeframe(timeframe)
}

// validateTickerAdvanced performs advanced validation for tickers
func (v *DataValidatorService) validateTickerAdvanced(ticker entities.Ticker) error {
	if !v.enableAnomalyDetection && !v.enableDuplicateDetection {
		return nil
	}

	key := ticker.Symbol + ":" + ticker.BrokerID

	v.mutex.Lock()
	defer v.mutex.Unlock()

	// Check for duplicates
	if v.enableDuplicateDetection {
		if lastPrice, exists := v.lastTickerPrices[key]; exists {
			// Check if this is a duplicate (same price and volume)
			if ticker.Price == lastPrice && ticker.Volume == ticker.Volume {
				return errors.New("duplicate ticker detected")
			}
		}
	}

	// Check for price anomalies
	if v.enableAnomalyDetection {
		if lastPrice, exists := v.lastTickerPrices[key]; exists {
			deviation := ((ticker.Price - lastPrice) / lastPrice) * 100
			if deviation < 0 {
				deviation = -deviation
			}

			if deviation > v.maxPriceDeviation {
				return errors.Errorf("price deviation too high: %.2f%% (max: %.2f%%)",
					deviation, v.maxPriceDeviation)
			}
		}
	}

	// Update tracking data
	v.lastTickerPrices[key] = ticker.Price

	return nil
}

// validateCandleAdvanced performs advanced validation for candles
func (v *DataValidatorService) validateCandleAdvanced(candle entities.Candle) error {
	if !v.enableAnomalyDetection && !v.enableDuplicateDetection {
		return nil
	}

	key := candle.Symbol + ":" + candle.BrokerID + ":" + candle.Timeframe

	v.mutex.Lock()
	defer v.mutex.Unlock()

	// Check for data consistency
	if v.enableAnomalyDetection {
		if lastCandle, exists := v.lastCandleData[key]; exists {
			// Check timestamp sequence
			if candle.Timestamp.Before(lastCandle.Timestamp) {
				return errors.New("candle timestamp is before previous candle")
			}

			// Check price continuity (close of previous should be near open of current)
			priceDiff := ((candle.Open - lastCandle.Close) / lastCandle.Close) * 100
			if priceDiff < 0 {
				priceDiff = -priceDiff
			}

			if priceDiff > v.maxPriceDeviation {
				return errors.Errorf("price gap too large between candles: %.2f%% (max: %.2f%%)",
					priceDiff, v.maxPriceDeviation)
			}

			// Check volume spikes
			if lastCandle.Volume > 0 {
				volumeRatio := candle.Volume / lastCandle.Volume
				if volumeRatio > v.maxVolumeSpike {
					return errors.Errorf("volume spike too high: %.2fx (max: %.2fx)",
						volumeRatio, v.maxVolumeSpike)
				}
			}
		}
	}

	// Update tracking data
	v.lastCandleData[key] = candle

	return nil
}

// SetAnomalyDetection enables or disables anomaly detection
func (v *DataValidatorService) SetAnomalyDetection(enabled bool) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.enableAnomalyDetection = enabled
}

// SetDuplicateDetection enables or disables duplicate detection
func (v *DataValidatorService) SetDuplicateDetection(enabled bool) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.enableDuplicateDetection = enabled
}

// SetMaxPriceDeviation sets the maximum allowed price deviation percentage
func (v *DataValidatorService) SetMaxPriceDeviation(deviation float64) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.maxPriceDeviation = deviation
}

// SetMaxVolumeSpike sets the maximum allowed volume spike multiplier
func (v *DataValidatorService) SetMaxVolumeSpike(spike float64) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.maxVolumeSpike = spike
}

// ClearTrackingData clears all tracking data (useful for testing or reset)
func (v *DataValidatorService) ClearTrackingData() {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	v.lastTickerPrices = make(map[string]float64)
	v.lastCandleData = make(map[string]entities.Candle)
}

// GetValidationStats returns validation statistics
func (v *DataValidatorService) GetValidationStats() map[string]interface{} {
	v.mutex.RLock()
	defer v.mutex.RUnlock()

	return map[string]interface{}{
		"anomaly_detection_enabled":   v.enableAnomalyDetection,
		"duplicate_detection_enabled": v.enableDuplicateDetection,
		"max_price_deviation":         v.maxPriceDeviation,
		"max_volume_spike":            v.maxVolumeSpike,
		"tracked_tickers":             len(v.lastTickerPrices),
		"tracked_candles":             len(v.lastCandleData),
	}
}
