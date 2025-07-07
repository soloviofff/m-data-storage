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

// BaseBroker предоставляет базовую реализацию интерфейса Broker
type BaseBroker struct {
	config       interfaces.BrokerConfig
	logger       *logrus.Logger
	connected    bool
	mu           sync.RWMutex
	
	// Каналы для данных
	tickerChan    chan entities.Ticker
	candleChan    chan entities.Candle
	orderBookChan chan entities.OrderBook
	
	// Управление подписками
	subscriptions map[string]entities.InstrumentSubscription
	subMu         sync.RWMutex
	
	// Управление жизненным циклом
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	
	// Статистика
	stats BrokerStats
}

// BrokerStats содержит статистику работы брокера
type BrokerStats struct {
	ConnectedAt       time.Time `json:"connected_at"`
	LastDataReceived  time.Time `json:"last_data_received"`
	TotalTickers      int64     `json:"total_tickers"`
	TotalCandles      int64     `json:"total_candles"`
	TotalOrderBooks   int64     `json:"total_orderbooks"`
	ActiveSubscriptions int     `json:"active_subscriptions"`
	ConnectionErrors  int64     `json:"connection_errors"`
	DataErrors        int64     `json:"data_errors"`
}

// NewBaseBroker создает новый базовый брокер
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

// Connect устанавливает соединение с брокером
func (b *BaseBroker) Connect(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if b.connected {
		return nil
	}
	
	b.logger.WithField("broker_id", b.config.ID).Info("Connecting to broker")
	
	// Базовая реализация - просто помечаем как подключенный
	// Конкретные брокеры должны переопределить этот метод
	b.connected = true
	b.stats.ConnectedAt = time.Now()
	
	b.logger.WithField("broker_id", b.config.ID).Info("Successfully connected to broker")
	return nil
}

// Disconnect отключается от брокера
func (b *BaseBroker) Disconnect() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if !b.connected {
		return nil
	}
	
	b.logger.WithField("broker_id", b.config.ID).Info("Disconnecting from broker")
	
	// Отменяем контекст для остановки всех горутин
	b.cancel()
	
	// Ждем завершения всех горутин
	b.wg.Wait()
	
	// Закрываем каналы
	close(b.tickerChan)
	close(b.candleChan)
	close(b.orderBookChan)
	
	b.connected = false
	
	b.logger.WithField("broker_id", b.config.ID).Info("Successfully disconnected from broker")
	return nil
}

// IsConnected возвращает статус соединения
func (b *BaseBroker) IsConnected() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connected
}

// GetInfo возвращает информацию о брокере
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

// GetSupportedInstruments возвращает поддерживаемые инструменты
func (b *BaseBroker) GetSupportedInstruments() []entities.InstrumentInfo {
	// Базовая реализация возвращает пустой список
	// Конкретные брокеры должны переопределить этот метод
	return []entities.InstrumentInfo{}
}

// Subscribe подписывается на инструменты
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

// Unsubscribe отписывается от инструментов
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

// GetTickerChannel возвращает канал тикеров
func (b *BaseBroker) GetTickerChannel() <-chan entities.Ticker {
	return b.tickerChan
}

// GetCandleChannel возвращает канал свечей
func (b *BaseBroker) GetCandleChannel() <-chan entities.Candle {
	return b.candleChan
}

// GetOrderBookChannel возвращает канал стаканов
func (b *BaseBroker) GetOrderBookChannel() <-chan entities.OrderBook {
	return b.orderBookChan
}

// Start запускает брокер
func (b *BaseBroker) Start(ctx context.Context) error {
	if !b.IsConnected() {
		if err := b.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
	}
	
	b.logger.WithField("broker_id", b.config.ID).Info("Broker started")
	return nil
}

// Stop останавливает брокер
func (b *BaseBroker) Stop() error {
	return b.Disconnect()
}

// Health проверяет здоровье брокера
func (b *BaseBroker) Health() error {
	if !b.IsConnected() {
		return fmt.Errorf("broker %s is not connected", b.config.ID)
	}
	
	// Проверяем, не переполнены ли каналы
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

// GetStats возвращает статистику брокера
func (b *BaseBroker) GetStats() BrokerStats {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.stats
}

// GetSubscriptions возвращает активные подписки
func (b *BaseBroker) GetSubscriptions() map[string]entities.InstrumentSubscription {
	b.subMu.RLock()
	defer b.subMu.RUnlock()
	
	result := make(map[string]entities.InstrumentSubscription)
	for k, v := range b.subscriptions {
		result[k] = v
	}
	return result
}

// SendTicker отправляет тикер в канал (для использования в наследниках)
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

// SendCandle отправляет свечу в канал (для использования в наследниках)
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

// SendOrderBook отправляет стакан в канал (для использования в наследниках)
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
