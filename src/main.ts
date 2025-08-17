import { startServer } from './interfaces/http/server';

const main = async () => {
	await startServer();
};

main().catch((err) => {
	console.error(err);
	process.exit(1);
});
