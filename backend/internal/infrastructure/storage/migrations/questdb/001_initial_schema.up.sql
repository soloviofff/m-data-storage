-- Migration: Initial Schema
-- Description: Create initial tables for time series data storage
-- Created: 2024-07-04

-- Tickers table for real-time price data
CREATE TABLE IF NOT EXISTS tickers (
    symbol SYMBOL CAPACITY 1000 CACHE,
    broker_id SYMBOL CAPACITY 100 CACHE,
    bid_price DOUBLE,
    ask_price DOUBLE,
    bid_size DOUBLE,
    ask_size DOUBLE,
    last_price DOUBLE,
    volume_24h DOUBLE,
    price_change_24h DOUBLE,
    price_change_percent_24h DOUBLE,
    high_24h DOUBLE,
    low_24h DOUBLE,
    open_24h DOUBLE,
    timestamp TIMESTAMP
) TIMESTAMP(timestamp) PARTITION BY DAY WAL;

-- Candles table for OHLCV data
CREATE TABLE IF NOT EXISTS candles (
    symbol SYMBOL CAPACITY 1000 CACHE,
    broker_id SYMBOL CAPACITY 100 CACHE,
    timeframe SYMBOL CAPACITY 20 CACHE,
    open_price DOUBLE,
    high_price DOUBLE,
    low_price DOUBLE,
    close_price DOUBLE,
    volume DOUBLE,
    quote_volume DOUBLE,
    trades_count LONG,
    open_interest DOUBLE,
    timestamp TIMESTAMP
) TIMESTAMP(timestamp) PARTITION BY DAY WAL;

-- Order books table for market depth data
CREATE TABLE IF NOT EXISTS order_books (
    symbol SYMBOL CAPACITY 1000 CACHE,
    broker_id SYMBOL CAPACITY 100 CACHE,
    side SYMBOL CAPACITY 2 CACHE, -- 'bid' or 'ask'
    price DOUBLE,
    quantity DOUBLE,
    level_index INT,
    update_id LONG,
    timestamp TIMESTAMP
) TIMESTAMP(timestamp) PARTITION BY DAY WAL;

-- Trades table for executed trades
CREATE TABLE IF NOT EXISTS trades (
    symbol SYMBOL CAPACITY 1000 CACHE,
    broker_id SYMBOL CAPACITY 100 CACHE,
    trade_id STRING,
    price DOUBLE,
    quantity DOUBLE,
    side SYMBOL CAPACITY 2 CACHE, -- 'buy' or 'sell'
    is_maker BOOLEAN,
    timestamp TIMESTAMP
) TIMESTAMP(timestamp) PARTITION BY DAY WAL;

-- Aggregated data table for pre-calculated metrics
CREATE TABLE IF NOT EXISTS aggregates (
    symbol SYMBOL CAPACITY 1000 CACHE,
    broker_id SYMBOL CAPACITY 100 CACHE,
    data_type SYMBOL CAPACITY 10 CACHE, -- 'volume', 'price', 'trades', etc.
    timeframe SYMBOL CAPACITY 20 CACHE,
    value DOUBLE,
    count LONG,
    min_value DOUBLE,
    max_value DOUBLE,
    avg_value DOUBLE,
    timestamp TIMESTAMP
) TIMESTAMP(timestamp) PARTITION BY DAY WAL;
