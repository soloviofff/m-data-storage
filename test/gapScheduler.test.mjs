import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
const require = createRequire(import.meta.url);
require('dotenv-expand').expand(require('dotenv').config());

const { __test_only__, runGapScheduler } = await import('../dist/infrastructure/services/gapScheduler.js');
const { getPool } = await import('../dist/infrastructure/db/client.js');

test('computeMissingMinuteRanges finds gaps', () => {
  const now = Date.now();
  const start = Math.floor((now - 10 * 60 * 1000) / 60000) * 60000; // 10 minutes ago
  const existing = [start, start + 60_000, start + 3 * 60_000]; // missing minute 2 and 4..10
  const ranges = __test_only__.computeMissingMinuteRanges(existing, start, start + 9 * 60_000);
  assert.equal(ranges.length > 0, true);
});

test('gap scheduler creates tasks for missing minutes', async () => {
  const pool = getPool();
  // Ensure is_active columns exist for tests
  await pool.query("ALTER TABLE registry.brokers ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true NOT NULL");
  await pool.query("ALTER TABLE registry.instruments ADD COLUMN IF NOT EXISTS is_active boolean DEFAULT true NOT NULL");
  // Ensure broker/instrument exist with correct relation
  await pool.query("INSERT INTO registry.brokers (id, system_name, name) VALUES (1, 'demo', 'Demo') ON CONFLICT DO NOTHING");
  await pool.query("INSERT INTO registry.instruments (id, broker_id, symbol, name) VALUES (1, 1, 'BTCUSDT', 'BTC') ON CONFLICT DO NOTHING");

  // Insert sparse data: 2 minutes present within last 10 minutes
  const base = Math.floor((Date.now() - 10 * 60 * 1000) / 60000) * 60000;
  await pool.query('DELETE FROM timeseries.ohlcv WHERE broker_id = 1 AND instrument_id = 1 AND ts >= to_timestamp(($1)::double precision/1000.0)', [base]);
  await pool.query('INSERT INTO timeseries.ohlcv (broker_id, instrument_id, ts, o, h, l, c, v) VALUES ($1,$2,to_timestamp(($3)::double precision/1000.0),$4,$5,$6,$7,$8)', [1,1, base, 1,1,1,1,1]);
  await pool.query('INSERT INTO timeseries.ohlcv (broker_id, instrument_id, ts, o, h, l, c, v) VALUES ($1,$2,to_timestamp(($3)::double precision/1000.0),$4,$5,$6,$7,$8)', [1,1, base + 60_000, 1,1,1,1,1]);

  const created = await runGapScheduler({ brokerId: 1, lookbackHours: 1 });
  assert.equal(created > 0, true);
});


