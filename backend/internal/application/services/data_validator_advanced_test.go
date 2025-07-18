package services

import (
	"testing"
	"time"

	"m-data-storage/internal/domain/entities"
)

func TestDataValidatorService_AdvancedValidation(t *testing.T) {
	validator := NewDataValidatorService()

	t.Run("ticker anomaly detection", func(t *testing.T) {
		// Enable anomaly detection
		validator.SetAnomalyDetection(true)
		validator.SetMaxPriceDeviation(50.0) // 50% max deviation

		// First ticker - should pass
		ticker1 := entities.Ticker{
			Symbol:    "BTCUSDT",
			BrokerID:  "test-broker",
			Price:     50000,
			Volume:    1.0,
			Timestamp: time.Now(),
			Market:    entities.MarketTypeSpot,
			Type:      entities.InstrumentTypeSpot,
		}
		err := validator.ValidateTicker(ticker1)
		if err != nil {
			t.Errorf("First ticker validation failed: %v", err)
		}

		// Second ticker with huge price deviation - should fail with anomaly detection
		ticker2 := entities.Ticker{
			Symbol:    "BTCUSDT",
			BrokerID:  "test-broker",
			Price:     100000, // 100% increase
			Volume:    1.0,
			Timestamp: time.Now().Add(time.Second),
			Market:    entities.MarketTypeSpot,
			Type:      entities.InstrumentTypeSpot,
		}
		err = validator.ValidateTicker(ticker2)
		if err == nil {
			t.Error("Expected price deviation error, but got none")
		}

		// Disable anomaly detection - now should pass
		validator.SetAnomalyDetection(false)
		err = validator.ValidateTicker(ticker2)
		if err != nil {
			t.Errorf("Ticker should pass with anomaly detection disabled: %v", err)
		}
	})

	t.Run("duplicate detection", func(t *testing.T) {
		// Reset validator state
		validator.ClearTrackingData()
		validator.SetDuplicateDetection(true)

		now := time.Now()
		
		// First ticker - should pass
		ticker1 := entities.Ticker{
			Symbol:    "ETHUSDT",
			BrokerID:  "test-broker",
			Price:     3000,
			Volume:    1.0,
			Timestamp: now,
			Market:    entities.MarketTypeSpot,
			Type:      entities.InstrumentTypeSpot,
		}
		err := validator.ValidateTicker(ticker1)
		if err != nil {
			t.Errorf("First ticker validation failed: %v", err)
		}

		// Duplicate ticker - should fail
		ticker2 := entities.Ticker{
			Symbol:    "ETHUSDT",
			BrokerID:  "test-broker",
			Price:     3000,
			Volume:    1.0,
			Timestamp: now,
			Market:    entities.MarketTypeSpot,
			Type:      entities.InstrumentTypeSpot,
		}
		err = validator.ValidateTicker(ticker2)
		if err == nil {
			t.Error("Expected duplicate detection error, but got none")
		}

		// Disable duplicate detection - now should pass
		validator.SetDuplicateDetection(false)
		err = validator.ValidateTicker(ticker2)
		if err != nil {
			t.Errorf("Ticker should pass with duplicate detection disabled: %v", err)
		}
	})

	t.Run("validation settings management", func(t *testing.T) {
		// Test setting anomaly detection
		validator.SetAnomalyDetection(false)
		stats := validator.GetValidationStats()
		if stats["anomaly_detection_enabled"].(bool) != false {
			t.Error("Anomaly detection should be disabled")
		}

		// Test setting price deviation
		validator.SetMaxPriceDeviation(25.0)
		stats = validator.GetValidationStats()
		if stats["max_price_deviation"].(float64) != 25.0 {
			t.Error("Max price deviation should be 25.0")
		}

		// Test setting volume spike
		validator.SetMaxVolumeSpike(5.0)
		stats = validator.GetValidationStats()
		if stats["max_volume_spike"].(float64) != 5.0 {
			t.Error("Max volume spike should be 5.0")
		}

		// Test clearing tracking data
		validator.ClearTrackingData()
		stats = validator.GetValidationStats()
		if stats["tracked_tickers"].(int) != 0 {
			t.Error("Tracked tickers should be 0 after clearing")
		}
		if stats["tracked_candles"].(int) != 0 {
			t.Error("Tracked candles should be 0 after clearing")
		}
	})
}
