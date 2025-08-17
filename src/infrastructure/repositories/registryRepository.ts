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
		where += ` AND m.broker_id = $${params.length}`;
	}
	const sql = `
		SELECT m.broker_id, m.instrument_id
		FROM registry.instrument_mappings m
		JOIN registry.brokers b ON b.id = m.broker_id
		JOIN registry.instruments i ON i.id = m.instrument_id
		WHERE ${where}
	`;
	const res = await pool.query(sql, params);
	return (res.rows ?? []) as BrokerInstrumentPair[];
}
