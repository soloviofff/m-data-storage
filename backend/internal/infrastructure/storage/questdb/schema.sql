-- QuestDB schema for time series data storage
-- This file contains all table definitions for storing time series market data

-- Tickers table - stores real-time ticker data
CREATE TABLE IF NOT EXISTS tickers (
    symbol SYMBOL CAPACITY 1000 CACHE,
    broker_id SYMBOL CAPACITY 100 CACHE,
    market SYMBOL CAPACITY 10 CACHE,
    instrument_type SYMBOL CAPACITY 10 CACHE,
    price DOUBLE,
    bid_price DOUBLE,
    ask_price DOUBLE,
    volume DOUBLE,
    quote_volume DOUBLE,
    price_change_24h DOUBLE,
    price_change_percent_24h DOUBLE,
    high_24h DOUBLE,
    low_24h DOUBLE,
    open_24h DOUBLE,
    close_24h DOUBLE,
    trades_count_24h LONG,
    timestamp TIMESTAMP
) TIMESTAMP(timestamp) PARTITION BY DAY WAL;

-- Candles table - stores OHLCV candle data
CREATE TABLE IF NOT EXISTS candles (
    symbol SYMBOL CAPACITY 1000 CACHE,
    broker_id SYMBOL CAPACITY 100 CACHE,
    market SYMBOL CAPACITY 10 CACHE,
    instrument_type SYMBOL CAPACITY 10 CACHE,
    timeframe SYMBOL CAPACITY 20 CACHE,
    open_price DOUBLE,
    high_price DOUBLE,
    low_price DOUBLE,
    close_price DOUBLE,
    volume DOUBLE,
    quote_volume DOUBLE,
    trades_count LONG,
    taker_buy_volume DOUBLE,
    taker_buy_quote_volume DOUBLE,
    timestamp TIMESTAMP
) TIMESTAMP(timestamp) PARTITION BY DAY WAL;

-- Order books table - stores order book snapshots
CREATE TABLE IF NOT EXISTS order_books (
    symbol SYMBOL CAPACITY 1000 CACHE,
    broker_id SYMBOL CAPACITY 100 CACHE,
    market SYMBOL CAPACITY 10 CACHE,
    instrument_type SYMBOL CAPACITY 10 CACHE,
    side SYMBOL CAPACITY 2 CACHE, -- 'bid' or 'ask'
    price DOUBLE,
    quantity DOUBLE,
    level_index INT,
    timestamp TIMESTAMP
) TIMESTAMP(timestamp) PARTITION BY DAY WAL;

-- Trades table - stores individual trade data (optional, for detailed analysis)
CREATE TABLE IF NOT EXISTS trades (
    symbol SYMBOL CAPACITY 1000 CACHE,
    broker_id SYMBOL CAPACITY 100 CACHE,
    market SYMBOL CAPACITY 10 CACHE,
    instrument_type SYMBOL CAPACITY 10 CACHE,
    trade_id SYMBOL CAPACITY 10000000 CACHE,
    price DOUBLE,
    quantity DOUBLE,
    quote_quantity DOUBLE,
    side SYMBOL CAPACITY 2 CACHE, -- 'buy' or 'sell'
    is_maker BOOLEAN,
    timestamp TIMESTAMP
) TIMESTAMP(timestamp) PARTITION BY DAY WAL;

-- Aggregated statistics table - stores pre-calculated aggregations
CREATE TABLE IF NOT EXISTS ticker_aggregates (
    symbol SYMBOL CAPACITY 1000 CACHE,
    broker_id SYMBOL CAPACITY 100 CACHE,
    market SYMBOL CAPACITY 10 CACHE,
    instrument_type SYMBOL CAPACITY 10 CACHE,
    interval_type SYMBOL CAPACITY 10 CACHE, -- '1m', '5m', '1h', '1d', etc.
    avg_price DOUBLE,
    min_price DOUBLE,
    max_price DOUBLE,
    total_volume DOUBLE,
    total_quote_volume DOUBLE,
    trades_count LONG,
    price_change DOUBLE,
    price_change_percent DOUBLE,
    timestamp TIMESTAMP
) TIMESTAMP(timestamp) PARTITION BY DAY WAL;

-- Candle aggregates table - stores pre-calculated candle aggregations
CREATE TABLE IF NOT EXISTS candle_aggregates (
    symbol SYMBOL CAPACITY 1000 CACHE,
    broker_id SYMBOL CAPACITY 100 CACHE,
    market SYMBOL CAPACITY 10 CACHE,
    instrument_type SYMBOL CAPACITY 10 CACHE,
    source_timeframe SYMBOL CAPACITY 20 CACHE,
    target_timeframe SYMBOL CAPACITY 20 CACHE,
    open_price DOUBLE,
    high_price DOUBLE,
    low_price DOUBLE,
    close_price DOUBLE,
    volume DOUBLE,
    quote_volume DOUBLE,
    trades_count LONG,
    timestamp TIMESTAMP
) TIMESTAMP(timestamp) PARTITION BY DAY WAL;
