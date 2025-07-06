-- Migration: Initial Schema
-- Description: Create initial tables for metadata storage
-- Created: 2024-07-04

-- Enable foreign keys
PRAGMA foreign_keys = ON;

-- Instruments table
CREATE TABLE IF NOT EXISTS instruments (
    id TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('spot', 'futures', 'stock', 'etf', 'bond')),
    market TEXT NOT NULL CHECK (market IN ('crypto', 'stock', 'forex', 'commodity')),
    base_asset TEXT,
    quote_asset TEXT,
    price_precision INTEGER NOT NULL DEFAULT 8,
    quantity_precision INTEGER NOT NULL DEFAULT 8,
    min_quantity REAL,
    max_quantity REAL,
    step_size REAL,
    tick_size REAL,
    settings TEXT, -- JSON
    broker_id TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(symbol, broker_id)
);

-- Subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id TEXT PRIMARY KEY,
    instrument_id TEXT NOT NULL,
    data_types TEXT NOT NULL, -- JSON array
    timeframes TEXT, -- JSON array, nullable for non-candle data
    settings TEXT, -- JSON
    broker_id TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (instrument_id) REFERENCES instruments(id) ON DELETE CASCADE
);

-- Broker configurations table
CREATE TABLE IF NOT EXISTS broker_configs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('crypto', 'stock')),
    connection_config TEXT NOT NULL, -- JSON
    auth_config TEXT NOT NULL, -- JSON
    rate_limits TEXT, -- JSON
    settings TEXT, -- JSON
    is_enabled BOOLEAN NOT NULL DEFAULT false,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- System configuration table
CREATE TABLE IF NOT EXISTS system_config (
    id INTEGER PRIMARY KEY CHECK (id = 1), -- Only one row allowed
    storage_retention INTEGER NOT NULL, -- in seconds
    vacuum_interval INTEGER NOT NULL,   -- in seconds
    max_storage_size INTEGER NOT NULL,  -- in bytes
    api_port INTEGER NOT NULL,
    api_host TEXT NOT NULL,
    read_timeout INTEGER NOT NULL,      -- in seconds
    write_timeout INTEGER NOT NULL,     -- in seconds
    shutdown_timeout INTEGER NOT NULL,   -- in seconds
    jwt_secret TEXT NOT NULL,
    api_key_header TEXT NOT NULL,
    allowed_origins TEXT NOT NULL,      -- JSON array
    metrics_enabled BOOLEAN NOT NULL,
    metrics_port INTEGER,
    tracing_enabled BOOLEAN NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_instruments_symbol ON instruments(symbol);
CREATE INDEX IF NOT EXISTS idx_instruments_broker_id ON instruments(broker_id);
CREATE INDEX IF NOT EXISTS idx_instruments_type ON instruments(type);
CREATE INDEX IF NOT EXISTS idx_instruments_market ON instruments(market);
CREATE INDEX IF NOT EXISTS idx_instruments_active ON instruments(is_active);

CREATE INDEX IF NOT EXISTS idx_subscriptions_instrument_id ON subscriptions(instrument_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_broker_id ON subscriptions(broker_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_active ON subscriptions(is_active);

CREATE INDEX IF NOT EXISTS idx_broker_configs_type ON broker_configs(type);
CREATE INDEX IF NOT EXISTS idx_broker_configs_enabled ON broker_configs(is_enabled);

-- Triggers for automatic timestamp updates
CREATE TRIGGER IF NOT EXISTS instruments_updated_at
AFTER UPDATE ON instruments
BEGIN
    UPDATE instruments SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS subscriptions_updated_at
AFTER UPDATE ON subscriptions
BEGIN
    UPDATE subscriptions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS broker_configs_updated_at
AFTER UPDATE ON broker_configs
BEGIN
    UPDATE broker_configs SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS system_config_updated_at
AFTER UPDATE ON system_config
BEGIN
    UPDATE system_config SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
