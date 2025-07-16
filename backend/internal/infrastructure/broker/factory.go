package broker

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/interfaces"
)

// Factory реализует интерфейс BrokerFactory
type Factory struct {
	logger *logrus.Logger
}

// NewFactory создает новую фабрику брокеров
func NewFactory(logger *logrus.Logger) *Factory {
	if logger == nil {
		logger = logrus.New()
	}

	return &Factory{
		logger: logger,
	}
}

// CreateBroker создает брокер на основе конфигурации
func (f *Factory) CreateBroker(config interfaces.BrokerConfig) (interfaces.Broker, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("broker %s is disabled", config.ID)
	}

	switch config.Type {
	case interfaces.BrokerTypeCrypto:
		return f.createCryptoBroker(config)
	case interfaces.BrokerTypeStock:
		return f.createStockBroker(config)
	default:
		return nil, fmt.Errorf("unsupported broker type: %s", config.Type)
	}
}

// GetSupportedTypes возвращает поддерживаемые типы брокеров
func (f *Factory) GetSupportedTypes() []interfaces.BrokerType {
	return []interfaces.BrokerType{
		interfaces.BrokerTypeCrypto,
		interfaces.BrokerTypeStock,
	}
}

// createCryptoBroker создает криптоброкер
func (f *Factory) createCryptoBroker(config interfaces.BrokerConfig) (interfaces.Broker, error) {
	f.logger.WithFields(logrus.Fields{
		"broker_id":   config.ID,
		"broker_name": config.Name,
		"broker_type": config.Type,
	}).Info("Creating crypto broker")

	// Проверяем, является ли это тестовым брокером
	if config.Name == "mock" || config.Name == "test" {
		return NewMockCryptoBroker(config, f.logger), nil
	}

	// В будущем здесь будет создание конкретных реализаций (Binance, Coinbase, etc.)
	// Пока возвращаем mock брокер для всех случаев
	return NewMockCryptoBroker(config, f.logger), nil
}

// createStockBroker создает фондовый брокер
func (f *Factory) createStockBroker(config interfaces.BrokerConfig) (interfaces.Broker, error) {
	f.logger.WithFields(logrus.Fields{
		"broker_id":   config.ID,
		"broker_name": config.Name,
		"broker_type": config.Type,
	}).Info("Creating stock broker")

	// Проверяем, является ли это тестовым брокером
	if config.Name == "mock" || config.Name == "test" {
		return NewMockStockBroker(config, f.logger), nil
	}

	// В будущем здесь будет создание конкретных реализаций (IEX, Alpha Vantage, etc.)
	// Пока возвращаем mock брокер для всех случаев
	return NewMockStockBroker(config, f.logger), nil
}

// ValidateConfig проверяет корректность конфигурации брокера
func (f *Factory) ValidateConfig(config interfaces.BrokerConfig) error {
	if config.ID == "" {
		return fmt.Errorf("broker ID cannot be empty")
	}

	if config.Name == "" {
		return fmt.Errorf("broker name cannot be empty")
	}

	if config.Type == "" {
		return fmt.Errorf("broker type cannot be empty")
	}

	// Проверяем, что тип поддерживается
	supportedTypes := f.GetSupportedTypes()
	supported := false
	for _, supportedType := range supportedTypes {
		if config.Type == supportedType {
			supported = true
			break
		}
	}

	if !supported {
		return fmt.Errorf("unsupported broker type: %s", config.Type)
	}

	// Проверяем конфигурацию соединения
	if config.Connection.WebSocketURL == "" && config.Connection.RestAPIURL == "" {
		return fmt.Errorf("at least one connection URL must be specified")
	}

	// Проверяем лимиты
	if config.Limits.MaxSubscriptions <= 0 {
		return fmt.Errorf("max subscriptions must be positive")
	}

	if config.Defaults.BufferSize <= 0 {
		return fmt.Errorf("buffer size must be positive")
	}

	if config.Defaults.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive")
	}

	return nil
}
