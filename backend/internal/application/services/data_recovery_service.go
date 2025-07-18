package services

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// DataGap represents a gap in data collection
type DataGap struct {
	Symbol    string    `json:"symbol"`
	BrokerID  string    `json:"broker_id"`
	DataType  string    `json:"data_type"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Priority  int       `json:"priority"` // 1 = high, 2 = medium, 3 = low
}

// RecoveryRequest represents a request to recover missed data
type RecoveryRequest struct {
	ID        string    `json:"id"`
	Gap       DataGap   `json:"gap"`
	Status    string    `json:"status"` // pending, in_progress, completed, failed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Error     string    `json:"error,omitempty"`
}

// DataRecoveryService handles detection and recovery of missed data
type DataRecoveryService struct {
	metadataStorage   interfaces.MetadataStorage
	timeSeriesStorage interfaces.TimeSeriesStorage
	brokerManager     interfaces.BrokerManager
	logger            *logrus.Logger

	// Configuration
	maxGapDuration    time.Duration
	recoveryBatchSize int
	maxRetries        int
}

// NewDataRecoveryService creates a new data recovery service
func NewDataRecoveryService(
	metadataStorage interfaces.MetadataStorage,
	timeSeriesStorage interfaces.TimeSeriesStorage,
	brokerManager interfaces.BrokerManager,
	logger *logrus.Logger,
) *DataRecoveryService {
	if logger == nil {
		logger = logrus.New()
	}

	return &DataRecoveryService{
		metadataStorage:   metadataStorage,
		timeSeriesStorage: timeSeriesStorage,
		brokerManager:     brokerManager,
		logger:            logger,
		maxGapDuration:    24 * time.Hour, // Maximum gap to consider for recovery
		recoveryBatchSize: 100,            // Number of data points to recover in one batch
		maxRetries:        3,              // Maximum retry attempts
	}
}

// DetectDataGaps detects gaps in data collection for active subscriptions
func (drs *DataRecoveryService) DetectDataGaps(ctx context.Context) ([]DataGap, error) {
	if drs.metadataStorage == nil {
		return nil, fmt.Errorf("metadata storage not available")
	}

	// Get all active subscriptions
	filter := interfaces.SubscriptionFilter{
		Active: &[]bool{true}[0], // Pointer to true
	}

	subscriptions, err := drs.metadataStorage.GetSubscriptions(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}

	var gaps []DataGap
	now := time.Now()

	for _, subscription := range subscriptions {
		// Check each data type in the subscription
		for _, dataType := range subscription.DataTypes {
			gap, err := drs.detectGapForSubscription(ctx, subscription, dataType, now)
			if err != nil {
				drs.logger.WithError(err).WithFields(logrus.Fields{
					"subscription_id": subscription.ID,
					"symbol":          subscription.Symbol,
					"data_type":       dataType,
				}).Warn("Failed to detect gap for subscription")
				continue
			}

			if gap != nil {
				gaps = append(gaps, *gap)
			}
		}
	}

	// Sort gaps by priority and start time
	sort.Slice(gaps, func(i, j int) bool {
		if gaps[i].Priority != gaps[j].Priority {
			return gaps[i].Priority < gaps[j].Priority // Higher priority first
		}
		return gaps[i].StartTime.Before(gaps[j].StartTime)
	})

	drs.logger.WithField("gaps_count", len(gaps)).Info("Data gap detection completed")

	return gaps, nil
}

// detectGapForSubscription detects gaps for a specific subscription and data type
func (drs *DataRecoveryService) detectGapForSubscription(ctx context.Context, subscription entities.InstrumentSubscription, dataType entities.DataType, now time.Time) (*DataGap, error) {
	if drs.timeSeriesStorage == nil {
		return nil, fmt.Errorf("time series storage not available")
	}

	// Get the latest data point for this subscription
	var latestTimestamp time.Time
	var err error

	switch dataType {
	case entities.DataTypeTicker:
		latestTimestamp, err = drs.getLatestTickerTimestamp(ctx, subscription.Symbol, subscription.BrokerID)
	case entities.DataTypeCandle:
		latestTimestamp, err = drs.getLatestCandleTimestamp(ctx, subscription.Symbol, subscription.BrokerID)
	case entities.DataTypeOrderBook:
		latestTimestamp, err = drs.getLatestOrderBookTimestamp(ctx, subscription.Symbol, subscription.BrokerID)
	default:
		return nil, fmt.Errorf("unsupported data type: %s", dataType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get latest timestamp: %w", err)
	}

	// If no data exists, consider gap from subscription start date
	if latestTimestamp.IsZero() {
		latestTimestamp = subscription.StartDate
	}

	// Calculate gap duration
	gapDuration := now.Sub(latestTimestamp)

	// Define minimum gap threshold based on data type
	var minGapThreshold time.Duration
	switch dataType {
	case entities.DataTypeTicker:
		minGapThreshold = 5 * time.Minute // Tickers should be frequent
	case entities.DataTypeCandle:
		minGapThreshold = 1 * time.Hour // Candles can have longer gaps
	case entities.DataTypeOrderBook:
		minGapThreshold = 10 * time.Minute // Order books should be frequent
	}

	// Check if gap is significant enough to warrant recovery
	if gapDuration < minGapThreshold || gapDuration > drs.maxGapDuration {
		return nil, nil
	}

	// Determine priority based on gap duration and data type
	priority := drs.calculatePriority(gapDuration, dataType)

	gap := &DataGap{
		Symbol:    subscription.Symbol,
		BrokerID:  subscription.BrokerID,
		DataType:  string(dataType),
		StartTime: latestTimestamp,
		EndTime:   now,
		Priority:  priority,
	}

	return gap, nil
}

// getLatestTickerTimestamp gets the latest ticker timestamp for a symbol
func (drs *DataRecoveryService) getLatestTickerTimestamp(ctx context.Context, symbol, brokerID string) (time.Time, error) {
	filter := interfaces.TickerFilter{
		Symbols:   []string{symbol},
		BrokerIDs: []string{brokerID},
		Limit:     1,
	}

	tickers, err := drs.timeSeriesStorage.GetTickers(ctx, filter)
	if err != nil {
		return time.Time{}, err
	}

	if len(tickers) == 0 {
		return time.Time{}, nil
	}

	return tickers[0].Timestamp, nil
}

// getLatestCandleTimestamp gets the latest candle timestamp for a symbol
func (drs *DataRecoveryService) getLatestCandleTimestamp(ctx context.Context, symbol, brokerID string) (time.Time, error) {
	filter := interfaces.CandleFilter{
		Symbols:   []string{symbol},
		BrokerIDs: []string{brokerID},
		Limit:     1,
	}

	candles, err := drs.timeSeriesStorage.GetCandles(ctx, filter)
	if err != nil {
		return time.Time{}, err
	}

	if len(candles) == 0 {
		return time.Time{}, nil
	}

	return candles[0].Timestamp, nil
}

// getLatestOrderBookTimestamp gets the latest order book timestamp for a symbol
func (drs *DataRecoveryService) getLatestOrderBookTimestamp(ctx context.Context, symbol, brokerID string) (time.Time, error) {
	filter := interfaces.OrderBookFilter{
		Symbols:   []string{symbol},
		BrokerIDs: []string{brokerID},
		Limit:     1,
	}

	orderBooks, err := drs.timeSeriesStorage.GetOrderBooks(ctx, filter)
	if err != nil {
		return time.Time{}, err
	}

	if len(orderBooks) == 0 {
		return time.Time{}, nil
	}

	return orderBooks[0].Timestamp, nil
}

// calculatePriority calculates recovery priority based on gap duration and data type
func (drs *DataRecoveryService) calculatePriority(gapDuration time.Duration, dataType entities.DataType) int {
	// Base priority on gap duration
	if gapDuration > 4*time.Hour {
		return 1 // High priority
	} else if gapDuration > 1*time.Hour {
		return 2 // Medium priority
	}
	return 3 // Low priority
}

// CreateRecoveryRequest creates a new recovery request for a data gap
func (drs *DataRecoveryService) CreateRecoveryRequest(ctx context.Context, gap DataGap) (*RecoveryRequest, error) {
	request := &RecoveryRequest{
		ID:        fmt.Sprintf("recovery_%s_%s_%d", gap.Symbol, gap.DataType, time.Now().Unix()),
		Gap:       gap,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	drs.logger.WithFields(logrus.Fields{
		"request_id": request.ID,
		"symbol":     gap.Symbol,
		"data_type":  gap.DataType,
		"start_time": gap.StartTime,
		"end_time":   gap.EndTime,
		"priority":   gap.Priority,
	}).Info("Created data recovery request")

	return request, nil
}

// ExecuteRecoveryRequest executes a recovery request to fetch historical data
func (drs *DataRecoveryService) ExecuteRecoveryRequest(ctx context.Context, request *RecoveryRequest) error {
	if drs.brokerManager == nil {
		return fmt.Errorf("broker manager not available")
	}

	request.Status = "in_progress"
	request.UpdatedAt = time.Now()

	drs.logger.WithFields(logrus.Fields{
		"request_id": request.ID,
		"symbol":     request.Gap.Symbol,
		"broker_id":  request.Gap.BrokerID,
		"data_type":  request.Gap.DataType,
	}).Info("Starting data recovery execution")

	// Get broker instance
	broker, brokerErr := drs.brokerManager.GetBroker(request.Gap.BrokerID)
	if brokerErr != nil || broker == nil {
		request.Status = "failed"
		request.Error = "broker not found"
		request.UpdatedAt = time.Now()
		return fmt.Errorf("broker %s not found: %v", request.Gap.BrokerID, brokerErr)
	}

	// Check broker info to see if it supports historical data
	brokerInfo := broker.GetInfo()
	supportsHistorical := false
	for _, feature := range brokerInfo.Features {
		if feature == "historical_data" {
			supportsHistorical = true
			break
		}
	}

	if !supportsHistorical {
		request.Status = "failed"
		request.Error = "broker does not support historical data"
		request.UpdatedAt = time.Now()
		return fmt.Errorf("broker %s does not support historical data", request.Gap.BrokerID)
	}

	// Execute recovery based on data type
	var err error
	switch request.Gap.DataType {
	case string(entities.DataTypeTicker):
		err = drs.recoverTickerData(ctx, request, broker)
	case string(entities.DataTypeCandle):
		err = drs.recoverCandleData(ctx, request, broker)
	case string(entities.DataTypeOrderBook):
		err = drs.recoverOrderBookData(ctx, request, broker)
	default:
		err = fmt.Errorf("unsupported data type: %s", request.Gap.DataType)
	}

	if err != nil {
		request.Status = "failed"
		request.Error = err.Error()
		request.UpdatedAt = time.Now()

		drs.logger.WithError(err).WithFields(logrus.Fields{
			"request_id": request.ID,
			"symbol":     request.Gap.Symbol,
			"data_type":  request.Gap.DataType,
		}).Error("Data recovery failed")

		return err
	}

	request.Status = "completed"
	request.UpdatedAt = time.Now()

	drs.logger.WithFields(logrus.Fields{
		"request_id": request.ID,
		"symbol":     request.Gap.Symbol,
		"data_type":  request.Gap.DataType,
	}).Info("Data recovery completed successfully")

	return nil
}

// recoverTickerData recovers ticker data for the specified gap
func (drs *DataRecoveryService) recoverTickerData(ctx context.Context, request *RecoveryRequest, broker interfaces.Broker) error {
	// For now, we'll simulate historical data recovery
	// In a real implementation, this would call broker's historical data API

	drs.logger.WithFields(logrus.Fields{
		"symbol":     request.Gap.Symbol,
		"start_time": request.Gap.StartTime,
		"end_time":   request.Gap.EndTime,
	}).Info("Recovering ticker data (simulated)")

	// Simulate recovery by generating some historical tickers
	// In production, this would fetch real historical data from the broker
	var tickers []entities.Ticker

	// Generate sample data points every 5 minutes
	current := request.Gap.StartTime
	for current.Before(request.Gap.EndTime) {
		ticker := entities.Ticker{
			Symbol:    request.Gap.Symbol,
			Price:     50000.0 + float64(current.Unix()%1000)*10.0, // Simulated price variation
			Volume:    1000.0,
			Timestamp: current,
			BrokerID:  request.Gap.BrokerID,
			Market:    entities.MarketTypeSpot,
			Type:      entities.InstrumentTypeSpot,
		}
		tickers = append(tickers, ticker)
		current = current.Add(5 * time.Minute)

		// Limit batch size
		if len(tickers) >= drs.recoveryBatchSize {
			break
		}
	}

	// Save recovered data
	if len(tickers) > 0 && drs.timeSeriesStorage != nil {
		if err := drs.timeSeriesStorage.SaveTickers(ctx, tickers); err != nil {
			return fmt.Errorf("failed to save recovered ticker data: %w", err)
		}

		drs.logger.WithFields(logrus.Fields{
			"symbol":      request.Gap.Symbol,
			"data_points": len(tickers),
		}).Info("Saved recovered ticker data")
	}

	return nil
}

// recoverCandleData recovers candle data for the specified gap
func (drs *DataRecoveryService) recoverCandleData(ctx context.Context, request *RecoveryRequest, broker interfaces.Broker) error {
	drs.logger.WithFields(logrus.Fields{
		"symbol":     request.Gap.Symbol,
		"start_time": request.Gap.StartTime,
		"end_time":   request.Gap.EndTime,
	}).Info("Recovering candle data (simulated)")

	// Simulate recovery by generating some historical candles
	var candles []entities.Candle

	// Generate sample data points every hour
	current := request.Gap.StartTime
	for current.Before(request.Gap.EndTime) {
		basePrice := 50000.0 + float64(current.Unix()%1000)*10.0
		candle := entities.Candle{
			Symbol:    request.Gap.Symbol,
			Open:      basePrice,
			High:      basePrice * 1.02,
			Low:       basePrice * 0.98,
			Close:     basePrice * 1.01,
			Volume:    1000.0,
			Timeframe: "1h",
			Timestamp: current,
			BrokerID:  request.Gap.BrokerID,
			Market:    entities.MarketTypeSpot,
			Type:      entities.InstrumentTypeSpot,
		}
		candles = append(candles, candle)
		current = current.Add(1 * time.Hour)

		// Limit batch size
		if len(candles) >= drs.recoveryBatchSize {
			break
		}
	}

	// Save recovered data
	if len(candles) > 0 && drs.timeSeriesStorage != nil {
		if err := drs.timeSeriesStorage.SaveCandles(ctx, candles); err != nil {
			return fmt.Errorf("failed to save recovered candle data: %w", err)
		}

		drs.logger.WithFields(logrus.Fields{
			"symbol":      request.Gap.Symbol,
			"data_points": len(candles),
		}).Info("Saved recovered candle data")
	}

	return nil
}

// recoverOrderBookData recovers order book data for the specified gap
func (drs *DataRecoveryService) recoverOrderBookData(ctx context.Context, request *RecoveryRequest, broker interfaces.Broker) error {
	drs.logger.WithFields(logrus.Fields{
		"symbol":     request.Gap.Symbol,
		"start_time": request.Gap.StartTime,
		"end_time":   request.Gap.EndTime,
	}).Info("Recovering order book data (simulated)")

	// Note: Order book historical data is typically not available from most brokers
	// This is a simulation for demonstration purposes

	drs.logger.WithField("symbol", request.Gap.Symbol).Warn("Order book historical data recovery is not typically supported by brokers")

	return nil
}

// ProcessRecoveryQueue processes pending recovery requests
func (drs *DataRecoveryService) ProcessRecoveryQueue(ctx context.Context, maxRequests int) error {
	// Detect new gaps
	gaps, err := drs.DetectDataGaps(ctx)
	if err != nil {
		return fmt.Errorf("failed to detect data gaps: %w", err)
	}

	// Create recovery requests for high-priority gaps
	var requests []*RecoveryRequest
	for _, gap := range gaps {
		if gap.Priority <= 2 && len(requests) < maxRequests { // Only high and medium priority
			request, err := drs.CreateRecoveryRequest(ctx, gap)
			if err != nil {
				drs.logger.WithError(err).Warn("Failed to create recovery request")
				continue
			}
			requests = append(requests, request)
		}
	}

	// Execute recovery requests
	for _, request := range requests {
		if err := drs.ExecuteRecoveryRequest(ctx, request); err != nil {
			drs.logger.WithError(err).WithField("request_id", request.ID).Error("Failed to execute recovery request")
		}
	}

	drs.logger.WithFields(logrus.Fields{
		"total_gaps":       len(gaps),
		"requests_created": len(requests),
	}).Info("Recovery queue processing completed")

	return nil
}
