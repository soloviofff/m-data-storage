package services

import (
	"m-data-storage/internal/domain/entities"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestDataProcessorService_AdaptiveFeatures(t *testing.T) {
	// Create a simple processor for testing adaptive features
	logger := logrus.New()
	processor := &DataProcessorService{
		batchSize:         10,
		minBatchSize:      5,
		maxBatchSize:      50,
		adaptiveThreshold: 0.7,
		adaptiveBatching:  true,
		priorityBuffering: false,
		logger:            logger,
	}

	t.Run("adaptive batch size calculation", func(t *testing.T) {
		// Test low utilization - should decrease batch size
		lowUtilSize := processor.calculateAdaptiveBatchSize(100, 1000) // 10% utilization
		if lowUtilSize >= processor.batchSize {
			t.Errorf("Expected smaller batch size for low utilization, got %d", lowUtilSize)
		}

		// Test high utilization - should increase batch size
		highUtilSize := processor.calculateAdaptiveBatchSize(800, 1000) // 80% utilization
		if highUtilSize <= processor.batchSize {
			t.Errorf("Expected larger batch size for high utilization, got %d", highUtilSize)
		}

		// Test max batch size limit
		maxUtilSize := processor.calculateAdaptiveBatchSize(999, 1000) // 99.9% utilization
		if maxUtilSize > processor.maxBatchSize {
			t.Errorf("Batch size exceeded maximum: %d > %d", maxUtilSize, processor.maxBatchSize)
		}

		// Test min batch size limit
		minUtilSize := processor.calculateAdaptiveBatchSize(1, 1000) // 0.1% utilization
		if minUtilSize < processor.minBatchSize {
			t.Errorf("Batch size below minimum: %d < %d", minUtilSize, processor.minBatchSize)
		}
	})

	t.Run("adaptive batching settings", func(t *testing.T) {
		// Test enabling/disabling adaptive batching
		processor.SetAdaptiveBatching(false)
		if processor.adaptiveBatching {
			t.Error("Adaptive batching should be disabled")
		}

		processor.SetAdaptiveBatching(true)
		if !processor.adaptiveBatching {
			t.Error("Adaptive batching should be enabled")
		}

		// Test batch size limits
		processor.SetBatchSizeLimits(5, 200)
		if processor.minBatchSize != 5 || processor.maxBatchSize != 200 {
			t.Errorf("Batch size limits not set correctly: min=%d, max=%d",
				processor.minBatchSize, processor.maxBatchSize)
		}

		// Test adaptive threshold
		processor.SetAdaptiveThreshold(0.8)
		if processor.adaptiveThreshold != 0.8 {
			t.Errorf("Adaptive threshold not set correctly: %f", processor.adaptiveThreshold)
		}
	})

	t.Run("priority buffering settings", func(t *testing.T) {
		// Test enabling/disabling priority buffering
		processor.SetPriorityBuffering(false)
		if processor.priorityBuffering {
			t.Error("Priority buffering should be disabled")
		}

		processor.SetPriorityBuffering(true)
		if !processor.priorityBuffering {
			t.Error("Priority buffering should be enabled")
		}
	})

	t.Run("buffer statistics", func(t *testing.T) {
		// Initialize channels for testing
		processor.tickerChan = make(chan entities.Ticker, 100)
		processor.candleChan = make(chan entities.Candle, 100)
		processor.orderBookChan = make(chan entities.OrderBook, 100)

		// Get buffer stats
		stats := processor.GetBufferStats()

		// Check that stats are properly initialized
		if stats.TickerChannelUtilization < 0 || stats.TickerChannelUtilization > 1 {
			t.Errorf("Invalid ticker channel utilization: %f", stats.TickerChannelUtilization)
		}

		if stats.CandleChannelUtilization < 0 || stats.CandleChannelUtilization > 1 {
			t.Errorf("Invalid candle channel utilization: %f", stats.CandleChannelUtilization)
		}

		if stats.OrderBookChannelUtilization < 0 || stats.OrderBookChannelUtilization > 1 {
			t.Errorf("Invalid orderbook channel utilization: %f", stats.OrderBookChannelUtilization)
		}

		// Check that overflow events counter is initialized
		if stats.OverflowEvents < 0 {
			t.Errorf("Invalid overflow events count: %d", stats.OverflowEvents)
		}

		// Check that adaptive batch sizes slice is initialized
		if stats.AdaptiveBatchSizes == nil {
			t.Error("Adaptive batch sizes slice should be initialized")
		}
	})
}
