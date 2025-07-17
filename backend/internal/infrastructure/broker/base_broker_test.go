package broker

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

func TestNewBaseBroker(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	logger := logrus.New()
	broker := NewBaseBroker(config, logger)

	assert.NotNil(t, broker)
	assert.Equal(t, config.ID, broker.config.ID)
	assert.Equal(t, config.Name, broker.config.Name)
	assert.Equal(t, config.Type, broker.config.Type)
	assert.False(t, broker.IsConnected())
	assert.NotNil(t, broker.GetTickerChannel())
	assert.NotNil(t, broker.GetCandleChannel())
	assert.NotNil(t, broker.GetOrderBookChannel())
}

func TestBaseBroker_Connect(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	broker := NewBaseBroker(config, logrus.New())
	ctx := context.Background()

	// Check initial state
	assert.False(t, broker.IsConnected())

	// Connect
	err := broker.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, broker.IsConnected())

	// Repeated connection should be safe
	err = broker.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, broker.IsConnected())
}

func TestBaseBroker_Disconnect(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	broker := NewBaseBroker(config, logrus.New())
	ctx := context.Background()

	// Connect
	err := broker.Connect(ctx)
	require.NoError(t, err)
	assert.True(t, broker.IsConnected())

	// Disconnect
	err = broker.Disconnect()
	require.NoError(t, err)
	assert.False(t, broker.IsConnected())

	// Repeated disconnection should be safe
	err = broker.Disconnect()
	require.NoError(t, err)
	assert.False(t, broker.IsConnected())
}

func TestBaseBroker_GetInfo(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	broker := NewBaseBroker(config, logrus.New())

	info := broker.GetInfo()
	assert.Equal(t, config.ID, info.ID)
	assert.Equal(t, config.Name, info.Name)
	assert.Equal(t, config.Type, info.Type)
	assert.Equal(t, "disconnected", info.Status)

	// Connect and check status
	err := broker.Connect(context.Background())
	require.NoError(t, err)

	info = broker.GetInfo()
	assert.Equal(t, "connected", info.Status)
}

func TestBaseBroker_Subscribe(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	broker := NewBaseBroker(config, logrus.New())
	ctx := context.Background()

	subscriptions := []entities.InstrumentSubscription{
		{
			Symbol: "BTCUSDT",
			Type:   entities.InstrumentTypeSpot,
			Market: entities.MarketTypeSpot,
		},
		{
			Symbol: "ETHUSDT",
			Type:   entities.InstrumentTypeSpot,
			Market: entities.MarketTypeSpot,
		},
	}

	err := broker.Subscribe(ctx, subscriptions)
	require.NoError(t, err)

	// Check that subscriptions are added
	brokerSubscriptions := broker.GetSubscriptions()
	assert.Len(t, brokerSubscriptions, 2)

	stats := broker.GetStats()
	assert.Equal(t, 2, stats.ActiveSubscriptions)
}

func TestBaseBroker_Unsubscribe(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	broker := NewBaseBroker(config, logrus.New())
	ctx := context.Background()

	subscriptions := []entities.InstrumentSubscription{
		{
			Symbol: "BTCUSDT",
			Type:   entities.InstrumentTypeSpot,
			Market: entities.MarketTypeSpot,
		},
		{
			Symbol: "ETHUSDT",
			Type:   entities.InstrumentTypeSpot,
			Market: entities.MarketTypeSpot,
		},
	}

	// Subscribe
	err := broker.Subscribe(ctx, subscriptions)
	require.NoError(t, err)

	// Unsubscribe from one instrument
	err = broker.Unsubscribe(ctx, subscriptions[:1])
	require.NoError(t, err)

	// Check that one subscription remains
	brokerSubscriptions := broker.GetSubscriptions()
	assert.Len(t, brokerSubscriptions, 1)

	stats := broker.GetStats()
	assert.Equal(t, 1, stats.ActiveSubscriptions)
}

func TestBaseBroker_SendData(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 10,
		},
	}

	broker := NewBaseBroker(config, logrus.New())

	// Test ticker sending
	ticker := entities.Ticker{
		Symbol:    "BTCUSDT",
		Price:     50000.0,
		Volume:    1.5,
		Timestamp: time.Now(),
	}

	broker.SendTicker(ticker)

	// Check that ticker is received
	select {
	case receivedTicker := <-broker.GetTickerChannel():
		assert.Equal(t, ticker.Symbol, receivedTicker.Symbol)
		assert.Equal(t, ticker.Price, receivedTicker.Price)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Ticker not received")
	}

	// Check statistics
	stats := broker.GetStats()
	assert.Equal(t, int64(1), stats.TotalTickers)
	assert.False(t, stats.LastDataReceived.IsZero())
}

func TestBaseBroker_Health(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 10,
		},
	}

	broker := NewBaseBroker(config, logrus.New())

	// Check health of disconnected broker
	err := broker.Health()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")

	// Connect and check health
	err = broker.Connect(context.Background())
	require.NoError(t, err)

	err = broker.Health()
	assert.NoError(t, err)
}

func TestBaseBroker_StartStop(t *testing.T) {
	config := interfaces.BrokerConfig{
		ID:   "test-broker",
		Name: "Test Broker",
		Type: interfaces.BrokerTypeCrypto,
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
	}

	broker := NewBaseBroker(config, logrus.New())
	ctx := context.Background()

	// Start broker
	err := broker.Start(ctx)
	require.NoError(t, err)
	assert.True(t, broker.IsConnected())

	// Stop broker
	err = broker.Stop()
	require.NoError(t, err)
	assert.False(t, broker.IsConnected())
}
