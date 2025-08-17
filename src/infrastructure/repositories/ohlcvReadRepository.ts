import { getPool } from '../db/client';

export type Timeframe = '1m' | '5m' | '15m' | '30m' | '1h' | '4h' | '1d';

export interface OhlcvRow {
	ts: Date;
	o: number | null;
	h: number | null;
	l: number | null;
	c: number | null;
	v: number | null;
}

export interface ReadPageParams {
	brokerId: number;
	instrumentId: number;
	upperBound: Date; // inclusive upper bound
	pageWindowMs: number; // typically 24h
	strictlyLessThanTs?: Date; // for keyset pagination: ts < last_ts
}

export async function readOneMinuteBars(params: ReadPageParams): Promise<OhlcvRow[]> {
	const { brokerId, instrumentId, upperBound, pageWindowMs, strictlyLessThanTs } = params;
	const lowerBound = new Date(upperBound.getTime() - pageWindowMs);
	const pool = getPool();
	const values: Array<number | Date> = [brokerId, instrumentId, lowerBound, upperBound];
	let where = 'broker_id = $1 AND instrument_id = $2 AND ts > $3 AND ts <= $4';
	if (strictlyLessThanTs) {
		values.push(strictlyLessThanTs);
		where += ` AND ts < $${values.length}`;
	}
	const sql = `
		SELECT ts, o, h, l, c, v
		FROM timeseries.ohlcv
		WHERE ${where}
		ORDER BY ts DESC
	`;
	const res = await pool.query(sql, values);
	type PgRow = {
		ts: string | Date;
		o: number | null;
		h: number | null;
		l: number | null;
		c: number | null;
		v: number | null;
	};
	return (res.rows as PgRow[]).map((r) => ({
		ts: new Date(r.ts),
		o: r.o,
		h: r.h,
		l: r.l,
		c: r.c,
		v: r.v,
	}));
}

export function timeframeToMinutes(tf: Timeframe): number {
	switch (tf) {
		case '1m':
			return 1;
		case '5m':
			return 5;
		case '15m':
			return 15;
		case '30m':
			return 30;
		case '1h':
			return 60;
		case '4h':
			return 240;
		case '1d':
			return 1440;
	}
}

export function floorToBucketStart(ts: Date, bucketMinutes: number): number {
	const epochSec = Math.floor(ts.getTime() / 1000);
	const minuteSec = Math.floor(epochSec / 60) * 60;
	const bucketSec = Math.floor(minuteSec / 60 / bucketMinutes) * bucketMinutes * 60;
	return bucketSec;
}

export interface OhlcvItem {
	ts: string; // ISO string in UTC
	o: number;
	h: number;
	l: number;
	c: number;
	v: number;
}

export function aggregateBars(rows: OhlcvRow[], tf: Timeframe): OhlcvItem[] {
	if (tf === '1m') {
		return rows.map((r) => ({
			ts: new Date(Math.floor(r.ts.getTime() / 60000) * 60000).toISOString(),
			o: r.o ?? 0,
			h: r.h ?? 0,
			l: r.l ?? 0,
			c: r.c ?? 0,
			v: r.v ?? 0,
		}));
	}
	const bucketMinutes = timeframeToMinutes(tf);
	const groups = new Map<
		number,
		{
			ts: number;
			o: number;
			h: number;
			l: number;
			c: number;
			v: number;
			firstTs: number;
			lastTs: number;
		}
	>();
	// Iterate in ascending time to compute o/c correctly
	const asc = [...rows].sort((a, b) => a.ts.getTime() - b.ts.getTime());
	for (const r of asc) {
		const bucketStartSec = floorToBucketStart(r.ts, bucketMinutes);
		const g = groups.get(bucketStartSec) ?? {
			ts: bucketStartSec * 1000,
			o: r.o ?? 0,
			h: r.h ?? 0,
			l: r.l ?? 0,
			c: r.c ?? 0,
			v: r.v ?? 0,
			firstTs: r.ts.getTime(),
			lastTs: r.ts.getTime(),
		};
		if (!groups.has(bucketStartSec)) {
			groups.set(bucketStartSec, g);
			continue;
		}
		// Update aggregates
		g.h = Math.max(g.h, r.h ?? g.h);
		g.l = Math.min(g.l, r.l ?? g.l);
		g.v += r.v ?? 0;
		// Open: first by time
		if (r.ts.getTime() < g.firstTs) {
			g.firstTs = r.ts.getTime();
			g.o = r.o ?? g.o;
		}
		// Close: last by time
		if (r.ts.getTime() >= g.lastTs) {
			g.lastTs = r.ts.getTime();
			g.c = r.c ?? g.c;
		}
	}
	const items = Array.from(groups.values())
		.sort((a, b) => b.ts - a.ts)
		.map((g) => ({
			ts: new Date(g.ts).toISOString(),
			o: g.o,
			h: g.h,
			l: g.l,
			c: g.c,
			v: g.v,
		}));
	return items;
}

export async function hasMoreData(params: ReadPageParams): Promise<boolean> {
	const { brokerId, instrumentId, upperBound, pageWindowMs, strictlyLessThanTs } = params;
	const lowerBound = new Date(upperBound.getTime() - pageWindowMs);
	const pool = getPool();
	const values: Array<number | Date> = [brokerId, instrumentId, lowerBound];
	let where = 'broker_id = $1 AND instrument_id = $2 AND ts <= $3';
	if (strictlyLessThanTs) {
		values.push(strictlyLessThanTs);
		where += ` AND ts < $${values.length}`;
	}
	const sql = `SELECT 1 FROM timeseries.ohlcv WHERE ${where} LIMIT 1`;
	const res = await pool.query(sql, values);
	return (res.rowCount ?? 0) > 0;
}
