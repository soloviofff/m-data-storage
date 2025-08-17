import { getPool } from '../db/client';
import { listBrokerInstrumentPairs } from '../repositories/registryRepository';
import { createTask } from '../repositories/taskRepository';

export interface GapSchedulerOptions {
	brokerId?: number;
	lookbackHours?: number; // how far back to search for gaps from now
}

// Compute missing minutes in [from, to] given a sorted array of existing minute timestamps (in ms)
function computeMissingMinuteRanges(
	existingMs: number[],
	fromMs: number,
	toMs: number,
): Array<{ from: number; to: number }> {
	const ranges: Array<{ from: number; to: number }> = [];
	let cursor = fromMs;
	const step = 60_000;
	const set = new Set(existingMs);
	while (cursor <= toMs) {
		if (!set.has(cursor)) {
			const start = cursor;
			while (cursor <= toMs && !set.has(cursor)) cursor += step;
			const end = cursor - step;
			ranges.push({ from: start, to: end });
		} else {
			cursor += step;
		}
	}
	return ranges;
}

export async function runGapScheduler(options: GapSchedulerOptions = {}): Promise<number> {
	const pool = getPool();
	const pairs = await listBrokerInstrumentPairs(options.brokerId);
	const lookbackMs = (options.lookbackHours ?? 24) * 60 * 60 * 1000;
	const now = Date.now();
	const fromMs = now - lookbackMs;
	let created = 0;

	for (const p of pairs) {
		const res = await pool.query(
			`SELECT FLOOR(EXTRACT(EPOCH FROM ts))::bigint * 1000 AS ms
       FROM timeseries.ohlcv
       WHERE broker_id = $1 AND instrument_id = $2
         AND ts >= to_timestamp(($3)::double precision/1000.0)
         AND ts <= to_timestamp(($4)::double precision/1000.0)
       ORDER BY ts ASC`,
			[p.broker_id, p.instrument_id, fromMs, now],
		);
		type MsRow = { ms: string | number };
		const existingMs: number[] = (res.rows as MsRow[]).map((r) => Number(r.ms));
		const gaps = computeMissingMinuteRanges(
			existingMs,
			Math.floor(fromMs / 60_000) * 60_000,
			Math.floor(now / 60_000) * 60_000,
		);
		for (const g of gaps) {
			const fromTs = new Date(g.from);
			const toTs = new Date(g.to);
			const idem = `gap:${p.broker_id}:${p.instrument_id}:${Math.floor(g.from / 60000)}-${Math.floor(g.to / 60000)}`;
			await createTask({
				brokerId: p.broker_id,
				instrumentId: p.instrument_id,
				fromTs,
				toTs,
				idempotencyKey: idem,
				priority: 1,
			});
			created += 1;
		}
	}
	return created;
}

export const __test_only__ = { computeMissingMinuteRanges };
