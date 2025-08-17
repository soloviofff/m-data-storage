import Fastify from 'fastify';
import { config } from '../../shared/config';
import { registerHealthRoutes } from './routes/health';
import { registerReadRoutes } from './routes/read';
import { registerIngestRoutes } from './routes/ingest';
import { registerTaskRoutes } from './routes/tasks';
import swagger from '@fastify/swagger';
import swaggerUi from '@fastify/swagger-ui';

export async function buildServer() {
	const prettyTransport =
		process.env.NODE_ENV === 'development' ? { target: 'pino-pretty' as const } : undefined;
	const app = Fastify({
		logger: { level: config.LOG_LEVEL, transport: prettyTransport },
	});

	await app.register(swagger, {
		openapi: {
			info: { title: 'm-data-storage API', version: '1.0.0' },
			servers: [{ url: config.SWAGGER_SERVER_URL }],
		},
	});
	await app.register(swaggerUi, {
		routePrefix: '/docs',
	});

	await registerHealthRoutes(app);
	await registerReadRoutes(app);
	await registerIngestRoutes(app);
	await registerTaskRoutes(app);

	// Global token validation for all protected endpoints (including read/write)
	app.addHook('onRequest', async (req, reply) => {
		// Apply auth only to API routes under /v1
		if (!req.url.startsWith('/v1')) return;
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
