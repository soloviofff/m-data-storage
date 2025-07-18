package services

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// DataQueryService implements the DataQuery interface
type DataQueryService struct {
	storageManager interfaces.StorageManager
	logger         *logrus.Logger
}

// NewDataQueryService creates a new data query service
func NewDataQueryService(
	storageManager interfaces.StorageManager,
	logger *logrus.Logger,
) *DataQueryService {
	if logger == nil {
		logger = logrus.New()
	}

	return &DataQueryService{
		storageManager: storageManager,
		logger:         logger,
	}
}

// GetTickers retrieves tickers based on filter
func (dqs *DataQueryService) GetTickers(ctx context.Context, filter interfaces.TickerFilter) ([]entities.Ticker, error) {
	if dqs.storageManager == nil {
		dqs.logger.Warn("StorageManager not available, returning empty ticker list")
		return []entities.Ticker{}, nil
	}

	// For now, delegate to storage manager
	// In the future, this could include caching, aggregation, etc.
	return dqs.storageManager.GetTickers(ctx, filter)
}

// GetCandles retrieves candles based on filter
func (dqs *DataQueryService) GetCandles(ctx context.Context, filter interfaces.CandleFilter) ([]entities.Candle, error) {
	if dqs.storageManager == nil {
		dqs.logger.Warn("StorageManager not available, returning empty candle list")
		return []entities.Candle{}, nil
	}

	// For now, delegate to storage manager
	// In the future, this could include caching, aggregation, etc.
	return dqs.storageManager.GetCandles(ctx, filter)
}

// GetOrderBooks retrieves order books based on filter
func (dqs *DataQueryService) GetOrderBooks(ctx context.Context, filter interfaces.OrderBookFilter) ([]entities.OrderBook, error) {
	if dqs.storageManager == nil {
		dqs.logger.Warn("StorageManager not available, returning empty order book list")
		return []entities.OrderBook{}, nil
	}

	// For now, delegate to storage manager
	// In the future, this could include caching, aggregation, etc.
	return dqs.storageManager.GetOrderBooks(ctx, filter)
}

// GetTickerAggregates retrieves aggregated ticker data
func (dqs *DataQueryService) GetTickerAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]interfaces.TickerAggregate, error) {
	// TODO: Implement ticker aggregation
	// This would involve grouping tickers by time intervals and calculating aggregates
	dqs.logger.WithFields(logrus.Fields{
		"symbol":     symbol,
		"interval":   interval,
		"start_time": startTime,
		"end_time":   endTime,
	}).Warn("Ticker aggregates not yet implemented")

	return []interfaces.TickerAggregate{}, nil
}

// GetCandleAggregates retrieves aggregated candle data
func (dqs *DataQueryService) GetCandleAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]interfaces.CandleAggregate, error) {
	// TODO: Implement candle aggregation
	// This would involve grouping candles by time intervals and calculating aggregates
	dqs.logger.WithFields(logrus.Fields{
		"symbol":     symbol,
		"interval":   interval,
		"start_time": startTime,
		"end_time":   endTime,
	}).Warn("Candle aggregates not yet implemented")

	return []interfaces.CandleAggregate{}, nil
}

// GetDataStats retrieves data statistics
func (dqs *DataQueryService) GetDataStats(ctx context.Context) (interfaces.DataStats, error) {
	// TODO: Implement data statistics
	// This would involve querying storage for various statistics
	dqs.logger.Info("Data statistics not yet implemented")

	return interfaces.DataStats{
		TotalRecords:    0,
		RecordsByType:   make(map[string]int64),
		RecordsByBroker: make(map[string]int64),
		RecordsBySymbol: make(map[string]int64),
		OldestRecord:    time.Now(),
		NewestRecord:    time.Now(),
		StorageSize:     0,
	}, nil
}
