import { FastifyInstance } from 'fastify';
import { z } from 'zod';
import {
	aggregateBars,
	hasMoreData,
	readOneMinuteBars,
	type Timeframe,
} from '../../../infrastructure/repositories/ohlcvReadRepository';

const querySchema = z.object({
	broker_id: z.coerce.number().int().positive(),
	instrument_id: z.coerce.number().int().positive(),
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
	app.get('/v1/ohlcv', async (req, reply) => {
		const parsed = querySchema.safeParse(req.query);
		if (!parsed.success) {
			return reply
				.code(400)
				.send({
					code: 'BAD_REQUEST',
					message: 'Invalid query',
					details: parsed.error.flatten(),
				});
		}
		const { broker_id, instrument_id, tf, pageToken } = parsed.data as {
			broker_id: number;
			instrument_id: number;
			tf: Timeframe;
			pageToken?: string;
		};

		const now = new Date();
		const upperBound = !pageToken ? now : new Date(Date.now());
		const strictlyLessThanTs = decodePageToken(pageToken);
		const pageWindowMs = 24 * 60 * 60 * 1000;

		const rows = await readOneMinuteBars({
			brokerId: broker_id,
			instrumentId: instrument_id,
			upperBound,
			pageWindowMs,
			strictlyLessThanTs,
		});
		const items = aggregateBars(rows, tf);
		let nextPageToken: string | undefined;
		if (rows.length > 0) {
			const lastTs = rows[rows.length - 1].ts;
			const more = await hasMoreData({
				brokerId: broker_id,
				instrumentId: instrument_id,
				upperBound,
				pageWindowMs,
				strictlyLessThanTs: lastTs,
			});
			if (more) nextPageToken = encodePageToken(lastTs);
		}
		return { items, nextPageToken };
	});
}
