import { getDb, getPool } from '../src/infrastructure/db/client';
import { brokers, instruments } from '../src/infrastructure/db/schema';

async function main() {
	const db = getDb();
	// Upsert broker with system_name
	await db
		.insert(brokers)
		.values([{ systemName: 'demo', name: 'Demo Broker' }])
		.onConflictDoNothing();

	// Resolve broker id
	const rows = await db.select({ id: brokers.id }).from(brokers).where(brokers.systemName.eq('demo'));
	const brokerId = rows?.[0]?.id as number;

	// Upsert instruments for this broker
	await db
		.insert(instruments)
		.values([
			{ brokerId, symbol: 'BTCUSDT', name: 'Bitcoin/USDT' },
			{ brokerId, symbol: 'ETHUSDT', name: 'Ethereum/USDT' },
		])
		.onConflictDoNothing();

	await getPool().end();
}

main().catch((err) => {
	console.error(err);
	process.exit(1);
});


