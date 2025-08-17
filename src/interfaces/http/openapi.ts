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
