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

// TestFullDataPipeline тестирует полный цикл работы системы:
// Конфигурация -> Создание брокеров -> Подключение к хранилищу -> Обработка данных -> Сохранение
func TestFullDataPipeline(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Уменьшаем логирование для тестов

	// 1. Создаем компоненты системы

	// Создаем мок хранилища
	mockStorageService := &MockStorageService{}
	mockStorageService.On("SaveTicker", mock.Anything, mock.AnythingOfType("entities.Ticker")).Return(nil)
	mockStorageService.On("SaveCandle", mock.Anything, mock.AnythingOfType("entities.Candle")).Return(nil)
	mockStorageService.On("SaveOrderBook", mock.Anything, mock.AnythingOfType("entities.OrderBook")).Return(nil).Maybe()

	// Создаем фабрику и менеджер брокеров
	factory := broker.NewFactory(logger)
	brokerManager := broker.NewManager(factory, logger)

	// Создаем интеграцию хранилища
	storageIntegration := NewBrokerStorageIntegration(brokerManager, mockStorageService, logger)

	// Создаем пайплайн данных
	pipelineConfig := DefaultDataPipelineConfig()
	pipelineConfig.AutoConnectBrokers = false            // Управляем подключением вручную
	pipelineConfig.HealthCheckInterval = 1 * time.Second // Ускоряем для тестов

	pipeline := NewDataPipelineService(brokerManager, storageIntegration, logger, pipelineConfig)

	// 2. Запускаем пайплайн
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := pipeline.Start(ctx)
	require.NoError(t, err)
	defer pipeline.Stop()

	// 3. Добавляем криптоброкер
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

	// 4. Добавляем фондовый брокер
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

	// 5. Подписываемся на инструменты
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

	// 6. Ждем генерации и обработки данных
	time.Sleep(3 * time.Second)

	// 7. Проверяем статистику
	stats := pipeline.GetStats()
	assert.False(t, stats.StartedAt.IsZero())
	assert.Equal(t, 2, stats.ConnectedBrokers)

	integrationStats := pipeline.GetIntegrationStats()
	assert.Equal(t, 2, integrationStats.ActiveBrokers)
	assert.True(t, integrationStats.TotalTickers > 0)
	assert.True(t, integrationStats.TotalCandles > 0)

	// 8. Проверяем здоровье системы
	err = pipeline.Health()
	assert.NoError(t, err)

	// 9. Проверяем, что данные сохраняются
	mockStorageService.AssertExpectations(t)

	// 10. Отписываемся от части инструментов
	err = pipeline.Unsubscribe(ctx, "crypto-test", cryptoSubscriptions[:1])
	require.NoError(t, err)

	// 11. Удаляем один брокер
	err = pipeline.RemoveBroker(ctx, "stock-test")
	require.NoError(t, err)

	// 12. Проверяем обновленную статистику
	finalStats := pipeline.GetStats()
	assert.Equal(t, 1, finalStats.ConnectedBrokers)

	finalIntegrationStats := pipeline.GetIntegrationStats()
	assert.Equal(t, 1, finalIntegrationStats.ActiveBrokers)
}

// TestBrokerFailureRecovery тестирует восстановление после сбоев брокера
func TestBrokerFailureRecovery(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	// Создаем компоненты
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

	// Запускаем пайплайн
	err := pipeline.Start(ctx)
	require.NoError(t, err)
	defer pipeline.Stop()

	// Добавляем брокер
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

	// Проверяем, что брокер работает
	time.Sleep(1 * time.Second)
	err = pipeline.Health()
	assert.NoError(t, err)

	// Симулируем сбой брокера (удаляем и добавляем снова)
	err = pipeline.RemoveBroker(ctx, "test-broker")
	require.NoError(t, err)

	// Ждем полного удаления
	time.Sleep(2 * time.Second)

	// Создаем новую конфигурацию с другим ID для избежания конфликтов
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

	// Проверяем восстановление
	time.Sleep(1 * time.Second)
	err = pipeline.Health()
	assert.NoError(t, err)

	stats := pipeline.GetStats()
	assert.Equal(t, 1, stats.ConnectedBrokers)
}

// TestConcurrentOperations тестирует конкурентные операции
func TestConcurrentOperations(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	// Создаем компоненты
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

	// Запускаем пайплайн
	err := pipeline.Start(ctx)
	require.NoError(t, err)
	defer pipeline.Stop()

	// Конкурентно добавляем несколько брокеров
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

	// Ждем завершения всех горутин
	for i := 0; i < numBrokers; i++ {
		<-done
	}

	// Проверяем, что все брокеры добавлены
	stats := pipeline.GetStats()
	assert.Equal(t, numBrokers, stats.ConnectedBrokers)

	// Проверяем здоровье
	err = pipeline.Health()
	assert.NoError(t, err)
}
