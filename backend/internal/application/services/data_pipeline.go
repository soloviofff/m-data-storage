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

// DataPipelineService управляет потоками данных от брокеров к хранилищу
type DataPipelineService struct {
	brokerManager      interfaces.BrokerManager
	storageIntegration interfaces.BrokerStorageIntegration
	logger             *logrus.Logger

	// Управление жизненным циклом
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Конфигурация
	config DataPipelineConfig

	// Статистика
	stats PipelineStats
	mu    sync.RWMutex
}

// DataPipelineConfig содержит конфигурацию пайплайна данных
type DataPipelineConfig struct {
	// Автоматическое подключение брокеров
	AutoConnectBrokers bool `yaml:"auto_connect_brokers" json:"auto_connect_brokers"`

	// Интервал проверки здоровья
	HealthCheckInterval time.Duration `yaml:"health_check_interval" json:"health_check_interval"`

	// Таймаут для операций
	OperationTimeout time.Duration `yaml:"operation_timeout" json:"operation_timeout"`

	// Автоматическое переподключение
	AutoReconnect bool `yaml:"auto_reconnect" json:"auto_reconnect"`

	// Интервал переподключения
	ReconnectInterval time.Duration `yaml:"reconnect_interval" json:"reconnect_interval"`
}

// PipelineStats содержит статистику пайплайна
type PipelineStats = interfaces.DataPipelineStats

// DefaultDataPipelineConfig возвращает конфигурацию по умолчанию
func DefaultDataPipelineConfig() DataPipelineConfig {
	return DataPipelineConfig{
		AutoConnectBrokers:  true,
		HealthCheckInterval: 30 * time.Second,
		OperationTimeout:    10 * time.Second,
		AutoReconnect:       true,
		ReconnectInterval:   5 * time.Second,
	}
}

// NewDataPipelineService создает новый сервис пайплайна данных
func NewDataPipelineService(
	brokerManager interfaces.BrokerManager,
	storageIntegration interfaces.BrokerStorageIntegration,
	logger *logrus.Logger,
	config DataPipelineConfig,
) *DataPipelineService {
	if logger == nil {
		logger = logrus.New()
	}

	return &DataPipelineService{
		brokerManager:      brokerManager,
		storageIntegration: storageIntegration,
		logger:             logger,
		config:             config,
		stats: PipelineStats{
			StartedAt: time.Now(),
		},
	}
}

// Start запускает пайплайн данных
func (dps *DataPipelineService) Start(ctx context.Context) error {
	dps.ctx, dps.cancel = context.WithCancel(ctx)

	dps.logger.Info("Starting data pipeline service")

	// Запускаем интеграцию хранилища
	if err := dps.storageIntegration.Start(dps.ctx); err != nil {
		return fmt.Errorf("failed to start storage integration: %w", err)
	}

	// Подключаем существующие брокеры, если включено автоподключение
	if dps.config.AutoConnectBrokers {
		if err := dps.connectAllBrokers(); err != nil {
			dps.logger.WithError(err).Warn("Failed to connect some brokers during startup")
		}
	}

	// Запускаем фоновые задачи
	dps.wg.Add(1)
	go dps.healthCheckWorker()

	if dps.config.AutoReconnect {
		dps.wg.Add(1)
		go dps.reconnectWorker()
	}

	dps.updateStats(func(stats *PipelineStats) {
		stats.StartedAt = time.Now()
	})

	dps.logger.Info("Data pipeline service started successfully")
	return nil
}

// Stop останавливает пайплайн данных
func (dps *DataPipelineService) Stop() error {
	dps.logger.Info("Stopping data pipeline service")

	if dps.cancel != nil {
		dps.cancel()
	}

	// Останавливаем интеграцию хранилища
	if err := dps.storageIntegration.Stop(); err != nil {
		dps.logger.WithError(err).Error("Failed to stop storage integration")
	}

	// Ждем завершения фоновых задач
	dps.wg.Wait()

	dps.logger.Info("Data pipeline service stopped")
	return nil
}

// AddBroker добавляет брокер в пайплайн
func (dps *DataPipelineService) AddBroker(ctx context.Context, config interfaces.BrokerConfig) error {
	// Добавляем брокер в менеджер
	if err := dps.brokerManager.AddBroker(ctx, config); err != nil {
		return fmt.Errorf("failed to add broker to manager: %w", err)
	}

	// Получаем брокер из менеджера
	broker, err := dps.brokerManager.GetBroker(config.ID)
	if err != nil {
		return fmt.Errorf("failed to get broker from manager: %w", err)
	}

	// Подключаем брокер
	if err := dps.connectBroker(ctx, config.ID, broker); err != nil {
		dps.logger.WithError(err).WithField("broker_id", config.ID).Error("Failed to connect broker")
		return fmt.Errorf("failed to connect broker: %w", err)
	}

	dps.logger.WithField("broker_id", config.ID).Info("Broker added to pipeline")
	return nil
}

// RemoveBroker удаляет брокер из пайплайна
func (dps *DataPipelineService) RemoveBroker(ctx context.Context, brokerID string) error {
	// Удаляем брокер из интеграции хранилища
	if err := dps.storageIntegration.RemoveBroker(brokerID); err != nil {
		dps.logger.WithError(err).WithField("broker_id", brokerID).Warn("Failed to remove broker from storage integration")
	}

	// Удаляем брокер из менеджера
	if err := dps.brokerManager.RemoveBroker(ctx, brokerID); err != nil {
		return fmt.Errorf("failed to remove broker from manager: %w", err)
	}

	// Обновляем статистику
	dps.updateStats(func(stats *PipelineStats) {
		if stats.ConnectedBrokers > 0 {
			stats.ConnectedBrokers--
		}
	})

	dps.logger.WithField("broker_id", brokerID).Info("Broker removed from pipeline")
	return nil
}

// Subscribe подписывается на инструменты брокера
func (dps *DataPipelineService) Subscribe(ctx context.Context, brokerID string, subscriptions []entities.InstrumentSubscription) error {
	broker, err := dps.brokerManager.GetBroker(brokerID)
	if err != nil {
		return fmt.Errorf("failed to get broker: %w", err)
	}

	if err := broker.Subscribe(ctx, subscriptions); err != nil {
		return fmt.Errorf("failed to subscribe to instruments: %w", err)
	}

	dps.logger.WithFields(logrus.Fields{
		"broker_id":     brokerID,
		"subscriptions": len(subscriptions),
	}).Info("Subscribed to instruments")

	return nil
}

// Unsubscribe отписывается от инструментов брокера
func (dps *DataPipelineService) Unsubscribe(ctx context.Context, brokerID string, subscriptions []entities.InstrumentSubscription) error {
	broker, err := dps.brokerManager.GetBroker(brokerID)
	if err != nil {
		return fmt.Errorf("failed to get broker: %w", err)
	}

	if err := broker.Unsubscribe(ctx, subscriptions); err != nil {
		return fmt.Errorf("failed to unsubscribe from instruments: %w", err)
	}

	dps.logger.WithFields(logrus.Fields{
		"broker_id":     brokerID,
		"subscriptions": len(subscriptions),
	}).Info("Unsubscribed from instruments")

	return nil
}

// GetStats возвращает статистику пайплайна
func (dps *DataPipelineService) GetStats() PipelineStats {
	dps.mu.RLock()
	defer dps.mu.RUnlock()
	return dps.stats
}

// GetIntegrationStats возвращает статистику интеграции хранилища
func (dps *DataPipelineService) GetIntegrationStats() interfaces.BrokerStorageIntegrationStats {
	return dps.storageIntegration.GetStats()
}

// Health проверяет здоровье пайплайна
func (dps *DataPipelineService) Health() error {
	// Проверяем интеграцию хранилища
	if err := dps.storageIntegration.Health(); err != nil {
		return fmt.Errorf("storage integration health check failed: %w", err)
	}

	// Проверяем брокеры
	brokerHealth := dps.brokerManager.Health()
	for brokerID, err := range brokerHealth {
		if err != nil {
			return fmt.Errorf("broker %s health check failed: %w", brokerID, err)
		}
	}

	return nil
}

// connectAllBrokers подключает все брокеры из менеджера
func (dps *DataPipelineService) connectAllBrokers() error {
	brokers := dps.brokerManager.GetAllBrokers()

	var errors []error
	for brokerID, broker := range brokers {
		ctx, cancel := context.WithTimeout(dps.ctx, dps.config.OperationTimeout)
		if err := dps.connectBroker(ctx, brokerID, broker); err != nil {
			errors = append(errors, fmt.Errorf("broker %s: %w", brokerID, err))
		}
		cancel()
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to connect %d brokers: %v", len(errors), errors)
	}

	return nil
}

// connectBroker подключает отдельный брокер
func (dps *DataPipelineService) connectBroker(ctx context.Context, brokerID string, broker interfaces.Broker) error {
	// Подключаем брокер
	if err := broker.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect broker: %w", err)
	}

	// Запускаем брокер
	if err := broker.Start(ctx); err != nil {
		return fmt.Errorf("failed to start broker: %w", err)
	}

	// Добавляем брокер в интеграцию хранилища
	if err := dps.storageIntegration.AddBroker(brokerID, broker); err != nil {
		return fmt.Errorf("failed to add broker to storage integration: %w", err)
	}

	dps.updateStats(func(stats *PipelineStats) {
		stats.ConnectedBrokers++
	})

	dps.logger.WithField("broker_id", brokerID).Info("Broker connected successfully")
	return nil
}

// healthCheckWorker выполняет периодические проверки здоровья
func (dps *DataPipelineService) healthCheckWorker() {
	defer dps.wg.Done()

	ticker := time.NewTicker(dps.config.HealthCheckInterval)
	defer ticker.Stop()

	dps.logger.Debug("Health check worker started")

	for {
		select {
		case <-dps.ctx.Done():
			dps.logger.Debug("Health check worker stopped")
			return

		case <-ticker.C:
			dps.performHealthCheck()
		}
	}
}

// reconnectWorker выполняет переподключение отключенных брокеров
func (dps *DataPipelineService) reconnectWorker() {
	defer dps.wg.Done()

	ticker := time.NewTicker(dps.config.ReconnectInterval)
	defer ticker.Stop()

	dps.logger.Debug("Reconnect worker started")

	for {
		select {
		case <-dps.ctx.Done():
			dps.logger.Debug("Reconnect worker stopped")
			return

		case <-ticker.C:
			dps.performReconnect()
		}
	}
}

// performHealthCheck выполняет проверку здоровья
func (dps *DataPipelineService) performHealthCheck() {
	dps.logger.Debug("Performing health check")

	err := dps.Health()

	dps.updateStats(func(stats *PipelineStats) {
		stats.LastHealthCheck = time.Now()
		if err != nil {
			stats.HealthChecksFailed++
		} else {
			stats.HealthChecksPassed++
		}
	})

	if err != nil {
		dps.logger.WithError(err).Warn("Health check failed")
	} else {
		dps.logger.Debug("Health check passed")
	}
}

// performReconnect пытается переподключить отключенные брокеры
func (dps *DataPipelineService) performReconnect() {
	brokers := dps.brokerManager.GetAllBrokers()

	for brokerID, broker := range brokers {
		if !broker.IsConnected() {
			dps.logger.WithField("broker_id", brokerID).Info("Attempting to reconnect broker")

			ctx, cancel := context.WithTimeout(dps.ctx, dps.config.OperationTimeout)
			if err := dps.connectBroker(ctx, brokerID, broker); err != nil {
				dps.logger.WithError(err).WithField("broker_id", brokerID).Error("Failed to reconnect broker")
				dps.updateStats(func(stats *PipelineStats) {
					stats.TotalErrors++
				})
			}
			cancel()
		}
	}
}

// updateStats обновляет статистику пайплайна
func (dps *DataPipelineService) updateStats(updateFunc func(*PipelineStats)) {
	dps.mu.Lock()
	defer dps.mu.Unlock()
	updateFunc(&dps.stats)
}
