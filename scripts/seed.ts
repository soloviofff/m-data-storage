import { getDb, getPool } from '../src/infrastructure/db/client';
import { brokers, instruments } from '../src/infrastructure/db/schema';

async function main() {
	const db = getDb();
	await db.insert(brokers).values([
		{ code: 'demo', name: 'Demo Broker' },
	]).onConflictDoNothing();

	await db.insert(instruments).values([
		{ symbol: 'BTCUSDT', name: 'Bitcoin/USDT' },
		{ symbol: 'ETHUSDT', name: 'Ethereum/USDT' },
	]).onConflictDoNothing();

	await getPool().end();
}

main().catch((err) => {
	console.error(err);
	process.exit(1);
});


