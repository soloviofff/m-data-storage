import { Pool } from 'pg';
import { drizzle } from 'drizzle-orm/node-postgres';
import * as schema from './schema';
import { config } from '../../shared/config';

let pool: Pool | null = null;

export function getPool(): Pool {
	if (!pool) {
		pool = new Pool({ connectionString: config.DATABASE_URL });
	}
	return pool;
}

export function getDb() {
	return drizzle(getPool(), { schema });
}
