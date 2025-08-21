import { FastifyInstance } from 'fastify';
import { z } from 'zod';
import {
	completeTask,
	reserveNextTasks,
} from '../../../infrastructure/repositories/taskRepository';
import { getPool } from '../../../infrastructure/db/client';

export async function registerTaskRoutes(app: FastifyInstance) {
	app.get(
		'/v1/tasks/next',
		{ schema: { summary: 'Reserve next tasks', tags: ['tasks'] } },
		async (req, reply) => {
			const schema = z.object({
				broker_system_name: z.string().min(1).optional(),
				instrument_symbols: z
					.string()
					.optional()
					.transform((v) =>
						v
							? v
									.split(',')
									.map((x) => x.trim())
									.filter(Boolean)
							: undefined,
					),
				limit: z.coerce.number().int().positive().max(100).default(10),
				leaseSeconds: z.coerce.number().int().positive().max(3600).default(60),
			});
			const parsed = schema.safeParse(req.query);
			if (!parsed.success) {
				return reply.code(400).send({
					code: 'BAD_REQUEST',
					message: 'Invalid query',
					details: parsed.error.flatten(),
				});
			}
			const { broker_system_name, instrument_symbols, limit, leaseSeconds } = parsed.data as {
				broker_system_name?: string;
				instrument_symbols?: string[];
				limit: number;
				leaseSeconds: number;
			};
			let brokerId: number | undefined;
			let instrumentIds: number[] | undefined;
			if (broker_system_name) {
				const pool = getPool();
				const br = await pool.query(
					'SELECT id FROM registry.brokers WHERE system_name = $1 AND is_active = true',
					[broker_system_name],
				);
				if ((br.rowCount ?? 0) === 0) {
					return reply
						.code(400)
						.send({ code: 'BAD_REQUEST', message: 'Unknown broker_system_name' });
				}
				brokerId = br.rows[0].id;
				if (instrument_symbols && instrument_symbols.length > 0) {
					const ins = await pool.query(
						'SELECT id FROM registry.instruments WHERE broker_id = $1 AND symbol = ANY($2::text[]) AND is_active = true',
						[brokerId, instrument_symbols],
					);
					instrumentIds = (ins.rows ?? []).map((r: { id: number }) => r.id);
				}
			}
			const items = await reserveNextTasks({
				brokerId,
				instrumentIds,
				limit,
				leaseSeconds,
			});
			// Map numeric ids to public identifiers
			if (!items || items.length === 0) return { items: [] };
			const pool = getPool();
			const uniqueBrokerIds = Array.from(new Set(items.map((i) => i.broker_id)));
			const uniqueInstrumentIds = Array.from(new Set(items.map((i) => i.instrument_id)));
			const br = await pool.query(
				'SELECT id, system_name FROM registry.brokers WHERE id = ANY($1::int[])',
				[uniqueBrokerIds],
			);
			const inss = await pool.query(
				'SELECT id, symbol FROM registry.instruments WHERE id = ANY($1::int[])',
				[uniqueInstrumentIds],
			);
			const brokerMap = new Map<number, string>(
				(br.rows ?? []).map((r: { id: number; system_name: string }) => [
					r.id,
					r.system_name,
				]),
			);
			const instrumentMap = new Map<number, string>(
				(inss.rows ?? []).map((r: { id: number; symbol: string }) => [r.id, r.symbol]),
			);
			const sanitized = items.map((i) => ({
				id: i.id,
				broker_system_name: brokerMap.get(i.broker_id),
				instrument_symbol: instrumentMap.get(i.instrument_id),
				from_ts: i.from_ts,
				to_ts: i.to_ts,
				status: i.status,
				priority: i.priority,
			}));
			return { items: sanitized };
		},
	);

	app.post(
		'/v1/tasks/:id/complete',
		{ schema: { summary: 'Complete reserved task', tags: ['tasks'] } },
		async (req, reply) => {
			const paramsSchema = z.object({ id: z.string().uuid() });
			const bodySchema = z.object({ status: z.enum(['done', 'failed']) });
			const params = paramsSchema.safeParse(req.params);
			const body = bodySchema.safeParse(req.body);
			if (!params.success || !body.success) {
				return reply.code(400).send({ code: 'BAD_REQUEST', message: 'Invalid request' });
			}
			const updated = await completeTask(params.data.id, body.data.status);
			if (!updated)
				return reply
					.code(409)
					.send({ code: 'CONFLICT', message: 'Task not reserved or not found' });
			// Map ids to public identifiers
			const pool = getPool();
			const br = await pool.query('SELECT system_name FROM registry.brokers WHERE id = $1', [
				updated.broker_id,
			]);
			const ins = await pool.query('SELECT symbol FROM registry.instruments WHERE id = $1', [
				updated.instrument_id,
			]);
			return {
				item: {
					id: updated.id,
					broker_system_name: br.rows?.[0]?.system_name,
					instrument_symbol: ins.rows?.[0]?.symbol,
					from_ts: updated.from_ts,
					to_ts: updated.to_ts,
					status: updated.status,
					priority: updated.priority,
				},
			};
		},
	);
}
