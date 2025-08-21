const { readFileSync, readdirSync } = require('fs');
const { Pool } = require('pg');
const path = require('path');
require('dotenv-expand').expand(require('dotenv').config());

async function run() {
  const pool = new Pool({ connectionString: process.env.DATABASE_URL_INNER });
  const client = await pool.connect();
  try {
    // Ensure required schemas and extension exist before applying migrations
    await client.query('CREATE SCHEMA IF NOT EXISTS registry');
    await client.query('CREATE SCHEMA IF NOT EXISTS timeseries');
    await client.query('CREATE EXTENSION IF NOT EXISTS timescaledb');

    // Create a simple migrations ledger
    await client.query(
      `CREATE TABLE IF NOT EXISTS registry.migrations (
         filename text PRIMARY KEY,
         applied_at timestamptz NOT NULL DEFAULT now()
       )`
    );
    const migrationsDir = path.join(__dirname, '..', 'drizzle', 'migrations');
    const files = readdirSync(migrationsDir)
      .filter((f) => f.endsWith('.sql'))
      .sort();
    for (const file of files) {
      const check = await client.query('SELECT 1 FROM registry.migrations WHERE filename = $1', [file]);
      if ((check.rowCount ?? 0) > 0) {
        continue; // already applied
      }
      const sql = readFileSync(path.join(migrationsDir, file), 'utf8');
      await client.query(sql);
      await client.query('INSERT INTO registry.migrations (filename) VALUES ($1)', [file]);
      console.log(`Applied migration: ${file}`);
    }
  } finally {
    client.release();
    await pool.end();
  }
}

run().catch((err) => {
  console.error(err);
  process.exit(1);
});
