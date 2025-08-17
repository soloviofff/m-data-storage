import { extendZodWithOpenApi } from '@asteasolutions/zod-to-openapi';
import { z } from 'zod';

extendZodWithOpenApi(z);

export const OhlcvItemSchema = z.object({
	ts: z.string().openapi({ example: '2025-08-17T08:00:00.000Z' }),
	o: z.number(),
	h: z.number(),
	l: z.number(),
	c: z.number(),
	v: z.number(),
});

export const ReadResponseSchema = z.object({
	items: z.array(OhlcvItemSchema),
	nextPageToken: z.string().optional(),
});

export const IngestBodySchema = z.object({
	broker_id: z.number().int().positive(),
	instrument_id: z.number().int().positive(),
	items: z.array(
		z.object({
			ts: z.string().openapi({ example: '2025-08-17T08:00:00.000Z' }),
			o: z.number(),
			h: z.number(),
			l: z.number(),
			c: z.number(),
			v: z.number(),
		}),
	),
});

export const IngestResponseSchema = z
	.object({
		inserted: z.number().int().nonnegative(),
		updated: z.number().int().nonnegative(),
		skipped: z.number().int().nonnegative(),
	})
	.openapi('IngestResponse');

// Registry schemas
export const BrokerItemSchema = z.object({
	id: z.number().int().positive(),
	code: z.string(),
	name: z.string(),
	is_active: z.boolean().openapi({ description: 'Whether broker is active for scheduling' }),
});

export const InstrumentItemSchema = z.object({
	id: z.number().int().positive(),
	symbol: z.string(),
	name: z.string().nullable().optional(),
	is_active: z.boolean().openapi({ description: 'Whether instrument is active for scheduling' }),
});

export const BrokersUpsertBodySchema = z
	.array(
		z.object({
			code: z.string(),
			name: z.string(),
			isActive: z.boolean().optional(),
		}),
	)
	.openapi('BrokersUpsertBody');

export const InstrumentsUpsertBodySchema = z
	.array(
		z.object({
			symbol: z.string(),
			name: z.string().optional(),
			isActive: z.boolean().optional(),
		}),
	)
	.openapi('InstrumentsUpsertBody');

export const BrokersListResponseSchema = z
	.object({ items: z.array(BrokerItemSchema) })
	.openapi('BrokersListResponse');

export const InstrumentsListResponseSchema = z
	.object({ items: z.array(InstrumentItemSchema) })
	.openapi('InstrumentsListResponse');

export const RemovedCountResponseSchema = z
	.object({ removed: z.number().int().nonnegative() })
	.openapi('RemovedCountResponse');

// Task schemas
export const TaskItemSchema = z.object({
	id: z.string().uuid(),
	broker_id: z.number().int().positive(),
	instrument_id: z.number().int().positive(),
	from_ts: z.string(),
	to_ts: z.string(),
	status: z.enum(['queued', 'reserved', 'done', 'failed']),
	priority: z.number().int(),
	idempotency_key: z.string(),
	reserved_until: z.string().nullable(),
});

export const TasksListResponseSchema = z
	.object({ items: z.array(TaskItemSchema) })
	.openapi('TasksListResponse');

export const TaskCompleteBodySchema = z
	.object({ status: z.enum(['done', 'failed']) })
	.openapi('TaskCompleteBody');

export const TaskCompleteResponseSchema = z
	.object({ item: TaskItemSchema })
	.openapi('TaskCompleteResponse');
