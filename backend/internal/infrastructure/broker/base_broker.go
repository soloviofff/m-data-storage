package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// BaseBroker provides base implementation of Broker interface
type BaseBroker struct {
	config    interfaces.BrokerConfig
	logger    *logrus.Logger
	connected bool
	mu        sync.RWMutex

	// Data channels
	tickerChan    chan entities.Ticker
	candleChan    chan entities.Candle
	orderBookChan chan entities.OrderBook

	// Subscription management
	subscriptions map[string]entities.InstrumentSubscription
	subMu         sync.RWMutex

	// Lifecycle management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Statistics
	stats BrokerStats
}

// BrokerStats contains broker operation statistics
type BrokerStats struct {
	ConnectedAt         time.Time `json:"connected_at"`
	LastDataReceived    time.Time `json:"last_data_received"`
	TotalTickers        int64     `json:"total_tickers"`
	TotalCandles        int64     `json:"total_candles"`
	TotalOrderBooks     int64     `json:"total_orderbooks"`
	ActiveSubscriptions int       `json:"active_subscriptions"`
	ConnectionErrors    int64     `json:"connection_errors"`
	DataErrors          int64     `json:"data_errors"`
}

// NewBaseBroker creates a new base broker
func NewBaseBroker(config interfaces.BrokerConfig, logger *logrus.Logger) *BaseBroker {
	if logger == nil {
		logger = logrus.New()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &BaseBroker{
		config:        config,
		logger:        logger,
		connected:     false,
		tickerChan:    make(chan entities.Ticker, config.Defaults.BufferSize),
		candleChan:    make(chan entities.Candle, config.Defaults.BufferSize),
		orderBookChan: make(chan entities.OrderBook, config.Defaults.BufferSize),
		subscriptions: make(map[string]entities.InstrumentSubscription),
		ctx:           ctx,
		cancel:        cancel,
		stats:         BrokerStats{},
	}
}

// Connect establishes connection to the broker
func (b *BaseBroker) Connect(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.connected {
		return nil
	}

	b.logger.WithField("broker_id", b.config.ID).Info("Connecting to broker")

	// Base implementation - just mark as connected
	// Concrete brokers should override this method
	b.connected = true
	b.stats.ConnectedAt = time.Now()

	b.logger.WithField("broker_id", b.config.ID).Info("Successfully connected to broker")
	return nil
}

// Disconnect disconnects from the broker
func (b *BaseBroker) Disconnect() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if !b.connected {
		return nil
	}

	b.logger.WithField("broker_id", b.config.ID).Info("Disconnecting from broker")

	// Cancel context to stop all goroutines
	b.cancel()

	// Wait for all goroutines to finish
	b.wg.Wait()

	// Close channels
	close(b.tickerChan)
	close(b.candleChan)
	close(b.orderBookChan)

	b.connected = false

	b.logger.WithField("broker_id", b.config.ID).Info("Successfully disconnected from broker")
	return nil
}

// IsConnected returns connection status
func (b *BaseBroker) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected
}

// GetInfo returns broker information
func (b *BaseBroker) GetInfo() interfaces.BrokerInfo {
	status := "disconnected"
	if b.IsConnected() {
		status = "connected"
	}

	return interfaces.BrokerInfo{
		ID:          b.config.ID,
		Name:        b.config.Name,
		Type:        b.config.Type,
		Description: fmt.Sprintf("Broker %s (%s)", b.config.Name, b.config.Type),
		Website:     "",
		Status:      status,
		Features:    []string{"tickers", "candles", "orderbooks"},
	}
}

// GetSupportedInstruments returns supported instruments
func (b *BaseBroker) GetSupportedInstruments() []entities.InstrumentInfo {
	// Base implementation returns empty list
	// Concrete brokers should override this method
	return []entities.InstrumentInfo{}
}

// Subscribe subscribes to instruments
func (b *BaseBroker) Subscribe(ctx context.Context, instruments []entities.InstrumentSubscription) error {
	b.subMu.Lock()
	defer b.subMu.Unlock()

	for _, instrument := range instruments {
		key := fmt.Sprintf("%s_%s_%s", instrument.Symbol, instrument.Type, instrument.Market)
		b.subscriptions[key] = instrument

		b.logger.WithFields(logrus.Fields{
			"broker_id": b.config.ID,
			"symbol":    instrument.Symbol,
			"type":      instrument.Type,
			"market":    instrument.Market,
		}).Info("Subscribed to instrument")
	}

	b.stats.ActiveSubscriptions = len(b.subscriptions)
	return nil
}

// Unsubscribe unsubscribes from instruments
func (b *BaseBroker) Unsubscribe(ctx context.Context, instruments []entities.InstrumentSubscription) error {
	b.subMu.Lock()
	defer b.subMu.Unlock()

	for _, instrument := range instruments {
		key := fmt.Sprintf("%s_%s_%s", instrument.Symbol, instrument.Type, instrument.Market)
		delete(b.subscriptions, key)

		b.logger.WithFields(logrus.Fields{
			"broker_id": b.config.ID,
			"symbol":    instrument.Symbol,
			"type":      instrument.Type,
			"market":    instrument.Market,
		}).Info("Unsubscribed from instrument")
	}

	b.stats.ActiveSubscriptions = len(b.subscriptions)
	return nil
}

// GetTickerChannel returns ticker channel
func (b *BaseBroker) GetTickerChannel() <-chan entities.Ticker {
	return b.tickerChan
}

// GetCandleChannel returns candle channel
func (b *BaseBroker) GetCandleChannel() <-chan entities.Candle {
	return b.candleChan
}

// GetOrderBookChannel returns order book channel
func (b *BaseBroker) GetOrderBookChannel() <-chan entities.OrderBook {
	return b.orderBookChan
}

// Start starts the broker
func (b *BaseBroker) Start(ctx context.Context) error {
	if !b.IsConnected() {
		if err := b.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
	}

	b.logger.WithField("broker_id", b.config.ID).Info("Broker started")
	return nil
}

// Stop stops the broker
func (b *BaseBroker) Stop() error {
	return b.Disconnect()
}

// Health checks broker health
func (b *BaseBroker) Health() error {
	if !b.IsConnected() {
		return fmt.Errorf("broker %s is not connected", b.config.ID)
	}

	// Check if channels are not overflowing
	if len(b.tickerChan) > cap(b.tickerChan)*9/10 {
		return fmt.Errorf("ticker channel is nearly full")
	}
	if len(b.candleChan) > cap(b.candleChan)*9/10 {
		return fmt.Errorf("candle channel is nearly full")
	}
	if len(b.orderBookChan) > cap(b.orderBookChan)*9/10 {
		return fmt.Errorf("orderbook channel is nearly full")
	}

	return nil
}

// GetStats returns broker statistics
func (b *BaseBroker) GetStats() BrokerStats {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.stats
}

// GetSubscriptions returns active subscriptions
func (b *BaseBroker) GetSubscriptions() map[string]entities.InstrumentSubscription {
	b.subMu.RLock()
	defer b.subMu.RUnlock()

	result := make(map[string]entities.InstrumentSubscription)
	for k, v := range b.subscriptions {
		result[k] = v
	}
	return result
}

// SendTicker sends ticker to channel (for use in derived classes)
func (b *BaseBroker) SendTicker(ticker entities.Ticker) {
	select {
	case b.tickerChan <- ticker:
		b.stats.TotalTickers++
		b.stats.LastDataReceived = time.Now()
	default:
		b.stats.DataErrors++
		b.logger.WithField("broker_id", b.config.ID).Warn("Ticker channel is full, dropping data")
	}
}

// SendCandle sends candle to channel (for use in derived classes)
func (b *BaseBroker) SendCandle(candle entities.Candle) {
	select {
	case b.candleChan <- candle:
		b.stats.TotalCandles++
		b.stats.LastDataReceived = time.Now()
	default:
		b.stats.DataErrors++
		b.logger.WithField("broker_id", b.config.ID).Warn("Candle channel is full, dropping data")
	}
}

// SendOrderBook sends order book to channel (for use in derived classes)
func (b *BaseBroker) SendOrderBook(orderBook entities.OrderBook) {
	select {
	case b.orderBookChan <- orderBook:
		b.stats.TotalOrderBooks++
		b.stats.LastDataReceived = time.Now()
	default:
		b.stats.DataErrors++
		b.logger.WithField("broker_id", b.config.ID).Warn("OrderBook channel is full, dropping data")
	}
}
