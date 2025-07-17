package broker

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m-data-storage/internal/domain/interfaces"
)

func TestNewManager(t *testing.T) {
	factory := NewFactory(logrus.New())
	manager := NewManager(factory, logrus.New())

	assert.NotNil(t, manager)
	assert.Equal(t, 0, manager.GetBrokerCount())
}

func TestManager_AddBroker(t *testing.T) {
	factory := NewFactory(logrus.New())
	manager := NewManager(factory, logrus.New())

	config := interfaces.BrokerConfig{
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
			MaxSubscriptions: 100,
		},
	}

	ctx := context.Background()
	err := manager.AddBroker(ctx, config)
	require.NoError(t, err)

	assert.Equal(t, 1, manager.GetBrokerCount())

	// Check that broker can be retrieved
	broker, err := manager.GetBroker("test-broker")
	require.NoError(t, err)
	assert.NotNil(t, broker)

	info := broker.GetInfo()
	assert.Equal(t, "test-broker", info.ID)
	assert.Equal(t, "Test Broker", info.Name)
}

func TestManager_AddBroker_Duplicate(t *testing.T) {
	factory := NewFactory(logrus.New())
	manager := NewManager(factory, logrus.New())

	config := interfaces.BrokerConfig{
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
			MaxSubscriptions: 100,
		},
	}

	ctx := context.Background()

	// Add first time
	err := manager.AddBroker(ctx, config)
	require.NoError(t, err)

	// Try to add second time
	err = manager.AddBroker(ctx, config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestManager_RemoveBroker(t *testing.T) {
	factory := NewFactory(logrus.New())
	manager := NewManager(factory, logrus.New())

	config := interfaces.BrokerConfig{
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
			MaxSubscriptions: 100,
		},
	}

	ctx := context.Background()

	// Add broker
	err := manager.AddBroker(ctx, config)
	require.NoError(t, err)
	assert.Equal(t, 1, manager.GetBrokerCount())

	// Remove broker
	err = manager.RemoveBroker(ctx, "test-broker")
	require.NoError(t, err)
	assert.Equal(t, 0, manager.GetBrokerCount())

	// Try to get removed broker
	_, err = manager.GetBroker("test-broker")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_GetAllBrokers(t *testing.T) {
	factory := NewFactory(logrus.New())
	manager := NewManager(factory, logrus.New())

	configs := []interfaces.BrokerConfig{
		{
			ID:      "broker1",
			Name:    "Broker 1",
			Type:    interfaces.BrokerTypeCrypto,
			Enabled: true,
			Connection: interfaces.ConnectionConfig{
				WebSocketURL: "wss://test1.com",
			},
			Defaults: interfaces.DefaultsConfig{
				BufferSize: 100,
				BatchSize:  10,
			},
			Limits: interfaces.LimitsConfig{
				MaxSubscriptions: 100,
			},
		},
		{
			ID:      "broker2",
			Name:    "Broker 2",
			Type:    interfaces.BrokerTypeStock,
			Enabled: true,
			Connection: interfaces.ConnectionConfig{
				WebSocketURL: "wss://test2.com",
			},
			Defaults: interfaces.DefaultsConfig{
				BufferSize: 100,
				BatchSize:  10,
			},
			Limits: interfaces.LimitsConfig{
				MaxSubscriptions: 100,
			},
		},
	}

	ctx := context.Background()

	// Add brokers
	for _, config := range configs {
		err := manager.AddBroker(ctx, config)
		require.NoError(t, err)
	}

	// Get all brokers
	brokers := manager.GetAllBrokers()
	assert.Len(t, brokers, 2)
	assert.Contains(t, brokers, "broker1")
	assert.Contains(t, brokers, "broker2")
}

func TestManager_StartStopAll(t *testing.T) {
	factory := NewFactory(logrus.New())
	manager := NewManager(factory, logrus.New())

	config := interfaces.BrokerConfig{
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
			MaxSubscriptions: 100,
		},
	}

	ctx := context.Background()

	// Add broker
	err := manager.AddBroker(ctx, config)
	require.NoError(t, err)

	// Start all brokers
	err = manager.StartAll(ctx)
	require.NoError(t, err)

	// Check that broker is connected
	broker, err := manager.GetBroker("test-broker")
	require.NoError(t, err)
	assert.True(t, broker.IsConnected())

	// Stop all brokers
	err = manager.StopAll()
	require.NoError(t, err)

	// Check that broker is disconnected
	assert.False(t, broker.IsConnected())
}

func TestManager_HealthCheck(t *testing.T) {
	factory := NewFactory(logrus.New())
	manager := NewManager(factory, logrus.New())

	config := interfaces.BrokerConfig{
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
			MaxSubscriptions: 100,
		},
	}

	ctx := context.Background()

	// Add broker
	err := manager.AddBroker(ctx, config)
	require.NoError(t, err)

	// Check health (broker not connected)
	health := manager.HealthCheck()
	assert.Len(t, health, 1)
	assert.Error(t, health["test-broker"])

	// Connect broker
	err = manager.StartAll(ctx)
	require.NoError(t, err)

	// Check health (broker connected)
	health = manager.HealthCheck()
	assert.Len(t, health, 1)
	assert.NoError(t, health["test-broker"])
}

func TestManager_GetConnectedBrokers(t *testing.T) {
	factory := NewFactory(logrus.New())
	manager := NewManager(factory, logrus.New())

	config := interfaces.BrokerConfig{
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
			MaxSubscriptions: 100,
		},
	}

	ctx := context.Background()

	// Add broker
	err := manager.AddBroker(ctx, config)
	require.NoError(t, err)

	// Check connected brokers (none yet)
	connected := manager.GetConnectedBrokers()
	assert.Len(t, connected, 0)

	// Connect broker
	err = manager.StartAll(ctx)
	require.NoError(t, err)

	// Check connected brokers
	connected = manager.GetConnectedBrokers()
	assert.Len(t, connected, 1)
	assert.Contains(t, connected, "test-broker")
}
