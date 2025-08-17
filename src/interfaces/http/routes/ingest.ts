import { FastifyInstance } from 'fastify';
import { z } from 'zod';
import {
	ingestOhlcvBatch,
	type IngestItem,
} from '../../../infrastructure/repositories/ohlcvIngestRepository';

const bodySchema = z.object({
	broker_id: z.coerce.number().int().positive(),
	instrument_id: z.coerce.number().int().positive(),
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
			const { broker_id, instrument_id, items } = parsed.data as {
				broker_id: number;
				instrument_id: number;
				items: IngestItem[];
			};
			try {
				const res = await ingestOhlcvBatch(broker_id, instrument_id, items);
				return res;
			} catch (err) {
				const message = err instanceof Error ? err.message : 'Bad ingest payload';
				return reply.code(400).send({ code: 'INVALID_INGEST', message });
			}
		},
	);
}
