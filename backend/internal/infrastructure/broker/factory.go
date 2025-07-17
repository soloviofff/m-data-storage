package broker

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/interfaces"
)

// Factory implements BrokerFactory interface
type Factory struct {
	logger *logrus.Logger
}

// NewFactory creates a new broker factory
func NewFactory(logger *logrus.Logger) *Factory {
	if logger == nil {
		logger = logrus.New()
	}

	return &Factory{
		logger: logger,
	}
}

// CreateBroker creates a broker based on configuration
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

// GetSupportedTypes returns supported broker types
func (f *Factory) GetSupportedTypes() []interfaces.BrokerType {
	return []interfaces.BrokerType{
		interfaces.BrokerTypeCrypto,
		interfaces.BrokerTypeStock,
	}
}

// createCryptoBroker creates a crypto broker
func (f *Factory) createCryptoBroker(config interfaces.BrokerConfig) (interfaces.Broker, error) {
	f.logger.WithFields(logrus.Fields{
		"broker_id":   config.ID,
		"broker_name": config.Name,
		"broker_type": config.Type,
	}).Info("Creating crypto broker")

	// Check if this is a test broker
	if config.Name == "mock" || config.Name == "test" {
		return NewMockCryptoBroker(config, f.logger), nil
	}

	// In the future, specific implementations will be created here (Binance, Coinbase, etc.)
	// For now, return mock broker for all cases
	return NewMockCryptoBroker(config, f.logger), nil
}

// createStockBroker creates a stock broker
func (f *Factory) createStockBroker(config interfaces.BrokerConfig) (interfaces.Broker, error) {
	f.logger.WithFields(logrus.Fields{
		"broker_id":   config.ID,
		"broker_name": config.Name,
		"broker_type": config.Type,
	}).Info("Creating stock broker")

	// Check if this is a test broker
	if config.Name == "mock" || config.Name == "test" {
		return NewMockStockBroker(config, f.logger), nil
	}

	// In the future, specific implementations will be created here (IEX, Alpha Vantage, etc.)
	// For now, return mock broker for all cases
	return NewMockStockBroker(config, f.logger), nil
}

// ValidateConfig validates broker configuration
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

	// Check that type is supported
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

	// Check connection configuration
	if config.Connection.WebSocketURL == "" && config.Connection.RestAPIURL == "" {
		return fmt.Errorf("at least one connection URL must be specified")
	}

	// Check limits
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
