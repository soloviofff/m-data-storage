-- TimescaleDB hypertable and policies for timeseries.ohlcv
SELECT create_hypertable('timeseries.ohlcv', by_range('ts'), if_not_exists => TRUE);

-- 7-day chunk interval
SELECT set_chunk_time_interval('timeseries.ohlcv', INTERVAL '7 days');

-- Compression policy: enable and apply for rows older than 30 days
ALTER TABLE timeseries.ohlcv SET (timescaledb.compress, timescaledb.compress_segmentby = 'broker_id,instrument_id');
SELECT add_compression_policy('timeseries.ohlcv', INTERVAL '30 days');

