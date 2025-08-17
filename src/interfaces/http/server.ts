import Fastify from 'fastify';
import { config } from '../../shared/config';
import { registerHealthRoutes } from './routes/health';
import { registerReadRoutes } from './routes/read';
import { registerIngestRoutes } from './routes/ingest';

export async function buildServer() {
	const prettyTransport =
		process.env.NODE_ENV === 'development' ? { target: 'pino-pretty' as const } : undefined;
	const app = Fastify({
		logger: { level: config.LOG_LEVEL, transport: prettyTransport },
	});

	await registerHealthRoutes(app);
	await registerReadRoutes(app);
	await registerIngestRoutes(app);

	// Global token validation for all protected endpoints (including read/write)
	app.addHook('onRequest', async (req, reply) => {
		// Allow health/ready without token for K8s/compose integrations
		if (req.url === '/health' || req.url === '/ready') return;
		const auth = req.headers['authorization'] || req.headers['x-api-key'];
		const token =
			typeof auth === 'string' && auth.startsWith('Bearer ')
				? auth.slice('Bearer '.length)
				: typeof auth === 'string'
					? auth
					: undefined;
		if (!token || token !== config.API_TOKEN) {
			return reply
				.code(401)
				.send({ code: 'UNAUTHORIZED', message: 'Invalid or missing token' });
		}
	});

	return app;
}

export async function startServer() {
	const app = await buildServer();
	await app.listen({ port: config.PORT, host: '0.0.0.0' });
	app.log.info({ port: config.PORT }, 'Server started');
}
