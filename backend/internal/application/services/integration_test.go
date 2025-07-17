package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
	"m-data-storage/internal/infrastructure/broker"
)

// TestFullDataPipeline tests the full system workflow:
// Configuration -> Broker creation -> Storage connection -> Data processing -> Saving
func TestFullDataPipeline(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce logging for tests

	// 1. Create system components

	// Create storage mock
	mockStorageService := &MockStorageService{}
	mockStorageService.On("SaveTicker", mock.Anything, mock.AnythingOfType("entities.Ticker")).Return(nil)
	mockStorageService.On("SaveCandle", mock.Anything, mock.AnythingOfType("entities.Candle")).Return(nil)
	mockStorageService.On("SaveOrderBook", mock.Anything, mock.AnythingOfType("entities.OrderBook")).Return(nil).Maybe()

	// Create broker factory and manager
	factory := broker.NewFactory(logger)
	brokerManager := broker.NewManager(factory, logger)

	// Create storage integration
	storageIntegration := NewBrokerStorageIntegration(brokerManager, mockStorageService, logger)

	// Create data pipeline
	pipelineConfig := DefaultDataPipelineConfig()
	pipelineConfig.AutoConnectBrokers = false            // Manage connections manually
	pipelineConfig.HealthCheckInterval = 1 * time.Second // Speed up for tests

	pipeline := NewDataPipelineService(brokerManager, storageIntegration, logger, pipelineConfig)

	// 2. Start pipeline
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := pipeline.Start(ctx)
	require.NoError(t, err)
	defer pipeline.Stop()

	// 3. Add crypto broker
	cryptoConfig := interfaces.BrokerConfig{
		ID:      "crypto-test",
		Name:    "Crypto Test Broker",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: true,
		Connection: interfaces.ConnectionConfig{
			WebSocketURL: "wss://test-crypto.com",
			RestAPIURL:   "https://api.test-crypto.com",
			Timeout:      5 * time.Second,
		},
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
		Limits: interfaces.LimitsConfig{
			MaxSubscriptions: 100,
		},
	}

	err = pipeline.AddBroker(ctx, cryptoConfig)
	require.NoError(t, err)

	// 4. Add stock broker
	stockConfig := interfaces.BrokerConfig{
		ID:      "stock-test",
		Name:    "Stock Test Broker",
		Type:    interfaces.BrokerTypeStock,
		Enabled: true,
		Connection: interfaces.ConnectionConfig{
			WebSocketURL: "wss://test-stock.com",
			RestAPIURL:   "https://api.test-stock.com",
			Timeout:      5 * time.Second,
		},
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 1000,
			BatchSize:  100,
		},
		Limits: interfaces.LimitsConfig{
			MaxSubscriptions: 50,
		},
	}

	err = pipeline.AddBroker(ctx, stockConfig)
	require.NoError(t, err)

	// 5. Subscribe to instruments
	cryptoSubscriptions := []entities.InstrumentSubscription{
		{
			Symbol: "BTCUSDT",
			Type:   entities.InstrumentTypeSpot,
			Market: entities.MarketTypeSpot,
		},
		{
			Symbol: "ETHUSDT",
			Type:   entities.InstrumentTypeFutures,
			Market: entities.MarketTypeFutures,
		},
	}

	err = pipeline.Subscribe(ctx, "crypto-test", cryptoSubscriptions)
	require.NoError(t, err)

	stockSubscriptions := []entities.InstrumentSubscription{
		{
			Symbol: "AAPL",
			Type:   entities.InstrumentTypeStock,
			Market: entities.MarketTypeStock,
		},
		{
			Symbol: "MSFT",
			Type:   entities.InstrumentTypeStock,
			Market: entities.MarketTypeStock,
		},
	}

	err = pipeline.Subscribe(ctx, "stock-test", stockSubscriptions)
	require.NoError(t, err)

	// 6. Wait for data generation and processing
	time.Sleep(3 * time.Second)

	// 7. Check statistics
	stats := pipeline.GetStats()
	assert.False(t, stats.StartedAt.IsZero())
	assert.Equal(t, 2, stats.ConnectedBrokers)

	integrationStats := pipeline.GetIntegrationStats()
	assert.Equal(t, 2, integrationStats.ActiveBrokers)
	assert.True(t, integrationStats.TotalTickers > 0)
	assert.True(t, integrationStats.TotalCandles > 0)

	// 8. Check system health
	err = pipeline.Health()
	assert.NoError(t, err)

	// 9. Check that data is being saved
	mockStorageService.AssertExpectations(t)

	// 10. Unsubscribe from some instruments
	err = pipeline.Unsubscribe(ctx, "crypto-test", cryptoSubscriptions[:1])
	require.NoError(t, err)

	// 11. Remove one broker
	err = pipeline.RemoveBroker(ctx, "stock-test")
	require.NoError(t, err)

	// 12. Check updated statistics
	finalStats := pipeline.GetStats()
	assert.Equal(t, 1, finalStats.ConnectedBrokers)

	finalIntegrationStats := pipeline.GetIntegrationStats()
	assert.Equal(t, 1, finalIntegrationStats.ActiveBrokers)
}

// TestBrokerFailureRecovery tests recovery after broker failures
func TestBrokerFailureRecovery(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	// Create components
	mockStorageService := &MockStorageService{}
	mockStorageService.On("SaveTicker", mock.Anything, mock.AnythingOfType("entities.Ticker")).Return(nil)
	mockStorageService.On("SaveCandle", mock.Anything, mock.AnythingOfType("entities.Candle")).Return(nil)
	mockStorageService.On("SaveOrderBook", mock.Anything, mock.AnythingOfType("entities.OrderBook")).Return(nil).Maybe()

	factory := broker.NewFactory(logger)
	brokerManager := broker.NewManager(factory, logger)
	storageIntegration := NewBrokerStorageIntegration(brokerManager, mockStorageService, logger)

	pipelineConfig := DefaultDataPipelineConfig()
	pipelineConfig.AutoReconnect = true
	pipelineConfig.ReconnectInterval = 1 * time.Second

	pipeline := NewDataPipelineService(brokerManager, storageIntegration, logger, pipelineConfig)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Start pipeline
	err := pipeline.Start(ctx)
	require.NoError(t, err)
	defer pipeline.Stop()

	// Add broker
	config := interfaces.BrokerConfig{
		ID:      "test-broker",
		Name:    "Test Broker",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: true,
		Connection: interfaces.ConnectionConfig{
			WebSocketURL: "wss://test.com",
			Timeout:      5 * time.Second,
		},
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
		Limits: interfaces.LimitsConfig{
			MaxSubscriptions: 50,
		},
	}

	err = pipeline.AddBroker(ctx, config)
	require.NoError(t, err)

	// Check that broker is working
	time.Sleep(1 * time.Second)
	err = pipeline.Health()
	assert.NoError(t, err)

	// Simulate broker failure (remove and add again)
	err = pipeline.RemoveBroker(ctx, "test-broker")
	require.NoError(t, err)

	// Wait for complete removal
	time.Sleep(2 * time.Second)

	// Create new configuration with different ID to avoid conflicts
	recoveryConfig := interfaces.BrokerConfig{
		ID:      "test-broker-recovery",
		Name:    "Test Broker Recovery",
		Type:    interfaces.BrokerTypeCrypto,
		Enabled: true,
		Connection: interfaces.ConnectionConfig{
			WebSocketURL: "wss://test.com",
			Timeout:      5 * time.Second,
		},
		Defaults: interfaces.DefaultsConfig{
			BufferSize: 100,
		},
		Limits: interfaces.LimitsConfig{
			MaxSubscriptions: 50,
		},
	}

	err = pipeline.AddBroker(ctx, recoveryConfig)
	require.NoError(t, err)

	// Check recovery
	time.Sleep(1 * time.Second)
	err = pipeline.Health()
	assert.NoError(t, err)

	stats := pipeline.GetStats()
	assert.Equal(t, 1, stats.ConnectedBrokers)
}

// TestConcurrentOperations tests concurrent operations
func TestConcurrentOperations(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	// Create components
	mockStorageService := &MockStorageService{}
	mockStorageService.On("SaveTicker", mock.Anything, mock.AnythingOfType("entities.Ticker")).Return(nil)
	mockStorageService.On("SaveCandle", mock.Anything, mock.AnythingOfType("entities.Candle")).Return(nil)
	mockStorageService.On("SaveOrderBook", mock.Anything, mock.AnythingOfType("entities.OrderBook")).Return(nil).Maybe()

	factory := broker.NewFactory(logger)
	brokerManager := broker.NewManager(factory, logger)
	storageIntegration := NewBrokerStorageIntegration(brokerManager, mockStorageService, logger)

	pipeline := NewDataPipelineService(brokerManager, storageIntegration, logger, DefaultDataPipelineConfig())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start pipeline
	err := pipeline.Start(ctx)
	require.NoError(t, err)
	defer pipeline.Stop()

	// Concurrently add multiple brokers
	const numBrokers = 5
	done := make(chan bool, numBrokers)

	for i := 0; i < numBrokers; i++ {
		go func(id int) {
			defer func() { done <- true }()

			config := interfaces.BrokerConfig{
				ID:      fmt.Sprintf("broker-%d", id),
				Name:    fmt.Sprintf("Broker %d", id),
				Type:    interfaces.BrokerTypeCrypto,
				Enabled: true,
				Connection: interfaces.ConnectionConfig{
					WebSocketURL: "wss://test.com",
					Timeout:      5 * time.Second,
				},
				Defaults: interfaces.DefaultsConfig{
					BufferSize: 100,
				},
				Limits: interfaces.LimitsConfig{
					MaxSubscriptions: 50,
				},
			}

			err := pipeline.AddBroker(ctx, config)
			assert.NoError(t, err)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numBrokers; i++ {
		<-done
	}

	// Check that all brokers are added
	stats := pipeline.GetStats()
	assert.Equal(t, numBrokers, stats.ConnectedBrokers)

	// Check health
	err = pipeline.Health()
	assert.NoError(t, err)
}
