-- Initialize database: enable TimescaleDB extension and create logical schemas
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Logical separation: registry (metadata) and timeseries (OHLCV hypertable)
CREATE SCHEMA IF NOT EXISTS registry;
CREATE SCHEMA IF NOT EXISTS timeseries;


