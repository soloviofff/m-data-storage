import {
	pgSchema,
	serial,
	integer,
	text,
	timestamp,
	uniqueIndex,
	index,
	uuid,
	real,
	doublePrecision,
} from 'drizzle-orm/pg-core';

// Schemas
export const registry = pgSchema('registry');
export const timeseries = pgSchema('timeseries');

// Registry tables
export const brokers = registry.table('brokers', {
	id: serial('id').primaryKey(),
	code: text('code').notNull().unique(),
	name: text('name').notNull(),
});

export const instruments = registry.table('instruments', {
	id: serial('id').primaryKey(),
	symbol: text('symbol').notNull().unique(),
	name: text('name'),
});

export const instrumentMappings = registry.table(
	'instrument_mappings',
	{
		id: serial('id').primaryKey(),
		brokerId: integer('broker_id')
			.notNull()
			.references(() => brokers.id),
		instrumentId: integer('instrument_id')
			.notNull()
			.references(() => instruments.id),
		externalSymbol: text('external_symbol').notNull(),
	},
	(t) => ({
		uniq: uniqueIndex('instrument_mappings_broker_external_symbol_uniq').on(
			t.brokerId,
			t.externalSymbol,
		),
	}),
);

export const tasks = registry.table(
	'tasks',
	{
		id: uuid('id').primaryKey(),
		brokerId: integer('broker_id')
			.notNull()
			.references(() => brokers.id),
		instrumentId: integer('instrument_id')
			.notNull()
			.references(() => instruments.id),
		fromTs: timestamp('from_ts', { withTimezone: true }).notNull(),
		toTs: timestamp('to_ts', { withTimezone: true }).notNull(),
		status: text('status').notNull(), // queued | reserved | done | failed
		priority: integer('priority').default(0).notNull(),
		idempotencyKey: text('idempotency_key').notNull().unique(),
		reservedUntil: timestamp('reserved_until', { withTimezone: true }),
	},
	(t) => ({
		idxInstrument: index('tasks_instrument_id_idx').on(t.instrumentId),
	}),
);

export const audits = registry.table('audits', {
	id: serial('id').primaryKey(),
	ts: timestamp('ts', { withTimezone: true }).defaultNow().notNull(),
	actor: text('actor'),
	action: text('action'),
	details: text('details'),
});

// Timeseries table definition (logical); will be materialized by SQL migration as hypertable
export const ohlcv = timeseries.table(
	'ohlcv',
	{
		brokerId: integer('broker_id').notNull(),
		instrumentId: integer('instrument_id').notNull(),
		ts: timestamp('ts', { withTimezone: true }).notNull(),
		o: real('o'),
		h: real('h'),
		l: real('l'),
		c: real('c'),
		v: doublePrecision('v'),
	},
	(t) => ({
		idxInstrumentTs: index('ohlcv_instrument_ts_idx').on(t.instrumentId, t.ts),
		uniq: uniqueIndex('ohlcv_broker_instrument_ts_uniq').on(t.brokerId, t.instrumentId, t.ts),
	}),
);
