-- Rollback for: Initial Schema
-- Description: Drop all initial tables and indexes
-- Created: 2024-07-04

-- Drop triggers
DROP TRIGGER IF EXISTS system_config_updated_at;
DROP TRIGGER IF EXISTS broker_configs_updated_at;
DROP TRIGGER IF EXISTS subscriptions_updated_at;
DROP TRIGGER IF EXISTS instruments_updated_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_broker_configs_enabled;
DROP INDEX IF EXISTS idx_broker_configs_type;
DROP INDEX IF EXISTS idx_subscriptions_active;
DROP INDEX IF EXISTS idx_subscriptions_broker_id;
DROP INDEX IF EXISTS idx_subscriptions_instrument_id;
DROP INDEX IF EXISTS idx_instruments_active;
DROP INDEX IF EXISTS idx_instruments_market;
DROP INDEX IF EXISTS idx_instruments_type;
DROP INDEX IF EXISTS idx_instruments_broker_id;
DROP INDEX IF EXISTS idx_instruments_symbol;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS system_config;
DROP TABLE IF EXISTS broker_configs;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS instruments;
