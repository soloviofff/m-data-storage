package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// BrokerStorageIntegration интегрирует брокеры с системой хранения
type BrokerStorageIntegration struct {
	brokerManager  interfaces.BrokerManager
	storageService interfaces.StorageService
	logger         *logrus.Logger

	// Управление жизненным циклом
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Активные интеграции по брокерам
	integrations map[string]*BrokerIntegration
	mu           sync.RWMutex

	// Статистика
	stats   IntegrationStats
	statsMu sync.RWMutex
}

// BrokerIntegration представляет интеграцию одного брокера с хранилищем
type BrokerIntegration struct {
	brokerID string
	broker   interfaces.Broker

	// Каналы для остановки воркеров
	stopChan chan struct{}
	wg       sync.WaitGroup

	// Статистика по брокеру
	stats BrokerIntegrationStats
	mu    sync.RWMutex
}

// IntegrationStats содержит общую статистику интеграции
type IntegrationStats = interfaces.BrokerStorageIntegrationStats

// BrokerIntegrationStats содержит статистику интеграции по брокеру
type BrokerIntegrationStats = interfaces.BrokerIntegrationStats

// NewBrokerStorageIntegration создает новый сервис интеграции
func NewBrokerStorageIntegration(
	brokerManager interfaces.BrokerManager,
	storageService interfaces.StorageService,
	logger *logrus.Logger,
) *BrokerStorageIntegration {
	if logger == nil {
		logger = logrus.New()
	}

	return &BrokerStorageIntegration{
		brokerManager:  brokerManager,
		storageService: storageService,
		logger:         logger,
		integrations:   make(map[string]*BrokerIntegration),
		stats: IntegrationStats{
			StartedAt: time.Now(),
		},
	}
}

// NewBrokerStorageIntegrationService создает новый сервис интеграции (возвращает интерфейс)
func NewBrokerStorageIntegrationService(
	brokerManager interfaces.BrokerManager,
	storageService interfaces.StorageService,
	logger *logrus.Logger,
) interfaces.BrokerStorageIntegration {
	return NewBrokerStorageIntegration(brokerManager, storageService, logger)
}

// Start запускает интеграцию всех брокеров с хранилищем
func (bis *BrokerStorageIntegration) Start(ctx context.Context) error {
	bis.ctx, bis.cancel = context.WithCancel(ctx)

	bis.logger.Info("Starting broker-storage integration")

	// Получаем всех брокеров
	brokers := bis.brokerManager.GetAllBrokers()

	bis.mu.Lock()
	defer bis.mu.Unlock()

	// Запускаем интеграцию для каждого брокера
	for brokerID, broker := range brokers {
		if err := bis.startBrokerIntegrationUnsafe(brokerID, broker); err != nil {
			bis.logger.WithError(err).WithField("broker_id", brokerID).Error("Failed to start broker integration")
			continue
		}
	}

	bis.updateStatsUnsafe()
	bis.logger.WithField("active_brokers", len(bis.integrations)).Info("Broker-storage integration started")

	return nil
}

// Stop останавливает интеграцию
func (bis *BrokerStorageIntegration) Stop() error {
	bis.logger.Info("Stopping broker-storage integration")

	if bis.cancel != nil {
		bis.cancel()
	}

	bis.mu.Lock()
	defer bis.mu.Unlock()

	// Останавливаем все интеграции брокеров
	for brokerID, integration := range bis.integrations {
		bis.stopBrokerIntegrationUnsafe(brokerID, integration)
	}

	// Ждем завершения всех воркеров
	bis.wg.Wait()

	bis.logger.Info("Broker-storage integration stopped")
	return nil
}

// AddBroker добавляет новый брокер в интеграцию
func (bis *BrokerStorageIntegration) AddBroker(brokerID string, broker interfaces.Broker) error {
	bis.mu.Lock()
	defer bis.mu.Unlock()

	if _, exists := bis.integrations[brokerID]; exists {
		return fmt.Errorf("broker %s already integrated", brokerID)
	}

	if err := bis.startBrokerIntegrationUnsafe(brokerID, broker); err != nil {
		return fmt.Errorf("failed to start broker integration: %w", err)
	}

	bis.updateStatsUnsafe()
	bis.logger.WithField("broker_id", brokerID).Info("Broker added to integration")

	return nil
}

// RemoveBroker удаляет брокер из интеграции
func (bis *BrokerStorageIntegration) RemoveBroker(brokerID string) error {
	bis.mu.Lock()
	defer bis.mu.Unlock()

	integration, exists := bis.integrations[brokerID]
	if !exists {
		return fmt.Errorf("broker %s not found in integration", brokerID)
	}

	bis.stopBrokerIntegrationUnsafe(brokerID, integration)
	delete(bis.integrations, brokerID)

	bis.updateStatsUnsafe()
	bis.logger.WithField("broker_id", brokerID).Info("Broker removed from integration")

	return nil
}

// GetStats возвращает статистику интеграции
func (bis *BrokerStorageIntegration) GetStats() IntegrationStats {
	bis.statsMu.RLock()
	defer bis.statsMu.RUnlock()
	return bis.stats
}

// GetBrokerStats возвращает статистику по конкретному брокеру
func (bis *BrokerStorageIntegration) GetBrokerStats(brokerID string) (BrokerIntegrationStats, error) {
	bis.mu.RLock()
	defer bis.mu.RUnlock()

	integration, exists := bis.integrations[brokerID]
	if !exists {
		return BrokerIntegrationStats{}, fmt.Errorf("broker %s not found", brokerID)
	}

	integration.mu.RLock()
	defer integration.mu.RUnlock()
	return integration.stats, nil
}

// Health проверяет здоровье интеграции
func (bis *BrokerStorageIntegration) Health() error {
	bis.mu.RLock()
	defer bis.mu.RUnlock()

	if len(bis.integrations) == 0 {
		return fmt.Errorf("no active broker integrations")
	}

	// Проверяем, что данные поступают
	bis.statsMu.RLock()
	lastData := bis.stats.LastDataReceived
	bis.statsMu.RUnlock()

	if !lastData.IsZero() && time.Since(lastData) > 5*time.Minute {
		return fmt.Errorf("no data received for %v", time.Since(lastData))
	}

	return nil
}

// startBrokerIntegrationUnsafe запускает интеграцию для брокера (без блокировки)
func (bis *BrokerStorageIntegration) startBrokerIntegrationUnsafe(brokerID string, broker interfaces.Broker) error {
	integration := &BrokerIntegration{
		brokerID: brokerID,
		broker:   broker,
		stopChan: make(chan struct{}),
		stats: BrokerIntegrationStats{
			BrokerID:  brokerID,
			StartedAt: time.Now(),
		},
	}

	// Запускаем воркеры для обработки данных
	integration.wg.Add(3)
	go bis.tickerWorker(integration)
	go bis.candleWorker(integration)
	go bis.orderBookWorker(integration)

	bis.integrations[brokerID] = integration

	bis.logger.WithField("broker_id", brokerID).Info("Broker integration started")
	return nil
}

// stopBrokerIntegrationUnsafe останавливает интеграцию брокера (без блокировки)
func (bis *BrokerStorageIntegration) stopBrokerIntegrationUnsafe(brokerID string, integration *BrokerIntegration) {
	close(integration.stopChan)
	integration.wg.Wait()
	bis.logger.WithField("broker_id", brokerID).Info("Broker integration stopped")
}

// updateStatsUnsafe обновляет общую статистику (без блокировки)
func (bis *BrokerStorageIntegration) updateStatsUnsafe() {
	bis.statsMu.Lock()
	defer bis.statsMu.Unlock()

	bis.stats.ActiveBrokers = len(bis.integrations)
}

// tickerWorker обрабатывает тикеры от брокера
func (bis *BrokerStorageIntegration) tickerWorker(integration *BrokerIntegration) {
	defer integration.wg.Done()

	tickerChan := integration.broker.GetTickerChannel()

	bis.logger.WithField("broker_id", integration.brokerID).Debug("Ticker worker started")

	for {
		select {
		case <-integration.stopChan:
			bis.logger.WithField("broker_id", integration.brokerID).Debug("Ticker worker stopped")
			return

		case <-bis.ctx.Done():
			bis.logger.WithField("broker_id", integration.brokerID).Debug("Ticker worker stopped by context")
			return

		case ticker, ok := <-tickerChan:
			if !ok {
				bis.logger.WithField("broker_id", integration.brokerID).Warn("Ticker channel closed")
				return
			}

			if err := bis.processTicker(integration, ticker); err != nil {
				bis.logger.WithError(err).WithFields(logrus.Fields{
					"broker_id": integration.brokerID,
					"symbol":    ticker.Symbol,
				}).Error("Failed to process ticker")

				integration.mu.Lock()
				integration.stats.Errors++
				integration.mu.Unlock()

				bis.statsMu.Lock()
				bis.stats.TotalErrors++
				bis.statsMu.Unlock()
			}
		}
	}
}

// candleWorker обрабатывает свечи от брокера
func (bis *BrokerStorageIntegration) candleWorker(integration *BrokerIntegration) {
	defer integration.wg.Done()

	candleChan := integration.broker.GetCandleChannel()

	bis.logger.WithField("broker_id", integration.brokerID).Debug("Candle worker started")

	for {
		select {
		case <-integration.stopChan:
			bis.logger.WithField("broker_id", integration.brokerID).Debug("Candle worker stopped")
			return

		case <-bis.ctx.Done():
			bis.logger.WithField("broker_id", integration.brokerID).Debug("Candle worker stopped by context")
			return

		case candle, ok := <-candleChan:
			if !ok {
				bis.logger.WithField("broker_id", integration.brokerID).Warn("Candle channel closed")
				return
			}

			if err := bis.processCandle(integration, candle); err != nil {
				bis.logger.WithError(err).WithFields(logrus.Fields{
					"broker_id": integration.brokerID,
					"symbol":    candle.Symbol,
				}).Error("Failed to process candle")

				integration.mu.Lock()
				integration.stats.Errors++
				integration.mu.Unlock()

				bis.statsMu.Lock()
				bis.stats.TotalErrors++
				bis.statsMu.Unlock()
			}
		}
	}
}

// orderBookWorker обрабатывает ордербуки от брокера
func (bis *BrokerStorageIntegration) orderBookWorker(integration *BrokerIntegration) {
	defer integration.wg.Done()

	orderBookChan := integration.broker.GetOrderBookChannel()

	bis.logger.WithField("broker_id", integration.brokerID).Debug("OrderBook worker started")

	for {
		select {
		case <-integration.stopChan:
			bis.logger.WithField("broker_id", integration.brokerID).Debug("OrderBook worker stopped")
			return

		case <-bis.ctx.Done():
			bis.logger.WithField("broker_id", integration.brokerID).Debug("OrderBook worker stopped by context")
			return

		case orderBook, ok := <-orderBookChan:
			if !ok {
				bis.logger.WithField("broker_id", integration.brokerID).Warn("OrderBook channel closed")
				return
			}

			if err := bis.processOrderBook(integration, orderBook); err != nil {
				bis.logger.WithError(err).WithFields(logrus.Fields{
					"broker_id": integration.brokerID,
					"symbol":    orderBook.Symbol,
				}).Error("Failed to process orderbook")

				integration.mu.Lock()
				integration.stats.Errors++
				integration.mu.Unlock()

				bis.statsMu.Lock()
				bis.stats.TotalErrors++
				bis.statsMu.Unlock()
			}
		}
	}
}

// processTicker обрабатывает тикер и сохраняет его в хранилище
func (bis *BrokerStorageIntegration) processTicker(integration *BrokerIntegration, ticker entities.Ticker) error {
	// Добавляем информацию о брокере к тикеру
	ticker.BrokerID = integration.brokerID

	// Сохраняем тикер через StorageService
	ctx, cancel := context.WithTimeout(bis.ctx, 10*time.Second)
	defer cancel()

	if err := bis.storageService.SaveTicker(ctx, ticker); err != nil {
		return fmt.Errorf("failed to save ticker: %w", err)
	}

	// Обновляем статистику
	now := time.Now()

	integration.mu.Lock()
	integration.stats.TickersProcessed++
	integration.stats.LastDataReceived = now
	integration.mu.Unlock()

	bis.statsMu.Lock()
	bis.stats.TotalTickers++
	bis.stats.LastDataReceived = now
	bis.statsMu.Unlock()

	bis.logger.WithFields(logrus.Fields{
		"broker_id": integration.brokerID,
		"symbol":    ticker.Symbol,
		"price":     ticker.Price,
	}).Debug("Ticker processed and saved")

	return nil
}

// processCandle обрабатывает свечу и сохраняет её в хранилище
func (bis *BrokerStorageIntegration) processCandle(integration *BrokerIntegration, candle entities.Candle) error {
	// Добавляем информацию о брокере к свече
	candle.BrokerID = integration.brokerID

	// Сохраняем свечу через StorageService
	ctx, cancel := context.WithTimeout(bis.ctx, 10*time.Second)
	defer cancel()

	if err := bis.storageService.SaveCandle(ctx, candle); err != nil {
		return fmt.Errorf("failed to save candle: %w", err)
	}

	// Обновляем статистику
	now := time.Now()

	integration.mu.Lock()
	integration.stats.CandlesProcessed++
	integration.stats.LastDataReceived = now
	integration.mu.Unlock()

	bis.statsMu.Lock()
	bis.stats.TotalCandles++
	bis.stats.LastDataReceived = now
	bis.statsMu.Unlock()

	bis.logger.WithFields(logrus.Fields{
		"broker_id": integration.brokerID,
		"symbol":    candle.Symbol,
		"timeframe": candle.Timeframe,
		"open":      candle.Open,
		"close":     candle.Close,
	}).Debug("Candle processed and saved")

	return nil
}

// processOrderBook обрабатывает ордербук и сохраняет его в хранилище
func (bis *BrokerStorageIntegration) processOrderBook(integration *BrokerIntegration, orderBook entities.OrderBook) error {
	// Добавляем информацию о брокере к ордербуку
	orderBook.BrokerID = integration.brokerID

	// Сохраняем ордербук через StorageService
	ctx, cancel := context.WithTimeout(bis.ctx, 10*time.Second)
	defer cancel()

	if err := bis.storageService.SaveOrderBook(ctx, orderBook); err != nil {
		return fmt.Errorf("failed to save orderbook: %w", err)
	}

	// Обновляем статистику
	now := time.Now()

	integration.mu.Lock()
	integration.stats.OrderBooksProcessed++
	integration.stats.LastDataReceived = now
	integration.mu.Unlock()

	bis.statsMu.Lock()
	bis.stats.TotalOrderBooks++
	bis.stats.LastDataReceived = now
	bis.statsMu.Unlock()

	bis.logger.WithFields(logrus.Fields{
		"broker_id": integration.brokerID,
		"symbol":    orderBook.Symbol,
		"bids":      len(orderBook.Bids),
		"asks":      len(orderBook.Asks),
	}).Debug("OrderBook processed and saved")

	return nil
}
