package broker

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m-data-storage/internal/domain/interfaces"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory(logrus.New())
	assert.NotNil(t, factory)
}

func TestFactory_GetSupportedTypes(t *testing.T) {
	factory := NewFactory(logrus.New())
	
	types := factory.GetSupportedTypes()
	assert.Len(t, types, 2)
	assert.Contains(t, types, interfaces.BrokerTypeCrypto)
	assert.Contains(t, types, interfaces.BrokerTypeStock)
}

func TestFactory_CreateBroker_Crypto(t *testing.T) {
	factory := NewFactory(logrus.New())
	
	config := interfaces.BrokerConfig{
		ID:      "crypto-broker",
		Name:    "Crypto Broker",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: true,
		Connection: interfaces.ConnectionConfig{
			WebSocketURL: "wss://crypto.com",
			RestAPIURL:   "https://api.crypto.com",
			Timeout:      30 * time.Second,
		},
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
		Limits: interfaces.LimitsConfig{
			MaxSubscriptions: 50,
		},
	}
	
	broker, err := factory.CreateBroker(config)
	require.NoError(t, err)
	assert.NotNil(t, broker)
	
	info := broker.GetInfo()
	assert.Equal(t, "crypto-broker", info.ID)
	assert.Equal(t, "Crypto Broker", info.Name)
	assert.Equal(t, interfaces.BrokerTypeCrypto, info.Type)
}

func TestFactory_CreateBroker_Stock(t *testing.T) {
	factory := NewFactory(logrus.New())
	
	config := interfaces.BrokerConfig{
		ID:      "stock-broker",
		Name:    "Stock Broker",
		Type:    interfaces.BrokerTypeStock,
		Enabled: true,
		Connection: interfaces.ConnectionConfig{
			WebSocketURL: "wss://stock.com",
			RestAPIURL:   "https://api.stock.com",
			Timeout:      30 * time.Second,
		},
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
		Limits: interfaces.LimitsConfig{
			MaxSubscriptions: 50,
		},
	}
	
	broker, err := factory.CreateBroker(config)
	require.NoError(t, err)
	assert.NotNil(t, broker)
	
	info := broker.GetInfo()
	assert.Equal(t, "stock-broker", info.ID)
	assert.Equal(t, "Stock Broker", info.Name)
	assert.Equal(t, interfaces.BrokerTypeStock, info.Type)
}

func TestFactory_CreateBroker_Disabled(t *testing.T) {
	factory := NewFactory(logrus.New())
	
	config := interfaces.BrokerConfig{
		ID:      "disabled-broker",
		Name:    "Disabled Broker",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: false, // Брокер отключен
		Connection: interfaces.ConnectionConfig{
			WebSocketURL: "wss://test.com",
		},
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
			BatchSize:  10,
		},
		Limits: interfaces.LimitsConfig{
			MaxSubscriptions: 50,
		},
	}
	
	broker, err := factory.CreateBroker(config)
	assert.Error(t, err)
	assert.Nil(t, broker)
	assert.Contains(t, err.Error(), "disabled")
}

func TestFactory_CreateBroker_UnsupportedType(t *testing.T) {
	factory := NewFactory(logrus.New())
	
	config := interfaces.BrokerConfig{
		ID:      "unknown-broker",
		Name:    "Unknown Broker",
		Type:    "unknown", // Неподдерживаемый тип
		Enabled: true,
		Connection: interfaces.ConnectionConfig{
			WebSocketURL: "wss://test.com",
		},
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
			BatchSize:  10,
		},
		Limits: interfaces.LimitsConfig{
			MaxSubscriptions: 50,
		},
	}
	
	broker, err := factory.CreateBroker(config)
	assert.Error(t, err)
	assert.Nil(t, broker)
	assert.Contains(t, err.Error(), "unsupported broker type")
}

func TestFactory_ValidateConfig(t *testing.T) {
	factory := NewFactory(logrus.New())
	
	tests := []struct {
		name        string
		config      interfaces.BrokerConfig
		expectError bool
		errorText   string
	}{
		{
			name: "valid config",
			config: interfaces.BrokerConfig{
				ID:      "test-broker",
				Name:    "Test Broker",
				Type:    interfaces.BrokerTypeCrypto,
				Enabled: true,
				Connection: interfaces.ConnectionConfig{
					WebSocketURL: "wss://test.com",
				},
				Defaults: interfaces.DefaultsConfig{
					BufferSize: 100,
					BatchSize:  10,
				},
				Limits: interfaces.LimitsConfig{
					MaxSubscriptions: 50,
				},
			},
			expectError: false,
		},
		{
			name: "empty ID",
			config: interfaces.BrokerConfig{
				Name: "Test Broker",
				Type: interfaces.BrokerTypeCrypto,
			},
			expectError: true,
			errorText:   "ID cannot be empty",
		},
		{
			name: "empty name",
			config: interfaces.BrokerConfig{
				ID:   "test-broker",
				Type: interfaces.BrokerTypeCrypto,
			},
			expectError: true,
			errorText:   "name cannot be empty",
		},
		{
			name: "empty type",
			config: interfaces.BrokerConfig{
				ID:   "test-broker",
				Name: "Test Broker",
			},
			expectError: true,
			errorText:   "type cannot be empty",
		},
		{
			name: "unsupported type",
			config: interfaces.BrokerConfig{
				ID:   "test-broker",
				Name: "Test Broker",
				Type: "unsupported",
			},
			expectError: true,
			errorText:   "unsupported broker type",
		},
		{
			name: "no connection URLs",
			config: interfaces.BrokerConfig{
				ID:   "test-broker",
				Name: "Test Broker",
				Type: interfaces.BrokerTypeCrypto,
				Connection: interfaces.ConnectionConfig{
					// Нет URL
				},
				Defaults: interfaces.DefaultsConfig{
					BufferSize: 100,
					BatchSize:  10,
				},
				Limits: interfaces.LimitsConfig{
					MaxSubscriptions: 50,
				},
			},
			expectError: true,
			errorText:   "connection URL must be specified",
		},
		{
			name: "invalid max subscriptions",
			config: interfaces.BrokerConfig{
				ID:   "test-broker",
				Name: "Test Broker",
				Type: interfaces.BrokerTypeCrypto,
				Connection: interfaces.ConnectionConfig{
					WebSocketURL: "wss://test.com",
				},
				Defaults: interfaces.DefaultsConfig{
					BufferSize: 100,
					BatchSize:  10,
				},
				Limits: interfaces.LimitsConfig{
					MaxSubscriptions: 0, // Неверное значение
				},
			},
			expectError: true,
			errorText:   "max subscriptions must be positive",
		},
		{
			name: "invalid buffer size",
			config: interfaces.BrokerConfig{
				ID:   "test-broker",
				Name: "Test Broker",
				Type: interfaces.BrokerTypeCrypto,
				Connection: interfaces.ConnectionConfig{
					WebSocketURL: "wss://test.com",
				},
				Defaults: interfaces.DefaultsConfig{
					BufferSize: 0, // Неверное значение
					BatchSize:  10,
				},
				Limits: interfaces.LimitsConfig{
					MaxSubscriptions: 50,
				},
			},
			expectError: true,
			errorText:   "buffer size must be positive",
		},
		{
			name: "invalid batch size",
			config: interfaces.BrokerConfig{
				ID:   "test-broker",
				Name: "Test Broker",
				Type: interfaces.BrokerTypeCrypto,
				Connection: interfaces.ConnectionConfig{
					WebSocketURL: "wss://test.com",
				},
				Defaults: interfaces.DefaultsConfig{
					BufferSize: 100,
					BatchSize:  0, // Неверное значение
				},
				Limits: interfaces.LimitsConfig{
					MaxSubscriptions: 50,
				},
			},
			expectError: true,
			errorText:   "batch size must be positive",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := factory.ValidateConfig(tt.config)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
