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
	
	err := manager.AddBroker(config)
	require.NoError(t, err)
	
	assert.Equal(t, 1, manager.GetBrokerCount())
	
	// Проверяем, что брокер можно получить
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
	
	// Добавляем первый раз
	err := manager.AddBroker(config)
	require.NoError(t, err)
	
	// Пытаемся добавить второй раз
	err = manager.AddBroker(config)
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
	
	// Добавляем брокер
	err := manager.AddBroker(config)
	require.NoError(t, err)
	assert.Equal(t, 1, manager.GetBrokerCount())
	
	// Удаляем брокер
	err = manager.RemoveBroker("test-broker")
	require.NoError(t, err)
	assert.Equal(t, 0, manager.GetBrokerCount())
	
	// Пытаемся получить удаленный брокер
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
	
	// Добавляем брокеры
	for _, config := range configs {
		err := manager.AddBroker(config)
		require.NoError(t, err)
	}
	
	// Получаем все брокеры
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
	
	// Добавляем брокер
	err := manager.AddBroker(config)
	require.NoError(t, err)
	
	// Запускаем все брокеры
	ctx := context.Background()
	err = manager.StartAll(ctx)
	require.NoError(t, err)
	
	// Проверяем, что брокер подключен
	broker, err := manager.GetBroker("test-broker")
	require.NoError(t, err)
	assert.True(t, broker.IsConnected())
	
	// Останавливаем все брокеры
	err = manager.StopAll()
	require.NoError(t, err)
	
	// Проверяем, что брокер отключен
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
	
	// Добавляем брокер
	err := manager.AddBroker(config)
	require.NoError(t, err)
	
	// Проверяем здоровье (брокер не подключен)
	health := manager.HealthCheck()
	assert.Len(t, health, 1)
	assert.Error(t, health["test-broker"])
	
	// Подключаем брокер
	ctx := context.Background()
	err = manager.StartAll(ctx)
	require.NoError(t, err)
	
	// Проверяем здоровье (брокер подключен)
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
	
	// Добавляем брокер
	err := manager.AddBroker(config)
	require.NoError(t, err)
	
	// Проверяем подключенные брокеры (пока нет)
	connected := manager.GetConnectedBrokers()
	assert.Len(t, connected, 0)
	
	// Подключаем брокер
	ctx := context.Background()
	err = manager.StartAll(ctx)
	require.NoError(t, err)
	
	// Проверяем подключенные брокеры
	connected = manager.GetConnectedBrokers()
	assert.Len(t, connected, 1)
	assert.Contains(t, connected, "test-broker")
}
