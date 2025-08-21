import { getPool } from '../db/client';

export interface BrokerInstrumentPair {
	broker_id: number;
	instrument_id: number;
}

export async function listBrokerInstrumentPairs(
	brokerId?: number,
): Promise<BrokerInstrumentPair[]> {
	const pool = getPool();
	const params: Array<number> = [];
	let where = 'b.is_active = true AND i.is_active = true';
	if (typeof brokerId === 'number') {
		params.push(brokerId);
		where += ` AND i.broker_id = $${params.length}`;
	}
	const sql = `
		SELECT i.broker_id, i.id AS instrument_id
		FROM registry.instruments i
		JOIN registry.brokers b ON b.id = i.broker_id
		WHERE ${where}
	`;
	const res = await pool.query(sql, params);
	return (res.rows ?? []) as BrokerInstrumentPair[];
}
