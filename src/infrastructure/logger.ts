import pino from 'pino';
import { config } from '../shared/config';

function resolvePrettyTransport() {
	if (process.env.NODE_ENV !== 'development') return undefined;
	try {
		// Check whether pino-pretty is available
		// eslint-disable-next-line @typescript-eslint/no-var-requires
		require('pino-pretty');
		return { target: 'pino-pretty' as const };
	} catch {
		return undefined;
	}
}

export const logger = pino({
	level: config.LOG_LEVEL,
	transport: resolvePrettyTransport(),
});
