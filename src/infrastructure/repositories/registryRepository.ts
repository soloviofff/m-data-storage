import { getPool } from '../db/client';

export interface BrokerInstrumentPair {
	broker_id: number;
	instrument_id: number;
}

export async function listBrokerInstrumentPairs(
	brokerId?: number,
): Promise<BrokerInstrumentPair[]> {
	const pool = getPool();
	const sql = brokerId
		? 'SELECT broker_id, instrument_id FROM registry.instrument_mappings WHERE broker_id = $1'
		: 'SELECT broker_id, instrument_id FROM registry.instrument_mappings';
	const res = await pool.query(sql, brokerId ? [brokerId] : []);
	return (res.rows ?? []) as BrokerInstrumentPair[];
}
