import { FastifyInstance } from 'fastify';
import { z } from 'zod';
import { getPool } from '../../../infrastructure/db/client';
import {
	aggregateBars,
	hasMoreData,
	readOneMinuteBars,
	type Timeframe,
} from '../../../infrastructure/repositories/ohlcvReadRepository';
// Swagger note: response schema omitted to avoid runtime JSON Schema issues

const querySchema = z.object({
	broker_system_name: z.string().min(1),
	instrument_symbol: z.string().min(1),
	tf: z.enum(['1m', '5m', '15m', '30m', '1h', '4h', '1d']).default('1m'),
	pageToken: z.string().optional(),
});

function decodePageToken(token?: string): Date | undefined {
	if (!token) return undefined;
	try {
		const json = JSON.parse(Buffer.from(token, 'base64url').toString('utf8')) as {
			last_ts: number;
		};
		if (!json || typeof json.last_ts !== 'number') return undefined;
		return new Date(json.last_ts * 60 * 1000);
	} catch {
		return undefined;
	}
}

function encodePageToken(date: Date): string {
	const payload = { last_ts: Math.floor(date.getTime() / 60000) };
	return Buffer.from(JSON.stringify(payload)).toString('base64url');
}

export async function registerReadRoutes(app: FastifyInstance) {
	app.get(
		'/v1/ohlcv',
		{ schema: { summary: 'Read OHLCV bars with optional aggregation', tags: ['read'] } },
		async (req, reply) => {
			const parsed = querySchema.safeParse(req.query);
			if (!parsed.success) {
				return reply.code(400).send({
					code: 'BAD_REQUEST',
					message: 'Invalid query',
					details: parsed.error.flatten(),
				});
			}
			const { broker_system_name, instrument_symbol, tf, pageToken } = parsed.data as {
				broker_system_name: string;
				instrument_symbol: string;
				tf: Timeframe;
				pageToken?: string;
			};

			// Resolve internal ids
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

			const now = new Date();
			const upperBound = !pageToken ? now : new Date(Date.now());
			const strictlyLessThanTs = decodePageToken(pageToken);
			const pageWindowMs = 24 * 60 * 60 * 1000;

			const rows = await readOneMinuteBars({
				brokerId,
				instrumentId,
				upperBound,
				pageWindowMs,
				strictlyLessThanTs,
			});
			const items = aggregateBars(rows, tf);
			let nextPageToken: string | undefined;
			if (rows.length > 0) {
				const lastTs = rows[rows.length - 1].ts;
				const more = await hasMoreData({
					brokerId,
					instrumentId,
					upperBound,
					pageWindowMs,
					strictlyLessThanTs: lastTs,
				});
				if (more) nextPageToken = encodePageToken(lastTs);
			}
			return { items, nextPageToken };
		},
	);
}
