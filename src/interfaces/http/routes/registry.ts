import { FastifyInstance } from 'fastify';
import { z } from 'zod';
import { getPool } from '../../../infrastructure/db/client';
import {
	BrokersUpsertBodySchema,
	InstrumentsUpsertBodySchema,
	BrokersListResponseSchema,
	InstrumentsListResponseSchema,
	RemovedCountResponseSchema,
} from '../openapi';

const BrokersSchema = z.array(
	z.object({
		code: z.string().min(1),
		name: z.string().min(1),
		isActive: z.boolean().optional(),
	}),
);
const InstrumentsSchema = z.array(
	z.object({
		symbol: z.string().min(1),
		name: z.string().optional(),
		isActive: z.boolean().optional(),
	}),
);

export async function registerRegistryRoutes(app: FastifyInstance) {
	app.post(
		'/v1/admin/brokers',
		{
			schema: {
				summary: 'Upsert brokers',
				tags: ['registry'],
				body: BrokersUpsertBodySchema,
				response: { 200: z.object({ ok: z.boolean() }) },
			},
		},
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
					values.push(b.code, b.name, b.isActive ?? true);
					return `($${base + 1}, $${base + 2}, $${base + 3})`;
				})
				.join(',');
			await pool.query(
				`INSERT INTO registry.brokers (code, name, is_active) VALUES ${tuples}
			 ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name, is_active = EXCLUDED.is_active`,
				values,
			);
			return { ok: true };
		},
	);

	app.get(
		'/v1/admin/brokers',
		{
			schema: {
				summary: 'List brokers',
				tags: ['registry'],
				response: { 200: BrokersListResponseSchema },
			},
		},
		async () => {
			const pool = getPool();
			const res = await pool.query(
				'SELECT id, code, name FROM registry.brokers ORDER BY id ASC',
			);
			return { items: res.rows };
		},
	);

	app.post(
		'/v1/admin/instruments',
		{
			schema: {
				summary: 'Upsert instruments',
				tags: ['registry'],
				body: InstrumentsUpsertBodySchema,
				response: { 200: z.object({ ok: z.boolean() }) },
			},
		},
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
					const base = idx * 3;
					values.push(i.symbol, i.name ?? null, i.isActive ?? true);
					return `($${base + 1}, $${base + 2}, $${base + 3})`;
				})
				.join(',');
			await pool.query(
				`INSERT INTO registry.instruments (symbol, name, is_active) VALUES ${tuples}
			 ON CONFLICT (symbol) DO UPDATE SET name = EXCLUDED.name, is_active = EXCLUDED.is_active`,
				values,
			);
			return { ok: true };
		},
	);

	app.get(
		'/v1/admin/instruments',
		{
			schema: {
				summary: 'List instruments',
				tags: ['registry'],
				response: { 200: InstrumentsListResponseSchema },
			},
		},
		async () => {
			const pool = getPool();
			const res = await pool.query(
				'SELECT id, symbol, name FROM registry.instruments ORDER BY id ASC',
			);
			return { items: res.rows };
		},
	);

	// Stop watching all instruments for a broker: remove mappings for broker_id
	app.delete(
		'/v1/admin/watch/broker/:broker_id',
		{
			schema: {
				summary: 'Unwatch all instruments for broker',
				tags: ['registry'],
				response: { 200: RemovedCountResponseSchema },
			},
		},
		async (req, reply) => {
			const schema = z.object({ broker_id: z.coerce.number().int().positive() });
			const parsed = schema.safeParse(req.params);
			if (!parsed.success)
				return reply.code(400).send({ code: 'BAD_REQUEST', message: 'Invalid broker_id' });
			const { broker_id } = parsed.data;
			const pool = getPool();
			const res = await pool.query(
				'DELETE FROM registry.instrument_mappings WHERE broker_id = $1',
				[broker_id],
			);
			return { removed: res.rowCount ?? 0 };
		},
	);

	// Stop watching instrument across all brokers: remove mappings for instrument_id
	app.delete(
		'/v1/admin/watch/instrument/:instrument_id',
		{
			schema: {
				summary: 'Unwatch instrument globally',
				tags: ['registry'],
				response: { 200: RemovedCountResponseSchema },
			},
		},
		async (req, reply) => {
			const schema = z.object({ instrument_id: z.coerce.number().int().positive() });
			const parsed = schema.safeParse(req.params);
			if (!parsed.success)
				return reply
					.code(400)
					.send({ code: 'BAD_REQUEST', message: 'Invalid instrument_id' });
			const { instrument_id } = parsed.data;
			const pool = getPool();
			const res = await pool.query(
				'DELETE FROM registry.instrument_mappings WHERE instrument_id = $1',
				[instrument_id],
			);
			return { removed: res.rowCount ?? 0 };
		},
	);

	// Stop watching a specific pair (broker, instrument)
	app.delete(
		'/v1/admin/watch/:broker_id/:instrument_id',
		{
			schema: {
				summary: 'Unwatch specific broker-instrument pair',
				tags: ['registry'],
				response: { 200: RemovedCountResponseSchema },
			},
		},
		async (req, reply) => {
			const schema = z.object({
				broker_id: z.coerce.number().int().positive(),
				instrument_id: z.coerce.number().int().positive(),
			});
			const parsed = schema.safeParse(req.params);
			if (!parsed.success)
				return reply.code(400).send({ code: 'BAD_REQUEST', message: 'Invalid ids' });
			const { broker_id, instrument_id } = parsed.data;
			const pool = getPool();
			const res = await pool.query(
				'DELETE FROM registry.instrument_mappings WHERE broker_id = $1 AND instrument_id = $2',
				[broker_id, instrument_id],
			);
			return { removed: res.rowCount ?? 0 };
		},
	);
}
