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

-- Broker configuration table
CREATE TABLE IF NOT EXISTS broker_config (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT false,
    config_json TEXT NOT NULL,          -- Full broker config as JSON
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Update trigger for system_config
CREATE TRIGGER IF NOT EXISTS system_config_updated_at
AFTER UPDATE ON system_config
BEGIN
    UPDATE system_config SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Update trigger for broker_config
CREATE TRIGGER IF NOT EXISTS broker_config_updated_at
AFTER UPDATE ON broker_config
BEGIN
    UPDATE broker_config SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 