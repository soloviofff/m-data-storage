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

test('health/docs are public, v1 requires token', async () => {
  const { app, baseUrl } = await startTestServer();
  try {
    let resp = await fetch(`${baseUrl}/health`);
    assert.equal(resp.status, 200);
    resp = await fetch(`${baseUrl}/docs`);
    assert.equal([200, 302].includes(resp.status), true);

    resp = await fetch(`${baseUrl}/v1/ohlcv?broker_system_name=demo&instrument_symbol=BTCUSDT`);
    assert.equal(resp.status, 401);
  } finally {
    await app.close();
  }
});


