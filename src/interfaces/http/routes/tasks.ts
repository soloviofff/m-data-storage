import { FastifyInstance } from 'fastify';
import { z } from 'zod';
import {
	completeTask,
	reserveNextTasks,
} from '../../../infrastructure/repositories/taskRepository';
import {
	TasksListResponseSchema,
	TaskCompleteBodySchema,
	TaskCompleteResponseSchema,
} from '../openapi';

export async function registerTaskRoutes(app: FastifyInstance) {
	app.get(
		'/v1/tasks/next',
		{
			schema: {
				summary: 'Reserve next tasks',
				tags: ['tasks'],
				response: { 200: TasksListResponseSchema },
			},
		},
		async (req, reply) => {
			const schema = z.object({
				broker_id: z.coerce.number().int().positive().optional(),
				instrument_ids: z
					.string()
					.optional()
					.transform((v) =>
						v
							? v
									.split(',')
									.map((x) => Number(x))
									.filter((n) => Number.isInteger(n) && n > 0)
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
			const { broker_id, instrument_ids, limit, leaseSeconds } = parsed.data as {
				broker_id?: number;
				instrument_ids?: number[];
				limit: number;
				leaseSeconds: number;
			};
			const items = await reserveNextTasks({
				brokerId: broker_id,
				instrumentIds: instrument_ids,
				limit,
				leaseSeconds,
			});
			return { items };
		},
	);

	app.post(
		'/v1/tasks/:id/complete',
		{
			schema: {
				summary: 'Complete reserved task',
				tags: ['tasks'],
				body: TaskCompleteBodySchema,
				response: {
					200: TaskCompleteResponseSchema,
					409: z.object({ code: z.string(), message: z.string() }),
				},
			},
		},
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
			return { item: updated };
		},
	);
}
