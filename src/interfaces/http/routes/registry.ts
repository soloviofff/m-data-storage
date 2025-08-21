import { FastifyInstance } from 'fastify';
import { z } from 'zod';
import { getPool } from '../../../infrastructure/db/client';

const BrokersSchema = z.array(
	z.object({
		system_name: z.string().min(1),
		name: z.string().min(1),
		isActive: z.boolean().optional(),
	}),
);
const InstrumentsSchema = z.array(
	z.object({
		broker_system_name: z.string().min(1),
		symbol: z.string().min(1),
		name: z.string().optional(),
		isActive: z.boolean().optional(),
	}),
);

export async function registerRegistryRoutes(app: FastifyInstance) {
	app.post(
		'/v1/admin/brokers',
		{ schema: { summary: 'Upsert brokers', tags: ['registry'] } },
		async (req, reply) => {
			const parsed = BrokersSchema.safeParse(req.body);
			if (!parsed.success)
				return reply
					.code(400)
					.send({ code: 'BAD_REQUEST', message: 'Invalid brokers payload' });
			const pool = getPool();
			const values: Array<string | boolean> = [];
			const tuples = parsed.data
				.map((b, i) => {
					const base = i * 3;
					values.push(b.system_name, b.name, b.isActive ?? true);
					return `($${base + 1}, $${base + 2}, $${base + 3})`;
				})
				.join(',');
			await pool.query(
				`INSERT INTO registry.brokers (system_name, name, is_active) VALUES ${tuples}
				 ON CONFLICT (system_name) DO UPDATE SET name = EXCLUDED.name, is_active = EXCLUDED.is_active`,
				values,
			);
			return { ok: true };
		},
	);

	app.get(
		'/v1/admin/brokers',
		{ schema: { summary: 'List brokers', tags: ['registry'] } },
		async () => {
			const pool = getPool();
			const res = await pool.query(
				'SELECT system_name, name, is_active FROM registry.brokers ORDER BY id ASC',
			);
			return { items: res.rows };
		},
	);

	app.post(
		'/v1/admin/instruments',
		{ schema: { summary: 'Upsert instruments', tags: ['registry'] } },
		async (req, reply) => {
			const parsed = InstrumentsSchema.safeParse(req.body);
			if (!parsed.success)
				return reply
					.code(400)
					.send({ code: 'BAD_REQUEST', message: 'Invalid instruments payload' });
			const pool = getPool();
			const values: Array<string | null | boolean> = [];
			const tuples = parsed.data
				.map((i, idx) => {
					const base = idx * 4;
					values.push(i.broker_system_name, i.symbol, i.name ?? null, i.isActive ?? true);
					return `($${base + 1}, $${base + 2}, $${base + 3}, $${base + 4})`;
				})
				.join(',');
			await pool.query(
				`WITH src(system_name, symbol, name, is_active) AS (
					VALUES ${tuples}
				)
				INSERT INTO registry.instruments (broker_id, symbol, name, is_active)
				SELECT b.id, s.symbol, s.name, s.is_active
				FROM src s
				JOIN registry.brokers b ON b.system_name = s.system_name
				ON CONFLICT (broker_id, symbol) DO UPDATE SET name = EXCLUDED.name, is_active = EXCLUDED.is_active`,
				values,
			);
			return { ok: true };
		},
	);

	app.get(
		'/v1/admin/instruments',
		{ schema: { summary: 'List instruments', tags: ['registry'] } },
		async () => {
			const pool = getPool();
			const res = await pool.query(
				`SELECT b.system_name, i.symbol, i.name, i.is_active
				 FROM registry.instruments i
				 JOIN registry.brokers b ON b.id = i.broker_id
				 ORDER BY b.system_name ASC, i.symbol ASC`,
			);
			return { items: res.rows };
		},
	);

	// Stop watching all instruments for a broker: remove mappings for broker system name
	app.delete(
		'/v1/admin/watch/broker/:broker_system_name',
		{ schema: { summary: 'Unwatch all instruments for broker', tags: ['registry'] } },
		async (req, reply) => {
			const schema = z.object({ broker_system_name: z.string().min(1) });
			const parsed = schema.safeParse(req.params);
			if (!parsed.success)
				return reply
					.code(400)
					.send({ code: 'BAD_REQUEST', message: 'Invalid broker_system_name' });
			const { broker_system_name } = parsed.data;
			const pool = getPool();
			const res = await pool.query(
				`DELETE FROM registry.instrument_mappings m
				 USING registry.brokers b
				 WHERE m.broker_id = b.id AND b.system_name = $1`,
				[broker_system_name],
			);
			return { removed: res.rowCount ?? 0 };
		},
	);

	// Stop watching instrument across all brokers: remove mappings by instrument symbol for all brokers
	app.delete(
		'/v1/admin/watch/instrument/:instrument_symbol',
		{ schema: { summary: 'Unwatch instrument globally', tags: ['registry'] } },
		async (req, reply) => {
			const schema = z.object({ instrument_symbol: z.string().min(1) });
			const parsed = schema.safeParse(req.params);
			if (!parsed.success)
				return reply
					.code(400)
					.send({ code: 'BAD_REQUEST', message: 'Invalid instrument_symbol' });
			const { instrument_symbol } = parsed.data;
			const pool = getPool();
			const res = await pool.query(
				`DELETE FROM registry.instrument_mappings m
				 USING registry.instruments i
				 WHERE m.instrument_id = i.id AND i.symbol = $1`,
				[instrument_symbol],
			);
			return { removed: res.rowCount ?? 0 };
		},
	);

	// Stop watching a specific pair (broker, instrument)
	app.delete(
		'/v1/admin/watch/:broker_system_name/:instrument_symbol',
		{ schema: { summary: 'Unwatch specific broker-instrument pair', tags: ['registry'] } },
		async (req, reply) => {
			const schema = z.object({
				broker_system_name: z.string().min(1),
				instrument_symbol: z.string().min(1),
			});
			const parsed = schema.safeParse(req.params);
			if (!parsed.success)
				return reply.code(400).send({ code: 'BAD_REQUEST', message: 'Invalid params' });
			const { broker_system_name, instrument_symbol } = parsed.data;
			const pool = getPool();
			const res = await pool.query(
				`DELETE FROM registry.instrument_mappings m
				 USING registry.brokers b, registry.instruments i
				 WHERE m.broker_id = b.id AND m.instrument_id = i.id
				   AND b.system_name = $1 AND i.symbol = $2`,
				[broker_system_name, instrument_symbol],
			);
			return { removed: res.rowCount ?? 0 };
		},
	);
}
