import assert from 'node:assert/strict';
import test from 'node:test';
import { createRequire } from 'node:module';
const require = createRequire(import.meta.url);
require('dotenv-expand').expand(require('dotenv').config());

const { buildServer } = await import('../dist/interfaces/http/server.js');

async function startTestServer() {
  const app = await buildServer();
  await app.listen({ port: 0, host: '127.0.0.1' });
  const addr = app.server.address();
  const port = typeof addr === 'object' && addr && 'port' in addr ? addr.port : 0;
  return { app, baseUrl: `http://127.0.0.1:${port}` };
}

test('ingest then read 1m and 5m', async () => {
  const { app, baseUrl } = await startTestServer();
  try {
    const authHeaders = { Authorization: `Bearer ${process.env.API_TOKEN || 'changeme'}` };
    const now = Date.now();
    const m0 = Math.floor((now - 10 * 60 * 1000) / 60000) * 60000; // t-10m aligned
    const body = {
      broker_system_name: 'demo',
      instrument_symbol: 'BTCUSDT',
      items: [
        { ts: new Date(m0).toISOString(), o: 100, h: 101, l: 99, c: 100.5, v: 10 },
        { ts: new Date(m0 + 60_000).toISOString(), o: 100.5, h: 102, l: 100, c: 101, v: 20 },
      ],
    };
    let resp = await fetch(`${baseUrl}/v1/ingest/ohlcv`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', ...authHeaders },
      body: JSON.stringify(body),
    });
    assert.equal(resp.status, 200);
    const ingestRes = await resp.json();
    assert.ok(typeof ingestRes.inserted === 'number');

    // Read 1m
    resp = await fetch(`${baseUrl}/v1/ohlcv?broker_system_name=demo&instrument_symbol=BTCUSDT&tf=1m`, { headers: authHeaders });
    assert.equal(resp.status, 200);
    const read1m = await resp.json();
    assert.ok(Array.isArray(read1m.items));
    assert.ok(read1m.items.length >= 2);
    assert.ok('ts' in read1m.items[0] && 'o' in read1m.items[0]);

    // Read 5m aggregated
    resp = await fetch(`${baseUrl}/v1/ohlcv?broker_system_name=demo&instrument_symbol=BTCUSDT&tf=5m`, { headers: authHeaders });
    assert.equal(resp.status, 200);
    const read5m = await resp.json();
    assert.ok(Array.isArray(read5m.items));
    assert.ok(read5m.items.length >= 1);
  } finally {
    await app.close();
  }
});


