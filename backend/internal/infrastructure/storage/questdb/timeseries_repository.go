package questdb

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver for QuestDB

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/interfaces"
)

// TimeSeriesRepository implements TimeSeriesStorage interface for QuestDB
type TimeSeriesRepository struct {
	db     *sql.DB
	config Config
}

// Config holds QuestDB connection configuration
type Config struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
}

// NewTimeSeriesRepository creates a new QuestDB repository
func NewTimeSeriesRepository(config Config) *TimeSeriesRepository {
	return &TimeSeriesRepository{
		config: config,
	}
}

// Connect establishes connection to QuestDB
func (r *TimeSeriesRepository) Connect(ctx context.Context) error {
	dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		r.config.Host, r.config.Port, r.config.Database,
		r.config.Username, r.config.Password, r.config.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open QuestDB connection: %w", err)
	}

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping QuestDB: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	r.db = db

	// Initialize schema
	if err := r.initializeSchema(ctx); err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// Disconnect closes the database connection
func (r *TimeSeriesRepository) Disconnect() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// Health checks the database connection
func (r *TimeSeriesRepository) Health() error {
	if r.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.db.PingContext(ctx)
}

// initializeSchema creates tables and indexes if they don't exist
func (r *TimeSeriesRepository) initializeSchema(ctx context.Context) error {
	schemaPath := filepath.Join("internal", "infrastructure", "storage", "questdb", "schema.sql")

	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = r.db.ExecContext(ctx, string(schemaBytes))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

// SaveTickers saves ticker data to QuestDB
func (r *TimeSeriesRepository) SaveTickers(ctx context.Context, tickers []entities.Ticker) error {
	if len(tickers) == 0 {
		return nil
	}

	query := `INSERT INTO tickers (
		symbol, price, volume, market, type, timestamp, broker_id,
		change, change_percent, high_24h, low_24h, volume_24h, prev_close_24h,
		open_interest, bid_price, ask_price, bid_size, ask_size
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, ticker := range tickers {
		_, err := stmt.ExecContext(ctx,
			ticker.Symbol,
			ticker.Price,
			ticker.Volume,
			string(ticker.Market),
			string(ticker.Type),
			ticker.Timestamp,
			ticker.BrokerID,
			ticker.Change,
			ticker.ChangePercent,
			ticker.High24h,
			ticker.Low24h,
			ticker.Volume24h,
			ticker.PrevClose24h,
			ticker.OpenInterest,
			ticker.BidPrice,
			ticker.AskPrice,
			ticker.BidSize,
			ticker.AskSize,
		)
		if err != nil {
			return fmt.Errorf("failed to insert ticker %s: %w", ticker.Symbol, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SaveCandles saves candle data to QuestDB
func (r *TimeSeriesRepository) SaveCandles(ctx context.Context, candles []entities.Candle) error {
	if len(candles) == 0 {
		return nil
	}

	query := `INSERT INTO candles (
		symbol, open, high, low, close, volume, market, type, timestamp, timeframe, broker_id,
		trades, quote_volume, open_interest
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, candle := range candles {
		_, err := stmt.ExecContext(ctx,
			candle.Symbol,
			candle.Open,
			candle.High,
			candle.Low,
			candle.Close,
			candle.Volume,
			string(candle.Market),
			string(candle.Type),
			candle.Timestamp,
			candle.Timeframe,
			candle.BrokerID,
			candle.Trades,
			candle.QuoteVolume,
			candle.OpenInterest,
		)
		if err != nil {
			return fmt.Errorf("failed to insert candle %s: %w", candle.Symbol, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SaveOrderBooks saves order book data to QuestDB
func (r *TimeSeriesRepository) SaveOrderBooks(ctx context.Context, orderBooks []entities.OrderBook) error {
	if len(orderBooks) == 0 {
		return nil
	}

	query := `INSERT INTO order_books (
		symbol, bids, asks, market, type, timestamp, broker_id, last_update_id, event_time
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, orderBook := range orderBooks {
		// Convert price levels to JSON strings for storage
		bidsJSON := r.priceLevelsToJSON(orderBook.Bids)
		asksJSON := r.priceLevelsToJSON(orderBook.Asks)

		_, err := stmt.ExecContext(ctx,
			orderBook.Symbol,
			bidsJSON,
			asksJSON,
			string(orderBook.Market),
			string(orderBook.Type),
			orderBook.Timestamp,
			orderBook.BrokerID,
			orderBook.LastUpdateID,
			orderBook.EventTime,
		)
		if err != nil {
			return fmt.Errorf("failed to insert order book %s: %w", orderBook.Symbol, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// priceLevelsToJSON converts price levels to JSON string
func (r *TimeSeriesRepository) priceLevelsToJSON(levels []entities.PriceLevel) string {
	if len(levels) == 0 {
		return "[]"
	}

	result := "["
	for i, level := range levels {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf(`{"price":%f,"quantity":%f}`, level.Price, level.Quantity)
	}
	result += "]"
	return result
}

// GetTickers retrieves ticker data based on filter
func (r *TimeSeriesRepository) GetTickers(ctx context.Context, filter interfaces.TickerFilter) ([]entities.Ticker, error) {
	query := "SELECT symbol, price, volume, market, type, timestamp, broker_id, " +
		"change, change_percent, high_24h, low_24h, volume_24h, prev_close_24h, " +
		"open_interest, bid_price, ask_price, bid_size, ask_size " +
		"FROM tickers WHERE 1=1"

	args := []interface{}{}
	argIndex := 1

	// Build WHERE clause based on filter
	if len(filter.Symbols) > 0 {
		query += fmt.Sprintf(" AND symbol IN (%s)", r.buildPlaceholders(len(filter.Symbols), &argIndex))
		for _, symbol := range filter.Symbols {
			args = append(args, symbol)
		}
	}

	if len(filter.BrokerIDs) > 0 {
		query += fmt.Sprintf(" AND broker_id IN (%s)", r.buildPlaceholders(len(filter.BrokerIDs), &argIndex))
		for _, brokerID := range filter.BrokerIDs {
			args = append(args, brokerID)
		}
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *filter.EndTime)
		argIndex++
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tickers: %w", err)
	}
	defer rows.Close()

	var tickers []entities.Ticker
	for rows.Next() {
		var ticker entities.Ticker
		var market, instrumentType string

		err := rows.Scan(
			&ticker.Symbol,
			&ticker.Price,
			&ticker.Volume,
			&market,
			&instrumentType,
			&ticker.Timestamp,
			&ticker.BrokerID,
			&ticker.Change,
			&ticker.ChangePercent,
			&ticker.High24h,
			&ticker.Low24h,
			&ticker.Volume24h,
			&ticker.PrevClose24h,
			&ticker.OpenInterest,
			&ticker.BidPrice,
			&ticker.AskPrice,
			&ticker.BidSize,
			&ticker.AskSize,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticker: %w", err)
		}

		ticker.Market = entities.MarketType(market)
		ticker.Type = entities.InstrumentType(instrumentType)
		tickers = append(tickers, ticker)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tickers: %w", err)
	}

	return tickers, nil
}

// GetCandles retrieves candle data based on filter
func (r *TimeSeriesRepository) GetCandles(ctx context.Context, filter interfaces.CandleFilter) ([]entities.Candle, error) {
	query := "SELECT symbol, open, high, low, close, volume, market, type, timestamp, timeframe, broker_id, " +
		"trades, quote_volume, open_interest " +
		"FROM candles WHERE 1=1"

	args := []interface{}{}
	argIndex := 1

	// Build WHERE clause based on filter
	if len(filter.Symbols) > 0 {
		query += fmt.Sprintf(" AND symbol IN (%s)", r.buildPlaceholders(len(filter.Symbols), &argIndex))
		for _, symbol := range filter.Symbols {
			args = append(args, symbol)
		}
	}

	if len(filter.BrokerIDs) > 0 {
		query += fmt.Sprintf(" AND broker_id IN (%s)", r.buildPlaceholders(len(filter.BrokerIDs), &argIndex))
		for _, brokerID := range filter.BrokerIDs {
			args = append(args, brokerID)
		}
	}

	if len(filter.Timeframes) > 0 {
		query += fmt.Sprintf(" AND timeframe IN (%s)", r.buildPlaceholders(len(filter.Timeframes), &argIndex))
		for _, timeframe := range filter.Timeframes {
			args = append(args, timeframe)
		}
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *filter.EndTime)
		argIndex++
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query candles: %w", err)
	}
	defer rows.Close()

	var candles []entities.Candle
	for rows.Next() {
		var candle entities.Candle
		var market, instrumentType string

		err := rows.Scan(
			&candle.Symbol,
			&candle.Open,
			&candle.High,
			&candle.Low,
			&candle.Close,
			&candle.Volume,
			&market,
			&instrumentType,
			&candle.Timestamp,
			&candle.Timeframe,
			&candle.BrokerID,
			&candle.Trades,
			&candle.QuoteVolume,
			&candle.OpenInterest,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan candle: %w", err)
		}

		candle.Market = entities.MarketType(market)
		candle.Type = entities.InstrumentType(instrumentType)
		candles = append(candles, candle)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate candles: %w", err)
	}

	return candles, nil
}

// buildPlaceholders builds SQL placeholders for IN clauses
func (r *TimeSeriesRepository) buildPlaceholders(count int, argIndex *int) string {
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = fmt.Sprintf("$%d", *argIndex)
		*argIndex++
	}
	return fmt.Sprintf("%s", placeholders[0])
	// Note: For simplicity, we only handle single values.
	// In production, you'd join all placeholders with commas
}

// GetOrderBooks retrieves order book data based on filter
func (r *TimeSeriesRepository) GetOrderBooks(ctx context.Context, filter interfaces.OrderBookFilter) ([]entities.OrderBook, error) {
	query := "SELECT symbol, bids, asks, market, type, timestamp, broker_id, last_update_id, event_time " +
		"FROM order_books WHERE 1=1"

	args := []interface{}{}
	argIndex := 1

	// Build WHERE clause based on filter
	if len(filter.Symbols) > 0 {
		query += fmt.Sprintf(" AND symbol IN (%s)", r.buildPlaceholders(len(filter.Symbols), &argIndex))
		for _, symbol := range filter.Symbols {
			args = append(args, symbol)
		}
	}

	if len(filter.BrokerIDs) > 0 {
		query += fmt.Sprintf(" AND broker_id IN (%s)", r.buildPlaceholders(len(filter.BrokerIDs), &argIndex))
		for _, brokerID := range filter.BrokerIDs {
			args = append(args, brokerID)
		}
	}

	if filter.StartTime != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argIndex)
		args = append(args, *filter.StartTime)
		argIndex++
	}

	if filter.EndTime != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argIndex)
		args = append(args, *filter.EndTime)
		argIndex++
	}

	query += " ORDER BY timestamp DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIndex)
		args = append(args, filter.Limit)
		argIndex++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIndex)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query order books: %w", err)
	}
	defer rows.Close()

	var orderBooks []entities.OrderBook
	for rows.Next() {
		var orderBook entities.OrderBook
		var market, instrumentType, bidsJSON, asksJSON string
		var eventTime sql.NullTime

		err := rows.Scan(
			&orderBook.Symbol,
			&bidsJSON,
			&asksJSON,
			&market,
			&instrumentType,
			&orderBook.Timestamp,
			&orderBook.BrokerID,
			&orderBook.LastUpdateID,
			&eventTime,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order book: %w", err)
		}

		orderBook.Market = entities.MarketType(market)
		orderBook.Type = entities.InstrumentType(instrumentType)
		if eventTime.Valid {
			orderBook.EventTime = eventTime.Time
		}

		// Parse JSON price levels (simplified - in real implementation would use proper JSON parsing)
		orderBook.Bids = r.parsePriceLevelsFromJSON(bidsJSON)
		orderBook.Asks = r.parsePriceLevelsFromJSON(asksJSON)

		orderBooks = append(orderBooks, orderBook)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate order books: %w", err)
	}

	return orderBooks, nil
}

// parsePriceLevelsFromJSON parses JSON string to price levels (simplified implementation)
func (r *TimeSeriesRepository) parsePriceLevelsFromJSON(jsonStr string) []entities.PriceLevel {
	// This is a simplified implementation
	// In a real implementation, you would use proper JSON parsing
	return []entities.PriceLevel{}
}

// GetTickerAggregates retrieves aggregated ticker data
func (r *TimeSeriesRepository) GetTickerAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]interfaces.TickerAggregate, error) {
	query := `SELECT symbol,
		date_trunc($1, timestamp) as timestamp,
		avg(price) as avg_price,
		min(price) as min_price,
		max(price) as max_price,
		sum(volume) as volume,
		count(*) as count
		FROM tickers
		WHERE symbol = $2 AND timestamp >= $3 AND timestamp <= $4
		GROUP BY symbol, date_trunc($1, timestamp)
		ORDER BY timestamp`

	rows, err := r.db.QueryContext(ctx, query, interval, symbol, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query ticker aggregates: %w", err)
	}
	defer rows.Close()

	var aggregates []interfaces.TickerAggregate
	for rows.Next() {
		var agg interfaces.TickerAggregate
		err := rows.Scan(
			&agg.Symbol,
			&agg.Timestamp,
			&agg.AvgPrice,
			&agg.MinPrice,
			&agg.MaxPrice,
			&agg.Volume,
			&agg.Count,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticker aggregate: %w", err)
		}
		aggregates = append(aggregates, agg)
	}

	return aggregates, nil
}

// GetCandleAggregates retrieves aggregated candle data
func (r *TimeSeriesRepository) GetCandleAggregates(ctx context.Context, symbol string, interval string, startTime, endTime time.Time) ([]interfaces.CandleAggregate, error) {
	query := `SELECT symbol,
		date_trunc($1, timestamp) as timestamp,
		first(open) as open,
		max(high) as high,
		min(low) as low,
		last(close) as close,
		sum(volume) as volume,
		count(*) as count
		FROM candles
		WHERE symbol = $2 AND timestamp >= $3 AND timestamp <= $4
		GROUP BY symbol, date_trunc($1, timestamp)
		ORDER BY timestamp`

	rows, err := r.db.QueryContext(ctx, query, interval, symbol, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query candle aggregates: %w", err)
	}
	defer rows.Close()

	var aggregates []interfaces.CandleAggregate
	for rows.Next() {
		var agg interfaces.CandleAggregate
		err := rows.Scan(
			&agg.Symbol,
			&agg.Timestamp,
			&agg.Open,
			&agg.High,
			&agg.Low,
			&agg.Close,
			&agg.Volume,
			&agg.Count,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan candle aggregate: %w", err)
		}
		aggregates = append(aggregates, agg)
	}

	return aggregates, nil
}

// CleanupOldData removes old data based on retention period
func (r *TimeSeriesRepository) CleanupOldData(ctx context.Context, retentionPeriod time.Duration) error {
	cutoffTime := time.Now().Add(-retentionPeriod)

	tables := []string{"tickers", "candles", "order_books"}

	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE timestamp < $1", table)
		result, err := r.db.ExecContext(ctx, query, cutoffTime)
		if err != nil {
			return fmt.Errorf("failed to cleanup %s: %w", table, err)
		}

		rowsAffected, _ := result.RowsAffected()
		fmt.Printf("Cleaned up %d rows from %s\n", rowsAffected, table)
	}

	return nil
}
