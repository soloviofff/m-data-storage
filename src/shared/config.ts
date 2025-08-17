import dotenv from 'dotenv';
import dotenvExpand from 'dotenv-expand';
const env = dotenv.config();
dotenvExpand.expand(env);
import { z } from 'zod';

const envSchema = z.object({
	NODE_ENV: z.enum(['development', 'test', 'production']).default('development'),
	PORT: z.coerce.number().int().positive().default(8080),
	LOG_LEVEL: z
		.enum(['fatal', 'error', 'warn', 'info', 'debug', 'trace', 'silent'])
		.default('info'),
	API_TOKEN: z.string().min(1),
	DATABASE_URL: z.string().min(1),
	FINALIZATION_WINDOW_HOURS: z.coerce.number().int().positive().default(24),
	SUPPORTED_TIMEFRAMES: z.string().default('1m,5m,15m,30m,1h,4h,1d'),
	SWAGGER_SERVER_URL: z.string().url().default('http://localhost:8080'),
});

const parsed = envSchema.safeParse(process.env);
if (!parsed.success) {
	console.error('Invalid environment variables', parsed.error.flatten());
	process.exit(1);
}

export const config = {
	...parsed.data,
	supportedTimeframes: parsed.data.SUPPORTED_TIMEFRAMES.split(','),
};
