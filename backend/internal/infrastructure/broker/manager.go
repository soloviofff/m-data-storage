package broker

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/interfaces"
)

// Manager реализует интерфейс BrokerManager
type Manager struct {
	brokers map[string]interfaces.Broker
	factory interfaces.BrokerFactory
	logger  *logrus.Logger
	mu      sync.RWMutex
}

// NewManager создает новый менеджер брокеров
func NewManager(factory interfaces.BrokerFactory, logger *logrus.Logger) *Manager {
	if logger == nil {
		logger = logrus.New()
	}

	return &Manager{
		brokers: make(map[string]interfaces.Broker),
		factory: factory,
		logger:  logger,
	}
}

// Initialize инициализирует менеджер брокеров
func (m *Manager) Initialize(ctx context.Context) error {
	m.logger.Info("Initializing broker manager")
	return nil
}

// AddBroker добавляет брокер в менеджер
func (m *Manager) AddBroker(ctx context.Context, config interfaces.BrokerConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.brokers[config.ID]; exists {
		return fmt.Errorf("broker with ID %s already exists", config.ID)
	}

	broker, err := m.factory.CreateBroker(config)
	if err != nil {
		return fmt.Errorf("failed to create broker %s: %w", config.ID, err)
	}

	m.brokers[config.ID] = broker

	m.logger.WithFields(logrus.Fields{
		"broker_id":   config.ID,
		"broker_name": config.Name,
		"broker_type": config.Type,
	}).Info("Broker added to manager")

	return nil
}

// RemoveBroker удаляет брокер из менеджера
func (m *Manager) RemoveBroker(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	broker, exists := m.brokers[id]
	if !exists {
		return fmt.Errorf("broker with ID %s not found", id)
	}

	// Останавливаем брокер перед удалением
	if err := broker.Stop(); err != nil {
		m.logger.WithField("broker_id", id).WithError(err).Warn("Error stopping broker")
	}

	delete(m.brokers, id)

	m.logger.WithField("broker_id", id).Info("Broker removed from manager")
	return nil
}

// GetBroker возвращает брокер по ID
func (m *Manager) GetBroker(id string) (interfaces.Broker, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	broker, exists := m.brokers[id]
	if !exists {
		return nil, fmt.Errorf("broker with ID %s not found", id)
	}

	return broker, nil
}

// GetAllBrokers возвращает все брокеры
func (m *Manager) GetAllBrokers() map[string]interfaces.Broker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]interfaces.Broker)
	for id, broker := range m.brokers {
		result[id] = broker
	}

	return result
}

// StartAll запускает все брокеры
func (m *Manager) StartAll(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errors []error

	for id, broker := range m.brokers {
		if err := broker.Start(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to start broker %s: %w", id, err))
			m.logger.WithField("broker_id", id).WithError(err).Error("Failed to start broker")
		} else {
			m.logger.WithField("broker_id", id).Info("Broker started successfully")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to start %d brokers: %v", len(errors), errors)
	}

	m.logger.Info("All brokers started successfully")
	return nil
}

// StopAll останавливает все брокеры
func (m *Manager) StopAll() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var errors []error

	for id, broker := range m.brokers {
		if err := broker.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop broker %s: %w", id, err))
			m.logger.WithField("broker_id", id).WithError(err).Error("Failed to stop broker")
		} else {
			m.logger.WithField("broker_id", id).Info("Broker stopped successfully")
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop %d brokers: %v", len(errors), errors)
	}

	m.logger.Info("All brokers stopped successfully")
	return nil
}

// Health проверяет здоровье всех брокеров
func (m *Manager) Health() map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	health := make(map[string]error)

	for id, broker := range m.brokers {
		health[id] = broker.Health()
	}

	return health
}

// HealthCheck проверяет здоровье всех брокеров (deprecated: use Health() instead)
func (m *Manager) HealthCheck() map[string]error {
	return m.Health()
}

// ListBrokers возвращает список информации о брокерах
func (m *Manager) ListBrokers() []interfaces.BrokerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var brokers []interfaces.BrokerInfo
	for _, broker := range m.brokers {
		brokers = append(brokers, broker.GetInfo())
	}

	return brokers
}

// GetBrokerCount возвращает количество брокеров
func (m *Manager) GetBrokerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.brokers)
}

// GetConnectedBrokers возвращает список подключенных брокеров
func (m *Manager) GetConnectedBrokers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var connected []string
	for id, broker := range m.brokers {
		if broker.IsConnected() {
			connected = append(connected, id)
		}
	}

	return connected
}

// GetBrokerInfo возвращает информацию о всех брокерах
func (m *Manager) GetBrokerInfo() map[string]interfaces.BrokerInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info := make(map[string]interfaces.BrokerInfo)
	for id, broker := range m.brokers {
		info[id] = broker.GetInfo()
	}

	return info
}

// Shutdown корректно завершает работу менеджера
func (m *Manager) Shutdown() error {
	m.logger.Info("Shutting down broker manager")

	if err := m.StopAll(); err != nil {
		m.logger.WithError(err).Error("Error stopping brokers during shutdown")
		return err
	}

	m.logger.Info("Broker manager shutdown completed")
	return nil
}
