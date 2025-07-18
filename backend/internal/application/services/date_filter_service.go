package services

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// DateFilterService handles date-based filtering and tracking
type DateFilterService struct {
	metadataStorage   interfaces.MetadataStorage
	timeSeriesStorage interfaces.TimeSeriesStorage
	logger            *logrus.Logger
}

// NewDateFilterService creates a new date filter service
func NewDateFilterService(
	metadataStorage interfaces.MetadataStorage,
	timeSeriesStorage interfaces.TimeSeriesStorage,
	logger *logrus.Logger,
) *DateFilterService {
	if logger == nil {
		logger = logrus.New()
	}

	return &DateFilterService{
		metadataStorage:   metadataStorage,
		timeSeriesStorage: timeSeriesStorage,
		logger:            logger,
	}
}

// FilterTickersBySubscriptionDate filters tickers based on subscription start dates
func (dfs *DateFilterService) FilterTickersBySubscriptionDate(ctx context.Context, filter interfaces.TickerFilter) ([]entities.Ticker, error) {
	// Get all tickers first
	tickers, err := dfs.timeSeriesStorage.GetTickers(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get tickers: %w", err)
	}

	// If no symbols specified, return all tickers
	if len(filter.Symbols) == 0 {
		return tickers, nil
	}

	// Get subscription start dates for each symbol
	subscriptionDates := make(map[string]time.Time)
	for _, symbol := range filter.Symbols {
		startDate, err := dfs.getSubscriptionStartDate(ctx, symbol, filter.BrokerIDs)
		if err != nil {
			dfs.logger.WithError(err).WithField("symbol", symbol).Warn("Failed to get subscription start date")
			continue
		}
		if !startDate.IsZero() {
			subscriptionDates[symbol] = startDate
		}
	}

	// Filter tickers based on subscription start dates
	filteredTickers := make([]entities.Ticker, 0, len(tickers))
	for _, ticker := range tickers {
		if startDate, exists := subscriptionDates[ticker.Symbol]; exists {
			// Only include tickers from subscription start date onwards
			if ticker.Timestamp.After(startDate) || ticker.Timestamp.Equal(startDate) {
				filteredTickers = append(filteredTickers, ticker)
			}
		} else {
			// If no subscription found, include the ticker (backward compatibility)
			filteredTickers = append(filteredTickers, ticker)
		}
	}

	dfs.logger.WithFields(logrus.Fields{
		"original_count": len(tickers),
		"filtered_count": len(filteredTickers),
		"symbols":        filter.Symbols,
	}).Debug("Filtered tickers by subscription date")

	return filteredTickers, nil
}

// FilterCandlesBySubscriptionDate filters candles based on subscription start dates
func (dfs *DateFilterService) FilterCandlesBySubscriptionDate(ctx context.Context, filter interfaces.CandleFilter) ([]entities.Candle, error) {
	// Get all candles first
	candles, err := dfs.timeSeriesStorage.GetCandles(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get candles: %w", err)
	}

	// If no symbols specified, return all candles
	if len(filter.Symbols) == 0 {
		return candles, nil
	}

	// Get subscription start dates for each symbol
	subscriptionDates := make(map[string]time.Time)
	for _, symbol := range filter.Symbols {
		startDate, err := dfs.getSubscriptionStartDate(ctx, symbol, filter.BrokerIDs)
		if err != nil {
			dfs.logger.WithError(err).WithField("symbol", symbol).Warn("Failed to get subscription start date")
			continue
		}
		if !startDate.IsZero() {
			subscriptionDates[symbol] = startDate
		}
	}

	// Filter candles based on subscription start dates
	filteredCandles := make([]entities.Candle, 0, len(candles))
	for _, candle := range candles {
		if startDate, exists := subscriptionDates[candle.Symbol]; exists {
			// Only include candles from subscription start date onwards
			if candle.Timestamp.After(startDate) || candle.Timestamp.Equal(startDate) {
				filteredCandles = append(filteredCandles, candle)
			}
		} else {
			// If no subscription found, include the candle (backward compatibility)
			filteredCandles = append(filteredCandles, candle)
		}
	}

	dfs.logger.WithFields(logrus.Fields{
		"original_count": len(candles),
		"filtered_count": len(filteredCandles),
		"symbols":        filter.Symbols,
	}).Debug("Filtered candles by subscription date")

	return filteredCandles, nil
}

// FilterOrderBooksBySubscriptionDate filters order books based on subscription start dates
func (dfs *DateFilterService) FilterOrderBooksBySubscriptionDate(ctx context.Context, filter interfaces.OrderBookFilter) ([]entities.OrderBook, error) {
	// Get all order books first
	orderBooks, err := dfs.timeSeriesStorage.GetOrderBooks(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get order books: %w", err)
	}

	// If no symbols specified, return all order books
	if len(filter.Symbols) == 0 {
		return orderBooks, nil
	}

	// Get subscription start dates for each symbol
	subscriptionDates := make(map[string]time.Time)
	for _, symbol := range filter.Symbols {
		startDate, err := dfs.getSubscriptionStartDate(ctx, symbol, filter.BrokerIDs)
		if err != nil {
			dfs.logger.WithError(err).WithField("symbol", symbol).Warn("Failed to get subscription start date")
			continue
		}
		if !startDate.IsZero() {
			subscriptionDates[symbol] = startDate
		}
	}

	// Filter order books based on subscription start dates
	filteredOrderBooks := make([]entities.OrderBook, 0, len(orderBooks))
	for _, orderBook := range orderBooks {
		if startDate, exists := subscriptionDates[orderBook.Symbol]; exists {
			// Only include order books from subscription start date onwards
			if orderBook.Timestamp.After(startDate) || orderBook.Timestamp.Equal(startDate) {
				filteredOrderBooks = append(filteredOrderBooks, orderBook)
			}
		} else {
			// If no subscription found, include the order book (backward compatibility)
			filteredOrderBooks = append(filteredOrderBooks, orderBook)
		}
	}

	dfs.logger.WithFields(logrus.Fields{
		"original_count": len(orderBooks),
		"filtered_count": len(filteredOrderBooks),
		"symbols":        filter.Symbols,
	}).Debug("Filtered order books by subscription date")

	return filteredOrderBooks, nil
}

// getSubscriptionStartDate retrieves the earliest subscription start date for a symbol
func (dfs *DateFilterService) getSubscriptionStartDate(ctx context.Context, symbol string, brokerIDs []string) (time.Time, error) {
	if dfs.metadataStorage == nil {
		return time.Time{}, fmt.Errorf("metadata storage not available")
	}

	// Create filter for subscriptions
	filter := interfaces.SubscriptionFilter{
		Symbols: []string{symbol},
	}

	// Add broker IDs if specified
	if len(brokerIDs) > 0 {
		filter.BrokerIDs = brokerIDs
	}

	// Get subscriptions for the symbol
	subscriptions, err := dfs.metadataStorage.GetSubscriptions(ctx, filter)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		return time.Time{}, nil
	}

	// Find the earliest start date
	var earliestDate time.Time
	for _, sub := range subscriptions {
		if earliestDate.IsZero() || sub.StartDate.Before(earliestDate) {
			earliestDate = sub.StartDate
		}
	}

	return earliestDate, nil
}
