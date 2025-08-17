import type { Config } from 'drizzle-kit';
import dotenv from 'dotenv';
import dotenvExpand from 'dotenv-expand';
const env = dotenv.config();
dotenvExpand.expand(env);

if (!process.env.DATABASE_URL) {
	throw new Error('DATABASE_URL is not set');
}

export default {
	dialect: 'postgresql',
	schema: './src/infrastructure/db/schema.ts',
	out: './drizzle/migrations',
	dbCredentials: {
		url: process.env.DATABASE_URL!,
	},
} satisfies Config;
