-- SQLite schema for metadata storage
-- This file contains all table definitions for storing metadata

-- Instruments table - stores information about trading instruments
CREATE TABLE IF NOT EXISTS instruments (
    symbol TEXT PRIMARY KEY,
    type TEXT NOT NULL, -- 'spot', 'futures', 'stock', 'etf', 'bond'
    market TEXT NOT NULL, -- 'spot', 'futures', 'stock'
    base_asset TEXT,
    quote_asset TEXT,
    min_price REAL,
    max_price REAL,
    min_quantity REAL,
    max_quantity REAL,
    price_precision INTEGER,
    quantity_precision INTEGER,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Subscriptions table - stores active data subscriptions
CREATE TABLE IF NOT EXISTS subscriptions (
    id TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    type TEXT NOT NULL, -- 'spot', 'futures', 'stock', 'etf', 'bond'
    market TEXT NOT NULL, -- 'spot', 'futures', 'stock'
    data_types TEXT NOT NULL, -- JSON array of data types
    start_date DATETIME NOT NULL,
    end_date DATETIME,
    settings TEXT, -- JSON settings
    broker_id TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (symbol) REFERENCES instruments(symbol) ON DELETE CASCADE
);

-- Broker configurations table - stores broker connection configs
CREATE TABLE IF NOT EXISTS broker_configs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- 'crypto', 'stock'
    enabled BOOLEAN NOT NULL DEFAULT false,
    config_json TEXT NOT NULL, -- Full broker config as JSON
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

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_instruments_type ON instruments(type);
CREATE INDEX IF NOT EXISTS idx_instruments_market ON instruments(market);
CREATE INDEX IF NOT EXISTS idx_instruments_active ON instruments(is_active);

CREATE INDEX IF NOT EXISTS idx_subscriptions_broker_id ON subscriptions(broker_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_symbol ON subscriptions(symbol);
CREATE INDEX IF NOT EXISTS idx_subscriptions_active ON subscriptions(is_active);
CREATE INDEX IF NOT EXISTS idx_subscriptions_start_date ON subscriptions(start_date);

CREATE INDEX IF NOT EXISTS idx_broker_configs_type ON broker_configs(type);
CREATE INDEX IF NOT EXISTS idx_broker_configs_enabled ON broker_configs(enabled);

-- Triggers for automatic updated_at timestamps
CREATE TRIGGER IF NOT EXISTS instruments_updated_at
AFTER UPDATE ON instruments
BEGIN
    UPDATE instruments SET updated_at = CURRENT_TIMESTAMP WHERE symbol = NEW.symbol;
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
