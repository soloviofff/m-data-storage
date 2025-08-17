import { getPool } from '../db/client';
import { config } from '../../shared/config';

export interface IngestItem {
	ts: string; // ISO minute-aligned UTC string
	o: number;
	h: number;
	l: number;
	c: number;
	v: number;
}

export interface IngestResult {
	inserted: number;
	updated: number;
	skipped: number;
}

function assertMinuteAligned(date: Date) {
	if (date.getUTCSeconds() !== 0 || date.getUTCMilliseconds() !== 0) {
		throw new Error('Timestamps must be aligned to exact minutes (ss=00, ms=000)');
	}
}

function ensureContinuousMinutes(dates: Date[]) {
	for (let i = 1; i < dates.length; i += 1) {
		const prev = dates[i - 1].getTime();
		const curr = dates[i].getTime();
		if (curr - prev !== 60_000) {
			throw new Error('Items must form a continuous range of minutes without gaps');
		}
	}
}

export async function ingestOhlcvBatch(
	brokerId: number,
	instrumentId: number,
	items: IngestItem[],
): Promise<IngestResult> {
	if (!Array.isArray(items) || items.length === 0) {
		throw new Error('Items array must be non-empty');
	}
	// Parse and validate timestamps
	const parsed = items.map((it) => {
		const d = new Date(it.ts);
		if (Number.isNaN(d.getTime())) throw new Error(`Invalid timestamp: ${it.ts}`);
		assertMinuteAligned(d);
		return { d, it };
	});
	// Ensure strictly increasing and continuous
	parsed.sort((a, b) => a.d.getTime() - b.d.getTime());
	ensureContinuousMinutes(parsed.map((p) => p.d));

	const now = Date.now();
	const windowMs = Number(config.FINALIZATION_WINDOW_HOURS) * 60 * 60 * 1000;
	const windowStart = now - windowMs;

	const recent: typeof parsed = [];
	const historical: typeof parsed = [];
	for (const p of parsed) {
		if (p.d.getTime() >= windowStart) recent.push(p);
		else historical.push(p);
	}

	const pool = getPool();
	let inserted = 0;
	let updated = 0;
	let skipped = 0;

	// Historical: DO NOTHING on conflict, use RETURNING to count inserted
	if (historical.length > 0) {
		const values: Array<number | string> = [];
		const tuples = historical
			.map((p, idx) => {
				const base = idx * 8;
				values.push(
					brokerId,
					instrumentId,
					p.d.toISOString(),
					p.it.o,
					p.it.h,
					p.it.l,
					p.it.c,
					p.it.v,
				);
				return `($${base + 1}, $${base + 2}, $${base + 3}, $${base + 4}, $${base + 5}, $${base + 6}, $${base + 7}, $${base + 8})`;
			})
			.join(',');
		const sql = `
			INSERT INTO timeseries.ohlcv (broker_id, instrument_id, ts, o, h, l, c, v)
			VALUES ${tuples}
			ON CONFLICT (broker_id, instrument_id, ts) DO NOTHING
			RETURNING 1
		`;
		const res = await pool.query(sql, values);
		inserted += res.rowCount ?? 0;
		skipped += historical.length - (res.rowCount ?? 0);
	}

	// Recent: DO UPDATE on conflict, differentiate inserted vs updated via xmax
	if (recent.length > 0) {
		const values: Array<number | string> = [];
		const tuples = recent
			.map((p, idx) => {
				const base = idx * 8;
				values.push(
					brokerId,
					instrumentId,
					p.d.toISOString(),
					p.it.o,
					p.it.h,
					p.it.l,
					p.it.c,
					p.it.v,
				);
				return `($${base + 1}, $${base + 2}, $${base + 3}, $${base + 4}, $${base + 5}, $${base + 6}, $${base + 7}, $${base + 8})`;
			})
			.join(',');
		const sql = `
			INSERT INTO timeseries.ohlcv (broker_id, instrument_id, ts, o, h, l, c, v)
			VALUES ${tuples}
			ON CONFLICT (broker_id, instrument_id, ts) DO UPDATE SET
				o = EXCLUDED.o,
				h = EXCLUDED.h,
				l = EXCLUDED.l,
				c = EXCLUDED.c,
				v = EXCLUDED.v
			RETURNING (xmax = 0) AS inserted
		`;
		const res = await pool.query(sql, values);
		const rows = (res.rows ?? []) as Array<{ inserted: boolean }>;
		for (const r of rows) {
			if (r.inserted) inserted += 1;
			else updated += 1;
		}
	}

	return { inserted, updated, skipped };
}
