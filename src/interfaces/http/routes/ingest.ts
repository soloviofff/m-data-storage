import { FastifyInstance } from 'fastify';
import { z } from 'zod';
import { getPool } from '../../../infrastructure/db/client';
import {
	ingestOhlcvBatch,
	type IngestItem,
} from '../../../infrastructure/repositories/ohlcvIngestRepository';

const bodySchema = z.object({
	broker_system_name: z.string().min(1),
	instrument_symbol: z.string().min(1),
	items: z
		.array(
			z.object({
				ts: z.string(),
				o: z.number(),
				h: z.number(),
				l: z.number(),
				c: z.number(),
				v: z.number(),
			}),
		)
		.min(1),
});

export async function registerIngestRoutes(app: FastifyInstance) {
	app.post(
		'/v1/ingest/ohlcv',
		{ schema: { summary: 'Ingest OHLCV batch', tags: ['ingest'] } },
		async (req, reply) => {
			const parsed = bodySchema.safeParse(req.body);
			if (!parsed.success) {
				return reply.code(400).send({
					code: 'BAD_REQUEST',
					message: 'Invalid body',
					details: parsed.error.flatten(),
				});
			}
			const { broker_system_name, instrument_symbol, items } = parsed.data as {
				broker_system_name: string;
				instrument_symbol: string;
				items: IngestItem[];
			};
			// Resolve internal ids from public identifiers
			const pool = getPool();
			const br = await pool.query(
				'SELECT id FROM registry.brokers WHERE system_name = $1 AND is_active = true',
				[broker_system_name],
			);
			if ((br.rowCount ?? 0) === 0) {
				return reply
					.code(400)
					.send({ code: 'BAD_REQUEST', message: 'Unknown broker_system_name' });
			}
			const brokerId: number = br.rows[0].id;
			const ins = await pool.query(
				'SELECT id FROM registry.instruments WHERE broker_id = $1 AND symbol = $2 AND is_active = true',
				[brokerId, instrument_symbol],
			);
			if ((ins.rowCount ?? 0) === 0) {
				return reply
					.code(400)
					.send({ code: 'BAD_REQUEST', message: 'Unknown instrument_symbol' });
			}
			const instrumentId: number = ins.rows[0].id;
			try {
				const res = await ingestOhlcvBatch(brokerId, instrumentId, items);
				return res;
			} catch (err) {
				const message = err instanceof Error ? err.message : 'Bad ingest payload';
				return reply.code(400).send({ code: 'INVALID_INGEST', message });
			}
		},
	);
}
